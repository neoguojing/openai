package main

import (
	"math"
	"sync"

	"github.com/neoguojing/log"

	"github.com/gin-gonic/gin"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/wechat/v2"
	"github.com/neoguojing/wechat/v2/aispeech"
	speechConfig "github.com/neoguojing/wechat/v2/aispeech/config"
	"github.com/neoguojing/wechat/v2/cache"
	"github.com/neoguojing/wechat/v2/officialaccount"
	offConfig "github.com/neoguojing/wechat/v2/officialaccount/config"
	"github.com/neoguojing/wechat/v2/officialaccount/message"
	"github.com/neoguojing/wechat/v2/officialaccount/server"
)

var (
	aiSpeechServer        *aispeech.CustomerService
	wc                    *wechat.Wechat
	officialAccount       *officialaccount.OfficialAccount
	once                  sync.Once
	officialAccountServer *server.Server
)

func aiBot(in string) string {
	text, _ := chat.Dialogue(models.Text, in, "", nil)
	return text
}

func init() {
	config := config.GetConfig()
	wc = wechat.NewWechat()
	memory := cache.NewMemory()

	cfg := &speechConfig.Config{
		AppID:          config.AISpeech.AppID,
		Token:          config.AISpeech.Token,
		EncodingAESKey: config.AISpeech.EncodingAESKey,
		Cache:          memory,
		AiBot:          aiBot,
	}
	aiSpeechServer = wc.GetAiSpeech(cfg)

	officeCfg := &offConfig.Config{
		AppID:          config.OfficeAccount.AppID,
		AppSecret:      config.OfficeAccount.AppSecret,
		Token:          config.OfficeAccount.Token,
		EncodingAESKey: config.OfficeAccount.EncodingAESKey,
		Cache:          memory,
	}
	officialAccount = wc.GetOfficialAccount(officeCfg)

}

func aispeechHandler(c *gin.Context) {
	aiSpeechServer.ServeHTTP(c.Writer, c.Request)
}

func officeAccountHandler(c *gin.Context) {
	log.Info(c.Request.Host)
	// 传入request和responseWriter
	officialAccountServer = officialAccount.GetServer(c.Request, c.Writer)
	// 设置接收消息的处理方法
	officialAccountServer.SetMessageHandler(func(msg *message.MixMessage) []message.Reply {
		var aiText string
		var err error
		if msg.MsgType == message.MsgTypeText {
			aiText, err = chat.Dialogue(models.Text, msg.Content, "", nil)
			if err != nil {
				log.Error(err.Error())
				return []message.Reply{{MsgType: message.MsgTypeText, MsgData: "ops"}}
			}
		} else if msg.MsgType == message.MsgTypeVoice {

		} else {

		}

		// 计算消息内容的长度
		messageLength := len(aiText)

		// 计算消息需要分成多少段
		segmentCount := int(math.Ceil(float64(messageLength) / 2048.0))
		replys := []message.Reply{}
		// 分段发送消息
		for i := 0; i < segmentCount; i++ {
			// 计算当前段的起始位置和长度
			start := i * 2048
			length := 2048
			if start+length > messageLength {
				length = messageLength - start
			}

			// 截取当前段的消息内容
			segment := aiText[start : start+length]
			text := message.NewText(segment)
			reply := message.Reply{MsgType: message.MsgTypeText, MsgData: text}
			replys = append(replys, reply)

		}

		log.Infof("segmentCount:%v", segmentCount)
		return replys
	})

	// 处理消息接收以及回复
	err := officialAccountServer.Serve()
	if err != nil {
		log.Error(err.Error())
		return
	}
}
