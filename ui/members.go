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

type membersView int

const (
	membersViewList membersView = iota
	membersViewDetail
	membersViewNameInput
	membersViewIPList
	membersViewIPAdd
	membersViewConfirmDelete
)

const memberSettingCount = 5

type membersModel struct {
	view          membersView
	networkID     string
	table         table.Model
	ipTable       table.Model
	members       []api.ControllerNetworkMember
	selectedID    string
	detail        *api.ControllerNetworkMember
	settingFocus  int
	nameInput     textinput.Model
	ipInput       textinput.Model
	width         int
	tableHeight   int
}

func newMembersModel() membersModel {
	ipInput := newInput("10.147.20.100/24")
	ipInput.CharLimit = 64
	return membersModel{
		nameInput: newInput("e.g. office-router"),
		ipInput:   ipInput,
	}
}

func memberDisplayName(mem api.ControllerNetworkMember) string {
	if n := strings.TrimSpace(mem.Name); n != "" {
		return n
	}
	return "—"
}

func ipAssignmentSummary(ips []string) string {
	if len(ips) == 0 {
		return "(none)"
	}
	if len(ips) == 1 {
		return ips[0]
	}
	return fmt.Sprintf("%d assigned (%s …)", len(ips), ips[0])
}

func (m *membersModel) setMembers(members []api.ControllerNetworkMember) {
	m.members = members
	rows := make([]table.Row, 0, len(members))
	for _, mem := range members {
		auth := "no"
		if mem.Authorized {
			auth = "yes"
		}
		auto := "yes"
		if mem.NoAutoAssignIps {
			auto = "no"
		}
		rows = append(rows, table.Row{
			mem.Address,
			truncate(memberDisplayName(mem), 16),
			auth,
			boolStr(mem.ActiveBridge),
			auto,
			truncate(joinStrings(mem.IPAssignments), 16),
			fmt.Sprintf("%d.%d.%d", mem.VMajor, mem.VMinor, mem.VRev),
		})
	}
	cols := []table.Column{
		{Title: "Node ID", Width: 12},
		{Title: "Name", Width: 16},
		{Title: "Auth", Width: 6},
		{Title: "Bridge", Width: 8},
		{Title: "AutoIP", Width: 8},
		{Title: "IPs", Width: 16},
		{Title: "Ver", Width: 8},
	}
	m.table = newTable(cols, rows, max(20, m.width-4), m.tableHeight)
}

func (m *membersModel) setIPTable(ips []string) {
	rows := make([]table.Row, 0, len(ips))
	for _, ip := range ips {
		rows = append(rows, table.Row{ip})
	}
	cols := []table.Column{{Title: "IP Assignment", Width: max(20, m.width-8)}}
	height := max(3, m.tableHeight-2)
	m.ipTable = newTable(cols, rows, max(20, m.width-4), height)
}

func (m *membersModel) resize(width, contentHeight int) {
	m.width = width
	m.tableHeight = max(3, contentHeight-6)
	if len(m.table.Rows()) > 0 {
		m.table.SetHeight(m.tableHeight)
		m.table.SetWidth(max(20, width-4))
	}
	if m.view == membersViewIPList {
		height := max(3, m.tableHeight-2)
		m.ipTable.SetHeight(height)
		m.ipTable.SetWidth(max(20, width-4))
	}
}

func (m *membersModel) selectedMemberID() string {
	if m.view == membersViewList && len(m.table.Rows()) > 0 {
		row := m.table.SelectedRow()
		if len(row) > 0 {
			return row[0]
		}
	}
	return m.selectedID
}

func (m *membersModel) selectedIP() string {
	if m.view == membersViewIPList && len(m.ipTable.Rows()) > 0 {
		row := m.ipTable.SelectedRow()
		if len(row) > 0 {
			return row[0]
		}
	}
	return ""
}

func (m *membersModel) findMember(id string) *api.ControllerNetworkMember {
	for i := range m.members {
		if m.members[i].Address == id {
			return &m.members[i]
		}
	}
	return nil
}

func (m *membersModel) openNameInput(mem *api.ControllerNetworkMember) {
	if mem == nil {
		return
	}
	m.nameInput.SetValue(mem.Name)
	m.nameInput.Focus()
	m.view = membersViewNameInput
}

func (m *membersModel) openIPList(mem *api.ControllerNetworkMember) {
	if mem == nil {
		return
	}
	m.detail = mem
	m.setIPTable(mem.IPAssignments)
	m.view = membersViewIPList
}

func (m *membersModel) openIPAdd() {
	m.ipInput.SetValue("")
	m.ipInput.Focus()
	m.view = membersViewIPAdd
}

func (m *membersModel) Update(msg tea.Msg) (membersModel, tea.Cmd) {
	var cmd tea.Cmd
	switch m.view {
	case membersViewList:
		m.table, cmd = m.table.Update(msg)
	case membersViewIPList:
		m.ipTable, cmd = m.ipTable.Update(msg)
	case membersViewNameInput:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case membersViewIPAdd:
		m.ipInput, cmd = m.ipInput.Update(msg)
	}
	return *m, cmd
}

