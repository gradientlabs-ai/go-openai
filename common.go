package openai

// common.go defines common types used throughout the OpenAI API.

// Usage Represents the total token usage per request to OpenAI.
type Usage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	TotalTokens      int     `json:"total_tokens"`

	// GroqCloud-specific fields (available in GroqCloud response)
	QueueTime        float64 `json:"queue_time,omitempty"`
	PromptTime       float64 `json:"prompt_time,omitempty"`
	CompletionTime   float64 `json:"completion_time,omitempty"`
	TotalTime        float64 `json:"total_time,omitempty"`
}
