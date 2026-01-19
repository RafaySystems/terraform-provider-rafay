package blueprint

import (
	"context"
	"testing"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	v3 "github.com/RafaySystems/rafay-common/pkg/hub/client/typed/infra/v3"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/encoding/protojson"
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

func blueprintProviderFactoriesWithMock(mockClient *MockBlueprintClient) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"rafay": func() (*schema.Provider, error) {
			provider := &schema.Provider{
				Schema: rafay.Schema(),
				ResourcesMap: map[string]*schema.Resource{
					"rafay_blueprint": rafay.ResourceBluePrint(),
				},
				DataSourcesMap: map[string]*schema.Resource{
					"rafay_blueprint": rafay.DataBluePrint(),
				},
				ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
					return &rafay.ProviderMeta{
						BlueprintClientFactory: func() (v3.BlueprintClient, error) {
							return mockClient, nil
						},
					}, nil
				},
			}
			return provider, nil
		},
	}
}

type blueprintTestConfig struct {
	mockClient        *MockBlueprintClient
	providerFactories map[string]func() (*schema.Provider, error)
}

func newBlueprintTestConfig() blueprintTestConfig {
	mockClient := new(MockBlueprintClient)
	return blueprintTestConfig{
		mockClient:        mockClient,
		providerFactories: blueprintProviderFactoriesWithMock(mockClient),
	}
}

func mustBlueprintFromJSON(t *testing.T, payload string) *infrapb.Blueprint {
	t.Helper()
	bp := &infrapb.Blueprint{}
	if err := protojson.Unmarshal([]byte(payload), bp); err != nil {
		t.Fatalf("failed to unmarshal blueprint JSON: %v", err)
	}
	return bp
}
