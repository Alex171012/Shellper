package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAIClient struct {
	baseURL string
	apiKey  string
}

type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type openAIChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type openAIDelta struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}

type openAIStreamChoice struct {
	Index        int         `json:"index"`
	Delta       openAIDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIStreamResponse struct {
	Choices []openAIStreamChoice `json:"choices"`
	Error   *openAIError         `json:"error,omitempty"`
}

func (c *OpenAIClient) buildURL() string {
	base := strings.TrimRight(c.baseURL, "/")
	return fmt.Sprintf("%s/chat/completions", base)
}

func (c *OpenAIClient) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error) {
	return c.chat(ctx, messages, opts, false, nil)
}

func (c *OpenAIClient) ChatStream(ctx context.Context, messages []Message, opts *ChatOptions, onChunk StreamHandler) (*ChatResponse, error) {
	return c.chat(ctx, messages, opts, true, onChunk)
}

func (c *OpenAIClient) chat(ctx context.Context, messages []Message, opts *ChatOptions, stream bool, onChunk StreamHandler) (*ChatResponse, error) {
	reqBody := openAIRequest{
		Model:       opts.Model,
		Messages:    messages,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Stream:      stream,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if !stream {
		var openAIResp openAIResponse
		if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
			return nil, fmt.Errorf("parse response: %w", err)
		}
		if openAIResp.Error != nil {
			return nil, fmt.Errorf("openai: %s (%s)", openAIResp.Error.Message, openAIResp.Error.Type)
		}
		choices := make([]Choice, len(openAIResp.Choices))
		for i, c := range openAIResp.Choices {
			choices[i] = Choice{
				Index:        c.Index,
				Message:      c.Message,
				FinishReason: c.FinishReason,
			}
		}
		return &ChatResponse{Choices: choices}, nil
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		if data == "[DONE]" {
			break
		}

		var chunk openAIStreamResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if chunk.Error != nil {
			return nil, fmt.Errorf("openai: %s (%s)", chunk.Error.Message, chunk.Error.Type)
		}

		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				if onChunk != nil {
					if err := onChunk(choice.Delta.Content); err != nil {
						return nil, err
					}
				}
			}
			if choice.FinishReason != nil {
				break
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read stream: %w", err)
	}

	return &ChatResponse{
		Choices: []Choice{
			{
				Index:   0,
				Message: Message{Role: "assistant", Content: fullContent.String()},
			},
		},
	}, nil
}
