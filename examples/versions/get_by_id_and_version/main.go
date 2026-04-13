// get_by_id_and_version demonstrates retrieving the fully-resolved metadata for
// a specific version of a package, including merged installer and locale data.
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

	ver, err := client.Versions.GetByIDAndVersion(context.Background(), "Microsoft.PowerShell", "7.4.6.0")
	if err != nil {
		log.Fatalf("GetByIDAndVersion: %v", err)
	}

	out, err := json.MarshalIndent(ver, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
