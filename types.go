package openai

type OpenAIRole string

const (
	User      OpenAIRole = "user"
	System    OpenAIRole = "sysmtem"
	Assistant OpenAIRole = "assistant"
)

type ModelInfo struct {
	ID         string   `json:"id"`
	Object     string   `json:"object"`
	OwnedBy    string   `json:"owned_by"`
	Permission []string `json:"permission"`
}

type CompletionRequest struct {
	Model       string  `json:"model"`
	Prompt      string  `json:"prompt"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	N           int     `json:"n"`
	Stream      bool    `json:"stream"`
	Logprobs    int     `json:"logprobs"`
	Stop        string  `json:"stop"`
}

type CompletionResponse struct {
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Created int    `json:"created"`
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"messages"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type EditChatRequest struct {
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"` //system, user, or assistant.
		Content string `json:"content"`
	} `json:"messages"`
	Instruction string `json:"instruction"`
}

type EditChatResponse struct {
	Choices []struct {
		Text         string      `json:"text"`
		Index        int         `json:"index"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Created int    `json:"created"`
	ID      string `json:"id"`
	Model   string `json:"model"`
	Object  string `json:"object"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ModelList struct {
	Data   []ModelInfo `json:"data"`
	Object string      `json:"object"`
}

func NewModelList() *ModelList {
	return &ModelList{
		Data: []ModelInfo{
			{
				ID:         "model-id-0",
				Object:     "model",
				OwnedBy:    "organization-owner",
				Permission: []string{"read", "write"},
			},
			{
				ID:         "model-id-1",
				Object:     "model",
				OwnedBy:    "organization-owner",
				Permission: []string{"read", "write"},
			},
			{
				ID:         "model-id-2",
				Object:     "model",
				OwnedBy:    "openai",
				Permission: []string{"read", "write"},
			},
		},
		Object: "list",
	}
}

type ImageRequest struct {
	Model          string `json:"model"`
	Prompt         string `json:"prompt"`
	Size           string `json:"size"`
	ResponseFormat string `json:"response_format"`
}

type ImageResponse struct {
	Created int `json:"created"`
	Data    []struct {
		URL string `json:"url"`
	} `json:"data"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingResponse struct {
	Model  string `json:"model"`
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type AudioRequest struct {
	Model string `json:"model"`
	File  string `json:"file"`
}

type AudioResponse struct {
	Text string `json:"text"`
}

type File struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int    `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type FileList struct {
	Data   []File `json:"data"`
	Object string `json:"object"`
}

func NewFile() *File {
	return &File{
		ID:        "file-XjGxS3KTG0uNmNOK362iJua3",
		Object:    "file",
		Bytes:     140,
		CreatedAt: 1613779121,
		Filename:  "mydata.jsonl",
		Purpose:   "fine-tune",
	}
}

type UploadResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type DeleteFileResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

type FileInfo struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int    `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type FineTuneJob struct {
	ID             string          `json:"id"`
	Object         string          `json:"object"`
	Model          string          `json:"model"`
	CreatedAt      int             `json:"created_at"`
	Events         []FineTuneEvent `json:"events"`
	FineTunedModel string          `json:"fine_tuned_model"`
	Hyperparams    struct {
		BatchSize        int     `json:"batch_size"`
		LearningRateMult float64 `json:"learning_rate_multiplier"`
		NEpochs          int     `json:"n_epochs"`
		PromptLossWeight float64 `json:"prompt_loss_weight"`
	} `json:"hyperparams"`
	OrganizationID  string `json:"organization_id"`
	ResultFiles     []File `json:"result_files"`
	Status          string `json:"status"`
	ValidationFiles []File `json:"validation_files"`
	TrainingFiles   []File `json:"training_files"`
	UpdatedAt       int    `json:"updated_at"`
}

type FineTuneEvent struct {
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type FineTuneJobList struct {
	Data   []FineTuneJob `json:"data"`
	Object string        `json:"object"`
}

type FineTuneJobDetail struct {
	ID             string          `json:"id"`
	Object         string          `json:"object"`
	Model          string          `json:"model"`
	CreatedAt      int             `json:"created_at"`
	Events         []FineTuneEvent `json:"events"`
	FineTunedModel string          `json:"fine_tuned_model"`
	Hyperparams    struct {
		BatchSize        int     `json:"batch_size"`
		LearningRateMult float64 `json:"learning_rate_multiplier"`
		NEpochs          int     `json:"n_epochs"`
		PromptLossWeight float64 `json:"prompt_loss_weight"`
	} `json:"hyperparams"`
	OrganizationID  string `json:"organization_id"`
	ResultFiles     []File `json:"result_files"`
	Status          string `json:"status"`
	ValidationFiles []File `json:"validation_files"`
	TrainingFiles   []File `json:"training_files"`
	UpdatedAt       int    `json:"updated_at"`
}

type FineTuneJobResultFile struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int    `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type FineTuneJobTrainingFile struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int    `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

type FineTuneJobEvent struct {
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type FineTuneJobHyperparams struct {
	BatchSize        int     `json:"batch_size"`
	LearningRateMult float64 `json:"learning_rate_multiplier"`
	NEpochs          int     `json:"n_epochs"`
	PromptLossWeight float64 `json:"prompt_loss_weight"`
}

type FineTuneJobResult struct {
	ID              string                    `json:"id"`
	Object          string                    `json:"object"`
	Model           string                    `json:"model"`
	CreatedAt       int                       `json:"created_at"`
	Events          []FineTuneJobEvent        `json:"events"`
	FineTunedModel  string                    `json:"fine_tuned_model"`
	Hyperparams     FineTuneJobHyperparams    `json:"hyperparams"`
	OrganizationID  string                    `json:"organization_id"`
	ResultFiles     []FineTuneJobResultFile   `json:"result_files"`
	Status          string                    `json:"status"`
	ValidationFiles []File                    `json:"validation_files"`
	TrainingFiles   []FineTuneJobTrainingFile `json:"training_files"`
	UpdatedAt       int                       `json:"updated_at"`
}

type Input struct {
	Text string `json:"text"`
}

type TextModerationResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Results []struct {
		Categories struct {
			Hate            bool `json:"hate"`
			HateThreatening bool `json:"hate/threatening"`
			SelfHarm        bool `json:"self-harm"`
			Sexual          bool `json:"sexual"`
			SexualMinors    bool `json:"sexual/minors"`
			Violence        bool `json:"violence"`
			ViolenceGraphic bool `json:"violence/graphic"`
		} `json:"categories"`
		CategoryScores struct {
			Hate            float64 `json:"hate"`
			HateThreatening float64 `json:"hate/threatening"`
			SelfHarm        float64 `json:"self-harm"`
			Sexual          float64 `json:"sexual"`
			SexualMinors    float64 `json:"sexual/minors"`
			Violence        float64 `json:"violence"`
			ViolenceGraphic float64 `json:"violence/graphic"`
		} `json:"category_scores"`
		Flagged bool `json:"flagged"`
	} `json:"results"`
}

type TextModerationRequest struct {
	Model string `json:"model"`
	Input Input  `json:"input"`
}

func NewTextModerationRequest(text string) *TextModerationRequest {
	return &TextModerationRequest{
		Model: "text-moderation-001",
		Input: Input{
			Text: text,
		},
	}
}
