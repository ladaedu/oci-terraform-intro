module "student1" {
  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
  tenancy_ocid          = "${var.tenancy_ocid}"
  user_name             = "${var.student1_name}"
  user_description      = "student account"
}
module "student2" {
  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
  tenancy_ocid          = "${var.tenancy_ocid}"
  user_name             = "${var.student2_name}"
  user_description      = "student account"
}
module "student3" {
  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
  tenancy_ocid          = "${var.tenancy_ocid}"
  user_name             = "${var.student3_name}"
  user_description      = "student account"
}
module "student4" {
  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
  tenancy_ocid          = "${var.tenancy_ocid}"
  user_name             = "${var.student4_name}"
  user_description      = "student account"
}
#module "student5" {
#  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
#  tenancy_ocid          = "${var.tenancy_ocid}"
#  user_name             = "${var.student5_name}"
#  user_description      = "student account"
#}
#module "student6" {
#  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
#  tenancy_ocid          = "${var.tenancy_ocid}"
#  user_name             = "${var.student6_name}"
#  user_description      = "student account"
#}
#module "student7" {
#  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
#  tenancy_ocid          = "${var.tenancy_ocid}"
#  user_name             = "${var.student7_name}"
#  user_description      = "student account"
#}
#module "student8" {
#  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
#  tenancy_ocid          = "${var.tenancy_ocid}"
#  user_name             = "${var.student8_name}"
#  user_description      = "student account"
#}
module "student9" {
  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
  tenancy_ocid          = "${var.tenancy_ocid}"
  user_name             = "${var.student9_name}"
  user_description      = "student account"
}
module "student10" {
  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
  tenancy_ocid          = "${var.tenancy_ocid}"
  user_name             = "${var.student10_name}"
  user_description      = "student account"
}
