variable "tenancy_ocid" {
}

variable "user_ocid" {
}

variable "fingerprint" {
}

variable "private_key_path" {
}

variable "ssh_public_key" {
}

variable "ssh_private_key" {
}

variable "region" {
  default = "eu-frankfurt-1"
}

/* Availability domain can be 0, 1 or 2 - use the one that has free resources */
#variable "availability_domain" {
#  default = 1
#}

variable "TestServerShape" {
  default = "VM.Standard2.1"
}

variable "InstanceImageOCID" {
  type = map(string)

  # TASK: set machine image for your environment, get it e.g. using command:
  #     oci compute image list --compartment-id "your compartment OCID" |less
  # and search for image with name Linux-7.6-2019, like written below (with different date).
  # TIP: the variable map can be (re-)defined also in env-vars file.
  default = {
	eu-frankfurt-1 = "ocid1.image.oc1.eu-frankfurt-1.aaaaaaaavz6p7tyrczcwd5uvq6x2wqkbwcrjjbuohbjomtzv32k5bq24rsha"
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
  default = 1
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
  default = "ocid1.compartment.oc1..aaaaaaaa5ho3ftokbmcdpn34mxhjcmuear2tnwyx54sxy6qpcqtaiwqucqlq"
}

