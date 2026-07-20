package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Backend      string `yaml:"backend"`
	Model        string `yaml:"model"`
	OllamaURL    string `yaml:"ollama_url"`
	OpenAIKey    string `yaml:"openai_key"`
	OpenAIBase   string `yaml:"openai_base"`
	AnthropicKey string `yaml:"anthropic_key"`
	Safety       string `yaml:"safety"`
	DefaultShell string `yaml:"default_shell"`
	DefaultMode  string `yaml:"default_mode"`
}

func Default() *Config {
	return &Config{
		Backend:      "ollama",
		Model:        "llama3.2",
		OllamaURL:    "http://localhost:11434",
		OpenAIBase:   "https://openrouter.ai/api/v1",
		Safety:       "strict",
		DefaultShell: "auto",
		DefaultMode:  "ask",
	}
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "shellper")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (*Config, error) {
	cfg := Default()

	path, err := configPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if v := os.Getenv("SHELLPER_BACKEND"); v != "" {
		cfg.Backend = v
	}
	if v := os.Getenv("SHELLPER_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("SHELLPER_OLLAMA_URL"); v != "" {
		cfg.OllamaURL = v
	}
	if v := os.Getenv("SHELLPER_OPENAI_KEY"); v != "" {
		cfg.OpenAIKey = v
	}
	if v := os.Getenv("SHELLPER_OPENAI_BASE"); v != "" {
		cfg.OpenAIBase = v
	}
	if v := os.Getenv("SHELLPER_ANTHROPIC_KEY"); v != "" {
		cfg.AnthropicKey = v
	}
	if v := os.Getenv("SHELLPER_SAFETY"); v != "" {
		cfg.Safety = v
	}
	if v := os.Getenv("SHELLPER_DEFAULT_SHELL"); v != "" {
		cfg.DefaultShell = v
	}
	if v := os.Getenv("SHELLPER_DEFAULT_MODE"); v != "" {
		cfg.DefaultMode = v
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("Config saved to %s\n", path)
	return nil
}
