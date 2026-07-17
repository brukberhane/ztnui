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
	membersViewAdd
	membersViewConfirmDelete
)

const memberSettingCount = 5

type membersModel struct {
	view         membersView
	networkID    string
	table        table.Model
	ipTable      table.Model
	members      []api.ControllerNetworkMember
	selectedID   string
	detail       *api.ControllerNetworkMember
	settingFocus int
	showHidden   bool
	visibleCount int
	hiddenCount  int
	nameInput    textinput.Model
	ipInput      textinput.Model
	addInput     textinput.Model
	width        int
	tableHeight  int
}

func newMembersModel() membersModel {
	ipInput := newInput("10.147.20.100/24")
	ipInput.CharLimit = 64
	addInput := newInput("10-char node ID")
	addInput.CharLimit = 10
	return membersModel{
		nameInput: newInput("e.g. office-router"),
		ipInput:   ipInput,
		addInput:  addInput,
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

func (m *membersModel) setMembers(members []api.ControllerNetworkMember, hidden map[string]bool) {
	m.members = members
	m.rebuildTable(hidden)
}

func (m *membersModel) rebuildTable(hidden map[string]bool) {
	rows := make([]table.Row, 0, len(m.members))
	m.visibleCount = 0
	m.hiddenCount = 0
	for _, mem := range m.members {
		isHidden := hidden[mem.Address]
		if isHidden {
			m.hiddenCount++
		}
		if isHidden && !m.showHidden {
			continue
		}
		m.visibleCount++
		auth := "no"
		if mem.Authorized {
			auth = "yes"
		}
		auto := "yes"
		if mem.NoAutoAssignIps {
			auto = "no"
		}
		name := memberDisplayName(mem)
		if isHidden {
			name = "[hidden] " + name
		}
		rows = append(rows, table.Row{
			mem.Address,
			truncate(name, 20),
			auth,
			boolStr(mem.ActiveBridge),
			auto,
			truncate(joinStrings(mem.IPAssignments), 16),
			fmt.Sprintf("%d.%d.%d", mem.VMajor, mem.VMinor, mem.VRev),
		})
	}
	cols := []table.Column{
		{Title: "Node ID", Width: 12},
		{Title: "Name", Width: 20},
		{Title: "Auth", Width: 6},
		{Title: "Bridge", Width: 8},
		{Title: "AutoIP", Width: 8},
		{Title: "IPs", Width: 16},
		{Title: "Ver", Width: 8},
	}
	m.table = newTable(cols, rows, max(20, m.width-4), m.tableHeight)
}

func (m *membersModel) toggleShowHidden(hidden map[string]bool) {
	m.showHidden = !m.showHidden
	m.rebuildTable(hidden)
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

func (m *membersModel) openAddMember() {
	m.addInput.SetValue("")
	m.addInput.Focus()
	m.view = membersViewAdd
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
	case membersViewAdd:
		m.addInput, cmd = m.addInput.Update(msg)
	}
	return *m, cmd
}

func (m *membersModel) View() string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render(fmt.Sprintf("Members — %s", m.networkID)))
	b.WriteString("\n")

	switch m.view {
	case membersViewList:
		if m.visibleCount == 0 {
			if m.hiddenCount > 0 && !m.showHidden {
				b.WriteString(SubtitleStyle.Render(fmt.Sprintf("%d hidden member(s) — press t to show", m.hiddenCount)))
			} else {
				b.WriteString(SubtitleStyle.Render("No members yet."))
				b.WriteString("\n")
				b.WriteString(HelpStyle.Render("Devices appear here when they join, or press + to authorize a node ID."))
			}
		} else {
			if m.hiddenCount > 0 {
				state := "off"
				if m.showHidden {
					state = "on"
				}
				b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Hidden: %d (show %s)", m.hiddenCount, state)))
				b.WriteString("\n")
			}
			b.WriteString(m.table.View())
		}
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("↑/↓ j/k navigate  l/enter detail  + authorize  H hide/unhide  t show hidden  a auth  delete remove  esc back  h help"))
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
	case membersViewAdd:
		b.WriteString(SubtitleStyle.Render("Authorize member by node ID"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Pre-approve a node before it joins, or authorize one waiting for approval."))
		b.WriteString("\n\n")
		b.WriteString("Node ID: ")
		b.WriteString(m.addInput.View())
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("enter authorize  esc cancel"))
	case membersViewConfirmDelete:
		b.WriteString(ErrorStyle.Render(fmt.Sprintf("Remove member %s?", m.selectedID)))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Deauthorizes, clears IPs, then deletes the controller record."))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("If the node is still joined to this network, it will reappear unauthorized until it leaves."))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("y confirm  esc cancel"))
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
	b.WriteString(HelpStyle.Render("+ add  x remove selected  esc back  h help  q quit"))
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

	b.WriteString(renderMemberNameRow("r", "Name", mem.Name, m.settingFocus == 0))
	b.WriteString(renderMemberToggleRow("a", "Authorized", "Allow this node on the network", mem.Authorized, m.settingFocus == 1))
	b.WriteString(renderMemberToggleRow("b", "Active bridge", "Relay L2 Ethernet frames", mem.ActiveBridge, m.settingFocus == 2))
	b.WriteString(renderMemberToggleRow("o", "Auto-assign IPs", "Let controller assign from pool", !mem.NoAutoAssignIps, m.settingFocus == 3))
	b.WriteString(renderMemberActionRow("i", "IP assignments", ipAssignmentSummary(mem.IPAssignments), m.settingFocus == 4))

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("H hide  delete remove  esc back  h help  q quit"))
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
