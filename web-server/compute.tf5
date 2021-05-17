resource "oci_core_instance" "WebServer" {
  count               = var.WebVMCount
  availability_domain = data.oci_identity_availability_domains.ADs.availability_domains[count.index % 3]["name"]

  ## If there are more instances than fault domains, terraform does automatic modulo
  #fault_domain        = lookup(data.oci_identity_fault_domains.FDs.fault_domains[count.index],"name")
  compartment_id = var.CompartmentOCID
  display_name   = "webServer${count.index}-${terraform.workspace}"

  source_details {
    source_type = "image"
    source_id   = var.InstanceImageOCID[var.region]
  }

  shape     = var.TestServerShape
  create_vnic_details {
    hostname_label = "web${count.index}"
    subnet_id = oci_core_subnet.PrivateSubnet.id
    assign_public_ip = false
  }

  metadata = {
    ssh_authorized_keys = file(var.ssh_public_key)
    user_data           = base64encode(file(var.WebServerBootStrap))
  }
  provisioner "file" {
    source      = "userdata/hello-plain-text.conf"
    destination = "nginx-demo.conf"
    connection {
      bastion_host        = oci_core_instance.Bastion[0].public_ip
      bastion_user        = "opc"
      bastion_private_key = file(var.ssh_private_key)
      type                = "ssh"
      host                = self.private_ip
      user                = "opc"
      private_key         = file(var.ssh_private_key)
    }
  }
}

