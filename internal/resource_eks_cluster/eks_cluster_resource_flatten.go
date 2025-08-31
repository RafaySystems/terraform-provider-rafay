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

var ngMapInUse = true
var existingAddonNames = []string{}

func FlattenEksCluster(ctx context.Context, c rafay.EKSCluster, data *EksClusterModel) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if data == nil {
		return diags
	}

	cv := NewClusterValueNull()
	d = cv.Flatten(ctx, c)
	diags = append(diags, d...)
	clusterElements := []attr.Value{
		cv,
	}
	data.Cluster, d = types.ListValue(ClusterValue{}.Type(ctx), clusterElements)
	diags = append(diags, d...)

	return diags
}

func FlattenEksClusterConfig(ctx context.Context, c rafay.EKSClusterConfig, data *EksClusterModel) diag.Diagnostics {
	var diags, d diag.Diagnostics

	// check ngMap are used
	ccList := make([]ClusterConfigValue, len(data.ClusterConfig.Elements()))
	d = data.ClusterConfig.ElementsAs(ctx, &ccList, false)
	diags = append(diags, d...)
	if len(ccList) > 0 && len(ccList[0].NodeGroups.Elements()) > 0 {
		ngMapInUse = false
	}

	// get existing addons
	if len(ccList) > 0 {
		addonsList := make([]AddonsValue, 0, len(ccList[0].Addons.Elements()))
		diags = ccList[0].Addons.ElementsAs(ctx, &addonsList, false)
		diags = append(diags, d...)
		for _, addon := range addonsList {
			existingAddonNames = append(existingAddonNames, getStringValue(addon.Name))
		}
	}

	cc := NewClusterConfigValueNull()
	d = cc.Flatten(ctx, c)
	diags = append(diags, d...)
	clusterElements := []attr.Value{
		cc,
	}
	data.ClusterConfig, d = types.ListValue(ClusterConfigValue{}.Type(ctx), clusterElements)
	diags = append(diags, d...)

	return diags
}

