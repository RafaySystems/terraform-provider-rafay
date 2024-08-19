package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/RafaySystems/terraform-provider-rafay/fwrafay/fwmodels/mks_cluster"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &MksClusterResource{}
var _ resource.ResourceWithImportState = &MksClusterResource{}

func NewMksClusterResource() resource.Resource {
	return &MksClusterResource{}
}

// MksClusterResource defines the resource implementation.
type MksClusterResource struct {
	client *http.Client
}

func (r *MksClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster"
}

func (r *MksClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "api version",
				MarkdownDescription: "api version",
				Default:             stringdefault.StaticString("infra.k8smgmt.io/v3"),
			},
			"kind": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Description:         "kind",
				MarkdownDescription: "kind",
				Default:             stringdefault.StaticString("Cluster"),
			},
			"metadata": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"annotations": schema.MapAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
						Description:         "annotations of the resource",
						MarkdownDescription: "annotations of the resource",
					},
					"created_by": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "Id of the Person",
								MarkdownDescription: "Id of the Person",
							},
							"is_ssouser": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "Whether person is logged in using sso",
								MarkdownDescription: "Whether person is logged in using sso",
							},
							"username": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "Username fo the Person",
								MarkdownDescription: "Username fo the Person",
							},
						},
						Optional: true,
						Computed: true,
					},
					"description": schema.StringAttribute{
						Optional:            true,
						Description:         "description of the resource",
						MarkdownDescription: "description of the resource",
					},
					"display_name": schema.StringAttribute{
						Optional:            true,
						Description:         "Display Name of the resource",
						MarkdownDescription: "Display Name of the resource",
					},
					"labels": schema.MapAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
						Description:         "labels of the resource",
						MarkdownDescription: "labels of the resource",
					},
					"modified_by": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "Id of the Person",
								MarkdownDescription: "Id of the Person",
							},
							"is_ssouser": schema.BoolAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "Whether person is logged in using sso",
								MarkdownDescription: "Whether person is logged in using sso",
							},
							"username": schema.StringAttribute{
								Optional:            true,
								Computed:            true,
								Description:         "Username fo the Person",
								MarkdownDescription: "Username fo the Person",
							},
						},
						Optional: true,
						Computed: true,
					},
					"name": schema.StringAttribute{
						Required:            true,
						Description:         "name of the resource",
						MarkdownDescription: "name of the resource",
					},
					"project": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "Project of the resource",
						MarkdownDescription: "Project of the resource",
						Default:             stringdefault.StaticString("defaultproject"),
					},
				},
				Required:            true,
				Description:         "metadata of the resource",
				MarkdownDescription: "metadata of the resource",
			},
			"spec": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"blueprint": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Required: true,
							},
							"version": schema.StringAttribute{
								Optional: true,
								Computed: true,
							},
						},
						Required: true,
					},
					"cloud_credentials": schema.StringAttribute{
						Optional:            true,
						Description:         "The credentials to be used to ssh into the  Clusster Nodes",
						MarkdownDescription: "The credentials to be used to ssh into the  Clusster Nodes",
					},
					"config": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"auto_approve_nodes": schema.BoolAttribute{
								Optional:            true,
								Description:         "Auto approves incoming nodes by default",
								MarkdownDescription: "Auto approves incoming nodes by default",
							},
							"ssh": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"passphrase": schema.StringAttribute{
										Optional:            true,
										Description:         "Provide ssh passphrase",
										MarkdownDescription: "Provide ssh passphrase",
									},
									"port": schema.StringAttribute{
										Required:            true,
										Description:         "Provide ssh port",
										MarkdownDescription: "Provide ssh port",
									},
									"private_key_path": schema.StringAttribute{
										Required:            true,
										Description:         "Provide local path to the private key",
										MarkdownDescription: "Provide local path to the private key",
									},
									"username": schema.StringAttribute{
										Required:            true,
										Description:         "Provide the ssh username",
										MarkdownDescription: "Provide the ssh username",
									},
								},
								Optional:            true,
								Description:         "SSH config for all the nodes within the cluster",
								MarkdownDescription: "SSH config for all the nodes within the cluster",
							},
							"dedicated_control_plane": schema.BoolAttribute{
								Optional:            true,
								Description:         "Select this option for preventing scheduling of user workloads on Control Plane nodes",
								MarkdownDescription: "Select this option for preventing scheduling of user workloads on Control Plane nodes",
							},
							"high_availability": schema.BoolAttribute{
								Optional:            true,
								Description:         "Select this option for highly available control plane. Minimum three control plane nodes are required",
								MarkdownDescription: "Select this option for highly available control plane. Minimum three control plane nodes are required",
							},
							"kubernetes_upgrade": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"params": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"worker_concurrency": schema.StringAttribute{
												Required:            true,
												Description:         "It can be number or percentage",
												MarkdownDescription: "It can be number or percentage",
											},
										},
										Required: true,
									},
									"strategy": schema.StringAttribute{
										Required:            true,
										Description:         "Kubernetes upgrade strategy for worker nodes and Valid options are: concurrent/sequential",
										MarkdownDescription: "Kubernetes upgrade strategy for worker nodes and Valid options are: concurrent/sequential",
									},
								},
								Optional: true,
							},
							"kubernetes_version": schema.StringAttribute{
								Required:            true,
								Description:         "Kubernetes version of the Control Plane",
								MarkdownDescription: "Kubernetes version of the Control Plane",
							},
							"location": schema.StringAttribute{
								Optional:            true,
								Description:         "The data center location where the cluster nodes will be launched",
								MarkdownDescription: "The data center location where the cluster nodes will be launched",
							},
							"network": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"cni": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Required:            true,
												Description:         "Provide the CNI name, e.g., Calico or Cilium",
												MarkdownDescription: "Provide the CNI name, e.g., Calico or Cilium",
											},
											"version": schema.StringAttribute{
												Required:            true,
												Description:         "Provide the CNI version, e.g., 3.26.1",
												MarkdownDescription: "Provide the CNI version, e.g., 3.26.1",
											},
										},
										Required:            true,
										Description:         "MKS Cluster CNI Specification",
										MarkdownDescription: "MKS Cluster CNI Specification",
									},
									"ipv6": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"pod_subnet": schema.StringAttribute{
												Optional:            true,
												Computed:            true,
												Description:         "Kubernetes pod subnet",
												MarkdownDescription: "Kubernetes pod subnet",
											},
											"service_subnet": schema.StringAttribute{
												Optional:            true,
												Computed:            true,
												Description:         "Kubernetes service subnet",
												MarkdownDescription: "Kubernetes service subnet",
											},
										},
										Optional: true,
										Computed: true,
									},
									"pod_subnet": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Description:         "Kubernetes pod subnet",
										MarkdownDescription: "Kubernetes pod subnet",
									},
									"service_subnet": schema.StringAttribute{
										Optional:            true,
										Computed:            true,
										Description:         "Kubernetes service subnet",
										MarkdownDescription: "Kubernetes service subnet",
									},
								},
								Required:            true,
								Description:         "MKS Cluster Network Specification",
								MarkdownDescription: "MKS Cluster Network Specification",
							},
							"nodes": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"arch": schema.StringAttribute{
											Required:            true,
											Description:         "System Architecture of the node",
											MarkdownDescription: "System Architecture of the node",
										},
										"hostname": schema.StringAttribute{
											Required:            true,
											Description:         "Hostname of the node",
											MarkdownDescription: "Hostname of the node",
										},
										"interface": schema.StringAttribute{
											Optional:            true,
											Description:         "Interface to be used on the node",
											MarkdownDescription: "Interface to be used on the node",
										},
										"labels": schema.MapAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Computed:            true,
											Description:         "labels to be added to the node",
											MarkdownDescription: "labels to be added to the node",
										},
										"operating_system": schema.StringAttribute{
											Required:            true,
											Description:         "OS of the node",
											MarkdownDescription: "OS of the node",
										},
										"private_ip": schema.StringAttribute{
											Required:            true,
											Description:         "Private ip address of the node",
											MarkdownDescription: "Private ip address of the node",
										},
										"roles": schema.ListAttribute{
											ElementType:         types.StringType,
											Required:            true,
											Description:         "Valid roles are: 'ControlPlane', 'Worker', 'Storage'",
											MarkdownDescription: "Valid roles are: 'ControlPlane', 'Worker', 'Storage'",
										},
										"ssh": schema.SingleNestedAttribute{
											Attributes: map[string]schema.Attribute{
												"ip_address": schema.StringAttribute{
													Optional:            true,
													Description:         "Use this to override node level ssh details",
													MarkdownDescription: "Use this to override node level ssh details",
												},
												"passphrase": schema.StringAttribute{
													Optional:            true,
													Description:         "SSH Passphrase",
													MarkdownDescription: "SSH Passphrase",
												},
												"port": schema.StringAttribute{
													Required:            true,
													Description:         "SSH Port",
													MarkdownDescription: "SSH Port",
												},
												"private_key_path": schema.StringAttribute{
													Required:            true,
													Description:         "Specify Path to SSH private key",
													MarkdownDescription: "Specify Path to SSH private key",
												},
												"username": schema.StringAttribute{
													Required:            true,
													Description:         "SSH Username",
													MarkdownDescription: "SSH Username",
												},
											},
											Optional:            true,
											Description:         "MKS Node SSH definition",
											MarkdownDescription: "MKS Node SSH definition",
										},
										"taints": schema.ListNestedAttribute{
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"effect": schema.StringAttribute{
														Optional: true,
													},
													"key": schema.StringAttribute{
														Optional: true,
													},
													"value": schema.StringAttribute{
														Optional: true,
													},
												},
											},
											Optional:            true,
											Description:         "taints to be added to the node",
											MarkdownDescription: "taints to be added to the node",
										},
									},
								},
								Required:            true,
								Description:         "holds node configuration for the cluster",
								MarkdownDescription: "holds node configuration for the cluster",
							},
						},
						Required:            true,
						Description:         "MKS V3 cluster specification",
						MarkdownDescription: "MKS V3 cluster specification",
					},
					"proxy": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"allow_insecure_bootstrap": schema.BoolAttribute{
								Optional: true,
							},
							"bootstrap_ca": schema.StringAttribute{
								Optional: true,
							},
							"enabled": schema.BoolAttribute{
								Required: true,
							},
							"http_proxy": schema.StringAttribute{
								Optional: true,
							},
							"https_proxy": schema.StringAttribute{
								Optional: true,
							},
							"no_proxy": schema.StringAttribute{
								Optional: true,
							},
							"proxy_auth": schema.StringAttribute{
								Optional: true,
							},
						},
						Optional: true,
					},
					"sharing": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Required: true,
							},
							"projects": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Required: true,
										},
									},
								},
								Required: true,
							},
						},
						Optional: true,
					},
					"system_components_placement": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"daemon_set_override": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"daemon_set_tolerations": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"effect": schema.StringAttribute{
													Optional: true,
												},
												"key": schema.StringAttribute{
													Optional: true,
												},
												"operator": schema.StringAttribute{
													Optional: true,
												},
												"toleration_seconds": schema.Int64Attribute{
													Optional: true,
												},
												"value": schema.StringAttribute{
													Optional: true,
												},
											},
										},
										Optional: true,
									},
									"node_selection_enabled": schema.BoolAttribute{
										Optional: true,
									},
								},
								Optional: true,
							},
							"node_selector": schema.MapAttribute{
								ElementType: types.StringType,
								Optional:    true,
							},
							"tolerations": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"effect": schema.StringAttribute{
											Optional: true,
										},
										"key": schema.StringAttribute{
											Optional: true,
										},
										"operator": schema.StringAttribute{
											Optional: true,
										},
										"toleration_seconds": schema.Int64Attribute{
											Optional: true,
										},
										"value": schema.StringAttribute{
											Optional: true,
										},
									},
								},
								Optional: true,
							},
						},
						Optional: true,
					},
					"type": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Description:         "The type of the cluster this spec corresponds to",
						MarkdownDescription: "The type of the cluster this spec corresponds to",
						Default:             stringdefault.StaticString("mks"),
					},
				},
				Required:            true,
				Description:         "cluster specification",
				MarkdownDescription: "cluster specification",
			},
		},
	}
}

func (r *MksClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *MksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	var data mks_cluster.MksClusterModel
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}


	// Convert from Terraform data model into API data model
    // createReq := infrapb.Cluster{
    //     ApiVersion: data.APIVersion.ValueString(),
	// 	Kind:       data.Kind.ValueString(),
	// 	Metadata: 
    // }

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", err))
		return
	}

	// Convert the Terraform model to a Hub model

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", err))
		return
	}

	hubCluster := &infrapb.Cluster{}
	err = json.Unmarshal(jsonBytes, hubCluster)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create cluster, got error: %s", err))
		return
	}

	err = client.InfraV3().Cluster().Apply(ctx, hubCluster, options.ApplyOptions{})

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data mks_cluster.MksClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data mks_cluster.MksClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data mks_cluster.MksClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *MksClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
