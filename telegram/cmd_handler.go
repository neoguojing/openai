package main

import (
	"errors"
	"fmt"
	"html/template"
	"math"
	"sort"
	"strconv"
	"strings"
	"math/rand"
    "time"

	"github.com/neoguojing/log"
	"github.com/neoguojing/openai/models"
	tgbotapi "github.com/neoguojing/telegram-bot-api/v5"
)



var (
	NO_JIEBA_ERROR = errors.New("b.jieba was nil")
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
	case "/report":
		reply = b.handleReport(args)
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

func (b *Bot) handleReport(args []string) string {
	var reply string
	if len(args) < 2 {
		reply = "need username and tag type:\n f: female\n m: man\n s: cheater\n a: admin\n other: "

	} else {
		userName := ""
		chatId, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			userName = args[0]
		}
		u := models.TelegramUserInfo{}
		user, err := u.FindByChatIDOrUsername(chatId, userName)
		if err != nil {
			logger.Errorf("Error finding user: %s", err)
			return err.Error()
		}

		err = u.UpdateTag(user.ChatID, models.USER_TAG(args[1]))
		if err != nil {
			logger.Errorf("Error handleStart: %s", err)
		}

		s := models.TelegramUserSummary{}
		err = s.UpdateLabel(user.ChatID, models.USER_TAG(args[1]))
		if err != nil {
			logger.Errorf("Error handleStart: %s", err)
		}
		reply = "report success"
	}
	return reply
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

func (b *Bot) search(args []string) (UserMap, error) {
	// TODO: Implement search functionality
	locations, keyword, err := b.handlePos(args)
	if err != nil {
		if err != NO_JIEBA_ERROR {
			logger.Errorf("Error handling POS: %s", err)
			return nil, err
		} else {
			keyword = args
		}

	}
	logger.Infof("TelegramProfile find-------------")
	p := &models.TelegramProfile{}
	var profiles []models.TelegramProfile
	if len(locations) != 0 && len(keyword) != 0 {
		profiles, err = p.FindByLocationAndKeyword(locations, keyword, 0, 0)
	} else if len(locations) != 0 {
		profiles, err = p.FindByLocations(locations, 0, 0)
	} else if len(keyword) != 0 {
		profiles, err = p.FindByKeywords(keyword, 0, 0)
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
		var profileMap = make(map[int64]models.TelegramProfile)
		var chatIDs = make([]int64, 0)
		for _, profile := range profiles {
			chatIDs = append(chatIDs, profile.ChatID)
			profileMap[profile.ChatID] = profile
		}

		s := &models.TelegramUserSummary{}
		var summrayMap = make(map[int64]models.TelegramUserSummary)
		summarys, err := s.FindByChatIDs(chatIDs)
		if err != nil {
			logger.Errorf("TelegramUserSummary Error finding users: %s", err)
			return nil, err
		}
		chatIDs = make([]int64, 0)
		for _, item := range summarys {
			chatIDs = append(chatIDs, item.ChatID)
			summrayMap[item.ChatID] = item
		}

		u := &models.TelegramUserInfo{}
		var userMap = make(map[int64]models.TelegramUserInfo)
		users, err := u.FindByChatIDs(chatIDs)
		if err != nil {
			logger.Errorf("TelegramUserInfo Error finding users: %s", err)
			return nil, err
		}
		for _, item := range users {
			chatIDs = append(chatIDs, item.ChatID)
			userMap[item.ChatID] = item
		}
		logger.Infof("found %d TelegramUserInfo", len(users))
		useFullArr := dataRecall(userMap, profileMap, summrayMap)

		return useFullArr, nil
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
	profiles, err := p.FindByLocations(args, 0, 0)
	if err != nil {
		logger.Errorf("Error finding profiles: %s", err)
		return err.Error()
	}

	if len(profiles) == 0 {
		logger.Errorf("Error finding profiles: no results found")
		return "No results found"
	}

	var profileMap = make(map[int64]models.TelegramProfile)
	var chatIDs = make([]int64, 0)
	for _, profile := range profiles {
		chatIDs = append(chatIDs, profile.ChatID)
		profileMap[profile.ChatID] = profile
	}

	s := &models.TelegramUserSummary{}
	var summrayMap = make(map[int64]models.TelegramUserSummary)
	summarys, err := s.FindByChatIDs(chatIDs)
	if err != nil {
		logger.Errorf("TelegramUserSummary Error finding users: %s", err)
		return err.Error()
	}

	chatIDs = make([]int64, 0)
	for _, item := range summarys {
		chatIDs = append(chatIDs, item.ChatID)
		summrayMap[item.ChatID] = item
	}

	u := &models.TelegramUserInfo{}
	var userMap = make(map[int64]models.TelegramUserInfo)
	users, err := u.FindByChatIDs(chatIDs)
	if err != nil {
		logger.Errorf("Error finding users: %s", err)
		return err.Error()
	}
	for _, item := range users {
		chatIDs = append(chatIDs, item.ChatID)
		userMap[item.ChatID] = item
	}

	useFullArr := dataRecall(userMap, profileMap, summrayMap)
	messages := generateTelegramMessages(useFullArr)
	reply := strings.Join(messages, "\n")

	return reply

}

func (b *Bot) handleUserName(args []string) string {
	if len(args) == 0 {
		logger.Errorf("Error handling user name: please provide a user name")
		return "Please provide a user name"
	}
	u := models.TelegramUserInfo{}
	user, err := u.FindByChatIDOrUsername(0, args[0])
	if err != nil {
		logger.Errorf("Error finding user: %s", err)
		return err.Error()
	}
	userFule := &UserInfoFull{
		User: *user,
	}
	replay, err := generateRecommendationMessage(userFule)
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
	if b.jieba == nil {
		return nil, nil, NO_JIEBA_ERROR
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

type UserInfoFull struct {
	User    models.TelegramUserInfo
	Profile models.TelegramProfile
	Message models.TelegramChatMessage
	Count   int64
	Score   float64
}

type UserMap []UserInfoFull

const TOPK = 5

func dataRecall(userInfos map[int64]models.TelegramUserInfo,
	profiles map[int64]models.TelegramProfile, summrays map[int64]models.TelegramUserSummary) UserMap {
	rand.Seed(time.Now().UnixNano())
	uMap := make(UserMap, 0)
	for id, u := range userInfos {
		uFull := UserInfoFull{
			User:  u,
			Score: 100,
		}

		if p, ok := profiles[id]; ok {
			uFull.Profile = p
		}

		if s, ok := summrays[id]; ok {
			uFull.Score *= math.Round(s.Confidence*100) / 100
		}

		uMap = append(uMap, uFull)
	}
	sort.Slice(uMap, func(i, j int) bool {
		return uMap[i].Score > uMap[j].Score
	})

	logger.Infof("dataRecall result:%v", uMap)
	if len(uMap) <= TOPK {
		return uMap
	}

	selected := make(UserMap, TOPK)
	for i:=0;i<TOPK;i++ {
		index := rand.Intn(len(uMap))
		selected[i] = uMap[index]
	}
	return selected
}

// 打分逻辑，匹配的关键值越靠前，则得分越高，得分越高则匹配度越高
func scoreUser(profile *models.TelegramProfile, keyword []string, location []string) float64 {
	score := 0.0
	kScore := 0.0
	lScore := 0.0

	if len(keyword) != 0 {
		for _, k := range keyword {
			pKeyWords := strings.Split(profile.Keywords, ",")
			for i, word := range pKeyWords {
				if len(k) > len(word) {
					if strings.Contains(k, word) {
						kScore += float64(len(pKeyWords) - i)
					}
				} else if len(word) > len(k) {
					if strings.Contains(word, k) {
						kScore += float64(len(pKeyWords) - i)
					}
				} else {
					if k == word {
						kScore += float64(len(pKeyWords) - i)
					}
				}
			}
		}
	}

	if len(location) != 0 {
		for _, k := range location {
			pKeyWords := strings.Split(profile.Location, ",")
			for i, word := range pKeyWords {
				if len(k) > len(word) {
					if strings.Contains(k, word) {
						lScore += float64(len(pKeyWords) - i)
					}
				} else if len(word) > len(k) {
					if strings.Contains(word, k) {
						lScore += float64(len(pKeyWords) - i)
					}
				} else {
					if k == word {
						lScore += float64(len(pKeyWords) - i)
					}
				}
			}
		}
	}
	score = (lScore + kScore) / 2.0
	return score
}

func generateRecommendationMessage(userInfo *UserInfoFull) (string, error) {
	messageTemplate := `👤 {{.Username}}-{{.ChatId}}
📝 {{.Bio}}
📝 {{.Urls}}
🕒 {{.UpdatedAt}}
🎖️ {{.Score}}
🔍 {{.Keywords}}
📍  {{.Location}}
📩 {{.LastMessageTime}}:{{.LastMessage}}
💬 {{.MessageTotal}}`

	tpl, err := template.New("recommendationMessage").Parse(messageTemplate)
	if err != nil {
		logger.Errorf("Error parsing message template: %s", err)
		return "", err
	}

	var tplData struct {
		FirstName       string
		LastName        string
		Username        string
		Bio             string
		Urls            string
		UpdatedAt       string
		Keywords        string
		Location        string
		LastMessageTime string
		LastMessage     string
		MessageTotal    int64
		Score           float64
		ChatId          int64
	}

	tplData.FirstName = userInfo.User.FirstName
	tplData.LastName = userInfo.User.LastName
	tplData.Username = "@" + userInfo.User.Username
	tplData.Bio = userInfo.User.Bio
	tplData.Urls = strings.Join(strings.Split(userInfo.Profile.Urls, ","), "\n")
	tplData.UpdatedAt = userInfo.User.UpdatedAt.Format("2006-01-02 15:04:05")
	tplData.Keywords = userInfo.Profile.Keywords
	tplData.Location = userInfo.Profile.Location
	tplData.LastMessageTime = userInfo.Message.UpdatedAt.Format("2006-01-02 15:04:05")
	tplData.LastMessage = userInfo.Message.Message
	tplData.MessageTotal = userInfo.Count
	tplData.Score = userInfo.Score
	tplData.ChatId = userInfo.User.ChatID
	var message strings.Builder
	err = tpl.Execute(&message, tplData)
	if err != nil {
		logger.Errorf("Error executing message template: %s", err)
		return "", err
	}

	return message.String(), nil
}

func generateTelegramMessages(userInfos UserMap) []string {
	var messages []string
	msg := &models.TelegramChatMessage{}
	for i, userInfo := range userInfos {
		m, err := msg.FindLatestMessageByChatID(userInfo.User.ChatID)
		if err != nil {
			log.Error(err.Error())
			m = msg
		}
		count, err := msg.CountMessagesByChatID(userInfo.User.ChatID)
		if err != nil {
			log.Error(err.Error())
		}
		userInfo.Message = *m
		userInfo.Count = count
		recomend, err := generateRecommendationMessage(&userInfo)
		if err != nil {
			log.Error(err.Error())
		} else {
			message := fmt.Sprintf("%d. %s", i+1, recomend)
			messages = append(messages, message)
		}

	}
	return messages
}
