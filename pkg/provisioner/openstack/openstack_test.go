package openstack

import (
	"os"
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

func TestCreateFrom(t *testing.T) {
	// required env vars
	// TODO: env vars should not be required to run provider, instead
	// it should be an explicit parameter or part of a config
	for k, v := range map[string]string{
		"OPENSTACK_USERNAME": "my_access_key",
		"OPENSTACK_PASSWORD": "my_secret_key",
		"OPENSTACK_AUTH_URL": "my_session_token",
	} {
		t.Logf("setting env var %s to %s", k, v)
		err := os.Setenv(k, v)
		if err != nil {
			t.Fatalf("unable to set required environment variable %s due to: %s", k, err)
		}
	}

	type args struct {
		config map[interface{}]interface{}
	}
	tests := []struct {
		name string
		args args
		want *Platform
	}{
		{
			name: "create OPENSTACK from nil config",
			args: args{
				config: nil,
			},
			want: defaultOS([]string{"my_access_key", "my_secret_key", "my_session_token"}),
		},
		{
			name: "create OPENSTACK from nil config with nil string value",
			args: args{
				config: nil,
			},
			want: modifiedOS("OpenstackRegion", nil, []string{"my_access_key", "my_secret_key", "my_session_token"}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CreateFrom("testCluster", tt.args.config, []string{"my_access_key", "my_secret_key", "my_session_token"}, tUI, version)
			assert.Equal(t, tt.want.Config(), got.Config(), tt.name)
		})
	}
}

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
			name: "new openstack with no env vars",
			args: args{
				envConfig: nil,
			},
			want: defaultOS([]string{"my_access_key", "my_secret_key", "my_session_token"}),
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

func defaultOS(credentials []string) *Platform {
	p := &Platform{
		name:    "openstack",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.ClusterName = "testCluster"
	p.Credentials(credentials...)
	return p
}

func modifiedOS(key string, value interface{}, credentials []string) *Platform {
	p := &Platform{
		name:    "openstack",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.ClusterName = "testCluster"
	config.SetField(p.config, key, value)
	p.Credentials(credentials...)
	return p
}

type miniConfig struct {
	Platforms map[string]interface{} `json:"platforms" yaml:"platforms" mapstructure:"platform"`
}
