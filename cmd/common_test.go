package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/sasha_tecno/shellper/internal/llm"
)

func TestExtractScript(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{
			name:     "simple code block",
			response: "```shell\necho hello\n```",
			want:     "echo hello",
		},
		{
			name:     "code block with plan",
			response: "## Plan\nDo x\n\n## Script\n```shell\necho hello\n```",
			want:     "echo hello",
		},
		{
			name:     "no code block returns whole",
			response: "just a response",
			want:     "just a response",
		},
		{
			name:     "multi-line script",
			response: "```shell\nls -la\necho done\n```",
			want:     "ls -la\necho done",
		},
		{
			name:     "empty code block",
			response: "```shell\n```",
			want:     "",
		},
		{
			name:     "multiple code blocks (extracts first)",
			response: "```shell\necho first\n```\n...\n```shell\necho second\n```",
			want:     "echo first",
		},
		{
			name:     "code block with bash tag",
			response: "```bash\necho hello\n```",
			want:     "echo hello",
		},
		{
			name:     "code block with language tag and extra spaces",
			response: "```shell \necho hello\n```",
			want:     "echo hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractScript(tt.response)
			if got != tt.want {
				t.Errorf("extractScript() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildSystemContext(t *testing.T) {
	ctx := buildSystemContext()
	if !strings.Contains(ctx, "OS:") {
		t.Errorf("buildSystemContext() missing OS")
	}
	if !strings.Contains(ctx, "Shell:") {
		t.Errorf("buildSystemContext() missing Shell")
	}
	if !strings.Contains(ctx, "Working directory:") {
		t.Errorf("buildSystemContext() missing Working directory")
	}
}

func TestBuildScriptPrompt(t *testing.T) {
	sysCtx := "System context:\n- OS: linux/amd64"

	prompt := buildScriptPrompt("list files", "ask", sysCtx, false)
	if len(prompt) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(prompt))
	}
	if prompt[0].Role != "system" {
		t.Errorf("first message role = %q, want system", prompt[0].Role)
	}
	if prompt[1].Role != "user" {
		t.Errorf("second message role = %q, want user", prompt[1].Role)
	}
	if !strings.Contains(prompt[1].Content, "list files") {
		t.Errorf("user message should contain query")
	}

	thinkPrompt := buildScriptPrompt("list files", "ask", sysCtx, true)
	thinkContent := thinkPrompt[0].Content
	if !strings.Contains(thinkContent, "think step-by-step") {
		t.Errorf("think prompt should mention thinking, got: %s", thinkContent)
	}
}

func TestBuildQAPrompt(t *testing.T) {
	sysCtx := "System context:\n- OS: linux/amd64"
	prompt := buildQAPrompt("what is a symlink", sysCtx)
	if len(prompt) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(prompt))
	}
	if !strings.Contains(prompt[0].Content, "Linux/shell assistant") {
		t.Errorf("system prompt should mention Linux/shell assistant")
	}
}

func TestBuildFixPrompt(t *testing.T) {
	prompt := buildFixPrompt("echo hello", "command not found")
	if len(prompt) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(prompt))
	}
	fixContent := prompt[1].Content
	if !strings.Contains(fixContent, "echo hello") {
		t.Errorf("fix prompt should contain original script")
	}
	if !strings.Contains(fixContent, "command not found") {
		t.Errorf("fix prompt should contain error output")
	}
}

func TestTrimHistory(t *testing.T) {
	msgs := make([]llm.Message, 25)
	for i := range msgs {
		msgs[i] = llm.Message{Role: "user", Content: "msg"}
	}

	trimmed := trimHistory(msgs)
	if len(trimmed) > maxHistory*2 {
		t.Errorf("trimmed length = %d, want <= %d", len(trimmed), maxHistory*2)
	}

	small := make([]llm.Message, 5)
	trimmedSmall := trimHistory(small)
	if len(trimmedSmall) != 5 {
		t.Errorf("small history should not be trimmed, got %d", len(trimmedSmall))
	}
}

func TestResolveShell(t *testing.T) {
	oldShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", oldShell)

	shell := resolveShell("auto")
	if shell != "bash" && shell != "zsh" && shell != "sh" {
		t.Errorf("resolveShell(auto) = %q, want a valid shell", shell)
	}

	if got := resolveShell("zsh"); got != "zsh" {
		t.Errorf("resolveShell(zsh) = %q, want zsh", got)
	}
	if got := resolveShell("bash"); got != "bash" {
		t.Errorf("resolveShell(bash) = %q, want bash", got)
	}

	os.Setenv("SHELL", "/usr/bin/zsh")
	if got := resolveShell("auto"); got != "zsh" {
		t.Errorf("resolveShell(auto) with SHELL=/usr/bin/zsh = %q, want zsh", got)
	}

	os.Setenv("SHELL", "/bin/bash")
	if got := resolveShell("auto"); got != "bash" {
		t.Errorf("resolveShell(auto) with SHELL=/bin/bash = %q, want bash", got)
	}
}
