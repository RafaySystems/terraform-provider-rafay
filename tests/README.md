# Terraform Provider Rafay - Tests

This directory contains all test files for the Terraform Provider Rafay.

## Directory Structure

```
tests/
├── framework/              # Plugin Framework tests
├── integration/            # Integration tests
│   ├── plan_only/          # Plan validation tests
│   └── negative/           # Error handling tests
├── README.md
└── TEST_COMMANDS.md
```

## Test Categories

### Framework Tests (`framework/`)
- Tests for new Terraform Plugin Framework implementation
- Currently includes MKS cluster resource test
- Separate from legacy SDKv2 tests in `rafay/` package

**Files:**
- `mks_cluster_resource_test.go` - MKS cluster resource test
- `provider_test.go` - Provider configuration setup

### Integration Tests (`integration/`)

#### Plan Only Tests (`plan_only/`)
- Test configuration validation without creating real resources
- Validate Terraform configurations can be parsed and planned successfully
- Use build tag: `//go:build planonly`

**Files:**
- `resource_aks_cluster_spec_plan_test.go` - AKS cluster spec plan validation (3 tests)
- `resource_aks_cluster_v3_plan_test.go` - AKS cluster v3 plan validation (6 tests)
- `resource_eks_cluster_plan_test.go` - EKS cluster plan validation (7 tests)

#### Negative Tests (`negative/`)
- Test error handling and validation for invalid configurations
- Ensure proper error messages for null/empty required fields
- Use build tag: `//go:build !planonly`

**Files:**
- `resource_aks_cluster_empty_null_negative_test.go` - AKS cluster error handling
- `resource_aks_cluster_spec_empty_null_negative_test.go` - AKS cluster spec error handling
- `resource_aks_cluster_v3_empty_null_negative_test.go` - AKS cluster v3 error handling
- `resource_aks_workload_identity_empty_null_negative_test.go` - AKS workload identity error handling
- `resource_eks_cluster_empty_null_negative_test.go` - EKS cluster error handling

## Build Tags

- Framework tests: `//go:build planonly` 
- Plan-only tests: `//go:build planonly`
- Negative tests: `//go:build !planonly`

## Running Tests

```bash
# Run all framework tests (plan-only)
go test -tags=planonly ./tests/framework/...

# Run all integration plan-only tests
go test -tags=planonly ./tests/integration/plan_only/...

# Run negative tests
go test ./tests/integration/negative/...

# Run all tests in the tests directory
go test ./tests/...
```

See `TEST_COMMANDS.md` for more detailed testing commands and examples.
