# Access Entries and Identity Mappings - Map-Based Design

## The Diff Problem with Lists

### ❌ OLD Approach (SDK v2 with Lists)

```hcl
resource "rafay_eks_cluster" "example" {
  cluster_config {
    # Access entries as a list - CAUSES UNWANTED DIFFS
    access_entries = [
      {
        principal_arn = "arn:aws:iam::123456789012:role/Developer"
        type          = "STANDARD"
        username      = "developer"
        groups        = ["system:masters"]
      },
      {
        principal_arn = "arn:aws:iam::123456789012:role/Admin"
        type          = "STANDARD"  
        username      = "admin"
        groups        = ["system:masters"]
      },
      {
        principal_arn = "arn:aws:iam::123456789012:role/ReadOnly"
        type          = "STANDARD"
        username      = "readonly"
        groups        = ["viewers"]
      }
    ]
    
    # Identity mappings as nested lists - ALSO PROBLEMATIC
    identity_mappings {
      arns = [
        {
          arn      = "arn:aws:iam::123456789012:role/NodeGroup"
          username = "system:node:{{EC2PrivateDNSName}}"
          group    = "system:bootstrappers"
        },
        {
          arn      = "arn:aws:iam::123456789012:role/Admin"
          username = "admin"
          group    = "system:masters"
        }
      ]
    }
  }
}
```

**Problem**: If you update the middle entry:
```hcl
access_entries = [
  { principal_arn = "arn:.../Developer", ... },  # Index 0
  { 
    principal_arn = "arn:.../Admin"
    type          = "EC2_LINUX"  # ← CHANGED THIS
    ...
  },  # Index 1
  { principal_arn = "arn:.../ReadOnly", ... },   # Index 2
]
```

**Terraform shows**:
```diff
  ~ resource "rafay_eks_cluster" "example" {
      ~ cluster_config {
          ~ access_entries[1] = {
              ~ type = "STANDARD" -> "EC2_LINUX"
            }
          # Sometimes also shows:
          ~ access_entries[2] = { ... }  # ← UNWANTED DIFF!
        }
    }
```

### Why This Happens
1. **Positional Matching**: Terraform compares lists by position
2. **Hash Changes**: When one item changes, list hash changes
3. **Reordering Confusion**: Terraform may think items shifted
4. **Plan Noise**: Shows changes that aren't really happening

---

## ✅ NEW Approach (Plugin Framework with Maps)

```hcl
resource "rafay_eks_cluster_v2" "example" {
  cluster_config = {
    # Access entries as a MAP - NO UNWANTED DIFFS!
    access_entries = {
      "developer-role" = {
        principal_arn = "arn:aws:iam::123456789012:role/Developer"
        type          = "STANDARD"
        username      = "developer"
        groups        = ["system:masters"]
      }
      "admin-role" = {
        principal_arn = "arn:aws:iam::123456789012:role/Admin"
        type          = "STANDARD"
        username      = "admin"
        groups        = ["system:masters"]
      }
      "readonly-role" = {
        principal_arn = "arn:aws:iam::123456789012:role/ReadOnly"
        type          = "STANDARD"
        username      = "readonly"
        groups        = ["viewers"]
      }
    }
    
    # Identity mappings with MAPS for ARNs
    identity_mappings = {
      arns = {
        "nodegroup-role" = {
          arn      = "arn:aws:iam::123456789012:role/NodeGroup"
          username = "system:node:{{EC2PrivateDNSName}}"
          groups   = ["system:bootstrappers", "system:nodes"]
        }
        "admin-role" = {
          arn      = "arn:aws:iam::123456789012:role/Admin"
          username = "admin"
          groups   = ["system:masters"]
        }
      }
      accounts = ["123456789012", "987654321098"]
    }
  }
}
```

**Now update the middle entry**:
```hcl
access_entries = {
  "developer-role" = { ... }  # Key: developer-role
  "admin-role" = {            # Key: admin-role
    principal_arn = "arn:.../Admin"
    type          = "EC2_LINUX"  # ← CHANGED THIS
    username      = "admin"
    groups        = ["system:masters"]
  }
  "readonly-role" = { ... }   # Key: readonly-role
}
```

