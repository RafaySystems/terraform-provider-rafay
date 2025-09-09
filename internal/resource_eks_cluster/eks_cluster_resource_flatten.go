package resource_eks_cluster

import (
	"context"
	"log"
	"strings"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	jsoniter "github.com/json-iterator/go"
)

func FlattenEksCluster(ctx context.Context, in rafay.EKSCluster, data *EksClusterModel) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if data == nil {
		return diags
	}

	cv := NewClusterValueNull()
	d = cv.Flatten(ctx, in)
	diags = append(diags, d...)
	data.Cluster, d = types.ListValue(ClusterValue{}.Type(ctx), []attr.Value{cv})
	diags = append(diags, d...)

	return diags
}

func FlattenEksClusterConfig(ctx context.Context, in rafay.EKSClusterConfig, data *EksClusterModel) diag.Diagnostics {
	var diags, d diag.Diagnostics

	// get cluster config from state
	ccList := make([]ClusterConfigValue, len(data.ClusterConfig.Elements()))
	d = data.ClusterConfig.ElementsAs(ctx, &ccList, false)
	diags = append(diags, d...)

	if len(ccList) > 0 {
		cc := NewClusterConfigValueNull()
		d = cc.Flatten(ctx, in, ccList[0])
		diags = append(diags, d...)
		data.ClusterConfig, d = types.ListValue(ClusterConfigValue{}.Type(ctx), []attr.Value{cc})
		diags = append(diags, d...)
	} else {
		data.ClusterConfig = types.ListNull(ClusterConfigValue{}.Type(ctx))
	}

	return diags
}

func (v *ClusterValue) Flatten(ctx context.Context, in rafay.EKSCluster) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in.Kind != "" {
		v.Kind = types.StringValue(in.Kind)
	}

	if in.Metadata != nil {
		md := NewMetadataValueNull()
		d = md.Flatten(ctx, in.Metadata)
		diags = append(diags, d...)
		mdElements := []attr.Value{
			md,
		}
		v.Metadata, d = types.ListValue(MetadataValue{}.Type(ctx), mdElements)
		diags = append(diags, d...)
	} else {
		v.Metadata = types.ListNull(MetadataValue{}.Type(ctx))
	}

	spec := types.ListNull(SpecValue{}.Type(ctx))
	if in.Spec != nil {
		sp := NewSpecValueNull()
		d = sp.Flatten(ctx, in.Spec)
		diags = append(diags, d...)
		specElements := []attr.Value{
			sp,
		}
		spec, d = types.ListValue(SpecValue{}.Type(ctx), specElements)
		diags = append(diags, d...)
	}
	v.Spec = spec

	v.state = attr.ValueStateKnown
	return diags
}

func (v *MetadataValue) Flatten(ctx context.Context, in *rafay.EKSClusterMetadata) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if in.Name != "" {
		v.Name = types.StringValue(in.Name)
	}
	if in.Project != "" {
		v.Project = types.StringValue(in.Project)
	}

	lbsMap := types.MapNull(types.StringType)
	if len(in.Labels) != 0 {
		lbs := map[string]attr.Value{}
		for key, val := range in.Labels {
			lbs[key] = types.StringValue(val)
		}
		lbsMap, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels = lbsMap

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SpecValue) Flatten(ctx context.Context, in *rafay.EKSSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.Blueprint != "" {
		v.Blueprint = types.StringValue(in.Blueprint)
	} else {
		v.Blueprint = types.StringNull()
	}
	if in.BlueprintVersion != "" {
		v.BlueprintVersion = types.StringValue(in.BlueprintVersion)
	} else {
		v.BlueprintVersion = types.StringNull()
	}
	if in.CloudProvider != "" {
		v.CloudProvider = types.StringValue(in.CloudProvider)
	} else {
		v.CloudProvider = types.StringNull()
	}
	if in.CniProvider != "" {
		v.CniProvider = types.StringValue(in.CniProvider)
	} else {
		v.CniProvider = types.StringNull()
	}
	if in.CrossAccountRoleArn != "" {
		v.CrossAccountRoleArn = types.StringValue(in.CrossAccountRoleArn)
	} else {
		v.CrossAccountRoleArn = types.StringNull()
	}
	if in.Type != "" {
		v.SpecType = types.StringValue(in.Type)
	} else {
		v.SpecType = types.StringNull()
	}

	if in.CniParams != nil {
		cp := NewCniParamsValueNull()
		d = cp.Flatten(ctx, in.CniParams)
		diags = append(diags, d...)
		cpElements := []attr.Value{
			cp,
		}
		v.CniParams, d = types.ListValue(CniParamsValue{}.Type(ctx), cpElements)
		diags = append(diags, d...)
	}

	proxycfgMap := types.MapNull(types.StringType)
	if in.ProxyConfig != nil {
		pc := map[string]attr.Value{}
		if in.ProxyConfig.HttpProxy != "" {
			pc["http_proxy"] = types.StringValue(in.ProxyConfig.HttpProxy)
		}
		if in.ProxyConfig.HttpsProxy != "" {
			pc["https_proxy"] = types.StringValue(in.ProxyConfig.HttpsProxy)
		}
		if in.ProxyConfig.NoProxy != "" {
			pc["no_proxy"] = types.StringValue(in.ProxyConfig.NoProxy)
		}
		if in.ProxyConfig.ProxyAuth != "" {
			pc["proxy_auth"] = types.StringValue(in.ProxyConfig.ProxyAuth)
		}
		if in.ProxyConfig.BootstrapCA != "" {
			pc["bootstrap_ca"] = types.StringValue(in.ProxyConfig.BootstrapCA)
		}
		if in.ProxyConfig.Enabled {
			pc["enabled"] = types.StringValue("true")
		}
		if in.ProxyConfig.AllowInsecureBootstrap {
			pc["allow_insecure_bootstrap"] = types.StringValue("true")
		}
		proxycfgMap, d = types.MapValue(types.StringType, pc)
		diags = append(diags, d...)
	}
	v.ProxyConfig = proxycfgMap

	if in.SystemComponentsPlacement != nil {
		scp := NewSystemComponentsPlacementValueNull()
		d = scp.Flatten(ctx, in.SystemComponentsPlacement)
		diags = append(diags, d...)
		v.SystemComponentsPlacement, d = types.ListValue(SystemComponentsPlacementValue{}.Type(ctx), []attr.Value{scp})
		diags = append(diags, d...)
	}

	if in.Sharing != nil {
		sh := NewSharingValueNull()
		d = sh.Flatten(ctx, in.Sharing)
		diags = append(diags, d...)
		v.Sharing, d = types.ListValue(SharingValue{}.Type(ctx), []attr.Value{sh})
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CniParamsValue) Flatten(ctx context.Context, in *rafay.CustomCni) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.CustomCniCidr != "" {
		v.CustomCniCidr = types.StringValue(in.CustomCniCidr)
	}

	customCniCrdSpec := types.ListNull(CustomCniCrdSpecValue{}.Type(ctx))
	if len(in.CustomCniCrdSpec) > 0 {
		csElements := make([]attr.Value, 0, len(in.CustomCniCrdSpec))
		for name, cniSpec := range in.CustomCniCrdSpec {
			cs := NewCustomCniCrdSpecValueNull()
			d = cs.Flatten(ctx, name, cniSpec)
			diags = append(diags, d...)
			csElements = append(csElements, cs)
		}
		customCniCrdSpec, d = types.ListValue(CustomCniCrdSpecValue{}.Type(ctx), csElements)
		diags = append(diags, d...)
	}
	v.CustomCniCrdSpec = customCniCrdSpec

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CustomCniCrdSpecValue) Flatten(ctx context.Context, name string, in []rafay.CustomCniSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(name)

	cniSpec := types.ListNull(CniSpecValue{}.Type(ctx))
	if len(in) > 0 {
		specElements := make([]attr.Value, 0, len(in))
		for _, spec := range in {
			s := NewCniSpecValueNull()
			d = s.Flatten(ctx, spec)
			diags = append(diags, d...)
			specElements = append(specElements, s)
		}
		cniSpec, d = types.ListValue(CniSpecValue{}.Type(ctx), specElements)
		diags = append(diags, d...)
	}
	v.CniSpec = cniSpec

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CniSpecValue) Flatten(ctx context.Context, in rafay.CustomCniSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if in.Subnet != "" {
		v.Subnet = types.StringValue(in.Subnet)
	}

	securityGroups := types.ListNull(types.StringType)
	if len(in.SecurityGroups) > 0 {
		sgElements := make([]attr.Value, 0, len(in.SecurityGroups))
		for _, sg := range in.SecurityGroups {
			sgElements = append(sgElements, types.StringValue(sg))
		}
		securityGroups, d = types.ListValue(types.StringType, sgElements)
		diags = append(diags, d...)
	}
	v.SecurityGroups = securityGroups

	v.state = attr.ValueStateKnown
	return diags
}

