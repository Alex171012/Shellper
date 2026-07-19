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

type streamChunkMsg struct {
	chunk string
	idx   int
}
type streamDoneMsg struct {
	response string
	idx      int
	err      error
}
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
		return m, nil

	case tea.KeyMsg:
		return m, m.handleKey(msg)

	case streamChunkMsg:
		m.handleStreamChunk(msg)
		return m, nil

	case streamDoneMsg:
		return m.handleStreamDone(msg)

	case execDoneMsg:
		return m.handleExecDone(msg)
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
	if m.cmdMenuShow {
		items := filteredCmds(m)
		count := len(items)
		if count > 7 {
			count = 7
		}
		h -= count + 2
	}
	if h < 5 {
		h = 5
	}
	return h
}

func filteredCmds(m *model) []cmdItem {
	if m.cmdMenuFilter == "" {
		return cmdMenuItems
	}
	var f []cmdItem
	for _, c := range cmdMenuItems {
		if strings.Contains(c.name, m.cmdMenuFilter) {
			f = append(f, c)
		}
	}
	return f
}

func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c":
		if m.status == statusGenerating || m.status == statusExecuting {
			m.cancel()
			m.finishStreamingMsg("cancelled")
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
		if m.cmdMenuShow {
			return m.selectCmdMenu()
		}
		if m.status == statusConfirming {
			return m.execConfirm(true)
		}
		if m.status != statusReady {
			return nil
		}
		return m.handleEnter()

	case "up":
		if m.cmdMenuShow {
			items := filteredCmds(m)
			if len(items) > 0 {
				m.cmdMenuSel = (m.cmdMenuSel - 1 + len(items)) % len(items)
			}
			return nil
		}
		m.messageVP.LineUp(1)
		return nil

	case "down":
		if m.cmdMenuShow {
			items := filteredCmds(m)
			if len(items) > 0 {
				m.cmdMenuSel = (m.cmdMenuSel + 1) % len(items)
			}
			return nil
		}
		m.messageVP.LineDown(1)
		return nil

	case "escape":
		if m.cmdMenuShow {
			m.cmdMenuShow = false
			m.cmdMenuFilter = ""
			return nil
		}
		if m.status == statusConfirming {
			return m.execConfirm(false)
		}
		return nil

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
			if m.cmdMenuShow {
				if strings.HasPrefix(m.input, "/") {
					m.cmdMenuFilter = m.input[1:]
				} else {
					m.cmdMenuShow = false
					m.cmdMenuFilter = ""
				}
				m.cmdMenuSel = 0
			}
		}
		return nil

	case "ctrl+w":
		idx := strings.LastIndex(strings.TrimRight(m.input, " "), " ")
		if idx < 0 {
			m.input = ""
		} else {
			m.input = m.input[:idx+1]
		}
		m.cmdMenuShow = false
		return nil

	case "ctrl+u":
		m.input = ""
		m.cmdMenuShow = false
		return nil
	}

	if m.status == statusConfirming {
		if msg.String() == "y" || msg.String() == "Y" {
			return m.execConfirm(true)
		}
		if msg.String() == "n" || msg.String() == "N" {
			return m.execConfirm(false)
		}
		return nil
	}

	if len(msg.String()) == 1 {
		ch := msg.String()
		if !m.cmdMenuShow && ch == "/" && m.input == "" {
			m.cmdMenuShow = true
			m.cmdMenuFilter = ""
			m.cmdMenuSel = 0
		}
		m.input += ch
		if m.cmdMenuShow {
			m.cmdMenuFilter = m.input[1:]
			m.cmdMenuSel = 0
		}
	}

	return nil
}

func (m *model) selectCmdMenu() tea.Cmd {
	items := filteredCmds(m)
	if len(items) == 0 {
		return nil
	}
	if m.cmdMenuSel >= len(items) {
		m.cmdMenuSel = 0
	}
	sel := items[m.cmdMenuSel]

	if sel.args != "" {
		m.input = "/" + sel.name + " "
	} else {
		m.input = "/" + sel.name
	}
	m.cmdMenuShow = false
	m.cmdMenuFilter = ""
	return nil
}

