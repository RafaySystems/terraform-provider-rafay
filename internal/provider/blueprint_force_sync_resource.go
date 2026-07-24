package provider

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource               = &BlueprintForceSyncResource{}
	_ resource.ResourceWithModifyPlan = &BlueprintForceSyncResource{}
)

func NewBlueprintForceSyncResource() resource.Resource {
	return &BlueprintForceSyncResource{}
}

// BlueprintForceSyncResource triggers a blueprint sync on a cluster, optionally
// assigning a blueprint name/version to the cluster first. It talks to the
// backend directly via rctl (like the legacy SDKv2 resource it replaces)
// rather than through the typed hub client, so it needs no provider-configured
// client.
type BlueprintForceSyncResource struct{}

type BlueprintForceSyncModel struct {
	ID               types.String `tfsdk:"id"`
	ClusterName      types.String `tfsdk:"cluster_name"`
	Project          types.String `tfsdk:"project"`
	BlueprintName    types.String `tfsdk:"blueprint_name"`
	BlueprintVersion types.String `tfsdk:"blueprint_version"`
	ForceSync        types.Bool   `tfsdk:"force_sync"`
	Addons           types.List   `tfsdk:"addons"`
}

func (r *BlueprintForceSyncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_force_sync"
}

func (r *BlueprintForceSyncResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Triggers a blueprint sync on a cluster, optionally assigning a blueprint name/version first.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal identifier (cluster_name/project).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the cluster to sync the blueprint to.",
			},
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Project the cluster belongs to.",
			},
			"blueprint_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Name of the blueprint to assign to the cluster before syncing. Leave unset to keep the cluster's current blueprint. Always reflects the blueprint actually assigned on the cluster, even if a requested change fails to apply.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"blueprint_version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Version of the blueprint to assign to the cluster before syncing. Leave unset to keep the cluster's current blueprint version. Always reflects the blueprint version actually assigned on the cluster, even if a requested change fails to apply.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"force_sync": schema.BoolAttribute{
				Optional:    true,
				WriteOnly:   true,
				Description: "Passed through to the backend's blueprint publish call to control how it handles a sync already in progress: false errors out, true restarts it. Every apply re-publishes regardless of this value — it only changes what's sent to the backend, matching the UI's publish action. This value is never stored in state.",
			},
			"addons": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				WriteOnly:   true,
				Description: "Subset of addons to sync. Only valid with force_sync=true. Never stored in state.",
			},
		},
	}
}

// ModifyPlan unconditionally forces a diff on id, so Update runs on every
// apply — matching the UI, where clicking "publish" always calls the
// backend regardless of a force flag. No dedicated "trigger" attribute is
// needed: id already exists and is already Computed/UseStateForUnknown for
// every other case. Create/Update always recompute id deterministically
// from cluster_name and project, so this resolves cleanly once the apply
// completes.
//
// This never talks to the backend itself — it only shapes the diff that the
// user reviews before approving `terraform apply`; the actual publish call
// (with whatever force_sync is set to) only happens inside Create/Update.
func (r *BlueprintForceSyncResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans have a null plan/config; nothing to force.
	if req.Plan.Raw.IsNull() {
		return
	}

	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)

	_, _, diags := readWriteOnlySyncOpts(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
}

// readWriteOnlySyncOpts reads force_sync and addons from config. Both are
// write-only (never in plan/state), so callers must use Config rather than
// the model from Plan.Get. Returns a diagnostic if addons is set without
// force_sync=true.
func readWriteOnlySyncOpts(ctx context.Context, config tfsdk.Config) (forceSync bool, addons []string, diags diag.Diagnostics) {
	var forceSyncAttr types.Bool
	diags.Append(config.GetAttribute(ctx, path.Root("force_sync"), &forceSyncAttr)...)
	if diags.HasError() {
		return
	}

	var addonsList types.List
	diags.Append(config.GetAttribute(ctx, path.Root("addons"), &addonsList)...)
	if diags.HasError() {
		return
	}

	if !addonsList.IsNull() && !addonsList.IsUnknown() {
		diags.Append(addonsList.ElementsAs(ctx, &addons, false)...)
		if diags.HasError() {
			return
		}
	}

	forceSync = forceSyncAttr.ValueBool()
	if len(addons) > 0 && !forceSync {
		diags.AddAttributeError(
			path.Root("addons"),
			"Invalid Configuration",
			"addons can only be used with force_sync=true",
		)
	}
	return
}

// blueprintSyncOutcome carries the edge/project IDs needed to poll for sync
// completion, plus the blueprint name/version actually observed on the
// cluster (as opposed to what was merely requested).
type blueprintSyncOutcome struct {
	edgeID            string
	projectID         string
	observedBlueprint string
	observedVersion   string
}

