// get_by_name demonstrates finding packages whose display name (PackageName from
// the defaultLocale manifest) matches the given string. Multiple packages from
// different publishers may share the same name.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
)

func main() {
	client, err := winget.NewClient(&config.Config{},
		winget.WithCacheDir("/tmp/winget-cache"),
	)
	if err != nil {
		log.Fatalf("creating client: %v", err)
	}

	defer func() {
		if closeErr := client.Close(); closeErr != nil {
			log.Printf("closing client: %v", closeErr)
		}
	}()

	results, err := client.Packages.GetByName(context.Background(), "PowerShell")
	if err != nil {
		log.Fatalf("GetByName: %v", err)
	}

	fmt.Printf("Found %d package(s) named \"PowerShell\"\n\n", len(results))

	for _, pkg := range results {
		out, err := json.MarshalIndent(pkg, "", "    ")
		if err != nil {
			log.Fatalf("marshalling result: %v", err)
		}

		fmt.Println(string(out))
	}
}
