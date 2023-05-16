
package openai

import (
	"testing"
)

var openai = NewOpenAI("xxxx")

func TestModelList(t *testing.T) {
	modelList, err := openai.Model().List()
	if err != nil {
		t.Errorf("Error retrieving model list: %v", err)
		return
	}
	if modelList == nil || len(modelList.Data) == 0 {
		t.Errorf("Expected non-nil model list with length > 0, but got %v", modelList)
		return
	}
}

func TestModelGet(t *testing.T) {
	modelInfo, err := openai.Model().Get("gpt-3.5-turbo")
	if err != nil {
		t.Errorf("Unexpected error occurred while retrieving model information: %v", err)
		return
	}
	t.Log(modelInfo)
}

func TestCompletions(t *testing.T) {
	message := "what is the AIGC"
	maxTokens := 100
	completionResponse, err := openai.Completions(message, maxTokens)
	if err != nil {
		t.Errorf("An error occurred while generating completions: %v", err)
		return
	}
	if completionResponse == nil {
		t.Errorf("Completion response is nil")
		return
	}
	t.Log(completionResponse)
}

func TestChatCompletions(t *testing.T) {
	message := "what is the AIGC"
	resp, err := openai.Chat().Completions(message)
	if err != nil {
		t.Errorf("An error occurred while generating chat completions: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Chat completion response is nil")
		return
	}
	t.Log(resp)
}

func TestChatEdits(t *testing.T) {
	message := "what is math"
	instrut := "use chinese"
	resp, err := openai.Chat().Edits(message, instrut)
	if err != nil {
		t.Errorf("An error occurred while generating chat edits: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Chat edit response is nil")
		return
	}
	t.Log(resp)
}

func TestImageGenerate(t *testing.T) {
	message := "A cute baby with swing"
	resp, err := openai.Image().Generate(message, 2, "1024x1024")
	if err != nil {
		t.Errorf("An error occurred while generating image: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Image generation response is nil")
		return
	}
	t.Log(resp)
}

func TestImageVariate(t *testing.T) {
	resp, err := openai.Image().Variate("./771ae33922b07e8ee52c059db27243b1.jpeg", 2, "1024x1024")
	if err != nil {
		t.Errorf("An error occurred while generating image variate: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Image variate response is nil")
		return
	}
	t.Log(resp)
}

func TestGetEmbeddings(t *testing.T) {
	resp, err := openai.GetEmbeddings("The food was delicious and the waiter...", "text-embedding-ada-002")
	if err != nil {
		t.Errorf("An error occurred while retrieving embeddings: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Embeddings response is nil")
		return
	}
	t.Log(resp)
}

func TestGetTuneFileList(t *testing.T) {
	resp, err := openai.TuneFile().List()
	if err != nil {
		t.Errorf("An error occurred while retrieving tune file list: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Tune file list response is nil")
		return
	}
	t.Log(resp)
}

func TestFineTuneList(t *testing.T) {
	resp, err := openai.FineTune().List()
	if err != nil {
		t.Errorf("An error occurred while retrieving fine tune list: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Fine tune list response is nil")
		return
	}
	t.Log(resp)
}

func TestModeration(t *testing.T) {
	resp, err := openai.Moderation("I want to kill them.")
	if err != nil {
		t.Errorf("An error occurred while moderating: %v", err)
		return
	}
	if resp == nil {
		t.Errorf("Moderation response is nil")
		return
	}
	t.Log(resp)
}

