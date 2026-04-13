// get_by_name_and_architecture demonstrates finding packages by display name and
// filtering the returned installer list to a specific architecture. Packages that
// have no installer for the requested architecture are excluded from the results.
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

	results, err := client.Packages.GetByNameAndArchitecture(context.Background(), "PowerShell", "x64")
	if err != nil {
		log.Fatalf("GetByNameAndArchitecture: %v", err)
	}

	fmt.Printf("Found %d package(s) named \"PowerShell\" with x64 installers\n\n", len(results))

	for _, pkg := range results {
		out, err := json.MarshalIndent(pkg, "", "    ")
		if err != nil {
			log.Fatalf("marshalling result: %v", err)
		}

		fmt.Println(string(out))
	}
}
