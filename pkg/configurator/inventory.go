package configurator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

// ValidAddressInventoryFields is a lookup for valid InventoryHost address fields
// the key names must match the values given in the yaml tag for the InventoryHost fields used for addressing
var ValidAddressInventoryFields = map[string]struct{}{
	"ansible_host":   struct{}{},
	"private_ip":     struct{}{},
	"public_ip":      struct{}{},
	"private_dns":    struct{}{},
	"public_dns":     struct{}{},
	"fqdn":           struct{}{},
	"hostname":       struct{}{},
	"hostname_short": struct{}{},
	"kubelet_taints": struct{}{},
	"kubelet_labels": struct{}{},
}

// ZeroPadLen is the length for the number following the role of the node. i.e. master001
const ZeroPadLen = 3

// Children is a children structure in an Ansible inventory
type Children struct {
	KubeCluster KubeCluster `json:"kube_cluster,omitempty" yaml:"kube_cluster" mapstructure:"kube_cluster"`
}

// KubeCluster is a kube_cluster structure in an Ansible inventory
type KubeCluster struct {
	Children map[string]InventoryHosts `json:"children,omitempty" yaml:"children" mapstructure:"children"`
}

// InventoryHosts is a hosts structure in an Ansible inventory
type InventoryHosts struct {
	Hosts map[string]*InventoryHost `json:"hosts,omitempty" yaml:"hosts" mapstructure:"hosts"`
}

// InventoryHost is a host structure in an Ansible inventory
// Note: update var ValidAddressInventoryFields if a new field is added
type InventoryHost struct {
	AnsibleHost   string   `json:"ansible_host,omitempty" yaml:"ansible_host" mapstructure:"ansible_host"`
	PrivateIP     string   `json:"private_ip,omitempty" yaml:"private_ip" mapstructure:"private_ip"`
	PublicIP      string   `json:"public_ip,omitempty" yaml:"public_ip" mapstructure:"public_ip"`
	PrivateDNS    string   `json:"private_dns,omitempty" yaml:"private_dns" mapstructure:"private_dns"`
	PublicDNS     string   `json:"public_dns,omitempty" yaml:"public_dns" mapstructure:"public_dns"`
	FQDN          string   `json:"fqdn,omitempty" yaml:"fqdn" mapstructure:"fqdn"`
	Hostname      string   `json:"hostname,omitempty" yaml:"hostname" mapstructure:"hostname"`
	HostnameShort string   `json:"hostname_short,omitempty" yaml:"hostname_short" mapstructure:"hostname_short"`
	KubeletTaints []string `json:"kubelet_taints,omitempty" yaml:"kubelet_taints" mapstructure:"kubelet_taints"`
	KubeletLabels []string `json:"kubelet_labels,omitempty" yaml:"kubelet_labels" mapstructure:"kubelet_labels"`
}

