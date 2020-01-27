package configurator

import (
	"encoding/json"
	"fmt"

	"github.com/johandry/merger"
)

// Config are all the settings to configure Kubernetes no matter the platform
type Config struct { //  aws
	ShellEditingMode                        string      `json:"shell_editing_mode,omitempty" yaml:"shell_editing_mode,omitempty" mapstructure:"shell_editing_mode,omitempty"`
	AddressInventoryField                   string      `json:"address_inventory_field" yaml:"address_inventory_field" mapstructure:"address_inventory_field"`
	EtcdInitialClusterToken                 string      `json:"etcd_initial_cluster_token" yaml:"etcd_initial_cluster_token" mapstructure:"etcd_initial_cluster_token"`
	KubeletMaxPods                          int         `json:"kubelet_max_pods" yaml:"kubelet_max_pods" mapstructure:"kubelet_max_pods"`
	KubeletSerializeImagePulls              bool        `json:"kubelet_serialize_image_pulls,omitempty" yaml:"kubelet_serialize_image_pulls,omitempty" mapstructure:"kubelet_serialize_image_pulls,omitempty"`
	KubeProxyMode                           string      `json:"kube_proxy_mode,omitempty" yaml:"kube_proxy_mode,omitempty" mapstructure:"kube_proxy_mode,omitempty"`
	KubeClusterCidr                         string      `json:"kube_cluster_cidr" yaml:"kube_cluster_cidr" mapstructure:"kube_cluster_cidr"`
	KubeServicesCidr                        string      `json:"kube_services_cidr" yaml:"kube_services_cidr" mapstructure:"kube_services_cidr"`
	KubeServiceIP                           string      `json:"kube_service_ip" yaml:"kube_service_ip" mapstructure:"kube_service_ip"`
	KubeAdvertiseAddress                    string      `json:"kube_advertise_address" yaml:"kube_advertise_address" mapstructure:"kube_advertise_address"`
	MasterSchedulableEnabled                bool        `json:"master_schedulable_enabled" yaml:"master_schedulable_enabled" mapstructure:"master_schedulable_enabled"`
	EtcdLocalProxyEnabled                   bool        `json:"enable_etcd_local_proxy" yaml:"enable_etcd_local_proxy" mapstructure:"enable_etcd_local_proxy"`
	DockerMaxConcurrentUploads              int         `json:"docker_max_concurrent_uploads,omitempty" yaml:"docker_max_concurrent_uploads,omitempty" mapstructure:"docker_max_concurrent_uploads,omitempty"`
	DockerMaxConcurrentDownloads            int         `json:"docker_max_concurrent_downloads,omitempty" yaml:"docker_max_concurrent_downloads,omitempty" mapstructure:"docker_max_concurrent_downloads,omitempty"`
	DockerLogMaxFiles                       string      `json:"docker_log_max_files,omitempty" yaml:"docker_log_max_files,omitempty" mapstructure:"docker_log_max_files"` // docker config takes it as a string
	DockerLogMaxSize                        string      `json:"docker_log_max_size,omitempty" yaml:"docker_log_max_size,omitempty" mapstructure:"docker_log_max_size"`
	DockerRegistryPort                      int         `json:"registry_port" yaml:"registry_port" mapstructure:"registry_port"`
	DockerRegistryPath                      string      `json:"docker_registry_path" yaml:"docker_registry_path" mapstructure:"docker_registry_path"`
	DownloadImagesIfMissing                 bool        `json:"download_images_if_missing" yaml:"download_images_if_missing" mapstructure:"download_images_if_missing"`
	HAProxyClientTimeout                    string      `json:"haproxy_client_timeout,omitempty" yaml:"haproxy_client_timeout,omitempty" mapstructure:"haproxy_client_timeout"`
	HAProxyServerTimeout                    string      `json:"haproxy_server_timeout,omitempty" yaml:"haproxy_server_timeout,omitempty" mapstructure:"haproxy_server_timeout"`
	DNSAAAADelayEnabled                     *bool       `json:"dns_aaaa_delay_enabled,omitempty" yaml:"dns_aaaa_delay_enabled,omitempty" mapstructure:"dns_aaaa_delay_enabled"`
	DNSArgs                                 string      `json:"dns_args" yaml:"dns_args" mapstructure:"dns_args"`
	EtcdDataDirectory                       string      `json:"etcd_data_directory,omitempty" yaml:"etcd_data_directory,omitempty" mapstructure:"etcd_data_directory"`
	EtcdDefragCrontabHour                   string      `json:"etcd_defrag_crontab_hour,omitempty" yaml:"etcd_defrag_crontab_hour,omitempty" mapstructure:"etcd_defrag_crontab_hour"`
	EtcdLogsCrontabHour                     string      `json:"etcd_logs_crontab_hour" yaml:"etcd_logs_crontab_hour" mapstructure:"etcd_logs_crontab_hour"`
	EtcdLogsCrontabMinute                   string      `json:"etcd_logs_crontab_minute" yaml:"etcd_logs_crontab_minute" mapstructure:"etcd_logs_crontab_minute"`
	EtcdLogsDaysToKeep                      int         `json:"etcd_logs_days_to_keep" yaml:"etcd_logs_days_to_keep" mapstructure:"etcd_logs_days_to_keep"`
	EtcdSnapshotsDirectory                  string      `json:"etcd_snapshots_directory,omitempty" yaml:"etcd_snapshots_directory,omitempty" mapstructure:"etcd_snapshots_directory"`
	EtcdQuotaBackendBytes                   int         `json:"etcd_quota_backend_bytes,omitempty" yaml:"etcd_quota_backend_bytes,omitempty" mapstructure:"etcd_quota_backend_bytes"`
	UseLocalImages                          bool        `json:"use_local_images" yaml:"use_local_images" mapstructure:"use_local_images"`
	ClusterIfaceName                        string      `json:"cluster_iface_name" yaml:"cluster_iface_name" mapstructure:"cluster_iface_name"`
	ClusterIface                            string      `json:"cluster_iface" yaml:"cluster_iface" mapstructure:"cluster_iface"`
	CniIface                                string      `json:"cni_iface" yaml:"cni_iface" mapstructure:"cni_iface"`
	CniIPEncapsulation                      string      `json:"cni_ip_encapsulation" yaml:"cni_ip_encapsulation" mapstructure:"cni_ip_encapsulation"`
	PublicVipIfaceName                      string      `json:"public_vip_iface_name" yaml:"public_vip_iface_name" mapstructure:"public_vip_iface_name"`
	PublicVipIface                          string      `json:"public_vip_iface" yaml:"public_vip_iface" mapstructure:"public_vip_iface"`
	KubeAuditLogMaxAge                      int         `json:"kube_audit_log_max_age" yaml:"kube_audit_log_max_age" mapstructure:"kube_audit_log_max_age"`
	KubeAuditLogMaxBackup                   int         `json:"kube_audit_log_max_backup" yaml:"kube_audit_log_max_backup" mapstructure:"kube_audit_log_max_backup"`
	KubeAuditLogMaxSize                     int         `json:"kube_audit_log_max_size" yaml:"kube_audit_log_max_size" mapstructure:"kube_audit_log_max_size"`
	NginxIngressEnabled                     bool        `json:"nginx_ingress_enabled" yaml:"nginx_ingress_enabled" mapstructure:"nginx_ingress_enabled"`
	NginxIngressControllerProxyBodySize     string      `json:"nginx_ingress_controller_proxy_body_size" yaml:"nginx_ingress_controller_proxy_body_size" mapstructure:"nginx_ingress_controller_proxy_body_size"`
	NginxIngressControllerErrorLogLevel     string      `json:"nginx_ingress_controller_error_log_level" yaml:"nginx_ingress_controller_error_log_level" mapstructure:"nginx_ingress_controller_error_log_level"`
	NginxIngressControllerSslProtocols      string      `json:"nginx_ingress_controller_ssl_protocols" yaml:"nginx_ingress_controller_ssl_protocols" mapstructure:"nginx_ingress_controller_ssl_protocols"`
	NginxIngressControllerProxyReadTimeout  string      `json:"nginx_ingress_controller_proxy_read_timeout" yaml:"nginx_ingress_controller_proxy_read_timeout" mapstructure:"nginx_ingress_controller_proxy_read_timeout"`
	NginxIngressControllerProxySendTimeout  string      `json:"nginx_ingress_controller_proxy_send_timeout" yaml:"nginx_ingress_controller_proxy_send_timeout" mapstructure:"nginx_ingress_controller_proxy_send_timeout"`
	NginxIngressControllerTLSCertLocalPath  string      `json:"nginx_ingress_controller_tls_cert_local_path" yaml:"nginx_ingress_controller_tls_cert_local_path" mapstructure:"nginx_ingress_controller_tls_cert_local_path"`
	NginxIngressControllerTLSKeyLocalPath   string      `json:"nginx_ingress_controller_tls_key_local_path" yaml:"nginx_ingress_controller_tls_key_local_path" mapstructure:"nginx_ingress_controller_tls_key_local_path"`
	NginxIngressControllerBasicAuthUsername string      `json:"nginx_ingress_controller_basic_auth_username" yaml:"nginx_ingress_controller_basic_auth_username" mapstructure:"nginx_ingress_controller_basic_auth_username"`
	NginxIngressControllerBasicAuthPassword string      `json:"nginx_ingress_controller_basic_auth_password" yaml:"nginx_ingress_controller_basic_auth_password" mapstructure:"nginx_ingress_controller_basic_auth_password"`
	DefaultIngressHost                      string      `json:"default_ingress_host" yaml:"default_ingress_host" mapstructure:"default_ingress_host"`
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
	HostTimeZone                            string      `json:"host_timezone,omitempty" yaml:"host_timezone" mapstructure:"host_timezone"`
	ControlPlaneTimeZone                    string      `json:"controlplane_timezone,omitempty" yaml:"controlplane_timezone" mapstructure:"controlplane_timezone"`
	CloudProviderEnabled                    bool        `json:"cloud_provider_enabled,omitempty" yaml:"cloud_provider_enabled" mapstructure:"cloud_provider_enabled"`
	PodEvictionTimeout                      string      `json:"pod_eviction_timeout" yaml:"pod_eviction_timeout" mapstructure:"pod_eviction_timeout"`
	TerminatedPodGCThreshold                int         `json:"terminated_pod_gc_threshold" yaml:"terminated_pod_gc_threshold" mapstructure:"terminated_pod_gc_threshold"`
	AdditionalRSharedMountPoints            []string    `json:"additional_rshared_mount_points,omitempty" yaml:"additional_rshared_mount_points,omitempty" mapstructure:"additional_rshared_mount_points"`
	WaitForReady                            int         `json:"wait_for_ready" yaml:"wait_for_ready" mapstructure:"wait_for_ready"`
	SysctlSettings                          interface{} `json:"sysctl_settings,omitempty" yaml:"sysctl_settings,omitempty" mapstructure:"sysctl_settings"`
}

// DefaultConfig returns the default configuration of the configurator based on
// the default inventory
func DefaultConfig(envConfig map[string]string) (*Config, error) {
	defaultConfig := Config{}
	jsonInventory, err := defaultInventoryVariables.JSON()
	if err != nil {
		return nil, err
	}

	// DEBUG:
	// fmt.Printf("[DEBUG] Inventory: %s\n", jsonInventory)
	if err = json.Unmarshal(jsonInventory, &defaultConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal the default inventory. %s", err)
	}

	config := Config{}
	if err := merger.Merge(&config, envConfig, defaultConfig); err != nil {
		return nil, err
	}
	// DEBUG:
	// fmt.Printf("[DEBUG] Config: %v\n", config)

	return &config, nil
}

// Map converts the current configuration to a map of string of strings
func (c *Config) Map() (map[string]interface{}, error) {
	var configM map[string]interface{}

	configB, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configB, &configM)

	return configM, nil
}
