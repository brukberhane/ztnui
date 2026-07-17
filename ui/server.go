package ui

import (
	"fmt"
	"strings"

	"github.com/brukberhane/ztnui/api"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type serverView int

const (
	serverViewList serverView = iota
	serverViewDetail
	serverViewEdit
	serverViewCreate
	serverViewRules
	serverViewCreateForm
)

type serverModel struct {
	view            serverView
	table           table.Model
	networks        map[string]*api.ControllerNetwork
	selectedID      string
	detail          *api.ControllerNetwork
	form            networkForm
	createName      textinput.Model
	createPoolStart textinput.Model
	createPoolEnd   textinput.Model
	createCIDR      textinput.Model
	rulesPreset     int
	pendingEdit     bool
	createFocus     int
	width           int
	tableHeight     int
}

func newServerModel() serverModel {
	s := serverModel{
		networks:        make(map[string]*api.ControllerNetwork),
		form:            newNetworkForm(),
		createName:      newInput("my-network"),
		createPoolStart: newInput("10.147.20.1"),
		createPoolEnd:   newInput("10.147.20.254"),
		createCIDR:      newInput("10.147.20.0/24"),
	}
	return s
}

func (s *serverModel) setNetworkList(ids []string) {
	rows := make([]table.Row, 0, len(ids))
	for _, id := range ids {
		name := "-"
		if net, ok := s.networks[id]; ok && net.Name != "" {
			name = net.Name
		}
		rows = append(rows, table.Row{id, truncate(name, 30)})
	}
	cols := []table.Column{
		{Title: "Network ID", Width: 18},
		{Title: "Name", Width: 30},
	}
	s.table = newTable(cols, rows, max(20, s.width-4), s.tableHeight)
}

func (s *serverModel) resize(width, contentHeight int) {
	s.width = width
	s.tableHeight = max(3, contentHeight-6)
	if len(s.table.Rows()) > 0 {
		s.table.SetHeight(s.tableHeight)
		s.table.SetWidth(max(20, width-4))
	}
}

func (s *serverModel) selectedNetworkID() string {
	if s.view == serverViewList && len(s.table.Rows()) > 0 {
		row := s.table.SelectedRow()
		if len(row) > 0 {
			return row[0]
		}
	}
	return s.selectedID
}

func (s *serverModel) Update(msg tea.Msg) (serverModel, tea.Cmd) {
	var cmd tea.Cmd
	switch s.view {
	case serverViewList:
		s.table, cmd = s.table.Update(msg)
	case serverViewEdit:
		switch s.form.focusIndex {
		case 0:
			s.form.name, cmd = s.form.name.Update(msg)
		case 3:
			s.form.mtu, cmd = s.form.mtu.Update(msg)
		case 4:
			s.form.multicastLimit, cmd = s.form.multicastLimit.Update(msg)
		case 8:
			s.form.poolStart, cmd = s.form.poolStart.Update(msg)
		case 9:
			s.form.poolEnd, cmd = s.form.poolEnd.Update(msg)
		case 10:
			s.form.routes, cmd = s.form.routes.Update(msg)
		case 11:
			s.form.dnsDomain, cmd = s.form.dnsDomain.Update(msg)
		case 12:
			s.form.dnsServers, cmd = s.form.dnsServers.Update(msg)
		case 13:
			s.form.rules, cmd = s.form.rules.Update(msg)
		}
	case serverViewCreateForm:
		switch s.createFocus {
		case 0:
			s.createName, cmd = s.createName.Update(msg)
		case 1:
			s.createPoolStart, cmd = s.createPoolStart.Update(msg)
		case 2:
			s.createPoolEnd, cmd = s.createPoolEnd.Update(msg)
		case 3:
			s.createCIDR, cmd = s.createCIDR.Update(msg)
		}
	case serverViewRules:
		// preset selection handled in app
	}
	return *s, cmd
}

func (s *serverModel) blurCreateForm() {
	s.createName.Blur()
	s.createPoolStart.Blur()
	s.createPoolEnd.Blur()
	s.createCIDR.Blur()
}

func (s *serverModel) focusCreateField() {
	s.blurCreateForm()
	switch s.createFocus {
	case 0:
		s.createName.Focus()
	case 1:
		s.createPoolStart.Focus()
	case 2:
		s.createPoolEnd.Focus()
	case 3:
		s.createCIDR.Focus()
	}
}

func (s *serverModel) View(controllerStatus *api.ControllerStatus, hasController, checked bool, loading bool, networkCount int) string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render("Server (Controller)"))
	b.WriteString("\n")

	if checked && !hasController {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render("This node is not a self-hosted ZeroTier controller."))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("The /controller API returned 404 or controller:false."))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("Enable controller in ZeroTier, or point Settings at a remote controller host."))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc back  h help  q quit"))
		return b.String()
	}

	if !checked || loading {
		b.WriteString("\n")
		b.WriteString(SubtitleStyle.Render("Checking controller status..."))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("esc back  h help  q quit"))
		return b.String()
	}

	if controllerStatus != nil {
		b.WriteString(SubtitleStyle.Render(fmt.Sprintf(
			"Controller: %s  API v%d",
			boolStr(controllerStatus.Controller),
			controllerStatus.APIVersion,
		)))
		b.WriteString("\n")
	}

	switch s.view {
	case serverViewList:
		if networkCount == 0 {
			b.WriteString(SubtitleStyle.Render("No controller networks yet."))
			b.WriteString("\n")
			b.WriteString(HelpStyle.Render("Press c to create a network."))
			b.WriteString("\n\n")
		} else {
			b.WriteString(s.table.View())
			b.WriteString("\n")
		}
		b.WriteString(HelpStyle.Render("↑/↓ j/k navigate  l/enter detail  c create  e edit  d delete  m members  esc back  h help  q quit"))
	case serverViewDetail:
		b.WriteString(s.renderDetail())
	case serverViewEdit:
		b.WriteString(s.renderEditForm())
	case serverViewCreate:
		b.WriteString(s.renderCreatePrompt())
	case serverViewCreateForm:
		b.WriteString(s.renderCreateForm())
	case serverViewRules:
		b.WriteString(s.renderRules())
	}
	return b.String()
}

