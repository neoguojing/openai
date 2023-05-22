package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neoguojing/openai"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var api *openai.OpenAI

// @title OpenAI API
// @version 1.0
// @description This is a sample OpenAI API server.
// @host localhost:8080
// @BasePath /openai/api/v1
func GenerateGinRouter(apiKey string) *gin.Engine {
	router := gin.Default()
	api = openai.NewOpenAI(apiKey)
	openaiGroup := router.Group("/openai/api/v1")
	openaiGroup.POST("/files/upload", uploadFile)
	openaiGroup.DELETE("/files/:file_id", deleteFile)
	openaiGroup.GET("/files/:file_id", getFile)
	openaiGroup.POST("/fine-tunes/:file_id", createFineTuneJob)
	openaiGroup.GET("/fine-tunes", getFineTuneJobList)
	openaiGroup.GET("/fine-tunes/:fine_tune_id", getFineTuneJob)
	openaiGroup.GET("/fine-tunes/:fine_tune_id/events", getFineTuneJobEvents)
	openaiGroup.DELETE("/fine-tunes/:fine_tune_id", deleteFineTuneJob)
	openaiGroup.PUT("/fine-tunes/:fine_tune_id", cancelFineTuneJob)
	openaiGroup.POST("/audio/transcriptions", transcribeAudio)
	openaiGroup.POST("/audio/translations", translateAudio)
	openaiGroup.POST("/embeddings", getEmbeddings)
	openaiGroup.POST("/images/generate", generateImage)
	openaiGroup.POST("/images/edit", editImage)
	openaiGroup.POST("/images/variate", variateImage)
	openaiGroup.POST("/chat/complete", completeChat)
	openaiGroup.POST("/chat/edit", editChat)
	openaiGroup.GET("/model/list", listModels)
	openaiGroup.GET("/model/:name", getModel)
	openaiGroup.POST("/completions", completeText)
	openaiGroup.POST("/moderations/:text", moderation)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	return router
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(err error) *ErrorResponse {
	return &ErrorResponse{
		Error: err.Error(),
	}
}

// @Summary Upload a file
// @Description Upload a file to be fine-tuned
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to be uploaded"
// @Success 200 {object} openai.FileInfo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/files/upload [post]
func uploadFile(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filePath := "/tmp/" + file.Filename
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	var fileInfo *openai.FileInfo
	fileInfo, err = api.TuneFile().Upload(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, fileInfo)
}

// @Summary Get file info
// @Description Get information about a fine-tuned file
// @Accept json
// @Produce json
// @Param file_id path string true "File ID"
// @Success 200 {object} openai.FileInfo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/files/{file_id} [get]
func getFile(c *gin.Context) {

	fileID := c.Param("file_id")
	var err error
	var fileInfo *openai.FileInfo
	fileInfo, err = api.TuneFile().Get(fileID) // get file info using file ID
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, fileInfo)

}

// @Summary Delete a file
// @Description Delete a fine-tuned file
// @Accept json
// @Produce json
// @Param file_id path string true "File ID"
// @Success 200 {object} openai.DeleteFileResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/files/{file_id} [delete]
func deleteFile(c *gin.Context) {
	fileID := c.Param("file_id")
	var err error
	var fileInfo *openai.DeleteFileResponse
	fileInfo, err = api.TuneFile().Delete(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, fileInfo)
}

// @Summary Create a fine-tune job
// @Description Create a fine-tune job using a file ID
// @Accept json
// @Produce json
// @Param file_id path string true "File ID"
// @Success 200 {object} openai.FineTuneJob
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/fine-tunes/{file_id} [post]
func createFineTuneJob(c *gin.Context) {

	fileID := c.Param("file_id")
	var err error
	var fineTuneJob *openai.FineTuneJob
	fineTuneJob, err = api.FineTune().Create(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, fineTuneJob)

}

