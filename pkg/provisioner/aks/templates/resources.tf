resource "azurerm_resource_group" "aks" {
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
