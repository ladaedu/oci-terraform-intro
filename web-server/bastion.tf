// Bastion
// running in Public subnet
// accessible with ssh
resource "oci_core_instance" "Bastion" {
  count               = min(var.BastionVMCount,2)
  availability_domain = data.oci_identity_availability_domains.ADs.availability_domains[count.index % 2]["name"]
  compartment_id      = var.CompartmentOCID
  display_name        = "bastion${count.index}-${terraform.workspace}"
  hostname_label      = "bastion${count.index}"

  source_details {
    source_type = "image"
    source_id   = var.InstanceImageOCID[var.region]
  }

  shape     = var.TestServerShape
  subnet_id = oci_core_subnet.BastionSubnet[count.index % 2].id

  metadata = {
    ssh_authorized_keys = file(var.ssh_public_key)
    user_data           = base64encode(file(var.BastionServerBootStrap))
  }
}

