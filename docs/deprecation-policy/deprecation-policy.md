# Deprecation Policy and Semantic Versioning

## Overview

This document outlines the deprecation policy and semantic versioning for the Rafay Terraform Provider, establishing clear guidelines for managing breaking changes, version upgrades, and maintaining backward compatibility. This policy ensures predictable upgrade paths and maintains user trust during provider evolution.

## Semantic Versioning Strategy

The Rafay Terraform Provider follows strict Semantic Versioning (SemVer) using the format `MAJOR.MINOR.PATCH`:

### MAJOR Version Changes (Breaking Changes)

**When to increment:** Incompatible API changes that break existing configurations.

**Examples:**
- Removing resources (e.g., removing `rafay_legacy_cluster`)
- Removing or renaming resource fields (e.g., removing `project_id` attribute from `rafay_cluster`)
- Changing default behavior that breaks existing configurations
- Schema changes that require state migration
- Removing/renaming resource data sources/types
- Changing resource names (config refactoring)
- Changing required vs optional field status

**Version Example:** `1.1.51` → `2.0.0`

### MINOR Version Changes (Backward Compatible)

**When to increment:** Adding new functionality in a backward-compatible manner.

**Examples:**
- Adding new resources (e.g., adding `rafay_environment_template`)
- Adding new optional arguments to existing resources
- Adding data sources
- New resources that don't break existing configurations
- Adding new computed attributes

**Version Example:** `1.1.51` → `1.2.0`

### PATCH Version Changes (Bug Fixes)

**When to increment:** Backward-compatible bug fixes and improvements.

**Examples:**
- Fixing diff suppression bugs
- Correcting documentation
- Tightening validation without breaking existing usage
- Patching crashes without breaking existing usage
- Fixing `terraform import` issues
- Performance improvements without behavior changes

**Version Example:** `1.1.51` → `1.1.52`

## Deprecation Process

Following the [AWS provider deprecation model](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/version-2-upgrade) and [version 3 upgrade guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/version-3-upgrade), our deprecation process ensures users have adequate time to migrate.

### Deprecation Timeline

