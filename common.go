package openai

// common.go defines common types used throughout the OpenAI API.

type PromptTokenDetails struct {
	CachedTokens int `json:"cached_tokens,omitempty"`
}

type CompletionTokenDetails struct {
	ReasoningTokens          int `json:"reasoning_tokens,omitempty"`
	AcceptedPredictionTokens int `json:"accepted_prediction_tokens,omitempty"`
	RejectedPredictionTokens int `json:"rejected_prediction_tokens,omitempty"`
}

// Usage Represents the total token usage per request to OpenAI.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`

	// GroqCloud-specific fields (available in GroqCloud response)
	QueueTime              float64                `json:"queue_time,omitempty"`
	PromptTime             float64                `json:"prompt_time,omitempty"`
	CompletionTime         float64                `json:"completion_time,omitempty"`
	TotalTime              float64                `json:"total_time,omitempty"`
	PromptTokenDetails     PromptTokenDetails     `json:"prompt_tokens_details,omitempty"`
	CompletionTokenDetails CompletionTokenDetails `json:"completion_tokens_details,omitempty"`
}
