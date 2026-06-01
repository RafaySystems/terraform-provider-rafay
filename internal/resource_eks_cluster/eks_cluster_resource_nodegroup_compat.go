package resource_eks_cluster

import (
	"context"
	"os"
	"strings"

	"github.com/RafaySystems/terraform-provider-rafay/rafay"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type clusterConfigExpandOpts struct {
	useNodeGroupsMap        bool
	useManagedNodeGroupsMap bool
}

func clusterConfigExpandOptsFromConfig(cc ClusterConfigValue) clusterConfigExpandOpts {
	return clusterConfigExpandOpts{
		useNodeGroupsMap:        clusterConfigUsesNodeGroupsMap(cc),
		useManagedNodeGroupsMap: clusterConfigUsesManagedNodeGroupsMap(cc),
	}
}

func clusterConfigUsesNodeGroupsMap(cc ClusterConfigValue) bool {
	return !cc.NodeGroupsMap.IsNull() && !cc.NodeGroupsMap.IsUnknown() && len(cc.NodeGroupsMap.Elements()) > 0
}

func clusterConfigUsesManagedNodeGroupsMap(cc ClusterConfigValue) bool {
	return !cc.ManagedNodegroupsMap.IsNull() && !cc.ManagedNodegroupsMap.IsUnknown() && len(cc.ManagedNodegroupsMap.Elements()) > 0
}

func clusterConfigUsesNodeGroupsList(cc ClusterConfigValue) bool {
	return !cc.NodeGroups.IsNull() && !cc.NodeGroups.IsUnknown() && len(cc.NodeGroups.Elements()) > 0
}

func clusterConfigUsesManagedNodeGroupsList(cc ClusterConfigValue) bool {
	return !cc.ManagedNodegroups.IsNull() && !cc.ManagedNodegroups.IsUnknown() && len(cc.ManagedNodegroups.Elements()) > 0
}

func isNodeGroupsMapMode(state ClusterConfigValue) bool {
	switch os.Getenv("TF_RAFAY_EKS_MIGRATE_TO_MAP") {
	case "true":
		return true
	case "false":
		return false
	}

	if clusterConfigUsesNodeGroupsList(state) {
		return false
	}

	return clusterConfigUsesNodeGroupsMap(state)
}

func isManagedNodeGroupsMapMode(state ClusterConfigValue) bool {
	switch os.Getenv("TF_RAFAY_EKS_MIGRATE_TO_MAP") {
	case "true":
		return true
	case "false":
		return false
	}

	if clusterConfigUsesManagedNodeGroupsList(state) {
		return false
	}

	return clusterConfigUsesManagedNodeGroupsMap(state)
}

// Exported test helpers and wrappers for external tests.
func NormalizeNodeGroupsPlanList(ctx context.Context, configList, stateList basetypes.ListValue) (basetypes.ListValue, diag.Diagnostics) {
	return normalizeNodeGroupsPlanList(ctx, configList, stateList)
}

func NormalizeManagedNodeGroupsPlanList(ctx context.Context, configList, stateList basetypes.ListValue) (basetypes.ListValue, diag.Diagnostics) {
	return normalizeManagedNodeGroupsPlanList(ctx, configList, stateList)
}

func OrderNodeGroupsFromState(ctx context.Context, apiGroups []*rafay.NodeGroup, stateList basetypes.ListValue) ([]*rafay.NodeGroup, diag.Diagnostics) {
	return orderNodeGroupsFromState(ctx, apiGroups, stateList)
}

func OrderManagedNodeGroupsFromState(ctx context.Context, apiGroups []*rafay.ManagedNodeGroup, stateList basetypes.ListValue) ([]*rafay.ManagedNodeGroup, diag.Diagnostics) {
	return orderManagedNodeGroupsFromState(ctx, apiGroups, stateList)
}

func AppendStateOnlyNodeGroupsListElements(ngElements []attr.Value, apiNames map[string]bool, stNgs []NodeGroupsValue) []attr.Value {
	return appendStateOnlyNodeGroupsListElements(ngElements, apiNames, stNgs)
}

func AppendStateOnlyManagedNodeGroupsListElements(mngElements []attr.Value, apiNames map[string]bool, stMngs []ManagedNodegroupsValue) []attr.Value {
	return appendStateOnlyManagedNodeGroupsListElements(mngElements, apiNames, stMngs)
}

func NewTestNodeGroupsValue(name string, desired int64) NodeGroupsValue {
	desiredInt := int(desired)
	ng := &rafay.NodeGroup{
		Name:             name,
		InstanceType:     "t3.medium",
		DesiredCapacity:  &desiredInt,
	}
	v := NewNodeGroupsValueNull()
	_ = v.Flatten(context.Background(), ng, NodeGroupsValue{})
	return v
}

func NewTestManagedNodegroupsValue(name string, desired int64) ManagedNodegroupsValue {
	desiredInt := int(desired)
	mng := &rafay.ManagedNodeGroup{
		Name:            name,
		InstanceType:    "t3.medium",
		DesiredCapacity: &desiredInt,
	}
	v := NewManagedNodegroupsValueNull()
	_ = v.Flatten(context.Background(), mng, ManagedNodegroupsValue{})
	return v
}

func NewTestClusterConfigValueWithManaged(ctx context.Context, managedNodegroups basetypes.ListValue) ClusterConfigValue {
	cc := NewTestClusterConfigValue(ctx, types.ListNull(NodeGroupsValue{}.Type(ctx)))
	cc.ManagedNodegroups = managedNodegroups
	return cc
}

func NewTestClusterConfigValue(ctx context.Context, nodeGroups basetypes.ListValue) ClusterConfigValue {
	cc := ClusterConfigValue{
		AccessConfig:            types.ListNull(AccessConfigValue{}.Type(ctx)),
		Addons:                    types.SetNull(AddonsValue{}.Type(ctx)),
		AddonsConfig:              types.ListNull(AddonsConfigValue{}.Type(ctx)),
		AutoModeConfig:            types.ListNull(AutoModeConfigValue{}.Type(ctx)),
		AutoZonalShiftConfig:      types.ListNull(AutoZonalShiftConfigValue{}.Type(ctx)),
		AvailabilityZones:         types.ListNull(types.StringType),
		CloudWatch:                types.ListNull(CloudWatchValue{}.Type(ctx)),
		DeleteProtectionConfig:    types.ListNull(DeleteProtectionConfigValue{}.Type(ctx)),
		FargateProfiles:           types.ListNull(FargateProfilesValue{}.Type(ctx)),
		Iam3:                      types.ListNull(Iam3Value{}.Type(ctx)),
		IdentityMappings:          types.ListNull(IdentityMappingsValue{}.Type(ctx)),
		IdentityProviders:         types.ListNull(IdentityProvidersValue{}.Type(ctx)),
		KubernetesNetworkConfig:   types.ListNull(KubernetesNetworkConfigValue{}.Type(ctx)),
		ManagedNodegroups:         types.ListNull(ManagedNodegroupsValue{}.Type(ctx)),
		ManagedNodegroupsMap:      types.MapNull(ManagedNodegroupsMapValue{}.Type(ctx)),
		Metadata2:                 types.ListNull(Metadata2Value{}.Type(ctx)),
		NodeGroups:                nodeGroups,
		NodeGroupsMap:             types.MapNull(NodeGroupsMapValue{}.Type(ctx)),
		PrivateCluster:            types.ListNull(PrivateClusterValue{}.Type(ctx)),
		SecretsEncryption:         types.ListNull(SecretsEncryptionValue{}.Type(ctx)),
		Vpc:                       types.ListNull(VpcValue{}.Type(ctx)),
		ZonalShiftConfig:          types.ListNull(ZonalShiftConfigValue{}.Type(ctx)),
		state:                     attr.ValueStateKnown,
	}
	return cc
}

func normalizeNodeGroupName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func nodeGroupNameFromValue(name basetypes.StringValue) string {
	return normalizeNodeGroupName(getStringValue(name))
}

func managedNodeGroupNameFromValue(name basetypes.StringValue) string {
	return normalizeNodeGroupName(getStringValue(name))
}

func normalizeNodeGroupsPlanList(ctx context.Context, configList, stateList basetypes.ListValue) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if configList.IsNull() || configList.IsUnknown() {
		return configList, diags
	}

	configElems := make([]NodeGroupsValue, 0, len(configList.Elements()))
	if d := configList.ElementsAs(ctx, &configElems, false); d.HasError() {
		return configList, append(diags, d...)
	}

	if stateList.IsNull() || stateList.IsUnknown() || len(stateList.Elements()) == 0 {
		return configList, diags
	}

	stateElems := make([]NodeGroupsValue, 0, len(stateList.Elements()))
	if d := stateList.ElementsAs(ctx, &stateElems, false); d.HasError() {
		return configList, append(diags, d...)
	}

	configByName := make(map[string]NodeGroupsValue, len(configElems))
	for _, ng := range configElems {
		name := nodeGroupNameFromValue(ng.Name)
		if name == "" {
			continue
		}
		configByName[name] = ng
	}

	stateByName := make(map[string]NodeGroupsValue, len(stateElems))
	for _, ng := range stateElems {
		name := nodeGroupNameFromValue(ng.Name)
		if name == "" {
			continue
		}
		stateByName[name] = ng
	}

	if len(configByName) != len(stateByName) {
		if isAddOnlyNodeGroups(configByName, stateByName) &&
			newNodeGroupNamesOnlyAfterExisting(configElems, stateByName) &&
			existingNodeGroupNamesMatchStateOrder(configElems, stateElems, stateByName) {
			return buildAppendNewNodeGroupsPlan(ctx, stateElems, configElems, stateByName)
		}
		return buildPlanListInConfigOrder(ctx, configElems, stateByName, NodeGroupsValue{}.Type(ctx))
	}

	reorderOnly := true
	for name, cfg := range configByName {
		st, ok := stateByName[name]
		if !ok {
			reorderOnly = false
			break
		}
		if !cfg.Equal(st) {
			reorderOnly = false
			break
		}
	}

	if reorderOnly {
		return stateList, diags
	}

	return buildPlanListInConfigOrder(ctx, configElems, stateByName, NodeGroupsValue{}.Type(ctx))
}

