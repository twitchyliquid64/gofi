package config

// Config stores logical configuration of the network.
type Config struct {
	SSID   string
	Pass   string
	Beacon bool
}

// Generate ingests the devices current config and modifies it based on the fields in Config.
func (b *Config) Generate(sysconf, mgmtconf []byte) (string, string, error) {
	config, err := Parse(sysconf)
	if err != nil {
		return "", "", err
	}
	if err = b.apply(config); err != nil {
		return "", "", err
	}
	var newSysConf string
	newSysConf, err = config.Serialize()
	return newSysConf, string(mgmtconf), err
}

func (b *Config) apply(config *Section) error {
	for _, section := range config.Get("wireless").Iterate() {
		section.Get("hide_ssid").SetVal("false")
	}
	for _, section := range config.Get("radio").Iterate() {
		section.Get("channel").SetVal("auto")
	}
	return nil
}
