package rafay

import (
	"context"
	"testing"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	v3 "github.com/RafaySystems/rafay-common/pkg/hub/client/typed/infra/v3"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBlueprintClient is a mock of BlueprintClient interface
type MockBlueprintClient struct {
	mock.Mock
}

func (m *MockBlueprintClient) Apply(ctx context.Context, blueprint *infrapb.Blueprint, opts options.ApplyOptions) error {
	args := m.Called(ctx, blueprint, opts)
	return args.Error(0)
}

func (m *MockBlueprintClient) Get(ctx context.Context, opts options.GetOptions) (*infrapb.Blueprint, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*infrapb.Blueprint), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockBlueprintClient) Delete(ctx context.Context, opts options.DeleteOptions) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

func (m *MockBlueprintClient) List(ctx context.Context, opts options.ListOptions) (*infrapb.BlueprintList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) != nil {
		return args.Get(0).(*infrapb.BlueprintList), args.Error(1)
	}
	return nil, args.Error(1)
}

func TestResourceBlueprintCreate(t *testing.T) {
	// Setup mock
	mockClient := new(MockBlueprintClient)

	// Override the client getter
	oldGetBlueprintClient := getBlueprintClient
	defer func() { getBlueprintClient = oldGetBlueprintClient }()
	getBlueprintClient = func() (v3.BlueprintClient, error) {
		return mockClient, nil
	}

	// Setup expectations
	mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Check if already exists check - read call?
	// In resourceBluePrintCreate:
	// create := isBlueprintAlreadyExists(ctx, d)
	//    -> resourceBluePrintRead
	//       -> client.Get
	// So we expect a Get call first which might return error/empty for "exists check" or we can mock it to return error 404

	// Wait, isBlueprintAlreadyExists creates a NEW resourceData, so it might check if id is set?
	// Let's check isBlueprintAlreadyExists impl if available or infer from usage.
	// Usually it checks if ID is set. If we create a fresh ResourceData, ID is empty.

	// Create Test Resource Data
	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint",
				"project": "test-project",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"version": "v1",
				"default_addons": []interface{}{
					map[string]interface{}{
						"enable_ingress": true,
					},
				},
			},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId("") // New resource

	// The code:
	// create := isBlueprintAlreadyExists(ctx, d)
	// diags := resourceBluePrintUpsert(ctx, d, m)

	// We expect Apply to be called.

	ctx := context.Background()
	diags := resourceBluePrintCreate(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}

func TestResourceBlueprintRead(t *testing.T) {
	mockClient := new(MockBlueprintClient)

	oldGetBlueprintClient := getBlueprintClient
	defer func() { getBlueprintClient = oldGetBlueprintClient }()
	getBlueprintClient = func() (v3.BlueprintClient, error) {
		return mockClient, nil
	}

	expectedBP := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "test-blueprint",
			Project: "test-project",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v1",
		},
	}

	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint" && opts.Project == "test-project"
	})).Return(expectedBP, nil)

	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint",
				"project": "test-project",
			},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId("test-blueprint")
	// Wait, getMetaName uses ID if set? Or metadata name?
	// In Read: Use GetMetaData(d) -> reads from config/schema

	ctx := context.Background()
	diags := resourceBluePrintRead(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}

func TestResourceBlueprintUpdate(t *testing.T) {
	mockClient := new(MockBlueprintClient)

	oldGetBlueprintClient := getBlueprintClient
	defer func() { getBlueprintClient = oldGetBlueprintClient }()
	getBlueprintClient = func() (v3.BlueprintClient, error) {
		return mockClient, nil
	}

	mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint",
				"project": "test-project",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"version": "v2",
			},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId("test-blueprint")

	ctx := context.Background()
	diags := resourceBluePrintUpdate(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}

func TestResourceBlueprintDelete(t *testing.T) {
	mockClient := new(MockBlueprintClient)

	oldGetBlueprintClient := getBlueprintClient
	defer func() { getBlueprintClient = oldGetBlueprintClient }()
	getBlueprintClient = func() (v3.BlueprintClient, error) {
		return mockClient, nil
	}

	mockClient.On("Delete", mock.Anything, mock.MatchedBy(func(opts options.DeleteOptions) bool {
		return opts.Name == "test-blueprint" && opts.Project == "test-project"
	})).Return(nil)

	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint",
				"project": "test-project",
			},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId("test-blueprint")

	ctx := context.Background()
	diags := resourceBluePrintDelete(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}