// InventoryVariables contain all the variables required for the Ansible inventory
type InventoryVariables struct {
	ShellEditingMode                        string      `json:"shell_editing_mode,omitempty" yaml:"shell_editing_mode,omitempty" mapstructure:"shell_editing_mode,omitempty"`
	AddressInventoryField                   string      `json:"address_inventory_field" yaml:"address_inventory_field" mapstructure:"address_inventory_field"`
	AlbDNSName                              string      `json:"alb_dns_name,omitempty" yaml:"alb_dns_name" mapstructure:"alb_dns_name"`                   // Just for AWS.     From provisioner
	VsphereDatacenter                       string      `json:"vsphere_datacenter,omitempty" yaml:"vsphere_datacenter" mapstructure:"vsphere_datacenter"` // Just for vSphere. From provisioner
	VsphereFolder                           string      `json:"vsphere_folder,omitempty" yaml:"vsphere_folder" mapstructure:"vsphere_folder"`             // Just for vSphere. From provisioner
	VsphereDatastore                        string      `json:"vsphere_datastore,omitempty" yaml:"vsphere_datastore" mapstructure:"vsphere_datastore"`    // Just for vSphere. From provisioner
	VsphereNet                              string      `json:"vsphere_net" yaml:"vsphere_net" mapstructure:"vsphere_net"`                                // Just for vSphere. From provisioner
	VsphereUsername                         string      `json:"vsphere_username,omitempty" yaml:"vsphere_username" mapstructure:"vsphere_username"`       // Just for vSphere. From environment
	VspherePassword                         string      `json:"vsphere_password,omitempty" yaml:"vsphere_password" mapstructure:"vsphere_password"`       // Just for vSphere. From environment
	VsphereServer                           string      `json:"vsphere_server,omitempty" yaml:"vsphere_server" mapstructure:"vsphere_server"`             // Just for vSphere. From environment
	VsphereResourcePool                     string      `json:"vsphere_resource_pool,omitempty" yaml:"vsphere_resource_pool,omitempty" mapstructure:"vsphere_resource_pool"`
	ClusterName                             string      `json:"cluster_name,omitempty" yaml:"cluster_name" mapstructure:"cluster_name"`                               // From top in config file.
	AnsibleUser                             string      `json:"ansible_user,omitempty" yaml:"ansible_user" mapstructure:"ansible_user"`                               // Calculated, depend of platform
	CloudProvider                           string      `json:"cloud_provider,omitempty" yaml:"cloud_provider" mapstructure:"cloud_provider"`                         // Calculated, depend of platform
	CloudProviderEnabled                    bool        `json:"cloud_provider_enabled,omitempty" yaml:"cloud_provider_enabled" mapstructure:"cloud_provider_enabled"` // Calculated, depend of platform
	KubeAPISslPort                          int         `json:"kube_api_ssl_port,omitempty" yaml:"kube_api_ssl_port" mapstructure:"kube_api_ssl_port"`                // From provisioner
	DisableMasterHA                         bool        `json:"disable_master_ha,omitempty" yaml:"disable_master_ha" mapstructure:"disable_master_ha"`                // From provisioner
	KubeVirtualIPApi                        string      `json:"kube_virtual_ip_api,omitempty" yaml:"kube_virtual_ip_api" mapstructure:"kube_virtual_ip_api"`          // From provisioner
	KubeVipAPISslPort                       int         `json:"kube_vip_api_ssl_port,omitempty" yaml:"kube_vip_api_ssl_port" mapstructure:"kube_vip_api_ssl_port"`    // From provisioner
	PublicVirtualIP                         string      `json:"public_virtual_ip,omitempty" yaml:"public_virtual_ip" mapstructure:"public_virtual_ip"`
	PublicVirtualIPSslPort                  int         `json:"public_virtual_ip_ssl_port,omitempty" yaml:"public_virtual_ip_ssl_port" mapstructure:"public_virtual_ip_ssl_port"`
	EtcdLocalProxyEnabled                   bool        `json:"enable_etcd_local_proxy" yaml:"enable_etcd_local_proxy" mapstructure:"enable_etcd_local_proxy"`
	EtcdInitialClusterToken                 string      `json:"etcd_initial_cluster_token,omitempty" yaml:"etcd_initial_cluster_token" mapstructure:"etcd_initial_cluster_token"` // From configurator in config file. This and the following fields
	KubeletMaxPods                          int         `json:"kubelet_max_pods,omitempty" yaml:"kubelet_max_pods" mapstructure:"kubelet_max_pods"`
	KubeletSerializeImagePulls              bool        `json:"kubelet_serialize_image_pulls,omitempty" yaml:"kubelet_serialize_image_pulls,omitempty" mapstructure:"kubelet_serialize_image_pulls,omitempty"`
	KubeProxyMode                           string      `json:"kube_proxy_mode,omitempty" yaml:"kube_proxy_mode,omitempty" mapstructure:"kube_proxy_mode,omitempty"`
	KubeClusterCidr                         string      `json:"kube_cluster_cidr,omitempty" yaml:"kube_cluster_cidr" mapstructure:"kube_cluster_cidr"`
	KubeServicesCidr                        string      `json:"kube_services_cidr,omitempty" yaml:"kube_services_cidr" mapstructure:"kube_services_cidr"`
	KubeServiceIP                           string      `json:"kube_service_ip,omitempty" yaml:"kube_service_ip" mapstructure:"kube_service_ip"`
	KubeAdvertiseAddress                    string      `json:"kube_advertise_address,omitempty" yaml:"kube_advertise_address" mapstructure:"kube_advertise_address"`
	MasterSchedulableEnabled                bool        `json:"master_schedulable_enabled,omitempty" yaml:"master_schedulable_enabled" mapstructure:"master_schedulable_enabled"`
	DockerMaxConcurrentUploads              int         `json:"docker_max_concurrent_uploads,omitempty" yaml:"docker_max_concurrent_uploads,omitempty" mapstructure:"docker_max_concurrent_uploads,omitempty"`
	DockerMaxConcurrentDownloads            int         `json:"docker_max_concurrent_downloads,omitempty" yaml:"docker_max_concurrent_downloads,omitempty" mapstructure:"docker_max_concurrent_downloads,omitempty"`
	DockerLogMaxFiles                       string      `json:"docker_log_max_files,omitempty" yaml:"docker_log_max_files,omitempty" mapstructure:"docker_log_max_files"` // docker config takes it as a string
	DockerLogMaxSize                        string      `json:"docker_log_max_size,omitempty" yaml:"docker_log_max_size,omitempty" mapstructure:"docker_log_max_size"`
	DockerRegistryPath                      string      `json:"docker_registry_path,omitempty" yaml:"docker_registry_path" mapstructure:"docker_registry_path"`
	DockerRegistryPort                      int         `json:"registry_port" yaml:"registry_port" mapstructure:"registry_port"`
	DownloadImagesIfMissing                 bool        `json:"download_images_if_missing,omitempty" yaml:"download_images_if_missing" mapstructure:"download_images_if_missing"`
	HAProxyClientTimeout                    string      `json:"haproxy_client_timeout,omitempty" yaml:"haproxy_client_timeout,omitempty" mapstructure:"haproxy_client_timeout"`
	HAProxyServerTimeout                    string      `json:"haproxy_server_timeout,omitempty" yaml:"haproxy_server_timeout,omitempty" mapstructure:"haproxy_server_timeout"`
	DNSAAAADelayEnabled                     bool        `json:"dns_aaaa_delay_enabled,omitempty" yaml:"dns_aaaa_delay_enabled,omitempty" mapstructure:"dns_aaaa_delay_enabled"`
	DNSArgs                                 string      `json:"dns_args,omitempty" yaml:"dns_args" mapstructure:"dns_args"`
	DNSServers                              []string    `json:"dns_servers,omitempty" yaml:"dns_servers" mapstructure:"dns_servers"`
	DNSSearch                               []string    `json:"dns_search,omitempty" yaml:"dns_search" mapstructure:"dns_search"`
	EtcdDataDirectory                       string      `json:"etcd_data_directory,omitempty" yaml:"etcd_data_directory,omitempty" mapstructure:"etcd_data_directory"`
	EtcdDefragCrontabHour                   string      `json:"etcd_defrag_crontab_hour,omitempty" yaml:"etcd_defrag_crontab_hour,omitempty" mapstructure:"etcd_defrag_crontab_hour"`
	EtcdLogsCrontabHour                     string      `json:"etcd_logs_crontab_hour,omitempty" yaml:"etcd_logs_crontab_hour" mapstructure:"etcd_logs_crontab_hour"`
	EtcdLogsCrontabMinute                   string      `json:"etcd_logs_crontab_minute,omitempty" yaml:"etcd_logs_crontab_minute" mapstructure:"etcd_logs_crontab_minute"`
	EtcdLogsDaysToKeep                      int         `json:"etcd_logs_days_to_keep,omitempty" yaml:"etcd_logs_days_to_keep" mapstructure:"etcd_logs_days_to_keep"`
	EtcdSnapshotsDirectory                  string      `json:"etcd_snapshots_directory,omitempty" yaml:"etcd_snapshots_directory,omitempty" mapstructure:"etcd_snapshots_directory"`
	EtcdQuotaBackendBytes                   int         `json:"etcd_quota_backend_bytes,omitempty" yaml:"etcd_quota_backend_bytes,omitempty" mapstructure:"etcd_quota_backend_bytes"`
	UseLocalImages                          bool        `json:"use_local_images,omitempty" yaml:"use_local_images" mapstructure:"use_local_images"`
	ClusterIfaceName                        string      `json:"cluster_iface_name,omitempty" yaml:"cluster_iface_name" mapstructure:"cluster_iface_name"`
	ClusterIface                            string      `json:"cluster_iface,omitempty" yaml:"cluster_iface" mapstructure:"cluster_iface"`
	CniIface                                string      `json:"cni_iface,omitempty" yaml:"cni_iface" mapstructure:"cni_iface"`
	PublicVipIfaceName                      string      `json:"public_vip_iface_name,omitempty" yaml:"public_vip_iface_name" mapstructure:"public_vip_iface_name"`
	PublicVipIface                          string      `json:"public_vip_iface,omitempty" yaml:"public_vip_iface" mapstructure:"public_vip_iface"`
	CniIPEncapsulation                      string      `json:"cni_ip_encapsulation,omitempty" yaml:"cni_ip_encapsulation" mapstructure:"cni_ip_encapsulation"`
	TimeServers                             []string    `json:"time_servers,omitempty" yaml:"time_servers" mapstructure:"time_servers"`
	KubeAuditLogMaxAge                      int         `json:"kube_audit_log_max_age,omitempty" yaml:"kube_audit_log_max_age" mapstructure:"kube_audit_log_max_age"`
	KubeAuditLogMaxBackup                   int         `json:"kube_audit_log_max_backup,omitempty" yaml:"kube_audit_log_max_backup" mapstructure:"kube_audit_log_max_backup"`
	KubeAuditLogMaxSize                     int         `json:"kube_audit_log_max_size,omitempty" yaml:"kube_audit_log_max_size" mapstructure:"kube_audit_log_max_size"`
	NginxIngressEnabled                     bool        `json:"nginx_ingress_enabled,omitempty" yaml:"nginx_ingress_enabled" mapstructure:"nginx_ingress_enabled"`
	NginxIngressControllerProxyBodySize     string      `json:"nginx_ingress_controller_proxy_body_size,omitempty" yaml:"nginx_ingress_controller_proxy_body_size" mapstructure:"nginx_ingress_controller_proxy_body_size"`
	NginxIngressControllerErrorLogLevel     string      `json:"nginx_ingress_controller_error_log_level,omitempty" yaml:"nginx_ingress_controller_error_log_level" mapstructure:"nginx_ingress_controller_error_log_level"`
	NginxIngressControllerSslProtocols      string      `json:"nginx_ingress_controller_ssl_protocols,omitempty" yaml:"nginx_ingress_controller_ssl_protocols" mapstructure:"nginx_ingress_controller_ssl_protocols"`
	NginxIngressControllerProxyReadTimeout  string      `json:"nginx_ingress_controller_proxy_read_timeout,omitempty" yaml:"nginx_ingress_controller_proxy_read_timeout" mapstructure:"nginx_ingress_controller_proxy_read_timeout"`
	NginxIngressControllerProxySendTimeout  string      `json:"nginx_ingress_controller_proxy_send_timeout,omitempty" yaml:"nginx_ingress_controller_proxy_send_timeout" mapstructure:"nginx_ingress_controller_proxy_send_timeout"`
	NginxIngressControllerTLSCertLocalPath  string      `json:"nginx_ingress_controller_tls_cert_local_path,omitempty" yaml:"nginx_ingress_controller_tls_cert_local_path" mapstructure:"nginx_ingress_controller_tls_cert_local_path"`
	NginxIngressControllerTLSKeyLocalPath   string      `json:"nginx_ingress_controller_tls_key_local_path,omitempty" yaml:"nginx_ingress_controller_tls_key_local_path" mapstructure:"nginx_ingress_controller_tls_key_local_path"`
	NginxIngressControllerBasicAuthUsername string      `json:"nginx_ingress_controller_basic_auth_username,omitempty" yaml:"nginx_ingress_controller_basic_auth_username" mapstructure:"nginx_ingress_controller_basic_auth_username"`
	NginxIngressControllerBasicAuthPassword string      `json:"nginx_ingress_controller_basic_auth_password,omitempty" yaml:"nginx_ingress_controller_basic_auth_password" mapstructure:"nginx_ingress_controller_basic_auth_password"`
	DefaultIngressHost                      string      `json:"default_ingress_host,omitempty" yaml:"default_ingress_host" mapstructure:"default_ingress_host"`
	RookEnabled                             bool        `json:"rook_enabled,omitempty" yaml:"rook_enabled" mapstructure:"rook_enabled"`
	RookCephStorageDeviceDirectories        []string    `json:"rook_ceph_storage_directories,omitempty" yaml:"rook_ceph_storage_directories" mapstructure:"rook_ceph_storage_directories"`
	RookCephStorageDeviceFilter             string      `json:"rook_ceph_storage_device_filter,omitempty" yaml:"rook_ceph_storage_device_filter" mapstructure:"rook_ceph_storage_device_filter"`
	RookDashboardEnabled                    bool        `json:"rook_dashboard_enabled,omitempty" yaml:"rook_dashboard_enabled" mapstructure:"rook_dashboard_enabled"`
	RookDashboardExternalEnabled            bool        `json:"rook_dashboard_external_enabled,omitempty" yaml:"rook_dashboard_external_enabled" mapstructure:"rook_dashboard_external_enabled"`
	RookDashboardPort                       int         `json:"rook_dashboard_port,omitempty" yaml:"rook_dashboard_port" mapstructure:"rook_dashboard_port"`
	RookObjectStoreEnabled                  bool        `json:"rook_object_store_enabled,omitempty" yaml:"rook_object_store_enabled" mapstructure:"rook_object_store_enabled"`
	RookObjectStoreRadosGatewayEnabled      bool        `json:"rook_object_store_rados_gateway_enabled,omitempty" yaml:"rook_object_store_rados_gateway_enabled" mapstructure:"rook_object_store_rados_gateway_enabled"`
	RookFileStoreEnabled                    bool        `json:"rook_file_store_enabled,omitempty" yaml:"rook_file_store_enabled" mapstructure:"rook_file_store_enabled"`
	DefaultStorageclass                     string      `json:"default_storageclass,omitempty" yaml:"default_storageclass" mapstructure:"default_storageclass"`
	HostTimeZone                            string      `json:"host_timezone,omitempty" yaml:"host_timezone,omitempty" mapstructure:"host_timezone"`
	ControlPlaneTimeZone                    string      `json:"controlplane_timezone,omitempty" yaml:"controlplane_timezone,omitempty" mapstructure:"controlplane_timezone"`
	PodEvictionTimeout                      string      `json:"pod_eviction_timeout,omitempty" yaml:"pod_eviction_timeout" mapstructure:"pod_eviction_timeout"`
	TerminatedPodGCThreshold                int         `json:"terminated_pod_gc_threshold,omitempty" yaml:"terminated_pod_gc_threshold" mapstructure:"terminated_pod_gc_threshold"`
	AdditionalRSharedMountPoints            []string    `json:"additional_rshared_mount_points,omitempty" yaml:"additional_rshared_mount_points,omitempty" mapstructure:"additional_rshared_mount_points"`
	SysctlSettings                          interface{} `json:"sysctl_settings,omitempty" yaml:"sysctl_settings,omitempty" mapstructure:"sysctl_settings"`
}