func (m *membersModel) View() string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render(fmt.Sprintf("Members — %s", m.networkID)))
	b.WriteString("\n")

	switch m.view {
	case membersViewList:
		if len(m.members) == 0 {
			b.WriteString(SubtitleStyle.Render("No members yet. Join devices to this network."))
		} else {
			b.WriteString(m.table.View())
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("↑/↓ j/k navigate  l/enter detail  a auth  b bridge  o auto-IP  i IPs  n name  delete remove  h back"))
	case membersViewDetail:
		b.WriteString(m.renderDetail())
	case membersViewNameInput:
		b.WriteString(SubtitleStyle.Render("Set member name"))
		b.WriteString("\n")
		b.WriteString(m.nameInput.View())
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("enter save  esc cancel"))
	case membersViewIPList:
		b.WriteString(m.renderIPList())
	case membersViewIPAdd:
		b.WriteString(SubtitleStyle.Render("Add IP assignment"))
		b.WriteString("\n")
		b.WriteString(m.ipInput.View())
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("enter add  esc cancel"))
	case membersViewConfirmDelete:
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Delete member %s? Press y to confirm, esc to cancel", m.selectedID)))
	}
	return b.String()
}

func (m *membersModel) renderIPList() string {
	if m.detail == nil {
		return "Loading..."
	}
	var b strings.Builder
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("IP assignments — %s", memberDisplayName(*m.detail))))
	b.WriteString("\n\n")
	if len(m.detail.IPAssignments) == 0 {
		b.WriteString(HelpStyle.Render("  No IPs assigned. Press + to add one."))
		b.WriteString("\n")
	} else {
		b.WriteString(m.ipTable.View())
		b.WriteString("\n")
	}
	b.WriteString(HelpStyle.Render("+ add  x remove selected  h back  q quit"))
	return b.String()
}

func (m *membersModel) renderDetail() string {
	if m.detail == nil {
		return "Loading member..."
	}
	mem := m.detail
	var b strings.Builder
	title := memberDisplayName(*mem)
	if title == "—" {
		title = mem.Address
	}
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Member %s (%s)", title, mem.Address)))
	b.WriteString("\n\n")
	lines := []struct{ k, v string }{
		{"Version", fmt.Sprintf("%d.%d.%d", mem.VMajor, mem.VMinor, mem.VRev)},
		{"Protocol", fmt.Sprintf("%d", mem.VProto)},
	}
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render(l.k+":"), ValueStyle.Render(l.v)))
	}
	if mem.Identity != "" {
		b.WriteString(fmt.Sprintf("  %s %s\n", LabelStyle.Render("Identity:"), ValueStyle.Render(truncate(mem.Identity, 60))))
	}

	b.WriteString("\n")
	b.WriteString(HeaderStyle.Render("Member settings"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("  j/k move  space/l toggle or open  i edit IPs"))
	b.WriteString("\n\n")

	b.WriteString(renderMemberNameRow("n", "Name", mem.Name, m.settingFocus == 0))
	b.WriteString(renderMemberToggleRow("a", "Authorized", "Allow this node on the network", mem.Authorized, m.settingFocus == 1))
	b.WriteString(renderMemberToggleRow("b", "Active bridge", "Relay L2 Ethernet frames", mem.ActiveBridge, m.settingFocus == 2))
	b.WriteString(renderMemberToggleRow("o", "Auto-assign IPs", "Let controller assign from pool", !mem.NoAutoAssignIps, m.settingFocus == 3))
	b.WriteString(renderMemberActionRow("i", "IP assignments", ipAssignmentSummary(mem.IPAssignments), m.settingFocus == 4))

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("delete remove  h back  q quit"))
	return b.String()
}

func renderMemberNameRow(key, label, value string, focused bool) string {
	prefix := "  "
	style := lipgloss.NewStyle()
	if focused {
		prefix = "> "
		style = MenuSelectedStyle
	}
	display := strings.TrimSpace(value)
	val := HelpStyle.Render("(unset)")
	if display != "" {
		val = ValueStyle.Render(truncate(display, 40))
	}
	line := fmt.Sprintf("%s[%s] %s: %s", prefix, key, label, val)
	out := style.Render(line)
	if focused {
		out += "\n" + HelpStyle.Render("      enter to edit")
	}
	return out + "\n"
}

func renderMemberActionRow(key, label, value string, focused bool) string {
	prefix := "  "
	style := lipgloss.NewStyle()
	if focused {
		prefix = "> "
		style = MenuSelectedStyle
	}
	val := ValueStyle.Render(truncate(value, 40))
	line := fmt.Sprintf("%s[%s] %s: %s", prefix, key, label, val)
	out := style.Render(line)
	if focused {
		out += "\n" + HelpStyle.Render("      enter/l to manage IPs")
	}
	return out + "\n"
}

func renderMemberToggleRow(key, label, desc string, on bool, focused bool) string {
	return renderClientSettingRow(key, label, desc, on, focused)
}
