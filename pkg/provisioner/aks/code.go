package aks

// Code generated automatically by 'go run codegen/main.go --pkg <pkg> --src <pkg>/templates --dst <pkg>/code.go'; DO NOT EDIT THIS FILE.

func init() {
	ResourceTemplates = map[string]string{
		"data-sources": dataSourcesTpl,
		"output":       outputTpl,
		"provider":     providerTpl,
		"resources":    resourcesTpl,
		"variables":    variablesTpl,
	}
}

// Expressions in the templates
/**
data-sources : {{ if and ( ne .VnetName "" ) ( ne .VnetResourceGroupName "" ) -}}
data-sources : {{ .VnetName }}
data-sources : {{ .VnetResourceGroupName }}
data-sources : {{- else -}}
data-sources : {{- end }}
data-sources : {{ .ResourceGroupLocation }}
output : {{ if and ( .ContainerRegistrySku ) ( ne .ContainerRegistrySku "" ) }}
output : {{ end }}
output : {{ if .Jumpbox }}
output : {{ end }}
provider : {{ .Environment }}
resources : {{ .ClusterName }}
resources : {{ .ResourceGroupLocation }}
resources : {{ .ClusterName }}
resources : {{ if or ( eq .VnetName "" ) ( eq .VnetResourceGroupName "" ) }}
resources : {{ .ClusterName }}
resources : {{ DefaultString .VnetAddressSpace "10.240.0.0/16" }}
resources : {{ end }}
resources : {{ .ClusterName }}
resources : {{ DefaultString .SubnetAddressPrefix "10.240.0.0/20" }}
resources : {{ if and ( ne .VnetName "" ) ( ne .VnetResourceGroupName "" ) -}}
resources : {{ .VnetResourceGroupName }}
resources : {{ .VnetName }}
resources : {{- else -}}
resources : {{- end }}
resources : {{ if ne .PrivateDNSZoneName "" }}
resources : {{ .PrivateDNSZoneName }}
resources : {{ end }}
resources : {{ if ne .PrivateDNSZoneName "" -}}
resources : {{- end }}
resources : {{ .ClusterName }}
resources : {{ if ne .DNSPrefix "" -}}
resources : {{ .DNSPrefix }}
resources : {{- else -}}
resources : {{ Dash ( Lower .ClusterName ) }}
resources : {{- end }}
resources : {{ if ne .KubernetesVersion "" -}}
resources : {{ .KubernetesVersion }}
resources : {{- else -}}
resources : {{- end }}
resources : {{ if .EnablePodSecurityPolicy -}}
resources : {{ BoolToString .EnablePodSecurityPolicy }}
resources : {{- end }}
resources : {{ DefaultString .AdminUsername "kubekit" }}
resources : {{ Trim .PublicKey }}
resources : {{ if ne .DNSServiceIP "" -}}
resources : {{ .DNSServiceIP }}
resources : {{- else -}}
resources : {{ DefaultString .ServiceCIDR "172.21.0.0/16" }}
resources : {{- end }}
resources : {{ DefaultString .DockerBridgeCIDR "172.17.0.1/16" }}
resources : {{ DefaultString .ServiceCIDR "172.21.0.0/16" }}
resources : {{ DefaultString .NetworkPolicy "calico" }}
resources : {{ range $k, $v := .NodePools }}
resources : {{ Lower ( Alphanumeric $v.Name ) }}
resources : {{ $v.Count }}
resources : {{ $v.Type }}
resources : {{ $v.VMSize }}
resources : {{ $v.RootVolumeSize }}
resources : {{ $v.MaxPods }}
resources : {{ $v.EnableAutoScaling }}
resources : {{ if .EnableAutoScaling }}
resources : {{ if gt $v.AutoScalingMinCount 0 }}
resources : {{ $v.AutoScalingMinCount }}
resources : {{ end }}
resources : {{ if gt $v.AutoScalingMaxCount 0 }}
resources : {{ $v.AutoScalingMaxCount }}
resources : {{ end }}
resources : {{ end }}
resources : {{ if gt (len $v.AvailabilityZones) 0 }}
resources : {{ QuoteList $v.AvailabilityZones }}
resources : {{ else }}
resources : {{ end }}
resources : {{ if gt (len $v.NodeTaints) 0 }}
resources : {{ QuoteList $v.NodeTaints }}
resources : {{ else }}
resources : {{ end }}
resources : {{ end }}
resources : {{ if and ( ne .ClusterClientID "" ) ( ne .ClusterClientSecret "" ) -}}
resources : {{ .ClusterClientID }}
resources : {{ .ClusterClientSecret }}
resources : {{- else -}}
resources : {{ .ClientID }}
resources : {{ .ClientSecret }}
resources : {{- end }}
resources : {{ .ClusterName }}
resources : {{ if and ( .ContainerRegistrySku ) ( ne .ContainerRegistrySku "" ) }}
resources : {{ AlphanumericHyphen ( Dash ( Lower .ClusterName ) ) }}
resources : {{ .ContainerRegistrySku }}
resources : {{ .ContainerRegistryAdminEnabled }}
resources : {{ end }}
resources : {{ if .Jumpbox }}
resources : {{ if .Jumpbox.EnablePublicIP }}
resources : {{ if ne .PrivateDNSZoneName "" -}}
resources : {{- end }}
resources : {{ if ne .PrivateDNSZoneName "" -}}
resources : {{- end }}
resources : {{ end }}
resources : {{ if ne .PrivateDNSZoneName "" -}}
resources : {{- end }}
resources : {{ .ClusterName }}
resources : {{ DefaultString .Jumpbox.VMSize "Standard_B1s" }}
resources : {{ DefaultString .Jumpbox.UbuntuServerSku "18.04-LTS" }}
resources : {{ DefaultInt .Jumpbox.RootVolumeSize 30 }}
resources : {{ DefaultString .Jumpbox.AdminUsername "kubekit-jumpbox" }}
resources : {{ Trim ( DefaultString .Jumpbox.PublicKey .PublicKey ) }}
resources : {{ DefaultString .Jumpbox.AdminUsername "kubekit-jumpbox" }}
resources : {{ if and (.Jumpbox.NSGRules) (gt (len .Jumpbox.NSGRules) 0) }}
resources : {{ if ne .PrivateDNSZoneName "" -}}
resources : {{- end }}
resources : {{ .ClusterName }}
resources : {{ range $i, $rule := .Jumpbox.NSGRules }}
resources : {{ $rule_priority := DefaultInt $rule.Priority ( Multiply ( len .Jumpbox.NSGRules ) 100 ) }}
resources : {{ DefaultString $rule.Name ( print "rule-priority" $rule_priority ) }}
resources : {{ if ne .PrivateDNSZoneName "" -}}
resources : {{- end }}
resources : {{ DefaultString $rule.Name ( print "rule-priority" $rule_priority ) }}
resources : {{ $rule_priority }}
resources : {{ DefaultString $rule.Direction "Inbound" }}
resources : {{ $rule.Access }}
resources : {{ DefaultString $rule.Protocol "Tcp" }}
resources : {{ DefaultString $rule.SourcePortRange "*" }}
resources : {{ DefaultString $rule.DestinationPortRange "*" }}
resources : {{ DefaultString $rule.SourceAddressPrefix "*" }}
resources : {{ DefaultString $rule.DestinationAddressPrefix "*" }}
resources : {{ end }}
resources : {{ end }}
resources : {{ end }}
**/

