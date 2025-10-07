# Documentation Suggestions for the `rafay/` Folder

## Top 5 Priority Documentation

### 1. Deprecation Policy (`docs/deprecation-policy.md`)
**Priority: Critical**

Based on the Rafay provider's adoption of strict semantic versioning and the need for clear upgrade guidance, this should include:

#### **Semantic Versioning Strategy**
- **MAJOR Version Changes (Breaking Changes)**
  - Removing resources (e.g., removing `rafay_legacy_cluster`)
  - Removing or renaming resource arguments (e.g., removing `project_id` attribute from `rafay_cluster`)
  - Changing default behavior that breaks existing configurations
  - Schema changes that require state migration
  - Next version example: `1.1.51` → `2.0.0`

- **MINOR Version Changes (Backward Compatible)**
  - Adding new resources (e.g., adding `rafay_environment_template`)
  - Adding new optional arguments to existing resources
  - Adding new data sources
  - New features that don't break existing configurations
  - Next version example: `1.1.51` → `1.2.0`

- **PATCH Version Changes (Bug Fixes)**
  - Fixing diff suppression bugs
  - Correcting documentation
  - Tightening validation without breaking usage
  - Patching crashes without breaking existing usage
  - Fixing `terraform import` issues
  - Next version example: `1.1.51` → `1.1.52`

#### **Deprecation Process**
Following the [AWS provider deprecation model](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/version-4-upgrade):

- **Deprecation Announcement**
  - Announce deprecations with at least one minor version lead time
  - Provide clear upgrade notes in `UPGRADE.md`
  - Include deprecation warnings in provider logs
  - Document in `CHANGELOG.md` with categorization

- **Resource Deprecation Examples**
  ```hcl
  # Example: Deprecating rafay_cluster in favor of rafay_eks_cluster
  resource "rafay_cluster" "example" {
    # DEPRECATED: Use rafay_eks_cluster instead
    # This resource will be removed in version 2.0.0
    name = "my-cluster"
  }
  
  # New recommended approach
  resource "rafay_eks_cluster" "example" {
    metadata {
      name = "my-cluster"
    }
  }
  ```

- **Argument Deprecation Examples**
  ```hcl
  resource "rafay_aks_cluster" "example" {
    # DEPRECATED: project_id is deprecated, use metadata.project instead
    # project_id will be removed in version 2.0.0
    project_id = "my-project"  # Deprecated
    
    metadata {
      project = "my-project"  # New approach
    }
  }
  ```

#### **Version Support Policy**
- **Support Windows**: Document support for current version (N) and previous version (N-1)
- **End-of-Life Timeline**: Clear communication when versions reach end-of-life
- **Security Updates**: Policy for backporting critical security fixes
- **Migration Assistance**: Provide automated migration tools where possible

#### **Communication Strategy**
- **CHANGELOG.md Integration**
  - Automated changelog generation via GitHub Actions
  - Clear categorization: BREAKING CHANGES, FEATURES, ENHANCEMENTS, BUG FIXES, DEPRECATIONS
  - Version-based organization with release dates
  - Direct links to upgrade guides

- **User Notification Process**
  - Provider log warnings for deprecated features
  - Documentation updates with migration examples
  - GitHub releases with detailed upgrade notes
  - Community forum announcements

#### **State Migration Support**
- **Automatic State Migration**
  - Built-in state migration for schema changes
  - Validation and integrity checks
  - Rollback procedures for failed migrations

- **Manual Migration Procedures**
  - Step-by-step guides for complex migrations
  - `terraform state mv` commands for resource renames
  - Import procedures for restructured resources

#### **Upgrade Documentation Structure**
Following AWS provider patterns:
```
docs/
├── guides/
│   ├── version-2-upgrade.md    # Major version upgrade guide
│   ├── version-1.5-upgrade.md  # Minor version with breaking changes
│   └── migration-examples/     # Specific migration scenarios
├── UPGRADE.md                  # Current upgrade notes
└── CHANGELOG.md               # Automated changelog
```

**Rationale:** Critical for maintaining user trust during resource evolution and ensuring smooth upgrades as the provider matures with 70+ resources.

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
- **Test Organization (Recommended Structure)**
  ```
  tests/
  ├── unit/                    # Unit tests for internal functions
  │   ├── expand/             # Expand function tests (from rafay/)
  │   ├── flatten/            # Flatten function tests (from rafay/)
  │   └── mocks/              # Mock infrastructure (test_mocks.go)
  ├── integration/            # Integration tests
  │   ├── plan_only/          # Plan validation tests
  │   ├── negative/           # Error handling tests
  │   └── acceptance/         # Full lifecycle tests
  └── framework/              # Plugin Framework tests
      ├── resources/          # Framework resource tests
      └── data_sources/       # Framework data source tests
  ```

- **Migration Strategy for Test Consolidation**
  - **Phase 1:** Create new `tests/unit/` structure
  - **Phase 2:** Move unit tests with package adjustments
  - **Phase 3:** Refactor to use public interfaces where possible
  - **Phase 4:** Maintain hybrid approach during SDKv2→Framework migration

- **Package Organization Options**
  - **Option A (Current):** Unit tests in `rafay` package for internal function access
  - **Option B (Consolidated):** All tests in `tests` package with exported functions
  - **Option C (Hybrid):** Structured `tests/` folder with appropriate package imports

- **Build Tag Usage**
  - Unit tests: No build tags (run always)
  - Integration tests: `//go:build integration`
  - Plan-only tests: `//go:build planonly`
  - Negative tests: `//go:build !planonly`

- **Mock Data Management**
  - Centralized mock infrastructure in `tests/unit/mocks/`
  - Reusable test utilities and helpers
  - Complex data structure mocking for EKS/AKS resources
  - Cross-package mock access patterns

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