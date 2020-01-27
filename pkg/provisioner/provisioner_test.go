package provisioner

import (
	"io/ioutil"
	"testing"

	"github.com/johandry/log"
	"github.com/kraken/ui"
	"github.com/liferaft/kubekit/pkg/provisioner/ec2"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

const (
	defaultAWSConfigPath = "testdata/aws_default.config.yml"
	noAwsConfigPath      = "testdata/no_aws.config.yml"
)

var (
	tUI     = ui.New(false, log.NewDefault())
	version = "1.0"
)

func TestNewPlatform(t *testing.T) {
	c := map[string]interface{}{}
	yaml.Unmarshal([]byte(tc2YamlPlatformConfig), &c)

	type args struct {
		name   string
		config interface{}
	}
	tests := []struct {
		name    string
		args    args
		creds   []string // credentials to provision
		want    string
		wantErr bool
	}{
		{
			name: "unknown platform",
			args: args{
				name:   "building",
				config: nil,
			},
			want:    "null\n",
			wantErr: true,
		},
		{
			name: "EC2 platform with no config",
			args: args{
				name:   "ec2",
				config: nil,
			},
			creds:   []string{"my_access_key", "my_secret_key", "my_session_token", "aws_region"},
			want:    string(mustLoad(t, noAwsConfigPath)),
			wantErr: false,
		},
		{
			name: "EC2 platform with config",
			args: args{
				name:   "ec2",
				config: c["ec2"],
			},
			creds:   []string{"my_access_key", "my_secret_key", "my_session_token", "us-west-2"},
			want:    string(mustLoad(t, defaultAWSConfigPath)),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewPlatform(tt.args.name, "testCluster", tt.args.config, tt.creds, tUI, version)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPlatform() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				assert.Equal(t, nil, got)
				return
			}
			assert.NotEqual(t, nil, got)
			conf := got.Config()
			result, err := yaml.Marshal(conf)
			assert.IsType(t, nil, err)
			assert.Equal(t, tt.want, string(result))
		})
	}
}

func TestSupportedPlatforms(t *testing.T) {
	type args struct {
		envConfig map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
		skip bool // should this test be skipped until fixed?
	}{
		{
			name: "Platforms with no env config",
			args: args{
				envConfig: nil,
			},
			want: map[string]interface{}{
				"ec2": func() *ec2.Config {
					cfg := ec2.Config{}
					mustParseYaml(t, mustLoad(t, defaultAWSConfigPath), &cfg)
					return &cfg
				}(),
			},
			skip: true, // TODO(po250005): fix or remove this flakey test
		},
	}
	for _, tt := range tests {
		if tt.skip {
			continue // skip this flakey test
		}
		t.Run(tt.name, func(t *testing.T) {
			got := SupportedPlatforms("testCluster", tt.args.envConfig, tUI, version)
			for pName, platform := range got {
				if confYaml, ok := tt.want[pName]; ok {
					assert.Equal(t, platform.Config(), &confYaml)
				} else {
					t.Logf("SupportedPlatforms() not found platform %q", pName)
					// TODO: add all platform configs
					return
				}
			}

		})
	}
}

func TestSupportedPlatformsName(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "all platform",
			want: []string{
				"aks",
				"ec2",
				"eks",
				"vsphere",
				"openstack",
				"raw",
				"vra",
				"stacki",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SupportedPlatformsName()
			assert.Equal(t, tt.want, got)
		})
	}
}

func mustLoad(tb testing.TB, srcP string) []byte {
	out, err := ioutil.ReadFile(srcP)
	if err != nil {
		tb.Fatalf("unable to open %s : %s", srcP, err)
	}
	return out
}

func mustParseYaml(tb testing.TB, raw []byte, dst interface{}) {
	err := yaml.Unmarshal(raw, dst)
	if err != nil {
		tb.Logf("can't marshal this input: %s", string(raw))
		tb.Fatalf("unable to unmarshal due to: %s", err)
	}
}

var tc2YamlPlatformConfig = string(`ec2:
  aws_env: aws-k8s
  kube_api_ssl_port: 8081
  disable_master_ha: true
  kube_virtual_ip_shortname: ""
  kube_virtual_ip_api: ""
  kube_vip_api_ssl_port: 8443
  public_apiserver_dns_name: ""
  private_apiserver_dns_name: ""
  username: ec2-user
  aws_region: us-west-2
  aws_vpc_id: vpc-8d56b9e9
  private_key_file: ""
  public_key_file: ""
  configure_from_private_net: false
  dns_servers:
  - 1.1.1.1
  - 8.8.8.8
  dns_search:
  - some.value.com
  - other.value.com
  time_servers:
  - 169.254.169.123
  - 1.2.3.4
  default_node_pool:
    count: 0
    connection_timeout: 5m
    aws_ami: ami-0b8485a3553c5d032
    aws_instance_type: m4.2xlarge
    root_volume_size: 200
    root_volume_type: gp2
    security_groups:
    - sg-502d9a37
    subnets:
    - subnet-5bddc82c
    kubelet_node_labels:
    - node-role.kubernetes.io/compute=""
    - node.kubernetes.io/compute=""
  node_pools:
    master:
      count: 1
      kubelet_node_labels:
      - node-role.kubernetes.io/master=""
      - node.kubernetes.io/master=""
      kubelet_node_taints:
      - node-role.kubernetes.io/master="":NoSchedule
      - node.kubernetes.io/master="":NoSchedule
    worker:
      count: 1
      kubelet_node_labels:
      - node-role.kubernetes.io/worker=""
      - node.kubernetes.io/worker=""
`)

var tc2YamlPlatform = string(`ec2:
  aws_env: aws-k8s
  username: ec2-user
  aws_vpc_id: vpc-8d56b9e9
  private_key_file: ""
  public_key_file: ""
  default_node_pool:
    count: 0
    connection_timeout: 5m
    aws_ami: ami-0b8485a3553c5d032
    aws_instance_type: m4.2xlarge
    root_volume_size: 200
    root_volume_type: gp2
    security_groups:
    - sg-502d9a37
    subnets:
    - subnet-5bddc82c
    kubelet_node_labels:
    - node-role.kubernetes.io/compute=""
    - node.kubernetes.io/compute=""
  node_pools:
    master:
      count: 1
      kubelet_node_labels:
      - node-role.kubernetes.io/master=""
      - node.kubernetes.io/master=""
      kubelet_node_taints:
      - node-role.kubernetes.io/master="":NoSchedule
      - node.kubernetes.io/master="":NoSchedule
    worker:
      count: 1
      kubelet_node_labels:
      - node-role.kubernetes.io/worker=""
      - node.kubernetes.io/worker=""
`)
