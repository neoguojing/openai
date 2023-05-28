package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/neoguojing/log"
	"github.com/neoguojing/openai"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/openai/role"
	"github.com/neoguojing/openwechat"
)

var (
	self *openwechat.Self
	chat *openai.Chat

	logger = log.NewLogger()
)

func main() {
	var err error
	role.LoadRoles2DB()

	config := config.GetConfig()
	if config.OpenAI.ApiKey == "" {
		logger.Error("pls provide a api key")
		return
	}
	gpt := openai.NewOpenAI(config.OpenAI.ApiKey)
	chat = gpt.Chat()
	if config.OpenAI.Role != "" {
		chat.Prepare("config.OpenAI.Role")
	}

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
	// Check if the config file exists
	_, err = os.Stat("config.json")
	var storage = openwechat.NewFileHotReloadStorage("config.json")
	if os.IsNotExist(err) {
		if err = bot.Login(); err != nil {
			logger.Error(err.Error())
			return
		}
		bot.SetHotStorage(storage)
		bot.DumpHotReloadStorage()
	} else {
		if err = bot.HotLogin(storage); err != nil {
			logger.Error(err.Error())
			os.Remove("config.json")
			fmt.Println("pls relogin~")
			return
		}
	}

	// 获取登陆的用户
	self, err = bot.GetCurrentUser()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// 获取所有的好友
	friends, err := self.Friends()
	logger.Info(fmt.Sprintf("friends: %v, err: %v", friends, err))

	// 获取所有的群组
	groups, err := self.Groups()
	logger.Info(fmt.Sprintf("groups: %v, err: %v", groups, err))

	bot.MessageHandler = MessageHandler

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

func MessageHandler(msg *openwechat.Message) {
	if !msg.IsText() && !msg.IsVoice() {
		return
	}

	sender, _ := msg.Sender()
	logger.Info(fmt.Sprintf("sender: %v, content: %v", sender.NickName, msg.Content))

	if msg.IsSendByGroup() {
		if !msg.IsAt() {
			return
		}
		group := openwechat.Group{sender}
		logger.Info(fmt.Sprintf("group inf: %v, content: %v", group.NickName, msg.Content))
		if msg.ToUserName == self.UserName {
			dumpText := "@" + sender.Self().NickName
			msg.Content = strings.ReplaceAll(msg.Content, dumpText, "")
			if msg.Content == "" {
				return
			}
			replayText, err := chatGPTReplay(msg)
			if err != nil {
				logger.Error(fmt.Sprintf("ReplyText: %v", err.Error()))
				msg.ReplyText("ops...")
				return
			}
			gSendor, err := msg.SenderInGroup()
			if err != nil {
				logger.Error(fmt.Sprintf("SendorInGroup: %v", err.Error()))
				msg.ReplyText("ops...")
				return
			}

			replayText = "@" + gSendor.NickName + " " + replayText
			_, err = msg.ReplyText(replayText)
			if err != nil {
				logger.Error(fmt.Sprintf("ReplyText: %v", err.Error()))
			}
		}

	} else if msg.IsSendByFriend() {
		replayText, err := chatGPTReplay(msg)
		if err != nil {
			logger.Error(fmt.Sprintf("ReplyText: %v", err.Error()))
			msg.ReplyText("ops...")
			return
		}
		_, err = msg.ReplyText(replayText)
		if err != nil {
			logger.Error(fmt.Sprintf("ReplyText: %v", err.Error()))
		}
	}
}

func chatGPTReplay(msg *openwechat.Message) (string, error) {
	var replayText string
	var err error
	if msg.IsVoice() {
		resp, err := msg.GetVoice()
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		fileName := msg.MsgId + ".mp3"
		logger.Info(fileName)

		replayText, err = chat.Dialogue(models.Voice, "", fileName, resp.Body)
		if err != nil {
			logger.Error(fmt.Sprintf("chatGPTVoice: %v", err.Error()))
			return "", err
		}
		logger.Info(fmt.Sprintf("chatGPTVoice replayText: %v", replayText))
	} else {
		replayText, err = chat.Dialogue(models.Text, msg.Content, "", nil)
		if err != nil {
			logger.Error(fmt.Sprintf("chatGPTReplay: %v", err.Error()))
			return "", err
		}
	}

	replayText = strings.TrimSpace(replayText)
	replayText = strings.Trim(replayText, "\n")
	logger.Info(fmt.Sprintf("replayText: %v", replayText))
	return replayText, nil
}
