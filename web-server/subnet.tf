resource "oci_core_route_table" "PrivateRoutingTable" {
  compartment_id = var.CompartmentOCID
  vcn_id         = oci_core_virtual_network.VCN.id
  display_name   = "Private Routing Table-${terraform.workspace}"

  route_rules {
    destination       = "0.0.0.0/0"
    network_entity_id = oci_core_nat_gateway.NATGateway.id
  }
}

# Private SecList - Private Network
resource "oci_core_security_list" "PrivateSubnetSeclist" {
  compartment_id = var.CompartmentOCID
  display_name   = "Private Subnet Seclist-${terraform.workspace}"
  vcn_id         = oci_core_virtual_network.VCN.id

  egress_security_rules {
    protocol    = "6"
    destination = "0.0.0.0/0"
  }

  ingress_security_rules {
    protocol = "all"
    source   = var.VCNCIDR
  }

  ingress_security_rules {
    # ssh
    tcp_options {
      max = 22
      min = 22
    }
    protocol = "6"
    source   = var.VCNCIDR
  }
  ingress_security_rules {
    # http
    tcp_options {
      max = 80
      min = 80
    }
    protocol = "6"
    source   = var.VCNCIDR
  }
  ingress_security_rules {
    # https
    tcp_options {
      max = 443
      min = 443
    }
    protocol = "6"
    source   = var.VCNCIDR
  }
}

resource "oci_core_subnet" "PrivateSubnet" {
  cidr_block                 = var.PrivateSubnetCIDR
  display_name               = "Private Subnet-${terraform.workspace}"
  dns_label                  = "private"
  compartment_id             = var.CompartmentOCID
  vcn_id                     = oci_core_virtual_network.VCN.id
  route_table_id             = oci_core_route_table.PrivateRoutingTable.id
  security_list_ids          = [oci_core_security_list.PrivateSubnetSeclist.id]
  dhcp_options_id            = oci_core_virtual_network.VCN.default_dhcp_options_id
  prohibit_public_ip_on_vnic = "true"
}

