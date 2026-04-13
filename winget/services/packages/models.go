// Package packages provides the primary query surface for WinGet packages.
package packages

import (
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/shared/models"
)

// ResourcePackage is the merged view of the latest (or requested) version of a
// WinGet package. It combines data from the version manifest, installer manifest,
// defaultLocale manifest, and any additional locale manifests.
type ResourcePackage struct {
	PackageIdentifier   string
	LatestVersion       string
	AvailableVersions   []string
	DefaultLocale       string
	Publisher           string
	PublisherURL        string
	PublisherSupportURL string
	PrivacyURL          string
	Author              string
	PackageName         string
	PackageURL          string
	License             string
	LicenseURL          string
	Copyright           string
	CopyrightURL        string
	ShortDescription    string
	Description         string
	Moniker             string
	Tags                []string
	Agreements          []models.Agreement
	ReleaseNotes        string
	ReleaseNotesURL     string
	PurchaseURL         string
	InstallationNotes   string
	Documentations      []models.Documentation
	Icons               []models.Icon
	Installers          []models.EffectiveInstaller
	Locales             []manifests.LocaleManifest
}

// ListResponse wraps a slice of packages with a total count.
type ListResponse struct {
	TotalCount int
	Results    []*ResourcePackage
}

// FilterOptions specifies criteria for filtering packages in Search and List
// operations. All non-empty string fields are case-insensitive. All slice
// fields use OR semantics (package must match at least one element).
type FilterOptions struct {
	// Publisher filters by exact match on the publisher component of the package
	// identifier (the part before the first dot), e.g. "Microsoft".
	Publisher string

	// NameContains filters by a substring match on PackageName from the
	// defaultLocale manifest.
	NameContains string

	// MonikerContains filters by a substring match on the Moniker field.
	MonikerContains string

	// TagsAny filters to packages that have at least one of the specified tags.
	TagsAny []string

	// License filters by a substring match on the License field.
	License string

	// InstallerType filters to packages that have at least one effective installer
	// with the given InstallerType (e.g. "msi", "exe", "msix").
	InstallerType string

	// Architecture filters to packages that have at least one effective installer
	// with the given architecture (e.g. "x64", "x86", "arm64").
	Architecture string

	// Scope filters to packages that have at least one effective installer with
	// the given scope (e.g. "user", "machine").
	Scope string

	// ProductCode filters to packages that have at least one effective installer
	// with an exact match on ProductCode (case-insensitive). Useful for looking
	// up packages by their Windows Installer product GUID.
	ProductCode string

	// PackageFamilyName filters to packages that have at least one effective
	// installer with an exact match on PackageFamilyName (case-insensitive).
	// Applies primarily to MSIX/APPX packages.
	PackageFamilyName string

	// CommandsAny filters to packages that register at least one of the given
	// CLI commands in any effective installer.
	CommandsAny []string

	// MinimumOSVersion filters to packages that have at least one effective
	// installer whose MinimumOSVersion contains the given string (e.g. "10.0.17763").
	MinimumOSVersion string

	// HasMoniker, when true, restricts results to packages with a non-empty Moniker.
	HasMoniker bool

	// Limit caps the number of results returned. 0 means no limit.
	Limit int

	// Offset skips the first N results (for pagination).
	Offset int
}

// InstallerFilterOptions specifies criteria for the SearchInstallers method,
// which operates at the individual installer level rather than the package level.
// Each matching result is a single resolved installer entry with its package context.
type InstallerFilterOptions struct {
	// Publisher narrows the candidate set to packages from the given publisher
	// (case-insensitive exact match on the publisher path component).
	Publisher string

	// Architecture filters to installers with the given architecture
	// (e.g. "x64", "x86", "arm64").
	Architecture string

	// InstallerType filters to installers with the given type
	// (e.g. "msi", "exe", "msix", "wix").
	InstallerType string

	// Scope filters to installers with the given scope ("user" or "machine").
	Scope string

	// ProductCode filters to installers with an exact match on ProductCode
	// (case-insensitive).
	ProductCode string

	// PackageFamilyName filters to installers with an exact match on
	// PackageFamilyName (case-insensitive).
	PackageFamilyName string

	// CommandsAny filters to installers that register at least one of the given
	// CLI commands.
	CommandsAny []string

	// MinimumOSVersion filters to installers whose MinimumOSVersion contains
	// the given string.
	MinimumOSVersion string

	// Limit caps the number of results returned. 0 means no limit.
	Limit int

	// Offset skips the first N results (for pagination).
	Offset int
}

// InstallerResult is a single resolved installer entry together with its
// package identifier, version, and display name. It is returned by SearchInstallers.
type InstallerResult struct {
	PackageIdentifier string
	PackageVersion    string
	PackageName       string
	Publisher         string
	Installer         models.EffectiveInstaller
}

// InstallerSearchResponse wraps the results of a SearchInstallers call.
type InstallerSearchResponse struct {
	// TotalCount is the number of matching installers before Limit/Offset are applied.
	TotalCount int
	Results    []*InstallerResult
}
