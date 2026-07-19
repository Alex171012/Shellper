package shell

import (
	"os"
	"os/exec"
	"strings"
)

func Detect() string {
	shell := os.Getenv("SHELL")
	if shell != "" {
		if strings.Contains(shell, "zsh") {
			return "zsh"
		}
		if strings.Contains(shell, "bash") {
			return "bash"
		}
		return shell
	}

	if _, err := exec.LookPath("bash"); err == nil {
		return "bash"
	}
	if _, err := exec.LookPath("sh"); err == nil {
		return "sh"
	}

	return "sh"
}

func GetShellCmd(shell string) string {
	switch shell {
	case "zsh":
		return "zsh"
	case "bash":
		return "bash"
	default:
		return "sh"
	}
}