// AllInventory contain all the variables and childrens of the Ansible inventory
type AllInventory struct {
	Variables InventoryVariables `json:"vars,omitempty" yaml:"vars" mapstructure:"vars"`
	Children  Children           `json:"children,omitempty" yaml:"children" mapstructure:"children"`
}

// Inventory contain all the variables and childrens of the Ansible inventory
type Inventory struct {
	All AllInventory `json:"all,omitempty" yaml:"all" mapstructure:"all"`
}

// if a new field is added in InventoryHost then add it here as key
var validAddressInventoryFields = map[string]struct{}{
	"ansible_host":   struct{}{},
	"private_ip":     struct{}{},
	"public_ip":      struct{}{},
	"private_dns":    struct{}{},
	"public_dns":     struct{}{},
	"fqdn":           struct{}{},
	"hostname":       struct{}{},
	"hostname_short": struct{}{},
	"kubelet_taints": struct{}{},
	"kubelet_labels": struct{}{},
}

var defaultInventoryVariables = InventoryVariables{
	ShellEditingMode:                   "",
	AddressInventoryField:              "private_ip",
	EtcdLocalProxyEnabled:              false,
	EtcdInitialClusterToken:            "0c3616cc-434e",
	KubeletMaxPods:                     110,
	KubeProxyMode:                      "iptables",
	KubeletSerializeImagePulls:         false,
	KubeClusterCidr:                    "172.24.0.0/16",
	KubeServicesCidr:                   "172.21.0.0/16",
	KubeServiceIP:                      "172.21.0.1",
	KubeAdvertiseAddress:               "{{ ansible_eth0.ipv4.address }}",
	DisableMasterHA:                    true,
	KubeVirtualIPApi:                   "",
	PublicVirtualIP:                    "",
	MasterSchedulableEnabled:           false,
	DockerMaxConcurrentUploads:         10,
	DockerMaxConcurrentDownloads:       10,
	DockerLogMaxFiles:                  "5", // docker config takes it as a string
	DockerLogMaxSize:                   "16m",
	DockerRegistryPath:                 "/var/lib/docker/registry",
	DownloadImagesIfMissing:            false,
	HAProxyClientTimeout:               "30m", // 30 minutes
	HAProxyServerTimeout:               "30m", // 30 minutes
	DNSAAAADelayEnabled:                true,
	EtcdDataDirectory:                  "/var/lib/etcd",
	EtcdDefragCrontabHour:              "1",
	EtcdLogsCrontabHour:                "*",
	EtcdLogsCrontabMinute:              "0,30",
	EtcdLogsDaysToKeep:                 30,
	UseLocalImages:                     true,
	ClusterIfaceName:                   "ansible_eth0",
	PublicVipIfaceName:                 "ansible_eth0",
	ClusterIface:                       "{{ hostvars[inventory_hostname][cluster_iface_name] }}",
	PublicVipIface:                     "{{ hostvars[inventory_hostname][public_vip_iface_name] }}",
	CniIface:                           "{{ cluster_iface.device }}",
	CniIPEncapsulation:                 "Always",
	KubeAuditLogMaxAge:                 30,
	KubeAuditLogMaxBackup:              10,
	KubeAuditLogMaxSize:                128,
	RookCephStorageDeviceDirectories:   []string{"/data/rook/storage/0"},
	DockerRegistryPort:                 5000,
	RookEnabled:                        true,
	RookDashboardEnabled:               true,
	RookDashboardExternalEnabled:       true,
	RookDashboardPort:                  7665,
	RookObjectStoreEnabled:             true,
	RookObjectStoreRadosGatewayEnabled: true,
	RookFileStoreEnabled:               true,
	RookCephStorageDeviceFilter:        "",
	DefaultStorageclass:                "rook-ceph-block-delete",
	PodEvictionTimeout:                 "2m",
	TerminatedPodGCThreshold:           100,
	AdditionalRSharedMountPoints:       []string{},
	CloudProviderEnabled:               false,
	SysctlSettings:                     "{{ sysctl_defaults }}",
}

