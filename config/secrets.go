package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zalando/go-keyring"
)

const (
	keyringService = "ztnui"
	keyringUser    = "zerotier-auth-token"
)

var (
	// ErrTokenUnavailable means no token could be resolved from any source.
	ErrTokenUnavailable = errors.New("zerotier auth token unavailable")
)

func tokenEncPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "token.enc"), nil
}

func tokenKeyPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".token.key"), nil
}

// HasStoredToken reports whether an encrypted or keyring token exists.
func HasStoredToken() bool {
	if token, err := loadKeyringToken(); err == nil && token != "" {
		return true
	}
	path, err := tokenEncPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// StoreToken saves the token in the OS keyring, or encrypted on disk as fallback.
func StoreToken(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token is empty")
	}

	if err := keyring.Set(keyringService, keyringUser, token); err == nil {
		_ = os.Remove(mustTokenEncPath())
		return nil
	}
	return storeEncryptedToken(token)
}

// LoadStoredToken loads a previously saved token.
func LoadStoredToken() (string, error) {
	if token, err := loadKeyringToken(); err == nil && token != "" {
		return token, nil
	}
	return loadEncryptedToken()
}

// ClearStoredToken removes saved credentials.
func ClearStoredToken() error {
	_ = keyring.Delete(keyringService, keyringUser)
	path, err := tokenEncPath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	keyPath, err := tokenKeyPath()
	if err != nil {
		return err
	}
	_ = os.Remove(keyPath)
	return nil
}

func loadKeyringToken() (string, error) {
	token, err := keyring.Get(keyringService, keyringUser)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(token), nil
}

func mustTokenEncPath() string {
	p, err := tokenEncPath()
	if err != nil {
		return filepath.Join(os.TempDir(), "ztnui-token.enc")
	}
	return p
}

func storeEncryptedToken(token string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	key, err := loadOrCreateTokenKey()
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(token), nil)
	payload := append(nonce, ciphertext...)
	encoded := base64.RawStdEncoding.EncodeToString(payload)

	encPath, err := tokenEncPath()
	if err != nil {
		return err
	}
	return os.WriteFile(encPath, []byte(encoded), 0o600)
}

func loadEncryptedToken() (string, error) {
	encPath, err := tokenEncPath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(encPath)
	if err != nil {
		return "", err
	}

	key, err := loadOrCreateTokenKey()
	if err != nil {
		return "", err
	}

	payload, err := base64.RawStdEncoding.DecodeString(strings.TrimSpace(string(data)))
	if err != nil {
		return "", fmt.Errorf("decode token file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(payload) < gcm.NonceSize() {
		return "", errors.New("token file corrupt")
	}

	nonce, ciphertext := payload[:gcm.NonceSize()], payload[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt token: %w", err)
	}
	return strings.TrimSpace(string(plain)), nil
}

func loadOrCreateTokenKey() ([]byte, error) {
	keyPath, err := tokenKeyPath()
	if err != nil {
		return nil, err
	}
	if data, err := os.ReadFile(keyPath); err == nil && len(data) == 32 {
		return data, nil
	}

	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	if err := os.WriteFile(keyPath, key, 0o600); err != nil {
		return nil, err
	}
	return key, nil
}
