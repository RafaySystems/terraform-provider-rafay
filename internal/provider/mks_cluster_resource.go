package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rctl/pkg/cluster"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	fw "github.com/RafaySystems/terraform-provider-rafay/internal/resource_mks_cluster"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                     = &MksClusterResource{}
	_ resource.ResourceWithConfigure        = &MksClusterResource{}
	_ resource.ResourceWithImportState      = &MksClusterResource{}
	_ resource.ResourceWithConfigValidators = &MksClusterResource{}
)

func NewMksClusterResource() resource.Resource {
	return &MksClusterResource{}
}

// MksClusterResource defines the resource implemSharentation.
type MksClusterResource struct {
	client typed.Client
}

func (r *MksClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster"
}

func (r *MksClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fw.MksClusterResourceSchema(ctx)
}

func (r *MksClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(typed.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *typed.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Save the client for use in CRUD operations
	r.client = client
}

func (r *MksClusterResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("spec").AtName("config").AtName("cluster_ssh"),
			path.MatchRoot("spec").AtName("cloud_credentials"),
		),
	}
}

func (r *MksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data fw.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub, daig := fw.ConvertMksClusterToHub(ctx, data)
	if daig.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", daig))
		return
	}

	// Create the cluster
	err := cluster.ApplyMksV3Cluster(ctx, r.client, hub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", err))
		return
	}

	// Wait for the cluster operation to complete
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(90) * time.Minute)
	daig = fw.WaitForClusterApplyOperation(ctx, r.client, hub, timeout, ticker)

	if daig.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", daig))
		return
	}

	// edge object is also created. Update cse.
	if hub.GetSpec().Sharing != nil {
		// explicitly set cse to false.

		clusterName := hub.Metadata.Name
		// get pid from name
		projectName := hub.GetMetadata().Project
		pid, err := getProjectIDFromName(projectName)
		if err != nil {
			tflog.Error(ctx, "failed to get project id", map[string]any{"clusterName": clusterName, "projectName": projectName})
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the project id, got error: %s", err))
			return
		}
		existingEdge, err := cluster.GetCluster(clusterName, pid, uaDef)
		if err != nil {
			tflog.Error(ctx, "failed to get v1 mks cluster", map[string]any{"clusterName": clusterName, "projectID": pid})
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the edge object, got error: %s", err))
			return
		}

		existingEdge.Settings[clusterSharingExtKey] = "false"
		err = cluster.UpdateCluster(existingEdge, uaDef)
		if err != nil {
			tflog.Error(ctx, "failed to update v1 mks cluster", map[string]any{"edgeObj": existingEdge})
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the edge object, got error: %s", err))
			return
		}

		// TODO(Akshay): change to Info log
		tflog.Error(ctx, "cluster is updated successfully")
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var state fw.MksClusterModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read the cluster from the Hub
	c, err := r.client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    state.Metadata.Name.ValueString(),
		Project: state.Metadata.Project.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the cluster, got error: %s", err))
		return
	}

	// get v1 settings for sharing
	clusterName := c.Metadata.Name
	projectName := c.GetMetadata().Project
	pid, err := getProjectIDFromName(projectName)
	if err != nil {
		tflog.Error(ctx, "failed to get project id", map[string]any{"clusterName": clusterName, "projectName": projectName})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the project id, got error: %s", err))
		return
	}
	edge, err := cluster.GetCluster(clusterName, pid, uaDef)
	if err != nil {
		tflog.Error(ctx, "failed to get v1 mks cluster", map[string]any{"clusterName": clusterName, "projectID": pid})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the edge object, got error: %s", err))
		return
	}

	cse := edge.Settings[clusterSharingExtKey]
	// TODO(Akshay): convert to Info later
	tflog.Error(ctx, "Got edge obj from backend", map[string]any{clusterSharingExtKey: cse})

	if cse == "true" {
		c.Spec.Sharing = nil
	}

	// Convert the Hub model to a Terraform model
	daigs := fw.ConvertMksClusterFromHub(ctx, c, &state)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster, got error: %s", daigs))
		return
	}
	// Save the refreshed state into Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

}

