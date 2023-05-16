package openai

import (
	"reflect"
	"testing"
)

func TestModelList(t *testing.T) {
	openai := NewOpenAI("your_api_key_here")
	modelList, err := openai.Model().List()
	if err != nil {
		t.Errorf("Error retrieving model list: %v", err)
	}
	if modelList == nil || len(modelList.Data) == 0 {
		t.Errorf("Expected non-nil model list with length > 0, but got %v", modelList)
	}
}

func TestModelGet(t *testing.T) {
	// Create a new Model instance with a mock API key and URL
	model := &Model{
		apiKey: "mock-api-key",
		url:    "https://api.openai.com/v1/models",
	}
	// Call the Get method with a mock model name
	modelInfo, err := model.Get("mock-model-name")

	// Check that the error is nil
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	// Check that the returned ModelInfo struct has the expected values
	expectedModelInfo := &ModelInfo{}
	if !reflect.DeepEqual(modelInfo, expectedModelInfo) {
		t.Errorf("unexpected ModelInfo: got %v, want %v", modelInfo, expectedModelInfo)
		return
	}
}

func TestCompletions(t *testing.T) {
	apiKey := "your-api-key"
	o := NewOpenAI(apiKey)
	message := "test message"
	maxTokens := 10
	completionResponse, err := o.Completions(message, maxTokens)
	if err != nil {
		t.Errorf("Completions() error = %v", err)
		return
	}
	if completionResponse == nil {
		t.Errorf("Completions() completionResponse is nil")
		return
	}
	// Add more assertions here to test the output of the function
}