// --- Cluster config ---

func (v *ClusterConfigValue) Flatten(ctx context.Context, in rafay.EKSClusterConfig, state ClusterConfigValue) diag.Diagnostics {
	var diags, d diag.Diagnostics

	// get existing addons from TF state
	var stAddonNames []string
	stAddons := make([]AddonsValue, 0, len(state.Addons.Elements()))
	diags = state.Addons.ElementsAs(ctx, &stAddons, false)
	diags = append(diags, d...)
	for _, addon := range stAddons {
		stAddonNames = append(stAddonNames, getStringValue(addon.Name))
	}

	// node groups list or map?
	isNodeGroupsMap := false
	if !state.NodeGroupsMap.IsNull() && !state.NodeGroupsMap.IsUnknown() &&
		len(state.NodeGroupsMap.Elements()) > 0 {
		isNodeGroupsMap = true
	}

	// managed node groups list or map?
	isManagedNodeGroupsMap := false
	if !state.ManagedNodegroupsMap.IsNull() && !state.ManagedNodegroupsMap.IsUnknown() &&
		len(state.ManagedNodegroupsMap.Elements()) > 0 {
		isManagedNodeGroupsMap = true
	}

	if in.APIVersion != "" {
		v.Apiversion = types.StringValue(in.APIVersion)
	}
	if in.Kind != "" {
		v.Kind = types.StringValue(in.Kind)
	}

	availabilityZones := types.ListNull(types.StringType)
	if len(in.AvailabilityZones) > 0 {
		azElements := []attr.Value{}
		for _, az := range in.AvailabilityZones {
			azElements = append(azElements, types.StringValue(az))
		}
		availabilityZones, d = types.ListValue(types.StringType, azElements)
		diags = append(diags, d...)
	}
	v.AvailabilityZones = availabilityZones

	if in.Metadata != nil {
		md := NewMetadata2ValueNull()
		d = md.Flatten(ctx, in.Metadata)
		diags = append(diags, d...)
		mdElements := []attr.Value{
			md,
		}
		v.Metadata2, d = types.ListValue(Metadata2Value{}.Type(ctx), mdElements)
		diags = append(diags, d...)
	}

	// node groups
	if len(in.NodeGroups) > 0 {
		if isNodeGroupsMap {
			stNgMaps := make(map[string]NodeGroupsMapValue, len(state.NodeGroupsMap.Elements()))
			d = state.NodeGroupsMap.ElementsAs(ctx, &stNgMaps, false)
			diags = append(diags, d...)

			nodegrp := map[string]attr.Value{}
			for _, ng := range in.NodeGroups {
				stNgMap := NodeGroupsMapValue{}
				if _, ok := stNgMaps[ng.Name]; ok {
					stNgMap = stNgMaps[ng.Name]
				}

				ngrp := NewNodeGroupsMapValueNull()
				d = ngrp.Flatten(ctx, ng, stNgMap)
				diags = append(diags, d...)
				nodegrp[ng.Name] = ngrp
			}
			v.NodeGroupsMap, d = types.MapValue(NodeGroupsMapValue{}.Type(ctx), nodegrp)
			diags = append(diags, d...)

			v.NodeGroups = types.ListNull(NodeGroupsValue{}.Type(ctx))
		} else {
			stNgs := make([]NodeGroupsValue, 0, len(state.NodeGroups.Elements()))
			d = state.NodeGroups.ElementsAs(ctx, &stNgs, false)
			diags = append(diags, d...)

			ngElements := []attr.Value{}
			for _, ng := range in.NodeGroups {
				stNg := NodeGroupsValue{}
				for _, sng := range stNgs {
					if strings.EqualFold(getStringValue(sng.Name), ng.Name) {
						stNg = sng
					}
				}

				ngList := NewNodeGroupsValueNull()
				d = ngList.Flatten(ctx, ng, stNg)
				diags = append(diags, d...)
				ngElements = append(ngElements, ngList)
			}
			v.NodeGroups, d = types.ListValue(NodeGroupsValue{}.Type(ctx), ngElements)
			diags = append(diags, d...)
			v.NodeGroupsMap = types.MapNull(NodeGroupsMapValue{}.Type(ctx))
		}
	}

	if in.KubernetesNetworkConfig != nil {
		netconf := NewKubernetesNetworkConfigValueNull()
		d = netconf.Flatten(ctx, in.KubernetesNetworkConfig)
		diags = append(diags, d...)
		v.KubernetesNetworkConfig, d = types.ListValue(KubernetesNetworkConfigValue{}.Type(ctx), []attr.Value{netconf})
		diags = append(diags, d...)
	}

	if in.IAM != nil {
		iam := NewIam3ValueNull()
		d = iam.Flatten(ctx, in.IAM)
		diags = append(diags, d...)
		v.Iam3, d = types.ListValue(Iam3Value{}.Type(ctx), []attr.Value{iam})
		diags = append(diags, d...)
	}

	identityProviders := types.ListNull(IdentityProvidersValue{}.Type(ctx))
	if len(in.IdentityProviders) > 0 {
		identityProvidersList := []attr.Value{}
		for _, identityProvider := range in.IdentityProviders {
			ip := NewIdentityProvidersValueNull()
			d = ip.Flatten(ctx, identityProvider)
			diags = append(diags, d...)
			identityProvidersList = append(identityProvidersList, ip)
		}
		identityProviders, d = types.ListValue(IdentityProvidersValue{}.Type(ctx), identityProvidersList)
		diags = append(diags, d...)
	}
	v.IdentityProviders = identityProviders

	if in.VPC != nil {
		vpc := NewVpcValueNull()
		d = vpc.Flatten(ctx, in.VPC)
		diags = append(diags, d...)
		v.Vpc, d = types.ListValue(VpcValue{}.Type(ctx), []attr.Value{vpc})
		diags = append(diags, d...)
	}

	addons := types.ListNull(AddonsValue{}.Type(ctx))
	if len(in.Addons) > 0 {
		addonsList := []attr.Value{}
		for _, add := range in.Addons {
			// Hack: Check if addon exists in the state. Addons (CoreDNS, Kube-Proxy, VPC CNI, EBS CSI Driver)
			// are auto created by EKS Cluster.
			exists := false
			for _, e := range stAddonNames {
				if strings.EqualFold(e, add.Name) {
					exists = true
				}
			}

			if exists {
				stAddon := AddonsValue{}
				for _, stAdd := range stAddons {
					if strings.EqualFold(getStringValue(stAdd.Name), add.Name) {
						stAddon = stAdd
					}
				}
				addon := NewAddonsValueNull()
				d = addon.Flatten(ctx, add, stAddon)
				diags = append(diags, d...)
				addonsList = append(addonsList, addon)
			}
		}
		addons, d = types.ListValue(AddonsValue{}.Type(ctx), addonsList)
		diags = append(diags, d...)
	}
	v.Addons = addons

	if in.PrivateCluster != nil {
		privateCluster := NewPrivateClusterValueNull()
		d = privateCluster.Flatten(ctx, in.PrivateCluster)
		diags = append(diags, d...)
		v.PrivateCluster, d = types.ListValue(PrivateClusterValue{}.Type(ctx), []attr.Value{privateCluster})
		diags = append(diags, d...)
	}

	// managed node groups
	if len(in.ManagedNodeGroups) > 0 {
		if isManagedNodeGroupsMap {
			stMngMaps := make(map[string]ManagedNodegroupsMapValue, len(state.ManagedNodegroupsMap.Elements()))
			d = state.ManagedNodegroupsMap.ElementsAs(ctx, &stMngMaps, false)
			diags = append(diags, d...)

			managednodegrp := map[string]attr.Value{}
			for _, mng := range in.ManagedNodeGroups {
				stMngMap := ManagedNodegroupsMapValue{}
				if _, ok := stMngMaps[mng.Name]; ok {
					stMngMap = stMngMaps[mng.Name]
				}

				mngm := NewManagedNodegroupsMapValueNull()
				d = mngm.Flatten(ctx, mng, stMngMap)
				diags = append(diags, d...)
				managednodegrp[mng.Name] = mngm
			}
			v.ManagedNodegroupsMap, d = types.MapValue(ManagedNodegroupsMapValue{}.Type(ctx), managednodegrp)
			diags = append(diags, d...)
			v.ManagedNodegroups = types.ListNull(ManagedNodegroupsValue{}.Type(ctx))
		} else {
			stMngs := make([]ManagedNodegroupsValue, 0, len(state.ManagedNodegroups.Elements()))
			d = state.ManagedNodegroups.ElementsAs(ctx, &stMngs, false)
			diags = append(diags, d...)

			mngElements := []attr.Value{}
			for _, mng := range in.ManagedNodeGroups {
				stMng := ManagedNodegroupsValue{}
				for _, smng := range stMngs {
					if strings.EqualFold(getStringValue(smng.Name), mng.Name) {
						stMng = smng
					}
				}

				mngList := NewManagedNodegroupsValueNull()
				d = mngList.Flatten(ctx, mng, stMng)
				diags = append(diags, d...)
				mngElements = append(mngElements, mngList)
			}
			v.ManagedNodegroups, d = types.ListValue(ManagedNodegroupsValue{}.Type(ctx), mngElements)
			diags = append(diags, d...)
			v.ManagedNodegroupsMap = types.MapNull(ManagedNodegroupsMapValue{}.Type(ctx))
		}
	}

	fargateProfiles := types.ListNull(FargateProfilesValue{}.Type(ctx))
	if len(in.FargateProfiles) > 0 {
		fargateProfilesList := []attr.Value{}
		for _, fargateProfile := range in.FargateProfiles {
			fp := NewFargateProfilesValueNull()
			d = fp.Flatten(ctx, fargateProfile)
			diags = append(diags, d...)
			fargateProfilesList = append(fargateProfilesList, fp.Name)
		}
		fargateProfiles, d = types.ListValue(FargateProfilesValue{}.Type(ctx), fargateProfilesList)
		diags = append(diags, d...)
	}
	v.FargateProfiles = fargateProfiles

	if in.CloudWatch != nil {
		cloudWatch := NewCloudWatchValueNull()
		d = cloudWatch.Flatten(ctx, in.CloudWatch)
		diags = append(diags, d...)
		v.CloudWatch, d = types.ListValue(CloudWatchValue{}.Type(ctx), []attr.Value{cloudWatch})
		diags = append(diags, d...)
	}

	if in.SecretsEncryption != nil {
		SecretsEncryption := NewSecretsEncryptionValueNull()
		d = SecretsEncryption.Flatten(ctx, in.SecretsEncryption)
		diags = append(diags, d...)
		v.SecretsEncryption, d = types.ListValue(SecretsEncryptionValue{}.Type(ctx), []attr.Value{SecretsEncryption})
		diags = append(diags, d...)
	}

	if in.IdentityMappings != nil {
		identityMappings := NewIdentityMappingsValueNull()
		d = identityMappings.Flatten(ctx, in.IdentityMappings)
		diags = append(diags, d...)
		v.IdentityMappings, d = types.ListValue(IdentityMappingsValue{}.Type(ctx), []attr.Value{identityMappings})
		diags = append(diags, d...)
	}

	if in.AccessConfig != nil {
		accessConfig := NewAccessConfigValueNull()
		d = accessConfig.Flatten(ctx, in.AccessConfig)
		diags = append(diags, d...)
		v.AccessConfig, d = types.ListValue(AccessConfigValue{}.Type(ctx), []attr.Value{accessConfig})
		diags = append(diags, d...)
	}

	if in.AddonsConfig != nil {
		addonsConfig := NewAddonsConfigValueNull()
		d = addonsConfig.Flatten(ctx, in.AddonsConfig)
		diags = append(diags, d...)
		v.AddonsConfig, d = types.ListValue(AddonsConfigValue{}.Type(ctx), []attr.Value{addonsConfig})
		diags = append(diags, d...)
	}

	if in.AutoModeConfig != nil {
		autoModeConfig := NewAutoModeConfigValueNull()
		d = autoModeConfig.Flatten(ctx, in.AutoModeConfig)
		diags = append(diags, d...)
		v.AutoModeConfig, d = types.ListValue(AutoModeConfigValue{}.Type(ctx), []attr.Value{autoModeConfig})
		diags = append(diags, d...)
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AutoModeConfigValue) Flatten(ctx context.Context, in *rafay.EKSAutoModeConfig) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Enabled = types.BoolValue(in.Enabled)
	v.NodeRoleArn = types.StringValue(in.NodeRoleARN)
	nodePools := types.ListNull(types.StringType)
	if len(in.NodePools) > 0 {
		nodepools := []attr.Value{}
		for _, nodepool := range in.NodePools {
			nodepools = append(nodepools, types.StringValue(nodepool))
		}
		nodePools, d = types.ListValue(types.StringType, nodepools)
		diags = append(diags, d...)
	}
	v.NodePools = nodePools

	return diags
}

