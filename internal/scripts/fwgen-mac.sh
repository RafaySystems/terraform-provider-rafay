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
sed -i '' 's/"metadata2"/"metadata"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"iam2"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"iam_node_group_with_addon_policies2"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"tolerations2"/"tolerations"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"iam3"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

sed -i '' 's/"availability_zones2"/"availability_zones"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"labels2"/"labels"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"tags2"/"tags"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"security_groups2"/"security_groups"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

sed -i '' 's/"subnets2"/"subnets"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"subnets3"/"subnets"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"subnets4"/"subnets"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"subnets5"/"subnets"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

sed -i '' 's/"attach_policy_arns3"/"attach_policy_arns"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i '' 's/"attach_policy_v2_2"/"attach_policy_v2"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go

