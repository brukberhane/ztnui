package api

import "encoding/json"

// Status is the local node status from GET /status.
type Status struct {
	Address              string        `json:"address"`
	Clock                int64         `json:"clock"`
	Config               StatusConfig  `json:"config"`
	Online               bool          `json:"online"`
	PlanetWorldID        int64         `json:"planetWorldId"`
	PlanetWorldTimestamp int64         `json:"planetWorldTimestamp"`
	PublicIdentity       string        `json:"publicIdentity"`
	TCPFallbackActive    bool          `json:"tcpFallbackActive"`
	Version              string        `json:"version"`
	VersionBuild         int           `json:"versionBuild"`
	VersionMajor         int           `json:"versionMajor"`
	VersionMinor         int           `json:"versionMinor"`
	VersionRev           int           `json:"versionRev"`
}

type StatusConfig struct {
	Settings StatusSettings `json:"settings"`
}

type StatusSettings struct {
	AllowTCPFallbackRelay bool     `json:"allowTcpFallbackRelay"`
	PortMappingEnabled    bool     `json:"portMappingEnabled"`
	PrimaryPort           int      `json:"primaryPort"`
	ListeningOn           []string `json:"listeningOn,omitempty"`
	SurfaceAddresses      []string `json:"surfaceAddresses,omitempty"`
	SecondaryPort         int      `json:"secondaryPort,omitempty"`
	TertiaryPort          int      `json:"tertiaryPort,omitempty"`
	ForceTCPRelay         bool     `json:"forceTcpRelay,omitempty"`
	SoftwareUpdate        string   `json:"softwareUpdate,omitempty"`
	SoftwareUpdateChannel string   `json:"softwareUpdateChannel,omitempty"`
}

// Network is a joined network membership from GET/POST /network.
type Network struct {
	AllowDNS               bool                    `json:"allowDNS"`
	AllowDefault           bool                    `json:"allowDefault"`
	AllowGlobal            bool                    `json:"allowGlobal"`
	AllowManaged           bool                    `json:"allowManaged"`
	AssignedAddresses      []string                `json:"assignedAddresses,omitempty"`
	Bridge                 bool                    `json:"bridge"`
	BroadcastEnabled       bool                    `json:"broadcastEnabled"`
	DNS                    NetworkDNS              `json:"dns,omitempty"`
	ID                     string                  `json:"id"`
	MAC                    string                  `json:"mac,omitempty"`
	MTU                    int                     `json:"mtu,omitempty"`
	MulticastSubscriptions []MulticastSubscription `json:"multicastSubscriptions,omitempty"`
	Name                   string                  `json:"name,omitempty"`
	NetconfRevision        int                     `json:"netconfRevision,omitempty"`
	PortDeviceName         string                  `json:"portDeviceName,omitempty"`
	PortError              int                     `json:"portError,omitempty"`
	Routes                 []NetworkRoute          `json:"routes,omitempty"`
	Status                 string                  `json:"status,omitempty"`
	Type                   string                  `json:"type,omitempty"`
}

type NetworkDNS struct {
	Domain  string   `json:"domain,omitempty"`
	Servers []string `json:"servers,omitempty"`
}

type MulticastSubscription struct {
	ADI int64  `json:"adi"`
	MAC string `json:"mac"`
}

type NetworkRoute struct {
	Flags  int    `json:"flags,omitempty"`
	Metric int    `json:"metric,omitempty"`
	Target string `json:"target"`
	Via    string `json:"via,omitempty"`
}

// Peer is a peer from GET /peer.
type Peer struct {
	Address      string     `json:"address"`
	IsBonded     bool       `json:"isBonded"`
	Latency      int        `json:"latency"`
	Paths        []PeerPath `json:"paths,omitempty"`
	Role         string     `json:"role,omitempty"`
	Version      string     `json:"version,omitempty"`
	VersionMajor int        `json:"versionMajor,omitempty"`
	VersionMinor int        `json:"versionMinor,omitempty"`
	VersionRev   int        `json:"versionRev,omitempty"`
}

