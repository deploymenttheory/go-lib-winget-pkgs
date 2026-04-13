package manifests_test

import (
	"testing"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/shared/models"
)

var powershellInstallerYAML = []byte(`
PackageIdentifier: Microsoft.PowerShell
PackageVersion: 7.5.0.0
MinimumOSVersion: 10.0.17763.0
InstallerType: wix
Scope: machine
InstallModes:
- interactive
- silent
- silentWithProgress
UpgradeBehavior: install
Commands:
- pwsh
ReleaseDate: 2025-01-22
Installers:
- Architecture: x64
  InstallerUrl: https://example.com/PowerShell-7.5.0-win-x64.msi
  InstallerSha256: 6B988B7E236A8E1CF1166D3BE289D3A20AA344499153BDAADD2F9FEDFFC6EDA9
  ProductCode: '{D012DCD1-67EA-4627-938F-19FD677FC03A}'
- Architecture: x86
  InstallerUrl: https://example.com/PowerShell-7.5.0-win-x86.msi
  InstallerSha256: 25BDF464E4050B7DD0E6034F2AE34D1111596A4A497EAA89862D0B5928825F58
  ProductCode: '{57413274-F91F-4B8E-B61A-A13FE70D4072}'
  Scope: user
ManifestType: installer
ManifestVersion: 1.9.0
`)

var powershellLocaleYAML = []byte(`
PackageIdentifier: Microsoft.PowerShell
PackageVersion: 7.5.0.0
PackageLocale: en-US
Publisher: Microsoft Corporation
PackageName: PowerShell
License: MIT
ShortDescription: Cross-platform task automation solution.
Moniker: pwsh
Tags:
- shell
- cross-platform
ManifestType: defaultLocale
ManifestVersion: 1.9.0
`)

var powershellVersionYAML = []byte(`
PackageIdentifier: Microsoft.PowerShell
PackageVersion: 7.5.0.0
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.9.0
`)

func TestParseVersionManifest(t *testing.T) {
	t.Parallel()

	m, err := manifests.ParseVersionManifest(powershellVersionYAML)
	if err != nil {
		t.Fatalf("ParseVersionManifest: %v", err)
	}

	if m.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want %q", m.PackageIdentifier, "Microsoft.PowerShell")
	}

	if m.PackageVersion != "7.5.0.0" {
		t.Errorf("PackageVersion = %q, want %q", m.PackageVersion, "7.5.0.0")
	}

	if m.DefaultLocale != "en-US" {
		t.Errorf("DefaultLocale = %q, want %q", m.DefaultLocale, "en-US")
	}
}

func TestParseInstallerManifest(t *testing.T) {
	t.Parallel()

	m, err := manifests.ParseInstallerManifest(powershellInstallerYAML)
	if err != nil {
		t.Fatalf("ParseInstallerManifest: %v", err)
	}

	if m.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want %q", m.PackageIdentifier, "Microsoft.PowerShell")
	}

	if m.InstallerType != "wix" {
		t.Errorf("InstallerType = %q, want %q", m.InstallerType, "wix")
	}

	if m.Scope != "machine" {
		t.Errorf("Scope = %q, want %q", m.Scope, "machine")
	}

	if len(m.Installers) != 2 {
		t.Fatalf("len(Installers) = %d, want 2", len(m.Installers))
	}

	if m.Installers[0].Architecture != "x64" {
		t.Errorf("Installers[0].Architecture = %q, want %q", m.Installers[0].Architecture, "x64")
	}
}

func TestParseDefaultLocaleManifest(t *testing.T) {
	t.Parallel()

	m, err := manifests.ParseDefaultLocaleManifest(powershellLocaleYAML)
	if err != nil {
		t.Fatalf("ParseDefaultLocaleManifest: %v", err)
	}

	if m.Publisher != "Microsoft Corporation" {
		t.Errorf("Publisher = %q, want %q", m.Publisher, "Microsoft Corporation")
	}

	if m.PackageName != "PowerShell" {
		t.Errorf("PackageName = %q, want %q", m.PackageName, "PowerShell")
	}

	if m.Moniker != "pwsh" {
		t.Errorf("Moniker = %q, want %q", m.Moniker, "pwsh")
	}

	if len(m.Tags) != 2 {
		t.Errorf("len(Tags) = %d, want 2", len(m.Tags))
	}
}

