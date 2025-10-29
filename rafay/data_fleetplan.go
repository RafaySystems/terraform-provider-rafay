package rafay

import (
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataFleetplan() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceFleetPlanRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute), // 10 minutes
		},
		SchemaVersion: 1,
		Schema:        resource.FleetPlanSchema.Schema,
	}
}
