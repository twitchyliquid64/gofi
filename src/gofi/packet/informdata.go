package packet

import "encoding/json"

// InformData captures the unpacked representation of an inform packet from the device.
type InformData struct {
	BootVersion string `json:"bootrom_version,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
	Hostname    string `json:"hostname,omitempty"`

	Interfaces []Interface `json:"if_table,omitempty"`
	RadioInfo  []RadioInfo `json:"radio_table,omitempty"`

	IsDiscovery bool `json:"discovery_response,omitempty"`
}

// Interface captures the interface information reported in a Inform packet.
type Interface struct {
	Drops      int    `json:"drops"`
	FullDuplex bool   `json:"full_duplex"`
	IP         string `json:"ip"`
	Latency    int    `json:"latency"`
	MAC        string `json:"mac"`
	Name       string `json:"name"`
	Netmask    string `json:"netmask"`
	NumPorts   int    `json:"num_port"`
	Speed      int    `json:"speed"`
	Up         bool   `json:"up"`
	Uptime     int    `json:"uptime"`
}

// RadioInfo describes a wireless interface on the device.
type RadioInfo struct {
	BuiltinAntennaGain int    `json:"builtin_ant_gain"`
	HasInternalAntenna bool   `json:"builtin_antenna"`
	MaxTx              int    `json:"max_txpower"`
	Name               string `json:"name"`
	Radio              string `json:"radio"`
}

// FormatDiscoveryResponse decodes a JSON inform payload representing a discoveryResponse packet
func FormatDiscoveryResponse(d []byte) (*InformData, error) {
	var out InformData
	return &out, json.Unmarshal(d, &out)
}
