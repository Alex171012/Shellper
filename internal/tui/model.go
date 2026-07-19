package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	streaming bool
}

type panelState int

const (
	panelHidden panelState = iota
	panelCollapsed
	panelExpanded
)

type cmdItem struct {
	name        string
	args        string
	description string
}

var cmdMenuItems = []cmdItem{
	{name: "mode", args: "ask|explain|qa", description: "Switch operation mode"},
	{name: "persona", args: "default|beginner|expert", description: "Set AI persona"},
	{name: "save", args: "[name]", description: "Save current session"},
	{name: "load", args: "<name>", description: "Load a saved session"},
	{name: "clear", args: "", description: "Clear conversation"},
	{name: "help", args: "", description: "Show help"},
	{name: "quit", args: "", description: "Exit Shellper"},
}

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
	program   *tea.Program
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

	input        string
	streamContent string

	cmdMenuShow   bool
	cmdMenuFilter string
	cmdMenuSel    int

	width  int
	height int

	llmHistory    []llm.Message
	sessionName   string
	loadedSession []llm.Message

	streamingMsgIdx int
}

func initialModel(client llm.Client, cfg *appConfig, sessionName string, loadedMsgs []llm.Message) *model {
	ctx, cancel := context.WithCancel(context.Background())
	m := &model{
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
		streamingMsgIdx: -1,
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
