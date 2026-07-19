package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/fatih/color"
	"shellper/internal/executor"
	"shellper/internal/llm"
	"shellper/internal/safety"
)

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
)

const maxHistory = 10

type appContext struct {
	llmClient llm.Client
	cfg       *appConfig
}

var appPersona string

type appConfig struct {
	model        string
	safetyMode   string
	shellCmd     string
	temperature  float64
	renderMD     bool
	maxAutoFix   int
	thinkEnabled bool
	reviewAll    bool
}

func renderMarkdown(text string) string {
	if !glamourEnabled() {
		return text
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return strings.TrimSpace(out)
}

func glamourEnabled() bool {
	if os.Getenv("SHELLPER_NO_COLOR") != "" {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return true
}

func buildSystemContext() string {
	hostname, _ := os.Hostname()
	cwd, _ := os.Getwd()
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	return fmt.Sprintf(`System context:
- OS: %s/%s
- Host: %s
- User: %s
- Shell: %s
- Working directory: %s
- Home: %s`,
		runtime.GOOS, runtime.GOARCH,
		hostname,
		user,
		shell,
		cwd,
		os.Getenv("HOME"),
	)
}

func buildScriptPrompt(query string, mode string, sysCtx string, think bool) []llm.Message {
	cb := "```"
	persona := getPersona(appPersona)

	var thinkBlock string
	if think || mode == "explain" {
		thinkBlock = `
- First, think step-by-step about what needs to be done
- Explain your plan briefly
- Then generate the commands`
	} else {
		thinkBlock = `
- Generate ONLY valid shell commands
- Output them in a code block — no explanations, no commentary`
	}

	personaRules := persona.Rules
	if personaRules == "" {
		personaRules = `
- Keep scripts simple, safe, and beginner-friendly`
	}

	system := fmt.Sprintf(`%s

%s

Rules:%s
%s
- Wrap the commands in a code block with shell language tag
- Prefer built-in commands over external tools when possible
- Never use sudo in generated commands
- If the request is impossible or dangerous, explain why instead`, persona.SystemIntro, sysCtx, thinkBlock, personaRules)

	if mode == "explain" {
		system += fmt.Sprintf(`

For each command, add a comment above it explaining what it does and why. Structure your response as:

## Plan
Brief explanation of what you'll do and why.

## Script
%s
# command explanation
actual command
%s`, cb+"shell", cb)
	} else {
		structure := fmt.Sprintf(`

## Script
%s
# commands only here
%s`, cb+"shell", cb)
		system += structure
		if think {
			system = strings.Replace(system, "## Script", "## Plan\nBrief step-by-step explanation.\n\n## Script", 1)
		}
	}

	return []llm.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: query},
	}
}

func buildQAPrompt(query string, sysCtx string) []llm.Message {
	system := fmt.Sprintf(`You are Shellper, a Linux/shell assistant.

%s

Rules:
- Answer questions about Linux, shell commands, bash/zsh, and system administration
- Be concise but thorough
- Provide examples when helpful
- If something is dangerous, warn about it
- Format responses in markdown for readability
- Don't generate commands unless asked`, sysCtx)

	return []llm.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: query},
	}
}

func buildFixPrompt(script string, errorOutput string) []llm.Message {
	return []llm.Message{
		{Role: "system", Content: `You are Shellper. The script you generated failed to execute correctly.
Analyze the error and produce a corrected version of the script.

Rules:
- Output the corrected script in a shell code block
- Keep the same task intent
- Only fix what's broken, don't add unnecessary changes`},
		{Role: "user", Content: fmt.Sprintf("Script that failed:\n```shell\n%s\n```\n\nError output:\n```\n%s\n```\n\nPlease provide a corrected script.", script, errorOutput)},
	}
}