func (v *AddonsConfigValue) Flatten(ctx context.Context, in *rafay.EKSAddonsConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.AutoApplyPodIdentityAssociations = types.BoolValue(in.AutoApplyPodIdentityAssociations)
	v.DisableEbsCsiDriver = types.BoolValue(in.DisableEBSCSIDriver)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AccessConfigValue) Flatten(ctx context.Context, in *rafay.EKSClusterAccess) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.BootstrapClusterCreatorAdminPermissions = types.BoolPointerValue(&in.BootstrapClusterCreatorAdminPermissions)
	v.AuthenticationMode = types.StringValue(in.AuthenticationMode)

	accessEntries := types.ListNull(AccessEntriesValue{}.Type(ctx))
	if len(in.AccessEntries) > 0 {
		accessEntriesList := []attr.Value{}
		for _, accessEntry := range in.AccessEntries {
			ae := NewAccessEntriesValueNull()
			d = ae.Flatten(ctx, accessEntry)
			diags = append(diags, d...)
			accessEntriesList = append(accessEntriesList, ae)
		}
		accessEntries, d = types.ListValue(AccessEntriesValue{}.Type(ctx), accessEntriesList)
		diags = append(diags, d...)
	}
	v.AccessEntries = accessEntries

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AccessEntriesValue) Flatten(ctx context.Context, in *rafay.EKSAccessEntry) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.PrincipalArn = types.StringValue(in.PrincipalARN)
	v.AccessEntriesType = types.StringValue(in.Type)
	v.KubernetesUsername = types.StringValue(in.KubernetesUsername)
	kubernetesGroups := types.ListNull(types.StringType)
	if len(in.KubernetesGroups) > 0 {
		groups := []attr.Value{}
		for _, group := range in.KubernetesGroups {
			groups = append(groups, types.StringValue(group))
		}
		kubernetesGroups, d = types.ListValue(types.StringType, groups)
		diags = append(diags, d...)
	}
	v.KubernetesGroups = kubernetesGroups

	tags := types.MapNull(types.StringType)
	if len(in.Tags) > 0 {
		tgs := make(map[string]attr.Value, len(in.Tags))
		for key, val := range in.Tags {
			tgs[key] = types.StringValue(val)
		}
		tags, d = types.MapValue(types.StringType, tgs)
		diags = append(diags, d...)
	}
	v.Tags = tags

	accessPolicies := types.ListNull(AccessPoliciesValue{}.Type(ctx))
	if len(in.AccessPolicies) > 0 {
		accessPoliciesList := []attr.Value{}
		for _, accessPolicy := range in.AccessPolicies {
			ap := NewAccessPoliciesValueNull()
			d = ap.Flatten(ctx, accessPolicy)
			diags = append(diags, d...)
			accessPoliciesList = append(accessPoliciesList, ap)
		}
		accessPolicies, d = types.ListValue(AccessPoliciesValue{}.Type(ctx), accessPoliciesList)
		diags = append(diags, d...)
	}
	v.AccessPolicies = accessPolicies

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AccessPoliciesValue) Flatten(ctx context.Context, in *rafay.EKSAccessPolicy) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.PolicyArn = types.StringValue(in.PolicyARN)

	accessScope := NewAccessScopeValueNull()
	d = accessScope.Flatten(ctx, in.AccessScope)
	diags = append(diags, d...)
	v.AccessScope, d = types.ListValue(AccessScopeValue{}.Type(ctx), []attr.Value{accessScope})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AccessScopeValue) Flatten(ctx context.Context, in *rafay.EKSAccessScope) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.AccessScopeType = types.StringValue(in.Type)
	namespaces := types.ListNull(types.StringType)
	if len(in.Namespaces) > 0 {
		namespacesList := []attr.Value{}
		for _, namespace := range in.Namespaces {
			namespacesList = append(namespacesList, types.StringValue(namespace))
		}
		namespaces, d = types.ListValue(types.StringType, namespacesList)
		diags = append(diags, d...)
	}
	v.Namespaces = namespaces

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IdentityMappingsValue) Flatten(ctx context.Context, in *rafay.EKSClusterIdentityMappings) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	accounts := types.ListNull(types.StringType)
	if len(in.Accounts) > 0 {
		accountsList := []attr.Value{}
		for _, account := range in.Accounts {
			accountsList = append(accountsList, types.StringValue(account))
		}
		accounts, d = types.ListValue(types.StringType, accountsList)
		diags = append(diags, d...)
	}
	v.Accounts = accounts

	arns := types.ListNull(ArnsValue{}.Type(ctx))
	if len(in.Arns) > 0 {
		arnElements := []attr.Value{}
		for _, arn := range in.Arns {
			arnsValue := NewArnsValueNull()
			d = arnsValue.Flatten(ctx, arn)
			diags = append(diags, d...)
			arnElements = append(arnElements, arnsValue)
		}
		arns, d = types.ListValue(ArnsValue{}.Type(ctx), arnElements)
		diags = append(diags, d...)
	}
	v.Arns = arns

	v.state = attr.ValueStateKnown
	return diags
}

