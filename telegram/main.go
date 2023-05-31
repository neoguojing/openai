package main

import (
	"fmt"
	"strings"

	"github.com/neoguojing/log"
	"github.com/neoguojing/openai"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/openai/role"
	tgbotapi "github.com/neoguojing/telegram-bot-api/v5"
)

var (
	logger = log.NewLogger()
	chat   *openai.Chat
)

// Define a struct to hold the bot and its configuration
type Bot struct {
	bot *tgbotapi.BotAPI
}

// Define a struct to hold the bot's configuration

// Define a function to create a new bot
func NewBot(config config.Config) (*Bot, error) {
	// Create a new bot instance
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		return nil, err
	}

	// Set the bot's debug mode
	bot.Debug = false

	// Return the bot instance
	return &Bot{
		bot: bot,
	}, nil
}

// Define a function to handle incoming messages
func (b *Bot) HandleMessage(update tgbotapi.Update) {
	// Check if the update is a message
	if update.Message == nil && update.ChannelPost == nil {
		return
	}

	// user := update.SentFrom()

	var msg *tgbotapi.MessageConfig
	chatType := update.FromChat()
	if chatType.IsChannel() || chatType.IsGroup() || chatType.IsSuperGroup() {
		logger.Infof("receive group msg:%v", update.ChannelPost)
		msg = b.publicMessge(update)
		if msg == nil {
			return
		}
	} else {
		logger.Infof("receive msg:%v", update.Message)
		msg = b.privateMessage(update)
		if msg == nil {
			return
		}
	}

	// Send the message
	_, err := b.bot.Send(msg)
	if err != nil {
		logger.Errorf("Error sending message: %s", err)
	}
}

func (b *Bot) publicMessge(update tgbotapi.Update) *tgbotapi.MessageConfig {
	logger.Infof("group msg:%v", update.ChannelPost.Text)

	userName, replayText := b.makeReplyText(update.ChannelPost)
	if replayText == "" {
		return nil
	}
	msg := tgbotapi.NewMessage(update.ChannelPost.Chat.ID, userName+" "+replayText)
	msg.ReplyToMessageID = update.ChannelPost.MessageID
	logger.Infof("group msg:%v,%v", userName, replayText)

	return &msg
}

func (b *Bot) privateMessage(update tgbotapi.Update) *tgbotapi.MessageConfig {
	_, replayText := b.makeReplyText(update.Message)
	if replayText == "" {
		return nil
	}
	// Create a new message to send back to the user
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, replayText)
	return &msg
}

func (b *Bot) getSendUserName(text string) (userName string, content string) {
	content = text
	if strings.HasPrefix(text, "@") {
		tmp := strings.SplitN(text, " ", 2)
		if len(tmp) <= 1 {
			return
		}
		userName = tmp[0]
		content = tmp[1]
		return
	}

	return
}

func (b *Bot) makeReplyText(message *tgbotapi.Message) (userName, replayText string) {

	// 判断消息类型分别处理
	var err error
	var request string
	if message.Voice != nil {

		replayText, err = chat.Dialogue(models.Voice, "", message.Voice.FileUniqueID, nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Voice: %v", err.Error()))
			return
		}
		logger.Info(fmt.Sprintf("Voice replayText: %v", replayText))
	} else {
		userName, request = b.getSendUserName(message.Text)
		replayText, err = chat.Dialogue(models.Text, request, "", nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Text: %v", err.Error()))
			return
		}
		logger.Info(fmt.Sprintf("Text: %v", replayText))
	}

	return
}

// Define a function to start the bot
func (b *Bot) Start() error {
	// Set up a new update channel
	updates := tgbotapi.NewUpdate(0)
	updates.Timeout = 60

	// Start the update loop
	updatesChan := b.bot.GetUpdatesChan(updates)

	// Handle incoming updates
	for update := range updatesChan {
		b.HandleMessage(update)
	}

	// Return nil if the loop exits cleanly
	return nil
}

// Define the main function
func main() {
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatalf("Error creating bot: %s", err)
	}

	if config.Telegram.Token == "" {
		logger.Fatal("need bot token")
	}

	role.LoadRoles2DB()

	gpt := openai.NewOpenAI(config.OpenAI.ApiKey)
	chat = gpt.Chat(openai.WithPlatform(models.Telegram))
	if config.OpenAI.Role != "" {
		chat.Prepare(config.OpenAI.Role)
	}

	// Create a new bot instance
	bot, err := NewBot(*config)
	if err != nil {
		logger.Fatalf("Error creating bot: %s", err)
	}

	// Start the bot
	if err := bot.Start(); err != nil {
		logger.Fatalf("Error starting bot: %s", err)
	}
}
