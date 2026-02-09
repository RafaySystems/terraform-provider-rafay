# State Order Approach - Explanation and Examples

## The Problem (In Simple Terms)

**Current situation:**
- You have a list of nodegroups in your Terraform config (resource.tf)
- The order matters to Terraform - it compares them position-by-position like items in a shopping list
- If you reorder the list in your config, Terraform thinks you're replacing ALL nodegroups (even though they're the same nodegroups, just reordered)
- This causes confusing diffs that show changes when nothing really changed

**Example of the problem:**
```hcl
# Your config has: ng-2, ng-4, ng-3, ng-1
# State has:       ng-1, ng-2, ng-3, ng-4

# Terraform compares position-by-position:
# Position 0: ng-2 (config) vs ng-1 (state) → DIFF!
# Position 1: ng-4 (config) vs ng-2 (state) → DIFF!
# Position 2: ng-3 (config) vs ng-3 (state) → Same ✓
# Position 3: ng-1 (config) vs ng-4 (state) → DIFF!
```

## The Solution (In Simple Terms)

**Instead of storing state in alphabetical order, store it in the SAME order as your config.**

Think of it like this:
- You write a shopping list: Milk, Eggs, Bread, Cheese
- Terraform's state file remembers: "Item 1 is Milk, Item 2 is Eggs, Item 3 is Bread, Item 4 is Cheese"
- When you refresh, Terraform checks what's actually at the store and **puts them back in YOUR list order**
- If you reorder your list to: Cheese, Milk, Bread, Eggs → Terraform reorders the state to match → no false diffs

**Key insight:** The provider will "follow" whatever order YOU put in the config file.

**Important timing details:**
- **After apply**: State matches config order (Read runs after apply and sorts to config order)
- **During Read**: Keeps prior state order to avoid API-order churn (API may return nodegroups in different order each call)

---

## Scenario Examples

### Scenario 1: User Reorders Nodegroups in Config

**Initial state:**
```hcl
# resource.tf
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**State file (after apply):**
```
managed_nodegroups[0]: ng-1
managed_nodegroups[1]: ng-2
managed_nodegroups[2]: ng-3
managed_nodegroups[3]: ng-4
```

**User reorders config:**
```hcl
# resource.tf (NEW ORDER)
managed_nodegroups { name = "ng-4" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-3" ... }
```

**First `terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          ~ managed_nodegroups[0]: ng-4 (was ng-1)
          ~ managed_nodegroups[1]: ng-2 (no change)
          ~ managed_nodegroups[2]: ng-1 (was ng-3)
          ~ managed_nodegroups[3]: ng-3 (was ng-4)
        }
    }
```

**After `terraform apply` (or `terraform refresh`):**
- State is reordered to match config
- State file now has: ng-4, ng-2, ng-1, ng-3

**Next `terraform plan`:**
```
No changes. Your infrastructure matches the configuration.
```

✅ **Result:** After one refresh cycle, reordering in config shows no changes.

---

### Scenario 2: User Adds a Nodegroup

#### 2a. Add at the Beginning

**Current config:**
```hcl
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**User adds ng-1 at the start:**
```hcl
managed_nodegroups { name = "ng-1" ... }  # NEW
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**`terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          + managed_nodegroups[0]: ng-1 (new)
          ~ managed_nodegroups[1]: ng-2 (was at position 0)
          ~ managed_nodegroups[2]: ng-3 (was at position 1)
          ~ managed_nodegroups[3]: ng-4 (was at position 2)
        }
    }
```

⚠️ **Note:** Adding at the beginning causes a "cascade" - all existing nodegroups shift positions. This is unavoidable with Terraform lists - it's how lists work. The provider/backend should ideally only create ng-1; plan noise is cosmetic.

#### 2b. Add at the End

**Current config:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
```

**User adds ng-4 at the end:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }  # NEW
```

**`terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          + managed_nodegroups[3]: ng-4 (new)
        }
    }
```

✅ **Result:** Clean addition, no cascade!

#### 2c. Add in the Middle

**Current config:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**User adds ng-2 in the middle:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }  # NEW
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**`terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          + managed_nodegroups[1]: ng-2 (new)
          ~ managed_nodegroups[2]: ng-3 (was at position 1)
          ~ managed_nodegroups[3]: ng-4 (was at position 2)
        }
    }
```

⚠️ **Note:** Same cascade issue - nodegroups after the insertion point shift positions. The provider/backend should ideally only create ng-2; plan noise is cosmetic.

---

### Scenario 3: User Removes a Nodegroup

#### 3a. Remove from the Beginning

**Current config:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**User removes ng-1:**
```hcl
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**`terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          - managed_nodegroups[0]: ng-1 (removed)
          ~ managed_nodegroups[0]: ng-2 (was at position 1)
          ~ managed_nodegroups[1]: ng-3 (was at position 2)
          ~ managed_nodegroups[2]: ng-4 (was at position 3)
        }
    }
```

