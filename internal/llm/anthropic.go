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

type AnthropicClient struct {
	apiKey  string
	baseURL string
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type anthropicReq struct {
	Model       string           `json:"model"`
	System      string           `json:"system,omitempty"`
	Messages    []anthropicMsg   `json:"messages"`
	MaxTokens   int              `json:"max_tokens"`
	Temperature float64          `json:"temperature,omitempty"`
	Stream      bool             `json:"stream,omitempty"`
}

type anthropicMsg struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicResp struct {
	Content []anthropicContent `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type anthropicDelta struct {
	Text string `json:"text"`
}

type anthropicStreamEvent struct {
	Type  string          `json:"type"`
	Delta *anthropicDelta `json:"delta,omitempty"`
}

func (c *AnthropicClient) base() string {
	if c.baseURL != "" {
		return strings.TrimRight(c.baseURL, "/")
	}
	return "https://api.anthropic.com"
}

func (c *AnthropicClient) Chat(ctx context.Context, messages []Message, opts *ChatOptions) (*ChatResponse, error) {
	return c.chat(ctx, messages, opts, false, nil)
}

func (c *AnthropicClient) ChatStream(ctx context.Context, messages []Message, opts *ChatOptions, onChunk StreamHandler) (*ChatResponse, error) {
	return c.chat(ctx, messages, opts, true, onChunk)
}

func (c *AnthropicClient) chat(ctx context.Context, messages []Message, opts *ChatOptions, stream bool, onChunk StreamHandler) (*ChatResponse, error) {
	system, msgs := splitSystemMessage(messages)

	reqBody := anthropicReq{
		Model:       opts.Model,
		System:      system,
		Messages:    msgs,
		MaxTokens:   4096,
		Temperature: opts.Temperature,
		Stream:      stream,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("anthropic request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	if !stream {
		var anthResp anthropicResp
		if err := json.NewDecoder(resp.Body).Decode(&anthResp); err != nil {
			return nil, fmt.Errorf("parse: %w", err)
		}
		if anthResp.Error != nil {
			return nil, fmt.Errorf("anthropic: %s", anthResp.Error.Message)
		}
		var text string
		for _, c := range anthResp.Content {
			if c.Type == "text" {
				text += c.Text
			}
		}
		return &ChatResponse{
			Choices: []Choice{{Message: Message{Role: "assistant", Content: text}}},
		}, nil
	}

	var fullContent strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")

		var evt anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}

		if evt.Delta != nil && evt.Delta.Text != "" {
			fullContent.WriteString(evt.Delta.Text)
			if onChunk != nil {
				onChunk(evt.Delta.Text)
			}
		}
	}

	return &ChatResponse{
		Choices: []Choice{{Message: Message{Role: "assistant", Content: fullContent.String()}}},
	}, nil
}

func splitSystemMessage(messages []Message) (string, []anthropicMsg) {
	var system string
	var msgs []anthropicMsg

	for _, m := range messages {
		if m.Role == "system" {
			system += m.Content + "\n"
		} else {
			msgs = append(msgs, anthropicMsg{
				Role:    m.Role,
				Content: []anthropicContent{{Type: "text", Text: m.Content}},
			})
		}
	}

	return strings.TrimSpace(system), msgs
}
