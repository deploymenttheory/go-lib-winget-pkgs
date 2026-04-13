package index_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor/mock"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/index"
)

func newMockWithPowerShell() *mock.Executor {
	m := mock.New()

	// Add a realistic set of manifest files for two versions of PowerShell
	// and one version of VS Code.
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.yaml",
		versionManifest("Microsoft.PowerShell", "7.5.0.0"),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.installer.yaml",
		[]byte("PackageIdentifier: Microsoft.PowerShell\nManifestType: installer\n"),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.locale.en-US.yaml",
		[]byte("PackageIdentifier: Microsoft.PowerShell\nManifestType: defaultLocale\n"),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.4.0.0/Microsoft.PowerShell.yaml",
		versionManifest("Microsoft.PowerShell", "7.4.0.0"),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.4.0.0/Microsoft.PowerShell.installer.yaml",
		[]byte("PackageIdentifier: Microsoft.PowerShell\nManifestType: installer\n"),
	)
	m.AddFile(
		"manifests/m/Microsoft/PowerShell/7.4.0.0/Microsoft.PowerShell.locale.en-US.yaml",
		[]byte("PackageIdentifier: Microsoft.PowerShell\nManifestType: defaultLocale\n"),
	)
	m.AddFile(
		"manifests/m/Microsoft/VisualStudioCode/1.90.0/Microsoft.VisualStudioCode.yaml",
		versionManifest("Microsoft.VisualStudioCode", "1.90.0"),
	)
	m.AddFile(
		"manifests/m/Microsoft/VisualStudioCode/1.90.0/Microsoft.VisualStudioCode.installer.yaml",
		[]byte("PackageIdentifier: Microsoft.VisualStudioCode\nManifestType: installer\n"),
	)
	m.AddFile(
		"manifests/m/Microsoft/VisualStudioCode/1.90.0/Microsoft.VisualStudioCode.locale.en-US.yaml",
		[]byte("PackageIdentifier: Microsoft.VisualStudioCode\nManifestType: defaultLocale\n"),
	)
	m.AddFile(
		"manifests/g/Google/Chrome/125.0.6422.60/Google.Chrome.yaml",
		versionManifest("Google.Chrome", "125.0.6422.60"),
	)
	m.AddFile(
		"manifests/g/Google/Chrome/125.0.6422.60/Google.Chrome.installer.yaml",
		[]byte("PackageIdentifier: Google.Chrome\nManifestType: installer\n"),
	)
	m.AddFile(
		"manifests/g/Google/Chrome/125.0.6422.60/Google.Chrome.locale.en-US.yaml",
		[]byte("PackageIdentifier: Google.Chrome\nManifestType: defaultLocale\n"),
	)

	return m
}

func versionManifest(id, version string) []byte {
	return []byte("PackageIdentifier: " + id + "\nPackageVersion: " + version + "\nDefaultLocale: en-US\nManifestType: version\nManifestVersion: 1.9.0\n")
}

