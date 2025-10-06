# Documentation Suggestions for the `rafay/` Folder

## Top 5 Priority Documentation

### 1. Deprecation Policy (`docs/deprecation-policy.md`)
**Priority: Critical**

Based on the ongoing migration from SDKv2 to Plugin Framework, this should include:

#### **Deprecation Strategy**
- **Criteria for Deprecation**
  - Technical debt assessment for legacy SDKv2 implementations
  - Performance and maintainability considerations
  - Plugin Framework feature parity requirements
  - User impact analysis for breaking changes

- **Clear Examples of SDKv2 vs Plugin Framework**
  - **SDKv2 Implementation Examples (`rafay/` folder):**
    - `rafay/provider.go` - SDKv2 provider using `schema.Provider` and `helper/schema`
    - `rafay/resource_group.go` - Simple resource with manual schema definition
    - `rafay/resource_driver.go` - Complex resource with expand/flatten patterns
    - `rafay/resource_eks_cluster.go` - Large complex resource (195KB, 7172 lines)
    - SDKv2 patterns: `schema.Resource`, `ResourceData`, `diag.Diagnostics`

  - **Plugin Framework Implementation Examples (`internal/` folder):**
    - `internal/provider/provider.go` - Framework provider using `provider.Provider` interface
    - `internal/resource_mks_cluster/mks_cluster_resource_gen.go` - Generated code from JSON schema
    - `internal/resource_mks_cluster/mks_cluster_resource_ext.go` - Custom conversion methods
    - `internal/provider/mks_cluster_resource_test.go` - Framework testing patterns
    - Framework patterns: `types.String`, `basetypes.BoolValue`, `schema.Schema`

- **Key Differences Illustrated:**
  ```go
  // SDKv2 Pattern (rafay/resource_group.go)
  func resourceGroup() *schema.Resource {
      return &schema.Resource{
          CreateContext: resourceGroupCreate,
          Schema: map[string]*schema.Schema{
              "name": {
                  Type:     schema.TypeString,
                  Required: true,
              },
          },
      }
  }

  // Plugin Framework Pattern (internal/provider/provider.go)
  type RafayFwProviderModel struct {
      ProviderConfigFile types.String `tfsdk:"provider_config_file"`
  }
  
  func (p *RafayFwProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
      resp.Schema = schema.Schema{
          Attributes: map[string]schema.Attribute{
              "provider_config_file": schema.StringAttribute{
                  Optional: true,
              },
          },
      }
  }
  ```

- **Deprecation Timeline**
  - **Phase 1 (Current):** Maintain SDKv2 stability while developing Plugin Framework
  - **Phase 2 (6-12 months):** Gradual deprecation warnings and migration paths
  - **Phase 3 (12+ months):** Full deprecation and removal of legacy code

#### **Communication Strategy**
- **User Notification Process**
  - Deprecation warnings in provider logs
  - Documentation updates with migration guides
  - Community communication through forums and GitHub
  - Version-specific deprecation notices

- **Migration Support**
  - Detailed migration guides for each deprecated resource
  - Compatibility matrices between old and new implementations
  - Support channels for migration assistance
  - Timeline extensions for complex migration scenarios

#### **Backward Compatibility**
- **State File Compatibility**
  - Automatic state migration utilities
  - Manual migration procedures for edge cases
  - Rollback procedures if migration fails
  - State validation and integrity checks

- **Configuration Compatibility**
  - Syntax preservation during migration
  - Deprecated argument handling
  - Default value migration strategies
  - Breaking change documentation

**Rationale:** Critical for managing the complex migration from SDKv2 to Plugin Framework while maintaining user trust and system stability.

### 2. Testing Guide (`docs/testing-guide.md`)
**Priority: High**

Based on analysis of the comprehensive testing infrastructure in the codebase, this should include:

#### **Unit Testing Framework**
- **Mock Infrastructure Usage**
  - How to leverage `test_mocks.go` (889 lines) for comprehensive testing
  - `MockTestData` structure with EKS and AKS mock data
  - `MockResourceData` creation for schema testing
  - `GetCtyValue` utility for complex type testing

