package ssh

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config store the SSH configuration
type Config struct {
	Address string
	config  *ssh.ClientConfig
	client  *ssh.Client
}

// New returns an instance of the SSH configuration
func New(username, address, privateKey, password string) (*Config, error) {
	// TODO: Should the HostKeyCallback be replace? this may not be recommended for production
	sshConfig := &ssh.ClientConfig{
		User:            username,
		Timeout:         5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	auth := []ssh.AuthMethod{}

	if len(privateKey) != 0 {
		publicKey, err := publicKey(privateKey)
		if err != nil {
			return nil, err
		}
		auth = append(auth, publicKey)
	}

	if len(password) != 0 {
		sshInteractive := func(user, instruction string, questions []string, echos []bool) (answers []string, err error) {
			answers = make([]string, len(questions))
			for n := range questions {
				answers[n] = password
			}

			return answers, nil
		}

		auth = append(auth, ssh.Password(password))
		// In case 'PasswordAuthentication' is set to No in /etc/ssh/sshd_config at the server, try with interactive method
		auth = append(auth, ssh.KeyboardInteractive(sshInteractive))
	}

	if len(auth) == 0 {
		return nil, fmt.Errorf("no authentication method defined for host %s", address)
	}

	sshConfig.Auth = auth
	sshConfig.SetDefaults()

	return &Config{
		Address: address,
		config:  sshConfig,
	}, nil
}

func publicKey(privateKey string) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

func (c *Config) setClient() error {
	if c.client != nil {
		c.client.Close()
		// return nil
	}

	client, err := ssh.Dial("tcp", c.Address+":22", c.config)
	c.client = client
	return err
}

// Close closes the client connection if exists
func (c *Config) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
