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
	modeStr := modeStyle.Render(strings.ToUpper(" " + m.modeName() + " "))
	statusStr := statusStyle.Render(" " + m.statusText() + " ")
	personaStr := headerInfoStyle.Render(" " + m.persona + " ")

	right := lipgloss.JoinHorizontal(lipgloss.Left, statusStr, personaStr)

	header := lipgloss.JoinHorizontal(lipgloss.Center,
		" Shellper ", modeStr, right,
	)
	header = headerStyle.Render(header)
	header = lipgloss.NewStyle().Width(m.width).Render(header)

	return header
}

func (m model) renderMessages() string {
	if len(m.messages) == 0 {
		return lipgloss.NewStyle().Padding(2, 2).
			Foreground(lipgloss.Color("#6272A4")).
			Render(`Welcome to Shellper!

  Type a message and press Enter.
  /help for commands.

Modes:
  ask      Generate and run scripts
  explain  Scripts with explanations
  qa       Linux Q&A (no execution)`)
	}

	var b strings.Builder
	for _, msg := range m.messages {
		var roleColor string
		switch msg.role {
		case "user":
			roleColor = "#50FA7B"
		case "assistant":
			roleColor = "#BD93F9"
		case "system":
			roleColor = "#6272A4"
		case "error":
			roleColor = "#FF5555"
		}

		author := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(roleColor)).Render(msg.role + ":")
		b.WriteString(author + "\n")
		b.WriteString(messageContentStyle.Render(msg.content) + "\n")

		if msg.script != "" {
			b.WriteString(scriptLabelStyle.Render("─── Script ───") + "\n")
			b.WriteString(messageContentStyle.Render(msg.script) + "\n")
		}

		if msg.output != "" {
			b.WriteString(outputLabelStyle.Render("─── Output ───") + "\n")
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
	modeTag := modeStyle.Render("[" + m.modeName() + "]")

	var prompt string
	if strings.HasPrefix(m.input, "/") {
		prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Render(m.input)
	} else {
		prompt = m.input
	}

	var confirmStr string
	if m.status == statusConfirming {
		confirmStr = warningStyle.Render(" Run? (y/N) ")
	} else if m.status == statusGenerating {
		confirmStr = statusStyle.Render(" Generating... ")
	}

	display := modeTag + " " + prompt
	if strings.HasPrefix(m.input, "/") {
		display += "█"
	} else if confirmStr != "" {
		display += confirmStr
	} else {
		display += "█"
	}

	display = lipgloss.NewStyle().Width(m.width - 4).Render(display)
	return inputStyle.Width(m.width - 2).Render(display)
}

func (m model) renderScriptPanel() string {
	if m.script == "" || m.scriptPanel != panelExpanded {
		return ""
	}

	content := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width - 8).
		Render(m.script)

	header := scriptLabelStyle.Render("📜 Script")
	border := panelBorderStyle.
		Width(m.width - 4).
		Render(header + "\n" + content)

	return border
}

func (m model) renderOutputPanel() string {
	if m.output == "" || m.outputPanel != panelExpanded {
		return ""
	}

	content := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width - 8).
		Render(m.output)

	header := outputLabelStyle.Render("📟 Output")
	border := panelBorderStyle.
		Width(m.width - 4).
		Render(header + "\n" + content)

	return border
}
