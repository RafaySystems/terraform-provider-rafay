# Documentation Suggestions for the `rafay/` Folder

Based on analysis of the [Terraform AWS Provider documentation structure](https://github.com/hashicorp/terraform-provider-aws/tree/main/docs) and examination of the `rafay/` folder codebase, here are the top 3 documentation pieces that would provide the most value.

## Top 3 Priority Documentation

### 1. Contributing Guide (`CONTRIBUTING.md`)
**Priority: Critical**

Following the AWS provider's pattern, this should include:

- **Development Environment Setup**
  - How to set up local development with `rctl` and `rafay-common` dependencies
  - Required Go version and development tools
  - Configuration of provider authentication and testing environments

- **Resource Development Guidelines**
  - Patterns for creating new resources using Terraform SDKv2 patterns
  - Standard resource lifecycle implementation (Create/Read/Update/Delete)
  - Schema definition best practices for complex nested configurations

- **Testing Standards**
  - How to use the comprehensive `test_mocks.go` infrastructure (889 lines)
  - Writing effective expand/flatten function tests
  - Integration testing with Rafay's control plane
  - Mock data patterns for EKS/AKS cluster testing

- **Code Organization**
  - Explaining the relationship between `rafay/` (legacy SDKv2) and `internal/` (new Plugin Framework)
  - File naming conventions and structure
  - When to modify existing vs create new resources

- **Pull Request Process**
  - Code review requirements and standards
  - Testing requirements before submission
  - Documentation update requirements

- **Migration Guidelines**
  - Understanding the ongoing migration from SDKv2 to Plugin Framework
  - How to work with legacy code during transition
  - Deprecation and compatibility considerations

**Rationale:** Critical given the complexity of the codebase (195KB `resource_eks_cluster.go` file alone) and the ongoing migration effort from SDKv2 to Plugin Framework.

### 2. Resource Implementation Guide (`docs/resource-guide.md`)
**Priority: High**

Based on the AWS provider's resource documentation patterns, this should cover:

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

- **Expand/Flatten Function Patterns**
  - Understanding the extensive utility functions in `utils.go` (2474 lines)
  - Data transformation between Terraform state and Rafay API
  - Handling complex nested data structures
  - JSON parsing and validation patterns

- **API Integration Standards**
  - Consistent patterns for Rafay control plane integration
  - Authentication and authorization handling
  - Error handling from API calls
  - Retry and timeout strategies

- **Testing Patterns**
  - How to use the mock infrastructure effectively
  - Unit testing strategies for expand/flatten functions
  - Integration testing approaches
  - Test data organization and reuse

**Rationale:** Essential for maintaining consistency across the massive codebase and helping developers understand established patterns.

### 3. Architecture Overview (`docs/architecture.md`)
**Priority: High**

Similar to how the AWS provider documents its internal structure, this should explain:

- **Package Structure**
  - Organization of resources, data sources, and utilities
  - File naming conventions and categorization
  - Relationship between different resource types

- **API Integration Architecture**
  - How the provider integrates with Rafay's control plane via `rctl` and `rafay-common`
  - Protocol buffer integration patterns
  - Client initialization and management

- **Authentication Flow**
  - Provider configuration mechanisms
  - Config file vs environment variable authentication
  - TLS configuration and security considerations

- **Resource Categories**
  - **Cluster Resources:** EKS (`resource_eks_cluster.go`), AKS (`resource_aks_cluster.go`, `resource_aks_cluster_v3.go`), GKE (`resource_gke_cluster.go`)
  - **Workload Resources:** Workloads, templates, blueprints, namespaces, projects
  - **Infrastructure Resources:** Agents, repositories, credentials, network policies
  - **Management Resources:** Users, groups, roles, policies

- **Legacy vs Modern Code**
  - Clear explanation of what's in `rafay/` (SDKv2) vs `internal/` (Plugin Framework)
  - Migration strategy and timeline
  - Compatibility considerations during transition

- **Dependencies and External Integrations**
  - Understanding the complex dependency tree with Rafay's internal packages
  - External tool integrations (`exec.go` for command execution)
  - Version compatibility requirements

- **Data Flow**
  - How Terraform configuration flows through expand functions to API calls
  - How API responses flow through flatten functions back to Terraform state
  - State management and drift detection

**Rationale:** Provides essential context for anyone working with the codebase, especially given the size (100+ files) and complexity of the implementation.

## Why These Three?

Looking at the [AWS provider's documentation approach](https://github.com/hashicorp/terraform-provider-aws), these three documents would:

1. **Enable Contribution** - The contributing guide removes barriers for new developers joining the project
2. **Ensure Consistency** - The resource guide maintains code quality across the large and complex codebase
3. **Provide Context** - The architecture overview helps developers understand the big picture and make informed decisions

These align with HashiCorp's documentation standards while addressing the specific challenges of the `rafay/` folder:
- **Size:** 100+ files with some extremely large implementations
- **Complexity:** Massive resource implementations with intricate nested configurations
- **Legacy Nature:** Built on Terraform SDKv2 during active migration to Plugin Framework
- **Domain Complexity:** Deep integration with Rafay's Kubernetes management platform

## Additional High-Priority Documentation

### 4. Testing Guide (`docs/testing-guide.md`)
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

### 5. Migration Roadmap (`docs/migration-roadmap.md`)
**Priority: High**

Based on the existing migration patterns and Plugin Framework transition, this should include:

#### **Current State Analysis**
- **Legacy SDKv2 Implementation (`rafay/` folder)**
  - 100+ resource and data source files
  - Complex expand/flatten function patterns
  - Extensive mock testing infrastructure
  - Integration with `rctl` and `rafay-common` packages

- **Modern Plugin Framework Implementation (`internal/` folder)**
  - New framework adoption in progress
  - Structured approach with generated code
  - Enhanced type safety and validation

#### **Migration Strategy**
- **Phase 1: Foundation (Current)**
  - Maintain SDKv2 implementation for stability
  - Develop Plugin Framework patterns in `internal/`
  - Establish migration guidelines and patterns
  - Create compatibility testing framework

- **Phase 2: Gradual Migration**
  - **Resource Prioritization**
    - Start with simpler resources (users, groups, basic configurations)
    - Progress to complex resources (EKS/AKS clusters)
    - Maintain backward compatibility during transition
  
  - **Migration Pattern per Resource**
    - Create Plugin Framework equivalent in `internal/`
    - Implement comprehensive test coverage
    - Validate feature parity with legacy implementation
    - Switch provider registration to new implementation
    - Deprecate legacy implementation

- **Phase 3: Completion and Cleanup**
  - Remove deprecated SDKv2 implementations
  - Clean up legacy test infrastructure
  - Update documentation and examples
  - Final compatibility validation

#### **Technical Migration Patterns**
- **From SDKv2 to Plugin Framework**
  - Schema definition transformation
  - Expand/flatten function migration to Value types
  - Context handling updates
  - Error handling pattern changes
  - Testing framework adaptation

- **Compatibility Considerations**
  - State file compatibility during migration
  - Configuration syntax preservation
  - Provider behavior consistency
  - Deprecation warnings and communication

#### **Risk Mitigation**
- **Testing Strategy**
  - Parallel testing of old and new implementations
  - State migration validation
  - Regression testing for existing users
  - Performance comparison testing

- **Rollback Planning**
  - Ability to revert to SDKv2 implementation
  - State rollback procedures
  - User communication for issues
  - Hotfix deployment strategies

#### **Timeline and Milestones**
- **Short-term (3-6 months)**
  - Complete migration guidelines documentation
  - Migrate 5-10 simple resources
  - Establish testing patterns for Plugin Framework
  - User communication about migration plans

- **Medium-term (6-12 months)**
  - Migrate complex resources (EKS, AKS clusters)
  - Comprehensive compatibility testing
  - Beta release with Plugin Framework resources
  - Gather user feedback and iterate

- **Long-term (12+ months)**
  - Complete migration of all resources
  - Remove legacy SDKv2 code
  - Final documentation updates
  - Provider v2.0 release

#### **Success Metrics**
- **Technical Metrics**
  - 100% feature parity with legacy implementation
  - No regression in functionality or performance
  - Comprehensive test coverage for new implementation
  - Clean separation of legacy vs modern code

- **User Experience Metrics**
  - Seamless migration for existing users
  - No breaking changes in configuration syntax
  - Improved error messages and validation
  - Enhanced provider performance and reliability

## Additional Considerations

While the top 5 priorities above address the most critical needs, other valuable documentation pieces could include:

- **Troubleshooting Guide** - Common issues and debugging approaches
- **Resource-Specific Guides** - Deep dives into complex resources like EKS clusters
- **Performance Optimization Guide** - Best practices for large-scale deployments
- **Security Best Practices** - Authentication, authorization, and secure configuration patterns

## Implementation Approach

Following the AWS provider model:
1. Start with the Contributing Guide to establish development standards
2. Create the Architecture Overview to provide necessary context
3. Develop the Resource Implementation Guide to ensure consistency
4. Iterate based on developer feedback and common questions

This documentation foundation would significantly improve developer experience and code maintainability for the `rafay/` folder.
