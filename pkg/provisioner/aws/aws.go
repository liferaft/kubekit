package aws

import (
	"github.com/kraken/ui"

	"github.com/kraken/terraformer"
)

// Platform implements the Provisioner interface for AWS
type Platform struct {
	name    string
	config  *Config
	t       *terraformer.Terraformer
	ui      *ui.UI
	version string
}

// New creates a new Plaform with the given environment configuration
func New(clusterName string, envConfig map[string]string, ui *ui.UI, version string) (*Platform, error) {
	config := &Config{}
	if err := config.MergeWithEnv(envConfig, defaultConfig); err != nil {
		return nil, err
	}
	config.ClusterName = clusterName

	return &Platform{
		name:    "aws",
		config:  config,
		ui:      ui,
		version: version,
	}, nil
}

// CreateFrom creates a new Plaftorm with the given configuration for AWS
func CreateFrom(clusterName string, config map[interface{}]interface{}, credentials []string, ui *ui.UI, version string) *Platform {
	if config == nil {
		return newPlatform(&defaultConfig, credentials, ui, version)
	}
	c := NewConfigFrom(config)
	c.ClusterName = clusterName

	return newPlatform(c, credentials, ui, version)
}

func newPlatform(c *Config, credentials []string, ui *ui.UI, version string) *Platform {
	p := Platform{
		name:    "aws",
		config:  c,
		ui:      ui,
		version: version,
	}
	p.Credentials(credentials...)
	return &p
}

// MergeWithEnv implements the MergeWithEnv method from the interfase
// Provisioner. It merges the environment variables with the existing configuration
func (p *Platform) MergeWithEnv(envConfig map[string]string) error {
	return p.config.MergeWithEnv(envConfig)
}
