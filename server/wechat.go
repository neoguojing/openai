package main

import (
	"fmt"
	"net/http"
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

var officeAccountHandlerStand = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from myHttpHandler!")
	log.Info(r.Host)

	// 传入request和responseWriter
	officialAccountServer = officialAccount.GetServer(r, w)
	// 设置接收消息的处理方法
	officialAccountServer.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
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
	err := officialAccountServer.Serve()
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 发送回复的消息
	err = officialAccountServer.Send()
	if err != nil {
		log.Error(err.Error())
		return
	}
})

func officeAccountHandler(c *gin.Context) {
	log.Info(c.Request.Host)

	// 传入request和responseWriter
	officialAccountServer = officialAccount.GetServer(c.Request, c.Writer)
	// 设置接收消息的处理方法
	officialAccountServer.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
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
	err := officialAccountServer.Serve()
	if err != nil {
		log.Error(err.Error())
		return
	}
	// 发送回复的消息
	err = officialAccountServer.Send()
	if err != nil {
		log.Error(err.Error())
		return
	}
}