func buildPlanListInConfigOrder(ctx context.Context, configElems []NodeGroupsValue, stateByName map[string]NodeGroupsValue, elemType attr.Type) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := make([]attr.Value, 0, len(configElems))
	for _, cfg := range configElems {
		name := nodeGroupNameFromValue(cfg.Name)
		if name == "" {
			result = append(result, cfg)
			continue
		}
		if st, ok := stateByName[name]; ok && cfg.Equal(st) {
			result = append(result, cfg)
			continue
		}
		result = append(result, cfg)
	}

	listVal, d := types.ListValue(elemType, result)
	diags = append(diags, d...)
	return listVal, diags
}

func isAddOnlyNodeGroups(configByName, stateByName map[string]NodeGroupsValue) bool {
	if len(configByName) <= len(stateByName) {
		return false
	}
	for name, st := range stateByName {
		cfg, ok := configByName[name]
		if !ok || !cfg.Equal(st) {
			return false
		}
	}
	return true
}

func isAddOnlyManagedNodeGroups(configByName, stateByName map[string]ManagedNodegroupsValue) bool {
	if len(configByName) <= len(stateByName) {
		return false
	}
	for name, st := range stateByName {
		cfg, ok := configByName[name]
		if !ok || !cfg.Equal(st) {
			return false
		}
	}
	return true
}

