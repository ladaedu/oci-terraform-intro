module "compartment" {
  #source                  = "oracle-terraform-modules/iam/oci//modules/iam-compartment"
  #version                 = "1.0.2"
  source                  =  "../terraform-oci-iam/modules/iam-compartment"
  tenancy_ocid            = "${var.tenancy_ocid}"
  compartment_name        = "${var.name}"
  compartment_description = "${var.name} compartment"
  compartment_create      = "${var.compartment_create}"
}