// LoadNilDefault will load assigned defaults when the current value is nil
// this is usually for dealing with omitted values for backwards compatibility
func (c *Config) LoadNilDefault() {
	if c.DNSAAAADelayEnabled == nil {
		c.DNSAAAADelayEnabled = &defaultInventoryVariables.DNSAAAADelayEnabled
	}

	if c.SysctlSettings == nil {
		c.SysctlSettings = &defaultInventoryVariables.SysctlSettings
	}
}

// Inventory creates an Ansible inventory from the Configurator information
func (c *Configurator) Inventory() (*Inventory, error) {
	var vars InventoryVariables

	c.config.LoadNilDefault()

	// since SysctlSettings is an interface that can either be a map or string we need to
	// pre-process it in the former case to the correct type for the json marshaller to understand whats going on
	if m, isMap := c.config.SysctlSettings.(map[interface{}]interface{}); isMap {
		c.config.SysctlSettings = toFlatStringMap(m)
	}

	confB, err := json.Marshal(c.config)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(confB, &vars); err != nil {
		return nil, err
	}

	hostsRoles := map[string]InventoryHosts{}

	for _, host := range c.Hosts {

		roleNameGroup := strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(strings.ToLower(host.RoleName[:len(host.RoleName)-ZeroPadLen]))

		tempLabels, err := getListFromNodePool(c.platformConfig, "kubelet_node_labels", roleNameGroup)
		if err != nil {
			c.ui.Log.Fatalf("unable to extract kubelet_node_labels due to error: %e", err)
		}
		tempTaints, err := getListFromNodePool(c.platformConfig, "kubelet_node_taints", roleNameGroup)
		if err != nil {
			c.ui.Log.Fatalf("unable to extract kubelet_node_taints due to error: %e", err)
		}

		ih := &InventoryHost{
			AnsibleHost:   host.PrivateIP,
			PrivateIP:     host.PrivateIP,
			PublicIP:      host.PublicIP,
			PrivateDNS:    host.PrivateDNS,
			PublicDNS:     host.PublicDNS,
			FQDN:          host.PrivateDNS,
			Hostname:      strings.Split(host.PrivateDNS, ".")[0],
			HostnameShort: strings.Split(host.PublicDNS, ".")[0],

			KubeletLabels: tempLabels,
			KubeletTaints: tempTaints,
		}

		hr, ok := hostsRoles[roleNameGroup]
		if !ok {
			hosts := map[string]*InventoryHost{}
			hr = InventoryHosts{
				Hosts: hosts,
			}
		}

		c.ui.Log.Debugf("Adding the host %s to the inventory group %s", host.RoleName, roleNameGroup)
		hr.Hosts[host.RoleName] = ih

		// TODO remove kube_ from here and all ansible playbooks before merge
		hostsRoles[roleNameGroup] = hr
	}

	i := new(Inventory)

	vars.ClusterName = c.clusterName
	vars.CloudProvider = c.platform

	if v, ok := c.platformConfig["dns_search"]; ok {
		if sa, ok := v.([]interface{}); ok {
			dnsSearch := []string{}
			for _, sb := range sa {
				dnsSearch = append(dnsSearch, sb.(string))
			}
			vars.DNSSearch = dnsSearch
		}
	}

	if v, ok := c.platformConfig["dns_servers"]; ok {
		if sa, ok := v.([]interface{}); ok {
			dnsServers := []string{}
			for _, sb := range sa {
				dnsServers = append(dnsServers, sb.(string))
			}
			vars.DNSServers = dnsServers
		}
	}

	if v, ok := c.platformConfig["time_servers"]; ok {
		if sa, ok := v.([]interface{}); ok {
			timeServers := []string{}
			for _, sb := range sa {
				timeServers = append(timeServers, sb.(string))
			}
			vars.TimeServers = timeServers
		}
	}

	if _, ok := ValidAddressInventoryFields[vars.AddressInventoryField]; !ok {
		vars.AddressInventoryField = "private_dns"
	} else if c.platform == "ec2" {
		// force the use of the private_ip field for aws as the others have issues
		// in particular with DNS, it takes a few seconds to resolve names so its not performant
		vars.AddressInventoryField = "private_ip"
	}

	// Variables from the platform configuration
	if v, ok := c.platformConfig["username"]; ok {
		vars.AnsibleUser = v.(string)
	}
	if v, ok := c.platformConfig["kube_api_ssl_port"]; ok {
		vars.KubeAPISslPort = int(v.(float64))
	}
	if v, ok := c.platformConfig["disable_master_ha"]; ok {
		vars.DisableMasterHA = v.(bool)
	}
	if v, ok := c.platformConfig["kube_virtual_ip_api"]; ok {
		vars.KubeVirtualIPApi = v.(string)
	}
	if v, ok := c.platformConfig["kube_vip_api_ssl_port"]; ok {
		vars.KubeVipAPISslPort = int(v.(float64))
	}
	if v, ok := c.platformConfig["public_virtual_ip"]; ok {
		vars.PublicVirtualIP = v.(string)
	}
	if v, ok := c.platformConfig["public_virtual_ip_ssl_port"]; ok {
		vars.PublicVirtualIPSslPort = int(v.(float64))
	}

	// Only required for AWS
	if c.platform == "ec2" {
		vars.AlbDNSName = c.address
	}

	// Only required for vSphere
	if c.platform == "vsphere" {
		if v, ok := c.platformConfig["datacenter"]; ok {
			vars.VsphereDatacenter = v.(string)
		}
		if v, ok := c.platformConfig["folder"]; ok {
			vars.VsphereFolder = v.(string)
		}
		if v, ok := c.platformConfig["datastore"]; ok {
			vars.VsphereDatastore = v.(string)
		}
		if v, ok := c.stateData["server"]; ok {
			vars.VsphereServer = v.(string)
		}
		if v, ok := c.platformConfig["vsphere_net"]; ok {
			vars.VsphereNet = v.(string)
		}
		if v, ok := c.platformConfig["resource_pool"]; ok {
			vars.VsphereResourcePool = v.(string)
		}
	}

	if c.platform == "stacki" {
		if vars.EtcdDataDirectory == "" {
			vars.EtcdDataDirectory = "/data/etcd"
		}
		if vars.EtcdSnapshotsDirectory == "" {
			vars.EtcdSnapshotsDirectory = "/data/etcd-snapshots"
		}
	} else {
		if vars.EtcdDataDirectory == "" {
			vars.EtcdDataDirectory = defaultInventoryVariables.EtcdDataDirectory
		}
		if vars.EtcdSnapshotsDirectory == "" {
			vars.EtcdSnapshotsDirectory = defaultInventoryVariables.EtcdSnapshotsDirectory
		}
	}

	i.All.Variables = vars

	i.All.Children.KubeCluster.Children = map[string]InventoryHosts{}

	for k, v := range hostsRoles {
		i.All.Children.KubeCluster.Children[k] = v
	}

	return i, nil
}

