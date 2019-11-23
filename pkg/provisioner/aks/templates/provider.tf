provider "azurerm" {
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
