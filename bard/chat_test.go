package bard

import (
	"os"
	"testing"
)

func TestChat(t *testing.T) {
	token := os.Getenv("BARD_API_KEY")
	timeout := 30         // You can set this to your desired default timeout
	language := "english" // You can set this to your desired default language

	chat := NewChatBard(token, timeout, language)
	chat.Start()
}
