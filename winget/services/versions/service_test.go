package versions_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor/mock"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/index"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/versions"
)

// newTestVersions builds the versions service stack on top of the provided mock executor.
func newTestVersions(t *testing.T, exec *mock.Executor) *versions.Versions {
	t.Helper()

	cfg := &config.Config{}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("config.Validate: %v", err)
	}

	idx, err := index.Build(context.Background(), exec, 1)
	if err != nil {
		t.Fatalf("index.Build: %v", err)
	}

	mfst := manifests.NewManifests(exec)

	return versions.NewVersions(exec, idx, mfst)
}

func newMockRepo() *mock.Executor {
	m := mock.New()

	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.yaml",
		[]byte("PackageIdentifier: Microsoft.PowerShell\nPackageVersion: 7.5.0.0\nDefaultLocale: en-US\nManifestType: version\nManifestVersion: 1.9.0\n"),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.installer.yaml",
		[]byte(`PackageIdentifier: Microsoft.PowerShell
PackageVersion: 7.5.0.0
InstallerType: wix
Scope: machine
Installers:
- Architecture: x64
  InstallerUrl: https://example.com/pwsh-x64.msi
  InstallerSha256: AABBCCDDEEFF00112233445566778899AABBCCDDEEFF00112233445566778899
ManifestType: installer
ManifestVersion: 1.9.0
`),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.locale.en-US.yaml",
		[]byte(`PackageIdentifier: Microsoft.PowerShell
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
`),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.4.6.0/Microsoft.PowerShell.yaml",
		[]byte("PackageIdentifier: Microsoft.PowerShell\nPackageVersion: 7.4.6.0\nDefaultLocale: en-US\nManifestType: version\nManifestVersion: 1.9.0\n"),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.4.6.0/Microsoft.PowerShell.installer.yaml",
		[]byte(`PackageIdentifier: Microsoft.PowerShell
PackageVersion: 7.4.6.0
InstallerType: wix
Scope: machine
Installers:
- Architecture: x64
  InstallerUrl: https://example.com/pwsh-7.4.6-x64.msi
  InstallerSha256: AABBCCDDEEFF00112233445566778899AABBCCDDEEFF00112233445566778899
ManifestType: installer
ManifestVersion: 1.9.0
`),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.4.6.0/Microsoft.PowerShell.locale.en-US.yaml",
		[]byte(`PackageIdentifier: Microsoft.PowerShell
PackageVersion: 7.4.6.0
PackageLocale: en-US
Publisher: Microsoft Corporation
PackageName: PowerShell
License: MIT
ShortDescription: Cross-platform task automation solution.
Moniker: pwsh
Tags:
- shell
ManifestType: defaultLocale
ManifestVersion: 1.9.0
`),
	)
	m.AddFile(
		"manifests/g/Google/Chrome/125.0.6422.60/Google.Chrome.yaml",
		[]byte("PackageIdentifier: Google.Chrome\nPackageVersion: 125.0.6422.60\nDefaultLocale: en-US\nManifestType: version\nManifestVersion: 1.9.0\n"),
	)
	m.AddFile(
		"manifests/g/Google/Chrome/125.0.6422.60/Google.Chrome.installer.yaml",
		[]byte(`PackageIdentifier: Google.Chrome
PackageVersion: 125.0.6422.60
InstallerType: exe
Scope: machine
Installers:
- Architecture: x64
  InstallerUrl: https://example.com/chrome.exe
  InstallerSha256: AABBCCDDEEFF00112233445566778899AABBCCDDEEFF00112233445566778899
ManifestType: installer
ManifestVersion: 1.9.0
`),
	)
	m.AddFile(
		"manifests/g/Google/Chrome/125.0.6422.60/Google.Chrome.locale.en-US.yaml",
		[]byte(`PackageIdentifier: Google.Chrome
PackageVersion: 125.0.6422.60
PackageLocale: en-US
Publisher: Google LLC
PackageName: Google Chrome
License: Proprietary
ShortDescription: Fast web browser.
Tags:
- browser
ManifestType: defaultLocale
ManifestVersion: 1.9.0
`),
	)

	return m
}

func TestListByID(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	resp, err := svc.ListByID(context.Background(), "Microsoft.PowerShell")
	if err != nil {
		t.Fatalf("ListByID: %v", err)
	}

	if resp.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", resp.PackageIdentifier)
	}

	if len(resp.Versions) != 2 {
		t.Errorf("len(Versions) = %d, want 2", len(resp.Versions))
	}

	// Versions must be sorted newest first.
	if len(resp.Versions) >= 2 && resp.Versions[0] != "7.5.0.0" {
		t.Errorf("Versions[0] = %q, want 7.5.0.0 (newest first)", resp.Versions[0])
	}
}

func TestListByID_CaseInsensitive(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	resp, err := svc.ListByID(context.Background(), "microsoft.powershell")
	if err != nil {
		t.Fatalf("ListByID(lowercase): %v", err)
	}

	if resp.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", resp.PackageIdentifier)
	}
}

func TestListByID_ValidationError(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	_, err := svc.ListByID(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestListByID_NotFound(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	_, err := svc.ListByID(context.Background(), "Nonexistent.Package")
	if err == nil {
		t.Error("expected not-found error, got nil")
	}
}

func TestGetByIDAndVersion(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	rv, err := svc.GetByIDAndVersion(context.Background(), "Microsoft.PowerShell", "7.4.6.0")
	if err != nil {
		t.Fatalf("GetByIDAndVersion: %v", err)
	}

	if rv.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", rv.PackageIdentifier)
	}

	if rv.PackageVersion != "7.4.6.0" {
		t.Errorf("PackageVersion = %q, want 7.4.6.0", rv.PackageVersion)
	}

	if rv.PackageName != "PowerShell" {
		t.Errorf("PackageName = %q, want PowerShell", rv.PackageName)
	}

	if rv.Publisher != "Microsoft Corporation" {
		t.Errorf("Publisher = %q, want Microsoft Corporation", rv.Publisher)
	}

	if len(rv.Installers) != 1 {
		t.Errorf("len(Installers) = %d, want 1", len(rv.Installers))
	}

	if rv.Installers[0].InstallerType != "wix" {
		t.Errorf("InstallerType = %q, want wix (inherited from root)", rv.Installers[0].InstallerType)
	}
}

func TestGetByIDAndVersion_CaseInsensitive(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	rv, err := svc.GetByIDAndVersion(context.Background(), "MICROSOFT.POWERSHELL", "7.5.0.0")
	if err != nil {
		t.Fatalf("GetByIDAndVersion(uppercase): %v", err)
	}

	if rv.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", rv.PackageIdentifier)
	}
}

func TestGetByIDAndVersion_NotFound(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	t.Run("unknown package", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByIDAndVersion(context.Background(), "Nonexistent.Package", "1.0.0")
		if err == nil {
			t.Error("expected not-found error for unknown package, got nil")
		}
	})

	t.Run("unknown version", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByIDAndVersion(context.Background(), "Microsoft.PowerShell", "0.0.0.0")
		if err == nil {
			t.Error("expected not-found error for unknown version, got nil")
		}
	})
}

func TestGetByIDAndVersion_ValidationErrors(t *testing.T) {
	t.Parallel()

	svc := newTestVersions(t, newMockRepo())

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByIDAndVersion(context.Background(), "", "7.5.0.0")
		if err == nil {
			t.Error("expected error for empty ID")
		}
	})

	t.Run("empty version", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByIDAndVersion(context.Background(), "Microsoft.PowerShell", "")
		if err == nil {
			t.Error("expected error for empty version")
		}
	})
}