func (v *ArnsValue) Flatten(ctx context.Context, in *rafay.IdentityMappingARN) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Arn = types.StringValue(in.Arn)
	v.Username = types.StringValue(in.Username)
	group := types.ListNull(types.StringType)
	if len(in.Group) > 0 {
		groups := []attr.Value{}
		for _, groupItem := range in.Group {
			groups = append(groups, types.StringValue(groupItem))
		}
		group, d = types.ListValue(types.StringType, groups)
		diags = append(diags, d...)
	}
	v.Group = group

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SecretsEncryptionValue) Flatten(ctx context.Context, in *rafay.SecretsEncryption) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.KeyArn = types.StringValue(in.KeyARN)
	v.EncryptExistingSecrets = types.BoolPointerValue(in.EncryptExistingSecrets)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CloudWatchValue) Flatten(ctx context.Context, in *rafay.EKSClusterCloudWatch) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	cloudLogging := NewCloudLoggingValueNull()
	d = cloudLogging.Flatten(ctx, in.ClusterLogging)
	diags = append(diags, d...)
	v.CloudLogging, d = types.ListValue(CloudLoggingValue{}.Type(ctx), []attr.Value{cloudLogging})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CloudLoggingValue) Flatten(ctx context.Context, in *rafay.EKSClusterCloudWatchLogging) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	enableTypes := types.ListNull(types.StringType)
	if len(in.EnableTypes) > 0 {
		enableTypesList := []attr.Value{}
		for _, enableType := range in.EnableTypes {
			enableTypesList = append(enableTypesList, types.StringValue(enableType))
		}
		enableTypes, d = types.ListValue(types.StringType, enableTypesList)
		diags = append(diags, d...)
	}
	v.EnableTypes = enableTypes

	v.LogRetentionInDays = types.Int64Value(int64(in.LogRetentionInDays))

	v.state = attr.ValueStateKnown
	return diags
}

