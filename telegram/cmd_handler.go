package main

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/neoguojing/openai/models"
	tgbotapi "github.com/neoguojing/telegram-bot-api/v5"
	"github.com/yanyiwu/gojieba"
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
	case "/key":
		reply = b.handleSearch(args)
		// Create a message with a photo
	case "/locate":
	case "/username":
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

	results, err := b.search(args)
	if err != nil {
		logger.Errorf("Error searching: %s", err)
		return "Error searching"
	}

	if len(results) == 0 {
		return "No results found"
	}

	var reply strings.Builder
	reply.WriteString("Results:\n")

	return reply.String()
}

func (b *Bot) search(args []string) ([]models.TelegramProfile, error) {
	// TODO: Implement search functionality
	locations, keyword, err := b.handlePos(args)
	if err != nil {
		return nil, err
	}

	var profiles []models.TelegramProfile
	if len(locations) != 0 && len(keyword) != 0 {
		p := &models.TelegramProfile{}
		profiles, err = p.FindByLocationAndKeyword(locations, keyword, 6, 0)
	} else if len(locations) != 0 {
		p := &models.TelegramProfile{}
		profiles, err = p.FindByLocations(locations, 6, 0)
	} else if len(keyword) != 0 {
		p := &models.TelegramProfile{}
		profiles, err = p.FindByKeywords(keyword, 6, 0)
	} else {
		return nil, errors.New("Please provide a query to search for")
	}
	return profiles, err
}

func (b *Bot) handleLocate(args []string) string {
	if len(args) == 0 {
		return "pls input location"
	}
	p := &models.TelegramProfile{}
	profiles, err := p.FindByLocations(args, 6, 0)
	if err != nil {
		return err.Error()
	}

	var chatIDs []int64
	for _, profile := range profiles {
		chatIDs = append(chatIDs, profile.ChatID)
	}

	u := &models.TelegramUserInfo{}
	users, err := u.FindByChatIDs(chatIDs)
	if err != nil {
		return err.Error()
	}

	var reply strings.Builder
	reply.WriteString("Users:\n")
	for _, user := range users {
		reply.WriteString(user.Username)
		reply.WriteString("\n")
	}

	return reply.String()

}

func (b *Bot) handleUserName(args []string) string {
	if len(args) == 0 {
		return "pls input user name"
	}
	u := &models.TelegramUserInfo{}
	user, err := u.FindByChatIDOrUsername(0, args[0])
	if err != nil {
		return err.Error()
	}

	replay, err := generateRecommendationMessage(*user)
	if err != nil {
		return err.Error()
	}
	return replay
}

func (b *Bot) handlePos(args []string) ([]string, []string, error) {
	if len(args) == 0 {
		return nil, nil, errors.New("Please provide a sentence to analyze")
	}
	x := gojieba.NewJieba()
	defer x.Free()

	var locations []string
	var nv []string
	for _, arg := range args {
		words := x.Tag(arg)
		for _, word := range words {
			if word.Tag == "ns" {
				locations = append(locations, word.Word)
			} else if word.Tag == "n" {
				nv = append(nv, word.Word)
			} else if word.Tag == "v" {
				nv = append(nv, word.Word)
			}
		}
	}

	return locations, nv, nil
}

func generateRecommendationMessage(userInfo models.TelegramUserInfo) (string, error) {
	messageTemplate := `👤 {{.Username}}
📝 {{.Bio}}
🕒 {{.UpdatedAt}}`
	tpl, err := template.New("recommendationMessage").Parse(messageTemplate)
	if err != nil {
		logger.Errorf("Error parsing message template: %s", err)
		return "", err
	}

	var tplData struct {
		FirstName string
		LastName  string
		Username  string
		Bio       string
		UpdatedAt string
	}
	tplData.FirstName = userInfo.FirstName
	tplData.LastName = userInfo.LastName
	tplData.Username = userInfo.Username
	tplData.Bio = userInfo.Bio
	tplData.UpdatedAt = userInfo.UpdatedAt.Format("2006-01-02 15:04:05")

	var message strings.Builder
	err = tpl.Execute(&message, tplData)
	if err != nil {
		logger.Errorf("Error executing message template: %s", err)
		return "", err
	}

	return message.String(), nil
}

func generateTelegramMessages(userInfos []models.TelegramUserInfo) []string {
	var messages []string
	for i, userInfo := range userInfos {
		recomend, _ := generateRecommendationMessage(userInfo)
		message := fmt.Sprintf("%d. %s", i+1, recomend)
		messages = append(messages, message)
	}
	return messages
}
