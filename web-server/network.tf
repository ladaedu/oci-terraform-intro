provider "oci" {
  tenancy_ocid     = var.tenancy_ocid
  user_ocid        = var.user_ocid
  fingerprint      = var.fingerprint
  private_key_path = var.private_key_path
  region           = var.region
}

resource "oci_core_virtual_network" "VCN" {
  cidr_block     = var.VCNCIDR
  compartment_id = var.CompartmentOCID
  display_name   = "Web VCN-${terraform.workspace}"
  dns_label      = "demo"
}

resource "oci_core_internet_gateway" "InetGW" {
  compartment_id = var.CompartmentOCID
  display_name   = "Internet GW -${terraform.workspace}"
  vcn_id         = oci_core_virtual_network.VCN.id
}

resource "oci_core_nat_gateway" "NATGateway" {
  compartment_id = var.CompartmentOCID
  vcn_id         = oci_core_virtual_network.VCN.id
  display_name   = "NAT Gateway-${terraform.workspace}"
}

