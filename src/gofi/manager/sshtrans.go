package manager

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

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

// GetConfig returns the currently applied configuration.
func GetConfig(addr, pass string) ([]byte, []byte, error) {
	c := &ssh.ClientConfig{User: "ubnt", HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	c.Auth = append(c.Auth, ssh.Password(pass))

	client, err := ssh.Dial("tcp", addr+":22", c)
	if err != nil {
		return nil, nil, err
	}
	defer client.Close()

	s, err := client.NewSession()
	if err != nil {
		return nil, nil, err
	}

	var system []byte
	system, err = s.Output("/bin/cat /tmp/system.cfg")
	if err != nil {
		return nil, nil, err
	}
	s.Close()

	s, err = client.NewSession()
	if err != nil {
		return nil, nil, err
	}

	var mgmt []byte
	mgmt, err = s.Output("/bin/cat /etc/persistent/cfg/mgmt")
	if err != nil {
		return nil, nil, err
	}

	return system, mgmt, nil
}
