package tests

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// -----------------------------
// TERRATEST EKS CLUSTER TESTS
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
func TestEKSClusterConfigs(t *testing.T) {
	testCases := []struct {
		name  string
		tfDir string
		vars  map[string]interface{}
	}{
		{
			"Minimal Blueprint Config",
			"fixtures/01-minimal-blueprint",
			map[string]interface{}{
				"instance_type": "m5.xlarge",
			},
		},
		/*{
			"Default Blueprint With IAM Config",
			"fixtures/02-default-blueprint-with-iam",
			map[string]interface{}{
				"volume_type": "gp2",
			},
		},
		{
			"Basic Cluster Config",
			"fixtures/03-basic-cluster",
			map[string]interface{}{},
		},
		{
			"Managed Node Group Cluster",
			"fixtures/04-managed-nodegroups",
			map[string]interface{}{},
		},
		{
			"Spot Instances Cluster",
			"fixtures/05-spot-instances",
			map[string]interface{}{},
		},
		{
			"Cluster With Private Networking",
			"fixtures/06-private-networking",
			map[string]interface{}{},
		},
		{
			"Cluster With IAM Custom",
			"fixtures/07-iam-custom",
			map[string]interface{}{},
		},
		{
			"Multi AZ Scaling Cluster",
			"fixtures/08-multi-az-scaling",
			map[string]interface{}{},
		},
		{
			"Cluster With Sharing Enabled",
			"fixtures/09-sharing-enabled",
			map[string]interface{}{},
		},
		{
			"Bottlerocket GPU",
			"fixtures/10-bottlerocket-gpu",
			map[string]interface{}{},
		},*/
	}

	for _, tc := range testCases {
		expectedName, err := randomString("tf-eks-test", 6)
		if err != nil {
			t.Fatalf("Failed to generate random string: %v", err)
		}
		t.Parallel()
		t.Run(tc.name, func(t *testing.T) {
			tfDir := filepath.Join(".", tc.tfDir)
			fmt.Printf("\n🔧 Running test for %s (%s)\n", tc.name, tfDir)

			// Variables to pass to our Terraform code using -var options
			vars := map[string]interface{}{
				"rafay_config_file": os.Getenv("RAFAY_CONFIG_FILE"),
				"name":              expectedName,
			}

			opts := &terraform.Options{
				TerraformDir: tfDir,
				// Variables to pass to our Terraform code using -var options
				Vars: vars,
				EnvVars: map[string]string{
					"TF_CLI_CONFIG_FILE": os.Getenv("TF_CLI_CONFIG_FILE"),
				},
				NoColor: true,
			}

			// Always clean up
			defer terraform.DestroyE(t, opts)

			// Run Terraform init & apply
			terraform.InitAndApply(t, opts)

			// Verify output
			clusterName := terraform.Output(t, opts, "cluster_name")
			assert.Equal(t, expectedName, clusterName, "Expected cluster name output")

			// Verify idempotence (no diff after apply)
			planArgs := terraform.FormatArgs(opts, "plan", "-detailed-exitcode")
			planOut, _ := terraform.RunTerraformCommandE(t, opts, planArgs...)
			assert.NotContains(t, planOut, "Plan: ", "Expected no changes after successful apply")

			// Simulate config change (modify variable)
			updatedOpts := *opts
			// Merge scenario-specific vars
			for k, v := range tc.vars {
				updatedOpts.Vars[k] = v
			}
			diffOut := terraform.RunTerraformCommand(t, &updatedOpts, terraform.FormatArgs(&updatedOpts, "plan")...)
			assert.Contains(t, diffOut, "Plan:", "Expected diff after modifying config")

			fmt.Printf("✅ Test passed for %s\n", tc.name)
		})
	}
}

func randomString(prefix string, maxLength int) (string, error) {
	allowedChars := "abcdefghijklmnopqrstuvwxyz0123456789"
	if maxLength <= 0 {
		return "", fmt.Errorf("maxLength must be positive")
	}

	result := make([]byte, maxLength)
	for i := 0; i < maxLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedChars))))
		if err != nil {
			return "", err
		}
		result[i] = allowedChars[n.Int64()]
	}
	return fmt.Sprintf("%s-%s", prefix, string(result)), nil
}
