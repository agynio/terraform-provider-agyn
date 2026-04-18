package main

import (
	"context"
	"flag"
	"log"

	"github.com/agynio/terraform-provider-agyn/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version = "0.6.9"
	commit  = ""
)

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run provider with debugger support")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/agynio/agyn",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version, commit), opts); err != nil {
		log.Fatalf("error serving provider: %v", err)
	}
}
