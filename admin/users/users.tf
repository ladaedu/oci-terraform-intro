module "iam_users" {
  source                =  "../tfmodules/terraform-oci-iam/modules/iam-user"
  tenancy_ocid          = var.tenancy_ocid
  users           = [{ 
      name        = var.student_name
      description = "student account"
      email       = var.student_name
    },{
      name        = var.student1_name
      description = "student1 account"
      email       = var.student1_name
    },{
      name        = var.student2_name
      description = "student2 account"
      email       = var.student2_name
    },{
      name        = var.student3_name
      description = "student3 account"
      email       = var.student3_name
    },{
      name        = var.student4_name
      description = "student4 account"
      email       = var.student4_name
    },{
      name        = var.student5_name
      description = "student5 account"
      email       = var.student5_name
    },{
      name        = var.student6_name
      description = "student6 account"
      email       = var.student6_name
    },{
      name        = var.student7_name
      description = "student7 account"
      email       = var.student7_name
#    },{
#      name        = var.student8_name
#      description = "student8 account"
#      email       = var.student8_name
    }]
}

module "reset_password_student" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,0)
    name         = element(module.iam_users.names,0)
    email        = element(module.iam_users.names,0)
    tenancy_name = var.tenancy_name
    group        = module.group_student.group_name
}
output "Email-student" {
    value       = module.reset_password_student.email_suggestion
    description = "Email suggestion"
}

module "reset_password_student1" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,1)
    name         = element(module.iam_users.names,1)
    email        = element(module.iam_users.names,1)
    tenancy_name = var.tenancy_name
    group        = module.group_student1.group_name
}
output "Email-student1" {
    value       = module.reset_password_student1.email_suggestion
    description = "Email suggestion"
}

module "reset_password_student2" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,2)
    name         = element(module.iam_users.names,2)
    email        = element(module.iam_users.names,2)
    tenancy_name = var.tenancy_name
    group        = module.group_student2.group_name
}
output "Email-student2" {
    value       = module.reset_password_student2.email_suggestion
    description = "Email suggestion"
}

module "reset_password_student3" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,3)
    name         = element(module.iam_users.names,3)
    email        = element(module.iam_users.names,3)
    tenancy_name = var.tenancy_name
    group        = module.group_student3.group_name
}
output "Email-student3" {
    value       = module.reset_password_student3.email_suggestion
    description = "Email suggestion"
}

module "reset_password_student4" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,4)
    name         = element(module.iam_users.names,4)
    email        = element(module.iam_users.names,4)
    tenancy_name = var.tenancy_name
    group        = module.group_student4.group_name
}
output "Email-student4" {
    value       = module.reset_password_student4.email_suggestion
    description = "Email suggestion"
}

module "reset_password_student5" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,5)
    name         = element(module.iam_users.names,5)
    email        = element(module.iam_users.names,5)
    tenancy_name = var.tenancy_name
    group        = module.group_student5.group_name
}
output "Email-student5" {
    value       = module.reset_password_student5.email_suggestion
    description = "Email suggestion"
}

module "reset_password_student6" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,6)
    name         = element(module.iam_users.names,6)
    email        = element(module.iam_users.names,6)
    tenancy_name = var.tenancy_name
    group        = module.group_student6.group_name
}
output "Email-student6" {
    value       = module.reset_password_student6.email_suggestion
    description = "Email suggestion"
}

module "reset_password_student7" {
    source       = "../tfmodules/reset-user-password"
    user_id      = element(module.iam_users.user_id,7)
    name         = element(module.iam_users.names,7)
    email        = element(module.iam_users.names,7)
    tenancy_name = var.tenancy_name
    group        = module.group_student7.group_name
}
output "Email-student7" {
    value       = module.reset_password_student7.email_suggestion
    description = "Email suggestion"
}

#module "reset_password_student8" {
#    source       = "../tfmodules/reset-user-password"
#    user_id      = element(module.iam_users.user_id,8)
#    name         = element(module.iam_users.names,8)
#    email        = element(module.iam_users.names,8)
#    tenancy_name = var.tenancy_name
#    group        = module.group_student8.group_name
#}
#output "Email-student8" {
#    value       = module.reset_password_student8.email_suggestion
#    description = "Email suggestion"
#}

