# Shellper — AI-Powered Shell Assistant

## Concept
Shellper takes natural language requests and generates/executes shell scripts (bash/zsh) using an LLM backend (Ollama by default, OpenAI-compatible as alternative).

## Tech Stack
- **Language**: Go
- **CLI**: cobra + viper
- **LLM**: Direct REST API calls to Ollama and OpenAI-compatible endpoints
- **Config**: YAML at `~/.config/shellper/config.yaml` + env vars

## CLI Commands

| Command | Description |
|---------|-------------|
| `shellper ask "..."` | One-shot: generate + run script |
| `shellper explain "..."` | Generate + explain + confirm before run |
| `shellper qa "..."` | Pure Q&A about Linux/shell (no execution) |

Run `shellper` without args for interactive REPL mode.

## Safety Tiers

| Tier | Examples | Behavior |
|------|----------|----------|
| Safe | ls, cat, echo, touch, pwd, grep, find, head, tail | Auto-approve |
| Risky | rm, mv, cp, chmod, apt/pip/npm install, kill, systemctl | Require y/N confirmation |
| Dangerous | rm -rf /, sudo, fork bomb, pipe to shell, dd, mkfs, shutdown | Blocked unless --force |

## Architecture

```
shellper/
├── cmd/
│   ├── root.go       # REPL mode
│   ├── ask.go        # ask subcommand
│   ├── explain.go    # explain subcommand
│   └── qa.go         # qa subcommand
├── internal/
│   ├── llm/
│   │   ├── client.go # Client interface
│   │   ├── ollama.go # Ollama backend
│   │   └── openai.go # OpenAI-compatible backend
│   ├── safety/
│   │   ├── checker.go # Safety classification
│   │   └── patterns.go # Pattern definitions
│   ├── executor/
│   │   └── shell.go  # Command execution
│   ├── shell/
│   │   └── detect.go # Shell auto-detection
│   └── config/
│       └── config.go # Configuration management
├── main.go
└── go.mod
```

## Pipeline
```
User input → LLM prompt → parse response → safety check
→ auto/confirm/block → execute in shell → show output
```

## Future
- Full desktop agent (expand safety/executor to file system, browser, etc.)
- Plugin system for extensible capabilities
- GUI/TUI interface