// isBlueprintSyncInProgress reports whether the cluster's ClusterBlueprintSync
// condition is currently unsettled (in progress, pending, or retrying).
// Checking this upfront lets triggerBlueprintSync fail fast with a clear,
// consistent message when force_sync=false and a sync is already running —
// rather than a confusing error surfacing from UpdateCluster, which (unlike
// PublishClusterBlueprint) has no force-override of its own.
func isBlueprintSyncInProgress(edgeID, projectID string) (bool, error) {
	c, err := cluster.GetClusterWithEdgeID(edgeID, projectID, uaDef)
	if err != nil {
		return false, err
	}
	for _, condition := range c.Cluster.Conditions {
		if condition.Type != models.ClusterBlueprintSync {
			continue
		}
		switch condition.Status {
		case models.InProgress, models.Pending, models.Retry:
			return true, nil
		}
	}
	return false, nil
}

// triggerBlueprintSync resolves the project/cluster, updates the cluster's
// assigned blueprint if blueprintName/blueprintVersion differ from what's
// currently set, and publishes a blueprint sync.
//
// The returned outcome's observedBlueprint/observedVersion always reflect
// what is actually assigned on the cluster: the requested values only if the
// update call succeeded, otherwise whatever was already there.
//
// When addons is non-empty, forceSync must be true (validated by the
// caller) and publish goes through PublishBlueprintCluster so only that
// subset is synced; otherwise PublishClusterBlueprint is used.
func triggerBlueprintSync(clusterName, projectName string, forceSync bool, blueprintName, blueprintVersion string, addons []string) (*blueprintSyncOutcome, error) {
	log.Printf("blueprint sync starting for cluster: %s, project: %s, force_sync: %v, addons: %v", clusterName, projectName, forceSync, addons)

	if len(addons) > 0 && !forceSync {
		return nil, fmt.Errorf("addons can only be used with force_sync=true")
	}

	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %q: %w", projectName, err)
	}

	clusterResp, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster %q: %w", clusterName, err)
	}

	outcome := &blueprintSyncOutcome{
		edgeID:            clusterResp.ID,
		projectID:         projectID,
		observedBlueprint: clusterResp.ClusterBlueprint,
		observedVersion:   clusterResp.ClusterBlueprintVersion,
	}

	if !forceSync {
		inProgress, err := isBlueprintSyncInProgress(outcome.edgeID, projectID)
		if err != nil {
			log.Printf("warning: unable to determine blueprint sync status for cluster %s: %s — proceeding anyway", clusterName, err)
		} else if inProgress {
			return outcome, fmt.Errorf("a blueprint sync is already in progress for cluster %q; set force_sync=true to restart it", clusterName)
		}
	}

	blueprintChanged := false
	if blueprintName != "" && clusterResp.ClusterBlueprint != blueprintName {
		clusterResp.ClusterBlueprint = blueprintName
		blueprintChanged = true
	}
	if blueprintVersion != "" && clusterResp.ClusterBlueprintVersion != blueprintVersion {
		clusterResp.ClusterBlueprintVersion = blueprintVersion
		blueprintChanged = true
	}

	// The publish call's own Metadata.ForceSync flag isn't sufficient on
	// its own — the backend also expects the cluster's ForceBlueprintSync
	// field set via UpdateCluster before a forced publish, or the publish
	// call fails. So UpdateCluster must run whenever force_sync=true, not
	// just when the requested blueprint name/version actually changed.
	clusterResp.ForceBlueprintSync = forceSync
	if blueprintChanged || forceSync {
		log.Printf("updating cluster %s blueprint to name=%q version=%q force_blueprint_sync=%v before sync", clusterName, clusterResp.ClusterBlueprint, clusterResp.ClusterBlueprintVersion, forceSync)
		if err := cluster.UpdateCluster(clusterResp, uaDef); err != nil {
			// Update failed server-side: outcome keeps the pre-update
			// observed values so the caller doesn't record the attempted
			// (but never applied) blueprint into state.
			return outcome, fmt.Errorf("failed to update blueprint for cluster %q: %w", clusterName, err)
		}
		outcome.observedBlueprint = clusterResp.ClusterBlueprint
		outcome.observedVersion = clusterResp.ClusterBlueprintVersion
	}

	if len(addons) > 0 {
		if err := cluster.PublishBlueprintCluster(clusterName, projectID, outcome.observedBlueprint, outcome.observedVersion, forceSync, addons); err != nil {
			return outcome, fmt.Errorf("failed to publish blueprint for cluster %q: %w", clusterName, err)
		}
	} else {
		if err := cluster.PublishClusterBlueprint(clusterName, projectID, forceSync); err != nil {
			return outcome, fmt.Errorf("failed to publish blueprint for cluster %q: %w", clusterName, err)
		}
	}
	log.Printf("blueprint publish triggered for cluster: %s", clusterName)

	return outcome, nil
}

