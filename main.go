// Package main is the entry point for the terraform-provider-libvirt binary.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/dmacvicar/terraform-provider-libvirt/v2/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// version is set at build time using ldflags
var version = "dev"

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/dmacvicar/libvirt",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