func newNodeGroupNamesOnlyAfterExisting(configElems []NodeGroupsValue, stateByName map[string]NodeGroupsValue) bool {
	sawNew := false
	for _, cfg := range configElems {
		name := nodeGroupNameFromValue(cfg.Name)
		if name == "" {
			continue
		}
		if _, exists := stateByName[name]; exists {
			if sawNew {
				return false
			}
			continue
		}
		sawNew = true
	}
	return true
}

func newManagedNodeGroupNamesOnlyAfterExisting(configElems []ManagedNodegroupsValue, stateByName map[string]ManagedNodegroupsValue) bool {
	sawNew := false
	for _, cfg := range configElems {
		name := managedNodeGroupNameFromValue(cfg.Name)
		if name == "" {
			continue
		}
		if _, exists := stateByName[name]; exists {
			if sawNew {
				return false
			}
			continue
		}
		sawNew = true
	}
	return true
}

func existingNodeGroupNamesMatchStateOrder(configElems []NodeGroupsValue, stateElems []NodeGroupsValue, stateByName map[string]NodeGroupsValue) bool {
	stateOrder := make([]string, 0, len(stateElems))
	for _, st := range stateElems {
		name := nodeGroupNameFromValue(st.Name)
		if name != "" {
			stateOrder = append(stateOrder, name)
		}
	}

	seen := make([]string, 0, len(stateOrder))
	for _, cfg := range configElems {
		name := nodeGroupNameFromValue(cfg.Name)
		if name == "" {
			continue
		}
		if _, exists := stateByName[name]; exists {
			seen = append(seen, name)
		}
	}

	if len(seen) != len(stateOrder) {
		return false
	}
	for i := range stateOrder {
		if seen[i] != stateOrder[i] {
			return false
		}
	}
	return true
}

