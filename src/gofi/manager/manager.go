package manager

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"gofi/adopt"
	"gofi/config"
	"gofi/packet"
	"gofi/serv"
	"strings"
)

// ap represents the state of an ap.
type ap interface {
	MAC() [6]byte
	IsAdopted() bool
	IsManaged() bool
	SetState(int)
	AuthKey() []byte
	SSHPw() string
	GetIP() string
	GetConfigVersion() string
	SetConfigVersion(string)
}

type discoveryStateInitialiser func(string, string, *packet.Discovery) (ap, *adopt.Config, error)

// Manager handles controller state.
type Manager struct {
	MacAddrToKey map[[6]byte]ap

	localAddr        string
	httpListenerAddr string
	serv             *serv.Serv
	networkConfig    *config.Config

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
		networkConfig:        &config.Config{},
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
				m.MacAddrToKey[accessPoint.MAC()] = accessPoint
				adoptErr := adopt.Adopt(adoptCfg)
				if adoptErr != nil {
					fmt.Printf("[ADOPT] Adopt failed: %s\n", adoptErr)
					continue
				}
				fmt.Printf("[DISCOVERY] Adoption for %s successful.\n", discoveryPkt.IPInfo)
			}
		}
	}
}

// HandleInform is called by the server when an inform packet is recieved.GenerateMgmtConf
func (m *Manager) HandleInform(informPkt *packet.Inform) ([]byte, error) {
	accessPoint, apKnown := m.MacAddrToKey[informPkt.APMAC]
	if !apKnown {
		return nil, errors.New("AP is not known")
	}

	d, err := informPkt.Payload(accessPoint.AuthKey())
	if err != nil {
		return nil, err
	}
	if !accessPoint.IsManaged() {
		return m.handleDiscoveryInform(informPkt, accessPoint, d)
	}
	return m.handleNormalInform(informPkt, accessPoint, d)
}

func (m *Manager) handleNormalInform(informPkt *packet.Inform, accessPoint ap, d []byte) ([]byte, error) {
	reply := informPkt.CloneForReply()
	var err error
	reply.Data, err = packet.MakeNoop(3)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[INFORM] Handled nominal inform for %x\n", accessPoint.MAC())
	return reply.Marshal(accessPoint.AuthKey())
}

func (m *Manager) handleDiscoveryInform(informPkt *packet.Inform, accessPoint ap, d []byte) ([]byte, error) {
	discoveryResponse, err := packet.FormatDiscoveryResponse(d)
	if err != nil {
		return nil, err
	}
	if !discoveryResponse.IsDiscovery {
		return nil, errors.New("Expected discovery response")
	}

	reply := informPkt.CloneForReply()
	if !accessPoint.IsAdopted() { //havent sent just the mgmt yet
		var r []byte
		r, err = GenerateRandomBytes(8)
		if err != nil {
			return nil, err
		}
		accessPoint.SetConfigVersion(hex.EncodeToString(r))
		fmt.Printf("[INFORM-PROVISION] New configuration version %q for %x\n", hex.EncodeToString(r), accessPoint.MAC())

		fmt.Printf("[INFORM-PROVISION] Sending management configuration to %x\n", accessPoint.MAC())
		var mgmtConf string
		mgmtConf, err = m.networkConfig.GenerateMgmtConf(hex.EncodeToString(accessPoint.AuthKey()), accessPoint.GetConfigVersion(), m.localAddr, m.httpListenerAddr)
		if err != nil {
			return nil, err
		}
		fmt.Println(mgmtConf)
		reply.Data, err = packet.MakeMgmtConfigUpdate(mgmtConf, accessPoint.GetConfigVersion())
		accessPoint.SetState(StateAdopted)
	} else if !accessPoint.IsManaged() { //mgmt sent, system not yet sent
		fmt.Printf("[INFORM-PROVISION] Sending system configuration to %x via HTTP response\n", accessPoint.MAC())
		sysconf, err := GetSysConfig(accessPoint.GetIP(), accessPoint.SSHPw()) //fetch config from AP
		if err != nil {
			return nil, err
		}
		newSysConf, err := m.networkConfig.GenerateSysConf(sysconf, accessPoint.GetConfigVersion()) //Make modifications based on desired settings
		if err != nil {
			return nil, err
		}

		var mgmtConf string
		mgmtConf, err = m.networkConfig.GenerateMgmtConf(hex.EncodeToString(accessPoint.AuthKey()), accessPoint.GetConfigVersion(), m.localAddr, m.httpListenerAddr)
		if err != nil {
			return nil, err
		}
		accessPoint.SetState(StateManaged)
		reply.Data, err = packet.MakeConfigUpdate(newSysConf, mgmtConf, accessPoint.GetConfigVersion())
		fmt.Println(string(reply.Data))
	}
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(reply.Data))
	return reply.Marshal(accessPoint.AuthKey())
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
// Sauce: https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand/
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
