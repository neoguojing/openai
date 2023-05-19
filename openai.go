package openai

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
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

func (o *Chat) Complete(content string) (*ChatResponse, error) {
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

func (o *OpenAI) GenerateGinRouter(apiKey string) *gin.Engine {
	router := gin.Default()
	api := NewOpenAI(apiKey)
	router.Group("/openai/api/v1")

	router.POST("/files/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		filePath := "/tmp/" + file.Filename
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		fileInfo, err := api.TuneFile().Upload(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fileInfo)

	})

	router.DELETE("/files/:file_id", func(c *gin.Context) {

		fileID := c.Param("file_id")
		fileInfo, err := api.TuneFile().Delete(fileID) // get file info using file ID
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fileInfo)
	})

	router.POST("/fine-tunes/:file_id", func(c *gin.Context) {
		fileID := c.Param("file_id")
		fineTuneJob, err := api.FineTune().Create(fileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fineTuneJob)

	})
	router.GET("/files/:file_id", func(c *gin.Context) {

		fileID := c.Param("file_id")
		fileInfo, err := api.TuneFile().Get(fileID) // get file info using file ID
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fileInfo)

	})

	router.POST("/fine-tunes", func(c *gin.Context) {
		fileID := c.PostForm("file_id")
		fineTuneJob, err := api.FineTune().Create(fileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fineTuneJob)

	})
	router.GET("/fine-tunes", func(c *gin.Context) {
		fineTuneJobList, err := api.FineTune().List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fineTuneJobList)

	})
	router.GET("/fine-tunes/:fine_tune_id", func(c *gin.Context) {
		fineTuneID := c.Param("fine_tune_id")
		fineTuneJob, err := api.FineTune().Get(fineTuneID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fineTuneJob)

	})
	router.GET("/fine-tunes/:fine_tune_id/events", func(c *gin.Context) {
		fineTuneID := c.Param("fine_tune_id")
		fineTuneJobEventList, err := api.FineTune().Events(fineTuneID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, fineTuneJobEventList)

	})
	router.DELETE("/fine-tunes/:fine_tune_id", func(c *gin.Context) {
		fineTuneID := c.Param("fine_tune_id")
		modelDelete, err := api.FineTune().Delete(fineTuneID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, modelDelete)

	})
	router.POST("/audio/transcriptions", func(c *gin.Context) {

		// code for audio transcriptions
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		filePath := "/tmp/" + file.Filename
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		transcription, err := api.Audio().Transcriptions(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, transcription)

	})
	router.POST("/audio/translations", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		filePath := "/tmp/" + file.Filename
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		translation, err := api.Audio().Translations(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, translation)

	})
	router.POST("/embeddings", func(c *gin.Context) {
		var input EmbeddingRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		embedding, err := api.GetEmbeddings(input.Input, input.Model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, embedding)

	})

	router.POST("/images/generate", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		filePath := "/tmp/" + file.Filename
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		imageInfo, err := api.Image().Generate(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, imageInfo)

	})
	router.POST("/images/edit", func(c *gin.Context) {

		// code for image editing
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		filePath := "/tmp/" + file.Filename
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		edit, err := api.Image().Edit(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, edit)

	})
	router.POST("/images/variate", func(c *gin.Context) {

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		filePath := "/tmp/" + file.Filename
		err = c.SaveUploadedFile(file, filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		variation, err := api.Image().Variate(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, variation)

	})

	router.POST("/chat/complete", func(c *gin.Context) {

		var input ChatInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := api.Chat().Complete("")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, response)

	})
	router.POST("/chat/edit", func(c *gin.Context) {

	})
	return router
}
