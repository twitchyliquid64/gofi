package manager

import (
	"fmt"
	"gofi/adopt"
	"gofi/packet"
	"gofi/serv"
	"strings"
)

// Manager handles controller state.
type Manager struct {
	MacAddrToKey map[[6]byte][]byte

	localAddr        string
	httpListenerAddr string
	serv             *serv.Serv
}

// New creates a new network manager (controller state).
func New(httpListenerAddr, localAddr string) (*Manager, error) {
	m := &Manager{
		MacAddrToKey:     map[[6]byte][]byte{},
		localAddr:        localAddr,
		httpListenerAddr: httpListenerAddr,
	}

	serv, err := serv.New(m, httpListenerAddr)
	if err != nil {
		return nil, err
	}
	m.serv = serv

	return m, nil
}

// Close shuts down server resources.
func (m *Manager) Close() error {
	return m.serv.Close()
}

// KeyForMac returns the data encryption key for a given MAC address.
func (m *Manager) KeyForMac(mac [6]byte) []byte {
	return m.KeyForMac(mac)
}

// Run starts the main loop for the manager.
func (m *Manager) Run() error {
	for {
		select {
		case discoveryPkt := <-m.serv.DiscoveryPackets:
			_, alreadyAdopted := m.MacAddrToKey[discoveryPkt.MAC]
			if !alreadyAdopted {
				discoveryPkt.Debug()
				adoptCfg := adopt.NewConfig(strings.Split(discoveryPkt.IPInfo.String(), ":")[0]+":22", m.localAddr+m.httpListenerAddr)
				m.MacAddrToKey[discoveryPkt.MAC] = adoptCfg.Key
				adoptErr := adopt.Adopt(adoptCfg)
				fmt.Println("adoption operation err:", adoptErr)
			}
		}
	}
}

// HandleInform is called by the server when an inform packet is recieved.
func (m *Manager) HandleInform(informPkt *packet.Inform) {
	key, keyKnown := m.MacAddrToKey[informPkt.APMAC]
	if !keyKnown {
		fmt.Printf("Cannot process inform from %x: No known key\n", key)
	} else {
		d, err := informPkt.Payload(key)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			discoveryResponse, err := packet.FormatDiscoveryResponse(d)
			if err != nil {
				fmt.Println("Error:", err)
			}
			fmt.Printf("%+v\n", discoveryResponse)
		}
	}
}
