package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const testProviderConfig = `
provider "rafay" {
	provider_config_file = "~/.rafay/cli/config.json"
	ignore_insecure_tls_error = true
}
`

var testFwProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"rafay": providerserver.NewProtocol6WithError(New("test")()),
}

func TestRafayProviderSchema(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"rafay": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig,
				Check: resource.ComposeTestCheckFunc(
					// Check if the provider config was applied correctly
					resource.TestCheckResourceAttr("provider.rafay", "provider_config_file", "~/.rafay/cli/config.json"),
					resource.TestCheckResourceAttr("provider.rafay", "ignore_insecure_tls_error", "true"),
				),
			},
		},
	})
}

// This function can be used to check for initialization of the provider
func TestRafayProvider_Configure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"rafay": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig,
				Check: resource.ComposeTestCheckFunc(
					// Add checks to verify that the provider initialized successfully
					func(s *terraform.State) error {
						// Add checks for client initialization, etc.
						return nil
					},
				),
			},
		},
	})
}