func TestFlattenInstallers_InheritsRootDefaults(t *testing.T) {
	t.Parallel()

	m, err := manifests.ParseInstallerManifest(powershellInstallerYAML)
	if err != nil {
		t.Fatalf("ParseInstallerManifest: %v", err)
	}

	effective := manifests.FlattenInstallers(m)

	if len(effective) != 2 {
		t.Fatalf("len(effective) = %d, want 2", len(effective))
	}

	// x64 installer should inherit root-level InstallerType, Scope, Commands.
	x64 := effective[0]

	if x64.InstallerType != "wix" {
		t.Errorf("x64 InstallerType = %q, want %q (inherited from root)", x64.InstallerType, "wix")
	}

	if x64.Scope != "machine" {
		t.Errorf("x64 Scope = %q, want %q (inherited from root)", x64.Scope, "machine")
	}

	if len(x64.Commands) == 0 || x64.Commands[0] != "pwsh" {
		t.Errorf("x64 Commands = %v, want [pwsh] (inherited from root)", x64.Commands)
	}

	if x64.ReleaseDate != "2025-01-22" {
		t.Errorf("x64 ReleaseDate = %q, want %q", x64.ReleaseDate, "2025-01-22")
	}
}

func TestFlattenInstallers_PerInstallerOverridesRoot(t *testing.T) {
	t.Parallel()

	m, err := manifests.ParseInstallerManifest(powershellInstallerYAML)
	if err != nil {
		t.Fatalf("ParseInstallerManifest: %v", err)
	}

	effective := manifests.FlattenInstallers(m)

	// x86 installer explicitly sets Scope: user, which should override root's machine.
	x86 := effective[1]

	if x86.Scope != "user" {
		t.Errorf("x86 Scope = %q, want %q (per-installer override)", x86.Scope, "user")
	}

	// x86 should still inherit InstallerType from root.
	if x86.InstallerType != "wix" {
		t.Errorf("x86 InstallerType = %q, want %q (inherited)", x86.InstallerType, "wix")
	}
}

func TestFlattenInstallers_BoolPointerInheritance(t *testing.T) {
	t.Parallel()

	trueVal := true

	m := &manifests.InstallerManifest{
		PackageIdentifier:       "Test.Package",
		PackageVersion:          "1.0.0",
		InstallerType:           "exe",
		RequireExplicitUpgrade:  &trueVal,
		Installers: []manifests.RawInstaller{
			{
				Architecture:    "x64",
				InstallerURL:    "https://example.com/setup.exe",
				InstallerSha256: "abc123",
				// RequireExplicitUpgrade is nil → should inherit root true
			},
		},
	}

	effective := manifests.FlattenInstallers(m)
	if len(effective) != 1 {
		t.Fatalf("len(effective) = %d, want 1", len(effective))
	}

	if !effective[0].RequireExplicitUpgrade {
		t.Error("RequireExplicitUpgrade should be true (inherited from root)")
	}
}

func TestFlattenInstallers_Empty(t *testing.T) {
	t.Parallel()

	m := &manifests.InstallerManifest{
		PackageIdentifier: "Empty.Package",
		PackageVersion:    "1.0.0",
		Installers:        nil,
	}

	effective := manifests.FlattenInstallers(m)
	if effective == nil {
		t.Error("FlattenInstallers should return a non-nil slice for empty input")
	}

	if len(effective) != 0 {
		t.Errorf("len(effective) = %d, want 0", len(effective))
	}
}

func TestParseInstallerManifest_InvalidYAML(t *testing.T) {
	t.Parallel()

	_, err := manifests.ParseInstallerManifest([]byte("not: valid: yaml: ["))
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestFlattenInstallers_SwitchesInheritance(t *testing.T) {
	t.Parallel()

	m := &manifests.InstallerManifest{
		PackageIdentifier: "Test.Package",
		PackageVersion:    "1.0.0",
		InstallerType:     "exe",
		InstallerSwitches: &models.InstallerSwitches{
			Silent:   "/S",
			Upgrade:  "/U",
		},
		Installers: []manifests.RawInstaller{
			{
				Architecture:    "x64",
				InstallerURL:    "https://example.com/setup.exe",
				InstallerSha256: "abc123",
				// No InstallerSwitches set — should inherit root
			},
			{
				Architecture:    "x86",
				InstallerURL:    "https://example.com/setup-x86.exe",
				InstallerSha256: "def456",
				InstallerSwitches: &models.InstallerSwitches{
					Silent: "/SILENT",
					// Upgrade not set
				},
			},
		},
	}

	effective := manifests.FlattenInstallers(m)

	// x64 inherits root switches entirely.
	if effective[0].InstallerSwitches.Silent != "/S" {
		t.Errorf("x64 Silent = %q, want /S", effective[0].InstallerSwitches.Silent)
	}

	if effective[0].InstallerSwitches.Upgrade != "/U" {
		t.Errorf("x64 Upgrade = %q, want /U", effective[0].InstallerSwitches.Upgrade)
	}

	// x86 overrides switches entirely (whole struct, not field-by-field).
	if effective[1].InstallerSwitches.Silent != "/SILENT" {
		t.Errorf("x86 Silent = %q, want /SILENT", effective[1].InstallerSwitches.Silent)
	}
}
