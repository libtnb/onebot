package onebot

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FileMessageHelper 文件消息辅助工具
type FileMessageHelper struct {
	client *Client
}

// NewFileHelper 创建文件消息辅助工具
func (c *Client) FileHelper() *FileMessageHelper {
	return &FileMessageHelper{client: c}
}

// SendImageFromURL 通过 URL 发送图片
func (h *FileMessageHelper) SendImageFromURL(targetType, targetID, imageURL string, caption ...string) error {
	// 上传图片
	fileName := filepath.Base(imageURL)
	if fileName == "" || fileName == "." {
		fileName = "image.jpg"
	}
	
	uploadResp, err := h.client.UploadFile("image", fileName, imageURL)
	if err != nil {
		return fmt.Errorf("上传图片失败: %w", err)
	}
	
	// 构建消息
	msg := Message{}
	if len(caption) > 0 && caption[0] != "" {
		msg = append(msg, Text(caption[0]))
	}
	msg = append(msg, Image(uploadResp.FileID))
	
	// 发送消息
	return h.sendToTarget(targetType, targetID, msg)
}

// SendImageFromFile 发送本地图片文件
func (h *FileMessageHelper) SendImageFromFile(targetType, targetID, filePath string, caption ...string) error {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}
	
	// 转换为 base64
	base64Data := base64.StdEncoding.EncodeToString(data)
	fileName := filepath.Base(filePath)
	
	// 尝试通过 base64 发送（某些实现支持）
	msg := Message{}
	if len(caption) > 0 && caption[0] != "" {
		msg = append(msg, Text(caption[0]))
	}
	msg = append(msg, MessageSegment{
		Type: "image",
		Data: map[string]any{
			"file": "base64://" + base64Data,
			"name": fileName,
		},
	})
	
	return h.sendToTarget(targetType, targetID, msg)
}

// SendImageFromBase64 发送 base64 编码的图片
func (h *FileMessageHelper) SendImageFromBase64(targetType, targetID, base64Data string, caption ...string) error {
	// 构建消息
	msg := Message{}
	if len(caption) > 0 && caption[0] != "" {
		msg = append(msg, Text(caption[0]))
	}
	
	// 清理 base64 数据（移除可能的前缀）
	base64Data = strings.TrimPrefix(base64Data, "data:image/jpeg;base64,")
	base64Data = strings.TrimPrefix(base64Data, "data:image/png;base64,")
	base64Data = strings.TrimPrefix(base64Data, "data:image/gif;base64,")
	base64Data = strings.TrimPrefix(base64Data, "data:image/webp;base64,")
	
	msg = append(msg, MessageSegment{
		Type: "image",
		Data: map[string]any{
			"file": "base64://" + base64Data,
		},
	})
	
	return h.sendToTarget(targetType, targetID, msg)
}

// SendFileFromURL 通过 URL 发送文件
func (h *FileMessageHelper) SendFileFromURL(targetType, targetID, fileURL string, caption ...string) error {
	// 上传文件
	fileName := filepath.Base(fileURL)
	if fileName == "" || fileName == "." {
		fileName = "file"
	}
	
	uploadResp, err := h.client.UploadFile("file", fileName, fileURL)
	if err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}
	
	// 构建消息
	msg := Message{}
	if len(caption) > 0 && caption[0] != "" {
		msg = append(msg, Text(caption[0]))
	}
	msg = append(msg, File(uploadResp.FileID))
	
	return h.sendToTarget(targetType, targetID, msg)
}

// DownloadFile 下载文件到本地
func (h *FileMessageHelper) DownloadFile(fileID, fileType, savePath string) error {
	// 获取文件信息
	fileInfo, err := h.client.GetFile(fileID, fileType)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	
	if fileInfo.URL == "" {
		return fmt.Errorf("文件 URL 为空")
	}
	
	// 下载文件
	resp, err := http.Get(fileInfo.URL)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 确保目录存在
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	// 保存文件
	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer out.Close()
	
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	
	return nil
}

