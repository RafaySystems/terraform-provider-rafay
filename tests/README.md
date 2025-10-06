# Test Organization

This directory contains all test files organized by type and purpose according to the Testing Guide.

## Directory Structure

```
tests/
├── unit/                    # Unit tests for internal functions
│   ├── expand/             # Expand function tests (moved from rafay/)
│   ├── flatten/            # Flatten function tests (moved from rafay/)
│   └── mocks/              # Mock infrastructure (test_mocks.go)
├── integration/            # Integration tests
│   ├── plan_only/          # Plan validation tests
│   ├── negative/           # Error handling tests
│   └── acceptance/         # Full lifecycle tests
└── framework/              # Plugin Framework tests
    ├── resources/          # Framework resource tests
    └── data_sources/       # Framework data source tests
```

## Test Categories

### Unit Tests (`unit/`)
- Test internal expand/flatten functions
- Use mock data from `unit/mocks/`
- Run in same package as implementation for internal function access

### Integration Tests (`integration/`)
- **Plan Only:** Test configuration validation without resource creation
- **Negative:** Test error handling and validation
- **Acceptance:** Full resource lifecycle tests (create/read/update/delete)

### Framework Tests (`framework/`)
- Tests for new Plugin Framework implementation
- Separate from legacy SDKv2 tests

## Build Tags

- Unit tests: No build tags (run always)
- Integration tests: `//go:build integration`
- Plan-only tests: `//go:build planonly`
- Negative tests: `//go:build !planonly`

## Running Tests

```bash
# Run all unit tests
go test ./tests/unit/...

# Run integration tests
go test -tags=integration ./tests/integration/...

# Run plan-only tests
go test -tags=planonly ./tests/integration/plan_only/...

# Run negative tests
go test -tags=!planonly ./tests/integration/negative/...
```
