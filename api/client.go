package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 15 * time.Second

// Client talks to the ZeroTier One service API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a client for the given base URL and auth token.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("X-ZT1-Auth", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Method:     method,
			Path:       path,
			Body:       strings.TrimSpace(string(respBody)),
		}
	}

	if out == nil || len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

// Status returns local node status.
func (c *Client) Status(ctx context.Context) (*Status, error) {
	var s Status
	if err := c.do(ctx, http.MethodGet, "/status", nil, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// Networks returns all joined networks.
func (c *Client) Networks(ctx context.Context) ([]Network, error) {
	var nets []Network
	if err := c.do(ctx, http.MethodGet, "/network", nil, &nets); err != nil {
		return nil, err
	}
	return nets, nil
}

// Network returns a single joined network.
func (c *Client) Network(ctx context.Context, id string) (*Network, error) {
	var net Network
	if err := c.do(ctx, http.MethodGet, "/network/"+id, nil, &net); err != nil {
		return nil, err
	}
	return &net, nil
}

// UpdateNetwork joins or updates a network membership.
func (c *Client) UpdateNetwork(ctx context.Context, id string, net *Network) (*Network, error) {
	var out Network
	if err := c.do(ctx, http.MethodPost, "/network/"+id, net, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LeaveNetwork removes membership from a network.
func (c *Client) LeaveNetwork(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/network/"+id, nil, nil)
}

// Peers returns all peers.
func (c *Client) Peers(ctx context.Context) ([]Peer, error) {
	var peers []Peer
	if err := c.do(ctx, http.MethodGet, "/peer", nil, &peers); err != nil {
		return nil, err
	}
	return peers, nil
}

// Peer returns a single peer.
func (c *Client) Peer(ctx context.Context, address string) (*Peer, error) {
	var peer Peer
	if err := c.do(ctx, http.MethodGet, "/peer/"+address, nil, &peer); err != nil {
		return nil, err
	}
	return &peer, nil
}

// ControllerStatus returns controller status.
func (c *Client) ControllerStatus(ctx context.Context) (*ControllerStatus, error) {
	var cs ControllerStatus
	if err := c.do(ctx, http.MethodGet, "/controller", nil, &cs); err != nil {
		return nil, err
	}
	return &cs, nil
}

// ControllerNetworks lists network IDs hosted by the controller.
func (c *Client) ControllerNetworks(ctx context.Context) ([]string, error) {
	var ids []string
	if err := c.do(ctx, http.MethodGet, "/controller/network", nil, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// ControllerNetwork returns controller network details.
func (c *Client) ControllerNetwork(ctx context.Context, id string) (*ControllerNetwork, error) {
	var net ControllerNetwork
	if err := c.do(ctx, http.MethodGet, "/controller/network/"+id, nil, &net); err != nil {
		return nil, err
	}
	return &net, nil
}

// UpdateControllerNetwork creates or updates a controller network by ID.
func (c *Client) UpdateControllerNetwork(ctx context.Context, id string, net *ControllerNetwork) (*ControllerNetwork, error) {
	var out ControllerNetwork
	if err := c.do(ctx, http.MethodPost, "/controller/network/"+id, net, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateControllerNetwork creates a network with a random ID for the controller.
func (c *Client) CreateControllerNetwork(ctx context.Context, controllerID string, net *ControllerNetwork) (*ControllerNetwork, error) {
	if controllerID == "" {
		return nil, fmt.Errorf("controller node ID is required")
	}
	paths := []string{
		"/controller/network/" + controllerID + "______",
		"/controller/network/" + controllerID,
	}
	var lastErr error
	for _, path := range paths {
		var out ControllerNetwork
		err := c.do(ctx, http.MethodPost, path, net, &out)
		if err == nil {
			return &out, nil
		}
		lastErr = err
		if !IsNotFound(err) {
			return nil, err
		}
	}
	return nil, lastErr
}

// DeleteControllerNetwork deletes a controller network.
func (c *Client) DeleteControllerNetwork(ctx context.Context, id string) error {
	return c.do(ctx, http.MethodDelete, "/controller/network/"+id, nil, nil)
}

// ControllerNetworkMembers lists member IDs and revision counters.
func (c *Client) ControllerNetworkMembers(ctx context.Context, networkID string) (MembersMap, error) {
	var members MembersMap
	if err := c.do(ctx, http.MethodGet, "/controller/network/"+networkID+"/member", nil, &members); err != nil {
		return nil, err
	}
	return members, nil
}

// ControllerNetworkMember returns a single member.
func (c *Client) ControllerNetworkMember(ctx context.Context, networkID, nodeID string) (*ControllerNetworkMember, error) {
	var member ControllerNetworkMember
	path := fmt.Sprintf("/controller/network/%s/member/%s", networkID, nodeID)
	if err := c.do(ctx, http.MethodGet, path, nil, &member); err != nil {
		return nil, err
	}
	return &member, nil
}

// UpdateMember updates a controller network member.
func (c *Client) UpdateMember(ctx context.Context, networkID, nodeID string, member *ControllerNetworkMember) (*ControllerNetworkMember, error) {
	var out ControllerNetworkMember
	path := fmt.Sprintf("/controller/network/%s/member/%s", networkID, nodeID)
	if err := c.do(ctx, http.MethodPost, path, member, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMember deletes a controller network member.
func (c *Client) DeleteMember(ctx context.Context, networkID, nodeID string) error {
	path := fmt.Sprintf("/controller/network/%s/member/%s", networkID, nodeID)
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}