// @Summary Get fine-tune job list
// @Description Get a list of all fine-tune jobs
// @Accept json
// @Produce json
// @Success 200 {object} openai.FineTuneJobList
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/fine-tunes [get]
func getFineTuneJobList(c *gin.Context) {
	var err error
	var fineTuneJobList *openai.FineTuneJobList
	fineTuneJobList, err = api.FineTune().List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, fineTuneJobList)
}

// @Summary Get fine-tune job
// @Description Get information about a fine-tune job
// @Accept json
// @Produce json
// @Param fine_tune_id path string true "Fine-tune job ID"
// @Success 200 {object} openai.FineTuneJob
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/fine-tunes/{fine_tune_id} [get]
func getFineTuneJob(c *gin.Context) {

	fineTuneID := c.Param("fine_tune_id")
	var err error
	var fineTuneJob *openai.FineTuneJob
	fineTuneJob, err = api.FineTune().Get(fineTuneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, fineTuneJob)
}

// @Summary Get fine-tune job events
// @Description Get events for a fine-tune job
// @Accept json
// @Produce json
// @Param fine_tune_id path string true "Fine-tune job ID"
// @Success 200 {object} openai.FineTuneJobEventList
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/fine-tunes/{fine_tune_id}/events [get]
func getFineTuneJobEvents(c *gin.Context) {

	fineTuneID := c.Param("fine_tune_id")
	var err error
	var fineTuneJobEventList *openai.FineTuneJobEventList
	fineTuneJobEventList, err = api.FineTune().Events(fineTuneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, fineTuneJobEventList)

}

// @Summary Delete a fine-tune job
// @Description Delete a fine-tune job using a fine-tune job ID
// @Accept json
// @Produce json
// @Param fine_tune_id path string true "Fine-tune job ID"
// @Success 200 {object} openai.JobDeleteInfo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/fine-tunes/{fine_tune_id} [delete]
func deleteFineTuneJob(c *gin.Context) {

	fineTuneID := c.Param("fine_tune_id")
	var err error
	var response *openai.JobDeleteInfo
	response, err = api.FineTune().Delete(fineTuneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)

}

// @Summary Cancel a fine-tune job
// @Description Cancel a fine-tune job using a fine-tune job ID
// @Accept json
// @Produce json
// @Param fine_tune_id path string true "Fine-tune job ID"
// @Success 200 {object} openai.FineTuneJob
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/fine-tunes/{fine_tune_id}/cancel [post]
func cancelFineTuneJob(c *gin.Context) {

	fineTuneID := c.Param("fine_tune_id")
	var err error
	var response *openai.FineTuneJob
	response, err = api.FineTune().Cancel(fineTuneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)

}

// @Summary Transcribe audio file
// @Description Transcribe an audio file to text
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Audio file to transcribe"
// @Success 200 {object} openai.AudioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/audio/transcriptions [post]
func transcribeAudio(c *gin.Context) {

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
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	var response *openai.AudioResponse
	response, err = api.Audio().Transcriptions(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response)

}

// @Summary Translate audio file
// @Description Translate an audio file to text
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Audio file to translate"
// @Success 200 {object} openai.AudioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/audio/translations [post]
func translateAudio(c *gin.Context) {

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
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	var response *openai.AudioResponse
	response, err = api.Audio().Translations(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)

}

// GetEmbeddings godoc
// @Summary Get embeddings
// @Description Get embeddings for a given input
// @Accept json
// @Produce json
// @Param input body openai.EmbeddingRequest true "Input for which embeddings are to be generated"
// @Success 200 {object} openai.EmbeddingResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/embeddings [post]
func getEmbeddings(c *gin.Context) {

	var input openai.EmbeddingRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	var err error
	var response *openai.EmbeddingResponse
	response, err = api.GetEmbeddings(input.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)

}

// @Summary Generate an image
// @Description Generate an image using OpenAI's DALL-E API
// @Accept json
// @Produce json
// @Param model query string true "Model to use for image generation"
// @Param n query int true "Number of images to generate"
// @Param size query int true "Size of the image to generate"
// @Success 200 {object} openai.ImageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/images [get]
// @Tags Images
func generateImage(c *gin.Context) {

	var input openai.ImageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	var err error
	var response *openai.ImageResponse
	response, err = api.Image().Generate(input.Model, input.N, input.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)

}

