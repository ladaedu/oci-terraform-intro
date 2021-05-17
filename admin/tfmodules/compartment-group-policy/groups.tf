module "group_and_policy" {
  source               = "../terraform-oci-iam/modules/iam-group"
  #source                = "oracle-terraform-modules/iam/oci//modules/iam-group"
  #version               = "2.0.0"
  tenancy_ocid          = var.tenancy_ocid
  group_name            = "${var.name}.grp"
  group_description     = "${var.name} group"
  user_ids              = []
  policy_compartment_id = var.tenancy_ocid
  policy_name           = "${var.name}.pl"
  policy_description    = "${var.name} policy"
  policy_statements = [
    "ALLOW GROUP ${module.group_and_policy.group_name} to manage all-resources IN COMPARTMENT ${module.compartment.compartment_name}",
    "ALLOW GROUP ${module.group_and_policy.group_name} to read all-resources IN TENANCY",
  ]
}

