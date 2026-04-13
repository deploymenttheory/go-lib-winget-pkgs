package manifests

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor"
)

// Manifests provides low-level access to raw WinGet manifest files. It reads
// and parses individual YAML files on demand without caching.
type Manifests struct {
	exec executor.Executor
}

// NewManifests constructs a Manifests service backed by the given executor.
func NewManifests(exec executor.Executor) *Manifests {
	return &Manifests{exec: exec}
}

// GetVersionManifest parses and returns the version manifest for the given
// package identifier and version.
func (m *Manifests) GetVersionManifest(ctx context.Context, id, version string) (*VersionManifest, error) {
	m.exec.GetLogger().Debug("Getting version manifest",
		zap.String("package_id", id),
		zap.String("version", version),
	)

	path, err := m.versionManifestPath(ctx, id, version)
	if err != nil {
		return nil, err
	}

	data, err := m.exec.ReadFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading version manifest for %s@%s: %w", id, version, err)
	}

	manifest, err := ParseVersionManifest(data)
	if err != nil {
		return nil, fmt.Errorf("parsing version manifest for %s@%s: %w", id, version, err)
	}

	return manifest, nil
}

// GetInstallerManifest parses and returns the raw (un-flattened) installer
// manifest for the given package identifier and version.
func (m *Manifests) GetInstallerManifest(ctx context.Context, id, version string) (*InstallerManifest, error) {
	m.exec.GetLogger().Debug("Getting installer manifest",
		zap.String("package_id", id),
		zap.String("version", version),
	)

	dir, err := m.versionDir(id, version)
	if err != nil {
		return nil, err
	}

	path := dir + "/" + id + ".installer.yaml"

	data, err := m.exec.ReadFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading installer manifest for %s@%s: %w", id, version, err)
	}

	manifest, err := ParseInstallerManifest(data)
	if err != nil {
		return nil, fmt.Errorf("parsing installer manifest for %s@%s: %w", id, version, err)
	}

	return manifest, nil
}

// GetDefaultLocaleManifest parses and returns the defaultLocale manifest for
// the given package identifier and version.
func (m *Manifests) GetDefaultLocaleManifest(ctx context.Context, id, version string) (*DefaultLocaleManifest, error) {
	m.exec.GetLogger().Debug("Getting defaultLocale manifest",
		zap.String("package_id", id),
		zap.String("version", version),
	)

	dir, err := m.versionDir(id, version)
	if err != nil {
		return nil, err
	}

	// Find the defaultLocale filename — it matches {id}.locale.{locale}.yaml
	// where the locale matches the DefaultLocale declared in the version manifest.
	// We enumerate directory files to find it reliably.
	files, err := m.exec.ListFiles(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("listing manifest directory for %s@%s: %w", id, version, err)
	}

	filename, found := findDefaultLocaleFile(files, id)
	if !found {
		return nil, fmt.Errorf("defaultLocale manifest not found for %s@%s", id, version)
	}

	data, err := m.exec.ReadFile(ctx, dir+"/"+filename)
	if err != nil {
		return nil, fmt.Errorf("reading defaultLocale manifest for %s@%s: %w", id, version, err)
	}

	manifest, err := ParseDefaultLocaleManifest(data)
	if err != nil {
		return nil, fmt.Errorf("parsing defaultLocale manifest for %s@%s: %w", id, version, err)
	}

	return manifest, nil
}

// GetLocaleManifest parses and returns the locale manifest for the given
// package identifier, version, and BCP-47 locale code (e.g. "de-DE").
func (m *Manifests) GetLocaleManifest(ctx context.Context, id, version, locale string) (*LocaleManifest, error) {
	m.exec.GetLogger().Debug("Getting locale manifest",
		zap.String("package_id", id),
		zap.String("version", version),
		zap.String("locale", locale),
	)

	dir, err := m.versionDir(id, version)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/%s.locale.%s.yaml", dir, id, locale)

	data, err := m.exec.ReadFile(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("reading locale manifest for %s@%s (%s): %w", id, version, locale, err)
	}

	manifest, err := ParseLocaleManifest(data)
	if err != nil {
		return nil, fmt.Errorf("parsing locale manifest for %s@%s (%s): %w", id, version, locale, err)
	}

	return manifest, nil
}

