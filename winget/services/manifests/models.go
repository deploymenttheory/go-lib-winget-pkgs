// Package manifests provides types and functions for reading and parsing
// WinGet manifest YAML files.
package manifests

import (
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/shared/models"
)

// VersionManifest is the index manifest for a specific package version.
// It ties together the installer and locale manifests.
//
//nolint:tagliatelle
type VersionManifest struct {
	PackageIdentifier string `yaml:"PackageIdentifier"`
	PackageVersion    string `yaml:"PackageVersion"`
	DefaultLocale     string `yaml:"DefaultLocale"`
	ManifestType      string `yaml:"ManifestType"`
	ManifestVersion   string `yaml:"ManifestVersion"`
}

// InstallerManifest contains all installation-related metadata for a package
// version. Root-level fields serve as defaults; each entry in Installers may
// override any of them.
//
//nolint:tagliatelle
type InstallerManifest struct {
	PackageIdentifier          string                        `yaml:"PackageIdentifier"`
	PackageVersion             string                        `yaml:"PackageVersion"`
	Channel                    string                        `yaml:"Channel,omitempty"`
	InstallerLocale            string                        `yaml:"InstallerLocale,omitempty"`
	Platform                   []string                      `yaml:"Platform,omitempty"`
	MinimumOSVersion           string                        `yaml:"MinimumOSVersion,omitempty"`
	InstallerType              string                        `yaml:"InstallerType,omitempty"`
	NestedInstallerType        string                        `yaml:"NestedInstallerType,omitempty"`
	NestedInstallerFiles       []models.NestedInstallerFile  `yaml:"NestedInstallerFiles,omitempty"`
	Scope                      string                        `yaml:"Scope,omitempty"`
	InstallModes               []string                      `yaml:"InstallModes,omitempty"`
	InstallerSwitches          *models.InstallerSwitches     `yaml:"InstallerSwitches,omitempty"`
	InstallerSuccessCodes      []int                         `yaml:"InstallerSuccessCodes,omitempty"`
	ExpectedReturnCodes        []models.ExpectedReturnCode   `yaml:"ExpectedReturnCodes,omitempty"`
	UpgradeBehavior            string                        `yaml:"UpgradeBehavior,omitempty"`
	Commands                   []string                      `yaml:"Commands,omitempty"`
	Protocols                  []string                      `yaml:"Protocols,omitempty"`
	FileExtensions             []string                      `yaml:"FileExtensions,omitempty"`
	Dependencies               *models.Dependencies          `yaml:"Dependencies,omitempty"`
	PackageFamilyName          string                        `yaml:"PackageFamilyName,omitempty"`
	ProductCode                string                        `yaml:"ProductCode,omitempty"`
	Capabilities               []string                      `yaml:"Capabilities,omitempty"`
	RestrictedCapabilities     []string                      `yaml:"RestrictedCapabilities,omitempty"`
	Markets                    *models.Markets               `yaml:"Markets,omitempty"`
	InstallerAbortsTerminal    *bool                         `yaml:"InstallerAbortsTerminal,omitempty"`
	ReleaseDate                string                        `yaml:"ReleaseDate,omitempty"`
	InstallLocationRequired    *bool                         `yaml:"InstallLocationRequired,omitempty"`
	RequireExplicitUpgrade     *bool                         `yaml:"RequireExplicitUpgrade,omitempty"`
	DisplayInstallWarnings     *bool                         `yaml:"DisplayInstallWarnings,omitempty"`
	UnsupportedOSArchitectures []string                      `yaml:"UnsupportedOSArchitectures,omitempty"`
	UnsupportedArguments       []string                      `yaml:"UnsupportedArguments,omitempty"`
	AppsAndFeaturesEntries     []models.AppsAndFeaturesEntry `yaml:"AppsAndFeaturesEntries,omitempty"`
	ElevationRequirement       string                        `yaml:"ElevationRequirement,omitempty"`
	InstallationMetadata       *models.InstallationMetadata  `yaml:"InstallationMetadata,omitempty"`
	DownloadCommandProhibited  *bool                         `yaml:"DownloadCommandProhibited,omitempty"`
	RepairBehavior             string                        `yaml:"RepairBehavior,omitempty"`
	ArchiveBinariesDependOnPath *bool                        `yaml:"ArchiveBinariesDependOnPath,omitempty"`
	Installers                 []RawInstaller                `yaml:"Installers"`
	ManifestType               string                        `yaml:"ManifestType"`
	ManifestVersion            string                        `yaml:"ManifestVersion"`
}

