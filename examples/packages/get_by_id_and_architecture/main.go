// get_by_id_and_architecture demonstrates retrieving the latest version of a
// package by identifier, with the Installers slice pre-filtered to only those
// matching the requested architecture. An error is returned if no installer for
// that architecture exists in the package.
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

	pkg, err := client.Packages.GetByIDAndArchitecture(context.Background(), "Microsoft.PowerShell", "x64")
	if err != nil {
		log.Fatalf("GetByIDAndArchitecture: %v", err)
	}

	out, err := json.MarshalIndent(pkg, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
