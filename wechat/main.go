package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/eatmoreapple/openwechat"
	"github.com/neoguojing/openai"
)

func main() {
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式

	// 注册消息处理函数
	bot.MessageHandler = func(msg *openwechat.Message) {
		if msg.IsText() && msg.Content == "ping" {
			msg.ReplyText("pong")
		}
	}
	// 注册登陆二维码回调
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl

	// 登陆
	if err := bot.Login(); err != nil {
		fmt.Println(err)
		return
	}

	// 获取登陆的用户
	self, err := bot.GetCurrentUser()
	if err != nil {
		fmt.Println(err)
		return
	}

	// 获取所有的好友
	friends, err := self.Friends()
	fmt.Println(friends, err)

	// 获取所有的群组
	groups, err := self.Groups()
	fmt.Println(groups, err)

	bot.MessageHandler = MessageHandler

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

func MessageHandler(msg *openwechat.Message) {
	if !msg.IsText() {
		return
	}
	log.Println(msg.Content)
	gptResp, err := openai.NewOpenAI("").Chat().Complete(msg.Content)
	if err != nil {
		log.Println("Complete: ", err.Error())
		return
	}
	if len(gptResp.Choices) == 0 {
		log.Println("Empty from gpt")
		return
	}
	replayText := gptResp.Choices[0].Message.Content
	replayText = strings.TrimSpace(replayText)
	replayText = strings.Trim(replayText, "\n")
	log.Println(replayText)
	if msg.IsSendByGroup() {
		if !msg.IsAt() {
			return
		}

	} else if msg.IsSendByFriend() {

	} else if msg.IsSendBySelf() {
		_, err = msg.ReplyText(replayText)
		if err != nil {
			log.Println("ReplyText: ", err.Error())
		}
	} else {

	}
}
