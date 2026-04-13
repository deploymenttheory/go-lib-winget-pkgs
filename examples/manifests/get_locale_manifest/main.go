// get_locale_manifest demonstrates low-level retrieval of a specific locale
// manifest for a package and version using a BCP-47 locale code (e.g. "de-DE").
// Use list_locales to discover which locale codes are available.
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

	manifest, err := client.Manifests.GetLocaleManifest(context.Background(), "Microsoft.PowerShell", "7.4.6.0", "de-DE")
	if err != nil {
		log.Fatalf("GetLocaleManifest: %v", err)
	}

	out, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
