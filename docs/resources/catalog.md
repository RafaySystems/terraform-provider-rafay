---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay_catalog Resource - terraform-provider-rafay"
subcategory: ""
description: |-
  
---

# rafay_catalog (Resource)



## Example Usage

```terraform
resource "rafay_catalog" "basic_custom_catalog" {
  metadata {
    name    = "terraform-test"
    project = "terraform"
  }
  spec {
    auto_sync  = false
    repository = "istio-terraform"
    type       = "HelmRepository"
    sharing {
      enabled = true
      projects {
        name = "defaultproject"
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `metadata` (Block List, Max: 1) Metadata of the catalog resource (see [below for nested schema](#nestedblock--metadata))
- `spec` (Block List, Max: 1) Specification of the catalog resource (see [below for nested schema](#nestedblock--spec))
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

- `auto_sync` (Boolean) flag to indicate if the catalog is synced periodically
- `icon_url` (String) icon url of catalog
- `repository` (String) catalog helm repository name
- `sharing` (Block List, Max: 1) catalog sharing configuration (see [below for nested schema](#nestedblock--spec--sharing))
- `type` (String) type of catalog

<a id="nestedblock--spec--sharing"></a>
### Nested Schema for `spec.sharing`

Optional:

- `enabled` (Boolean) flag to specify if sharing is enabled for resource
- `projects` (Block List) list of projects this resource is shared to (see [below for nested schema](#nestedblock--spec--sharing--projects))

<a id="nestedblock--spec--sharing--projects"></a>
### Nested Schema for `spec.sharing.projects`

Optional:

- `name` (String) name of the project




<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `update` (String)

