package main

import (
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
)

var (
	aiSpeechServer  *aispeech.CustomerService
	wc              *wechat.Wechat
	officialAccount *officialaccount.OfficialAccount
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
	// 传入request和responseWriter
	server := officialAccount.GetServer(c.Request, c.Writer)
	// 设置接收消息的处理方法
	server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		var aiText string
		var err error
		if msg.MsgType == message.MsgTypeText {
			aiText, err = chat.Dialogue(models.Text, msg.Content, "", nil)
			if err != nil {
				log.Error(err.Error())
				return &message.Reply{}
			}
		} else if msg.MsgType == message.MsgTypeVoice {

		} else {

		}
		text := message.NewText(aiText)
		return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
	})

	// 处理消息接收以及回复
	err := server.Serve()
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 发送回复的消息
	server.Send()
}