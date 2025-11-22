package onebot

import (
	"encoding/json"
)

// ActionRequest 动作请求
type ActionRequest struct {
	Action string         `json:"action"`         // 动作名称
	Params map[string]any `json:"params"`         // 动作参数
	Echo   string         `json:"echo,omitempty"` // 用于标识请求的字符串
	Self   *Self          `json:"self,omitempty"` // 机器人自身标识（多账号时需要）
}

// ActionResponse 动作响应
type ActionResponse struct {
	Status  string `json:"status"`         // 执行状态：ok/failed
	Retcode int64  `json:"retcode"`        // 返回码
	Data    any    `json:"data"`           // 响应数据
	Message string `json:"message"`        // 错误信息
	Echo    string `json:"echo,omitempty"` // 原样返回请求中的 echo
}

// NewActionRequest 创建动作请求
func NewActionRequest(action string, params map[string]any) *ActionRequest {
	return &ActionRequest{
		Action: action,
		Params: params,
	}
}

// WithEcho 设置 echo
func (r *ActionRequest) WithEcho(echo string) *ActionRequest {
	r.Echo = echo
	return r
}

// WithSelf 设置机器人自身标识
func (r *ActionRequest) WithSelf(self *Self) *ActionRequest {
	r.Self = self
	return r
}

// IsOK 判断响应是否成功
func (r *ActionResponse) IsOK() bool {
	return r.Status == "ok" && r.Retcode == 0
}

// UnmarshalData 将响应数据解析到指定结构
func (r *ActionResponse) UnmarshalData(v any) error {
	data, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// 返回码常量
const (
	RetcodeOK                     = 0     // 成功
	RetcodeBadRequest             = 10001 // 无效的动作请求
	RetcodeUnsupportedAction      = 10002 // 不支持的动作
	RetcodeBadParam               = 10003 // 无效的动作请求参数
	RetcodeUnsupportedParam       = 10004 // 不支持的动作请求参数
	RetcodeUnsupportedSegment     = 10005 // 不支持的消息段类型
	RetcodeBadSegmentData         = 10006 // 无效的消息段参数
	RetcodeUnsupportedSegmentData = 10007 // 不支持的消息段参数
	RetcodeWhoAmI                 = 10101 // 未指定机器人账号
	RetcodeUnknownSelf            = 10102 // 未知的机器人账号
	RetcodeBadHandler             = 20001 // 动作处理器实现错误
	RetcodeInternalHandlerError   = 20002 // 动作处理器运行时异常
)

// 标准动作响应数据结构

// SendMessageResponse send_message 动作的响应数据
type SendMessageResponse struct {
	MessageID string  `json:"message_id"` // 消息 ID
	Time      float64 `json:"time"`       // 消息发送时间
}

// GetVersionResponse get_version 动作的响应数据
type GetVersionResponse struct {
	Impl          string `json:"impl"`           // OneBot 实现名称
	Version       string `json:"version"`        // OneBot 实现版本
	OneBotVersion string `json:"onebot_version"` // OneBot 标准版本
}

// GetStatusResponse get_status 动作的响应数据
type GetStatusResponse struct {
	Good bool `json:"good"` // 是否各项状态都符合预期
	Bots []struct {
		Self   Self `json:"self"`   // 机器人自身标识
		Online bool `json:"online"` // 是否在线
	} `json:"bots"` // 机器人账号状态列表
}

// GetSupportedActionsResponse get_supported_actions 动作的响应数据
type GetSupportedActionsResponse []string

// GetSelfInfoResponse get_self_info 动作的响应数据
type GetSelfInfoResponse struct {
	UserID   string `json:"user_id"`   // 用户 ID
	UserName string `json:"user_name"` // 用户名称/昵称
}

// GetUserInfoResponse get_user_info 动作的响应数据
type GetUserInfoResponse struct {
	UserID   string `json:"user_id"`   // 用户 ID
	UserName string `json:"user_name"` // 用户名称/昵称
}

// GetGroupInfoResponse get_group_info 动作的响应数据
type GetGroupInfoResponse struct {
	GroupID   string `json:"group_id"`   // 群 ID
	GroupName string `json:"group_name"` // 群名称
}

// GetGroupListResponse get_group_list 动作的响应数据
type GetGroupListResponse []GetGroupInfoResponse

// GetGroupMemberInfoResponse get_group_member_info 动作的响应数据
type GetGroupMemberInfoResponse struct {
	UserID   string `json:"user_id"`   // 用户 ID
	UserName string `json:"user_name"` // 用户名称/昵称
}

// GetGroupMemberListResponse get_group_member_list 动作的响应数据
type GetGroupMemberListResponse []GetGroupMemberInfoResponse

// GetFriendListResponse get_friend_list 动作的响应数据
type GetFriendListResponse []GetUserInfoResponse

// UploadFileResponse upload_file 动作的响应数据
type UploadFileResponse struct {
	FileID string `json:"file_id"` // 文件 ID
}

// GetFileResponse get_file 动作的响应数据
type GetFileResponse struct {
	Name string `json:"name"` // 文件名
	URL  string `json:"url"`  // 文件 URL
	// Headers map[string]string `json:"headers,omitempty"` // 下载时需要添加的请求头
	// Path    string            `json:"path,omitempty"`    // 文件路径（本地文件）
	// Data    []byte            `json:"data,omitempty"`    // 文件数据（base64 编码）
	// Sha256  string            `json:"sha256,omitempty"`  // 文件 SHA256 校验和
}
