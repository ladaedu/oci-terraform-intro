variable "tenancy_ocid" {}
variable "user_ocid" {}
variable "fingerprint" {}
variable "private_key_path" {}
variable "home_region" {}
variable "compartment_create" {
  default = true
}
variable "group_create" {
  default = true
}
variable "policy_create" {
  default = true
}
variable "name" {}
