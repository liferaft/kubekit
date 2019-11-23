package vsphere

import (
	"os"
	"testing"

	"github.com/johandry/log"
	"github.com/stretchr/testify/assert"
	"github.com/kraken/ui"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	yaml "gopkg.in/yaml.v2"
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
		"VSPHERE_SERVER":   "my_access_key",
		"VSPHERE_USERNAME": "my_secret_key",
		"VSPHERE_PASSWORD": "my_session_token",
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
			name: "create %s from nil config",
			args: args{
				config: nil,
			},
			want: defaultVS([]string{"my_session_token", "my_access_key", "my_secret_key"}),
		},
		{
			name: "create vsphere from nil config with nil string value",
			args: args{
				config: nil,
			},
			want: modifiedVS("Datacenter", nil, []string{"my_session_token", "my_access_key", "my_secret_key"}),
		},
		{
			name: "create from config",
			args: args{
				config: newConfigFromYaml(),
			},
			want: &Platform{
				name: "vsphere",
				config: &Config{
					ClusterName: "testCluster",
					Username:    "root",
					Datacenter:  "# Required value. Example: Vagrant",
					DefaultNodePool: NodePool{
						TemplateName: "vmware-kubekit-os-01.48-19.11.00-200G",
						CPUs:         8,
						Memory:       16384,
						RootVolSize:  200,
						LinkedClone:  true,
					},
					NodePools: map[string]NodePool{
						"master": NodePool{
							Count: 1,
						},
						"worker": NodePool{
							Count: 1,
						},
					},
					KubeAPISSLPort:    6443,
					KubeVIPAPISSLPort: 8443,
					VsphereUsername:   "my_access_key",
					VspherePassword:   "my_secret_key",
					VsphereServer:     "my_session_token",
					DNSServers:        []string{"153.64.180.100", "153.64.251.200"},
					TimeServers:       []string{"0.us.pool.ntp.org", "1.us.pool.ntp.org", "2.us.pool.ntp.org"},
					Folder:            "# Required value. Example: Discovered virtual machine/ja186051",
					DNSSearch:         []string{},
					VsphereNet:        "# Required value. Example: dvpg_vm_550",
					ResourcePool:      "# Required value. Example: sd_vgrnt_01/Resources/vagrant01",
					Datastore:         "# Required value. Example: sd_labs_19_vgrnt_dsc/sd_labs_19_vgrnt03",
				},
				ui: tUI,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CreateFrom("testCluster", tt.args.config, []string{"my_session_token", "my_access_key", "my_secret_key"}, tUI, version)
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
			name: "new vsphere with no env vars",
			args: args{
				envConfig: nil,
			},
			want: defaultVS([]string{"my_access_key", "my_secret_key", "my_session_token"}),
		},
		{
			name: "new vsphere with env vars",
			args: args{
				envConfig: map[string]string{
					"Datastore": "test",
				},
			},
			want: modifiedVS("Datastore", "test", []string{"my_access_key", "my_secret_key", "my_session_token"}),
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

func defaultVS(credentials []string) *Platform {
	p := &Platform{
		name:    "vsphere",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.ClusterName = "testCluster"
	p.Credentials(credentials...)
	return p
}

func modifiedVS(key string, value interface{}, credentials []string) *Platform {
	p := &Platform{
		name:    "vsphere",
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

func newConfigFromYaml() map[interface{}]interface{} {
	yamlStr := `
platforms:
  vsphere:
    kube_api_ssl_port: 6443
    disable_master_ha: false
    kube_virtual_ip_shortname: ""
    kube_virtual_ip_api: ""
    kube_vip_api_ssl_port: 8443
    public_apiserver_dns_name: ""
    private_apiserver_dns_name: ""
    username: root
    datacenter: '# Required value. Example: Vagrant'
    datastore: '# Required value. Example: sd_labs_19_vgrnt_dsc/sd_labs_19_vgrnt03'
    resource_pool: '# Required value. Example: sd_vgrnt_01/Resources/vagrant01'
    vsphere_net: '# Required value. Example: dvpg_vm_550'
    folder: '# Required value. Example: Discovered virtual machine/ja186051'
    domain: ""
    dns_servers:
    - 153.64.180.100
    - 153.64.251.200
    dns_search: []
    time_servers:
    - 0.us.pool.ntp.org
    - 1.us.pool.ntp.org
    - 2.us.pool.ntp.org
    private_key_file: ""
    public_key_file: ""
    default_node_pool:
      template_name: vmware-kubekit-os-01.48-19.11.00-200G
      cpus: 8
      memory: 16384
      root_vol_size: 200
      linked_clone: true
    node_pools:
      master:
        count: 1
      worker:
        count: 1
`
	c := miniConfig{}
	yaml.Unmarshal([]byte(yamlStr), &c)
	p := c.Platforms["vsphere"]

	cnf := p.(map[interface{}]interface{})
	return cnf
}