const dataSourcesTpl = `data "azurerm_subscription" "aks" {}

data "azurerm_client_config" "aks" {}

data "azurerm_virtual_network" "aks" {
  depends_on = [
    "azurerm_subnet.aks",
  ]

  {{ if and ( ne .VnetName "" ) ( ne .VnetResourceGroupName "" ) -}}
  name                = "{{ .VnetName }}"
  resource_group_name = "{{ .VnetResourceGroupName }}"
  {{- else -}}
  name                = "${azurerm_virtual_network.aks.name}"
  resource_group_name = "${azurerm_resource_group.aks.name}"
  {{- end }}
}

data "azurerm_kubernetes_service_versions" "aks" {
  location = "{{ .ResourceGroupLocation }}"
}
`

const outputTpl = `output "kubernetes_version" {
  value = "${azurerm_kubernetes_cluster.aks.kubernetes_version}"
}

output "vnet" {
  value = "${data.azurerm_virtual_network.aks.name}"
}

output "node_resource_group" {
  value = "${azurerm_kubernetes_cluster.aks.node_resource_group}"
}

output "kubeconfig" {
  value = "${azurerm_kubernetes_cluster.aks.kube_config_raw}"
}

output "fqdn" {
  value = "${azurerm_kubernetes_cluster.aks.fqdn}"
}

output "host" {
  value = "${azurerm_kubernetes_cluster.aks.kube_config.0.host}"
}

output "client_key" {
  value = "${azurerm_kubernetes_cluster.aks.kube_config.0.client_key}"
}

output "client_cert" {
  value = "${azurerm_kubernetes_cluster.aks.kube_config.0.client_certificate}"
}

output "ca_cert" {
  value = "${azurerm_kubernetes_cluster.aks.kube_config.0.cluster_ca_certificate}"
}

{{ if and ( .ContainerRegistrySku ) ( ne .ContainerRegistrySku "" ) }}
// a registry can have geospatial replication, which means multiple servers/users
// but we will disable geospatial replication to only allow for 1 server and user for now
output "container_registry_server" {
  value = "${azurerm_container_registry.aks.login_server}"
}

output "container_registry_admin_enabled" {
  value = "${azurerm_container_registry.aks.admin_enabled}"
}

output "container_registry_admin_users" {
  value = "${azurerm_container_registry.aks.admin_username}"
}

output "container_registry_admin_passwords" {
  value = "${azurerm_container_registry.aks.admin_password}"
}
{{ end }}

{{ if .Jumpbox }}
output "jumpbox" {
  value = "${azurerm_network_interface.jumpbox.private_ip_address}"
}
{{ end }}
`