func existingManagedNodeGroupNamesMatchStateOrder(configElems []ManagedNodegroupsValue, stateElems []ManagedNodegroupsValue, stateByName map[string]ManagedNodegroupsValue) bool {
	stateOrder := make([]string, 0, len(stateElems))
	for _, st := range stateElems {
		name := managedNodeGroupNameFromValue(st.Name)
		if name != "" {
			stateOrder = append(stateOrder, name)
		}
	}

	seen := make([]string, 0, len(stateOrder))
	for _, cfg := range configElems {
		name := managedNodeGroupNameFromValue(cfg.Name)
		if name == "" {
			continue
		}
		if _, exists := stateByName[name]; exists {
			seen = append(seen, name)
		}
	}

	if len(seen) != len(stateOrder) {
		return false
	}
	for i := range stateOrder {
		if seen[i] != stateOrder[i] {
			return false
		}
	}
	return true
}

func buildAppendNewNodeGroupsPlan(
	ctx context.Context,
	stateElems []NodeGroupsValue,
	configElems []NodeGroupsValue,
	stateByName map[string]NodeGroupsValue,
) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := make([]attr.Value, 0, len(configElems))
	for _, st := range stateElems {
		name := nodeGroupNameFromValue(st.Name)
		if name == "" {
			continue
		}
		result = append(result, st)
	}
	for _, cfg := range configElems {
		name := nodeGroupNameFromValue(cfg.Name)
		if name == "" {
			continue
		}
		if _, exists := stateByName[name]; !exists {
			result = append(result, cfg)
		}
	}

	listVal, d := types.ListValue(NodeGroupsValue{}.Type(ctx), result)
	diags = append(diags, d...)
	return listVal, diags
}

func buildAppendNewManagedNodeGroupsPlan(
	ctx context.Context,
	stateElems []ManagedNodegroupsValue,
	configElems []ManagedNodegroupsValue,
	stateByName map[string]ManagedNodegroupsValue,
) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := make([]attr.Value, 0, len(configElems))
	for _, st := range stateElems {
		name := managedNodeGroupNameFromValue(st.Name)
		if name == "" {
			continue
		}
		result = append(result, st)
	}
	for _, cfg := range configElems {
		name := managedNodeGroupNameFromValue(cfg.Name)
		if name == "" {
			continue
		}
		if _, exists := stateByName[name]; !exists {
			result = append(result, cfg)
		}
	}

	listVal, d := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), result)
	diags = append(diags, d...)
	return listVal, diags
}

