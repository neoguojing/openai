package openai

import (
	"encoding/json"
	"errors"
	"io"
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

func WithRole(role OpenAIRole) ChatOption {
	return func(c *Chat) {
		c.role = role
	}
}

type Image struct {
	apiKey   string
	url      string
	filePath string
}

type TuneFile struct {
	apiKey   string
	url      string
	filePath string
}

type FineTune struct {
	apiKey   string
	url      string
	filePath string
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

func (o *OpenAI) Completions(message string) (*CompletionResponse, error) {
	o.url = "https://api.openai.com/v1/completions"
	client := resty.New()
	req := CompletionRequest{
		Model:       "text-davinci-003",
		Prompt:      message,
		MaxTokens:   4097,
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

func (o *OpenAI) Image() *Image {

	return &Image{
		url:    "https://api.openai.com/v1/images/",
		apiKey: o.apiKey,
	}
}

func (o *Image) Generate(prompt string, n int) (*ImageResponse, error) {
	url := "https://api.openai.com/v1/images/generations"

	if n <= 0 {
		n = 1
	} else if n > 10 {
		n = 10
	}

	req := ImageRequest{
		Prompt: prompt,
		N:      n,
		Size:   Size1024,
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

func (o *Image) EditDirect(fileName string, input io.Reader, maskName string, mask io.Reader,
	prompt string, n int, size ImageSizeSupported) (*ImageResponse, error) {
	url := "https://api.openai.com/v1/images/edits"
	client := resty.New()

	req := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetFileReader("image", fileName, input)

	if mask != nil {
		req.SetFileReader("mask", maskName, mask)
	}

	if n <= 0 {
		n = 1
	} else if n > 10 {
		n = 10
	}

	resp, err := req.SetFormData(map[string]string{
		"prompt": prompt,
		"n":      strconv.Itoa(n),
		"size":   string(size),
	}).Post(url)
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

func (o *Image) Edit(imagePath string, maskPath string, prompt string, n int, size ImageSizeSupported) (*ImageResponse, error) {
	if imagePath == "" {
		return nil, errors.New("u need to upload a file")
	}
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(imagePath)

	var maskName string
	var mask *os.File
	if maskPath != "" {
		mask, err = os.Open(maskPath)
		if err != nil {
			return nil, err
		}
		defer mask.Close()
		maskName = filepath.Base(maskPath)
	}

	return o.EditDirect(fileName, file, maskName, mask, prompt, n, size)
}

func (o *Image) VariateDirect(fileName string, input io.Reader, n int, size ImageSizeSupported) (*ImageResponse, error) {
	url := "https://api.openai.com/v1/images/variations"

	if n <= 0 {
		n = 1
	} else if n > 10 {
		n = 10
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetFileReader("image", fileName, input).
		SetFormData(map[string]string{
			"n":    strconv.Itoa(n),
			"size": string(size),
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

func (o *Image) Variate(imagePath string, n int, size ImageSizeSupported) (*ImageResponse, error) {
	if imagePath == "" {
		return nil, errors.New("u need to upload a file")
	}

	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(imagePath)

	return o.VariateDirect(fileName, file, n, size)
}

func (o *OpenAI) GetEmbeddings(input string) (*EmbeddingResponse, error) {
	url := "https://api.openai.com/v1/embeddings"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetResult(&EmbeddingResponse{}).
		SetBody(EmbeddingRequest{
			Input: input,
			Model: "text-embedding-ada-002",
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

func (o *TuneFile) UploadDirect(fileName string, input io.Reader) (*FileInfo, error) {
	url := "https://api.openai.com/v1/files"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFormData(map[string]string{
			"purpose": "fine-tune",
		}).
		SetFileReader("file", fileName, input).
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
func (o *TuneFile) Upload(filePath string) (*FileInfo, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(filePath)
	return o.UploadDirect(fileName, file)
}

// New code starts here
func (o *TuneFile) Delete(fileID string) (*DeleteFileResponse, error) {
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

func (o *FineTune) Delete(fine_tune_id string) (*JobDeleteInfo, error) {
	url := "https://api.openai.com/v1/fine-tunes/" + fine_tune_id
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Delete(url)
	if err != nil {
		return nil, err
	}
	var modelDelete JobDeleteInfo
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
