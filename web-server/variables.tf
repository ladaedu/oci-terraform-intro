variable "tenancy_ocid" {
  default = "ocid1.tenancy.oc1..aaaaaaaah3b24zkkewpfygiw3rekqn3idilrt2qrjzkcdxbu5yhqpet4ox4a"
}

variable "user_ocid" {
  default = "ocid1.user.oc1..aaaaaaaaq623ujht4oif2p6qkmoo5xm4i44eizvvxwlybgri2cqoqukqjela"
}

variable "fingerprint" {
  default = "9a:80:84:d0:dd:92:e7:66:ed:8d:90:84:e5:45:ea:27"
}

variable "private_key_path" {
  default = "/home/vit/Downloads/vitek.skrhak-06-07-08-49.pem"
}

variable "ssh_public_key" {
  default = "/home/vit/.ssh/id_rsa.pub"
}

variable "ssh_private_key" {
  default = "/home/vit/.ssh/id_rsa"
}

variable "region" {
  default = "eu-frankfurt-1"
}

/* Availability domain can be 0, 1 or 2 - use the one that has free resources */
variable "availability_domain" {
  default = 1
}

variable "TestServerShape" {
  default = "VM.Standard.E4.Flex"
}

variable "InstanceImageOCID" {
  type = map(string)

  # TASK: set machine image for your environment, get it e.g. using command:
  #     oci compute image list --compartment-id "your compartment OCID" |less
  # and search for image with name Linux-7.6-2019, like written below (with different date).
  # TIP: the variable map can be (re-)defined also in env-vars file.
  default = {
    // Oracle-Linux-7.9-2021.05.12-0
    eu-frankfurt-1 = "ocid1.image.oc1.eu-frankfurt-1.aaaaaaaaprt6uk32tylin3owcddyllao3uthmo7vheqepeybvjj6to7xkdgq",

  }
}

####################################################################################################
variable "WebServerBootStrap" {
  default = "./userdata/webServer"
}

variable "BastionServerBootStrap" {
  default = "./userdata/bastionServer"
}

variable "WebVMCount" {
  default = 1
}

variable "BastionVMCount" {
  default = 2
}

####################################################################################################
variable "VCNCIDR" {
  default = "10.0.0.0/16"
}

variable "PrivateSubnetCIDR" {
  default = "10.0.0.0/24"
}

variable "BastionSubnetCIDRs" {
  # same count as ADs in region
  default = ["10.0.100.0/28", "10.0.100.16/28", "10.0.100.32/28"]
}

variable "CompartmentOCID" {
  default = "ocid1.compartment.oc1..aaaaaaaaicbtnwruibyhgerp77cep5i6gnn6o6ouz74yyok4dgu2gsjhslga"
}

