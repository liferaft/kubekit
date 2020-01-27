data "azurerm_subscription" "aks" {}

data "azurerm_client_config" "aks" {}

data "azurerm_virtual_network" "aks" {
  depends_on = [
    "azurerm_subnet.aks",
  ]

  {{ if and ( ne .VnetName "" ) ( ne .VnetResourceGroupName "" ) -}}
  name                = "{{ .VnetName }}"
  resource_group_name = "{{ .VnetResourceGroupName }}"
  {{- else -}}
  name                = azurerm_virtual_network.aks.name
  resource_group_name = azurerm_resource_group.aks.name
  {{- end }}
}

data "azurerm_kubernetes_service_versions" "aks" {
  location = "{{ .ResourceGroupLocation }}"
}