// RawInstaller is a single entry in InstallerManifest.Installers. Fields left
// at their zero value inherit the corresponding root-level default from the
// parent InstallerManifest. Bool fields use pointers so that unset (nil) can be
// distinguished from explicitly false.
//
//nolint:tagliatelle
type RawInstaller struct {
	Architecture               string                        `yaml:"Architecture"`
	InstallerURL               string                        `yaml:"InstallerUrl"`
	InstallerSha256            string                        `yaml:"InstallerSha256"`
	SignatureSha256             string                        `yaml:"SignatureSha256,omitempty"`
	InstallerType              string                        `yaml:"InstallerType,omitempty"`
	NestedInstallerType        string                        `yaml:"NestedInstallerType,omitempty"`
	NestedInstallerFiles       []models.NestedInstallerFile  `yaml:"NestedInstallerFiles,omitempty"`
	Scope                      string                        `yaml:"Scope,omitempty"`
	Platform                   []string                      `yaml:"Platform,omitempty"`
	MinimumOSVersion           string                        `yaml:"MinimumOSVersion,omitempty"`
	InstallModes               []string                      `yaml:"InstallModes,omitempty"`
	InstallerSwitches          *models.InstallerSwitches     `yaml:"InstallerSwitches,omitempty"`
	InstallerSuccessCodes      []int                         `yaml:"InstallerSuccessCodes,omitempty"`
	ExpectedReturnCodes        []models.ExpectedReturnCode   `yaml:"ExpectedReturnCodes,omitempty"`
	UpgradeBehavior            string                        `yaml:"UpgradeBehavior,omitempty"`
	Commands                   []string                      `yaml:"Commands,omitempty"`
	Protocols                  []string                      `yaml:"Protocols,omitempty"`
	FileExtensions             []string                      `yaml:"FileExtensions,omitempty"`
	Dependencies               *models.Dependencies          `yaml:"Dependencies,omitempty"`
	PackageFamilyName          string                        `yaml:"PackageFamilyName,omitempty"`
	ProductCode                string                        `yaml:"ProductCode,omitempty"`
	Capabilities               []string                      `yaml:"Capabilities,omitempty"`
	RestrictedCapabilities     []string                      `yaml:"RestrictedCapabilities,omitempty"`
	Markets                    *models.Markets               `yaml:"Markets,omitempty"`
	InstallerAbortsTerminal    *bool                         `yaml:"InstallerAbortsTerminal,omitempty"`
	ReleaseDate                string                        `yaml:"ReleaseDate,omitempty"`
	InstallLocationRequired    *bool                         `yaml:"InstallLocationRequired,omitempty"`
	RequireExplicitUpgrade     *bool                         `yaml:"RequireExplicitUpgrade,omitempty"`
	DisplayInstallWarnings     *bool                         `yaml:"DisplayInstallWarnings,omitempty"`
	UnsupportedOSArchitectures []string                      `yaml:"UnsupportedOSArchitectures,omitempty"`
	UnsupportedArguments       []string                      `yaml:"UnsupportedArguments,omitempty"`
	AppsAndFeaturesEntries     []models.AppsAndFeaturesEntry `yaml:"AppsAndFeaturesEntries,omitempty"`
	ElevationRequirement       string                        `yaml:"ElevationRequirement,omitempty"`
	InstallationMetadata       *models.InstallationMetadata  `yaml:"InstallationMetadata,omitempty"`
	DownloadCommandProhibited  *bool                         `yaml:"DownloadCommandProhibited,omitempty"`
	RepairBehavior             string                        `yaml:"RepairBehavior,omitempty"`
	ArchiveBinariesDependOnPath *bool                        `yaml:"ArchiveBinariesDependOnPath,omitempty"`
	Channel                    string                        `yaml:"Channel,omitempty"`
}

