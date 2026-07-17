package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Controller != "localhost" {
		t.Fatalf("controller: %q", cfg.Controller)
	}
	if cfg.Port != 9993 {
		t.Fatalf("port: %d", cfg.Port)
	}
	if cfg.BaseURL() != "http://localhost:9993" {
		t.Fatalf("base url: %q", cfg.BaseURL())
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	path := filepath.Join(dir, "ztnui.json")
	data, _ := json.Marshal(fileConfig{
		Controller: "10.0.0.1",
		Port:       8888,
		Token:      "abc123def456789012345678",
	})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(orig) }()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Controller != "10.0.0.1" {
		t.Fatalf("controller: %q", cfg.Controller)
	}
	if cfg.Port != 8888 {
		t.Fatalf("port: %d", cfg.Port)
	}
	if cfg.Token != "abc123def456789012345678" {
		t.Fatalf("token in memory: %q", cfg.Token)
	}
	if cfg.BaseURL() != "http://10.0.0.1:8888" {
		t.Fatalf("base url: %q", cfg.BaseURL())
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(saved), "abc123") {
		t.Fatal("plaintext token still in config file after migration")
	}
}

func TestResolveTokenFromConfig(t *testing.T) {
	cfg := &Config{Token: "  mytoken  "}
	token, err := cfg.ResolveToken()
	if err != nil {
		t.Fatal(err)
	}
	if token != "mytoken" {
		t.Fatalf("token: %q", token)
	}
}

func TestSetMemberHidden(t *testing.T) {
	cfg := Default()
	cfg.SetMemberHidden("net1", "abc123def0", true)
	if !cfg.IsMemberHidden("net1", "abc123def0") {
		t.Fatal("expected member hidden")
	}
	cfg.SetMemberHidden("net1", "abc123def0", false)
	if cfg.IsMemberHidden("net1", "abc123def0") {
		t.Fatal("expected member unhidden")
	}
}

func TestDefaultTokenPathLinux(t *testing.T) {
	path := DefaultTokenPath()
	if path != "/var/lib/zerotier-one/authtoken.secret" {
		t.Fatalf("unexpected path: %q", path)
	}
}
