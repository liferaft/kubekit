package openstack

import (
	"encoding/json"
	"fmt"

	"github.com/johandry/merger"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

// KubekitOS is the latest or stable KubekitOS template name
const KubekitOS = "c0b5a78d-6b21-4f44-a839-1a16c9f5e525"

// DefaultConfig is the default configuration for a openstack platform
var defaultConfig = Config{
	Username:            "kubekit",
	KubeAPISSLPort:      6443,
	DisableMasterHA:     true,
	KubeVIPAPISSLPort:   8443,
	DefaultNodePool:     defaultNodePool,
	OpenstackTenantName: requiredValue + "kubekit",
	OpenstackDomainName: "Default",
	OpenstackNetName:    requiredValue + "kubekit-net",
	DNSServers:          []string{"153.64.180.100", "153.64.251.200"},
	TimeServers:         []string{"0.us.pool.ntp.org", "1.us.pool.ntp.org", "2.us.pool.ntp.org"},
	NodePools: map[string]NodePool{
		"master": defaultMasterNodePool,
		"worker": defaultWorkerNodePool,
	},
	// TODO: other default values?
}

var defaultNodePool = NodePool{
	Count:             1,
	OpenstackImageID:  KubekitOS,
	OpenstackFlavorID: "29",
	SecurityGroups:    []string{"default"},
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

// Config defines the Openstack configuration parameters in the Cluster config file
type Config struct {
	// Following fields are platform generic
	// TODO: refactor common fields into common data structure
	ClusterName             string   `json:"-" yaml:"-" mapstructure:"clustername"`
	KubeAPISSLPort          int      `json:"kube_api_ssl_port" yaml:"kube_api_ssl_port" mapstructure:"kube_api_ssl_port"`
	DisableMasterHA         bool     `json:"disable_master_ha" yaml:"disable_master_ha" mapstructure:"disable_master_ha"`
	KubeVirtualIPShortname  string   `json:"kube_virtual_ip_shortname" yaml:"kube_virtual_ip_shortname" mapstructure:"kube_virtual_ip_shortname"`
	KubeVirtualIPApi        string   `json:"kube_virtual_ip_api" yaml:"kube_virtual_ip_api" mapstructure:"kube_virtual_ip_api"`
	KubeVIPAPISSLPort       int      `json:"kube_vip_api_ssl_port" yaml:"kube_vip_api_ssl_port" mapstructure:"kube_vip_api_ssl_port"`
	PublicAPIServerDNSName  string   `json:"public_apiserver_dns_name" yaml:"public_apiserver_dns_name" mapstructure:"public_apiserver_dns_name"`
	PrivateAPIServerDNSName string   `json:"private_apiserver_dns_name" yaml:"private_apiserver_dns_name" mapstructure:"private_apiserver_dns_name"`
	Username                string   `json:"username" yaml:"username" mapstructure:"username"`
	PrivateKey              string   `json:"private_key,omitempty" yaml:"private_key,omitempty" mapstructure:"private_key"`
	PrivateKeyFile          string   `json:"private_key_file" yaml:"private_key_file" mapstructure:"private_key_file"`
	PublicKey               string   `json:"public_key,omitempty" yaml:"public_key,omitempty" mapstructure:"public_key"`
	PublicKeyFile           string   `json:"public_key_file" yaml:"public_key_file" mapstructure:"public_key_file"`
	DNSServers              []string `json:"dns_servers" yaml:"dns_servers" mapstructure:"dns_servers"`
	DNSSearch               []string `json:"dns_search" yaml:"dns_search" mapstructure:"dns_search"`
	TimeServers             []string `json:"time_servers" yaml:"time_servers" mapstructure:"time_servers"`

	// Following are openstack specific fields
	OpenstackTenantName string              `json:"openstack_tenant_name,omitempty" yaml:"openstack_tenant_name" mapstructure:"openstack_tenant_name"`
	OpenstackAuthURL    string              `json:"-" yaml:"-" mapstructure:"-"`
	OpenstackUserName   string              `json:"-" yaml:"-" mapstructure:"-"`
	OpenstackPassword   string              `json:"-" yaml:"-" mapstructure:"-"`
	OpenstackDomainName string              `json:"openstack_domain_name,omitempty" yaml:"openstack_domain_name" mapstructure:"openstack_domain_name"`
	OpenstackRegion     string              `json:"openstack_region,omitempty" yaml:"openstack_region" mapstructure:"openstack_region"`
	OpenstackNetName    string              `json:"openstack_net_name,omitempty" yaml:"openstack_net_name" mapstructure:"openstack_net_name"`
	DefaultNodePool     NodePool            `json:"default_node_pool" yaml:"default_node_pool" yaml:"default_node_pool" mapstructure:"default_node_pool"`
	NodePools           map[string]NodePool `json:"node_pools" yaml:"node_pools" yaml:"node_pools" mapstructure:"node_pools"`
}

// NodePool defines the settings for group of instances on Openstack
type NodePool struct {
	Name              string   `json:"-" yaml:"-" mapstructure:"name"`
	Count             int      `json:"count" yaml:"count" mapstructure:"count"`
	OpenstackImageID  string   `json:"openstack_image_id,omitempty" yaml:"openstack_image_id,omitempty" mapstructure:"openstack_image_id"`
	OpenstackFlavorID string   `json:"openstack_flavor_id,omitempty" yaml:"openstack_flavor_id,omitempty" mapstructure:"openstack_flavor_id"`
	SecurityGroups    []string `json:"security_groups,omitempty" yaml:"security_groups,omitempty" mapstructure:"security_groups"`
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

// NewConfigFrom returns a new openstack  configuration from a map, usually from a
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
		case "dns_servers":
			c.DNSServers = config.GetListFromInterface(v)
		case "dns_search":
			c.DNSSearch = config.GetListFromInterface(v)
		case "time_servers":
			c.TimeServers = config.GetListFromInterface(v)
		case "openstack_password", "openstack_user_name", "openstack_auth_url":
			panic(fmt.Errorf("field %s is obsolete, please remove it from cluster.yaml and use kubekit login for credentials", name))
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
