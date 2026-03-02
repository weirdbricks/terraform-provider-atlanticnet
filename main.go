package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/weirdbricks/terraform-provider-atlanticnet/internal/provider"
)

// version is injected at build time via: go build -ldflags="-X main.version=0.1.0"
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debug mode for use with delve or other debuggers")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/weirdbricks/atlanticnet",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
