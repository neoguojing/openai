package main

import (
	"github.com/gin-gonic/gin"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/wechat/v2"
	"github.com/neoguojing/wechat/v2/aispeech"
	speechConfig "github.com/neoguojing/wechat/v2/aispeech/config"
	"github.com/neoguojing/wechat/v2/cache"
)

var (
	aiSpeechServer *aispeech.CustomerService
)

func aiBot(in string) string {
	text, _ := chat.Dialogue(models.Text, in, "", nil)
	return text
}

func init() {
	config := config.GetConfig()
	wc := wechat.NewWechat()
	memory := cache.NewMemory()
	cfg := &speechConfig.Config{
		AppID:          config.AISpeech.AppID,
		Token:          config.AISpeech.Token,
		EncodingAESKey: config.AISpeech.EncodingAESKey,
		Cache:          memory,
		AiBot:          aiBot,
	}
	aiSpeechServer = wc.GetAiSpeech(cfg)
}

func aispeechHandler(c *gin.Context) {
	aiSpeechServer.ServeHTTP(c.Writer, c.Request)
}