type PeerPath struct {
	Active        bool   `json:"active"`
	Address       string `json:"address"`
	Expired       bool   `json:"expired"`
	LastReceive   int64  `json:"lastReceive"`
	LastSend      int64  `json:"lastSend"`
	Preferred     bool   `json:"preferred"`
	TrustedPathID int    `json:"trustedPathId"`
}

// ControllerStatus from GET /controller.
type ControllerStatus struct {
	Controller bool  `json:"controller"`
	APIVersion int   `json:"apiVersion"`
	Clock      int64 `json:"clock"`
}

// ControllerNetwork from controller endpoints.
type ControllerNetwork struct {
	ID                string            `json:"id,omitempty"`
	NwID              string            `json:"nwid,omitempty"`
	ObjType           string            `json:"objtype,omitempty"`
	Name              string            `json:"name,omitempty"`
	CreationTime      float64           `json:"creationTime,omitempty"`
	Private           bool              `json:"private,omitempty"`
	EnableBroadcast   bool              `json:"enableBroadcast,omitempty"`
	V4AssignMode      V4AssignMode      `json:"v4AssignMode,omitempty"`
	V6AssignMode      V6AssignMode      `json:"v6AssignMode,omitempty"`
	MTU               int               `json:"mtu,omitempty"`
	MulticastLimit    int               `json:"multicastLimit,omitempty"`
	Revision          int               `json:"revision,omitempty"`
	Routes            []Route           `json:"routes,omitempty"`
	IPAssignmentPools []IPAssignmentPool `json:"ipAssignmentPools,omitempty"`
	Rules             []json.RawMessage `json:"rules,omitempty"`
	Capabilities      []json.RawMessage `json:"capabilities,omitempty"`
	Tags              []json.RawMessage `json:"tags,omitempty"`
	DNS               ControllerDNS     `json:"dns,omitempty"`
	RemoteTraceTarget string            `json:"remoteTraceTarget,omitempty"`
	RemoteTraceLevel  int               `json:"remoteTraceLevel,omitempty"`
}

type V4AssignMode struct {
	ZT bool `json:"zt,omitempty"`
}

type V6AssignMode struct {
	SixPlane bool `json:"6plane,omitempty"`
	RFC4193  bool `json:"rfc4193,omitempty"`
	ZT       bool `json:"zt,omitempty"`
}

type Route struct {
	Target string  `json:"target"`
	Via    *string `json:"via"`
}

type IPAssignmentPool struct {
	IPRangeStart string `json:"ipRangeStart"`
	IPRangeEnd   string `json:"ipRangeEnd"`
}

type ControllerDNS struct {
	Domain  string   `json:"domain,omitempty"`
	Servers []string `json:"servers,omitempty"`
}

// ControllerNetworkMember from controller member endpoints.
type ControllerNetworkMember struct {
	ID             string   `json:"id,omitempty"`
	Address        string   `json:"address,omitempty"`
	NwID           string   `json:"nwid,omitempty"`
	Name           string   `json:"name,omitempty"`
	Authorized     bool     `json:"authorized,omitempty"`
	ActiveBridge   bool     `json:"activeBridge,omitempty"`
	NoAutoAssignIps bool    `json:"noAutoAssignIps,omitempty"`
	Identity       string   `json:"identity,omitempty"`
	IPAssignments  []string `json:"ipAssignments,omitempty"`
	Revision       int      `json:"revision,omitempty"`
	VMajor         int      `json:"vMajor,omitempty"`
	VMinor         int      `json:"vMinor,omitempty"`
	VRev           int      `json:"vRev,omitempty"`
	VProto         int      `json:"vProto,omitempty"`
}

// MembershipConfig returns only the client-side toggles for POST /network/{id}.
func (n *Network) MembershipConfig() *Network {
	return &Network{
		AllowDNS:     n.AllowDNS,
		AllowDefault: n.AllowDefault,
		AllowGlobal:  n.AllowGlobal,
		AllowManaged: n.AllowManaged,
	}
}

// MembersMap is the response from GET /controller/network/{id}/member.
type MembersMap map[string]int
