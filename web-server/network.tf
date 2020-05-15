resource "oci_core_vcn" "VCN" {
  cidr_block     = var.VCNCIDR
  compartment_id = var.CompartmentOCID
  display_name   = "Web VCN-${terraform.workspace}"
  dns_label      = "demo"
}

resource "oci_core_internet_gateway" "InetGW" {
  compartment_id = var.CompartmentOCID
  display_name   = "Internet GW -${terraform.workspace}"
  vcn_id         = oci_core_vcn.VCN.id
}

resource "oci_core_nat_gateway" "NATGateway" {
  compartment_id = var.CompartmentOCID
  vcn_id         = oci_core_vcn.VCN.id
  display_name   = "NAT Gateway-${terraform.workspace}"
}
