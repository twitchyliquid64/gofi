package config

import (
	"testing"

	"github.com/kylelemons/godebug/diff"
)

var expected = `aaa.1.br.devname=br0
aaa.1.devname=ath0
aaa.1.driver=madwifi
aaa.1.eapol_version=2
aaa.1.ssid=kek
aaa.1.status=enabled
aaa.1.verbose=2
aaa.1.wpa.1.pairwise=CCMP
aaa.1.wpa.group_rekey=0
aaa.1.wpa.key.1.mgmt=WPA-PSK
aaa.1.wpa.psk=the_shrekkening
aaa.1.wpa=2
aaa.status=enabled
bandsteering.mode=equal
bandsteering.status=enabled
bridge.1.devname=br0
bridge.1.fd=1
bridge.1.port.1.devname=eth0
bridge.1.port.2.devname=ath0
bridge.1.stp.status=disabled
bridge.status=enabled
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
ntpclient.1.server=0.ubnt.pool.ntp.org
ntpclient.1.status=enabled
ntpclient.status=enabled
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
radio.1.txpower_mode=auto
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
radio.2.txpower_mode=auto
radio.countrycode=36
radio.status=enabled
route.status=enabled
syslog.file=/var/log/messages
syslog.level=8
syslog.remote.ip=192.168.1.1
syslog.remote.port=514
syslog.remote.status=disabled
syslog.rotate=1
syslog.size=200
syslog.status=enabled
wireless.1.addmtikie=disabled
wireless.1.authmode=1
wireless.1.autowds=disabled
wireless.1.devname=ath0
wireless.1.hide_ssid=false
wireless.1.is_guest=false
wireless.1.l2_isolation=disabled
wireless.1.mac_acl.policy=deny
wireless.1.mac_acl.status=enabled
wireless.1.mode=master
wireless.1.parent=wifi0
wireless.1.schedule_enabled=disabled
wireless.1.security=none
wireless.1.ssid=kek
wireless.1.status=enabled
wireless.1.uapsd=disabled
wireless.1.usage=user
wireless.1.vport=disabled
wireless.1.vwire=disabled
wireless.1.wds=disabled
wireless.1.wmm=enabled
wireless.status=enabled`

func TestBuildACLR(t *testing.T) {
	c := Config{
		Networks: []Network{
			Network{
				SSID: "kek",
				Pass: "the_shrekkening",
			},
		},
		Bandsteer: SteerSettings{
			Enabled: true,
			Mode:    SteerBalance,
		},
	}
	out, err := c.GenerateSysConf("UAP-AC-LR", "123") //Make modifications based on desired settings
	if err != nil {
		t.Fatal(err)
	}
	if out != expected {
		t.Log(diff.Diff(expected, out))
		t.Error("Output mismatch")
	}
}

var expectedRadius = `aaa.1.br.devname=br0
aaa.1.devname=ath0
aaa.1.driver=madwifi
aaa.1.eapol_version=2
aaa.1.radius.acct.1.ip=192.168.1.3
aaa.1.radius.acct.1.port=1813
aaa.1.radius.acct.1.secret=secret
aaa.1.radius.auth.1.ip=192.168.1.3
aaa.1.radius.auth.1.port=1813
aaa.1.radius.auth.1.secret=secret
aaa.1.ssid=kek
aaa.1.status=enabled
aaa.1.verbose=2
aaa.1.wpa.1.pairwise=CCMP
aaa.1.wpa.group_rekey=0
aaa.1.wpa.key.1.mgmt=WPA-EAP
aaa.1.wpa.psk=the_shrekkening
aaa.1.wpa=2
aaa.status=enabled
bandsteering.mode=equal
bandsteering.status=enabled
bridge.1.devname=br0
bridge.1.fd=1
bridge.1.port.1.devname=eth0
bridge.1.port.2.devname=ath0
bridge.1.stp.status=disabled
bridge.status=enabled
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
ntpclient.1.server=0.ubnt.pool.ntp.org
ntpclient.1.status=enabled
ntpclient.status=enabled
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
radio.1.txpower_mode=auto
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
radio.2.txpower_mode=auto
radio.countrycode=36
radio.status=enabled
route.status=enabled
syslog.file=/var/log/messages
syslog.level=8
syslog.remote.ip=192.168.1.1
syslog.remote.port=514
syslog.remote.status=disabled
syslog.rotate=1
syslog.size=200
syslog.status=enabled
wireless.1.addmtikie=disabled
wireless.1.authmode=1
wireless.1.autowds=disabled
wireless.1.devname=ath0
wireless.1.hide_ssid=false
wireless.1.is_guest=false
wireless.1.l2_isolation=disabled
wireless.1.mac_acl.policy=deny
wireless.1.mac_acl.status=enabled
wireless.1.mode=master
wireless.1.parent=wifi0
wireless.1.schedule_enabled=disabled
wireless.1.security=none
wireless.1.ssid=kek
wireless.1.status=enabled
wireless.1.uapsd=disabled
wireless.1.usage=user
wireless.1.vport=disabled
wireless.1.vwire=disabled
wireless.1.wds=disabled
wireless.1.wmm=enabled
wireless.status=enabled`

func TestBuildACLRRadius(t *testing.T) {
	c := Config{
		Networks: []Network{
			Network{
				SSID:         "kek",
				Pass:         "the_shrekkening",
				Kind:         WpaEapRadius,
				RadiusIP:     "192.168.1.3",
				RadiusPort:   1813,
				RadiusSecret: "secret",
			},
		},
		Bandsteer: SteerSettings{
			Enabled: true,
			Mode:    SteerBalance,
		},
	}
	out, err := c.GenerateSysConf("UAP-AC-LR", "123") //Make modifications based on desired settings
	if err != nil {
		t.Fatal(err)
	}
	if out != expectedRadius {
		t.Log(diff.Diff(expectedRadius, out))
		t.Error("Output mismatch")
	}
}
