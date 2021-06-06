variable "tenancy_ocid" {
  default = "ocid1.tenancy.oc1..aaaaaaaah3b24zkkewpfygiw3rekqn3idilrt2qrjzkcdxbu5yhqpet4ox4a"
}

variable "user_ocid" {
  default = "ocid1.user.oc1..aaaaaaaawcnv6rupe65khoaixjfozzazx45v74z7piphbley4ogfdsikntgq"
}

variable "fingerprint" {
  default = "af:08:97:b1:11:d2:b1:96:79:fa:e7:19:d9:1e:c0:2c"
}

variable "private_key_path" {
  default = "/home/jirka/Downloads/jiriklepl-05-18-14-32.pem"
}

variable "ssh_public_key" {
}

variable "ssh_private_key" {
}

variable "region" {
  default = "eu-frankfurt-1"
}

/* Availability domain can be 0, 1 or 2 - use the one that has free resources */
variable "availability_domain" {
  default = 1
}

variable "TestServerShape" {
  default = "VM.Standard2.1"
}

variable "InstanceImageOCID" {
  type = map(string)

  default = {
    us-frankfurt-1 = "ocid1.image.oc1.eu-frankfurt-1.aaaaaaaad7zlitmahc5w5m7hv47jiwooiy7vqethdsl4vexdc3sfcrzkvi3q"
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
  default = 3
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
  default = "ocid1.image.oc1.eu-frankfurt-1.aaaaaaaavz6p7tyrczcwd5uvq6x2wqkbwcrjjbuohbjomtzv32k5bq24rsha"
}

