// get_by_id_and_version demonstrates retrieving a specific version of a package
// by its package identifier and version string.
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

	pkg, err := client.Packages.GetByIDAndVersion(context.Background(), "Microsoft.PowerShell", "7.4.6.0")
	if err != nil {
		log.Fatalf("GetByIDAndVersion: %v", err)
	}

	out, err := json.MarshalIndent(pkg, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
