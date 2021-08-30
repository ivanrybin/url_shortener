package main

import (
	"log"

	"url_shortener/cmd/url_shortener_client/cmd"

	"github.com/spf13/pflag"
)

func main() {
	// server address
	var address string
	pflag.StringVarP(&address, "address", "a", "localhost:9876", "server address")
	pflag.Parse()

	if err := cmd.NewCLI(address).Execute(); err != nil {
		log.Fatalf("cannot execute command: %v", err)
	}
}
