package vra

import (
	"testing"

	"github.com/johandry/log"
	"github.com/stretchr/testify/assert"
	"github.com/kraken/ui"
	"github.com/kubekit/kubekit/pkg/provisioner/config"
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
			name: "new vra with no env vars",
			args: args{
				envConfig: nil,
			},
			want: defaultVRA([]string{"my_access_key", "my_secret_key", "my_session_token", "my_aws_region"}),
		},
		{
			name: "create vra from nil config with nil string value",
			args: args{
				envConfig: nil,
			},
			want: modifiedVRA("Username", nil, []string{"my_access_key", "my_secret_key", "my_session_token"}),
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

func defaultVRA(credentials []string) *Platform {
	p := &Platform{
		name:    "vra",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.clusterName = "testCluster"
	return p
}

func modifiedVRA(key string, value interface{}, credentials []string) *Platform {
	p := &Platform{
		name:    "vra",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.clusterName = "testCluster"
	config.SetField(p.config, key, value)
	return p
}

type miniConfig struct {
	Platforms map[string]interface{} `json:"platforms" yaml:"platforms" mapstructure:"platform"`
}
