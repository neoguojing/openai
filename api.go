package openai

import (
	"net/http"

	"github.com/gin-gonic/gin"
	// ginSwagger "github.com/swaggo/gin-swagger"
	// _ "github.com/neoguojing/openai/docs"
)

var api *OpenAI

func GenerateGinRouter(apiKey string) *gin.Engine {
	router := gin.Default()

	// router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api = NewOpenAI(apiKey)
	openaiGroup := router.Group("/openai/api/v1")
	openaiGroup.POST("/files/upload", uploadFile)
	openaiGroup.DELETE("/files/:file_id", deleteFile)
	openaiGroup.GET("/files/:file_id", getFile)
	openaiGroup.POST("/fine-tunes/:file_id", createFineTuneJob)
	openaiGroup.GET("/fine-tunes", getFineTuneJobList)
	openaiGroup.GET("/fine-tunes/:fine_tune_id", getFineTuneJob)
	openaiGroup.GET("/fine-tunes/:fine_tune_id/events", getFineTuneJobEvents)
	openaiGroup.DELETE("/fine-tunes/:fine_tune_id", deleteFineTuneJob)
	openaiGroup.POST("/audio/transcriptions", transcribeAudio)
	openaiGroup.POST("/audio/translations", translateAudio)
	openaiGroup.POST("/embeddings", getEmbeddings)
	openaiGroup.POST("/images/generate", generateImage)
	openaiGroup.POST("/images/edit", editImage)
	openaiGroup.POST("/images/variate", variateImage)
	openaiGroup.POST("/chat/complete", completeChat)
	openaiGroup.POST("/chat/edit", editChat)

	return router
}

func uploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filePath := "/tmp/" + file.Filename
	err = c.SaveUploadedFile(file, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fileInfo, err := api.TuneFile().Upload(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fileInfo)
}

func getFile(c *gin.Context) {

	fileID := c.Param("file_id")
	fileInfo, err := api.TuneFile().Get(fileID) // get file info using file ID
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, fileInfo)

}

func deleteFile(c *gin.Context) {

	fileID := c.Param("file_id")
	fileInfo, err := api.TuneFile().Delete(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fileInfo)
}

func createFineTuneJob(c *gin.Context) {
	fileID := c.Param("file_id")
	fineTuneJob, err := api.FineTune().Create(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, fineTuneJob)

}

func getFineTuneJobList(c *gin.Context) {
	fineTuneJobList, err := api.FineTune().List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fineTuneJobList)
}

func getFineTuneJob(c *gin.Context) {
	fineTuneID := c.Param("fine_tune_id")
	fineTuneJob, err := api.FineTune().Get(fineTuneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fineTuneJob)
}

func getFineTuneJobEvents(c *gin.Context) {
	fineTuneID := c.Param("fine_tune_id")
	fineTuneJobEventList, err := api.FineTune().Events(fineTuneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, fineTuneJobEventList)

}

func deleteFineTuneJob(c *gin.Context) {
	fineTuneID := c.Param("fine_tune_id")
	modelDelete, err := api.FineTune().Delete(fineTuneID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, modelDelete)

}

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

}

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

}

func getEmbeddings(c *gin.Context) {
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

}

func generateImage(c *gin.Context) {

	var input ImageRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageInfo, err := api.Image().Generate(input.Model, input.N, input.Size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, imageInfo)

}

func editImage(c *gin.Context) {

	image, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filePath := "/tmp/" + image.Filename
	err = c.SaveUploadedFile(image, filePath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	prompt := c.PostForm("prompt")
	edit, err := api.Image().Edit(filePath, "", prompt, 1, Size1024)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, edit)

}

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	variation, err := api.Image().Variate(filePath, 1, Size1024)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, variation)
}

func completeChat(c *gin.Context) {
	var input DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	response, err := api.Chat().Complete(input.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, response)
}

func editChat(c *gin.Context) {
	var input DialogRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
		return
	}
	response, err := api.Chat().Edits(input.Input, input.Instruction)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, response)
}
