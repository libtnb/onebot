package onebot

import (
	"encoding/json"
)

// Event OneBot 事件基础结构
type Event struct {
	ID         string  `json:"id"`          // 事件唯一标识符
	Self       *Self   `json:"self"`        // 机器人自身标识（元事件可为空）
	Time       float64 `json:"time"`        // 事件发生时间（Unix 时间戳）
	Type       string  `json:"type"`        // 事件类型：meta/message/notice/request
	DetailType string  `json:"detail_type"` // 事件详细类型
	SubType    string  `json:"sub_type"`    // 事件子类型
}

// MessageEvent 消息事件
type MessageEvent struct {
	Event
	MessageID  string  `json:"message_id"`         // 消息唯一 ID
	Message    Message `json:"message"`            // 消息内容
	AltMessage string  `json:"alt_message"`        // 消息内容的替代表示
	UserID     string  `json:"user_id"`            // 用户 ID
	GroupID    string  `json:"group_id,omitempty"` // 群 ID（群消息才有）
}

// NoticeEvent 通知事件
type NoticeEvent struct {
	Event
	UserID     string `json:"user_id,omitempty"`     // 用户 ID
	GroupID    string `json:"group_id,omitempty"`    // 群 ID
	OperatorID string `json:"operator_id,omitempty"` // 操作者 ID
}

// MetaEvent 元事件
type MetaEvent struct {
	Event
	Interval int64      `json:"interval,omitempty"` // 心跳间隔（毫秒）
	Status   *BotStatus `json:"status,omitempty"`   // 状态信息
}

// BotStatus 机器人状态
type BotStatus struct {
	Good bool `json:"good"` // 是否各项状态都符合预期
	Bots []struct {
		Self   Self `json:"self"`   // 机器人自身标识
		Online bool `json:"online"` // 是否在线
	} `json:"bots"` // 机器人账号状态列表
}

// RequestEvent 请求事件
type RequestEvent struct {
	Event
	UserID  string `json:"user_id,omitempty"`  // 用户 ID
	GroupID string `json:"group_id,omitempty"` // 群 ID
	Comment string `json:"comment,omitempty"`  // 验证信息
}

// ParseEvent 解析事件 JSON
func ParseEvent(data []byte) (any, error) {
	var base Event
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, err
	}

	switch base.Type {
	case "message":
		var event MessageEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil
	case "notice":
		var event NoticeEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil
	case "meta":
		var event MetaEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil
	case "request":
		var event RequestEvent
		if err := json.Unmarshal(data, &event); err != nil {
			return nil, err
		}
		return &event, nil
	default:
		// 返回基础事件结构，支持扩展事件类型
		return &base, nil
	}
}

// IsPrivateMessage 判断是否为私聊消息
func (e *MessageEvent) IsPrivateMessage() bool {
	return e.DetailType == "private"
}

// IsGroupMessage 判断是否为群消息
func (e *MessageEvent) IsGroupMessage() bool {
	return e.DetailType == "group"
}
