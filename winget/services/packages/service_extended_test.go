package packages_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor/mock"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/packages"
)

// newRichMockRepo builds a mock repo with richer YAML that includes ProductCode,
// Commands, PackageFamilyName, and MinimumOSVersion so the extended filter and
// arch-based methods can be exercised.
func newRichMockRepo() *mock.Executor {
	m := mock.New()

	// Microsoft.PowerShell — wix, x64 only, with ProductCode + Commands
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
MinimumOSVersion: 10.0.17763.0
Commands:
- pwsh
Installers:
- Architecture: x64
  InstallerUrl: https://example.com/pwsh-x64.msi
  InstallerSha256: AABBCCDDEEFF00112233445566778899AABBCCDDEEFF00112233445566778899
  ProductCode: '{D012DCD1-67EA-4627-938F-19FD677FC03A}'
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
ManifestType: defaultLocale
ManifestVersion: 1.9.0
`),
	)

	// Google.Chrome — exe, x64, with PackageFamilyName + MinimumOSVersion
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
MinimumOSVersion: 10.0.19041.0
Installers:
- Architecture: x64
  InstallerUrl: https://example.com/chrome.exe
  InstallerSha256: AABBCCDDEEFF00112233445566778899AABBCCDDEEFF00112233445566778899
  PackageFamilyName: Google.Chrome_8wekyb3d8bbwe
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

// --- GetByIDAndArchitecture ---

func TestGetByIDAndArchitecture(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	pkg, err := svc.GetByIDAndArchitecture(context.Background(), "Microsoft.PowerShell", "x64")
	if err != nil {
		t.Fatalf("GetByIDAndArchitecture: %v", err)
	}

	if pkg.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", pkg.PackageIdentifier)
	}

	if len(pkg.Installers) != 1 {
		t.Fatalf("len(Installers) = %d, want 1", len(pkg.Installers))
	}

	if pkg.Installers[0].Architecture != "x64" {
		t.Errorf("Architecture = %q, want x64", pkg.Installers[0].Architecture)
	}
}

func TestGetByIDAndArchitecture_CaseInsensitiveID(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	pkg, err := svc.GetByIDAndArchitecture(context.Background(), "microsoft.powershell", "x64")
	if err != nil {
		t.Fatalf("GetByIDAndArchitecture(lowercase id): %v", err)
	}

	if pkg.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", pkg.PackageIdentifier)
	}
}

func TestGetByIDAndArchitecture_CaseInsensitiveArch(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	pkg, err := svc.GetByIDAndArchitecture(context.Background(), "Microsoft.PowerShell", "X64")
	if err != nil {
		t.Fatalf("GetByIDAndArchitecture(uppercase arch): %v", err)
	}

	if len(pkg.Installers) != 1 {
		t.Errorf("len(Installers) = %d, want 1", len(pkg.Installers))
	}
}

func TestGetByIDAndArchitecture_ArchNotPresent(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	_, err := svc.GetByIDAndArchitecture(context.Background(), "Microsoft.PowerShell", "arm64")
	if err == nil {
		t.Error("expected error when architecture is not present, got nil")
	}
}

func TestGetByIDAndArchitecture_PackageNotFound(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	_, err := svc.GetByIDAndArchitecture(context.Background(), "Nonexistent.Package", "x64")
	if err == nil {
		t.Error("expected not-found error, got nil")
	}
}

func TestGetByIDAndArchitecture_ValidationErrors(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	t.Run("empty id", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByIDAndArchitecture(context.Background(), "", "x64")
		if err == nil {
			t.Error("expected error for empty ID")
		}
	})

	t.Run("empty arch", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByIDAndArchitecture(context.Background(), "Microsoft.PowerShell", "")
		if err == nil {
			t.Error("expected error for empty architecture")
		}
	})
}

// --- GetByNameAndArchitecture ---

func TestGetByNameAndArchitecture(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	results, err := svc.GetByNameAndArchitecture(context.Background(), "PowerShell", "x64")
	if err != nil {
		t.Fatalf("GetByNameAndArchitecture: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}

	if results[0].PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", results[0].PackageIdentifier)
	}

	for _, inst := range results[0].Installers {
		if inst.Architecture != "x64" {
			t.Errorf("unexpected non-x64 installer: %s", inst.Architecture)
		}
	}
}

func TestGetByNameAndArchitecture_ArchNotPresent(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	// arm64 doesn't exist in the mock — expect empty results, not an error.
	results, err := svc.GetByNameAndArchitecture(context.Background(), "PowerShell", "arm64")
	if err != nil {
		t.Fatalf("GetByNameAndArchitecture: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0 (no arm64 installers)", len(results))
	}
}

func TestGetByNameAndArchitecture_ValidationErrors(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	t.Run("empty name", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByNameAndArchitecture(context.Background(), "", "x64")
		if err == nil {
			t.Error("expected error for empty name")
		}
	})

	t.Run("empty arch", func(t *testing.T) {
		t.Parallel()

		_, err := svc.GetByNameAndArchitecture(context.Background(), "PowerShell", "")
		if err == nil {
			t.Error("expected error for empty architecture")
		}
	})
}

// --- Extended FilterOptions (Search) ---

func TestSearch_ByProductCode(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		ProductCode: "{D012DCD1-67EA-4627-938F-19FD677FC03A}",
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

	if resp.Results[0].PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", resp.Results[0].PackageIdentifier)
	}
}

func TestSearch_ByPackageFamilyName(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		PackageFamilyName: "Google.Chrome_8wekyb3d8bbwe",
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
		t.Errorf("PackageIdentifier = %q, want Google.Chrome", resp.Results[0].PackageIdentifier)
	}
}

func TestSearch_ByCommandsAny(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		CommandsAny: []string{"pwsh"},
	})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1 (only PowerShell registers pwsh)", resp.TotalCount)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty")
	}

	if resp.Results[0].PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", resp.Results[0].PackageIdentifier)
	}
}

func TestSearch_ByMinimumOSVersion(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	// "10.0.19041" matches only Chrome (10.0.19041.0), not PowerShell (10.0.17763.0).
	resp, err := svc.Search(context.Background(), &packages.FilterOptions{
		MinimumOSVersion: "10.0.19041",
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
		t.Errorf("PackageIdentifier = %q, want Google.Chrome", resp.Results[0].PackageIdentifier)
	}
}

// --- SearchInstallers ---

func TestSearchInstallers_NilFilter(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), nil)
	if err != nil {
		t.Fatalf("SearchInstallers(nil): %v", err)
	}

	// Two packages, one installer each.
	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", resp.TotalCount)
	}
}

func TestSearchInstallers_ByArchitecture(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		Architecture: "x64",
	})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2 (both packages have x64)", resp.TotalCount)
	}

	for _, r := range resp.Results {
		if r.Installer.Architecture != "x64" {
			t.Errorf("unexpected architecture %q in results", r.Installer.Architecture)
		}
	}
}

func TestSearchInstallers_ByInstallerType(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		InstallerType: "wix",
	})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1 (only PowerShell is wix)", resp.TotalCount)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty")
	}

	if resp.Results[0].PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want Microsoft.PowerShell", resp.Results[0].PackageIdentifier)
	}
}

func TestSearchInstallers_ByProductCode(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		ProductCode: "{D012DCD1-67EA-4627-938F-19FD677FC03A}",
	})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", resp.TotalCount)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty")
	}

	if resp.Results[0].PackageName != "PowerShell" {
		t.Errorf("PackageName = %q, want PowerShell", resp.Results[0].PackageName)
	}
}

func TestSearchInstallers_ByCommandsAny(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		CommandsAny: []string{"pwsh"},
	})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", resp.TotalCount)
	}
}

func TestSearchInstallers_ByPackageFamilyName(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		PackageFamilyName: "Google.Chrome_8wekyb3d8bbwe",
	})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1", resp.TotalCount)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty")
	}

	if resp.Results[0].PackageIdentifier != "Google.Chrome" {
		t.Errorf("PackageIdentifier = %q, want Google.Chrome", resp.Results[0].PackageIdentifier)
	}
}

func TestSearchInstallers_ByMinimumOSVersion(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		MinimumOSVersion: "10.0.17763",
	})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if resp.TotalCount != 1 {
		t.Errorf("TotalCount = %d, want 1 (only PowerShell has 10.0.17763.0)", resp.TotalCount)
	}
}

func TestSearchInstallers_Limit(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{Limit: 1})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if resp.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2 (total before limit)", resp.TotalCount)
	}

	if len(resp.Results) != 1 {
		t.Errorf("len(Results) = %d, want 1 (limited)", len(resp.Results))
	}
}

func TestSearchInstallers_Offset(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{Offset: 1})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Errorf("len(Results) = %d, want 1 (offset by 1 of 2)", len(resp.Results))
	}
}

func TestSearchInstallers_PopulatesDisplayFields(t *testing.T) {
	t.Parallel()

	svc := newTestPackages(t, newRichMockRepo())

	resp, err := svc.SearchInstallers(context.Background(), &packages.InstallerFilterOptions{
		InstallerType: "wix",
	})
	if err != nil {
		t.Fatalf("SearchInstallers: %v", err)
	}

	if len(resp.Results) == 0 {
		t.Fatal("Results is empty")
	}

	r := resp.Results[0]

	if r.PackageName == "" {
		t.Error("PackageName should be populated from defaultLocale manifest")
	}

	if r.Publisher == "" {
		t.Error("Publisher should be populated from defaultLocale manifest")
	}

	if r.PackageVersion == "" {
		t.Error("PackageVersion should be populated")
	}
}