⚠️ **Note:** Cascade - all nodegroups shift up one position. The provider/backend should ideally only delete ng-1; plan noise is cosmetic.

#### 3b. Remove from the End

**Current config:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**User removes ng-4:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
```

**`terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          - managed_nodegroups[3]: ng-4 (removed)
        }
    }
```

✅ **Result:** Clean removal, no cascade!

#### 3c. Remove from the Middle

**Current config:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**User removes ng-2:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**`terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          - managed_nodegroups[1]: ng-2 (removed)
          ~ managed_nodegroups[1]: ng-3 (was at position 2)
          ~ managed_nodegroups[2]: ng-4 (was at position 3)
        }
    }
```

⚠️ **Note:** Cascade - nodegroups after the removal point shift positions. The provider/backend should ideally only delete ng-2; plan noise is cosmetic.

---

### Scenario 4: User Modifies a Nodegroup

**Current config:**
```hcl
managed_nodegroups {
  name = "ng-2"
  instance_type = "t3.large"
  desired_capacity = 1
}
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**User changes instance_type of ng-2:**
```hcl
managed_nodegroups {
  name = "ng-2"
  instance_type = "t3.xlarge"  # CHANGED
  desired_capacity = 1
}
managed_nodegroups { name = "ng-3" ... }
managed_nodegroups { name = "ng-4" ... }
```

**`terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          ~ managed_nodegroups[0]: {
              ~ instance_type: "t3.large" -> "t3.xlarge"
            }
        }
    }
```

✅ **Result:** Only the modified nodegroup shows changes, regardless of position!

---

## Out-of-Band Changes

### Scenario: Nodegroup Added Outside Terraform

**Current config:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
```

**Someone manually creates ng-3 via the Rafay console (outside Terraform)**

**`terraform refresh` (or `terraform plan`):**
```
Terraform detected the following changes made outside of Terraform:

  # rafay_eks_cluster.ekscluster-basic has changed
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          + managed_nodegroups[2]: ng-3 (detected outside Terraform)
        }
    }
```

**After refresh, state now has:** ng-1, ng-2, ng-3

**Next `terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          - managed_nodegroups[2]: ng-3 (not in config, will be removed)
        }
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

**Explanation:** Since ng-3 is not in your config, Terraform wants to delete it to match your desired state. You have two choices:
1. Add ng-3 to your config to keep it
2. Run `terraform apply` to remove ng-3 and sync infrastructure with config

---

### Scenario: Nodegroup Deleted Outside Terraform

**Current config:**
```hcl
managed_nodegroups { name = "ng-1" ... }
managed_nodegroups { name = "ng-2" ... }
managed_nodegroups { name = "ng-3" ... }
```

**Someone manually deletes ng-2 via the Rafay console**

**`terraform refresh`:**
```
Terraform detected the following changes made outside of Terraform:

  # rafay_eks_cluster.ekscluster-basic has changed
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          - managed_nodegroups[1]: ng-2 (deleted outside Terraform)
          ~ managed_nodegroups[1]: ng-3 (was at position 2)
        }
    }
```

**After refresh, state has:** ng-1, ng-3

**Next `terraform plan`:**
```
Terraform will perform the following actions:

  # rafay_eks_cluster.ekscluster-basic will be updated in-place
  ~ resource "rafay_eks_cluster" "ekscluster-basic" {
      ~ cluster_config {
          + managed_nodegroups[1]: ng-2 (will be recreated)
          ~ managed_nodegroups[2]: ng-3 (was at position 1)
        }
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

**Explanation:** Terraform sees ng-2 in your config but not in reality, so it will recreate it. You have two choices:
1. Remove ng-2 from config if you wanted it deleted
2. Run `terraform apply` to recreate ng-2

---

## Summary: What Works vs. What Doesn't

| Scenario | Result |
|----------|--------|
| ✅ Reorder nodegroups in config | No changes after apply |
| ✅ Add nodegroup at END | Clean add, no cascade |
| ⚠️ Add nodegroup at START/MIDDLE | Cascade (plan noise; provider/backend should ideally only create the new nodegroup) |
| ✅ Remove nodegroup from END | Clean removal, no cascade |
| ⚠️ Remove nodegroup from START/MIDDLE | Cascade (plan noise; provider/backend should ideally only delete the removed nodegroup) |
| ✅ Modify any nodegroup | Only that nodegroup shows changes |
| ✅ Out-of-band additions | Detected in refresh, can be kept or removed |
| ✅ Out-of-band deletions | Detected in refresh, will be recreated or config updated |

**The cascade issue is a Terraform limitation with lists, not a provider bug.** The provider/backend should ideally only perform the actual required operations (create/delete/update specific nodegroups) - the position shifts are only visible in the plan output.

## Alternative: Use Set Instead of List

If cascades are unacceptable, we can change `managed_nodegroups` from a **list** to a **set**. With sets:
- Order doesn't matter at all
- No cascades ever
- But modifications show as "remove old + add new" instead of "update in-place"

Trade-off: Better for add/remove scenarios, slightly worse for modify scenarios.
