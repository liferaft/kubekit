package vsphere

import (
	"encoding/json"
	"fmt"

	"github.com/johandry/merger"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

// KubeOS is the latest or stable KubeOS template name
const KubeOS = "Templates/LinkedClones/vmware-kubekit-os-01.48-19.11.00-200G"

// DefaultConfig is the default configuration for a vSphere platform
var defaultConfig = Config{
	Username:          "root",
	Datacenter:        requiredValue + "Vagrant",
	Datastore:         requiredValue + "sd_labs_19_vgrnt_dsc/sd_labs_19_vgrnt03",
	ResourcePool:      requiredValue + "sd_vgrnt_01/Resources/vagrant01",
	VsphereNet:        requiredValue + "dvpg_vm_550",
	Folder:            requiredValue + "Discovered virtual machine/kubekit",
	KubeAPISSLPort:    6443,
	DisableMasterHA:   true,
	KubeVIPAPISSLPort: 8443,
	DNSServers:        []string{"153.64.180.100", "153.64.251.200"},
	TimeServers:       []string{"0.us.pool.ntp.org", "1.us.pool.ntp.org", "2.us.pool.ntp.org"},
	DefaultNodePool:   defaultNodePool,
	NodePools: map[string]NodePool{
		"master": defaultMasterNodePool,
		"worker": defaultWorkerNodePool,
	},
}

var defaultNodePool = NodePool{
	TemplateName: KubeOS,
	CPUs:         8,
	Memory:       24576,
	RootVolSize:  200,
	LinkedClone:  true,
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

// Config defines the vSphere configuration parameters in the Cluster config file
type Config struct {
	ClusterName             string              `json:"-" yaml:"-" mapstructure:"clustername"`
	KubeAPISSLPort          int                 `json:"kube_api_ssl_port" yaml:"kube_api_ssl_port" mapstructure:"kube_api_ssl_port"`
	DisableMasterHA         bool                `json:"disable_master_ha" yaml:"disable_master_ha" mapstructure:"disable_master_ha"`
	KubeVirtualIPShortname  string              `json:"kube_virtual_ip_shortname" yaml:"kube_virtual_ip_shortname" mapstructure:"kube_virtual_ip_shortname"`
	KubeVirtualIPApi        string              `json:"kube_virtual_ip_api" yaml:"kube_virtual_ip_api" mapstructure:"kube_virtual_ip_api"`
	KubeVIPAPISSLPort       int                 `json:"kube_vip_api_ssl_port" yaml:"kube_vip_api_ssl_port" mapstructure:"kube_vip_api_ssl_port"`
	PublicVirtualIP         string              `json:"public_virtual_ip,omitempty" yaml:"public_virtual_ip,omitempty" mapstructure:"public_virtual_ip"`
	PublicVirtualIPSSLPort  int                 `json:"public_virtual_ip_ssl_port,omitempty" yaml:"public_virtual_ip_ssl_port,omitempty" mapstructure:"public_virtual_ip_ssl_port"`
	PublicAPIServerDNSName  string              `json:"public_apiserver_dns_name" yaml:"public_apiserver_dns_name" mapstructure:"public_apiserver_dns_name"`
	PrivateAPIServerDNSName string              `json:"private_apiserver_dns_name" yaml:"private_apiserver_dns_name" mapstructure:"private_apiserver_dns_name"`
	Username                string              `json:"username" yaml:"username" mapstructure:"name"`
	VsphereUsername         string              `json:"-" yaml:"-" mapstructure:"-"`
	VspherePassword         string              `json:"-" yaml:"-" mapstructure:"-"`
	VsphereServer           string              `json:"-" yaml:"-" mapstructure:"vsphere_server"`
	Datacenter              string              `json:"datacenter" yaml:"datacenter" mapstructure:"datacenter"`
	Datastore               string              `json:"datastore" yaml:"datastore" mapstructure:"datastore"`
	ResourcePool            string              `json:"resource_pool" yaml:"resource_pool" mapstructure:"resource_pool"`
	VsphereNet              string              `json:"vsphere_net" yaml:"vsphere_net" mapstructure:"vsphere_net"`
	Folder                  string              `json:"folder" yaml:"folder" mapstructure:"folder"`
	Domain                  string              `json:"domain" yaml:"domain" mapstructure:"domain"`
	DNSServers              []string            `json:"dns_servers" yaml:"dns_servers" mapstructure:"dns_servers"`
	DNSSearch               []string            `json:"dns_search" yaml:"dns_search" mapstructure:"dns_search"`
	TimeServers             []string            `json:"time_servers" yaml:"time_servers" mapstructure:"time_servers"`
	PrivateKey              string              `json:"private_key,omitempty" yaml:"private_key,omitempty" mapstructure:"private_key"`
	PrivateKeyFile          string              `json:"private_key_file" yaml:"private_key_file" mapstructure:"private_key_file"`
	PublicKey               string              `json:"public_key,omitempty" yaml:"public_key,omitempty" mapstructure:"public_key"`
	PublicKeyFile           string              `json:"public_key_file" yaml:"public_key_file" mapstructure:"public_key_file"`
	DefaultNodePool         NodePool            `json:"default_node_pool" yaml:"default_node_pool" mapstructure:"default_node_pool"`
	NodePools               map[string]NodePool `json:"node_pools" yaml:"node_pools" mapstructure:"node_pools"`
}

// Address defines a static IP and an optional predefined hostname for an instance to be used in the node pool
type Address struct {
	IP       string `json:"ip" yaml:"ip" mapstructure:"ip"`
	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty" mapstructure:"hostname"`
}

// NodePool defines the settings for group of instances on vSphere
type NodePool struct {
	Name              string    `json:"-" yaml:"-" mapstructure:"name"`
	Count             int       `json:"count" yaml:"count" mapstructure:"count"`
	TemplateName      string    `json:"template_name,omitempty" yaml:"template_name,omitempty" mapstructure:"template_name"`
	CPUs              int       `json:"cpus,omitempty" yaml:"cpus,omitempty" mapstructure:"cpus"`
	Memory            int       `json:"memory,omitempty" yaml:"memory,omitempty" mapstructure:"memory"`
	RootVolSize       int       `json:"root_vol_size,omitempty" yaml:"root_vol_size,omitempty" mapstructure:"root_vol_size"`
	LinkedClone       bool      `json:"linked_clone,omitempty" yaml:"linked_clone,omitempty" mapstructure:"linked_clone"`
	KubeletNodeLabels []string  `json:"kubelet_node_labels,omitempty" yaml:"kubelet_node_labels,omitempty" mapstructure:"kubelet_node_labels"`
	KubeletNodeTaints []string  `json:"kubelet_node_taints,omitempty" yaml:"kubelet_node_taints,omitempty" mapstructure:"kubelet_node_taints"`
	AddressPool       []Address `json:"address_pool,omitempty" yaml:"address_pool,omitempty" mapstructure:"address_pool"`
	IPNetmask         *int      `json:"ip_netmask,omitempty" yaml:"ip_netmask,omitempty" mapstructure:"ip_netmask"`
	IPGateway         string    `json:"ip_gateway,omitempty" yaml:"ip_gateway,omitempty" mapstructure:"ip_gateway"`
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

// NewConfigFrom returns a new vSphere  configuration from a map, usually from a
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
		case "vsphere_password", "vsphere_username", "vsphere_server":
			panic(fmt.Errorf("field %s is obsolete, please remove it from cluster.yaml and use kubekit login for credentials", name))
		default:
			config.SetField(c, name, v)
		}
	}
}

func getAddress(m map[interface{}]interface{}) Address {
	addr := Address{}
	for k, v := range m {
		config.SetField(&addr, k.(string), v)
	}
	return addr
}

func getAddressPool(l []interface{}) []Address {
	var addrPool []Address
	for _, v := range l {
		mapVal := v.(map[interface{}]interface{})
		addrPool = append(addrPool, getAddress(mapVal))
	}
	return addrPool
}

func getNodePool(m map[interface{}]interface{}) NodePool {
	n := NodePool{}
	for k, v := range m {
		name := k.(string)
		switch name {
		case "kubelet_node_labels":
			n.KubeletNodeLabels = config.GetListFromInterface(v)
		case "kubelet_node_taints":
			n.KubeletNodeTaints = config.GetListFromInterface(v)
		case "address_pool":
			listVal := v.([]interface{})
			n.AddressPool = getAddressPool(listVal)
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
