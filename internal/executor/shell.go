package executor

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

type Result struct {
	Output   string
	ExitCode int
}

func Run(script string, shellCmd string) (*Result, error) {
	cmd := exec.Command(shellCmd, "-c", script)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("execution error: %w", err)
		}
	}

	output := strings.TrimSpace(stdout.String())
	if stderr.Len() > 0 {
		if output != "" {
			output += "\n"
		}
		output += strings.TrimSpace(stderr.String())
	}

	return &Result{Output: output, ExitCode: exitCode}, nil
}

func PrintResult(result *Result) {
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if result.ExitCode == 0 {
		fmt.Println(green("✓ Exit code: 0"))
	} else {
		fmt.Println(red(fmt.Sprintf("✗ Exit code: %d", result.ExitCode)))
	}

	if result.Output != "" {
		fmt.Println(yellow("─── Output ───"))
		fmt.Println(result.Output)
		fmt.Println(yellow("──────────────"))
	}
}
