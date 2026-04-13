// search_installers demonstrates the flat installer search, which returns one
// result per matching installer entry rather than one per package. This is
// useful for queries that operate at the installer level — for example, finding
// every x64 MSI that registers a specific CLI command, or locating packages by
// ProductCode or PackageFamilyName.
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

	resp, err := client.Packages.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		Architecture:  "x64",
		InstallerType: "msi",
		Publisher:     "Microsoft",
		Limit:         20,
	})
	if err != nil {
		log.Fatalf("SearchInstallers: %v", err)
	}

	fmt.Printf("Found %d x64 MSI installer(s) from Microsoft (showing %d)\n\n",
		resp.TotalCount, len(resp.Results))

	for _, r := range resp.Results {
		fmt.Printf("  %-50s  %-15s  arch=%-6s  type=%s\n",
			r.PackageIdentifier,
			r.PackageVersion,
			r.Installer.Architecture,
			r.Installer.InstallerType,
		)
	}
}
