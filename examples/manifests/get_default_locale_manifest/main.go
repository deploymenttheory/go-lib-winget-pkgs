// get_default_locale_manifest demonstrates low-level retrieval of the
// defaultLocale manifest for a specific package and version. This manifest
// contains the canonical English-language metadata such as PackageName,
// Publisher, ShortDescription, Tags, and License.
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

	manifest, err := client.Manifests.GetDefaultLocaleManifest(context.Background(), "Microsoft.PowerShell", "7.4.6.0")
	if err != nil {
		log.Fatalf("GetDefaultLocaleManifest: %v", err)
	}

	out, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
