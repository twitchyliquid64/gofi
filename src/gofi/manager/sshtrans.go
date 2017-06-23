package manager

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// NOTE: Deprecated method. Don't do this.
func applyConfig(addr, pass string) error {
	c := &ssh.ClientConfig{User: "ubnt", HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	c.Auth = append(c.Auth, ssh.Password(pass))

	client, err := ssh.Dial("tcp", addr+":22", c)
	if err != nil {
		return err
	}
	defer client.Close()

	s, err := client.NewSession()
	if err != nil {
		return err
	}

	var ubntConfOutput []byte
	ubntConfOutput, err = s.Output("ubntconf -c /tmp/system.cfg")
	if err != nil {
		return err
	}
	s.Close()

	s, err = client.NewSession()
	if err != nil {
		return err
	}

	var cfgtmdOutput []byte
	cfgtmdOutput, err = s.Output("cfgmtd -w -p /etc/")
	fmt.Printf("[CONF-ENABLE] ubnt: %q, cfgtmd: %q\n", string(ubntConfOutput), string(cfgtmdOutput))
	return err
}

// NOTE: Deprecated method. Don't do this.
func setSystemConfig(addr, pass, cfg string) error {
	c := &ssh.ClientConfig{User: "ubnt", HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	c.Auth = append(c.Auth, ssh.Password(pass))

	client, err := ssh.Dial("tcp", addr+":22", c)
	if err != nil {
		return err
	}
	defer client.Close()

	s, err := client.NewSession()
	if err != nil {
		return err
	}
	defer s.Close()
	w, err := s.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer w.Close()
		fmt.Fprint(w, cfg)
	}()
	return s.Run("cat - > /tmp/system.cfg")
}

// GetSysConfig returns the currently applied system configuration.
// Probably dont use this approach.
func GetSysConfig(addr, pass string) ([]byte, error) {
	c := &ssh.ClientConfig{User: "ubnt", HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	c.Auth = append(c.Auth, ssh.Password(pass))

	client, err := ssh.Dial("tcp", addr+":22", c)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	s, err := client.NewSession()
	if err != nil {
		return nil, err
	}

	var system []byte
	system, err = s.Output("/bin/cat /tmp/system.cfg")
	if err != nil {
		return nil, err
	}
	s.Close()

	return system, nil
}
