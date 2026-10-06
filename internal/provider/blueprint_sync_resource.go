package provider

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &BlueprintSyncResource{}
	_ resource.ResourceWithModifyPlan     = &BlueprintSyncResource{}
	_ resource.ResourceWithValidateConfig = &BlueprintSyncResource{}
)

// partialSuccessStatus is a terminal ClusterBlueprintSync condition status
// the backend can report (e.g. some but not all steps of the sync
// completed) that isn't among the constants the vendored rctl SDK's
// models.RafayConditionStatus enum defines. It's still just a string under
// the hood, so this compares fine against condition.Status.
const partialSuccessStatus = models.RafayConditionStatus("PartialSuccess")

func NewBlueprintSyncResource() resource.Resource {
	return &BlueprintSyncResource{}
}

// BlueprintSyncResource triggers a blueprint sync on a cluster, optionally
// assigning a blueprint name/version to the cluster first. It talks to the
// backend directly via rctl (like the legacy SDKv2 resource it replaces)
// rather than through the typed hub client, so it needs no provider-configured
// client.
type BlueprintSyncResource struct{}

type BlueprintSyncModel struct {
	ID               types.String `tfsdk:"id"`
	ClusterName      types.String `tfsdk:"cluster_name"`
	Project          types.String `tfsdk:"project"`
	BlueprintName    types.String `tfsdk:"blueprint_name"`
	BlueprintVersion types.String `tfsdk:"blueprint_version"`
	ForceSync        types.Bool   `tfsdk:"force_sync"`
	Addons           types.List   `tfsdk:"addons"`
	OptionalAddons   types.List   `tfsdk:"optional_addons"`
}

func (r *BlueprintSyncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blueprint_sync"
}

func (r *BlueprintSyncResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Description: "Subset of blueprint addons to sync. Only valid with force_sync=true. When unset, the full blueprint is synced. This value is never stored in state.",
			},
			"optional_addons": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Description: "Optional blueprint addons to deploy on this cluster when publishing. Addons marked optional on the blueprint are skipped unless listed here; [] deselects all of them. Leave unset to keep the cluster's current selection; removing it after it was set deselects all of them. Always reflects the selection on the cluster, so a plan shows any difference from it.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ValidateConfig rejects addons without force_sync=true, matching the rctl
// CLI rule that selective addon sync is only allowed with --force-sync.
func (r *BlueprintSyncResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var addons types.List
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("addons"), &addons)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if addons.IsNull() || addons.IsUnknown() || len(addons.Elements()) == 0 {
		return
	}

	var forceSync types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("force_sync"), &forceSync)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if forceSync.IsNull() || forceSync.IsUnknown() || !forceSync.ValueBool() {
		resp.Diagnostics.AddAttributeError(
			path.Root("addons"),
			"Invalid Configuration",
			"addons can only be used with force_sync=true",
		)
	}
}

// ModifyPlan forces a diff on id when force_sync=true, so Update re-syncs on
// every apply that asks for it. Otherwise the plan only changes when the
// blueprint or the optional add-on selection does, and an unchanged
// configuration plans no changes. Create/Update always recompute id
// deterministically from cluster_name and project, so a forced id resolves
// cleanly once the apply completes.
//
// This never talks to the backend itself — it only shapes the diff that the
// user reviews before approving `terraform apply`; the actual publish call
// only happens inside Create/Update.
func (r *BlueprintSyncResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy plans have a null plan/config; nothing to force.
	if req.Plan.Raw.IsNull() {
		return
	}

	if !req.State.Raw.IsNull() {
		var configured, planned types.List
		resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("optional_addons"), &configured)...)
		resp.Diagnostics.Append(resp.Plan.GetAttribute(ctx, path.Root("optional_addons"), &planned)...)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("optional_addons"), plannedOptionalAddons(ctx, configured, req.Private, planned))...)
		if mustRecordOptionalAddonsConfigured(ctx, configured, req.Private) {
			resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)
		}
	}

	var forceSync types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("force_sync"), &forceSync)...)
	if resp.Diagnostics.HasError() || !forceSync.ValueBool() {
		return
	}
	resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("id"), types.StringUnknown())...)
}

