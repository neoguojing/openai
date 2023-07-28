package main

import (
	"context"
	"strconv"
	"sync"
	"time"

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
	officialAccountServer.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
		reply := message.Reply{}
		msgId := strconv.FormatInt(msg.MsgID, 10)
		log.Infof("-------------receive msg:%v,%s", msgId, msg.Content)
		if officialAccount.GetContext().Cache.IsExist(msg.Content) {
			msgs := officialAccount.GetContext().Cache.Get(msg.Content)
			if msgs != nil {
				reply = msgs.(message.Reply)
				log.Infof("-------------msg cached:%v，%v", msgId, msgs)
			}
		} else {

			var aiText string
			var err error
			if msg.MsgType == message.MsgTypeText {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				defer cancel()

				done := make(chan bool)
				go func() {
					aiText, err = chat.Dialogue(models.Text, msg.Content, "", nil)
					// 计算消息内容的长度
					messageLength := len(aiText)

					if messageLength > 2048 {
						start := 0
						length := 2048
						segment := aiText[start : start+length]
						text := message.NewText(segment)
						reply = message.Reply{MsgType: message.MsgTypeText, MsgData: text}
					} else {
						text := message.NewText(aiText)
						reply = message.Reply{MsgType: message.MsgTypeText, MsgData: text}
					}

					officialAccount.GetContext().Cache.Set(msg.Content, reply, time.Minute*10)
					log.Infof("-------------msg reply prepare finished:%v,%v", msgId, reply)
					done <- true

				}()

				select {
				case <-done:
				case <-ctx.Done():
					if ctx.Err() == context.DeadlineExceeded {
						// 上下文对象已超时，返回固定内容
						text := message.NewText("内容生成中，请重试~")
						return &message.Reply{MsgType: message.MsgTypeText, MsgData: text}
					}
				}

				if err != nil {
					log.Error(err.Error())
					return &message.Reply{MsgType: message.MsgTypeText, MsgData: "ops"}
				}

				log.Infof("-------------chat.Dialogue:%v", aiText)
			} else if msg.MsgType == message.MsgTypeVoice {

			} else {

			}

		}

		return &reply
	})

	// 处理消息接收以及回复
	err := officialAccountServer.Serve()
	if err != nil {
		log.Error(err.Error())
		return
	}

	officialAccountServer.Send()
}
