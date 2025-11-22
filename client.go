package onebot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

// Client OneBot 12 客户端
type Client struct {
	// 配置
	url           string
	accessToken   string
	self          *Self
	logger        *slog.Logger
	reconnect     bool
	reconnectWait time.Duration
	heartbeat     time.Duration
	timeout       time.Duration

	// 运行时状态
	conn      *websocket.Conn
	mu        sync.RWMutex
	closed    atomic.Bool
	connected atomic.Bool

	// 事件处理
	eventHandlers map[string][]EventHandler
	handlerMu     sync.RWMutex

	// 动作处理
	actionChan   chan *actionCall
	responseChan map[string]chan *ActionResponse
	responseMu   sync.RWMutex

	// 上下文
	ctx    context.Context
	cancel context.CancelFunc
}

// EventHandler 事件处理函数
type EventHandler func(event any)

// actionCall 内部动作调用结构
type actionCall struct {
	request  *ActionRequest
	response chan *ActionResponse
}

// New 创建新的 OneBot 客户端
func New(url string, opts ...Option) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		url:           url,
		logger:        slog.Default(),
		reconnect:     true,
		reconnectWait: 5 * time.Second,
		heartbeat:     30 * time.Second,
		timeout:       30 * time.Second,
		eventHandlers: make(map[string][]EventHandler),
		actionChan:    make(chan *actionCall, 100),
		responseChan:  make(map[string]chan *ActionResponse),
		ctx:           ctx,
		cancel:        cancel,
	}

	// 应用选项
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// Connect 连接到 OneBot 实现
func (c *Client) Connect() error {
	return c.connect()
}

// connect 内部连接方法
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.CloseNow()
	}

	// 设置请求头
	opts := &websocket.DialOptions{
		HTTPHeader: make(map[string][]string),
	}
	if c.accessToken != "" {
		opts.HTTPHeader["Authorization"] = []string{"Bearer " + c.accessToken}
	}

	// 连接 WebSocket
	conn, _, err := websocket.Dial(c.ctx, c.url, opts)
	if err != nil {
		return fmt.Errorf("连接 WebSocket 失败: %w", err)
	}

	c.conn = conn
	c.connected.Store(true)
	c.logger.Info("已连接到 OneBot 实现", "url", c.url)

	// 启动读写协程
	go c.readLoop()
	go c.writeLoop()

	return nil
}

// readLoop 读取消息循环
func (c *Client) readLoop() {
	defer func() {
		c.connected.Store(false)
		c.handleDisconnect()
	}()

	for {
		_, data, err := c.conn.Read(c.ctx)
		if err != nil {
			if !c.closed.Load() {
				c.logger.Error("读取消息失败", "error", err)
			}
			return
		}

		// 尝试解析为动作响应
		var response ActionResponse
		if err := json.Unmarshal(data, &response); err == nil && response.Echo != "" {
			c.handleActionResponse(&response)
			continue
		}

		// 解析为事件
		event, err := ParseEvent(data)
		if err != nil {
			c.logger.Error("解析事件失败", "error", err, "data", string(data))
			continue
		}

		c.handleEvent(event)
	}
}

// writeLoop 写入消息循环
func (c *Client) writeLoop() {
	ticker := time.NewTicker(c.heartbeat)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return

		case <-ticker.C:
			// 发送心跳（如果需要）
			// OneBot 12 标准中心跳是通过元事件推送的，不需要主动发送

		case call := <-c.actionChan:
			// 设置 echo
			if call.request.Echo == "" {
				call.request.Echo = uuid.New().String()
			}

			// 注册响应通道
			c.responseMu.Lock()
			c.responseChan[call.request.Echo] = call.response
			c.responseMu.Unlock()

			// 发送请求
			data, err := json.Marshal(call.request)
			if err != nil {
				c.logger.Error("序列化动作请求失败", "error", err)
				call.response <- &ActionResponse{
					Status:  "failed",
					Retcode: RetcodeBadRequest,
					Message: err.Error(),
				}
				continue
			}

			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn != nil {
				err = conn.Write(c.ctx, websocket.MessageText, data)
				if err != nil {
					c.logger.Error("发送动作请求失败", "error", err)
					call.response <- &ActionResponse{
						Status:  "failed",
						Retcode: RetcodeInternalHandlerError,
						Message: err.Error(),
					}
				} else {
					c.logger.Debug("发送动作请求", "action", call.request.Action)
				}
			}
		}
	}
}

