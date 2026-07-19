# Shellper

> **AI-powered shell assistant** — turn natural language into bash/zsh scripts.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
![Tests](https://img.shields.io/badge/tests-21%20passing-brightgreen)

---

## Features

| | |
|---|---|
| **Natural language to scripts** | `shellper ask "compress all logs older than 7 days"` |
| **Learn mode** | `shellper explain "how to find large files"` — shows plan + explanations |
| **Linux Q&A** | `shellper qa "what is a symlink"` — no execution, just answers |
| **REPL mode** | `shellper` — interactive session with conversation memory |
| **3-tier safety** | Safe (auto-run) · Risky (confirm) · Dangerous (blocked) |
| **Streaming** | See tokens arrive in real-time as the AI generates |
| **Multiple backends** | Ollama (default) · OpenAI-compatible (OpenRouter, etc.) |
| **Custom personas** | Default · Beginner · Expert — tune tone and depth |
| **Auto-fix** | Failed scripts are sent back to the AI for automatic correction |
| **Session save/load** | Save REPL conversations, continue later |
| **Markdown rendering** | Beautiful terminal output via Charm Glamour |

---

## Quick start

### 1. Install

```bash
git clone <your-repo-url>
cd shellper
go build -o shellper .
```

### 2. Run

```bash
# Make sure Ollama is running with a model pulled
ollama pull llama3.2

# One-shot: generate + run a script
./shellper ask "create a project directory structure for a Go app"

# Learn mode: shows plan and explains each command
./shellper explain "backup my home directory to /backup"

# Q&A: just ask a question
./shellper qa "how does piping work in bash"

# Interactive session
./shellper
```

---

## Usage

### CLI commands

```
shellper ask "query"        Generate and run a script (fast, no planning)
shellper explain "query"    Generate with plan + explanations
shellper qa "question"      Ask a Linux/shell question
shellper session save/load  Save and load REPL conversations
shellper                    Interactive REPL mode
```

### Flags

| Flag | Description |
|------|-------------|
| `--think` | Enable planning step in `ask` mode |
| `--review` | Review all scripts before execution (even safe ones) |
| `--force` | Allow dangerous commands (requires typing "yes") |
| `--persona default\|beginner\|expert` | Set AI persona |

### Examples

```bash
# Fast execution
shellper ask "list all files modified in the last 24 hours"

# With planning
shellper --think ask "set up a git repo with initial commit"

# Safe learning
shellper explain "how to grep recursively for a pattern"

# Expert mode — concise, advanced commands
shellper --persona expert ask "find top 10 largest files"

# Beginner mode — extra safety, verbose explanations
shellper --persona beginner explain "what does chmod 755 do"

# Review everything
shellper --review ask "rename all .txt files to .md"
```

---

## Configuration

Config file: `~/.config/shellper/config.yaml`

```yaml
backend: ollama           # "ollama" or "openai"
model: llama3.2
ollama_url: http://localhost:11434
openai_base: https://openrouter.ai/api/v1
openai_key: ""            # or use SHELLPER_OPENAI_KEY env var
safety: strict            # "strict" or "permissive"
default_shell: auto       # "auto", "bash", or "zsh"
default_mode: ask         # "ask", "explain", or "qa"
```

Every field can be set via environment variable (`SHELLPER_*`).

---

## Safety

Shellper uses a **3-tier pattern-based safety system**:

| Tier | Examples | Behaviour |
|------|----------|-----------|
| 🔵 **Safe** | `ls`, `cat`, `echo`, `touch`, `grep`, `find` | Auto-run |
| 🟡 **Risky** | `rm`, `mv`, `chmod`, `apt install`, `kill` | Require `y/N` confirmation |
| 🔴 **Dangerous** | `rm -rf /`, `sudo`, fork bombs, `dd`, `shutdown` | **Blocked** unless `--force` (requires typing `yes`) |

All scripts are checked before execution regardless of thinking mode.

---

## Architecture

```
shellper/
├── cmd/              # CLI commands (ask, explain, qa, session, REPL)
├── internal/
│   ├── config/       # YAML config + env var overrides
│   ├── llm/          # Ollama + OpenAI backends with streaming
│   ├── safety/       # 3-tier pattern-based checker
│   ├── executor/     # Shell command execution
│   └── shell/        # Shell auto-detection
├── main.go
├── go.mod
└── PLAN.md
```

---

## Backends

### Ollama (default)

```bash
ollama pull llama3.2     # or qwen2.5-coder:7b for better code
```

### OpenAI-compatible (OpenRouter, etc.)

```yaml
# config.yaml
backend: openai
model: gpt-4o-mini
openai_key: sk-...       # or export SHELLPER_OPENAI_KEY=sk-...
openai_base: https://openrouter.ai/api/v1
```

---

## Development

```bash
go build -o shellper .   # Build
go test ./...             # Run tests (21 tests)
go vet ./...              # Lint
```

---

## License

MIT © [Alex171012](LICENSE)

---

Built with [DeepSeek-v4-Flash](https://deepseek.com) — the AI that designed, implemented, and documented every line of Shellper.
