/* Availability domain can be 0, 1 or 2 - use the one that has free resources */
variable "availability_domain" {
  default = 2
}

//variable "TestServerShape" {
//  default = "VM.Standard1.1"
//}

variable "InstanceImageOCID" {
  type = "map"

  default = {
    // Oracle-Linux-7.8-2020.04.17-0
    eu-frankfurt-1 = "ocid1.image.oc1.eu-frankfurt-1.aaaaaaaavz6p7tyrczcwd5uvq6x2wqkbwcrjjbuohbjomtzv32k5bq24rsha"
  }
}

####################################################################################################
variable "TestServerBootStrap" {
  default = "compute-userdata"
}

variable "TestVMCount" {
  default = "1"
}
