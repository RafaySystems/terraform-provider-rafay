---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay_chargeback_share Resource - terraform-provider-rafay"
subcategory: ""
description: |-
  
---

# rafay_chargeback_share (Resource)
Chargeback group share is to enable/disable the Share unallocated cost based on tenancy or resource allocation.

## Example Usage

Example Chargeback Share resource :

```terraform
resource "rafay_chargeback_share" "tfdemochargebackshare" {
  metadata {
    name = "chargebackshare"
  }
  spec {
    share_enabled = true
    share_type = "tenancy"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

***Required***

- `metadata` (Block List, Max: 1) Metadata of the chargeback group report resource (see [below for nested schema](#nestedblock--metadata))
- `spec` (Block List, Max: 1) Specification of the chargeback group report resource (see [below for nested schema](#nestedblock--spec))

***Optional***

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`

***Required***

- `name` (String) name of the resource

<a id="nestedblock--spec"></a>
### Nested Schema for `spec`

***Required***

- `share_enabled` (Boolean) Enable/disable the Share unallocated cost (see [below for nested schema](#nestedblock--spec--time))
- `share_type` (String) Share type. Valid values are `allocation` and `tenancy`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

***Optional***
- `create` - (String) Sets the timeout duration for creating a resource. The default timeout is 10 minutes. 
- `update` - (String) Sets the timeout duration for updating a resource. The default timeout is 10 minutes. 

## Attribute Reference

---

- `id` - (String) The ID of the resource, generated by the system after you create the resource.