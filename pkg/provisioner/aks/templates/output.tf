output "kubernetes_version" {
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
