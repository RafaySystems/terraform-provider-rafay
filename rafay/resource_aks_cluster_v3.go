package rafay

import (
	"context"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAKSClusterV3() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAKSClusterV3Create,
		ReadContext:   resourceAKSClusterV3Read,
		UpdateContext: resourceAKSClusterV3Update,
		DeleteContext: resourceAKSClusterV3Delete,
		Importer: &schema.ResourceImporter{
			State: resourceAKSClusterV3Import,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterSchema.Schema,
	}
}

func resourceAKSClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// calls upsert
	return resourceAKSClusterV3Upsert(ctx, d, m)
}

func resourceAKSClusterV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceAKSClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// calls upsert
	return resourceAKSClusterV3Upsert(ctx, d, m)
}

func resourceAKSClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceAKSClusterV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

func resourceAKSClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
