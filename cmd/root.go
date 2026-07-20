package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"shellper/internal/config"
	"shellper/internal/executor"
	"shellper/internal/llm"
	"shellper/internal/safety"
	"github.com/spf13/cobra"
)

var replHistory []llm.Message

func NewRootCmd() *cobra.Command {
	var force, think, review bool
	var persona string

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: config error: %v\n", err)
		cfg = config.Default()
	}

	app := &appContext{
		llmClient: llm.NewClient(cfg.Backend, cfg.OllamaURL, cfg.OpenAIBase, cfg.OpenAIKey, cfg.AnthropicKey),
		cfg: &appConfig{
			model:        cfg.Model,
			safetyMode:   cfg.Safety,
			shellCmd:     resolveShell(cfg.DefaultShell),
			temperature:  0.7,
			renderMD:     true,
			maxAutoFix:   3,
			thinkEnabled: false,
			reviewAll:    false,
		},
	}

	rootCmd := &cobra.Command{
		Use:   "shellper",
		Short: "AI-powered shell assistant",
		Long: `Shellper — your AI-powered shell companion.

Generate, explain, and execute shell scripts using natural language.
Supports Ollama and OpenAI-compatible backends.

Modes:
  ask      Generate and run a script immediately (no thinking by default)
  explain  Generate and explain a script, always shows plan
  qa       Ask Linux/shell questions (no execution)

Run without arguments for interactive REPL mode.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if force {
				app.cfg.safetyMode = "permissive"
			}
			if think {
				app.cfg.thinkEnabled = true
			}
			if review {
				app.cfg.reviewAll = true
			}
			if persona != "" {
				if _, ok := builtinPersonas[persona]; !ok {
					return fmt.Errorf("unknown persona %q. Available:\n%s", persona, listPersonas())
				}
				appPersona = persona
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			showWelcome(cfg)
			return runREPL(app)
		},
	}

	rootCmd.PersistentFlags().BoolVar(&force, "force", false, "Bypass safety checks (dangerous)")
	rootCmd.PersistentFlags().BoolVar(&think, "think", false, "Enable thinking/planning in ask mode")
	rootCmd.PersistentFlags().BoolVar(&review, "review", false, "Review all scripts before execution (even safe ones)")
	rootCmd.PersistentFlags().StringVar(&persona, "persona", "", "Persona to use (default, beginner, expert). Overrides config")
	rootCmd.AddCommand(askCmd(app))
	rootCmd.AddCommand(explainCmd(app))
	rootCmd.AddCommand(qaCmd(app))
	rootCmd.AddCommand(sessionCmd())
	rootCmd.AddCommand(tuiCmd())

	return rootCmd
}

func runREPL(app *appContext) error {
	scanner := bufio.NewScanner(os.Stdin)
	currentMode := "ask"

	fmt.Println(cyan(bold("Shellper — interactive mode")))
	fmt.Println(yellow("Prefix your request to switch mode:"))
	fmt.Println(yellow("  ask:     generate and run a script (fast)"))
	fmt.Println(yellow("  explain: generate with plan + explanations"))
	fmt.Println(yellow("  qa:      ask a question (no execution)"))
	fmt.Println(yellow("  help     show help"))
	fmt.Println(yellow("  exit     quit"))
	fmt.Println()

	for {
		fmt.Print(green("shellper> "))
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "help" {
			fmt.Println(cyan(bold("Shellper — interactive mode")))
			fmt.Println(yellow("Modes:"))
			fmt.Println(yellow("  ask:     \"ask: create a file\" — fast, no planning"))
			fmt.Println(yellow("  explain: \"explain: find large files\" — shows plan"))
			fmt.Println(yellow("  qa:      \"qa: what is a symlink\" — Q&A only"))
			fmt.Println()
			fmt.Println(yellow("Flags in effect:"))
			fmt.Println(yellow("  think=" + fmt.Sprint(app.cfg.thinkEnabled) + "  review=" + fmt.Sprint(app.cfg.reviewAll)))
			fmt.Println(yellow("  force=" + fmt.Sprint(app.cfg.safetyMode == "permissive") + "  persona=" + appPersona))
			fmt.Println(yellow("Default mode: " + currentMode))
			continue
		}

		mode := currentMode
		query := input

		if strings.HasPrefix(strings.ToLower(input), "ask:") {
			mode = "ask"
			query = strings.TrimSpace(input[4:])
			currentMode = "ask"
		} else if strings.HasPrefix(strings.ToLower(input), "explain:") {
			mode = "explain"
			query = strings.TrimSpace(input[8:])
			currentMode = "explain"
		} else if strings.HasPrefix(strings.ToLower(input), "qa:") {
			mode = "qa"
			query = strings.TrimSpace(input[3:])
			currentMode = "qa"
		}

		if query == "" {
			fmt.Println(red("Empty query. Type 'help' for usage."))
			continue
		}

		ctx, cancel := context.WithCancel(context.Background())

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		go func() {
			<-sigCh
			cancel()
		}()

		switch mode {
		case "qa":
			messages := append(replHistory, buildQAPrompt(query, buildSystemContext())...)
			response, err := getLLMResponse(ctx, app, messages)
			if err != nil {
				fmt.Println(red("Error: " + err.Error()))
			} else {
				if app.cfg.renderMD {
					fmt.Println(renderMarkdown(response))
				} else {
					fmt.Println(response)
				}
				replHistory = append(replHistory,
					llm.Message{Role: "user", Content: query},
					llm.Message{Role: "assistant", Content: response},
				)
				replHistory = trimHistory(replHistory)
			}
		case "ask", "explain":
			think := mode == "explain" || app.cfg.thinkEnabled
			sysCtx := buildSystemContext()
			messages := buildScriptPrompt(query, mode, sysCtx, think)
			if len(replHistory) > 0 {
				messages = append(replHistory, messages...)
			}

			response, err := getLLMResponse(ctx, app, messages)
			if err != nil {
				fmt.Println(red("Error: " + err.Error()))
				cancel()
				continue
			}

			script := extractScript(response)
			if strings.TrimSpace(script) == "" {
				if app.cfg.renderMD {
					fmt.Println(renderMarkdown(response))
				} else {
					fmt.Println(response)
				}
				cancel()
				continue
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

			safetyResult := safety.Check(script, app.cfg.safetyMode)
			action := confirmScript(script, safetyResult, app, scanner)
			if action == confirmAbort {
				cancel()
				continue
			}

			result, err := executor.Run(script, app.cfg.shellCmd)
			if err != nil {
				fmt.Println(red("Error: " + err.Error()))
				cancel()
				continue
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

				safetyResult := safety.Check(fixedScript, app.cfg.safetyMode)
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

			if result.ExitCode == 0 {
				replHistory = append(replHistory,
					llm.Message{Role: "user", Content: query},
					llm.Message{Role: "assistant", Content: response},
				)
				replHistory = trimHistory(replHistory)
			}
		}

		cancel()
	}

	return nil
}
