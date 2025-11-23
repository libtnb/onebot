package example

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/libtnb/onebot"
)

// Example_FileAndImageHandler 展示如何处理文件和图片消息
func Example_FileAndImageHandler() {
	// 创建客户端
	client, err := onebot.New("ws://localhost:5700",
		onebot.WithAccessToken("your-token"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 连接
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 处理接收到的消息
	client.On("message", func(event any) {
		msg := event.(*onebot.MessageEvent)

		// 遍历消息段，查找文件和图片
		for _, seg := range msg.Message {
			switch seg.Type {
			case "image":
				handleImageMessage(client, msg, seg)
			case "file":
				handleFileMessage(client, msg, seg)
			case "voice":
				handleVoiceMessage(client, msg, seg)
			case "video":
				handleVideoMessage(client, msg, seg)
			}
		}
	})

	select {}
}

// handleImageMessage 处理图片消息
func handleImageMessage(client *onebot.Client, msg *onebot.MessageEvent, seg onebot.MessageSegment) {
	fileID, _ := seg.Data["file_id"].(string)
	log.Printf("收到图片消息: file_id=%s", fileID)

	// 获取图片文件信息
	fileInfo, err := client.GetFile(fileID, "image")
	if err != nil {
		log.Printf("获取图片信息失败: %v", err)
		return
	}

	log.Printf("图片信息: name=%s, url=%s", fileInfo.Name, fileInfo.URL)

	// 如果有 URL，可以下载图片
	if fileInfo.URL != "" {
		if err := downloadFile(fileInfo.URL, "./downloads/"+fileInfo.Name); err != nil {
			log.Printf("下载图片失败: %v", err)
		} else {
			log.Printf("图片已下载到: ./downloads/%s", fileInfo.Name)
		}
	}

	// 回复确认消息
	reply := onebot.Message{
		onebot.Text("收到图片: "),
		onebot.Image(fileID), // 转发同一张图片
	}

	if msg.IsPrivateMessage() {
		client.SendPrivateMessage(msg.UserID, reply)
	} else if msg.IsGroupMessage() {
		client.SendGroupMessage(msg.GroupID, reply)
	}
}

// handleFileMessage 处理文件消息
func handleFileMessage(client *onebot.Client, msg *onebot.MessageEvent, seg onebot.MessageSegment) {
	fileID, _ := seg.Data["file_id"].(string)
	log.Printf("收到文件消息: file_id=%s", fileID)

	// 获取文件信息
	fileInfo, err := client.GetFile(fileID, "file")
	if err != nil {
		log.Printf("获取文件信息失败: %v", err)
		return
	}

	log.Printf("文件信息: name=%s, url=%s", fileInfo.Name, fileInfo.URL)

	// 回复确认消息
	reply := onebot.Message{
		onebot.Text(fmt.Sprintf("已收到文件: %s", fileInfo.Name)),
	}

	if msg.IsPrivateMessage() {
		client.SendPrivateMessage(msg.UserID, reply)
	} else if msg.IsGroupMessage() {
		client.SendGroupMessage(msg.GroupID, reply)
	}
}

// handleVoiceMessage 处理语音消息
func handleVoiceMessage(client *onebot.Client, msg *onebot.MessageEvent, seg onebot.MessageSegment) {
	fileID, _ := seg.Data["file_id"].(string)
	log.Printf("收到语音消息: file_id=%s", fileID)

	// 回复确认
	reply := onebot.Message{
		onebot.Text("收到语音消息"),
	}

	if msg.IsPrivateMessage() {
		client.SendPrivateMessage(msg.UserID, reply)
	} else if msg.IsGroupMessage() {
		client.SendGroupMessage(msg.GroupID, reply)
	}
}

// handleVideoMessage 处理视频消息
func handleVideoMessage(client *onebot.Client, msg *onebot.MessageEvent, seg onebot.MessageSegment) {
	fileID, _ := seg.Data["file_id"].(string)
	log.Printf("收到视频消息: file_id=%s", fileID)

	// 回复确认
	reply := onebot.Message{
		onebot.Text("收到视频消息"),
	}

	if msg.IsPrivateMessage() {
		client.SendPrivateMessage(msg.UserID, reply)
	} else if msg.IsGroupMessage() {
		client.SendGroupMessage(msg.GroupID, reply)
	}
}

// Example_SendFiles 展示如何发送文件和图片
func Example_SendFiles() {
	client, _ := onebot.New("ws://localhost:5700")
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 方法1: 通过 URL 上传文件并发送
	sendImageByURL(client, "user123", "https://example.com/image.jpg")

	// 方法2: 使用已有的 file_id 发送
	sendImageByFileID(client, "user123", "existing_file_id_123")

	// 方法3: 通过本地文件发送（需要先上传）
	sendLocalFile(client, "user123", "./local_image.jpg")

	// 方法4: 发送 base64 编码的图片（某些 OneBot 实现支持）
	sendBase64Image(client, "user123", "./local_image.jpg")

	// 方法5: 组合消息（文字+图片+文件）
	sendMixedMessage(client, "group123")
}

// sendImageByURL 通过 URL 发送图片
func sendImageByURL(client *onebot.Client, userID string, imageURL string) {
	// 先上传文件获取 file_id
	uploadResp, err := client.UploadFile("image", "image.jpg", imageURL)
	if err != nil {
		log.Printf("上传图片失败: %v", err)
		return
	}

	// 使用获取到的 file_id 发送图片
	msg := onebot.Message{
		onebot.Text("这是一张通过 URL 上传的图片："),
		onebot.Image(uploadResp.FileID),
	}

	if _, err := client.SendPrivateMessage(userID, msg); err != nil {
		log.Printf("发送图片消息失败: %v", err)
	}
}

// sendImageByFileID 使用已有的 file_id 发送图片
func sendImageByFileID(client *onebot.Client, userID string, fileID string) {
	msg := onebot.Message{
		onebot.Text("这是一张使用 file_id 的图片："),
		onebot.Image(fileID),
	}

	if _, err := client.SendPrivateMessage(userID, msg); err != nil {
		log.Printf("发送图片消息失败: %v", err)
	}
}

// sendLocalFile 发送本地文件
func sendLocalFile(client *onebot.Client, userID string, filePath string) {
	// 方法1: 如果 OneBot 实现支持本地文件路径
	// 某些实现可能支持 file:// 协议
	absPath, _ := filepath.Abs(filePath)
	fileURL := "file://" + absPath
	uploadResp, err := client.UploadFile("image", filepath.Base(filePath), fileURL)
	if err != nil {
		log.Printf("上传本地文件失败: %v", err)
		
		// 方法2: 启动临时 HTTP 服务器提供文件
		serveLocalFile(client, userID, filePath)
		return
	}

	msg := onebot.Message{
		onebot.Text("本地图片："),
		onebot.Image(uploadResp.FileID),
	}
	client.SendPrivateMessage(userID, msg)
}

// serveLocalFile 通过临时 HTTP 服务器提供本地文件
func serveLocalFile(client *onebot.Client, userID string, filePath string) {
	// 启动临时服务器
	port := "18080"
	fileName := filepath.Base(filePath)
	
	http.HandleFunc("/"+fileName, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	})

	go http.ListenAndServe(":"+port, nil)
	time.Sleep(100 * time.Millisecond) // 等待服务器启动

	// 通过 HTTP URL 上传
	fileURL := fmt.Sprintf("http://localhost:%s/%s", port, fileName)
	uploadResp, err := client.UploadFile("image", fileName, fileURL)
	if err != nil {
		log.Printf("通过 HTTP 服务器上传失败: %v", err)
		return
	}

	msg := onebot.Message{
		onebot.Text("本地图片（通过 HTTP 服务）："),
		onebot.Image(uploadResp.FileID),
	}
	client.SendPrivateMessage(userID, msg)
}

// sendBase64Image 发送 base64 编码的图片
func sendBase64Image(client *onebot.Client, userID string, filePath string) {
	// 读取文件
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("读取文件失败: %v", err)
		return
	}

	// base64 编码
	base64Data := base64.StdEncoding.EncodeToString(data)
	
	// 某些 OneBot 实现支持直接发送 base64 数据
	// 注意：这是扩展功能，不是标准的一部分
	msg := onebot.Message{
		onebot.Text("Base64 图片："),
		{
			Type: "image",
			Data: map[string]any{
				"file": "base64://" + base64Data,
				// 或者某些实现使用
				// "data": base64Data,
				// "type": "base64",
			},
		},
	}

	if _, err := client.SendPrivateMessage(userID, msg); err != nil {
		log.Printf("发送 base64 图片失败（可能不支持）: %v", err)
	}
}

