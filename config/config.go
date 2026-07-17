package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds ztnui connection settings.
// Token is runtime-only and never written to ztnui.json in plaintext.
type Config struct {
	Controller    string              `json:"controller"`
	Port          int                 `json:"port"`
	Token         string              `json:"-"`
	HiddenMembers map[string][]string `json:"hiddenMembers,omitempty"` // networkID -> node IDs (local UI filter)
}

type fileConfig struct {
	Controller    string              `json:"controller"`
	Port          int                 `json:"port"`
	Token         string              `json:"token,omitempty"`
	HiddenMembers map[string][]string `json:"hiddenMembers,omitempty"`
}

// Default returns default configuration values.
func Default() *Config {
	return &Config{
		Controller: "localhost",
		Port:       9993,
	}
}

// BaseURL returns the controller API base URL.
func (c *Config) BaseURL() string {
	host := c.Controller
	if host == "" {
		host = "localhost"
	}
	port := c.Port
	if port == 0 {
		port = 9993
	}
	return fmt.Sprintf("http://%s:%d", host, port)
}

// ResolveToken returns token from memory, secure store, or OS default path.
func (c *Config) ResolveToken() (string, error) {
	if t := strings.TrimSpace(c.Token); t != "" {
		return t, nil
	}
	if t, err := LoadStoredToken(); err == nil && t != "" {
		c.Token = t
		return t, nil
	}
	data, err := os.ReadFile(DefaultTokenPath())
	if err != nil {
		if errors.Is(err, os.ErrPermission) || os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %v", ErrTokenUnavailable, err)
		}
		return "", fmt.Errorf("read token from %s: %w", DefaultTokenPath(), err)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", fmt.Errorf("%w: token file empty", ErrTokenUnavailable)
	}
	return token, nil
}

// SaveToken stores token securely and clears it from memory config.
func (c *Config) SaveToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token is empty")
	}
	if err := StoreToken(token); err != nil {
		return err
	}
	c.Token = token
	return nil
}

// Load reads config from cwd or ~/.config/ztnui/ztnui.json.
func Load() (*Config, error) {
	cfg := Default()

	if data, err := os.ReadFile("ztnui.json"); err == nil {
		return parseConfigFile(cfg, data, "ztnui.json")
	}

	path, err := ConfigFilePath()
	if err != nil {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	return parseConfigFile(cfg, data, path)
}

func parseConfigFile(cfg *Config, data []byte, sourcePath string) (*Config, error) {
	var fc fileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.Controller = fc.Controller
	cfg.Port = fc.Port
	if len(fc.HiddenMembers) > 0 {
		cfg.HiddenMembers = fc.HiddenMembers
	}

	// Migrate legacy plaintext token into secure storage.
	if t := strings.TrimSpace(fc.Token); t != "" {
		_ = StoreToken(t)
		cfg.Token = t
		if err := cfg.Save(); err != nil {
			return nil, err
		}
		if sourcePath != "" {
			if err := writeConfigFile(sourcePath, fileConfig{
				Controller:    fc.Controller,
				Port:          fc.Port,
				HiddenMembers: fc.HiddenMembers,
			}); err != nil {
				return nil, err
			}
		}
	}
	return cfg, nil
}

func writeConfigFile(path string, fc fileConfig) error {
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

// HiddenMembersSet returns hidden node IDs for a network as a lookup set.
func (c *Config) HiddenMembersSet(networkID string) map[string]bool {
	set := make(map[string]bool)
	if c == nil || c.HiddenMembers == nil {
		return set
	}
	for _, id := range c.HiddenMembers[networkID] {
		set[id] = true
	}
	return set
}

// IsMemberHidden reports whether a node is hidden for a network.
func (c *Config) IsMemberHidden(networkID, nodeID string) bool {
	return c.HiddenMembersSet(networkID)[nodeID]
}

// SetMemberHidden adds or removes a locally hidden member for a network.
func (c *Config) SetMemberHidden(networkID, nodeID string, hidden bool) {
	if c.HiddenMembers == nil {
		c.HiddenMembers = make(map[string][]string)
	}
	ids := make([]string, 0, len(c.HiddenMembers[networkID]))
	for _, id := range c.HiddenMembers[networkID] {
		if id != nodeID {
			ids = append(ids, id)
		}
	}
	if hidden {
		ids = append(ids, nodeID)
	}
	if len(ids) == 0 {
		delete(c.HiddenMembers, networkID)
	} else {
		c.HiddenMembers[networkID] = ids
	}
}

// Save writes non-secret settings to ~/.config/ztnui/ztnui.json.
func (c *Config) Save() error {
	path, err := ConfigFilePath()
	if err != nil {
		return err
	}
	fc := fileConfig{
		Controller:    c.Controller,
		Port:          c.Port,
		HiddenMembers: c.HiddenMembers,
	}
	return writeConfigFile(path, fc)
}