**Terraform shows** (PERFECT!):
```diff
  ~ resource "rafay_eks_cluster_v2" "example" {
      ~ cluster_config = {
          ~ access_entries = {
              ~ "admin-role" = {
                  ~ type = "STANDARD" -> "EC2_LINUX"
                }
              # "developer-role" unchanged
              # "readonly-role" unchanged
            }
        }
    }
```

**Only the changed item shows!** ✅

---

## Benefits of Map-Based Access Entries

### 1. Precise Diffs
```diff
# Only what changed
~ access_entries["admin-role"].type: "STANDARD" -> "EC2_LINUX"

# NOT this:
~ access_entries[1].type: "STANDARD" -> "EC2_LINUX"
~ access_entries[2]: (unwanted noise)
```

### 2. Stable References
```hcl
# Easy to reference in for_each
resource "aws_iam_role_policy_attachment" "access" {
  for_each = var.access_entries
  
  role       = each.value.principal_arn
  policy_arn = "..."
}

# Can reference by key
output "admin_access" {
  value = rafay_eks_cluster_v2.example.cluster_config.access_entries["admin-role"]
}
```

### 3. No Order Dependency
```hcl
# These are IDENTICAL (order doesn't matter with maps)
access_entries = {
  "a" = { ... }
  "b" = { ... }
}

access_entries = {
  "b" = { ... }
  "a" = { ... }
}
```

### 4. Clear Intent
```hcl
# Map key describes the entry
"developer-role" = { ... }  # Clear what this is
"admin-role" = { ... }       # Self-documenting
"readonly-role" = { ... }    # Easy to understand

# vs list index
[0] = { ... }  # What is this?
[1] = { ... }  # Which role?
[2] = { ... }  # Have to read the content
```

---

## Complete Schema Definition

### Access Entries (Map by Key)

```hcl
access_entries = {
  "unique-key-1" = {
    principal_arn           = "arn:aws:iam::123456789012:role/MyRole"
    type                    = "STANDARD"  # or "EC2_LINUX", "EC2_WINDOWS", "FARGATE_LINUX"
    username                = "my-user"
    groups                  = ["group1", "group2"]
    kubernetes_groups       = ["system:masters"]
    
    # For EKS Pod Identity
    cluster_id              = "cluster-12345"
    
    # Access policies
    access_policies = {
      "policy-1" = {
        policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy"
        access_scope = {
          type       = "cluster"  # or "namespace"
          namespaces = ["default", "kube-system"]
        }
      }
    }
  }
}
```

### Identity Mappings (Nested Maps)

```hcl
identity_mappings = {
  # Map of ARN mappings
  arns = {
    "node-role" = {
      arn      = "arn:aws:iam::123456789012:role/NodeInstanceRole"
      username = "system:node:{{EC2PrivateDNSName}}"
      groups   = ["system:bootstrappers", "system:nodes"]
    }
    "admin-role" = {
      arn      = "arn:aws:iam::123456789012:role/AdminRole"
      username = "admin"
      groups   = ["system:masters"]
    }
    "readonly-role" = {
      arn      = "arn:aws:iam::123456789012:role/ReadOnlyRole"  
      username = "readonly:{{SessionName}}"
      groups   = ["viewers"]
    }
  }
  
  # List of account IDs (lists are OK when there's no name/key)
  accounts = [
    "123456789012",
    "987654321098"
  ]
}
```

---

## Real-World Example

### Scenario: Managing Multiple Teams

