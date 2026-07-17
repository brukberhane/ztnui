package ui

import (
	"fmt"
	"strings"
)

type helpSection struct {
	title string
	rows  [][2]string
}

var helpSections = []helpSection{
	{
		title: "Global",
		rows: [][2]string{
			{"q / ctrl+c", "quit"},
			{"esc", "back / close"},
			{"h", "help"},
			{"n", "node info"},
			{",", "settings"},
			{"tab / H / L", "switch tabs"},
			{"r", "refresh"},
		},
	},
	{
		title: "Client networks",
		rows: [][2]string{
			{"↑/↓ / j/k", "navigate"},
			{"l / enter", "detail"},
			{"+", "join network"},
			{"x", "leave network"},
			{"p", "peers"},
		},
	},
	{
		title: "Client detail",
		rows: [][2]string{
			{"↑/↓ / j/k", "move"},
			{"space / l / enter", "toggle"},
			{"d", "allow DNS"},
			{"g", "allow default route"},
			{"G", "allow global IPs"},
			{"m", "allow managed IPs"},
		},
	},
	{
		title: "Server networks",
		rows: [][2]string{
			{"↑/↓ / j/k", "navigate"},
			{"l / enter", "detail"},
			{"c", "create network"},
			{"e", "edit network"},
			{"d", "delete network"},
			{"m", "members"},
		},
	},
	{
		title: "Server detail",
		rows: [][2]string{
			{"e", "edit network"},
			{"R", "rules presets"},
			{"m", "members"},
			{"d", "delete network"},
		},
	},
	{
		title: "Members",
		rows: [][2]string{
			{"↑/↓ / j/k", "navigate"},
			{"l / enter", "detail"},
			{"+", "authorize node by ID"},
			{"H", "hide / unhide (local)"},
			{"t", "toggle show hidden"},
			{"a", "toggle auth"},
			{"b", "toggle bridge"},
			{"o", "toggle auto-IP"},
			{"i", "IP assignments"},
			{"r", "rename"},
			{"delete", "deauth + remove (node must leave to stay gone)"},
		},
	},
	{
		title: "Forms",
		rows: [][2]string{
			{"tab / shift+tab", "next / previous field"},
			{"space", "toggle checkbox"},
			{"enter / ctrl+s", "submit"},
			{"esc", "cancel"},
		},
	},
}

func (m Model) viewHelpOverlay() string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render("Help"))
	b.WriteString("\n\n")
	for _, sec := range helpSections {
		b.WriteString(SubtitleStyle.Render(sec.title))
		b.WriteString("\n")
		for _, r := range sec.rows {
			b.WriteString(fmt.Sprintf("  %-18s %s\n", KeyStyle.Render(r[0]), HelpStyle.Render(r[1])))
		}
		b.WriteString("\n")
	}
	b.WriteString(HelpStyle.Render("esc/q/h close"))
	return b.String()
}