// optionalAddonsConfiguredKey is the private state key recording whether the
// last apply had optional_addons set in its config.
const optionalAddonsConfiguredKey = "optional_addons_configured"

type privateState interface {
	GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics)
}

// optionalAddonsWereConfigured reports whether the last apply set
// optional_addons in its config.
func optionalAddonsWereConfigured(ctx context.Context, private privateState) bool {
	if private == nil {
		return false
	}
	v, _ := private.GetKey(ctx, optionalAddonsConfiguredKey)
	return string(v) == "true"
}

// mustRecordOptionalAddonsConfigured reports whether optional_addons is set
// but the last apply did not record it: it was set with nothing else to change,
// so no apply ran, or before this was tracked. One update records it (without
// a sync, see blueprintSyncNeeded); otherwise removing it later would keep the
// selection instead of deselecting it.
func mustRecordOptionalAddonsConfigured(ctx context.Context, configured types.List, private privateState) bool {
	return !configured.IsNull() && !optionalAddonsWereConfigured(ctx, private)
}

// blueprintSyncNeeded reports whether an update has anything to sync. An update
// that only records the optional_addons flag does not.
func blueprintSyncNeeded(plan, state BlueprintSyncModel, forceSync bool) bool {
	return forceSync ||
		!plan.BlueprintName.Equal(state.BlueprintName) ||
		!plan.BlueprintVersion.Equal(state.BlueprintVersion) ||
		!plan.OptionalAddons.Equal(state.OptionalAddons)
}

// plannedOptionalAddons plans optional_addons. Unset in the config keeps the
// cluster's selection, unless the last apply set it: removing it from the
// config then deselects every optional add-on, as an explicit [] does.
func plannedOptionalAddons(ctx context.Context, configured types.List, private privateState, planned types.List) types.List {
	if configured.IsNull() && optionalAddonsWereConfigured(ctx, private) {
		return optionalAddonsState(nil)
	}
	return planned
}

// optionalAddonsState is the state value of a selection read from the cluster:
// none selected is [], so it compares equal to an explicit [] in config.
func optionalAddonsState(selection []string) types.List {
	elems := make([]attr.Value, 0, len(selection))
	for _, name := range selection {
		elems = append(elems, types.StringValue(name))
	}
	return types.ListValueMust(types.StringType, elems)
}

// appliedOptionalAddons is the optional_addons value stored after an apply: the
// planned one when known (the configured list, or the prior state when unset),
// otherwise the selection the cluster was left with.
func appliedOptionalAddons(planned types.List, observed []string) types.List {
	if planned.IsUnknown() {
		return optionalAddonsState(observed)
	}
	return planned
}

// readStringListFromConfig extracts a string list from config.
// Returns nil when the attribute is unset or unknown, and a non-nil empty
// slice for an explicit [] (callers rely on telling the two apart).
func readStringListFromConfig(ctx context.Context, config tfsdk.Config, attr string) ([]string, diag.Diagnostics) {
	var listAttr types.List
	diags := config.GetAttribute(ctx, path.Root(attr), &listAttr)
	if diags.HasError() {
		return nil, diags
	}
	if listAttr.IsNull() || listAttr.IsUnknown() {
		return nil, diags
	}

	var items []string
	diags.Append(listAttr.ElementsAs(ctx, &items, false)...)
	if diags.HasError() {
		return nil, diags
	}
	return items, diags
}

