package manager

import "golang.org/x/crypto/ssh"

// GetConfig returns the currently applied configuration.
func GetConfig(addr, pass string) ([]byte, error) {
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

	var o []byte
	o, err = s.Output("/bin/cat /tmp/system.cfg")
	if err != nil {
		return nil, err
	}
	return o, nil
}
