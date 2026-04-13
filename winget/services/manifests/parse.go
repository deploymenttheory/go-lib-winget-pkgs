package manifests

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/shared/models"
)

// ParseVersionManifest parses the raw YAML bytes of a version manifest.
func ParseVersionManifest(data []byte) (*VersionManifest, error) {
	var m VersionManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing version manifest: %w", err)
	}

	return &m, nil
}

// ParseInstallerManifest parses the raw YAML bytes of an installer manifest.
func ParseInstallerManifest(data []byte) (*InstallerManifest, error) {
	var m InstallerManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing installer manifest: %w", err)
	}

	return &m, nil
}

// ParseDefaultLocaleManifest parses the raw YAML bytes of a defaultLocale manifest.
func ParseDefaultLocaleManifest(data []byte) (*DefaultLocaleManifest, error) {
	var m DefaultLocaleManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing defaultLocale manifest: %w", err)
	}

	return &m, nil
}

// ParseLocaleManifest parses the raw YAML bytes of a locale manifest.
func ParseLocaleManifest(data []byte) (*LocaleManifest, error) {
	var m LocaleManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing locale manifest: %w", err)
	}

	return &m, nil
}

// FlattenInstallers resolves the two-level inheritance model used by WinGet
// installer manifests: root-level fields act as defaults and per-installer
// entries may override any of them. The returned slice contains one fully
// resolved EffectiveInstaller per entry in manifest.Installers.
func FlattenInstallers(manifest *InstallerManifest) []models.EffectiveInstaller {
	result := make([]models.EffectiveInstaller, 0, len(manifest.Installers))

	for _, raw := range manifest.Installers {
		eff := models.EffectiveInstaller{
			Architecture:               raw.Architecture,
			InstallerURL:               raw.InstallerURL,
			InstallerSha256:            raw.InstallerSha256,
			SignatureSha256:             raw.SignatureSha256,
			InstallerType:              coalesceStr(raw.InstallerType, manifest.InstallerType),
			NestedInstallerType:        coalesceStr(raw.NestedInstallerType, manifest.NestedInstallerType),
			NestedInstallerFiles:       coalesceSlice(raw.NestedInstallerFiles, manifest.NestedInstallerFiles),
			Scope:                      coalesceStr(raw.Scope, manifest.Scope),
			Platform:                   coalesceStrSlice(raw.Platform, manifest.Platform),
			MinimumOSVersion:           coalesceStr(raw.MinimumOSVersion, manifest.MinimumOSVersion),
			InstallModes:               coalesceStrSlice(raw.InstallModes, manifest.InstallModes),
			InstallerSwitches:          resolveSwitches(raw.InstallerSwitches, manifest.InstallerSwitches),
			InstallerSuccessCodes:      coalesceIntSlice(raw.InstallerSuccessCodes, manifest.InstallerSuccessCodes),
			ExpectedReturnCodes:        coalesceReturnCodes(raw.ExpectedReturnCodes, manifest.ExpectedReturnCodes),
			UpgradeBehavior:            coalesceStr(raw.UpgradeBehavior, manifest.UpgradeBehavior),
			Commands:                   coalesceStrSlice(raw.Commands, manifest.Commands),
			Protocols:                  coalesceStrSlice(raw.Protocols, manifest.Protocols),
			FileExtensions:             coalesceStrSlice(raw.FileExtensions, manifest.FileExtensions),
			Dependencies:               coalesceDeps(raw.Dependencies, manifest.Dependencies),
			PackageFamilyName:          coalesceStr(raw.PackageFamilyName, manifest.PackageFamilyName),
			ProductCode:                coalesceStr(raw.ProductCode, manifest.ProductCode),
			Capabilities:               coalesceStrSlice(raw.Capabilities, manifest.Capabilities),
			RestrictedCapabilities:     coalesceStrSlice(raw.RestrictedCapabilities, manifest.RestrictedCapabilities),
			Markets:                    coalesceMarkets(raw.Markets, manifest.Markets),
			InstallerAbortsTerminal:    coalesceBool(raw.InstallerAbortsTerminal, manifest.InstallerAbortsTerminal),
			ReleaseDate:                coalesceStr(raw.ReleaseDate, manifest.ReleaseDate),
			InstallLocationRequired:    coalesceBool(raw.InstallLocationRequired, manifest.InstallLocationRequired),
			RequireExplicitUpgrade:     coalesceBool(raw.RequireExplicitUpgrade, manifest.RequireExplicitUpgrade),
			DisplayInstallWarnings:     coalesceBool(raw.DisplayInstallWarnings, manifest.DisplayInstallWarnings),
			UnsupportedOSArchitectures: coalesceStrSlice(raw.UnsupportedOSArchitectures, manifest.UnsupportedOSArchitectures),
			UnsupportedArguments:       coalesceStrSlice(raw.UnsupportedArguments, manifest.UnsupportedArguments),
			AppsAndFeaturesEntries:     coalesceARPEntries(raw.AppsAndFeaturesEntries, manifest.AppsAndFeaturesEntries),
			ElevationRequirement:       coalesceStr(raw.ElevationRequirement, manifest.ElevationRequirement),
			InstallationMetadata:       coalesceMetadata(raw.InstallationMetadata, manifest.InstallationMetadata),
			DownloadCommandProhibited:  coalesceBool(raw.DownloadCommandProhibited, manifest.DownloadCommandProhibited),
			RepairBehavior:             coalesceStr(raw.RepairBehavior, manifest.RepairBehavior),
			ArchiveBinariesDependOnPath: coalesceBool(raw.ArchiveBinariesDependOnPath, manifest.ArchiveBinariesDependOnPath),
			Channel:                    coalesceStr(raw.Channel, manifest.Channel),
		}

		result = append(result, eff)
	}

	return result
}

