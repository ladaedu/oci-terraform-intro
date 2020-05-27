resource "oci_core_route_table" "PublicRoutingTable" {
  compartment_id = var.CompartmentOCID
  vcn_id         = oci_core_virtual_network.VCN.id
  display_name   = "Public Routing Table-${terraform.workspace}"

  route_rules {
    destination       = "0.0.0.0/0"
    network_entity_id = oci_core_internet_gateway.InetGW.id
  }
}

# Bastion SecList - Public Internet
resource "oci_core_security_list" "BastionSubnetSeclist" {
  compartment_id = var.CompartmentOCID
  display_name   = "Bastion Subnet Seclist-${terraform.workspace}"
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

resource "oci_core_subnet" "BastionSubnet" {
  count               = min(var.BastionVMCount, 3)
  availability_domain = data.oci_identity_availability_domains.ADs.availability_domains[count.index]["name"]
  cidr_block          = var.BastionSubnetCIDRs[count.index]
  display_name        = "Bastion Subnet-${count.index}-${terraform.workspace}"
  dns_label           = "bastion${count.index}"
  compartment_id      = var.CompartmentOCID
  vcn_id              = oci_core_virtual_network.VCN.id
  route_table_id      = oci_core_route_table.PublicRoutingTable.id
  security_list_ids   = [oci_core_security_list.BastionSubnetSeclist.id]
  dhcp_options_id     = oci_core_virtual_network.VCN.default_dhcp_options_id
}

