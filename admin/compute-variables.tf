/* Availability domain can be 0, 1 or 2 - use the one that has free resources */
variable "availability_domain" {
  default = 2
}

#variable "TestServerShape" {
#  default = "VM.Standard1.1"
#}

variable "InstanceImageOCID" {
  type = "map"

  default = {
    // Oracle-Linux-7.6-2019.02.20-0
    us-phoenix-1 = "ocid1.image.oc1.phx.aaaaaaaacss7qgb6vhojblgcklnmcbchhei6wgqisqmdciu3l4spmroipghq"
    us-ashburn-1 = "ocid1.image.oc1.iad.aaaaaaaajnvrzitemn2k5gfkfq2hwfs4bid577u2jbrzla42wxo2qc77gwxa"
    uk-london-1  = "ocid1.image.oc1.uk-london-1.aaaaaaaarruepdlahln5fah4lvm7tsf4was3wdx75vfs6vljdke65imbqnhq"
  }
}

####################################################################################################
variable "TestServerBootStrap" {
  default = "compute-userdata"
}

variable "TestVMCount" {
  default = "1"
}