// blueprintSyncOutcome carries the edge/project IDs needed to poll for sync
// completion, plus the blueprint name/version actually observed on the
// cluster (as opposed to what was merely requested).
type blueprintSyncOutcome struct {
	edgeID                 string
	projectID              string
	observedBlueprint      string
	observedVersion        string
	observedOptionalAddons []string
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
// When addons is non-empty or optionalAddons is set (even to []), a selective
// sync is published via PublishBlueprintCluster (addons requires
// forceSync=true, enforced by ValidateConfig). Otherwise the full-blueprint
// PublishClusterBlueprint path is used. See blueprintSyncPublishSelection.
//
// The returned outcome's observedBlueprint/observedVersion always reflect
// what is actually assigned on the cluster: the requested values only if the
// update call succeeded, otherwise whatever was already there.
func triggerBlueprintSync(clusterName, projectName string, forceSync bool, blueprintName, blueprintVersion string, addons, optionalAddons []string) (*blueprintSyncOutcome, error) {
	log.Printf("blueprint sync starting for cluster: %s, project: %s, force_sync: %v, addons: %v, optional_addons: %v", clusterName, projectName, forceSync, addons, optionalAddons)

	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		return nil, fmt.Errorf("failed to get project %q: %w", projectName, err)
	}

	clusterResp, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster %q: %w", clusterName, err)
	}

	outcome := &blueprintSyncOutcome{
		edgeID:                 clusterResp.ID,
		projectID:              projectID,
		observedBlueprint:      clusterResp.ClusterBlueprint,
		observedVersion:        clusterResp.ClusterBlueprintVersion,
		observedOptionalAddons: clusterResp.OptionalAddons,
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
	// GetCluster read back the selection of the current blueprint; send the one
	// requested for the new blueprint so the update is valid against it. When
	// optional_addons is unset the current selection is kept as-is
	if blueprintChanged && optionalAddons != nil {
		clusterResp.OptionalAddons = optionalAddons
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

	if selective, selection := blueprintSyncPublishSelection(addons, optionalAddons, clusterResp.OptionalAddons); selective {
		if err := cluster.PublishBlueprintCluster(clusterName, projectID, outcome.observedBlueprint, outcome.observedVersion, forceSync, addons, selection); err != nil {
			return outcome, fmt.Errorf("failed to publish blueprint for cluster %q: %w", clusterName, err)
		}
		outcome.observedOptionalAddons = selection
	} else {
		if err := cluster.PublishClusterBlueprint(clusterName, projectID, forceSync); err != nil {
			return outcome, fmt.Errorf("failed to publish blueprint for cluster %q: %w", clusterName, err)
		}
	}
	log.Printf("blueprint publish triggered for cluster: %s", clusterName)

	return outcome, nil
}

// blueprintSyncPublishSelection picks the publish path and the optional add-on
// selection to send. PublishBlueprintCluster always replaces the selection, so
// when optional_addons is unset (nil) it is given the cluster's current one;
// an explicit [] (non-nil) deselects everything. With neither addons nor
// optional_addons set, the full publish is used, which keeps the selection.
func blueprintSyncPublishSelection(addons, optionalAddons, current []string) (selective bool, selection []string) {
	if optionalAddons == nil {
		return len(addons) > 0, current
	}
	return true, optionalAddons
}

// blueprintSyncResult is the terminal outcome of a ClusterBlueprintSync
// condition. blueprintSyncStillRunning means no terminal status has been
// observed yet and polling should continue.
type blueprintSyncResult int

const (
	blueprintSyncStillRunning blueprintSyncResult = iota
	blueprintSyncSucceeded
	blueprintSyncPartialSuccess
	blueprintSyncFailed
)

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
func blueprintSyncConditionStatus(edgeID, projectID string) (blueprintSyncResult, error) {
	c, err := cluster.GetClusterWithEdgeID(edgeID, projectID, uaDef)
	if err != nil {
		return blueprintSyncStillRunning, err
	}
	for _, condition := range c.Cluster.Conditions {
		if condition.Type != models.ClusterBlueprintSync {
			continue
		}
		switch condition.Status {
		case models.Failed:
			return blueprintSyncFailed, nil
		case models.Success:
			return blueprintSyncSucceeded, nil
		case partialSuccessStatus:
			return blueprintSyncPartialSuccess, nil
		}
	}
	return blueprintSyncStillRunning, nil
}

// pollBlueprintSync waits for a previously triggered blueprint sync to reach
// a terminal condition, or for ctx to time out. A blueprint resync on an
// already-ready cluster can finish in a few seconds, so this checks on a
// ~10-15s cadence rather than the 30s+ cadence used for full cluster
// provisioning — the overall 20-minute Create/Update timeout still bounds
// how long a genuinely slow sync gets to run.
//
// partialSuccess is true when the sync settled at PartialSuccess: treated as
// a completed (non-error) apply, but the caller should surface it as a
// warning rather than silently reporting full success.
func pollBlueprintSync(ctx context.Context, edgeID, projectID, clusterName string) (partialSuccess bool, err error) {
	// Allow the backend a brief moment to transition conditions before the
	// first check.
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("context cancelled before blueprint sync polling started")
	case <-time.After(10 * time.Second):
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false, fmt.Errorf("blueprint sync timed out for cluster: %s", clusterName)
		case <-ticker.C:
			log.Printf("checking blueprint sync status for cluster: %s", clusterName)
			result, err := blueprintSyncConditionStatus(edgeID, projectID)
			if err != nil {
				log.Printf("error checking blueprint sync status for %s: %s — will retry", clusterName, err.Error())
				continue
			}
			switch result {
			case blueprintSyncFailed:
				return false, fmt.Errorf("blueprint sync failed for cluster: %s", clusterName)
			case blueprintSyncSucceeded:
				log.Printf("blueprint sync completed successfully for cluster: %s", clusterName)
				return false, nil
			case blueprintSyncPartialSuccess:
				log.Printf("blueprint sync completed with partial success for cluster: %s", clusterName)
				return true, nil
			}
			log.Printf("blueprint sync still in progress for cluster: %s", clusterName)
		}
	}
}

