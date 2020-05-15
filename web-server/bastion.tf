// Bastion
// running in Public subnet
// accessible with ssh
resource "oci_core_instance" "Bastion" {
  count = var.BastionVMCount
  availability_domain = lookup(data.oci_identity_availability_domains.ADs.availability_domains[count.index % length(data.oci_identity_availability_domains.ADs.availability_domains)],"name")
  compartment_id      = var.CompartmentOCID
  display_name        = "bastion${count.index}-${terraform.workspace}"
  hostname_label = "bastion${count.index}"

  source_details {
    source_type = "image"
    source_id   = var.InstanceImageOCID[var.region]
  }

  shape     = var.TestServerShape
  
  create_vnic_details {
    subnet_id = oci_core_subnet.BastionSubnet[count.index % length(oci_core_subnet.BastionSubnet)].id
  }

  metadata = {
    ssh_authorized_keys = file(var.ssh_public_key)
    user_data           = base64encode(file(var.BastionServerBootStrap))
  }
}
