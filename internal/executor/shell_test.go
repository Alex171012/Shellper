package executor

import (
	"strings"
	"testing"
)

func TestRunEcho(t *testing.T) {
	result, err := Run(`echo "hello shellper"`, "bash")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if !strings.Contains(result.Output, "hello shellper") {
		t.Errorf("Output = %q, should contain 'hello shellper'", result.Output)
	}
}

func TestRunExitCode(t *testing.T) {
	result, err := Run("exit 42", "bash")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.ExitCode != 42 {
		t.Errorf("ExitCode = %d, want 42", result.ExitCode)
	}
}

func TestRunStderr(t *testing.T) {
	result, err := Run(`echo "stdout" && echo "stderr" >&2`, "bash")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if !strings.Contains(result.Output, "stdout") {
		t.Errorf("Output missing stdout")
	}
	if !strings.Contains(result.Output, "stderr") {
		t.Errorf("Output missing stderr")
	}
}

func TestRunMultiLine(t *testing.T) {
	script := `echo "line1"
echo "line2"
echo "line3"`
	result, err := Run(script, "bash")
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if !strings.Contains(result.Output, "line1") {
		t.Errorf("Output missing line1")
	}
}

func TestRunWithZsh(t *testing.T) {
	result, err := Run(`echo "zsh test"`, "zsh")
	if err != nil {
		t.Fatalf("Run() with zsh error: %v", err)
	}
	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0", result.ExitCode)
	}
	if !strings.Contains(result.Output, "zsh test") {
		t.Errorf("Output = %q", result.Output)
	}
}

func TestRunInvalidCommand(t *testing.T) {
	result, err := Run("nonexistent_command_xyz", "bash")
	if err != nil {
		t.Fatalf("Run() should not return error for failed command: %v", err)
	}
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code for invalid command")
	}
}
