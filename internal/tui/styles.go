package tui

import "github.com/charmbracelet/lipgloss"

var (
	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7B56DB")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	headerInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))
	modeTag    = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	statusTag  = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))

	userLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#50FA7B")).
			Render

	assistantLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#BD93F9")).
			Render

	systemLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6272A4")).
			Render

	errorLabel = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555")).
			Render

	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#44475A")).
			Render

	codeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#282A36")).
			Padding(0, 2).
			Margin(0, 2)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#6272A4")).
			BorderTop(true).
			Padding(0, 1)

	panelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#44475A")).
			Padding(0, 1)

	scriptLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C")).
			Bold(true)

	outputLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	welcomeStyle = lipgloss.NewStyle().
			Padding(2, 3).
			Foreground(lipgloss.Color("#6272A4"))

)
