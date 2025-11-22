package onebot

import (
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client, err := New("ws://localhost:5700",
		WithAccessToken("test-token"),
		WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("创建客户端失败: %v", err)
	}
	
	if client.url != "ws://localhost:5700" {
		t.Errorf("URL 设置错误: got %s, want %s", client.url, "ws://localhost:5700")
	}
	
	if client.accessToken != "test-token" {
		t.Errorf("AccessToken 设置错误: got %s, want %s", client.accessToken, "test-token")
	}
	
	if client.timeout != 5*time.Second {
		t.Errorf("Timeout 设置错误: got %v, want %v", client.timeout, 5*time.Second)
	}
}

func TestMessageSegmentConstructors(t *testing.T) {
	// 测试文本消息段
	text := Text("Hello, World!")
	if text.Type != "text" {
		t.Errorf("文本消息段类型错误: got %s, want %s", text.Type, "text")
	}
	if text.Data["text"] != "Hello, World!" {
		t.Errorf("文本消息段内容错误: got %v, want %s", text.Data["text"], "Hello, World!")
	}
	
	// 测试图片消息段
	image := Image("file123")
	if image.Type != "image" {
		t.Errorf("图片消息段类型错误: got %s, want %s", image.Type, "image")
	}
	if image.Data["file_id"] != "file123" {
		t.Errorf("图片消息段 file_id 错误: got %v, want %s", image.Data["file_id"], "file123")
	}
	
	// 测试位置消息段
	location := Location(31.0, 121.0, "标题", "内容")
	if location.Type != "location" {
		t.Errorf("位置消息段类型错误: got %s, want %s", location.Type, "location")
	}
	if location.Data["latitude"] != 31.0 {
		t.Errorf("位置消息段纬度错误: got %v, want %f", location.Data["latitude"], 31.0)
	}
	
	// 测试提及消息段
	mention := Mention("user123")
	if mention.Type != "mention" {
		t.Errorf("提及消息段类型错误: got %s, want %s", mention.Type, "mention")
	}
	if mention.Data["user_id"] != "user123" {
		t.Errorf("提及消息段 user_id 错误: got %v, want %s", mention.Data["user_id"], "user123")
	}
	
	// 测试提及所有人消息段
	mentionAll := MentionAll()
	if mentionAll.Type != "mention_all" {
		t.Errorf("提及所有人消息段类型错误: got %s, want %s", mentionAll.Type, "mention_all")
	}
}

func TestMessageToAltMessage(t *testing.T) {
	msg := Message{
		Text("你好"),
		Image("file123"),
		Voice("voice456"),
		Mention("user789"),
		MentionAll(),
		Text("！"),
	}
	
	expected := "你好[图片][语音]@user789@全体成员！"
	result := msg.ToAltMessage()
	
	if result != expected {
		t.Errorf("ToAltMessage 结果错误: got %s, want %s", result, expected)
	}
}

func TestActionRequest(t *testing.T) {
	params := map[string]any{
		"user_id": "123",
		"message": "test",
	}
	
	req := NewActionRequest("send_message", params)
	req.WithEcho("echo123").WithSelf(&Self{Platform: "qq", UserID: "bot123"})
	
	if req.Action != "send_message" {
		t.Errorf("Action 错误: got %s, want %s", req.Action, "send_message")
	}
	
	if req.Echo != "echo123" {
		t.Errorf("Echo 错误: got %s, want %s", req.Echo, "echo123")
	}
	
	if req.Self.Platform != "qq" {
		t.Errorf("Self.Platform 错误: got %s, want %s", req.Self.Platform, "qq")
	}
}

func TestActionResponse(t *testing.T) {
	resp := &ActionResponse{
		Status:  "ok",
		Retcode: 0,
		Data: map[string]any{
			"message_id": "msg123",
			"time":       1234567890.5,
		},
		Message: "",
		Echo:    "echo123",
	}
	
	if !resp.IsOK() {
		t.Error("IsOK() 应该返回 true")
	}
	
	var result SendMessageResponse
	err := resp.UnmarshalData(&result)
	if err != nil {
		t.Fatalf("UnmarshalData 失败: %v", err)
	}
	
	if result.MessageID != "msg123" {
		t.Errorf("MessageID 错误: got %s, want %s", result.MessageID, "msg123")
	}
	
	if result.Time != 1234567890.5 {
		t.Errorf("Time 错误: got %f, want %f", result.Time, 1234567890.5)
	}
}

func TestEventParsing(t *testing.T) {
	// 测试解析私聊消息事件
	privateMsg := `{
		"id": "test-id",
		"self": {"platform": "qq", "user_id": "bot123"},
		"time": 1234567890.5,
		"type": "message",
		"detail_type": "private",
		"sub_type": "",
		"message_id": "msg123",
		"message": [{"type": "text", "data": {"text": "Hello"}}],
		"alt_message": "Hello",
		"user_id": "user456"
	}`
	
	event, err := ParseEvent([]byte(privateMsg))
	if err != nil {
		t.Fatalf("解析私聊消息失败: %v", err)
	}
	
	msgEvent, ok := event.(*MessageEvent)
	if !ok {
		t.Fatal("类型断言失败，应该是 MessageEvent")
	}
	
	if !msgEvent.IsPrivateMessage() {
		t.Error("IsPrivateMessage() 应该返回 true")
	}
	
	if msgEvent.IsGroupMessage() {
		t.Error("IsGroupMessage() 应该返回 false")
	}
	
	if msgEvent.UserID != "user456" {
		t.Errorf("UserID 错误: got %s, want %s", msgEvent.UserID, "user456")
	}
	
	// 测试解析群消息事件
	groupMsg := `{
		"id": "test-id2",
		"self": {"platform": "qq", "user_id": "bot123"},
		"time": 1234567890.5,
		"type": "message",
		"detail_type": "group",
		"sub_type": "",
		"message_id": "msg456",
		"message": [{"type": "text", "data": {"text": "World"}}],
		"alt_message": "World",
		"user_id": "user789",
		"group_id": "group123"
	}`
	
	event2, err := ParseEvent([]byte(groupMsg))
	if err != nil {
		t.Fatalf("解析群消息失败: %v", err)
	}
	
	msgEvent2, ok := event2.(*MessageEvent)
	if !ok {
		t.Fatal("类型断言失败，应该是 MessageEvent")
	}
	
	if msgEvent2.IsPrivateMessage() {
		t.Error("IsPrivateMessage() 应该返回 false")
	}
	
	if !msgEvent2.IsGroupMessage() {
		t.Error("IsGroupMessage() 应该返回 true")
	}
	
	if msgEvent2.GroupID != "group123" {
		t.Errorf("GroupID 错误: got %s, want %s", msgEvent2.GroupID, "group123")
	}
}

func TestEventHandlerRegistration(t *testing.T) {
	client, _ := New("ws://localhost:5700")
	
	client.On("message.private", func(event any) {
		// 处理消息
	})
	
	// 检查处理器是否注册
	if len(client.eventHandlers["message.private"]) != 1 {
		t.Error("事件处理器没有正确注册")
	}
	
	// 注册通配符处理器
	client.On("*", func(event any) {})
	
	if len(client.eventHandlers["*"]) != 1 {
		t.Error("通配符事件处理器没有正确注册")
	}
}

func TestTimestamp(t *testing.T) {
	now := time.Now()
	ts := Timestamp()
	
	// 检查时间戳是否合理（在当前时间附近）
	diff := ts - float64(now.Unix())
	if diff < -1 || diff > 1 {
		t.Errorf("Timestamp() 返回的时间戳不正确: %f (diff: %f)", ts, diff)
	}
}
