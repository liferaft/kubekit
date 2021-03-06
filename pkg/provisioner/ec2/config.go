package ec2

import (
	"encoding/json"

	"github.com/johandry/merger"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

// KubeOS is the latest image generated for KubeKit
const KubeOS = "ami-0b8485a3553c5d032"

// defaultConfig is the default configuration for AWS platform
var defaultConfig = Config{
	Username:          "ec2-user",
	AwsEnv:            "aws-k8s",
	AwsVpcID:          requiredValue + "vpc-8d56b9e9",
	KubeAPISSLPort:    8081,
	DisableMasterHA:   true,
	KubeVIPAPISSLPort: 8443,
	TimeServers:       []string{"169.254.169.123"},
	DefaultNodePool:   defaultNodePool,
	NodePools: map[string]NodePool{
		"master": defaultMasterNodePool,
		"worker": defaultWorkerNodePool,
	},
	ElasticFileshares: map[string]config.ElasticFileshare{},
}

var defaultNodePool = NodePool{
	SecurityGroups:    []string{requiredValue + "sg-502d9a37"},
	ConnectionTimeout: "5m",
	Ami:               KubeOS,
	InstanceType:      "m4.2xlarge",
	Subnets:           []string{requiredValue + "subnet-5bddc82c"},
	RootVolumeSize:    200,
	RootVolumeType:    "gp2",
	PGStrategy:        "",
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/compute=""`,
		`node.kubernetes.io/compute=""`,
	},
}

var defaultMasterNodePool = NodePool{
	Name:  "master",
	Count: 1,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/master=""`,
		`node.kubernetes.io/master=""`,
	},
	KubeletNodeTaints: []string{
		`node-role.kubernetes.io/master="":NoSchedule`,
		`node.kubernetes.io/master="":NoSchedule`,
	},
}

