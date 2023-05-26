package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/neoguojing/openai"
	"github.com/neoguojing/openwechat"
	"gopkg.in/yaml.v2"
)

var (
	self *openwechat.Self
	gpt  *openai.OpenAI
)

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal(err.Error())
	}
	if config.OpenAI.ApiKey == "" {
		log.Fatal("pls provide a api key")
	}
	gpt = openai.NewOpenAI(config.OpenAI.ApiKey)
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
			fmt.Println(err)
			return
		}
		bot.SetHotStorage(storage)
		bot.DumpHotReloadStorage()
	} else {
		if err = bot.HotLogin(storage); err != nil {
			fmt.Println(err)
			os.Remove("config.json")
			fmt.Println("pls relogin~")
			return
		}
	}

	// 获取登陆的用户
	self, err = bot.GetCurrentUser()
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
	if !msg.IsText() && !msg.IsVoice() {
		return
	}

	sender, _ := msg.Sender()
	log.Println(sender.NickName, msg.Content)

	if msg.IsSendByGroup() {
		if !msg.IsAt() {
			return
		}
		group := openwechat.Group{sender}
		log.Println("group inf:", group.NickName, msg.Content)
		if msg.ToUserName == self.UserName {
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

			replayText = "@" + gSendor.NickName + " " + replayText
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
	}
}

func chatGPTReplay(msg *openwechat.Message) (string, error) {
	var err error
	if msg.IsVoice() {
		msg.Content, err = chatGPTVoice(msg)
		if err != nil {
			log.Println("chatGPTVoice: ", err.Error())
			return "", err
		}
		log.Println("chatGPTVoice content:", msg.Content)
	}

	if msg.Content == "" {
		return "", fmt.Errorf("chatGPTReplay:empty msg content")
	}

	gptResp, err := gpt.Chat().Complete(msg.Content)
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

func chatGPTVoice(msg *openwechat.Message) (string, error) {

	resp, err := msg.GetVoice()
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile("", msg.MsgId+"*.mp3")
	if err != nil {
		return "", err
	}
	defer file.Close()
	// Copy the response body to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	log.Println(file.Name())
	audioResp, err := gpt.Audio().Transcriptions(file.Name())
	if err != nil {
		return "", err
	}

	log.Println("TranscriptionsDirect:", audioResp.Text)

	return audioResp.Text, nil
}

func getConfig() (*Config, error) {
	config := &Config{}
	file, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

type OpenAIConfig struct {
	ApiKey string `yaml:"api_key"`
}

type Config struct {
	OpenAI OpenAIConfig `yaml:"openai"`
}
