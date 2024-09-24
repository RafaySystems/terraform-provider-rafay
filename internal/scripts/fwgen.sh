#!/bin/bash

for file in $(find internal/resource_* -type f -name "*.json"); do
    echo "Generating framework provider code for ${file}..."
    tfplugingen-framework generate resources \
        --input=${file} \
        --output=internal/ \
    ${file}
done