- **Expand Function Testing**
  - Pattern analysis from `resource_eks_cluster_expand_test.go` (771 lines)
  - Table-driven test structure for comprehensive coverage
  - Testing complex nested structures (cluster metadata, VPC configs, IAM settings)
  - Benchmark testing patterns for performance validation
  - Example test cases:
    ```go
    // From actual test file
    func TestExpandEKSCluster(t *testing.T) {
        tests := []struct {
            name     string
            input    []interface{}
            expected *EKSCluster
        }{
            // Multiple test scenarios including edge cases
        }
    }
    ```

- **Flatten Function Testing**
  - Pattern analysis from `resource_eks_cluster_flatten_test.go` (834 lines)
  - State reconstruction testing from API responses
  - Null value handling and edge case validation
  - Complex nested data structure flattening

#### **Integration Testing Framework**
- **Plan-Only Testing**
  - Pattern analysis from `tests/resource_eks_cluster_plan_test.go` (438 lines)
  - Using external providers for registry-based testing
  - Environment setup with dummy credentials for plan validation
  - Configuration validation without actual resource creation
  - Example scenarios:
    - Basic cluster configurations
    - Node group defaults validation
    - VPC and networking configurations
    - IAM and OIDC service account testing
    - Access configuration and policies
    - Addon configurations

- **Negative Testing**
  - Pattern analysis from `tests/resource_eks_cluster_empty_null_negative_test.go` (447 lines)
  - Empty vs null value handling
  - Required field validation
  - Error message pattern matching
  - Build tag usage (`//go:build !planonly`) for test categorization

#### **Testing Best Practices**
- **Test Organization**
  - Separation of unit tests (in `rafay/`) vs integration tests (in `tests/`)
  - Build tag usage for different test categories
  - External provider configuration for registry testing

- **Mock Data Management**
  - Centralized mock data in `test_mocks.go`
  - Reusable test utilities and helpers
  - Complex data structure mocking for EKS/AKS resources

- **Test Coverage Strategies**
  - Expand/flatten function pairing validation
  - Schema validation testing
  - Error handling and edge case coverage
  - Performance benchmark integration

#### **Test Execution Patterns**
- **Local Development Testing**
  - Unit test execution with mock data
  - Plan-only integration testing
  - Test environment setup and teardown

- **CI/CD Integration**
  - Build tag separation for different test environments
  - External provider testing with registry
  - Negative test validation in pipelines

**Rationale:** Essential for maintaining code quality and reliability during the migration period and beyond.

### 3. Examples Directory (`examples/`)
**Priority: Critical**

