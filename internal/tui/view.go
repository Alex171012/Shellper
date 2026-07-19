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
	input := m.renderInput()

	m.messageVP.SetContent(m.renderMessages())

	body := m.messageVP.View()

	if m.status == statusConfirming {
		body += "\n" + m.renderConfirmPanel()
	}

	if m.scriptPanel == panelExpanded && m.script != "" {
		body += "\n" + m.renderScriptPanel()
	}

	if m.outputPanel == panelExpanded && m.output != "" {
		body += "\n" + m.renderOutputPanel()
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

		if msg.role == "assistant" && strings.Contains(msg.content, "```") {
			b.WriteString(messageContentStyle.Render(msg.content) + "\n")
		} else if msg.role == "assistant" {
			rendered := renderMarkdown(msg.content)
			b.WriteString(messageContentStyle.Render(rendered) + "\n")
		} else {
			b.WriteString(messageContentStyle.Render(msg.content) + "\n")
		}

		if msg.exitCode != 0 {
			b.WriteString(errorStyle.Render(fmt.Sprintf("exit code: %d\n", msg.exitCode)))
		}

		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderConfirmPanel() string {
	if m.script == "" {
		return ""
	}

	hasSudo := strings.Contains(m.script, "sudo ")
	hasRm := strings.Contains(m.script, "rm ")
	hasChmod := strings.Contains(m.script, "chmod ")

	var warnings []string
	if hasSudo {
		warnings = append(warnings, "🔒 Uses sudo — system will ask for password")
	}
	if hasRm {
		warnings = append(warnings, "⚠ Uses rm — files will be deleted")
	}
	if hasChmod {
		warnings = append(warnings, "⚠ Uses chmod — file permissions will change")
	}

	highlighted := highlightShell(m.script)

	var b strings.Builder
	b.WriteString(scriptLabelStyle.Render("═══ Script Preview ═══"))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(highlighted))
	b.WriteString("\n")

	if len(warnings) > 0 {
		b.WriteString(warningStyle.Render("─── Safety Warnings ───"))
		b.WriteString("\n")
		for _, w := range warnings {
			b.WriteString("  " + w + "\n")
		}
		b.WriteString("\n")
	}

	border := panelBorderStyle.
		Width(m.width - 4).
		Render(b.String() + warningStyle.Render("  Run? (y/N)  "))

	return border
}

func (m model) renderInput() string {
	modeTag := modeStyle.Render("[" + m.modeName() + "]")

	var prompt string
	if strings.HasPrefix(m.input, "/") {
		prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Render(m.input)
	} else {
		prompt = m.input
	}

	var suffix string
	if m.status == statusGenerating {
		suffix = statusStyle.Render(" Generating... ")
	} else if m.status == statusExecuting {
		suffix = statusStyle.Render(" Executing... ")
	} else if m.status == statusConfirming {
		suffix = ""
	} else {
		suffix = "█"
	}

	display := modeTag + " " + prompt + suffix

	if m.status == statusConfirming {
		display = modeTag + warningStyle.Render(" Press y (run) or n (cancel) ")
	}

	display = lipgloss.NewStyle().Width(m.width - 4).Render(display)
	return inputStyle.Width(m.width - 2).Render(display)
}

func (m model) renderScriptPanel() string {
	if m.script == "" || m.scriptPanel != panelExpanded {
		return ""
	}

	highlighted := highlightShell(m.script)

	content := lipgloss.NewStyle().
		Padding(0, 1).
		Width(m.width - 8).
		Render(highlighted)

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
