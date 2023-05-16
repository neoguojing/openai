package openai

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type OpenAIOption func(*OpenAI)

func WithModel(model string) OpenAIOption {
	return func(o *OpenAI) {
		o.model = model
	}
}

type OpenAI struct {
	apiKey string
	url    string
	model  string
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

type ChatOption func(*Chat)

func WithChatModel(model string) ChatOption {
	return func(c *Chat) {
		c.model = model
	}
}

func WithRole(role OpenAIRole) ChatOption {
	return func(c *Chat) {
		c.role = role
	}
}

type Image struct {
	apiKey string
	url    string
}

type Audio struct {
	apiKey string
	url    string
}

type TuneFile struct {
	apiKey string
	url    string
}

type FineTune struct {
	apiKey string
	url    string
}

func NewOpenAI(apiKey string, opts ...OpenAIOption) *OpenAI {
	o := &OpenAI{apiKey: apiKey, model: "gpt-3.5-turbo"}
	for _, opt := range opts {
		opt(o)
	}
	return o
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

func (o *Model) Get(model string) (*ModelInfo, error) {
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
		Model:       o.model,
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

func (o *Chat) Completions(content string) (*ChatResponse, error) {
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
	return &output, nil
}

func (o *OpenAI) Image() *Image {

	return &Image{
		url:    "https://api.openai.com/v1/images/",
		apiKey: o.apiKey,
	}
}

func (o *Image) Generate(prompt string, n int, size string) (*ImageResponse, error) {
	url := "https://api.openai.com/v1/images/generations"

	req := ImageRequest{
		Prompt: prompt,
		N:      n,
		Size:   size,
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
	var imageResponse ImageResponse
	err = json.Unmarshal(resp.Body(), &imageResponse)
	if err != nil {
		return nil, err
	}
	return &imageResponse, nil
}

func (o *Image) Edit(imagePath string, maskPath string, prompt string, n int, size string) (*ImageResponse, error) {
	url := "https://api.openai.com/v1/images/edits"
	client := resty.New()

	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(imagePath)

	mask, err := os.Open(maskPath)
	if err != nil {
		return nil, err
	}
	defer mask.Close()
	maskName := filepath.Base(maskPath)

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetFileReader("image", fileName, file).
		SetFileReader("mask", maskName, mask).
		SetFormData(map[string]string{
			"prompt": prompt,
			"n":      strconv.Itoa(n),
			"size":   size,
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	var imageResponse ImageResponse
	err = json.Unmarshal(resp.Body(), &imageResponse)
	if err != nil {
		return nil, err
	}
	return &imageResponse, nil
}

func (o *Image) Variate(imagePath string, n int, size string) (*ImageResponse, error) {
	url := "https://api.openai.com/v1/images/variations"

	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(imagePath)

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetFileReader("image", fileName, file).
		SetFormData(map[string]string{
			"n":    strconv.Itoa(n),
			"size": size,
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	var imageResponse ImageResponse
	err = json.Unmarshal(resp.Body(), &imageResponse)
	if err != nil {
		return nil, err
	}
	return &imageResponse, nil
}

func (o *OpenAI) GetEmbeddings(input string, model string) (*EmbeddingResponse, error) {
	url := "https://api.openai.com/v1/embeddings"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetResult(&EmbeddingResponse{}).
		SetBody(EmbeddingRequest{
			Input: input,
			Model: model,
		}).
		Post(url)
	if err != nil {
		return nil, err
	}

	var response EmbeddingResponse
	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (o *OpenAI) Audio() *Audio {

	return &Audio{
		url:    "https://api.openai.com/v1/audio/",
		apiKey: o.apiKey,
	}
}

func (o *Audio) Transcriptions(filePath string) (*AudioResponse, error) {
	url := "https://api.openai.com/v1/audio/transcriptions"
	client := resty.New()

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(filePath)

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFileReader("file", fileName, file).
		SetFormData(map[string]string{
			"model": "whisper-1",
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	var audioResponse AudioResponse
	err = json.Unmarshal(resp.Body(), &audioResponse)
	if err != nil {
		return nil, err
	}
	return &audioResponse, nil
}

func (o *Audio) Translations(filePath string) (*AudioResponse, error) {
	url := "https://api.openai.com/v1/audio/translations"
	client := resty.New()
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(filePath)

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFileReader("file", fileName, file).
		SetFormData(map[string]string{
			"model": "whisper-1",
		}).
		Post(url)
	if err != nil {
		return nil, err
	}
	var audioResponse AudioResponse
	err = json.Unmarshal(resp.Body(), &audioResponse)
	if err != nil {
		return nil, err
	}
	return &audioResponse, nil
}

func (o *OpenAI) TuneFile() *TuneFile {
	return &TuneFile{
		url:    "https://api.openai.com/v1/models",
		apiKey: o.apiKey,
	}
}

func (o *TuneFile) List() (*FileList, error) {
	url := "https://api.openai.com/v1/files"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	var fileList FileList
	err = json.Unmarshal(resp.Body(), &fileList)
	if err != nil {
		return nil, err
	}
	return &fileList, nil
}

// New code starts here
func (o *TuneFile) Upload(filePath string) (*FileInfo, error) {
	url := "https://api.openai.com/v1/files"
	client := resty.New()
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(filePath)
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFormData(map[string]string{
			"purpose": "fine-tune",
		}).
		SetFileReader("file", fileName, file).
		Post(url)
	if err != nil {
		return nil, err
	}
	var fileInfo FileInfo
	err = json.Unmarshal(resp.Body(), &fileInfo)
	if err != nil {
		return nil, err
	}
	return &fileInfo, nil
}

// New code starts here
func (o *TuneFile) DeleteFile(fileID string) (*DeleteFileResponse, error) {
	url := "https://api.openai.com/v1/files/" + fileID
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Delete(url)
	if err != nil {
		return nil, err
	}
	var deleteFileResponse DeleteFileResponse
	err = json.Unmarshal(resp.Body(), &deleteFileResponse)
	if err != nil {
		return nil, err
	}
	return &deleteFileResponse, nil
}

func (o *TuneFile) Get(fileID string) (*FileInfo, error) {
	url := "https://api.openai.com/v1/files/" + fileID
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	var fileInfo FileInfo
	err = json.Unmarshal(resp.Body(), &fileInfo)
	if err != nil {
		return nil, err
	}
	return &fileInfo, nil
}

func (o *TuneFile) Content(fileID string, filePath string) error {
	url := "https://api.openai.com/v1/files/" + fileID + "/content"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, resp.Body(), 0644)
	if err != nil {
		return err
	}
	return nil
}

func (o *OpenAI) FineTune() *FineTune {
	return &FineTune{
		url:    "https://api.openai.com/v1/fine-tunes",
		apiKey: o.apiKey,
	}
}

func (o *FineTune) Create(fileID string) (*FineTuneJob, error) {
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
	var fineTuneJob FineTuneJob
	err = json.Unmarshal(resp.Body(), &fineTuneJob)
	if err != nil {
		return nil, err
	}
	return &fineTuneJob, nil
}

func (o *FineTune) List() (*FineTuneJobList, error) {
	url := "https://api.openai.com/v1/fine-tunes"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	var fineTuneJobList FineTuneJobList
	err = json.Unmarshal(resp.Body(), &fineTuneJobList)
	if err != nil {
		return nil, err
	}
	return &fineTuneJobList, nil
}

func (o *FineTune) Get(fine_tune_id string) (*FineTuneJob, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	var fineTuneJob FineTuneJob
	err = json.Unmarshal(resp.Body(), &fineTuneJob)
	if err != nil {
		return nil, err
	}
	return &fineTuneJob, nil
}

// New code starts here
func (o *FineTune) Cancel(fine_tune_id string) (*FineTuneJob, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id + "/cancel"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Post(url)
	if err != nil {
		return nil, err
	}
	var fineTuneJob FineTuneJob
	err = json.Unmarshal(resp.Body(), &fineTuneJob)
	if err != nil {
		return nil, err
	}
	return &fineTuneJob, nil
}

func (o *FineTune) Events(fine_tune_id string) (*FineTuneJobEventList, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id + "/events"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	var fineTuneJobEventList FineTuneJobEventList
	err = json.Unmarshal(resp.Body(), &fineTuneJobEventList)
	if err != nil {
		return nil, err
	}
	return &fineTuneJobEventList, nil
}

func (o *FineTune) Delete(fine_tune_id string) (*ModelDelete, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Delete(url)
	if err != nil {
		return nil, err
	}
	var modelDelete ModelDelete
	err = json.Unmarshal(resp.Body(), &modelDelete)
	if err != nil {
		return nil, err
	}
	return &modelDelete, nil
}

func (o *OpenAI) Moderation(input string) (*TextModerationResponse, error) {
	url := "https://api.openai.com/v1/moderations"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetResult(&TextModerationResponse{}).
		SetBody(TextModerationRequest{Input: input}).
		Post(url)
	if err != nil {
		return nil, err
	}
	var textModerationResponse TextModerationResponse
	err = json.Unmarshal(resp.Body(), &textModerationResponse)
	if err != nil {
		return nil, err
	}
	return &textModerationResponse, nil
}