Based on AWS provider patterns observed in their [release history](https://github.com/hashicorp/terraform-provider-aws/releases):

1. **Deprecation Announcement** (Version N)
   - Announce deprecations with at least one minor version lead time
   - Add deprecation warnings to provider logs using standardized format
   - Update documentation with deprecation notices and migration paths
   - Include in `CHANGELOG.md` under **DEPRECATIONS** section
   - Provide automated detection tools where possible

2. **Grace Period** (Version N+1 to N+X)
   - Maintain full backward compatibility
   - Provide comprehensive migration examples and upgrade guides
   - Continue deprecation warnings with version-specific messaging
   - Offer automated migration tools and state migration utilities

3. **Breaking Change Implementation** (Next Major Version)
   - Implement breaking changes in major version releases only
   - Remove deprecated functionality with clear upgrade documentation
   - Provide automated state migration where technically feasible
   - Include detailed upgrade guides with before/after examples

### Minimum Deprecation Periods

Following AWS provider standards:

- **Resources:** Minimum 6 months or 2 minor versions, whichever is longer
- **Arguments/Attributes:** Minimum 3 months or 1 minor version, whichever is longer  
- **Data Sources:** Minimum 3 months or 1 minor version, whichever is longer
- **Provider Configuration:** Minimum 9 months or 3 minor versions, whichever is longer
- **Default Behavior Changes:** Minimum 6 months with opt-in mechanisms where possible

## Implementation Guide: Go Code to Terraform Resource Mapping

This section demonstrates how Go code deprecation warnings translate to user-facing Terraform warnings.

### 1. Deprecating Individual Fields/Arguments

**Go Implementation (SDKv2):**

```go
package rafay

import (
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAKSCluster() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      "project_id": {
        Type:       schema.TypeString,
        Optional:   true,
        Deprecated: "Argument `project_id` is deprecated and will be removed in v2.0.0. Use `metadata.project` instead for consistency with Kubernetes conventions.",
        Description: "DEPRECATED: Use metadata.project instead.",
      },
      "metadata": {
        Type:     schema.TypeList,
        Optional: true,
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "project": {
              Type:        schema.TypeString,
              Required:    true,
              Description: "Project name for the cluster.",
            },
            // ... other metadata fields
          },
        },
      },
      "tags": {
        Type:       schema.TypeMap,
        Optional:   true,
        Deprecated: "Argument `tags` is deprecated and will be removed in v2.0.0. Use `resource_tags` block instead for enhanced tagging capabilities with propagation support.",
        Elem: &schema.Schema{
          Type: schema.TypeString,
        },
      },
      "resource_tags": {
        Type:     schema.TypeList,
        Optional: true,
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "tags": {
              Type:     schema.TypeMap,
              Optional: true,
              Elem: &schema.Schema{
                Type: schema.TypeString,
              },
            },
            "propagate_to_resources": {
              Type:        schema.TypeBool,
              Optional:    true,
              Default:     false,
              Description: "Propagate tags to underlying Azure resources.",
            },
          },
        },
      },
    },
  }
}
```

**Resulting Terraform Warning:**

```
│ Warning: Argument is deprecated
│ 
│   on main.tf line 5, in resource "rafay_aks_cluster" "example":
│    5:   project_id = "my-project"
│ 
│ Argument `project_id` is deprecated and will be removed in v2.0.0. Use
│ `metadata.project` instead for consistency with Kubernetes conventions.
```

**User-Facing Terraform Configuration:**

```terraform
resource "rafay_aks_cluster" "example" {
  # DEPRECATED: Will generate warning during terraform plan/apply
  project_id = "my-project"
  
  # NEW: Recommended approach
  metadata {
    name    = "my-cluster"
    project = "my-project"
  }
}
```

### 2. Deprecating Entire Resources

**Go Implementation (SDKv2):**

```go
package rafay

func resourceCluster() *schema.Resource {
  return &schema.Resource{
    DeprecationMessage: "Resource `rafay_cluster` is deprecated and will be removed in v2.0.0. Use `rafay_eks_cluster`, `rafay_aks_cluster`, or `rafay_gke_cluster` instead for enhanced functionality and cloud-specific features.",
    
    Schema: map[string]*schema.Schema{
      "name": {
        Type:        schema.TypeString,
        Required:    true,
        Description: "Cluster name.",
      },
      "project_id": {
        Type:        schema.TypeString,
        Required:    true,
        Description: "Project ID.",
      },
      // ... other legacy fields
    },
    
    CreateContext: resourceClusterCreate,
    ReadContext:   resourceClusterRead,
    UpdateContext: resourceClusterUpdate,
    DeleteContext: resourceClusterDelete,
  }
}
```

**Resulting Terraform Warning:**

```
│ Warning: Resource is deprecated
│ 
│   on main.tf line 1, in resource "rafay_cluster" "example":
│    1: resource "rafay_cluster" "example" {
│ 
│ Resource `rafay_cluster` is deprecated and will be removed in v2.0.0. Use
│ `rafay_eks_cluster`, `rafay_aks_cluster`, or `rafay_gke_cluster` instead for
│ enhanced functionality and cloud-specific features.
```

### 3. Deprecating Data Sources

**Go Implementation (SDKv2):**

```go
package rafay

func dataSourceClusters() *schema.Resource {
  return &schema.Resource{
    DeprecationMessage: "Data source `rafay_clusters` is deprecated and will be removed in v2.0.0. Use cloud-specific data sources `rafay_eks_clusters`, `rafay_aks_clusters`, or `rafay_gke_clusters` instead for better performance and enhanced filtering.",
    
    ReadContext: dataSourceClustersRead,
    
    Schema: map[string]*schema.Schema{
      "project": {
        Type:        schema.TypeString,
        Required:    true,
        Description: "Project name.",
      },
      // ... other fields
    },
  }
}
```

**Resulting Terraform Warning:**

```
│ Warning: Data source is deprecated
│ 
│   on main.tf line 10, in data "rafay_clusters" "all":
│   10: data "rafay_clusters" "all" {
│ 
│ Data source `rafay_clusters` is deprecated and will be removed in v2.0.0. Use
│ cloud-specific data sources `rafay_eks_clusters`, `rafay_aks_clusters`, or
│ `rafay_gke_clusters` instead for better performance and enhanced filtering.
```

### 4. Deprecating Nested Blocks

**Go Implementation (SDKv2):**

```go
func resourceEKSCluster() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      "spec": {
        Type:     schema.TypeList,
        Required: true,
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "cluster_config": {
              Type:     schema.TypeList,
              Required: true,
              MaxItems: 1,
              Elem: &schema.Resource{
                Schema: map[string]*schema.Schema{
                  "vpc_config": {
                    Type:       schema.TypeList,
                    Optional:   true,
                    Deprecated: "Block `vpc_config` is deprecated and will be removed in v2.0.0. Use `network_config` block instead for enhanced networking features including IPv6 and security group management.",
                    MaxItems:   1,
                    Elem: &schema.Resource{
                      Schema: map[string]*schema.Schema{
                        "subnet_ids": {
                          Type:     schema.TypeList,
                          Required: true,
                          Elem: &schema.Schema{
                            Type: schema.TypeString,
                          },
                        },
                      },
                    },
                  },
                  "network_config": {
                    Type:     schema.TypeList,
                    Optional: true,
                    MaxItems: 1,
                    Elem: &schema.Resource{
                      Schema: map[string]*schema.Schema{
                        "subnet_ids": {
                          Type:     schema.TypeList,
                          Required: true,
                          Elem: &schema.Schema{
                            Type: schema.TypeString,
                          },
                        },
                        "ipv6_enabled": {
                          Type:     schema.TypeBool,
                          Optional: true,
                          Default:  false,
                        },
                        "security_group_ids": {
                          Type:     schema.TypeList,
                          Optional: true,
                          Elem: &schema.Schema{
                            Type: schema.TypeString,
                          },
                        },
                      },
                    },
                  },
                },
              },
            },
          },
        },
      },
    },
  }
}
```

### 5. Deprecating Specific Values with Validation

**Go Implementation (SDKv2):**

```go
package rafay

import (
  "fmt"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validateCapacityType(val interface{}, key string) (warns []string, errs []error) {
  v := val.(string)
  
  deprecatedValues := map[string]string{
    "SPOT": "Value `SPOT` for `capacity_type` is deprecated. Use `MIXED` with spot configuration for better cost optimization and availability.",
  }
  
  if msg, deprecated := deprecatedValues[v]; deprecated {
    warns = append(warns, fmt.Sprintf("%s: %s", key, msg))
  }
  
  validValues := []string{"ON_DEMAND", "SPOT", "MIXED"}
  for _, valid := range validValues {
    if v == valid {
      return warns, errs
    }
  }
  
  errs = append(errs, fmt.Errorf("%s must be one of %v, got: %s", key, validValues, v))
  return warns, errs
}

func resourceEKSClusterNodeGroup() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      "capacity_type": {
        Type:         schema.TypeString,
        Optional:     true,
        Default:      "ON_DEMAND",
        ValidateFunc: validateCapacityType,
        Description:  "Capacity type for node group: ON_DEMAND, SPOT (deprecated), or MIXED.",
      },
    },
  }
}
```

**Resulting Terraform Warning:**

```
│ Warning: capacity_type: Value `SPOT` for `capacity_type` is deprecated. Use
│ `MIXED` with spot configuration for better cost optimization and availability.
```

### 6. Plugin Framework Deprecation (New Framework Resources)

**Go Implementation (Plugin Framework):**

```go
package resource_eks_cluster

import (
  "context"
  "github.com/hashicorp/terraform-plugin-framework/resource/schema"
  "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
  "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *EKSClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
  resp.Schema = schema.Schema{
    Attributes: map[string]schema.Attribute{
      "project_id": schema.StringAttribute{
        Optional:           true,
        DeprecationMessage: "Attribute `project_id` is deprecated and will be removed in v2.0.0. Use `metadata.project` instead.",
        PlanModifiers: []planmodifier.String{
          stringplanmodifier.UseStateForUnknown(),
        },
      },
    },
    Blocks: map[string]schema.Block{
      "metadata": schema.SingleNestedBlock{
        Attributes: map[string]schema.Attribute{
          "project": schema.StringAttribute{
            Required:    true,
            Description: "Project name for the cluster.",
          },
        },
      },
    },
  }
}
```

### 7. State Migration for Deprecated Resources

**Go Implementation (SDKv2):**

```go
package rafay

func resourceAKSClusterV0() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      "project_id": {
        Type:     schema.TypeString,
        Required: true,
      },
      "tags": {
        Type:     schema.TypeMap,
        Optional: true,
        Elem: &schema.Schema{
          Type: schema.TypeString,
        },
      },
    },
  }
}

func resourceAKSCluster() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      "metadata": {
        Type:     schema.TypeList,
        Required: true,
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "project": {
              Type:     schema.TypeString,
              Required: true,
            },
          },
        },
      },
      "resource_tags": {
        Type:     schema.TypeList,
        Optional: true,
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "tags": {
              Type:     schema.TypeMap,
              Optional: true,
              Elem: &schema.Schema{
                Type: schema.TypeString,
              },
            },
          },
        },
      },
    },
    
    SchemaVersion: 1,
    StateUpgraders: []schema.StateUpgrader{
      {
        Type:    resourceAKSClusterV0().CoreConfigSchema().ImpliedType(),
        Upgrade: resourceAKSClusterStateUpgradeV0,
        Version: 0,
      },
    },
  }
}

func resourceAKSClusterStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
  // Migrate project_id to metadata.project
  if projectID, ok := rawState["project_id"].(string); ok {
    rawState["metadata"] = []interface{}{
      map[string]interface{}{
        "project": projectID,
      },
    }
    delete(rawState, "project_id")
  }
  
  // Migrate tags to resource_tags
  if tags, ok := rawState["tags"].(map[string]interface{}); ok {
    rawState["resource_tags"] = []interface{}{
      map[string]interface{}{
        "tags": tags,
      },
    }
    delete(rawState, "tags")
  }
  
  return rawState, nil
}
```

### Complete Deprecation Lifecycle Example

**Phase 1: v1.5.0 - Add Deprecation Warnings**

```go
// rafay/resource_aks_cluster.go
func resourceAKSCluster() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      "project_id": {
        Type:       schema.TypeString,
        Optional:   true,
        Deprecated: "Deprecated in v1.5.0, will be removed in v2.0.0. Use metadata.project instead.",
      },
      "metadata": {
        Type:     schema.TypeList,
        Optional: true,  // Still optional to maintain backward compatibility
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "project": {
              Type:     schema.TypeString,
              Required: true,
            },
          },
        },
      },
    },
    SchemaVersion: 0,
  }
}
```

**Phase 2: v1.6.0 - Add State Upgrader**

```go
// rafay/resource_aks_cluster.go
func resourceAKSCluster() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      "project_id": {
        Type:       schema.TypeString,
        Optional:   true,
        Deprecated: "Deprecated in v1.5.0, will be removed in v2.0.0. Use metadata.project instead.",
      },
      "metadata": {
        Type:     schema.TypeList,
        Optional: true,
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "project": {
              Type:     schema.TypeString,
              Required: true,
            },
          },
        },
      },
    },
    SchemaVersion: 1,
    StateUpgraders: []schema.StateUpgrader{
      {
        Type:    resourceAKSClusterV0().CoreConfigSchema().ImpliedType(),
        Upgrade: resourceAKSClusterStateUpgradeV0,
        Version: 0,
      },
    },
  }
}
```

**Phase 3: v2.0.0 - Remove Deprecated Field**

```go
// rafay/resource_aks_cluster.go
func resourceAKSCluster() *schema.Resource {
  return &schema.Resource{
    Schema: map[string]*schema.Schema{
      // project_id field removed entirely
      "metadata": {
        Type:     schema.TypeList,
        Required: true,  // Now required since old field is removed
        MaxItems: 1,
        Elem: &schema.Resource{
          Schema: map[string]*schema.Schema{
            "project": {
              Type:     schema.TypeString,
              Required: true,
            },
          },
        },
      },
    },
    SchemaVersion: 1,  // Keep schema version for historical state migrations
    StateUpgraders: []schema.StateUpgrader{
      {
        Type:    resourceAKSClusterV0().CoreConfigSchema().ImpliedType(),
        Upgrade: resourceAKSClusterStateUpgradeV0,
        Version: 0,
      },
    },
  }
}
```

## Deprecation Examples

### Resource Deprecation

Following AWS provider patterns for resource lifecycle management:

```terraform
# DEPRECATED: rafay_cluster resource (Deprecated in v1.2.0, removed in v2.0.0)
# Use rafay_eks_cluster instead for enhanced functionality and better AWS integration
resource "rafay_cluster" "example" {
  name       = "my-cluster"
  project_id = "my-project"
  
  # Legacy configuration structure
  config = {
    kubernetes_version = "1.24"
    node_count        = 3
  }
}

# NEW: Recommended approach using enhanced resource structure
resource "rafay_eks_cluster" "example" {
  metadata {
    name    = "my-cluster"
    project = "my-project"
    labels = {
      environment = "production"
      team        = "platform"
    }
  }
  
  spec {
    cluster_config {
      version = "1.24"
      
      # Enhanced networking configuration
      vpc_config {
        subnet_ids = ["subnet-12345", "subnet-67890"]
      }
      
      # Improved node group management
      node_groups {
        name          = "primary"
        instance_type = "m5.large"
        desired_size  = 3
        min_size      = 1
        max_size      = 10
      }
    }
  }
}
```

### Argument Deprecation

Based on AWS provider argument restructuring patterns:

```terraform
resource "rafay_aks_cluster" "example" {
  # DEPRECATED: project_id argument (Deprecated in v1.3.0, removed in v2.0.0)
  # Use metadata.project instead for consistency with Kubernetes conventions
  project_id = "my-project"  # Will generate deprecation warning
  
  # DEPRECATED: Simple tags map (Deprecated in v1.3.0, removed in v2.0.0)  
  # Use resource_tags block for enhanced tagging capabilities
  tags = {
    Environment = "production"
    Team        = "platform"
  }
  
  # NEW: Recommended metadata structure
  metadata {
    name    = "my-cluster"
    project = "my-project"  # Replaces deprecated project_id
    
    # Enhanced labeling system
    labels = {
      "rafay.io/environment" = "production"
      "rafay.io/team"        = "platform"
      "rafay.io/cost-center" = "engineering"
    }
  }
  
  # NEW: Enhanced resource tagging (replaces simple tags map)
  resource_tags {
    tags = {
      Environment   = "production"
      Team         = "platform"
      CostCenter   = "engineering"
      ManagedBy    = "terraform"
    }
    
    # Propagate tags to underlying Azure resources
    propagate_to_resources = true
  }
  
  spec {
    cluster_config {
      # Configuration remains compatible
      kubernetes_version = "1.24"
      
      # Enhanced node pool configuration
      node_pools {
        name = "system"
        mode = "System"
        
        # Improved scaling configuration
        auto_scaling {
          enabled   = true
          min_count = 1
          max_count = 10
        }
      }
    }
  }
}
```

### Data Source Deprecation

Following AWS provider data source evolution patterns:

```terraform
# DEPRECATED: rafay_clusters data source (Deprecated in v1.4.0, removed in v2.0.0)
# Use type-specific data sources for better performance and filtering
data "rafay_clusters" "example" {
  project = "my-project"
  
  # Limited filtering capabilities
  filter {
    name   = "status"
    values = ["READY"]
  }
}

# NEW: Type-specific data sources with enhanced filtering
data "rafay_eks_clusters" "production_eks" {
  metadata {
    project = "my-project"
  }
  
  # Enhanced filtering with multiple criteria
  filter {
    cluster_status = ["READY", "UPDATING"]
    
    # Filter by Kubernetes version range
    kubernetes_version_min = "1.23"
    kubernetes_version_max = "1.26"
    
    # Filter by labels
    label_selector = {
      "environment" = "production"
      "team"        = "platform"
    }
  }
  
  # Improved sorting and pagination
  sort_by    = "created_at"
  sort_order = "desc"
  limit      = 50
}

data "rafay_aks_clusters" "production_aks" {
  metadata {
    project = "my-project"
  }
  
  # AKS-specific filtering capabilities
  filter {
    cluster_status = ["READY"]
    
    # Filter by Azure region
    azure_region = "East US 2"
    
    # Filter by node pool configuration
    min_node_count = 3
    max_node_count = 100
  }
}
```

### Provider Configuration Changes

Following AWS provider configuration evolution:

```terraform
# DEPRECATED: Legacy provider configuration (Deprecated in v1.5.0, removed in v2.0.0)
provider "rafay" {
  # DEPRECATED: Individual authentication fields
  api_key    = var.rafay_api_key     # Use api_credentials block instead
  api_secret = var.rafay_api_secret  # Use api_credentials block instead
  
  # DEPRECATED: Simple endpoint configuration
  console_url = "https://console.rafay.dev"  # Use endpoints block instead
  
  # DEPRECATED: Basic retry configuration
  max_retries = 3  # Use retry_config block for enhanced control
}

# NEW: Enhanced provider configuration with structured blocks
provider "rafay" {
  # NEW: Structured API credentials with multiple authentication methods
  api_credentials {
    # Option 1: API Key authentication
    api_key    = var.rafay_api_key
    api_secret = var.rafay_api_secret
    
    # Option 2: Service account authentication (new capability)
    # service_account_key_file = "/path/to/service-account.json"
    
    # Option 3: OIDC authentication (new capability)  
    # oidc_token_source = "environment"
  }
  
  # NEW: Structured endpoint configuration
  endpoints {
    console_url = "https://console.rafay.dev"
    api_url     = "https://api.rafay.dev"
    
    # Regional endpoint support (new capability)
    region = "us-west-2"
    
    # Custom endpoints for private deployments
    # custom_endpoints = {
    #   console = "https://rafay.internal.company.com"
    #   api     = "https://api.rafay.internal.company.com"
    # }
  }
  
  # NEW: Enhanced retry and timeout configuration
  retry_config {
    max_retries      = 5
    retry_delay_base = "1s"
    retry_delay_max  = "30s"
    
    # Exponential backoff configuration
    backoff_multiplier = 2.0
    
    # Jitter to prevent thundering herd
    enable_jitter = true
  }
  
  # NEW: Request timeout configuration
  timeout_config {
    default_timeout = "30s"
    
    # Operation-specific timeouts
    create_timeout = "10m"
    update_timeout = "10m"
    delete_timeout = "5m"
  }
  
  # NEW: Enhanced logging and debugging
  logging_config {
    level = "INFO"  # DEBUG, INFO, WARN, ERROR
    
    # Log request/response for debugging
    log_requests  = false
    log_responses = false
    
    # Structured logging format
    format = "json"  # json, text
  }
}
```

### Default Behavior Changes

Following AWS provider patterns for behavioral changes:

```terraform
# Example: Node group defaults evolution (similar to AWS EKS defaults)

# BEFORE v2.0.0: Manual node group configuration required
resource "rafay_eks_cluster" "example" {
  metadata {
    name    = "my-cluster"
    project = "my-project"
  }
  
  spec {
    cluster_config {
      version = "1.24"
      
      # BEFORE: Manual node group configuration was required
      node_groups {
        name          = "primary"
        instance_type = "m5.large"
        desired_size  = 3
        min_size      = 1
        max_size      = 10
        
        # Manual AMI selection required
        ami_type = "AL2_x86_64"
        
        # Manual capacity type selection
        capacity_type = "ON_DEMAND"
      }
    }
  }
}

# AFTER v2.0.0: Intelligent defaults with opt-out capability
resource "rafay_eks_cluster" "example" {
  metadata {
    name    = "my-cluster" 
    project = "my-project"
  }
  
  spec {
    cluster_config {
      version = "1.24"
      
      # NEW: Automatic default node group creation (can be disabled)
      auto_create_node_group = true  # Default: true (breaking change)
      
      # NEW: Intelligent defaults based on cluster configuration
      default_node_group {
        # Automatically selects appropriate instance types based on workload hints
        auto_instance_selection = true
        
        # Intelligent scaling based on cluster usage patterns  
        auto_scaling_policy = "balanced"  # cost_optimized, performance_optimized, balanced
        
        # Automatic AMI selection based on Kubernetes version
        auto_ami_selection = true
        
        # Mixed capacity for cost optimization (new default behavior)
        capacity_type = "MIXED"  # Changed from ON_DEMAND default
        
        # Spot instance configuration for mixed capacity
        spot_allocation_strategy = "price-capacity-optimized"
        spot_max_price_percentage = 50  # % of On-Demand price
      }
      
      # Override defaults when needed
      node_groups {
        name = "custom-workload"
        
        # Explicit configuration overrides defaults
        instance_type = "c5.xlarge"
        capacity_type = "ON_DEMAND"
        
        # Custom node group still benefits from new features
        auto_ami_selection = true
      }
    }
  }
}

# Migration helper for preserving v1.x behavior
resource "rafay_eks_cluster" "legacy_behavior" {
  metadata {
    name    = "legacy-cluster"
    project = "my-project"
  }
  
  spec {
    cluster_config {
      version = "1.24"
      
      # Disable new default behaviors to preserve v1.x compatibility
      auto_create_node_group = false
      
      # Explicit node group configuration (v1.x style)
      node_groups {
        name          = "primary"
        instance_type = "m5.large"
        desired_size  = 3
        min_size      = 1
        max_size      = 10
        
        # Explicit v1.x defaults
        ami_type      = "AL2_x86_64"
        capacity_type = "ON_DEMAND"
        
        # Opt out of new automatic features
        auto_ami_selection = false
      }
    }
  }
}

## Version Support Policy

### Support Windows
- **Current Version (N):** Full support including new features and bug fixes
- **Previous Version (N-1):** Security updates and critical bug fixes only
- **Older Versions (N-2 and below):** End-of-life, no support

### End-of-Life Timeline
- **Minor Versions:** Supported for 12 months after release
- **Major Versions:** Previous major version supported for 18 months after new major release
- **Security Updates:** Critical security fixes backported to N-1 for 6 months

### Migration Assistance
- Automated migration tools for common scenarios
- Step-by-step migration guides
- Professional services for complex migrations

## Communication Strategy

### CHANGELOG.md Integration

Based on AWS provider [changelog patterns](https://github.com/hashicorp/terraform-provider-aws/releases), automated changelog generation via GitHub Actions with clear categorization:

- **BREAKING CHANGES:** Schema changes, resource removals, behavior changes requiring user action
- **FEATURES:** New resources, data sources, and major functionality additions
- **ENHANCEMENTS:** Improvements to existing resources, performance optimizations, new optional arguments
- **BUG FIXES:** Issue resolutions, crash fixes, and patches
- **DEPRECATIONS:** Advance notice of upcoming changes with removal timelines
- **DOCUMENTATION:** Documentation updates, example improvements, and clarifications

### AWS Provider Changelog Format Example

Following the standardized format from AWS provider releases:

```markdown
## 2.0.0 (January 15, 2025)

BREAKING CHANGES:

* provider: Remove deprecated `project_id` argument from all cluster resources. Use `metadata.project` instead ([#123](https://github.com/RafaySystems/terraform-provider-rafay/issues/123))
* resource/rafay_eks_cluster: Change default `capacity_type` from `ON_DEMAND` to `MIXED` for cost optimization ([#124](https://github.com/RafaySystems/terraform-provider-rafay/issues/124))
* resource/rafay_aks_cluster: Remove deprecated `tags` argument. Use `resource_tags` block instead ([#125](https://github.com/RafaySystems/terraform-provider-rafay/issues/125))

FEATURES:

* **New Resource:** `rafay_environment_template` ([#126](https://github.com/RafaySystems/terraform-provider-rafay/issues/126))
* **New Data Source:** `rafay_cost_profiles` ([#127](https://github.com/RafaySystems/terraform-provider-rafay/issues/127))
* resource/rafay_eks_cluster: Add `auto_create_node_group` argument for intelligent defaults ([#128](https://github.com/RafaySystems/terraform-provider-rafay/issues/128))

ENHANCEMENTS:

* resource/rafay_aks_cluster: Add `resource_tags` block with propagation support ([#129](https://github.com/RafaySystems/terraform-provider-rafay/issues/129))
* data-source/rafay_eks_clusters: Add enhanced filtering with `label_selector` and version ranges ([#130](https://github.com/RafaySystems/terraform-provider-rafay/issues/130))
* provider: Add structured `api_credentials`, `endpoints`, and `retry_config` blocks ([#131](https://github.com/RafaySystems/terraform-provider-rafay/issues/131))

BUG FIXES:

* resource/rafay_eks_cluster: Fix state inconsistency when node groups are modified outside Terraform ([#132](https://github.com/RafaySystems/terraform-provider-rafay/issues/132))
* resource/rafay_aks_cluster: Prevent `terraform import` failures for clusters with custom node pools ([#133](https://github.com/RafaySystems/terraform-provider-rafay/issues/133))

DEPRECATIONS:

* data-source/rafay_clusters: Deprecate in favor of type-specific data sources `rafay_eks_clusters` and `rafay_aks_clusters`. Will be removed in v3.0.0 ([#134](https://github.com/RafaySystems/terraform-provider-rafay/issues/134))
```

### User Notification Process

Following AWS provider communication patterns:

1. **Provider Log Warnings**
   ```
   [WARN] rafay_cluster resource is deprecated and will be removed in version 2.0.0. Use rafay_eks_cluster instead. See upgrade guide: https://registry.terraform.io/providers/RafaySystems/rafay/latest/docs/guides/version-2-upgrade
   
   [WARN] rafay_aks_cluster.example: Argument "project_id" is deprecated and will be removed in version 2.0.0. Use "metadata.project" instead.
   
   [WARN] rafay_aks_cluster.example: Argument "tags" is deprecated and will be removed in version 2.0.0. Use "resource_tags" block instead for enhanced tagging capabilities.
   ```

2. **Documentation Updates**
   - Deprecation notices prominently displayed in resource documentation
   - Migration examples with side-by-side comparisons
   - Clear timelines with specific version numbers
   - Links to upgrade guides and automated migration tools

3. **GitHub Releases**
   - Detailed upgrade notes with each release following AWS provider format
   - Direct links to migration guides and documentation
   - Highlight breaking changes and deprecations in release notes
   - Provide downloadable migration scripts where applicable

## State Migration Support

### Automatic State Migration

For compatible schema changes:
- Built-in state migration during `terraform plan`
- Validation and integrity checks
- Automatic backup creation before migration
- Rollback procedures for failed migrations

### Manual Migration Procedures

For complex changes requiring user intervention, following AWS provider migration patterns:

1. **Resource Renames**
   ```bash
   # Example: Migrating from rafay_cluster to rafay_eks_cluster
   
   # Step 1: Remove the old resource from state (without destroying)
   terraform state rm rafay_cluster.example
   
   # Step 2: Import the existing cluster with new resource type
   terraform import rafay_eks_cluster.example cluster-id
   
   # Step 3: Update configuration file to use new resource
   # (See deprecation examples above for configuration changes)
   
   # Step 4: Verify the plan shows no changes
   terraform plan
   ```

2. **Argument Restructuring**
   ```bash
   # Example: Migrating from project_id to metadata.project
   
   # Step 1: Update configuration file
   # Change from:
   #   project_id = "my-project"
   # To:
   #   metadata {
   #     project = "my-project"
   #   }
   
   # Step 2: Run terraform plan to validate changes
   terraform plan
   
   # Step 3: Apply changes (should show in-place update)
   terraform apply
   ```

3. **Provider Configuration Migration**
   ```bash
   # Example: Migrating to structured provider configuration
   
   # Step 1: Update provider block configuration
   # (See provider configuration examples above)
   
   # Step 2: Re-initialize Terraform
   terraform init -upgrade
   
   # Step 3: Verify connectivity with new configuration
   terraform plan
   ```

4. **Complex Data Structure Migration**
   ```bash
   # Example: Migrating from simple tags to resource_tags block
   
   # Step 1: Export current state for backup
   terraform show -json > backup-state.json
   
   # Step 2: Update configuration to use new structure
   # (See argument deprecation examples above)
   
   # Step 3: Plan and apply changes
   terraform plan
   terraform apply
   
   # Step 4: Verify tags are correctly applied to cloud resources
   # Check via cloud provider console or CLI
   ```

### Automated Migration Tools

Following AWS provider automation patterns:

```bash
# Rafay Provider Migration CLI Tool (planned for v2.0 release)

# Check for deprecated usage in current configuration
rafay-migrate scan --path ./terraform/

# Generate migration plan
rafay-migrate plan --from-version 1.x --to-version 2.0

# Apply automated migrations where possible
rafay-migrate apply --backup-state

# Validate migration results
rafay-migrate validate --post-migration-check
```

## Upgrade Documentation Structure

```
docs/
├── guides/
│   ├── version-2-upgrade.md      # Major version upgrade guide
│   ├── version-1.5-upgrade.md    # Minor version with deprecations
│   └── migration-examples/       # Specific migration scenarios
│       ├── cluster-migration.md
│       ├── rbac-migration.md
│       └── state-migration.md
├── UPGRADE.md                    # Current upgrade notes
└── CHANGELOG.md                  # Automated changelog
```

## Implementation Guidelines

### For Developers

1. **Before Deprecating:**
   - Ensure replacement functionality exists
   - Create comprehensive migration documentation
   - Implement deprecation warnings
   - Update tests to cover both old and new approaches

2. **During Deprecation Period:**
   - Maintain backward compatibility
   - Provide migration assistance
   - Update documentation regularly

3. **Before Removal:**
   - Confirm adequate notice period has passed
   - Verify migration paths are well-documented
   - Ensure automated tools are available
   - Coordinate with release management

### For Users

1. **Stay Informed:**
   - Subscribe to release notifications
   - Review changelog for each update
   - Monitor provider logs for warnings

2. **Plan Migrations:**
   - Test migrations in non-production environments
   - Review upgrade guides before applying updates
   - Schedule migration windows appropriately
   - Maintain backup configurations

## Exceptions and Special Cases

### Emergency Deprecations

In rare cases where security vulnerabilities or critical bugs require immediate action:
- Minimum 30-day notice for emergency deprecations
- Immediate patch release with fixes
- Accelerated migration assistance
- Clear communication about urgency

## Compliance and Monitoring

### Deprecation Tracking

- Automated tracking of deprecated features
- Regular reviews of deprecation timelines
- Metrics on migration adoption rates
- User feedback collection and analysis

### Quality Assurance

- All deprecations must include migration paths
- Automated testing of upgrade scenarios
- Documentation review requirements

## Contact and Support

For questions about deprecations or migration assistance:

- **Documentation:** [Provider Documentation](https://registry.terraform.io/providers/RafaySystems/rafay/latest/docs)
- **GitHub Issues:** [Rafay TF Provider](https://github.com/RafaySystems/terraform-provider-rafay/issues)
- **Professional Services:** [Book a Demo](https://rafay.co/)

---

**Document Version:** 1.0  
**Last Updated:** October 2024  
**Next Review:** January 2025

This deprecation policy ensures predictable, user-friendly evolution of the Rafay Terraform Provider while maintaining stability and trust in production environments.