func (r *MksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var plan fw.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub, daigs := fw.ConvertMksClusterToHub(ctx, plan)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the cluster, got error: %s", daigs))
		return
	}

	// Get the cluster if present
	clusterName := hub.Metadata.Name
	// get pid from name
	projectName := hub.GetMetadata().Project
	pid, err := getProjectIDFromName(projectName)
	if err != nil {
		tflog.Error(ctx, "failed to get project id", map[string]any{"clusterName": clusterName, "projectName": projectName})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the project id, got error: %s", err))
		return
	}
	existingEdge, err := cluster.GetCluster(clusterName, pid, uaDef)
	if err != nil {
		tflog.Error(ctx, "failed to get v1 mks cluster", map[string]any{"clusterName": clusterName, "projectID": pid})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the edge object, got error: %s", err))
		return
	}
	cse := existingEdge.Settings[clusterSharingExtKey]

	if hub.Spec.Sharing != nil && cse == "true" {
		resp.Diagnostics.AddError("Client Error", "Cluster sharing is currently managed through the external 'rafay_cluster_sharing' resource. To prevent configuration conflicts, please remove the sharing settings from the 'rafay_mks_cluster' resource and manage sharing exclusively via the external resource.")
		return
	}

	// Call the Hub to Apply the cluster
	err = cluster.ApplyMksV3Cluster(ctx, r.client, hub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", err))
		return
	}

	existingEdge, err = cluster.GetCluster(clusterName, pid, uaDef)
	if err != nil {
		tflog.Error(ctx, "failed to get v1 mks cluster", map[string]any{"clusterName": clusterName, "projectID": pid})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the edge object, got error: %s", err))
		return
	}
	cse = existingEdge.Settings[clusterSharingExtKey]

	// sharing is removed. Unset cse flag.
	if hub.Spec.Sharing == nil && cse == "false" {
		existingEdge.Settings[clusterSharingExtKey] = ""
		err = cluster.UpdateCluster(existingEdge, uaDef)
		if err != nil {
			tflog.Error(ctx, "failed to update v1 mks cluster", map[string]any{"edgeObj": existingEdge})
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the edge object, got error: %s", err))
			return
		}

		tflog.Debug(ctx, "cse flag unset")
	}

	// sharing is present. Set cse flag to false.
	if hub.Spec.Sharing != nil && cse != "false" {
		existingEdge.Settings[clusterSharingExtKey] = "false"
		err = cluster.UpdateCluster(existingEdge, uaDef)
		if err != nil {
			tflog.Error(ctx, "failed to update v1 mks cluster", map[string]any{"edgeObj": existingEdge})
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the edge object, got error: %s", err))
			return
		}

		tflog.Debug(ctx, "cse flag set to false")
	}

	// Wait for the cluster operation to complete
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(90) * time.Minute)
	daigs = fw.WaitForClusterApplyOperation(ctx, r.client, hub, timeout, ticker)

	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", daigs))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data fw.MksClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.InfraV3().Cluster().Delete(ctx, options.DeleteOptions{
		Name:    data.Metadata.Name.ValueString(),
		Project: data.Metadata.Project.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete cluster, got error: %s", err))
		return
	}

	// Wait for the cluster deletion to be completed
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(30) * time.Minute)
	daigs := fw.WaitForClusterDeleteOperation(ctx, r.client, data.Metadata.Name.ValueString(), data.Metadata.Project.ValueString(), timeout, ticker)

	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete the cluster, got error: %s", daigs))
		return
	}
}

func (r *MksClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name/project. Got: %q", req.ID),
		)
		return
	}

	name := idParts[0]
	project := idParts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("metadata").AtName("name"), name)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("metadata").AtName("project"), project)...)
}
