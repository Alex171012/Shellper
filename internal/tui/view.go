package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	header := m.renderHeader()
	input := m.renderInput()

	m.messageVP.SetContent(m.renderMessages())

	body := m.messageVP.View()

	if m.messageVP.TotalLineCount() > m.messageVP.VisibleLineCount() {
		atTop := m.messageVP.YOffset <= 0
		atBottom := m.messageVP.YOffset+m.messageVP.VisibleLineCount() >= m.messageVP.TotalLineCount()
		if !atTop {
			body += "\n" + dimStyle.Padding(0, 2).Render("▲ PgUp")
		}
		if !atBottom {
			body += "\n" + dimStyle.Padding(0, 2).Render("▼ PgDn")
		}
	}

	if m.cmdMenuShow {
		body += "\n" + m.renderCmdMenu()
	}

	if m.status == statusConfirming {
		body += "\n" + m.renderConfirmPanel()
	}

	if m.scriptPanel == panelExpanded && m.script != "" {
		body += "\n" + m.renderScriptPanel()
	}

	if m.outputPanel == panelExpanded && m.output != "" {
		body += "\n" + m.renderOutputPanel()
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, body, input)
}

func (m *model) renderHeader() string {
	mode := modeTag.Render(strings.ToUpper(" " + m.modeName() + " "))
	status := statusTag.Render(" " + m.statusText() + " ")
	persona := headerInfo.Render(" " + m.persona + " ")
	right := lipgloss.JoinHorizontal(lipgloss.Left, status, persona)
	line := lipgloss.JoinHorizontal(lipgloss.Center, " Shellper ", mode, right)
	return headerStyle.Width(m.width).Render(line)
}

func (m *model) renderMessages() string {
	if len(m.messages) == 0 {
		return welcomeStyle.Render(
			"Welcome to Shellper\n\n" +
				"Type a message and press Enter.\n" +
				"Type / or press / to open the command menu.\n\n" +
				"  " + modeTag.Render("[ask]") + "  Generate and run scripts\n" +
				"  " + modeTag.Render("[explain]") + "  Scripts with explanations\n" +
				"  " + modeTag.Render("[qa]") + "  Linux Q&A (no execution)",
		)
	}

	var b strings.Builder
	for _, msg := range m.messages {
		label := m.messageLabel(msg)
		ts := timestampStyle(msg.time.Format("15:04"))
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, label, " ", ts) + "\n")

		switch msg.role {
		case "assistant":
			if msg.streaming {
				content := msg.content
				if content == "" {
					content = "▊"
				} else {
					content = content + "▊"
				}
				b.WriteString(content + "\n")
			} else {
				b.WriteString(m.renderAssistantContent(msg.content) + "\n")
			}
		case "error":
			b.WriteString(errorStyle.Render(msg.content) + "\n")
		default:
			b.WriteString(msg.content + "\n")
		}

		if msg.exitCode != 0 {
			b.WriteString(errorStyle.Render(fmt.Sprintf("  exit code: %d", msg.exitCode)) + "\n")
		}

		b.WriteString("\n")
	}

	return b.String()
}

func (m *model) messageLabel(msg messageEntry) string {
	switch msg.role {
	case "user":
		return userLabel("You")
	case "assistant":
		if msg.streaming {
			return statusTag.Render("Shellper ●")
		}
		return assistantLabel("Shellper")
	case "system":
		return systemLabel("System")
	case "error":
		return errorLabel("Error")
	}
	return msg.role
}

func (m *model) renderAssistantContent(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")
	inCodeBlock := false
	var codeBuf strings.Builder

	flushCode := func() {
		if codeBuf.Len() > 0 {
			highlighted := highlightShell(codeBuf.String())
			result.WriteString(codeBlockStyle.Render(highlighted))
			result.WriteString("\n")
			codeBuf.Reset()
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			if inCodeBlock {
				flushCode()
				inCodeBlock = false
			} else {
				flushCode()
				inCodeBlock = true
			}
			continue
		}
		if inCodeBlock {
			if codeBuf.Len() > 0 {
				codeBuf.WriteString("\n")
			}
			codeBuf.WriteString(line)
		} else {
			result.WriteString(line + "\n")
		}
	}

	if inCodeBlock {
		flushCode()
	}

	out := result.String()
	out = renderMarkdown(out)
	return out
}

