---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay Provider"
subcategory: ""
description: |-
  
---

# Rafay Provider

Organizations that have invested in building complex Terraform based "Infrastructure as Code" for cluster provisioning can seamlessly integrate and use the controller for workload related operations. 

There are examples in the [GitHub repo](https://github.com/RafaySystems/terraform-provider-rafay). 


---

## Terraform Resource List 

### Released 

| Resource                                  | Version Released |
| ----------------------------------------- | ---------------- |
| `rafay_mks_cluster`                       | 1.1.36           |
| `rafay_driver` (deprecated)               | 1.1.22           |
| `rafay_workflow_handler`                  | 1.1.43           |
| `rafay_environment`                       | 1.1.18           |
| `rafay_environment_template`              | 1.1.18           |
| `rafay_resource_template`                 | 1.1.18           |
| `rafay_static_resource`                   | 1.1.18           |
| `rafay_config_context`                    | 1.1.18           |
| `rafay_gke_cluster`                       | 1.1.18           |
| `rafay_aks_cluster_v3`                    | 1.1.11           |
| `rafay_tag_group`                         | 1.1.11           |
| `rafay_project_tags_association`          | 1.1.11           |
| `rafay_container_registry`                | 1.1.9            |
| `rafay_chargeback_group`                  | 1.1.9            |
| `rafay_chargeback_group_report`           | 1.1.9            |
| `rafay_chargeback_share`                  | 1.1.9            |
| `rafay_cloud_credentials_v3`              | 1.1.9            |
| `rafay_cluster_sharing`                   | 1.1.4            |
| `catalog`                           	    | 1.1.2            |
| `secret_group`                            | 1.1.2            |
| `secret_provider`                         | 1.1.2            |
| `mesh_profile`                            | 1.1.2            |
| `cluster_mesh_rule`                       | 1.1.2            |
| `cluster_mesh_policy`                     | 1.1.2            |
| `namespace_mesh_rule`                     | 1.1.2            |
| `namespace_mesh_policy`                   | 1.1.2            |
| `cost_profile`                            | 1.1.2            |
| `download_kubeconfig`                     | 1.1.2            |
| `access_apikey`                           | 1.1.1            |
| `opa_installation_profile`                | 1.1.1            |
| `cluster_network_policy`                  | 1.1.0            |  
| `cluster_network_policy_rule`             | 1.1.0            |  
| `network_policy_profile`                  | 1.1.0            |  
| `namespace_network_policy_rule`           | 1.1.0            | 
| `namespace_network_policy`                | 1.1.0            | 
| `opa_constraint`                          | 1.1.0            |  
| `opa_constraint_template`                 | 1.1.0            |  
| `secret_sealer`                           | 1.1.0            |  
| `user`                                    | 1.1.0            |
| `pipeline`                                | 1.1.0            |
| `blueprint` (with OPA support)            | 1.0.2            |
| `workload`                                | 1.0.0            |
| `workloadtemplate`                        | 1.0.0            |
| `aks_cluster_spec`                        | 1.0.0            |
| `eks_cluster_spec`                        | 1.0.0            |
| `opa_policy`                              | 1.0.0            |
| `group`                                   | 0.9.12           |
| `groupassociation`                        | 0.9.12           |
| `import_cluster`                          | 0.9.12           |
| `addon`                                   | 0.9.11           |
| `agent`                                   | 0.9.11           |
| `aks_cluster`                             | 0.9.11           |
| `blueprint`                               | 0.9.11           |
| `cloud_credential`                        | 0.9.11           |
| `cluster_override`                        | 0.9.11           |
| `eks_cluster`                             | 0.9.11           |
| `namespace`                               | 0.9.11           |
| `project`                                 | 0.9.11           |
| `repositories`                            | 0.9.11           |
| | |
 
### In Development 

- `infra_provisioner`


---

## Example Usage

```terraform
terraform {
  required_providers {
    rafay = {
      version = ">= 0.1"
      source  = "RafaySystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = "/Users/tf_user/rafay_config.json"
}
```

## Authentication

The Rafay provider offers a flexible means of providing credentials for
authentication. The following methods are supported, in this order, and
explained below:

- Environment variables
- Credentials/configuration file


### Environment Variables

You can provide your credentials via the `RCTL_REST_ENDPOINT`, `RCTL_API_KEY`,
and `RCTL_PROJECT` environment variables, representing your Rafay
Console Endpoint, Rafay Access Key, and Rafay Project respectively.


```terraform
provider "rafay" {}
```

Usage:

```sh
$ export RCTL_API_KEY="rafayaccesskey"
$ export RCTL_REST_ENDPOINT="console.rafay.dev"
$ export RCTL_PROJECT="defaultproject"
$ terraform plan
```
>! Note: For `RCTL_API_KEY`, use the entire output of the generated API key.

### Credentials/configuration file

You can use an [Rafay credentials or configuration file](https://docs.rafay.co/cli/config/#config-file) to specify your credentials. You can specify a location of the configuration file in the Terraform configuration by providing the `provider_config_file`  

Usage:

```terraform
provider "rafay" {
  provider_config_file = "/Users/tf_user/rafay_config.json"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- **ignore_insecure_tls_error** (Boolean)
- **provider_config_file** (String)
