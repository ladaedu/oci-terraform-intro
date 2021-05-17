resource "oci_core_virtual_network" "VCN" {
  cidr_block     = var.VCNCIDR
  compartment_id = var.CompartmentOCID
  display_name   = "Web VCN-${terraform.workspace}"
  dns_label      = "linux"
}

resource "oci_core_internet_gateway" "InetGW" {
  compartment_id = var.CompartmentOCID
  display_name   = "Internet GW-${terraform.workspace}"
  vcn_id         = oci_core_virtual_network.VCN.id
}
