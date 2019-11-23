package aks

import (
	"testing"

	"github.com/johandry/log"
	"github.com/stretchr/testify/assert"
	"github.com/kraken/ui"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
)

var (
	tUI     = ui.New(false, log.NewDefault())
	version = "1.0"
)

func TestNew(t *testing.T) {
	type args struct {
		envConfig map[string]string
	}
	tests := []struct {
		name string
		args args
		want *Platform
	}{
		{
			name: "new aks with no env vars",
			args: args{
				envConfig: nil,
			},
			want: defaultAKS([]string{"my_access_key", "my_secret_key", "my_session_token", "my_aws_region"}),
		},
		{
			name: "create aks from nil config with nil string value",
			args: args{
				envConfig: nil,
			},
			want: modifiedAKS("VnetName", nil, []string{"my_access_key", "my_secret_key", "my_session_token"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, _ := New("testCluster", tt.args.envConfig, tUI, version)
			assert.Equal(t, tt.want.Config(), got.Config(), tt.name)
		})
	}
}

func defaultAKS(credentials []string) *Platform {
	p := &Platform{
		name:    "aks",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.ClusterName = "testCluster"
	return p
}

func modifiedAKS(key string, value interface{}, credentials []string) *Platform {
	p := &Platform{
		name:    "aks",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.ClusterName = "testCluster"
	config.SetField(p.config, key, value)
	return p
}

type miniConfig struct {
	Platforms map[string]interface{} `json:"platforms" yaml:"platforms" mapstructure:"platform"`
}