func normalizeManagedNodeGroupsPlanList(ctx context.Context, configList, stateList basetypes.ListValue) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	if configList.IsNull() || configList.IsUnknown() {
		return configList, diags
	}

	configElems := make([]ManagedNodegroupsValue, 0, len(configList.Elements()))
	if d := configList.ElementsAs(ctx, &configElems, false); d.HasError() {
		return configList, append(diags, d...)
	}

	if stateList.IsNull() || stateList.IsUnknown() || len(stateList.Elements()) == 0 {
		return configList, diags
	}

	stateElems := make([]ManagedNodegroupsValue, 0, len(stateList.Elements()))
	if d := stateList.ElementsAs(ctx, &stateElems, false); d.HasError() {
		return configList, append(diags, d...)
	}

	configByName := make(map[string]ManagedNodegroupsValue, len(configElems))
	for _, mng := range configElems {
		name := managedNodeGroupNameFromValue(mng.Name)
		if name == "" {
			continue
		}
		configByName[name] = mng
	}

	stateByName := make(map[string]ManagedNodegroupsValue, len(stateElems))
	for _, mng := range stateElems {
		name := managedNodeGroupNameFromValue(mng.Name)
		if name == "" {
			continue
		}
		stateByName[name] = mng
	}

	if len(configByName) != len(stateByName) {
		if isAddOnlyManagedNodeGroups(configByName, stateByName) &&
			newManagedNodeGroupNamesOnlyAfterExisting(configElems, stateByName) &&
			existingManagedNodeGroupNamesMatchStateOrder(configElems, stateElems, stateByName) {
			return buildAppendNewManagedNodeGroupsPlan(ctx, stateElems, configElems, stateByName)
		}
		return buildManagedPlanListInConfigOrder(ctx, configElems, stateByName)
	}

	reorderOnly := true
	for name, cfg := range configByName {
		st, ok := stateByName[name]
		if !ok {
			reorderOnly = false
			break
		}
		if !cfg.Equal(st) {
			reorderOnly = false
			break
		}
	}

	if reorderOnly {
		return stateList, diags
	}

	return buildManagedPlanListInConfigOrder(ctx, configElems, stateByName)
}

func buildManagedPlanListInConfigOrder(ctx context.Context, configElems []ManagedNodegroupsValue, stateByName map[string]ManagedNodegroupsValue) (basetypes.ListValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	result := make([]attr.Value, 0, len(configElems))
	for _, cfg := range configElems {
		name := managedNodeGroupNameFromValue(cfg.Name)
		if name == "" {
			result = append(result, cfg)
			continue
		}
		if st, ok := stateByName[name]; ok && cfg.Equal(st) {
			result = append(result, cfg)
			continue
		}
		result = append(result, cfg)
	}

	listVal, d := types.ListValue(ManagedNodegroupsValue{}.Type(ctx), result)
	diags = append(diags, d...)
	return listVal, diags
}

func orderNodeGroupsFromState(ctx context.Context, apiGroups []*rafay.NodeGroup, stateList basetypes.ListValue) ([]*rafay.NodeGroup, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(apiGroups) == 0 {
		return apiGroups, diags
	}

	apiByName := make(map[string]*rafay.NodeGroup, len(apiGroups))
	for _, ng := range apiGroups {
		if ng == nil {
			continue
		}
		apiByName[normalizeNodeGroupName(ng.Name)] = ng
	}

	if stateList.IsNull() || stateList.IsUnknown() || len(stateList.Elements()) == 0 {
		return apiGroups, diags
	}

	stateElems := make([]NodeGroupsValue, 0, len(stateList.Elements()))
	if d := stateList.ElementsAs(ctx, &stateElems, false); d.HasError() {
		return apiGroups, append(diags, d...)
	}

	ordered := make([]*rafay.NodeGroup, 0, len(apiGroups))
	seen := make(map[string]bool, len(apiGroups))

	for _, st := range stateElems {
		name := nodeGroupNameFromValue(st.Name)
		if name == "" {
			continue
		}
		if ng, ok := apiByName[name]; ok {
			ordered = append(ordered, ng)
			seen[name] = true
		}
	}

	for _, ng := range apiGroups {
		if ng == nil {
			continue
		}
		name := normalizeNodeGroupName(ng.Name)
		if !seen[name] {
			ordered = append(ordered, ng)
		}
	}

	return ordered, diags
}

