package adopt

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/ssh"
)

// Config specifies all the information to perform an adopt operation.
type Config struct {
	APAddr         string
	ControllerAddr string
	Pass           string

	Key []byte
}

// Adopt performs an adopt operation.
func Adopt(cfg *Config) error {
	c := &ssh.ClientConfig{User: "ubnt", HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	c.Auth = append(c.Auth, ssh.Password(cfg.Pass))

	client, err := ssh.Dial("tcp", cfg.APAddr, c)
	if err != nil {
		return err
	}
	defer client.Close()

	s, err := client.NewSession()
	if err != nil {
		return err
	}

	_, err = s.CombinedOutput("/usr/bin/syswrapper.sh set-adopt http://" + cfg.ControllerAddr + "/inform " + hex.EncodeToString(cfg.Key))
	if err != nil {
		return err
	}
	return nil
}

// NewConfig creates a Config with a random encryption key.
func NewConfig(apAddr, controllerAddr, pass string) *Config {
	b, err := GenerateRandomBytes(16)
	if err != nil {
		panic(err)
	}
	return &Config{
		APAddr:         apAddr,
		ControllerAddr: controllerAddr,
		Key:            b,
		Pass:           pass,
	}
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
