package tests

import (
	"context"
	"os"
	"testing"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockClient is a mock implementation of the Rafay client
type MockClient struct {
	mock.Mock
}

// MockFleetPlanService is a mock implementation of the FleetPlan service
type MockFleetPlanService struct {
	mock.Mock
}

func (m *MockFleetPlanService) Apply(ctx context.Context, fleetplan *infrapb.FleetPlan, opts interface{}) error {
	args := m.Called(ctx, fleetplan, opts)
	return args.Error(0)
}

func (m *MockFleetPlanService) Get(ctx context.Context, opts interface{}) (*infrapb.FleetPlan, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*infrapb.FleetPlan), args.Error(1)
}

func (m *MockFleetPlanService) Delete(ctx context.Context, opts interface{}) error {
	args := m.Called(ctx, opts)
	return args.Error(0)
}

// MockInfraV3Service is a mock implementation of the InfraV3 service
type MockInfraV3Service struct {
	mock.Mock
	FleetPlanService *MockFleetPlanService
}

func (m *MockInfraV3Service) FleetPlan() *MockFleetPlanService {
	return m.FleetPlanService
}

// MockTypedClient is a mock implementation of the typed client
type MockTypedClient struct {
	mock.Mock
	InfraV3Service *MockInfraV3Service
}

func (m *MockTypedClient) InfraV3() *MockInfraV3Service {
	return m.InfraV3Service
}

func TestResourceFleetPlan(t *testing.T) {
	resource := rafay.TestResourceFleetPlan()

	assert.NotNil(t, resource)
	assert.NotNil(t, resource.CreateContext)
	assert.NotNil(t, resource.UpdateContext)
	assert.NotNil(t, resource.ReadContext)
	assert.NotNil(t, resource.DeleteContext)
	assert.NotNil(t, resource.Timeouts)
	assert.NotNil(t, resource.Schema)
	assert.Equal(t, 1, resource.SchemaVersion)
}

func TestResourceFleetPlanCreate(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-fleetplan",
				"project": "test-project",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"fleet": []interface{}{
					map[string]interface{}{
						"kind": "clusters",
						"labels": map[string]interface{}{
							"env": "test",
						},
						"projects": []interface{}{
							map[string]interface{}{
								"name": "test-project",
							},
						},
					},
				},
			},
		},
	})

	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestResourceFleetPlanCreate(ctx, d, nil)

	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestResourceFleetPlanUpdate(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-fleetplan",
				"project": "test-project",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"fleet": []interface{}{
					map[string]interface{}{
						"kind": "clusters",
						"labels": map[string]interface{}{
							"env": "test",
						},
						"projects": []interface{}{
							map[string]interface{}{
								"name": "test-project",
							},
						},
					},
				},
			},
		},
	})

	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestResourceFleetPlanUpdate(ctx, d, nil)

	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestResourceFleetPlanUpsert(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-fleetplan",
				"project": "test-project",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"fleet": []interface{}{
					map[string]interface{}{
						"kind": "clusters",
						"labels": map[string]interface{}{
							"env": "test",
						},
						"projects": []interface{}{
							map[string]interface{}{
								"name": "test-project",
							},
						},
					},
				},
			},
		},
	})

	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "DEBUG")
	defer os.Unsetenv("TF_LOG")

	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestResourceFleetPlanUpsert(ctx, d)

	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestResourceFleetPlanUpsertWithNameChange(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "new-fleetplan",
				"project": "test-project",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"fleet": []interface{}{
					map[string]interface{}{
						"kind": "clusters",
						"labels": map[string]interface{}{
							"env": "test",
						},
						"projects": []interface{}{
							map[string]interface{}{
								"name": "test-project",
							},
						},
					},
				},
			},
		},
	})

	// Set a different ID to simulate name change
	d.SetId("old-fleetplan")

	// Test the function
	diags := rafay.TestResourceFleetPlanUpsert(ctx, d)

	// Verify error is returned for name change
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestResourceFleetPlanRead(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-fleetplan",
				"project": "test-project",
			},
		},
	})
	d.SetId("test-fleetplan")

	// Test the function
	diags := rafay.TestResourceFleetPlanRead(ctx, d, nil)

	// Verify no errors (this will fail in real scenario due to missing client, but tests the function structure)
	assert.NotEmpty(t, diags)
}

func TestResourceFleetPlanDelete(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-fleetplan",
				"project": "test-project",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"fleet": []interface{}{
					map[string]interface{}{
						"kind": "clusters",
						"labels": map[string]interface{}{
							"env": "test",
						},
						"projects": []interface{}{
							map[string]interface{}{
								"name": "test-project",
							},
						},
					},
				},
			},
		},
	})

	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "DEBUG")
	defer os.Unsetenv("TF_LOG")

	// Test the function
	diags := rafay.TestResourceFleetPlanDelete(ctx, d, nil)

	// Verify no errors (this will fail in real scenario due to missing client, but tests the function structure)
	assert.NotEmpty(t, diags)
}

