package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/brukberhane/ztnui/api"
	"github.com/brukberhane/ztnui/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	screenAuth = iota
	screenClient
	screenServer
	screenMembers
)

const (
	overlayNone = iota
	overlayHelp
	overlayNodeInfo
	overlaySettings
)

type appState struct {
	status             *api.Status
	networks           []api.Network
	peers              []api.Peer
	controllerStatus   *api.ControllerStatus
	controllerNetworks []string
	hasController      bool
	controllerChecked  bool
	lastError          string
	lastSuccess        string
	loading            bool
}

type Model struct {
	width, height int
	screen        int
	prevScreen    int
	state         appState
	client        clientModel
	server        serverModel
	members       membersModel
	settings      settingsForm
	auth          authModel
	needsAuth     bool
	apiClient     *api.Client
	cfg           *config.Config
	ready         bool
	overlay       int
}

func NewModel(cfg *config.Config, client *api.Client, needsAuth bool, authReason string) Model {
	screen := screenClient
	if needsAuth {
		screen = screenAuth
	}
	return Model{
		screen:    screen,
		cfg:       cfg,
		apiClient: client,
		needsAuth: needsAuth,
		auth:      newAuthModel(authReason),
		client:    newClientModel(),
		server:    newServerModel(),
		members:   newMembersModel(),
		settings:  newSettingsForm(),
	}
}

