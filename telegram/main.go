package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/neoguojing/log"
	"github.com/neoguojing/openai"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/openai/role"
	tgbotapi "github.com/neoguojing/telegram-bot-api/v5"
	"github.com/yanyiwu/gojieba"
)

var (
	logger = log.NewLogger()
	chat   *openai.Chat
)

// Define a struct to hold the bot and its configuration
type Bot struct {
	bot   *tgbotapi.BotAPI
	jieba *gojieba.Jieba
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
	// dictDir := filepath.Join(filepath.Dir(os.Args[0]), "dict")
	// jiebaPath := filepath.Join(dictDir, "jieba.dict.utf8")
	// hmmPath := filepath.Join(dictDir, "hmm_model.utf8")
	// idfPath := filepath.Join(dictDir, "idf.utf8")
	// stopwordPath := filepath.Join(dictDir, "stop_words.utf8")
	// userPath := filepath.Join(dictDir, "user.dict.utf8")
	// Return the bot instance
	return &Bot{
		bot: bot,
		// jieba: gojieba.NewJieba(jiebaPath, hmmPath, idfPath, stopwordPath, userPath),
	}, nil
}

// Define a function to handle incoming messages
func (b *Bot) HandleMessage(update tgbotapi.Update) {
	// Check if the update is a message
	if update.Message == nil && update.ChannelPost == nil {
		return
	}

	// user := update.SentFrom()

	if update.Message != nil && update.Message.IsCommand() {
		go b.handleCommand(update.Message)
		return
	}
	if update.ChannelPost != nil && update.ChannelPost.IsCommand() {
		go b.handleCommand(update.ChannelPost)
		return
	}

	var msg *tgbotapi.MessageConfig
	chatType := update.FromChat()
	if chatType.IsChannel() {
		logger.Infof("receive channel msg:%v", update.ChannelPost)
		if !b.IsAtMe(update.ChannelPost.Text) {
			go b.MessageTypeHandler(update.ChannelPost)
			return
		}
		msg = b.publicMessge(update)
		if msg == nil {
			return
		}
	} else if chatType.IsPrivate() {
		logger.Infof("receive private msg:%v", update.Message)
		msg = b.privateMessage(update)
		if msg == nil {
			return
		}
	} else if chatType.IsGroup() || chatType.IsSuperGroup() {
		logger.Infof("receive group or supper group msg:%v", update.Message.Text)
		if !b.IsAtMe(update.Message.Text) {
			go b.MessageTypeHandler(update.Message)
			return
		}
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

func (b *Bot) IsAtMe(userName string) bool {
	me := b.bot.Self.UserName
	logger.Infof("%s-%s", me, userName)
	if strings.Contains(userName, "@") && strings.Contains(userName, me) {
		return true
	}
	return false
}

func (b *Bot) publicMessge(update tgbotapi.Update) *tgbotapi.MessageConfig {
	logger.Infof("group msg:%v", update.ChannelPost.Text)

	userName, replayText := b.makeReplyText(update.ChannelPost)
	if replayText == "" {
		return nil
	}

	msg := tgbotapi.NewMessage(update.ChannelPost.Chat.ID, replayText)
	msg.ReplyToMessageID = update.ChannelPost.MessageID
	logger.Infof("group msg:%v,%v", userName, replayText)

	return &msg
}

func (b *Bot) privateMessage(update tgbotapi.Update) *tgbotapi.MessageConfig {
	sendUserName, replayText := b.makeReplyText(update.Message)
	if replayText == "" {
		return nil
	}

	if sendUserName != "" {
		fromUserName := update.Message.From.UserName
		replayText = "@" + fromUserName + " " + replayText
	} else {
		err := b.MessageTypeHandler(update.Message)
		if err != nil {
			return nil
		}
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

func (b *Bot) MessageTypeHandler(msg *tgbotapi.Message) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("MessageTypeHandler recovered:", r)
		}
	}()
	var err error
	var fileID string
	var mediaType models.MediaType
	if msg.Text != "" {
		fileID = ""
		mediaType = models.Text
	} else if msg.Voice != nil {
		fileID = ""
		mediaType = models.Voice
	} else if msg.Video != nil {
		fileID = ""
		mediaType = models.Video
	} else if msg.Audio != nil {
		fileID = ""
		mediaType = models.Voice
	} else if msg.Document != nil {
		fileID = ""
		mediaType = models.File
	} else if msg.Photo != nil {
		fileID = ""
		mediaType = models.Picture
	}

	var reader io.ReadCloser
	if fileID != "" {
		url, err := b.bot.GetFileDirectURL(fileID)
		if err != nil {
			logger.Error(fmt.Sprintf("Voice GetFileDirectURL: %v", err.Error()))
			return err
		}
		reader, err = b.DownloadFile(url)
		if err != nil {
			logger.Error(fmt.Sprintf("Voice DownloadFile: %v", err.Error()))
			return err
		}
		defer reader.Close()
	}

	err = chat.Recorder(mediaType, msg.Text, fileID, reader)

	return err
}

func (b *Bot) makeReplyText(message *tgbotapi.Message) (userName, replayText string) {

	// 判断消息类型分别处理
	var err error
	var request string
	if message.Voice != nil {
		url, err := b.bot.GetFileDirectURL(message.Voice.FileID)
		if err != nil {
			logger.Error(fmt.Sprintf("Voice GetFileDirectURL: %v", err.Error()))
			return
		}
		reader, err := b.DownloadFile(url)
		if err != nil {
			logger.Error(fmt.Sprintf("Voice DownloadFile: %v", err.Error()))
			return
		}
		defer reader.Close()
		replayText, err = chat.Dialogue(models.Voice, "", message.Voice.FileID, reader)
		if err != nil {
			logger.Error(fmt.Sprintf("Voice: %v", err.Error()))
			return
		}
		logger.Info(fmt.Sprintf("Voice replayText: %v", replayText))
	} else if message.Text != "" {
		userName, request = b.getSendUserName(message.Text)
		replayText, err = chat.Dialogue(models.Text, request, "", nil)
		if err != nil {
			logger.Error(fmt.Sprintf("Text: %v", err.Error()))
			return
		}
		logger.Info(fmt.Sprintf("Text: %v", replayText))
	} else {
		err := b.MessageTypeHandler(message)
		if err != nil {
			return
		}
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

func (b *Bot) Destroy() {
	if b.jieba != nil {
		b.jieba.Free()
	}
}

// Download file from url use resty
func (b *Bot) DownloadFile(url string) (io.ReadCloser, error) {
	// Create a new Resty client
	client := resty.New()

	// Send the GET request and get the response
	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}
	// Return the response body as an io Reader
	return resp.RawResponse.Body, nil
}

// Define the main function
func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("main recovered:", r)
		}
	}()
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
		bot.Destroy()
		logger.Fatalf("Error starting bot: %s", err)
	}
}
