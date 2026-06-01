package resource_eks_cluster

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// EksClusterResourceSchemaPatched returns the generated schema with provider patches
// applied. Map mirror fields are Computed+Optional so they can be maintained in
// state without appearing in user config or plan diffs.
func EksClusterResourceSchemaPatched(ctx context.Context) schema.Schema {
	s := EksClusterResourceSchema(ctx)

	patchStringUseStateForUnknown(s.Attributes, "id")

	if ccBlock, ok := s.Blocks["cluster_config"].(schema.ListNestedBlock); ok {
		patchClusterConfigMapAttributes(ccBlock.NestedObject.Attributes)
	}

	return s
}

func patchStringUseStateForUnknown(attrs map[string]schema.Attribute, name string) {
	attr, ok := attrs[name]
	if !ok {
		return
	}

	str, ok := attr.(schema.StringAttribute)
	if !ok {
		return
	}

	str.PlanModifiers = append(str.PlanModifiers, stringplanmodifier.UseStateForUnknown())
	attrs[name] = str
}

func patchClusterConfigMapAttributes(attrs map[string]schema.Attribute) {
	if attrs == nil {
		return
	}

	patchMapNestedComputedOptional(attrs, "node_groups_map")
	patchMapNestedComputedOptional(attrs, "managed_nodegroups_map")
}

func patchMapNestedComputedOptional(attrs map[string]schema.Attribute, name string) {
	attr, ok := attrs[name]
	if !ok {
		return
	}

	m, ok := attr.(schema.MapNestedAttribute)
	if !ok {
		return
	}

	m.Computed = true
	m.Optional = true
	attrs[name] = m
}
