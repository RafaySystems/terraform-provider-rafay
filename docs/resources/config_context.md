---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay_config_context Resource - terraform-provider-rafay"
subcategory: ""
description: |-
  
---

# rafay_config_context (Resource)



## Example Usage

```terraform
resource "rafay_config_context" "config-context-example" {
  metadata {
    name        = var.name
    project     = var.project
    description = "this is a test config context created from terraform"
  }
  spec {
    envs {
      key       = "name-modified"
      value     = "modified-value"
      sensitive = false
    }
    envs {
      key       = "name-new"
      value     = "new-value"
      sensitive = false
    }
    files {
      name      = "file://variables.tf"
      sensitive = true
    }
    variables {
      name       = "new-variable"
      value_type = "text"
      value      = "new-value"
      options {
        override {
          type              = "restricted"
          restricted_values = ["new-value", "modified-value"]
        }
        description = "this is a dummy variable"
        sensitive   = false
        required    = true
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `metadata` (Block List, Max: 1) Metadata of the config context resource (see [below for nested schema](#nestedblock--metadata))
- `spec` (Block List, Max: 1) Specification of the config context resource (see [below for nested schema](#nestedblock--spec))

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`


***Required***

- `name` (String) name of the resource
- `project` (String) Project of the resource

***Optional***

- `description` (String) description of the resource


<a id="nestedblock--spec"></a>
### Nested Schema for `spec`

***Optional***

- `envs` (Block List) Environment variables data (see [below for nested schema](#nestedblock--spec--envs))
- `files` (Block List) File path information (see [below for nested schema](#nestedblock--spec--files))
- `sharing` (Block List, Max: 1) Defines if this is shared with other projects (see [below for nested schema](#nestedblock--spec--sharing))
- `variables` (Block List) Variables data for config context (see [below for nested schema](#nestedblock--spec--variables))

<a id="nestedblock--spec--sharing"></a>
### Nested Schema for `spec.sharing`

***Optional***

- `enabled` (Boolean) flag to specify if sharing is enabled for resource
- `projects` (Block List) list of projects this resource is shared to (see [below for nested schema](#nestedblock--spec--sharing--projects))

<a id="nestedblock--spec--sharing--projects"></a>
### Nested Schema for `spec.sharing.projects`

***Required***

- `name` (String) name of the project, '*' if to be shared with all projects

<a id="nestedblock--spec--envs"></a>
### Nested Schema for `spec.envs`

***Required***

- `key` (String) Key of the environment variable to be set
- `value` (String) Value of the environment variable to be set

***Optional***

- `sensitive` (Boolean) Determines whether the value is sensitive or not, accordingly applies encryption on it


<a id="nestedblock--spec--files"></a>
### Nested Schema for `spec.files`

***Required***

- `name` (String) relative path of a artifact

***Optional***

- `sensitive` (Boolean) Determines whether the value is sensitive or not, accordingly applies encryption on it

<a id="nestedblock--spec--variables"></a>
### Nested Schema for `spec.variables`

***Required***

- `name` (String) Name of the variable
- `value_type` (String) Specify the variable value type, Supported types are `text`, `expression`, `json`, `hcl`.
- `value` (String) Value of the variable in the specified format

***Optional***

- `options` (Block List, Max: 1) Provide the variable options (see [below for nested schema](#nestedblock--spec--variables--options))

<a id="nestedblock--spec--variables--options"></a>
### Nested Schema for `spec.variables.options`

***Optional***

- `description` (String) Description of the variable
- `override` (Block List, Max: 1) Determines whether the variable can be overridden (see [below for nested schema](#nestedblock--spec--variables--options--override))
- `required` (Boolean) Specify whether this variable is required or optional, by default it is optional
- `sensitive` (Boolean) Determines whether the value is sensitive or not, accordingly applies encryption on it

<a id="nestedblock--spec--variables--options--override"></a>
### Nested Schema for `spec.variables.options.override`

***Optional***

- `restricted_values` (List of String) If the override type is restricted, specify the values it is restricted to
- `type` (String) Specify the type of override this variable supports, Available types are `allowed`, `notallowed`, `restricted`


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

***Optional***

- `create` (String)
- `delete` (String)
- `update` (String)