func (s *serverModel) renderDetail() string {
	if s.detail == nil {
		return "Loading network..."
	}
	n := s.detail
	var b strings.Builder
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Network %s", n.NwID)))
	b.WriteString("\n\n")
	lines := []struct{ k, v string }{
		{"Name", n.Name},
		{"Private", boolStr(n.Private)},
		{"Broadcast", boolStr(n.EnableBroadcast)},
		{"MTU", fmt.Sprintf("%d", n.MTU)},
		{"Multicast Limit", fmt.Sprintf("%d", n.MulticastLimit)},
		{"v4 ZT", boolStr(n.V4AssignMode.ZT)},
		{"DNS Domain", n.DNS.Domain},
		{"DNS Servers", joinStrings(n.DNS.Servers)},
	}
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render(l.k+":"), ValueStyle.Render(l.v)))
	}
	if len(n.IPAssignmentPools) > 0 {
		p := n.IPAssignmentPools[0]
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("IP Pool:"),
			ValueStyle.Render(p.IPRangeStart+" - "+p.IPRangeEnd)))
	}
	if len(n.Routes) > 0 {
		b.WriteString("\n  Routes:\n")
		for _, r := range n.Routes {
			via := "null"
			if r.Via != nil {
				via = *r.Via
			}
			b.WriteString(fmt.Sprintf("    %s via %s\n", r.Target, via))
		}
	}
	b.WriteString(fmt.Sprintf("  %s %d rule(s)\n", LabelStyle.Render("Rules:"), len(n.Rules)))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("e edit  R rules  m members  d delete  esc back  h help  r refresh  q quit"))
	return b.String()
}

