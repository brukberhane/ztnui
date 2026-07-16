package ui

import "github.com/charmbracelet/lipgloss"

// Styles use terminal default colors (bold/faint/reverse) so the TUI
// inherits the user's terminal theme.

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true)

	SubtitleStyle = lipgloss.NewStyle().
			Faint(true)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	MenuSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Bold(true).
				Reverse(true)

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Reverse(true).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Reverse(true)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Faint(true)

	LabelStyle = lipgloss.NewStyle().
			Faint(true).
			Width(20)

	ValueStyle = lipgloss.NewStyle()

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			MarginBottom(1)

	TabStyle = lipgloss.NewStyle().
			Padding(0, 1)

	TabActiveStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Reverse(true)
)

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func joinStrings(ss []string) string {
	if len(ss) == 0 {
		return "-"
	}
	out := ss[0]
	for i := 1; i < len(ss); i++ {
		out += ", " + ss[i]
	}
	return out
}
