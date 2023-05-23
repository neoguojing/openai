package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/neoguojing/openai"
	"github.com/neoguojing/openwechat"
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

	sender, _ := msg.Sender()
	log.Println(sender.NickName, msg.Content)

	if msg.IsSendByGroup() {
		if !msg.IsAt() {
			return
		}
		// group := openwechat.Group(sender)
		// log.Println(group.NickName, msg.Content)
		recv, _ := msg.Receiver()
		if recv.IsSelf() {
			dumpText := "@" + sender.Self().NickName
			msg.Content = strings.ReplaceAll(msg.Content, dumpText, "")
			if msg.Content == "" {
				return
			}
			replayText, err := chatGPTReplay(msg)
			if err != nil {
				log.Println("ReplyText: ", err.Error())
				msg.ReplyText("ops...")
				return
			}
			gSendor, err := msg.SenderInGroup()
			if err != nil {
				log.Println("SendorInGroup: ", err.Error())
				msg.ReplyText("ops...")
				return
			}

			replayText = "@" + gSendor.NickName + replayText
			_, err = msg.ReplyText(replayText)
			if err != nil {
				log.Println("ReplyText: ", err.Error())
			}
		}

	} else if msg.IsSendByFriend() {
		replayText, err := chatGPTReplay(msg)
		if err != nil {
			log.Println("ReplyText: ", err.Error())
			msg.ReplyText("ops...")
			return
		}
		_, err = msg.ReplyText(replayText)
		if err != nil {
			log.Println("ReplyText: ", err.Error())
		}
	} else if msg.IsSendBySelf() {

	} else {

	}
}

func chatGPTReplay(msg *openwechat.Message) (string, error) {
	gptResp, err := openai.NewOpenAI("").
		Chat().Complete(msg.Content)
	if err != nil {
		log.Println("Complete: ", err.Error())
		return "", err
	}
	if len(gptResp.Choices) == 0 {
		log.Println("Empty from gpt")
		return "", err
	}
	replayText := gptResp.Choices[0].Message.Content
	replayText = strings.TrimSpace(replayText)
	replayText = strings.Trim(replayText, "\n")
	log.Println(replayText)
	return replayText, nil
}
