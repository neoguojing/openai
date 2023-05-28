package openai

import (
	"encoding/json"
	"errors"
	"io"
	"log"

	"github.com/go-resty/resty/v2"
	"github.com/neoguojing/openai/models"
)

type Chat struct {
	apiKey      string
	url         string
	model       string
	role        OpenAIRole
	request     string
	response    string
	instruction string
	audio       Audio
}

type ChatOption func(*Chat)

func WithChatModel(model string) ChatOption {
	return func(c *Chat) {
		c.model = model
	}
}

func WithChatInput(text string) ChatOption {
	return func(c *Chat) {
		c.request = text
	}
}

func (o *OpenAI) Chat(opts ...ChatOption) *Chat {
	c := &Chat{
		url:    "https://api.openai.com/v1/chat/completions",
		apiKey: o.apiKey,
		model:  "gpt-3.5-turbo",
		role:   User,
	}

	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (o *Chat) Prepare(roleName string) *Chat {
	roles, err := models.SearchRoleByName(roleName)
	if err != nil {
		log.Println(err)
		return nil
	}

	if len(roles) == 0 {
		log.Println("roles was empty")
		return nil
	}
	chatResponse, err := o.Complete(roles[0].Desc)
	if err != nil {
		log.Println(err)
		return nil
	}
	log.Println(chatResponse.Choices[0].Message.Content)
	return o
}

func (o *OpenAI) PreProcessForChat(media models.MediaType, text string, filePath string,
	reader io.Reader) *Chat {
	var input string
	if media == models.Voice {
		audioResp, err := o.Audio().TranscriptionsDirect(filePath, reader)
		if err != nil {
			log.Println(err)
			return nil
		}
		input = audioResp.Text
	} else if media == models.Picture {
	} else if media == models.Text {
		input = text
	} else if media == models.Video {
	} else if media == models.File {
	}

	return o.Chat(WithChatInput(input))
}

func (o *Chat) CompleteWithPrepareInput() (*ChatResponse, error) {
	return o.Complete(o.request)
}

func (o *Chat) Complete(content string) (*ChatResponse, error) {
	if content == "" {
		return nil, errors.New("empty input")
	}

	client := resty.New()
	req := ChatRequest{
		Model: o.model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    string(o.role),
				Content: content,
			},
		},
	}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(req).
		Post(o.url)
	if err != nil {
		return nil, err
	}
	var chatResponse ChatResponse
	err = json.Unmarshal(resp.Body(), &chatResponse)
	if err != nil {
		return nil, err
	}
	o.request = content
	o.response, _ = chatResponse.GetContent()
	return &chatResponse, nil
}

func (o *Chat) Edits(content string, instruction string) (*EditChatResponse, error) {
	url := "https://api.openai.com/v1/edits"

	req := EditChatRequest{
		Model: "text-davinci-edit-001",
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    string(o.role),
				Content: content,
			},
		},
		Instruction: instruction,
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
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

	o.request = content
	o.response, _ = output.GetContent()
	o.instruction = instruction
	return &output, nil
}
