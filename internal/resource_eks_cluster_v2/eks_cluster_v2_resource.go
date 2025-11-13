package resource_eks_cluster_v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/clusterctl"
	"github.com/RafaySystems/rctl/pkg/config"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/go-yaml/yaml"

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

	// Convert Terraform model to API structs
	eksCluster, eksClusterConfig, diags := convertModelToClusterSpec(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster name and project for API calls
	clusterName := eksCluster.Metadata.Name
	projectName := eksCluster.Metadata.Project

	tflog.Info(ctx, "Creating cluster", map[string]interface{}{
		"name":    clusterName,
		"project": projectName,
	})

	// Get project ID from name
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get project ID",
			fmt.Sprintf("Could not convert project name %s to ID: %s", projectName, err.Error()),
		)
		return
	}

	// Set cluster sharing external flag if spec.sharing is specified
	var cse string
	if eksCluster.Spec.Sharing != nil {
		cse = "false"
	}

	// Encode cluster spec to YAML
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	if err := encoder.Encode(eksCluster); err != nil {
		resp.Diagnostics.AddError(
			"Failed to encode cluster spec",
			fmt.Sprintf("Could not encode cluster metadata: %s", err.Error()),
		)
		return
	}
	if err := encoder.Encode(eksClusterConfig); err != nil {
		resp.Diagnostics.AddError(
			"Failed to encode cluster config",
			fmt.Sprintf("Could not encode cluster config: %s", err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Cluster YAML", map[string]interface{}{
		"yaml": b.String(),
	})

	// Apply cluster via clusterctl
	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()

	response, err := clusterctl.Apply(logger, rctlConfig, clusterName, b.Bytes(), false, false, false, false, uaDef, cse)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create cluster",
			fmt.Sprintf("clusterctl.Apply failed: %s", err.Error()),
		)
		return
	}

	// Parse response to get task ID
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse API response",
			fmt.Sprintf("Could not parse clusterctl response: %s", err.Error()),
		)
		return
	}

	if res.TaskSetID == "" {
		tflog.Warn(ctx, "No task ID returned, cluster may already exist or operation completed immediately")
	} else {
		tflog.Info(ctx, "Cluster creation task started", map[string]interface{}{
			"taskSetID": res.TaskSetID,
		})
	}

	// Wait a moment for cluster to be created in database
	time.Sleep(10 * time.Second)

	// Get cluster to set ID
	s, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
	if errGet != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster after creation",
			fmt.Sprintf("Could not retrieve cluster %s: %s", clusterName, errGet.Error()),
		)
		return
	}

	// Set ID early so partial state is saved if operation times out
	data.ID = types.StringValue(s.ID)

	tflog.Info(ctx, "Cluster created in database, waiting for provisioning to complete (15-20 minutes typical)")

	// Poll for cluster readiness
	if res.TaskSetID != "" {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				resp.Diagnostics.AddWarning(
					"Cluster creation timeout",
					fmt.Sprintf("Cluster %s provisioning timed out. The cluster may still be provisioning in the background. Check the Rafay console for status.", clusterName),
				)
				// Still save the state with what we have
				break LOOP

			case <-ticker.C:
				tflog.Debug(ctx, "Checking cluster provision status...")

				// Check cluster status
				check, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
				if errGet != nil {
					resp.Diagnostics.AddError(
						"Failed to check cluster status",
						fmt.Sprintf("Could not retrieve cluster status: %s", errGet.Error()),
					)
					return
				}

				// Get detailed status via clusterctl
				rctlConfig.ProjectID = projectID
				statusResp, err := clusterctl.Status(logger, rctlConfig, res.TaskSetID)
				if err != nil {
					tflog.Error(ctx, "Failed to get cluster status", map[string]interface{}{
						"error": err.Error(),
					})
					continue
				}

				sres := clusterCTLResponse{}
				err = json.Unmarshal([]byte(statusResp), &sres)
				if err != nil {
					tflog.Error(ctx, "Failed to parse status response", map[string]interface{}{
						"error": err.Error(),
					})
					continue
				}

				tflog.Info(ctx, "Cluster status", map[string]interface{}{
					"status": sres.Status,
				})

				if strings.Contains(sres.Status, "STATUS_COMPLETE") {
					// Check cluster conditions for blueprint sync
					conditionsFailure, clusterReadiness, err := getClusterConditions(check.ID, projectID)
					if err != nil {
						resp.Diagnostics.AddError(
							"Failed to check cluster conditions",
							fmt.Sprintf("Could not check cluster readiness: %s", err.Error()),
						)
						return
					}

					if conditionsFailure {
						resp.Diagnostics.AddError(
							"Blueprint sync failed",
							fmt.Sprintf("Blueprint sync failed for cluster %s", clusterName),
						)
						return
					} else if clusterReadiness {
						tflog.Info(ctx, "Cluster is ready")
						break LOOP
					} else {
						tflog.Info(ctx, "Cluster provisioning complete, waiting for cluster to be ready...")
					}
				} else if strings.Contains(sres.Status, "STATUS_FAILED") {
					resp.Diagnostics.AddError(
						"Cluster provisioning failed",
						fmt.Sprintf("Failed to provision cluster %s: %s", clusterName, statusResp),
					)
					return
				}
			}
		}
	}

	// Handle cluster sharing external flag
	edgeDb, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster after provisioning",
			fmt.Sprintf("Could not retrieve cluster: %s", err.Error()),
		)
		return
	}

	cseFromDb := edgeDb.Settings[clusterSharingExtKey]
	if cseFromDb != "true" {
		if eksCluster.Spec.Sharing == nil && cseFromDb != "" {
			// Reset cse as sharing is removed
			edgeDb.Settings[clusterSharingExtKey] = ""
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "Failed to reset cluster sharing flag", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
		if eksCluster.Spec.Sharing != nil && cseFromDb != "false" {
			// Explicitly set cse to false
			edgeDb.Settings[clusterSharingExtKey] = "false"
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "Failed to set cluster sharing flag", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}

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

	// Extract cluster metadata from state
	var clusterModel ClusterModel
	diags := data.Cluster.As(ctx, &clusterModel, types.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var metadataModel ClusterMetadataModel
	diags = clusterModel.Metadata.As(ctx, &metadataModel, types.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterName := metadataModel.Name.ValueString()
	projectName := metadataModel.Project.ValueString()

	tflog.Info(ctx, "Reading cluster", map[string]interface{}{
		"name":    clusterName,
		"project": projectName,
	})

	// Get project ID
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get project ID",
			fmt.Sprintf("Could not convert project name %s to ID: %s", projectName, err.Error()),
		)
		return
	}

	// Get cluster from API
	c, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		tflog.Error(ctx, "Failed to get cluster", map[string]interface{}{
			"error": err.Error(),
		})
		if strings.Contains(err.Error(), "not found") {
			// Cluster no longer exists, remove from state
			tflog.Warn(ctx, "Cluster not found, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read cluster",
			fmt.Sprintf("Could not retrieve cluster %s: %s", clusterName, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Got cluster from backend", map[string]interface{}{
		"cluster_id": c.ID,
	})

	// Get cluster spec
	logger := glogger.GetLogger()
	rctlCfg := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, projectID, uaDef)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster spec",
			fmt.Sprintf("Could not retrieve cluster spec for %s: %s", clusterName, err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Got cluster spec from backend")

	// Parse YAML
	decoder := yaml.NewDecoder(bytes.NewReader([]byte(clusterSpecYaml)))

	var eksCluster rafay.EKSCluster
	if err := decoder.Decode(&eksCluster); err != nil {
		resp.Diagnostics.AddError(
			"Failed to decode cluster metadata",
			fmt.Sprintf("Could not parse cluster YAML: %s", err.Error()),
		)
		return
	}

	var eksClusterConfig rafay.EKSClusterConfig
	if err := decoder.Decode(&eksClusterConfig); err != nil {
		resp.Diagnostics.AddError(
			"Failed to decode cluster config",
			fmt.Sprintf("Could not parse cluster config YAML: %s", err.Error()),
		)
		return
	}

	// Convert API response to model
	newData, convertDiags := convertClusterSpecToModel(ctx, &eksCluster, &eksClusterConfig)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve ID
	newData.ID = data.ID

	resp.Diagnostics.Append(resp.State.Set(ctx, newData)...)

	tflog.Info(ctx, "Successfully read EKS cluster v2")
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

	// Extract cluster metadata from plan
	var clusterModel ClusterModel
	diags := data.Cluster.As(ctx, &clusterModel, types.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var metadataModel ClusterMetadataModel
	diags = clusterModel.Metadata.As(ctx, &metadataModel, types.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterName := metadataModel.Name.ValueString()
	projectName := metadataModel.Project.ValueString()

	tflog.Info(ctx, "Updating cluster", map[string]interface{}{
		"name":    clusterName,
		"project": projectName,
	})

	// Get project ID
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get project ID",
			fmt.Sprintf("Could not convert project name %s to ID: %s", projectName, err.Error()),
		)
		return
	}

	// Get current cluster
	c, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get cluster",
			fmt.Sprintf("Could not retrieve cluster %s: %s", clusterName, err.Error()),
		)
		return
	}

	// Verify cluster ID matches
	if c.ID != data.ID.ValueString() {
		resp.Diagnostics.AddError(
			"Cluster ID mismatch",
			fmt.Sprintf("State ID %s does not match current ID %s", data.ID.ValueString(), c.ID),
		)
		return
	}

	// Check cluster sharing external flag
	cse := c.Settings[clusterSharingExtKey]
	tflog.Debug(ctx, "Cluster sharing external flag", map[string]interface{}{
		"cse": cse,
	})

	// If cluster sharing is externally managed, check for conflicts
	if cse == "true" {
		var stateData EKSClusterV2ResourceModel
		diags := req.State.Get(ctx, &stateData)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Check if sharing configuration changed
		if !data.Cluster.Equal(stateData.Cluster) {
			var stateCluster, planCluster ClusterModel
			data.Cluster.As(ctx, &planCluster, types.ObjectAsOptions{})
			stateData.Cluster.As(ctx, &stateCluster, types.ObjectAsOptions{})

			if !planCluster.Spec.Equal(stateCluster.Spec) {
				resp.Diagnostics.AddError(
					"Cluster sharing externally managed",
					"Cluster sharing is currently managed through the external 'rafay_cluster_sharing' resource. To prevent configuration conflicts, please remove the sharing settings from the 'rafay_eks_cluster_v2' resource and manage sharing exclusively via the external resource.",
				)
				return
			}
		}
	}

	// Update follows the same process as create (upsert pattern)
	// Convert model to API structs
	eksCluster, eksClusterConfig, convertDiags := convertModelToClusterSpec(ctx, &data)
	resp.Diagnostics.Append(convertDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Encode to YAML
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	if err := encoder.Encode(eksCluster); err != nil {
		resp.Diagnostics.AddError(
			"Failed to encode cluster spec",
			fmt.Sprintf("Could not encode cluster metadata: %s", err.Error()),
		)
		return
	}
	if err := encoder.Encode(eksClusterConfig); err != nil {
		resp.Diagnostics.AddError(
			"Failed to encode cluster config",
			fmt.Sprintf("Could not encode cluster config: %s", err.Error()),
		)
		return
	}

	// Apply update
	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()

	response, err := clusterctl.Apply(logger, rctlConfig, clusterName, b.Bytes(), false, false, false, false, uaDef, "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to update cluster",
			fmt.Sprintf("clusterctl.Apply failed: %s", err.Error()),
		)
		return
	}

	// Parse response
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse API response",
			fmt.Sprintf("Could not parse clusterctl response: %s", err.Error()),
		)
		return
	}

	tflog.Info(ctx, "Cluster update task started", map[string]interface{}{
		"taskSetID": res.TaskSetID,
	})

	// Poll for update completion (similar to create)
	if res.TaskSetID != "" {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

	LOOP:
		for {
			select {
			case <-ctx.Done():
				resp.Diagnostics.AddWarning(
					"Cluster update timeout",
					fmt.Sprintf("Cluster %s update timed out. The cluster may still be updating in the background. Check the Rafay console for status.", clusterName),
				)
				break LOOP

			case <-ticker.C:
				rctlConfig.ProjectID = projectID
				statusResp, err := clusterctl.Status(logger, rctlConfig, res.TaskSetID)
				if err != nil {
					tflog.Error(ctx, "Failed to get cluster status", map[string]interface{}{
						"error": err.Error(),
					})
					continue
				}

				sres := clusterCTLResponse{}
				err = json.Unmarshal([]byte(statusResp), &sres)
				if err != nil {
					tflog.Error(ctx, "Failed to parse status response", map[string]interface{}{
						"error": err.Error(),
					})
					continue
				}

				if strings.Contains(sres.Status, "STATUS_COMPLETE") {
					tflog.Info(ctx, "Cluster update complete")
					break LOOP
				} else if strings.Contains(sres.Status, "STATUS_FAILED") {
					resp.Diagnostics.AddError(
						"Cluster update failed",
						fmt.Sprintf("Failed to update cluster %s: %s", clusterName, statusResp),
					)
					return
				}
			}
		}
	}

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

	// Extract cluster metadata from state
	var clusterModel ClusterModel
	diags := data.Cluster.As(ctx, &clusterModel, types.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var metadataModel ClusterMetadataModel
	diags = clusterModel.Metadata.As(ctx, &metadataModel, types.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterName := metadataModel.Name.ValueString()
	projectName := metadataModel.Project.ValueString()

	tflog.Info(ctx, "Deleting cluster", map[string]interface{}{
		"name":    clusterName,
		"project": projectName,
	})

	// Get project ID
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get project ID",
			fmt.Sprintf("Could not convert project name %s to ID: %s", projectName, err.Error()),
		)
		return
	}

	// Delete cluster
	errDel := cluster.DeleteCluster(clusterName, projectID, false, uaDef)
	if errDel != nil {
		resp.Diagnostics.AddError(
			"Failed to delete cluster",
			fmt.Sprintf("Could not delete cluster %s: %s", clusterName, errDel.Error()),
		)
		return
	}

	tflog.Info(ctx, "Cluster deletion initiated, polling for completion...")

	// Poll for deletion completion
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			resp.Diagnostics.AddWarning(
				"Cluster deletion timeout",
				fmt.Sprintf("Cluster %s deletion timed out. The cluster may still be deleting in the background. Check the Rafay console for status.", clusterName),
			)
			// Consider deletion successful after timeout
			return

		case <-ticker.C:
			tflog.Debug(ctx, "Checking if cluster still exists...")
			check, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
			if errGet != nil {
				// Error getting cluster usually means it's been deleted
				tflog.Info(ctx, "Cluster not found, deletion complete", map[string]interface{}{
					"error": errGet.Error(),
				})
				return
			}
			if check == nil {
				tflog.Info(ctx, "Cluster deleted successfully")
				return
			}
			tflog.Info(ctx, "Cluster still exists, waiting 60 more seconds...", map[string]interface{}{
				"cluster_id": check.ID,
			})
		}
	}
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

// Constants
const (
	clusterSharingExtKey = "cluster_sharing_external"
	uaDef                = "terraform-provider-rafay-eks-cluster-v2"
)

// clusterCTLResponse represents the response from clusterctl operations
type clusterCTLResponse struct {
	TaskSetID string `json:"taskset_id,omitempty"`
	Status    string `json:"status,omitempty"`
}

// getProjectIDFromName converts a project name to project ID
func getProjectIDFromName(projectName string) (string, error) {
	// This function needs to be imported from the rafay package or implemented here
	// For now, assuming it's available from the rafay package
	return rafay.GetProjectIDFromName(projectName)
}

// getClusterConditions checks cluster conditions for blueprint sync and readiness
func getClusterConditions(clusterID, projectID string) (conditionsFailure, clusterReadiness bool, err error) {
	// This function needs to be imported from the rafay package or implemented here
	// For now, assuming it's available from the rafay package
	return rafay.GetClusterConditions(clusterID, projectID)
}

