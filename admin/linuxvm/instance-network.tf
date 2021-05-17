resource "oci_core_route_table" "PublicRoutingTable" {
  compartment_id = var.CompartmentOCID
  vcn_id         = oci_core_virtual_network.VCN.id
  display_name   = "Public Routing Table-${terraform.workspace}"

  route_rules {
    destination       = "0.0.0.0/0"
    network_entity_id = oci_core_internet_gateway.InetGW.id
  }
}

# LinuxVM SecList - Public Internet
resource "oci_core_security_list" "LinuxVMSubnetSeclist" {
  compartment_id = var.CompartmentOCID
  display_name   = "LinuxVM Subnet Seclist-${terraform.workspace}"
  vcn_id         = oci_core_virtual_network.VCN.id

  egress_security_rules {
    protocol    = "6"
    destination = "0.0.0.0/0"
  }

  ingress_security_rules {
    # ssh
    tcp_options {
      max = 22
      min = 22
    }

    protocol = "6"
    source   = "0.0.0.0/0"
  }
}

resource "oci_core_subnet" "LinuxVMSubnet" {
  cidr_block          = var.LinuxVMSubnetCIDR
  display_name        = "LinuxVM Subnet-${terraform.workspace}"
  dns_label           = "vm"
  compartment_id      = var.CompartmentOCID
  vcn_id              = oci_core_virtual_network.VCN.id
  route_table_id      = oci_core_route_table.PublicRoutingTable.id
  security_list_ids   = [oci_core_security_list.LinuxVMSubnetSeclist.id]
  dhcp_options_id     = oci_core_virtual_network.VCN.default_dhcp_options_id
}

