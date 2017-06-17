package manager

import (
	"net"
	"strings"
)

// BasicClient is an in-memory representation of AP state.
type BasicClient struct {
	IsManagedMode bool
	EncryptionKey []byte
	MACAddr       [6]byte
	IP            net.Addr
}

// MAC returns the MAC address of the AP.
func (c *BasicClient) MAC() [6]byte {
	return c.MACAddr
}

// IsManaged returns true if the AP is 'adopted'
func (c *BasicClient) IsManaged() bool {
	return c.IsManagedMode
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
