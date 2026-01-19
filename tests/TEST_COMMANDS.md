# Test Commands Reference

This document provides a comprehensive reference for the test commands available in the Makefile.

## Test Commands Overview

### Primary Commands (Daily Use)

#### `make test` **COMPREHENSIVE**
|- **Purpose:** Run ALL tests (rafay package + framework + integration + negative)
- **What it runs:**
  1. Rafay package tests (expand/flatten functions)
  2. Framework tests with `TF_ACC=1` and `-tags=planonly` (MKS cluster resource test)
  3. Integration plan-only tests with `TF_ACC=1` and `-tags=planonly` (AKS Cluster Spec, AKS Cluster V3, EKS validation)
  4. Negative tests with `TF_ACC=1` (error handling validation)
- **Usage:** Complete test suite in one command - **this is what you want for thorough testing**
- **Note:** Sets all required tags and environment variables automatically

#### `make test-cover` **COMPREHENSIVE + COVERAGE**
- **Purpose:** Run all tests with coverage reporting
- **What it runs:** Same as `make test` but with `-cover` flag for each suite
- **Usage:** Complete test suite with coverage analysis

### Targeted Test Commands

#### `make test-rafay`
- **Purpose:** Run tests in the rafay/ package only
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay`
- **Usage:** Test expand/flatten functions and internal logic in the rafay package
- **Note:** These are unit-style tests that run alongside the resource/data source implementations
- **Tags/Env:** None required

#### `make test-framework`
- **Purpose:** Run framework tests only (MKS cluster resource test)
- **Command:** `TF_ACC=1 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=planonly ./tests/framework/...`
- **Usage:** Test new Plugin Framework implementation (plan-only)
- **Tests:** MKS cluster resource test
- **Tags/Env:** `-tags=planonly` and `TF_ACC=1` (both auto-set)

#### `make test-integration`
- **Purpose:** Run integration plan-only tests
- **Command:** `TF_ACC=1 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=planonly ./tests/integration/plan_only/...`
- **Usage:** Test Terraform plan validation without creating real resources
- **Tests:** AKS Cluster Spec, AKS Cluster V3, and EKS cluster plan validation (3 test files)
- **Tags/Env:** `-tags=planonly` and `TF_ACC=1` (both auto-set)

#### `make test-negative`
- **Purpose:** Run negative validation tests
- **Command:** `TF_ACC=1 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./tests/integration/negative/...`
- **Usage:** Test error handling for invalid configurations
- **Tests:** Validation of null/empty field handling
- **Tags/Env:** `TF_ACC=1` (auto-set)

### Acceptance Testing

#### `make testacc`
- **Purpose:** Run Terraform acceptance tests
- **Command:** `TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m`
- **Usage:** Full resource lifecycle tests with real API (requires credentials)

#### `make test-ci`
- **Purpose:** Run CI tests with acceptance tests enabled
- **Command:** `TF_ACC=1 GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay ./tests/...`
- **Usage:** CI/CD pipeline testing

## Test Organization

The tests are organized in the following structure:

```
tests/
├── framework/              # Plugin Framework tests
│   ├── mks_cluster_resource_test.go
│   └── provider_test.go
├── integration/            # Integration tests
│   ├── plan_only/          # Plan validation tests (3 test files)
│   │   ├── resource_aks_cluster_spec_plan_test.go
│   │   ├── resource_aks_cluster_v3_plan_test.go
│   │   └── resource_eks_cluster_plan_test.go
│   └── negative/           # Error handling tests (5 test files)
└── TEST_COMMANDS.md

rafay/                      # Package with resource implementations
└── test_*.go               # Unit-style tests alongside implementations
```

## Environment Setup

### For Plan-Only Tests (Framework & Integration)
No real credentials needed - tests validate configuration parsing only:
```bash
export RCTL_API_KEY="dummy"
export RCTL_PROJECT="defaultproject"
export RCTL_REST_ENDPOINT="console.example.dev"
```

Or use provider config file:
```bash
# Ensure ~/.rafay/cli/config.json exists with any values
```

### For Real Acceptance Tests
Requires actual API credentials:
```bash
export RCTL_API_KEY="your-actual-api-key"
export RCTL_PROJECT="your-project"
export RCTL_REST_ENDPOINT="your-endpoint"
```

## Usage Examples

```bash
# ⭐ RECOMMENDED: Run all tests comprehensively
# This automatically sets all required tags and TF_ACC
make test

# Run all tests with coverage reporting
make test-cover

# Quick targeted testing during development:
make test-rafay         # Only rafay package (expand/flatten)
make test-framework     # Only MKS cluster tests (auto sets -tags=planonly)
make test-integration   # Only AKS/EKS plan tests (auto sets -tags=planonly)
make test-negative      # Only error handling (auto sets TF_ACC=1)

# Full acceptance tests with real API (requires credentials)
make testacc
```

## Build Tags

The tests use build tags for categorization:

- `planonly` - For plan-only validation tests (framework and integration/plan_only)
- `!planonly` - For negative validation tests (integration/negative)

## Quick Reference

| Command | What It Tests | Build Tags / Env | Requires Real API |
|---------|--------------|------------------|-------------------|
| `make test` | **Everything (comprehensive)** | **Auto-sets all** | **No** |
| `make test-cover` | Everything + coverage | Auto-sets all | No |
| `make test-rafay` | rafay/ package tests | None | No |
| `make test-framework` | MKS cluster resource test | planonly (auto) | No |
| `make test-integration` | Plan validation | planonly (auto) | No |
| `make test-negative` | Error handling | TF_ACC=1 (auto) | No |
| `make testacc` | Full lifecycle | TF_ACC=1 | Yes |
| `make test-ci` | CI pipeline | TF_ACC=1 | Yes |

**Note:** All specific test commands (`test-framework`, `test-integration`, `test-negative`) automatically set their required build tags and environment variables.

## Development Workflow

**During Development (Quick Feedback):**
```bash
# Test only what you're working on:
make test-rafay         # If working on expand/flatten functions
make test-framework     # If working on MKS cluster implementation
make test-integration   # If working on AKS/EKS resources
make test-negative      # If working on validation logic
```

**Before Committing (Comprehensive Check):**
```bash
# ⭐ Run the comprehensive suite - catches everything!
make test

# This automatically runs:
# [1/4] rafay package tests
# [2/4] framework tests (with -tags=planonly)
# [3/4] integration plan-only tests (with -tags=planonly)
# [4/4] negative tests (with TF_ACC=1)
```

**With Coverage Analysis:**
```bash
# Same comprehensive suite with coverage
make test-cover
```

**CI/CD Pipeline:**
```bash
# Runs all tests including acceptance tests with real API
make test-ci
```

## Key Features

1. **`make test` is comprehensive** - runs all test suites in one command
2. **Build tags are automatically set** - no need to remember `-tags=planonly`
3. **TF_ACC is automatically set** - negative tests execute without manual configuration
4. **Clear progress indicators** - see which suite is running (1/4, 2/4, etc.)
5. **Individual commands available** - use them for fast iteration during development
