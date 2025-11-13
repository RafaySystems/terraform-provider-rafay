package resource_eks_cluster_v2

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &EKSClusterV2Resource{}
var _ resource.ResourceWithImportState = &EKSClusterV2Resource{}
var _ resource.ResourceWithConfigure = &EKSClusterV2Resource{}

func NewEKSClusterV2Resource() resource.Resource {
	return &EKSClusterV2Resource{}
}

// EKSClusterV2Resource defines the resource implementation.
type EKSClusterV2Resource struct {
	client interface{}
}

// EKSClusterV2ResourceModel describes the resource data model using maps.
type EKSClusterV2ResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Cluster       types.Object `tfsdk:"cluster"` // Single nested
	ClusterConfig types.Object `tfsdk:"cluster_config"` // Single nested
	Timeouts      types.Object `tfsdk:"timeouts"`
}

// ClusterModel represents the cluster metadata (formerly a list, now a single nested object)
type ClusterModel struct {
	Kind     types.String `tfsdk:"kind"`
	Metadata types.Object `tfsdk:"metadata"` // Single nested
	Spec     types.Object `tfsdk:"spec"`     // Single nested
}

// ClusterMetadataModel represents the cluster metadata fields
type ClusterMetadataModel struct {
	Name    types.String `tfsdk:"name"`
	Project types.String `tfsdk:"project"`
	Labels  types.Map    `tfsdk:"labels"` // Map instead of TypeMap
}

// ClusterSpecModel represents the cluster spec
type ClusterSpecModel struct {
	Type                       types.String `tfsdk:"type"`
	Blueprint                  types.String `tfsdk:"blueprint"`
	BlueprintVersion           types.String `tfsdk:"blueprint_version"`
	CloudProvider              types.String `tfsdk:"cloud_provider"`
	CrossAccountRoleArn        types.String `tfsdk:"cross_account_role_arn"`
	CniProvider                types.String `tfsdk:"cni_provider"`
	CniParams                  types.Object `tfsdk:"cni_params"` // Single nested
	ProxyConfig                types.Map    `tfsdk:"proxy_config"`
	SystemComponentsPlacement  types.Object `tfsdk:"system_components_placement"` // Single nested
	Sharing                    types.Object `tfsdk:"sharing"` // Single nested
}

// SharingModel represents cluster sharing configuration
type SharingModel struct {
	Enabled  types.Bool `tfsdk:"enabled"`
	Projects types.Map  `tfsdk:"projects"` // Map of project names to project objects
}

// CNIParamsModel represents CNI parameters
type CNIParamsModel struct {
	CustomCniCidr    types.String `tfsdk:"custom_cni_cidr"`
	CustomCniCredits types.String `tfsdk:"custom_cni_credits"`
}

// SystemComponentsPlacementModel represents system components placement
type SystemComponentsPlacementModel struct {
	NodeSelector types.Map    `tfsdk:"node_selector"`
	Tolerations  types.Map    `tfsdk:"tolerations"` // Map of toleration key to toleration config
	DaemonsetNodeSelector types.Map `tfsdk:"daemonset_node_selector"`
	DaemonsetTolerations  types.Map `tfsdk:"daemonset_tolerations"` // Map of toleration key to toleration config
}

// ClusterConfigModel represents the EKS-specific configuration (formerly a list)
type ClusterConfigModel struct {
	APIVersion types.String `tfsdk:"apiversion"`
	Kind       types.String `tfsdk:"kind"`
	Metadata   types.Object `tfsdk:"metadata"` // Single nested
	VPC        types.Object `tfsdk:"vpc"`      // Single nested
	NodeGroups types.Map    `tfsdk:"node_groups"` // Map of node group name to node group config
	ManagedNodeGroups types.Map `tfsdk:"managed_node_groups"` // Map instead of list
	IdentityProviders types.Map `tfsdk:"identity_providers"` // Map of provider name to provider config
	EncryptionConfig  types.Object `tfsdk:"encryption_config"`
	AccessEntries     types.Map    `tfsdk:"access_entries"` // Map by principal ARN or name
	IdentityMappings  types.Object `tfsdk:"identity_mappings"` // Single nested with ARN and account maps
}

