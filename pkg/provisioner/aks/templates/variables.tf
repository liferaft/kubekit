variable "subscription_id" {}
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
