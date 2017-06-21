package manager

import (
	"net"
	"strings"
)

// BasicClient is an in-memory representation of AP state.
type BasicClient struct {
	state         int
	EncryptionKey []byte
	MACAddr       [6]byte
	IP            net.Addr
	CfgVersion    string
}

// MAC returns the MAC address of the AP.
func (c *BasicClient) MAC() [6]byte {
	return c.MACAddr
}

// SetState is called when the AP transistions to a new state.
func (c *BasicClient) SetState(state int) {
	c.state = state
}

// GetConfigVersion fetches the config version.
func (c *BasicClient) GetConfigVersion() string {
	return c.CfgVersion
}

// SetConfigVersion stores the config version.
func (c *BasicClient) SetConfigVersion(cfgv string) {
	c.CfgVersion = cfgv
}

// GetState returns the state of the device.
func (c *BasicClient) GetState() int {
	return c.state
}

// AuthKey returns the authentication key for all communications.
func (c *BasicClient) AuthKey() []byte {
	return c.EncryptionKey
}

// SSHPw returns the password to login via SSH.
func (c *BasicClient) SSHPw() string {
	return "ubnt"
}

// GetIP returns the IP as a string.
func (c *BasicClient) GetIP() string {
	return strings.Split(c.IP.String(), ":")[0]
}
