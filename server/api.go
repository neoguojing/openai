package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	midware "github.com/neoguojing/gin-midware"
	"github.com/neoguojing/openai"
	"github.com/neoguojing/openai/models"
	docs "github.com/neoguojing/openai/server/docs"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var (
	api  *openai.OpenAI
	chat *openai.Chat
)

// @title OpenAI API
// @version 1.0
// @description This is a sample OpenAI API server.
// @host localhost:8080
// @BasePath /openai/api/v1
func GenerateGinRouter(apiKey string) *gin.Engine {
	router := gin.Default()
	keyFunc := func(c *gin.Context) string {
		userAgent := c.Request.Header.Get("User-Agent")
		acceptLanguage := c.Request.Header.Get("Accept-Language")
		forwardedFor := c.Request.Header.Get("X-Forwarded-For")

		// 使用这些值生成用户唯一标识符
		return GenerateUserIdentifier(userAgent, acceptLanguage, forwardedFor)
	}
	router.Use(midware.GinRateLimiter(keyFunc, 10, 1*time.Second))
	docs.SwaggerInfo.BasePath = "/openai/api/v1"

	api = openai.NewOpenAI(apiKey)
	chat = api.Chat(openai.WithPlatform(models.HttpServer))
	openaiGroup := router.Group("/openai/api/v1")
	openaiGroup.POST("/files/upload", uploadFile)
	openaiGroup.DELETE("/files/:file_id", deleteFile)
	openaiGroup.GET("/files/:file_id", getFile)
	openaiGroup.GET("/files", listFiles)
	openaiGroup.POST("/fine-tunes/:file_id", createFineTuneJob)
	openaiGroup.GET("/fine-tunes", getFineTuneJobList)
	openaiGroup.GET("/fine-tunes/:fine_tune_id", getFineTuneJob)
	openaiGroup.GET("/fine-tunes/:fine_tune_id/events", getFineTuneJobEvents)
	openaiGroup.DELETE("/fine-tunes/:fine_tune_id", deleteFineTuneJob)
	openaiGroup.PUT("/fine-tunes/:fine_tune_id/cancel", cancelFineTuneJob)
	openaiGroup.POST("/audio/transcriptions", transcribeAudio)
	openaiGroup.POST("/audio/translations", translateAudio)
	openaiGroup.POST("/embeddings", getEmbeddings)
	openaiGroup.POST("/images/generate", generateImage)
	openaiGroup.POST("/images/edit", editImage)
	openaiGroup.POST("/images/variate", variateImage)
	openaiGroup.POST("/chat", completeChat)
	openaiGroup.POST("/chat/edit", editChat)
	openaiGroup.POST("/chat/voice", voiceChat)
	openaiGroup.PUT("/chat/:role", setRoleForChat)
	openaiGroup.GET("/models", listModels)
	openaiGroup.GET("/model/:name", getModel)
	openaiGroup.POST("/completions", completeText)
	openaiGroup.POST("/moderations", moderation)
	openaiGroup.POST("/aispeech", aispeechHandler)
	openaiGroup.POST("/officeaccount", officeAccountHandler)
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
// @Router /files/upload [post]
func uploadFile(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()
	var fileInfo *openai.FileInfo
	fileInfo, err = api.TuneFile().UploadDirect(file.Filename, reader)
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
// @Router /files/{file_id} [get]
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

// @Summary List file info
// @Description List information about the fine-tuned files
// @Accept json
// @Produce json
// @Success 200 {object} openai.FileList
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /files [get]
func listFiles(c *gin.Context) {

	var err error
	var fileInfo *openai.FileList
	fileInfo, err = api.TuneFile().List() // get file info using file ID
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
// @Router /files/{file_id} [delete]
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
// @Router /fine-tunes/{file_id} [post]
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
// @Router /fine-tunes [get]
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
// @Router /fine-tunes/{fine_tune_id} [get]
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
// @Router /fine-tunes/{fine_tune_id}/events [get]
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
// @Router /fine-tunes/{fine_tune_id} [delete]
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
// @Router /fine-tunes/{fine_tune_id}/cancel [post]
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
// @Router /audio/transcriptions [post]
func transcribeAudio(c *gin.Context) {

	// code for audio transcriptions
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()
	var response *openai.AudioResponse
	response, err = api.Audio().TranscriptionsDirect(file.Filename, reader)
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
// @Router /audio/translations [post]
func translateAudio(c *gin.Context) {

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()
	var response *openai.AudioResponse
	response, err = api.Audio().TranslationsDirect(file.Filename, reader)
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
// @Router /embeddings [post]
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
// @Param input body openai.ImageRequest true "Model to use for image generation"
// @Success 200 {object} openai.ImageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /images/generate [post]
// @Tags Images
func generateImage(c *gin.Context) {

	var input openai.ImageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	var err error
	var response *openai.ImageResponse

	response, err = api.Image().Generate(input.Prompt, input.N)
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
// @Router /images/edit [post]
// @Tags Images
func editImage(c *gin.Context) {

	image, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	reader, err := image.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	prompt := c.PostForm("prompt")
	var response *openai.ImageResponse
	response, err = api.Image().EditDirect(image.Filename, reader, "", nil, prompt, 1, openai.Size1024)
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
// @Success 200 {object} openai.ImageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /images/variate [post]
// @Tags Images
func variateImage(c *gin.Context) {

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}
	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	var response *openai.ImageResponse
	response, err = api.Image().VariateDirect(file.Filename, reader, 1, openai.Size1024)
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
// @Success 200 {object} openai.AudioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chat [post]
func completeChat(c *gin.Context) {

	var input openai.DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err))
		return
	}

	var err error
	var response = openai.AudioResponse{}
	text, err := chat.Dialogue(models.Text, input.Input, "", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	response.Text = text
	c.JSON(http.StatusOK, response)
}

// @Description 使用语音进行对话
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Audio file to transcribe"
// @Success 200 {object} openai.AudioResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chat/voice [post]
func voiceChat(c *gin.Context) {
	// code for audio transcriptions
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	reader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defer reader.Close()
	var response = openai.AudioResponse{}
	text, err := chat.Dialogue(models.Voice, "", file.Filename, reader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}
	response.Text = text
	c.JSON(http.StatusOK, response)
}

// @Description 设置AI角色
// @Accept json
// @Produce json
// @Param role path string true "role name"
// @Success 200 {object} openai.ChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chat/{role} [post]
func setRoleForChat(c *gin.Context) {

	role := c.Param("role")
	roles, err := models.SearchRoleByName(role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
		return
	}

	var roleDesc string
	var response *openai.ChatResponse
	if len(roles) > 0 {
		roleDesc = roles[0].Desc
		response, err = chat.Complete(roleDesc)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewErrorResponse(err))
			return
		}
	}
	c.JSON(http.StatusOK, response)
}

// @Summary Edit a chat prompt
// @Description Edit a chat prompt using OpenAI's API
// @Accept json
// @Produce json
// @Param input body openai.DialogRequest true "Input for chat prompt"
// @Success 200 {object} openai.EditChatResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /chat/edit [post]
func editChat(c *gin.Context) {

	var input openai.DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}

	var err error
	var response *openai.EditChatResponse
	response, err = chat.Edits(input.Input, input.Instruction)
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
// @Router /models [get]
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
// @Router /model/{name} [get]
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
// @Success 200 {object} openai.CompletionResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /completions [post]
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
// @Router /moderations [post]
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
