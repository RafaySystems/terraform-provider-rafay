# Test Commands Reference

This document provides a comprehensive reference for the new organized test commands added to the Makefile.

## New Organized Test Commands

### Unit Tests
#### `make test-unit`
- **Purpose:** Run unit tests for internal functions in rafay/ package
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay`
- **Usage:** Test expand/flatten functions and internal logic

#### `make test-unit-cover`
- **Purpose:** Run unit tests with coverage reporting
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -cover ./rafay`
- **Usage:** Unit testing with coverage analysis

### Integration Tests
#### `make test-integration`
- **Purpose:** Run all integration tests
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=integration ./tests/integration/...`
- **Usage:** Comprehensive integration testing

#### `make test-plan-only`
- **Purpose:** Run plan-only integration tests
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=planonly ./tests/integration/plan_only/`
- **Usage:** Test configuration validation without resource creation

#### `make test-negative`
- **Purpose:** Run negative integration tests
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags='!planonly' ./tests/integration/negative/`
- **Usage:** Test error handling and edge cases

#### `make test-integration-cover`
- **Purpose:** Run integration tests with coverage
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -cover -tags=integration ./tests/integration/...`
- **Usage:** Integration testing with coverage analysis

### Plugin Framework Tests
#### `make test-framework`
- **Purpose:** Run Plugin Framework tests
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./internal/provider/`
- **Usage:** Test new Plugin Framework implementation

### Comprehensive Tests
#### `make test-all-organized`
- **Purpose:** Run all tests with organized structure
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay ./tests/... ./internal/provider/`
- **Usage:** Run all test categories in one command

#### `make test-all-cover`
- **Purpose:** Run all tests with coverage
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -cover ./rafay ./tests/... ./internal/provider/`
- **Usage:** Comprehensive testing with coverage analysis

## Environment Setup

### For Plan-Only Tests
```bash
export RCTL_API_KEY="dummy"
export RCTL_PROJECT="default"
export RCTL_REST_ENDPOINT="console.example.dev"
```

### For Real Integration Tests
```bash
export RCTL_API_KEY="your-actual-api-key"
export RCTL_PROJECT="your-project"
export RCTL_REST_ENDPOINT="your-endpoint"
```

## Test Organization

The tests are organized according to the new structure:

```
tests/
├── unit/                    # Unit tests for internal functions
│   ├── expand/             # Expand function tests
│   ├── flatten/            # Flatten function tests  
│   └── mocks/              # Mock infrastructure
├── integration/            # Integration tests
│   ├── plan_only/          # Plan validation tests
│   ├── negative/           # Error handling tests
│   └── acceptance/         # Full lifecycle tests
└── framework/              # Plugin Framework tests
    ├── resources/          # Framework resource tests
    └── data_sources/       # Framework data source tests
```

## Usage Examples

```bash
# Quick unit testing during development
make test-unit

# Test configuration validation
make test-plan-only

# Test error handling
make test-negative

# Test new Plugin Framework features
make test-framework

# Comprehensive testing before release
make test-all-cover
```

## Build Tags

The organized tests use build tags for categorization:

- `integration` - For integration tests
- `planonly` - For plan-only tests  
- `!planonly` - For non-plan-only tests

This allows fine-grained control over which tests run in different environments.

## Integration with Existing Commands

These new commands complement the existing Makefile targets:
- `make test` - Original test command (preserved)
- `make testacc` - Acceptance tests (preserved)
- `make test-migrate` - Migration tests (preserved)

The new organized commands provide more granular control over test execution while maintaining compatibility with existing workflows.