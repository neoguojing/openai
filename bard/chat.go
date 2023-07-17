package bard

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/danielpark/bardapi"
)

type ChatBard struct {
	userPrompt  string
	bard        *bardapi.Bard
	chatHistory []map[string]string
	language    string
	timeout     int
	token       string
}

func NewChatBard(token string, timeout int, language string) *ChatBard {
	bard := bardapi.NewBard(token, language, timeout)
	return &ChatBard{
		userPrompt:  ">>> ",
		bard:        bard,
		chatHistory: []map[string]string{},
		language:    language,
		timeout:     timeout,
		token:       token,
	}
}

func (c *ChatBard) Start() {
	fmt.Println("Welcome to Chatbot")
	fmt.Println("If you enter quit, q, or stop, the chat will end.")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print(c.userPrompt)
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)
		userInput = strings.ToLower(userInput)

		if userInput == "quit" || userInput == "q" || userInput == "stop" {
			break
		}

		if !c.isValidInput(userInput) {
			fmt.Println("Invalid input! Please try again.")
			continue
		}

		response, err := c.bard.GetAnswer(userInput)
		if err != nil {
			fmt.Println("Error occurred:", err.Error())
		} else {
			c.displayResponse(response)
			c.addToChatHistory(userInput, response.Content)
		}
	}

	fmt.Println("Chat Ended.")
}

func (c *ChatBard) isValidInput(userInput string) bool {
	if userInput == "" {
		return false
	}
	if len(userInput) > 1000 {
		return false
	}
	return true
}

func (c *ChatBard) displayResponse(response *bardapi.Response) {
	if len(response.Images) > 0 {
		fmt.Printf("Chatbot: %s\n\nImage links: %v\n", response.Content, response.Images)
	} else {
		fmt.Println("Chatbot:", response.Content)
	}
}

func (c *ChatBard) addToChatHistory(userInput string, chatbotResponse string) {
	entry := map[string]string{
		"User":    userInput,
		"Chatbot": chatbotResponse,
	}
	c.chatHistory = append(c.chatHistory, entry)
}

func (c *ChatBard) DisplayChatHistory() {
	fmt.Println("Chat History")

	for _, entry := range c.chatHistory {
		fmt.Println("User:", entry["User"])
		fmt.Println("Chatbot:", entry["Chatbot"])
	}
}

// func main() {
// 	token := os.Getenv("BARD_API_KEY")
// 	timeout := 30         // You can set this to your desired default timeout
// 	language := "english" // You can set this to your desired default language

// 	chat := NewChatBard(token, timeout, language)
// 	chat.Start()
// }
