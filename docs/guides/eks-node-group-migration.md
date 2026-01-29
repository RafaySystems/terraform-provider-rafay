---
page_title: "Rafay EKS Cluster resource - Node group migration"

---

# Rafay EKS Cluster resource - Node group migration

## Overview

<div style="border: 2px solid #e67e22; background:#fff4e5; padding:12px; border-radius:6px; margin:12px 0;"> ⚠️ <strong>Important</strong><br><br>
This migration guidance applies specifically to the <strong>EKS cluster type</strong> in the Terraform Provider for the Kubernetes Operations Platform.
</div>

The Terraform provider for EKS cluster management has been enhanced to deliver more consistent configuration handling and resolve previously encountered diff inconsistencies.

The latest version of the Terraform provider introduces several enhancements, including:

* Consistent and predictable diff behavior
* Improved handling of deeply nested objects
* A clearer and more flexible schema design
* Better alignment with Terraform’s long-term architectural direction

To take advantage of these improvements, the EKS cluster resource now supports a new map-based node group structure (`node_groups_map`, `managed_nodegroups_map`).This format offers a more organized way to define node groups, reduces unnecessary diffs, and makes ongoing configuration updates smoother and more predictable.

<div style="border: 2px solid #448aff; background:#edf3ff; padding:12px; border-radius:6px; margin:12px 0;"> ✏️ <strong>Note</strong><br><br>
Migration to the map based structure is <strong>optional</strong>. Your existing list-based configuration will continue to work as is without requiring changes.<br><br>

However, if you want to <strong>avoid unnecessary diffs</strong> and take advantage of the <strong>improved stability and behavior</strong> introduced in the latest provider, we recommend updating to the map-based format.</div>

---

## Enhancements in Node Group Management


<div style="border: 2px solid #e67e22; background:#fff4e5; padding:12px; border-radius:6px; margin:12px 0;"> ⚠️ <strong>Deprecated Node Group Configuration Blocks</strong><br><br>


<strong>For existing clusters using Terraform</strong>, you may see deprecation warnings for node group configuration blocks: <br><br>

<strong>Self-Managed Node Groups:</strong><br>
- The <code>node_groups</code> block is deprecated and will be removed in a future release<br>
- Use <code>node_groups_map</code> for new configs<br>
- Existing setups still work, but migration is recommended<br><br>

<strong>Managed Node Groups:</strong><br>
- The <code>managed_nodegroups</code> block is deprecated and will be removed in a future release<br>
- Use <code>managed_nodegroups_map</code> for new configs<br>
- Existing setups still work, but migration is recommended<br><br>

<strong>Note:</strong><br>
These warnings do not impact functionality. Existing configurations will continue to work, but we recommend migrating to the new map-based format for better maintainability and to leverage the fixes. <br><br>

</div>

To take advantage of these improvements, the provider now supports a map-based structure for node groups and managed node groups. This structure offers better clarity, simplifies future updates, and enables smoother lifecycle operations.

---

## Migration Steps

Follow the controlled workflow below to migrate node groups from list format to map format without triggering misleading diffs.

**Step 1. Upgrade the Terraform Provider**

Upgrade to the latest version of the Terraform provider, which includes the enhanced EKS resource and support for the map-based node group structure.

```bash
terraform init -upgrade
```

**Step 2. Set Migration Flag**

Expose the environment variable to indicate that a migration is in progress. This environment variable suppresses misleading Terraform diffs when migrating node groups from list to map format.

```bash
export TF_RAFAY_EKS_MIGRATE_TO_MAP=true
```


<div style="border: 2px solid #448aff; background:#edf3ff; padding:12px; border-radius:6px; margin:12px 0;"> 💡 <strong>Important</strong><br><br>
Ensure the environment variable <code>TF_RAFAY_EKS_MIGRATE_TO_MAP</code> is passed to the Terraform runtime when running in automated environments such as CI/CD pipelines.
</div>

**Step 3. Refresh Terraform State**

Refresh local or remote Terraform state so the provider can realign internal structures.

```bash
terraform refresh
```

**Step 4.Update Configuration to the New Structure**

Modify node group and managed node group definitions in the configuration file to use the new **map-based structure**.

A breakdown of the format change is described in the **Migrating Node Groups to the Map Format** section below.


**Step 5. Validate Configuration**

Run a Terraform plan to confirm that the migration completed successfully.

```bash
terraform plan
```

If no diffs appear, the configuration and migration steps are complete.

---


## Migrating Node Groups to the Map Format

This section expands **Step 4 Update Configuration to the New Structure** and explains how to manually convert node groups from the list-based structure to the new map-based structure.

**Change in the Top-Level Definition**

The first change appears at the top level of the node group definition:

***Old (List Block)***

Uses repeated blocks for each node group.

```hcl
node_groups {
```

***New (Map Attribute)***

Uses a single map where each key represents a node group name.

```hcl
node_groups_map = {
```

This change replaces multiple repeated `node_groups {}` blocks with one unified map. Each node group is now defined using its name as the map key, allowing Terraform to clearly identify and track each node group independently.

- Old Format: List-Based Node Groups

```hcl
node_groups {
  name               = "ng-1"
  desired_capacity   = 1
  # … other fields
}
```

In this format, Terraform interprets each `node_groups {}` block as an item in a list. Adding, removing, or reordering these blocks often caused diff inconsistencies because list ordering mattered.


- New Format: Map-Based Node Groups

```hcl
node_groups_map = {
  "ng-1" = {
    desired_capacity = 1
    # … other fields
  }
}
```

In this updated format:

* `"ng-1"` becomes the **map key**
* All node group fields move inside the map value
* Terraform tracks node groups by **name**, not by position

