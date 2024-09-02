package convertor

import (
	"context"
	"time"

	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
)

func WaitForClusterOperation(ctx context.Context, client typed.Client, cluster *infrapb.Cluster, timeout <-chan time.Time, ticker *time.Ticker) diag.Diagnostics {
	var diags diag.Diagnostics
	for {
		select {
		case <-timeout:
			// Timeout reached
			diags.AddError("Timeout reached while waiting for cluster operation to complete", "")
			return diags

		case <-ticker.C:
			uCluster, err := client.InfraV3().Cluster().Status(ctx, options.StatusOptions{
				Name:    cluster.Metadata.Name,
				Project: cluster.Metadata.Project,
			})
			if err != nil {
				// Error occurred while fetching cluster status
				diags.AddError("Error occurred while fetching cluster status", err.Error())
				return diags
			}

			if uCluster == nil {
				continue
			}

			if uCluster.Status != nil && uCluster.Status.Mks != nil {
				uClusterCommonStatus := uCluster.Status.CommonStatus
				switch uClusterCommonStatus.ConditionStatus {
				case commonpb.ConditionStatus_StatusSubmitted:
					// Submitted
					continue
				case commonpb.ConditionStatus_StatusOK:
					// Completed
					return diags
				case commonpb.ConditionStatus_StatusFailed:
					failureReason := uClusterCommonStatus.Reason
					diags.AddError("Cluster operation failed", failureReason)
				}
			}
		}
	}
}
