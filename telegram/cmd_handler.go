package main

import (
	"strings"

	tgbotapi "github.com/neoguojing/telegram-bot-api/v5"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) {

	// Split the command into its parts
	command := strings.Split(message.Text, " ")[0]
	args := strings.Split(message.Text, " ")[1:]

	var reply string
	// Handle the command
	switch command {
	case "/start":
		reply = b.handleStart(args)
	case "/help":
		reply = b.handleHelp(args)
	case "/search":
		reply = b.handleSearch(args)
		// Create a message with a photo
	case "/photo":
		photoConfig := tgbotapi.NewPhoto(message.Chat.ID, nil)
		photoConfig.Caption = "This is a random photo"
		photoConfig.ParseMode = tgbotapi.ModeMarkdown

		_, err := b.bot.Send(photoConfig)
		if err != nil {
			logger.Errorf("Error sending photo: %s", err)
		}
		return
	default:
		_, reply = b.makeReplyText(message)
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, reply)
	_, err := b.bot.Send(msg)
	if err != nil {
		logger.Errorf("Error sending message: %s", err)
	}
}

func (b *Bot) handleStart(args []string) string {
	return "Welcome to my bot!"
}

func (b *Bot) handleHelp(args []string) string {
	return "Here are the available commands:\n/search [query] - search for something\n/help - show this help message"
}

func (b *Bot) handleSearch(args []string) string {
	if len(args) == 0 {
		return "Please provide a query to search for"
	}

	query := strings.Join(args, " ")
	results, err := b.search(query)
	if err != nil {
		logger.Errorf("Error searching: %s", err)
		return "Error searching"
	}

	if len(results) == 0 {
		return "No results found"
	}

	var reply strings.Builder
	reply.WriteString("Results:\n")
	for _, result := range results {
		reply.WriteString(result.Title)
		reply.WriteString("\n")
		reply.WriteString(result.URL)
		reply.WriteString("\n\n")
	}

	return reply.String()
}

func (b *Bot) search(query string) ([]*Result, error) {
	// TODO: Implement search functionality
	return nil, nil
}

type Result struct {
	Title string
	URL   string
}
