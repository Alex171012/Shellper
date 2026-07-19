package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"shellper/internal/llm"
	"shellper/internal/shell"
)

type TUIOpts struct {
	Backend      string
	Model        string
	OllamaURL    string
	OpenAIBase   string
	OpenAIKey    string
	SafetyMode   string
	DefaultShell string
	Force        bool
	Think        bool
	Review       bool
	SessionName  string
}

func StartTUI(opts TUIOpts) error {
	client := llm.NewClient(opts.Backend, opts.OllamaURL, opts.OpenAIBase, opts.OpenAIKey)

	shellCmd := opts.DefaultShell
	if shellCmd == "auto" {
		shellCmd = shell.Detect()
	}

	safetyMode := opts.SafetyMode
	if opts.Force {
		safetyMode = "permissive"
	}

	cfg := &appConfig{
		model:        opts.Model,
		safetyMode:   safetyMode,
		shellCmd:     shellCmd,
		temperature:  0.7,
		maxAutoFix:   3,
		renderMD:     true,
		thinkEnabled: opts.Think,
		reviewAll:    opts.Review,
	}

	var loadedMsgs []llm.Message
	if opts.SessionName != "" {
		var err error
		loadedMsgs, err = loadTUISession(opts.SessionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: couldn't load session %q: %v\n", opts.SessionName, err)
		}
	}

	m := initialModel(client, cfg, opts.SessionName, loadedMsgs)
	p := tea.NewProgram(&m, tea.WithAltScreen())

	_, err := p.Run()
	return err
}

func extractTUIScript(response string) string {
	lines := strings.Split(response, "\n")
	inBlock := false
	closed := false
	var scriptLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inBlock {
				inBlock = false
				closed = true
				continue
			}
			if !closed {
				inBlock = true
			}
			continue
		}
		if inBlock {
			scriptLines = append(scriptLines, line)
		}
	}

	if !closed && len(scriptLines) == 0 {
		return response
	}

	return strings.Join(scriptLines, "\n")
}

func buildTUIPrompt(query, mode, sysCtx string, think bool, persona string) []llm.Message {
	cb := "```"

	intro := "You are Shellper, an AI that generates shell scripts (bash/zsh) based on user requests."
	extraRules := "- Keep scripts simple, safe, and beginner-friendly"

	if persona == "beginner" {
		intro = "You are Shellper, a friendly AI tutor that teaches shell scripting to beginners."
		extraRules = "- Prioritize safety above all else\n- Explain every command in simple terms\n- Prefer the safest approach"
	} else if persona == "expert" {
		intro = "You are Shellper, an expert sysadmin AI assistant."
		extraRules = "- Use the most efficient commands\n- Minimize comments\n- Assume the user knows shell scripting"
	}

	thinkBlock := ""
	if think || mode == "explain" {
		thinkBlock = "\n- First, think step-by-step about what needs to be done\n- Explain your plan briefly\n- Then generate the commands"
	} else {
		thinkBlock = "\n- Generate ONLY valid shell commands\n- Output them in a code block — no explanations"
	}

	system := fmt.Sprintf(`%s

%s

Rules:%s
%s
- Wrap the commands in a code block with shell language tag
- Never use sudo in generated commands
- If the request is impossible or dangerous, explain why instead`, intro, sysCtx, thinkBlock, extraRules)

	if mode == "explain" {
		system += fmt.Sprintf(`

For each command, add a comment explaining it. Structure as:

## Plan
Brief explanation.

## Script
%s
# command explanation
actual command
%s`, cb+"shell", cb)
	} else {
		system += fmt.Sprintf(`

## Script
%s
# commands only here
%s`, cb+"shell", cb)
		if think {
			system = strings.Replace(system, "## Script", "## Plan\nBrief explanation.\n\n## Script", 1)
		}
	}

	return []llm.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: query},
	}
}
