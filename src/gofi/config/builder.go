package config

import (
	"errors"
	"strconv"
	"strings"
)

// Network setups
const (
	WpaPsk       = 0
	WpaEapRadius = 1
)

// Network represents configuration for a wireless SSID.
type Network struct {
	Kind     int
	SSID     string
	Pass     string
	Is5Ghz   bool
	NoBeacon bool
	Channel  int

	RadiusIP     string
	RadiusPort   int
	RadiusSecret string
}

// band steering modes
const (
	SteerPrefer5G = 0
	SteerBalance  = 1
)

// SteerSettings represents band steering settings
type SteerSettings struct {
	Enabled bool
	Mode    int
}

// SwitchSettings specifies options for any switches attached to the network.
type SwitchSettings struct {
}

// Config stores logical configuration of the network.
type Config struct {
	Networks        []Network
	Bandsteer       SteerSettings
	Txpower         int
	MinRSSI         int
	MinRSSIInterval int

	SwitchConfig SwitchSettings
}

var baseTwoRadioDevice = `
# enable stuff
radio.status=enabled
radio.countrycode=36
aaa.status=enabled
wireless.status=enabled

# network / routing config
bridge.1.devname=br0
bridge.1.fd=1
bridge.1.port.1.devname=eth0
#bridge.1.port.2.devname=ath0
#bridge.1.port.3.devname=ath1
bridge.1.stp.status=disabled
bridge.status=enabled
route.status=enabled

# system config
ntpclient.1.server=0.ubnt.pool.ntp.org
ntpclient.1.status=enabled
ntpclient.status=enabled
dhcpc.1.devname=br0
dhcpc.1.status=enabled
dhcpc.status=enabled
dhcpd.1.status=disabled
dhcpd.status=disabled
ebtables.1.cmd=-t broute -A BROUTING -p 0x888e -i ath0 -j DROP
ebtables.status=enabled
httpd.status=disabled
mgmt.discovery.status=enabled
mgmt.flavor=ace
mgmt.is_default=true
syslog.file=/var/log/messages
syslog.level=8
syslog.remote.ip=192.168.1.1
syslog.remote.port=514
syslog.remote.status=disabled
syslog.rotate=1
syslog.size=200
syslog.status=enabled
netconf.1.autoip.status=disabled
netconf.1.devname=br0
netconf.1.ip=192.168.1.20
netconf.1.netmask=255.255.255.0
netconf.1.status=enabled
netconf.1.up=enabled
netconf.2.autoip.status=disabled
netconf.2.devname=eth0
netconf.2.ip=0.0.0.0
netconf.2.promisc=enabled
netconf.2.status=enabled
netconf.2.up=enabled
netconf.3.autoip.status=disabled
netconf.3.devname=ath0
netconf.3.ip=0.0.0.0
netconf.3.promisc=enabled
netconf.3.status=enabled
netconf.3.up=disabled
netconf.4.autoip.status=disabled
netconf.4.devname=ath1
netconf.4.ip=0.0.0.0
netconf.4.promisc=enabled
netconf.4.status=enabled
netconf.4.up=disabled
netconf.status=enabled

# bandsteering / air time fairness
bandsteering.status=disabled
bandsteering.mode=prefer_5g
# atf.status=enabled
# atf.mode=disabled

# Radio 1 defaults - 2.4Ghz
radio.1.ack.auto=disabled
radio.1.acktimeout=64
radio.1.ampdu.status=enabled
radio.1.channel=auto
radio.1.cwm.enable=0
radio.1.cwm.mode=0
radio.1.devname=ath0
radio.1.forbiasauto=0
radio.1.ieee_mode=11nght20
radio.1.mode=master
radio.1.phyname=wifi0
radio.1.rate.auto=enabled
radio.1.rate.mcs=auto
radio.1.status=enabled
radio.1.txpower=auto
radio.1.hard_noisefloor.status=disabled
radio.1.ubntroam.status=disabled
radio.1.bgscan.status=disabled

# Radio 2 defaults - 5.0Ghz
radio.2.ack.auto=disabled
radio.2.acktimeout=64
radio.2.ampdu.status=enabled
radio.2.channel=auto
radio.2.clksel=1
radio.2.cwm.enable=0
radio.2.cwm.mode=1
radio.2.devname=ath1
radio.2.forbiasauto=0
radio.2.ieee_mode=11naht40
radio.2.mode=master
radio.2.phyname=wifi1
radio.2.rate.auto=enabled
radio.2.rate.mcs=auto
radio.2.status=enabled
radio.2.txpower=auto
radio.2.hard_noisefloor.status=disabled
radio.2.ubntroam.status=disabled
radio.2.bgscan.status=disabled
# radio.2.virtual.1.devname=ath2
# radio.2.virtual.1.status=enabled

`

