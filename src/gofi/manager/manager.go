package manager

import (
	"errors"
	"fmt"
	"gofi/adopt"
	"gofi/packet"
	"gofi/serv"
	"strings"
)

// ap represents the state of an ap.
type ap interface {
	MAC() [6]byte
	IsManaged() bool
	AuthKey() []byte
	SSHPw() string
	GetIP() string
}

type discoveryStateInitialiser func(string, string, *packet.Discovery) (ap, *adopt.Config, error)

// Manager handles controller state.
type Manager struct {
	MacAddrToKey map[[6]byte]ap

	localAddr        string
	httpListenerAddr string
	serv             *serv.Serv

	discoveryInitializer discoveryStateInitialiser
}

// defaultStateInitializer stores AP state in memory.
func defaultStateInitializer(localAddr, listenerAddr string, discoveryPkt *packet.Discovery) (ap, *adopt.Config, error) {
	discoveryPkt.Debug()
	adoptCfg := adopt.NewConfig(strings.Split(discoveryPkt.IPInfo.String(), ":")[0]+":22", localAddr+listenerAddr, "ubnt")
	return &BasicClient{
		EncryptionKey: adoptCfg.Key,
		MACAddr:       discoveryPkt.MAC,
		IP:            discoveryPkt.IPInfo,
	}, adoptCfg, nil
}

// New creates a new AP manager (controller state).
func New(httpListenerAddr, localAddr string, stateInitializer discoveryStateInitialiser) (*Manager, error) {
	m := &Manager{
		MacAddrToKey:         map[[6]byte]ap{},
		localAddr:            localAddr,
		httpListenerAddr:     httpListenerAddr,
		discoveryInitializer: defaultStateInitializer,
	}
	if stateInitializer != nil {
		m.discoveryInitializer = stateInitializer
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

// Run starts the main loop for the manager.
func (m *Manager) Run() error {
	for {
		select {
		case discoveryPkt := <-m.serv.DiscoveryPackets:
			_, alreadyAdopted := m.MacAddrToKey[discoveryPkt.MAC]
			if !alreadyAdopted {
				accessPoint, adoptCfg, err := m.discoveryInitializer(m.localAddr, m.httpListenerAddr, discoveryPkt)
				if err != nil {
					fmt.Printf("[DISCOVERY] State initializer returned error: %s\n", err)
					fmt.Printf("[DISCOVERY] Aborting processing of discovery from %s\n", discoveryPkt.IPInfo)
					continue
				}
				adoptErr := adopt.Adopt(adoptCfg)
				if adoptErr != nil {
					fmt.Printf("[ADOPT] Adopt failed: %s\n", adoptErr)
					continue
				}
				fmt.Printf("[DISCOVERY] Adoption for %s successful.\n", discoveryPkt.IPInfo)
				m.MacAddrToKey[accessPoint.MAC()] = accessPoint
			}
		}
	}
}

// HandleInform is called by the server when an inform packet is recieved.
func (m *Manager) HandleInform(informPkt *packet.Inform) ([]byte, error) {
	accessPoint, apKnown := m.MacAddrToKey[informPkt.APMAC]
	if !apKnown {
		return nil, errors.New("AP is not known")
	}

	d, err := informPkt.Payload(accessPoint.AuthKey())
	if err != nil {
		return nil, err
	}
	fmt.Println(string(d))
	if !accessPoint.IsManaged() {
		return m.handleDiscoveryInform(informPkt, accessPoint, d)
	}
	return nil, errors.New("Don't know how to handle inform for given AP state")
}

func (m *Manager) handleDiscoveryInform(informPkt *packet.Inform, accessPoint ap, d []byte) ([]byte, error) {
	discoveryResponse, err := packet.FormatDiscoveryResponse(d)
	if err != nil {
		return nil, err
	}
	if !discoveryResponse.IsDiscovery {
		return nil, errors.New("Expected discovery response")
	}

	config, err := GetConfig(accessPoint.GetIP(), accessPoint.SSHPw())
	if err != nil {
		return nil, err
	}
	fmt.Println(string(config))

	// we reuse the existing packet structure
	reply := informPkt.CloneForReply()
	reply.Data, err = packet.MakeNoop(5)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(reply.Data))
	return reply.Marshal(accessPoint.AuthKey())
}