func TestBuild_Count(t *testing.T) {
	t.Parallel()

	m := newMockWithPowerShell()

	idx, err := index.Build(context.Background(), m, 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// 3 distinct packages: Microsoft.PowerShell, Microsoft.VisualStudioCode, Google.Chrome
	if idx.Count() != 3 {
		t.Errorf("Count = %d, want 3", idx.Count())
	}
}

func TestBuild_GetByID(t *testing.T) {
	t.Parallel()

	idx, err := index.Build(context.Background(), newMockWithPowerShell(), 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	entry, ok := idx.GetByID("Microsoft.PowerShell")
	if !ok {
		t.Fatal("GetByID(Microsoft.PowerShell): not found")
	}

	if entry.PackageIdentifier != "Microsoft.PowerShell" {
		t.Errorf("PackageIdentifier = %q, want %q", entry.PackageIdentifier, "Microsoft.PowerShell")
	}

	if entry.Publisher != "Microsoft" {
		t.Errorf("Publisher = %q, want %q", entry.Publisher, "Microsoft")
	}

	if len(entry.Versions) != 2 {
		t.Errorf("len(Versions) = %d, want 2", len(entry.Versions))
	}
}

func TestBuild_GetByID_CaseInsensitive(t *testing.T) {
	t.Parallel()

	idx, err := index.Build(context.Background(), newMockWithPowerShell(), 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	_, ok := idx.GetByID("microsoft.powershell")
	if !ok {
		t.Error("GetByID should be case-insensitive")
	}

	_, ok = idx.GetByID("MICROSOFT.POWERSHELL")
	if !ok {
		t.Error("GetByID should be case-insensitive (upper)")
	}
}

func TestBuild_GetByID_NotFound(t *testing.T) {
	t.Parallel()

	idx, err := index.Build(context.Background(), newMockWithPowerShell(), 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	_, ok := idx.GetByID("Nonexistent.Package")
	if ok {
		t.Error("GetByID should return false for nonexistent package")
	}
}

func TestBuild_VersionsSortedDescending(t *testing.T) {
	t.Parallel()

	idx, err := index.Build(context.Background(), newMockWithPowerShell(), 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	entry, ok := idx.GetByID("Microsoft.PowerShell")
	if !ok {
		t.Fatal("GetByID: not found")
	}

	if len(entry.Versions) < 2 {
		t.Fatalf("expected at least 2 versions, got %d", len(entry.Versions))
	}

	// 7.5.0.0 > 7.4.0.0 — newest should be first.
	if entry.Versions[0].Version != "7.5.0.0" {
		t.Errorf("Versions[0] = %q, want 7.5.0.0 (newest first)", entry.Versions[0].Version)
	}

	if entry.Versions[1].Version != "7.4.0.0" {
		t.Errorf("Versions[1] = %q, want 7.4.0.0", entry.Versions[1].Version)
	}
}

func TestBuild_LatestVersion(t *testing.T) {
	t.Parallel()

	idx, err := index.Build(context.Background(), newMockWithPowerShell(), 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	entry, _ := idx.GetByID("Microsoft.PowerShell")

	latest := entry.LatestVersion()
	if latest.Version != "7.5.0.0" {
		t.Errorf("LatestVersion = %q, want 7.5.0.0", latest.Version)
	}
}

func TestBuild_GetByPublisher(t *testing.T) {
	t.Parallel()

	idx, err := index.Build(context.Background(), newMockWithPowerShell(), 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	microsoftEntries := idx.GetByPublisher("Microsoft")
	if len(microsoftEntries) != 2 {
		t.Errorf("GetByPublisher(Microsoft) = %d entries, want 2", len(microsoftEntries))
	}

	googleEntries := idx.GetByPublisher("Google")
	if len(googleEntries) != 1 {
		t.Errorf("GetByPublisher(Google) = %d entries, want 1", len(googleEntries))
	}

	unknown := idx.GetByPublisher("Unknown")
	if len(unknown) != 0 {
		t.Errorf("GetByPublisher(Unknown) = %d entries, want 0", len(unknown))
	}
}

func TestBuild_All(t *testing.T) {
	t.Parallel()

	idx, err := index.Build(context.Background(), newMockWithPowerShell(), 1)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	all := idx.All()
	if len(all) != 3 {
		t.Errorf("All() = %d entries, want 3", len(all))
	}
}

func TestBuild_ErrNotFound(t *testing.T) {
	t.Parallel()

	if !errors.Is(index.ErrNotFound, index.ErrNotFound) {
		t.Error("ErrNotFound should satisfy errors.Is with itself")
	}
}

func TestBuild_ContextCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	m := newMockWithPowerShell()

	// Build may return an error due to cancellation, or may succeed if all
	// files are already in memory and the walk completes before the cancellation
	// is observed. Either outcome is acceptable — we just must not panic.
	_, _ = index.Build(ctx, m, 1)
}
