package openai

import "github.com/neoguojing/openai/config"

type IChat interface {
	Complete(string) (*ChatResponse, error)
}

type ChatFactory struct {
	config *config.Config
}

func NewChatFactory() *ChatFactory {
	return &ChatFactory{
		config: config.GetConfig(),
	}
}

var GlobalChatFactory *ChatFactory

func init() {
	GlobalChatFactory = NewChatFactory()
}

type ChatType string

const (
	Baidu  ChatType = "baidu"
	Claude ChatType = "claude"
)

func (f *ChatFactory) GetChat(chatType ChatType) IChat {
	switch chatType {
	case Baidu:
		b := f.config.Baidu
		if b.Key != "" && b.Secret != "" {
			client := NewBaiduClient(b.Key, b.Secret)
			return client
		}
	case Claude:
		b := f.config.Claude
		if b.ApiKey != "" {
			client := NewClaudeClient(b.ApiKey)
			return client
		}
	}
	return nil
}
