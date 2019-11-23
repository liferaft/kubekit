package aks

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/johandry/merger"
	"github.com/liferaft/kubekit/pkg/provisioner/config"
	"github.com/liferaft/kubekit/pkg/provisioner/utils"
)

const (
	loginComment  = "# fill in if a different from the login"
	requiredValue = "# Required value. Example: "
	findByTagKey  = "yaml"

	nodePoolTypeVMSS = "VirtualMachineScaleSets"
	nodePoolTypeAS   = "AvailabilitySet"
)

// defaultConfig is the default configuration for AWS platform
var defaultConfig = Config{
	Environment:           "public",
	ResourceGroupLocation: requiredValue + "West US",
	//PreviewFeatures:       defaultPreviewFeatures,
	VnetAddressSpace:    "10.240.0.0/16",
	SubnetAddressPrefix: "10.240.0.0/20",
	PrivateDNSZoneName:  "",
	ServiceCIDR:         "172.21.0.0/16",
	DockerBridgeCIDR:    "172.17.0.1/16",
	DNSServiceIP:        "172.21.0.10",
	NetworkPolicy:       "calico",
	AdminUsername:       "kubekit",
	//ContainerRegistrySku:          "Basic",
	//ContainerRegistryAdminEnabled: false,
	DefaultNodePool: defaultNodePool,
	NodePools: map[string]NodePool{
		"fastcompute": fastComputeNodePool,
		"slowcompute": slowComputeNodePool,
	},
}

var defaultPreviewFeatures = []PreviewFeature{
	// Pod Security Policy
	{
		Namespace: "Microsoft.ContainerService",
		Name:      "PodSecurityPolicyPreview",
	},
	// Azure Policy (OpenPolicyAgent)
	{
		Namespace: "Microsoft.ContainerService",
		Name:      "AKS-AzurePolicyAutoApprove",
	},
	{
		Namespace: "Microsoft.PolicyInsights",
		Name:      "AKS-DataplaneAutoApprove",
	},
}

var defaultNodePool = NodePool{
	VMSize:         "Standard_F8s_v2",
	RootVolumeSize: 100,
	MaxPods:        30, // matches the default advanced networking max pod size per node
	DockerRoot:     "/mnt",
	Type:           nodePoolTypeVMSS,
}

var fastComputeNodePool = NodePool{
	Name:           "fastcompute",
	Count:          3,
	VMSize:         "Standard_F8s_v2",
	RootVolumeSize: 100,
	MaxPods:        30, // matches the default advanced networking max pod size per node
	DockerRoot:     "/mnt",
	Type:           nodePoolTypeVMSS,
}

var slowComputeNodePool = NodePool{
	Name:           "slowcompute",
	Count:          3,
	VMSize:         "Standard_F8s_v2",
	RootVolumeSize: 100,
	MaxPods:        30, // matches the default advanced networking max pod size per node
	DockerRoot:     "/mnt",
	Type:           nodePoolTypeVMSS,
}