// VPCModel represents VPC configuration
type VPCModel struct {
	Region                   types.String `tfsdk:"region"`
	CIDR                     types.String `tfsdk:"cidr"`
	ClusterResourcesVpcConfig types.Object `tfsdk:"cluster_resources_vpc_config"`
	Subnets                  types.Object `tfsdk:"subnets"` // Single nested with public/private maps
	NAT                      types.Object `tfsdk:"nat"`
	SecurityGroup            types.Object `tfsdk:"security_group"`
}

// SubnetsModel represents subnet configuration with maps
type SubnetsModel struct {
	Public  types.Map `tfsdk:"public"`  // Map of AZ to subnet config
	Private types.Map `tfsdk:"private"` // Map of AZ to subnet config
}

// NodeGroupModel represents a node group configuration
type NodeGroupModel struct {
	Name                      types.String `tfsdk:"name"`
	AMI                       types.String `tfsdk:"ami"`
	IAM                       types.Object `tfsdk:"iam"`
	InstanceType              types.String `tfsdk:"instance_type"`
	AvailabilityZones         types.List   `tfsdk:"availability_zones"`
	DesiredCapacity           types.Int64  `tfsdk:"desired_capacity"`
	MinSize                   types.Int64  `tfsdk:"min_size"`
	MaxSize                   types.Int64  `tfsdk:"max_size"`
	VolumeSize                types.Int64  `tfsdk:"volume_size"`
	VolumeType                types.String `tfsdk:"volume_type"`
	Labels                    types.Map    `tfsdk:"labels"`
	Tags                      types.Map    `tfsdk:"tags"`
	PrivateNetworking         types.Bool   `tfsdk:"private_networking"`
	SecurityGroups            types.Object `tfsdk:"security_groups"`
	SSH                       types.Object `tfsdk:"ssh"`
	Taints                    types.Map    `tfsdk:"taints"` // Map of taint key to taint config
	UpdateConfig              types.Object `tfsdk:"update_config"`
	ScalingConfig             types.Object `tfsdk:"scaling_config"`
}

func (r *EKSClusterV2Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eks_cluster_v2"
}

