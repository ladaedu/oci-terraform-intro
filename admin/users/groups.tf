module "group_student" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student.grp"
  group_description     = "student.grp group"
  user_ids              = [element(module.iam_users.user_id,0)]
}

module "group_student1" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student1.grp"
  group_description     = "student1.grp group"
  user_ids              = [element(module.iam_users.user_id,1)]
}

module "group_student2" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student2.grp"
  group_description     = "student2.grp group"
  user_ids              = [element(module.iam_users.user_id,2)]
}

module "group_student3" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student3.grp"
  group_description     = "student3.grp group"
  user_ids              = [element(module.iam_users.user_id,3)]
}

module "group_student4" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student4.grp"
  group_description     = "student4.grp group"
  user_ids              = [element(module.iam_users.user_id,4)]
}

module "group_student5" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student5.grp"
  group_description     = "student5.grp group"
  user_ids              = [element(module.iam_users.user_id,5)]
}

module "group_student6" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student6.grp"
  group_description     = "student6.grp group"
  user_ids              = [element(module.iam_users.user_id,6)]
}

module "group_student7" {
  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
  tenancy_ocid          = "${var.tenancy_ocid}"
  group_create          = false
  group_name            = "student7.grp"
  group_description     = "student7.grp group"
  user_ids              = [element(module.iam_users.user_id,7)]
}

#module "group_student8" {
#  source                  =  "../tfmodules/terraform-oci-iam/modules/iam-group"
#  tenancy_ocid          = "${var.tenancy_ocid}"
#  group_create          = false
#  group_name            = "student8.grp"
#  group_description     = "student8.grp group"
#  user_ids              = [element(module.iam_users.user_id,8)]
#}