This approach provides predictable diffs and smoother updates.


### Field-Level Changes

<div style="width:570px; overflow-x:auto; white-space:nowrap; padding:8px;">
<table><thead>
  <tr>
    <th>Field</th>
    <th>List Block</th>
    <th>Map Attribute / List</th>
    <th>Change</th>
  </tr></thead>
<tbody>
  <tr>
    <td><code>name</td>
    <td><code>name = "ng-1"</td>
    <td><code>"ng-1" = {}</td>
    <td>Name is now the map key</td>
  </tr><br>
  <tr>
    <td><code>iam</td>
    <td><code>iam { … }</td>
    <td><code>iam = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>iam_node_group_with_addon_policies</td>
    <td><code>iam_node_group_with_addon_policies { … }</td>
    <td><code>iam_node_group_with_addon_policies = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>statement</td>
    <td><code>statement { … }</td>
    <td><code>statement = [{ … }]</td>
    <td>Block → List of objects</td>
  </tr>
  <tr>
    <td><code>ssh</td>
    <td><code>ssh { … }</td>
    <td><code>ssh = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>placement</td>
    <td><code>placement { … }</td>
    <td><code>placement = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>instance_selector</td>
    <td><code>instance_selector { … }</td>
    <td><code>instance_selector = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>bottle_rocket</td>
    <td><code>bottle_rocket { … }</td>
    <td><code>bottle_rocket = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>instances_distribution</td>
    <td><code>instances_distribution { … }</td>
    <td><code>instances_distribution = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>asg_metrics_collection</td>
    <td><code>asg_metrics_collection { … }</td>
    <td><code>asg_metrics_collection = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>update_config</td>
    <td><code>update_config { … }</td>
    <td><code>update_config = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>kubelet_extra_config</td>
    <td><code>kubelet_extra_config { … }</td>
    <td><code>kubelet_extra_config = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>security_groups</td>
    <td><code>security_groups { … }</td>
    <td><code>security_groups = { … }</td>
    <td>Block → Attribute</td>
  </tr>
  <tr>
    <td><code>taints</td>
    <td><code>taints { … }</td>
    <td><code>taints = [{ … }]</td>
    <td>Block → List of maps</td>
  </tr>
</tbody>
</table>
</div>

<div style="border: 2px solid #448aff; background:#edf3ff; padding:12px; border-radius:6px; margin:12px 0;"> 💡 <strong>Important</strong><br><br>
These changes apply to both <code>node_groups_map</code> and <code>managed_nodegroups_map</code>.
</div>

## Examples

Here is an exmple of node group:

**List based Node group example**

```hcl
node_groups {
  disable_imdsv1     = false
  disable_pods_imds  = false
  efa_enabled        = false
  volume_iops        = 3000
  volume_throughput  = 125 
  name               = "ng-1"
  ami_family         = "AmazonLinux2"

  iam {
    iam_node_group_with_addon_policies {
      alb_ingress     = false
      app_mesh        = false
      app_mesh_review = false
      cert_manager    = false
      cloud_watch     = false
      ebs             = false
      efs             = false
      external_dns    = false
      fsx             = false
      xray            = false
      image_builder   = true
      auto_scaler     = true
    }
  }

  instance_type      = "t3.xlarge"
  desired_capacity   = 1
  min_size           = 1
  max_size           = 2
  max_pods_per_node  = 50
  version            = "1.31"
  volume_size        = 80
  volume_type        = "gp3"
  private_networking = false
}
```

**Map based Node group example:**

```hcl
node_groups_map = {
  "ng-1" = { 
    disable_imdsv1     = false
    disable_pods_imds  = false
    efa_enabled        = false
    volume_iops        = 3000
    volume_throughput  = 125 
    ami_family         = "AmazonLinux2"

    iam = {
      iam_node_group_with_addon_policies = {
        alb_ingress     = false
        app_mesh        = false
        app_mesh_review = false
        cert_manager    = false
        cloud_watch     = false
        ebs             = false
        efs             = false
        external_dns    = false
        fsx             = false
        xray            = false
        image_builder   = true
        auto_scaler     = true
      }
    }

    instance_type      = "t3.xlarge"
    desired_capacity   = 1
    min_size           = 1
    max_size           = 2
    max_pods_per_node  = 50
    version            = "1.31"
    volume_size        = 80
    volume_type        = "gp3"
    private_networking = false
  }
}
```

---

## Post-Migration Note: Proxy Config Diff

After upgrading to the latest provider, a diff may appear during the next `terraform plan` if the configuration includes an empty `proxy_config {}` block or `value = ""` attribute in taints:

```
~ cluster {
    ~ spec {
        + proxy_config = {}
    }
}
Plan: 0 to add, 1 to change, 0 to destroy.
```

```
~ "ng1" = {
    ~ taints                     = [
        - {
            - effect = "NoExecute" -> null
            - key    = "app" -> null
          },
        + {
            + effect = "NoExecute"
            + key    = "app"
              # (1 unchanged attribute hidden)
          },
      ]
      # (21 unchanged attributes hidden)
  },
  # (1 unchanged element hidden)
```

**How to Resolve It**

Two options are available:

1. **Remove the empty block/attribute**: Delete the `proxy_config {}` entry if proxy configuration is not required. Remove `value = ""` from taints if not required.
2. **Run `terraform apply` once**: This updates the backend state and prevents future diffs.

<div style="border: 2px solid #448aff; background:#edf3ff; padding:12px; border-radius:6px; margin:12px 0;"> ✏️ <strong>Note</strong><br><br>
An empty <code>proxy_config {}</code> block and empty <code>value = ""</code> attribute behaves the same as removing it entirely.
</div>



---