//go:generate go tool tfplugindocs
package main

import (
	"context"

	tfprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nautobot/terraform-provider-nautobot/internal/provider"
)

// This gets overridden at build time via -ldflags "-X main.version=..."
var version = "dev"

func main() {
	providerserver.Serve(
		context.Background(),
		func() tfprovider.Provider { return provider.New(version) },
		providerserver.ServeOpts{
			Address: "registry.terraform.io/nautobot/nautobot",
		},
	)
}
