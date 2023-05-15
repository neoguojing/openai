package openai

import (
	"strconv"

	"github.com/go-resty/resty/v2"
)

type OpenAI struct {
	apiKey string
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{apiKey: apiKey}
}

func (o *OpenAI) GetModels() ([]byte, error) {
	url := "https://api.openai.com/v1/models"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("OpenAI-Organization", "org-U3jJBNZ72nnwuS5qRKQOVhcS").
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) Chat(message string) ([]byte, error) {
	url := "https://api.openai.com/v1/chat/completions"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(`{
            "model": "gpt-3.5-turbo",
            "messages": [{"role": "user", "content": "` + message + `"}],
            "temperature": 0.7
        }`).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetTextDavinci003() ([]byte, error) {
	url := "https://api.openai.com/v1/models/text-davinci-003"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetCompletions(prompt string, maxTokens int, temperature float64) ([]byte, error) {
	url := "https://api.openai.com/v1/completions"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(`{
			"model": "text-davinci-003",
			"prompt": "` + prompt + `",
			"max_tokens": ` + strconv.Itoa(maxTokens) + `,
			"temperature": ` + strconv.FormatFloat(temperature, 'f', 1, 64) + `
		}`).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
}

func (o *OpenAI) GetEdits(prompt string, instruction string) ([]byte, error) {
	url := "https://api.openai.com/v1/edits"
	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetBody(`{
			"model": "text-davinci-edit-001",
			"input": "` + prompt + `",
			"instruction": "` + instruction + `"
		}`).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Body(), nil
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
