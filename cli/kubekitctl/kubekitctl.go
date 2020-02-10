package kubekitctl

import (
	"os"
	"path/filepath"

	"github.com/johandry/log"
	"github.com/liferaft/kubekit/cli"
	clientv1 "github.com/liferaft/kubekit/pkg/client/v1"
	homedir "github.com/mitchellh/go-homedir"
)

const apiVersion = "v1"

const (
	defDebug          = false
	defInsecure       = false
	defPort           = "5823"
	defHost           = "localhost"
	defNoGRPC         = false
	defNoHTTP         = false
	defCertDir        = "server/pki"
	defKubeKitHomeDir = ".kubekit.d"
)

// Config store all the KubeKit Client parameters
type Config struct {
	Debug       bool   `json:"debug" yaml:"debug" toml:"debug" mapstructure:"debug"`
	Insecure    bool   `json:"insecure" yaml:"insecure" toml:"insecure" mapstructure:"insecure"`
	Host        string `json:"host" yaml:"host" toml:"host" mapstructure:"host"`
	Port        string `json:"port" yaml:"port" toml:"port" mapstructure:"port"`
	PortGRPC    string `json:"grpc-port" yaml:"grpc-port" toml:"grpc-port" mapstructure:"grpc-port"`
	PortHealthz string `json:"healthz-port" yaml:"healthz-port" toml:"healthz-port" mapstructure:"healthz-port"`
	NoHTTP      bool   `json:"no-http" yaml:"no-http" toml:"no-http" mapstructure:"no-http"`
	NoGRPC      bool   `json:"no-grpc" yaml:"no-grpc" toml:"no-grpc" mapstructure:"no-grpc"`
	CertDir     string `json:"cert-dir" yaml:"cert-dir" toml:"cert-dir" mapstructure:"cert-dir"`
	CertFile    string `json:"tls-cert-file" yaml:"tls-cert-file" toml:"tls-cert-file" mapstructure:"tls-cert-file"`
	KeyFile     string `json:"tls-private-key-file" yaml:"tls-private-key-file" toml:"tls-private-key-file" mapstructure:"tls-private-key-file"`
	CAFile      string `json:"ca-file" yaml:"ca-file" toml:"ca-file" mapstructure:"ca-file"`

	Logger *log.Logger
	client *clientv1.Config
}

var config *Config

func (c *Config) init() error {
	if c.NoGRPC && c.NoHTTP {
		return cli.UserErrorf("no-http and no-grpc cannot be used at same time")
	}

	if c.PortGRPC == "" {
		c.PortGRPC = c.Port
	}
	if c.PortHealthz == "" {
		c.PortHealthz = c.Port
	}

	// If no certDir is given, ...
	if len(c.CertDir) == 0 {
		var kkHomeDir string
		if kkHome := os.Getenv("KUBEKIT_HOME"); kkHome != "" {
			// ... use $KUBEKIT_HOME/server/pki ...
			kkHomeDir = kkHome
		} else {
			// ... but if KUBEKIT_HOME is not set, use $HOME/.kubekit.d/server/pki
			home, err := homedir.Dir()
			if err != nil {
				panic(err)
			}
			kkHomeDir = filepath.Join(home, defKubeKitHomeDir)
		}

		c.CertDir = filepath.Join(kkHomeDir, defCertDir)
	}

	if len(c.CertFile) == 0 {
		c.CertFile = filepath.Join(c.CertDir, "kubekit-client.crt")
	}
	if len(c.KeyFile) == 0 {
		c.KeyFile = filepath.Join(c.CertDir, "kubekit-client.key")
	}
	if len(c.CAFile) == 0 {
		c.CAFile = filepath.Join(c.CertDir, "kubekit-ca.crt")
	}

	c.client = clientv1.New(apiVersion, c.Logger)

	if !c.Insecure {
		if err := c.client.WithTLS(c.CertDir, c.CertFile, c.KeyFile, c.CAFile); err != nil {
			return err
		}
	}

	if !c.NoGRPC {
		if err := c.client.WithGRPC(c.Host, c.PortGRPC); err != nil {
			return err
		}
	}
	if !c.NoHTTP {
		c.client.WithHTTP(c.Host, c.Port, c.PortHealthz)
	}
	return nil
}
