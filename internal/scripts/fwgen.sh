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

sed -i 's/"metadata2"/"metadata"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"iam2"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"iam_node_group_with_addon_policies2"/"iam_node_group_with_addon_policies"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"tolerations2"/"tolerations"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go
sed -i 's/"iam3"/"iam"/g' internal/resource_eks_cluster/eks_cluster_resource_gen.go