func orderManagedNodeGroupsFromState(ctx context.Context, apiGroups []*rafay.ManagedNodeGroup, stateList basetypes.ListValue) ([]*rafay.ManagedNodeGroup, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(apiGroups) == 0 {
		return apiGroups, diags
	}

	apiByName := make(map[string]*rafay.ManagedNodeGroup, len(apiGroups))
	for _, mng := range apiGroups {
		if mng == nil {
			continue
		}
		apiByName[normalizeNodeGroupName(mng.Name)] = mng
	}

	if stateList.IsNull() || stateList.IsUnknown() || len(stateList.Elements()) == 0 {
		return apiGroups, diags
	}

	stateElems := make([]ManagedNodegroupsValue, 0, len(stateList.Elements()))
	if d := stateList.ElementsAs(ctx, &stateElems, false); d.HasError() {
		return apiGroups, append(diags, d...)
	}

	ordered := make([]*rafay.ManagedNodeGroup, 0, len(apiGroups))
	seen := make(map[string]bool, len(apiGroups))

	for _, st := range stateElems {
		name := managedNodeGroupNameFromValue(st.Name)
		if name == "" {
			continue
		}
		if mng, ok := apiByName[name]; ok {
			ordered = append(ordered, mng)
			seen[name] = true
		}
	}

	for _, mng := range apiGroups {
		if mng == nil {
			continue
		}
		name := normalizeNodeGroupName(mng.Name)
		if !seen[name] {
			ordered = append(ordered, mng)
		}
	}

	return ordered, diags
}

func buildNodeGroupsMapFromAPI(ctx context.Context, groups []*rafay.NodeGroup, stateMap basetypes.MapValue) (basetypes.MapValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(groups) == 0 {
		return types.MapNull(NodeGroupsMapValue{}.Type(ctx)), diags
	}

	stNgMaps := make(map[string]NodeGroupsMapValue, len(stateMap.Elements()))
	if !stateMap.IsNull() && !stateMap.IsUnknown() {
		if d := stateMap.ElementsAs(ctx, &stNgMaps, false); d.HasError() {
			diags = append(diags, d...)
		}
	}

	nodegrp := make(map[string]attr.Value, len(groups))
	for _, ng := range groups {
		if ng == nil || ng.Name == "" {
			continue
		}

		stNgMap := NodeGroupsMapValue{}
		if existing, ok := stNgMaps[ng.Name]; ok {
			stNgMap = existing
		} else if existing, ok := stNgMaps[normalizeNodeGroupName(ng.Name)]; ok {
			stNgMap = existing
		}

		ngrp := NewNodeGroupsMapValueNull()
		if d := ngrp.Flatten(ctx, ng, stNgMap); d.HasError() {
			diags = append(diags, d...)
		}
		nodegrp[ng.Name] = ngrp
	}

	mapVal, d := types.MapValue(NodeGroupsMapValue{}.Type(ctx), nodegrp)
	diags = append(diags, d...)
	return mapVal, diags
}

func buildManagedNodeGroupsMapFromAPI(ctx context.Context, groups []*rafay.ManagedNodeGroup, stateMap basetypes.MapValue) (basetypes.MapValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(groups) == 0 {
		return types.MapNull(ManagedNodegroupsMapValue{}.Type(ctx)), diags
	}

	stMngMaps := make(map[string]ManagedNodegroupsMapValue, len(stateMap.Elements()))
	if !stateMap.IsNull() && !stateMap.IsUnknown() {
		if d := stateMap.ElementsAs(ctx, &stMngMaps, false); d.HasError() {
			diags = append(diags, d...)
		}
	}

	managednodegrp := make(map[string]attr.Value, len(groups))
	for _, mng := range groups {
		if mng == nil || mng.Name == "" {
			continue
		}

		stMngMap := ManagedNodegroupsMapValue{}
		if existing, ok := stMngMaps[mng.Name]; ok {
			stMngMap = existing
		} else if existing, ok := stMngMaps[normalizeNodeGroupName(mng.Name)]; ok {
			stMngMap = existing
		}

		mngm := NewManagedNodegroupsMapValueNull()
		if d := mngm.Flatten(ctx, mng, stMngMap); d.HasError() {
			diags = append(diags, d...)
		}
		managednodegrp[mng.Name] = mngm
	}

	mapVal, d := types.MapValue(ManagedNodegroupsMapValue{}.Type(ctx), managednodegrp)
	diags = append(diags, d...)
	return mapVal, diags
}

