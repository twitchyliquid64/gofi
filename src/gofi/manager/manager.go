package manager

// Implements main controller logic.

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

	"github.com/kylelemons/godebug/pretty"
)

// States which can be passed to SetState()
const (
	StateUnknown int = iota
	StateAdopting
	StateAdopted
	StateProvisioning
	StateManaged
)

// AP represents the controller state of an access point.
type AP interface {
	MAC() [6]byte
	GetState() int
	SetState(int)
	AuthKey() []byte
	SSHPw() string
	GetIP() string
	GetConfigVersion() string
	SetConfigVersion(string)
}

type discoveryStateInitialiser func(string, string, *packet.Discovery) (AP, *adopt.Config, error)

// Manager handles controller state.
type Manager struct {
	MacAddrToKey map[[6]byte]AP

	localAddr        string
	httpListenerAddr string
	serv             *serv.Serv
	networkConfig    *config.Config

	discoveryInitializer discoveryStateInitialiser
}

// defaultStateInitializer stores AP state in memory.
func defaultStateInitializer(localAddr, listenerAddr string, discoveryPkt *packet.Discovery) (AP, *adopt.Config, error) {
	discoveryPkt.Debug()
	adoptCfg := adopt.NewConfig(strings.Split(discoveryPkt.IPInfo.String(), ":")[0]+":22", localAddr+listenerAddr, "ubnt")
	return &BasicClient{
		EncryptionKey: adoptCfg.Key,
		MACAddr:       discoveryPkt.MAC,
		IP:            discoveryPkt.IPInfo,
	}, adoptCfg, nil
}

// New creates a new AP manager (controller state).
func New(httpListenerAddr, localAddr string, conf *config.Config, stateInitializer discoveryStateInitialiser) (*Manager, error) {
	m := &Manager{
		MacAddrToKey:         map[[6]byte]AP{},
		localAddr:            localAddr,
		httpListenerAddr:     httpListenerAddr,
		discoveryInitializer: defaultStateInitializer,
		networkConfig:        conf,
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
				setAPConfigDirty(accessPoint)
				accessPoint.SetState(StateAdopting)
				m.MacAddrToKey[accessPoint.MAC()] = accessPoint
				adoptErr := adopt.Adopt(adoptCfg)
				if adoptErr != nil {
					fmt.Printf("[ADOPT] [%x] Adopt failed: %s\n", discoveryPkt.MAC, adoptErr)
					continue
				}
				accessPoint.SetState(StateAdopted)
				fmt.Printf("[DISCOVERY] [%x] Adoption for %s successful.\n", discoveryPkt.MAC, discoveryPkt.IPInfo)
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

	informPayload, err := packet.UnpackInform(d)
	if err != nil {
		return nil, err
	}
	pretty.Print(informPayload)

	if informPayload.ConfigVersion != accessPoint.GetConfigVersion() {
		if accessPoint.GetState() == StateAdopted {
			accessPoint.SetState(StateProvisioning)
		}
		fmt.Printf("[INFORM] [%x] AP config version is %q, but we are at %q\n", accessPoint.MAC(), informPayload.ConfigVersion, accessPoint.GetConfigVersion())
		return m.handleInformSendConfig(informPayload, informPkt, accessPoint, d)
	}

	if accessPoint.GetState() == StateProvisioning {
		accessPoint.SetState(StateManaged)
	}
	return m.handleNormalInform(informPayload, informPkt, accessPoint, d)
}

// handles an inform packet with a noop when no action needs to be taken.
func (m *Manager) handleNormalInform(informPayload *packet.InformData, informPkt *packet.Inform, accessPoint AP, d []byte) ([]byte, error) {
	var err error

	reply := informPkt.CloneForReply()
	reply.Data, err = packet.MakeNoop(3)
	if err != nil {
		return nil, err
	}
	fmt.Printf("[INFORM] [%x] Handled nominal inform\n", accessPoint.MAC())
	return reply.Marshal(accessPoint.AuthKey())
}

// handles an inform by generating a response to set the configuration.
func (m *Manager) handleInformSendConfig(informPayload *packet.InformData, informPkt *packet.Inform, accessPoint AP, d []byte) ([]byte, error) {
	reply := informPkt.CloneForReply()
	fmt.Printf("[INFORM] [%x] Sending system configuration\n", accessPoint.MAC())
	newSysConf, err := m.networkConfig.GenerateSysConf(informPayload.ModelName, accessPoint.GetConfigVersion()) //Make modifications based on desired settings
	if err != nil {
		return nil, err
	}

	var mgmtConf string
	mgmtConf, err = m.networkConfig.GenerateMgmtConf(hex.EncodeToString(accessPoint.AuthKey()), accessPoint.GetConfigVersion(), m.localAddr, m.httpListenerAddr)
	if err != nil {
		return nil, err
	}
	accessPoint.SetState(StateManaged)
	fmt.Println(newSysConf)
	reply.Data, err = packet.MakeConfigUpdate(newSysConf, mgmtConf, accessPoint.GetConfigVersion())
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

func setAPConfigDirty(accessPoint AP) error {
	r, err := GenerateRandomBytes(8)
	if err != nil {
		return err
	}
	fmt.Printf("[MANAGER] [%x] New configuration version %q\n", accessPoint.MAC(), hex.EncodeToString(r))
	accessPoint.SetConfigVersion(hex.EncodeToString(r))
	return nil
}
