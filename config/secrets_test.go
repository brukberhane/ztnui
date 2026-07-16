package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncryptedTokenRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	token := "0123456789abcdef012345678"
	if err := StoreToken(token); err != nil {
		t.Fatal(err)
	}

	got, err := LoadStoredToken()
	if err != nil {
		t.Fatal(err)
	}
	if got != token {
		t.Fatalf("token: got %q want %q", got, token)
	}

	encPath := filepath.Join(dir, ".config", "ztnui", "token.enc")
	if data, err := os.ReadFile(encPath); err == nil {
		if strings.Contains(string(data), token) {
			t.Fatal("token stored in plaintext on disk")
		}
	}
}

func TestClearStoredToken(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	if err := StoreToken("abc123def456789012345678"); err != nil {
		t.Fatal(err)
	}
	if err := ClearStoredToken(); err != nil {
		t.Fatal(err)
	}
	if HasStoredToken() {
		t.Fatal("expected no stored token after clear")
	}
}
