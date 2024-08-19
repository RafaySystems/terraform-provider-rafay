// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/RafaySystems/terraform-provider-rafay/fwrafay/fwmodels/mks_cluster"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MksClusterDataSource{}

func NewMksClusterDataSource() datasource.DataSource {
	return &MksClusterDataSource{}
}

// ExampleDataSource defines the data source implementation.
type MksClusterDataSource struct {
	client *http.Client
}

func (d *MksClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster"
}

func (d *MksClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_version": schema.StringAttribute{
				Computed:            true,
				Description:         "api version",
				MarkdownDescription: "api version",
			},
			"kind": schema.StringAttribute{
				Computed:            true,
				Description:         "kind",
				MarkdownDescription: "kind",
			},
			"metadata": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"annotations": schema.MapAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						Description:         "annotations of the resource",
						MarkdownDescription: "annotations of the resource",
					},
					"created_by": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:            true,
								Description:         "Id of the Person",
								MarkdownDescription: "Id of the Person",
							},
							"is_ssouser": schema.BoolAttribute{
								Computed:            true,
								Description:         "Whether person is logged in using sso",
								MarkdownDescription: "Whether person is logged in using sso",
							},
							"username": schema.StringAttribute{
								Computed:            true,
								Description:         "Username fo the Person",
								MarkdownDescription: "Username fo the Person",
							},
						},
						Computed: true,
					},
					"description": schema.StringAttribute{
						Computed:            true,
						Description:         "description of the resource",
						MarkdownDescription: "description of the resource",
					},
					"display_name": schema.StringAttribute{
						Computed:            true,
						Description:         "Display Name of the resource",
						MarkdownDescription: "Display Name of the resource",
					},
					"labels": schema.MapAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						Description:         "labels of the resource",
						MarkdownDescription: "labels of the resource",
					},
					"modified_by": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								Computed:            true,
								Description:         "Id of the Person",
								MarkdownDescription: "Id of the Person",
							},
							"is_ssouser": schema.BoolAttribute{
								Computed:            true,
								Description:         "Whether person is logged in using sso",
								MarkdownDescription: "Whether person is logged in using sso",
							},
							"username": schema.StringAttribute{
								Computed:            true,
								Description:         "Username fo the Person",
								MarkdownDescription: "Username fo the Person",
							},
						},
						Computed: true,
					},
					"name": schema.StringAttribute{
						Computed:            true,
						Description:         "name of the resource",
						MarkdownDescription: "name of the resource",
					},
					"project": schema.StringAttribute{
						Computed:            true,
						Description:         "Project of the resource",
						MarkdownDescription: "Project of the resource",
					},
				},
				Computed:            true,
				Description:         "metadata of the resource",
				MarkdownDescription: "metadata of the resource",
			},
			"spec": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"blueprint": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed: true,
							},
							"version": schema.StringAttribute{
								Computed: true,
							},
						},
						Computed: true,
					},
					"cloud_credentials": schema.StringAttribute{
						Computed:            true,
						Description:         "The credentials to be used to ssh into the  Clusster Nodes",
						MarkdownDescription: "The credentials to be used to ssh into the  Clusster Nodes",
					},
					"config": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"auto_approve_nodes": schema.BoolAttribute{
								Computed:            true,
								Description:         "Auto approves incoming nodes by default",
								MarkdownDescription: "Auto approves incoming nodes by default",
							},
							"cluster_ssh": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"passphrase": schema.StringAttribute{
										Computed:            true,
										Description:         "Provide ssh passphrase",
										MarkdownDescription: "Provide ssh passphrase",
									},
									"port": schema.StringAttribute{
										Computed:            true,
										Description:         "Provide ssh port",
										MarkdownDescription: "Provide ssh port",
									},
									"private_key_path": schema.StringAttribute{
										Computed:            true,
										Description:         "Provide local path to the private key",
										MarkdownDescription: "Provide local path to the private key",
									},
									"username": schema.StringAttribute{
										Computed:            true,
										Description:         "Provide the ssh username",
										MarkdownDescription: "Provide the ssh username",
									},
								},
								Computed:            true,
								Description:         "SSH config for all the nodes within the cluster",
								MarkdownDescription: "SSH config for all the nodes within the cluster",
							},
							"dedicated_control_plane": schema.BoolAttribute{
								Computed:            true,
								Description:         "Select this option for preventing scheduling of user workloads on Control Plane nodes",
								MarkdownDescription: "Select this option for preventing scheduling of user workloads on Control Plane nodes",
							},
							"dedicated_masters_enabled": schema.BoolAttribute{
								Computed:            true,
								Description:         "This is deprecated in favour of dedicatedControlPlane",
								MarkdownDescription: "This is deprecated in favour of dedicatedControlPlane",
								DeprecationMessage:  "This attribute is deprecated.",
							},
							"high_availability": schema.BoolAttribute{
								Computed:            true,
								Description:         "Select this option for highly available control plane. Minimum three control plane nodes are required",
								MarkdownDescription: "Select this option for highly available control plane. Minimum three control plane nodes are required",
							},
							"kubernetes_upgrade": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"params": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"worker_concurrency": schema.StringAttribute{
												Computed:            true,
												Description:         "It can be number or percentage",
												MarkdownDescription: "It can be number or percentage",
											},
										},
										Computed: true,
									},
									"strategy": schema.StringAttribute{
										Computed:            true,
										Description:         "Kubernetes upgrade strategy for worker nodes and Valid options are: concurrent/sequential",
										MarkdownDescription: "Kubernetes upgrade strategy for worker nodes and Valid options are: concurrent/sequential",
									},
								},
								Computed: true,
							},
							"kubernetes_version": schema.StringAttribute{
								Computed:            true,
								Description:         "Kubernetes version of the Control Plane",
								MarkdownDescription: "Kubernetes version of the Control Plane",
							},
							"location": schema.StringAttribute{
								Computed:            true,
								Description:         "The data center location where the cluster nodes will be launched",
								MarkdownDescription: "The data center location where the cluster nodes will be launched",
							},
							"network": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"cni": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Computed:            true,
												Description:         "Provide the CNI name, e.g., Calico or Cilium",
												MarkdownDescription: "Provide the CNI name, e.g., Calico or Cilium",
											},
											"version": schema.StringAttribute{
												Computed:            true,
												Description:         "Provide the CNI version, e.g., 3.26.1",
												MarkdownDescription: "Provide the CNI version, e.g., 3.26.1",
											},
										},
										Computed:            true,
										Description:         "MKS Cluster CNI Specification",
										MarkdownDescription: "MKS Cluster CNI Specification",
									},
									"ipv6": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"pod_subnet": schema.StringAttribute{
												Computed:            true,
												Description:         "Kubernetes pod subnet",
												MarkdownDescription: "Kubernetes pod subnet",
											},
											"service_subnet": schema.StringAttribute{
												Computed:            true,
												Description:         "Kubernetes service subnet",
												MarkdownDescription: "Kubernetes service subnet",
											},
										},
										Computed: true,
									},
									"pod_subnet": schema.StringAttribute{
										Computed:            true,
										Description:         "Kubernetes pod subnet",
										MarkdownDescription: "Kubernetes pod subnet",
									},
									"service_subnet": schema.StringAttribute{
										Computed:            true,
										Description:         "Kubernetes service subnet",
										MarkdownDescription: "Kubernetes service subnet",
									},
								},
								Computed:            true,
								Description:         "MKS Cluster Network Specification",
								MarkdownDescription: "MKS Cluster Network Specification",
							},
							"nodes": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"arch": schema.StringAttribute{
											Computed:            true,
											Description:         "System Architecture of the node",
											MarkdownDescription: "System Architecture of the node",
										},
										"hostname": schema.StringAttribute{
											Computed:            true,
											Description:         "Hostname of the node",
											MarkdownDescription: "Hostname of the node",
										},
										"interface": schema.StringAttribute{
											Computed:            true,
											Description:         "Interface to be used on the node",
											MarkdownDescription: "Interface to be used on the node",
										},
										"labels": schema.MapAttribute{
											ElementType:         types.StringType,
											Computed:            true,
											Description:         "labels to be added to the node",
											MarkdownDescription: "labels to be added to the node",
										},
										"operating_system": schema.StringAttribute{
											Computed:            true,
											Description:         "OS of the node",
											MarkdownDescription: "OS of the node",
										},
										"private_ip": schema.StringAttribute{
											Computed:            true,
											Description:         "Private ip address of the node",
											MarkdownDescription: "Private ip address of the node",
										},
										"roles": schema.ListAttribute{
											ElementType:         types.StringType,
											Computed:            true,
											Description:         "Valid roles are: 'Master/ControlPlane', 'Worker', 'Storage'",
											MarkdownDescription: "Valid roles are: 'Master/ControlPlane', 'Worker', 'Storage'",
										},
										"ssh": schema.SingleNestedAttribute{
											Attributes: map[string]schema.Attribute{
												"ip_address": schema.StringAttribute{
													Computed:            true,
													Description:         "Use this to override node level ssh details",
													MarkdownDescription: "Use this to override node level ssh details",
												},
												"passphrase": schema.StringAttribute{
													Computed:            true,
													Description:         "SSH Passphrase",
													MarkdownDescription: "SSH Passphrase",
												},
												"port": schema.StringAttribute{
													Computed:            true,
													Description:         "SSH Port",
													MarkdownDescription: "SSH Port",
												},
												"private_key_path": schema.StringAttribute{
													Computed:            true,
													Description:         "Specify Path to SSH private key",
													MarkdownDescription: "Specify Path to SSH private key",
												},
												"username": schema.StringAttribute{
													Computed:            true,
													Description:         "SSH Username",
													MarkdownDescription: "SSH Username",
												},
											},
											Computed:            true,
											Description:         "MKS Node SSH definition",
											MarkdownDescription: "MKS Node SSH definition",
										},
										"taints": schema.ListNestedAttribute{
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"effect": schema.StringAttribute{
														Computed: true,
													},
													"key": schema.StringAttribute{
														Computed: true,
													},
													"value": schema.StringAttribute{
														Computed: true,
													},
												},
											},
											Computed:            true,
											Description:         "taints to be added to the node",
											MarkdownDescription: "taints to be added to the node",
										},
									},
								},
								Computed:            true,
								Description:         "holds node configuration for the cluster",
								MarkdownDescription: "holds node configuration for the cluster",
							},
						},
						Computed:            true,
						Description:         "MKS V3 cluster specification",
						MarkdownDescription: "MKS V3 cluster specification",
					},
					"proxy": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"allow_insecure_bootstrap": schema.BoolAttribute{
								Computed: true,
							},
							"bootstrap_ca": schema.StringAttribute{
								Computed: true,
							},
							"enabled": schema.BoolAttribute{
								Computed: true,
							},
							"http_proxy": schema.StringAttribute{
								Computed: true,
							},
							"https_proxy": schema.StringAttribute{
								Computed: true,
							},
							"no_proxy": schema.StringAttribute{
								Computed: true,
							},
							"proxy_auth": schema.StringAttribute{
								Computed: true,
							},
						},
						Computed: true,
					},
					"sharing": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Computed: true,
							},
							"projects": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								Computed: true,
							},
						},
						Computed: true,
					},
					"system_components_placement": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"daemon_set_override": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"daemon_set_tolerations": schema.ListNestedAttribute{
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"effect": schema.StringAttribute{
													Computed: true,
												},
												"key": schema.StringAttribute{
													Computed: true,
												},
												"operator": schema.StringAttribute{
													Computed: true,
												},
												"toleration_seconds": schema.Int64Attribute{
													Computed: true,
												},
												"value": schema.StringAttribute{
													Computed: true,
												},
											},
										},
										Computed: true,
									},
									"node_selection_enabled": schema.BoolAttribute{
										Computed: true,
									},
								},
								Computed: true,
							},
							"node_selector": schema.MapAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"tolerations": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"effect": schema.StringAttribute{
											Computed: true,
										},
										"key": schema.StringAttribute{
											Computed: true,
										},
										"operator": schema.StringAttribute{
											Computed: true,
										},
										"toleration_seconds": schema.Int64Attribute{
											Computed: true,
										},
										"value": schema.StringAttribute{
											Computed: true,
										},
									},
								},
								Computed: true,
							},
						},
						Computed: true,
					},
					"type": schema.StringAttribute{
						Computed:            true,
						Description:         "The type of the cluster this spec corresponds to",
						MarkdownDescription: "The type of the cluster this spec corresponds to",
					},
				},
				Computed:            true,
				Description:         "cluster specification",
				MarkdownDescription: "cluster specification",
			},
		},
	}
}

func (d *MksClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *MksClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data mks_cluster.MksClusterModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
