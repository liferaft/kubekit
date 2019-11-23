package aks

import (
	"github.com/johandry/log"

	"github.com/kraken/terraformer"
	"github.com/kraken/ui"
)

// Platform implements the Provisioner interface for Azure AKS
type Platform struct {
	name    string
	config  *Config
	t       *terraformer.Terraformer
	logger  *log.Logger
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
		name:    "aks",
		config:  config,
		ui:      ui,
		version: version,
	}, nil
}

// CreateFrom creates a new Platform with the given configuration for Azure AKS
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
		name:    "aks",
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
