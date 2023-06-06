package test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

func TestResourceConfigContext(t *testing.T) {
	expectedName := "test-config-context"
	// Make a copy of the terraform module to a temporary directory. This allows running multiple tests in parallel
	// against the same terraform module.
	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "../", "examples/resources/rafay_config_context")

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: exampleFolder,

		// Variables to pass to our Terraform code using -var options
		Vars: map[string]interface{}{
			"rafay_config_file": os.Getenv("RAFAY_CONFIG_FILE"),
			"name":              expectedName,
		},

		// Environment variables to set when running Terraform
		EnvVars: map[string]string{
			"TF_CLI_CONFIG_FILE": os.Getenv("TF_CLI_CONFIG_FILE"),
		},
	})

	// run `terraform destroy` to clean up any resources that were created
	defer terraform.Destroy(t, terraformOptions)

	// run `terraform init` and `terraform apply` and fail the test if there are any errors
	terraform.InitAndApply(t, terraformOptions)

	resourceName := terraform.Output(t, terraformOptions, "resource_name")

	assert.Equal(t, expectedName, resourceName)
}
