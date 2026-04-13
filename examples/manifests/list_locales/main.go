// list_locales demonstrates retrieving all available locale codes for a specific
// package and version. The returned BCP-47 codes can be passed to
// GetLocaleManifest to retrieve the translated metadata.
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

	locales, err := client.Manifests.ListLocales(context.Background(), "Microsoft.PowerShell", "7.4.6.0")
	if err != nil {
		log.Fatalf("ListLocales: %v", err)
	}

	fmt.Printf("Available locales (%d):\n", len(locales))

	for _, locale := range locales {
		fmt.Printf("  %s\n", locale)
	}
}
