package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	messages := m.renderMessages()
	input := m.renderInput()
	script := m.renderScriptPanel()
	output := m.renderOutputPanel()

	m.messageVP.SetContent(messages)

	body := m.messageVP.View()

	if m.scriptPanel == panelExpanded && m.script != "" {
		body += "\n" + script
	}

	if m.outputPanel == panelExpanded && m.output != "" {
		body += "\n" + output
	}

	components := []string{header, body, input}
	return lipgloss.JoinVertical(lipgloss.Left, components...)
}

func (m model) renderHeader() string {
	modeStr := " " + m.modeName()
	modeStr = modeStyle.Render(strings.ToUpper(modeStr))

	statusStr := " " + m.statusText()
	statusStr = statusStyle.Render(statusStr)

	personaStr := " " + m.persona
	personaStr = headerInfoStyle.Render(personaStr)

	right := lipgloss.JoinHorizontal(lipgloss.Left,
		statusStr,
		" │ ", personaStr,
	)

	left := lipgloss.JoinHorizontal(lipgloss.Left,
		" Shellper ", modeStr,
	)

	header := lipgloss.JoinHorizontal(lipgloss.Center, left, right)
	header = headerStyle.Render(header)
	header = lipgloss.NewStyle().Width(m.width).Render(header)

	return header
}

func (m model) renderMessages() string {
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().Padding(2, 2).
			Foreground(lipgloss.Color("#6272A4")).
			Render("Welcome to Shellper!\n\nType a message to start.\n\nModes:\n  ask      Generate and run scripts\n  explain  Scripts with explanations\n  qa       Linux Q&A (no execution)\n\nCommands:\n  :help    Show all commands")
	}

	var b strings.Builder
	for _, msg := range m.messages {
		author := messageAuthorStyle.Render(msg.role + ":")
		b.WriteString(author + "\n")
		b.WriteString(messageContentStyle.Render(msg.content) + "\n")

		if msg.script != "" {
			scriptLabel := scriptLabelStyle.Render("─── Script ───")
			b.WriteString(scriptLabel + "\n")
			b.WriteString(messageContentStyle.Render(msg.script) + "\n")
		}

		if msg.output != "" {
			outputLabel := outputLabelStyle.Render("─── Output ───")
			b.WriteString(outputLabel + "\n")
			b.WriteString(messageContentStyle.Render(msg.output) + "\n")
		}

		if msg.exitCode != 0 {
			b.WriteString(errorStyle.Render(fmt.Sprintf("exit code: %d\n", msg.exitCode)))
		}

		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderInput() string {
	if m.commandMode {
		prompt := helpKeyStyle.Render(":") + m.commandBuf + "█"
		return inputStyle.Width(m.width - 2).Render(prompt)
	}

	var modeTag string
	switch m.currentMode {
	case modeAsk:
		modeTag = "[ask]"
	case modeExplain:
		modeTag = "[explain]"
	case modeQA:
		modeTag = "[qa]"
	}

	var confirmStr string
	if m.status == statusConfirming {
		confirmStr = warningStyle.Render(" Run? (y/N) ")
	} else if m.status == statusGenerating {
		confirmStr = statusStyle.Render(" Generating... ")
	}

	modeTag = modeStyle.Render(modeTag)

	var inputLine string
	if m.inputFocused {
		inputLine = modeTag + " " + m.input + "█"
	} else {
		if m.status == statusConfirming {
			inputLine = modeTag + " " + confirmStr
		} else {
			inputLine = modeTag + " " + m.input
		}
	}

	inputLine = lipgloss.NewStyle().Width(m.width - 4).Render(inputLine)

	return inputStyle.Width(m.width - 2).Render(inputLine)
}

func (m model) renderScriptPanel() string {
	if m.script == "" || m.scriptPanel != panelExpanded {
		return ""
	}

	lines := strings.Split(m.script, "\n")
	display := strings.Join(lines, "\n")

	content := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width - 8).
		Render(display)

	header := scriptLabelStyle.Render("📜 Script (J/K to scroll)")
	border := panelBorderStyle.
		Width(m.width - 4).
		Render(header + "\n" + content)

	return border
}

func (m model) renderOutputPanel() string {
	if m.output == "" || m.outputPanel != panelExpanded {
		return ""
	}

	lines := strings.Split(m.output, "\n")
	display := strings.Join(lines, "\n")

	content := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width - 8).
		Render(display)

	header := outputLabelStyle.Render("📟 Output")
	border := panelBorderStyle.
		Width(m.width - 4).
		Render(header + "\n" + content)

	return border
}
