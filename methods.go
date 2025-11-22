package onebot

import (
	"encoding/json"
	"fmt"
	"time"
)

// 消息相关方法

// SendPrivateMessage 发送私聊消息
func (c *Client) SendPrivateMessage(userID string, message Message) (*SendMessageResponse, error) {
	params := map[string]any{
		"detail_type": "private",
		"user_id":     userID,
		"message":     message,
	}

	resp, err := c.Call("send_message", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("发送消息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result SendMessageResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SendGroupMessage 发送群消息
func (c *Client) SendGroupMessage(groupID string, message Message) (*SendMessageResponse, error) {
	params := map[string]any{
		"detail_type": "group",
		"group_id":    groupID,
		"message":     message,
	}

	resp, err := c.Call("send_message", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("发送消息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result SendMessageResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SendMessage 通用发送消息
func (c *Client) SendMessage(detailType string, params map[string]any) (*SendMessageResponse, error) {
	params["detail_type"] = detailType

	resp, err := c.Call("send_message", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("发送消息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result SendMessageResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteMessage 撤回消息
func (c *Client) DeleteMessage(messageID string) error {
	params := map[string]any{
		"message_id": messageID,
	}

	resp, err := c.Call("delete_message", params)
	if err != nil {
		return err
	}

	if !resp.IsOK() {
		return fmt.Errorf("撤回消息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	return nil
}

// 元动作方法

// GetVersion 获取版本信息
func (c *Client) GetVersion() (*GetVersionResponse, error) {
	resp, err := c.Call("get_version", nil)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取版本失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result GetVersionResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetStatus 获取状态
func (c *Client) GetStatus() (*GetStatusResponse, error) {
	resp, err := c.Call("get_status", nil)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取状态失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result GetStatusResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetSupportedActions 获取支持的动作列表
func (c *Client) GetSupportedActions() ([]string, error) {
	resp, err := c.Call("get_supported_actions", nil)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取支持的动作失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result []string
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetLatestEvents 获取最新事件（仅 HTTP 通信方式支持）
func (c *Client) GetLatestEvents(limit int, timeout time.Duration) ([]any, error) {
	params := map[string]any{
		"limit":   limit,
		"timeout": int64(timeout.Seconds()),
	}

	resp, err := c.CallWithTimeout("get_latest_events", params, timeout+5*time.Second)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取最新事件失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	// 解析事件列表
	if data, ok := resp.Data.([]any); ok {
		var events []any
		for _, item := range data {
			if eventData, err := json.Marshal(item); err == nil {
				if event, err := ParseEvent(eventData); err == nil {
					events = append(events, event)
				}
			}
		}
		return events, nil
	}

	return nil, fmt.Errorf("响应数据格式错误")
}

// 用户信息方法

// GetSelfInfo 获取机器人自身信息
func (c *Client) GetSelfInfo() (*GetSelfInfoResponse, error) {
	resp, err := c.Call("get_self_info", nil)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取自身信息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result GetSelfInfoResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetUserInfo 获取用户信息
func (c *Client) GetUserInfo(userID string) (*GetUserInfoResponse, error) {
	params := map[string]any{
		"user_id": userID,
	}

	resp, err := c.Call("get_user_info", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取用户信息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result GetUserInfoResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFriendList 获取好友列表
func (c *Client) GetFriendList() ([]GetUserInfoResponse, error) {
	resp, err := c.Call("get_friend_list", nil)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取好友列表失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result []GetUserInfoResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// 群组相关方法

// GetGroupInfo 获取群信息
func (c *Client) GetGroupInfo(groupID string) (*GetGroupInfoResponse, error) {
	params := map[string]any{
		"group_id": groupID,
	}

	resp, err := c.Call("get_group_info", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取群信息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result GetGroupInfoResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetGroupList 获取群列表
func (c *Client) GetGroupList() ([]GetGroupInfoResponse, error) {
	resp, err := c.Call("get_group_list", nil)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取群列表失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result []GetGroupInfoResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// GetGroupMemberInfo 获取群成员信息
func (c *Client) GetGroupMemberInfo(groupID, userID string) (*GetGroupMemberInfoResponse, error) {
	params := map[string]any{
		"group_id": groupID,
		"user_id":  userID,
	}

	resp, err := c.Call("get_group_member_info", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取群成员信息失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result GetGroupMemberInfoResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetGroupMemberList 获取群成员列表
func (c *Client) GetGroupMemberList(groupID string) ([]GetGroupMemberInfoResponse, error) {
	params := map[string]any{
		"group_id": groupID,
	}

	resp, err := c.Call("get_group_member_list", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取群成员列表失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result []GetGroupMemberInfoResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// 文件相关方法

// UploadFile 上传文件
func (c *Client) UploadFile(fileType string, name string, url string) (*UploadFileResponse, error) {
	params := map[string]any{
		"type": fileType,
		"name": name,
		"url":  url,
	}

	resp, err := c.Call("upload_file", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("上传文件失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result UploadFileResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetFile 获取文件
func (c *Client) GetFile(fileID string, fileType string) (*GetFileResponse, error) {
	params := map[string]any{
		"file_id": fileID,
		"type":    fileType,
	}

	resp, err := c.Call("get_file", params)
	if err != nil {
		return nil, err
	}

	if !resp.IsOK() {
		return nil, fmt.Errorf("获取文件失败: %s (code: %d)", resp.Message, resp.Retcode)
	}

	var result GetFileResponse
	if err := resp.UnmarshalData(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
