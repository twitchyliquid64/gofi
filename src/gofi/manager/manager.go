package manager

import (
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

	sysconf, mgmtconf, err := GetConfig(accessPoint.GetIP(), accessPoint.SSHPw()) //fetch config from AP
	if err != nil {
		return nil, err
	}
	newSysConf, newMgmtConf, err := m.networkConfig.Generate(sysconf, mgmtconf) //Make modifications based on desired settings
	if err != nil {
		return nil, err
	}

	reply := informPkt.CloneForReply()
	if !accessPoint.IsAdopted() { //havent sent just the mgmt yet
		fmt.Printf("[INFORM-PROVISION] Sending management configuration to %x\n", accessPoint.MAC())
		reply.Data, err = packet.MakeMgmtConfigUpdate(newMgmtConf)
		accessPoint.SetState(StateAdopted)
	} else if !accessPoint.IsManaged() { //mgmt sent, system not yet sent
		fmt.Printf("[INFORM-PROVISION] Sending system configuration to %x via SSH\n", accessPoint.MAC())
		err = setSystemConfig(accessPoint.GetIP(), accessPoint.SSHPw(), newSysConf)
		if err != nil {
			return nil, err
		}
		err = applyConfig(accessPoint.GetIP(), accessPoint.SSHPw())
		if err != nil {
			return nil, err
		}
		accessPoint.SetState(StateManaged)
		reply.Data, err = packet.MakeNoop(3)
	}
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(reply.Data))
	return reply.Marshal(accessPoint.AuthKey())
}
