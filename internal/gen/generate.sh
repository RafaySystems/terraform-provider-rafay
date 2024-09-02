for file in $(find internal/gen/fwspec -type f -name "*.json"); do
    echo "Generating framework provider code for ${file}..."
    tfplugingen-framework generate resources \
        --input=${file} \
        --output=internal/gen \
    ${file}
done