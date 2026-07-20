package llm

import "context"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
	Error   string   `json:"error,omitempty"`
}

type StreamHandler func(chunk string) error

type Client interface {
	Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error)
	ChatStream(ctx context.Context, messages []Message, opts *ChatOptions, onChunk StreamHandler) (*ChatResponse, error)
}

type ChatOptions struct {
	Model       string
	Temperature float64
	MaxTokens   int
}

func NewClient(backend, ollamaURL, openaiBase, openaiKey, anthropicKey string) Client {
	switch backend {
	case "openai":
		return &OpenAIClient{
			baseURL: openaiBase,
			apiKey:  openaiKey,
		}
	case "anthropic":
		return &AnthropicClient{
			apiKey:  anthropicKey,
			baseURL: ollamaURL,
		}
	default:
		return &OllamaClient{
			baseURL: ollamaURL,
		}
	}
}
