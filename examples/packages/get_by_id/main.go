// get_by_id demonstrates retrieving the latest version of a package by its
// package identifier.
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

	pkg, err := client.Packages.GetByID(context.Background(), "Microsoft.PowerShell")
	if err != nil {
		log.Fatalf("GetByID: %v", err)
	}

	out, err := json.MarshalIndent(pkg, "", "    ")
	if err != nil {
		log.Fatalf("marshalling result: %v", err)
	}

	fmt.Println(string(out))
}