func (m *model) handleEnter() tea.Cmd {
	text := strings.TrimSpace(m.input)
	if text == "" {
		return nil
	}

	m.input = ""
	m.cmdMenuShow = false

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

func (m *model) execConfirm(approve bool) tea.Cmd {
	if !approve {
		m.messages = append(m.messages, messageEntry{
			role: "system", content: "Cancelled.", time: time.Now(),
		})
		m.script = ""
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return nil
	}

	m.status = statusExecuting
	return m.executeScript()
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
		var b strings.Builder
		b.WriteString("Commands:\n")
		for _, c := range cmdMenuItems {
			args := c.args
			if args != "" {
				args = " " + args
			}
			b.WriteString(fmt.Sprintf("  /%s%s  — %s\n", c.name, args, c.description))
		}
		b.WriteString("\nKeys:\n  PgUp/PgDn  Scroll  Tab  Cycle panels\n  Ctrl+C  Cancel/Quit  Ctrl+W  Del word\n  /  Show command menu")
		m.messages = append(m.messages, messageEntry{
			role: "system", content: b.String(), time: time.Now(),
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
	m.streamingMsgIdx = -1

	messages := buildTUIPrompt(query, m.modeName(), buildSystemCtx(), m.currentMode == modeExplain || m.cfg.thinkEnabled, m.persona)
	m.llmHistory = append(m.llmHistory, messages...)

	idx := len(m.messages)
	m.streamingMsgIdx = idx
	m.messages = append(m.messages, messageEntry{
		role: "assistant", content: "", time: time.Now(), streaming: true,
	})
	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()

	prog := m.program
	client := m.llmClient
	modelName := m.cfg.model
	temp := m.cfg.temperature
	ctx := m.appCtx
	histCopy := make([]llm.Message, len(m.llmHistory))
	copy(histCopy, m.llmHistory)

	return func() tea.Msg {
		var fullContent strings.Builder
		onChunk := llm.StreamHandler(func(chunk string) error {
			fullContent.WriteString(chunk)
			if prog != nil {
				prog.Send(streamChunkMsg{chunk: chunk, idx: idx})
			}
			return nil
		})

		opts := &llm.ChatOptions{Model: modelName, Temperature: temp}
		resp, err := client.ChatStream(ctx, histCopy, opts, onChunk)
		if err != nil {
			return streamDoneMsg{idx: idx, err: err}
		}
		if len(resp.Choices) > 0 {
			return streamDoneMsg{
				response: strings.TrimSpace(resp.Choices[0].Message.Content),
				idx:      idx,
			}
		}
		return streamDoneMsg{idx: idx, err: fmt.Errorf("empty response")}
	}
}

func (m *model) startQA(query string) tea.Cmd {
	m.status = statusGenerating

	messages := buildQARequest(query, buildSystemCtx())
	m.llmHistory = append(m.llmHistory, messages...)

	idx := len(m.messages)
	m.streamingMsgIdx = idx
	m.messages = append(m.messages, messageEntry{
		role: "assistant", content: "", time: time.Now(), streaming: true,
	})
	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()

	prog := m.program
	client := m.llmClient
	modelName := m.cfg.model
	temp := m.cfg.temperature
	ctx := m.appCtx
	histCopy := make([]llm.Message, len(m.llmHistory))
	copy(histCopy, m.llmHistory)

	return func() tea.Msg {
		var fullContent strings.Builder
		onChunk := llm.StreamHandler(func(chunk string) error {
			fullContent.WriteString(chunk)
			if prog != nil {
				prog.Send(streamChunkMsg{chunk: chunk, idx: idx})
			}
			return nil
		})

		opts := &llm.ChatOptions{Model: modelName, Temperature: temp}
		resp, err := client.ChatStream(ctx, histCopy, opts, onChunk)
		if err != nil {
			return streamDoneMsg{idx: idx, err: err}
		}
		if len(resp.Choices) > 0 {
			return streamDoneMsg{
				response: strings.TrimSpace(resp.Choices[0].Message.Content),
				idx:      idx,
			}
		}
		return streamDoneMsg{idx: idx, err: fmt.Errorf("empty response")}
	}
}

func (m *model) handleStreamChunk(msg streamChunkMsg) {
	if msg.idx >= 0 && msg.idx < len(m.messages) {
		m.messages[msg.idx].content += msg.chunk
		m.messages[msg.idx].streaming = true
	}
	if m.messageVP.AtBottom() {
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
	}
}

func (m *model) finishStreamingMsg(endText string) {
	if m.streamingMsgIdx >= 0 && m.streamingMsgIdx < len(m.messages) {
		m.messages[m.streamingMsgIdx].streaming = false
		if m.messages[m.streamingMsgIdx].content == "" {
			m.messages[m.streamingMsgIdx].content = endText
		}
	}
	m.streamingMsgIdx = -1
}

func (m *model) handleStreamDone(msg streamDoneMsg) (tea.Model, tea.Cmd) {
	m.finishStreamingMsg("")

	if msg.err != nil {
		m.messages = append(m.messages, messageEntry{
			role: "error", content: msg.err.Error(), time: time.Now(),
		})
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return m, nil
	}

	fullContent := strings.TrimSpace(msg.response)
	if msg.idx >= 0 && msg.idx < len(m.messages) {
		m.messages[msg.idx].content = fullContent
	}
	m.llmHistory = append(m.llmHistory,
		llm.Message{Role: "assistant", Content: fullContent},
	)

	if m.currentMode == modeQA {
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return m, nil
	}

	script := extractTUIScript(fullContent)
	if strings.TrimSpace(script) == "" {
		m.status = statusReady
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return m, nil
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
		m.messageVP.SetContent(m.renderMessages())
		m.messageVP.GotoBottom()
		return m, m.executeScript()
	}

	m.messageVP.SetContent(m.renderMessages())
	m.messageVP.GotoBottom()
	return m, nil
}

func (m *model) handleExecDone(msg execDoneMsg) (tea.Model, tea.Cmd) {
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
	return m, nil
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
		return execDoneMsg{result: result, script: m.script}
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
