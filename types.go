package openai

import (
	"errors"
	"fmt"
)

// OpenAIRole 是 OpenAI 的角色类型
type OpenAIRole string

const (
	User      OpenAIRole = "user"
	System    OpenAIRole = "sysmtem"
	Assistant OpenAIRole = "assistant"
)

type ImageSizeSupported string

const (
	Size256  ImageSizeSupported = "256x256"
	Size512  ImageSizeSupported = "512x512"
	Size1024 ImageSizeSupported = "1024x1024"
)

// ModelInfo 是模型信息
type ModelInfo struct {
	// 模型 ID
	ID string `json:"id"`
	// 模型对象
	Object string `json:"object"`
	// 模型所属者
	OwnedBy string `json:"owned_by"`
	// 模型权限
	Permission []ModelPermission `json:"permission"`
}

// ModelPermission 是模型权限
type ModelPermission struct {
	// 模型 ID
	ID string `json:"id"`
	// 是否允许创建引擎
	AllowCreateEngine bool `json:"allow_create_engine"`
	// 是否允许采样
	AllowSampling bool `json:"allow_sampling"`
	// 是否允许记录概率
	AllowLogprobs bool `json:"allow_logprobs"`
	// 是否允许搜索索引
	AllowSearchIndices bool `json:"allow_search_indices"`
	// 是否允许查看
	AllowView bool `json:"allow_view"`
	// 是否允许微调
	AllowFineTuning bool `json:"allow_fine_tuning"`
	// 组织
	Organization string `json:"organization"`
	// 组
	Group interface{} `json:"group"`
	// 是否阻塞
	IsBlocking bool `json:"is_blocking"`
}

// CompletionRequest represents a request to generate text completion.
type CompletionRequest struct {
	// Model is the ID of the model to use for text completion.
	Model string `json:"model"`
	// Prompt is the text prompt to use for text completion.
	Prompt string `json:"prompt"`
	// MaxTokens is the maximum number of tokens to generate in the completion.
	MaxTokens int `json:"max_tokens"`
	// Temperature is the sampling temperature to use for text completion.
	Temperature float64 `json:"temperature"`
	// TopP is the top-p sampling cutoff to use for text completion.
	TopP float64 `json:"top_p"`
	// N is the number of completions to generate.
	N int `json:"n"`
	// Stream specifies whether to stream the response or wait for the entire response.
	Stream bool `json:"stream"`
	// Logprobs specifies the number of log probabilities to generate.
	Logprobs int `json:"logprobs"`
	// Stop is the stop sequence to use for text completion.
	Stop string `json:"stop"`
}

