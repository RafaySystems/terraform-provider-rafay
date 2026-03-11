package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
)

const mksClusterType = "mks"

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ClusterKubernetesVersionsDataSource{}

func NewClusterKubernetesVersionsDataSource() datasource.DataSource {
	return &ClusterKubernetesVersionsDataSource{}
}

// ClusterKubernetesVersionsDataSource defines the data source implementation.
type ClusterKubernetesVersionsDataSource struct {
	client typed.Client
}

// ClusterKubernetesVersionsModel describes the data source data model.
type ClusterKubernetesVersionsModel struct {
	Project                  types.String `tfsdk:"project"`
	ClusterName              types.String `tfsdk:"cluster_name"`
	ClusterType              types.String `tfsdk:"cluster_type"`
	DefaultVersion           types.String `tfsdk:"default_version"`
	LatestVersion            types.String `tfsdk:"latest_version"`
	IncludeDeprecatedVersion types.Bool   `tfsdk:"include_deprecated_version"`
	Versions                 types.List   `tfsdk:"versions"`
}

// kubernetesVersionsRequest is the request body for the Kubernetes versions API (rafay-common).
type kubernetesVersionsRequest struct {
	ClusterType               string `json:"cluster_type"`
	IncludeDeprecatedVersions bool   `json:"include_deprecated_versions"`
}

// kubernetesVersionsResponse is the API response body.
type kubernetesVersionsResponse struct {
	KubernetesVersions []string `json:"kubernetes_versions"`
	DefaultVersion     string   `json:"default_version"`
	LatestVersion      string   `json:"latest_version"`
}

func (d *ClusterKubernetesVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_kubernetes_versions"
}

func (d *ClusterKubernetesVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get the available Kubernetes versions for an MKS cluster from the Rafay API.",
		Attributes: map[string]schema.Attribute{
			"project": schema.StringAttribute{
				Required:    true,
				Description: "Project that contains the MKS cluster.",
			},
			"cluster_name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the MKS cluster to fetch Kubernetes versions for.",
			},
			"cluster_type": schema.StringAttribute{
				Computed:    true,
				Description: "The cluster type sent to the API; typically \"mks\".",
			},
			"default_version": schema.StringAttribute{
				Computed:    true,
				Description: "The default Kubernetes version for the cluster (from API).",
			},
			"latest_version": schema.StringAttribute{
				Computed:    true,
				Description: "The latest supported Kubernetes version (from API).",
			},
			"include_deprecated_version": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, deprecated Kubernetes versions are included. Defaults to false.",
			},
			"versions": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of available Kubernetes versions (from API).",
			},
		},
	}
}

func (d *ClusterKubernetesVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(typed.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected typed.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *ClusterKubernetesVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ClusterKubernetesVersionsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	includeDeprecated := false
	if !config.IncludeDeprecatedVersion.IsNull() {
		includeDeprecated = config.IncludeDeprecatedVersion.ValueBool()
	}

	clusterType := mksClusterType
	body := kubernetesVersionsRequest{
		ClusterType:               clusterType,
		IncludeDeprecatedVersions: includeDeprecated,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		resp.Diagnostics.AddError("Failed to build request", err.Error())
		return
	}

	extResp, err := d.client.InfraV3().Cluster().ExtApi().KubernetesVersions(ctx, options.ExtOptions{
		Project: config.Project.ValueString(),
		Name:    config.ClusterName.ValueString(),
		Body:    bodyBytes,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch Kubernetes versions", err.Error())
		return
	}

	if extResp.Status != 200 {
		resp.Diagnostics.AddError("API error", fmt.Sprintf("Kubernetes versions API returned status %d: %s", extResp.Status, string(extResp.Body)))
		return
	}

	var apiResp kubernetesVersionsResponse
	if err := json.Unmarshal(extResp.Body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Failed to parse API response", err.Error())
		return
	}

	versionsList, diags := types.ListValueFrom(ctx, types.StringType, apiResp.KubernetesVersions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := ClusterKubernetesVersionsModel{
		Project:                  config.Project,
		ClusterName:              config.ClusterName,
		ClusterType:              types.StringValue(clusterType),
		DefaultVersion:           types.StringValue(apiResp.DefaultVersion),
		LatestVersion:            types.StringValue(apiResp.LatestVersion),
		IncludeDeprecatedVersion: types.BoolValue(includeDeprecated),
		Versions:                 versionsList,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
