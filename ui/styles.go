package ui

import "github.com/charmbracelet/lipgloss"

// Palette.
var (
	AccentColor = lipgloss.Color("#00ADD8")
	GoodColor   = lipgloss.Color("#00D26A")
	BadColor    = lipgloss.Color("#FF5F5F")
	WarnColor   = lipgloss.Color("#F5C211")
	MutedColor  = lipgloss.Color("#9CA3AF")
)

var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor)

	SubtitleStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(MutedColor)

	MenuItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	MenuSelectedStyle = lipgloss.NewStyle().
				PaddingLeft(1).
				Bold(true).
				Background(AccentColor).
				Foreground(lipgloss.Color("#000000"))

	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(1, 2)

	StatusBarStyle = lipgloss.NewStyle().
			Background(AccentColor).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(BadColor)

	SuccessStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(GoodColor)

	HelpStyle = lipgloss.NewStyle().
			Faint(true).
			Foreground(MutedColor)

	LabelStyle = lipgloss.NewStyle().
			Foreground(MutedColor).
			Width(20)

	ValueStyle = lipgloss.NewStyle().
			Foreground(AccentColor)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(AccentColor).
			MarginBottom(1)

	TabStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(MutedColor)

	TabActiveStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Background(AccentColor).
			Foreground(lipgloss.Color("#000000"))

	OverlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(1, 2)

	KeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(AccentColor)
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