// @Summary Edit an image using OpenAI's DALL-E API
// @Description Edit an image using OpenAI's DALL-E API
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image to edit"
// @Param prompt formData string false "Prompt for image editing"
// @Success 200 {object} openai.ImageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/images/edit [post]
// @Tags Images
func editImage(c *gin.Context) {

	image, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}
	filePath := "/tmp/" + image.Filename
	err = c.SaveUploadedFile(image, filePath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	prompt := c.PostForm("prompt")
	var response *openai.ImageResponse
	response, err = api.Image().Edit(filePath, "", prompt, 1, openai.Size1024)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)

}

// @Summary Generate image variations
// @Description Generate variations of an image using OpenAI's DALL-E API
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image to generate variations of"
// @Param n query int true "Number of variations to generate"
// @Param size query int true "Size of the variations to generate"
// @Success 200 {object} openai.ImageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/images/variations [post]
// @Tags Images
func variateImage(c *gin.Context) {

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
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	var response *openai.ImageResponse
	response, err = api.Image().Variate(filePath, 1, openai.Size1024)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Description 使用OpenAI的API完成聊天提示
// @Accept json
// @Produce json
// @Param input body openai.DialogRequest true "聊天提示的输入"
// @Success 200 {object} openai.ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/chat [post]
func completeChat(c *gin.Context) {

	var input openai.DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	var err error
	var response *openai.ChatResponse
	response, err = api.Chat().Complete(input.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Edit a chat prompt
// @Description Edit a chat prompt using OpenAI's API
// @Accept json
// @Produce json
// @Param input body openai.DialogRequest true "Input for chat prompt"
// @Param instruction body string true "Instruction for chat prompt editing"
// @Success 200 {object} openai.EditChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/chat/edit [post]
func editChat(c *gin.Context) {

	var input openai.DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var err error
	var response *openai.EditChatResponse
	response, err = api.Chat().Edits(input.Input, input.Instruction)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary List models
// @Description List all available models
// @Accept json
// @Produce json
// @Success 200 {object} openai.ModelList
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/models [get]
// @Tags Models
func listModels(c *gin.Context) {

	var err error
	var response *openai.ModelList
	response, err = api.Model().List()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Get a model
// @Description Get information about a specific OpenAI model
// @Accept json
// @Produce json
// @Param name path string true "Name of the model"
// @Success 200 {object} openai.ModelInfo
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/models/{name} [get]
// @Tags Models
func getModel(c *gin.Context) {

	name := c.Param("name")
	var err error
	var response *openai.ModelInfo
	response, err = api.Model().Get(name)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Complete a text prompt
// @Description Complete a text prompt using OpenAI's API
// @Accept json
// @Produce json
// @Param input body openai.DialogRequest true "Input for text prompt"
// @Param model query string true "Name of the model to use for completion"
// @Param temperature query float64 true "Sampling temperature to use for completion"
// @Param max_tokens query int true "Maximum number of tokens to generate for completion"
// @Param n query int true "Number of completions to generate"
// @Param stop query string true "Sequence to stop generation at"
// @Param presence query string true "Sequence to force into the generated text"
// @Param frequency_penalty query float64 true "Frequency penalty to use for completion"
// @Param presence_penalty query float64 true "Presence penalty to use for completion"
// @Success 200 {object} openai.CompletionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/completions [post]
func completeText(c *gin.Context) {

	var input openai.DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	var err error
	var response *openai.CompletionResponse
	response, err = api.Completions(input.Input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Moderation
// @Description Check if text contains inappropriate content using OpenAI's API
// @Accept json
// @Produce json
// @Param input body openai.DialogRequest true "Input for moderation"
// @Success 200 {object} openai.TextModerationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /openai/api/v1/moderations [post]
func moderation(c *gin.Context) {
	var input openai.DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}
	var err error
	var response *openai.TextModerationResponse
	response, err = api.Moderation(input.Input)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, response)
}
