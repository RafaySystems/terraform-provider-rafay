---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay_infra_provisioner Resource - terraform-provider-rafay"
subcategory: ""
description: |-
  
---

# rafay_infra_provisioner (Resource)



## Example Usage

```terraform
resource "rafay_infra_provisioner" "tfdemoinfraprovisioner1" {
  metadata {
    name    = "tfdemoinfraprovisioner1"
    project = "upgrade"
  }
  spec {
    config {
      backend_file_path {
        name    = "some-name"
        project = "upgrade"
      }
      backend_vars = [
        {
          key   = "string"
          type  = "Plain"
          value = "string"
        },
      ]
      env_vars = [
        {
          key   = "string"
          type  = "Plain"
          value = "string"
        },
      ]
      inputVars = [
        {
          key   = "string"
          type  = "Plain"
          value = "string"
        },
      ]
      tf_vars_file_path {
        name    = "some-name"
        project = "upgrade"
      }
      version = "1.0.0"
    }
    folder_path {
      name    = "some-name"
      project = "upgrade"
    }
    repository = "gitops"
    revision   = "string"
    type       = "Terraform"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `metadata` (Block List, Max: 1) Metadata of the infrastructure provisioner resource (see [below for nested schema](#nestedblock--metadata))
- `spec` (Block List, Max: 1) Specification of the infrastructure provisioner resource (see [below for nested schema](#nestedblock--spec))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`

Optional:

- `annotations` (Map of String) annotations of the resource
- `description` (String) description of the resource
- `labels` (Map of String) labels of the resource
- `name` (String) name of the resource
- `project` (String) Project of the resource


<a id="nestedblock--spec"></a>
### Nested Schema for `spec`

Optional:

- `config` (Block List, Max: 1) (see [below for nested schema](#nestedblock--spec--config))
- `folder_path` (Block List, Max: 1) infrastructure provisioner git repository relative folder path (see [below for nested schema](#nestedblock--spec--folder_path))
- `repository` (String) infrastructure provisioner git repository name
- `revision` (String) infrastructure provisioner git repository branch or tag
- `secret` (Block List, Max: 1) InfraProvisioner secrets (see [below for nested schema](#nestedblock--spec--secret))
- `type` (String) type of infrastructure provisioner

<a id="nestedblock--spec--config"></a>
### Nested Schema for `spec.config`

Optional:

- `backend_file_path` (Block List, Max: 1) terraform state store backend file path (see [below for nested schema](#nestedblock--spec--config--backend_file_path))
- `backend_vars` (Block List) terraform state store backend variables (see [below for nested schema](#nestedblock--spec--config--backend_vars))
- `env_vars` (Block List) environment variables (see [below for nested schema](#nestedblock--spec--config--env_vars))
- `input_vars` (Block List) terraform input variables (see [below for nested schema](#nestedblock--spec--config--input_vars))
- `secret_groups` (List of String) Pipeline secrets groups
- `tf_vars_file_path` (Block List, Max: 1) terraform input variables file path (see [below for nested schema](#nestedblock--spec--config--tf_vars_file_path))
- `version` (String) terraform version

<a id="nestedblock--spec--config--backend_file_path"></a>
### Nested Schema for `spec.config.backend_file_path`

Optional:

- `name` (String) relative path of a artifact


<a id="nestedblock--spec--config--backend_vars"></a>
### Nested Schema for `spec.config.backend_vars`

Optional:

- `key` (String) variable key
- `type` (String) variable type
- `value` (String) variable value


<a id="nestedblock--spec--config--env_vars"></a>
### Nested Schema for `spec.config.env_vars`

Optional:

- `key` (String) variable key
- `type` (String) variable type
- `value` (String) variable value


<a id="nestedblock--spec--config--input_vars"></a>
### Nested Schema for `spec.config.input_vars`

Optional:

- `key` (String) variable key
- `type` (String) variable type
- `value` (String) variable value


<a id="nestedblock--spec--config--tf_vars_file_path"></a>
### Nested Schema for `spec.config.tf_vars_file_path`

Optional:

- `name` (String) relative path of a artifact



<a id="nestedblock--spec--folder_path"></a>
### Nested Schema for `spec.folder_path`

Optional:

- `name` (String) relative path of a artifact


<a id="nestedblock--spec--secret"></a>
### Nested Schema for `spec.secret`

Optional:

- `name` (String) relative path of a artifact



<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