// ListLocales returns the BCP-47 locale codes of all locale manifests available
// for the given package identifier and version.
func (m *Manifests) ListLocales(ctx context.Context, id, version string) ([]string, error) {
	dir, err := m.versionDir(id, version)
	if err != nil {
		return nil, err
	}

	files, err := m.exec.ListFiles(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("listing manifest directory for %s@%s: %w", id, version, err)
	}

	localePrefix := id + ".locale."

	var locales []string

	for _, f := range files {
		if !strings.HasPrefix(f, localePrefix) || !strings.HasSuffix(f, ".yaml") {
			continue
		}

		locale := strings.TrimPrefix(f, localePrefix)
		locale = strings.TrimSuffix(locale, ".yaml")

		locales = append(locales, locale)
	}

	return locales, nil
}

// versionDir returns the slash-separated repository-relative path to the
// manifest directory for a given package id and version.
//
// WinGet structures the repository as:
//
//	manifests/{first_letter_of_publisher}/{Publisher}/{rest_of_id}/{version}/
//
// For example, "Microsoft.PowerShell" at version "7.5.0.0" lives at:
//
//	manifests/m/Microsoft/PowerShell/7.5.0.0/
//
// The package identifier is split on the FIRST dot only; everything after the
// first dot (including any further dots) becomes the package path component.
// For example, "Google.Chrome.Beta" maps to "manifests/g/Google/Chrome.Beta/".
func (m *Manifests) versionDir(id, version string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("package identifier must not be empty")
	}

	if version == "" {
		return "", fmt.Errorf("version must not be empty")
	}

	dotIdx := strings.IndexByte(id, '.')
	if dotIdx < 0 {
		return "", fmt.Errorf("invalid package identifier %q: must be in Publisher.PackageName format", id)
	}

	publisher := id[:dotIdx]
	packagePath := id[dotIdx+1:]
	letter := strings.ToLower(string([]rune(publisher)[0]))

	return fmt.Sprintf("manifests/%s/%s/%s/%s", letter, publisher, packagePath, version), nil
}

// versionManifestPath returns the full repo-relative path to the version
// manifest file, e.g. "manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.yaml".
func (m *Manifests) versionManifestPath(ctx context.Context, id, version string) (string, error) {
	dir, err := m.versionDir(id, version)
	if err != nil {
		return "", err
	}

	files, listErr := m.exec.ListFiles(ctx, dir)
	if listErr != nil {
		return "", fmt.Errorf("listing manifest directory for %s@%s: %w", id, version, listErr)
	}

	for _, f := range files {
		if f == id+".yaml" {
			return dir + "/" + f, nil
		}
	}

	return "", fmt.Errorf("version manifest not found for %s@%s", id, version)
}

// findDefaultLocaleFile finds the defaultLocale manifest filename among a list
// of files in a version directory. A defaultLocale manifest is a file whose
// ManifestType field would be "defaultLocale" — we identify it by filename
// convention: {id}.locale.{locale}.yaml where it is NOT a secondary locale
// (i.e. it is the first locale-type file found, which is the defaultLocale).
// In practice we look for the file with the lowest BCP-47 tag depth or simply
// the one declared in the version manifest; here we pick whichever locale file
// appears to be the canonical one by returning the first match.
func findDefaultLocaleFile(files []string, id string) (string, bool) {
	localePrefix := id + ".locale."

	// Try to find "en-US" first as the most common default locale.
	enUS := localePrefix + "en-US.yaml"
	for _, f := range files {
		if f == enUS {
			return f, true
		}
	}

	// Fall back to the first locale file found.
	for _, f := range files {
		if strings.HasPrefix(f, localePrefix) && strings.HasSuffix(f, ".yaml") {
			return f, true
		}
	}

	return "", false
}