// Config defines the Azure AKS configuration parameters in the Cluster config file
type Config struct {
	ClusterName                   string              `json:"-" yaml:"-" mapstructure:"-"`
	PrivateKey                    string              `json:"private_key,omitempty" yaml:"private_key,omitempty" mapstructure:"private_key"`
	PrivateKeyFile                string              `json:"private_key_file" yaml:"private_key_file" mapstructure:"private_key_file"`
	PublicKey                     string              `json:"public_key,omitempty" yaml:"public_key,omitempty" mapstructure:"public_key"`
	PublicKeyFile                 string              `json:"public_key_file" yaml:"public_key_file" mapstructure:"public_key_file"`
	SubscriptionID                string              `json:"subscription_id,omitempty" yaml:"subscription_id,omitempty" mapstructure:"subscription_id"`
	TenantID                      string              `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty" mapstructure:"tenant_id"`
	ClientID                      string              `json:"client_id,omitempty" yaml:"client_id,omitempty" mapstructure:"client_id"`
	ClientSecret                  string              `json:"client_secret,omitempty" yaml:"client_secret,omitempty" mapstructure:"client_secret"`
	Environment                   string              `json:"environment" yaml:"environment" mapstructure:"environment"`
	ResourceGroupLocation         string              `json:"resource_group_location" yaml:"resource_group_location" mapstructure:"resource_group_location"`
	DNSPrefix                     string              `json:"dns_prefix,omitempty" yaml:"dns_prefix,omitempty" mapstructure:"dns_prefix"`
	ContainerRegistryHost         Registry            `json:"container_registry,omitempty" yaml:"container_registry,omitempty" mapstructure:"container_registry,omitempty"`
	ContainerRegistrySku          string              `json:"container_registry_sku,omitempty" yaml:"container_registry_sku,omitempty" mapstructure:"container_registry_sku"`
	ContainerRegistryAdminEnabled bool                `json:"container_registry_admin_enabled,omitempty" yaml:"container_registry_admin_enabled,omitempty" mapstructure:"container_registry_admin_enabled"`
	PreviewFeatures               []PreviewFeature    `json:"preview_features,omitempty" yaml:"preview_features,omitempty" mapstructure:"preview_features,omitempty"`
	VnetName                      string              `json:"vnet_name" yaml:"vnet_name" mapstructure:"vnet_name"`
	VnetResourceGroupName         string              `json:"vnet_resource_group_name" yaml:"vnet_resource_group_name" mapstructure:"vnet_resource_group_name"`
	VnetAddressSpace              string              `json:"vnet_address_space" yaml:"vnet_address_space" mapstructure:"vnet_address_space"`
	SubnetAddressPrefix           string              `json:"subnet_address_prefix" yaml:"subnet_address_prefix" mapstructure:"subnet_address_prefix"`
	PrivateDNSZoneName            string              `json:"private_dns_zone_name" yaml:"private_dns_zone_name" mapstructure:"private_dns_zone_name"`
	ServiceCIDR                   string              `json:"service_cidr" yaml:"service_cidr" mapstructure:"service_cidr"`
	DockerBridgeCIDR              string              `json:"docker_bridge_cidr" yaml:"docker_bridge_cidr" mapstructure:"docker_bridge_cidr"`
	DNSServiceIP                  string              `json:"dns_service_ip" yaml:"dns_service_ip" mapstructure:"dns_service_ip"`
	KubernetesVersion             string              `json:"kubernetes_version" yaml:"kubernetes_version" mapstructure:"kubernetes_version"`
	AdminUsername                 string              `json:"admin_username" yaml:"admin_username" mapstructure:"admin_username"`
	NetworkPolicy                 string              `json:"network_policy" yaml:"network_policy" mapstructure:"network_policy"`
	EnablePodSecurityPolicy       bool                `json:"enable_pod_security_policy" yaml:"enable_pod_security_policy" mapstructure:"enable_pod_security_policy"`
	ClusterClientID               string              `json:"cluster_client_id" yaml:"cluster_client_id" mapstructure:"cluster_client_id"`
	ClusterClientSecret           string              `json:"cluster_client_secret" yaml:"cluster_client_secret" mapstructure:"cluster_client_secret"`
	DefaultNodePool               NodePool            `json:"default_node_pool" yaml:"default_node_pool" mapstructure:"default_node_pool"`
	NodePools                     map[string]NodePool `json:"node_pools" yaml:"node_pools" mapstructure:"node_pools"`
	Jumpbox                       *Jumpbox            `json:"jumpbox,omitempty" yaml:"jumpbox,omitempty" mapstructure:"jumpbox"`
}

