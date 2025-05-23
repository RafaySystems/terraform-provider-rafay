---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay_blueprints Data Source - terraform-provider-rafay"
subcategory: ""
description: |-
  
---

# rafay_blueprints (Data Source)

## Example Usage

```terraform
data "rafay_blueprints" "list" {
  projectname = "defaultproject"
}

output "blueprint_list" {
  description = "blueprints list"
  value       = data.rafay_blueprints.list.blueprints
}

```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `projectname` (String) Project name from where blueprints to be listed

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `blueprints` (List of Object) A list of blueprints (see [below for nested schema](#nestedatt--blueprints))
- `id` (String) The ID of this resource.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String)


<a id="nestedatt--blueprints"></a>
### Nested Schema for `blueprints`

Read-Only:

- `deployed_clusters` (String) Deployed clusters count
- `name` (String) Name of the blueprint
- `ownership` (String) Ownership of the blueprint
- `versions` (String) Version count of the blueprint
