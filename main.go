package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/brukberhane/ztnui/api"
	"github.com/brukberhane/ztnui/config"
	"github.com/brukberhane/ztnui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	token, err := cfg.ResolveToken()
	needsAuth := false
	authReason := ""
	var client *api.Client

	if err != nil {
		if errors.Is(err, config.ErrTokenUnavailable) || os.IsPermission(err) {
			needsAuth = true
			authReason = fmt.Sprintf(
				"Cannot read %s.\nPaste token from that file, or run with permissions to read it.",
				config.DefaultTokenPath(),
			)
		} else {
			needsAuth = true
			authReason = err.Error()
		}
	} else {
		client = api.NewClient(cfg.BaseURL(), token)
	}

	model := ui.NewModel(cfg, client, needsAuth, authReason)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
