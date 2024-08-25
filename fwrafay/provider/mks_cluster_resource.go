package provider

import (
	"context"
	"fmt"

	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/cluster"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/RafaySystems/terraform-provider-rafay/fwrafay/types/mks_cluster"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MksClusterResource{}

func NewMksClusterResource() resource.Resource {
	return &MksClusterResource{}
}

// MksClusterResource defines the resource implementation.
type MksClusterResource struct {
	client typed.Client
}

func (r *MksClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster"
}

func (r *MksClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = mks_cluster.MksClusterResourceSchema(ctx)
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

	r.client = client
}

func (r *MksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data mks_cluster.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub := &infrapb.Cluster{}
	daig := mks_cluster.ConvertMksClusterToHub(ctx, &data, hub)
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
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var plan mks_cluster.MksClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read the cluster from the Hub
	c, err := r.client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    plan.Metadata.Name.ValueString(),
		Project: plan.Metadata.Project.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster, got error: %s", err))
		return
	}

	// Convert the Hub model to a Terraform model and save it into the state
	daigs := mks_cluster.ConvertMksClusterFromHub(ctx, c, &plan)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster, got error: %s", daigs))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var plan mks_cluster.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub := &infrapb.Cluster{}
	daigs := mks_cluster.ConvertMksClusterToHub(ctx, &plan, hub)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the cluster, got error: %s", daigs))
		return
	}

	// Call the Hub to update the cluster
	err := cluster.ApplyMksV3Cluster(ctx, r.client, hub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Read Terraform prior state data into the model
	var data mks_cluster.MksClusterModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	err := r.client.InfraV3().Cluster().Delete(ctx, options.DeleteOptions{
		Name:    data.Metadata.Name.ValueString(),
		Project: data.Metadata.Project.ValueString(),
	})

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete cluster, got error: %s", err))
		return
	}
}