func toFlatStringMap(m map[interface{}]interface{}) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		result[fmt.Sprintf("%v", k)] = fmt.Sprintf("%v", v)
	}
	return result
}

// Yaml returns the Inventory structure in YAML format
func (i *Inventory) Yaml() ([]byte, error) {
	return yaml.Marshal(i)
}

// JSON returns the Inventory structure in JSON format
func (i *Inventory) JSON() ([]byte, error) {
	return json.Marshal(i)
}

// JSON returns the Inventory Variables structure in JSON format
func (iv *InventoryVariables) JSON() ([]byte, error) {
	return json.Marshal(iv)
}

// Map returns the Inventory Variables structure as a map of strings
func (iv *InventoryVariables) Map() (map[string]interface{}, error) {
	b, err := iv.JSON()
	if err != nil {
		return nil, err
	}
	var bucket interface{}
	err = json.Unmarshal(b, &bucket)
	if err != nil {
		return nil, err
	}
	return bucket.(map[string]interface{}), nil
}

func getListFromNodePool(m map[string]interface{}, key string, pool string) ([]string, error) {
	switch pool {
	case "default_node_pool":
		nodePool, ok := m["default_node_pool"].(map[string]interface{})
		if ok {
			return getListFromNodePool(nodePool, key, "target_node_pool")
		}
	case "target_node_pool":
		if targetList, ok := m[key]; ok {
			if reflect.TypeOf(targetList).Kind() == reflect.Slice {
				l := []string{}
				for _, v := range targetList.([]interface{}) {
					if reflect.TypeOf(v).Kind() == reflect.String {
						l = append(l, v.(string))
					} else {
						return []string{}, fmt.Errorf("cannot parse item as string")
					}
				}
				return l, nil
			}
			return []string{}, fmt.Errorf("cannot parse item as list")
		}
	default:
		nodePools, ok := m["node_pools"].(map[string]interface{})
		if !ok {
			return getListFromNodePool(m, key, "default_node_pool")
		}
		nodePool, ok := nodePools[pool].(map[string]interface{})
		if !ok {
			return getListFromNodePool(m, key, "default_node_pool")
		}
		targetList, err := getListFromNodePool(nodePool, key, "target_node_pool")
		if err == nil && len(targetList) == 0 {
			return getListFromNodePool(m, key, "default_node_pool")
		} else if len(targetList) == 0 {
			return []string{}, err
		}
		return targetList, err
	}
	return []string{}, nil
}
