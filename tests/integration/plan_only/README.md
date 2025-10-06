# Integration Tests - Plan Only

This directory contains plan-only integration tests that validate Terraform configurations without creating actual resources.

## Purpose

Plan-only tests verify:
- Configuration syntax and validation
- Schema compliance
- Default value assignment
- Provider configuration
- Resource attribute validation

## Test Files

- `resource_eks_cluster_plan_test.go` - EKS cluster configuration validation

## Build Tags

These tests use the `//go:build planonly` build tag.

## Running

```bash
# Run plan-only tests
go test -tags=planonly ./tests/integration/plan_only/...
```

## Environment Setup

Tests use dummy environment variables for plan validation:
- `RCTL_API_KEY=dummy`
- `RCTL_PROJECT=default`  
- `RCTL_REST_ENDPOINT=console.example.dev`