```hcl
resource "rafay_eks_cluster_v2" "production" {
  cluster = {
    metadata = {
      name    = "prod-cluster"
      project = "production"
    }
    spec = {
      type           = "aws-eks"
      cloud_provider = "aws-prod-creds"
    }
  }
  
  cluster_config = {
    metadata = {
      name    = "prod-cluster"
      region  = "us-west-2"
      version = "1.28"
    }
    
    # Access entries organized by team
    access_entries = {
      # Platform team - Full admin
      "platform-team-admin" = {
        principal_arn     = "arn:aws:iam::123456789012:role/PlatformTeamRole"
        type              = "STANDARD"
        username          = "platform-admin"
        kubernetes_groups = ["system:masters"]
      }
      
      # Development team - Namespace access
      "dev-team-developer" = {
        principal_arn     = "arn:aws:iam::123456789012:role/DevTeamRole"
        type              = "STANDARD"
        username          = "developer"
        kubernetes_groups = ["developers"]
        access_policies = {
          "namespace-admin" = {
            policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSEditPolicy"
            access_scope = {
              type       = "namespace"
              namespaces = ["dev", "staging"]
            }
          }
        }
      }
      
      # Operations team - Read access
      "ops-team-viewer" = {
        principal_arn     = "arn:aws:iam::123456789012:role/OpsTeamRole"
        type              = "STANDARD"
        username          = "ops-viewer"
        kubernetes_groups = ["viewers"]
        access_policies = {
          "cluster-view" = {
            policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy"
            access_scope = {
              type = "cluster"
            }
          }
        }
      }
      
      # CI/CD pipeline - Deployment access
      "cicd-pipeline" = {
        principal_arn     = "arn:aws:iam::123456789012:role/CICDRole"
        type              = "STANDARD"
        username          = "cicd-deployer"
        kubernetes_groups = ["deployers"]
      }
      
      # EC2 node groups - Node identity
      "node-group-identity" = {
        principal_arn = "arn:aws:iam::123456789012:role/EKSNodeRole"
        type          = "EC2_LINUX"
        username      = "system:node:{{EC2PrivateDNSName}}"
        kubernetes_groups = ["system:bootstrappers", "system:nodes"]
      }
    }
    
    # Identity mappings for aws-auth compatibility
    identity_mappings = {
      arns = {
        "legacy-admin" = {
          arn      = "arn:aws:iam::123456789012:role/LegacyAdminRole"
          username = "legacy-admin"
          groups   = ["system:masters"]
        }
      }
      accounts = ["123456789012"]
    }
  }
}
```

### Update Operation - No Noise!

**You decide to change dev team access**:
```hcl
"dev-team-developer" = {
  principal_arn     = "arn:aws:iam::123456789012:role/DevTeamRole"
  type              = "STANDARD"
  username          = "developer"
  kubernetes_groups = ["developers", "approvers"]  # ← Added "approvers"
  access_policies = {
    "namespace-admin" = {
      policy_arn = "arn:aws:eks::aws:cluster-access-policy/AmazonEKSEditPolicy"
      access_scope = {
        type       = "namespace"
        namespaces = ["dev", "staging", "qa"]  # ← Added "qa"
      }
    }
  }
}
```

**Terraform plan shows** (ONLY what changed):
```diff
  ~ access_entries = {
      ~ "dev-team-developer" = {
          ~ kubernetes_groups = [
              "developers",
            + "approvers",
            ]
          ~ access_policies = {
              ~ "namespace-admin" = {
                  ~ access_scope = {
                      ~ namespaces = [
                          "dev",
                          "staging",
                        + "qa",
                        ]
                    }
                }
            }
        }
      # platform-team-admin: no changes
      # ops-team-viewer: no changes  
      # cicd-pipeline: no changes
      # node-group-identity: no changes
    }
```

**Perfect! No unwanted diffs on other entries!** ✅

---

## Migration from List to Map

### Before (List):
```hcl
access_entries = [
  { principal_arn = "arn:.../Role1", type = "STANDARD", username = "user1" },
  { principal_arn = "arn:.../Role2", type = "STANDARD", username = "user2" },
]
```

### After (Map):
```hcl
access_entries = {
  "role1-entry" = { principal_arn = "arn:.../Role1", type = "STANDARD", username = "user1" }
  "role2-entry" = { principal_arn = "arn:.../Role2", type = "STANDARD", username = "user2" }
}
```

### Tips for Choosing Keys:
1. **Descriptive names**: `"platform-admin"` not `"entry1"`
2. **Stable identifiers**: Use role name, team name, purpose
3. **Avoid ARNs as keys**: Too long, use descriptive names instead
4. **Consistent naming**: `"team-role"`, not mixing styles

---

## Summary

| Aspect | Lists (Old) | Maps (New) |
|--------|------------|-----------|
| **Update middle item** | Shows unwanted diffs ❌ | Only changed item ✅ |
| **Reference by key** | Not possible ❌ | `["admin-role"]` ✅ |
| **Order matters** | Yes ❌ | No ✅ |
| **for_each friendly** | Complex ❌ | Natural ✅ |
| **Self-documenting** | Need comments ❌ | Key describes it ✅ |
| **State stability** | Shifts with changes ❌ | Stable ✅ |
| **Plan noise** | High ❌ | Minimal ✅ |

**Conclusion**: Map-based access entries eliminate unwanted diffs and provide a much better user experience! 🎉