func (m *model) renderCmdMenu() string {
	items := filteredCmds(m)
	if len(items) == 0 {
		return dimStyle.Padding(0, 2).Render("No matching commands")
	}

	var b strings.Builder
	b.WriteString(dimStyle.Padding(0, 2).Render("Commands:") + "\n")

	for i, item := range items {
		mark := "  "
		style := dimStyle
		if i == m.cmdMenuSel {
			mark = "▸ "
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
		}

		args := item.args
		if args != "" {
			args = " " + args
		}
		line := fmt.Sprintf("%s/%s%s  %s", mark, item.name, args, item.description)
		b.WriteString(style.Padding(0, 2).Render(line) + "\n")
	}

	width := m.width - 4
	if width < 40 {
		width = 40
	}
	return panelBorder.Width(width).Render(b.String())
}

func (m *model) renderConfirmPanel() string {
	if m.script == "" {
		return ""
	}

	hasSudo := strings.Contains(m.script, "sudo ")
	hasRm := strings.Contains(m.script, "rm ")
	hasChmod := strings.Contains(m.script, "chmod ")
	hasRedirect := strings.Contains(m.script, "> ") || strings.Contains(m.script, ">> ")

	var warnings []string
	if hasSudo {
		warnings = append(warnings, "🔒 Uses sudo — system will ask for password")
	}
	if hasRm {
		warnings = append(warnings, "⚠ Uses rm — files will be deleted")
	}
	if hasChmod {
		warnings = append(warnings, "⚠ Uses chmod — file permissions change")
	}
	if hasRedirect {
		warnings = append(warnings, "⚠ Uses output redirection — files may be overwritten")
	}

	highlighted := highlightShell(m.script)

	var b strings.Builder
	b.WriteString(scriptLabel.Render("═══ Confirm Script ═══") + "\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Margin(0, 1).Render(highlighted) + "\n")

	if len(warnings) > 0 {
		b.WriteString(warningStyle.Render("─── Warnings ───") + "\n")
		for _, w := range warnings {
			b.WriteString("  " + w + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(warningStyle.Render("  Run? (y/N)"))

	width := m.width - 4
	if width < 40 {
		width = 40
	}
	return panelBorder.Width(width).Render(b.String())
}

func (m *model) renderInput() string {
	tag := modeTag.Render("[" + m.modeName() + "]")

	switch m.status {
	case statusGenerating:
		return inputStyle.Width(m.width - 2).Render(
			tag + "  " + statusTag.Render("Generating…"),
		)
	case statusExecuting:
		return inputStyle.Width(m.width - 2).Render(
			tag + "  " + statusTag.Render("Executing…"),
		)
	case statusConfirming:
		return inputStyle.Width(m.width - 2).Render(
			tag + "  " + warningStyle.Render("y (run) / n (cancel)"),
		)
	}

	if m.cmdMenuShow {
		return inputStyle.Width(m.width - 2).Render(
			tag + "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Render(m.input+"█"),
		)
	}

	prompt := m.input
	if m.input == "" {
		prompt = dimStyle.Render("Type a message…  /  for menu")
	} else if strings.HasPrefix(m.input, "/") {
		prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Render(m.input)
	}

	display := tag + " " + prompt
	if m.input != "" && !strings.HasPrefix(m.input, "/") {
		display += "█"
	}
	display = lipgloss.NewStyle().Width(m.width - 4).Render(display)
	return inputStyle.Width(m.width - 2).Render(display)
}

func (m *model) renderScriptPanel() string {
	if m.script == "" || m.scriptPanel != panelExpanded {
		return ""
	}

	highlighted := highlightShell(m.script)
	content := lipgloss.NewStyle().Padding(0, 1).Width(m.width - 8).Render(highlighted)
	header := scriptLabel.Render("📜 Script (Tab to close)")
	return panelBorder.Width(m.width - 4).Render(header + "\n" + content)
}

func (m *model) renderOutputPanel() string {
	if m.output == "" || m.outputPanel != panelExpanded {
		return ""
	}

	content := lipgloss.NewStyle().Padding(0, 1).Width(m.width - 8).Render(m.output)
	header := outputLabel.Render("📟 Output (Tab to close)")
	return panelBorder.Width(m.width - 4).Render(header + "\n" + content)
}