func (s *serverModel) renderEditForm() string {
	f := s.form
	var b strings.Builder
	b.WriteString(SubtitleStyle.Render("Edit Network"))
	b.WriteString("\n\n")
	b.WriteString(renderFormField("Name", f.focusIndex == 0, f.name.View()))
	b.WriteString(renderToggle("Private", f.focusIndex == 1, f.private))
	b.WriteString(renderToggle("Broadcast", f.focusIndex == 2, f.enableBroadcast))
	b.WriteString(renderFormField("MTU", f.focusIndex == 3, f.mtu.View()))
	b.WriteString(renderFormField("Multicast Limit", f.focusIndex == 4, f.multicastLimit.View()))
	b.WriteString(renderToggle("v4 ZT", f.focusIndex == 5, f.v4ZT))
	b.WriteString(renderToggle("v6 6plane", f.focusIndex == 6, f.v6SixPlane))
	b.WriteString(renderToggle("v6 rfc4193", f.focusIndex == 7, f.v6RFC4193))
	b.WriteString(renderFormField("Pool Start", f.focusIndex == 8, f.poolStart.View()))
	b.WriteString(renderFormField("Pool End", f.focusIndex == 9, f.poolEnd.View()))
	b.WriteString(renderFormField("Routes (target;via,...)", f.focusIndex == 10, f.routes.View()))
	b.WriteString(renderFormField("DNS Domain", f.focusIndex == 11, f.dnsDomain.View()))
	b.WriteString(renderFormField("DNS Servers", f.focusIndex == 12, f.dnsServers.View()))
	b.WriteString("\nRules JSON:\n")
	b.WriteString(f.rules.View())
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("tab/shift+tab focus  space toggle  enter/ctrl+s save  esc back  q quit"))
	return b.String()
}

func (s *serverModel) renderCreatePrompt() string {
	return SubtitleStyle.Render("Create Network") + "\n\n" +
		"  [1] Quick create (name + IP range)\n" +
		"  [2] Blank network (random ID)\n\n" +
		HelpStyle.Render("1/2 select  esc back  h help  q quit")
}

func (s *serverModel) renderCreateForm() string {
	var b strings.Builder
	b.WriteString(SubtitleStyle.Render("Quick Create Network"))
	b.WriteString("\n\n")
	b.WriteString(renderFormField("Name", s.createFocus == 0, s.createName.View()))
	b.WriteString(renderFormField("Pool Start", s.createFocus == 1, s.createPoolStart.View()))
	b.WriteString(renderFormField("Pool End", s.createFocus == 2, s.createPoolEnd.View()))
	b.WriteString(renderFormField("CIDR", s.createFocus == 3, s.createCIDR.View()))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("tab next field  enter submit  esc back  ctrl+s save  q quit"))
	return b.String()
}

func (s *serverModel) renderRules() string {
	presets := []string{"allow-all", "drop-all", "allow-network"}
	var b strings.Builder
	b.WriteString(SubtitleStyle.Render("Rules Presets"))
	b.WriteString("\n\n")
	for i, p := range presets {
		style := MenuItemStyle
		if i == s.rulesPreset {
			style = MenuSelectedStyle
		}
		b.WriteString(style.Render(fmt.Sprintf("  %s", p)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓ j/k select  l/enter apply  esc back  h help  q quit"))
	return b.String()
}

func renderFormField(label string, focused bool, input string) string {
	prefix := "  "
	if focused {
		prefix = "> "
	}
	return fmt.Sprintf("%s%s %s\n", prefix, LabelStyle.Render(label+":"), input)
}

func renderToggle(label string, focused bool, val bool) string {
	prefix := "  "
	if focused {
		prefix = "> "
	}
	mark := "[ ]"
	if val {
		mark = "[x]"
	}
	return fmt.Sprintf("%s%s %s %s\n", prefix, LabelStyle.Render(label+":"), mark, boolStr(val))
}

func buildCreateNetwork(name, poolStart, poolEnd, cidr string) *api.ControllerNetwork {
	return &api.ControllerNetwork{
		Name:            name,
		Private:         true,
		EnableBroadcast: true,
		V4AssignMode:    api.V4AssignMode{ZT: true},
		IPAssignmentPools: []api.IPAssignmentPool{
			{IPRangeStart: poolStart, IPRangeEnd: poolEnd},
		},
		Routes: []api.Route{{Target: cidr, Via: nil}},
	}
}