// blueprintSyncConditionStatus reports the terminal state of the cluster's
// ClusterBlueprintSync condition specifically. This is deliberately not
// getClusterConditions/ClusterReady: that's a general cluster-readiness
// signal meant for full cluster provisioning (used by the AKS/EKS
// resources), and on a cluster that's already up and running — the normal
// case here, since this resource resyncs an existing cluster rather than
// creating one — ClusterReady is typically already Success and stays that
// way regardless of whether the sync we just triggered succeeds or fails.
// Watching ClusterBlueprintSync's own status is the only reliable way to
// tell whether *this* sync actually completed.
func blueprintSyncConditionStatus(edgeID, projectID string) (failed bool, succeeded bool, err error) {
	c, err := cluster.GetClusterWithEdgeID(edgeID, projectID, uaDef)
	if err != nil {
		return false, false, err
	}
	for _, condition := range c.Cluster.Conditions {
		if condition.Type != models.ClusterBlueprintSync {
			continue
		}
		switch condition.Status {
		case models.Failed:
			return true, false, nil
		case models.Success:
			return false, true, nil
		}
	}
	return false, false, nil
}

// pollBlueprintSync waits for a previously triggered blueprint sync to reach
// a terminal (succeeded or failed) condition, or for ctx to time out. A
// blueprint resync on an already-ready cluster can finish in a few seconds,
// so this checks on a ~10-15s cadence rather than the 30s+ cadence used for
// full cluster provisioning — the overall 20-minute Create/Update timeout
// still bounds how long a genuinely slow sync gets to run.
func pollBlueprintSync(ctx context.Context, edgeID, projectID, clusterName string) error {
	// Allow the backend a brief moment to transition conditions before the
	// first check.
	select {
	case <-ctx.Done():
		return fmt.Errorf("context cancelled before blueprint sync polling started")
	case <-time.After(10 * time.Second):
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("blueprint sync timed out for cluster: %s", clusterName)
		case <-ticker.C:
			log.Printf("checking blueprint sync status for cluster: %s", clusterName)
			failed, succeeded, err := blueprintSyncConditionStatus(edgeID, projectID)
			if err != nil {
				log.Printf("error checking blueprint sync status for %s: %s — will retry", clusterName, err.Error())
				continue
			}
			if failed {
				return fmt.Errorf("blueprint sync failed for cluster: %s", clusterName)
			}
			if succeeded {
				log.Printf("blueprint sync completed successfully for cluster: %s", clusterName)
				return nil
			}
			log.Printf("blueprint sync still in progress for cluster: %s", clusterName)
		}
	}
}

func (r *BlueprintForceSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	var plan BlueprintForceSyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	forceSync, addons, diags := readWriteOnlySyncOpts(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan, diags = applyBlueprintSync(ctx, plan, forceSync, addons)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		// Not calling resp.State.Set leaves the resource absent from
		// state, so a retry cleanly re-attempts Create.
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read is intentionally a no-op: resp.State already defaults to the prior
// state. force_sync forces a diff via ModifyPlan instead of a Read-time side
// effect, so `terraform plan`/refresh never talks to the backend — the sync
// only runs inside Create/Update, which Terraform only calls after the user
// approves `terraform apply`.
func (r *BlueprintForceSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *BlueprintForceSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	var plan BlueprintForceSyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	forceSync, addons, diags := readWriteOnlySyncOpts(ctx, req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan, diags = applyBlueprintSync(ctx, plan, forceSync, addons)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		// resp.State was pre-populated by the framework with the prior
		// (last-known-good) state and is left untouched here, so a failed
		// apply never drifts state away from reality.
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// applyBlueprintSync triggers a sync, waits for completion, and returns the
// plan model updated with the observed blueprint name/version and id.
func applyBlueprintSync(ctx context.Context, plan BlueprintForceSyncModel, forceSync bool, addons []string) (BlueprintForceSyncModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	clusterName := plan.ClusterName.ValueString()
	projectName := plan.Project.ValueString()

	outcome, err := triggerBlueprintSync(clusterName, projectName, forceSync, plan.BlueprintName.ValueString(), plan.BlueprintVersion.ValueString(), addons)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to sync blueprint for cluster %q: %s", clusterName, err))
		return plan, diags
	}

	if err := pollBlueprintSync(ctx, outcome.edgeID, outcome.projectID, clusterName); err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Blueprint sync did not complete for cluster %q: %s", clusterName, err))
		return plan, diags
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", clusterName, projectName))
	plan.BlueprintName = types.StringValue(outcome.observedBlueprint)
	plan.BlueprintVersion = types.StringValue(outcome.observedVersion)
	return plan, diags
}

func (r *BlueprintForceSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Removing from Terraform state only; blueprint sync cannot be "undone".
	log.Println("blueprint_sync destroy: removing from Terraform state, no API call made")
}
