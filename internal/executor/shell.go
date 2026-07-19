package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/creack/pty"
	"github.com/fatih/color"
)

type Result struct {
	Output   string
	ExitCode int
}

func needsPTY(script string) bool {
	return strings.Contains(script, "sudo ")
}

func Run(script string, shellCmd string) (*Result, error) {
	if needsPTY(script) {
		return runWithPTY(script, shellCmd)
	}
	return runDirect(script, shellCmd)
}

func runDirect(script string, shellCmd string) (*Result, error) {
	cmd := exec.Command(shellCmd, "-c", script)
	var stdout, stderr strings.Builder
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

func runWithPTY(script string, shellCmd string) (*Result, error) {
	cmd := exec.Command(shellCmd, "-c", script)

	f, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("PTY start: %w", err)
	}
	defer f.Close()

	go func() {
		io.Copy(f, os.Stdin)
	}()

	var output strings.Builder
	writer := io.MultiWriter(&output, os.Stdout)
	io.Copy(writer, f)

	err = cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &Result{Output: strings.TrimSpace(output.String()), ExitCode: exitCode}, nil
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
