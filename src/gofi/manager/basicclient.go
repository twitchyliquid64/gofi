package manager

import (
	"net"
	"strings"
)

// States which can be passed to SetState()
const (
	StateUnknown int = iota
	StateAdopted
	StateManaged
)

// BasicClient is an in-memory representation of AP state.
type BasicClient struct {
	IsManagedMode bool
	IsAdoptedMode bool
	EncryptionKey []byte
	MACAddr       [6]byte
	IP            net.Addr
}

// MAC returns the MAC address of the AP.
func (c *BasicClient) MAC() [6]byte {
	return c.MACAddr
}

// SetState is called when the AP transistions to a new state.
func (c *BasicClient) SetState(state int) {
	switch state {
	case StateManaged:
		c.IsManagedMode = true
		fallthrough
	case StateAdopted:
		c.IsAdoptedMode = true
	}
}

// IsManaged returns true if the AP is 'adopted' AND configured
func (c *BasicClient) IsManaged() bool {
	return c.IsManagedMode
}

// IsAdopted returns true if the AP is 'adopted'
func (c *BasicClient) IsAdopted() bool {
	return c.IsAdoptedMode
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