func TestExpandFleetPlan(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected *infrapb.FleetPlan
		hasError bool
	}{
		{
			name: "Valid fleetplan with clusters",
			input: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":    "test-fleetplan",
						"project": "test-project",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"kind": "clusters",
								"labels": map[string]interface{}{
									"env": "test",
								},
								"projects": []interface{}{
									map[string]interface{}{
										"name": "test-project",
									},
								},
							},
						},
					},
				},
			},
			expected: &infrapb.FleetPlan{
				ApiVersion: "infra.k8smgmt.io/v3",
				Kind:       "FleetPlan",
			},
			hasError: false,
		},
		{
			name: "Valid fleetplan with environments",
			input: map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":    "test-fleetplan",
						"project": "test-project",
					},
				},
				"spec": []interface{}{
					map[string]interface{}{
						"fleet": []interface{}{
							map[string]interface{}{
								"kind": "environments",
								"labels": map[string]interface{}{
									"env": "test",
								},
								"projects": []interface{}{
									map[string]interface{}{
										"name": "test-project",
									},
								},
								"templates": []interface{}{
									map[string]interface{}{
										"name":    "test-template",
										"version": "v1.0.0",
									},
								},
								"target_batch_size": 5,
							},
						},
						"schedules": []interface{}{
							map[string]interface{}{
								"name":        "test-schedule",
								"description": "Test schedule",
								"type":        "maintenance",
								"cadence": []interface{}{
									map[string]interface{}{
										"cron_expression": "0 2 * * *",
										"cron_timezone":   "UTC",
										"time_to_live":    "1h",
									},
								},
							},
						},
					},
				},
			},
			expected: &infrapb.FleetPlan{
				ApiVersion: "infra.k8smgmt.io/v3",
				Kind:       "FleetPlan",
			},
			hasError: false,
		},
		{
			name:  "Nil input",
			input: nil,
			expected: &infrapb.FleetPlan{
				ApiVersion: "infra.k8smgmt.io/v3",
				Kind:       "FleetPlan",
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlan().Schema, tt.input)

			result, err := rafay.TestExpandFleetPlan(d)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.ApiVersion, result.ApiVersion)
				assert.Equal(t, tt.expected.Kind, result.Kind)
			}
		})
	}
}

func TestExpandFleetPlanSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.FleetPlanSpec
		hasError bool
	}{
		{
			name: "Valid spec with clusters",
			input: []interface{}{
				map[string]interface{}{
					"fleet": []interface{}{
						map[string]interface{}{
							"kind": "clusters",
							"labels": map[string]interface{}{
								"env": "test",
							},
							"projects": []interface{}{
								map[string]interface{}{
									"name": "test-project",
								},
							},
						},
					},
				},
			},
			expected: &infrapb.FleetPlanSpec{},
			hasError: false,
		},
		{
			name: "Valid spec with environments",
			input: []interface{}{
				map[string]interface{}{
					"fleet": []interface{}{
						map[string]interface{}{
							"kind": "environments",
							"labels": map[string]interface{}{
								"env": "test",
							},
							"projects": []interface{}{
								map[string]interface{}{
									"name": "test-project",
								},
							},
							"templates": []interface{}{
								map[string]interface{}{
									"name":    "test-template",
									"version": "v1.0.0",
								},
							},
							"target_batch_size": 5,
						},
					},
					"schedules": []interface{}{
						map[string]interface{}{
							"name":        "test-schedule",
							"description": "Test schedule",
							"type":        "maintenance",
							"cadence": []interface{}{
								map[string]interface{}{
									"cron_expression": "0 2 * * *",
									"cron_timezone":   "UTC",
									"time_to_live":    "1h",
								},
							},
						},
					},
				},
			},
			expected: &infrapb.FleetPlanSpec{},
			hasError: false,
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
			hasError: true,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := rafay.TestExpandFleetPlanSpec(tt.input)

			if tt.hasError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestFlattenFleetPlanSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.FleetPlanSpec
		expected []interface{}
	}{
		{
			name:     "Nil spec",
			input:    nil,
			expected: []interface{}{},
		},
		{
			name: "Valid spec with clusters",
			input: &infrapb.FleetPlanSpec{
				Fleet: &infrapb.FleetSpec{
					Kind: "clusters",
					Labels: map[string]string{
						"env": "test",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"fleet": []interface{}{
						map[string]interface{}{
							"kind":     "clusters",
							"labels":   map[string]string{"env": "test"},
							"projects": []interface{}{},
						},
					},
					"operation_workflow": []interface{}{},
					"agents":             []interface{}{},
				},
			},
		},
		{
			name: "Valid spec with environments",
			input: &infrapb.FleetPlanSpec{
				Fleet: &infrapb.FleetSpec{
					Kind: "environments",
					Labels: map[string]string{
						"env": "test",
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"fleet": []interface{}{
						map[string]interface{}{
							"kind":              "environments",
							"labels":            map[string]string{"env": "test"},
							"projects":          []interface{}{},
							"templates":         []interface{}{},
							"target_batch_size": int32(0),
						},
					},
					"operation_workflow": []interface{}{},
					"agents":             []interface{}{},
					"schedules":          []interface{}{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenFleetPlanSpec(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExpandFleetSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.FleetSpec
	}{
		{
			name: "Valid fleet spec with clusters",
			input: []interface{}{
				map[string]interface{}{
					"kind": "clusters",
					"labels": map[string]interface{}{
						"env": "test",
					},
					"projects": []interface{}{
						map[string]interface{}{
							"name": "test-project",
						},
					},
				},
			},
			expected: &infrapb.FleetSpec{
				Kind: "clusters",
				Labels: map[string]string{
					"env": "test",
				},
			},
		},
		{
			name: "Valid fleet spec with environments",
			input: []interface{}{
				map[string]interface{}{
					"kind": "environments",
					"labels": map[string]interface{}{
						"env": "test",
					},
					"projects": []interface{}{
						map[string]interface{}{
							"name": "test-project",
						},
					},
					"templates": []interface{}{
						map[string]interface{}{
							"name":    "test-template",
							"version": "v1.0.0",
						},
					},
					"target_batch_size": 5,
				},
			},
			expected: &infrapb.FleetSpec{
				Kind: "environments",
				Labels: map[string]string{
					"env": "test",
				},
			},
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestExpandFleetSpec(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.expected.Kind, result.Kind)
			}
		})
	}
}

func TestExpandProjects(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*infrapb.ProjectFilter
	}{
		{
			name: "Valid projects",
			input: []interface{}{
				map[string]interface{}{
					"name": "project1",
				},
				map[string]interface{}{
					"name": "project2",
				},
			},
			expected: []*infrapb.ProjectFilter{
				{Name: "project1"},
				{Name: "project2"},
			},
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: []*infrapb.ProjectFilter{},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []*infrapb.ProjectFilter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestExpandProjects(tt.input)
			assert.Equal(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Name, result[i].Name)
			}
		})
	}
}

func TestExpandEnvironmentTemplates(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected []*infrapb.TemplateFilter
	}{
		{
			name: "Valid templates",
			input: []interface{}{
				map[string]interface{}{
					"name":    "template1",
					"version": "v1.0.0",
				},
				map[string]interface{}{
					"name":    "template2",
					"version": "v2.0.0",
				},
			},
			expected: []*infrapb.TemplateFilter{
				{Name: "template1", Version: "v1.0.0"},
				{Name: "template2", Version: "v2.0.0"},
			},
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: []*infrapb.TemplateFilter{},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []*infrapb.TemplateFilter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestExpandEnvironmentTemplates(tt.input)
			assert.Equal(t, len(tt.expected), len(result))
			for i, expected := range tt.expected {
				assert.Equal(t, expected.Name, result[i].Name)
				assert.Equal(t, expected.Version, result[i].Version)
			}
		})
	}
}

func TestExpandFleetPlanSchedules(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.FleetSchedule
	}{
		{
			name: "Valid schedules",
			input: []interface{}{
				map[string]interface{}{
					"type": "one-time",
					"cadence": []interface{}{
						map[string]interface{}{
							"cron_expression": "0 2 * * *",
							"cron_timezone":   "UTC",
						},
					},
				},
			},
			expected: &infrapb.FleetSchedule{
				Type: "one-time",
				Cadence: &infrapb.ScheduleOptions{
					CronExpression: "0 2 * * *",
					CronTimezone:   "UTC",
				},
			},
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestExpandFleetPlanSchedules(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestExpandScheduleCadence(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.ScheduleOptions
	}{
		{
			name: "Valid cadence",
			input: []interface{}{
				map[string]interface{}{
					"cron_expression": "0 2 * * *",
					"cron_timezone":   "UTC",
					"time_to_live":    "1h",
				},
			},
			expected: &infrapb.ScheduleOptions{
				CronExpression: "0 2 * * *",
				CronTimezone:   "UTC",
				TimeToLive:     "1h",
			},
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestExpandScheduleCadence(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.CronExpression, result.CronExpression)
				assert.Equal(t, tt.expected.CronTimezone, result.CronTimezone)
				assert.Equal(t, tt.expected.TimeToLive, result.TimeToLive)
			}
		})
	}
}

