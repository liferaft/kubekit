package vra

import (
	"github.com/johandry/merger"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

// DefaultConfig is the default configuration for a vrealize platform
var defaultConfig = Config{
	Username:          "root",
	Password:          "TCAMPass123",
	PrivateKeyFile:    requiredValue + "/home/username/.ssh/id_rsa",
	PublicKeyFile:     requiredValue + "/home/username/.ssh/id_rsa.pub",
	APIAddress:        requiredValue + "39.80.0.50",
	KubeAPISSLPort:    6443,
	KubeVIPAPISSLPort: 8443,
	KubeVirtualIPApi:  "",
	DisableMasterHA:   true,
	DNSServers:        []string{"153.64.180.100", "153.64.251.200"},
	TimeServers:       []string{"0.us.pool.ntp.org", "1.us.pool.ntp.org", "2.us.pool.ntp.org"},
	DefaultNodePool:   defaultNodePool,
	NodePools: map[string]NodePool{
		"master": defaultMasterNodePool,
		"worker": defaultWorkerNodePool,
	},
}

var defaultNodePool = NodePool{
	Name:  "default",
	Count: 0,
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
	Nodes: []Node{
		Node{
			PublicIP:   requiredValue + "10.25.150.100",
			PrivateIP:  requiredValue + "39.80.50.100",
			PublicDNS:  requiredValue + "master1.vra.test",
			PrivateDNS: requiredValue + "master1",
		},
	},
}

var defaultWorkerNodePool = NodePool{
	Name:  "worker",
	Count: 3,
	KubeletNodeLabels: []string{
		`node-role.kubernetes.io/worker=""`,
		`node.kubernetes.io/worker=""`,
	},
	Nodes: []Node{
		Node{
			PublicIP:   requiredValue + "10.25.150.101",
			PrivateIP:  requiredValue + "39.80.50.101",
			PublicDNS:  requiredValue + "worker1.vra.test",
			PrivateDNS: requiredValue + "worker1",
		},
		Node{
			PublicIP:   requiredValue + "10.25.150.102",
			PrivateIP:  requiredValue + "39.80.50.102",
			PublicDNS:  requiredValue + "worker2.vra.test",
			PrivateDNS: requiredValue + "worker2",
		},
		Node{
			PublicIP:   requiredValue + "10.25.150.103",
			PrivateIP:  requiredValue + "39.80.50.103",
			PublicDNS:  requiredValue + "worker3.vra.test",
			PrivateDNS: requiredValue + "worker3",
		},
	},
}

const requiredValue = "# Required value. Example: "

// Config defines the vrealize configuration parameters in the Cluster config file
type Config struct {
	clusterName            string
	APIAddress             string              `json:"api_address" yaml:"api_address" mapstructure:"api_address"`
	KubeAPISSLPort         int                 `json:"kube_api_ssl_port" yaml:"kube_api_ssl_port" mapstructure:"kube_api_ssl_port"`
	DisableMasterHA        bool                `json:"disable_master_ha" yaml:"disable_master_ha" mapstructure:"disable_master_ha"`
	KubeVirtualIPShortname string              `json:"kube_virtual_ip_shortname" yaml:"kube_virtual_ip_shortname" mapstructure:"kube_virtual_ip_shortname"`
	KubeVirtualIPApi       string              `json:"kube_virtual_ip_api" yaml:"kube_virtual_ip_api" mapstructure:"kube_virtual_ip_api"`
	KubeVIPAPISSLPort      int                 `json:"kube_vip_api_ssl_port" yaml:"kube_vip_api_ssl_port" mapstructure:"kube_vip_api_ssl_port"`
	Username               string              `json:"username" yaml:"username" mapstructure:"username"`
	Password               string              `json:"password" yaml:"password" mapstructure:"password"`
	PrivateKey             string              `json:"private_key,omitempty" yaml:"private_key,omitempty" mapstructure:"private_key"`
	PrivateKeyFile         string              `json:"private_key_file" yaml:"private_key_file" mapstructure:"private_key_file"`
	PublicKey              string              `json:"public_key,omitempty" yaml:"public_key,omitempty" mapstructure:"public_key"`
	PublicKeyFile          string              `json:"public_key_file" yaml:"public_key_file" mapstructure:"public_key_file"`
	DNSServers             []string            `json:"dns_servers" yaml:"dns_servers" mapstructure:"dns_servers"`
	DNSSearch              []string            `json:"dns_search" yaml:"dns_search" mapstructure:"dns_search"`
	TimeServers            []string            `json:"time_servers" yaml:"time_servers" mapstructure:"time_servers"`
	DefaultNodePool        NodePool            `json:"default_node_pool" yaml:"default_node_pool" mapstructure:"default_node_pool"`
	NodePools              map[string]NodePool `json:"node_pools,omitempty" yaml:"node_pools,omitempty" mapstructure:"node_pools"`
}

// NodePool defines the settings for group of instances on vra
type NodePool struct {
	Name              string   `json:"-" yaml:"-" mapstructure:"name"`
	Count             int      `json:"count" yaml:"count" mapstructure:"count"`
	KubeletNodeLabels []string `json:"kubelet_node_labels,omitempty" yaml:"kubelet_node_labels,omitempty" mapstructure:"kubelet_node_labels"`
	KubeletNodeTaints []string `json:"kubelet_node_taints,omitempty" yaml:"kubelet_node_taints,omitempty" mapstructure:"kubelet_node_taints"`
	Nodes             []Node   `json:"address_pool" yaml:"address_pool" mapstructure:"address_pool"`
}

// Node encapsulate a node or host information to save in the configuration
type Node struct {
	PublicIP   string `json:"public_ip" yaml:"public_ip" mapstructure:"public_ip"`
	PrivateIP  string `json:"private_ip" yaml:"private_ip" mapstructure:"private_ip"`
	PublicDNS  string `json:"public_dns" yaml:"public_dns" mapstructure:"public_dns"`
	PrivateDNS string `json:"private_dns" yaml:"private_dns" mapstructure:"private_dns"`
}

func getNodesFromInterface(m []interface{}) []Node {
	nodes := []Node{}
	if len(m) > 0 {
		for _, n := range m {
			node := Node{}
			fieldMap := n.(map[interface{}]interface{})
			for k, v := range fieldMap {
				config.SetField(&node, k.(string), v)
			}
			nodes = append(nodes, node)
		}
	}
	return nodes
}

func getNodePool(m map[interface{}]interface{}) NodePool {
	n := NodePool{}
	for k, v := range m {
		name := k.(string)
		switch name {
		case "address_pool":
			n.Nodes = getNodesFromInterface(v.([]interface{}))
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

// NewConfigFrom returns a new Stacki configuration from a map, usually from a
// cluster config file
func NewConfigFrom(m map[interface{}]interface{}) *Config {
	c := &Config{}
	c.MergeWithMapConfig(m)
	return c
}

// MergeWithEnv merges this configuration with the given configuration in
// a map[string]string, usually from environment variables
func (c *Config) MergeWithEnv(envConf map[string]string, conf ...Config) error {
	partialEnvConf, nodesEnvConf := utils.RemoveEnv(envConf, "nodes")
	var err error
	if len(conf) == 0 {
		err = merger.Merge(c, partialEnvConf)
	} else {
		err = merger.Merge(c, partialEnvConf, conf[0])
	}
	if err != nil {
		return err
	}
	c.MergeNodePools(nodesEnvConf)
	return nil
}

// MergeWithMapConfig merges this configuration with the given configuration in
// a map[string], usually from a cluster config file
func (c *Config) MergeWithMapConfig(m map[interface{}]interface{}) {
	for k, v := range m {
		name := k.(string)
		switch name {
		case "default_node_pool":
			m1 := v.(map[interface{}]interface{})
			c.DefaultNodePool = getNodePool(m1)
		case "node_pools":
			m1 := v.(map[interface{}]interface{})
			c.NodePools = getNodePools(m1)
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
