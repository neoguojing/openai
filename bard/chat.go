package bard

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ChatBard struct {
	userPrompt  string
	bard        *Bard
	chatHistory []map[string]string
	language    string
	timeout     int
	token       string
}

func NewChatBard(token string, timeout int, language string) *ChatBard {
	bard := NewBard(token, timeout, nil, nil, "", language, false, "")
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
			c.addToChatHistory(userInput, response["content"].(string))
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

func (c *ChatBard) displayResponse(response map[string]interface{}) {
	if len(response["images"].(map[string]interface{})) > 0 {
		fmt.Printf("Chatbot: %s\n\nImage links: %v\n", response["content"], response["images"])
	} else {
		fmt.Println("Chatbot:", response["content"])
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