// sendMixedMessage 发送混合消息
func sendMixedMessage(client *onebot.Client, groupID string) {
	// 假设已经有一些 file_id
	imageFileID := "image_123"
	fileFileID := "file_456"
	voiceFileID := "voice_789"

	msg := onebot.Message{
		onebot.Text("这是一条包含多种媒体的消息：\n"),
		onebot.Text("1. 图片："),
		onebot.Image(imageFileID),
		onebot.Text("\n2. 文件："),
		onebot.File(fileFileID),
		onebot.Text("\n3. 语音："),
		onebot.Voice(voiceFileID),
		onebot.Text("\n请查收！"),
	}

	if _, err := client.SendGroupMessage(groupID, msg); err != nil {
		log.Printf("发送混合消息失败: %v", err)
	}
}

// Example_FileBot 完整的文件处理机器人示例
func Example_FileBot() {
	client, _ := onebot.New("ws://localhost:5700")
	defer client.Close()

	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 创建下载目录
	os.MkdirAll("./downloads", 0755)
	os.MkdirAll("./uploads", 0755)

	// 处理命令
	client.On("message", func(event any) {
		msg := event.(*onebot.MessageEvent)

		// 检查文本命令
		if len(msg.Message) > 0 && msg.Message[0].Type == "text" {
			text, _ := msg.Message[0].Data["text"].(string)
			
			switch text {
			case "/sendimage":
				// 发送示例图片
				handleSendImage(client, msg)
			case "/sendfile":
				// 发送示例文件
				handleSendFile(client, msg)
			case "/help":
				// 发送帮助信息
				help := onebot.Message{
					onebot.Text("文件机器人命令：\n"),
					onebot.Text("/sendimage - 发送示例图片\n"),
					onebot.Text("/sendfile - 发送示例文件\n"),
					onebot.Text("直接发送图片或文件，我会保存并回复"),
				}
				sendReply(client, msg, help)
			}
		}

		// 处理接收到的文件
		for _, seg := range msg.Message {
			switch seg.Type {
			case "image", "file", "voice", "video":
				handleReceivedFile(client, msg, seg)
			}
		}
	})

	select {}
}

