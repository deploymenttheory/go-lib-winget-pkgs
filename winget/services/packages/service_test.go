package packages_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor/mock"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/index"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/packages"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/versions"
)

// newTestPackages builds the full service stack on top of the provided mock executor.
func newTestPackages(t *testing.T, exec *mock.Executor) *packages.Packages {
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
	ver := versions.NewVersions(exec, idx, mfst)

	return packages.NewPackages(exec, idx, ver, mfst)
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

func TestGetByID(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	pkg, err := svc.GetByID(context.Background(), "Microsoft.PowerShell")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if pkg.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", pkg.PackageIdentifier)
	}

	if pkg.PackageName != "PowerShell" {
		t.Errorf("PackageName = %q, want PowerShell", pkg.PackageName)
	}

	if pkg.Moniker != "pwsh" {
		t.Errorf("Moniker = %q, want pwsh", pkg.Moniker)
	}

	if len(pkg.Installers) != 1 {
		t.Errorf("len(Installers) = %d, want 1", len(pkg.Installers))
	}

	if pkg.Installers[0].InstallerType != "wix" {
		t.Errorf("InstallerType = %q, want wix (inherited from root)", pkg.Installers[0].InstallerType)
	}
}

func TestGetByID_CaseInsensitive(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	pkg, err := svc.GetByID(context.Background(), "microsoft.powershell")
	if err != nil {
		t.Fatalf("GetByID(lowercase): %v", err)
	}

	if pkg.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", pkg.PackageIdentifier)
	}
}

func TestGetByID_ValidationError(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	_, err := svc.GetByID(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty ID, got nil")
	}
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	_, err := svc.GetByID(context.Background(), "Nonexistent.Package")
	if err == nil {
		t.Error("expected not-found error, got nil")
	}
}

func TestGetByIDAndVersion_ValidationErrors(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByIDAndVersion(context.Background(), "", "1.0.0")
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

func TestListAll(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	resp, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}

	if len(resp.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(resp.Results))
	}
}

func TestListByPublisher(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	resp, err := svc.ListByPublisher(context.Background(), "Microsoft")
	if err != nil {
		t.Fatalf("ListByPublisher: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", resp.TotalCount)
	}

	if resp.Results[0].PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("Results[0].PackageIdentifier = %q, want Microsoft.PowerShell", resp.Results[0].PackageIdentifier)
	}
}

func TestListByPublisher_ValidationError(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	_, err := svc.ListByPublisher(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty publisher")
	}
}

func TestSearch_ByNameContains(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		NameContains: "powershell",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", resp.TotalCount)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty, cannot check PackageName")
	}

	if resp.Results[0].PackageName != "PowerShell" {
		t.Errorf("PackageName = %q, want PowerShell", resp.Results[0].PackageName)
	}
}

func TestSearch_ByInstallerType(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		InstallerType: "exe",
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1 (only Chrome is exe)", resp.TotalCount)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty, cannot check further")
	}
}

func TestSearch_ByTagsAny(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		TagsAny: []string{"browser"},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", resp.TotalCount)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty")
	}

	if resp.Results[0].PackageIdentifier != "Google.Chrome" {
		t.Errorf("unexpected result: %s", resp.Results[0].PackageIdentifier)
	}
}

func TestSearch_Limit(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{Limit: 1})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("TotalCount should reflect total before limit: got %d, want 2", resp.TotalCount)
	}

	if len(resp.Results) != 1 {
		t.Errorf("len(Results) = %d, want 1 (limited)", len(resp.Results))
	}
}

func TestSearch_Offset(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{Offset: 1})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("len(Results) = %d, want 1 (offset by 1 of 2)", len(resp.Results))
	}
}

func TestSearch_NilFilter(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	// A nil filter should behave identically to an empty FilterOptions.
	resp, err := svc.Search(context.Background(), nil)
	if err != nil {
		t.Fatalf("Search(nil): %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestGetByName(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	results, err := svc.GetByName(context.Background(), "PowerShell")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	if results[0].PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", results[0].PackageIdentifier)
	}
}

func TestGetByName_ValidationError(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newMockRepo())

	_, err := svc.GetByName(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty name")
	}
}
