package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Backend != "ollama" {
		t.Errorf("Default backend = %q, want ollama", cfg.Backend)
	}
	if cfg.Model != "llama3.2" {
		t.Errorf("Default model = %q, want llama3.2", cfg.Model)
	}
	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("Default OllamaURL = %q", cfg.OllamaURL)
	}
	if cfg.Safety != "strict" {
		t.Errorf("Default Safety = %q, want strict", cfg.Safety)
	}
	if cfg.DefaultShell != "auto" {
		t.Errorf("Default DefaultShell = %q, want auto", cfg.DefaultShell)
	}
}

func TestLoadWithEnvOverrides(t *testing.T) {
	os.Setenv("SHELLPER_BACKEND", "openai")
	os.Setenv("SHELLPER_MODEL", "gpt-4o-mini")
	os.Setenv("SHELLPER_OLLAMA_URL", "http://10.0.0.1:11434")
	os.Setenv("SHELLPER_OPENAI_KEY", "sk-test-key")
	os.Setenv("SHELLPER_OPENAI_BASE", "https://api.openai.com/v1")
	os.Setenv("SHELLPER_SAFETY", "permissive")
	os.Setenv("SHELLPER_DEFAULT_SHELL", "zsh")
	os.Setenv("SHELLPER_DEFAULT_MODE", "explain")

	defer func() {
		os.Unsetenv("SHELLPER_BACKEND")
		os.Unsetenv("SHELLPER_MODEL")
		os.Unsetenv("SHELLPER_OLLAMA_URL")
		os.Unsetenv("SHELLPER_OPENAI_KEY")
		os.Unsetenv("SHELLPER_OPENAI_BASE")
		os.Unsetenv("SHELLPER_SAFETY")
		os.Unsetenv("SHELLPER_DEFAULT_SHELL")
		os.Unsetenv("SHELLPER_DEFAULT_MODE")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Backend != "openai" {
		t.Errorf("Backend = %q, want openai", cfg.Backend)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Errorf("Model = %q, want gpt-4o-mini", cfg.Model)
	}
	if cfg.OpenAIKey != "sk-test-key" {
		t.Errorf("OpenAIKey = %q", cfg.OpenAIKey)
	}
	if cfg.Safety != "permissive" {
		t.Errorf("Safety = %q, want permissive", cfg.Safety)
	}
	if cfg.DefaultShell != "zsh" {
		t.Errorf("DefaultShell = %q, want zsh", cfg.DefaultShell)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", oldHome)

	orig := &Config{
		Backend:      "openai",
		Model:        "gpt-4",
		OllamaURL:    "http://localhost:11434",
		OpenAIKey:    "sk-test",
		OpenAIBase:   "https://openrouter.ai/api/v1",
		Safety:       "permissive",
		DefaultShell: "bash",
		DefaultMode:  "ask",
	}

	if err := Save(orig); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if loaded.Backend != orig.Backend {
		t.Errorf("Backend = %q, want %q", loaded.Backend, orig.Backend)
	}
	if loaded.Model != orig.Model {
		t.Errorf("Model = %q, want %q", loaded.Model, orig.Model)
	}
	if loaded.Safety != orig.Safety {
		t.Errorf("Safety = %q, want %q", loaded.Safety, orig.Safety)
	}
}
