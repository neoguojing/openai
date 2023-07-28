package openai

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/neoguojing/log"

	"github.com/go-resty/resty/v2"
	"github.com/neoguojing/openai/config"
	"github.com/neoguojing/openai/models"
	"github.com/neoguojing/openai/utils"
)

var (
	baseFilePath = os.Getenv(config.EnvFilePath)
)

type Chat struct {
	apiKey    string
	url       string
	model     string
	role      OpenAIRole
	audio     *Audio
	client    *resty.Client
	recorder  *models.Recorder
	platform  models.Platform
	completor IChat
	cache     *utils.LRUCache
}

type ChatOption func(*Chat)

func WithChatModel(model string) ChatOption {
	return func(c *Chat) {
		c.model = model
	}
}

func WithProxy(proxyURL string) ChatOption {

	return func(c *Chat) {
		c.client.SetProxy(proxyURL)
	}
}

func WithPlatform(p models.Platform) ChatOption {
	return func(c *Chat) {
		c.platform = p
	}
}

func WithComplete(chatType ChatType) ChatOption {
	completor := GlobalChatFactory.GetChat(chatType)
	return func(c *Chat) {
		c.completor = completor
	}
}

func lruCallBack(key string, value interface{}, freq int) {

}

func (o *OpenAI) Chat(opts ...ChatOption) *Chat {
	c := &Chat{
		url:      "https://api.openai.com/v1/chat/completions",
		apiKey:   o.apiKey,
		model:    "gpt-3.5-turbo",
		role:     User,
		client:   resty.New(),
		audio:    o.Audio(),
		recorder: models.GetRecorder(),
		cache:    utils.NewLRUCache(500, lruCallBack),
	}

	for _, opt := range opts {
		opt(c)
	}
	c.init()
	return c
}

func (c *Chat) init() {
	// load cache
	record := models.ChatRecord{}
	records, err := record.GetChatRecordsByFrequency(0, 500)
	if err != nil {
		log.Error("Chat:init:%v", err.Error())
	} else {
		for i := len(records) - 1; i >= 0; i-- {
			c.cache.Set(records[i].Request, records[i].Reply, 0)
		}

	}

}

func (c *Chat) Prepare(roleName string) *Chat {
	roles, err := models.SearchRoleByName(roleName)
	if err != nil {
		log.Error(err.Error())
		return nil
	}

	if len(roles) == 0 {
		log.Error("roles was empty")
		return nil
	}
	chatResponse, err := c.Complete(roles[0].Desc)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	content, _ := chatResponse.GetContent()
	log.Info(content)
	return c
}

func (c *Chat) save(filePath string, reader io.Reader) (string, error) {
	go func() error {

		if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
			if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
				log.Error(err.Error())
				return err
			}
		}

		file, err := os.Create(filePath)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		defer file.Close()

		_, err = io.Copy(file, reader)
		if err != nil {
			log.Error(err.Error())
			return err
		}
		return nil
	}()

	return filePath, nil
}

func (c *Chat) Dialogue(media models.MediaType, text string, filePath string,
	reader io.Reader) (string, error) {
	if text == "" && reader == nil {
		return "", errors.New("empty input")
	}

	var input string
	var dstFilePath string
	if media == models.Voice {
		audioResp, err := c.audio.TranscriptionsDirect(filePath, reader)
		if err != nil {
			log.Error(err.Error())
			return "", err
		}
		input = audioResp.Text
		dst := filepath.Join(baseFilePath, string(models.Voice), filePath)
		dstFilePath, _ = c.save(dst, reader)
	} else if media == models.Picture {
	} else if media == models.Text {
		input = text
	} else if media == models.Video {
	} else if media == models.File {
	}

	if value := c.cache.Get(input); value != nil {
		reply := value.(string)
		return reply, nil
	}

	resp, err := c.Complete(input)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}
	reply, err := resp.GetContent()
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	c.cache.Set(input, reply, 0)

	record := models.ChatRecord{
		Request:   input,
		Reply:     reply,
		MediaType: media,
		FilePath:  dstFilePath,
		Frequency: 1,
	}
	elem := models.Element{
		Operation:  models.Create,
		ChatRecord: &record,
	}
	c.recorder.Send(elem)

	return reply, nil
}

// 默认对话
func (c *Chat) Complete(content string) (*ChatResponse, error) {
	if content == "" {
		return nil, errors.New("empty input")
	}

	if c.completor != nil {
		return c.completor.Complete(content)
	}

	req := ChatRequest{
		Model: c.model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    string(c.role),
				Content: content,
			},
		},
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+c.apiKey).
		SetBody(req).
		Post(c.url)
	if err != nil {
		return nil, err
	}
	var chatResponse ChatResponse
	err = json.Unmarshal(resp.Body(), &chatResponse)
	if err != nil {
		return nil, err
	}

	return &chatResponse, nil
}

func (c *Chat) Edits(content string, instruction string) (*EditChatResponse, error) {
	url := "https://api.openai.com/v1/edits"

	req := EditChatRequest{
		Model: "text-davinci-edit-001",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    string(c.role),
				Content: content,
			},
		},
		Instruction: instruction,
	}

	resp, err := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+c.apiKey).
		SetBody(req).
		Post(url)
	if err != nil {
		return nil, err
	}
	var output EditChatResponse
	err = json.Unmarshal(resp.Body(), &output)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (c *Chat) Recorder(media models.MediaType, text string, filePath string,
	reader io.Reader) error {

	record := models.ChatRecord{
		Request:   text,
		MediaType: media,
	}
	var dstFilePath string
	switch media {
	case models.File:
		dst := filepath.Join(baseFilePath, string(models.File), filePath)
		dstFilePath, _ = c.save(dst, reader)
	case models.Picture:
		dst := filepath.Join(baseFilePath, string(models.Picture), filePath)
		dstFilePath, _ = c.save(dst, reader)
	case models.Video:
		dst := filepath.Join(baseFilePath, string(models.Video), filePath)
		dstFilePath, _ = c.save(dst, reader)
	case models.Voice:
		dst := filepath.Join(baseFilePath, string(models.Voice), filePath)
		dstFilePath, _ = c.save(dst, reader)
	case models.Text:
		record.Request = text
	default:
		return errors.New("not support type")
	}
	record.FilePath = dstFilePath

	elem := models.Element{
		Operation:  models.Create,
		ChatRecord: &record,
	}
	c.recorder.Send(elem)
	return nil
}
