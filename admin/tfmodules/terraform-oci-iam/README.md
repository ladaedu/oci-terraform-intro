# Oracle Cloud Infrastructure IAM Terraform Module

The Oracle Cloud Infrastructure (OCI) IAM module  allows you to create identity and access management (IAM) groups so you can control who has access to your OCI cloud resources. IAM allows you to control the level of access for each user group, and limits access to specific OCI resources.

This readme provides an overview of IAM, the component parts of IAM, and provides examples scenarios to help you understand how to work with the IAM Terrform module.

## Prerequisites

1. [Download and install Terraform](https://www.terraform.io/downloads.html) (v0.10.3 or later)
2. [Download and install the OCI Terraform Provider](https://github.com/oracle/terraform-provider-oci) (v2.1.10 or later)
3. Export OCI credentials using guidance at [Export Credentials](https://github.com/oracle/terraform-provider-oci#export-credentials).

## IAM Components

Following are descriptions of the main components of IAM. For more information, and to see example scenarios, see [Overview of IAM](https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).

### Compartment

A compartment is a collection of related resources that helps you organize and isolate your OCI cloud resources. Use compartments to isolate reousrces so you can measure usage, control billing, as well as control user access using IAM policies. Compartments also allow you to separate business units. For more information, see [Setting Up Your Tenancy](https://docs.cloud.oracle.com/iaas/Content/GSG/Concepts/settinguptenancy.htm).

### User

An individual employee or system that needs to manage or use your company's Oracle Cloud Infrastructure resources. Users might need to launch instances, manage remote disks, work with your virtual cloud network, etc. End users of your application are not typically IAM users. Users have one or more IAM credentials see [User Credentials](https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/usercredentials.htm).

### Group and Dynamic Group

A collection of users who all need the same type of access to a particular set of resources or compartment. For more information, see [Managing Groups](https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managinggroups.htm). See also [Managing Dynamic Groups](https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingdynamicgroups.htm).

### Policy

Policies specify who can access which resources, and which actions they can take on each resource. Access is granted at the group and compartment level, which means you can write a policy that gives a group a specific type of access within a specific compartment, and give them different access to a different compartment.

You can also grant access at the tenancy scope. If you give a group access to the tenancy, the group automatically gets the same type of access to all the compartments inside the tenancy.

For more information about using IAM polcies, see [Getting Started with Polcies](https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policygetstarted.htm). See also [Managing Policies](https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingpolicies.htm) and [Policy Syntax](https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/policysyntax.htm).

## How to use this module

This module has the following folder structure:

* [modules](https://bitbucket.aka.lgl.grungy.us/projects/TFS/repos/terraform-oci-iam/browse/modules): This folder includes four sub modules for creating IAM resource in OCI.
* [example](https://bitbucket.aka.lgl.grungy.us/projects/TFS/repos/terraform-oci-iam/browse/example): This folder contains an example of how to use the module.

## Usage
`iam-compartment`:

```hcl
module "iam_compartment" {
  source                    = "./modules/iam-compartment"
  tenancy_ocid              = "${var.tenancy_ocid}"
  compartment_name          = "tf_example_compartment"
  compartment_description   = "compartment created by terraform"
}
```

```hcl
module "iam_user" {
  source                    = "../modules/iam-user"
  tenancy_ocid              = "${var.tenancy_ocid}"
  user_name                 = "tf_example_user@oracle.com"
  user_description          = "user created by terraform"
}
```

```hcl
module "iam_group" {
  source                    = "../modules/iam-group"
  tenancy_ocid              = "${var.tenancy_ocid}"
  group_name                = "tf_example_group"
  group_description         = "group created by terraform"
  add_user_to_group         = true
  user_count                = 2
  user_ids                  = ["${module.iam_user1.user_id}", "${module.iam_user2.user_id}"]
  policy_compartment_id     = "ocid1.compartment.oc1..xxxxxxxxxxxxxx"
  policy_name               = "tf-example-policy"
  policy_description        = "policy created by terraform"
  policy_statements         = ["Allow group tf_example_group to read instances in compartment tf_example_compartment", "Allow group tf_example_group to inspect instances in compartment tf_example_compartment"]
}
```

```hcl
module "iam_dynamic_group" {
  source                    = "../modules/iam-dynamic-group"
  tenancy_ocid              = "${var.tenancy_ocid}"
  dynamic_group_name        = "tf_example_dynamic_group"
  dynamic_group_description = "dynamic group created by terraform"
  dynamic_group_rule        = "instance.compartment.id = ocid1.compartment.oc1..xxxxxxxxxxxxxx"
  add_user_to_group         = true
  user_count                = 1
  user_ids                  = ["${module.iam_user.user_id}"]
  policy_create             = false
}
```

**Following are arguments available to the IAM module:**

Argument | Description
--- | ---
tenancy_ocid | Unique identifier (OCID) of the tenancy.
compartment_name | The name you assign to the compartment. The name must be unique across all compartments in a given tenancy.
compartment_description | Description of the compartment. You can edit the description.
compartment_create | Specifies whether the module should create a compartment. If true, the compartment will be managed by the module. In this case, the user must have permission to create the compartment. If false, compartment data will be returned about any existing compartments. If no compartment is found, an empty string is returned for the compartment ID.
group_name | The name you assign to the IAM group when created. The name must be unique across all compartments in the tenancy.
group_description | Description of the IAM group. The description is editable.
group_create | Specifies whether the module should create the group. If true, the user must have permission to create a group. If false, group data is returned for all existing groups. If no groups are found, an empty string is returned for the group ID.
dynamic_group_name | Name given to the dynamic group during creation. The name must be unique across all compartments in the tenancy.
dynamic_group_description | Description of the dynamic group. The description is editable.
dynamic_group_create | Specifies whether the module should create a dynamic group. If true, the user must have permission to create a dymaic group. If false, data is returned for any existing dynami groups, and an empty string is returned for the dymaic group ID.
dynamic_group_rule | Define a matching rule or a set of matching rules to define the group members.
user_name | The name you assign to the user during creation. The name must be unique across all compartments in the tenancy.
user_description | Description of the user. The description is editable.
user_create | Specifies whether the module should create a user. If true, the user must have permissions to create a user. If false, user data will be returned about existing users. If no users are found, an empty string is returned for the user ID.
user_count | Number of user to be added to a group.
policy_name | The name you assign to the IAM policy.
policy_description | Description of the IAM policy. The description is editable.
policy_statements | The policy definition expressed as one or more policy statements.
policy_create | Specifices whether the modules should create the IAM policy.

## Contributing

This project is open source. Oracle appreciates any contributions that are made by the open source community.