// handleEvent 处理事件
func (c *Client) handleEvent(event any) {
	c.handlerMu.RLock()
	defer c.handlerMu.RUnlock()

	// 获取事件类型
	var eventType string
	switch e := event.(type) {
	case *MessageEvent:
		eventType = "message"
		if e.DetailType != "" {
			eventType = "message." + e.DetailType
		}
		c.logger.Info("收到消息事件",
			"type", e.DetailType,
			"user_id", e.UserID,
			"message", e.AltMessage)
	case *NoticeEvent:
		eventType = "notice"
		if e.DetailType != "" {
			eventType = "notice." + e.DetailType
		}
		c.logger.Info("收到通知事件", "type", e.DetailType)
	case *MetaEvent:
		eventType = "meta"
		if e.DetailType != "" {
			eventType = "meta." + e.DetailType
		}
		c.logger.Debug("收到元事件", "type", e.DetailType)
	case *RequestEvent:
		eventType = "request"
		if e.DetailType != "" {
			eventType = "request." + e.DetailType
		}
		c.logger.Info("收到请求事件", "type", e.DetailType)
	case *Event:
		eventType = e.Type
		if e.DetailType != "" {
			eventType = e.Type + "." + e.DetailType
		}
		c.logger.Info("收到事件", "type", eventType)
	}

	// 调用通用处理器
	if handlers, ok := c.eventHandlers["*"]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}

	// 调用类型处理器
	if handlers, ok := c.eventHandlers[eventType]; ok {
		for _, handler := range handlers {
			go handler(event)
		}
	}
}

// handleActionResponse 处理动作响应
func (c *Client) handleActionResponse(response *ActionResponse) {
	c.responseMu.Lock()
	ch, ok := c.responseChan[response.Echo]
	if ok {
		delete(c.responseChan, response.Echo)
	}
	c.responseMu.Unlock()

	if ok && ch != nil {
		c.logger.Debug("收到动作响应",
			"echo", response.Echo,
			"status", response.Status,
			"retcode", response.Retcode)
		select {
		case ch <- response:
		case <-time.After(time.Second):
			c.logger.Error("发送响应到通道超时", "echo", response.Echo)
		}
	} else {
		c.logger.Warn("收到未知的动作响应", "echo", response.Echo)
	}
}

// handleDisconnect 处理断开连接
func (c *Client) handleDisconnect() {
	if c.closed.Load() {
		return
	}

	c.logger.Warn("WebSocket 连接断开")

	if c.reconnect {
		go func() {
			for !c.closed.Load() {
				c.logger.Info("尝试重连...", "wait", c.reconnectWait)
				time.Sleep(c.reconnectWait)

				if err := c.connect(); err != nil {
					c.logger.Error("重连失败", "error", err)
				} else {
					break
				}
			}
		}()
	}
}

// Close 关闭客户端
func (c *Client) Close() error {
	c.closed.Store(true)
	c.cancel()

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close(websocket.StatusNormalClosure, "客户端关闭")
	}

	return nil
}

// IsConnected 检查是否已连接
func (c *Client) IsConnected() bool {
	return c.connected.Load()
}

// On 注册事件处理器
// eventType 可以是 "*"（所有事件）、"message"（所有消息）、"message.private"（私聊消息）等
func (c *Client) On(eventType string, handler EventHandler) {
	c.handlerMu.Lock()
	defer c.handlerMu.Unlock()

	c.eventHandlers[eventType] = append(c.eventHandlers[eventType], handler)
}

// Call 调用动作
func (c *Client) Call(action string, params map[string]any) (*ActionResponse, error) {
	return c.CallWithTimeout(action, params, c.timeout)
}

// CallWithTimeout 调用动作（带超时）
func (c *Client) CallWithTimeout(action string, params map[string]any, timeout time.Duration) (*ActionResponse, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("未连接到 OneBot 实现")
	}

	request := NewActionRequest(action, params)
	if c.self != nil {
		request.WithSelf(c.self)
	}

	call := &actionCall{
		request:  request,
		response: make(chan *ActionResponse, 1),
	}

	select {
	case c.actionChan <- call:
	case <-time.After(timeout):
		return nil, fmt.Errorf("发送动作请求超时")
	}

	select {
	case response := <-call.response:
		return response, nil
	case <-time.After(timeout):
		// 清理响应通道
		c.responseMu.Lock()
		delete(c.responseChan, request.Echo)
		c.responseMu.Unlock()
		return nil, fmt.Errorf("等待动作响应超时")
	}
}