func (v *ClusterValue) Flatten(ctx context.Context, c rafay.EKSCluster) diag.Diagnostics {
	var diags, d diag.Diagnostics
	v.Kind = types.StringValue(c.Kind)

	md := NewMetadataValueNull()
	d = md.Flatten(ctx, c.Metadata)
	diags = append(diags, d...)
	mdElements := []attr.Value{
		md,
	}
	v.Metadata, d = types.ListValue(MetadataValue{}.Type(ctx), mdElements)
	diags = append(diags, d...)

	spec := NewSpecValueNull()
	d = spec.Flatten(ctx, c.Spec)
	diags = append(diags, d...)
	specElements := []attr.Value{
		spec,
	}
	v.Spec, d = types.ListValue(SpecValue{}.Type(ctx), specElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *MetadataValue) Flatten(ctx context.Context, md *rafay.EKSClusterMetadata) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(md.Name)
	v.Project = types.StringValue(md.Project)

	lbsMap := types.MapNull(types.StringType)
	if len(md.Labels) != 0 {
		lbs := map[string]attr.Value{}
		for key, val := range md.Labels {
			lbs[key] = types.StringValue(val)
		}
		lbsMap, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels = lbsMap

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SpecValue) Flatten(ctx context.Context, spec *rafay.EKSSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if spec == nil {
		return diags
	}

	v.Blueprint = types.StringValue(spec.Blueprint)
	v.BlueprintVersion = types.StringValue(spec.BlueprintVersion)
	v.CloudProvider = types.StringValue(spec.CloudProvider)
	v.CniProvider = types.StringValue(spec.CniProvider)
	v.CrossAccountRoleArn = types.StringValue(spec.CrossAccountRoleArn)
	v.SpecType = types.StringValue(spec.Type)

	cp := NewCniParamsValueNull()
	d = cp.Flatten(ctx, spec.CniParams)
	diags = append(diags, d...)
	cpElements := []attr.Value{
		cp,
	}
	v.CniParams, d = types.ListValue(CniParamsValue{}.Type(ctx), cpElements)

	proxycfgMap := types.MapNull(types.StringType)
	if spec.ProxyConfig != nil {
		pc := map[string]attr.Value{}
		if len(spec.ProxyConfig.HttpProxy) > 0 {
			pc["http_proxy"] = types.StringValue(spec.ProxyConfig.HttpProxy)
		}
		if len(spec.ProxyConfig.HttpsProxy) > 0 {
			pc["https_proxy"] = types.StringValue(spec.ProxyConfig.HttpsProxy)
		}
		if len(spec.ProxyConfig.NoProxy) > 0 {
			pc["no_proxy"] = types.StringValue(spec.ProxyConfig.NoProxy)
		}
		if len(spec.ProxyConfig.ProxyAuth) > 0 {
			pc["proxy_auth"] = types.StringValue(spec.ProxyConfig.ProxyAuth)
		}
		if len(spec.ProxyConfig.BootstrapCA) > 0 {
			pc["bootstrap_ca"] = types.StringValue(spec.ProxyConfig.BootstrapCA)
		}
		if spec.ProxyConfig.Enabled {
			pc["enabled"] = types.StringValue("true")
		}
		if spec.ProxyConfig.AllowInsecureBootstrap {
			pc["allow_insecure_bootstrap"] = types.StringValue("true")
		}
	}
	v.ProxyConfig = proxycfgMap

	scp := NewSystemComponentsPlacementValueNull()
	d = scp.Flatten(ctx, spec.SystemComponentsPlacement)
	diags = append(diags, d...)
	v.SystemComponentsPlacement, d = scp.ToObjectValue(ctx)
	diags = append(diags, d...)

	sh := NewSharingValueNull()
	d = sh.Flatten(ctx, spec.Sharing)
	diags = append(diags, d...)
	v.Sharing, d = sh.ToObjectValue(ctx)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *CniParamsValue) Flatten(ctx context.Context, cp *rafay.CustomCni) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if cp == nil {
		return diags
	}

	v.CustomCniCidr = types.StringValue(cp.CustomCniCidr)

	customCniCrdSpec := types.ListNull(CustomCniCrdSpecValue{}.Type(ctx))
	if len(cp.CustomCniCrdSpec) > 0 {
		csElements := make([]attr.Value, 0, len(cp.CustomCniCrdSpec))
		for name, cniSpec := range cp.CustomCniCrdSpec {
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

func (v *CustomCniCrdSpecValue) Flatten(ctx context.Context, name string, cs []rafay.CustomCniSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(name)

	cniSpec := types.ListNull(CniSpecValue{}.Type(ctx))
	if len(cs) > 0 {
		specElements := make([]attr.Value, 0, len(cs))
		for _, spec := range cs {
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

func (v *CniSpecValue) Flatten(ctx context.Context, spec rafay.CustomCniSpec) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Subnet = types.StringValue(spec.Subnet)

	securityGroups := types.ListNull(types.StringType)
	if len(spec.SecurityGroups) > 0 {
		sgElements := make([]attr.Value, 0, len(spec.SecurityGroups))
		for _, sg := range spec.SecurityGroups {
			sgElements = append(sgElements, types.StringValue(sg))
		}
		securityGroups, d = types.ListValue(types.StringType, sgElements)
		diags = append(diags, d...)
	}
	v.SecurityGroups = securityGroups

	v.state = attr.ValueStateKnown
	return diags
}

func (v *ClusterConfigValue) Flatten(ctx context.Context, in rafay.EKSClusterConfig) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Apiversion = types.StringValue(in.APIVersion)
	v.Kind = types.StringValue(in.Kind)

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

	md := NewMetadata2ValueNull()
	d = md.Flatten(ctx, in.Metadata)
	diags = append(diags, d...)
	mdElements := []attr.Value{
		md,
	}
	v.Metadata2, d = types.ListValue(Metadata2Value{}.Type(ctx), mdElements)
	diags = append(diags, d...)

	if ngMapInUse {
		// TODO(Akshay): Update later
		ngMap := types.MapNull(NodeGroupsMapValue{}.Type(ctx))
		if len(in.NodeGroups) != 0 {
			nodegrp := map[string]attr.Value{}
			for _, ng := range in.NodeGroups {
				ngrp := NewNodeGroupsMapValueNull()
				d = ngrp.Flatten(ctx, ng)
				diags = append(diags, d...)
				nodegrp[ng.Name] = ngrp
			}
			ngMap, d = types.MapValue(NodeGroupsMapValue{}.Type(ctx), nodegrp)
			diags = append(diags, d...)
		}

		v.NodeGroupsMap = ngMap
		v.NodeGroups = types.ListNull(NodeGroupsValue{}.Type(ctx))
	} else {
		ngElements := []attr.Value{}
		for _, ng := range in.NodeGroups {
			ngList := NewNodeGroupsValueNull()
			d = ngList.Flatten(ctx, ng)
			diags = append(diags, d...)
			ngElements = append(ngElements, ngList)
		}
		v.NodeGroups, d = types.ListValue(NodeGroupsValue{}.Type(ctx), ngElements)
		diags = append(diags, d...)
		v.NodeGroupsMap = types.MapNull(NodeGroupsMapValue{}.Type(ctx))
	}

	netconf := NewKubernetesNetworkConfigValueNull()
	d = netconf.Flatten(ctx, in.KubernetesNetworkConfig)
	diags = append(diags, d...)
	v.KubernetesNetworkConfig, d = types.ListValue(KubernetesNetworkConfigValue{}.Type(ctx), []attr.Value{netconf})
	diags = append(diags, d...)

	iam := NewIam3ValueNull()
	d = iam.Flatten(ctx, in.IAM)
	diags = append(diags, d...)
	v.Iam3, d = types.ListValue(Iam3Value{}.Type(ctx), []attr.Value{iam})
	diags = append(diags, d...)

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

	vpc := NewVpcValueNull()
	d = vpc.Flatten(ctx, in.VPC)
	diags = append(diags, d...)
	v.Vpc, d = types.ListValue(VpcValue{}.Type(ctx), []attr.Value{vpc})
	diags = append(diags, d...)

	addons := types.ListNull(AddonsValue{}.Type(ctx))
	if len(in.Addons) > 0 {
		addonsList := []attr.Value{}
		for _, add := range in.Addons {
			// Hack: Check if addon exists in state (ie. provided by user while creating)
			exists := false
			for _, e := range existingAddonNames {
				if strings.EqualFold(e, add.Name) {
					exists = true
				}
			}

			if exists {
				addon := NewAddonsValueNull()
				d = addon.Flatten(ctx, add)
				diags = append(diags, d...)
				addonsList = append(addonsList, addon)
			}
		}
		addons, d = types.ListValue(AddonsValue{}.Type(ctx), addonsList)
		diags = append(diags, d...)
	}
	v.Addons = addons

	privateCluster := NewPrivateClusterValueNull()
	d = privateCluster.Flatten(ctx, in.PrivateCluster)
	diags = append(diags, d...)
	v.PrivateCluster, d = types.ListValue(PrivateClusterValue{}.Type(ctx), []attr.Value{privateCluster})
	diags = append(diags, d...)

	// managed node groups
	managedNodegroups := types.ListNull(ManagedNodegroupsValue{}.Type(ctx))
	if len(in.ManagedNodeGroups) > 0 {
		mngElements := []attr.Value{}
		for _, mng := range in.ManagedNodeGroups {
			mngList := NewManagedNodegroupsValueNull()
			d = mngList.Flatten(ctx, mng)
			diags = append(diags, d...)
			mngElements = append(mngElements, mngList)
		}
		managedNodegroups, d = types.ListValue(ManagedNodegroupsValue{}.Type(ctx), mngElements)
		diags = append(diags, d...)
	}
	v.ManagedNodegroups = managedNodegroups

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

	cloudWatch := NewCloudWatchValueNull()
	d = cloudWatch.Flatten(ctx, in.CloudWatch)
	diags = append(diags, d...)
	v.CloudWatch, d = types.ListValue(CloudWatchValue{}.Type(ctx), []attr.Value{cloudWatch})
	diags = append(diags, d...)

	SecretsEncryption := NewSecretsEncryptionValueNull()
	d = SecretsEncryption.Flatten(ctx, in.SecretsEncryption)
	diags = append(diags, d...)
	v.SecretsEncryption, d = types.ListValue(SecretsEncryptionValue{}.Type(ctx), []attr.Value{SecretsEncryption})
	diags = append(diags, d...)

	identityMappings := NewIdentityMappingsValueNull()
	d = identityMappings.Flatten(ctx, in.IdentityMappings)
	diags = append(diags, d...)
	v.IdentityMappings, d = types.ListValue(IdentityMappingsValue{}.Type(ctx), []attr.Value{identityMappings})
	diags = append(diags, d...)

	accessConfig := NewAccessConfigValueNull()
	d = accessConfig.Flatten(ctx, in.AccessConfig)
	diags = append(diags, d...)
	v.AccessConfig, d = types.ListValue(AccessConfigValue{}.Type(ctx), []attr.Value{accessConfig})
	diags = append(diags, d...)

	addonsConfig := NewAddonsConfigValueNull()
	d = addonsConfig.Flatten(ctx, in.AddonsConfig)
	diags = append(diags, d...)
	v.AddonsConfig, d = types.ListValue(AddonsConfigValue{}.Type(ctx), []attr.Value{addonsConfig})
	diags = append(diags, d...)

	autoModeConfig := NewAutoModeConfigValueNull()
	d = autoModeConfig.Flatten(ctx, in.AutoModeConfig)
	diags = append(diags, d...)
	v.AutoModeConfig, d = types.ListValue(AutoModeConfigValue{}.Type(ctx), []attr.Value{autoModeConfig})
	diags = append(diags, d...)

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

func (v *ManagedNodegroupsValue) Flatten(ctx context.Context, ng *rafay.ManagedNodeGroup) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(ng.Name)
	v.AmiFamily = types.StringValue(ng.AMIFamily)
	v.DesiredCapacity = types.Int64Value(int64(*ng.DesiredCapacity))
	v.DisableImdsv1 = types.BoolValue(*ng.DisableIMDSv1)
	v.DisablePodsImds = types.BoolValue(*ng.DisablePodIMDS)
	v.EfaEnabled = types.BoolValue(*ng.EFAEnabled)
	v.InstanceType = types.StringValue(ng.InstanceType)
	v.MaxPodsPerNode = types.Int64Value(int64(*ng.MaxPodsPerNode))
	v.MaxSize = types.Int64Value(int64(*ng.MaxSize))
	v.MinSize = types.Int64Value(int64(*ng.MinSize))
	v.PrivateNetworking = types.BoolValue(*ng.PrivateNetworking)
	v.Version = types.StringValue(ng.Version)
	v.VolumeIops = types.Int64Value(int64(*ng.VolumeIOPS))
	v.VolumeSize = types.Int64Value(int64(*ng.VolumeSize))
	v.VolumeThroughput = types.Int64Value(int64(*ng.VolumeThroughput))
	v.VolumeType = types.StringValue(ng.VolumeType)
	v.EbsOptimized = types.BoolValue(*ng.EBSOptimized)
	v.VolumeName = types.StringValue(ng.VolumeName)
	v.VolumeEncrypted = types.BoolValue(*ng.VolumeEncrypted)
	v.VolumeKmsKeyId = types.StringValue(ng.VolumeKmsKeyID)
	v.OverrideBootstrapCommand = types.StringValue(ng.OverrideBootstrapCommand)

	preBootstrapCommands := types.ListNull(types.StringType)
	if len(ng.PreBootstrapCommands) > 0 {
		pbElements := []attr.Value{}
		for _, pb := range ng.PreBootstrapCommands {
			pbElements = append(pbElements, types.StringValue(pb))
		}
		preBootstrapCommands, d = types.ListValue(types.StringType, pbElements)
		diags = append(diags, d...)
	}
	v.PreBootstrapCommands = preBootstrapCommands

	asgSuspendProcesses := types.ListNull(types.StringType)
	if len(ng.ASGSuspendProcesses) > 0 {
		aspElements := []attr.Value{}
		for _, asp := range ng.ASGSuspendProcesses {
			aspElements = append(aspElements, types.StringValue(asp))
		}
		asgSuspendProcesses, d = types.ListValue(types.StringType, aspElements)
		diags = append(diags, d...)
	}
	v.AsgSuspendProcesses = asgSuspendProcesses

	v.EnableDetailedMonitoring = types.BoolPointerValue(ng.EnableDetailedMonitoring)

	availabilityZones := types.ListNull(types.StringType)
	if len(ng.AvailabilityZones) > 0 {
		azElements := []attr.Value{}
		for _, az := range ng.AvailabilityZones {
			azElements = append(azElements, types.StringValue(az))
		}
		availabilityZones, d = types.ListValue(types.StringType, azElements)
		diags = append(diags, d...)
	}
	v.AvailabilityZones = availabilityZones
	snElements := []attr.Value{}
	for _, sn := range ng.Subnets {
		snElements = append(snElements, types.StringValue(sn))
	}
	v.Subnets, d = types.ListValue(types.StringType, snElements)
	diags = append(diags, d...)
	v.InstancePrefix = types.StringValue(ng.InstancePrefix)
	v.InstanceName = types.StringValue(ng.InstanceName)
	lbsMap := types.MapNull(types.StringType)
	if len(ng.Labels) != 0 {
		lbs := map[string]attr.Value{}
		for key, val := range ng.Labels {
			lbs[key] = types.StringValue(val)
		}
		lbsMap, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels = lbsMap
	tagMap := types.MapNull(types.StringType)
	if len(ng.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range ng.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags = tagMap
	v.Ami = types.StringValue(ng.AMI)
	v.Spot = types.BoolPointerValue(ng.Spot)
	instanceTypes := types.ListNull(types.StringType)
	if len(ng.InstanceTypes) > 0 {
		instanceTypesList := []attr.Value{}
		for _, val := range ng.InstanceTypes {
			instanceTypesList = append(instanceTypesList, types.StringValue(val))
		}
		instanceTypes, d = types.ListValue(types.StringType, instanceTypesList)
		diags = append(diags, d...)
	}
	v.InstanceTypes = instanceTypes

	// blocks start here
	iam := NewIam4ValueNull()
	d = iam.Flatten(ctx, ng.IAM)
	diags = append(diags, d...)
	iamElements := []attr.Value{iam}
	v.Iam4, d = types.ListValue(Iam4Value{}.Type(ctx), iamElements)
	diags = append(diags, d...)

	ssh := NewSsh4ValueNull()
	d = ssh.Flatten(ctx, ng.SSH)
	diags = append(diags, d...)
	v.Ssh4, d = types.ListValue(Ssh4Value{}.Type(ctx), []attr.Value{ssh})
	diags = append(diags, d...)

	placement := NewPlacement4ValueNull()
	d = placement.Flatten(ctx, ng.Placement)
	diags = append(diags, d...)
	v.Placement4, d = types.ListValue(Placement4Value{}.Type(ctx), []attr.Value{placement})
	diags = append(diags, d...)

	instanceSel := NewInstanceSelector4ValueNull()
	d = instanceSel.Flatten(ctx, ng.InstanceSelector)
	diags = append(diags, d...)
	v.InstanceSelector4, d = types.ListValue(InstanceSelector4Value{}.Type(ctx), []attr.Value{instanceSel})
	diags = append(diags, d...)

	bottlerkt := NewBottleRocket4ValueNull()
	d = bottlerkt.Flatten(ctx, ng.Bottlerocket)
	diags = append(diags, d...)
	v.BottleRocket4, d = types.ListValue(BottleRocket4Value{}.Type(ctx), []attr.Value{bottlerkt})
	diags = append(diags, d...)

	taints := types.ListNull(Taints4Value{}.Type(ctx))
	if len(ng.Taints) > 0 {
		taintsList := []attr.Value{}
		for _, val := range ng.Taints {
			taint := NewTaints4ValueNull()
			d = taint.Flatten(ctx, val)
			diags = append(diags, d...)
			taintsList = append(taintsList, taint)
		}
		taints, d = types.ListValue(Taints4Value{}.Type(ctx), taintsList)
		diags = append(diags, d...)
	}
	v.Taints4 = taints

	updateConfig := NewUpdateConfig4ValueNull()
	d = updateConfig.Flatten(ctx, ng.UpdateConfig)
	diags = append(diags, d...)
	v.UpdateConfig4, d = types.ListValue(UpdateConfig4Value{}.Type(ctx), []attr.Value{updateConfig})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

// ////////managed node groups start
func (v *UpdateConfig4Value) Flatten(ctx context.Context, in *rafay.NodeGroupUpdateConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.MaxUnavailable != nil {
		v.MaxUnavaliable = types.Int64Value(int64(*in.MaxUnavailable))
	}
	if in.MaxUnavailablePercentage != nil {
		v.MaxUnavaliablePercetage = types.Int64Value(int64(*in.MaxUnavailablePercentage))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Taints4Value) Flatten(ctx context.Context, in rafay.NodeGroupTaint) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Key = types.StringValue(in.Key)
	v.Value = types.StringValue(in.Value)
	v.Effect = types.StringValue(in.Effect)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *BottleRocket4Value) Flatten(ctx context.Context, in *rafay.NodeGroupBottlerocket) diag.Diagnostics {
	var diags diag.Diagnostics

	v.EnableAdminContainer = types.BoolPointerValue(in.EnableAdminContainer)

	if in.Settings != nil && len(in.Settings) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Settings)
		if err != nil {
			diags.AddError("Bottle rocket marshal error", err.Error())
		}
		v.Settings = types.StringValue(string(jsonStr))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *InstanceSelector4Value) Flatten(ctx context.Context, instanceSel *rafay.InstanceSelector) diag.Diagnostics {
	var diags diag.Diagnostics

	if instanceSel.VCPUs != nil {
		v.Vcpus = types.Int64Value(int64(*instanceSel.VCPUs))
	}
	v.Memory = types.StringValue(instanceSel.Memory)
	if instanceSel.GPUs != nil {
		v.Gpus = types.Int64Value(int64(*instanceSel.GPUs))
	}
	v.CpuArchitecture = types.StringValue(instanceSel.CPUArchitecture)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Placement4Value) Flatten(ctx context.Context, placement *rafay.Placement) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Group = types.StringValue(placement.GroupName)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Ssh4Value) Flatten(ctx context.Context, ssh *rafay.NodeGroupSSH) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Allow = types.BoolPointerValue(ssh.Allow)
	v.PublicKey = types.StringValue(ssh.PublicKey)
	v.PublicKeyName = types.StringValue(ssh.PublicKeyName)

	sourceSecurityGroupIds := types.ListNull(types.StringType)
	if len(ssh.SourceSecurityGroupIDs) > 0 {
		ids := []attr.Value{}
		for _, id := range ssh.SourceSecurityGroupIDs {
			ids = append(ids, types.StringValue(id))
		}
		sourceSecurityGroupIds, d = types.ListValue(types.StringType, ids)
		diags = append(diags, d...)
	}
	v.SourceSecurityGroupIds = sourceSecurityGroupIds

	v.EnableSsm = types.BoolPointerValue(ssh.EnableSSM)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Iam4Value) Flatten(ctx context.Context, iam *rafay.NodeGroupIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if iam == nil {
		return diags
	}

	// TODO(Akshay): Check if Attach Policy v2. based on that populate attach_policy and attach_policy_v2
	if iam.AttachPolicy != nil {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := json2.Marshal(iam.AttachPolicy)
		if err != nil {
			diags.AddError("Attach Policy Marshal Error", err.Error())
		}
		v.AttachPolicyV2 = types.StringValue(string(jsonBytes))
	}

	attachPolicy := NewAttachPolicyValueNull()
	d = attachPolicy.Flatten(ctx, iam.AttachPolicy)
	diags = append(diags, d...)
	v.AttachPolicy4, d = types.ListValue(AttachPolicy4Value{}.Type(ctx), []attr.Value{attachPolicy})
	diags = append(diags, d...)

	attachPolicyArns := types.ListNull(types.StringType)
	if len(iam.AttachPolicyARNs) > 0 {
		arns := []attr.Value{}
		for _, arn := range iam.AttachPolicyARNs {
			arns = append(arns, types.StringValue(arn))
		}
		attachPolicyArns, d = types.ListValue(types.StringType, arns)
		diags = append(diags, d...)
	}
	v.AttachPolicyArns = attachPolicyArns

	v.InstanceProfileArn = types.StringValue(iam.InstanceProfileARN)
	v.InstanceRoleArn = types.StringValue(iam.InstanceRoleARN)
	v.InstanceRoleName = types.StringValue(iam.InstanceRoleName)
	v.InstanceRolePermissionBoundary = types.StringValue(iam.InstanceRolePermissionsBoundary)

	addonPolicies := NewIamNodeGroupWithAddonPoliciesValueNull()
	d = addonPolicies.Flatten(ctx, iam.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies4, d = types.ListValue(IamNodeGroupWithAddonPolicies4Value{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AttachPolicy4Value) Flatten(ctx context.Context, attachpol *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if attachpol == nil {
		return diags
	}

	v.Version = types.StringValue(attachpol.Version)
	v.Id = types.StringValue(attachpol.Id)

	statement4 := types.ListNull(Statement4Value{}.Type(ctx))
	if len(attachpol.Statement) > 0 {
		stms := []attr.Value{}
		for _, stm := range attachpol.Statement {
			sv := NewStatementValueNull()
			d = sv.Flatten(ctx, stm)
			diags = append(diags, d...)
			stms = append(stms, sv)
		}
		statement4, d = types.ListValue(Statement4Value{}.Type(ctx), stms)
		diags = append(diags, d...)
	}
	v.Statement4 = statement4

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Statement4Value) Flatten(ctx context.Context, stm rafay.InlineStatement) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if len(stm.Effect) > 0 {
		v.Effect = types.StringValue(stm.Effect)
	}
	if len(stm.Sid) > 0 {
		v.Sid = types.StringValue(stm.Sid)
	}
	if stm.Action != nil && len(stm.Action.([]interface{})) > 0 {
		actEle := []attr.Value{}
		for _, act := range stm.Action.([]interface{}) {
			actEle = append(actEle, types.StringValue(act.(string)))
		}
		v.Action, d = types.ListValue(types.StringType, actEle)
		diags = append(diags, d...)
	}
	if stm.NotAction != nil && len(stm.NotAction.([]interface{})) > 0 {
		naEle := []attr.Value{}
		for _, na := range stm.NotAction.([]interface{}) {
			naEle = append(naEle, types.StringValue(na.(string)))
		}
		v.NotAction, d = types.ListValue(types.StringType, naEle)
		diags = append(diags, d...)
	}
	if len(stm.Resource.(string)) > 0 {
		v.Resource = types.StringValue(stm.Resource.(string))
	}
	if stm.NotResource != nil && len(stm.NotResource.([]interface{})) > 0 {
		nrEle := []attr.Value{}
		for _, nr := range stm.NotResource.([]interface{}) {
			nrEle = append(nrEle, types.StringValue(nr.(string)))
		}
		v.NotResource, d = types.ListValue(types.StringType, nrEle)
		diags = append(diags, d...)
	}

	if len(stm.Condition) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Condition)
		if err != nil {
			diags.AddError("condition marshal error", err.Error())
		}
		v.Condition = types.StringValue(string(jsonStr))
	}

	if len(stm.Principal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Principal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.Principal = types.StringValue(string(jsonStr))

	}

	if len(stm.NotPrincipal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.NotPrincipal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.NotPrincipal = types.StringValue(string(jsonStr))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamNodeGroupWithAddonPolicies4Value) Flatten(ctx context.Context, addonPolicies *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if addonPolicies == nil {
		return diags
	}

	v.AlbIngress = types.BoolPointerValue(addonPolicies.AWSLoadBalancerController)
	v.AppMesh = types.BoolPointerValue(addonPolicies.AppMesh)
	v.AppMeshReview = types.BoolPointerValue(addonPolicies.AppMeshPreview)
	v.AutoScaler = types.BoolPointerValue(addonPolicies.AutoScaler)
	v.CertManager = types.BoolPointerValue(addonPolicies.CertManager)
	v.CloudWatch = types.BoolPointerValue(addonPolicies.CloudWatch)
	v.Ebs = types.BoolPointerValue(addonPolicies.EBS)
	v.Efs = types.BoolPointerValue(addonPolicies.EFS)
	v.ExternalDns = types.BoolPointerValue(addonPolicies.ExternalDNS)
	v.Fsx = types.BoolPointerValue(addonPolicies.FSX)
	v.ImageBuilder = types.BoolPointerValue(addonPolicies.ImageBuilder)
	v.Xray = types.BoolPointerValue(addonPolicies.XRay)

	v.state = attr.ValueStateKnown
	return diags
}

//////////managed node groups stops

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

func (v *AddonsValue) Flatten(ctx context.Context, in *rafay.Addon) diag.Diagnostics {
	var diags, d diag.Diagnostics

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

	// TODO(Akshay): Check if Attach Policy v2. based on that populate attach_policy and attach_policy_v2
	if in.AttachPolicy != nil {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := json2.Marshal(in.AttachPolicy)
		if err != nil {
			diags.AddError("Attach Policy Marshal Error", err.Error())
		}
		v.AttachPolicyV22 = types.StringValue(string(jsonBytes))
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

	attachPolicy := NewAttachPolicy3ValueNull()
	d = attachPolicy.Flatten(ctx, in.AttachPolicy)
	diags = append(diags, d...)
	v.AttachPolicy3, d = types.ListValue(AttachPolicy3Value{}.Type(ctx), []attr.Value{attachPolicy})
	diags = append(diags, d...)

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

func (v *Statement2Value) Flatten(ctx context.Context, stm rafay.InlineStatement) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if len(stm.Effect) > 0 {
		v.Effect = types.StringValue(stm.Effect)
	}
	if len(stm.Sid) > 0 {
		v.Sid = types.StringValue(stm.Sid)
	}
	if stm.Action != nil && len(stm.Action.([]interface{})) > 0 {
		actEle := []attr.Value{}
		for _, act := range stm.Action.([]interface{}) {
			actEle = append(actEle, types.StringValue(act.(string)))
		}
		v.Action, d = types.ListValue(types.StringType, actEle)
		diags = append(diags, d...)
	}
	if stm.NotAction != nil && len(stm.NotAction.([]interface{})) > 0 {
		naEle := []attr.Value{}
		for _, na := range stm.NotAction.([]interface{}) {
			naEle = append(naEle, types.StringValue(na.(string)))
		}
		v.NotAction, d = types.ListValue(types.StringType, naEle)
		diags = append(diags, d...)
	}
	if len(stm.Resource.(string)) > 0 {
		v.Resource = types.StringValue(stm.Resource.(string))
	}
	if stm.NotResource != nil && len(stm.NotResource.([]interface{})) > 0 {
		nrEle := []attr.Value{}
		for _, nr := range stm.NotResource.([]interface{}) {
			nrEle = append(nrEle, types.StringValue(nr.(string)))
		}
		v.NotResource, d = types.ListValue(types.StringType, nrEle)
		diags = append(diags, d...)
	}

	if len(stm.Condition) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Condition)
		if err != nil {
			diags.AddError("condition marshal error", err.Error())
		}
		v.Condition = types.StringValue(string(jsonStr))
	}

	if len(stm.Principal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Principal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.Principal = types.StringValue(string(jsonStr))

	}

	if len(stm.NotPrincipal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.NotPrincipal)
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

func (v *PublicValue) Flatten(ctx context.Context, nm string, in rafay.AZSubnetSpec) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Name = types.StringValue(nm)
	v.Id = types.StringValue(in.ID)
	v.Az = types.StringValue(in.AZ)
	v.Cidr = types.StringValue(in.CIDR)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PrivateValue) Flatten(ctx context.Context, nm string, in rafay.AZSubnetSpec) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Name = types.StringValue(nm)
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

func (v *Metadata2Value) Flatten(ctx context.Context, md *rafay.EKSClusterConfigMetadata) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(md.Name)
	v.Region = types.StringValue(md.Region)
	v.Version = types.StringValue(md.Version)

	tagMap := types.MapNull(types.StringType)
	if len(md.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range md.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags = tagMap

	antsMap := types.MapNull(types.StringType)
	if len(md.Annotations) != 0 {
		ants := map[string]attr.Value{}
		for key, val := range md.Annotations {
			ants[key] = types.StringValue(val)
		}
		antsMap, d = types.MapValue(types.StringType, ants)
		diags = append(diags, d...)
	}
	v.Annotations = antsMap

	v.state = attr.ValueStateKnown
	return diags
}

func (v *NodeGroupsMapValue) Flatten(ctx context.Context, ng *rafay.NodeGroup) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.AmiFamily = types.StringValue(ng.AMIFamily)
	v.DesiredCapacity = types.Int64Value(int64(*ng.DesiredCapacity))
	v.DisableImdsv1 = types.BoolValue(*ng.DisableIMDSv1)
	v.DisablePodsImds = types.BoolValue(*ng.DisablePodIMDS)
	v.EfaEnabled = types.BoolValue(*ng.EFAEnabled)
	v.InstanceType = types.StringValue(ng.InstanceType)
	v.MaxPodsPerNode = types.Int64Value(int64(ng.MaxPodsPerNode))
	v.MaxSize = types.Int64Value(int64(*ng.MaxSize))
	v.MinSize = types.Int64Value(int64(*ng.MinSize))
	v.PrivateNetworking = types.BoolValue(*ng.PrivateNetworking)
	v.Version = types.StringValue(ng.Version)
	v.VolumeIops = types.Int64Value(int64(*ng.VolumeIOPS))
	v.VolumeSize = types.Int64Value(int64(*ng.VolumeSize))
	v.VolumeThroughput = types.Int64Value(int64(*ng.VolumeThroughput))
	v.VolumeType = types.StringValue(ng.VolumeType)

	iam := NewIam2ValueNull()
	d = iam.Flatten(ctx, ng.IAM)
	diags = append(diags, d...)
	iamElements := []attr.Value{iam}
	v.Iam2, d = types.ListValue(Iam2Value{}.Type(ctx), iamElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *Iam2Value) Flatten(ctx context.Context, iam *rafay.NodeGroupIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if iam == nil {
		return diags
	}

	addonPolicies := NewIamNodeGroupWithAddonPolicies2ValueNull()
	d = addonPolicies.Flatten(ctx, iam.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies2, d = types.ListValue(IamNodeGroupWithAddonPolicies2Value{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamNodeGroupWithAddonPolicies2Value) Flatten(ctx context.Context, addonPolicies *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if addonPolicies == nil {
		return diags
	}

	v.AlbIngress = types.BoolPointerValue(addonPolicies.AWSLoadBalancerController)
	v.AppMesh = types.BoolPointerValue(addonPolicies.AppMesh)
	v.AppMeshReview = types.BoolPointerValue(addonPolicies.AppMeshPreview)
	v.AutoScaler = types.BoolPointerValue(addonPolicies.AutoScaler)
	v.CertManager = types.BoolPointerValue(addonPolicies.CertManager)
	v.CloudWatch = types.BoolPointerValue(addonPolicies.CloudWatch)
	v.Ebs = types.BoolPointerValue(addonPolicies.EBS)
	v.Efs = types.BoolPointerValue(addonPolicies.EFS)
	v.ExternalDns = types.BoolPointerValue(addonPolicies.ExternalDNS)
	v.Fsx = types.BoolPointerValue(addonPolicies.FSX)
	v.ImageBuilder = types.BoolPointerValue(addonPolicies.ImageBuilder)
	v.Xray = types.BoolPointerValue(addonPolicies.XRay)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *NodeGroupsValue) Flatten(ctx context.Context, ng *rafay.NodeGroup) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Name = types.StringValue(ng.Name)
	v.AmiFamily = types.StringValue(ng.AMIFamily)
	if ng.DesiredCapacity != nil {
		v.DesiredCapacity = types.Int64Value(int64(*ng.DesiredCapacity))
	}
	v.DisableImdsv1 = types.BoolPointerValue(ng.DisableIMDSv1)
	v.DisablePodsImds = types.BoolPointerValue(ng.DisablePodIMDS)
	v.EfaEnabled = types.BoolPointerValue(ng.EFAEnabled)
	v.InstanceType = types.StringValue(ng.InstanceType)
	v.MaxPodsPerNode = types.Int64Value(int64(ng.MaxPodsPerNode))
	if ng.MaxSize != nil {
		v.MaxSize = types.Int64Value(int64(*ng.MaxSize))
	}
	if ng.MinSize != nil {
		v.MinSize = types.Int64Value(int64(*ng.MinSize))
	}
	v.PrivateNetworking = types.BoolPointerValue(ng.PrivateNetworking)
	v.Version = types.StringValue(ng.Version)
	if ng.VolumeIOPS != nil {
		v.VolumeIops = types.Int64Value(int64(*ng.VolumeIOPS))
	}
	if ng.VolumeSize != nil {
		v.VolumeSize = types.Int64Value(int64(*ng.VolumeSize))
	}
	if ng.VolumeThroughput != nil {
		v.VolumeThroughput = types.Int64Value(int64(*ng.VolumeThroughput))
	}
	v.VolumeType = types.StringValue(ng.VolumeType)
	// TODO: SubnetCidr is missing
	v.ClusterDns = types.StringValue(ng.ClusterDNS)
	v.EbsOptimized = types.BoolPointerValue(ng.EBSOptimized)
	v.VolumeName = types.StringValue(ng.VolumeName)
	v.VolumeEncrypted = types.BoolPointerValue(ng.VolumeEncrypted)
	v.VolumeKmsKeyId = types.StringValue(ng.VolumeKmsKeyID)
	v.OverrideBootstrapCommand = types.StringValue(ng.OverrideBootstrapCommand)

	preBootstrapCommands := types.ListNull(types.StringType)
	if len(ng.PreBootstrapCommands) > 0 {
		pbElements := []attr.Value{}
		for _, pb := range ng.PreBootstrapCommands {
			pbElements = append(pbElements, types.StringValue(pb))
		}
		preBootstrapCommands, d = types.ListValue(types.StringType, pbElements)
		diags = append(diags, d...)
	}
	v.PreBootstrapCommands = preBootstrapCommands

	asgSuspendProcess := types.ListNull(types.StringType)
	if len(ng.ASGSuspendProcesses) > 0 {
		aspElements := []attr.Value{}
		for _, asp := range ng.ASGSuspendProcesses {
			aspElements = append(aspElements, types.StringValue(asp))
		}
		asgSuspendProcess, d = types.ListValue(types.StringType, aspElements)
		diags = append(diags, d...)
	}
	v.AsgSuspendProcesses = asgSuspendProcess

	targetGroupArns := types.ListNull(types.StringType)
	if len(ng.TargetGroupARNs) > 0 {
		tgaElements := []attr.Value{}
		for _, tga := range ng.TargetGroupARNs {
			tgaElements = append(tgaElements, types.StringValue(tga))
		}
		targetGroupArns, d = types.ListValue(types.StringType, tgaElements)
		diags = append(diags, d...)
	}
	v.TargetGroupArns = targetGroupArns

	classicLoadBalancerNames := types.ListNull(types.StringType)
	if len(ng.ClassicLoadBalancerNames) > 0 {
		clbElements := []attr.Value{}
		for _, clb := range ng.ClassicLoadBalancerNames {
			clbElements = append(clbElements, types.StringValue(clb))
		}
		classicLoadBalancerNames, d = types.ListValue(types.StringType, clbElements)
		diags = append(diags, d...)
	}
	v.ClassicLoadBalancerNames = classicLoadBalancerNames

	v.CpuCredits = types.StringValue(ng.CPUCredits)
	v.EnableDetailedMonitoring = types.BoolPointerValue(ng.EnableDetailedMonitoring)
	v.InstanceType = types.StringValue(ng.InstanceType)

	availabilityZones2 := types.ListNull(types.StringType)
	if len(ng.AvailabilityZones) > 0 {
		azElements := []attr.Value{}
		for _, az := range ng.AvailabilityZones {
			azElements = append(azElements, types.StringValue(az))
		}
		availabilityZones2, d = types.ListValue(types.StringType, azElements)
		diags = append(diags, d...)
	}
	v.AvailabilityZones2 = availabilityZones2

	subnets := types.ListNull(types.StringType)
	if len(ng.Subnets) > 0 {
		snElements := []attr.Value{}
		for _, sn := range ng.Subnets {
			snElements = append(snElements, types.StringValue(sn))
		}
		subnets, d = types.ListValue(types.StringType, snElements)
		diags = append(diags, d...)
	}
	v.Subnets = subnets

	v.InstancePrefix = types.StringValue(ng.InstancePrefix)
	v.InstanceName = types.StringValue(ng.InstanceName)

	lbsMap := types.MapNull(types.StringType)
	if len(ng.Labels) != 0 {
		lbs := map[string]attr.Value{}
		for key, val := range ng.Labels {
			lbs[key] = types.StringValue(val)
		}
		lbsMap, d = types.MapValue(types.StringType, lbs)
		diags = append(diags, d...)
	}
	v.Labels2 = lbsMap

	tagMap := types.MapNull(types.StringType)
	if len(ng.Tags) != 0 {
		tag := map[string]attr.Value{}
		for key, val := range ng.Tags {
			tag[key] = types.StringValue(val)
		}
		tagMap, d = types.MapValue(types.StringType, tag)
		diags = append(diags, d...)
	}
	v.Tags2 = tagMap

	v.Ami = types.StringValue(ng.AMI)

	iam := NewIamValueNull()
	d = iam.Flatten(ctx, ng.IAM)
	diags = append(diags, d...)
	iamElements := []attr.Value{iam}
	v.Iam, d = types.ListValue(IamValue{}.Type(ctx), iamElements)
	diags = append(diags, d...)

	ssh := NewSshValueNull()
	d = ssh.Flatten(ctx, ng.SSH)
	diags = append(diags, d...)
	v.Ssh, d = types.ListValue(SshValue{}.Type(ctx), []attr.Value{ssh})
	diags = append(diags, d...)

	placement := NewPlacementValueNull()
	d = placement.Flatten(ctx, ng.Placement)
	diags = append(diags, d...)
	v.Placement, d = types.ListValue(PlacementValue{}.Type(ctx), []attr.Value{placement})
	diags = append(diags, d...)

	instanceSel := NewInstanceSelectorValueNull()
	d = instanceSel.Flatten(ctx, ng.InstanceSelector)
	diags = append(diags, d...)
	v.InstanceSelector, d = types.ListValue(InstanceSelectorValue{}.Type(ctx), []attr.Value{instanceSel})
	diags = append(diags, d...)

	bottlerkt := NewBottleRocketValueNull()
	d = bottlerkt.Flatten(ctx, ng.Bottlerocket)
	diags = append(diags, d...)
	v.BottleRocket, d = types.ListValue(BottleRocketValue{}.Type(ctx), []attr.Value{bottlerkt})
	diags = append(diags, d...)

	instDistribution := NewInstancesDistributionValueNull()
	d = instDistribution.Flatten(ctx, ng.InstancesDistribution)
	diags = append(diags, d...)
	v.InstancesDistribution, d = types.ListValue(InstancesDistributionValue{}.Type(ctx), []attr.Value{instDistribution})
	diags = append(diags, d...)

	asgMetricsCollection := types.ListNull(AsgMetricsCollectionValue{}.Type(ctx))
	if len(ng.ASGMetricsCollection) > 0 {
		amcList := []attr.Value{}
		for _, val := range ng.ASGMetricsCollection {
			amc := NewAsgMetricsCollectionValueNull()
			d = amc.Flatten(ctx, val)
			diags = append(diags, d...)
			amcList = append(amcList, amc)
		}
		asgMetricsCollection, d = types.ListValue(AsgMetricsCollectionValue{}.Type(ctx), amcList)
		diags = append(diags, d...)
	}
	v.AsgMetricsCollection = asgMetricsCollection

	taints := types.ListNull(TaintsValue{}.Type(ctx))
	if len(ng.Taints) > 0 {
		taintsList := []attr.Value{}
		for _, val := range ng.Taints {
			taint := NewTaintsValueNull()
			d = taint.Flatten(ctx, val)
			diags = append(diags, d...)
			taintsList = append(taintsList, taint)
		}
		taints, d = types.ListValue(TaintsValue{}.Type(ctx), taintsList)
		diags = append(diags, d...)
	}
	v.Taints = taints

	updateConfig := NewUpdateConfigValueNull()
	d = updateConfig.Flatten(ctx, ng.UpdateConfig)
	diags = append(diags, d...)
	v.UpdateConfig, d = types.ListValue(UpdateConfigValue{}.Type(ctx), []attr.Value{updateConfig})
	diags = append(diags, d...)

	kubeletExtraConfig := NewKubeletExtraConfigValueNull()
	d = kubeletExtraConfig.Flatten(ctx, ng.KubeletExtraConfig)
	diags = append(diags, d...)
	v.KubeletExtraConfig, d = types.ListValue(KubeletExtraConfigValue{}.Type(ctx), []attr.Value{kubeletExtraConfig})
	diags = append(diags, d...)

	secGroup := NewSecurityGroups2ValueNull()
	d = secGroup.Flatten(ctx, ng.SecurityGroups)
	diags = append(diags, d...)
	v.SecurityGroups2, d = types.ListValue(SecurityGroups2Value{}.Type(ctx), []attr.Value{secGroup})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *KubeletExtraConfigValue) Flatten(ctx context.Context, in *rafay.KubeletExtraConfig) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	kubeReserved := types.MapNull(types.StringType)
	if len(in.KubeReserved) > 0 {
		kr := map[string]attr.Value{}
		for key, val := range in.KubeReserved {
			kr[key] = types.StringValue(val)
		}
		kubeReserved, d = types.MapValue(types.StringType, kr)
		diags = append(diags, d...)
	}
	v.KubeReserved = kubeReserved
	v.KubeReservedCgroup = types.StringValue(in.KubeReservedCGroup)

	sysReserved := types.MapNull(types.StringType)
	if len(in.SystemReserved) > 0 {
		sr := map[string]attr.Value{}
		for key, val := range in.SystemReserved {
			sr[key] = types.StringValue(val)
		}
		sysReserved, d = types.MapValue(types.StringType, sr)
		diags = append(diags, d...)
	}
	v.SystemReserved = sysReserved

	evictionHard := types.MapNull(types.StringType)
	if len(in.EvictionHard) > 0 {
		eh := map[string]attr.Value{}
		for key, val := range in.EvictionHard {
			eh[key] = types.StringValue(val)
		}
		evictionHard, d = types.MapValue(types.StringType, eh)
		diags = append(diags, d...)
	}
	v.EvictionHard = evictionHard

	featureGates := types.MapNull(types.BoolType)
	if len(in.FeatureGates) > 0 {
		fg := map[string]attr.Value{}
		for key, val := range in.FeatureGates {
			fg[key] = types.BoolValue(val)
		}
		featureGates, d = types.MapValue(types.BoolType, fg)
		diags = append(diags, d...)
	}
	v.FeatureGates = featureGates

	v.state = attr.ValueStateKnown
	return diags
}

func (v *UpdateConfigValue) Flatten(ctx context.Context, in *rafay.NodeGroupUpdateConfig) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	if in.MaxUnavailable != nil {
		v.MaxUnavaliable = types.Int64Value(int64(*in.MaxUnavailable))
	}
	if in.MaxUnavailablePercentage != nil {
		v.MaxUnavaliablePercetage = types.Int64Value(int64(*in.MaxUnavailablePercentage))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *TaintsValue) Flatten(ctx context.Context, in rafay.NodeGroupTaint) diag.Diagnostics {
	var diags diag.Diagnostics

	v.Key = types.StringValue(in.Key)
	v.Value = types.StringValue(in.Value)
	v.Effect = types.StringValue(in.Effect)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AsgMetricsCollectionValue) Flatten(ctx context.Context, in rafay.MetricsCollection) diag.Diagnostics {
	var diags, d diag.Diagnostics

	v.Granularity = types.StringValue(in.Granularity)

	metrics := types.ListNull(types.StringType)
	if len(in.Metrics) > 0 {
		metricsList := []attr.Value{}
		for _, metric := range in.Metrics {
			metricsList = append(metricsList, types.StringValue(metric))
		}
		metrics, d = types.ListValue(types.StringType, metricsList)
		diags = append(diags, d...)
	}
	v.Metrics = metrics

	v.state = attr.ValueStateKnown
	return diags
}

func (v *InstancesDistributionValue) Flatten(ctx context.Context, in *rafay.NodeGroupInstancesDistribution) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if in == nil {
		return diags
	}

	instanceTypes := types.ListNull(types.StringType)
	if len(in.InstanceTypes) > 0 {
		instanceTypesList := []attr.Value{}
		for _, it := range in.InstanceTypes {
			instanceTypesList = append(instanceTypesList, types.StringValue(it))
		}
		instanceTypes, d = types.ListValue(types.StringType, instanceTypesList)
		diags = append(diags, d...)
	}
	v.InstanceTypes = instanceTypes

	v.MaxPrice = types.Float64PointerValue(in.MaxPrice)
	if in.OnDemandBaseCapacity != nil {
		v.OnDemandBaseCapacity = types.Int64Value(int64(*in.OnDemandBaseCapacity))
	}

	if in.OnDemandPercentageAboveBaseCapacity != nil {
		v.OnDemandPercentageAboveBaseCapacity = types.Int64Value(int64(*in.OnDemandPercentageAboveBaseCapacity))
	}
	if in.SpotInstancePools != nil {
		v.SpotInstancePools = types.Int64Value(int64(*in.SpotInstancePools))
	}
	v.SpotAllocationStrategy = types.StringValue(in.SpotAllocationStrategy)
	v.CapacityRebalance = types.BoolPointerValue(in.CapacityRebalance)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *BottleRocketValue) Flatten(ctx context.Context, in *rafay.NodeGroupBottlerocket) diag.Diagnostics {
	var diags diag.Diagnostics
	if in == nil {
		return diags
	}

	v.EnableAdminContainer = types.BoolPointerValue(in.EnableAdminContainer)

	if in.Settings != nil && len(in.Settings) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(in.Settings)
		if err != nil {
			diags.AddError("Bottle rocket marshal error", err.Error())
		}
		v.Settings = types.StringValue(string(jsonStr))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *InstanceSelectorValue) Flatten(ctx context.Context, instanceSel *rafay.InstanceSelector) diag.Diagnostics {
	var diags diag.Diagnostics

	if instanceSel == nil {
		return diags
	}

	if instanceSel.VCPUs != nil {
		v.Vcpus = types.Int64Value(int64(*instanceSel.VCPUs))
	}
	v.Memory = types.StringValue(instanceSel.Memory)

	if instanceSel.GPUs != nil {
		v.Gpus = types.Int64Value(int64(*instanceSel.GPUs))
	}
	v.CpuArchitecture = types.StringValue(instanceSel.CPUArchitecture)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *PlacementValue) Flatten(ctx context.Context, placement *rafay.Placement) diag.Diagnostics {
	var diags diag.Diagnostics

	if placement == nil {
		return diags
	}

	v.Group = types.StringValue(placement.GroupName)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SshValue) Flatten(ctx context.Context, ssh *rafay.NodeGroupSSH) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if ssh == nil {
		return diags
	}

	v.Allow = types.BoolPointerValue(ssh.Allow)
	v.PublicKey = types.StringValue(ssh.PublicKey)
	v.PublicKeyName = types.StringValue(ssh.PublicKeyName)

	sourceSecurityGroupIds := types.ListNull(types.StringType)
	if len(ssh.SourceSecurityGroupIDs) > 0 {
		ids := []attr.Value{}
		for _, id := range ssh.SourceSecurityGroupIDs {
			ids = append(ids, types.StringValue(id))
		}
		sourceSecurityGroupIds, d = types.ListValue(types.StringType, ids)
		diags = append(diags, d...)
	}
	v.SourceSecurityGroupIds = sourceSecurityGroupIds

	v.EnableSsm = types.BoolPointerValue(ssh.EnableSSM)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SecurityGroups2Value) Flatten(ctx context.Context, sg *rafay.NodeGroupSGs) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if sg == nil {
		return diags
	}

	v.WithShared = types.BoolPointerValue(sg.WithShared)
	v.WithLocal = types.BoolPointerValue(sg.WithLocal)

	attachIds := types.ListNull(types.StringType)
	if len(sg.AttachIDs) > 0 {
		aidsElements := []attr.Value{}
		for _, aid := range sg.AttachIDs {
			aidsElements = append(aidsElements, types.StringValue(aid))
		}
		attachIds, d = types.ListValue(types.StringType, aidsElements)
		diags = append(diags, d...)
	}
	v.AttachIds = attachIds

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamValue) Flatten(ctx context.Context, iam *rafay.NodeGroupIAM) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if iam == nil {
		return diags
	}

	// TODO(Akshay): Check if Attach Policy v2. based on that populate attach_policy and attach_policy_v2
	if iam.AttachPolicy != nil {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonBytes, err := json2.Marshal(iam.AttachPolicy)
		if err != nil {
			diags.AddError("Attach Policy Marshal Error", err.Error())
		}
		v.AttachPolicyV2 = types.StringValue(string(jsonBytes))
	}

	attachPolicy := NewAttachPolicyValueNull()
	d = attachPolicy.Flatten(ctx, iam.AttachPolicy)
	diags = append(diags, d...)
	v.AttachPolicy, d = types.ListValue(AttachPolicyValue{}.Type(ctx), []attr.Value{attachPolicy})
	diags = append(diags, d...)

	attachPolicyArns := types.ListNull(types.StringType)
	if len(iam.AttachPolicyARNs) > 0 {
		arns := []attr.Value{}
		for _, arn := range iam.AttachPolicyARNs {
			arns = append(arns, types.StringValue(arn))
		}
		attachPolicyArns, d = types.ListValue(types.StringType, arns)
		diags = append(diags, d...)
	}
	v.AttachPolicyArns = attachPolicyArns

	v.InstanceProfileArn = types.StringValue(iam.InstanceProfileARN)
	v.InstanceRoleArn = types.StringValue(iam.InstanceRoleARN)
	v.InstanceRoleName = types.StringValue(iam.InstanceRoleName)
	v.InstanceRolePermissionBoundary = types.StringValue(iam.InstanceRolePermissionsBoundary)

	addonPolicies := NewIamNodeGroupWithAddonPoliciesValueNull()
	d = addonPolicies.Flatten(ctx, iam.WithAddonPolicies)
	diags = append(diags, d...)
	addonPoliciesElements := []attr.Value{addonPolicies}
	v.IamNodeGroupWithAddonPolicies, d = types.ListValue(IamNodeGroupWithAddonPoliciesValue{}.Type(ctx), addonPoliciesElements)
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *AttachPolicyValue) Flatten(ctx context.Context, attachpol *rafay.InlineDocument) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if attachpol == nil {
		return diags
	}

	v.Version = types.StringValue(attachpol.Version)
	v.Id = types.StringValue(attachpol.Id)

	statement := types.ListNull(StatementValue{}.Type(ctx))
	if len(attachpol.Statement) > 0 {
		stms := []attr.Value{}
		for _, stm := range attachpol.Statement {
			sv := NewStatementValueNull()
			d = sv.Flatten(ctx, stm)
			diags = append(diags, d...)
			stms = append(stms, sv)
		}
		statement, d = types.ListValue(StatementValue{}.Type(ctx), stms)
		diags = append(diags, d...)
	}
	v.Statement = statement

	v.state = attr.ValueStateKnown
	return diags
}

