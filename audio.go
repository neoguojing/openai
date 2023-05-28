package openai

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
)

type Audio struct {
	apiKey   string
	url      string
	model    string
	filePath string
	text     string
}

func (o *OpenAI) Audio() *Audio {

	return &Audio{
		url:    "https://api.openai.com/v1/audio/",
		apiKey: o.apiKey,
		model:  "whisper-1",
	}
}

func (o *Audio) TranscriptionsDirect(filePath string, input io.Reader) (*AudioResponse, error) {
	if input == nil {
		return nil, errors.New("empty input")
	}

	url := "https://api.openai.com/v1/audio/transcriptions"
	client := resty.New()

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFileReader("file", filePath, input).
		SetFormData(map[string]string{
			"model": o.model,
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
	o.filePath = filePath
	o.text = audioResponse.Text

	return &audioResponse, nil
}

func (o *Audio) Transcriptions(filePath string) (*AudioResponse, error) {

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(filePath)

	return o.TranscriptionsDirect(fileName, file)
}

func (o *Audio) TranslationsDirect(filePath string, input io.Reader) (*AudioResponse, error) {
	url := "https://api.openai.com/v1/audio/translations"
	client := resty.New()

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+o.apiKey).
		SetHeader("Content-Type", "multipart/form-data").
		SetFileReader("file", filePath, input).
		SetFormData(map[string]string{
			"model": o.model,
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
	o.filePath = filePath
	o.text = audioResponse.Text
	return &audioResponse, nil
}

func (o *Audio) Translations(filePath string) (*AudioResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	fileName := filepath.Base(filePath)

	return o.TranslationsDirect(fileName, file)
}
