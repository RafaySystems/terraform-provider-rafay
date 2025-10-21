package tests

import (
	"context"
	"os"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestDataFleetplans(t *testing.T) {
	resource := rafay.TestDataFleetplans()
	
	assert.NotNil(t, resource)
	assert.NotNil(t, resource.ReadContext)
	assert.NotNil(t, resource.Timeouts)
	assert.NotNil(t, resource.Schema)
	assert.Equal(t, 1, resource.SchemaVersion)
	
	// Test timeout values
	assert.Equal(t, 10*60, int(resource.Timeouts.Read.Seconds())) // 10 minutes
}

func TestDataFleetplansRead(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplans().Schema, map[string]interface{}{
		"project": "test-project",
		"type":    "clusters",
	})
	
	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestDataFleetplansRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplansSchema(t *testing.T) {
	resource := rafay.TestDataFleetplans()
	schema := resource.Schema
	
	// Test project field
	projectField := schema["project"]
	assert.NotNil(t, projectField)
	assert.True(t, projectField.Required)
	assert.Equal(t, "Project name from where fleetplans to be listed", projectField.Description)
	
	// Test type field
	typeField := schema["type"]
	assert.NotNil(t, typeField)
	assert.False(t, typeField.Required)
	assert.Equal(t, "clusters", typeField.Default)
	assert.Equal(t, "Resource type of the fleet plan", typeField.Description)
	
	// Test fleetplans field
	fleetplansField := schema["fleetplans"]
	assert.NotNil(t, fleetplansField)
	assert.True(t, fleetplansField.Computed)
	assert.Equal(t, "List of fleetplans", fleetplansField.Description)
}

func TestDataFleetplansWithDefaultType(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplans().Schema, map[string]interface{}{
		"project": "test-project",
		// type not specified, should use default
	})
	
	// Test the function
	diags := rafay.TestDataFleetplansRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplansWithEnvironmentsType(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplans().Schema, map[string]interface{}{
		"project": "test-project",
		"type":    "environments",
	})
	
	// Test the function
	diags := rafay.TestDataFleetplansRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplansWithEmptyData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplans().Schema, map[string]interface{}{})
	
	// Test with empty data - should still work but will fail due to missing client
	diags := rafay.TestDataFleetplansRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplansWithDebugLogging(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplans().Schema, map[string]interface{}{
		"project": "test-project",
		"type":    "clusters",
	})
	
	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "DEBUG")
	defer os.Unsetenv("TF_LOG")
	
	// Test the function
	diags := rafay.TestDataFleetplansRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplansWithTraceLogging(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplans().Schema, map[string]interface{}{
		"project": "test-project",
		"type":    "clusters",
	})
	
	// Test with TF_LOG environment variable
	os.Setenv("TF_LOG", "TRACE")
	defer os.Unsetenv("TF_LOG")
	
	// Test the function
	diags := rafay.TestDataFleetplansRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplansTimeouts(t *testing.T) {
	resource := rafay.TestDataFleetplans()
	timeouts := resource.Timeouts
	
	// Test Read timeout (10 minutes)
	assert.NotNil(t, timeouts.Read)
	assert.Equal(t, 10*60, int(timeouts.Read.Seconds()))
}

func TestDataFleetplansResourceData(t *testing.T) {
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplans().Schema, map[string]interface{}{
		"project": "test-project",
		"type":    "clusters",
	})
	
	// Test that data is set correctly
	assert.Equal(t, "test-project", d.Get("project"))
	assert.Equal(t, "clusters", d.Get("type"))
}

func TestDataFleetplansFleetplansSchema(t *testing.T) {
	resource := rafay.TestDataFleetplans()
	schema := resource.Schema
	
	// Test fleetplans field structure
	fleetplansField := schema["fleetplans"]
	assert.NotNil(t, fleetplansField)
	assert.True(t, fleetplansField.Computed)
	assert.Equal(t, "List of fleetplans", fleetplansField.Description)
}
