package eks

import (
	"encoding/json"

	"github.com/johandry/merger"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

// EKSAmi is the latest or stable EKS AMI ID on us-west-2 region
const EKSAmi = "" // empty string will get the latest

// EKSGPUAmi is the latest or stable EKS AMI ID that supports GPU on us-west-2 region
const EKSGPUAmi = "ami-038a987c6425a84ad" //amazon-eks-node-1.14-v20190906

// Find the latest EKS AMI at https://docs.aws.amazon.com/eks/latest/userguide/eks-optimized-ami.html

// defaultConfig is the default configuration for AWS platform
var defaultConfig = Config{
	Username:              "ec2-user",
	AwsVpcID:              requiredValue + "vpc-8d56b9e9",
	MaxPods:               110,
	MaxMapCount:           262144,
	KubernetesVersion:     "",
	EndpointPublicAccess:  true,
	EndpointPrivateAccess: false,
	ClusterLogsTypes: []string{
		"api",
		"audit",
		"authenticator",
		"controllerManager",
		"scheduler",
	},
	ClusterSecurityGroups: []string{
		requiredValue + "sg-502d9a37",
	},
	IngressSubnets: []string{
		requiredValue + "subnet-5bddc82c",
		requiredValue + "subnet-478a4123",
	},
	DefaultNodePool: defaultNodePool,
	NodePools: map[string]NodePool{
		"compute_fast_ephemeral": ephemeralFastNodePool,
		"compute_slow_ephemeral": ephemeralSlowNodePool,
		"persistent_storage":     persistentNodePool,
	},
	ElasticFileshares: map[string]config.ElasticFileshare{},
}

// var defaultElasticFileshare = ElasticFileshare{
// 	PerformanceMode: "generalPurpose",
// 	ThroughputMode:  "bursting",
// 	Encrypted:       false,
// }

var defaultNodePool = NodePool{
	//AwsAmi: EKSAmi,
	SecurityGroups: []string{
		requiredValue + "sg-502d9a37",
	},
	RootVolumeSize: 100,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/compute=""`,
		`node.kubernetes.io/compute=""`,
	},
	KubeletNodeTaints: []string{
		"",
	},
	Subnets: []string{
		requiredValue + "subnet-5bddc82c",
	},
}

var ephemeralFastNodePool = NodePool{
	Name:            "compute_fast_ephemeral",
	Count:           1,
	AwsInstanceType: "m5d.2xlarge",
	RootVolumeSize:  100,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/compute=""`,
		`node.kubernetes.io/compute=""`,
		"ephemeral-volumes=fast",
	},
}

var ephemeralSlowNodePool = NodePool{
	Name:            "compute_slow_ephemeral",
	Count:           1,
	AwsInstanceType: "m5.2xlarge",
	RootVolumeSize:  100,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/compute=""`,
		`node.kubernetes.io/compute=""`,
		"ephemeral-volumes=slow",
	},
}

var persistentNodePool = NodePool{
	Name:            "persistent_storage",
	Count:           3,
	AwsInstanceType: "i3.2xlarge",
	RootVolumeSize:  100,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/persistent=""`,
		`node.kubernetes.io/persistent=""`,
		"ephemeral-volumes=slow",
		"storage=persistent",
	},
	KubeletNodeTaints: []string{
		"storage=persistent:NoSchedule",
	},
	PGStrategy: "spread",
}