const providerTpl = `provider "azurerm" {
  subscription_id = "${var.subscription_id}"
  tenant_id       = "${var.tenant_id}"
  environment     = "{{ .Environment }}"
  client_id       = "${var.client_id}"
  client_secret   = "${var.client_secret}"
  client_certificate_password = "${var.client_certificate_password}"
  client_certificate_path     = "${var.client_certificate_path}"
  msi_endpoint    = "${var.msi_endpoint}"
  use_msi         = "${var.use_msi}"
}

terraform {
  backend "azurerm" {}
}
`

const resourcesTpl = `resource "azurerm_resource_group" "aks" {
  name     = "{{ .ClusterName }}"
  location = "{{ .ResourceGroupLocation }}"

  tags = {
    ClusterName = "{{ .ClusterName }}"
  }
}

{{ if or ( eq .VnetName "" ) ( eq .VnetResourceGroupName "" ) }}
resource "azurerm_virtual_network" "aks" {
  name                = "{{ .ClusterName }}"
  address_space       = ["{{ DefaultString .VnetAddressSpace "10.240.0.0/16" }}"]  // even though its a list, it takes a single value
  location            = "${azurerm_resource_group.aks.location}"
  resource_group_name = "${azurerm_resource_group.aks.name}"
}
{{ end }}

resource "azurerm_subnet" "aks" {
  name                 = "{{ .ClusterName }}"
  address_prefix       = "{{ DefaultString .SubnetAddressPrefix "10.240.0.0/20" }}"

  {{ if and ( ne .VnetName "" ) ( ne .VnetResourceGroupName "" ) -}}
  // user provided vnet
  resource_group_name  = "{{ .VnetResourceGroupName }}"
  virtual_network_name = "{{ .VnetName }}"
  {{- else -}}
  // provisioned vnet
  resource_group_name  = "${azurerm_resource_group.aks.name}"
  virtual_network_name = "${azurerm_virtual_network.aks.name}"
  {{- end }}
}

{{ if ne .PrivateDNSZoneName "" }}
resource "azurerm_private_dns_zone" "aks" {
  name                = "{{ .PrivateDNSZoneName }}"
  resource_group_name = "${azurerm_resource_group.aks.name}"
}
{{ end }}

resource "azurerm_kubernetes_cluster" "aks" {
  depends_on = [
    "data.azurerm_kubernetes_service_versions.aks",
    "azurerm_resource_group.aks",
    {{ if ne .PrivateDNSZoneName "" -}}
    "azurerm_subnet.aks",
    {{- end }}
  ]

  name                = "{{ .ClusterName }}"
  location            = "${azurerm_resource_group.aks.location}"
  resource_group_name = "${azurerm_resource_group.aks.name}"

  {{ if ne .DNSPrefix "" -}}
  dns_prefix          = "{{ .DNSPrefix }}"
  {{- else -}}
  dns_prefix          = "{{ Dash ( Lower .ClusterName ) }}"
  {{- end }}

  {{ if ne .KubernetesVersion "" -}}
  kubernetes_version  = "{{ .KubernetesVersion }}"
  {{- else -}}
  kubernetes_version  = "${data.azurerm_kubernetes_service_versions.aks.latest_version}"
  {{- end }}

//  kubernetes_version  = "${var.kubernetes_version == "" ? data.azurerm_kubernetes_service_versions.aks.latest_version}"

  role_based_access_control {
    enabled = true
  }

  {{ if .EnablePodSecurityPolicy -}}
  enable_pod_security_policy = "{{ BoolToString .EnablePodSecurityPolicy }}"
  {{- end }}
//  node_resource_group = "${azurerm_resource_group.aks.name}"

  linux_profile {
    admin_username = "{{ DefaultString .AdminUsername "kubekit" }}"

    ssh_key {
      key_data = "{{ Trim .PublicKey }}"
    }
  }

  network_profile {
    network_plugin     = "azure"

    {{ if ne .DNSServiceIP "" -}}
    dns_service_ip     = "{{ .DNSServiceIP }}"
    {{- else -}}
    dns_service_ip     = "${cidrhost("{{ DefaultString .ServiceCIDR "172.21.0.0/16" }}", 10)}"
    {{- end }}

    docker_bridge_cidr = "{{ DefaultString .DockerBridgeCIDR "172.17.0.1/16" }}"
    service_cidr       = "{{ DefaultString .ServiceCIDR "172.21.0.0/16" }}"
    network_policy     = "{{ DefaultString .NetworkPolicy "calico" }}"
  }

  {{ range $k, $v := .NodePools }}
  agent_pool_profile {
    name                = "{{ Lower ( Alphanumeric $v.Name ) }}"
    count               = "{{ $v.Count }}"
    type                = "{{ $v.Type }}"
    vm_size             = "{{ $v.VMSize }}"
    os_type             = "Linux"
    os_disk_size_gb     = "{{ $v.RootVolumeSize }}"
    max_pods            = "{{ $v.MaxPods }}"
    vnet_subnet_id      = "${azurerm_subnet.aks.id}"
    enable_auto_scaling = "{{ $v.EnableAutoScaling }}"

    {{ if .EnableAutoScaling }}
    {{ if gt $v.AutoScalingMinCount 0 }}
    min_count           = "{{ $v.AutoScalingMinCount }}"
    {{ end }}
    {{ if gt $v.AutoScalingMaxCount 0 }}
    max_count           = "{{ $v.AutoScalingMaxCount }}"
    {{ end }}
    {{ end }}

    {{ if gt (len $v.AvailabilityZones) 0 }}
    availability_zones  = [{{ QuoteList $v.AvailabilityZones }}]
    {{ else }}
    availability_zones  = []
    {{ end }}
    {{ if gt (len $v.NodeTaints) 0 }}
    node_taints         = [{{ QuoteList $v.NodeTaints }}]
    {{ else }}
    node_taints         = []
    {{ end }}
  }

  {{ end }}

  service_principal {
    {{ if and ( ne .ClusterClientID "" ) ( ne .ClusterClientSecret "" ) -}}
    client_id     = "{{ .ClusterClientID }}"
    client_secret = "{{ .ClusterClientSecret }}"
    {{- else -}}
    client_id     = "{{ .ClientID }}"
    client_secret = "{{ .ClientSecret }}"
    {{- end }}
  }

  tags = {
    ClusterName = "{{ .ClusterName }}"
  }
}

{{ if and ( .ContainerRegistrySku ) ( ne .ContainerRegistrySku "" ) }}
// currently setup without georeplication
resource "azurerm_container_registry" "aks" {
  name                     = "{{ AlphanumericHyphen ( Dash ( Lower .ClusterName ) ) }}"
  location                 = "${azurerm_resource_group.aks.location}"
  resource_group_name      = "${azurerm_resource_group.aks.name}"
  sku                      = "{{ .ContainerRegistrySku }}"
  admin_enabled            = {{ .ContainerRegistryAdminEnabled }}
  //georeplication_locations = "${var.container_registry_georeplication_locations}"
}
{{ end }}

{{ if .Jumpbox }}
{{ if .Jumpbox.EnablePublicIP }}
resource "azurerm_public_ip" "jumpbox" {
  depends_on = [
    "azurerm_subnet.aks",
    {{ if ne .PrivateDNSZoneName "" -}}
    "azurerm_private_dns_zone.aks",
    {{- end }}
    "azurerm_kubernetes_cluster.aks",
  ]

  name                = "jumpbox"
  location            = "${azurerm_resource_group.aks.location}"
  resource_group_name = "${azurerm_resource_group.aks.name}"
  allocation_method   = "Dynamic"
}

resource "azurerm_network_interface" "jumpbox" {
  depends_on = [
    "azurerm_subnet.aks",
    {{ if ne .PrivateDNSZoneName "" -}}
    "azurerm_private_dns_zone.aks",
    {{- end }}
    "azurerm_kubernetes_cluster.aks",
  ]

  name                = "jumpbox"
  location            = "${azurerm_resource_group.aks.location}"
  resource_group_name = "${azurerm_resource_group.aks.name}"

  ip_configuration {
    name                          = "jumpbox"
    subnet_id                     = "${azurerm_subnet.aks.id}"
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = "${azurerm_public_ip.jumpbox.id}"
  }
}
{{ end }}

resource "azurerm_virtual_machine" "jumpbox" {
  depends_on = [
    "azurerm_subnet.aks",
    {{ if ne .PrivateDNSZoneName "" -}}
    "azurerm_private_dns_zone.aks",
    {{- end }}
    "azurerm_kubernetes_cluster.aks",
  ]

  name                  = "{{ .ClusterName }}-jumpbox"
  location              = "${azurerm_resource_group.aks.location}"
  resource_group_name   = "${azurerm_resource_group.aks.name}"
  network_interface_ids = ["${azurerm_network_interface.jumpbox.id}"]
  vm_size               = "{{ DefaultString .Jumpbox.VMSize "Standard_B1s" }}"

  storage_image_reference {
    publisher = "Canonical"
    offer     = "UbuntuServer"
    sku       = "{{ DefaultString .Jumpbox.UbuntuServerSku "18.04-LTS" }}"
    version   = "latest"
  }

  storage_os_disk {
    name              = "osdisk"
    caching           = "ReadWrite"
    create_option     = "FromImage"
    managed_disk_type = "Standard_LRS"
    disk_size_gb      = "{{ DefaultInt .Jumpbox.RootVolumeSize 30 }}"
  }

  os_profile {
    computer_name  = "jumpbox"
    admin_username = "{{ DefaultString .Jumpbox.AdminUsername "kubekit-jumpbox" }}"
  }

  os_profile_linux_config {
    disable_password_authentication = true

    ssh_keys {
      key_data = "{{ Trim ( DefaultString .Jumpbox.PublicKey .PublicKey ) }}"
      path     = "/home/{{ DefaultString .Jumpbox.AdminUsername "kubekit-jumpbox" }}/.ssh/authorized_keys"
    }
  }
}

{{ if and (.Jumpbox.NSGRules) (gt (len .Jumpbox.NSGRules) 0) }}
// if no rules get added, it will inherit from the global network security group rules
resource "azurerm_network_security_group" "jumpbox" {
depends_on = [
"azurerm_subnet.aks",
{{ if ne .PrivateDNSZoneName "" -}}
"azurerm_private_dns_zone.aks",
{{- end }}
"azurerm_kubernetes_cluster.aks",
]

name                = "{{ .ClusterName }}-jumpbox"
location            = "${azurerm_resource_group.aks.location}"
resource_group_name = "${azurerm_resource_group.aks.name}"
}

{{ range $i, $rule := .Jumpbox.NSGRules }}
{{ $rule_priority := DefaultInt $rule.Priority ( Multiply ( len .Jumpbox.NSGRules ) 100 ) }}
resource "azurerm_network_security_rule" "jumpbox-{{ DefaultString $rule.Name ( print "rule-priority" $rule_priority ) }}" {
  depends_on = [
    "azurerm_subnet.aks",
    {{ if ne .PrivateDNSZoneName "" -}}
    "azurerm_private_dns_zone.aks",
    {{- end }}
    "azurerm_kubernetes_cluster.aks",
  ]

  resource_group_name         = "${azurerm_resource_group.aks.name}"
  network_security_group_name = "${azurerm_network_security_group.jumpbox.name}"

  name                        = "{{ DefaultString $rule.Name ( print "rule-priority" $rule_priority ) }}"
  priority                    = "{{ $rule_priority }}"
  direction                   = "{{ DefaultString $rule.Direction "Inbound" }}"
  access                      = "{{ $rule.Access }}"
  protocol                    = "{{ DefaultString $rule.Protocol "Tcp" }}"
  source_port_range           = "{{ DefaultString $rule.SourcePortRange "*" }}"
  destination_port_range      = "{{ DefaultString $rule.DestinationPortRange "*" }}"
  source_address_prefix       = "{{ DefaultString $rule.SourceAddressPrefix "*" }}"
  destination_address_prefix  = "{{ DefaultString $rule.DestinationAddressPrefix "*" }}"

}

{{ end }}
{{ end }}
{{ end }}
`

const variablesTpl = `variable "subscription_id" {}
variable "tenant_id" {}

variable "client_id" {
  default = ""
}
variable "client_secret" {
  default = ""
}

variable "client_certificate_password" {
  default = ""
}
variable "client_certificate_path" {
  default = ""
}
variable "msi_endpoint" {
  default = ""
}
variable "use_msi" {
  default = false
}
`