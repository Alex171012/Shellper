package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"shellper/internal/executor"
	"shellper/internal/llm"
	"shellper/internal/safety"
)

type errMsg struct{ err error }
type streamDoneMsg struct{ response string }
type execDoneMsg struct {
	result *executor.Result
	script string
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.messageVP.Width = msg.Width - 4
		m.messageVP.Height = m.calcMessagesHeight()
		m.scriptVP.Width = msg.Width - 6
		m.scriptVP.Height = m.calcScriptHeight()
		return m, nil

	case tea.KeyMsg:
		return m, m.handleKey(msg)

	case errMsg:
		m.messages = append(m.messages, messageEntry{
			role: "error", content: msg.err.Error(), time: time.Now(),
		})
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return m, nil

	case streamDoneMsg:
		*m = m.handleStreamDone(msg)
		return m, nil

	case execDoneMsg:
		*m = m.handleExecDone(msg)
		return m, nil
	}

	return m, nil
}

func (m *model) calcMessagesHeight() int {
	h := m.height
	if m.outputPanel == panelExpanded {
		h -= 6
	}
	if m.scriptPanel == panelExpanded {
		h -= 10
	}
	h -= 4
	if h < 5 {
		h = 5
	}
	return h
}

func (m *model) calcScriptHeight() int {
	if m.scriptPanel == panelExpanded {
		return 8
	}
	return 0
}

func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		if m.status == statusGenerating || m.status == statusExecuting {
			m.status = statusReady
			return nil
		}
		return tea.Quit

	case "pgup":
		m.messageVP.HalfViewUp()
		return nil
	case "pgdown":
		m.messageVP.HalfViewDown()
		return nil

	case "tab":
		if m.scriptPanel == panelExpanded {
			m.scriptPanel = panelCollapsed
			m.outputPanel = panelExpanded
		} else if m.outputPanel == panelExpanded {
			m.outputPanel = panelCollapsed
		} else {
			m.scriptPanel = panelExpanded
		}
		return nil

	case "enter":
		if m.status != statusReady {
			return nil
		}
		return m.handleEnter()

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return nil

	case "ctrl+w":
		idx := strings.LastIndex(strings.TrimRight(m.input, " "), " ")
		if idx < 0 {
			m.input = ""
		} else {
			m.input = m.input[:idx+1]
		}
		return nil

	case "ctrl+u":
		m.input = ""
		return nil
	}

	if m.status == statusConfirming {
		return m.handleConfirm(msg)
	}

	if len(msg.String()) == 1 {
		m.input += msg.String()
	}

	return nil
}

func (m *model) handleEnter() tea.Cmd {
	text := strings.TrimSpace(m.input)
	if text == "" {
		return nil
	}

	m.input = ""

	if strings.HasPrefix(text, "/") {
		return m.execCommand(text[1:])
	}

	m.messages = append(m.messages, messageEntry{
		role: "user", content: text, time: time.Now(),
	})
	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()

	if m.currentMode == modeQA {
		return m.startQA(text)
	}
	return m.startScriptGeneration(text)
}

func (m *model) handleConfirm(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "y", "Y":
		m.status = statusExecuting
		return m.executeScript()
	case "n", "N", "escape":
		m.messages = append(m.messages, messageEntry{
			role: "system", content: "Cancelled.", time: time.Now(),
		})
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return nil
	}
	return nil
}

func (m *model) execCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "quit", "exit", "q":
		return tea.Quit

	case "help":
		m.messages = append(m.messages, messageEntry{
			role: "system",
			content: `Commands:
  /help              Show this help
  /mode ask|explain|qa   Switch mode
  /persona <name>    Set persona (default, beginner, expert)
  /save [name]       Save session
  /load <name>       Load session
  /clear             Clear conversation
  /quit              Quit

Modes:
  ask      Fast script generation
  explain  Script with plan + explanation
  qa       Q&A (no execution)

Keys:
  PgUp/PgDn   Scroll messages
  Tab         Cycle panels (script → output → hide)
  Ctrl+C      Cancel / Quit
  Ctrl+W      Delete word
  Ctrl+U      Clear line`,
			time: time.Now(),
		})

	case "mode":
		if len(parts) > 1 {
			switch parts[1] {
			case "ask":
				m.currentMode = modeAsk
			case "explain":
				m.currentMode = modeExplain
			case "qa":
				m.currentMode = modeQA
			}
		}

	case "persona":
		if len(parts) > 1 {
			m.persona = parts[1]
		}

	case "clear":
		m.messages = nil
		m.llmHistory = nil
		m.script = ""
		m.output = ""
		m.scriptPanel = panelHidden
		m.outputPanel = panelHidden

	case "save":
		name := "unnamed"
		if len(parts) > 1 {
			name = parts[1]
		}
		if err := saveTUISession(name, m.llmHistory); err != nil {
			m.messages = append(m.messages, messageEntry{
				role: "error", content: "Save failed: " + err.Error(), time: time.Now(),
			})
		} else {
			m.messages = append(m.messages, messageEntry{
				role: "system", content: "Session saved: " + name, time: time.Now(),
			})
		}

	case "load":
		if len(parts) > 1 {
			msgs, err := loadTUISession(parts[1])
			if err != nil {
				m.messages = append(m.messages, messageEntry{
					role: "error", content: "Load failed: " + err.Error(), time: time.Now(),
				})
			} else {
				m.messages = nil
				m.llmHistory = msgs
				for _, msg := range msgs {
					m.messages = append(m.messages, messageEntry{
						role: msg.Role, content: msg.Content, time: time.Now(),
					})
				}
				m.messages = append(m.messages, messageEntry{
					role: "system", content: "Session loaded: " + parts[1], time: time.Now(),
				})
			}
		}
	}

	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()
	return nil
}

