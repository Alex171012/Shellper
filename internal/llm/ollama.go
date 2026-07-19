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

type OllamaClient struct {
	baseURL string
}

type ollamaRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ollamaStreamMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaStreamResponse struct {
	Message ollamaStreamMessage `json:"message"`
	Done    bool                `json:"done"`
	Error   string              `json:"error,omitempty"`
}

func (c *OllamaClient) buildURL() string {
	return fmt.Sprintf("%s/api/chat", strings.TrimRight(c.baseURL, "/"))
}

func (c *OllamaClient) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error) {
	return c.chat(ctx, messages, opts, false, nil)
}

func (c *OllamaClient) ChatStream(ctx context.Context, messages []Message, opts *ChatOptions, onChunk StreamHandler) (*ChatResponse, error) {
	return c.chat(ctx, messages, opts, true, onChunk)
}

func (c *OllamaClient) chat(ctx context.Context, messages []Message, opts *ChatOptions, stream bool, onChunk StreamHandler) (*ChatResponse, error) {
	reqBody := ollamaRequest{
		Model:    opts.Model,
		Messages: messages,
		Stream:   stream,
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

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if !stream {
		var ollamaResp ollamaStreamResponse
		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return nil, fmt.Errorf("parse response: %w", err)
		}
		if ollamaResp.Error != "" {
			return nil, fmt.Errorf("ollama: %s", ollamaResp.Error)
		}
		return &ChatResponse{
			Choices: []Choice{
				{
					Index:   0,
					Message: Message{Role: ollamaResp.Message.Role, Content: ollamaResp.Message.Content},
				},
			},
		}, nil
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var chunk ollamaStreamResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue
		}

		if chunk.Error != "" {
			return nil, fmt.Errorf("ollama: %s", chunk.Error)
		}

		if chunk.Message.Content != "" {
			fullContent.WriteString(chunk.Message.Content)
			if onChunk != nil {
				if err := onChunk(chunk.Message.Content); err != nil {
					return nil, err
				}
			}
		}

		if chunk.Done {
			break
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
