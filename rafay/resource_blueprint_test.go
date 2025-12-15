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

func TestResourceBlueprint(t *testing.T) {
	// Setup shared mock
	mockClient := new(MockBlueprintClient)

	// Override the client getter globally for this test suite
	oldGetBlueprintClient := getBlueprintClient
	defer func() { getBlueprintClient = oldGetBlueprintClient }()
	getBlueprintClient = func() (v3.BlueprintClient, error) {
		return mockClient, nil
	}

	tests := []struct {
		name string
		run  func(*testing.T, *MockBlueprintClient)
	}{
		{"Create", testResourceBlueprintCreate},
		{"Read", testResourceBlueprintRead},
		{"Update", testResourceBlueprintUpdate},
		{"Delete", testResourceBlueprintDelete},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, mockClient)
			// Reset expectations after each test to ensure isolation
			mockClient.ExpectedCalls = nil
			mockClient.Calls = nil
		})
	}
}

func testResourceBlueprintCreate(t *testing.T, mockClient *MockBlueprintClient) {
	// Setup expectations
	mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create Test Resource Data
	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint-create",
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
	d.SetId("")

	ctx := context.Background()
	diags := resourceBluePrintCreate(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}

func testResourceBlueprintRead(t *testing.T, mockClient *MockBlueprintClient) {
	expectedBP := &infrapb.Blueprint{
		Metadata: &commonpb.Metadata{
			Name:    "test-blueprint-read",
			Project: "test-project",
		},
		Spec: &infrapb.BlueprintSpec{
			Version: "v1",
		},
	}

	mockClient.On("Get", mock.Anything, mock.MatchedBy(func(opts options.GetOptions) bool {
		return opts.Name == "test-blueprint-read" && opts.Project == "test-project"
	})).Return(expectedBP, nil)

	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint-read",
				"project": "test-project",
			},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId("test-blueprint-read")

	ctx := context.Background()
	diags := resourceBluePrintRead(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}

func testResourceBlueprintUpdate(t *testing.T, mockClient *MockBlueprintClient) {
	mockClient.On("Apply", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint-update",
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
	d.SetId("test-blueprint-update")

	ctx := context.Background()
	diags := resourceBluePrintUpdate(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}

func testResourceBlueprintDelete(t *testing.T, mockClient *MockBlueprintClient) {
	mockClient.On("Delete", mock.Anything, mock.MatchedBy(func(opts options.DeleteOptions) bool {
		return opts.Name == "test-blueprint-delete" && opts.Project == "test-project"
	})).Return(nil)

	resourceSchema := resourceBluePrint().Schema
	resourceDataMap := map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-blueprint-delete",
				"project": "test-project",
			},
		},
	}
	d := schema.TestResourceDataRaw(t, resourceSchema, resourceDataMap)
	d.SetId("test-blueprint-delete")

	ctx := context.Background()
	diags := resourceBluePrintDelete(ctx, d, nil)

	assert.False(t, diags.HasError())
	mockClient.AssertExpectations(t)
}