func (r *BlueprintSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	var plan BlueprintSyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// force_sync and addons are write-only: they're never in state, so they
	// must be read from the raw config, not the plan/state model above. So is
	// optional_addons: the plan holds the prior state when it is unset, and
	// only the config tells unset (keep the selection) from a set list.
	var forceSync types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("force_sync"), &forceSync)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addons, diags := readStringListFromConfig(ctx, req.Config, "addons")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	optionalAddons, diags := readStringListFromConfig(ctx, req.Config, "optional_addons")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterName := plan.ClusterName.ValueString()
	projectName := plan.Project.ValueString()

	outcome, err := triggerBlueprintSync(clusterName, projectName, forceSync.ValueBool(), plan.BlueprintName.ValueString(), plan.BlueprintVersion.ValueString(), addons, optionalAddons)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to sync blueprint for cluster %q: %s", clusterName, err))
		// Not calling resp.State.Set leaves the resource absent from
		// state, so a retry cleanly re-attempts Create.
		return
	}

	partialSuccess, err := pollBlueprintSync(ctx, outcome.edgeID, outcome.projectID, clusterName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Blueprint sync did not complete for cluster %q: %s", clusterName, err))
		return
	}
	if partialSuccess {
		resp.Diagnostics.AddWarning("Blueprint Sync Partially Succeeded", fmt.Sprintf("Blueprint sync for cluster %q completed with PartialSuccess: some part of the sync did not complete cleanly. Check the cluster in the Rafay console for details.", clusterName))
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", clusterName, projectName))
	plan.BlueprintName = types.StringValue(outcome.observedBlueprint)
	plan.BlueprintVersion = types.StringValue(outcome.observedVersion)
	plan.OptionalAddons = appliedOptionalAddons(plan.OptionalAddons, outcome.observedOptionalAddons)
	// Write-only attributes must not be stored in state.
	plan.ForceSync = types.BoolNull()
	plan.Addons = types.ListNull(types.StringType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, optionalAddonsConfiguredKey, []byte(strconv.FormatBool(optionalAddons != nil)))...)
}

