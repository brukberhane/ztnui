package ui

import (
	"context"
	"strings"

	"github.com/brukberhane/ztnui/api"
	"github.com/charmbracelet/bubbles/textinput"
)

type authModel struct {
	input  textinput.Model
	reason string
}

func newAuthModel(reason string) authModel {
	ti := textinput.New()
	ti.Placeholder = "paste 24-char token"
	ti.CharLimit = 128
	ti.Width = 48
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()
	return authModel{input: ti, reason: reason}
}

func (a authModel) View() string {
	var b strings.Builder
	b.WriteString(HeaderStyle.Render("Authentication required"))
	b.WriteString("\n\n")
	if a.reason != "" {
		b.WriteString(SubtitleStyle.Render(a.reason))
		b.WriteString("\n\n")
	}
	b.WriteString("Paste ZeroTier auth token (from authtoken.secret):\n\n")
	b.WriteString("  ")
	b.WriteString(a.input.View())
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("enter save  esc quit  ctrl+c quit"))
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Token stored encrypted (OS keyring or local AES-GCM file)"))
	return b.String()
}

func validateToken(client *api.Client) error {
	_, err := client.Status(context.Background())
	return err
}
