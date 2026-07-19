package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"shellper/internal/llm"
)

type mode int

const (
	modeAsk mode = iota
	modeExplain
	modeQA
)

type status int

const (
	statusReady status = iota
	statusGenerating
	statusConfirming
	statusExecuting
)

type messageEntry struct {
	role     string
	content  string
	time     time.Time
	script   string
	output   string
	exitCode int
}

type panelState int

const (
	panelHidden panelState = iota
	panelCollapsed
	panelExpanded
)

type appConfig struct {
	model        string
	safetyMode   string
	shellCmd     string
	temperature  float64
	maxAutoFix   int
	renderMD     bool
	thinkEnabled bool
	reviewAll    bool
}

type model struct {
	appCtx    context.Context
	cancel    context.CancelFunc
	llmClient llm.Client
	cfg       *appConfig

	currentMode mode
	status      status
	persona     string

	messages  []messageEntry
	messageVP viewport.Model

	script      string
	scriptPanel panelState
	scriptVP    viewport.Model

	output      string
	outputPanel panelState

	input string

	width  int
	height int

	llmHistory    []llm.Message
	sessionName   string
	loadedSession []llm.Message
}

func initialModel(client llm.Client, cfg *appConfig, sessionName string, loadedMsgs []llm.Message) model {
	ctx, cancel := context.WithCancel(context.Background())
	m := model{
		appCtx:        ctx,
		cancel:        cancel,
		llmClient:     client,
		cfg:           cfg,
		currentMode:   modeAsk,
		status:        statusReady,
		persona:       "default",
		scriptPanel:   panelHidden,
		outputPanel:   panelHidden,
		sessionName:   sessionName,
		loadedSession: loadedMsgs,
	}

	if len(loadedMsgs) > 0 {
		for _, msg := range loadedMsgs {
			m.messages = append(m.messages, messageEntry{
				role: msg.Role, content: msg.Content, time: time.Now(),
			})
		}
		m.llmHistory = loadedMsgs
	}

	m.messageVP = viewport.New(80, 20)
	m.scriptVP = viewport.New(80, 8)

	return m
}

func (m *model) modeName() string {
	switch m.currentMode {
	case modeAsk:
		return "ask"
	case modeExplain:
		return "explain"
	case modeQA:
		return "qa"
	}
	return "ask"
}

func (m *model) statusText() string {
	switch m.status {
	case statusReady:
		return "ready"
	case statusGenerating:
		return "generating..."
	case statusConfirming:
		return "confirm?"
	case statusExecuting:
		return "executing..."
	}
	return ""
}