func (m Model) Init() tea.Cmd {
	if m.needsAuth || m.apiClient == nil {
		return nil
	}
	return tea.Batch(
		fetchStatus(m.apiClient),
		fetchNetworks(m.apiClient),
		tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return tickMsg{} }),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		_, _, _, _, contentH := m.layout()
		m.client.resize(msg.Width, contentH)
		m.server.resize(msg.Width, contentH)
		m.members.resize(msg.Width, contentH)
		if len(m.state.networks) > 0 {
			m.client.setNetworks(m.state.networks)
		}
		if len(m.state.controllerNetworks) > 0 && m.state.hasController {
			m.server.setNetworkList(m.state.controllerNetworks)
		}
		return m, nil

	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if key == "q" && !m.inTextInput() {
			if m.overlay != overlayNone {
				m.overlay = overlayNone
				m.state.lastError = ""
				return m, nil
			}
			return m, tea.Quit
		}

		if m.inTextInput() {
			if key == "esc" {
				if m.screen == screenAuth {
					return m, tea.Quit
				}
				if m.overlay == overlaySettings {
					m.overlay = overlayNone
					m.state.lastError = ""
					return m, nil
				}
				m, cmd := m.goBack()
				return m, cmd
			}
			if isSubmitKey(key) {
				m, cmd, handled := m.handleActionKeys(msg)
				if handled {
					if cmd != nil {
						cmds = append(cmds, cmd)
					}
					return m, tea.Batch(cmds...)
				}
			}
		} else {
			if key == "h" {
				m, cmd := m.toggleOverlay(overlayHelp)
				return m, cmd
			}
			if key == "n" {
				m, cmd := m.toggleOverlay(overlayNodeInfo)
				return m, cmd
			}
			if key == "," {
				m, cmd := m.toggleOverlay(overlaySettings)
				return m, cmd
			}
			if m.overlay == overlayNone && m.isRootTab() {
				m, cmd, ok := m.handleRootTabKeys(key)
				if ok {
					return m, cmd
				}
			}
			if isBackKey(key) {
				if m.screen == screenAuth {
					return m, tea.Quit
				}
				m, cmd := m.goBack()
				return m, cmd
			}
			if key == "r" {
				m, cmd := m.refresh()
				return m, cmd
			}

			m, cmd, handled := m.handleActionKeys(msg)
			if handled {
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			}
		}

	case tickMsg:
		if m.apiClient != nil && m.screen == screenClient && m.client.view == clientViewPeers {
			cmds = append(cmds, fetchPeers(m.apiClient))
		}
		return m, tea.Batch(cmds...)

	case statusMsg:
		m.state.loading = false
		if msg.err != nil {
			m.state.lastError = msg.err.Error()
		} else {
			m.state.status = msg.status
			m.ready = true
		}

	case networksMsg:
		m.state.loading = false
		if msg.err != nil {
			m.state.lastError = msg.err.Error()
		} else {
			m.state.networks = msg.networks
			_, _, _, _, contentH := m.layout()
			m.client.resize(m.width, contentH)
			m.client.setNetworks(msg.networks)
		}

	case networkMsg:
		m.state.loading = false
		if msg.err != nil {
			m.state.lastError = msg.err.Error()
		} else {
			m.client.detail = msg.network
			m.client.selectedID = msg.network.ID
		}

	case networkUpdatedMsg:
		m.state.loading = false
		if msg.err != nil {
			if m.screen == screenClient {
				m.state.lastError = msg.err.Error()
			}
		} else {
			m.state.lastError = ""
			m.state.lastSuccess = "Network settings updated"
			m.client.detail = msg.network
			m.client.view = clientViewDetail
		}

	case networkLeftMsg:
		m.state.loading = false
		if msg.err != nil {
			m.state.lastError = msg.err.Error()
		} else {
			m.state.lastSuccess = fmt.Sprintf("Left network %s", msg.id)
			cmds = append(cmds, fetchNetworks(m.apiClient))
		}

	case peersMsg:
		m.state.loading = false
		if msg.err != nil {
			m.state.lastError = msg.err.Error()
		} else {
			m.state.peers = msg.peers
			_, _, _, _, contentH := m.layout()
			m.client.resize(m.width, contentH)
			m.client.setPeers(msg.peers)
		}

	case controllerStatusMsg:
		m.state.loading = false
		m.state.controllerChecked = true
		if msg.err != nil {
			if api.IsNotFound(msg.err) && api.IsControllerPath(msg.err) {
				m.state.hasController = false
				m.state.controllerStatus = nil
				m.state.lastError = ""
			} else if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			m.state.controllerStatus = msg.status
			m.state.hasController = msg.status != nil && msg.status.Controller
			m.state.lastError = ""
			if m.state.hasController && m.isControllerScreen() {
				cmds = append(cmds, fetchControllerNetworks(m.apiClient))
			}
		}

	case controllerNetworksMsg:
		m.state.loading = false
		if msg.err != nil {
			if api.IsNotFound(msg.err) && api.IsControllerPath(msg.err) {
				m.state.hasController = false
				if m.isControllerScreen() {
					m.state.lastError = ""
				}
			} else if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			m.state.controllerNetworks = msg.ids
			_, _, _, _, contentH := m.layout()
			m.server.resize(m.width, contentH)
			m.server.setNetworkList(msg.ids)
			// fetch names for each network
			for _, id := range msg.ids {
				cmds = append(cmds, fetchControllerNetwork(m.apiClient, id))
			}
		}

	case controllerNetworkMsg:
		m.state.loading = false
		if msg.err != nil {
			if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			id := msg.network.NwID
			if id == "" {
				id = msg.network.ID
			}
			m.server.networks[id] = msg.network
			if m.screen == screenServer && m.server.view == serverViewList {
				m.server.setNetworkList(m.state.controllerNetworks)
			}
			if m.server.selectedID == id || m.server.pendingEdit {
				m.server.detail = msg.network
				m.server.selectedID = id
			}
			if m.server.pendingEdit {
				m.server.form = newNetworkForm()
				m.server.form.loadFrom(msg.network)
				m.server.form.editing = true
				m.server.view = serverViewEdit
				m.server.pendingEdit = false
			}
		}

	case controllerNetworkUpdatedMsg:
		m.state.loading = false
		if msg.err != nil {
			if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			m.state.lastSuccess = "Controller network saved"
			id := msg.network.NwID
			if id == "" {
				id = msg.network.ID
			}
			m.server.networks[id] = msg.network
			m.server.detail = msg.network
			m.server.view = serverViewDetail
		}

	case controllerNetworkCreatedMsg:
		m.state.loading = false
		if msg.err != nil {
			if api.IsNotFound(msg.err) && api.IsControllerPath(msg.err) {
				m.state.hasController = false
				m.state.controllerChecked = true
				m.state.lastError = ""
				m.server.view = serverViewList
			} else if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			id := msg.network.NwID
			if id == "" {
				id = msg.network.ID
			}
			m.state.lastSuccess = fmt.Sprintf("Created network %s", id)
			m.server.selectedID = id
			m.server.detail = msg.network
			m.server.view = serverViewDetail
			cmds = append(cmds, fetchControllerNetworks(m.apiClient))
		}

	case controllerNetworkDeletedMsg:
		m.state.loading = false
		if msg.err != nil {
			if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			m.state.lastSuccess = fmt.Sprintf("Deleted network %s", msg.id)
			delete(m.server.networks, msg.id)
			m.server.view = serverViewList
			cmds = append(cmds, fetchControllerNetworks(m.apiClient))
		}

	case membersMsg:
		m.state.loading = false
		if msg.err != nil {
			if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			_, _, _, _, contentH := m.layout()
			m.members.resize(m.width, contentH)
			m.members.setMembers(msg.members, m.cfg.HiddenMembersSet(m.members.networkID))
		}

	case memberUpdatedMsg:
		m.state.loading = false
		if msg.err != nil {
			if m.isControllerScreen() {
				m.state.lastError = msg.err.Error()
			}
		} else {
			m.state.lastSuccess = "Member updated"
			m.members.detail = msg.member
			for i := range m.members.members {
				if m.members.members[i].Address == msg.member.Address {
					m.members.members[i] = *msg.member
					break
				}
			}
			_, _, _, _, contentH := m.layout()
			m.members.resize(m.width, contentH)
			m.members.setMembers(m.members.members, m.cfg.HiddenMembersSet(m.members.networkID))
			if m.members.view == membersViewNameInput {
				m.members.view = membersViewDetail
			} else if m.members.view == membersViewIPAdd {
				m.members.view = membersViewIPList
			}
			if m.members.view == membersViewIPList {
				m.members.setIPTable(msg.member.IPAssignments)
			}
			cmds = append(cmds, fetchMembers(m.apiClient, m.members.networkID))
		}

	case memberDeletedMsg:
		m.state.loading = false
		if msg.err != nil {
			m.state.lastError = msg.err.Error()
		} else {
			m.state.lastSuccess = fmt.Sprintf(
				"Removed member %s — reappears unauthorized if node still joined",
				msg.nodeID,
			)
			m.members.view = membersViewList
			cmds = append(cmds, fetchMembers(m.apiClient, m.members.networkID))
		}

	case memberAuthorizedMsg:
		m.state.loading = false
		if msg.err != nil {
			m.state.lastError = msg.err.Error()
		} else {
			m.state.lastSuccess = fmt.Sprintf("Authorized member %s", msg.nodeID)
			m.members.view = membersViewList
			cmds = append(cmds, fetchMembers(m.apiClient, m.members.networkID))
		}
	}

	// delegate to sub-models
	var subCmd tea.Cmd
	if m.overlay == overlaySettings {
		switch m.settings.focusIndex {
		case 0:
			m.settings.controller, subCmd = m.settings.controller.Update(msg)
		case 1:
			m.settings.port, subCmd = m.settings.port.Update(msg)
		case 2:
			m.settings.token, subCmd = m.settings.token.Update(msg)
		}
	} else if m.overlay == overlayNone {
		switch m.screen {
		case screenAuth:
			m.auth.input, subCmd = m.auth.input.Update(msg)
		case screenClient:
			m.client, subCmd = m.client.Update(msg)
		case screenServer:
			m.server, subCmd = m.server.Update(msg)
		case screenMembers:
			m.members, subCmd = m.members.Update(msg)
		}
	}
	if subCmd != nil {
		cmds = append(cmds, subCmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) inTextInput() bool {
	if m.overlay == overlaySettings {
		return true
	}
	switch m.screen {
	case screenAuth:
		return true
	case screenClient:
		return m.client.view == clientViewJoin
	case screenServer:
		return m.server.view == serverViewEdit || m.server.view == serverViewCreateForm
	case screenMembers:
		return m.members.view == membersViewNameInput || m.members.view == membersViewIPAdd || m.members.view == membersViewAdd
	}
	return false
}

func (m Model) layout() (bannerH, titleH, tabsH, statusH, contentH int) {
	bannerH = 0
	if m.state.lastError != "" || m.state.lastSuccess != "" {
		bannerH = 1
	}
	titleH = 1
	tabsH = 0
	if m.showTabBar() {
		tabsH = 1
	}
	statusH = 1
	contentH = m.height - bannerH - titleH - tabsH - statusH
	if contentH < 1 {
		contentH = 1
	}
	return
}

func (m Model) isControllerScreen() bool {
	return m.screen == screenServer || m.screen == screenMembers
}

func (m Model) handleActionKeys(msg tea.KeyMsg) (Model, tea.Cmd, bool) {
	key := msg.String()

	switch m.overlay {
	case overlayHelp:
		return m, nil, true
	case overlayNodeInfo:
		return m.handleNodeInfoKeys(key)
	case overlaySettings:
		return m.handleSettingsKeys(key)
	}

	switch m.screen {
	case screenAuth:
		return m.handleAuthKeys(key)
	case screenClient:
		return m.handleClientKeys(key)
	case screenServer:
		return m.handleServerKeys(key)
	case screenMembers:
		return m.handleMembersKeys(key)
	}
	return m, nil, false
}

func (m Model) handleNodeInfoKeys(key string) (Model, tea.Cmd, bool) {
	// r is handled globally; esc/q/h are handled by overlay toggles.
	return m, nil, true
}

func (m Model) closeOverlay() Model {
	m.overlay = overlayNone
	m.state.lastError = ""
	return m
}

func (m Model) goBack() (Model, tea.Cmd) {
	if m.overlay != overlayNone {
		return m.closeOverlay(), nil
	}
	if m.isControllerScreen() {
		m.state.lastError = ""
	}
	switch m.screen {
	case screenClient:
		switch m.client.view {
		case clientViewNetworks:
			return m, nil
		default:
			m.client.view = clientViewNetworks
		}
	case screenServer:
		switch m.server.view {
		case serverViewList:
			return m, nil
		case serverViewCreateForm:
			m.server.blurCreateForm()
			m.server.view = serverViewCreate
		case serverViewCreate, serverViewDetail, serverViewRules:
			m.server.view = serverViewList
		case serverViewEdit:
			m.server.view = serverViewDetail
		}
	case screenMembers:
		switch m.members.view {
		case membersViewList:
			m.screen = screenServer
			m.server.view = serverViewDetail
		case membersViewNameInput:
			m.members.view = membersViewDetail
		case membersViewAdd:
			m.members.view = membersViewList
		case membersViewIPAdd:
			m.members.view = membersViewIPList
		case membersViewIPList:
			m.members.view = membersViewDetail
		default:
			m.members.view = membersViewList
		}
	}
	return m, nil
}

func (m Model) refresh() (Model, tea.Cmd) {
	m.state.loading = true
	m.state.lastError = ""

	if m.overlay == overlayNodeInfo {
		return m, fetchStatus(m.apiClient)
	}
	if m.overlay == overlaySettings || m.overlay == overlayHelp {
		return m, nil
	}

	switch m.screen {
	case screenClient:
		switch m.client.view {
		case clientViewPeers:
			return m, fetchPeers(m.apiClient)
		case clientViewDetail:
			id := m.client.selectedID
			return m, fetchNetwork(m.apiClient, id)
		default:
			return m, tea.Batch(fetchStatus(m.apiClient), fetchNetworks(m.apiClient))
		}
	case screenServer:
		if m.server.view == serverViewDetail && m.state.hasController {
			return m, fetchControllerNetwork(m.apiClient, m.server.selectedID)
		}
		if m.state.hasController {
			return m, fetchControllerNetworks(m.apiClient)
		}
		return m, fetchControllerStatus(m.apiClient)
	case screenMembers:
		return m, fetchMembers(m.apiClient, m.members.networkID)
	}
	return m, nil
}

func (m Model) handleAuthKeys(key string) (Model, tea.Cmd, bool) {
	if !isSubmitKey(key) {
		return m, nil, false
	}
	token := strings.TrimSpace(m.auth.input.Value())
	if len(token) != 24 {
		m.state.lastError = "Token must be 24 characters"
		return m, nil, true
	}
	if err := m.cfg.SaveToken(token); err != nil {
		m.state.lastError = err.Error()
		return m, nil, true
	}
	client := api.NewClient(m.cfg.BaseURL(), token)
	if err := validateToken(client); err != nil {
		_ = config.ClearStoredToken()
		m.state.lastError = "Token rejected: " + err.Error()
		return m, nil, true
	}
	m.apiClient = client
	m.needsAuth = false
	m.screen = screenClient
	m.state.lastError = ""
	m.state.lastSuccess = "Token saved securely"
	return m, tea.Batch(
		fetchStatus(m.apiClient),
		fetchNetworks(m.apiClient),
		tea.Tick(time.Second*5, func(t time.Time) tea.Msg { return tickMsg{} }),
	), true
}

func (m Model) handleClientKeys(key string) (Model, tea.Cmd, bool) {
	switch m.client.view {
	case clientViewNetworks:
		switch key {
		case "+":
			m.client.view = clientViewJoin
			m.client.joinInput.SetValue("")
			m.client.joinInput.Focus()
			return m, nil, true
		case "x":
			id := m.client.selectedNetworkID()
			if id != "" {
				m.state.loading = true
				return m, leaveNetwork(m.apiClient, id), true
			}
		case "p":
			m.client.view = clientViewPeers
			m.state.loading = true
			return m, fetchPeers(m.apiClient), true
		default:
			if isSelectKey(key) {
				id := m.client.selectedNetworkID()
				if id != "" {
					m.client.selectedID = id
					m.client.settingFocus = 0
					m.client.view = clientViewDetail
					m.state.loading = true
					return m, fetchNetwork(m.apiClient, id), true
				}
			}
		}
	case clientViewPeers:
		return m, nil, false
	case clientViewDetail:
		if m.client.detail == nil {
			return m, nil, true
		}
		switch key {
		case "up", "k":
			if m.client.settingFocus > 0 {
				m.client.settingFocus--
			}
			return m, nil, true
		case "down", "j":
			if m.client.settingFocus < 3 {
				m.client.settingFocus++
			}
			return m, nil, true
		case " ", "l", "enter":
			return m.toggleClientSetting(m.client.settingFocus)
		case "d":
			m.client.settingFocus = 0
			return m.toggleClientSetting(0)
		case "g":
			m.client.settingFocus = 1
			return m.toggleClientSetting(1)
		case "G":
			m.client.settingFocus = 2
			return m.toggleClientSetting(2)
		case "m":
			m.client.settingFocus = 3
			return m.toggleClientSetting(3)
		}
		return m, nil, true
	case clientViewJoin:
		if !isSubmitKey(key) {
			return m, nil, false
		}
		id := strings.TrimSpace(m.client.joinInput.Value())
		if len(id) == 16 {
			m.state.loading = true
			return m, joinNetwork(m.apiClient, id), true
		}
		m.state.lastError = "Network ID must be 16 characters"
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) serverControllerBlocked() bool {
	return m.state.controllerChecked && !m.state.hasController
}

func (m Model) controllerNodeID() (string, bool) {
	if m.state.status == nil || len(m.state.status.Address) != 10 {
		return "", false
	}
	return m.state.status.Address, true
}

func (m Model) createControllerNetwork(net *api.ControllerNetwork) (Model, tea.Cmd, bool) {
	if m.serverControllerBlocked() {
		return m, nil, true
	}
	addr, ok := m.controllerNodeID()
	if !ok {
		m.state.lastError = "Node ID unavailable — press r to refresh status"
		return m, fetchStatus(m.apiClient), true
	}
	m.state.loading = true
	m.state.lastError = ""
	return m, createControllerNetwork(m.apiClient, addr, net), true
}

func (m Model) toggleClientSetting(index int) (Model, tea.Cmd, bool) {
	if m.client.detail == nil {
		return m, nil, true
	}
	net := *m.client.detail
	switch index {
	case 0:
		net.AllowDNS = !net.AllowDNS
	case 1:
		net.AllowDefault = !net.AllowDefault
	case 2:
		net.AllowGlobal = !net.AllowGlobal
	case 3:
		net.AllowManaged = !net.AllowManaged
	default:
		return m, nil, true
	}
	m.state.lastError = ""
	m.state.loading = true
	return m, updateNetwork(m.apiClient, net.ID, &net), true
}

func (m Model) handleServerKeys(key string) (Model, tea.Cmd, bool) {
	if m.serverControllerBlocked() {
		return m, nil, true
	}
	if !m.state.controllerChecked && m.server.view == serverViewList {
		return m, nil, true
	}

	switch m.server.view {
	case serverViewList:
		switch key {
		case "c":
			m.server.view = serverViewCreate
			return m, nil, true
		case "e":
			id := m.server.selectedNetworkID()
			if id != "" {
				m.server.selectedID = id
				m.server.pendingEdit = true
				m.state.loading = true
				return m, fetchControllerNetwork(m.apiClient, id), true
			}
		case "d":
			id := m.server.selectedNetworkID()
			if id != "" {
				m.state.loading = true
				return m, deleteControllerNetwork(m.apiClient, id), true
			}
		case "m":
			id := m.server.selectedNetworkID()
			if id != "" {
				m.members.networkID = id
				m.screen = screenMembers
				m.members.view = membersViewList
				m.state.loading = true
				return m, fetchMembers(m.apiClient, id), true
			}
		default:
			if isSelectKey(key) {
				id := m.server.selectedNetworkID()
				if id != "" {
					m.server.selectedID = id
					m.server.view = serverViewDetail
					m.state.loading = true
					return m, fetchControllerNetwork(m.apiClient, id), true
				}
			}
		}
		return m, nil, false
	case serverViewDetail:
		switch key {
		case "e":
			if m.server.detail != nil {
				m.server.form = newNetworkForm()
				m.server.form.loadFrom(m.server.detail)
				m.server.form.editing = true
				m.server.view = serverViewEdit
			}
			return m, nil, true
		case "R":
			m.server.view = serverViewRules
			m.server.rulesPreset = 0
			return m, nil, true
		case "m":
			m.members.networkID = m.server.selectedID
			m.screen = screenMembers
			m.members.view = membersViewList
			m.state.loading = true
			return m, fetchMembers(m.apiClient, m.server.selectedID), true
		case "d":
			m.state.loading = true
			return m, deleteControllerNetwork(m.apiClient, m.server.selectedID), true
		}
		return m, nil, true
	case serverViewCreate:
		switch key {
		case "1":
			m.server.view = serverViewCreateForm
			m.server.createFocus = 0
			m.server.blurCreateForm()
			m.server.createName.Focus()
			return m, nil, true
		case "2":
			return m.createControllerNetwork(&api.ControllerNetwork{})
		}
		return m, nil, true
	case serverViewCreateForm:
		switch key {
		case "tab":
			m.server.createFocus = (m.server.createFocus + 1) % 4
			m.server.focusCreateField()
			return m, nil, true
		case "shift+tab":
			m.server.createFocus = (m.server.createFocus + 3) % 4
			m.server.focusCreateField()
			return m, nil, true
		default:
			if isSubmitKey(key) {
				return m.createControllerNetwork(buildCreateNetwork(
					m.server.createName.Value(),
					m.server.createPoolStart.Value(),
					m.server.createPoolEnd.Value(),
					m.server.createCIDR.Value(),
				))
			}
		}
		return m, nil, false
	case serverViewEdit:
		switch key {
		case "tab":
			m.server.form.nextFocus()
			return m, nil, true
		case "shift+tab":
			m.server.form.prevFocus()
			return m, nil, true
		case " ":
			m.server.form.toggleCurrent()
			return m, nil, true
		case "ctrl+s", "enter":
			net, err := m.server.form.toControllerNetwork()
			if err != nil {
				m.state.lastError = err.Error()
				return m, nil, true
			}
			m.state.loading = true
			return m, updateControllerNetwork(m.apiClient, m.server.selectedID, net), true
		}
		return m, nil, false
	case serverViewRules:
		presets := []string{"allow-all", "drop-all", "allow-network"}
		switch key {
		case "up", "k":
			if m.server.rulesPreset > 0 {
				m.server.rulesPreset--
			}
			return m, nil, true
		case "down", "j":
			if m.server.rulesPreset < len(presets)-1 {
				m.server.rulesPreset++
			}
			return m, nil, true
		default:
			if isSelectKey(key) {
				if m.server.detail != nil {
					preset := presets[m.server.rulesPreset]
					m.server.form = newNetworkForm()
					m.server.form.loadFrom(m.server.detail)
					m.server.form.rules.SetValue(RulesPreset(preset))
					m.server.view = serverViewEdit
					m.server.form.focusField(13)
				}
				return m, nil, true
			}
		}
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) toggleMemberHidden(nodeID string) (Model, tea.Cmd, bool) {
	if nodeID == "" {
		return m, nil, true
	}
	networkID := m.members.networkID
	hidden := m.cfg.IsMemberHidden(networkID, nodeID)
	m.cfg.SetMemberHidden(networkID, nodeID, !hidden)
	if err := m.cfg.Save(); err != nil {
		m.state.lastError = err.Error()
		return m, nil, true
	}
	if hidden {
		m.state.lastSuccess = fmt.Sprintf("Unhidden member %s", nodeID)
	} else {
		m.state.lastSuccess = fmt.Sprintf("Hidden member %s (local only)", nodeID)
	}
	m.members.rebuildTable(m.cfg.HiddenMembersSet(networkID))
	return m, nil, true
}

func (m Model) handleMembersKeys(key string) (Model, tea.Cmd, bool) {
	switch m.members.view {
	case membersViewList:
		switch key {
		case "t":
			m.members.toggleShowHidden(m.cfg.HiddenMembersSet(m.members.networkID))
			return m, nil, true
		case "H":
			id := m.members.selectedMemberID()
			return m.toggleMemberHidden(id)
		case "+":
			m.members.openAddMember()
			return m, nil, true
		case "a":
			return m.toggleMemberField("authorized")
		case "b":
			return m.toggleMemberField("bridge")
		case "o":
			return m.toggleMemberField("noAutoAssign")
		case "r":
			id := m.members.selectedMemberID()
			if mem := m.members.findMember(id); mem != nil {
				m.members.selectedID = id
				m.members.detail = mem
				m.members.openNameInput(mem)
			}
			return m, nil, true
		case "i":
			id := m.members.selectedMemberID()
			if mem := m.members.findMember(id); mem != nil {
				m.members.selectedID = id
				m.members.openIPList(mem)
			}
			return m, nil, true
		case "delete", "backspace":
			id := m.members.selectedMemberID()
			if id != "" {
				m.members.selectedID = id
				m.members.view = membersViewConfirmDelete
			}
			return m, nil, true
		default:
			if isSelectKey(key) {
				id := m.members.selectedMemberID()
				if id != "" {
					m.members.selectedID = id
					m.members.detail = m.members.findMember(id)
					m.members.settingFocus = 0
					m.members.view = membersViewDetail
				}
				return m, nil, true
			}
		}
		return m, nil, false
	case membersViewDetail:
		if m.members.detail == nil {
			return m, nil, true
		}
		switch key {
		case "up", "k":
			if m.members.settingFocus > 0 {
				m.members.settingFocus--
			}
			return m, nil, true
		case "down", "j":
			if m.members.settingFocus < memberSettingCount-1 {
				m.members.settingFocus++
			}
			return m, nil, true
		case " ", "l", "enter":
			switch m.members.settingFocus {
			case 0:
				m.members.openNameInput(m.members.detail)
				return m, nil, true
			case 1:
				return m.toggleMemberField("authorized")
			case 2:
				return m.toggleMemberField("bridge")
			case 3:
				return m.toggleMemberField("noAutoAssign")
			case 4:
				m.members.openIPList(m.members.detail)
				return m, nil, true
			}
			return m, nil, true
		case "r":
			m.members.settingFocus = 0
			m.members.openNameInput(m.members.detail)
			return m, nil, true
		case "a":
			m.members.settingFocus = 1
			return m.toggleMemberField("authorized")
		case "b":
			m.members.settingFocus = 2
			return m.toggleMemberField("bridge")
		case "o":
			m.members.settingFocus = 3
			return m.toggleMemberField("noAutoAssign")
		case "i":
			m.members.settingFocus = 4
			m.members.openIPList(m.members.detail)
			return m, nil, true
		case "H":
			return m.toggleMemberHidden(m.members.detail.Address)
		case "delete", "backspace":
			m.members.selectedID = m.members.detail.Address
			m.members.view = membersViewConfirmDelete
			return m, nil, true
		}
		return m, nil, true
	case membersViewNameInput:
		if !isSubmitKey(key) {
			return m, nil, false
		}
		id := m.members.selectedID
		name := strings.TrimSpace(m.members.nameInput.Value())
		return m.updateMemberFields(id, func(mem *api.ControllerNetworkMember) {
			mem.Name = name
		})
	case membersViewIPList:
		switch key {
		case "+":
			m.members.openIPAdd()
			return m, nil, true
		case "x", "delete", "backspace":
			ip := m.members.selectedIP()
			if ip == "" {
				return m, nil, true
			}
			id := m.members.selectedID
			return m.updateMemberFields(id, func(mem *api.ControllerNetworkMember) {
				mem.IPAssignments = removeString(mem.IPAssignments, ip)
			})
		}
		return m, nil, false
	case membersViewIPAdd:
		if !isSubmitKey(key) {
			return m, nil, false
		}
		ip := strings.TrimSpace(m.members.ipInput.Value())
		if ip == "" {
			m.state.lastError = "IP address required"
			return m, nil, true
		}
		id := m.members.selectedID
		return m.updateMemberFields(id, func(mem *api.ControllerNetworkMember) {
			for _, existing := range mem.IPAssignments {
				if existing == ip {
					return
				}
			}
			mem.IPAssignments = append(mem.IPAssignments, ip)
		})
	case membersViewAdd:
		if !isSubmitKey(key) {
			return m, nil, false
		}
		nodeID := strings.TrimSpace(m.members.addInput.Value())
		if !isValidNodeID(nodeID) {
			m.state.lastError = "Node ID must be 10 hex characters"
			return m, nil, true
		}
		m.state.loading = true
		return m, authorizeMember(m.apiClient, m.members.networkID, nodeID), true
	case membersViewConfirmDelete:
		if key == "y" {
			m.state.loading = true
			return m, deleteMember(m.apiClient, m.members.networkID, m.members.selectedID), true
		}
		if key == "esc" {
			m.members.view = membersViewList
			return m, nil, true
		}
		return m, nil, true
	}
	return m, nil, false
}

func (m Model) memberBase(id string) *api.ControllerNetworkMember {
	if mem := m.members.findMember(id); mem != nil {
		return mem
	}
	if m.members.detail != nil && m.members.detail.Address == id {
		return m.members.detail
	}
	return nil
}

func (m Model) updateMemberFields(id string, apply func(*api.ControllerNetworkMember)) (Model, tea.Cmd, bool) {
	base := m.memberBase(id)
	if base == nil {
		return m, nil, true
	}
	updated := *base
	apply(&updated)
	m.state.loading = true
	return m, updateMember(m.apiClient, m.members.networkID, id, &updated), true
}

func (m Model) toggleMemberField(field string) (Model, tea.Cmd, bool) {
	id := m.members.selectedMemberID()
	if id == "" {
		id = m.members.selectedID
	}
	return m.updateMemberFields(id, func(mem *api.ControllerNetworkMember) {
		switch field {
		case "authorized":
			mem.Authorized = !mem.Authorized
		case "bridge":
			mem.ActiveBridge = !mem.ActiveBridge
		case "noAutoAssign":
			mem.NoAutoAssignIps = !mem.NoAutoAssignIps
		}
	})
}

func (m *Model) loadSettingsForm() {
	m.settings.controller.SetValue(m.cfg.Controller)
	m.settings.port.SetValue(strconv.Itoa(m.cfg.Port))
	m.settings.token.SetValue("")
	if config.HasStoredToken() {
		m.settings.token.Placeholder = "stored securely (paste to replace)"
	} else {
		m.settings.token.Placeholder = "paste token to store securely"
	}
}

func (m Model) handleSettingsKeys(key string) (Model, tea.Cmd, bool) {
	switch key {
	case "tab":
		m.settings.focusField((m.settings.focusIndex + 1) % 3)
		return m, nil, true
	case "shift+tab":
		m.settings.focusField((m.settings.focusIndex + 2) % 3)
		return m, nil, true
	case "ctrl+s", "enter":
		port, err := strconv.Atoi(strings.TrimSpace(m.settings.port.Value()))
		if err != nil || port < 1 || port > 65535 {
			m.state.lastError = "Invalid port"
			return m, nil, true
		}
		m.cfg.Controller = strings.TrimSpace(m.settings.controller.Value())
		m.cfg.Port = port
		newToken := strings.TrimSpace(m.settings.token.Value())
		if newToken != "" {
			if len(newToken) != 24 {
				m.state.lastError = "Token must be 24 characters"
				return m, nil, true
			}
			if err := m.cfg.SaveToken(newToken); err != nil {
				m.state.lastError = err.Error()
				return m, nil, true
			}
		}
		if err := m.cfg.Save(); err != nil {
			m.state.lastError = err.Error()
			return m, nil, true
		}
		token, err := m.cfg.ResolveToken()
		if err != nil {
			m.needsAuth = true
			m.screen = screenAuth
			m.auth = newAuthModel(err.Error())
			m.state.lastError = "Token required"
			return m, nil, true
		}
		m.apiClient = api.NewClient(m.cfg.BaseURL(), token)
		if err := validateToken(m.apiClient); err != nil {
			m.state.lastError = "Token rejected: " + err.Error()
			return m, nil, true
		}
		m.needsAuth = false
		m.state.lastSuccess = "Settings saved"
		m.overlay = overlayNone
		return m, fetchStatus(m.apiClient), true
	}
	return m, nil, false
}

func (m Model) toggleOverlay(o int) (Model, tea.Cmd) {
	if m.overlay == o {
		m.overlay = overlayNone
		m.state.lastError = ""
		return m, nil
	}
	m.overlay = o
	m.state.lastError = ""
	switch o {
	case overlayNodeInfo:
		m.state.loading = true
		return m, fetchStatus(m.apiClient)
	case overlaySettings:
		m.loadSettingsForm()
	}
	return m, nil
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	_, _, _, _, contentH := m.layout()
	var sections []string

	if m.state.lastError != "" {
		sections = append(sections, ErrorStyle.Width(m.width).Render("✗ "+m.state.lastError))
	} else if m.state.lastSuccess != "" {
		sections = append(sections, SuccessStyle.Width(m.width).Render("✓ "+m.state.lastSuccess))
	}

	title := TitleStyle.Render("ztnui") + " " + HelpStyle.Render("ZeroTier Node & Controller TUI")
	sections = append(sections, title)

	if m.showTabBar() {
		sections = append(sections, m.renderTabBar())
	}

	var mainContent string
	switch m.overlay {
	case overlayHelp:
		mainContent = m.viewHelpOverlay()
	case overlayNodeInfo:
		mainContent = m.viewNodeInfoOverlay()
	case overlaySettings:
		mainContent = m.viewSettings()
	default:
		switch m.screen {
		case screenAuth:
			mainContent = m.auth.View()
		case screenClient:
			mainContent = m.client.View(m.state.status)
		case screenServer:
			mainContent = m.server.View(
				m.state.controllerStatus,
				m.state.hasController,
				m.state.controllerChecked,
				m.state.loading,
				len(m.state.controllerNetworks),
			)
		case screenMembers:
			mainContent = m.members.View()
		}
	}

	sections = append(sections, lipgloss.Place(m.width, contentH, lipgloss.Left, lipgloss.Top, mainContent))

	statusText := m.buildStatusBar()
	sections = append(sections, StatusBarStyle.Width(m.width).Render(statusText))

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(
		lipgloss.JoinVertical(lipgloss.Left, sections...),
	)
}

func (m Model) viewSettings() string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render("Settings"))
	b.WriteString("\n\n")
	b.WriteString(renderFormField("Controller", m.settings.focusIndex == 0, m.settings.controller.View()))
	b.WriteString(renderFormField("Port", m.settings.focusIndex == 1, m.settings.port.View()))
	b.WriteString(renderFormField("Auth token", m.settings.focusIndex == 2, m.settings.token.View()))
	if config.HasStoredToken() {
		b.WriteString(HelpStyle.Render("  Token on disk: encrypted (keyring or AES-GCM)"))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("tab focus  enter/ctrl+s save  esc close  q quit"))
	b.WriteString("\n")
	b.WriteString(SubtitleStyle.Render(fmt.Sprintf("Config: %s", m.cfg.BaseURL())))
	return b.String()
}

func (m Model) viewNodeInfoOverlay() string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render("Node Info"))
	b.WriteString("\n\n")
	if m.state.status == nil {
		b.WriteString("Loading...")
	} else {
		b.WriteString(renderNodeInfo(m.state.status))
	}
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("esc/q close  h help  r refresh"))
	return b.String()
}

func (m Model) buildStatusBar() string {
	parts := []string{}
	if m.state.loading {
		parts = append(parts, "Loading...")
	}
	switch m.overlay {
	case overlayHelp:
		parts = append(parts, "help")
	case overlayNodeInfo:
		parts = append(parts, "node info")
	case overlaySettings:
		parts = append(parts, "settings")
	}
	if m.state.status != nil {
		parts = append(parts, fmt.Sprintf("node:%s", m.state.status.Address))
	}
	parts = append(parts, fmt.Sprintf("api:%s", m.cfg.BaseURL()))
	if m.state.controllerStatus != nil {
		parts = append(parts, fmt.Sprintf("controller:%s", boolStr(m.state.controllerStatus.Controller)))
	}
	parts = append(parts, "tab/H/L switch  h help  n node  , settings  q quit  r refresh")
	return strings.Join(parts, " │ ")
}