func (v *StatementValue) Flatten(ctx context.Context, stm rafay.InlineStatement) diag.Diagnostics {
	var diags, d diag.Diagnostics

	if len(stm.Effect) > 0 {
		v.Effect = types.StringValue(stm.Effect)
	}
	if len(stm.Sid) > 0 {
		v.Sid = types.StringValue(stm.Sid)
	}
	if stm.Action != nil && len(stm.Action.([]interface{})) > 0 {
		actEle := []attr.Value{}
		for _, act := range stm.Action.([]interface{}) {
			actEle = append(actEle, types.StringValue(act.(string)))
		}
		v.Action, d = types.ListValue(types.StringType, actEle)
		diags = append(diags, d...)
	}
	if stm.NotAction != nil && len(stm.NotAction.([]interface{})) > 0 {
		naEle := []attr.Value{}
		for _, na := range stm.NotAction.([]interface{}) {
			naEle = append(naEle, types.StringValue(na.(string)))
		}
		v.NotAction, d = types.ListValue(types.StringType, naEle)
		diags = append(diags, d...)
	}
	if len(stm.Resource.(string)) > 0 {
		v.Resource = types.StringValue(stm.Resource.(string))
	}
	if stm.NotResource != nil && len(stm.NotResource.([]interface{})) > 0 {
		nrEle := []attr.Value{}
		for _, nr := range stm.NotResource.([]interface{}) {
			nrEle = append(nrEle, types.StringValue(nr.(string)))
		}
		v.NotResource, d = types.ListValue(types.StringType, nrEle)
		diags = append(diags, d...)
	}

	if len(stm.Condition) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Condition)
		if err != nil {
			diags.AddError("condition marshal error", err.Error())
		}
		v.Condition = types.StringValue(string(jsonStr))
	}

	if len(stm.Principal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.Principal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.Principal = types.StringValue(string(jsonStr))

	}

	if len(stm.NotPrincipal) > 0 {
		var json2 = jsoniter.ConfigCompatibleWithStandardLibrary
		jsonStr, err := json2.Marshal(stm.NotPrincipal)
		if err != nil {
			log.Println("attach policy marshal err:", err)
		}
		v.NotPrincipal = types.StringValue(string(jsonStr))
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *IamNodeGroupWithAddonPoliciesValue) Flatten(ctx context.Context, addonPolicies *rafay.NodeGroupIAMAddonPolicies) diag.Diagnostics {
	var diags diag.Diagnostics
	if addonPolicies == nil {
		return diags
	}

	v.AlbIngress = types.BoolPointerValue(addonPolicies.AWSLoadBalancerController)
	v.AppMesh = types.BoolPointerValue(addonPolicies.AppMesh)
	v.AppMeshReview = types.BoolPointerValue(addonPolicies.AppMeshPreview)
	v.AutoScaler = types.BoolPointerValue(addonPolicies.AutoScaler)
	v.CertManager = types.BoolPointerValue(addonPolicies.CertManager)
	v.CloudWatch = types.BoolPointerValue(addonPolicies.CloudWatch)
	v.Ebs = types.BoolPointerValue(addonPolicies.EBS)
	v.Efs = types.BoolPointerValue(addonPolicies.EFS)
	v.ExternalDns = types.BoolPointerValue(addonPolicies.ExternalDNS)
	v.Fsx = types.BoolPointerValue(addonPolicies.FSX)
	v.ImageBuilder = types.BoolPointerValue(addonPolicies.ImageBuilder)
	v.Xray = types.BoolPointerValue(addonPolicies.XRay)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SystemComponentsPlacementValue) Flatten(ctx context.Context, scp *rafay.SystemComponentsPlacement) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if scp == nil {
		return diags
	}

	ns := types.MapNull(types.StringType)
	if len(scp.NodeSelector) != 0 {
		nodeSelector := map[string]attr.Value{}
		for key, val := range scp.NodeSelector {
			nodeSelector[key] = types.StringValue(val)
		}
		ns, d = types.MapValue(types.StringType, nodeSelector)
		diags = append(diags, d...)
	}
	v.NodeSelector = ns

	tolerations := types.ListNull(TolerationsValue{}.Type(ctx))
	if len(scp.Tolerations) > 0 {
		tolerationsList := make([]attr.Value, 0, len(scp.Tolerations))
		for _, t := range scp.Tolerations {
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
	d = daemonsetOverride.Flatten(ctx, scp.DaemonsetOverride)
	diags = append(diags, d...)
	v.DaemonsetOverride, d = types.ListValue(DaemonsetOverrideValue{}.Type(ctx), []attr.Value{daemonsetOverride})
	diags = append(diags, d...)

	v.state = attr.ValueStateKnown
	return diags
}

func (v *TolerationsValue) Flatten(ctx context.Context, t *rafay.Tolerations) diag.Diagnostics {
	var diags diag.Diagnostics
	if t == nil {
		return diags
	}

	v.Key = types.StringValue(t.Key)
	v.Operator = types.StringValue(t.Operator)
	v.Value = types.StringValue(t.Value)
	v.Effect = types.StringValue(t.Effect)
	if t.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64Value(int64(*t.TolerationSeconds))
	} else {
		v.TolerationSeconds = types.Int64Null()
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *DaemonsetOverrideValue) Flatten(ctx context.Context, dso *rafay.DaemonsetOverride) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if dso == nil {
		return diags
	}

	v.NodeSelectionEnabled = types.BoolPointerValue(dso.NodeSelectionEnabled)

	tolerations2 := types.ListNull(Tolerations2Value{}.Type(ctx))
	if len(dso.Tolerations) > 0 {
		tolerationsList := make([]attr.Value, 0, len(dso.Tolerations))
		for _, t := range dso.Tolerations {
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

func (v *Tolerations2Value) Flatten(ctx context.Context, t *rafay.Tolerations) diag.Diagnostics {
	var diags diag.Diagnostics
	if t == nil {
		return diags
	}

	v.Key = types.StringValue(t.Key)
	v.Operator = types.StringValue(t.Operator)
	v.Value = types.StringValue(t.Value)
	v.Effect = types.StringValue(t.Effect)
	if t.TolerationSeconds != nil {
		v.TolerationSeconds = types.Int64Value(int64(*t.TolerationSeconds))
	} else {
		v.TolerationSeconds = types.Int64Null()
	}

	v.state = attr.ValueStateKnown
	return diags
}

func (v *SharingValue) Flatten(ctx context.Context, sh *rafay.V1ClusterSharing) diag.Diagnostics {
	var diags, d diag.Diagnostics
	if sh == nil {
		return diags
	}

	v.Enabled = types.BoolPointerValue(sh.Enabled)

	projects := types.ListNull(ProjectsValue{}.Type(ctx))
	if len(sh.Projects) > 0 {
		projectsList := make([]attr.Value, 0, len(sh.Projects))
		for _, p := range sh.Projects {
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

func (v *ProjectsValue) Flatten(ctx context.Context, p *rafay.V1ClusterSharingProject) diag.Diagnostics {
	var diags diag.Diagnostics
	if p == nil {
		return diags
	}

	v.Name = types.StringValue(p.Name)

	v.state = attr.ValueStateKnown
	return diags
}
