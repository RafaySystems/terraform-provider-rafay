# EKS Cluster V2 Resource - Plugin Framework with Maps

This is a complete rewrite of the EKS cluster resource using Terraform Plugin Framework with map-based schema instead of list-based schema.

## Key Changes from SDK v2 Version

### 1. Framework Migration
- **Old**: Terraform Plugin SDK v2 (`github.com/hashicorp/terraform-plugin-sdk/v2`)
- **New**: Terraform Plugin Framework (`github.com/hashicorp/terraform-plugin-framework`)

### 2. Schema Structure

#### Lists → Single Nested Objects
Blocks that were lists with `MinItems: 1, MaxItems: 1` are now single nested objects:

**Old (SDK v2)**:
```hcl
cluster {
  metadata {
    name    = "my-cluster"
    project = "my-project"
  }
}
```

**New (Plugin Framework)**:
```hcl
cluster = {
  metadata = {
    name    = "my-cluster"
    project = "my-project"
  }
}
```

#### Lists → Maps for Collections
Collections like node groups and subnets now use maps with meaningful keys:

**Old (SDK v2)**:
```hcl
node_groups = [
  {
    name = "ng-1"
    ...
  },
  {
    name = "ng-2"
    ...
  }
]
```

**New (Plugin Framework)**:
```hcl
node_groups = {
  "ng-1" = {
    name = "ng-1"
    ...
  }
  "ng-2" = {
    name = "ng-2"
    ...
  }
}
```

#### Subnet Configuration
Subnets are organized by availability zone:

**Old**:
```hcl
subnets = {
  public = [
    { id = "subnet-1", cidr = "10.0.1.0/24", az = "us-west-2a" },
    { id = "subnet-2", cidr = "10.0.2.0/24", az = "us-west-2b" }
  ]
}
```

**New**:
```hcl
subnets = {
  public = {
    "us-west-2a" = { id = "subnet-1", cidr = "10.0.1.0/24", az = "us-west-2a" }
    "us-west-2b" = { id = "subnet-2", cidr = "10.0.2.0/24", az = "us-west-2b" }
  }
}
```

### 3. Project Sharing
Now uses map for projects:

**Old**:
```hcl
sharing {
  enabled = true
  projects = [
    { name = "project1" },
    { name = "project2" }
  ]
}
```

**New**:
```hcl
sharing = {
  enabled = true
  projects = {
    "project1" = { name = "project1" }
    "project2" = { name = "project2" }
  }
}
```

### 4. Taints
Taints are now keyed by taint key:

**Old**:
```hcl
taints = [
  { key = "dedicated", value = "gpu", effect = "NoSchedule" }
]
```

**New**:
```hcl
taints = {
  "dedicated" = { key = "dedicated", value = "gpu", effect = "NoSchedule" }
}
```

## Benefits of Map-Based Schema

### 1. Better User Experience
- More intuitive configuration
- Easier to reference specific items (e.g., `node_groups["primary"]`)
- Natural key-value relationships

### 2. Improved State Management
- Reduced plan noise
- Better diff handling
- More stable resource addresses

### 3. Terraform Best Practices
- Maps are recommended over lists for named collections
- Better for_each compatibility
- Clearer intent in configuration

### 4. Plugin Framework Advantages
- Better validation support
- Type-safe attribute handling
- Improved error messages
- Better nested object support

## Usage Example