Following the [AWS provider examples structure](https://github.com/hashicorp/terraform-provider-aws/tree/main/examples), this should include:

#### **Basic Examples**
- **Simple Resource Configurations**
  - Basic EKS cluster setup with minimal configuration
  - AKS cluster with default settings
  - Simple project and namespace creation
  - Basic user and group management

- **Common Use Cases**
  - Multi-region cluster deployment
  - Development vs production environment patterns
  - Basic CI/CD pipeline configurations
  - Simple workload deployment examples

#### **Advanced Examples**
- **Complex Cluster Configurations**
  - EKS cluster with custom VPC, IAM roles, and node groups
  - AKS cluster with managed identity and advanced networking
  - GKE cluster with custom configurations
  - Multi-cluster management scenarios

- **Enterprise Patterns**
  - RBAC and security policy configurations
  - Cost management and chargeback setups
  - Fleet management with multiple clusters
  - GitOps workflow implementations

#### **Integration Examples**
- **Multi-Resource Scenarios**
  - Complete application deployment workflows
  - Infrastructure and application lifecycle management
  - Disaster recovery configurations
  - Monitoring and alerting setups

- **Migration Examples**
  - Legacy to modern resource migration patterns
  - State migration examples
  - Configuration upgrade patterns
  - Compatibility examples during deprecation

#### **Example Structure**
Following AWS provider patterns:
```
examples/
├── basic/
│   ├── eks-cluster/
│   ├── aks-cluster/
│   └── project-setup/
├── advanced/
│   ├── multi-cluster/
│   ├── enterprise-rbac/
│   └── gitops-workflow/
├── integrations/
│   ├── complete-application/
│   └── disaster-recovery/
└── migration/
    ├── sdkv2-to-plugin-framework/
    └── state-migration/
```

**Rationale:** Critical for user adoption and understanding, especially during the migration period where clear examples help users navigate changes.

### 4. Automated Changelog (`CHANGELOG.md` + GitHub Actions)
**Priority: High**

Implementation of an AI-powered changelog generation system:

#### **GitHub Actions Workflow**
- **Trigger Configuration**
  - Activate on merge to main branch
  - Process pull request metadata and commit history
  - Generate changelog entries automatically

- **AI Integration**
  - Parse commit messages using conventional commit patterns
  - Analyze pull request descriptions and comments
  - Categorize changes (features, bug fixes, breaking changes, deprecations)
  - Generate human-readable changelog entries

#### **Changelog Structure**
- **Version-based Organization**
  - Semantic versioning (major.minor.patch)
  - Release date and version tags
  - Clear categorization of changes

- **Change Categories**
  - **BREAKING CHANGES:** SDKv2 to Plugin Framework migrations
  - **FEATURES:** New resources and functionality
  - **ENHANCEMENTS:** Improvements to existing resources
  - **BUG FIXES:** Issue resolutions and patches
  - **DEPRECATIONS:** Legacy feature deprecations
  - **DOCUMENTATION:** Documentation updates and improvements

#### **Implementation Details**
- **GitHub Action Workflow**
  ```yaml
  name: Generate Changelog
  on:
    push:
      branches: [main]
  jobs:
    changelog:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v3
        - name: Generate Changelog
          uses: ai-changelog-generator@v1
          with:
            token: ${{ secrets.GITHUB_TOKEN }}
            output: CHANGELOG.md
  ```

- **AI Processing Logic**
  - Natural language processing of commit messages
  - Merge request description analysis
  - Automatic categorization and formatting
  - Duplicate detection and consolidation

**Rationale:** Essential for tracking the complex migration process and keeping users informed of changes, especially during the SDKv2 to Plugin Framework transition.

### 5. Resource Implementation Guide (`docs/resource-implementation-guide.md`)
**Priority: Medium**

Selected as the fifth priority from the lower priority items, this should cover:

#### **SDKv2 Implementation Patterns**
- **Standard Resource Lifecycle**
  - The Create/Read/Update/Delete patterns used across 70+ resources
  - Context handling and timeout configurations
  - Error handling and diagnostics patterns
  - Import functionality implementation

- **Schema Definition Patterns**
  - How complex nested schemas are structured (EKS/AKS cluster configurations)
  - Optional vs required field patterns
  - List and map handling in Terraform schemas
  - Schema versioning strategies

#### **Expand/Flatten Function Patterns**
- **Data Transformation**
  - Understanding the extensive utility functions in `utils.go` (2474 lines)
  - Data transformation between Terraform state and Rafay API
  - Handling complex nested data structures
  - JSON parsing and validation patterns

- **Testing Integration**
  - How to use the mock infrastructure effectively
  - Unit testing strategies for expand/flatten functions
  - Integration testing approaches
  - Test data organization and reuse

#### **Migration Considerations**
- **Plugin Framework Preparation**
  - Identifying resources ready for migration
  - Compatibility requirements during transition
  - Testing strategies for parallel implementations
  - Deprecation planning for legacy resources

**Rationale:** Important for maintaining consistency during the migration period and helping developers understand the established patterns in the legacy codebase.

## Lower Priority Documentation

The following documentation pieces are valuable but lower priority:

- **Contributing Guide (`CONTRIBUTING.md`)** - Development environment setup and contribution guidelines
- **Architecture Overview (`docs/architecture.md`)** - High-level system architecture and component relationships
- **Troubleshooting Guide (`docs/troubleshooting.md`)** - Common issues and debugging approaches
- **Performance Optimization Guide (`docs/performance.md`)** - Best practices for large-scale deployments
- **Security Best Practices (`docs/security.md`)** - Authentication, authorization, and secure configuration patterns

## Implementation Approach

Following the reordered priorities:

1. **Start with Deprecation Policy** to establish clear migration expectations
2. **Implement Testing Guide** to ensure code quality during transition
3. **Create Examples Directory** to help users navigate changes
4. **Set up Automated Changelog** to track migration progress
5. **Develop Resource Implementation Guide** for consistency maintenance

This prioritization focuses on managing the SDKv2 to Plugin Framework migration while maintaining user experience and code quality throughout the transition period.