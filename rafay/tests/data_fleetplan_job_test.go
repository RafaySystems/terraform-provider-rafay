package tests

import (
	"context"
	"os"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestDataFleetplanJob(t *testing.T) {
	resource := rafay.TestDataFleetplanJob()
	
	assert.NotNil(t, resource)
	assert.NotNil(t, resource.ReadContext)
	assert.NotNil(t, resource.Timeouts)
	assert.NotNil(t, resource.Schema)
	assert.Equal(t, 1, resource.SchemaVersion)
	
	// Test timeout values
	assert.Equal(t, 10*60, int(resource.Timeouts.Read.Seconds())) // 10 minutes
}

func TestDataFleetplanJobRead(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJob().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
		"name":           "test-job",
	})
	
	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestDataFleetplanJobRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobSchema(t *testing.T) {
	resource := rafay.TestDataFleetplanJob()
	schema := resource.Schema
	
	// Test project field
	projectField := schema["project"]
	assert.NotNil(t, projectField)
	assert.True(t, projectField.Required)
	assert.Equal(t, "Project name from where environments to be listed", projectField.Description)
	
	// Test fleetplan_name field
	fleetplanNameField := schema["fleetplan_name"]
	assert.NotNil(t, fleetplanNameField)
	assert.True(t, fleetplanNameField.Required)
	assert.Equal(t, "FleetPlan name", fleetplanNameField.Description)
	
	// Test name field
	nameField := schema["name"]
	assert.NotNil(t, nameField)
	assert.True(t, nameField.Required)
	assert.Equal(t, "FleetPlan job name", nameField.Description)
	
	// Test status field
	statusField := schema["status"]
	assert.NotNil(t, statusField)
	assert.True(t, statusField.Computed)
	assert.Equal(t, "Fleetplan job", statusField.Description)
}

func TestDataFleetplanJobWithEmptyData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJob().Schema, map[string]interface{}{})
	
	// Test with empty data - should still work but will fail due to missing client
	diags := rafay.TestDataFleetplanJobRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobWithValidData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJob().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
		"name":           "test-job",
	})
	
	// Test the function
	diags := rafay.TestDataFleetplanJobRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobWithDebugLogging(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJob().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
		"name":           "test-job",
	})
	
	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "DEBUG")
	defer os.Unsetenv("TF_LOG")
	
	// Test the function
	diags := rafay.TestDataFleetplanJobRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobWithTraceLogging(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJob().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
		"name":           "test-job",
	})
	
	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "TRACE")
	defer os.Unsetenv("TF_LOG")
	
	// Test the function
	diags := rafay.TestDataFleetplanJobRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobTimeouts(t *testing.T) {
	resource := rafay.TestDataFleetplanJob()
	timeouts := resource.Timeouts
	
	// Test Read timeout (10 minutes)
	assert.NotNil(t, timeouts.Read)
	assert.Equal(t, 10*60, int(timeouts.Read.Seconds()))
}

func TestDataFleetplanJobResourceData(t *testing.T) {
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJob().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
		"name":           "test-job",
	})
	
	// Test that data is set correctly
	assert.Equal(t, "test-project", d.Get("project"))
	assert.Equal(t, "test-fleetplan", d.Get("fleetplan_name"))
	assert.Equal(t, "test-job", d.Get("name"))
}

func TestDataFleetplanJobStatusSchema(t *testing.T) {
	resource := rafay.TestDataFleetplanJob()
	schema := resource.Schema
	
	// Test status field structure
	statusField := schema["status"]
	assert.NotNil(t, statusField)
	assert.True(t, statusField.Computed)
	assert.Equal(t, "Fleetplan job", statusField.Description)
}
