package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"shellper/internal/llm"
)

type TUISession struct {
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
	Messages  []llm.Message `json:"messages"`
}

func sessionDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "shellper", "tui-sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func saveTUISession(name string, history []llm.Message) error {
	dir, err := sessionDir()
	if err != nil {
		return err
	}

	path := filepath.Join(dir, name+".json")
	sess := TUISession{
		Name:      name,
		CreatedAt: time.Now(),
		Messages:  history,
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func loadTUISession(name string) ([]llm.Message, error) {
	dir, err := sessionDir()
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("session %q not found", name)
	}

	var sess TUISession
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	return sess.Messages, nil
}
