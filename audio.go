package openai

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"go.beyondstorage.io/v5/services"
	"go.beyondstorage.io/v5/types"

	// Add fs support
	_ "go.beyondstorage.io/services/fs/v4"
	// Add s3 support
	_ "go.beyondstorage.io/services/s3/v3"
)

var (
	store types.Storager
)

func init() {
	var err error

	// Get current path
	currentPath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dst := filepath.Join(currentPath, "data")
	// Initialize storage
	store, err = services.NewStoragerFromString("fs://" + dst)
	if err != nil {
		log.Fatal(err)
	}

}

type Audio struct {
	apiKey   string
	url      string
	model    string
	filePath string
	text     string
	store    types.Storager
	client   *resty.Client
}

func (o *OpenAI) Audio() *Audio {

	return &Audio{
		url:    "https://api.openai.com/v1/audio/",
		apiKey: o.apiKey,
		model:  "whisper-1",
		store:  store,
		client: resty.New(),
	}
}

func (o *Audio) save(filePath string, input io.Reader) error {
	fileName := filepath.Base(filePath)
	_, err := o.store.Write(fileName, input, 10240)
	return err
}

func (o *Audio) TranscriptionsDirect(filePath string, input io.Reader) (*AudioResponse, error) {
	url := "https://api.openai.com/v1/audio/transcriptions"

	resp, err := o.client.R().
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
	o.text = audioResponse.Text
	o.save(filePath, input)
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

	resp, err := o.client.R().
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
	o.text = audioResponse.Text
	o.save(filePath, input)
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
