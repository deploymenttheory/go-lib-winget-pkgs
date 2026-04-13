// Package models contains shared types used across the winget library.
package models

// InstallerSwitches defines CLI switches passed to the installer.
//
//nolint:tagliatelle
type InstallerSwitches struct {
	Silent             string `yaml:"Silent,omitempty"`
	SilentWithProgress string `yaml:"SilentWithProgress,omitempty"`
	Interactive        string `yaml:"Interactive,omitempty"`
	InstallLocation    string `yaml:"InstallLocation,omitempty"`
	Log                string `yaml:"Log,omitempty"`
	Upgrade            string `yaml:"Upgrade,omitempty"`
	Custom             string `yaml:"Custom,omitempty"`
	Repair             string `yaml:"Repair,omitempty"`
}

// EffectiveInstaller is a fully-resolved installer entry. All fields are the
// result of merging the InstallerManifest root defaults with the per-installer
// overrides. Consumers never see raw unresolved data.
type EffectiveInstaller struct {
	Architecture               string
	InstallerURL               string
	InstallerSha256            string
	SignatureSha256             string
	InstallerType              string
	NestedInstallerType        string
	NestedInstallerFiles       []NestedInstallerFile
	Scope                      string
	Platform                   []string
	MinimumOSVersion           string
	InstallModes               []string
	InstallerSwitches          InstallerSwitches
	InstallerSuccessCodes      []int
	ExpectedReturnCodes        []ExpectedReturnCode
	UpgradeBehavior            string
	Commands                   []string
	Protocols                  []string
	FileExtensions             []string
	Dependencies               *Dependencies
	PackageFamilyName          string
	ProductCode                string
	Capabilities               []string
	RestrictedCapabilities     []string
	Markets                    *Markets
	InstallerAbortsTerminal    bool
	ReleaseDate                string
	InstallLocationRequired    bool
	RequireExplicitUpgrade     bool
	DisplayInstallWarnings     bool
	UnsupportedOSArchitectures []string
	UnsupportedArguments       []string
	AppsAndFeaturesEntries     []AppsAndFeaturesEntry
	ElevationRequirement       string
	InstallationMetadata       *InstallationMetadata
	DownloadCommandProhibited  bool
	RepairBehavior             string
	ArchiveBinariesDependOnPath bool
	Channel                    string
}

// Dependencies declares runtime dependencies for the installer.
//
//nolint:tagliatelle
type Dependencies struct {
	WindowsFeatures      []string           `yaml:"WindowsFeatures,omitempty"`
	WindowsLibraries     []string           `yaml:"WindowsLibraries,omitempty"`
	PackageDependencies  []PackageDependency `yaml:"PackageDependencies,omitempty"`
	ExternalDependencies []string           `yaml:"ExternalDependencies,omitempty"`
}

// PackageDependency identifies another WinGet package required at runtime.
//
//nolint:tagliatelle
type PackageDependency struct {
	PackageIdentifier string `yaml:"PackageIdentifier"`
	MinimumVersion    string `yaml:"MinimumVersion,omitempty"`
}

// Markets restricts which countries/regions a package is available in.
//
//nolint:tagliatelle
type Markets struct {
	AllowedMarkets  []string `yaml:"AllowedMarkets,omitempty"`
	ExcludedMarkets []string `yaml:"ExcludedMarkets,omitempty"`
}

// AppsAndFeaturesEntry overrides ARP (Add/Remove Programs) registry values.
//
//nolint:tagliatelle
type AppsAndFeaturesEntry struct {
	DisplayName    string `yaml:"DisplayName,omitempty"`
	Publisher      string `yaml:"Publisher,omitempty"`
	DisplayVersion string `yaml:"DisplayVersion,omitempty"`
	ProductCode    string `yaml:"ProductCode,omitempty"`
	UpgradeCode    string `yaml:"UpgradeCode,omitempty"`
	InstallerType  string `yaml:"InstallerType,omitempty"`
}

// InstallationMetadata provides file-based detection metadata for the installer.
//
//nolint:tagliatelle
type InstallationMetadata struct {
	DefaultInstallLocation string                  `yaml:"DefaultInstallLocation,omitempty"`
	Files                  []InstallationMetadataFile `yaml:"Files,omitempty"`
}

// InstallationMetadataFile describes a single installed file used for detection.
//
//nolint:tagliatelle
type InstallationMetadataFile struct {
	RelativeFilePath    string `yaml:"RelativeFilePath"`
	FileSha256          string `yaml:"FileSha256,omitempty"`
	FileType            string `yaml:"FileType,omitempty"`
	InvocationParameter string `yaml:"InvocationParameter,omitempty"`
	DisplayName         string `yaml:"DisplayName,omitempty"`
}

// ExpectedReturnCode maps a non-zero installer exit code to a named response.
//
//nolint:tagliatelle
type ExpectedReturnCode struct {
	InstallerReturnCode int    `yaml:"InstallerReturnCode"`
	ReturnResponse      string `yaml:"ReturnResponse"`
	ReturnResponseURL   string `yaml:"ReturnResponseUrl,omitempty"`
}

// NestedInstallerFile describes a file extracted from an archive installer.
//
//nolint:tagliatelle
type NestedInstallerFile struct {
	RelativeFilePath     string `yaml:"RelativeFilePath"`
	PortableCommandAlias string `yaml:"PortableCommandAlias,omitempty"`
}

// Agreement describes a license or EULA agreement associated with a package.
//
//nolint:tagliatelle
type Agreement struct {
	AgreementLabel string `yaml:"AgreementLabel,omitempty"`
	Agreement      string `yaml:"Agreement,omitempty"`
	AgreementURL   string `yaml:"AgreementUrl,omitempty"`
}

// Documentation is a labelled link to external documentation.
//
//nolint:tagliatelle
type Documentation struct {
	DocumentLabel string `yaml:"DocumentLabel,omitempty"`
	DocumentURL   string `yaml:"DocumentUrl,omitempty"`
}

// Icon describes a package icon asset.
//
//nolint:tagliatelle
type Icon struct {
	IconURL        string `yaml:"IconUrl"`
	IconFileType   string `yaml:"IconFileType"`
	IconResolution string `yaml:"IconResolution,omitempty"`
	IconTheme      string `yaml:"IconTheme,omitempty"`
	IconSha256     string `yaml:"IconSha256,omitempty"`
}