var perNetworkBase = `
aaa.XREPX.br.devname=br0
aaa.XREPX.devname=ath0
aaa.XREPX.driver=madwifi
aaa.XREPX.ssid=kek
aaa.XREPX.status=enabled
aaa.XREPX.verbose=2
aaa.XREPX.wpa=2
aaa.XREPX.eapol_version=2
aaa.XREPX.wpa.group_rekey=0
aaa.XREPX.wpa.1.pairwise=CCMP
aaa.XREPX.wpa.key.1.mgmt=WPA-PSK
aaa.XREPX.wpa.psk=ee

wireless.XREPX.addmtikie=disabled
wireless.XREPX.authmode=1
wireless.XREPX.autowds=disabled
wireless.XREPX.devname=ath0
wireless.XREPX.hide_ssid=false
wireless.XREPX.is_guest=false
wireless.XREPX.l2_isolation=disabled
wireless.XREPX.mac_acl.policy=deny
wireless.XREPX.mac_acl.status=enabled
wireless.XREPX.mode=master
wireless.XREPX.parent=wifi0
wireless.XREPX.schedule_enabled=disabled
wireless.XREPX.security=none
wireless.XREPX.ssid=kek
wireless.XREPX.status=enabled
wireless.XREPX.uapsd=disabled
wireless.XREPX.usage=user
wireless.XREPX.vport=disabled
wireless.XREPX.vwire=disabled
wireless.XREPX.wds=disabled
wireless.XREPX.wmm=enabled
wireless.XREPX.puren=0
wireless.XREPX.pureg=1
`

// GenerateSysConf ingests the devices current config and modifies it based on the fields in Config.
func (b *Config) GenerateSysConf(modelName, configVersion string) (string, error) {
	var conf *Section
	var err error

	if len(b.Networks) == 0 {
		return "", errors.New("At least one network must be specified")
	}

	if len(b.Networks) > 2 {
		return "", errors.New("we do not currently support more than 2 networks")
		// To do that, we have to implement all of this nonsense
		// # radio.2.virtual.1.devname=ath2
		// # radio.2.virtual.1.status=enabled
	}

	switch modelName {
	case "USW-8P-60":
		conf, err = Parse([]byte(basicSwitchConfig))
	case "UAP-AC":
		fallthrough
	case "UAP-AC-LR":
		conf, err = Parse([]byte(baseTwoRadioDevice))
	default:
		return "", errors.New("Cannot handle model " + modelName)
	}

	if err != nil {
		return "", err
	}
	switch modelName {
	case "USW-8P-60":
		if err = b.applySwitchConf(conf, configVersion); err != nil {
			return "", err
		}
	default:
		if err = b.applySysConf(conf, configVersion); err != nil {
			return "", err
		}
	}

	var newSysConf string
	newSysConf, err = conf.Serialize()
	if err != nil {
		return "", err
	}

	return newSysConf, err
}

func (b *Config) GenerateMgmtConf(auth, configVersion, localAddr, listenerAddr string) (string, error) {
	configMgmt, err := Parse([]byte(`
		mgmt.is_default=false
		mgmt.authkey=41d6529fd555fbb1bdeeafeb995510fa
		mgmt.cfgversion=f1bb359840b519a4
		mgmt.servers.1.url=http://172.16.0.38:6080/inform
		mgmt.selfrun_guest=pass
		selfrun_guest=pass
		led_enabled=true
		cfgversion=f1bb359840b519a4
		authkey=41d6529fd555fbb1bdeeafeb995510fa
		`))
	if err != nil {
		return "", err
	}
	configMgmt.Get("mgmt").Get("servers").Get("1").Get("url").SetVal("http://" + localAddr + listenerAddr + "/inform")
	configMgmt.Get("mgmt").Get("authkey").SetVal(auth)
	configMgmt.Get("authkey").SetVal(auth)
	configMgmt.Get("mgmt").Get("cfgversion").SetVal(configVersion)
	configMgmt.Get("cfgversion").SetVal(configVersion)
	return configMgmt.Serialize()
}