// GetImageFromMessage 从消息中提取图片
func (h *FileMessageHelper) GetImageFromMessage(msg *MessageEvent) []string {
	var images []string
	for _, seg := range msg.Message {
		if seg.Type == "image" {
			if fileID, ok := seg.Data["file_id"].(string); ok {
				images = append(images, fileID)
			}
		}
	}
	return images
}

// GetFileFromMessage 从消息中提取文件
func (h *FileMessageHelper) GetFileFromMessage(msg *MessageEvent) []string {
	var files []string
	for _, seg := range msg.Message {
		if seg.Type == "file" {
			if fileID, ok := seg.Data["file_id"].(string); ok {
				files = append(files, fileID)
			}
		}
	}
	return files
}

// GetMediaFromMessage 从消息中提取所有媒体文件（图片、语音、视频、文件）
func (h *FileMessageHelper) GetMediaFromMessage(msg *MessageEvent) map[string][]string {
	media := make(map[string][]string)
	
	for _, seg := range msg.Message {
		switch seg.Type {
		case "image", "voice", "audio", "video", "file":
			if fileID, ok := seg.Data["file_id"].(string); ok {
				media[seg.Type] = append(media[seg.Type], fileID)
			}
		}
	}
	
	return media
}

// sendToTarget 根据目标类型发送消息
func (h *FileMessageHelper) sendToTarget(targetType, targetID string, msg Message) error {
	switch targetType {
	case "private":
		_, err := h.client.SendPrivateMessage(targetID, msg)
		return err
	case "group":
		_, err := h.client.SendGroupMessage(targetID, msg)
		return err
	default:
		params := map[string]any{
			targetType + "_id": targetID,
			"message": msg,
		}
		_, err := h.client.SendMessage(targetType, params)
		return err
	}
}

// ImageBuilder 图片消息构建器
type ImageBuilder struct {
	segments Message
}

// NewImageBuilder 创建图片消息构建器
func NewImageBuilder() *ImageBuilder {
	return &ImageBuilder{
		segments: Message{},
	}
}

// AddText 添加文本
func (b *ImageBuilder) AddText(text string) *ImageBuilder {
	b.segments = append(b.segments, Text(text))
	return b
}

// AddImage 添加图片（通过 file_id）
func (b *ImageBuilder) AddImage(fileID string) *ImageBuilder {
	b.segments = append(b.segments, Image(fileID))
	return b
}

// AddImageURL 添加图片 URL（某些实现支持）
func (b *ImageBuilder) AddImageURL(url string) *ImageBuilder {
	b.segments = append(b.segments, MessageSegment{
		Type: "image",
		Data: map[string]any{
			"url": url,
		},
	})
	return b
}

// AddImageBase64 添加 base64 图片（某些实现支持）
func (b *ImageBuilder) AddImageBase64(base64Data string) *ImageBuilder {
	b.segments = append(b.segments, MessageSegment{
		Type: "image",
		Data: map[string]any{
			"file": "base64://" + base64Data,
		},
	})
	return b
}

// AddFile 添加文件
func (b *ImageBuilder) AddFile(fileID string) *ImageBuilder {
	b.segments = append(b.segments, File(fileID))
	return b
}

// Build 构建消息
func (b *ImageBuilder) Build() Message {
	return b.segments
}

// ExtractTextFromMessage 从消息中提取纯文本
func ExtractTextFromMessage(msg *MessageEvent) string {
	var text strings.Builder
	for _, seg := range msg.Message {
		if seg.Type == "text" {
			if t, ok := seg.Data["text"].(string); ok {
				text.WriteString(t)
			}
		}
	}
	return text.String()
}

// HasImage 检查消息是否包含图片
func HasImage(msg *MessageEvent) bool {
	for _, seg := range msg.Message {
		if seg.Type == "image" {
			return true
		}
	}
	return false
}

// HasFile 检查消息是否包含文件
func HasFile(msg *MessageEvent) bool {
	for _, seg := range msg.Message {
		if seg.Type == "file" {
			return true
		}
	}
	return false
}

// HasMedia 检查消息是否包含媒体文件
func HasMedia(msg *MessageEvent) bool {
	for _, seg := range msg.Message {
		switch seg.Type {
		case "image", "voice", "audio", "video", "file":
			return true
		}
	}
	return false
}
