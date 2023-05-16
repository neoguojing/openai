package openai

import (
	"encoding/json"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Option func(*Chat)

func WithModel(model string) Option {
	return func(c *Chat) {
		c.model = model
	}
}

func WithRole(role OpenAIRole) Option {
	return func(c *Chat) {
		c.role = role
	}
}

type OpenAI struct {
	apiKey string
	url    string
}

type Model struct {
	ModelList
	apiKey string
	url    string
}

type Chat struct {
	apiKey string
	url    string
	model  string
	role   OpenAIRole
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{apiKey: apiKey}
}

func (o *OpenAI) Model() *Model {
	return &Model{
		url:    "https://api.openai.com/v1/models",
		apiKey: o.apiKey,
	}
}

func (o *Model) List() (*ModelList, error) {
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("OpenAI-Organization", "org-U3jJBNZ72nnwuS5qRKQOVhcS").
		Get(o.url)
	if err != nil {
		return nil, err
	}
	var modelList ModelList
	err = json.Unmarshal(resp.Body(), &modelList)
	if err != nil {
		return nil, err
	}
	o.ModelList = modelList
	return &modelList, nil
}

func (o *Model) GetModelInfo(model string) (*ModelInfo, error) {
	o.url += "/" + model
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(o.url)
	if err != nil {
		return nil, err
	}
	var modelInfo ModelInfo
	err = json.Unmarshal(resp.Body(), &modelInfo)
	if err != nil {
		return nil, err
	}
	return &modelInfo, nil
}

func (o *OpenAI) Completions(message string, maxTokens int) (*CompletionResponse, error) {
	o.url = "https://api.openai.com/v1/completions"
	client := resty.New()
	req := CompletionRequest{
		Model:       "gpt-3.5-turbo",
		Prompt:      message,
		MaxTokens:   maxTokens,
		Temperature: 0.7,
	}
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(req).
		Post(o.url)
	if err != nil {
		return nil, err
	}
	var completionResponse CompletionResponse
	err = json.Unmarshal(resp.Body(), &completionResponse)
	if err != nil {
		return nil, err
	}
	return &completionResponse, nil
}

func (o *OpenAI) Chat() *Chat {

	return &Chat{
		url:    "https://api.openai.com/v1/chat/completions",
		apiKey: o.apiKey,
		model:  "gpt-3.5-turbo",
	}
}

func (o *Chat) Completions(content string) (*ChatResponse, error) {
	client := resty.New()
	req := ChatRequest{
		Model: o.model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
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
	return &chatResponse, nil
}

func (o *Chat) Edits(content string, instruction string) (*EditChatResponse, error) {
	url := "https://api.openai.com/v1/edits"

	req := EditChatRequest{
		Model: o.model,
		Messages: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{
				Role:    "user",
				Content: content,
			},
		},
		Instruction: "",
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
	return &output, nil
}

func (o *OpenAI) GetImageGenerations(prompt string, n int, size string) ([]byte, error) {
	url := "https://api.openai.com/v1/images/generations"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(`{
			"prompt": "` + prompt + `",
			"n": ` + strconv.Itoa(n) + `,
			"size": "` + size + `"
		}`).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetImageEdits(imagePath string, maskPath string, prompt string, n int, size string) ([]byte, error) {
	url := "https://api.openai.com/v1/images/edits"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetFileReader("image", imagePath, nil).
		SetFileReader("mask", maskPath, nil).
		SetFormData(map[string]string{
			"prompt": prompt,
			"n":      strconv.Itoa(n),
			"size":   size,
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetImageVariations(imagePath string, n int, size string) ([]byte, error) {
	url := "https://api.openai.com/v1/images/variations"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetFileReader("image", imagePath, nil).
		SetFormData(map[string]string{
			"n":    strconv.Itoa(n),
			"size": size,
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetEmbeddings(input string, model string) ([]byte, error) {
	url := "https://api.openai.com/v1/embeddings"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(`{
			"input": "` + input + `",
			"model": "` + model + `"
		}`).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetAudioTranscriptions(filePath string) ([]byte, error) {
	url := "https://api.openai.com/v1/audio/transcriptions"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFileReader("file", filePath, nil).
		SetFormData(map[string]string{
			"model": "whisper-1",
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetAudioTranslations(filePath string) ([]byte, error) {
	url := "https://api.openai.com/v1/audio/translations"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFileReader("file", filePath, nil).
		SetFormData(map[string]string{
			"model": "whisper-1",
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetFiles() ([]byte, error) {
	url := "https://api.openai.com/v1/files"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

// New code starts here
func (o *OpenAI) UploadFile(filePath string) ([]byte, error) {
	url := "https://api.openai.com/v1/files"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFormData(map[string]string{
			"purpose": "fine-tune",
		}).
		SetFileReader("file", filePath, nil).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

// New code starts here
func (o *OpenAI) DeleteFile(fileID string) ([]byte, error) {
	url := "https://api.openai.com/v1/files/" + fileID
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Delete(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetFile(fileID string) ([]byte, error) {
	url := "https://api.openai.com/v1/files/" + fileID
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) DownloadFile(fileID string) ([]byte, error) {
	url := "https://api.openai.com/v1/files/" + fileID + "/content"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) FineTune(fileID string) ([]byte, error) {
	url := "https://api.openai.com/v1/fine-tunes"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(`{
			"training_file": "` + fileID + `"
		}`).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) ListFineTunes() ([]byte, error) {
	url := "https://api.openai.com/v1/fine-tunes"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetFineTune(fine_tune_id string) ([]byte, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

// New code starts here
func (o *OpenAI) CancelFineTune(fine_tune_id string) ([]byte, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id + "/cancel"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetEvents() ([]byte, error) {
	url := "https://api.openai.com/v1/fine-tunes/ft-AF1WoRqd3aJAHsqc9NY7iL8F/events"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) DeleteFineTune(fine_tune_id string) ([]byte, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Delete(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GenerateModeration(input string) ([]byte, error) {
	url := "https://api.openai.com/v1/moderations"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(`{
			"input": "` + input + `"
		}`).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}
