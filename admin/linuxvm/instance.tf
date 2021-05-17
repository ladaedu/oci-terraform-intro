// LinuxVM
// running in Public subnet
// accessible with ssh
resource "oci_core_instance" "LinuxVM" {
  availability_domain = data.oci_identity_availability_domains.ADs.availability_domains[var.availability_domain]["name"]
  compartment_id      = var.CompartmentOCID
  display_name        = "linuxvm-${terraform.workspace}"

  source_details {
    source_type = "image"
#    source_id   = var.InstanceImageOCID[var.region]
    source_id   = data.oci_core_images.oraclelinux.images.0.id
  }

  shape     = "VM.Standard.E3.Flex"
  shape_config {
        ocpus = 1
        memory_in_gbs = 1
        baseline_ocpu_utilization = "BASELINE_1_8"
  }
  create_vnic_details {
    hostname_label      = "linuxvm"
    subnet_id = oci_core_subnet.LinuxVMSubnet.id
  }

  metadata = {
    ssh_authorized_keys = file(var.ssh_public_key)
    user_data           = base64encode(file(var.LinuxVMServerBootStrap))
  }
}

