# Go Library for WinGet Package Metadata

[![Go Report Card](https://goreportcard.com/badge/github.com/deploymenttheory/go-lib-winget-pkgs)](https://goreportcard.com/report/github.com/deploymenttheory/go-lib-winget-pkgs)
[![GoDoc](https://pkg.go.dev/badge/github.com/deploymenttheory/go-lib-winget-pkgs)](https://pkg.go.dev/github.com/deploymenttheory/go-lib-winget-pkgs)
[![License](https://img.shields.io/github/license/deploymenttheory/go-lib-winget-pkgs)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/deploymenttheory/go-lib-winget-pkgs)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/deploymenttheory/go-lib-winget-pkgs)](https://github.com/deploymenttheory/go-lib-winget-pkgs/releases)
![Status: Experimental](https://img.shields.io/badge/status-experimental-yellow)

A Go library for programmatic access to [WinGet](https://github.com/microsoft/winget-pkgs) package metadata. Manages a local clone of the `microsoft/winget-pkgs` repository and exposes a clean, service-oriented API for querying packages, versions, and manifests without any external API calls. Uses an in-memory index built from the manifest tree for fast lookups across 75 000+ packages.

## Quick Start

```bash
go get github.com/deploymenttheory/go-lib-winget-pkgs
```

```go
import (
    "context"
    "fmt"

    "github.com/deploymenttheory/go-lib-winget-pkgs/winget"
    "github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
)

func main() {
    client, err := winget.NewClient(&config.Config{},
        winget.WithCacheDir("/tmp/winget-cache"),
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()

    pkg, err := client.Packages.GetByID(context.Background(), "Microsoft.PowerShell")
    if err != nil {
        panic(err)
    }

    fmt.Printf("%s %s — %s\n", pkg.PackageName, pkg.LatestVersion, pkg.ShortDescription)
    fmt.Printf("Installers: %d\n", len(pkg.Installers))
}
```

On first run the library clones `microsoft/winget-pkgs` into the cache directory and builds an in-memory index. Subsequent runs use the cached clone and pull only the latest changes.

## Examples

The [examples directory](examples/) contains a working `main.go` for every service operation:

**Packages** — [`examples/packages/`](examples/packages/)

| Example | Operation |
|---|---|
| [`get_by_id/`](examples/packages/get_by_id/) | Fetch the latest version of a package by identifier |
| [`get_by_id_and_version/`](examples/packages/get_by_id_and_version/) | Fetch a specific version by identifier |
| [`get_by_name/`](examples/packages/get_by_name/) | Find packages by display name |
| [`get_by_id_and_architecture/`](examples/packages/get_by_id_and_architecture/) | Fetch latest version filtered to one architecture |
| [`get_by_name_and_architecture/`](examples/packages/get_by_name_and_architecture/) | Find packages by name filtered to one architecture |
| [`list_all/`](examples/packages/list_all/) | List every package in the index |
| [`list_by_publisher/`](examples/packages/list_by_publisher/) | List packages by publisher |
| [`search/`](examples/packages/search/) | Filter packages by publisher, name, tags, installer type, and more |
| [`search_installers/`](examples/packages/search_installers/) | Flat installer-level search (one result per installer entry) |

**Versions** — [`examples/versions/`](examples/versions/)

| Example | Operation |
|---|---|
| [`list_by_id/`](examples/versions/list_by_id/) | List all available versions for a package |
| [`get_by_id_and_version/`](examples/versions/get_by_id_and_version/) | Fetch a fully-resolved version with all manifests merged |

**Manifests** — [`examples/manifests/`](examples/manifests/)

| Example | Operation |
|---|---|
| [`get_version_manifest/`](examples/manifests/get_version_manifest/) | Read the raw version manifest |
| [`get_installer_manifest/`](examples/manifests/get_installer_manifest/) | Read the raw installer manifest |
| [`get_default_locale_manifest/`](examples/manifests/get_default_locale_manifest/) | Read the default locale manifest |
| [`get_locale_manifest/`](examples/manifests/get_locale_manifest/) | Read a specific locale manifest |
| [`list_locales/`](examples/manifests/list_locales/) | List all available locales for a package version |

## Configuration

### Creating a client

```go
import (
    "github.com/deploymenttheory/go-lib-winget-pkgs/winget"
    "github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
)

// Defaults: clones to ~/.cache/go-lib-winget-pkgs, shallow clone, no auto-refresh
client, err := winget.NewClient(&config.Config{})

// Custom cache directory
client, err := winget.NewClient(&config.Config{},
    winget.WithCacheDir("/var/cache/winget"),
)
```

### Config fields

```go
&config.Config{
    RepoURL:     "https://github.com/microsoft/winget-pkgs", // default
    CacheDir:    "~/.cache/go-lib-winget-pkgs",              // default
    CloneDepth:  1,                                           // 0 = full clone (default: 1)
    AutoRefresh: 0,                                           // 0 = never auto-refresh (default)
    WorkerCount: runtime.NumCPU(),                            // concurrent YAML parsers (default)
    Timeout:     5 * time.Minute,                             // clone/pull timeout (default)
}
```

### Client options

```go
winget.WithCacheDir("/var/cache/winget")          // Override local clone directory
winget.WithRepoURL("https://...")                  // Use a mirror or fork
winget.WithCloneDepth(0)                           // Full clone instead of shallow
winget.WithAutoRefresh(6 * time.Hour)              // Pull and re-index on a schedule
winget.WithWorkerCount(8)                          // Concurrent YAML parsers at index build
winget.WithTimeout(10 * time.Minute)               // Clone/pull timeout
winget.WithLogger(zapLogger)                       // Structured logging with go.uber.org/zap
```

### Example: production configuration

```go
import (
    "time"
    "go.uber.org/zap"
    "github.com/deploymenttheory/go-lib-winget-pkgs/winget"
    "github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
)

logger, _ := zap.NewProduction()

client, err := winget.NewClient(
    &config.Config{},
    winget.WithCacheDir("/var/cache/winget"),
    winget.WithCloneDepth(1),
    winget.WithAutoRefresh(6*time.Hour),
    winget.WithWorkerCount(8),
    winget.WithLogger(logger),
)
if err != nil {
    panic(err)
}
defer client.Close()
```

## Service API

### Packages

```go
// Latest version by identifier (case-insensitive)
pkg, err := client.Packages.GetByID(ctx, "Microsoft.PowerShell")

// Specific version
pkg, err := client.Packages.GetByIDAndVersion(ctx, "Microsoft.PowerShell", "7.4.6.0")

// By display name (may return multiple on collision)
pkgs, err := client.Packages.GetByName(ctx, "PowerShell")

// Latest version filtered to one architecture — errors if no match
pkg, err := client.Packages.GetByIDAndArchitecture(ctx, "Microsoft.PowerShell", "x64")

// By name filtered to one architecture — excludes packages with no match
pkgs, err := client.Packages.GetByNameAndArchitecture(ctx, "PowerShell", "x64")

// All packages (latest version of each)
resp, err := client.Packages.ListAll(ctx)

// By publisher path component
resp, err := client.Packages.ListByPublisher(ctx, "Microsoft")

// Filtered search
resp, err := client.Packages.Search(ctx, &packages.FilterOptions{
    Publisher:     "Microsoft",
    InstallerType: "msi",
    Architecture:  "x64",
    Limit:         50,
})

// Flat installer-level search (one result per installer entry)
resp, err := client.Packages.SearchInstallers(ctx, &packages.InstallerFilterOptions{
    Architecture:  "x64",
    InstallerType: "msi",
    Publisher:     "Microsoft",
    Limit:         20,
})
```

### Versions

```go
// All available versions for a package, sorted descending
resp, err := client.Versions.ListByID(ctx, "Microsoft.PowerShell")

// Fully-resolved version with all manifests merged
ver, err := client.Versions.GetByIDAndVersion(ctx, "Microsoft.PowerShell", "7.4.6.0")
```

### Manifests (low-level)

```go
// Raw version manifest
vm, err := client.Manifests.GetVersionManifest(ctx, "Microsoft.PowerShell", "7.5.0.0")

// Raw installer manifest (un-flattened)
im, err := client.Manifests.GetInstallerManifest(ctx, "Microsoft.PowerShell", "7.5.0.0")

// Default locale manifest
dlm, err := client.Manifests.GetDefaultLocaleManifest(ctx, "Microsoft.PowerShell", "7.5.0.0")

// Specific locale manifest
lm, err := client.Manifests.GetLocaleManifest(ctx, "Microsoft.PowerShell", "7.5.0.0", "de-DE")

// All available locale codes
locales, err := client.Manifests.ListLocales(ctx, "Microsoft.PowerShell", "7.5.0.0")
```

### Refreshing the index

```go
// Pull latest from upstream and rebuild the in-memory index
if err := client.Refresh(ctx); err != nil {
    log.Printf("refresh failed: %v", err)
}
```

## FilterOptions reference

```go
// packages.FilterOptions — used with Search
type FilterOptions struct {
    Publisher         string   // exact match (case-insensitive)
    NameContains      string   // substring match on PackageName
    MonikerContains   string   // substring match on Moniker
    TagsAny           []string // package must have at least one tag
    License           string   // substring match on License
    InstallerType     string   // "msi", "exe", "msix", "wix", etc.
    Architecture      string   // "x64", "x86", "arm64", "neutral"
    Scope             string   // "user" or "machine"
    HasMoniker        bool     // only packages with a Moniker set
    ProductCode       string   // exact match on installer ProductCode
    PackageFamilyName string   // exact match on PackageFamilyName
    CommandsAny       []string // installer must register at least one command
    MinimumOSVersion  string   // exact match on MinimumOSVersion
    Limit             int      // 0 = unlimited
    Offset            int
}

// packages.InstallerFilterOptions — used with SearchInstallers
type InstallerFilterOptions struct {
    Publisher         string
    Architecture      string
    InstallerType     string
    Scope             string
    ProductCode       string
    PackageFamilyName string
    CommandsAny       []string
    MinimumOSVersion  string
    Limit             int
    Offset            int
}
```

## Error handling

```go
import "github.com/deploymenttheory/go-lib-winget-pkgs/winget"

pkg, err := client.Packages.GetByID(ctx, "Unknown.Package")
if winget.IsNotFound(err) {
    // package does not exist in the index
}
if winget.IsInvalidManifest(err) {
    // YAML could not be parsed
}
if winget.IsCloneFailed(err) {
    // initial clone failed (network, disk, permissions)
}
```

## Architecture

```
winget/
├── winget.go                   # Client struct, NewClient, Refresh, Close
├── with_options.go             # ClientOption functional options
├── config/config.go            # Config struct and defaults
├── executor/                   # Abstracts git + filesystem
│   ├── interface.go            # Executor interface
│   ├── gogit.go                # go-git implementation
│   └── mock/executor.go        # In-memory mock for unit tests
├── index/                      # In-memory package index
│   ├── index.go                # Build, thread-safe lookups
│   └── entry.go                # IndexEntry, VersionEntry
├── shared/models/models.go     # Shared types (EffectiveInstaller, Dependencies, …)
└── services/
    ├── packages/               # Packages service
    ├── versions/               # Versions service
    └── manifests/              # Low-level manifest access + parser
```

The executor interface decouples all services from the filesystem, making every service fully unit-testable without a network connection or a cloned repository.

## Documentation

- [WinGet manifest schema](https://github.com/microsoft/winget-pkgs/tree/master/doc/manifest)
- [GoDoc](https://pkg.go.dev/github.com/deploymenttheory/go-lib-winget-pkgs)

## Contributing

Contributions are welcome. Please read our [Contributing Guidelines](CONTRIBUTING.md) before submitting pull requests.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.

## Support

- **Issues:** [GitHub Issues](https://github.com/deploymenttheory/go-lib-winget-pkgs/issues)
- **WinGet manifest schema:** [microsoft/winget-pkgs](https://github.com/microsoft/winget-pkgs/tree/master/doc/manifest)

## Disclaimer

This is a community library and is not affiliated with or endorsed by Microsoft.
