package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"gofi/adopt"
	"gofi/config"
	"gofi/manager"
	"gofi/packet"
	"strings"
)

// ap is a DAO proxying configuration storage and requests through the global config/statefile.
// ap implements the manager.AP interface.
type ap struct {
	HexAddr string
	MAddr   [6]byte

	IP string
}

func (a *ap) MAC() [6]byte {
	return a.MAddr
}

func (a *ap) GetState() int {
	return localState.AccessPoints[a.HexAddr].State
}

func (a *ap) SetState(s int) {
	ac := localState.AccessPoints[a.HexAddr]
	ac.State = s
	localState.AccessPoints[a.HexAddr] = ac
	flushConfig()
}

func (a *ap) AuthKey() []byte {
	return localState.AccessPoints[a.HexAddr].AuthKey
}

func (a *ap) SSHPw() string {
	return localState.AccessPoints[a.HexAddr].SSHPw
}

func (a *ap) GetIP() string {
	return a.IP
}

func (a *ap) GetConfigVersion() string {
	return localState.AccessPoints[a.HexAddr].ConfigVersion
}

func (a *ap) SetConfigVersion(c string) {
	ac := localState.AccessPoints[a.HexAddr]
	ac.ConfigVersion = c
	localState.AccessPoints[a.HexAddr] = ac
	flushConfig()
}

func (a *ap) GetConfig() *config.Config {
	c := &config.Config{
		Networks: []config.Network{
			config.Network{
				SSID: *ssid,
				Pass: *password,
			},
		},
		Bandsteer: config.SteerSettings{
			Enabled: *bandSteer,
			Mode:    config.SteerPrefer5G,
		},
		Txpower: *txPower,
		MinRSSI: *minRSSI,
	}

	if *do5G {
		c.Networks = append(c.Networks, config.Network{
			SSID:   *ssid,
			Pass:   *password,
			Is5Ghz: true,
		})
	}
	return c
}

func onControllerDoesntKnowAP(ip string, i *packet.Inform) (manager.AP, error) {
	haddr := hex.EncodeToString(i.APMAC[:])
	_, known := localState.AccessPoints[haddr]
	if !known {
		return nil, errors.New("Ap " + haddr + " not known")
	}
	return &ap{
		HexAddr: haddr,
		MAddr:   i.APMAC,
		IP:      strings.Split(ip, ":")[0],
	}, nil
}

func onDiscoveryPacket(localAddr, listenerAddr string, discoveryPkt *packet.Discovery) (manager.AP, *adopt.Config, error) {
	var adoptCfg *adopt.Config
	haddr := hex.EncodeToString(discoveryPkt.MAC[:])

	if _, isKnown := localState.AccessPoints[haddr]; isKnown {
		fmt.Printf("Should not need to adopt %x - already known\n", discoveryPkt.MAC)
	} else {
		adoptCfg = adopt.NewConfig(strings.Split(discoveryPkt.IPInfo.String(), ":")[0]+":22", localAddr+listenerAddr, "ubnt")
		localState.AccessPoints[haddr] = apState{
			Mac:     discoveryPkt.MAC,
			AuthKey: adoptCfg.Key,
			SSHPw:   "ubnt",
		}
		flushConfig()
	}

	return &ap{
		HexAddr: haddr,
		MAddr:   discoveryPkt.MAC,
		IP:      strings.Split(discoveryPkt.IPInfo.String(), ":")[0],
	}, adoptCfg, nil
}
