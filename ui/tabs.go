package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var mainTabs = []struct {
	screen int
	label  string
}{
	{screenClient, "Client"},
	{screenServer, "Server"},
	{screenNodeInfo, "Node Info"},
	{screenSettings, "Settings"},
}

func (m Model) showTabBar() bool {
	switch m.screen {
	case screenClient, screenServer, screenNodeInfo, screenSettings:
		return true
	default:
		return false
	}
}

func (m Model) tabIndex() int {
	for i, t := range mainTabs {
		if m.screen == t.screen {
			return i
		}
	}
	return -1
}

func (m Model) isRootTab() bool {
	if m.tabIndex() < 0 {
		return false
	}
	switch m.screen {
	case screenClient:
		return m.client.view == clientViewNetworks
	case screenServer:
		return m.server.view == serverViewList
	default:
		return true
	}
}

func (m Model) handleRootTabKeys(key string) (Model, tea.Cmd, bool) {
	switch key {
	case "n":
		m, cmd := m.activateTab(screenNodeInfo)
		return m, cmd, true
	case ",":
		if m.screen != screenSettings {
			m, cmd := m.activateTab(screenSettings)
			return m, cmd, true
		}
	case "c":
		if m.screen != screenServer {
			m, cmd := m.activateTab(screenClient)
			return m, cmd, true
		}
	case "s":
		if m.screen != screenClient {
			m, cmd := m.activateTab(screenServer)
			return m, cmd, true
		}
	}

	prev, next := m.tabNavKeys(key)
	if prev {
		m, cmd := m.switchTab(-1)
		return m, cmd, true
	}
	if next {
		m, cmd := m.switchTab(1)
		return m, cmd, true
	}
	return m, nil, false
}

func (m Model) tabNavKeys(key string) (prev, next bool) {
	if m.screen == screenSettings {
		return key == "H" || key == "h", key == "L"
	}
	switch key {
	case "shift+tab", "H", "h":
		return true, false
	case "tab", "L":
		return false, true
	}
	return false, false
}

func (m Model) switchTab(delta int) (Model, tea.Cmd) {
	idx := m.tabIndex()
	if idx < 0 {
		return m, nil
	}
	next := (idx + delta + len(mainTabs)) % len(mainTabs)
	return m.activateTab(mainTabs[next].screen)
}

func (m Model) activateTab(screen int) (Model, tea.Cmd) {
	if m.screen == screen && m.tabIndex() >= 0 {
		switch screen {
		case screenClient:
			if m.client.view == clientViewNetworks {
				return m.refresh()
			}
		case screenServer:
			if m.server.view == serverViewList {
				return m.refresh()
			}
		}
	}

	m.screen = screen
	m.state.lastError = ""

	switch screen {
	case screenClient:
		m.client.view = clientViewNetworks
		m.state.loading = true
		return m, tea.Batch(fetchStatus(m.apiClient), fetchNetworks(m.apiClient))
	case screenServer:
		m.server.view = serverViewList
		m.state.loading = true
		m.state.controllerChecked = false
		cmds := []tea.Cmd{fetchControllerStatus(m.apiClient)}
		if m.state.status == nil {
			cmds = append(cmds, fetchStatus(m.apiClient))
		}
		return m, tea.Batch(cmds...)
	case screenNodeInfo:
		return m, fetchStatus(m.apiClient)
	case screenSettings:
		m.loadSettingsForm()
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) renderTabBar() string {
	parts := make([]string, 0, len(mainTabs))
	for _, t := range mainTabs {
		style := TabStyle
		if m.screen == t.screen {
			style = TabActiveStyle
		}
		parts = append(parts, style.Render(t.label))
	}
	hint := HelpStyle.Render("  tab/H/L switch")
	return lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(parts, " │ ")) + hint
}
