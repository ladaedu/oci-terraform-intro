module "group_student1" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_name            = "student1.grp"
  group_create          = false
  policy_create         = false
  user_count            = 1
  user_ids              = [ "${module.student1.user_id}" ]
}

module "group_student2" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_name            = "student2.grp"
  group_create          = false
  policy_create         = false
  user_count            = 1
  user_ids              = [ "${module.student2.user_id}" ]
}

module "group_student3" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_name            = "student3.grp"
  group_create          = false
  policy_create         = false
  user_count            = 1
  user_ids              = [ "${module.student3.user_id}" ]
}

module "group_student4" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_name            = "student4.grp"
  group_create          = false
  policy_create         = false
  user_count            = 1
  user_ids              = [ "${module.student4.user_id}" ]
}

module "group_student5" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_name            = "student5.grp"
  group_create          = false
  policy_create         = false
  user_count            = 1
  user_ids              = [ "${module.student5.user_id}" ]
}

module "group_student6" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_name            = "student6.grp"
  group_create          = false
  policy_create         = false
  user_count            = 1
  user_ids              = [ "${module.student6.user_id}" ]
}

