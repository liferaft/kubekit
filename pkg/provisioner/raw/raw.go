package raw

import (
	"github.com/kraken/ui"
)

// Platform implements the Provisioner interface for vSphere
type Platform struct {
	name    string
	config  *Config
	ui      *ui.UI
	version string
}

// New creates a new Plaform with the given environment configuration
func New(clusterName string, envConfig map[string]string, ui *ui.UI, version string) (*Platform, error) {
	config := &Config{}

	if err := config.MergeWithEnv(envConfig, defaultConfig); err != nil {
		return nil, err
	}
	config.clusterName = clusterName

	return &Platform{
		name:    "raw",
		config:  config,
		ui:      ui,
		version: version,
	}, nil
}

// CreateFrom creates a new Plaftorm with the given configuration for vSphere
func CreateFrom(clusterName string, config map[interface{}]interface{}, credentials []string, ui *ui.UI, version string) *Platform {
	if config == nil {
		return newPlatform(&defaultConfig, ui, version)
	}
	c := NewConfigFrom(config)
	c.clusterName = clusterName

	return newPlatform(c, ui, version)
}

func newPlatform(c *Config, ui *ui.UI, version string) *Platform {
	p := &Platform{
		name:    "raw",
		config:  c,
		ui:      ui,
		version: version,
	}

	return p
}

// MergeWithEnv implements the MergeWithEnv method from the interfase
// Provisioner. It merges the environment variables with the existing configuration
func (p *Platform) MergeWithEnv(envConfig map[string]string) error {
	return p.config.MergeWithEnv(envConfig)
}
