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