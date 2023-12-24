package openai

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/neoguojing/log"
)

const (
	CLAUDE_V2      = "claude-2"
	CLAUDE_INSTANT = "claude-instant-1"
)

type ClaudeClient struct {
	client *resty.Client
	model  string
}

func NewClaudeClient(apiKey string) *ClaudeClient {
	client := resty.New()
	client.SetHeaders(map[string]string{
		"Accept":            "application/json",
		"Anthropic-Version": "2023-06-01",
		"Content-Type":      "application/json",
		"X-Api-Key":         apiKey,
	})
	return &ClaudeClient{client: client, model: CLAUDE_V2}
}

func (c *ClaudeClient) Complete(input string) (*ChatResponse, error) {
	resp, err := c.client.R().
		SetBody(ClaudeRequest{
			Model:             c.model,
			Prompt:            input,
			MaxTokensToSample: 256,
		}).
		Post("https://api.anthropic.com/v1/complete")

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	var result ClaudeResponse
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}

	return ConvertClaudeToOpenai(&result), nil
}