package onebot

import "fmt"

// Error OneBot 错误
type Error struct {
	Code    int64  // 错误码
	Message string // 错误信息
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return fmt.Sprintf("onebot error %d: %s", e.Code, e.Message)
}

// NewError 创建新错误
func NewError(code int64, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// IsError 检查是否为特定错误码
func IsError(err error, code int64) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}

// 预定义错误
var (
	ErrNotConnected     = NewError(-1, "未连接到 OneBot 实现")
	ErrTimeout          = NewError(-2, "操作超时")
	ErrInvalidResponse  = NewError(-3, "无效的响应")
)
