# Integration Tests - Negative Testing

This directory contains negative integration tests that validate error handling and edge cases.

## Purpose

Negative tests verify:
- Empty vs null value handling
- Required field validation
- Error message patterns
- Invalid configuration rejection
- Boundary condition handling

## Test Files

- `resource_eks_cluster_empty_null_negative_test.go` - EKS cluster error handling validation

## Build Tags

These tests use the `//go:build !planonly` build tag to exclude them from plan-only test runs.

## Running

```bash
# Run negative tests
go test -tags=!planonly ./tests/integration/negative/...
```

## Test Patterns

Tests validate specific error conditions:
- Missing required arguments
- Invalid field values
- Configuration conflicts
- Provider authentication failures
