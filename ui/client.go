package ui

import (
	"fmt"
	"strings"

	"github.com/brukberhane/ztnui/api"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type clientView int

const (
	clientViewNetworks clientView = iota
	clientViewPeers
	clientViewDetail
	clientViewJoin
)

type clientModel struct {
	view         clientView
	table        table.Model
	joinInput    textinput.Model
	selectedID   string
	detail       *api.Network
	settingFocus int
	width        int
	height       int
	tableHeight  int
}

func newClientModel() clientModel {
	c := clientModel{
		joinInput: newInput("16-char network ID"),
	}
	c.joinInput.CharLimit = 16
	return c
}

func (c *clientModel) setNetworks(networks []api.Network) {
	rows := make([]table.Row, 0, len(networks))
	for _, n := range networks {
		rows = append(rows, table.Row{
			n.ID,
			truncate(n.Name, 20),
			n.Status,
			n.Type,
			truncate(joinStrings(n.AssignedAddresses), 24),
		})
	}
	cols := []table.Column{
		{Title: "ID", Width: 18},
		{Title: "Name", Width: 20},
		{Title: "Status", Width: 10},
		{Title: "Type", Width: 10},
		{Title: "Addresses", Width: 24},
	}
	c.table = newTable(cols, rows, max(20, c.width-4), c.tableHeight)
}

func (c *clientModel) setPeers(peers []api.Peer) {
	rows := make([]table.Row, 0, len(peers))
	for _, p := range peers {
		pathCount := len(p.Paths)
		rows = append(rows, table.Row{
			p.Address,
			p.Role,
			p.Version,
			fmt.Sprintf("%d", p.Latency),
			fmt.Sprintf("%d", pathCount),
		})
	}
	cols := []table.Column{
		{Title: "Address", Width: 12},
		{Title: "Role", Width: 8},
		{Title: "Version", Width: 10},
		{Title: "Latency", Width: 8},
		{Title: "Paths", Width: 6},
	}
	c.table = newTable(cols, rows, max(20, c.width-4), c.tableHeight)
}

func (c *clientModel) resize(width, contentHeight int) {
	c.width = width
	c.tableHeight = max(3, contentHeight-6)
	if len(c.table.Rows()) > 0 {
		c.table.SetHeight(c.tableHeight)
		c.table.SetWidth(max(20, width-4))
	}
}

func (c *clientModel) selectedNetworkID() string {
	if c.view == clientViewNetworks && len(c.table.Rows()) > 0 {
		row := c.table.SelectedRow()
		if len(row) > 0 {
			return row[0]
		}
	}
	return c.selectedID
}

func (c *clientModel) Update(msg tea.Msg) (clientModel, tea.Cmd) {
	var cmd tea.Cmd
	switch c.view {
	case clientViewJoin:
		c.joinInput, cmd = c.joinInput.Update(msg)
	case clientViewNetworks, clientViewPeers:
		c.table, cmd = c.table.Update(msg)
	}
	return *c, cmd
}

func (c *clientModel) View(status *api.Status) string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render("Client"))
	b.WriteString("\n")

	switch c.view {
	case clientViewNetworks:
		b.WriteString(SubtitleStyle.Render("Joined Networks"))
		b.WriteString("\n")
		b.WriteString(c.table.View())
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("↑/↓ j/k navigate  l/enter detail  + join  x leave  p peers  esc back  h help  q quit"))
	case clientViewPeers:
		b.WriteString(SubtitleStyle.Render("Peers"))
		b.WriteString("\n")
		b.WriteString(c.table.View())
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("↑/↓ j/k navigate  esc back  h help  r refresh  q quit"))
	case clientViewDetail:
		b.WriteString(c.renderDetail())
	case clientViewJoin:
		b.WriteString(SubtitleStyle.Render("Join Network"))
		b.WriteString("\n")
		b.WriteString("Network ID: ")
		b.WriteString(c.joinInput.View())
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("enter submit  esc cancel  ctrl+c quit"))
	default:
		if status != nil {
			b.WriteString(renderNodeInfo(status))
		}
	}
	return b.String()
}

