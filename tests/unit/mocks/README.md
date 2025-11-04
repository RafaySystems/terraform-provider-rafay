# Unit Test Mocks

This directory contains mock infrastructure for unit testing.

## Files

- `test_mocks.go` - Comprehensive mock data structures and utilities

## Usage

The mock infrastructure provides:

- `MockTestData` - Central structure containing EKS and AKS mock data
- `MockEKSData` - Mock data for EKS cluster testing
- `MockAKSData` - Mock data for AKS cluster testing
- `GetCtyValue()` - Utility for converting test data to cty.Value
- `MockResourceData()` - Creates mock ResourceData for schema testing
- `MockDiagnostics()` - Creates mock diagnostics
- `MockContext()` - Creates mock context

## Import Usage

```go
import "github.com/RafaySystems/terraform-provider-rafay/tests/unit/mocks"

// Use mock data
mockData := mocks.NewMockTestData()
eksData := mockData.EKS
```

## Cross-Package Access

Since these mocks are in a separate package, they reference rafay package types with the `rafay.` prefix.
