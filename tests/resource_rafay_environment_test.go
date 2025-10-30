package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

// -----------------------------
// TERRATEST Rafay Environment TESTS
// -----------------------------
// TO RUN THESE TESTS, MAKE SURE YOU HAVE THE FOLLOWING ENV VARIABLES SET:
// RAFAY_CONFIG_FILE - Path to your Rafay config file
// TF_CLI_CONFIG_FILE - Path to your Terraform CLI config file ( dev.tfrc )
/*
sample dev.tfrc config file:
provider_installation {

  # Use /home/developer/tmp/terraform-null as an overridden package directory
  # for the hashicorp/null provider. This disables the version and checksum
  # verifications for this provider and forces Terraform to look for the
  # null provider plugin in the given directory.
  dev_overrides {
    "rafaysystems/rafay" = "/Users/niravparikh/development/rafay/terraform-provider-rafay"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
*/
// -----------------------------
func TestRafayEnvironmentScale(t *testing.T) {
	testCases := []struct {
		name            string
		tfDir           string
		numEnvironments int
		etName          string
		etVersion       string
		agent           string
		project         string
	}{
		{
			name:            "Scale Test 100 Environments",
			tfDir:           "fixtures/rafay-environment-scale",
			numEnvironments: 100,
			etName:          "test-np-et",
			etVersion:       "withschd",
			agent:           "nirav-qc-k8s-eks-one",
			project:         "nirav",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tfDir := filepath.Join(".", tc.tfDir)
			fmt.Printf("\n🔧 Running test for %s (%s)\n", tc.name, tfDir)

			vars := map[string]interface{}{
				"rafay_config_file": os.Getenv("RAFAY_CONFIG_FILE"),
				"num_environments":  tc.numEnvironments,
				"et_name":           tc.etName,
				"et_version":        tc.etVersion,
				"agent":             tc.agent,
				"project":           tc.project,
				"name_prefix":       "scale-test-env",
			}

			opts := &terraform.Options{
				TerraformDir: tfDir,
				Vars:         vars,
				EnvVars: map[string]string{
					"TF_CLI_CONFIG_FILE": os.Getenv("TF_CLI_CONFIG_FILE"),
				},
				NoColor: true,
			}

			defer terraform.DestroyE(t, opts)
			terraform.InitAndApply(t, opts)
			fmt.Printf("✅ Test passed for %s\n", tc.name)
		})
	}
}
