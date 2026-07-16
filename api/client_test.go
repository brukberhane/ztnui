package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusUnmarshal(t *testing.T) {
	raw := `{"address":"abc123","clock":1234,"online":true,"version":"1.14.0","versionMajor":1,"versionMinor":14,"versionRev":0,"config":{"settings":{"primaryPort":9993}}}`
	var s Status
	if err := json.Unmarshal([]byte(raw), &s); err != nil {
		t.Fatal(err)
	}
	if s.Address != "abc123" {
		t.Fatalf("address: got %q", s.Address)
	}
	if !s.Online {
		t.Fatal("expected online")
	}
	if s.Config.Settings.PrimaryPort != 9993 {
		t.Fatalf("port: got %d", s.Config.Settings.PrimaryPort)
	}
}

func TestNetworkUnmarshal(t *testing.T) {
	raw := `{"id":"565799d8f620c5c5","name":"test","status":"OK","type":"PRIVATE","allowDNS":true,"allowDefault":false,"allowGlobal":false,"allowManaged":true,"assignedAddresses":["10.147.20.1"]}`
	var n Network
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		t.Fatal(err)
	}
	if n.ID != "565799d8f620c5c5" {
		t.Fatalf("id: got %q", n.ID)
	}
	if !n.AllowDNS {
		t.Fatal("expected allowDNS")
	}
	if len(n.AssignedAddresses) != 1 {
		t.Fatalf("addresses: got %v", n.AssignedAddresses)
	}
}

func TestControllerNetworkUnmarshal(t *testing.T) {
	raw := `{"id":"3e245e31af000001","nwid":"3e245e31af000001","name":"my-net","private":true,"ipAssignmentPools":[{"ipRangeStart":"10.1.1.1","ipRangeEnd":"10.1.1.50"}],"routes":[{"target":"10.1.1.0/24","via":null}],"v4AssignMode":{"zt":true}}`
	var n ControllerNetwork
	if err := json.Unmarshal([]byte(raw), &n); err != nil {
		t.Fatal(err)
	}
	if n.Name != "my-net" {
		t.Fatalf("name: got %q", n.Name)
	}
	if len(n.IPAssignmentPools) != 1 || n.IPAssignmentPools[0].IPRangeStart != "10.1.1.1" {
		t.Fatalf("pools: %+v", n.IPAssignmentPools)
	}
}

func TestClientStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-ZT1-Auth") != "testtoken" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.URL.Path != "/status" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"address":"deadbeef","online":true,"version":"1.0.0"}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "testtoken")
	status, err := client.Status(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if status.Address != "deadbeef" {
		t.Fatalf("address: got %q", status.Address)
	}
}

func TestClientUnauthorized(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "bad")
	_, err := client.Status(t.Context())
	if err == nil {
		t.Fatal("expected error")
	}
	if !IsUnauthorized(err) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

func TestControllerNetworkMemberUnmarshal(t *testing.T) {
	raw := `{"id":"eb8d45c5c9","address":"eb8d45c5c9","name":"office-router","authorized":true,"activeBridge":true,"noAutoAssignIps":true,"ipAssignments":["10.147.20.50"]}`
	var m ControllerNetworkMember
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		t.Fatal(err)
	}
	if m.Name != "office-router" {
		t.Fatalf("name: got %q", m.Name)
	}
	if !m.ActiveBridge {
		t.Fatal("expected activeBridge")
	}
	if !m.NoAutoAssignIps {
		t.Fatal("expected noAutoAssignIps")
	}
	if len(m.IPAssignments) != 1 || m.IPAssignments[0] != "10.147.20.50" {
		t.Fatalf("ips: %v", m.IPAssignments)
	}
}

func TestClientLeaveNetwork(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/network/abc123" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "token")
	if err := client.LeaveNetwork(t.Context(), "abc123"); err != nil {
		t.Fatal(err)
	}
}
