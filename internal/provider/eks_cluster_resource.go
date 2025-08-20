package provider

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/RafaySystems/rctl/pkg/cluster"
	config "github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/terraform-provider-rafay/internal/resource_eks_cluster"
	"github.com/RafaySystems/terraform-provider-rafay/rafay"

	"github.com/RafaySystems/rctl/pkg/clusterctl"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/go-yaml/yaml"
)

var _ resource.Resource = (*eksClusterResource)(nil)

func NewEksClusterResource() resource.Resource {
	return &eksClusterResource{}
}

type eksClusterResource struct{}

type eksClusterResourceModel struct {
	Id types.String `tfsdk:"id"`
}

func (r *eksClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eks_cluster"
}

func (r *eksClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_eks_cluster.EksClusterResourceSchema(ctx)
}

func (r *eksClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resource_eks_cluster.EksClusterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ccList := make([]resource_eks_cluster.ClusterConfigValue, 0, len(data.ClusterConfig.Elements()))
	d := data.ClusterConfig.ElementsAs(ctx, &ccList, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	cc := ccList[0]
	_ = cc

	if !cc.NodeGroups.IsNull() && !cc.NodeGroupsMap.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Only one of 'node_groups' or 'node_groups_map' can be set at a time. Please remove one of them.",
		)
	}
}

