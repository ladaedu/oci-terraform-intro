# Gets a list of Availability Domains
data "oci_identity_availability_domains" "ADs" {
  compartment_id = var.tenancy_ocid
}

# Get fault domains for selected AD
#data "oci_identity_fault_domains" "FDs" {
#    availability_domain = lookup(data.oci_identity_availability_domains.ADs.availability_domains[var.availability_domain],"name")
#    compartment_id      = var.CompartmentOCID
#}

# Tenancy
data "oci_identity_tenancy" "tenancy" {
  tenancy_id = var.tenancy_ocid
}

# Gets home region
# data "oci_identity_regions" "home-region" {
#  filter {
#    name   = "key"
#    values = ["${data.oci_identity_tenancy.tenancy.home_region_key}"]
#  }
#}

# get latest Oracle Linux image
data "oci_core_images" "oraclelinux" {
  compartment_id = var.CompartmentOCID
  operating_system = "Oracle Linux"
  operating_system_version = "7.9"
  filter {
    name = "display_name"
    values = ["^([a-zA-z]+)-([a-zA-z]+)-([\\.0-9]+)-([\\.0-9-]+)$"]
    regex = true
  }
}
