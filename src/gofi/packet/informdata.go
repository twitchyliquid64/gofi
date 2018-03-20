package packet

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// InformData captures the unpacked representation of an inform packet from the device.
type InformData struct {
	ConfigVersion   string `json:"cfgversion,omitempty"`
	BootVersion     string `json:"bootrom_version,omitempty"`
	Fingerprint     string `json:"fingerprint,omitempty"`
	Model           string `json:"model,omitempty"`
	ModelName       string `json:"model_display,omitempty"`
	InformURL       string `json:"inform_url,omitempty"`
	IsDefaultConfig bool   `json:"default,omitempty"`
	State           int    `json:"state,omitempty"`

	Hostname string `json:"hostname,omitempty"`
	IP       string `json:"ip,omitempty"`
	Mac      string `json:"mac,omitempty"`
	Netmask  string `json:"netmask,omitempty"`

	CountryCode int         `json:"country_code,omitempty"`
	Interfaces  []Interface `json:"if_table,omitempty"`
	RadioInfo   []RadioInfo `json:"radio_table,omitempty"`
	Nets        []Net       `json:"vap_table"`

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

	RxBytes     int64 `json:"rx_bytes"`
	RxDropped   int64 `json:"rx_dropped"`
	RxErrors    int64 `json:"rx_errors"`
	RxMulticast int64 `json:"rx_multicast"`
	RxPackets   int64 `json:"rx_packets"`

	TxBytes     int64 `json:"tx_bytes"`
	TxDropped   int64 `json:"tx_dropped"`
	TxErrors    int64 `json:"tx_errors"`
	TxMulticast int64 `json:"tx_multicast"`
	TxPackets   int64 `json:"tx_packets"`
}

// RadioInfo describes a wireless interface on the device.
type RadioInfo struct {
	BuiltinAntennaGain int             `json:"builtin_ant_gain"`
	HasInternalAntenna bool            `json:"builtin_antenna"`
	MaxTx              int             `json:"max_txpower"`
	Name               string          `json:"name"`
	Radio              string          `json:"radio"`
	DetectedNets       []NearbyNetwork `json:"scan_table"`
}

// NearbyNetwork describes a network detected by the AP.
type NearbyNetwork struct {
	Age      int    `json:"age"`
	BSSID    string `json:"bssid"`
	SSID     string `json:"essid"`
	Security string `json:"security"`

	Channel int  `json:"channel"`
	Freq    int  `json:"freq"`
	IsAdhoc bool `json:"is_adhoc"`
	RSSI    int  `json:"rssi"`
}

// Net represents information about a network being run from the AP.
type Net struct {
	BSSID   string `json:"bssid"`
	SSID    string `json:"essid"`
	CCQ     int    `json:"ccq"`
	Channel int    `json:"channel"`
	Name    string `json:"name"`
	Radio   string `json:"radio"`

	RxBytes   int64 `json:"rx_bytes"`
	TxBytes   int64 `json:"tx_bytes"`
	RxErrors  int64 `json:"rx_errors"`
	TxErrors  int64 `json:"tx_errors"`
	RxPackets int64 `json:"rx_packets"`
	TxPackets int64 `json:"tx_packets"`

	Stations []Station `json:"sta_table"`
}

// Station represents information about a station associated with the AP.
type Station struct {
	AuthTime   int64  `json:"auth_time"`
	Authorised bool   `json:"authorized"`
	Hostname   string `json:"hostname"`
	IP         string `json:"ip"`
	MAC        string `json:"mac"`
	State      int    `json:"state"`
	Uptime     int    `json:"uptime"`

	Idletime    int  `json:"idletime"`
	Is11N       bool `json:"is_11n"`
	CCQ         int  `json:"ccq"`
	Noise       int  `json:"noise"`
	RSSI        int  `json:"rssi"`
	Signal      int  `json:"signal"`
	PowerSaving bool `json:"state_pwrmgt"`
}

// UnpackInform decodes a JSON inform payload
func UnpackInform(d []byte) (*InformData, error) {
	var out InformData
	return &out, json.Unmarshal(d, &out)
}

// CommandData encapsulates the data sent in an instruction to a AP.
type CommandData struct {
	Type            string `json:"_type"`
	ServerTimestamp string `json:"server_time_in_utc,omitempty"`

	Cmd             string `json:"cmd,omitempty"`
	DatetimeRFC3339 string `json:"datetime,omitempty"`
	TimeStr         string `json:"time,omitempty"`

	Interval int `json:"interval,omitempty"`

	ConfigVersion    string `json:"cfgversion,omitempty"`
	ManagementConfig string `json:"mgmt_cfg,omitempty"`
	SystemConfig     string `json:"system_cfg,omitempty"`
	BlockedStations  string `json:"blocked_sta"`
}

// MakeNoop creates the payload section of a noop response.
func MakeNoop(pollDelay int) ([]byte, error) {
	return json.Marshal(CommandData{Type: "noop", Interval: pollDelay})
}

// MakeLocate creates the payload section of a locate response.
func MakeLocate() ([]byte, error) {
	return json.Marshal(CommandData{
		Type:            "cmd",
		Cmd:             "locate",
		ServerTimestamp: unixMicroPSTString(),
		DatetimeRFC3339: time.Now().Format(time.RFC3339),
		TimeStr:         fmt.Sprint(time.Now().Unix()),
	})
}

// MakeConfigUpdate creates the payload section of a response which sets all configuration.
func MakeConfigUpdate(sysCfg, mgmtCfg, configVersion string) ([]byte, error) {
	return json.Marshal(CommandData{
		Type:             "setparam",
		SystemConfig:     sysCfg,
		ServerTimestamp:  unixMicroPSTString(),
		ManagementConfig: mgmtCfg,
		ConfigVersion:    configVersion,
	})
}

// InformCfgVersion minimally parses the Inform payload, returning the current config version.
func InformCfgVersion(d []byte) string {
	var u struct {
		V string `json:"cfgversion"`
	}
	err := json.Unmarshal(d, &u)
	if err != nil {
		fmt.Println("InformCfgVersion() err:", err)
		return "?ERR?"
	}
	return u.V
}

//Credit: mcrute - https://github.com/mcrute/go-inform/blob/master/inform/tx_messages.go
func unixMicroPST() int64 {
	tnano := time.Now().UnixNano()
	return tnano / int64(time.Millisecond)
}

func unixMicroPSTString() string {
	return strconv.FormatInt(unixMicroPST(), 10)
}
