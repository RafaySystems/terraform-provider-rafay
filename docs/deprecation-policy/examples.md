
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
        Deprecated: "Configure metadata.project instead. This attribute will be removed in v2.0.0 of the provider.",
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
        Deprecated: "Configure resource_tags block instead. This attribute will be removed in v2.0.0 of the provider.",
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
│ Configure metadata.project instead. This attribute will be removed in v2.0.0
│ of the provider.
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
    DeprecationMessage: "Use rafay_eks_cluster, rafay_aks_cluster, or rafay_gke_cluster resource instead. This resource will be removed in v2.0.0 of the provider.",
    
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
│ Use rafay_eks_cluster, rafay_aks_cluster, or rafay_gke_cluster resource
│ instead. This resource will be removed in v2.0.0 of the provider.
```

### 3. Deprecating Data Sources

**Go Implementation (SDKv2):**

```go
package rafay

func dataSourceClusters() *schema.Resource {
  return &schema.Resource{
    DeprecationMessage: "Use rafay_eks_clusters, rafay_aks_clusters, or rafay_gke_clusters data source instead. This data source will be removed in v2.0.0 of the provider.",
    
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
│ Use rafay_eks_clusters, rafay_aks_clusters, or rafay_gke_clusters data source
│ instead. This data source will be removed in v2.0.0 of the provider.
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
                    Deprecated: "Configure network_config block instead. This block will be removed in v2.0.0 of the provider.",
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
    "SPOT": "Use MIXED capacity type with spot configuration instead. The SPOT value will be removed in the next major version of the provider.",
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
│ Warning: capacity_type: Use MIXED capacity type with spot configuration
│ instead. The SPOT value will be removed in the next major version of the
│ provider.
```

### 6. Plugin Framework Deprecation (New Framework Resources)

**Go Implementation (Plugin Framework) - Attribute Deprecation:**

```go
package resource_eks_cluster

import (
  "context"
  "github.com/hashicorp/terraform-plugin-framework/resource"
  "github.com/hashicorp/terraform-plugin-framework/resource/schema"
  "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
  "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

func (r *EKSClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
  resp.Schema = schema.Schema{
    Attributes: map[string]schema.Attribute{
      "project_id": schema.StringAttribute{
        Optional:           true,
        DeprecationMessage: "Configure metadata.project instead. This attribute will be removed in the next major version of the provider.",
        Description:        "DEPRECATED: Use metadata block with project field instead.",
        PlanModifiers: []planmodifier.String{
          stringplanmodifier.UseStateForUnknown(),
        },
      },
    },
    Blocks: map[string]schema.Block{
      "metadata": schema.SingleNestedBlock{
        Description: "Metadata configuration for the cluster.",
        Attributes: map[string]schema.Attribute{
          "project": schema.StringAttribute{
            Required:    true,
            Description: "Project name for the cluster.",
          },
          "name": schema.StringAttribute{
            Required:    true,
            Description: "Cluster name.",
          },
        },
      },
    },
  }
}
```

**Go Implementation (Plugin Framework) - Block Deprecation:**

```go
func (r *EKSClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
  resp.Schema = schema.Schema{
    Blocks: map[string]schema.Block{
      "vpc_config": schema.SingleNestedBlock{
        DeprecationMessage: "Configure network_config block instead. This block will be removed in the next major version of the provider.",
        Description:        "DEPRECATED: Use network_config block for enhanced networking features.",
        Attributes: map[string]schema.Attribute{
          "subnet_ids": schema.ListAttribute{
            Required:    true,
            ElementType: types.StringType,
            Description: "List of subnet IDs.",
          },
        },
      },
      "network_config": schema.SingleNestedBlock{
        Description: "Enhanced network configuration with IPv6 and security group support.",
        Attributes: map[string]schema.Attribute{
          "subnet_ids": schema.ListAttribute{
            Required:    true,
            ElementType: types.StringType,
            Description: "List of subnet IDs for the cluster.",
          },
          "ipv6_enabled": schema.BoolAttribute{
            Optional:    true,
            Description: "Enable IPv6 networking for the cluster.",
          },
          "security_group_ids": schema.ListAttribute{
            Optional:    true,
            ElementType: types.StringType,
            Description: "Additional security group IDs.",
          },
        },
      },
    },
  }
}
```

**Resulting Terraform Warning:**

```
│ Warning: Attribute Deprecated
│ 
│   with rafay_eks_cluster.example,
│   on main.tf line 5, in resource "rafay_eks_cluster" "example":
│    5:   project_id = "my-project"
│ 
│ Configure metadata.project instead. This attribute will be removed in the
│ next major version of the provider.
```

**Best Practices for Plugin Framework Deprecations:**

1. **Practitioner-Focused Messaging**: Use clear, actionable language that tells users what to do (e.g., "Configure X instead" not "Attribute X is deprecated")
2. **Consistent Format**: Follow the pattern: "Configure {new_option} instead. This {attribute|block} will be removed in the next major version of the provider."
3. **Both Attributes and Blocks**: The `DeprecationMessage` field works for both schema attributes and blocks
4. **Maintain Functionality**: Deprecated features must remain fully functional during the grace period
5. **Update Documentation**: Mark deprecated fields in the description as well for better visibility

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