package test

import (
	"os"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"github.com/stretchr/testify/assert"
)

func TestResourceGkeCluster(t *testing.T) {
	expectedName := randomString("terratest-gke-cluster", 5)

	// Make a copy to allow running multiple tests in parallel.
	exampleFolder := test_structure.CopyTerraformFolderToTemp(t, "../", "examples/resources/rafay_gke_cluster")
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: exampleFolder,
		Vars: map[string]interface{}{
			"rafay_config_file": os.Getenv("RAFAY_CONFIG_FILE"),
			"name":              expectedName,
		},
		EnvVars: map[string]string{
			"TF_CLI_CONFIG_FILE": os.Getenv("TF_CLI_CONFIG_FILE"),
		},
	})

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	resourceName := terraform.Output(t, terraformOptions, "resource_name")
	assert.Equal(t, expectedName, resourceName)
}
