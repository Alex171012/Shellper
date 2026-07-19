package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7B56DB"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	danger    = lipgloss.Color("#FF5555")
	warning   = lipgloss.Color("#F1FA8C")
)

var (
	appStyle = lipgloss.NewStyle().
			Margin(0, 0)

	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#7B56DB")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 1)

	headerInfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB86C"))

	modeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD"))

	messageAuthorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#BD93F9"))

	messageContentStyle = lipgloss.NewStyle().
				Padding(0, 2)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#6272A4")).
			BorderTop(true).
			Padding(0, 1)

	panelBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#44475A")).
				Padding(0, 1)

	scriptLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F1FA8C")).
				Bold(true)

	outputLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")).
				Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(danger).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(warning)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtle).
			Padding(0, 2)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)
)