func (v *FargateProfilesValue) Flatten(ctx context.Context, in *rafay.FargateProfile) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Name = types.StringValue(in.Name)
	v.PodExecutionRoleArn = types.StringValue(in.PodExecutionRoleARN)
	subnets := types.ListNull(types.StringType)
	if len(in.Subnets) > 0 {
		subnetsList := []attr.Value{}
		for _, subnet := range in.Subnets {
			subnetsList = append(subnetsList, types.StringValue(subnet))
		}
		subnets, d = types.ListValue(types.StringType, subnetsList)
		diags = append(diags, d...)
	}
	v.Subnets = subnets
	tags := types.MapNull(types.StringType)
	if len(in.Tags) > 0 {
		tgs := make(map[string]attr.Value, len(in.Tags))
		for key, val := range in.Tags {
			tgs[key] = types.StringValue(val)
		}
		tags, d = types.MapValue(types.StringType, tgs)
		diags = append(diags, d...)
	}
	v.Tags = tags
	v.Status = types.StringValue(in.Status)

	selectors := types.ListNull(SelectorsValue{}.Type(ctx))
	if len(in.Selectors) > 0 {
		selectorsList := []attr.Value{}
		for _, sel := range in.Selectors {
			selectorsValue := NewSelectorsValueNull()
			d = selectorsValue.Flatten(ctx, sel)
			diags = append(diags, d...)
			selectorsList = append(selectorsList, selectorsValue)
		}
		selectors, d = types.ListValue(SelectorsValue{}.Type(ctx), selectorsList)
		diags = append(diags, d...)
	}
	v.Selectors = selectors

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SelectorsValue) Flatten(ctx context.Context, in rafay.FargateProfileSelector) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Namespace = types.StringValue(in.Namespace)
	labels := map[string]attr.Value{}
	for key, val := range in.Labels {
		labels[key] = types.StringValue(val)
	}
	v.Labels, d = types.MapValue(types.StringType, labels)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PrivateClusterValue) Flatten(ctx context.Context, in *rafay.PrivateCluster) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Enabled = types.BoolPointerValue(in.Enabled)
	v.SkipEndpointCreation = types.BoolPointerValue(in.SkipEndpointCreation)
	additionalEndpointServices := types.ListNull(types.StringType)
	if len(in.AdditionalEndpointServices) > 0 {
		additionalEndpointServicesList := []attr.Value{}
		for _, val := range in.AdditionalEndpointServices {
			additionalEndpointServicesList = append(additionalEndpointServicesList, types.StringValue(val))
		}
		additionalEndpointServices, d = types.ListValue(types.StringType, additionalEndpointServicesList)
		diags = append(diags, d...)
	}
	v.AdditionalEndpointServices = additionalEndpointServices

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AddonsValue) Flatten(ctx context.Context, in *rafay.Addon, state AddonsValue) diag.Diagnostics {
	var diags, d diag.Diagnostics

	var isPolicyV1, isPolicyV2 bool
	if !state.AttachPolicyV22.IsNull() && !state.AttachPolicyV22.IsUnknown() &&
		getStringValue(state.AttachPolicyV22) != "" {
		isPolicyV2 = true
	}
	if !state.AttachPolicy3.IsNull() && !state.AttachPolicy3.IsUnknown() &&
		len(state.AttachPolicy3.Elements()) > 0 {
		isPolicyV1 = true
	}

	v.Name = types.StringValue(in.Name)
	v.Version = types.StringValue(in.Version)
	v.ServiceAccountRoleArn = types.StringValue(in.ServiceAccountRoleARN)
	attachPolicyArns3 := types.ListNull(types.StringType)
	if len(in.AttachPolicyARNs) > 0 {
		policyARNs := []attr.Value{}
		for _, arn := range in.AttachPolicyARNs {
			policyARNs = append(policyARNs, types.StringValue(arn))
		}
		attachPolicyArns3, d = types.ListValue(types.StringType, policyARNs)
		diags = append(diags, d...)
	}
	v.AttachPolicyArns3 = attachPolicyArns3

	// Rafay supports two formats for AttachPolicy
	if in.AttachPolicy != nil {
		if isPolicyV1 && !isPolicyV2 {
			attachPolicy := NewAttachPolicy3ValueNull()
			d = attachPolicy.Flatten(ctx, in.AttachPolicy)
			diags = append(diags, d...)
			v.AttachPolicy3, d = types.ListValue(AttachPolicy3Value{}.Type(ctx), []attr.Value{attachPolicy})
			diags = append(diags, d...)
		} else {
			var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
			jsonBytes, err := json2.Marshal(in.AttachPolicy)
			if err != nil {
				diags.AddError("Attach Policy Marshal Error", err.Error())
			}
			v.AttachPolicyV22 = types.StringValue(string(jsonBytes))
		}
	}

	v.PermissionsBoundary2 = types.StringValue(in.PermissionsBoundary)
	tags := types.MapNull(types.StringType)
	if len(in.Tags) > 0 {
		tgs := map[string]attr.Value{}
		for key, val := range in.Tags {
			tgs[key] = types.StringValue(val)
		}
		tags, d = types.MapValue(types.StringType, tgs)
		diags = append(diags, d...)
	}
	v.Tags4 = tags

	v.ConfigurationValues = types.StringValue(in.ConfigurationValues)
	v.UseDefaultPodIdentityAssociations = types.BoolValue(in.UseDefaultPodIdentityAssociations)

	policies := NewWellKnownPolicies3ValueNull()
	d = policies.Flatten(ctx, in.WellKnownPolicies)
	diags = append(diags, d...)
	v.WellKnownPolicies3, d = types.ListValue(WellKnownPolicies3Value{}.Type(ctx), []attr.Value{policies})
	diags = append(diags, d...)

	podIdentityAssociations2 := types.ListNull(PodIdentityAssociations2Value{}.Type(ctx))
	if len(in.PodIdentityAssociations) > 0 {
		pias := []attr.Value{}
		for _, p := range in.PodIdentityAssociations {
			pia := NewPodIdentityAssociations2ValueNull()
			d = pia.Flatten(ctx, p)
			diags = append(diags, d...)
			pias = append(pias, pia)
		}
		podIdentityAssociations2, d = types.ListValue(PodIdentityAssociations2Value{}.Type(ctx), pias)
		diags = append(diags, d...)
	}
	v.PodIdentityAssociations2 = podIdentityAssociations2

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PodIdentityAssociations2Value) Flatten(ctx context.Context, in *rafay.IAMPodIdentityAssociation) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Namespace = types.StringValue(in.Namespace)
	v.ServiceAccountName = types.StringValue(in.ServiceAccountName)
	v.RoleArn = types.StringValue(in.RoleARN)
	v.CreateServiceAccount = types.BoolValue(in.CreateServiceAccount)
	v.RoleName = types.StringValue(in.RoleName)
	v.PermissionBoundaryArn = types.StringValue(in.PermissionsBoundaryARN)
	permissionPolicyArns := types.ListNull(types.StringType)
	if len(in.PermissionPolicyARNs) > 0 {
		arns := []attr.Value{}
		for _, arn := range in.PermissionPolicyARNs {
			arns = append(arns, types.StringValue(arn))
		}
		permissionPolicyArns, d = types.ListValue(types.StringType, arns)
		diags = append(diags, d...)
	}
	v.PermissionPolicyArns = permissionPolicyArns
	if len(in.PermissionPolicy) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.PermissionPolicy)
		if err != nil {
			diags.AddError("Permission policy marshal error", err.Error())
		}
		v.PermissionPolicy = types.StringValue(string(jsonStr))
	}
	tagMap := types.MapNull(types.StringType)
	if len(in.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range in.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags = tagMap

	policies := NewWellKnownPolicies4ValueNull()
	d = policies.Flatten(ctx, in.WellKnownPolicies)
	diags = append(diags, d...)
	v.WellKnownPolicies4, d = types.ListValue(WellKnownPolicies4Value{}.Type(ctx), []attr.Value{policies})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *WellKnownPolicies4Value) Flatten(ctx context.Context, in *rafay.WellKnownPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.ImageBuilder = types.BoolPointerValue(in.ImageBuilder)
	v.AwsLoadBalancerController = types.BoolPointerValue(in.AWSLoadBalancerController)
	v.ExternalDns = types.BoolPointerValue(in.ExternalDNS)
	v.CertManager = types.BoolPointerValue(in.CertManager)
	v.EbsCsiController = types.BoolPointerValue(in.EBSCSIController)
	v.EfsCsiController = types.BoolPointerValue(in.EFSCSIController)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *WellKnownPolicies3Value) Flatten(ctx context.Context, in *rafay.WellKnownPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.ImageBuilder = types.BoolPointerValue(in.ImageBuilder)
	v.AwsLoadBalancerController = types.BoolPointerValue(in.AWSLoadBalancerController)
	v.ExternalDns = types.BoolPointerValue(in.ExternalDNS)
	v.CertManager = types.BoolPointerValue(in.CertManager)
	v.EbsCsiController = types.BoolPointerValue(in.EBSCSIController)
	v.EfsCsiController = types.BoolPointerValue(in.EFSCSIController)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AttachPolicy3Value) Flatten(ctx context.Context, attachpol *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if attachpol == nil {
		return diags
	}

	v.Version = types.StringValue(attachpol.Version)
	v.Id = types.StringValue(attachpol.Id)

	statement2 := types.ListNull(StatementValue{}.Type(ctx))
	if len(attachpol.Statement) > 0 {
		stms := []attr.Value{}
		for _, stm := range attachpol.Statement {
			sv := NewStatementValueNull()
			d = sv.Flatten(ctx, stm)
			diags = append(diags, d...)
			stms = append(stms, sv)
		}
		statement2, d = types.ListValue(StatementValue{}.Type(ctx), stms)
		diags = append(diags, d...)
	}
	v.Statement2 = statement2

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Statement2Value) Flatten(ctx context.Context, in rafay.InlineStatement) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if len(in.Effect) > 0 {
		v.Effect = types.StringValue(in.Effect)
	}
	if len(in.Sid) > 0 {
		v.Sid = types.StringValue(in.Sid)
	}
	if in.Action != nil && len(in.Action.([]interface{})) > 0 {
		actEle := []attr.Value{}
		for _, act := range in.Action.([]interface{}) {
			actEle = append(actEle, types.StringValue(act.(string)))
		}
		v.Action, d = types.ListValue(types.StringType, actEle)
		diags = append(diags, d...)
	}
	if in.NotAction != nil && len(in.NotAction.([]interface{})) > 0 {
		naEle := []attr.Value{}
		for _, na := range in.NotAction.([]interface{}) {
			naEle = append(naEle, types.StringValue(na.(string)))
		}
		v.NotAction, d = types.ListValue(types.StringType, naEle)
		diags = append(diags, d...)
	}
	if len(in.Resource.(string)) > 0 {
		v.Resource = types.StringValue(in.Resource.(string))
	}
	if in.NotResource != nil && len(in.NotResource.([]interface{})) > 0 {
		nrEle := []attr.Value{}
		for _, nr := range in.NotResource.([]interface{}) {
			nrEle = append(nrEle, types.StringValue(nr.(string)))
		}
		v.NotResource, d = types.ListValue(types.StringType, nrEle)
		diags = append(diags, d...)
	}

	if len(in.Condition) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Condition)
		if err != nil {
			diags.AddError("condition marshal error", err.Error())
		}
		v.Condition = types.StringValue(string(jsonStr))
	}

	if len(in.Principal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Principal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.Principal = types.StringValue(string(jsonStr))

	}

	if len(in.NotPrincipal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.NotPrincipal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.NotPrincipal = types.StringValue(string(jsonStr))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *VpcValue) Flatten(ctx context.Context, in *rafay.EKSClusterVPC) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Id = types.StringValue(in.ID)
	v.Cidr = types.StringValue(in.CIDR)
	v.Ipv6Cidr = types.StringValue(in.IPv6Cidr)
	v.Ipv6Pool = types.StringValue(in.IPv6Pool)
	v.SecurityGroup = types.StringValue(in.SecurityGroup)
	extraCidrs := types.ListNull(types.StringType)
	if len(in.ExtraCIDRs) > 0 {
		extraCidrsList := []attr.Value{}
		for _, cidr := range in.ExtraCIDRs {
			extraCidrsList = append(extraCidrsList, types.StringValue(cidr))
		}
		extraCidrs, d = types.ListValue(types.StringType, extraCidrsList)
		diags = append(diags, d...)
	}
	v.ExtraCidrs = extraCidrs

	extraIPv6Cidrs := types.ListNull(types.StringType)
	if len(in.ExtraIPv6CIDRs) > 0 {
		extraIPv6CidrsList := []attr.Value{}
		for _, cidr := range in.ExtraIPv6CIDRs {
			extraIPv6CidrsList = append(extraIPv6CidrsList, types.StringValue(cidr))
		}
		extraIPv6Cidrs, d = types.ListValue(types.StringType, extraIPv6CidrsList)
		diags = append(diags, d...)
	}
	v.ExtraIpv6Cidrs = extraIPv6Cidrs

	v.SharedNodeSecurityGroup = types.StringValue(in.SharedNodeSecurityGroup)
	v.ManageSharedNodeSecurityGroupRules = types.BoolPointerValue(in.ManageSharedNodeSecurityGroupRules)
	v.AutoAllocateIpv6 = types.BoolPointerValue(in.AutoAllocateIPv6)

	publicAccessCidrs := types.ListNull(types.StringType)
	if len(in.PublicAccessCIDRs) > 0 {
		publicAccessCidrsList := []attr.Value{}
		for _, cidr := range in.PublicAccessCIDRs {
			publicAccessCidrsList = append(publicAccessCidrsList, types.StringValue(cidr))
		}
		publicAccessCidrs, d = types.ListValue(types.StringType, publicAccessCidrsList)
		diags = append(diags, d...)
	}
	v.PublicAccessCidrs = publicAccessCidrs

	subnets := NewSubnets3ValueNull()
	d = subnets.Flatten(ctx, in.Subnets)
	diags = append(diags, d...)
	v.Subnets3, d = types.ListValue(Subnets3Value{}.Type(ctx), []attr.Value{subnets})
	diags = append(diags, d...)

	nat := NewNatValueNull()
	d = nat.Flatten(ctx, in.NAT)
	diags = append(diags, d...)
	v.Nat, d = types.ListValue(NatValue{}.Type(ctx), []attr.Value{nat})
	diags = append(diags, d...)

	clusterEndpoints := NewClusterEndpointsValueNull()
	d = clusterEndpoints.Flatten(ctx, in.ClusterEndpoints)
	diags = append(diags, d...)
	v.ClusterEndpoints, d = types.ListValue(ClusterEndpointsValue{}.Type(ctx), []attr.Value{clusterEndpoints})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *ClusterEndpointsValue) Flatten(ctx context.Context, in *rafay.ClusterEndpoints) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.PrivateAccess = types.BoolPointerValue(in.PrivateAccess)
	v.PublicAccess = types.BoolPointerValue(in.PublicAccess)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *NatValue) Flatten(ctx context.Context, in *rafay.ClusterNAT) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Gateway = types.StringValue(in.Gateway)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Subnets3Value) Flatten(ctx context.Context, in *rafay.ClusterSubnets) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	private := types.ListNull(PrivateValue{}.Type(ctx))
	if len(in.Private) > 0 {
		privates := []attr.Value{}
		for nm, pvSubnet := range in.Private {
			privateValue := NewPrivateValueNull()
			d = privateValue.Flatten(ctx, nm, pvSubnet)
			diags = append(diags, d...)
			privates = append(privates, privateValue)
		}
		private, d = types.ListValue(PrivateValue{}.Type(ctx), privates)
		diags = append(diags, d...)
	}
	v.Private = private

	public := types.ListNull(PublicValue{}.Type(ctx))
	if len(in.Public) > 0 {
		publics := []attr.Value{}
		for nm, puSubnet := range in.Public {
			publicValue := NewPublicValueNull()
			d = publicValue.Flatten(ctx, nm, puSubnet)
			diags = append(diags, d...)
			publics = append(publics, publicValue)
		}
		public, d = types.ListValue(PublicValue{}.Type(ctx), publics)
		diags = append(diags, d...)
	}
	v.Public = public

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PublicValue) Flatten(ctx context.Context, name string, in rafay.AZSubnetSpec) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Name = types.StringValue(name)
	v.Id = types.StringValue(in.ID)
	v.Az = types.StringValue(in.AZ)
	v.Cidr = types.StringValue(in.CIDR)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PrivateValue) Flatten(ctx context.Context, name string, in rafay.AZSubnetSpec) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Name = types.StringValue(name)
	v.Id = types.StringValue(in.ID)
	v.Az = types.StringValue(in.AZ)
	v.Cidr = types.StringValue(in.CIDR)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IdentityProvidersValue) Flatten(ctx context.Context, in *rafay.IdentityProvider) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.IdentityProvidersType = types.StringValue(in.Type)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Iam3Value) Flatten(ctx context.Context, in *rafay.EKSClusterIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.ServiceRoleArn = types.StringValue(in.ServiceRoleARN)
	v.ServiceRolePermissionBoundary = types.StringValue(in.ServiceRolePermissionsBoundary)
	v.FargatePodExecutionRoleArn = types.StringValue(in.FargatePodExecutionRoleARN)
	v.FargatePodExecutionRolePermissionsBoundary = types.StringValue(in.FargatePodExecutionRolePermissionsBoundary)
	v.WithOidc = types.BoolPointerValue(in.WithOIDC)
	v.VpcResourceControllerPolicy = types.BoolPointerValue(in.VPCResourceControllerPolicy)

	podIdentityAssociations := types.ListNull(PodIdentityAssociationsValue{}.Type(ctx))
	if len(in.PodIdentityAssociations) > 0 {
		pias := []attr.Value{}
		for _, p := range in.PodIdentityAssociations {
			pia := NewPodIdentityAssociationsValueNull()
			d = pia.Flatten(ctx, p)
			diags = append(diags, d...)
			pias = append(pias, pia)
		}
		podIdentityAssociations, d = types.ListValue(PodIdentityAssociationsValue{}.Type(ctx), pias)
		diags = append(diags, d...)
	}
	v.PodIdentityAssociations = podIdentityAssociations

	serviceAccounts := types.ListNull(ServiceAccountsValue{}.Type(ctx))
	if len(in.ServiceAccounts) > 0 {
		serviceAccountsList := []attr.Value{}
		for _, sa := range in.ServiceAccounts {
			serviceAccount := NewServiceAccountsValueNull()
			d = serviceAccount.Flatten(ctx, sa)
			diags = append(diags, d...)
			serviceAccountsList = append(serviceAccountsList, serviceAccount)
		}
		serviceAccounts, d = types.ListValue(ServiceAccountsValue{}.Type(ctx), serviceAccountsList)
		diags = append(diags, d...)
	}
	v.ServiceAccounts = serviceAccounts

	v.state = attr.ValueStateKnown
	return diags
}