func (r *eksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	newCluster, d := resource_eks_cluster.ExpandEksCluster(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	newClusterConfig, d := resource_eks_cluster.ExpandEksClusterConfig(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	tflog.Debug(ctx, "cluster value", map[string]any{"newCluster": newCluster, "newClusterConfig": newClusterConfig})

	// Call API to create EKS cluster
	clusterName := newCluster.Metadata.Name
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	if err := encoder.Encode(newCluster); err != nil {
		log.Printf("error encoding cluster: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}
	if err := encoder.Encode(newClusterConfig); err != nil {
		log.Printf("error encoding cluster config: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster config, got error: %s", err))
		return
	}
	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()
	_, err := clusterctl.Apply(logger, rctlConfig, clusterName, b.Bytes(), false, false, false, false, uaDef, "")
	if err != nil {
		log.Printf("cluster error 1: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *eksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "EKS Cluster Read existing data", map[string]interface{}{"data": data})

	// TODO(Akshay): handle null/unknown cluster

	clusterEls := make([]resource_eks_cluster.ClusterValue, 0, len(data.Cluster.Elements()))
	d := data.Cluster.ElementsAs(ctx, &clusterEls, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	metadataEls := make([]resource_eks_cluster.MetadataValue, 0, len(clusterEls[0].Metadata.Elements()))
	d = clusterEls[0].Metadata.ElementsAs(ctx, &metadataEls, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	mdO, d := metadataEls[0].ToObjectValue(ctx)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	var md resource_eks_cluster.MetadataType
	mdObj, d := md.ValueFromObject(ctx, mdO)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	mdValue, ok := mdObj.(resource_eks_cluster.MetadataValue)
	if !ok {
		resp.Diagnostics.AddError("Invalid Metadata", "Expected MetadataValue type but got a different type.")
		return
	}
	clusterName := mdValue.Name.ValueString()
	projectName := mdValue.Project.ValueString()
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get project ID from name '%s', got error: %s", projectName, err))
		return
	}

	// read prior state of cluster config to find out which one out of "node_groups" and "node_groups_map" is being used by user.
	ngMapInUse := true
	ccList := make([]resource_eks_cluster.ClusterConfigValue, 0, len(data.ClusterConfig.Elements()))
	d = data.ClusterConfig.ElementsAs(ctx, &ccList, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	cConfig := ccList[0]
	if !cConfig.NodeGroups.IsNull() {
		ngMapInUse = false
	}

	// Read API call logic
	c, err := cluster.GetCluster(clusterName, projectID, "")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", err))
		return
	}
	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlConfig, c.Name, projectID, "")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster spec, got error: %s", err))
		return
	}
	decoder := yaml.NewDecoder(bytes.NewReader([]byte(clusterSpecYaml)))
	clusterSpec := rafay.EKSCluster{}
	err = decoder.Decode(&clusterSpec)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode the cluster spec, got error: %s", err))
		return
	}
	clusterConfigSpec := rafay.EKSClusterConfig{}
	err = decoder.Decode(&clusterConfigSpec)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode the cluster config spec, got error: %s", err))
		return
	}
	tflog.Debug(ctx, "EKS Cluster Read API data", map[string]interface{}{
		"clusterSpec":       clusterSpec,
		"clusterConfigSpec": clusterConfigSpec,
	})

	_ = ngMapInUse

	// Update the model with the data from the API response
	diags := resource_eks_cluster.FlattenEksCluster(ctx, clusterSpec, &data)
	if diags.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster, got error: %s", diags))
		return
	}

	diags = resource_eks_cluster.FlattenEksClusterConfig(ctx, clusterConfigSpec, &data)
	if diags.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster config, got error: %s", diags))
		return
	}

	// mdv := map[string]attr.Value{
	// 	"name":    types.StringValue(clusterSpec.Metadata.Name),
	// 	"project": types.StringValue(clusterSpec.Metadata.Project),
	// }
	// md1, d := resource_eks_cluster.NewMetadataValue(resource_eks_cluster.MetadataValue{}.AttributeTypes(ctx), mdv)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	// mdElements := []attr.Value{
	// 	md1,
	// }
	// fmd, d := types.ListValue(resource_eks_cluster.MetadataValue{}.Type(ctx), mdElements)
	//
	// specv := map[string]attr.Value{
	// 	"cloud_provider": types.StringValue(clusterSpec.Spec.CloudProvider),
	// 	"blueprint":      types.StringValue(clusterSpec.Spec.Blueprint),
	// 	"cni_provider":   types.StringValue(clusterSpec.Spec.CniProvider),
	// 	"type":           types.StringValue(clusterSpec.Spec.Type),
	// }
	// spec1, d := resource_eks_cluster.NewSpecValue(resource_eks_cluster.SpecValue{}.AttributeTypes(ctx), specv)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	// specElements := []attr.Value{
	// 	spec1,
	// }
	// fspec, d := types.ListValue(resource_eks_cluster.SpecValue{}.Type(ctx), specElements)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	//
	// cv := map[string]attr.Value{
	// 	"kind":     types.StringValue(clusterSpec.Kind),
	// 	"metadata": fmd,
	// 	"spec":     fspec,
	// }
	// fcv, d := resource_eks_cluster.NewClusterValue(resource_eks_cluster.ClusterValue{}.AttributeTypes(ctx), cv)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	// clusterElements := []attr.Value{
	// 	fcv,
	// }
	// data.Cluster, d = types.ListValue(resource_eks_cluster.ClusterValue{}.Type(ctx), clusterElements)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	//
	// /// Start of ClusterConfig
	// tgs := clusterConfigSpec.Metadata.Tags
	// tgsElements := make(map[string]attr.Value, len(tgs))
	// for tk, tv := range tgs {
	// 	tgsElements[tk] = types.StringValue(tv)
	// }
	// tgsV, d := types.MapValue(types.StringType, tgsElements)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	// md2v := map[string]attr.Value{
	// 	"name":    types.StringValue(clusterConfigSpec.Metadata.Name),
	// 	"region":  types.StringValue(clusterConfigSpec.Metadata.Region),
	// 	"version": types.StringValue(clusterConfigSpec.Metadata.Version),
	// 	"tags":    tgsV,
	// }
	// md2, d := resource_eks_cluster.NewMetadata2Value(resource_eks_cluster.Metadata2Value{}.AttributeTypes(ctx), md2v)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	// md2Elements := []attr.Value{
	// 	md2,
	// }
	// fmd2, d := types.ListValue(resource_eks_cluster.Metadata2Value{}.Type(ctx), md2Elements)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	//
	// ngs := clusterConfigSpec.NodeGroups
	// var ngList basetypes.ListValue
	// var ngMap basetypes.MapValue
	//
	// if ngMapInUse {
	// 	ngMapElements := make(map[string]attr.Value, len(ngs))
	// 	for _, ng := range ngs {
	// 		iamaddon2 := map[string]attr.Value{
	// 			"alb_ingress":     types.BoolValue(false),
	// 			"app_mesh":        types.BoolValue(*ng.IAM.WithAddonPolicies.AppMesh),
	// 			"app_mesh_review": types.BoolValue(*ng.IAM.WithAddonPolicies.AppMeshPreview),
	// 			"cert_manager":    types.BoolValue(*ng.IAM.WithAddonPolicies.CertManager),
	// 			"cloud_watch":     types.BoolValue(*ng.IAM.WithAddonPolicies.CloudWatch),
	// 			"ebs":             types.BoolValue(*ng.IAM.WithAddonPolicies.EBS),
	// 			"efs":             types.BoolValue(*ng.IAM.WithAddonPolicies.EFS),
	// 			"external_dns":    types.BoolValue(*ng.IAM.WithAddonPolicies.ExternalDNS),
	// 			"fsx":             types.BoolValue(*ng.IAM.WithAddonPolicies.FSX),
	// 			"xray":            types.BoolValue(*ng.IAM.WithAddonPolicies.XRay),
	// 			"image_builder":   types.BoolValue(*ng.IAM.WithAddonPolicies.ImageBuilder),
	// 			"auto_scaler":     types.BoolValue(*ng.IAM.WithAddonPolicies.AutoScaler),
	// 		}
	// 		iamaddonv2, d := resource_eks_cluster.NewIamNodeGroupWithAddonPolicies2Value(resource_eks_cluster.IamNodeGroupWithAddonPolicies2Value{}.AttributeTypes(ctx), iamaddon2)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		iamaddonElements2 := []attr.Value{
	// 			iamaddonv2,
	// 		}
	// 		fiamaddon2, d := types.ListValue(resource_eks_cluster.IamNodeGroupWithAddonPolicies2Value{}.Type(ctx), iamaddonElements2)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		iamv2 := map[string]attr.Value{
	// 			"iam_node_group_with_addon_policies": fiamaddon2,
	// 		}
	// 		iamo2, d := resource_eks_cluster.NewIam2Value(resource_eks_cluster.Iam2Value{}.AttributeTypes(ctx), iamv2)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		iam2Elements := []attr.Value{
	// 			iamo2,
	// 		}
	// 		fiam2, d := types.ListValue(resource_eks_cluster.Iam2Value{}.Type(ctx), iam2Elements)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		ngmapv := map[string]attr.Value{
	// 			//"name":               types.StringValue(ng.Name),
	// 			"ami_family":         types.StringValue(ng.AMIFamily),
	// 			"instance_type":      types.StringValue(ng.InstanceType),
	// 			"desired_capacity":   types.Int64Value(int64(*ng.DesiredCapacity)),
	// 			"min_size":           types.Int64Value(int64(*ng.MinSize)),
	// 			"max_size":           types.Int64Value(int64(*ng.MaxSize)),
	// 			"max_pods_per_node":  types.Int64Value(int64(ng.MaxPodsPerNode)),
	// 			"version":            types.StringValue(ng.Version),
	// 			"disable_imdsv1":     types.BoolValue(*ng.DisableIMDSv1),
	// 			"disable_pods_imds":  types.BoolValue(*ng.DisablePodIMDS),
	// 			"efa_enabled":        types.BoolValue(*ng.EFAEnabled),
	// 			"private_networking": types.BoolValue(*ng.PrivateNetworking),
	// 			"volume_iops":        types.Int64Value(int64(*ng.VolumeIOPS)),
	// 			"volume_size":        types.Int64Value(int64(*ng.VolumeSize)),
	// 			"volume_throughput":  types.Int64Value(int64(*ng.VolumeThroughput)),
	// 			"volume_type":        types.StringValue(ng.VolumeType),
	// 			"iam":                fiam2,
	// 		}
	// 		ngmapo, d := resource_eks_cluster.NewNodeGroupsMapValue(resource_eks_cluster.NodeGroupsMapValue{}.AttributeTypes(ctx), ngmapv)
	// 		ngMapElements[ng.Name] = ngmapo
	// 	}
	//
	// 	ngMap, d = types.MapValue(resource_eks_cluster.NodeGroupsMapValue{}.Type(ctx), ngMapElements)
	// 	if d.HasError() {
	// 		resp.Diagnostics.Append(d...)
	// 		return
	// 	}
	//
	// } else {
	// 	ngElements := []attr.Value{}
	// 	for _, ng := range ngs {
	// 		iamaddon := map[string]attr.Value{
	// 			"alb_ingress":     types.BoolValue(false),
	// 			"app_mesh":        types.BoolValue(*ng.IAM.WithAddonPolicies.AppMesh),
	// 			"app_mesh_review": types.BoolValue(*ng.IAM.WithAddonPolicies.AppMeshPreview),
	// 			"cert_manager":    types.BoolValue(*ng.IAM.WithAddonPolicies.CertManager),
	// 			"cloud_watch":     types.BoolValue(*ng.IAM.WithAddonPolicies.CloudWatch),
	// 			"ebs":             types.BoolValue(*ng.IAM.WithAddonPolicies.EBS),
	// 			"efs":             types.BoolValue(*ng.IAM.WithAddonPolicies.EFS),
	// 			"external_dns":    types.BoolValue(*ng.IAM.WithAddonPolicies.ExternalDNS),
	// 			"fsx":             types.BoolValue(*ng.IAM.WithAddonPolicies.FSX),
	// 			"xray":            types.BoolValue(*ng.IAM.WithAddonPolicies.XRay),
	// 			"image_builder":   types.BoolValue(*ng.IAM.WithAddonPolicies.ImageBuilder),
	// 			"auto_scaler":     types.BoolValue(*ng.IAM.WithAddonPolicies.AutoScaler),
	// 		}
	// 		iamaddonv, d := resource_eks_cluster.NewIamNodeGroupWithAddonPoliciesValue(resource_eks_cluster.IamNodeGroupWithAddonPoliciesValue{}.AttributeTypes(ctx), iamaddon)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		iamaddonElements := []attr.Value{
	// 			iamaddonv,
	// 		}
	// 		fiamaddon, d := types.ListValue(resource_eks_cluster.IamNodeGroupWithAddonPoliciesValue{}.Type(ctx), iamaddonElements)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		iamv := map[string]attr.Value{
	// 			"iam_node_group_with_addon_policies": fiamaddon,
	// 		}
	// 		iamo, d := resource_eks_cluster.NewIamValue(resource_eks_cluster.IamValue{}.AttributeTypes(ctx), iamv)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		iamElements := []attr.Value{
	// 			iamo,
	// 		}
	// 		fiam, d := types.ListValue(resource_eks_cluster.IamValue{}.Type(ctx), iamElements)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	//
	// 		ngv := map[string]attr.Value{
	// 			"name":               types.StringValue(ng.Name),
	// 			"ami_family":         types.StringValue(ng.AMIFamily),
	// 			"instance_type":      types.StringValue(ng.InstanceType),
	// 			"desired_capacity":   types.Int64Value(int64(*ng.DesiredCapacity)),
	// 			"min_size":           types.Int64Value(int64(*ng.MinSize)),
	// 			"max_size":           types.Int64Value(int64(*ng.MaxSize)),
	// 			"max_pods_per_node":  types.Int64Value(int64(ng.MaxPodsPerNode)),
	// 			"version":            types.StringValue(ng.Version),
	// 			"disable_imdsv1":     types.BoolValue(*ng.DisableIMDSv1),
	// 			"disable_pods_imds":  types.BoolValue(*ng.DisablePodIMDS),
	// 			"efa_enabled":        types.BoolValue(*ng.EFAEnabled),
	// 			"private_networking": types.BoolValue(*ng.PrivateNetworking),
	// 			"volume_iops":        types.Int64Value(int64(*ng.VolumeIOPS)),
	// 			"volume_size":        types.Int64Value(int64(*ng.VolumeSize)),
	// 			"volume_throughput":  types.Int64Value(int64(*ng.VolumeThroughput)),
	// 			"volume_type":        types.StringValue(ng.VolumeType),
	// 			"iam":                fiam,
	// 		}
	// 		ngo, d := resource_eks_cluster.NewNodeGroupsValue(resource_eks_cluster.NodeGroupsValue{}.AttributeTypes(ctx), ngv)
	// 		if d.HasError() {
	// 			resp.Diagnostics.Append(d...)
	// 			return
	// 		}
	// 		ngElements = append(ngElements, ngo)
	// 	}
	//
	// 	ngList, d = types.ListValue(resource_eks_cluster.NodeGroupsValue{}.Type(ctx), ngElements)
	// 	if d.HasError() {
	// 		resp.Diagnostics.Append(d...)
	// 		return
	// 	}
	//
	// }
	//
	// cc := map[string]attr.Value{
	// 	"apiversion": types.StringValue(clusterConfigSpec.APIVersion),
	// 	"kind":       types.StringValue(clusterConfigSpec.Kind),
	// 	"metadata":   fmd2,
	// }
	// if ngMapInUse {
	// 	cc["node_groups_map"] = ngMap
	// 	cc["node_groups"] = types.ListNull(resource_eks_cluster.NodeGroupsValue{}.Type(ctx))
	// } else {
	// 	cc["node_groups"] = ngList
	// 	cc["node_groups_map"] = types.MapNull(resource_eks_cluster.NodeGroupsMapValue{}.Type(ctx))
	// }
	//
	// fcc, d := resource_eks_cluster.NewClusterConfigValue(resource_eks_cluster.ClusterConfigValue{}.AttributeTypes(ctx), cc)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }
	// ccElements := []attr.Value{
	// 	fcc,
	// }
	// data.ClusterConfig, d = types.ListValue(resource_eks_cluster.ClusterConfigValue{}.Type(ctx), ccElements)
	// if d.HasError() {
	// 	resp.Diagnostics.Append(d...)
	// 	return
	// }

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *eksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	updatedCluster, d := resource_eks_cluster.ExpandEksCluster(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	updatedClusterConfig, d := resource_eks_cluster.ExpandEksClusterConfig(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	tflog.Debug(ctx, "updated value", map[string]any{"updatedCluster": updatedCluster, "updatedClusterConfig": updatedClusterConfig})

	// Call API to update EKS cluster
	clusterName := updatedCluster.Metadata.Name
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	if err := encoder.Encode(updatedCluster); err != nil {
		log.Printf("error encoding cluster: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}
	if err := encoder.Encode(updatedClusterConfig); err != nil {
		log.Printf("error encoding cluster config: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster config, got error: %s", err))
		return
	}
	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()
	_, err := clusterctl.Apply(logger, rctlConfig, clusterName, b.Bytes(), false, false, false, false, uaDef, "")
	if err != nil {
		log.Printf("cluster error 1: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *eksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from state
	clusterList := make([]resource_eks_cluster.ClusterValue, 0, len(data.Cluster.Elements()))
	d := data.Cluster.ElementsAs(ctx, &clusterList, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	mdList := make([]resource_eks_cluster.MetadataValue, 0, len(clusterList[0].Metadata.Elements()))
	d = clusterList[0].Metadata.ElementsAs(ctx, &mdList, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	mdO, d := mdList[0].ToObjectValue(ctx)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	var md resource_eks_cluster.MetadataType
	mdObj, d := md.ValueFromObject(ctx, mdO)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	mdValue, ok := mdObj.(resource_eks_cluster.MetadataValue)
	if !ok {
		resp.Diagnostics.AddError("Invalid Metadata", "Expected MetadataValue type but got a different type.")
		return
	}
	clusterName := mdValue.Name.ValueString()
	projectName := mdValue.Project.ValueString()
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get project ID from name '%s', got error: %s", projectName, err))
		return
	}

	// Delete API call logic
	err = cluster.DeleteCluster(clusterName, projectID, false, uaDef)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to delete cluster, got error: %s", err),
		)
	}
}