func (c *clientModel) renderDetail() string {
	if c.detail == nil {
		return "Loading network..."
	}
	n := c.detail
	var b strings.Builder
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Network %s", n.ID)))
	b.WriteString("\n\n")

	info := []struct{ k, v string }{
		{"Name", n.Name},
		{"Status", n.Status},
		{"Type", n.Type},
		{"MAC", n.MAC},
		{"Device", n.PortDeviceName},
		{"MTU", fmt.Sprintf("%d", n.MTU)},
		{"Addresses", joinStrings(n.AssignedAddresses)},
		{"DNS domain", n.DNS.Domain},
		{"DNS servers", joinStrings(n.DNS.Servers)},
	}
	for _, l := range info {
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render(l.k+":"), ValueStyle.Render(l.v)))
	}
	if len(n.Routes) > 0 {
		b.WriteString("\n  Routes:\n")
		for _, r := range n.Routes {
			b.WriteString(fmt.Sprintf("    %s via %s\n", r.Target, r.Via))
		}
	}

	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render("Client settings"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("  j/k move  space/l toggle  (or press key in brackets)"))
	b.WriteString("\n\n")

	settings := []struct {
		key   string
		label string
		desc  string
		on    bool
	}{
		{"d", "Allow DNS", "Manage system DNS settings", n.AllowDNS},
		{"g", "Allow default route", "Route all traffic through ZeroTier", n.AllowDefault},
		{"G", "Allow global IPs", "Non-private RFC1918 address ranges", n.AllowGlobal},
		{"m", "Allow managed IPs", "Controller-assigned addresses and routes", n.AllowManaged},
	}
	for i, s := range settings {
		b.WriteString(renderClientSettingRow(s.key, s.label, s.desc, s.on, c.settingFocus == i))
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("esc back  h help  r refresh  q quit"))
	return b.String()
}

func renderClientSettingRow(key, label, desc string, on bool, focused bool) string {
	prefix := "  "
	style := lipgloss.NewStyle()
	if focused {
		prefix = "> "
		style = MenuSelectedStyle
	}
	state := ValueStyle.Render(boolStr(on))
	if on {
		state = SuccessStyle.Render(boolStr(on))
	}
	line := fmt.Sprintf("%s[%s] %s: %s", prefix, key, label, state)
	out := style.Render(line)
	if focused && desc != "" {
		out += "\n" + HelpStyle.Render("      "+desc)
	}
	return out + "\n"
}

func renderNodeInfo(s *api.Status) string {
	var b strings.Builder
	lines := []struct{ k, v string }{
		{"Address", s.Address},
		{"Version", s.Version},
		{"Online", boolStr(s.Online)},
		{"TCP Fallback", boolStr(s.TCPFallbackActive)},
		{"Primary Port", fmt.Sprintf("%d", s.Config.Settings.PrimaryPort)},
		{"Port Mapping", boolStr(s.Config.Settings.PortMappingEnabled)},
		{"Listening On", joinStrings(s.Config.Settings.ListeningOn)},
		{"Surface Addrs", joinStrings(s.Config.Settings.SurfaceAddresses)},
	}
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render(l.k+":"), ValueStyle.Render(l.v)))
	}
	return b.String()
}

func renderStatusPanel(s *api.Status) string {
	if s == nil {
		return BoxStyle.Render("Node: connecting...")
	}
	return BoxStyle.Width(40).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("Node"),
			fmt.Sprintf("%s %s", LabelStyle.Render("ID:"), ValueStyle.Render(s.Address)),
			fmt.Sprintf("%s %s", LabelStyle.Render("Ver:"), ValueStyle.Render(s.Version)),
			fmt.Sprintf("%s %s", LabelStyle.Render("Online:"), ValueStyle.Render(boolStr(s.Online))),
		),
	)
}