// CompletionResponse represents a response to generate text completion.
type CompletionResponse struct {
	// Choices is an array of choices for text completion.
	Choices []struct {
		// Text is the generated text for the choice.
		Text string `json:"text"`
		// Index is the index of the choice.
		Index int `json:"index"`
		// Logprobs is the log probabilities for the choice.
		Logprobs interface{} `json:"logprobs"`
		// FinishReason is the reason for finishing the choice.
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	// Created is the timestamp for when the response was created.
	Created int `json:"created"`
	// ID is the ID of the response.
	ID string `json:"id"`
	// Model is the ID of the model used for text completion.
	Model string `json:"model"`
	// Object is the type of object for the response.
	Object string `json:"object"`
	// Usage is the usage statistics for the response.
	Usage struct {
		// PromptTokens is the number of tokens in the prompt.
		PromptTokens int `json:"prompt_tokens"`
		// CompletionTokens is the number of tokens in the completion.
		CompletionTokens int `json:"completion_tokens"`
		// TotalTokens is the total number of tokens.
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// ChatRequest represents a request to generate a chat response.
type ChatRequest struct {
	// Model is the ID of the model to use for generating the chat response.
	Model string `json:"model"`
	// Messages is an array of messages in the chat.
	Messages []struct {
		// Role is the role of the message sender (system, user, or assistant).
		Role string `json:"role"`
		// Content is the content of the message.
		Content string `json:"content"`
	} `json:"messages"`
}

type ChatResponseOption func(*ChatResponse)

func WithContentLengthLimit(length int) ChatResponseOption {
	return func(resp *ChatResponse) {
		if len(resp.Choices) > 0 {
			if len(resp.Choices[0].Message.Content) > length {
				trimContent := resp.Choices[0].Message.Content[:length]
				resp.Choices[0].Message.Content = fmt.Sprintf("%s,[content longger than %v] ", trimContent, length)
			}
		}
	}
}

// ChatResponse represents a response to generate a chat response.
type ChatResponse struct {
	// ID is the ID of the response.
	ID string `json:"id"`
	// Object is the type of object for the response.
	Object string `json:"object"`
	// Created is the timestamp for when the response was created.
	Created int `json:"created"`
	// Choices is an array of choices for text completion.
	Choices []struct {
		// Index is the index of the choice.
		Index int `json:"index"`
		// Message is the message object for the choice.
		Message struct {
			// Role is the role of the message sender (system, user, or assistant).
			Role string `json:"role"`
			// Content is the content of the message.
			Content string `json:"content"`
		} `json:"message"`
		// FinishReason is the reason for finishing the choice.
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	// Usage is the usage statistics for the response.
	Usage struct {
		// PromptTokens is the number of tokens in the prompt.
		PromptTokens int `json:"prompt_tokens"`
		// CompletionTokens is the number of tokens in the completion.
		CompletionTokens int `json:"completion_tokens"`
		// TotalTokens is the total number of tokens.
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// CheckChatResponse checks if the chat response is valid.
func (r *ChatResponse) GetContent(options ...ChatResponseOption) (string, error) {
	if r == nil {
		return "", errors.New("response is nil")
	}

	if len(r.Choices) == 0 {
		return "", errors.New("response choices is empty")
	}
	content := r.Choices[0].Message.Content
	if content == "" {
		return "", errors.New("response choice message content is empty")
	}

	return content, nil
}

// EditChatRequest represents a request to edit a chat response.
type EditChatRequest struct {
	// Model is the ID of the model to use for generating the chat response.
	Model string `json:"model"`
	// Messages is an array of messages in the chat.
	Messages []struct {
		// Role is the role of the message sender (system, user, or assistant).
		Role string `json:"role"`
		// Content is the content of the message.
		Content string `json:"content"`
	} `json:"messages"`
	// Instruction is the instruction for editing the chat response.
	Instruction string `json:"instruction"`
}

// EditChatResponse represents a response to edit a chat response.
type EditChatResponse struct {
	// Choices is an array of choices for text completion.
	Choices []struct {
		// Text is the generated text for the choice.
		Text string `json:"text"`
		// Index is the index of the choice.
		Index int `json:"index"`
		// Logprobs is the log probabilities for the choice.
		Logprobs interface{} `json:"logprobs"`
		// FinishReason is the reason for finishing the choice.
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	// Created is the timestamp for when the response was created.
	Created int `json:"created"`
	// ID is the ID of the response.
	ID string `json:"id"`
	// Model is the ID of the model used for text completion.
	Model string `json:"model"`
	// Object is the type of object for the response.
	Object string `json:"object"`
	// Usage is the usage statistics for the response.
	Usage struct {
		// PromptTokens is the number of tokens in the prompt.
		PromptTokens int `json:"prompt_tokens"`
		// CompletionTokens is the number of tokens in the completion.
		CompletionTokens int `json:"completion_tokens"`
		// TotalTokens is the total number of tokens.
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// CheckChatResponse checks if the chat response is valid.
func (r *EditChatResponse) GetContent() (string, error) {
	if r == nil {
		return "", errors.New("response is nil")
	}

	if len(r.Choices) == 0 {
		return "", errors.New("response choices is empty")
	}
	content := r.Choices[0].Text
	if content == "" {
		return "", errors.New("response choice message content is empty")
	}

	return content, nil
}

// ModelList represents a list of models.
type ModelList struct {
	// Data is an array of model information.
	Data []ModelInfo `json:"data"`
	// Object is the type of object for the response.
	Object string `json:"object"`
}

// ImageRequest represents a request to generate an image.
type ImageRequest struct {
	// Model is the ID of the model to use for generating the image.
	Model string `json:"model"`
	// Prompt is the prompt to use for generating the image.
	Prompt string `json:"prompt"`
	// Size is the size of the image to generate.
	Size ImageSizeSupported `json:"size"`
	// N is the number of images to generate.
	N int `json:"n"`
	// ResponseFormat is the format of the response.
	ResponseFormat string `json:"response_format"`
}

// ImageResponse represents a response to generate an image.
type ImageResponse struct {
	// Created is the timestamp for when the response was created.
	Created int `json:"created"`
	// Data is an array of image URLs.
	Data []struct {
		// URL is the URL of the generated image.
		URL string `json:"url"`
	} `json:"data"`
}

// EmbeddingRequest represents a request to generate an embedding.
type EmbeddingRequest struct {
	// Model is the ID of the model to use for generating the embedding.
	Model string `json:"model"`
	// Input is the input text to generate the embedding for.
	Input string `json:"input"`
}

// EmbeddingResponse represents a response to generate an embedding.
type EmbeddingResponse struct {
	// Model is the ID of the model used for generating the embedding.
	Model string `json:"model"`
	// Object is the type of object for the response.
	Object string `json:"object"`
	// Data is an array of embedding information.
	Data []struct {
		// Object is the type of object for the response.
		Object string `json:"object"`
		// Embedding is the embedding generated for the input text.
		Embedding []float64 `json:"embedding"`
		// Index is the index of the input text.
		Index int `json:"index"`
	} `json:"data"`
	// Usage is the usage statistics for the response.
	Usage struct {
		// PromptTokens is the number of tokens in the prompt.
		PromptTokens int `json:"prompt_tokens"`
		// TotalTokens is the total number of tokens.
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// AudioResponse represents a response to generate audio.
type AudioResponse struct {
	// Text is the text used to generate the audio.
	Text string `json:"text"`
}

// FileList represents a list of files.
type FileList struct {
	// Data is an array of file information.
	Data []FileInfo `json:"data"`
	// Object is the type of object for the response.
	Object string `json:"object"`
}

// DeleteFileResponse represents a response to delete a file.
type DeleteFileResponse struct {
	// ID is the ID of the deleted file.
	ID string `json:"id"`
	// Object is the type of object for the response.
	Object string `json:"object"`
	// Deleted is a boolean indicating whether the file was successfully deleted.
	Deleted bool `json:"deleted"`
}

// FileInfo 文件信息
type FileInfo struct {
	// ID 是文件的唯一标识符。
	ID string `json:"id"`
	// Object 是响应的对象类型。
	Object string `json:"object"`
	// Bytes 是文件的大小（以字节为单位）。
	Bytes int `json:"bytes"`
	// CreatedAt 是文件创建的时间戳。
	CreatedAt int `json:"created_at"`
	// Filename 是文件的名称。
	Filename string `json:"filename"`
	// Purpose 是文件的用途。
	Purpose string `json:"purpose"`
}

// FineTuneJob represents a job for fine-tuning a model.
type FineTuneJob struct {
	// ID is the ID of the fine-tune job.
	ID string `json:"id"`
	// Object is the type of object for the response.
	Object string `json:"object"`
	// Model is the ID of the model being fine-tuned.
	Model string `json:"model"`
	// CreatedAt is the timestamp for when the fine-tune job was created.
	CreatedAt int `json:"created_at"`
	// Events is an array of events for the fine-tune job.
	Events []FineTuneEvent `json:"events"`
	// FineTunedModel is the ID of the fine-tuned model.
	FineTunedModel string `json:"fine_tuned_model"`
	// Hyperparams is the hyperparameters for the fine-tune job.
	Hyperparams FineTuneJobHyperparams `json:"hyperparams"`
	// OrganizationID is the ID of the organization that owns the fine-tune job.
	OrganizationID string `json:"organization_id"`
	// ResultFiles is an array of files generated by the fine-tune job.
	ResultFiles []FileInfo `json:"result_files"`
	// Status is the status of the fine-tune job.
	Status string `json:"status"`
	// ValidationFiles is an array of validation files for the fine-tune job.
	ValidationFiles []FileInfo `json:"validation_files"`
	// TrainingFiles is an array of training files for the fine-tune job.
	TrainingFiles []FileInfo `json:"training_files"`
	// UpdatedAt is the timestamp for when the fine-tune job was last updated.
	UpdatedAt int `json:"updated_at"`
}

// FineTuneEvent represents an event for a fine-tune job.
type FineTuneEvent struct {
	// Object is the type of object for the response.
	Object string `json:"object"`
	// CreatedAt is the timestamp for when the event was created.
	CreatedAt int `json:"created_at"`
	// Level is the level of the event.
	Level string `json:"level"`
	// Message is the message for the event.
	Message string `json:"message"`
}

// FineTuneJobList represents a list of fine-tune jobs.
type FineTuneJobList struct {
	// Data is an array of fine-tune job information.
	Data []FineTuneJob `json:"data"`
	// Object is the type of object for the response.
	Object string `json:"object"`
}

// FineTuneJobEventList represents a list of events for a fine-tune job.
type FineTuneJobEventList struct {
	// Data is an array of fine-tune job event information.
	Data []FineTuneEvent `json:"data"`
	// Object is the type of object for the response.
	Object string `json:"object"`
}

// FineTuneJobHyperparams represents the hyperparameters for a fine-tune job.
type FineTuneJobHyperparams struct {
	// BatchSize is the batch size for the fine-tune job.
	BatchSize int `json:"batch_size"`
	// LearningRateMult is the learning rate multiplier for the fine-tune job.
	LearningRateMult float64 `json:"learning_rate_multiplier"`
	// NEpochs is the number of epochs for the fine-tune job.
	NEpochs int `json:"n_epochs"`
	// PromptLossWeight is the prompt loss weight for the fine-tune job.
	PromptLossWeight float64 `json:"prompt_loss_weight"`
}

// JobDeleteInfo represents a response to delete a model.
type JobDeleteInfo struct {
	// ID is the ID of the deleted model.
	ID string `json:"id"`
	// Object is the type of object for the response.
	Object string `json:"object"`
	// Deleted is a boolean indicating whether the model was successfully deleted.
	Deleted bool `json:"deleted"`
}

// TextModerationResponse represents a response to a text moderation request.
type TextModerationResponse struct {
	// ID is the ID of the text moderation request.
	ID string `json:"id"`
	// Model is the ID of the model used for text moderation.
	Model string `json:"model"`
	// Results is an array of text moderation results.
	Results []struct {
		// Categories is a struct containing boolean values for different categories of text moderation.
		Categories struct {
			// Hate is a boolean indicating whether the text contains hate speech.
			Hate bool `json:"hate"`
			// HateThreatening is a boolean indicating whether the text contains threatening hate speech.
			HateThreatening bool `json:"hate/threatening"`
			// SelfHarm is a boolean indicating whether the text contains self-harm content.
			SelfHarm bool `json:"self-harm"`
			// Sexual is a boolean indicating whether the text contains sexual content.
			Sexual bool `json:"sexual"`
			// SexualMinors is a boolean indicating whether the text contains sexual content involving minors.
			SexualMinors bool `json:"sexual/minors"`
			// Violence is a boolean indicating whether the text contains violent content.
			Violence bool `json:"violence"`
			// ViolenceGraphic is a boolean indicating whether the text contains graphic violent content.
			ViolenceGraphic bool `json:"violence/graphic"`
		} `json:"categories"`
		// CategoryScores is a struct containing float values for the scores of different categories of text moderation.
		CategoryScores struct {
			// Hate is the score for hate speech.
			Hate float64 `json:"hate"`
			// HateThreatening is the score for threatening hate speech.
			HateThreatening float64 `json:"hate/threatening"`
			// SelfHarm is the score for self-harm content.
			SelfHarm float64 `json:"self-harm"`
			// Sexual is the score for sexual content.
			Sexual float64 `json:"sexual"`
			// SexualMinors is the score for sexual content involving minors.
			SexualMinors float64 `json:"sexual/minors"`
			// Violence is the score for violent content.
			Violence float64 `json:"violence"`
			// ViolenceGraphic is the score for graphic violent content.
			ViolenceGraphic float64 `json:"violence/graphic"`
		} `json:"category_scores"`
		// Flagged is a boolean indicating whether the text was flagged for moderation.
		Flagged bool `json:"flagged"`
	} `json:"results"`
}

// TextModerationRequest represents a request for text moderation.
type TextModerationRequest struct {
	// Model is the ID of the model used for text moderation.
	Model string `json:"model"`
	// Input is the text to be moderated.
	Input string `json:"input"`
}

// DialogRequest represents a request for a dialog.
type DialogRequest struct {
	// Instruction is the instruction for the dialog.
	Instruction string `json:"instruction"`
	// Input is the input for the dialog.
	Input string `json:"input"`
}