func (v *ServiceAccountsValue) Flatten(ctx context.Context, in *rafay.EKSClusterIAMServiceAccount) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	attachPolicyArns2 := types.ListNull(types.StringType)
	if len(in.AttachPolicyARNs) > 0 {
		arns := []attr.Value{}
		for _, arn := range in.AttachPolicyARNs {
			arns = append(arns, types.StringValue(arn))
		}
		attachPolicyArns2, d = types.ListValue(types.StringType, arns)
		diags = append(diags, d...)
	}
	v.AttachPolicyArns2 = attachPolicyArns2

	if in.AttachPolicy != nil && len(in.AttachPolicy) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.AttachPolicy)
		if err != nil {
			diags.AddError("attach policy marshal error", err.Error())
		}
		v.AttachPolicy = types.StringValue(string(jsonStr))
	}

	v.AttachRoleArn = types.StringValue(in.AttachRoleARN)
	v.PermissionsBoundary = types.StringValue(in.PermissionsBoundary)
	v.RoleName = types.StringValue(in.RoleName)
	v.RoleOnly = types.BoolPointerValue(in.RoleOnly)

	tagMap := types.MapNull(types.StringType)
	if len(in.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range in.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags3 = tagMap

	md := NewMetadata3ValueNull()
	d = md.Flatten(ctx, in.Metadata)
	diags = append(diags, d...)
	v.Metadata3, d = types.ListValue(Metadata3Value{}.Type(ctx), []attr.Value{md})
	diags = append(diags, d...)

	policies := NewWellKnownPolicies2ValueNull()
	d = policies.Flatten(ctx, in.WellKnownPolicies)
	diags = append(diags, d...)
	v.WellKnownPolicies2, d = types.ListValue(WellKnownPolicies2Value{}.Type(ctx), []attr.Value{policies})
	diags = append(diags, d...)

	status := NewStatusValueNull()
	d = status.Flatten(ctx, in.Status)
	diags = append(diags, d...)
	v.Status, d = types.ListValue(StatusValue{}.Type(ctx), []attr.Value{status})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *StatusValue) Flatten(ctx context.Context, in *rafay.ClusterIAMServiceAccountStatus) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.RoleArn = types.StringValue(in.RoleARN)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *WellKnownPolicies2Value) Flatten(ctx context.Context, in *rafay.WellKnownPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.ImageBuilder = types.BoolPointerValue(in.ImageBuilder)
	v.AwsLoadBalancerController = types.BoolPointerValue(in.AWSLoadBalancerController)
	v.ExternalDns = types.BoolPointerValue(in.ExternalDNS)
	v.CertManager = types.BoolPointerValue(in.CertManager)
	v.EbsCsiController = types.BoolPointerValue(in.EBSCSIController)
	v.EfsCsiController = types.BoolPointerValue(in.EFSCSIController)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Metadata3Value) Flatten(ctx context.Context, in *rafay.EKSClusterIAMMeta) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Name = types.StringValue(in.Name)
	v.Namespace = types.StringValue(in.Namespace)
	labels := types.MapNull(types.StringType)
	if len(in.Labels) > 0 {
		lbs := map[string]attr.Value{}
		for key, val := range in.Labels {
			lbs[key] = types.StringValue(val)
		}
		labels, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels = labels

	annotations := types.MapNull(types.StringType)
	if len(in.Annotations) > 0 {
		anots := map[string]attr.Value{}
		for key, val := range in.Annotations {
			anots[key] = types.StringValue(val)
		}
		annotations, d = types.MapValue(types.StringType, anots)
		diags = append(diags, d...)
	}
	v.Annotations = annotations

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PodIdentityAssociationsValue) Flatten(ctx context.Context, in *rafay.IAMPodIdentityAssociation) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Namespace = types.StringValue(in.Namespace)
	v.ServiceAccountName = types.StringValue(in.ServiceAccountName)
	v.RoleArn = types.StringValue(in.RoleARN)
	v.CreateServiceAccount = types.BoolValue(in.CreateServiceAccount)
	v.RoleName = types.StringValue(in.RoleName)
	v.PermissionBoundaryArn = types.StringValue(in.PermissionsBoundaryARN)
	permissionPolicyArns := types.ListNull(types.StringType)
	if len(in.PermissionPolicyARNs) > 0 {
		arns := []attr.Value{}
		for _, arn := range in.PermissionPolicyARNs {
			arns = append(arns, types.StringValue(arn))
		}
		permissionPolicyArns, d = types.ListValue(types.StringType, arns)
		diags = append(diags, d...)
	}
	v.PermissionPolicyArns = permissionPolicyArns
	if len(in.PermissionPolicy) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.PermissionPolicy)
		if err != nil {
			diags.AddError("Permission policy marshal error", err.Error())
		}
		v.PermissionPolicy = types.StringValue(string(jsonStr))
	}
	tagMap := types.MapNull(types.StringType)
	if len(in.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range in.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags = tagMap

	policies := NewWellKnownPoliciesValueNull()
	d = policies.Flatten(ctx, in.WellKnownPolicies)
	diags = append(diags, d...)
	v.WellKnownPolicies, d = types.ListValue(WellKnownPoliciesValue{}.Type(ctx), []attr.Value{policies})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *WellKnownPoliciesValue) Flatten(ctx context.Context, in *rafay.WellKnownPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.ImageBuilder = types.BoolPointerValue(in.ImageBuilder)
	v.AwsLoadBalancerController = types.BoolPointerValue(in.AWSLoadBalancerController)
	v.ExternalDns = types.BoolPointerValue(in.ExternalDNS)
	v.CertManager = types.BoolPointerValue(in.CertManager)
	v.EbsCsiController = types.BoolPointerValue(in.EBSCSIController)
	v.EfsCsiController = types.BoolPointerValue(in.EFSCSIController)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *KubernetesNetworkConfigValue) Flatten(ctx context.Context, in *rafay.KubernetesNetworkConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.IpFamily = types.StringValue(in.IPFamily)
	v.ServiceIpv4Cidr = types.StringValue(in.ServiceIPv4CIDR)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Metadata2Value) Flatten(ctx context.Context, in *rafay.EKSClusterConfigMetadata) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(in.Name)
	v.Region = types.StringValue(in.Region)
	v.Version = types.StringValue(in.Version)

	tagMap := types.MapNull(types.StringType)
	if len(in.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range in.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags = tagMap

	antsMap := types.MapNull(types.StringType)
	if len(in.Annotations) != 0 {
		ants := map[string]attr.Value{}
		for key, val := range in.Annotations {
			ants[key] = types.StringValue(val)
		}
		antsMap, d = types.MapValue(types.StringType, ants)
		diags = append(diags, d...)
	}
	v.Annotations = antsMap

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SystemComponentsPlacementValue) Flatten(ctx context.Context, in *rafay.SystemComponentsPlacement) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	ns := types.MapNull(types.StringType)
	if len(in.NodeSelector) != 0 {
		nodeSelector := map[string]attr.Value{}
		for key, val := range in.NodeSelector {
			nodeSelector[key] = types.StringValue(val)
		}
		ns, d = types.MapValue(types.StringType, nodeSelector)
		diags = append(diags, d...)
	}
	v.NodeSelector = ns

	tolerations := types.ListNull(TolerationsValue{}.Type(ctx))
	if len(in.Tolerations) > 0 {
		tolerationsList := make([]attr.Value, 0, len(in.Tolerations))
		for _, t := range in.Tolerations {
			tol := NewTolerationsValueNull()
			d = tol.Flatten(ctx, t)
			diags = append(diags, d...)
			tolerationsList = append(tolerationsList, tol)
		}
		tolerations, d = types.ListValue(TolerationsValue{}.Type(ctx), tolerationsList)
		diags = append(diags, d...)
	}
	v.Tolerations = tolerations

	// DaemonsetOverride
	daemonsetOverride := NewDaemonsetOverrideValueNull()
	d = daemonsetOverride.Flatten(ctx, in.DaemonsetOverride)
	diags = append(diags, d...)
	v.DaemonsetOverride, d = types.ListValue(DaemonsetOverrideValue{}.Type(ctx), []attr.Value{daemonsetOverride})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *TolerationsValue) Flatten(ctx context.Context, in *rafay.Tolerations) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Key = types.StringValue(in.Key)
	v.Operator = types.StringValue(in.Operator)
	v.Value = types.StringValue(in.Value)
	v.Effect = types.StringValue(in.Effect)
	if in.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64Value(int64(*in.TolerationSeconds))
	} else {
		v.TolerationSeconds = types.Int64Null()
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *DaemonsetOverrideValue) Flatten(ctx context.Context, in *rafay.DaemonsetOverride) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.NodeSelectionEnabled = types.BoolPointerValue(in.NodeSelectionEnabled)

	tolerations2 := types.ListNull(Tolerations2Value{}.Type(ctx))
	if len(in.Tolerations) > 0 {
		tolerationsList := make([]attr.Value, 0, len(in.Tolerations))
		for _, t := range in.Tolerations {
			tol := NewTolerations2ValueNull()
			d = tol.Flatten(ctx, t)
			diags = append(diags, d...)
			tolerationsList = append(tolerationsList, tol)
		}
		tolerations2, d = types.ListValue(Tolerations2Value{}.Type(ctx), tolerationsList)
		diags = append(diags, d...)
	}
	v.Tolerations2 = tolerations2

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Tolerations2Value) Flatten(ctx context.Context, in *rafay.Tolerations) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Key = types.StringValue(in.Key)
	v.Operator = types.StringValue(in.Operator)
	v.Value = types.StringValue(in.Value)
	v.Effect = types.StringValue(in.Effect)
	if in.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64Value(int64(*in.TolerationSeconds))
	} else {
		v.TolerationSeconds = types.Int64Null()
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SharingValue) Flatten(ctx context.Context, in *rafay.V1ClusterSharing) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Enabled = types.BoolPointerValue(in.Enabled)

	projects := types.ListNull(ProjectsValue{}.Type(ctx))
	if len(in.Projects) > 0 {
		projectsList := make([]attr.Value, 0, len(in.Projects))
		for _, p := range in.Projects {
			proj := NewProjectsValueNull()
			d = proj.Flatten(ctx, p)
			diags = append(diags, d...)
			projectsList = append(projectsList, proj)
		}
		projects, d = types.ListValue(ProjectsValue{}.Type(ctx), projectsList)
		diags = append(diags, d...)
	}
	v.Projects = projects

	v.state = attr.ValueStateKnown
	return diags
}

func (v *ProjectsValue) Flatten(ctx context.Context, in *rafay.V1ClusterSharingProject) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.Name = types.StringValue(in.Name)

	v.state = attr.ValueStateKnown
	return diags
}