func appendStateOnlyNodeGroupsListElements(ngElements []attr.Value, apiNames map[string]bool, stNgs []NodeGroupsValue) []attr.Value {
	for _, sng := range stNgs {
		name := nodeGroupNameFromValue(sng.Name)
		if name == "" || apiNames[name] {
			continue
		}
		ngElements = append(ngElements, sng)
	}
	return ngElements
}

func appendStateOnlyManagedNodeGroupsListElements(mngElements []attr.Value, apiNames map[string]bool, stMngs []ManagedNodegroupsValue) []attr.Value {
	for _, smng := range stMngs {
		name := managedNodeGroupNameFromValue(smng.Name)
		if name == "" || apiNames[name] {
			continue
		}
		mngElements = append(mngElements, smng)
	}
	return mngElements
}

func buildNodeGroupsMapFromListElements(ctx context.Context, list []attr.Value, stateMap basetypes.MapValue) (basetypes.MapValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(list) == 0 {
		return types.MapNull(NodeGroupsMapValue{}.Type(ctx)), diags
	}

	stNgMaps := make(map[string]NodeGroupsMapValue, len(stateMap.Elements()))
	if !stateMap.IsNull() && !stateMap.IsUnknown() {
		if d := stateMap.ElementsAs(ctx, &stNgMaps, false); d.HasError() {
			diags = append(diags, d...)
		}
	}

	nodegrp := make(map[string]attr.Value, len(list))
	for _, el := range list {
		ng, ok := el.(NodeGroupsValue)
		if !ok {
			continue
		}
		name := getStringValue(ng.Name)
		if name == "" {
			continue
		}

		stNgMap := NodeGroupsMapValue{}
		if existing, ok := stNgMaps[name]; ok {
			stNgMap = existing
		}

		apiNg, d := ng.Expand(ctx)
		diags = append(diags, d...)
		ngrp := NewNodeGroupsMapValueNull()
		diags = append(diags, ngrp.Flatten(ctx, apiNg, stNgMap)...)
		nodegrp[name] = ngrp
	}

	mapVal, d := types.MapValue(NodeGroupsMapValue{}.Type(ctx), nodegrp)
	diags = append(diags, d...)
	return mapVal, diags
}

func buildManagedNodeGroupsMapFromListElements(ctx context.Context, list []attr.Value, stateMap basetypes.MapValue) (basetypes.MapValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if len(list) == 0 {
		return types.MapNull(ManagedNodegroupsMapValue{}.Type(ctx)), diags
	}

	stMngMaps := make(map[string]ManagedNodegroupsMapValue, len(stateMap.Elements()))
	if !stateMap.IsNull() && !stateMap.IsUnknown() {
		if d := stateMap.ElementsAs(ctx, &stMngMaps, false); d.HasError() {
			diags = append(diags, d...)
		}
	}

	managednodegrp := make(map[string]attr.Value, len(list))
	for _, el := range list {
		mng, ok := el.(ManagedNodegroupsValue)
		if !ok {
			continue
		}
		name := getStringValue(mng.Name)
		if name == "" {
			continue
		}

		stMngMap := ManagedNodegroupsMapValue{}
		if existing, ok := stMngMaps[name]; ok {
			stMngMap = existing
		}

		apiMng, d := mng.Expand(ctx)
		diags = append(diags, d...)
		mngm := NewManagedNodegroupsMapValueNull()
		diags = append(diags, mngm.Flatten(ctx, apiMng, stMngMap)...)
		managednodegrp[name] = mngm
	}

	mapVal, d := types.MapValue(ManagedNodegroupsMapValue{}.Type(ctx), managednodegrp)
	diags = append(diags, d...)
	return mapVal, diags
}
