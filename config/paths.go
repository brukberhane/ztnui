package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// DefaultTokenPath returns the OS-specific path to authtoken.secret.
func DefaultTokenPath() string {
	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "ZeroTier", "authtoken.secret")
		}
		return filepath.Join(home, "Library", "Application Support", "ZeroTier", "authtoken.secret")
	case "windows":
		return filepath.Join(os.Getenv("ProgramData"), "ZeroTier", "One", "authtoken.secret")
	default:
		return "/var/lib/zerotier-one/authtoken.secret"
	}
}

// ConfigDir returns the ztnui config directory.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ztnui"), nil
}

// ConfigFilePath returns the path to ztnui.json in the config dir.
func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "ztnui.json"), nil
}
