package claude

import (
	"github.com/go-resty/resty/v2"
	"github.com/neoguojing/log"
)

const (
	CLAUDE_V2      = "claude-2"
	CLAUDE_INSTANT = "claude-instant-1"
)

type CompletionResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
	Model      string `json:"model"`
}

type Request struct {
	Model             string `json:"model"`
	Prompt            string `json:"prompt"`
	MaxTokensToSample int    `json:"max_tokens_to_sample"`
}

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

func (c *ClaudeClient) Complete(input string) (*CompletionResponse, error) {
	resp := &CompletionResponse{}
	_, err := c.client.R().
		SetBody(Request{
			Model:             c.model,
			Prompt:            input,
			MaxTokensToSample: 256,
		}).
		SetResult(resp).
		Post("https://api.anthropic.com/v1/complete")

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return resp, nil
}
