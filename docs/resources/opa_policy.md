---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "rafay_opa_policy Resource - terraform-provider-rafay"
subcategory: ""
description: |-
  
---

# rafay_opa_policy (Resource)

Create an Open Policy Agent (OPA) policy resource. 

## Example Usage

---

```terraform
#Basic example for opa policy
resource "rafay_opa_policy" "tftestopapolicy1" {
  metadata {
    name    = "tftestopapolicy1"
    project = "terraform"
  }
  spec {
    constraint_list {
      name = "tfdemoopaconstraint1"
      version = "v1"
    }
    sharing {
      enabled = false
    }
    version = "v0"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

***Required***

- `metadata` (Block List, Max: 1) Metadata of the policy resource (see [below for nested schema](#nestedblock--metadata))
- `spec` (Block List, Max: 1) Specification of the policy resource (see [below for nested schema](#nestedblock--spec))

***Optional***
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--metadata"></a>
### Nested Schema for `metadata`

***Required***

- `name` (String) name of the resource
- `project` (String) Project of the resource


<a id="nestedblock--spec"></a>
### Nested Schema for `spec`

***Required***

- `constraint_list` (Block List) List of constraints (see [below for nested schema](#nestedblock--spec--constraint_list))
- `version` (String) version of the policy

***Optional***

- `sharing` (Block List, Max: 1) policy sharing configuration (see [below for nested schema](#nestedblock--spec--sharing))
- 
<a id="nestedblock--spec--constraint_list"></a>
### Nested Schema for `spec.constraint_list`

***Required***

- `name` (String) Name of constraint
- `version` (String) version of constraint

<a id="nestedblock--spec--sharing"></a>
### Nested Schema for `spec.sharing`

***Required***

- `enabled` (Boolean) flag to specify if sharing is enabled for resource
- `projects` (Block List) list of projects this resource is shared to (see [below for nested schema](#nestedblock--spec--sharing--projects))

<a id="nestedblock--spec--sharing--projects"></a>
### Nested Schema for `spec.sharing.projects`

***Required***

- `name` (String) name of the project

Note: To share a resource across all projects in an organisation, below spec can be used
 ```
     sharing {
      enabled = true
      projects {
        name = "*"
      }
    }
```

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

***Optional***
- `create` - (String) Sets the timeout duration for creating a resource. The default timeout is 10 minutes. 
- `delete` - (String) Sets the timeout duration for deleting a resource. The default timeout is 10 minutes. 
- `update` - (String) Sets the timeout duration for updating a resource. The default timeout is 10 minutes. 


## Attribute Reference

---

- `id` - (String) The ID of the resource, generated by the system after you create the resource.