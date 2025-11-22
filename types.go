// Package onebot 提供 OneBot 12 标准的客户端实现
package onebot

import (
	"encoding/json"
	"time"
)

// Self 机器人自身标识
type Self struct {
	Platform string `json:"platform"` // 机器人平台名称
	UserID   string `json:"user_id"`  // 机器人用户 ID
}

// MessageSegment 消息段
type MessageSegment struct {
	Type string         `json:"type"` // 消息段类型
	Data map[string]any `json:"data"` // 消息段数据
}

// Message 消息类型，消息段数组
type Message []MessageSegment

// UnmarshalJSON 自定义反序列化，支持字符串和数组两种格式
func (m *Message) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*m = Message{{
			Type: "text",
			Data: map[string]any{"text": str},
		}}
		return nil
	}

	// 解析为消息段数组
	return json.Unmarshal(data, (*[]MessageSegment)(m))
}

// ToAltMessage 将消息转换为替代文本表示
func (m Message) ToAltMessage() string {
	var result string
	for _, seg := range m {
		switch seg.Type {
		case "text":
			if text, ok := seg.Data["text"].(string); ok {
				result += text
			}
		case "image":
			result += "[图片]"
		case "voice":
			result += "[语音]"
		case "audio":
			result += "[音频]"
		case "video":
			result += "[视频]"
		case "file":
			result += "[文件]"
		case "location":
			result += "[位置]"
		case "reply":
			result += "[回复]"
		case "mention":
			if userID, ok := seg.Data["user_id"].(string); ok {
				result += "@" + userID
			}
		case "mention_all":
			result += "@全体成员"
		default:
			result += "[" + seg.Type + "]"
		}
	}
	return result
}

// Timestamp 返回当前时间戳（浮点秒）
func Timestamp() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// 消息段构造函数

// Text 创建纯文本消息段
func Text(text string) MessageSegment {
	return MessageSegment{
		Type: "text",
		Data: map[string]any{
			"text": text,
		},
	}
}

// Image 创建图片消息段
func Image(fileID string) MessageSegment {
	return MessageSegment{
		Type: "image",
		Data: map[string]any{
			"file_id": fileID,
		},
	}
}

// Voice 创建语音消息段
func Voice(fileID string) MessageSegment {
	return MessageSegment{
		Type: "voice",
		Data: map[string]any{
			"file_id": fileID,
		},
	}
}

// Audio 创建音频消息段
func Audio(fileID string) MessageSegment {
	return MessageSegment{
		Type: "audio",
		Data: map[string]any{
			"file_id": fileID,
		},
	}
}

// Video 创建视频消息段
func Video(fileID string) MessageSegment {
	return MessageSegment{
		Type: "video",
		Data: map[string]any{
			"file_id": fileID,
		},
	}
}

// File 创建文件消息段
func File(fileID string) MessageSegment {
	return MessageSegment{
		Type: "file",
		Data: map[string]any{
			"file_id": fileID,
		},
	}
}

// Location 创建位置消息段
func Location(latitude, longitude float64, title, content string) MessageSegment {
	return MessageSegment{
		Type: "location",
		Data: map[string]any{
			"latitude":  latitude,
			"longitude": longitude,
			"title":     title,
			"content":   content,
		},
	}
}

// Reply 创建回复消息段
func Reply(messageID string, userID ...string) MessageSegment {
	data := map[string]any{
		"message_id": messageID,
	}
	if len(userID) > 0 {
		data["user_id"] = userID[0]
	}
	return MessageSegment{
		Type: "reply",
		Data: data,
	}
}

// Mention 创建提及（@）消息段
func Mention(userID string) MessageSegment {
	return MessageSegment{
		Type: "mention",
		Data: map[string]any{
			"user_id": userID,
		},
	}
}

// MentionAll 创建提及所有人消息段
func MentionAll() MessageSegment {
	return MessageSegment{
		Type: "mention_all",
		Data: map[string]any{},
	}
}
