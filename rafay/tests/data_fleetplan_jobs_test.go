package tests

import (
	"context"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestDataFleetplanJobs(t *testing.T) {
	resource := rafay.TestDataFleetplanJobs()
	
	assert.NotNil(t, resource)
	assert.NotNil(t, resource.ReadContext)
	assert.NotNil(t, resource.Timeouts)
	assert.NotNil(t, resource.Schema)
	assert.Equal(t, 1, resource.SchemaVersion)
	
	// Test timeout values
	assert.Equal(t, 10*60, int(resource.Timeouts.Read.Seconds())) // 10 minutes
}

func TestDataFleetplanJobsRead(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJobs().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
	})
	
	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestDataFleetplanJobsRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobsSchema(t *testing.T) {
	resource := rafay.TestDataFleetplanJobs()
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
	
	// Test jobs field
	jobsField := schema["jobs"]
	assert.NotNil(t, jobsField)
	assert.True(t, jobsField.Computed)
	assert.Equal(t, "List of fleetplan jobs", jobsField.Description)
}

func TestDataFleetplanJobsWithEmptyData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJobs().Schema, map[string]interface{}{})
	
	// Test with empty data - should still work but will fail due to missing client
	diags := rafay.TestDataFleetplanJobsRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobsWithValidData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJobs().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
	})
	
	// Test the function
	diags := rafay.TestDataFleetplanJobsRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobsTimeouts(t *testing.T) {
	resource := rafay.TestDataFleetplanJobs()
	timeouts := resource.Timeouts
	
	// Test Read timeout (10 minutes)
	assert.NotNil(t, timeouts.Read)
	assert.Equal(t, 10*60, int(timeouts.Read.Seconds()))
}

func TestDataFleetplanJobsResourceData(t *testing.T) {
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJobs().Schema, map[string]interface{}{
		"project":        "test-project",
		"fleetplan_name": "test-fleetplan",
	})
	
	// Test that data is set correctly
	assert.Equal(t, "test-project", d.Get("project"))
	assert.Equal(t, "test-fleetplan", d.Get("fleetplan_name"))
}

func TestDataFleetplanJobsJobsSchema(t *testing.T) {
	resource := rafay.TestDataFleetplanJobs()
	schema := resource.Schema
	
	// Test jobs field structure
	jobsField := schema["jobs"]
	assert.NotNil(t, jobsField)
	assert.True(t, jobsField.Computed)
	assert.Equal(t, "List of fleetplan jobs", jobsField.Description)
}

func TestDataFleetplanJobsWithDifferentProjects(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplanJobs().Schema, map[string]interface{}{
		"project":        "different-project",
		"fleetplan_name": "different-fleetplan",
	})
	
	// Test the function
	diags := rafay.TestDataFleetplanJobsRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanJobsSchemaValidation(t *testing.T) {
	resource := rafay.TestDataFleetplanJobs()
	schema := resource.Schema
	
	// Test that all required fields are present
	assert.Contains(t, schema, "project")
	assert.Contains(t, schema, "fleetplan_name")
	assert.Contains(t, schema, "jobs")
	
	// Test that required fields are actually required
	assert.True(t, schema["project"].Required)
	assert.True(t, schema["fleetplan_name"].Required)
	
	// Test that computed fields are actually computed
	assert.True(t, schema["jobs"].Computed)
}