```hcl
resource "rafay_eks_cluster_v2" "example" {
  cluster = {
    kind = "Cluster"
    metadata = {
      name    = "eks-demo"
      project = "defaultproject"
      labels = {
        environment = "production"
        team        = "platform"
      }
    }
    spec = {
      type                    = "aws-eks"
      blueprint               = "default"
      cloud_provider          = "aws-creds"
      cross_account_role_arn  = "arn:aws:iam::123456789012:role/CrossAccountRole"
      cni_provider            = "aws-cni"
      
      proxy_config = {
        http_proxy  = "http://proxy.example.com:8080"
        https_proxy = "https://proxy.example.com:8443"
        no_proxy    = "localhost,127.0.0.1,.svc"
      }
      
      system_components_placement = {
        node_selector = {
          "node-type" = "system"
        }
        tolerations = [
          {
            key      = "system"
            operator = "Equal"
            value    = "true"
            effect   = "NoSchedule"
          }
        ]
      }
      
      sharing = {
        enabled = true
        projects = {
          "dev-team" = { name = "dev-team" }
          "ops-team" = { name = "ops-team" }
        }
      }
    }
  }
  
  cluster_config = {
    apiversion = "rafay.io/v1alpha5"
    kind       = "ClusterConfig"
    
    metadata = {
      name    = "eks-demo"
      region  = "us-west-2"
      version = "1.28"
      tags = {
        Environment = "production"
        ManagedBy   = "terraform"
      }
    }
    
    vpc = {
      cidr = "10.0.0.0/16"
      cluster_resources_vpc_config = {
        endpoint_private_access = true
        endpoint_public_access  = true
        public_access_cidrs     = ["0.0.0.0/0"]
      }
      
      subnets = {
        public = {
          "us-west-2a" = {
            id   = "subnet-public-a"
            cidr = "10.0.1.0/24"
            az   = "us-west-2a"
          }
          "us-west-2b" = {
            id   = "subnet-public-b"
            cidr = "10.0.2.0/24"
            az   = "us-west-2b"
          }
        }
        private = {
          "us-west-2a" = {
            id   = "subnet-private-a"
            cidr = "10.0.11.0/24"
            az   = "us-west-2a"
          }
          "us-west-2b" = {
            id   = "subnet-private-b"
            cidr = "10.0.12.0/24"
            az   = "us-west-2b"
          }
        }
      }
      
      nat = {
        gateway = "Single"
      }
    }
    
    node_groups = {
      "primary" = {
        name                = "primary-ng"
        ami                 = "auto"
        instance_type       = "t3.large"
        desired_capacity    = 3
        min_size            = 2
        max_size            = 5
        volume_size         = 80
        volume_type         = "gp3"
        private_networking  = true
        availability_zones  = ["us-west-2a", "us-west-2b"]
        
        labels = {
          "node-group" = "primary"
          "workload"   = "general"
        }
        
        tags = {
          "Name"      = "eks-demo-primary"
          "NodeGroup" = "primary"
        }
        
        taints = {
          "dedicated" = {
            key    = "dedicated"
            value  = "general"
            effect = "PreferNoSchedule"
          }
        }
        
        iam = {
          instance_profile_arn = "arn:aws:iam::123456789012:instance-profile/eks-node"
        }
        
        ssh = {
          public_key_name = "my-keypair"
        }
      }
      
      "gpu" = {
        name                = "gpu-ng"
        ami                 = "auto"
        instance_type       = "g4dn.xlarge"
        desired_capacity    = 1
        min_size            = 0
        max_size            = 3
        volume_size         = 100
        volume_type         = "gp3"
        private_networking  = true
        availability_zones  = ["us-west-2a"]
        
        labels = {
          "node-group"        = "gpu"
          "workload"          = "ml"
          "nvidia.com/gpu"    = "true"
        }
        
        taints = {
          "gpu" = {
            key    = "nvidia.com/gpu"
            value  = "true"
            effect = "NoSchedule"
          }
        }
      }
    }
    
    managed_node_groups = {
      "managed-primary" = {
        name           = "managed-primary"
        ami_type       = "AL2_x86_64"
        instance_types = ["t3.medium", "t3.large"]
        desired_size   = 2
        min_size       = 1
        max_size       = 4
        disk_size      = 50
        
        labels = {
          "managed"  = "true"
          "workload" = "general"
        }
        
        tags = {
          "ManagedBy" = "EKS"
        }
        
        update_config = {
          max_unavailable = 1
        }
      }
    }
  }
  
  timeouts = {
    create = "100m"
    update = "130m"
    delete = "70m"
  }
}
```

## Migration Guide

### Step 1: Update Provider Version
```hcl
terraform {
  required_providers {
    rafay = {
      source  = "registry.terraform.io/rafaysystems/rafay"
      version = ">= 1.2.0"  # Version with Plugin Framework
    }
  }
}
```

### Step 2: Change Resource Type
```hcl
# Old
resource "rafay_eks_cluster" "example" { ... }

# New
resource "rafay_eks_cluster_v2" "example" { ... }
```

### Step 3: Convert Lists to Objects/Maps

**Metadata** (list → object):
```hcl
# Old
metadata {
  name = "my-cluster"
}

# New
metadata = {
  name = "my-cluster"
}
```

**Node Groups** (list → map):
```hcl
# Old
node_groups = [
  { name = "ng-1", ... }
]

# New
node_groups = {
  "ng-1" = { name = "ng-1", ... }
}
```

### Step 4: State Migration
```bash
# Import existing resources into new resource type
terraform import rafay_eks_cluster_v2.example defaultproject/eks-demo

# Or use state manipulation
terraform state mv rafay_eks_cluster.example rafay_eks_cluster_v2.example
```

## Implementation Status

### ✅ Completed
- [x] Schema definition with Plugin Framework
- [x] Map-based node groups
- [x] Map-based subnets
- [x] Map-based taints
- [x] Single nested objects for metadata
- [x] Basic resource structure
- [x] Helper functions framework

### 🚧 In Progress
- [ ] Complete CRUD implementations
- [ ] Full data model conversions
- [ ] Comprehensive validation
- [ ] Error handling
- [ ] State upgrade functions

### 📋 TODO
- [ ] Complete API integration
- [ ] Add comprehensive tests
- [ ] Update documentation
- [ ] Add examples
- [ ] Performance optimization

## Files Structure

```
internal/resource_eks_cluster_v2/
├── README.md                      # This file
├── eks_cluster_v2_resource.go     # Main resource implementation
├── eks_cluster_v2_helpers.go      # Helper functions and converters
└── eks_cluster_v2_validators.go   # Custom validators (future)
```

## Testing

```bash
# Build provider
make build

# Install locally
make install

# Run tests
go test -v ./internal/resource_eks_cluster_v2/...

# Run acceptance tests
TF_ACC=1 go test -v ./internal/resource_eks_cluster_v2/... -run TestAccEKSClusterV2_
```

## Contributing

When contributing to this resource:

1. Follow Plugin Framework best practices
2. Use maps for named collections
3. Use single nested objects instead of lists with MaxItems=1
4. Add comprehensive validation
5. Include unit tests
6. Update documentation
7. Test state upgrades

## Resources

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Schema Design Best Practices](https://developer.hashicorp.com/terraform/plugin/framework/schemas)
- [Maps vs Lists in Terraform](https://developer.hashicorp.com/terraform/language/attr-as-blocks)
- [State Upgrades](https://developer.hashicorp.com/terraform/plugin/framework/migrating/resources/state-upgrade)

