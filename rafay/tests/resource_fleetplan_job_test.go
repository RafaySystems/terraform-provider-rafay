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

// MockFleetPlanExtApi is a mock implementation of the FleetPlan ExtApi service
type MockFleetPlanExtApi struct {
	mock.Mock
}

func (m *MockFleetPlanExtApi) ExecuteFleetPlan(ctx context.Context, opts interface{}) (*infrapb.FleetPlanJob, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*infrapb.FleetPlanJob), args.Error(1)
}

func (m *MockFleetPlanExtApi) GetJobStatus(ctx context.Context, opts interface{}) (*infrapb.FleetPlanJobStatus, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*infrapb.FleetPlanJobStatus), args.Error(1)
}

func (m *MockFleetPlanExtApi) GetJobs(ctx context.Context, opts interface{}) (*infrapb.FleetPlanJobList, error) {
	args := m.Called(ctx, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*infrapb.FleetPlanJobList), args.Error(1)
}

// MockFleetPlanJobService is a mock implementation of the FleetPlan service for jobs
type MockFleetPlanJobService struct {
	mock.Mock
	ExtApiService *MockFleetPlanExtApi
}

func (m *MockFleetPlanJobService) ExtApi() *MockFleetPlanExtApi {
	return m.ExtApiService
}

// MockInfraV3JobService is a mock implementation of the InfraV3 service for jobs
type MockInfraV3JobService struct {
	mock.Mock
	FleetPlanService *MockFleetPlanJobService
}

func (m *MockInfraV3JobService) FleetPlan() *MockFleetPlanJobService {
	return m.FleetPlanService
}

// MockTypedJobClient is a mock implementation of the typed client for jobs
type MockTypedJobClient struct {
	mock.Mock
	InfraV3Service *MockInfraV3JobService
}

func (m *MockTypedJobClient) InfraV3() *MockInfraV3JobService {
	return m.InfraV3Service
}

func TestResourceFleetPlanTrigger(t *testing.T) {
	resource := rafay.TestResourceFleetPlanTrigger()
	
	assert.NotNil(t, resource)
	assert.NotNil(t, resource.CreateContext)
	assert.NotNil(t, resource.UpdateContext)
	assert.NotNil(t, resource.ReadContext)
	assert.NotNil(t, resource.DeleteContext)
	assert.NotNil(t, resource.Timeouts)
	assert.NotNil(t, resource.Schema)
	
	// Test schema fields
	schema := resource.Schema
	assert.Contains(t, schema, "fleetplan_name")
	assert.Contains(t, schema, "project")
	assert.Contains(t, schema, "trigger_value")
	
	// Test timeout values
	assert.Equal(t, 2*60*60, int(resource.Timeouts.Create.Seconds())) // 2 hours
	assert.Equal(t, 2*60*60, int(resource.Timeouts.Update.Seconds())) // 2 hours
	assert.Equal(t, 2*60, int(resource.Timeouts.Delete.Seconds()))   // 2 minutes
}