func (r *EKSClusterV2Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "EKS Cluster resource using Plugin Framework with map-based schema",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for the cluster",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster": schema.SingleNestedAttribute{
				MarkdownDescription: "Rafay specific cluster configuration",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"kind": schema.StringAttribute{
						MarkdownDescription: "The type of resource. Supported value is `Cluster`.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("Cluster"),
					},
					"metadata": schema.SingleNestedAttribute{
						MarkdownDescription: "Contains data that helps uniquely identify the resource.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "The name of the EKS cluster in Rafay console. This must be unique in your organization.",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"project": schema.StringAttribute{
								MarkdownDescription: "The name of the Rafay project the cluster will be created in.",
								Required:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"labels": schema.MapAttribute{
								MarkdownDescription: "The labels for the cluster in Rafay console.",
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"spec": schema.SingleNestedAttribute{
						MarkdownDescription: "The specification associated with the cluster.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"type": schema.StringAttribute{
								MarkdownDescription: "The cluster type. Supported value is `aws-eks`.",
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString("aws-eks"),
								Validators: []validator.String{
									stringvalidator.OneOf("aws-eks"),
								},
							},
							"blueprint": schema.StringAttribute{
								MarkdownDescription: "The blueprint associated with the cluster.",
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString("default"),
							},
							"blueprint_version": schema.StringAttribute{
								MarkdownDescription: "The blueprint version associated with the cluster.",
								Optional:            true,
							},
							"cloud_provider": schema.StringAttribute{
								MarkdownDescription: "The cloud credentials provider used to create and manage the cluster.",
								Required:            true,
							},
							"cross_account_role_arn": schema.StringAttribute{
								MarkdownDescription: "Role ARN of the linked account",
								Optional:            true,
							},
							"cni_provider": schema.StringAttribute{
								MarkdownDescription: "The container network interface (CNI) provider.",
								Optional:            true,
								Computed:            true,
								Default:             stringdefault.StaticString("aws-cni"),
							},
							"cni_params": schema.SingleNestedAttribute{
								MarkdownDescription: "The container network interface (CNI) parameters.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"custom_cni_cidr": schema.StringAttribute{
										MarkdownDescription: "Custom CNI CIDR block.",
										Optional:            true,
									},
									"custom_cni_credits": schema.StringAttribute{
										MarkdownDescription: "Custom CNI credits.",
										Optional:            true,
									},
								},
							},
							"proxy_config": schema.MapAttribute{
								MarkdownDescription: "The proxy configuration for the cluster.",
								Optional:            true,
								ElementType:         types.StringType,
							},
							"system_components_placement": schema.SingleNestedAttribute{
								MarkdownDescription: "Configure tolerations and nodeSelector for Rafay system components.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
								"node_selector": schema.MapAttribute{
									MarkdownDescription: "Node selector for system components.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"tolerations": schema.MapNestedAttribute{
									MarkdownDescription: "Tolerations for system components mapped by toleration key.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												MarkdownDescription: "The toleration key.",
												Required: true,
											},
											"operator": schema.StringAttribute{
												MarkdownDescription: "The toleration operator (Equal, Exists).",
												Optional: true,
											},
											"value": schema.StringAttribute{
												MarkdownDescription: "The toleration value.",
												Optional: true,
											},
											"effect": schema.StringAttribute{
												MarkdownDescription: "The toleration effect (NoSchedule, PreferNoSchedule, NoExecute).",
												Optional: true,
											},
										},
									},
								},
								"daemonset_node_selector": schema.MapAttribute{
									MarkdownDescription: "Node selector for daemonset components.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"daemonset_tolerations": schema.MapNestedAttribute{
									MarkdownDescription: "Tolerations for daemonset components mapped by toleration key.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												MarkdownDescription: "The toleration key.",
												Required: true,
											},
											"operator": schema.StringAttribute{
												MarkdownDescription: "The toleration operator (Equal, Exists).",
												Optional: true,
											},
											"value": schema.StringAttribute{
												MarkdownDescription: "The toleration value.",
												Optional: true,
											},
											"effect": schema.StringAttribute{
												MarkdownDescription: "The toleration effect (NoSchedule, PreferNoSchedule, NoExecute).",
												Optional: true,
											},
										},
									},
								},
								},
							},
							"sharing": schema.SingleNestedAttribute{
								MarkdownDescription: "The sharing configuration for the cluster.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"enabled": schema.BoolAttribute{
										MarkdownDescription: "Enable sharing for this resource.",
										Optional:            true,
									},
									"projects": schema.MapNestedAttribute{
										MarkdownDescription: "Map of projects this resource is shared with (key: project name).",
										Optional:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													MarkdownDescription: "The name of the project to share the resource.",
													Required:            true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"cluster_config": schema.SingleNestedAttribute{
				MarkdownDescription: "EKS specific cluster configuration",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"apiversion": schema.StringAttribute{
						MarkdownDescription: "API version for EKS config.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("rafay.io/v1alpha5"),
					},
					"kind": schema.StringAttribute{
						MarkdownDescription: "Kind of EKS config.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("ClusterConfig"),
					},
					"metadata": schema.SingleNestedAttribute{
						MarkdownDescription: "EKS cluster metadata.",
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "Name of the EKS cluster.",
								Required:            true,
							},
							"region": schema.StringAttribute{
								MarkdownDescription: "AWS region for the cluster.",
								Required:            true,
							},
							"version": schema.StringAttribute{
								MarkdownDescription: "Kubernetes version.",
								Required:            true,
							},
							"tags": schema.MapAttribute{
								MarkdownDescription: "Tags for the EKS cluster.",
								Optional:            true,
								ElementType:         types.StringType,
							},
						},
					},
					"vpc": schema.SingleNestedAttribute{
						MarkdownDescription: "VPC configuration for the cluster.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"region": schema.StringAttribute{
								MarkdownDescription: "AWS region.",
								Optional:            true,
							},
							"cidr": schema.StringAttribute{
								MarkdownDescription: "CIDR block for the VPC.",
								Optional:            true,
							},
							"cluster_resources_vpc_config": schema.SingleNestedAttribute{
								MarkdownDescription: "VPC config for cluster resources.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"endpoint_private_access": schema.BoolAttribute{
										Optional: true,
									},
									"endpoint_public_access": schema.BoolAttribute{
										Optional: true,
									},
									"public_access_cidrs": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
							"subnets": schema.SingleNestedAttribute{
								MarkdownDescription: "Subnet configuration using maps by availability zone.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"public": schema.MapNestedAttribute{
										MarkdownDescription: "Public subnets mapped by availability zone.",
										Optional:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													MarkdownDescription: "Subnet ID.",
													Optional:            true,
												},
												"cidr": schema.StringAttribute{
													MarkdownDescription: "CIDR block.",
													Optional:            true,
												},
												"az": schema.StringAttribute{
													MarkdownDescription: "Availability zone.",
													Optional:            true,
												},
											},
										},
									},
									"private": schema.MapNestedAttribute{
										MarkdownDescription: "Private subnets mapped by availability zone.",
										Optional:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													MarkdownDescription: "Subnet ID.",
													Optional:            true,
												},
												"cidr": schema.StringAttribute{
													MarkdownDescription: "CIDR block.",
													Optional:            true,
												},
												"az": schema.StringAttribute{
													MarkdownDescription: "Availability zone.",
													Optional:            true,
												},
											},
										},
									},
								},
							},
							"nat": schema.SingleNestedAttribute{
								MarkdownDescription: "NAT Gateway configuration.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"gateway": schema.StringAttribute{
										Optional: true,
									},
								},
							},
							"security_group": schema.SingleNestedAttribute{
								MarkdownDescription: "Security group configuration.",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"attach_ids": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
									},
								},
							},
						},
					},
					"node_groups": schema.MapNestedAttribute{
						MarkdownDescription: "Self-managed node groups mapped by name.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Node group name.",
									Required:            true,
								},
								"ami": schema.StringAttribute{
									MarkdownDescription: "AMI ID.",
									Optional:            true,
								},
								"instance_type": schema.StringAttribute{
									MarkdownDescription: "EC2 instance type.",
									Optional:            true,
								},
								"availability_zones": schema.ListAttribute{
									MarkdownDescription: "Availability zones.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"desired_capacity": schema.Int64Attribute{
									MarkdownDescription: "Desired capacity.",
									Optional:            true,
								},
								"min_size": schema.Int64Attribute{
									MarkdownDescription: "Minimum size.",
									Optional:            true,
								},
								"max_size": schema.Int64Attribute{
									MarkdownDescription: "Maximum size.",
									Optional:            true,
								},
								"volume_size": schema.Int64Attribute{
									MarkdownDescription: "EBS volume size.",
									Optional:            true,
								},
								"volume_type": schema.StringAttribute{
									MarkdownDescription: "EBS volume type.",
									Optional:            true,
								},
								"labels": schema.MapAttribute{
									MarkdownDescription: "Kubernetes labels.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"tags": schema.MapAttribute{
									MarkdownDescription: "AWS tags.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"private_networking": schema.BoolAttribute{
									MarkdownDescription: "Enable private networking.",
									Optional:            true,
								},
								"taints": schema.MapNestedAttribute{
									MarkdownDescription: "Kubernetes taints mapped by key.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												Required: true,
											},
											"value": schema.StringAttribute{
												Optional: true,
											},
											"effect": schema.StringAttribute{
												Optional: true,
											},
										},
									},
								},
								"iam": schema.SingleNestedAttribute{
									MarkdownDescription: "IAM configuration.",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"instance_profile_arn": schema.StringAttribute{
											Optional: true,
										},
										"instance_role_arn": schema.StringAttribute{
											Optional: true,
										},
									},
								},
								"security_groups": schema.SingleNestedAttribute{
									MarkdownDescription: "Security groups.",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"attach_ids": schema.ListAttribute{
											Optional:    true,
											ElementType: types.StringType,
										},
									},
								},
								"ssh": schema.SingleNestedAttribute{
									MarkdownDescription: "SSH configuration.",
									Optional:            true,
									Attributes: map[string]schema.Attribute{
										"public_key_name": schema.StringAttribute{
											Optional: true,
										},
										"source_security_group_ids": schema.ListAttribute{
											Optional:    true,
											ElementType: types.StringType,
										},
									},
								},
							},
						},
					},
					"managed_node_groups": schema.MapNestedAttribute{
						MarkdownDescription: "EKS managed node groups mapped by name.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Managed node group name.",
									Required:            true,
								},
								"ami_type": schema.StringAttribute{
									MarkdownDescription: "AMI type.",
									Optional:            true,
								},
								"instance_types": schema.ListAttribute{
									MarkdownDescription: "Instance types.",
									Optional:            true,
									ElementType:         types.StringType,
								},
								"desired_size": schema.Int64Attribute{
									Optional: true,
								},
								"min_size": schema.Int64Attribute{
									Optional: true,
								},
								"max_size": schema.Int64Attribute{
									Optional: true,
								},
								"disk_size": schema.Int64Attribute{
									Optional: true,
								},
								"labels": schema.MapAttribute{
									Optional:    true,
									ElementType: types.StringType,
								},
								"tags": schema.MapAttribute{
									Optional:    true,
									ElementType: types.StringType,
								},
								"taints": schema.MapNestedAttribute{
									MarkdownDescription: "Kubernetes taints mapped by key.",
									Optional:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												Required: true,
											},
											"value": schema.StringAttribute{
												Optional: true,
											},
											"effect": schema.StringAttribute{
												Optional: true,
											},
										},
									},
								},
								"subnet_ids": schema.ListAttribute{
									Optional:    true,
									ElementType: types.StringType,
								},
								"remote_access": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"ec2_ssh_key": schema.StringAttribute{
											Optional: true,
										},
										"source_security_groups": schema.ListAttribute{
											Optional:    true,
											ElementType: types.StringType,
										},
									},
								},
								"update_config": schema.SingleNestedAttribute{
									Optional: true,
									Attributes: map[string]schema.Attribute{
										"max_unavailable": schema.Int64Attribute{
											Optional: true,
										},
										"max_unavailable_percentage": schema.Int64Attribute{
											Optional: true,
										},
									},
								},
							},
						},
					},
					"identity_providers": schema.MapNestedAttribute{
						MarkdownDescription: "Identity providers for the cluster mapped by provider name.",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									MarkdownDescription: "The type of identity provider (e.g., oidc).",
									Optional: true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the identity provider.",
									Required: true,
								},
								"issuer_url": schema.StringAttribute{
									MarkdownDescription: "The issuer URL for the OIDC provider.",
									Optional: true,
								},
								"client_id": schema.StringAttribute{
									MarkdownDescription: "The client ID for the OIDC provider.",
									Optional: true,
								},
							},
						},
					},
					"encryption_config": schema.SingleNestedAttribute{
						MarkdownDescription: "Encryption configuration.",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"provider": schema.StringAttribute{
								Optional: true,
							},
							"resources": schema.ListAttribute{
								Optional:    true,
								ElementType: types.StringType,
							},
						},
					},
				},
			},
			"timeouts": schema.SingleNestedAttribute{
				MarkdownDescription: "Timeouts for CRUD operations.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"create": schema.StringAttribute{
						Optional: true,
					},
					"update": schema.StringAttribute{
						Optional: true,
					},
					"delete": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (r *EKSClusterV2Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(interface{})
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client interface, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *EKSClusterV2Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data EKSClusterV2ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Creating EKS cluster v2")

	// TODO: Extract cluster configuration from data
	// TODO: Call Rafay API to create cluster
	// TODO: Wait for cluster to be ready
	// TODO: Set ID and state

	// Placeholder implementation
	clusterName := "placeholder-cluster-name"
	projectName := "placeholder-project"

	// Set ID
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", projectName, clusterName))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Successfully created EKS cluster v2", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *EKSClusterV2Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EKSClusterV2ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Reading EKS cluster v2", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// TODO: Call Rafay API to get cluster details
	// TODO: Handle cluster not found (remove from state)
	// TODO: Update data model with latest state

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *EKSClusterV2Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data EKSClusterV2ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Updating EKS cluster v2", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// TODO: Extract cluster configuration changes
	// TODO: Call Rafay API to update cluster
	// TODO: Wait for update to complete

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	tflog.Info(ctx, "Successfully updated EKS cluster v2")
}

func (r *EKSClusterV2Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data EKSClusterV2ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Deleting EKS cluster v2", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// TODO: Call Rafay API to delete cluster
	// TODO: Wait for deletion to complete

	tflog.Info(ctx, "Successfully deleted EKS cluster v2")
}

func (r *EKSClusterV2Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions for data conversion

func parseClusterID(id string) (project, cluster string, err error) {
	parts := splitID(id, 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid cluster ID format, expected: project/cluster, got: %s", id)
	}
	return parts[0], parts[1], nil
}

func splitID(id string, expectedParts int) []string {
	parts := make([]string, 0, expectedParts)
	for i, part := range splitString(id, "/") {
		if i >= expectedParts {
			break
		}
		parts = append(parts, part)
	}
	return parts
}

func splitString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