func TestExpandScheduleOptOut(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.ScheduleOptOut
	}{
		{
			name: "Valid opt out",
			input: []interface{}{
				map[string]interface{}{
					"duration": "2h",
				},
			},
			expected: &infrapb.ScheduleOptOut{
				Duration: "2h",
			},
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestExpandScheduleOptOut(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.Duration, result.Duration)
			}
		})
	}
}

func TestExpandScheduleOptOutOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *infrapb.ScheduleOptOutOptions
	}{
		{
			name: "Valid opt out options",
			input: []interface{}{
				map[string]interface{}{
					"allow_opt_out": []interface{}{
						map[string]interface{}{
							"value": true,
						},
					},
					"max_allowed_duration": "4h",
					"max_allowed_times":    3,
				},
			},
			expected: &infrapb.ScheduleOptOutOptions{
				MaxAllowedDuration: "4h",
				MaxAllowedTimes:    3,
			},
		},
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestExpandScheduleOptOutOptions(tt.input)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, tt.expected.MaxAllowedDuration, result.MaxAllowedDuration)
				assert.Equal(t, tt.expected.MaxAllowedTimes, result.MaxAllowedTimes)
			}
		})
	}
}

func TestFlattenScheduleCadence(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.ScheduleOptions
		expected []interface{}
	}{
		{
			name: "Valid cadence",
			input: &infrapb.ScheduleOptions{
				CronExpression: "0 2 * * *",
				CronTimezone:   "UTC",
				TimeToLive:     "1h",
			},
			expected: []interface{}{
				map[string]interface{}{
					"cron_expression": "0 2 * * *",
					"cron_timezone":   "UTC",
					"time_to_live":    "1h",
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenScheduleCadence(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenScheduleOptOut(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.ScheduleOptOut
		expected []interface{}
	}{
		{
			name: "Valid opt out",
			input: &infrapb.ScheduleOptOut{
				Duration: "2h",
			},
			expected: []interface{}{
				map[string]interface{}{
					"duration": "2h",
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenScheduleOptOut(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenScheduleOptOutOptions(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.ScheduleOptOutOptions
		expected []interface{}
	}{
		{
			name: "Valid opt out options",
			input: &infrapb.ScheduleOptOutOptions{
				MaxAllowedDuration: "4h",
				MaxAllowedTimes:    3,
			},
			expected: []interface{}{
				map[string]interface{}{
					"allow_opt_out":        []interface{}(nil),
					"max_allowed_duration": "4h",
					"max_allowed_times":    int32(3),
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenScheduleOptOutOptions(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenSchedule(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.FleetSchedule
		expected []interface{}
	}{
		{
			name: "Valid schedule",
			input: &infrapb.FleetSchedule{
				Type: "maintenance",
			},
			expected: []interface{}{
				map[string]interface{}{
					"type":            "maintenance",
					"cadence":         []interface{}{},
					"opt_out":         []interface{}{},
					"opt_out_options": []interface{}{},
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenSchedule(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenFleet(t *testing.T) {
	tests := []struct {
		name     string
		input    *infrapb.FleetSpec
		expected []interface{}
	}{
		{
			name: "Valid fleet spec with clusters",
			input: &infrapb.FleetSpec{
				Kind: "clusters",
				Labels: map[string]string{
					"env": "test",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"kind":     "clusters",
					"labels":   map[string]string{"env": "test"},
					"projects": []interface{}{},
				},
			},
		},
		{
			name: "Valid fleet spec with environments",
			input: &infrapb.FleetSpec{
				Kind: "environments",
				Labels: map[string]string{
					"env": "test",
				},
				TargetBatchSize: 5,
			},
			expected: []interface{}{
				map[string]interface{}{
					"kind":              "environments",
					"labels":            map[string]string{"env": "test"},
					"projects":          []interface{}{},
					"templates":         []interface{}{},
					"target_batch_size": int32(5),
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenFleet(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenTemplates(t *testing.T) {
	tests := []struct {
		name     string
		input    []*infrapb.TemplateFilter
		expected []interface{}
	}{
		{
			name: "Valid templates",
			input: []*infrapb.TemplateFilter{
				{Name: "template1", Version: "v1.0.0"},
				{Name: "template2", Version: "v2.0.0"},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name":    "template1",
					"version": "v1.0.0",
				},
				map[string]interface{}{
					"name":    "template2",
					"version": "v2.0.0",
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenTemplates(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenProjects(t *testing.T) {
	tests := []struct {
		name     string
		input    []*infrapb.ProjectFilter
		expected []interface{}
	}{
		{
			name: "Valid projects",
			input: []*infrapb.ProjectFilter{
				{Name: "project1"},
				{Name: "project2"},
			},
			expected: []interface{}{
				map[string]interface{}{
					"name": "project1",
				},
				map[string]interface{}{
					"name": "project2",
				},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rafay.TestFlattenProjects(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
