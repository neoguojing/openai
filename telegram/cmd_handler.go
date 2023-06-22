package main

import (
	"errors"
	"fmt"
	"html/template"
	"strings"

	"github.com/neoguojing/log"
	"github.com/neoguojing/openai/models"
	tgbotapi "github.com/neoguojing/telegram-bot-api/v5"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("handleCommand recovered:", r)
		}
	}()
	log.Infof("recieve a command message:%v", message.Text)
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
	case "/locate":
		reply = b.handleLocate(args)
	case "/username":
		reply = b.handleUserName(args)
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
	var reply string
	if len(args) == 0 {
		reply = "Welcome!"
	} else {
		reply = fmt.Sprintf("Welcome, %s!", args[0])
	}
	return reply
}

func (b *Bot) handleHelp(args []string) string {
	var commands []string
	commands = append(commands, "/search [query] - search for something")
	commands = append(commands, "/help - show this help message")
	reply := strings.Join(commands, "\n")
	return reply
}

func (b *Bot) handleSearch(args []string) string {
	if len(args) == 0 {
		logger.Errorf("Error searching: please provide a query to search for")
		return "Please provide a query to search for"
	}

	results, err := b.search(args)
	if err != nil {
		logger.Errorf("Error searching: %s", err)
		return "Error searching"
	}

	if len(results) == 0 {
		logger.Errorf("Error searching: no results found")
		return "No results found"
	}

	logger.Infof("generateTelegramMessages")
	messages := generateTelegramMessages(results)
	reply := strings.Join(messages, "\n")
	logger.Infof("generateTelegramMessages-----Done")
	return reply
}

func (b *Bot) search(args []string) ([]models.TelegramUserInfo, error) {
	// TODO: Implement search functionality
	locations, keyword, err := b.handlePos(args)
	if err != nil {
		logger.Errorf("Error handling POS: %s", err)
		return nil, err
	}
	logger.Infof("TelegramProfile find-------------")
	p := &models.TelegramProfile{}
	var profiles []models.TelegramProfile
	if len(locations) != 0 && len(keyword) != 0 {
		profiles, err = p.FindByLocationAndKeyword(locations, keyword, 6, 0)
	} else if len(locations) != 0 {
		profiles, err = p.FindByLocations(locations, 6, 0)
	} else if len(keyword) != 0 {
		profiles, err = p.FindByKeywords(keyword, 6, 0)
	} else {
		err = errors.New("please provide a query to search for")
		logger.Errorf("Error searching: %s", err)
		return nil, err
	}

	if err != nil {
		logger.Errorf("Error searching: %s", err)
		return nil, err
	}
	logger.Infof("found %d TelegramProfiles", len(profiles))
	if len(profiles) != 0 {
		logger.Infof("TelegramUserInfo find-------------")
		var chatIDs []int64
		for _, profile := range profiles {
			chatIDs = append(chatIDs, profile.ChatID)
		}

		u := &models.TelegramUserInfo{}
		users, err := u.FindByChatIDs(chatIDs)
		if err != nil {
			logger.Errorf("Error finding users: %s", err)
			return nil, err
		}
		logger.Infof("found %d TelegramProfiles", len(users))
		return users, nil
	}
	err = errors.New("no results found")
	logger.Errorf("Error searching: %s", err)
	return nil, err
}

func (b *Bot) handleLocate(args []string) string {
	if len(args) == 0 {
		return "pls input location"
	}
	p := &models.TelegramProfile{}
	profiles, err := p.FindByLocations(args, 6, 0)
	if err != nil {
		logger.Errorf("Error finding profiles: %s", err)
		return err.Error()
	}

	if len(profiles) == 0 {
		logger.Errorf("Error finding profiles: no results found")
		return "No results found"
	}

	var chatIDs []int64
	for _, profile := range profiles {
		chatIDs = append(chatIDs, profile.ChatID)
	}

	u := &models.TelegramUserInfo{}
	users, err := u.FindByChatIDs(chatIDs)
	if err != nil {
		logger.Errorf("Error finding users: %s", err)
		return err.Error()
	}

	messages := generateTelegramMessages(users)
	reply := strings.Join(messages, "\n")

	return reply

}

func (b *Bot) handleUserName(args []string) string {
	if len(args) == 0 {
		logger.Errorf("Error handling user name: please provide a user name")
		return "Please provide a user name"
	}
	u := &models.TelegramUserInfo{}
	user, err := u.FindByChatIDOrUsername(0, args[0])
	if err != nil {
		logger.Errorf("Error finding user: %s", err)
		return err.Error()
	}

	replay, err := generateRecommendationMessage(*user)
	if err != nil {
		logger.Errorf("Error generating recommendation message: %s", err)
		return err.Error()
	}
	return replay
}

func (b *Bot) handlePos(args []string) ([]string, []string, error) {
	logger.Infof("handlePos-------------")
	if len(args) == 0 {
		return nil, nil, errors.New("please provide a sentence to analyze")
	}

	var locations []string
	var nv []string
	for _, arg := range args {
		words := b.jieba.Tag(arg)
		for _, word := range words {
			tags := strings.Split(word, "/")
			logger.Info(tags...)
			if len(tags) > 1 {
				switch tags[1] {
				case "ns":
					locations = append(locations, tags[0])
				case "n", "v":
					nv = append(nv, tags[0])
				}
			}
		}
	}
	logger.Infof("locations=%v", locations)
	logger.Infof("nv=%v", nv)
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
