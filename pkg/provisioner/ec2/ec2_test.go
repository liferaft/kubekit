package ec2

import (
	"os"
	"testing"

	"github.com/johandry/log"
	"github.com/kraken/ui"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	"github.com/stretchr/testify/assert"
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
		"AWS_ACCESS_KEY_ID":     "my_access_key",
		"AWS_SECRET_ACCESS_KEY": "my_secret_key",
		"AWS_SESSION_TOKEN":     "my_session_token",
		"AWS_DEFAULT_REGION":    "my_aws_region",
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
			name: "create from nil config",
			args: args{
				config: nil,
			},
			want: defaultAWS([]string{"my_access_key", "my_secret_key", "my_session_token", "my_aws_region"}),
		},
		{
			name: "create from config",
			args: args{
				config: newConfigFromYaml(),
			},
			want: &Platform{
				name: "ec2",
				config: &Config{
					ClusterName:       "testCluster",
					Username:          "ec2-user",
					AwsVpcID:          "vpc-8d56b9e9",
					ElasticFileshares: map[string]config.ElasticFileshare{},
					DefaultNodePool: NodePool{
						SecurityGroups:    []string{"sg-502d9a37"},
						ConnectionTimeout: "5m",
						Ami:               "ami-abc65dd3",
						Subnets:           []string{"subnet-5bddc82c"},
						InstanceType:      "m4.2xlarge",
						RootVolumeSize:    200,
						RootVolumeType:    "gp2",
						KubeletNodeLabels: []string{
							`node-role.kubernetes.io/compute=""`,
							`node.kubernetes.io/compute=""`,
						},
					},
					NodePools: map[string]NodePool{
						"master": NodePool{
							Count: 1,
							KubeletNodeLabels: []string{
								`node-role.kubernetes.io/master=""`,
								`node.kubernetes.io/master=""`,
							},
							KubeletNodeTaints: []string{
								`node-role.kubernetes.io/master="":NoSchedule`,
								`node.kubernetes.io/master="":NoSchedule`,
							},
						},
						"worker": NodePool{
							Count: 1,
							KubeletNodeLabels: []string{
								`node-role.kubernetes.io/worker=""`,
								`node.kubernetes.io/worker=""`,
							},
						},
					},
					AwsAccessKey:    "my_access_key",
					AwsSecretKey:    "my_secret_key",
					AwsSessionToken: "my_session_token",
					AwsRegion:       "my_aws_region",
				},
				ui: tUI,
			},
		},
		{
			name: "create from config from nil string in yaml",
			args: args{
				config: newConfigFromBadYaml(),
			},
			want: &Platform{
				name: "ec2",
				config: &Config{
					ClusterName:       "testCluster",
					Username:          "ec2-user",
					AwsVpcID:          "vpc-8d56b9e9",
					ElasticFileshares: map[string]config.ElasticFileshare{},
					DefaultNodePool: NodePool{
						SecurityGroups:    []string{},
						ConnectionTimeout: "",
						Ami:               "ami-abc65dd3",
						Subnets:           []string{"subnet-5bddc82c"},
						InstanceType:      "m4.2xlarge",
						RootVolumeSize:    200,
						RootVolumeType:    "gp2",
						KubeletNodeLabels: []string{
							`node-role.kubernetes.io/compute=""`,
							`node.kubernetes.io/compute=""`,
						},
					},
					NodePools: map[string]NodePool{
						"master": NodePool{
							Count: 1,
							KubeletNodeLabels: []string{
								`node-role.kubernetes.io/master=""`,
								`node.kubernetes.io/master=""`,
							},
							KubeletNodeTaints: []string{
								`node-role.kubernetes.io/master="":NoSchedule`,
								`node.kubernetes.io/master="":NoSchedule`,
							},
						},
						"worker": NodePool{
							Count: 1,
							KubeletNodeLabels: []string{
								`node-role.kubernetes.io/worker=""`,
								`node.kubernetes.io/worker=""`,
							},
						},
					},
					AwsAccessKey:    "my_access_key",
					AwsSecretKey:    "my_secret_key",
					AwsSessionToken: "my_session_token",
					AwsRegion:       "my_aws_region",
				},
				ui: tUI,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := CreateFrom("testCluster", tt.args.config, []string{"my_access_key", "my_secret_key", "my_session_token", "my_aws_region"}, tUI, version)
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
			name: "new AWS with no env vars",
			args: args{
				envConfig: nil,
			},
			want: defaultAWS([]string{"my_access_key", "my_secret_key", "my_session_token", "my_aws_region"}),
		},
		{
			name: "new AWS with a username in env vars",
			args: args{
				envConfig: map[string]string{
					"aws_username": "myuser",
					"username":     "fakeuser",
				},
			},
			want: &Platform{
				name: "ec2",
				config: &Config{ClusterName: "testCluster",
					AwsEnv:                  "aws-k8s",
					KubeAPISSLPort:          8081,
					DisableMasterHA:         true,
					KubeVIPAPISSLPort:       8443,
					Username:                "fakeuser",
					AwsAccessKey:            "my_access_key",
					AwsSecretKey:            "my_secret_key",
					AwsSessionToken:         "my_session_token",
					AwsRegion:               "my_aws_region",
					AwsVpcID:                "# Required value. Example: vpc-8d56b9e9",
					ConfigureFromPrivateNet: false,
					TimeServers:             []string{"169.254.169.123"},
					ElasticFileshares:       map[string]config.ElasticFileshare{},
					DefaultNodePool: NodePool{
						SecurityGroups:    []string{"# Required value. Example: sg-502d9a37"},
						ConnectionTimeout: "5m",
						Ami:               "ami-0b8485a3553c5d032",
						InstanceType:      "m4.2xlarge",
						Subnets:           []string{"# Required value. Example: subnet-5bddc82c"},
						RootVolumeSize:    200,
						RootVolumeType:    "gp2",
						KubeletNodeLabels: []string{
							`node-role.kubernetes.io/compute=""`,
							`node.kubernetes.io/compute=""`,
						},
					},
					NodePools: map[string]NodePool{
						"master": NodePool{
							Count: 1,
							Name:  "master",
							KubeletNodeLabels: []string{
								`node-role.kubernetes.io/master=""`,
								`node.kubernetes.io/master=""`,
							},
							KubeletNodeTaints: []string{
								`node-role.kubernetes.io/master="":NoSchedule`,
								`node.kubernetes.io/master="":NoSchedule`,
							},
						},
						"worker": NodePool{
							Count: 1,
							Name:  "worker",
							KubeletNodeLabels: []string{
								`node-role.kubernetes.io/worker=""`,
								`node.kubernetes.io/worker=""`,
							},
						},
					},
				},
			},
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

func defaultAWS(credentials []string) *Platform {
	p := &Platform{
		name:    "ec2",
		config:  &defaultConfig,
		ui:      tUI,
		version: version,
	}
	p.config.ClusterName = "testCluster"
	p.Credentials(credentials...)
	return p
}

func newConfig() *Config {
	c := &Config{
		ClusterName:       "testCluster",
		Username:          "ec2-user",
		ElasticFileshares: map[string]config.ElasticFileshare{},
		DefaultNodePool: NodePool{
			ConnectionTimeout: "5m",
			Ami:               KubeOS,
			InstanceType:      "m4.2xlarge",
			RootVolumeSize:    200,
			RootVolumeType:    "gp2",
			KubeletNodeLabels: []string{
				`node-role.kubernetes.io/compute=""`,
				`node.kubernetes.io/compute=""`,
			},
		},
		NodePools: map[string]NodePool{
			"master": NodePool{
				Count: 1,
				KubeletNodeLabels: []string{
					`node-role.kubernetes.io/master=""`,
					`node.kubernetes.io/master=""`,
				},
				KubeletNodeTaints: []string{
					`node-role.kubernetes.io/master="":NoSchedule`,
					`node.kubernetes.io/master="":NoSchedule`,
				},
			},
			"worker": NodePool{
				Count: 1,
				KubeletNodeLabels: []string{
					`node-role.kubernetes.io/worker=""`,
					`node.kubernetes.io/worker=""`,
				},
			},
		},
	}
	return c
}

type miniConfig struct {
	Platforms map[string]interface{} `json:"platforms" yaml:"platforms" mapstructure:"platform"`
}

func newConfigFromYaml() map[interface{}]interface{} {
	yamlStr := `
platforms:
  ec2:
    username: ec2-user
    aws_vpc_id: vpc-8d56b9e9
    private_key_file: ""
    public_key_file: ""
    default_node_pool:
      connection_timeout: 5m
      aws_ami: ami-abc65dd3
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
    elastic_fileshares: {}
`
	c := miniConfig{}
	yaml.Unmarshal([]byte(yamlStr), &c)
	p := c.Platforms["ec2"]
	cnf := p.(map[interface{}]interface{})
	return cnf
}

func newConfigFromBadYaml() map[interface{}]interface{} {
	yamlStr := `
platforms:
  ec2:
    username: ec2-user
    aws_vpc_id: vpc-8d56b9e9
    private_key_file: ""
    public_key_file: ""
    default_node_pool:
      connection_timeout:
      aws_ami: ami-abc65dd3
      aws_instance_type: m4.2xlarge
      root_volume_size: 200
      root_volume_type: gp2
      security_groups: []
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
    elastic_fileshares: {}
`
	c := miniConfig{}
	yaml.Unmarshal([]byte(yamlStr), &c)
	p := c.Platforms["ec2"]
	cnf := p.(map[interface{}]interface{})
	return cnf
}
