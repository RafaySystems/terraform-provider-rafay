# Test Commands Reference

This document provides a comprehensive reference for the streamlined test commands in the Makefile.

## Streamlined Test Commands

### Primary Commands (Daily Use)
#### `make test`
- **Purpose:** Run all tests (unit + integration) - the default command
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay ./tests/...`
- **Usage:** Run all test categories in one command

#### `make test-cover`
- **Purpose:** Run all tests with coverage reporting
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -cover ./rafay ./tests/...`
- **Usage:** Comprehensive testing with coverage analysis

### Unit Tests
#### `make test-unit`
- **Purpose:** Run unit tests for internal functions in rafay/ package
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v ./rafay`
- **Usage:** Test expand/flatten functions and internal logic

### Specialized Commands (When Needed)
#### `make test-integration`
- **Purpose:** Run integration tests only
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=integration ./tests/integration/...`
- **Usage:** Integration testing focus

#### `make test-api`
- **Purpose:** Run tests that require real API credentials
- **Command:** `GOLANG_PROTOBUF_REGISTRATION_CONFLICT=ignore go test -v -tags=integration ./tests/integration/acceptance/`
- **Usage:** Real API credential tests (internal use)

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
# Default: Run all tests (unit + integration)
make test

# Run all tests with coverage
make test-cover

# Quick unit testing during development
make test-unit

# Integration testing only
make test-integration

# Tests requiring real API credentials
make test-api
```

## Build Tags

The organized tests use build tags for categorization:

- `integration` - For integration tests
- `planonly` - For plan-only tests  
- `!planonly` - For non-plan-only tests

This allows fine-grained control over which tests run in different environments.

## Integration with Existing Commands

These streamlined commands complement the existing Makefile targets:
- `make test` - Now the primary command for all tests
- `make test-cover` - Primary command with coverage
- `make testacc` - Acceptance tests (preserved)
- `make test-migrate` - Migration tests (preserved)

The streamlined commands provide a cleaner interface while maintaining compatibility with existing workflows.