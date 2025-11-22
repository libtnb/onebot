package onebot

import (
	"log/slog"
	"time"
)

// Option 客户端选项
type Option func(*Client)

// WithAccessToken 设置访问令牌
func WithAccessToken(token string) Option {
	return func(c *Client) {
		c.accessToken = token
	}
}

// WithSelf 设置机器人自身标识（用于多账号场景）
func WithSelf(platform, userID string) Option {
	return func(c *Client) {
		c.self = &Self{
			Platform: platform,
			UserID:   userID,
		}
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// WithReconnect 设置是否自动重连
func WithReconnect(enabled bool) Option {
	return func(c *Client) {
		c.reconnect = enabled
	}
}

// WithReconnectWait 设置重连等待时间
func WithReconnectWait(wait time.Duration) Option {
	return func(c *Client) {
		c.reconnectWait = wait
	}
}

// WithHeartbeat 设置心跳间隔
func WithHeartbeat(interval time.Duration) Option {
	return func(c *Client) {
		c.heartbeat = interval
	}
}

// WithTimeout 设置默认超时时间
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.timeout = timeout
	}
}
