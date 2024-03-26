package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"errors"

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

	tGroups openwechat.Groups

	logger = log.NewLogger()

	NoMatchReplay = errors.New("no match wechat replay")
)

func main() {
	var err error
	role.LoadRoles2DB()

	config := config.GetConfig()
	if config.OpenAI.ApiKey == "" {
		logger.Error("pls provide a api key")
		// return
	} else {
		gpt := openai.NewOpenAI(config.OpenAI.ApiKey)
		chat = gpt.Chat(openai.WithPlatform(models.Wechat), openai.WithComplete(openai.Baidu))
		if config.OpenAI.Role != "" {
			chat.Prepare(config.OpenAI.Role)
		}
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
	// friends, err := self.Friends()
	// logger.Info(fmt.Sprintf("friends: %v, err: %v", friends, err))

	// 获取所有的群组
	groups, err := self.Groups()
	logger.Info(fmt.Sprintf("groups: %v, err: %v", groups, err))
	tgNames := config.Wechat.Groups
	for _, group := range groups {
		for _, name := range tgNames {
			if strings.HasPrefix(group.NickName, name) {
				logger.Infof("group:%v", *group.User)
				tGroups = append(tGroups, group)
			}
		}
	}
	logger.Info(fmt.Sprintf("tgroups: %v, err: %v", tGroups, err))

	bot.MessageHandler = MessageHandler
	defer models.GetRecorder().Exit()

	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

func MessageHandler(msg *openwechat.Message) {
	if !msg.IsText() && !msg.IsVoice() {
		// err := mutiMediaRecord(msg)
		// if err != nil {
		// 	logger.Error(err.Error())
		// 	return
		// }
		// return
	}

	sender, _ := msg.Sender()
	logger.Info(fmt.Sprintf("sender: %v, content: %v", sender.NickName, msg.Url))
	msgByte, _ := json.Marshal(msg)
	logger.Infof("message struct:%s", string(msgByte))
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
			// replayText, err := chatGPTReplay(msg)
			replayText, err := matchReplay(msg)
			if err != nil && err != NoMatchReplay {
				logger.Error(fmt.Sprintf("ReplyText: %v", err.Error()))
				msg.ReplyText("ops...")
				return
			}

			if replayText == "" {
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
		} else {

		}

	} else if msg.IsSendByFriend() {
		// replayText, err := chatGPTReplay(msg)
		replayText, err := matchReplay(msg)
		if err != nil && err != NoMatchReplay {
			logger.Error(fmt.Sprintf("ReplyText: %v", err.Error()))
			msg.ReplyText("ops...")
			return
		}

		if replayText == "" {
			return
		}

		_, err = msg.ReplyText(replayText)
		if err != nil {
			logger.Error(fmt.Sprintf("ReplyText: %v", err.Error()))
		}
	} else if msg.IsSendBySelf() {
		if msg.IsText() {
			err := self.SendTextToGroups(msg.Content, time.Second*3, tGroups...)
			if err != nil {
				logger.Error(fmt.Sprintf("SendTextToGroups: %v", err.Error()))
			}
		}

		logger.Errorf("other %v,%v", msg.MsgType, msg.Url)

		if msg.IsPicture() {
			resp, err := msg.GetPicture()
			if err != nil {
				return
			}
			defer resp.Body.Close()
			logger.Errorf("picture %v", msg)
			err = self.SendImageToGroups(resp.Body, time.Second*3, tGroups...)
			if err != nil {
				logger.Error(fmt.Sprintf("SendTextToGroups: %v", err.Error()))
			}
		}

		if msg.IsVideo() {
			// logger.Errorf("card %v", msg)
			resp, err := msg.GetVideo()
			if err != nil {
				return
			}
			defer resp.Body.Close()
			err = self.SendVideoToGroups(resp.Body, time.Second*3, tGroups...)
			if err != nil {
				logger.Error(fmt.Sprintf("SendTextToGroups: %v", err.Error()))
			}
		}

		if msg.IsMedia() {

			// sendMsg := openwechat.NewMediaSendMessage(msg.MsgType, msg.FromUserName, msg.ToUserName, msg.MediaId)
			// sentMsg := openwechat.SentMessage{
			// 	self: msg.Owner().Self(),
			// }
			// sentMsg.SendMessage = sendMsg
			// sentMsg.MsgId = msg.MsgId

			// err := self.ForwardMessageToGroups(sentMsg, time.Second*3, tGroups...)
			err := self.SendTextToGroups(msg.Url, time.Second*3, tGroups...)
			if err != nil {
				logger.Error(fmt.Sprintf("SendTextToGroups: %v", err.Error()))
			}
		}

	} else {
		logger.Warning("unhandled msg!!!!")
	}

}

func matchReplay(msg *openwechat.Message) (string, error) {
	config := config.GetConfig()
	wechatConfig := config.Wechat.Qas

	if msg.IsText() {
		for _, qa := range wechatConfig {
			if strings.Contains(msg.Content, qa.Q) {
				return qa.A, nil
			}
		}
	}

	return "", NoMatchReplay
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

func mutiMediaRecord(msg *openwechat.Message) error {
	var err error
	switch msg.MsgType {
	case openwechat.MsgTypeVoice:
		resp, err := msg.GetVoice()
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		fileName := msg.MsgId + ".mp3"
		logger.Info(fileName)
		err = chat.Recorder(models.Voice, "", fileName, resp.Body)
	case openwechat.MsgTypeImage:
		resp, err := msg.GetPicture()
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		fileName := msg.MsgId + ".jpeg"
		logger.Info(fileName)
		err = chat.Recorder(models.Picture, "", fileName, resp.Body)
	case openwechat.MsgTypeText:
		err = chat.Recorder(models.Text, msg.Content, "", nil)
	case openwechat.MsgTypeVideo:
		resp, err := msg.GetVideo()
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		fileName := msg.MsgId + ".mp4"
		logger.Info(fileName)
		err = chat.Recorder(models.Video, "", fileName, resp.Body)
	default:
		return errors.New("no support msg type")
	}
	return err
}
