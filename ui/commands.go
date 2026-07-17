package ui

import (
	"context"

	"github.com/brukberhane/ztnui/api"
	tea "github.com/charmbracelet/bubbletea"
)

// Message types for async API results.

type statusMsg struct {
	status *api.Status
	err    error
}

type networksMsg struct {
	networks []api.Network
	err      error
}

type networkMsg struct {
	network *api.Network
	err     error
}

type networkUpdatedMsg struct {
	network *api.Network
	err     error
}

type networkLeftMsg struct {
	id  string
	err error
}

type peersMsg struct {
	peers []api.Peer
	err   error
}

type controllerStatusMsg struct {
	status *api.ControllerStatus
	err    error
}

type controllerNetworksMsg struct {
	ids []string
	err error
}

type controllerNetworkMsg struct {
	network *api.ControllerNetwork
	err     error
}

type controllerNetworkUpdatedMsg struct {
	network *api.ControllerNetwork
	err     error
}

type controllerNetworkCreatedMsg struct {
	network *api.ControllerNetwork
	err     error
}

type controllerNetworkDeletedMsg struct {
	id  string
	err error
}

type membersMsg struct {
	members []api.ControllerNetworkMember
	err     error
}

type memberMsg struct {
	member *api.ControllerNetworkMember
	err    error
}

type memberUpdatedMsg struct {
	member *api.ControllerNetworkMember
	err    error
}

type memberDeletedMsg struct {
	nodeID string
	err    error
}

type memberAuthorizedMsg struct {
	nodeID string
	member *api.ControllerNetworkMember
	err    error
}

type errorMsg struct {
	err error
}

type successMsg struct {
	text string
}

type tickMsg struct{}

func fetchStatus(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		s, err := client.Status(context.Background())
		return statusMsg{status: s, err: err}
	}
}

func fetchNetworks(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		n, err := client.Networks(context.Background())
		return networksMsg{networks: n, err: err}
	}
}

func fetchNetwork(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		n, err := client.Network(context.Background(), id)
		return networkMsg{network: n, err: err}
	}
}

func updateNetwork(client *api.Client, id string, net *api.Network) tea.Cmd {
	return func() tea.Msg {
		n, err := client.UpdateNetwork(context.Background(), id, net.MembershipConfig())
		return networkUpdatedMsg{network: n, err: err}
	}
}

func leaveNetwork(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.LeaveNetwork(context.Background(), id)
		return networkLeftMsg{id: id, err: err}
	}
}

func joinNetwork(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		n, err := client.UpdateNetwork(context.Background(), id, &api.Network{})
		return networkUpdatedMsg{network: n, err: err}
	}
}

func fetchPeers(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		p, err := client.Peers(context.Background())
		return peersMsg{peers: p, err: err}
	}
}

func fetchControllerStatus(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		s, err := client.ControllerStatus(context.Background())
		return controllerStatusMsg{status: s, err: err}
	}
}

func fetchControllerNetworks(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		ids, err := client.ControllerNetworks(context.Background())
		return controllerNetworksMsg{ids: ids, err: err}
	}
}

func fetchControllerNetwork(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		n, err := client.ControllerNetwork(context.Background(), id)
		return controllerNetworkMsg{network: n, err: err}
	}
}

func updateControllerNetwork(client *api.Client, id string, net *api.ControllerNetwork) tea.Cmd {
	return func() tea.Msg {
		n, err := client.UpdateControllerNetwork(context.Background(), id, net)
		return controllerNetworkUpdatedMsg{network: n, err: err}
	}
}

func createControllerNetwork(client *api.Client, controllerID string, net *api.ControllerNetwork) tea.Cmd {
	return func() tea.Msg {
		n, err := client.CreateControllerNetwork(context.Background(), controllerID, net)
		return controllerNetworkCreatedMsg{network: n, err: err}
	}
}

func deleteControllerNetwork(client *api.Client, id string) tea.Cmd {
	return func() tea.Msg {
		err := client.DeleteControllerNetwork(context.Background(), id)
		return controllerNetworkDeletedMsg{id: id, err: err}
	}
}

func fetchMembers(client *api.Client, networkID string) tea.Cmd {
	return func() tea.Msg {
		memberMap, err := client.ControllerNetworkMembers(context.Background(), networkID)
		if err != nil {
			return membersMsg{err: err}
		}
		members := make([]api.ControllerNetworkMember, 0, len(memberMap))
		for nodeID := range memberMap {
			m, err := client.ControllerNetworkMember(context.Background(), networkID, nodeID)
			if err != nil {
				return membersMsg{err: err}
			}
			members = append(members, *m)
		}
		return membersMsg{members: members}
	}
}

func updateMember(client *api.Client, networkID, nodeID string, member *api.ControllerNetworkMember) tea.Cmd {
	return func() tea.Msg {
		m, err := client.UpdateMember(context.Background(), networkID, nodeID, member)
		return memberUpdatedMsg{member: m, err: err}
	}
}

func deleteMember(client *api.Client, networkID, nodeID string) tea.Cmd {
	return func() tea.Msg {
		// ZeroTier recreates member records when a node still contacts the network.
		// Deauthorize and clear IPs before delete (per controller docs / issue #859).
		if mem, err := client.ControllerNetworkMember(context.Background(), networkID, nodeID); err == nil && mem != nil {
			mem.Authorized = false
			mem.IPAssignments = nil
			_, _ = client.UpdateMember(context.Background(), networkID, nodeID, mem)
		}
		err := client.DeleteMember(context.Background(), networkID, nodeID)
		return memberDeletedMsg{nodeID: nodeID, err: err}
	}
}

func authorizeMember(client *api.Client, networkID, nodeID string) tea.Cmd {
	return func() tea.Msg {
		m, err := client.UpdateMember(context.Background(), networkID, nodeID, &api.ControllerNetworkMember{
			Authorized: true,
		})
		return memberAuthorizedMsg{nodeID: nodeID, member: m, err: err}
	}
}

// RulesPreset returns JSON for a named flow-rule preset.
func RulesPreset(name string) string {
	switch name {
	case "allow-all":
		return `[{"type":"ACTION_ACCEPT"}]`
	case "drop-all":
		return `[{"type":"ACTION_DROP"}]`
	case "allow-network":
		return `[{"type":"MATCH_ETHERTYPE","not":true,"ethertype":"0x0806"},{"type":"MATCH_ZT_DEST","not":true,"zt":"0xffffffffffff"},{"type":"ACTION_ACCEPT"},{"type":"ACTION_DROP"}]`
	default:
		return `[]`
	}
}
