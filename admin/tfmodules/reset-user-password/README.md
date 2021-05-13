# Reset user password

This [Terraform module](https://www.terraform.io/docs/modules/index.html) resets user password.

Note: output added by Ladislav.Dobias@oracle.com


```hcl
module "reset_user_someone" {
    source   = "./modules/reset-user-password"
    username = "some-user"
}
```

Note the following parameter:

Argument | Description
--- | ---
username | (Required) username