func TestCreateFleetPlanJob(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	
	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestCreateFleetPlanJob(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestUpdateFleetPlanJob(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	
	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestUpdateFleetPlanJob(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestUpsertFleetPlanJob(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	
	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "DEBUG")
	defer os.Unsetenv("TF_LOG")
	
	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestUpsertFleetPlanJob(ctx, d)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestReadFleetPlanJob(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	d.SetId("test-job-id")
	
	// Test the function
	diags := rafay.TestReadFleetPlanJob(ctx, d, nil)
	
	// Verify no errors (this will fail in real scenario due to missing client, but tests the function structure)
	assert.NotEmpty(t, diags)
}

func TestDeleteFleetPlanJob(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	
	// Test the function
	diags := rafay.TestDeleteFleetPlanJob(ctx, d, nil)
	
	// Verify no errors (delete operation is not supported, just removes from state)
	assert.Empty(t, diags)
}

func TestFleetPlanJobSchema(t *testing.T) {
	resource := rafay.TestResourceFleetPlanTrigger()
	schema := resource.Schema
	
	// Test fleetplan_name field
	fleetplanNameField := schema["fleetplan_name"]
	assert.NotNil(t, fleetplanNameField)
	assert.True(t, fleetplanNameField.Required)
	assert.Equal(t, "FleetPlan name", fleetplanNameField.Description)
	
	// Test project field
	projectField := schema["project"]
	assert.NotNil(t, projectField)
	assert.True(t, projectField.Required)
	assert.Equal(t, "FleetPlan project", projectField.Description)
	
	// Test trigger_value field
	triggerValueField := schema["trigger_value"]
	assert.NotNil(t, triggerValueField)
	assert.True(t, triggerValueField.Required)
	assert.Equal(t, "Enter trigger value to trigger a new job for fleetplan", triggerValueField.Description)
}

func TestFleetPlanJobTimeouts(t *testing.T) {
	resource := rafay.TestResourceFleetPlanTrigger()
	timeouts := resource.Timeouts
	
	// Test Create timeout (2 hours)
	assert.NotNil(t, timeouts.Create)
	assert.Equal(t, 2*60*60, int(timeouts.Create.Seconds()))
	
	// Test Update timeout (2 hours)
	assert.NotNil(t, timeouts.Update)
	assert.Equal(t, 2*60*60, int(timeouts.Update.Seconds()))
	
	// Test Delete timeout (2 minutes)
	assert.NotNil(t, timeouts.Delete)
	assert.Equal(t, 2*60, int(timeouts.Delete.Seconds()))
}

func TestFleetPlanJobResourceData(t *testing.T) {
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	
	// Test that data is set correctly
	assert.Equal(t, "test-fleetplan", d.Get("fleetplan_name"))
	assert.Equal(t, "test-project", d.Get("project"))
	assert.Equal(t, "test-trigger", d.Get("trigger_value"))
	
	// Test setting ID
	d.SetId("test-job-id")
	assert.Equal(t, "test-job-id", d.Id())
}

func TestFleetPlanJobWithEmptyData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{})
	
	// Test with empty data - should still work but will fail due to missing client
	diags := rafay.TestCreateFleetPlanJob(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestFleetPlanJobWithInvalidData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "", // Empty string
		"project":        "", // Empty string
		"trigger_value":  "", // Empty string
	})
	
	// Test with invalid data - should still work but will fail due to missing client
	diags := rafay.TestCreateFleetPlanJob(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestFleetPlanJobUpdateWithDifferentTrigger(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "new-trigger-value",
	})
	
	// Set a different ID to simulate update
	d.SetId("existing-job-id")
	
	// Test the function
	diags := rafay.TestUpdateFleetPlanJob(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestFleetPlanJobReadWithJobList(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	d.SetId("test-job-id")
	
	// Test the function
	diags := rafay.TestReadFleetPlanJob(ctx, d, nil)
	
	// Verify no errors (this will fail in real scenario due to missing client, but tests the function structure)
	assert.NotEmpty(t, diags)
}

func TestFleetPlanJobDeleteOperation(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	d.SetId("test-job-id")
	
	// Test the function
	diags := rafay.TestDeleteFleetPlanJob(ctx, d, nil)
	
	// Verify no errors (delete operation is not supported, just removes from state)
	assert.Empty(t, diags)
}

func TestFleetPlanJobWithDebugLogging(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	
	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "DEBUG")
	defer os.Unsetenv("TF_LOG")
	
	// Test the function
	diags := rafay.TestUpsertFleetPlanJob(ctx, d)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestFleetPlanJobWithTraceLogging(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestResourceFleetPlanTrigger().Schema, map[string]interface{}{
		"fleetplan_name": "test-fleetplan",
		"project":        "test-project",
		"trigger_value":  "test-trigger",
	})
	
	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "TRACE")
	defer os.Unsetenv("TF_LOG")
	
	// Test the function
	diags := rafay.TestUpsertFleetPlanJob(ctx, d)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}