func extractScript(response string) string {
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

func getLLMResponse(ctx context.Context, app *appContext, messages []llm.Message) (string, error) {
	return getLLMResponseStream(ctx, app, messages, true)
}

func getLLMResponseSilent(ctx context.Context, app *appContext, messages []llm.Message) (string, error) {
	return getLLMResponseStream(ctx, app, messages, false)
}

func getLLMResponseStream(ctx context.Context, app *appContext, messages []llm.Message, showStream bool) (string, error) {
	opts := &llm.ChatOptions{
		Model:       app.cfg.model,
		Temperature: app.cfg.temperature,
	}

	if !showStream {
		resp, err := app.llmClient.Chat(ctx, messages, opts)
		if err != nil {
			return "", fmt.Errorf("LLM request failed: %w", err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("LLM returned empty response")
		}
		return strings.TrimSpace(resp.Choices[0].Message.Content), nil
	}

	fmt.Print(green(bold("Generating...")))
	headerLen := len("Generating...")

	onChunk := llm.StreamHandler(func(chunk string) error {
		if headerLen > 0 {
			fmt.Print("\r" + strings.Repeat(" ", headerLen) + "\r")
			headerLen = 0
		}
		fmt.Print(chunk)
		return nil
	})

	resp, err := app.llmClient.ChatStream(ctx, messages, opts, onChunk)
	if err != nil {
		return "", fmt.Errorf("LLM request failed: %w", err)
	}

	fmt.Println()

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned empty response")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

type confirmAction int

const (
	confirmRun confirmAction = iota
	confirmAbort
	confirmForce
)

func confirmScript(script string, safetyResult *safety.Result, app *appContext, scanner *bufio.Scanner) confirmAction {
	summary := safetyResult.Tier
	tierStr := ""
	tierColor := green

	switch {
	case summary == safety.TierSafe && app.cfg.reviewAll:
		tierStr = "REVIEW"
		tierColor = cyan
	case summary == safety.TierSafe:
		if app.cfg.thinkEnabled || app.cfg.reviewAll {
			tierStr = "SAFE"
			tierColor = green
		} else {
			return confirmRun
		}
	case summary == safety.TierRisky:
		tierStr = "RISKY"
		tierColor = yellow
	case summary == safety.TierDangerous:
		tierStr = "DANGEROUS"
		tierColor = red
	}

	fmt.Println()
	fmt.Println(tierColor(bold(fmt.Sprintf("═══ Safety: %s ═══", tierStr))))

	if len(safetyResult.Blocks) > 0 {
		for _, block := range safetyResult.Blocks {
			fmt.Println(tierColor("  • " + block))
		}
		fmt.Println()
	}

	if safetyResult.IsDangerous() {
		if app.cfg.safetyMode == "permissive" {
			fmt.Println(red(bold("⚠ DANGEROUS COMMANDS DETECTED")))
			fmt.Println(red("Type 'yes' to confirm you want to run this:"))
			if !scanner.Scan() {
				return confirmAbort
			}
			if strings.TrimSpace(strings.ToLower(scanner.Text())) != "yes" {
				fmt.Println("Cancelled.")
				return confirmAbort
			}
			return confirmForce
		}
		fmt.Println(red(bold("⚠ BLOCKED: Dangerous command detected")))
		fmt.Println(yellow("Use --force to allow dangerous commands"))
		return confirmAbort
	}

	if strings.Contains(script, "sudo ") {
		fmt.Println(yellow(bold("🔒 This script uses sudo — your system will ask for your password.")))
		fmt.Println(yellow("   Shellper does NOT see or store your password."))
		fmt.Println()
	}

	fmt.Print(tierColor("Run this script? (y/N): "))
	if !scanner.Scan() {
		return confirmAbort
	}
	confirm := strings.TrimSpace(strings.ToLower(scanner.Text()))
	if confirm != "y" {
		fmt.Println("Cancelled.")
		return confirmAbort
	}
	return confirmRun
}

func handleScriptExecution(ctx context.Context, app *appContext, query string, mode string, think bool) error {
	sysCtx := buildSystemContext()
	messages := buildScriptPrompt(query, mode, sysCtx, think)
	response, err := getLLMResponse(ctx, app, messages)
	if err != nil {
		return err
	}

	script := extractScript(response)

	if strings.TrimSpace(script) == "" {
		if app.cfg.renderMD {
			fmt.Println(renderMarkdown(response))
		} else {
			fmt.Println(response)
		}
		return nil
	}

	var planPart string
	if idx := strings.Index(response, "## Script"); idx > 0 {
		planPart = strings.TrimSpace(response[:idx])
	}
	if planPart != "" {
		if app.cfg.renderMD {
			fmt.Println(renderMarkdown(planPart))
		} else {
			fmt.Println(planPart)
		}
		fmt.Println()
	}

	fmt.Println(cyan("─── Generated Script ───"))
	fmt.Println(script)
	fmt.Println(cyan("─────────────────────────"))

	scanner := bufio.NewScanner(os.Stdin)
	safetyResult := safety.Check(script, app.cfg.safetyMode)

	action := confirmScript(script, safetyResult, app, scanner)
	if action == confirmAbort {
		return nil
	}

	result, err := executor.Run(script, app.cfg.shellCmd)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	executor.PrintResult(result)

	autoFixCount := 0
	for result.ExitCode != 0 && autoFixCount < app.cfg.maxAutoFix {
		autoFixCount++
		fmt.Println()
		fmt.Println(yellow(bold(fmt.Sprintf("Attempting auto-fix (%d/%d)...", autoFixCount, app.cfg.maxAutoFix))))

		fixMessages := buildFixPrompt(script, result.Output)
		fixResponse, err := getLLMResponseSilent(ctx, app, fixMessages)
		if err != nil {
			fmt.Println(red("Auto-fix request failed: " + err.Error()))
			break
		}

		fixedScript := extractScript(fixResponse)
		if strings.TrimSpace(fixedScript) == "" || fixedScript == script {
			fmt.Println(yellow("No useful fix generated."))
			break
		}

		fmt.Println(cyan("─── Fixed Script ───"))
		fmt.Println(fixedScript)
		fmt.Println(cyan("────────────────────"))

		safetyResult = safety.Check(fixedScript, app.cfg.safetyMode)
		action = confirmScript(fixedScript, safetyResult, app, scanner)
		if action == confirmAbort {
			break
		}

		result, err = executor.Run(fixedScript, app.cfg.shellCmd)
		if err != nil {
			fmt.Println(red("Auto-fix execution failed: " + err.Error()))
			break
		}

		executor.PrintResult(result)
		script = fixedScript
	}

	if autoFixCount > 0 && result.ExitCode != 0 {
		fmt.Println(yellow("Auto-fix exhausted. Try rephrasing your request."))
	}

	return nil
}

func trimHistory(history []llm.Message) []llm.Message {
	if len(history) > maxHistory*2 {
		return history[len(history)-maxHistory*2:]
	}
	return history
}

func resolveShell(cfgShell string) string {
	if cfgShell == "auto" {
		shell := os.Getenv("SHELL")
		if strings.Contains(shell, "zsh") {
			return "zsh"
		}
		if strings.Contains(shell, "bash") {
			return "bash"
		}
		if _, err := exec.LookPath("bash"); err == nil {
			return "bash"
		}
		return "sh"
	}
	return cfgShell
}
