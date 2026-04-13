// get_installer_manifest demonstrates low-level retrieval of the raw (un-flattened)
// installer manifest for a specific package and version. Use the Packages or
// Versions services to obtain fully resolved EffectiveInstaller slices instead.
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

	manifest, err := client.Manifests.GetInstallerManifest(context.Background(), "Microsoft.PowerShell", "7.4.6.0")
	if err != nil {
		log.Fatalf("GetInstallerManifest: %v", err)
	}

	out, err := json.MarshalIndent(manifest, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
