#!/bin/bash

# Add Go bin directory to PATH
export PATH="$HOME/go/bin:$PATH"

# Install the Terraform Plugin Code Generator Framework
echo "Installing tfplugingen-framework..."
go install github.com/hashicorp/terraform-plugin-codegen-framework/cmd/tfplugingen-framework@v0.4.1

# Check if the tool was installed successfully
if ! command -v tfplugingen-framework &> /dev/null; then
    echo "Error: tfplugingen-framework not found. Please check your Go installation and PATH."
    echo "Trying to use full path..."
    TFPLUGIN_FRAMEWORK="$HOME/go/bin/tfplugingen-framework"
    if [ ! -f "$TFPLUGIN_FRAMEWORK" ]; then
        echo "Error: Tool not found at $TFPLUGIN_FRAMEWORK"
        exit 1
    fi
else
    TFPLUGIN_FRAMEWORK="tfplugingen-framework"
fi

for file in $(find internal/resource_* -type f -name "*.json"); do
    echo "Generating framework provider code for ${file}..."
    $TFPLUGIN_FRAMEWORK generate resources \
        --input=${file} \
        --output=internal/ \
    ${file}
done

# Use macOS-compatible sed syntax (empty extension after -i)

# cluster fields
sed -i '' 's/"tolerations2"/"tolerations"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"security_groups2"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go


# cluster config fields
sed -i '' 's/"metadata2"/"metadata"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"iam3"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy_arns2"/"attach_policy_arns"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"tags3"/"tags"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"metadata3"/"metadata"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"well_known_policies2"/"well_known_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"subnets3"/"subnets"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy_arns3"/"attach_policy_arns"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy_v2_2"/"attach_policy_v2"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"permissions_boundary2"/"permissions_boundary"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"tags4"/"tags"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy3"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"statement2"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"well_known_policies3"/"well_known_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"pod_identity_associations2"/"pod_identity_associations"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"well_known_policies4"/"well_known_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

# managed node group
sed -i '' 's/"iam4"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"iam_node_group_with_addon_policies4"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy4"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"statement4"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"ssh4"/"ssh"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"placement4"/"placement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"instance_selector4"/"instance_selector"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"bottle_rocket4"/"bottle_rocket"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"taints4"/"taints"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"update_config4"/"update_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"launch_template4"/"launch_template"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"security_groups4"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"node_repair_config4"/"node_repair_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go


# managed node group map
sed -i '' 's/"iam5"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"iam_node_group_with_addon_policies5"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy5"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"statement5"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"ssh5"/"ssh"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"placement5"/"placement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"instance_selector5"/"instance_selector"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"bottle_rocket5"/"bottle_rocket"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"taints5"/"taints"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"update_config5"/"update_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"launch_template5"/"launch_template"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"security_groups5"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"node_repair_config5"/"node_repair_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

# node groups map
sed -i '' 's/"iam6"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"iam_node_group_with_addon_policies6"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy6"/"attach_policy"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"statement6"/"statement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"ssh6"/"ssh"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"placement6"/"placement"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"instances_distribution6"/"instances_distribution"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"instance_selector6"/"instance_selector"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"bottle_rocket6"/"bottle_rocket"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"taints6"/"taints"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"update_config6"/"update_config"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"launch_template6"/"launch_template"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"security_groups6"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
