package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/sasha_tecno/shellper/internal/config"
)

var welcomeShown bool

func showWelcome(cfg *config.Config) {
	if welcomeShown {
		return
	}
	welcomeShown = true

	configPath, err := configPath()
	if err != nil {
		return
	}

	_, err = os.Stat(configPath)
	isFirstRun := os.IsNotExist(err)

	if isFirstRun {
		fmt.Println()
		fmt.Println(cyan(bold("╔══════════════════════════════════════════╗")))
		fmt.Println(cyan(bold("║       Welcome to Shellper! v0.1.0        ║")))
		fmt.Println(cyan(bold("║   AI-powered shell assistant             ║")))
		fmt.Println(cyan(bold("╚══════════════════════════════════════════╝")))
		fmt.Println()
		fmt.Println(yellow("First run detected — creating default config at:"))
		fmt.Println(yellow("  " + configPath))
		fmt.Println()

		if err := config.Save(cfg); err != nil {
			fmt.Println(red("Failed to save config: " + err.Error()))
		}

		fmt.Println(green("Quick start:"))
		fmt.Println(green("  shellper ask \"list files in current directory\""))
		fmt.Println(green("  shellper explain \"how to find large files\""))
		fmt.Println(green("  shellper qa \"what is a symlink\""))
		fmt.Println(green("  shellper                    # interactive mode"))
		fmt.Println()
	}

	if strings.HasPrefix(cfg.Backend, "ollama") {
		checkOllamaStatus(cfg.OllamaURL)
		fmt.Println()
	}
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + "/.config/shellper/config.yaml", nil
}

func checkOllamaStatus(url string) {
	base := strings.TrimRight(url, "/")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/tags", nil)
	if err != nil {
		fmt.Println(yellow("⚠ Could not check Ollama status"))
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(yellow(bold("⚠ Ollama is not running!")))
		fmt.Println(yellow(fmt.Sprintf("  Start it with: ollama serve")))
		fmt.Println(yellow(fmt.Sprintf("  Or set a different backend in ~/.config/shellper/config.yaml")))
		return
	}
	resp.Body.Close()

	fmt.Println(green("✓ Ollama is running at " + base))
}
