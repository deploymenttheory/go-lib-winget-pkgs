// Package versions provides access to the available versions of WinGet packages.
package versions

import (
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/shared/models"
)

// ResourceVersion is the fully-resolved metadata for a specific version of a
// WinGet package. It merges the version, installer, and defaultLocale manifests
// into a single struct.
type ResourceVersion struct {
	PackageIdentifier   string
	PackageVersion      string
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

// ListVersionsResponse contains the available versions for a package.
type ListVersionsResponse struct {
	PackageIdentifier string
	Versions          []string // sorted newest first
}
