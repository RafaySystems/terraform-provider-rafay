//go:build planonly
// +build planonly

package framework

import (
	"github.com/RafaySystems/terraform-provider-rafay/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	// testProviderConfig is a shared configuration to combine with the actual
	// test configuration so the Rafay client is properly configured.
	// It is also possible to use the RCTL_ environment variables instead,
	// such as updating the Makefile and running the testing through that tool.

	testProviderConfig = `
provider "rafay" {
	provider_config_file = "~/.rafay/cli/config.json"
	ignore_insecure_tls_error = true
}
`
)

var (
	// testFwProviderFactories are used to instantiate a provider during
	// acceptance testing. The factory function will be invoked for every Terraform
	// CLI command executed to create a provider server to which the CLI can
	// reattach.

	testFwProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"rafay": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)