var defaultGPUNodePool = NodePool{
	AwsAmi:          EKSGPUAmi,
	Name:            "gpu",
	Count:           0,
	AwsInstanceType: "p2.xlarge",
	RootVolumeSize:  100,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/gpu=""`,
		`node.kubernetes.io/gpu=""`,
		"ephemeral-volumes=slow",
	},
}

const requiredValue = "# Required value. Example: "

// Config defines the AWS configuration parameters in the Cluster config file
type Config struct {
	ClusterName           string                             `json:"-" yaml:"-" mapstructure:"clustername"`
	Username              string                             `json:"username" yaml:"username" mapstructure:"username"`
	AwsAccessKey          string                             `json:"-" yaml:"-" mapstructure:"-"`
	AwsSecretKey          string                             `json:"-" yaml:"-" mapstructure:"-"`
	AwsSessionToken       string                             `json:"-" yaml:"-" mapstructure:"-"`
	AwsRegion             string                             `json:"aws_region,omitempty" yaml:"aws_region,omitempty" mapstructure:"aws_region"`
	AwsVpcID              string                             `json:"aws_vpc_id" yaml:"aws_vpc_id" mapstructure:"aws_vpc_id"`
	IngressSubnets        []string                           `json:"ingress_subnets" yaml:"ingress_subnets,omitempty" mapstructure:"ingress_subnets"`
	ClusterSecurityGroups []string                           `json:"cluster_security_groups" yaml:"cluster_security_groups" mapstructure:"cluster_security_groups"`
	PrivateKey            string                             `json:"private_key,omitempty" yaml:"private_key,omitempty" mapstructure:"private_key"`
	PrivateKeyFile        string                             `json:"private_key_file" yaml:"private_key_file" mapstructure:"private_key_file"`
	PublicKey             string                             `json:"public_key,omitempty" yaml:"public_key,omitempty" mapstructure:"public_key"`
	PublicKeyFile         string                             `json:"public_key_file" yaml:"public_key_file" mapstructure:"public_key_file"`
	KubernetesVersion     string                             `json:"kubernetes_version" yaml:"kubernetes_version" mapstructure:"kubernetes_version"`
	EndpointPublicAccess  bool                               `json:"endpoint_public_access" yaml:"endpoint_public_access" mapstructure:"endpoint_public_access"`
	EndpointPrivateAccess bool                               `json:"endpoint_private_access" yaml:"endpoint_private_access" mapstructure:"endpoint_private_access"`
	Route53Name           []string                           `json:"route_53_name" yaml:"route_53_name" mapstructure:"route_53_name"`
	ClusterLogsTypes      []string                           `json:"cluster_logs_types" yaml:"cluster_logs_types" mapstructure:"cluster_logs_types"`
	S3Buckets             []string                           `json:"s3_buckets" yaml:"s3_buckets" mapstructure:"s3_buckets"`
	MaxPods               int                                `json:"max_pods,omitempty" yaml:"max_pods,omitempty" mapstructure:"max_pods"`
	MaxMapCount           int                                `json:"max_map_count,omitempty" yaml:"max_map_count,omitempty" mapstructure:"max_map_count"`
	DefaultNodePool       NodePool                           `json:"default_node_pool" yaml:"default_node_pool" mapstructure:"default_node_pool"`
	NodePools             map[string]NodePool                `json:"node_pools" yaml:"node_pools" mapstructure:"node_pools"`
	ElasticFileshares     map[string]config.ElasticFileshare `json:"elastic_fileshares,omitempty" yaml:"elastic_fileshares,omitempty" mapstructure:"elastic_fileshares"`
}

// NodePool defines the settings for group of instances on AWS
type NodePool struct {
	Name              string   `json:"-" yaml:"-" mapstructure:"name"`
	Count             int      `json:"count" yaml:"count" mapstructure:"count"`
	AwsAmi            string   `json:"aws_ami,omitempty" yaml:"aws_ami,omitempty" mapstructure:"aws_ami"`
	AwsInstanceType   string   `json:"aws_instance_type,omitempty" yaml:"aws_instance_type,omitempty" mapstructure:"aws_instance_type"`
	KubeletNodeLabels []string `json:"kubelet_node_labels,omitempty" yaml:"kubelet_node_labels,omitempty" mapstructure:"kubelet_node_labels"`
	KubeletNodeTaints []string `json:"kubelet_node_taints,omitempty" yaml:"kubelet_node_taints,omitempty" mapstructure:"kubelet_node_taints"`
	RootVolumeSize    int      `json:"root_volume_size,omitempty" yaml:"root_volume_size,omitempty" mapstructure:"root_volume_size"`
	PGStrategy        string   `json:"placementgroup_strategy,omitempty" yaml:"placementgroup_strategy,omitempty" mapstructure:"placementgroup_strategy"`
	Subnets           []string `json:"worker_pool_subnets,omitempty" yaml:"worker_pool_subnets,omitempty" mapstructure:"worker_pool_subnets"`
	SecurityGroups    []string `json:"security_groups,omitempty" yaml:"security_groups,omitempty" mapstructure:"security_groups"`
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
	var partialEnvConf map[string]string
	var nodePoolsEnvConf map[string]string
	var elasticPoolsEnvConf map[string]string
	partialEnvConf, nodePoolsEnvConf = utils.RemoveEnv(envConf, "node_pools")
	partialEnvConf, elasticPoolsEnvConf = utils.RemoveEnv(partialEnvConf, "elastic_fileshares")
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
	c.MergeElasticFileshares(elasticPoolsEnvConf)
	return nil
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
		case "s3_buckets":
			c.S3Buckets = config.GetListFromInterface(v)
		case "route_53_name":
			c.Route53Name = config.GetListFromInterface(v)
		case "ingress_subnets":
			c.IngressSubnets = config.GetListFromInterface(v)
		case "cluster_security_groups":
			c.ClusterSecurityGroups = config.GetListFromInterface(v)
		case "cluster_logs_types":
			c.ClusterLogsTypes = config.GetListFromInterface(v)
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
		case "worker_pool_subnets":
			n.Subnets = config.GetListFromInterface(v)
		case "security_groups":
			n.SecurityGroups = config.GetListFromInterface(v)
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
