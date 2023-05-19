package openai

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

type ModelInfo struct {
	ID         string            `json:"id"`
	Object     string            `json:"object"`
	OwnedBy    string            `json:"owned_by"`
	Permission []ModelPermission `json:"permission"`
}

type ModelPermission struct {
	ID                 string      `json:"id"`
	Object             string      `json:"object"`
	Created            int         `json:"created"`
	AllowCreateEngine  bool        `json:"allow_create_engine"`
	AllowSampling      bool        `json:"allow_sampling"`
	AllowLogprobs      bool        `json:"allow_logprobs"`
	AllowSearchIndices bool        `json:"allow_search_indices"`
	AllowView          bool        `json:"allow_view"`
	AllowFineTuning    bool        `json:"allow_fine_tuning"`
	Organization       string      `json:"organization"`
	Group              interface{} `json:"group"`
	IsBlocking         bool        `json:"is_blocking"`
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

type ImageRequest struct {
	Model          string             `json:"model"`
	Prompt         string             `json:"prompt"`
	Size           ImageSizeSupported `json:"size"`
	N              int                `json:"n"`
	ResponseFormat string             `json:"response_format"`
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
type AudioResponse struct {
	Text string `json:"text"`
}

type FileList struct {
	Data   []FileInfo `json:"data"`
	Object string     `json:"object"`
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
	ID              string                 `json:"id"`
	Object          string                 `json:"object"`
	Model           string                 `json:"model"`
	CreatedAt       int                    `json:"created_at"`
	Events          []FineTuneEvent        `json:"events"`
	FineTunedModel  string                 `json:"fine_tuned_model"`
	Hyperparams     FineTuneJobHyperparams `json:"hyperparams"`
	OrganizationID  string                 `json:"organization_id"`
	ResultFiles     []FileInfo             `json:"result_files"`
	Status          string                 `json:"status"`
	ValidationFiles []FileInfo             `json:"validation_files"`
	TrainingFiles   []FileInfo             `json:"training_files"`
	UpdatedAt       int                    `json:"updated_at"`
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

type FineTuneJobEventList struct {
	Data   []FineTuneEvent `json:"data"`
	Object string          `json:"object"`
}

type FineTuneJobHyperparams struct {
	BatchSize        int     `json:"batch_size"`
	LearningRateMult float64 `json:"learning_rate_multiplier"`
	NEpochs          int     `json:"n_epochs"`
	PromptLossWeight float64 `json:"prompt_loss_weight"`
}

type ModelDelete struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
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
	Input string `json:"input"`
}

type DialogRequest struct {
	Instruction string `json:"instruction"`
	Input       string `json:"input"`
}
