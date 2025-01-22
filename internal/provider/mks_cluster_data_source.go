package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	fw "github.com/RafaySystems/terraform-provider-rafay/internal/resource_mks_cluster"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MksClusterDataSource{}

func NewMksClusterDataSource() datasource.DataSource {
	return &MksClusterDataSource{}
}

func MksClusterDataSourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_version": schema.StringAttribute{
				Computed:    true,
				Description: "api version",
			},
			"kind": schema.StringAttribute{
				Computed:    true,
				Description: "kind",
			},
			"metadata": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"annotations": schema.MapAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "annotations of the resource",
					},
					"description": schema.StringAttribute{
						Computed:    true,
						Description: "description of the resource",
					},
					"labels": schema.MapAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "labels of the resource",
					},
					"name": schema.StringAttribute{
						Required:    true,
						Description: "name of the resource",
					},
					"project": schema.StringAttribute{
						Required:    true,
						Description: "Project of the resource",
					},
				},
				CustomType: fw.MetadataType{
					ObjectType: types.ObjectType{
						AttrTypes: fw.MetadataValue{}.AttributeTypes(ctx),
					},
				},
				Required:    true,
				Description: "metadata of the resource",
			},
			"spec": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"blueprint": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed: true,
							},
							"version": schema.StringAttribute{
								Computed:    true,
								Description: "Version of the blueprint",
							},
						},
						CustomType: fw.BlueprintType{
							ObjectType: types.ObjectType{
								AttrTypes: fw.BlueprintValue{}.AttributeTypes(ctx),
							},
						},
						Computed: true,
					},
					"cloud_credentials": schema.StringAttribute{
						Computed:    true,
						Description: "The credentials to be used to ssh into the  Clusster Nodes",
					},
					"config": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"auto_approve_nodes": schema.BoolAttribute{
								Computed:    true,
								Description: "Auto approves incoming nodes by default",
							},
							"cluster_ssh": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"passphrase": schema.StringAttribute{
										Computed:    true,
										Description: "Provide ssh passphrase",
									},
									"port": schema.StringAttribute{
										Computed:    true,
										Description: "Provide ssh port",
									},
									"private_key_path": schema.StringAttribute{
										Computed:    true,
										Description: "Provide local path to the private key",
									},
									"username": schema.StringAttribute{
										Computed:    true,
										Description: "Provide the ssh username",
									},
								},
								CustomType: fw.ClusterSshType{
									ObjectType: types.ObjectType{
										AttrTypes: fw.ClusterSshValue{}.AttributeTypes(ctx),
									},
								},
								Computed:    true,
								Description: "SSH config for all the nodes within the cluster",
							},
							"dedicated_control_plane": schema.BoolAttribute{
								Computed:    true,
								Description: "Select this option for preventing scheduling of user workloads on Control Plane nodes",
							},
							"high_availability": schema.BoolAttribute{
								Computed:    true,
								Description: "Select this option for highly available control plane. Minimum three control plane nodes are required",
							},
							"kubernetes_upgrade": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"params": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"worker_concurrency": schema.StringAttribute{
												Computed:    true,
												Description: "It can be number or percentage",
											},
										},
										CustomType: fw.ParamsType{
											ObjectType: types.ObjectType{
												AttrTypes: fw.ParamsValue{}.AttributeTypes(ctx),
											},
										},
										Computed: true,
									},
									"strategy": schema.StringAttribute{
										Computed:    true,
										Description: "Kubernetes upgrade strategy for worker nodes and Valid options are: concurrent/sequential",
									},
								},
								CustomType: fw.KubernetesUpgradeType{
									ObjectType: types.ObjectType{
										AttrTypes: fw.KubernetesUpgradeValue{}.AttributeTypes(ctx),
									},
								},
								Computed: true,
							},
							"kubernetes_version": schema.StringAttribute{
								Computed:    true,
								Description: "Kubernetes version of the Control Plane",
							},
							"installer_ttl": schema.Int64Attribute{
								Computed:    true,
								Description: "Installer TTL Configuration",
							},
							"kubelet_extra_args": schema.MapAttribute{
								ElementType: types.StringType,
								Computed:    true,
								Description: "Cluster kubelet extra args",
							},
							"location": schema.StringAttribute{
								Computed:    true,
								Description: "The data center location where the cluster nodes will be launched",
							},
							"network": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"cni": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Computed:    true,
												Description: "Provide the CNI name, e.g., Calico or Cilium",
											},
											"version": schema.StringAttribute{
												Computed:    true,
												Description: "Provide the CNI version, e.g., 3.26.1",
											},
										},
										CustomType: fw.CniType{
											ObjectType: types.ObjectType{
												AttrTypes: fw.CniValue{}.AttributeTypes(ctx),
											},
										},
										Computed:    true,
										Description: "MKS Cluster CNI Specification",
									},
									"ipv6": schema.SingleNestedAttribute{
										Attributes: map[string]schema.Attribute{
											"pod_subnet": schema.StringAttribute{
												Computed:    true,
												Description: "Kubernetes pod subnet",
											},
											"service_subnet": schema.StringAttribute{
												Computed:    true,
												Description: "Kubernetes service subnet",
											},
										},
										CustomType: fw.Ipv6Type{
											ObjectType: types.ObjectType{
												AttrTypes: fw.Ipv6Value{}.AttributeTypes(ctx),
											},
										},
										Computed: true,
									},
									"pod_subnet": schema.StringAttribute{
										Computed:    true,
										Description: "Kubernetes pod subnet",
									},
									"service_subnet": schema.StringAttribute{
										Computed:    true,
										Description: "Kubernetes service subnet",
									},
								},
								CustomType: fw.NetworkType{
									ObjectType: types.ObjectType{
										AttrTypes: fw.NetworkValue{}.AttributeTypes(ctx),
									},
								},
								Computed:    true,
								Description: "MKS Cluster Network Specification",
							},
							"nodes": schema.MapNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"arch": schema.StringAttribute{
											Computed:    true,
											Description: "System Architecture of the node",
										},
										"hostname": schema.StringAttribute{
											Computed:    true,
											Description: "Hostname of the node",
										},
										"interface": schema.StringAttribute{
											Computed:    true,
											Description: "Interface to be used on the node",
										},
										"labels": schema.MapAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "labels to be added to the node",
										},
										"kubelet_extra_args": schema.MapAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "Node kubelet extra args",
										},
										"operating_system": schema.StringAttribute{
											Computed:    true,
											Description: "OS of the node",
										},
										"private_ip": schema.StringAttribute{
											Computed:    true,
											Description: "Private ip address of the node",
										},
										"roles": schema.SetAttribute{
											ElementType: types.StringType,
											Computed:    true,
											Description: "Valid roles are: 'ControlPlane', 'Worker', 'Storage'",
										},
										"ssh": schema.SingleNestedAttribute{
											Attributes: map[string]schema.Attribute{
												"ip_address": schema.StringAttribute{
													Computed:    true,
													Description: "Use this to override node level ssh details",
												},
												"passphrase": schema.StringAttribute{
													Computed:    true,
													Description: "SSH Passphrase",
												},
												"port": schema.StringAttribute{
													Computed:    true,
													Description: "SSH Port",
												},
												"private_key_path": schema.StringAttribute{
													Computed:    true,
													Description: "Specify Path to SSH private key",
												},
												"username": schema.StringAttribute{
													Computed:    true,
													Description: "SSH Username",
												},
											},
											CustomType: fw.SshType{
												ObjectType: types.ObjectType{
													AttrTypes: fw.SshValue{}.AttributeTypes(ctx),
												},
											},
											Computed:    true,
											Description: "MKS Node SSH definition",
										},
										"taints": schema.SetNestedAttribute{
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
												CustomType: fw.TaintsType{
													ObjectType: types.ObjectType{
														AttrTypes: fw.TaintsValue{}.AttributeTypes(ctx),
													},
												},
											},
											Computed:    true,
											Description: "taints to be added to the node",
										},
									},
									CustomType: fw.NodesType{
										ObjectType: types.ObjectType{
											AttrTypes: fw.NodesValue{}.AttributeTypes(ctx),
										},
									},
								},
								Computed:    true,
								Description: "holds node configuration for the cluster",
							},
						},
						CustomType: fw.ConfigType{
							ObjectType: types.ObjectType{
								AttrTypes: fw.ConfigValue{}.AttributeTypes(ctx),
							},
						},
						Computed:    true,
						Description: "MKS V3 cluster specification",
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
						CustomType: fw.ProxyType{
							ObjectType: types.ObjectType{
								AttrTypes: fw.ProxyValue{}.AttributeTypes(ctx),
							},
						},
						Computed: true,
					},
					"sharing": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								Computed: true,
							},
							"projects": schema.SetNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											Computed: true,
										},
									},
									CustomType: fw.ProjectsType{
										ObjectType: types.ObjectType{
											AttrTypes: fw.ProjectsValue{}.AttributeTypes(ctx),
										},
									},
								},
								Computed: true,
							},
						},
						CustomType: fw.SharingType{
							ObjectType: types.ObjectType{
								AttrTypes: fw.SharingValue{}.AttributeTypes(ctx),
							},
						},
						Computed: true,
					},
					"system_components_placement": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"daemon_set_override": schema.SingleNestedAttribute{
								Attributes: map[string]schema.Attribute{
									"daemon_set_tolerations": schema.SetNestedAttribute{
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
											CustomType: fw.DaemonSetTolerationsType{
												ObjectType: types.ObjectType{
													AttrTypes: fw.DaemonSetTolerationsValue{}.AttributeTypes(ctx),
												},
											},
										},
										Computed: true,
									},
									"node_selection_enabled": schema.BoolAttribute{
										Computed: true,
									},
								},
								CustomType: fw.DaemonSetOverrideType{
									ObjectType: types.ObjectType{
										AttrTypes: fw.DaemonSetOverrideValue{}.AttributeTypes(ctx),
									},
								},
								Computed: true,
							},
							"node_selector": schema.MapAttribute{
								ElementType: types.StringType,
								Computed:    true,
							},
							"tolerations": schema.SetNestedAttribute{
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
									CustomType: fw.TolerationsType{
										ObjectType: types.ObjectType{
											AttrTypes: fw.TolerationsValue{}.AttributeTypes(ctx),
										},
									},
								},
								Computed: true,
							},
						},
						CustomType: fw.SystemComponentsPlacementType{
							ObjectType: types.ObjectType{
								AttrTypes: fw.SystemComponentsPlacementValue{}.AttributeTypes(ctx),
							},
						},
						Computed: true,
					},
					"type": schema.StringAttribute{
						Computed:    true,
						Description: "The type of the cluster this spec corresponds to",
					},
				},
				CustomType: fw.SpecType{
					ObjectType: types.ObjectType{
						AttrTypes: fw.SpecValue{}.AttributeTypes(ctx),
					},
				},
				Computed:    true,
				Description: "cluster specification",
			},
		},
	}
}

// MksClusterDataSource defines the data source implementation of MksClusterResource.
type MksClusterDataSource struct {
	client typed.Client
}

func (d *MksClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mks_cluster"
}

func (d *MksClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = MksClusterDataSourceSchema(ctx)
}

func (d *MksClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(typed.Client)

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
	var data fw.MksClusterModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	fmt.Println("mks cluster data", data)
	// Fetch the cluster from the Hub
	hub, err := d.client.InfraV3().Cluster().Get(ctx, options.GetOptions{
		Name:    data.Metadata.Name.ValueString(),
		Project: data.Metadata.Project.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch data", err.Error())
		return
	}
	fmt.Println("mks cluster data hub", hub)
	// convert the hub respo into the TF model
	resp.Diagnostics.Append(fw.ConvertMksClusterFromHub(ctx, hub, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