// Read refreshes optional_addons from the cluster so a plan shows how the
// configured selection differs from the one on the cluster. It only reads: the
// sync still only runs inside Create/Update, after the user approves `terraform
// apply`. The rest of the state is left as it was.
func (r *BlueprintSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BlueprintSyncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterName := state.ClusterName.ValueString()
	projectID, err := getProjectIDFromName(state.Project.ValueString())
	if err == nil {
		var c *models.ClusterDetails
		if c, err = cluster.GetCluster(clusterName, projectID, uaDef); err == nil {
			state.OptionalAddons = optionalAddonsState(c.OptionalAddons)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}
	// keep the prior state, as before Read refreshed anything
	resp.Diagnostics.AddWarning("Unable to refresh optional add-ons",
		fmt.Sprintf("Could not read the optional add-on selection of cluster %q, so the plan compares against the last applied one: %s", clusterName, err))
}

func (r *BlueprintSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Minute)
	defer cancel()

	var plan BlueprintSyncModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var forceSync types.Bool
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("force_sync"), &forceSync)...)
	if resp.Diagnostics.HasError() {
		return
	}

	addons, diags := readStringListFromConfig(ctx, req.Config, "addons")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	optionalAddons, diags := readStringListFromConfig(ctx, req.Config, "optional_addons")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	configured := optionalAddons != nil
	// removed from a config that set it: ModifyPlan planned [], deselect all
	if !configured && optionalAddonsWereConfigured(ctx, req.Private) {
		optionalAddons = []string{}
	}

	clusterName := plan.ClusterName.ValueString()
	projectName := plan.Project.ValueString()

	var state BlueprintSyncModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if !blueprintSyncNeeded(plan, state, forceSync.ValueBool()) {
		// only the optional_addons flag to record (see ModifyPlan): no sync
		plan.ID = types.StringValue(fmt.Sprintf("%s/%s", clusterName, projectName))
		plan.ForceSync = types.BoolNull()
		plan.Addons = types.ListNull(types.StringType)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.Append(resp.Private.SetKey(ctx, optionalAddonsConfiguredKey, []byte(strconv.FormatBool(configured)))...)
		return
	}

	outcome, err := triggerBlueprintSync(clusterName, projectName, forceSync.ValueBool(), plan.BlueprintName.ValueString(), plan.BlueprintVersion.ValueString(), addons, optionalAddons)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to sync blueprint for cluster %q: %s", clusterName, err))
		// resp.State was pre-populated by the framework with the prior
		// (last-known-good) state and is left untouched here, so a failed
		// apply never drifts state away from reality.
		return
	}

	partialSuccess, err := pollBlueprintSync(ctx, outcome.edgeID, outcome.projectID, clusterName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Blueprint sync did not complete for cluster %q: %s", clusterName, err))
		return
	}
	if partialSuccess {
		resp.Diagnostics.AddWarning("Blueprint Sync Partially Succeeded", fmt.Sprintf("Blueprint sync for cluster %q completed with PartialSuccess: some part of the sync did not complete cleanly. Check the cluster in the Rafay console for details.", clusterName))
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", clusterName, projectName))
	plan.BlueprintName = types.StringValue(outcome.observedBlueprint)
	plan.BlueprintVersion = types.StringValue(outcome.observedVersion)
	plan.OptionalAddons = appliedOptionalAddons(plan.OptionalAddons, outcome.observedOptionalAddons)
	// Write-only attributes must not be stored in state.
	plan.ForceSync = types.BoolNull()
	plan.Addons = types.ListNull(types.StringType)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	resp.Diagnostics.Append(resp.Private.SetKey(ctx, optionalAddonsConfiguredKey, []byte(strconv.FormatBool(configured)))...)
}

func (r *BlueprintSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Removing from Terraform state only; blueprint sync cannot be "undone".
	log.Println("blueprint_sync destroy: removing from Terraform state, no API call made")
}