// NSGRule defines the network security group rule to be added
type NSGRule struct {
	Name                     string `json:"name" yaml:"name" mapstructure:"name"`
	Priority                 int    `json:"priority" yaml:"priority" mapstructure:"priority"`
	Direction                string `json:"direction" yaml:"direction" mapstructure:"direction"`
	Access                   string `json:"access" yaml:"access" mapstructure:"access"`
	Protocol                 string `json:"protocol" yaml:"protocol" mapstructure:"protocol"`
	SourcePortRange          string `json:"source_port_range" yaml:"source_port_range" mapstructure:"source_port_range"`
	DestinationPortRange     string `json:"destination_port_range" yaml:"destination_port_range" mapstructure:"destination_port_range"`
	SourceAddressPrefix      string `json:"source_address_prefix" yaml:"source_address_prefix" mapstructure:"source_address_prefix"`
	DestinationAddressPrefix string `json:"destination_address_prefix" yaml:"destination_address_prefix" mapstructure:"destination_address_prefix"`
}

// Jumpbox contains information needed when creating a jumpbox to the cluster
type Jumpbox struct {
	AdminUsername   string    `json:"admin_username" yaml:"admin_username" mapstructure:"admin_username"`
	PrivateKey      string    `json:"-" yaml:"-" mapstructure:"-"`
	PrivateKeyFile  string    `json:"private_key_file" yaml:"private_key_file" mapstructure:"private_key_file"`
	PublicKey       string    `json:"public_key,omitempty" yaml:"public_key,omitempty" mapstructure:"public_key"`
	PublicKeyFile   string    `json:"public_key_file" yaml:"public_key_file" mapstructure:"public_key_file"`
	VMSize          string    `json:"vm_size" yaml:"vm_size" mapstructure:"vm_size"`
	RootVolumeSize  int       `json:"root_volume_size" yaml:"root_volume_size" mapstructure:"root_volume_size"`
	EnablePublicIP  bool      `json:"enable_public_ip" yaml:"enable_public_ip" mapstructure:"enable_public_ip"`
	TimeoutMinutes  int       `json:"timeout_minutes" yaml:"timeout_minutes" mapstructure:"timeout_minutes"`
	UploadKubeconfg bool      `json:"upload_kubeconfig" yaml:"upload_kubeconfig" mapstructure:"upload_kubeconfig"`
	NSGRules        []NSGRule `json:"network_security_group_rules,omitempty" yaml:"network_security_group_rules,omitempty" mapstructure:"network_security_group_rules"`
	Commands        []string  `json:"commands,omitempty" yaml:"commands,omitempty" mapstructure:"commands"`
	FileUploads     []string  `json:"file_uploads,omitempty" yaml:"file_uploads,omitempty" mapstructure:"file_uploads"`
	UbuntuServerSku string    `json:"ubuntu_server_sku" yaml:"ubuntu_server_sku" mapstructure:"ubuntu_server_sku"`
}

// PreviewFeature represents the preview feature information when enabling such
type PreviewFeature struct {
	Namespace string `json:"namespace" yaml:"namespace" mapstructure:"namespace"`
	Name      string `json:"name" yaml:"name" mapstructure:"name"`
}

// Registry represents the ACR login info
type Registry struct {
	Host     string `json:"host" yaml:"host" mapstructure:"host"`
	Username string `json:"username" yaml:"username" mapstructure:"username"`
	Password string `json:"password" yaml:"password" mapstructure:"password"`
}

// DataDisk defines the disk information
type DataDisk struct {
	MountPoint string `json:"mount_point" yaml:"mount_point" mapstructure:"mount_point"`
	VolumeSize int    `json:"volume_size" yaml:"volume_size" mapstructure:"volume_size"`
}

