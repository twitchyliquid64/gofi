package config

// Config stores logical configuration of the network.
type Config struct {
	SSID string
	Pass string
}

// GenerateSysConf ingests the devices current config and modifies it based on the fields in Config.
func (b *Config) GenerateSysConf(sysconf []byte, configVersion string) (string, error) {
	configSys, err := Parse(sysconf)
	if err != nil {
		return "", err
	}
	if err = b.applySysConf(configSys, configVersion); err != nil {
		return "", err
	}
	var newSysConf string
	newSysConf, err = configSys.Serialize()
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
	config.Get("aaa").Get("1").Get("ssid").SetVal(b.SSID)
	config.Get("wireless").Get("1").Get("ssid").SetVal(b.SSID)
	config.Get("aaa").Get("1").Get("wpa").Get("psk").SetVal(b.Pass)

	for _, section := range config.Get("wireless").Iterate() {
		section.Get("hide_ssid").SetVal("false")
	}
	for _, section := range config.Get("radio").Iterate() {
		section.Get("channel").SetVal("auto")
	}
	return nil
}
