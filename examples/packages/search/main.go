// search demonstrates filtered package search across the WinGet catalogue using
// structured FilterOptions. Multiple criteria are combined with AND semantics;
// within TagsAny a single matching tag is sufficient.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/packages"
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

	resp, err := client.Packages.Search(context.Background(), &packages.FilterOptions{
		Publisher:     "Microsoft",
		InstallerType: "msi",
		Architecture:  "x64",
		Limit:         20,
	})
	if err != nil {
		log.Fatalf("Search: %v", err)
	}

	fmt.Printf("Found %d packages (showing %d)\n\n", resp.TotalCount, len(resp.Results))

	for _, pkg := range resp.Results {
		fmt.Printf("  %-50s  %-20s  %s\n", pkg.PackageIdentifier, pkg.LatestVersion, pkg.ShortDescription)
	}
}
