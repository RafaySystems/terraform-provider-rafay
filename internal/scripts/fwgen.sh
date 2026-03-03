#!/bin/bash

# Install the Terraform Plugin Code Generator Framework
go install github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@v0.4.1

for file in $(find internal/resource_* -type f -name "*.json"); do
    echo "Generating framework provider code for ${file}..."
    tfplugingen-framework generate resources \
        --input=${file} \
        --output=internal/ \
    ${file}
done

# cluster fields
sed -i 's/"tolerations2"/"tolerations"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"security_groups2"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go


# cluster config fields
sed -i 's/"metadata2"/"metadata"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"iam3"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"attach_policy_arns2"/"attach_policy_arns"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"tags3"/"tags"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"metadata3"/"metadata"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"well_known_policies2"/"well_known_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"subnets3"/"subnets"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"attach_policy_arns3"/"attach_policy_arns"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"attach_policy_v2_2"/"attach_policy_v2"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"permissions_boundary2"/"permissions_boundary"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"tags4"/"tags"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"attach_policy3"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"statement2"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"well_known_policies3"/"well_known_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"pod_identity_associations2"/"pod_identity_associations"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"well_known_policies4"/"well_known_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

# managed node group
sed -i 's/"iam4"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"iam_node_group_with_addon_policies4"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"attach_policy4"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"statement4"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"ssh4"/"ssh"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"placement4"/"placement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"instance_selector4"/"instance_selector"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"bottle_rocket4"/"bottle_rocket"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"taints4"/"taints"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"update_config4"/"update_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"launch_template4"/"launch_template"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"security_groups4"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"node_repair_config4"/"node_repair_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"node_repair_config_overrides4"/"node_repair_config_overrides"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

# managed node group map
sed -i 's/"iam5"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"iam_node_group_with_addon_policies5"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"attach_policy5"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"statement5"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"ssh5"/"ssh"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"placement5"/"placement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"instance_selector5"/"instance_selector"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"bottle_rocket5"/"bottle_rocket"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"taints5"/"taints"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"update_config5"/"update_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"launch_template5"/"launch_template"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"security_groups5"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"node_repair_config5"/"node_repair_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"node_repair_config_overrides5"/"node_repair_config_overrides"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

# node groups map
sed -i 's/"iam6"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"iam_node_group_with_addon_policies6"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"attach_policy6"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"statement6"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"ssh6"/"ssh"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"placement6"/"placement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"instances_distribution6"/"instances_distribution"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"instance_selector6"/"instance_selector"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"bottle_rocket6"/"bottle_rocket"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"taints6"/"taints"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"update_config6"/"update_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"launch_template6"/"launch_template"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"security_groups6"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

# Remove duplicate NodeRepairConfigOverrides type/value block (generated for both node_repair_config4 and node_repair_config5)
# Use exact string match for portability (regex \{\} can behave differently across awk implementations)
awk '$0 == "var _ basetypes.ObjectTypable = NodeRepairConfigOverridesType{}" { n++; if(n==2) { skip=1; next } }
skip { if ($0 == "var _ basetypes.ObjectTypable = Placement4Type{}") { skip=0; print; } next }
!skip { print }' internal/resource_eks_cluster/eks_cluster_resource_gen.go > internal/resource_eks_cluster/eks_cluster_resource_gen.go.tmp && \
	mv internal/resource_eks_cluster/eks_cluster_resource_gen.go.tmp internal/resource_eks_cluster/eks_cluster_resource_gen.go

# Add Version to schema if not already present (run once after codegen; re-run is idempotent)
grep -q 'Version: 1' internal/resource_eks_cluster/eks_cluster_resource_gen.go || \
	sed -i 's/return schema.Schema{$/return schema.Schema{\n\t\tVersion: 1,/' internal/resource_eks_cluster/eks_cluster_resource_gen.go