func (b *Config) applySysConf(config *Section, configVersion string) error {

	for i, net := range b.Networks {
		index := strconv.Itoa(i + 1)
		base := strings.Replace(perNetworkBase, "XREPX", index, -1)
		netSpecific, err := Parse([]byte(base))
		if err != nil {
			return err
		}

		netSpecific.Get("aaa").Get(index).Get("devname").SetVal("ath" + strconv.Itoa(i))
		netSpecific.Get("wireless").Get(index).Get("devname").SetVal("ath" + strconv.Itoa(i))
		// bridge.1.port.1.devname=eth0
		netSpecific.Get("bridge").Get("1").Get("port").Get(strconv.Itoa(i + 2)).Get("devname").SetVal("ath" + strconv.Itoa(i))

		netSpecific.Get("aaa").Get(index).Get("ssid").SetVal(net.SSID)
		netSpecific.Get("wireless").Get(index).Get("ssid").SetVal(net.SSID)
		netSpecific.Get("aaa").Get(index).Get("wpa").Get("psk").SetVal(net.Pass)

		if net.NoBeacon {
			netSpecific.Get("wireless").Get(index).Get("hide_ssid").SetVal("true")
		} else {
			netSpecific.Get("wireless").Get(index).Get("hide_ssid").SetVal("false")
		}

		if net.Is5Ghz {
			netSpecific.Get("wireless").Get(index).Get("parent").SetVal("wifi1")
		} else {
			netSpecific.Get("wireless").Get(index).Get("parent").SetVal("wifi0")
		}
		if net.Channel != 0 {
			netSpecific.Get("wireless").Get(index).Get("channel").SetVal(strconv.Itoa(net.Channel))
		}

		switch net.Kind {
		case WpaEapRadius:
			netSpecific.Get("aaa").Get(index).Get("radius").Get("acct").Get("1").Get("ip").SetVal(net.RadiusIP)
			netSpecific.Get("aaa").Get(index).Get("radius").Get("acct").Get("1").Get("secret").SetVal(net.RadiusSecret)
			netSpecific.Get("aaa").Get(index).Get("radius").Get("acct").Get("1").Get("port").SetVal(strconv.Itoa(net.RadiusPort))
			netSpecific.Get("aaa").Get(index).Get("radius").Get("auth").Get("1").Get("ip").SetVal(net.RadiusIP)
			netSpecific.Get("aaa").Get(index).Get("radius").Get("auth").Get("1").Get("secret").SetVal(net.RadiusSecret)
			netSpecific.Get("aaa").Get(index).Get("radius").Get("auth").Get("1").Get("port").SetVal(strconv.Itoa(net.RadiusPort))

			netSpecific.Get("aaa").Get(index).Get("wpa").Get("key").Get("1").Get("mgmt").SetVal("WPA-EAP")
		}

		config.Consume(netSpecific)
	}

	if b.Bandsteer.Enabled {
		//bandsteering.status=disabled
		config.Get("bandsteering").Get("status").SetVal("enabled")
		switch b.Bandsteer.Mode {
		case SteerPrefer5G:
			config.Get("bandsteering").Get("mode").SetVal("prefer_5g")
		case SteerBalance:
			config.Get("bandsteering").Get("mode").SetVal("equal")
		}
	}

	if b.Txpower != 0 {
		config.Get("radio").Get("1").Get("txpower").SetVal(strconv.Itoa(b.Txpower))
		config.Get("radio").Get("2").Get("txpower").SetVal(strconv.Itoa(b.Txpower))
		config.Get("radio").Get("1").Get("txpower_mode").SetVal("custom")
		config.Get("radio").Get("2").Get("txpower_mode").SetVal("custom")
	} else {
		config.Get("radio").Get("1").Get("txpower_mode").SetVal("auto")
		config.Get("radio").Get("2").Get("txpower_mode").SetVal("auto")
	}

	if b.MinRSSI != 0 {
		config.Get("stamgr").Get("1").Get("minrssi").Get("status").SetVal("true")
		config.Get("stamgr").Get("1").Get("minrssi").Get("rssi").SetVal(strconv.Itoa(b.MinRSSI))
		config.Get("stamgr").Get("1").Get("radio").SetVal("ng")
		config.Get("stamgr").Get("1").Get("status").SetVal("true")
		config.Get("stamgr").Get("1").Get("loadbalance").Get("status").SetVal("false")
		config.Get("stamgr").Get("2").Get("minrssi").Get("status").SetVal("true")
		config.Get("stamgr").Get("2").Get("minrssi").Get("rssi").SetVal(strconv.Itoa(b.MinRSSI))
		config.Get("stamgr").Get("2").Get("radio").SetVal("na")
		config.Get("stamgr").Get("2").Get("status").SetVal("true")
		config.Get("stamgr").Get("2").Get("loadbalance").Get("status").SetVal("false")

		config.Get("stamgr").Get("status").SetVal("enabled")
		config.Get("stamgr").Get("interval").SetVal("2")
		if b.MinRSSIInterval != 0 {
			config.Get("stamgr").Get("interval").SetVal(strconv.Itoa(b.MinRSSIInterval))
		}
		config.Get("ubntroam").Get("status").SetVal("disabled")    //ubntroam.status=disabled
		config.Get("connectivity").Get("status").SetVal("enabled") //connectivity.status=enabled
	}

	return nil
}