func (m *model) startScriptGeneration(query string) tea.Cmd {
	m.status = statusGenerating
	m.scriptPanel = panelHidden
	m.outputPanel = panelHidden

	messages := buildTUIPrompt(query, m.modeName(), buildSystemCtx(), m.currentMode == modeExplain || m.cfg.thinkEnabled, m.persona)
	m.llmHistory = append(m.llmHistory, messages...)

	return func() tea.Msg {
		var fullResponse strings.Builder

		onChunk := llm.StreamHandler(func(chunk string) error {
			fullResponse.WriteString(chunk)
			return nil
		})

		opts := &llm.ChatOptions{
			Model:       m.cfg.model,
			Temperature: m.cfg.temperature,
		}

		resp, err := m.llmClient.ChatStream(m.appCtx, m.llmHistory, opts, onChunk)
		if err != nil {
			return errMsg{err: err}
		}
		if len(resp.Choices) > 0 {
			content := strings.TrimSpace(resp.Choices[0].Message.Content)
			m.llmHistory = append(m.llmHistory,
				llm.Message{Role: "assistant", Content: content},
			)
			return streamDoneMsg{response: content}
		}
		return errMsg{err: fmt.Errorf("empty response")}
	}
}

func (m *model) startQA(query string) tea.Cmd {
	m.status = statusGenerating

	messages := buildQARequest(query, buildSystemCtx())
	m.llmHistory = append(m.llmHistory, messages...)

	return func() tea.Msg {
		var fullResponse strings.Builder

		onChunk := llm.StreamHandler(func(chunk string) error {
			fullResponse.WriteString(chunk)
			return nil
		})

		opts := &llm.ChatOptions{
			Model:       m.cfg.model,
			Temperature: m.cfg.temperature,
		}

		resp, err := m.llmClient.ChatStream(m.appCtx, m.llmHistory, opts, onChunk)
		if err != nil {
			return errMsg{err: err}
		}
		if len(resp.Choices) > 0 {
			content := strings.TrimSpace(resp.Choices[0].Message.Content)
			m.llmHistory = append(m.llmHistory,
				llm.Message{Role: "assistant", Content: content},
			)
			return streamDoneMsg{response: content}
		}
		return errMsg{err: fmt.Errorf("empty response")}
	}
}

func (m *model) handleStreamDone(msg streamDoneMsg) model {
	m.messages = append(m.messages, messageEntry{
		role: "assistant", content: msg.response, time: time.Now(),
	})

	if m.currentMode == modeQA {
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return *m
	}

	script := extractTUIScript(msg.response)
	if strings.TrimSpace(script) == "" {
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return *m
	}

	m.script = script
	m.scriptPanel = panelExpanded
	m.messageVP.Height = m.calcMessagesHeight()
	m.scriptVP.SetContent(m.script)
	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()

	safetyResult := safety.Check(script, m.cfg.safetyMode)
	if safetyResult.IsDangerous() {
		if m.cfg.safetyMode == "permissive" {
			m.status = statusConfirming
		} else {
			m.messages = append(m.messages, messageEntry{
				role: "system", content: "⚠ BLOCKED: Dangerous command. Use --force.", time: time.Now(),
			})
			m.status = statusReady
		}
	} else if safetyResult.IsRisky() || m.cfg.reviewAll {
		m.status = statusConfirming
	} else {
		m.status = statusExecuting
		return *m
	}

	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()
	return *m
}

func (m *model) handleExecDone(msg execDoneMsg) model {
	m.output = msg.result.Output
	m.outputPanel = panelExpanded

	if len(m.messages) > 0 {
		idx := len(m.messages) - 1
		m.messages[idx].script = msg.script
		m.messages[idx].output = msg.result.Output
		m.messages[idx].exitCode = msg.result.ExitCode
	}

	m.status = statusReady
	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()
	return *m
}

func (m *model) executeScript() tea.Cmd {
	return func() tea.Msg {
		result, err := executor.Run(m.script, m.cfg.shellCmd)
		if err != nil {
			return execDoneMsg{
				result: &executor.Result{Output: err.Error(), ExitCode: 1},
				script: m.script,
			}
		}
		return execDoneMsg{
			result: result, script: m.script,
		}
	}
}

func buildSystemCtx() string {
	return "System context for the assistant."
}

func buildQARequest(query string, sysCtx string) []llm.Message {
	return []llm.Message{
		{Role: "system", Content: "You are Shellper, a Linux/shell assistant. Answer concisely."},
		{Role: "user", Content: query},
	}
}
