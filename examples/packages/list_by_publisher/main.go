// list_by_publisher demonstrates listing the latest version of all packages
// whose publisher component (the part of the identifier before the first dot)
// matches the given value.
package main

import (
	"context"
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

	resp, err := client.Packages.ListByPublisher(context.Background(), "Microsoft")
	if err != nil {
		log.Fatalf("ListByPublisher: %v", err)
	}

	fmt.Printf("Microsoft packages: %d\n\n", resp.TotalCount)

	for _, pkg := range resp.Results {
		fmt.Printf("  %-50s  %s\n", pkg.PackageIdentifier, pkg.LatestVersion)
	}
}
