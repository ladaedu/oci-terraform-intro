---
title: Introduction to Terraform using OCI
author: Vít Kotačka, Ladislav Dobiáš
...


# OCI & Terraform & Terratest

## Agenda

- Login to OCI console
- Prereqisities
- Setup OCI API key
- Today's Goals with Terraform
- Terraform - setup
- Terraform - first test
- Terraform - steps
- Terratest


## Login to OCI console

- OCI - Oracle Cloud Infrastructure
- console URL: [https://console.us-ashburn-1.oraclecloud.com/?tenant=czechedu](https://console.us-ashburn-1.oraclecloud.com/?tenant=czechedu)
    - user: email
    - password: generated, need to be changed on first login

- authorization:
    - every student is in one of `student*` groups
    - every group `student*` can:
        - do all in their compartment (same name as the group)
        - read all resources
        - (these policies would be too open for real production environment)

- quota:
    - important:
        - virtual machine shapes: 3x 15 VM.Standard2.1 (1 in each AD)
        - loadbalancers: 15 in region


## Prereqisities

This you should have installed (can be in docker, too):

- git
- openssl
- terraform, e.g.:

    ```
    wget https://releases.hashicorp.com/terraform/0.11.14/terraform_0.11.14_linux_amd64.zip
    unzip terraform_0.11.14_linux_amd64.zip
    mv terraform ~/bin
    ```
- go 1.11+

Optional (recommended - for OCI API key setup,...):

- oci cli - install OCI cli: [https://docs.cloud.oracle.com/iaas/Content/API/SDKDocs/cliinstall.htm](https://docs.cloud.oracle.com/iaas/Content/API/SDKDocs/cliinstall.htm)
    ```
    bash -c "$(curl -L https://raw.githubusercontent.com/oracle/oci-cli/master/scripts/install/install.sh)"
    ```


## Setup OCI API key

- 2 possibilities:
    - manual way:
        - see [https://docs.cloud.oracle.com/iaas/Content/API/Concepts/apisigningkey.htm](https://docs.cloud.oracle.com/iaas/Content/API/Concepts/apisigningkey.htm)
    - using OCI cli - generate OCI API key to ~/.oci:

          ```
          oci setup config
          ```

          - provide:
              - user OCID - get it from UI console
              - tenancy OCI (also from UI): `ocid1.tenancy.oc1..aaaaaaaag5ufgrzvunkiqspyc2ithwyyhlsl4ydvpiyj3zebc6mbn5kbs7oa`
              - region: `us-ashburn-1`
        - simple tests:

            ```
            oci iam region list
            oci compute image list --compartment-id ocid1.compartment.oc1..aaaaaaaamixpp5zksaik4p5e2itnzyfasnbhrnhb2zvim6vxy6kyfwbabkya
            ```


## Today's Goals with Terraform - simple webserver with bastion

Deployment diagram - simple:
![Simple web server](TF-simple.png)

This would be achieved at the step #6.


## Today's Goals with Terraform - more webservers with bastion and load balancer

Deployment diagram - with LB:
![Web server with LB](TF-loadbalancer.png)

This would be achieved at the last step.


## Terraform - setup

- get sources:

    ```
    git clone https://github.com/ladaedu/oci-terraform-intro
    cd oci-terraform-intro/web-server
    ```

- edit variables in env-vars.example that are not commented out, copy it first:

    ```
    cp env-vars.example env-vars
    ```

    - use data from `~/.oci/config`

- source it:

    ```
    . env-vars
    ```

## Terraform - first test

- list current `*.tf` files:

    ```
    ls *.tf
    ```

    - output (recommened to look inside the files): `network.tf  variables.tf`

- init terraform (download providers, modules,...):

    ```
    alias tf=terraform
    tf init
    ```

- plan

    ```
    tf plan
    ```

- apply

    ```
    tf apply
    ```


## Terraform - steps overview

0. VCN, gateways
1. Datasources - ADs, Tenancy
2. Bastion - network: routing table, seclist, subnet
3. Bastion VM
4. Private Subnet for Web servers - network: routing table, seclist, subnet
5. Web server
6. Outputs - IP addresses
7. Load balancer + add some web servers

## Terraform - next step

- rename next steps TF file, e.g. `*.tf1` to `*.tf`:

    ```
    orig=$(echo *1);link=${orig%?};echo ln -s $orig $link
    ```
- for other steps, replace `1` with next numbers

- plan

    ```
    tf plan
    ```

- apply

    ```
    tf apply
    ```
- check what was created in UI console

## Terratest

Terratest will create its own environment, so destroy your environment first, to avoid problems with quota.

- destroy the deployment:

    ```
    tf destroy
    ```

- run terratest:

    ```
    cd terratest
    go test -v -run TestTerraform
    ```
## Thank you

Questions?

# Backup slides

## Terraform graph

- generate graph - using Graphviz:

    ```
    tf graph
    ```
- generate graph with colors:
    ```
    ./tf-graph.sh
    ```

## References

- Terraform:
    - download: [https://www.terraform.io/downloads.html](https://www.terraform.io/downloads.html)
    - OCI provider docs: [https://www.terraform.io/docs/providers/oci/](https://www.terraform.io/docs/providers/oci/)
- OCI:
    - [Overview of Networking](https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm)
    - [Regions and Availability Domains](https://docs.cloud.oracle.com/iaas/Content/General/Concepts/regions.htm)
    - [Regional Subnets](https://docs.cloud.oracle.com/iaas/releasenotes/changes/08c01d20-c829-47f2-8d54-9e9958f50ba8/)
    - [Overview of Load Balancing](https://docs.cloud.oracle.com/iaas/Content/Balance/Concepts/balanceoverview.htm)

