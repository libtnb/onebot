package example

import (
	"log"
	"log/slog"
	"time"

	"github.com/libtnb/onebot"
)

func Example_Basic() {
	// 创建客户端
	client, err := onebot.New("ws://localhost:5700",
		onebot.WithAccessToken("your-access-token"),
		onebot.WithLogger(slog.Default()),
		onebot.WithReconnect(true),
		onebot.WithReconnectWait(5*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 连接到 OneBot 实现
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 注册事件处理器
	client.On("message.private", func(event any) {
		msg := event.(*onebot.MessageEvent)
		log.Printf("收到私聊消息: %s", msg.AltMessage)

		// 回复消息
		reply := onebot.Message{
			onebot.Text("你发送了: "),
			onebot.Text(msg.AltMessage),
		}

		if _, err := client.SendPrivateMessage(msg.UserID, reply); err != nil {
			log.Printf("发送消息失败: %v", err)
		}
	})

	// 等待事件
	select {}
}

func Example_EchoBot() {
	// 创建回声机器人客户端
	client, _ := onebot.New("ws://localhost:5700")
	defer client.Close()

	// 连接
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 处理所有消息事件
	client.On("message", func(event any) {
		msg := event.(*onebot.MessageEvent)

		// 构造回复
		var reply onebot.Message
		reply = append(reply, onebot.Reply(msg.MessageID))
		reply = append(reply, onebot.Text("Echo: "))
		reply = append(reply, msg.Message...)

		// 根据消息类型回复
		switch msg.DetailType {
		case "private":
			client.SendPrivateMessage(msg.UserID, reply)
		case "group":
			client.SendGroupMessage(msg.GroupID, reply)
		}
	})

	// 运行
	select {}
}

func Example_CommandBot() {
	client, _ := onebot.New("ws://localhost:5700",
		onebot.WithLogger(slog.Default()),
	)
	defer client.Close()

	// 连接
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 处理消息
	client.On("message", func(event any) {
		msg := event.(*onebot.MessageEvent)

		// 检查是否为文本消息
		if len(msg.Message) == 0 {
			return
		}

		// 获取第一个消息段
		first := msg.Message[0]
		if first.Type != "text" {
			return
		}

		text, _ := first.Data["text"].(string)

		// 处理命令
		switch text {
		case "/help":
			help := onebot.Message{
				onebot.Text("可用命令:\n"),
				onebot.Text("/help - 显示帮助\n"),
				onebot.Text("/version - 查看版本\n"),
				onebot.Text("/status - 查看状态"),
			}

			if msg.IsPrivateMessage() {
				client.SendPrivateMessage(msg.UserID, help)
			} else if msg.IsGroupMessage() {
				client.SendGroupMessage(msg.GroupID, help)
			}

		case "/version":
			version, err := client.GetVersion()
			if err != nil {
				log.Printf("获取版本失败: %v", err)
				return
			}

			reply := onebot.Message{
				onebot.Text("OneBot 版本信息:\n"),
				onebot.Text("实现: " + version.Impl + "\n"),
				onebot.Text("版本: " + version.Version + "\n"),
				onebot.Text("OneBot 标准: " + version.OneBotVersion),
			}

			if msg.IsPrivateMessage() {
				client.SendPrivateMessage(msg.UserID, reply)
			} else if msg.IsGroupMessage() {
				client.SendGroupMessage(msg.GroupID, reply)
			}

		case "/status":
			status, err := client.GetStatus()
			if err != nil {
				log.Printf("获取状态失败: %v", err)
				return
			}

			statusText := "正常"
			if !status.Good {
				statusText = "异常"
			}

			reply := onebot.Message{
				onebot.Text("机器人状态: " + statusText),
			}

			if msg.IsPrivateMessage() {
				client.SendPrivateMessage(msg.UserID, reply)
			} else if msg.IsGroupMessage() {
				client.SendGroupMessage(msg.GroupID, reply)
			}
		}
	})

	// 运行
	select {}
}

func Example_MultiBot() {
	// 多账号场景
	client, _ := onebot.New("ws://multi-bot.example.com:5700",
		onebot.WithSelf("wechat", "bot123456"), // 指定要操作的机器人账号
		onebot.WithAccessToken("token"),
	)
	defer client.Close()

	// 连接
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 监听事件
	client.On("*", func(event any) {
		// 处理所有事件
		switch e := event.(type) {
		case *onebot.MessageEvent:
			log.Printf("[%s] 收到消息: %s", e.Self.UserID, e.AltMessage)
		case *onebot.NoticeEvent:
			log.Printf("[%s] 收到通知: %s", e.Self.UserID, e.DetailType)
		}
	})

	// 运行
	select {}
}

func Example_WeChatBridge() {
	// 创建微信桥接客户端
	client, _ := onebot.New("ws://wechat-onebot.example.com:5700",
		onebot.WithSelf("wechat", "wxid_xxxxx"),
		onebot.WithLogger(slog.Default()),
	)
	defer client.Close()

	// 连接
	if err := client.Connect(); err != nil {
		log.Fatal(err)
	}

	// 处理微信私聊消息
	client.On("message.private", func(event any) {
		msg := event.(*onebot.MessageEvent)

		// 记录消息（带微信扩展字段）
		log.Printf("微信私聊 [%s]: %s", msg.UserID, msg.AltMessage)

		// 自动回复
		if msg.AltMessage == "你好" {
			reply := onebot.Message{
				onebot.Text("你好，有什么可以帮助你的吗？"),
			}
			client.SendPrivateMessage(msg.UserID, reply)
		}
	})

	// 处理微信群消息
	client.On("message.group", func(event any) {
		msg := event.(*onebot.MessageEvent)

		// 检查是否被 @
		for _, seg := range msg.Message {
			if seg.Type == "mention" {
				if selfID, ok := seg.Data["user_id"].(string); ok && selfID == "wxid_xxxxx" {
					// 被 @ 了，回复消息
					reply := onebot.Message{
						onebot.Mention(msg.UserID),
						onebot.Text(" 收到！"),
					}
					client.SendGroupMessage(msg.GroupID, reply)
					break
				}
			}
		}
	})

	// 运行
	select {}
}
