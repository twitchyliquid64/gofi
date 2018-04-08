package config

var basicSwitchConfig = `
# vlan
vlan.status=disabled
# bridge
bridge.status=disabled

# dhcpd / dhcpc
dhcpc.1.devname=eth0
dhcpc.1.status=enabled
dhcpc.status=enabled
dhcpd.1.status=disabled
dhcpd.status=disabled

ebtables.status=disabled
httpd.status=disabled

netconf.1.autoip.status=disabled
netconf.1.devname=eth0
netconf.1.ip=0.0.0.0
netconf.1.status=enabled
netconf.1.up=enabled
netconf.status=enabled

route.status=enabled

# ntpclient
ntpclient.status=enabled
ntpclient.1.status=enabled
ntpclient.1.server=0.ubnt.pool.ntp.org
ntpclient.2.status=enabled
ntpclient.2.server=1.ubnt.pool.ntp.org
ntpclient.3.status=enabled
ntpclient.3.server=2.ubnt.pool.ntp.org
ntpclient.4.status=enabled
ntpclient.4.server=3.ubnt.pool.ntp.org

radio.status=disabled
stamgr.status=disabled
switch.status=enabled

syslog.file=/var/log/messages
syslog.level=8
syslog.remote.status=disabled
syslog.rotate=1
syslog.size=200
syslog.status=enabled

# switch
switch.managementvlan=1
switch.wevent.idp=enabled
switch.wevent.mcip=
switch.wevent.key=
switch.jumboframes=disabled
switch.mtu=9216
switch.stp.version=rstp
switch.stp.priority=32768
switch.stp.status=enabled
switch.dot1x.status=disabled
switch.vlan.1.id=1
switch.vlan.1.mode=untagged
switch.vlan.1.status=enabled
switch.dhcp_snoop.status=enabled
switch.port.1.name=Port 1
switch.port.1.lldpmed.opmode=enabled
switch.port.1.lldpmed.topology_notify=disabled
switch.port.1.opmode=switch
switch.port.2.name=Port 2
switch.port.2.lldpmed.opmode=enabled
switch.port.2.lldpmed.topology_notify=disabled
switch.port.2.opmode=switch
switch.port.3.name=Port 3
switch.port.3.lldpmed.opmode=enabled
switch.port.3.lldpmed.topology_notify=disabled
switch.port.3.opmode=switch
switch.port.4.name=Port 4
switch.port.4.lldpmed.opmode=enabled
switch.port.4.lldpmed.topology_notify=disabled
switch.port.4.opmode=switch
switch.port.5.name=Port 5
switch.port.5.lldpmed.opmode=enabled
switch.port.5.lldpmed.topology_notify=disabled
switch.port.5.opmode=switch
switch.port.5.poe=auto
switch.port.6.name=Port 6
switch.port.6.lldpmed.opmode=enabled
switch.port.6.lldpmed.topology_notify=disabled
switch.port.6.opmode=switch
switch.port.6.poe=auto
switch.port.7.name=Port 7
switch.port.7.lldpmed.opmode=enabled
switch.port.7.lldpmed.topology_notify=disabled
switch.port.7.opmode=switch
switch.port.7.poe=auto
switch.port.8.name=Port 8
switch.port.8.lldpmed.opmode=enabled
switch.port.8.lldpmed.topology_notify=disabled
switch.port.8.opmode=switch
switch.port.8.poe=auto

users.1.name=ubnt
users.1.password=VvpvCwhccFv6Q
users.1.status=enabled
users.status=enabled
`

func (b *Config) applySwitchConf(config *Section, configVersion string) error {
	return nil
}
