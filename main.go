package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	framework "github.com/RafaySystems/terraform-provider-rafay/internal/provider"
	legacySDK "github.com/RafaySystems/terraform-provider-rafay/rafay"
)

var (
	// these will be set by the goreleaser configuration
	// to appropriate values for the compiled binary
	version string = "dev"

	// goreleaser can also pass the specific commit if you want
	// commit  string = ""
)

func main() {
	ctx := context.Background()

	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	// Upgrade the legacy provider to protocol v6
	upgradedSdkServer, err := tf5to6server.UpgradeServer(ctx, legacySDK.New(version)().GRPCProvider)
	if err != nil {
		log.Fatal(err)
	}

	// Create the provider server functions list
	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(framework.New(version)()), // terraform-plugin-framework provider
		func() tfprotov6.ProviderServer {
			return upgradedSdkServer
		},
	}

	// Create a new mux server
	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)

	if err != nil {
		log.Fatal(err)
	}

	// Serve the provider
	var serveOpts []tf6server.ServeOpt

	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"registry.terraform.io/RafaySystems/rafay",
		muxServer.ProviderServer,
		serveOpts...,
	)

	if err != nil {
		log.Fatal(err)
	}
}
