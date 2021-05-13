module "compartment" {
  source                  = "../terraform-oci-iam/modules/iam-compartment"
  #source                  = "oracle-terraform-modules/iam/oci//modules/iam-compartment"
  #version                 = "2.0.0"
  tenancy_ocid            = var.tenancy_ocid
  compartment_name        = var.name
  compartment_description = "${var.name} compartment"
}