// coalesceStr returns override if non-empty, otherwise fallback.
func coalesceStr(override, fallback string) string {
	if override != "" {
		return override
	}

	return fallback
}

// coalesceStrSlice returns override if non-nil (even if empty), otherwise fallback.
func coalesceStrSlice(override, fallback []string) []string {
	if override != nil {
		return override
	}

	return fallback
}

func coalesceIntSlice(override, fallback []int) []int {
	if override != nil {
		return override
	}

	return fallback
}

func coalesceSlice(override, fallback []models.NestedInstallerFile) []models.NestedInstallerFile {
	if override != nil {
		return override
	}

	return fallback
}

func coalesceReturnCodes(override, fallback []models.ExpectedReturnCode) []models.ExpectedReturnCode {
	if override != nil {
		return override
	}

	return fallback
}

func coalesceARPEntries(override, fallback []models.AppsAndFeaturesEntry) []models.AppsAndFeaturesEntry {
	if override != nil {
		return override
	}

	return fallback
}

// coalesceBool returns the value of override if set (non-nil), otherwise the
// value of fallback. If both are nil, returns false.
func coalesceBool(override, fallback *bool) bool {
	if override != nil {
		return *override
	}

	if fallback != nil {
		return *fallback
	}

	return false
}

func resolveSwitches(override, fallback *models.InstallerSwitches) models.InstallerSwitches {
	if override != nil {
		return *override
	}

	if fallback != nil {
		return *fallback
	}

	return models.InstallerSwitches{}
}

func coalesceDeps(override, fallback *models.Dependencies) *models.Dependencies {
	if override != nil {
		return override
	}

	return fallback
}

func coalesceMarkets(override, fallback *models.Markets) *models.Markets {
	if override != nil {
		return override
	}

	return fallback
}

func coalesceMetadata(override, fallback *models.InstallationMetadata) *models.InstallationMetadata {
	if override != nil {
		return override
	}

	return fallback
}
