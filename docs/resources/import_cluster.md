---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay_import_cluster Resource - terraform-provider-rafay"
subcategory: ""
description: |-

---

# rafay_import_cluster (Resource)

Import a cluster.

## Example Usage


```terraform
resource "rafay_import_cluster" "import_cluster" {
  clustername     = "terraform-importcluster"
  projectname     = "terraform"
  blueprint       = "default"
  kubeconfig_path = "<file-path/kubeconfig.yaml>"
  location        = "losangeles-us"
  values_path     = "<optional_path/values.yaml>"
  bootstrap_path  = "<optional_path/bootstrap.yaml>"
  labels = {
    "key1" = "value1"
    "key2" = "value2"
  }
  kubernetes_provider   = "AKS"
  provision_environment = "CLOUD"
}

output "values_data" {
  value = rafay_import_cluster.import_cluster.values_data
}

output "values_path" {
  value = rafay_import_cluster.import_cluster.values_path
}

output "bootstrap_data" {
  value = rafay_import_cluster.import_cluster.bootstrap_data
}

output "bootstrap_path" {
  value = rafay_import_cluster.import_cluster.bootstrap_path
}
```

<!-- schema generated by tfplugindocs -->
## Argument Reference

***Required***

- `blueprint` - (String) The name of the blueprint used for the cluster.
- `clustername` - (String) The name of the cluster.
- `projectname` - (String) The name of the project the cluster belongs to.

***Optional***

- `blueprint_version` - (String) The version of the blueprint.
- `description` - (String) The description for the cluster.
- `kubeconfig_path` - (String) The path to the kubeconfig file.
- `labels` - (Block) Labels are key/value pairs that are attached to the object.
- `kubernetes_provider` (String)  This field is used to define the Kubernetes provider. Supported values are `EKS`, `AKS`, `GKE`, `OPENSHIFT`, `OTHER`, `RKE` and `EKSANYWHERE`
- `provision_environment` (String) This field is used to define the type of environment. The supported values are `CLOUD` and `ONPREM`
- `values_path` - (String) The path to save the `values.yaml` file to. This is an optional parameter. If path is provided values.yaml will be downloaded to that path. Otherwise values.yaml will be downloaded to current directory and this output variable will be populated with path to the downloaded file.
- `bootstrap_path` - (String) The path to save the `bootstrap.yaml` file to. This is an optional parameter. If path is provided bootstrap.yaml will be downloaded to that path. Otherwise bootstrap.yaml will be downloaded to current directory and this output variable will be populated with path to the downloaded file.
- `timeouts` - (Block) Sets the duration of time the create, delete, and update functions are allowed to run. If the function takes longer than this, it is assumed the function has failed. The default is 10 minutes. (See [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

***Optional***

- `create` - (String) Sets the timeout duration for creating a resource. The default timeout is 10 minutes.
- `delete` - (String) Sets the timeout duration for deleting a resource. The default timeout is 10 minutes.
- `update` - (String) Sets the timeout duration for updating a resource. The default timeout is 10 minutes.

## Attribute Reference

- `id` - (String) The ID of the resource, generated by the system after you create the resource.
  
---

# rafay_import_cluster (Data Source)

```terraform
data "rafay_import_cluster" "import-sample-cluster" {
  metadata = {
    name    = "import-cluster"
    project = "sample-project"
  }
}

output "import_cluster" {
  description = "import_cluster"
  value       = data.rafay_import_cluster.import-sample-cluster
}
```
