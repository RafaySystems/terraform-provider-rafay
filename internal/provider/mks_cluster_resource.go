package provider

import (
	"context"
	"fmt"
	"time"

	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/cluster"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	convertor "github.com/RafaySystems/terraform-provider-rafay/internal/convertor"
	fw "github.com/RafaySystems/terraform-provider-rafay/internal/gen/resource_mks_cluster"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &MksClusterResource{}
	_ resource.ResourceWithImportState = &MksClusterResource{}
)

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

	r.client = client
}

func (r *MksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Read Terraform plan data into the model
	var data fw.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub := &infrapb.Cluster{}
	daig := convertor.ConvertMksClusterToHub(ctx, &data, hub)
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
	daig = convertor.WaitForClusterOperation(ctx, r.client, hub, timeout, ticker)

	if daig.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", daig))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read Terraform prior state data into the model
	var plan fw.MksClusterModel

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read the cluster, got error: %s", err))
		return
	}

	// Convert the Hub model to a Terraform model and save it into the state
	daigs := convertor.ConvertMksClusterFromHub(ctx, c, &plan)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster, got error: %s", daigs))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)

}

func (r *MksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read Terraform plan data into the model
	var plan fw.MksClusterModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert the Terraform model to a Hub model
	hub := &infrapb.Cluster{}
	daigs := convertor.ConvertMksClusterToHub(ctx, &plan, hub)
	if daigs.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the cluster, got error: %s", daigs))
		return
	}

	// Call the Hub to Apply the cluster
	err := cluster.ApplyMksV3Cluster(ctx, r.client, hub)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update cluster, got error: %s", err))
		return
	}

	// Wait for the cluster operation to complete
	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
	timeout := time.After(time.Duration(90) * time.Minute)
	daigs = convertor.WaitForClusterOperation(ctx, r.client, hub, timeout, ticker)

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
}

func (r *MksClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Todo
	resource.ImportStatePassthroughID(ctx, path.Root("metadata").AtName("name"), req, resp)
}