// DefaultLocaleManifest contains the primary (canonical-locale) metadata for a
// package version. Publisher, PackageName, License and ShortDescription are
// required by the WinGet schema; all other fields are optional.
//
//nolint:tagliatelle
type DefaultLocaleManifest struct {
	PackageIdentifier   string               `yaml:"PackageIdentifier"`
	PackageVersion      string               `yaml:"PackageVersion"`
	PackageLocale       string               `yaml:"PackageLocale"`
	Publisher           string               `yaml:"Publisher"`
	PublisherURL        string               `yaml:"PublisherUrl,omitempty"`
	PublisherSupportURL string               `yaml:"PublisherSupportUrl,omitempty"`
	PrivacyURL          string               `yaml:"PrivacyUrl,omitempty"`
	Author              string               `yaml:"Author,omitempty"`
	PackageName         string               `yaml:"PackageName"`
	PackageURL          string               `yaml:"PackageUrl,omitempty"`
	License             string               `yaml:"License"`
	LicenseURL          string               `yaml:"LicenseUrl,omitempty"`
	Copyright           string               `yaml:"Copyright,omitempty"`
	CopyrightURL        string               `yaml:"CopyrightUrl,omitempty"`
	ShortDescription    string               `yaml:"ShortDescription"`
	Description         string               `yaml:"Description,omitempty"`
	Moniker             string               `yaml:"Moniker,omitempty"`
	Tags                []string             `yaml:"Tags,omitempty"`
	Agreements          []models.Agreement   `yaml:"Agreements,omitempty"`
	ReleaseNotes        string               `yaml:"ReleaseNotes,omitempty"`
	ReleaseNotesURL     string               `yaml:"ReleaseNotesUrl,omitempty"`
	PurchaseURL         string               `yaml:"PurchaseUrl,omitempty"`
	InstallationNotes   string               `yaml:"InstallationNotes,omitempty"`
	Documentations      []models.Documentation `yaml:"Documentations,omitempty"`
	Icons               []models.Icon        `yaml:"Icons,omitempty"`
	ManifestType        string               `yaml:"ManifestType"`
	ManifestVersion     string               `yaml:"ManifestVersion"`
}

// LocaleManifest contains translated metadata for a specific locale. It is
// structurally similar to DefaultLocaleManifest but all metadata fields are
// optional and there is no Moniker field.
//
//nolint:tagliatelle
type LocaleManifest struct {
	PackageIdentifier   string               `yaml:"PackageIdentifier"`
	PackageVersion      string               `yaml:"PackageVersion"`
	PackageLocale       string               `yaml:"PackageLocale"`
	Publisher           string               `yaml:"Publisher,omitempty"`
	PublisherURL        string               `yaml:"PublisherUrl,omitempty"`
	PublisherSupportURL string               `yaml:"PublisherSupportUrl,omitempty"`
	PrivacyURL          string               `yaml:"PrivacyUrl,omitempty"`
	Author              string               `yaml:"Author,omitempty"`
	PackageName         string               `yaml:"PackageName,omitempty"`
	PackageURL          string               `yaml:"PackageUrl,omitempty"`
	License             string               `yaml:"License,omitempty"`
	LicenseURL          string               `yaml:"LicenseUrl,omitempty"`
	Copyright           string               `yaml:"Copyright,omitempty"`
	CopyrightURL        string               `yaml:"CopyrightUrl,omitempty"`
	ShortDescription    string               `yaml:"ShortDescription,omitempty"`
	Description         string               `yaml:"Description,omitempty"`
	Tags                []string             `yaml:"Tags,omitempty"`
	Agreements          []models.Agreement   `yaml:"Agreements,omitempty"`
	ReleaseNotes        string               `yaml:"ReleaseNotes,omitempty"`
	ReleaseNotesURL     string               `yaml:"ReleaseNotesUrl,omitempty"`
	PurchaseURL         string               `yaml:"PurchaseUrl,omitempty"`
	InstallationNotes   string               `yaml:"InstallationNotes,omitempty"`
	Documentations      []models.Documentation `yaml:"Documentations,omitempty"`
	Icons               []models.Icon        `yaml:"Icons,omitempty"`
	ManifestType        string               `yaml:"ManifestType"`
	ManifestVersion     string               `yaml:"ManifestVersion"`
}