var defaultWorkerNodePool = NodePool{
	Name:  "worker",
	Count: 1,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/worker=""`,
		`node.kubernetes.io/worker=""`,
	},
}

const requiredValue = "# Required value. Example: "

// Config defines the AWS configuration parameters in the Cluster config file
type Config struct {
	ClusterName             string                             `json:"-" yaml:"-" mapstructure:"clustername"`
	AwsEnv                  string                             `json:"aws_env" yaml:"aws_env" mapstructure:"aws_env"`
	KubeAPISSLPort          int                                `json:"kube_api_ssl_port" yaml:"kube_api_ssl_port" mapstructure:"kube_api_ssl_port"`
	DisableMasterHA         bool                               `json:"disable_master_ha" yaml:"disable_master_ha" mapstructure:"disable_master_ha"`
	KubeVirtualIPShortname  string                             `json:"kube_virtual_ip_shortname" yaml:"kube_virtual_ip_shortname" mapstructure:"kube_virtual_ip_shortname"`
	KubeVirtualIPApi        string                             `json:"kube_virtual_ip_api" yaml:"kube_virtual_ip_api" mapstructure:"kube_virtual_ip_api"`
	KubeVIPAPISSLPort       int                                `json:"kube_vip_api_ssl_port" yaml:"kube_vip_api_ssl_port" mapstructure:"kube_vip_api_ssl_port"`
	PublicAPIServerDNSName  string                             `json:"public_apiserver_dns_name" yaml:"public_apiserver_dns_name" mapstructure:"public_apiserver_dns_name"`
	PrivateAPIServerDNSName string                             `json:"private_apiserver_dns_name" yaml:"private_apiserver_dns_name" mapstructure:"private_apiserver_dns_name"`
	Username                string                             `json:"username" yaml:"username" mapstructure:"username"`
	AwsAccessKey            string                             `json:"-" yaml:"-" mapstructure:"-"`
	AwsSecretKey            string                             `json:"-" yaml:"-" mapstructure:"-"`
	AwsSessionToken         string                             `json:"-" yaml:"-" mapstructure:"-"`
	AwsRegion               string                             `json:"aws_region,omitempty" yaml:"aws_region,omitempty" mapstructure:"aws_region"`
	AwsVpcID                string                             `json:"aws_vpc_id" yaml:"aws_vpc_id" mapstructure:"aws_vpc_id"`
	PrivateKey              string                             `json:"private_key,omitempty" yaml:"private_key,omitempty" mapstructure:"private_key"`
	PrivateKeyFile          string                             `json:"private_key_file" yaml:"private_key_file" mapstructure:"private_key_file"`
	PublicKey               string                             `json:"public_key,omitempty" yaml:"public_key,omitempty" mapstructure:"public_key"`
	PublicKeyFile           string                             `json:"public_key_file" yaml:"public_key_file" mapstructure:"public_key_file"`
	ConfigureFromPrivateNet bool                               `json:"configure_from_private_net" yaml:"configure_from_private_net" mapstructure:"configure_from_private_net"`
	DNSServers              []string                           `json:"dns_servers" yaml:"dns_servers" mapstructure:"dns_servers"`
	DNSSearch               []string                           `json:"dns_search" yaml:"dns_search" mapstructure:"dns_search"`
	TimeServers             []string                           `json:"time_servers" yaml:"time_servers" mapstructure:"time_servers"`
	DefaultNodePool         NodePool                           `json:"default_node_pool" yaml:"default_node_pool" mapstructure:"default_node_pool"`
	NodePools               map[string]NodePool                `json:"node_pools" yaml:"node_pools" mapstructure:"node_pools"`
	ElasticFileshares       map[string]config.ElasticFileshare `json:"elastic_fileshares,omitempty" yaml:"elastic_fileshares,omitempty" mapstructure:"elastic_fileshares"`
}

// NodePool defines the settings for group of instances on AWS
type NodePool struct {
	Name              string   `json:"-" yaml:"-" mapstructure:"name"`
	Count             int      `json:"count" yaml:"count" mapstructure:"count"`
	ConnectionTimeout string   `json:"connection_timeout,omitempty" yaml:"connection_timeout,omitempty" mapstructure:"connection_timeout"`
	Ami               string   `json:"aws_ami,omitempty" yaml:"aws_ami,omitempty" mapstructure:"aws_ami"`
	InstanceType      string   `json:"aws_instance_type,omitempty" yaml:"aws_instance_type,omitempty" mapstructure:"aws_instance_type"`
	RootVolumeSize    int      `json:"root_volume_size,omitempty" yaml:"root_volume_size,omitempty" mapstructure:"root_volume_size"`
	RootVolumeType    string   `json:"root_volume_type,omitempty" yaml:"root_volume_type,omitempty" mapstructure:"root_volume_type"`
	PGStrategy        string   `json:"placementgroup_strategy,omitempty" yaml:"placementgroup_strategy,omitempty" mapstructure:"placementgroup_strategy"`
	SecurityGroups    []string `json:"security_groups,omitempty" yaml:"security_groups,omitempty" mapstructure:"security_groups"`
	Subnets           []string `json:"subnets,omitempty" yaml:"subnets,omitempty" mapstructure:"subnets"`
	KubeletNodeLabels []string `json:"kubelet_node_labels,omitempty" yaml:"kubelet_node_labels,omitempty" mapstructure:"kubelet_node_labels"`
	KubeletNodeTaints []string `json:"kubelet_node_taints,omitempty" yaml:"kubelet_node_taints,omitempty" mapstructure:"kubelet_node_taints"`
}

// MergeNodePools merges the node pools in this configuration with the given
// environment configuration for node pools
func (c *Config) MergeNodePools(nodePoolsEnvConf map[string]string) {
	if len(nodePoolsEnvConf) == 0 {
		return
	}
	nodePoolsMap := merger.TransformMap(nodePoolsEnvConf)
	nodePools, ok := nodePoolsMap["node_pools"].(map[string]interface{})
	if !ok {
		return
	}
	for nodePool := range nodePools {
		n := NodePool{}
		nodePoolEnv := utils.TrimLeft(nodePoolsEnvConf, "node_pools__"+nodePool+"__")
		merger.Merge(&n, nodePoolEnv, c.NodePools[nodePool])
		c.NodePools[nodePool] = n
	}
}

// NewConfigFrom returns a new AWS configuration from a map, usually from a
// cluster config file
func NewConfigFrom(m map[interface{}]interface{}) *Config {
	c := &Config{}
	c.MergeWithMapConfig(m)
	return c
}

// MergeWithEnv merges this configuration with the given configuration in
// a map[string]string, usually from environment variables
func (c *Config) MergeWithEnv(envConf map[string]string, conf ...Config) error {
	partialEnvConf, nodePoolsEnvConf := utils.RemoveEnv(envConf, "node_pools")
	var err error
	if len(conf) == 0 {
		err = merger.Merge(c, partialEnvConf)
	} else {
		err = merger.Merge(c, partialEnvConf, conf[0])
	}
	if err != nil {
		return err
	}
	c.MergeNodePools(nodePoolsEnvConf)
	return nil
}

// MergeElasticFileshares merges the elastic fileshares in this configuration with the given
// environment configuration for elastic fileshares
func (c *Config) MergeElasticFileshares(elasticFilesharesEnvConf map[string]string) {
	if len(elasticFilesharesEnvConf) == 0 {
		return
	}
	elasticFilesharesMap := merger.TransformMap(elasticFilesharesEnvConf)
	elasticFileshares, ok := elasticFilesharesMap["elastic_fileshares"].(map[string]interface{})
	if !ok {
		return
	}
	for elasticFileshare := range elasticFileshares {
		n := config.ElasticFileshare{}
		elasticFileshareEnv := utils.TrimLeft(elasticFilesharesEnvConf, "elastic_fileshares__"+elasticFileshare+"__")
		merger.Merge(&n, elasticFileshareEnv, c.ElasticFileshares[elasticFileshare])
		c.ElasticFileshares[elasticFileshare] = n
	}
}

// MergeWithMapConfig merges this configuration with the given configuration in
// a map[string], usually from a cluster config file
func (c *Config) MergeWithMapConfig(m map[interface{}]interface{}) {
	for k, v := range m {
		// vType := reflect.ValueOf(v).Type().Kind().String()
		name := k.(string)
		switch name {
		case "default_node_pool":
			m1 := v.(map[interface{}]interface{})
			c.DefaultNodePool = getNodePool(m1)
		case "node_pools":
			m1 := v.(map[interface{}]interface{})
			c.NodePools = getNodePools(m1)
		case "elastic_fileshares":
			m1 := v.(map[interface{}]interface{})
			c.ElasticFileshares = config.GetElasticFileshares(m1)
		case "dns_servers":
			c.DNSServers = config.GetListFromInterface(v)
		case "dns_search":
			c.DNSSearch = config.GetListFromInterface(v)
		case "time_servers":
			c.TimeServers = config.GetListFromInterface(v)
		default:
			config.SetField(c, name, v)
		}
	}
}

func getNodePool(m map[interface{}]interface{}) NodePool {
	n := NodePool{}
	for k, v := range m {
		name := k.(string)
		switch name {
		case "security_groups":
			n.SecurityGroups = config.GetListFromInterface(v)
		case "subnets":
			n.Subnets = config.GetListFromInterface(v)
		case "kubelet_node_labels":
			n.KubeletNodeLabels = config.GetListFromInterface(v)
		case "kubelet_node_taints":
			n.KubeletNodeTaints = config.GetListFromInterface(v)
		default:
			config.SetField(&n, name, v)
		}
	}
	return n
}

func getNodePools(m map[interface{}]interface{}) map[string]NodePool {
	nPools := make(map[string]NodePool, len(m))
	for k, v := range m {
		m1 := v.(map[interface{}]interface{})
		nPool := getNodePool(m1)
		nPools[k.(string)] = nPool
	}
	return nPools
}

func (c *Config) copyWithDefaults() Config {
	cfg := *c
	marshalled, _ := json.Marshal(c.DefaultNodePool)
	nodePools := make(map[string]NodePool, len(cfg.NodePools))
	for k, v := range cfg.NodePools {
		n := NodePool{}
		json.Unmarshal(marshalled, &n)

		a, _ := json.Marshal(v)
		json.Unmarshal(a, &n)

		n.Name = k
		nodePools[k] = n
	}
	cfg.NodePools = nodePools
	return cfg
}
