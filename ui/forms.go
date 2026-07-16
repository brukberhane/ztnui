package ui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/brukberhane/ztnui/api"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
)

// networkForm holds editable controller network fields.
type networkForm struct {
	name            textinput.Model
	private         bool
	enableBroadcast bool
	mtu             textinput.Model
	multicastLimit  textinput.Model
	v4ZT            bool
	v6SixPlane      bool
	v6RFC4193       bool
	v6ZT            bool
	poolStart       textinput.Model
	poolEnd         textinput.Model
	routes          textinput.Model
	dnsDomain       textinput.Model
	dnsServers      textinput.Model
	rules           textarea.Model
	focusIndex      int
	editing         bool
	networkID       string
}

func newNetworkForm() networkForm {
	f := networkForm{
		name:           newInput("Network name"),
		mtu:            newInput("2800"),
		multicastLimit: newInput("32"),
		poolStart:      newInput("10.147.20.1"),
		poolEnd:        newInput("10.147.20.254"),
		routes:         newInput("10.147.20.0/24"),
		dnsDomain:      newInput("zt.local"),
		dnsServers:     newInput("10.147.20.1"),
		private:        true,
		v4ZT:           true,
	}
	f.rules = textarea.New()
	f.rules.Placeholder = `[{"type":"ACTION_ACCEPT"}]`
	f.rules.SetHeight(4)
	f.rules.SetWidth(60)
	f.name.Focus()
	return f
}

func newInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 256
	ti.Width = 40
	return ti
}

func (f *networkForm) loadFrom(net *api.ControllerNetwork) {
	f.networkID = net.NwID
	if f.networkID == "" {
		f.networkID = net.ID
	}
	f.name.SetValue(net.Name)
	f.private = net.Private
	f.enableBroadcast = net.EnableBroadcast
	if net.MTU > 0 {
		f.mtu.SetValue(strconv.Itoa(net.MTU))
	}
	if net.MulticastLimit > 0 {
		f.multicastLimit.SetValue(strconv.Itoa(net.MulticastLimit))
	}
	f.v4ZT = net.V4AssignMode.ZT
	f.v6SixPlane = net.V6AssignMode.SixPlane
	f.v6RFC4193 = net.V6AssignMode.RFC4193
	f.v6ZT = net.V6AssignMode.ZT
	if len(net.IPAssignmentPools) > 0 {
		f.poolStart.SetValue(net.IPAssignmentPools[0].IPRangeStart)
		f.poolEnd.SetValue(net.IPAssignmentPools[0].IPRangeEnd)
	}
	var routeStrs []string
	for _, r := range net.Routes {
		via := ""
		if r.Via != nil {
			via = *r.Via
		}
		routeStrs = append(routeStrs, fmt.Sprintf("%s;%s", r.Target, via))
	}
	f.routes.SetValue(strings.Join(routeStrs, ","))
	f.dnsDomain.SetValue(net.DNS.Domain)
	f.dnsServers.SetValue(strings.Join(net.DNS.Servers, ","))
	if len(net.Rules) > 0 {
		data, _ := json.MarshalIndent(net.Rules, "", "  ")
		f.rules.SetValue(string(data))
	}
}

func (f *networkForm) toControllerNetwork() (*api.ControllerNetwork, error) {
	mtu, _ := strconv.Atoi(strings.TrimSpace(f.mtu.Value()))
	ml, _ := strconv.Atoi(strings.TrimSpace(f.multicastLimit.Value()))

	var routes []api.Route
	for _, part := range strings.Split(f.routes.Value(), ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		fields := strings.SplitN(part, ";", 2)
		target := strings.TrimSpace(fields[0])
		var via *string
		if len(fields) == 2 && strings.TrimSpace(fields[1]) != "" {
			v := strings.TrimSpace(fields[1])
			via = &v
		}
		routes = append(routes, api.Route{Target: target, Via: via})
	}

	var rules []json.RawMessage
	rulesText := strings.TrimSpace(f.rules.Value())
	if rulesText != "" {
		if err := json.Unmarshal([]byte(rulesText), &rules); err != nil {
			return nil, fmt.Errorf("invalid rules JSON: %w", err)
		}
	}

	net := &api.ControllerNetwork{
		Name:            strings.TrimSpace(f.name.Value()),
		Private:         f.private,
		EnableBroadcast: f.enableBroadcast,
		MTU:             mtu,
		MulticastLimit:  ml,
		V4AssignMode:    api.V4AssignMode{ZT: f.v4ZT},
		V6AssignMode: api.V6AssignMode{
			SixPlane: f.v6SixPlane,
			RFC4193:  f.v6RFC4193,
			ZT:       f.v6ZT,
		},
		IPAssignmentPools: []api.IPAssignmentPool{
			{
				IPRangeStart: strings.TrimSpace(f.poolStart.Value()),
				IPRangeEnd:   strings.TrimSpace(f.poolEnd.Value()),
			},
		},
		Routes: routes,
		DNS: api.ControllerDNS{
			Domain:  strings.TrimSpace(f.dnsDomain.Value()),
			Servers: splitCSV(f.dnsServers.Value()),
		},
		Rules: rules,
	}
	return net, nil
}

func (f *networkForm) focusCount() int {
	return 14 // name..rules fields
}

func (f *networkForm) blurAll() {
	f.name.Blur()
	f.mtu.Blur()
	f.multicastLimit.Blur()
	f.poolStart.Blur()
	f.poolEnd.Blur()
	f.routes.Blur()
	f.dnsDomain.Blur()
	f.dnsServers.Blur()
	f.rules.Blur()
}

func (f *networkForm) focusField(idx int) {
	f.blurAll()
	f.focusIndex = idx
	switch idx {
	case 0:
		f.name.Focus()
	case 3:
		f.mtu.Focus()
	case 4:
		f.multicastLimit.Focus()
	case 8:
		f.poolStart.Focus()
	case 9:
		f.poolEnd.Focus()
	case 10:
		f.routes.Focus()
	case 11:
		f.dnsDomain.Focus()
	case 12:
		f.dnsServers.Focus()
	case 13:
		f.rules.Focus()
	}
}

func (f *networkForm) nextFocus() {
	f.focusField((f.focusIndex + 1) % f.focusCount())
}

func (f *networkForm) prevFocus() {
	f.focusField((f.focusIndex - 1 + f.focusCount()) % f.focusCount())
}

func (f *networkForm) toggleCurrent() {
	switch f.focusIndex {
	case 1:
		f.private = !f.private
	case 2:
		f.enableBroadcast = !f.enableBroadcast
	case 5:
		f.v4ZT = !f.v4ZT
	case 6:
		f.v6SixPlane = !f.v6SixPlane
	case 7:
		f.v6RFC4193 = !f.v6RFC4193
	}
}

// settingsForm for connection config.
type settingsForm struct {
	controller textinput.Model
	port       textinput.Model
	token      textinput.Model
	focusIndex int
}

func newSettingsForm() settingsForm {
	s := settingsForm{
		controller: newInput("localhost"),
		port:       newInput("9993"),
		token:      newInput("paste to store securely"),
	}
	s.token.EchoMode = textinput.EchoPassword
	s.token.EchoCharacter = '•'
	s.controller.Focus()
	return s
}

func (s *settingsForm) blurAll() {
	s.controller.Blur()
	s.port.Blur()
	s.token.Blur()
}

func (s *settingsForm) focusField(idx int) {
	s.blurAll()
	s.focusIndex = idx
	switch idx {
	case 0:
		s.controller.Focus()
	case 1:
		s.port.Focus()
	case 2:
		s.token.Focus()
	}
}
