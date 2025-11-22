# OneBot Go 客户端库

[![Go Reference](https://pkg.go.dev/badge/github.com/libtnb/onebot.svg)](https://pkg.go.dev/github.com/libtnb/onebot)
[![Go Report Card](https://goreportcard.com/badge/github.com/libtnb/onebot)](https://goreportcard.com/report/github.com/libtnb/onebot)

一个符合 [OneBot 12](https://12.onebot.dev/) 标准的 Go 客户端库，用于连接和控制 OneBot 实现。

## 特性

- ✅ 完整支持 OneBot 12 标准协议
- ✅ 正向 WebSocket 连接
- ✅ 自动重连机制
- ✅ 结构化日志支持
- ✅ 类型安全的消息构造器
- ✅ 灵活的事件处理机制
- ✅ 支持多账号场景

## 安装

```bash
go get github.com/libtnb/onebot
```

## 快速开始

```go
package main

import (
    "log"
    "github.com/libtnb/onebot"
)

func main() {
    // 创建客户端
    client, err := onebot.New("ws://localhost:5700",
        onebot.WithAccessToken("your-token"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // 连接到 OneBot 实现
    if err := client.Connect(); err != nil {
        log.Fatal(err)
    }
    
    // 注册消息处理器
    client.On("message.private", func(event any) {
        msg := event.(*onebot.MessageEvent)
        log.Printf("收到私聊消息: %s", msg.AltMessage)
        
        // 回复消息
        reply := onebot.Message{
            onebot.Text("收到消息: " + msg.AltMessage),
        }
        client.SendPrivateMessage(msg.UserID, reply)
    })
    
    // 保持运行
    select {}
}
```

## 配置选项

```go
client, err := onebot.New("ws://localhost:5700",
    // 访问令牌
    onebot.WithAccessToken("your-token"),
    
    // 机器人标识（多账号场景）
    onebot.WithSelf("platform", "user_id"),
    
    // 自定义日志
    onebot.WithLogger(slog.Default()),
    
    // 自动重连
    onebot.WithReconnect(true),
    onebot.WithReconnectWait(5*time.Second),
    
    // 超时设置
    onebot.WithTimeout(30*time.Second),
)
```

## 事件处理

### 监听所有事件

```go
client.On("*", func(event any) {
    // 处理所有事件
})
```

### 监听特定类型事件

```go
// 私聊消息
client.On("message.private", func(event any) {
    msg := event.(*onebot.MessageEvent)
    // 处理私聊消息
})

// 群消息
client.On("message.group", func(event any) {
    msg := event.(*onebot.MessageEvent)
    // 处理群消息
})

// 通知事件
client.On("notice", func(event any) {
    notice := event.(*onebot.NoticeEvent)
    // 处理通知
})
```

## 消息构造

```go
// 纯文本
msg := onebot.Message{
    onebot.Text("Hello, World!"),
}

// 图片
msg := onebot.Message{
    onebot.Text("看这张图片: "),
    onebot.Image("file_id_123"),
}

// @某人
msg := onebot.Message{
    onebot.Mention("user_id"),
    onebot.Text(" 你好！"),
}

// 回复消息
msg := onebot.Message{
    onebot.Reply("message_id"),
    onebot.Text("回复内容"),
}

// 复杂消息
msg := onebot.Message{
    onebot.Text("位置分享: "),
    onebot.Location(31.032315, 121.447127, "上海交大", "闵行区东川路800号"),
}
```

## 发送消息

```go
// 发送私聊消息
resp, err := client.SendPrivateMessage("user_id", msg)

// 发送群消息
resp, err := client.SendGroupMessage("group_id", msg)

// 通用发送
params := map[string]any{
    "user_id": "123456",
    "message": msg,
}
resp, err := client.SendMessage("private", params)
```

## 调用动作

```go
// 获取版本信息
version, err := client.GetVersion()

// 获取状态
status, err := client.GetStatus()

// 获取支持的动作
actions, err := client.GetSupportedActions()

// 撤回消息
err := client.DeleteMessage("message_id")

// 自定义动作调用
params := map[string]any{
    "param1": "value1",
}
resp, err := client.Call("custom_action", params)
```

## 错误处理

```go
resp, err := client.SendPrivateMessage("user_id", msg)
if err != nil {
    // 网络或协议错误
    log.Printf("发送失败: %v", err)
    return
}

if !resp.IsOK() {
    // OneBot 返回的错误
    log.Printf("动作执行失败: %s (code: %d)", resp.Message, resp.Retcode)
}
```

## 完整示例

### 回声机器人

```go
client.On("message", func(event any) {
    msg := event.(*onebot.MessageEvent)
    
    // 构造回复
    reply := onebot.Message{
        onebot.Reply(msg.MessageID),
        onebot.Text("Echo: "),
    }
    reply = append(reply, msg.Message...)
    
    // 根据消息类型回复
    if msg.IsPrivateMessage() {
        client.SendPrivateMessage(msg.UserID, reply)
    } else if msg.IsGroupMessage() {
        client.SendGroupMessage(msg.GroupID, reply)
    }
})
```

### 命令机器人

```go
client.On("message", func(event any) {
    msg := event.(*onebot.MessageEvent)
    
    // 解析命令
    if len(msg.Message) > 0 && msg.Message[0].Type == "text" {
        text := msg.Message[0].Data["text"].(string)
        
        switch text {
        case "/help":
            // 发送帮助信息
        case "/status":
            // 获取并发送状态
        }
    }
})
```

## 协议支持

### 通信方式

- [x] 正向 WebSocket
- [ ] 反向 WebSocket（计划中）
- [ ] HTTP Webhook（计划中）

### 事件类型

- [x] 消息事件 (message)
- [x] 通知事件 (notice)
- [x] 请求事件 (request)
- [x] 元事件 (meta)

### 消息段

- [x] 纯文本 (text)
- [x] 提及 (mention)
- [x] 提及所有人 (mention_all)
- [x] 图片 (image)
- [x] 语音 (voice)
- [x] 音频 (audio)
- [x] 视频 (video)
- [x] 文件 (file)
- [x] 位置 (location)
- [x] 回复 (reply)

## 依赖

- Go 1.24+
- [github.com/coder/websocket](https://github.com/coder/websocket) - WebSocket 客户端
- [github.com/google/uuid](https://github.com/google/uuid) - UUID 生成

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [OneBot 12 标准](https://12.onebot.dev/)
- [OneBot 生态](https://onebot.dev/)
