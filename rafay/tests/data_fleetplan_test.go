package tests

import (
	"context"
	"testing"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestDataFleetplan(t *testing.T) {
	resource := rafay.TestDataFleetplan()
	
	assert.NotNil(t, resource)
	assert.NotNil(t, resource.ReadContext)
	assert.NotNil(t, resource.Timeouts)
	assert.NotNil(t, resource.Schema)
	assert.Equal(t, 1, resource.SchemaVersion)
	
	// Test timeout values
	assert.Equal(t, 10*60, int(resource.Timeouts.Read.Seconds())) // 10 minutes
}

func TestDataFleetplanRead(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-fleetplan",
				"project": "test-project",
			},
		},
	})
	
	// Test the function - this will fail due to missing client, but tests the function structure
	diags := rafay.TestDataFleetplanRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanSchema(t *testing.T) {
	resource := rafay.TestDataFleetplan()
	schema := resource.Schema
	
	// Test that schema is not nil
	assert.NotNil(t, schema)
	
	// Test that schema has the expected structure (inherited from FleetPlanSchema)
	assert.Contains(t, schema, "metadata")
	assert.Contains(t, schema, "spec")
}

func TestDataFleetplanWithEmptyData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplan().Schema, map[string]interface{}{})
	
	// Test with empty data - should still work but will fail due to missing client
	diags := rafay.TestDataFleetplanRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanWithValidData(t *testing.T) {
	ctx := context.Background()
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplan().Schema, map[string]interface{}{
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
					},
				},
			},
		},
	})
	
	// Test the function
	diags := rafay.TestDataFleetplanRead(ctx, d, nil)
	
	// Verify that errors are returned due to missing client (expected behavior)
	assert.NotEmpty(t, diags)
	assert.True(t, diags.HasError())
}

func TestDataFleetplanTimeouts(t *testing.T) {
	resource := rafay.TestDataFleetplan()
	timeouts := resource.Timeouts
	
	// Test Read timeout (10 minutes)
	assert.NotNil(t, timeouts.Read)
	assert.Equal(t, 10*60, int(timeouts.Read.Seconds()))
}

func TestDataFleetplanResourceData(t *testing.T) {
	d := schema.TestResourceDataRaw(t, rafay.TestDataFleetplan().Schema, map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":    "test-fleetplan",
				"project": "test-project",
			},
		},
	})
	
	// Test that data is set correctly
	metadata := d.Get("metadata").([]interface{})
	assert.NotNil(t, metadata)
	assert.Len(t, metadata, 1)
	
	metadataMap := metadata[0].(map[string]interface{})
	assert.Equal(t, "test-fleetplan", metadataMap["name"])
	assert.Equal(t, "test-project", metadataMap["project"])
}