// handleSendImage 处理发送图片命令
func handleSendImage(client *onebot.Client, msg *onebot.MessageEvent) {
	// 这里使用一个示例图片 URL
	imageURL := "https://via.placeholder.com/300x200.png?text=Hello+OneBot"
	
	// 上传图片
	uploadResp, err := client.UploadFile("image", "example.png", imageURL)
	if err != nil {
		reply := onebot.Message{
			onebot.Text("上传图片失败: " + err.Error()),
		}
		sendReply(client, msg, reply)
		return
	}

	// 发送图片消息
	reply := onebot.Message{
		onebot.Text("这是一张示例图片：\n"),
		onebot.Image(uploadResp.FileID),
	}
	sendReply(client, msg, reply)
}

// handleSendFile 处理发送文件命令
func handleSendFile(client *onebot.Client, msg *onebot.MessageEvent) {
	// 创建一个示例文本文件
	fileName := fmt.Sprintf("example_%d.txt", time.Now().Unix())
	filePath := filepath.Join("./uploads", fileName)
	
	content := fmt.Sprintf("这是一个示例文件\n创建时间: %s\n发送给: %s", 
		time.Now().Format("2006-01-02 15:04:05"),
		msg.UserID)
	
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		reply := onebot.Message{
			onebot.Text("创建文件失败: " + err.Error()),
		}
		sendReply(client, msg, reply)
		return
	}

	// 通过临时 HTTP 服务器提供文件
	port := "18081"
	http.HandleFunc("/"+fileName, func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	})
	go http.ListenAndServe(":"+port, nil)
	time.Sleep(100 * time.Millisecond)

	// 上传文件
	fileURL := fmt.Sprintf("http://localhost:%s/%s", port, fileName)
	uploadResp, err := client.UploadFile("file", fileName, fileURL)
	if err != nil {
		reply := onebot.Message{
			onebot.Text("上传文件失败: " + err.Error()),
		}
		sendReply(client, msg, reply)
		return
	}

	// 发送文件消息
	reply := onebot.Message{
		onebot.Text("这是一个示例文件：\n"),
		onebot.File(uploadResp.FileID),
	}
	sendReply(client, msg, reply)
}

// handleReceivedFile 处理接收到的文件
func handleReceivedFile(client *onebot.Client, msg *onebot.MessageEvent, seg onebot.MessageSegment) {
	fileID, _ := seg.Data["file_id"].(string)
	fileType := seg.Type
	
	// 获取文件信息
	fileInfo, err := client.GetFile(fileID, fileType)
	if err != nil {
		log.Printf("获取文件信息失败: %v", err)
		return
	}

	// 下载文件
	if fileInfo.URL != "" {
		fileName := fileInfo.Name
		if fileName == "" {
			fileName = fmt.Sprintf("%s_%s", fileID, fileType)
		}
		
		savePath := filepath.Join("./downloads", fileName)
		if err := downloadFile(fileInfo.URL, savePath); err != nil {
			log.Printf("下载文件失败: %v", err)
			reply := onebot.Message{
				onebot.Text(fmt.Sprintf("下载%s失败: %v", fileType, err)),
			}
			sendReply(client, msg, reply)
			return
		}

		// 回复确认消息
		reply := onebot.Message{
			onebot.Text(fmt.Sprintf("已保存%s: %s\n", fileType, fileName)),
			onebot.Text(fmt.Sprintf("文件大小: %s", getFileSize(savePath))),
		}
		
		// 如果是图片，回传缩略图
		if fileType == "image" {
			reply = append(reply, onebot.Text("\n原图："))
			reply = append(reply, onebot.Image(fileID))
		}
		
		sendReply(client, msg, reply)
	}
}

// sendReply 根据消息类型发送回复
func sendReply(client *onebot.Client, originalMsg *onebot.MessageEvent, reply onebot.Message) {
	if originalMsg.IsPrivateMessage() {
		client.SendPrivateMessage(originalMsg.UserID, reply)
	} else if originalMsg.IsGroupMessage() {
		client.SendGroupMessage(originalMsg.GroupID, reply)
	}
}

// downloadFile 下载文件
func downloadFile(url string, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// getFileSize 获取文件大小（人类可读格式）
func getFileSize(filePath string) string {
	info, err := os.Stat(filePath)
	if err != nil {
		return "unknown"
	}
	
	size := info.Size()
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