// NodePool defines the settings for group of instances on Azure AKS
type NodePool struct {
	Name                string      `json:"-" yaml:"-" mapstructure:"-"`
	Count               int         `json:"count,omitempty" yaml:"count,omitempty" mapstructure:"count"`
	VMSize              string      `json:"vm_size" yaml:"vm_size" mapstructure:"vm_size"`
	RootVolumeSize      int         `json:"root_volume_size" yaml:"root_volume_size" mapstructure:"root_volume_size"`
	MaxPods             int         `json:"max_pods" yaml:"max_pods" mapstructure:"max_pods"`
	Type                string      `json:"type" yaml:"type" mapstructure:"type"`
	EphemeralMountPoint string      `json:"ephemeral_mount_point,omitempty" yaml:"ephemeral_mount_point,omitempty" mapstructure:"ephemeral_mount_point,omitempty"`
	DockerRoot          string      `json:"docker_root,omitempty" yaml:"docker_root,omitempty" mapstructure:"docker_root"`
	DataDisks           *[]DataDisk `json:"data_disks,omitempty" yaml:"data_disks,omitempty" mapstructure:"data_disks,omitempty"`

	AvailabilityZones   []string `json:"availability_zones,omitempty" yaml:"availability_zones,omitempty" mapstructure:"availability_zones,omitempty"`
	NodeTaints          []string `json:"node_taints,omitempty" yaml:"node_taints,omitempty" mapstructure:"node_taints"`
	EnableAutoScaling   bool     `json:"enable_auto_scaling,omitempty" yaml:"enable_auto_scaling,omitempty" mapstructure:"enable_auto_scaling"`
	AutoScalingMinCount int      `json:"auto_scaling_min_count,omitempty" yaml:"auto_scaling_min_count,omitempty" mapstructure:"auto_scaling_min_count"`
	AutoScalingMaxCount int      `json:"auto_scaling_max_count,omitempty" yaml:"auto_scaling_max_count,omitempty" mapstructure:"auto_scaling_max_count"`
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

// NewConfigFrom returns a new AKS configuration from a map, usually from a
// cluster config file
func NewConfigFrom(m map[interface{}]interface{}) *Config {
	c := &Config{}
	c.MergeWithMapConfig(m)

	if c.ClusterClientSecret == "" {
		c.ClusterClientID = c.ClientID
	}
	if c.ClusterClientSecret == "" {
		c.ClusterClientSecret = c.ClientSecret
	}

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

func getRegistryHost(m map[interface{}]interface{}) Registry {
	registry := Registry{}
	for k, v := range m {
		name := k.(string)
		config.SetField(&registry, name, v)
	}
	return registry
}

// MergeWithMapConfig merges this configuration with the given configuration in
// a map[string], usually from a cluster config file
func (c *Config) MergeWithMapConfig(m map[interface{}]interface{}) {
	for k, v := range m {
		// vType := reflect.ValueOf(v).Type().Kind().String()
		name := k.(string)
		switch name {
		case "container_registry":
			m1 := v.(map[interface{}]interface{})
			c.ContainerRegistryHost = getRegistryHost(m1)
		case "default_node_pool":
			m1 := v.(map[interface{}]interface{})
			c.DefaultNodePool = getNodePool("", m1)
		case "node_pools":
			m1 := v.(map[interface{}]interface{})
			c.NodePools = getNodePools(m1)
		case "preview_features":
			l1 := v.([]interface{})
			c.PreviewFeatures = getPreviewFeatures(l1)
		case "jumpbox":
			m1 := v.(map[interface{}]interface{})
			c.Jumpbox = getJumpbox(m1)
		default:
			config.SetField(c, name, v)
		}
	}
}

func getDataDisk(m interface{}) DataDisk {
	d := DataDisk{}
	for k, v := range m.(map[interface{}]interface{}) {
		name := k.(string)
		setField(&d, name, v)
	}
	return d
}

func getDataDisks(l []interface{}) *[]DataDisk {
	var dataDisks []DataDisk
	for _, i := range l {
		dataDisks = append(dataDisks, getDataDisk(i))
	}
	return &dataDisks
}

//TODO These functions: getNodePool and getNodePools can be made to
//use dynamic structure return.  Don't have the time to do it.
//As every platform uses these 2 functions

func getNodePool(name string, m map[interface{}]interface{}) NodePool {
	n := NodePool{
		Name: name,
	}
	for k, v := range m {
		name := k.(string)
		switch name {
		case "data_disks":
			l1 := v.([]interface{})
			n.DataDisks = getDataDisks(l1)
		case "availability_zones":
			l1 := v.([]interface{})
			n.AvailabilityZones = getListOfStrings(l1)
		case "node_taints":
			l1 := v.([]interface{})
			n.NodeTaints = getListOfStrings(l1)
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
		nPools[k.(string)] = getNodePool(k.(string), m1)
	}
	return nPools
}

func getPreviewFeature(m interface{}) PreviewFeature {
	previewFeature := PreviewFeature{}
	for k, v := range m.(map[interface{}]interface{}) {
		name := k.(string)
		setField(&previewFeature, name, v)
	}
	return previewFeature
}

func getPreviewFeatures(l []interface{}) []PreviewFeature {
	var previewFeatures []PreviewFeature
	for _, i := range l {
		previewFeatures = append(previewFeatures, getPreviewFeature(i))
	}
	return previewFeatures
}

func getListOfStrings(l []interface{}) []string {
	var strList []string
	for _, i := range l {
		strList = append(strList, fmt.Sprintf("%v", i))
	}
	return strList
}

func getNSGRule(m interface{}) NSGRule {
	nsgRule := NSGRule{}
	for k, v := range m.(map[interface{}]interface{}) {
		name := k.(string)
		setField(&nsgRule, name, v)
	}
	return nsgRule
}

func getNSGRules(l []interface{}) []NSGRule {
	var nsgRules []NSGRule
	for _, i := range l {
		nsgRules = append(nsgRules, getNSGRule(i))
	}
	return nsgRules
}

func getJumpbox(m map[interface{}]interface{}) *Jumpbox {
	j := Jumpbox{}
	for k, v := range m {
		name := k.(string)
		switch name {
		case "network_security_group_rules":
			l1 := v.([]interface{})
			j.NSGRules = getNSGRules(l1)
		case "commands":
			l1 := v.([]interface{})
			j.Commands = getListOfStrings(l1)
		case "file_uploads":
			l1 := v.([]interface{})
			j.FileUploads = getListOfStrings(l1)
		default:
			setField(&j, name, v)
		}
	}
	if len(j.PublicKey) == 0 && len(j.PublicKeyFile) > 0 {
		keyData, err := ioutil.ReadFile(j.PublicKeyFile)
		if err == nil {
			j.PublicKey = string(keyData)
		}
	}
	return &j
}

func (c *Config) copyWithDefaults() Config {
	config := *c
	marshalled, _ := json.Marshal(c.DefaultNodePool)
	nodePools := make(map[string]NodePool, len(config.NodePools))
	for k, v := range config.NodePools {
		n := NodePool{}
		json.Unmarshal(marshalled, &n)

		a, _ := json.Marshal(v)
		json.Unmarshal(a, &n)

		n.Name = k
		nodePools[k] = n
	}
	config.NodePools = nodePools
	return config
}

func setField(c interface{}, name string, value interface{}) {
	sValue := reflect.ValueOf(c).Elem()
	sFieldValue := sValue.FieldByName(name)
	if !sFieldValue.IsValid() {
		sFieldValue = fieldByTag(sValue, name)
	}
	if !sFieldValue.IsValid() {
		panic(fmt.Errorf("field %q not found", name))
	}
	if !sFieldValue.CanSet() {
		panic(fmt.Errorf("cannot set value to field %q", name))
	}

	sFieldType := sFieldValue.Type()
	if sFieldType.Kind() == reflect.String && value == nil {
		sFieldValue.SetString("")
		return
	}

	v := reflect.ValueOf(value)

	if sFieldType != v.Type() {
		panic(fmt.Errorf("value type of field %q does not match with the config field type (%s != %s)", name, sFieldType.Kind().String(), v.Type().Kind().String()))
	}

	sFieldValue.Set(v)
}

func fieldByTag(value reflect.Value, name string) reflect.Value {
	for i := 0; i < value.NumField(); i++ {
		retValue := value.Field(i)
		tag := value.Type().Field(i).Tag
		allTagValues := tag.Get(findByTagKey)
		// the tag value may have more than one value, take the first one
		tagValue := strings.Split(allTagValues, ",")[0]
		if tagValue == name {
			return retValue
		}
	}
	panic(fmt.Errorf("not found field with %s tag named %q", findByTagKey, name))
}
