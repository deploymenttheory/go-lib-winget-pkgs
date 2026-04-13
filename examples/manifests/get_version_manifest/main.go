// get_version_manifest demonstrates low-level retrieval of the version manifest
// for a specific package and version. The version manifest contains the
// PackageIdentifier, PackageVersion, DefaultLocale, and ManifestVersion fields.
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

	manifest, err := client.Manifests.GetVersionManifest(context.Background(), "Microsoft.PowerShell", "7.4.6.0")
	if err != nil {
		log.Fatalf("GetVersionManifest: %v", err)
	}

	out, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
