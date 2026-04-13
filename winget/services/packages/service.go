package packages

import (
	"context"
	"fmt"
	"strings"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/index"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/versions"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/shared/models"
)

// Packages provides the primary query surface for WinGet packages.
type Packages struct {
	exec executor.Executor
	idx  *index.Index
	ver  *versions.Versions
	mfst *manifests.Manifests
}

// NewPackages constructs a Packages service.
func NewPackages(
	exec executor.Executor,
	idx *index.Index,
	ver *versions.Versions,
	mfst *manifests.Manifests,
) *Packages {
	return &Packages{
		exec: exec,
		idx:  idx,
		ver:  ver,
		mfst: mfst,
	}
}

// GetByID returns the latest version of the package with the given identifier
// (case-insensitive). ErrNotFound is returned if the package does not exist.
func (p *Packages) GetByID(ctx context.Context, id string) (*ResourcePackage, error) {
	if id == "" {
		return nil, fmt.Errorf("package identifier must not be empty")
	}

	entry, ok := p.idx.GetByID(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", index.ErrNotFound, id)
	}

	latest := entry.LatestVersion()

	return p.loadPackage(ctx, entry, latest.Version)
}

// GetByIDAndVersion returns the specific version of the package with the given
// identifier (case-insensitive).
func (p *Packages) GetByIDAndVersion(ctx context.Context, id, version string) (*ResourcePackage, error) {
	if id == "" {
		return nil, fmt.Errorf("package identifier must not be empty")
	}

	if version == "" {
		return nil, fmt.Errorf("version must not be empty")
	}

	entry, ok := p.idx.GetByID(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", index.ErrNotFound, id)
	}

	var found bool

	for _, ve := range entry.Versions {
		if ve.Version == version {
			found = true

			break
		}
	}

	if !found {
		return nil, fmt.Errorf("%w: %s@%s", index.ErrNotFound, id, version)
	}

	return p.loadPackage(ctx, entry, version)
}

// GetByName searches for packages whose PackageName (from the defaultLocale
// manifest) matches name case-insensitively. Multiple results are possible when
// packages from different publishers share the same display name.
func (p *Packages) GetByName(ctx context.Context, name string) ([]*ResourcePackage, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	nameLow := strings.ToLower(name)
	all := p.idx.All()

	var results []*ResourcePackage

	for _, entry := range all {
		latest := entry.LatestVersion()
		if latest.Version == "" {
			continue
		}

		locale, err := p.mfst.GetDefaultLocaleManifest(ctx, entry.PackageIdentifier, latest.Version)
		if err != nil {
			continue
		}

		if strings.ToLower(locale.PackageName) == nameLow {
			pkg, pkgErr := p.loadPackage(ctx, entry, latest.Version)
			if pkgErr != nil {
				continue
			}

			results = append(results, pkg)
		}
	}

	return results, nil
}

// ListAll returns the latest version of every package in the index.
func (p *Packages) ListAll(ctx context.Context) (*ListResponse, error) {
	return p.Search(ctx, &FilterOptions{})
}

// ListByPublisher returns the latest version of all packages whose publisher
// component (case-insensitive) matches the given string exactly.
func (p *Packages) ListByPublisher(ctx context.Context, publisher string) (*ListResponse, error) {
	if publisher == "" {
		return nil, fmt.Errorf("publisher must not be empty")
	}

	return p.Search(ctx, &FilterOptions{Publisher: publisher})
}

// Search filters packages using the provided FilterOptions. Publisher filtering
// is evaluated from the index (no file I/O). All other field filters require
// loading the defaultLocale and installer manifests.
func (p *Packages) Search(ctx context.Context, filter *FilterOptions) (*ListResponse, error) {
	if filter == nil {
		filter = &FilterOptions{}
	}

	// Determine the candidate set from the index.
	var candidates []*index.IndexEntry

	if filter.Publisher != "" {
		candidates = p.idx.GetByPublisher(filter.Publisher)
	} else {
		candidates = p.idx.All()
	}

	// Load and filter packages. Manifest reads are required for field-level filters.
	needsManifests := filter.NameContains != "" ||
		filter.MonikerContains != "" ||
		len(filter.TagsAny) > 0 ||
		filter.License != "" ||
		filter.InstallerType != "" ||
		filter.Architecture != "" ||
		filter.Scope != "" ||
		filter.ProductCode != "" ||
		filter.PackageFamilyName != "" ||
		len(filter.CommandsAny) > 0 ||
		filter.MinimumOSVersion != "" ||
		filter.HasMoniker

	var results []*ResourcePackage

	for _, entry := range candidates {
		latest := entry.LatestVersion()
		if latest.Version == "" {
			continue
		}

		if needsManifests {
			pkg, err := p.loadPackage(ctx, entry, latest.Version)
			if err != nil {
				continue
			}

			if !matchesFilter(pkg, filter) {
				continue
			}

			results = append(results, pkg)
		} else {
			// Publisher-only filter: build a lightweight ResourcePackage from the
			// index without reading any YAML files.
			results = append(results, &ResourcePackage{
				PackageIdentifier: entry.PackageIdentifier,
				LatestVersion:     latest.Version,
				Publisher:         entry.Publisher,
				AvailableVersions: versionStrings(entry.Versions),
			})
		}
	}

	// Apply offset and limit.
	total := len(results)

	if filter.Offset > 0 {
		if filter.Offset >= len(results) {
			results = nil
		} else {
			results = results[filter.Offset:]
		}
	}

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return &ListResponse{
		TotalCount: total,
		Results:    results,
	}, nil
}

// loadPackage reads and merges all manifests for a specific package version into
// a ResourcePackage.
func (p *Packages) loadPackage(ctx context.Context, entry *index.IndexEntry, version string) (*ResourcePackage, error) {
	rv, err := p.ver.GetByIDAndVersion(ctx, entry.PackageIdentifier, version)
	if err != nil {
		return nil, fmt.Errorf("loading version %s of %s: %w", version, entry.PackageIdentifier, err)
	}

	return &ResourcePackage{
		PackageIdentifier:   entry.PackageIdentifier,
		LatestVersion:       entry.LatestVersion().Version,
		AvailableVersions:   versionStrings(entry.Versions),
		DefaultLocale:       rv.DefaultLocale,
		Publisher:           rv.Publisher,
		PublisherURL:        rv.PublisherURL,
		PublisherSupportURL: rv.PublisherSupportURL,
		PrivacyURL:          rv.PrivacyURL,
		Author:              rv.Author,
		PackageName:         rv.PackageName,
		PackageURL:          rv.PackageURL,
		License:             rv.License,
		LicenseURL:          rv.LicenseURL,
		Copyright:           rv.Copyright,
		CopyrightURL:        rv.CopyrightURL,
		ShortDescription:    rv.ShortDescription,
		Description:         rv.Description,
		Moniker:             rv.Moniker,
		Tags:                rv.Tags,
		Agreements:          rv.Agreements,
		ReleaseNotes:        rv.ReleaseNotes,
		ReleaseNotesURL:     rv.ReleaseNotesURL,
		PurchaseURL:         rv.PurchaseURL,
		InstallationNotes:   rv.InstallationNotes,
		Documentations:      rv.Documentations,
		Icons:               rv.Icons,
		Installers:          rv.Installers,
		Locales:             rv.Locales,
	}, nil
}

// matchesFilter returns true if pkg satisfies all non-empty fields in filter.
func matchesFilter(pkg *ResourcePackage, filter *FilterOptions) bool {
	if filter.NameContains != "" {
		if !strings.Contains(strings.ToLower(pkg.PackageName), strings.ToLower(filter.NameContains)) {
			return false
		}
	}

	if filter.MonikerContains != "" {
		if !strings.Contains(strings.ToLower(pkg.Moniker), strings.ToLower(filter.MonikerContains)) {
			return false
		}
	}

	if len(filter.TagsAny) > 0 {
		if !hasAnyTag(pkg.Tags, filter.TagsAny) {
			return false
		}
	}

	if filter.License != "" {
		if !strings.Contains(strings.ToLower(pkg.License), strings.ToLower(filter.License)) {
			return false
		}
	}

	if filter.HasMoniker && pkg.Moniker == "" {
		return false
	}

	if filter.InstallerType != "" || filter.Architecture != "" || filter.Scope != "" ||
		filter.ProductCode != "" || filter.PackageFamilyName != "" ||
		len(filter.CommandsAny) > 0 || filter.MinimumOSVersion != "" {
		if !hasMatchingInstaller(pkg.Installers, filter) {
			return false
		}
	}

	return true
}

// hasAnyTag returns true if tags contains at least one element from want
// (case-insensitive).
func hasAnyTag(tags []string, want []string) bool {
	for _, w := range want {
		wLow := strings.ToLower(w)

		for _, t := range tags {
			if strings.ToLower(t) == wLow {
				return true
			}
		}
	}

	return false
}

// hasMatchingInstaller returns true if at least one effective installer satisfies
// all non-empty installer-specific fields in filter.
func hasMatchingInstaller(installers []models.EffectiveInstaller, filter *FilterOptions) bool {
	for _, inst := range installers {
		if filter.InstallerType != "" &&
			!strings.EqualFold(inst.InstallerType, filter.InstallerType) {
			continue
		}

		if filter.Architecture != "" &&
			!strings.EqualFold(inst.Architecture, filter.Architecture) {
			continue
		}

		if filter.Scope != "" &&
			!strings.EqualFold(inst.Scope, filter.Scope) {
			continue
		}

		if filter.ProductCode != "" &&
			!strings.EqualFold(inst.ProductCode, filter.ProductCode) {
			continue
		}

		if filter.PackageFamilyName != "" &&
			!strings.EqualFold(inst.PackageFamilyName, filter.PackageFamilyName) {
			continue
		}

		if len(filter.CommandsAny) > 0 && !hasAnyCommand(inst.Commands, filter.CommandsAny) {
			continue
		}

		if filter.MinimumOSVersion != "" &&
			!strings.Contains(inst.MinimumOSVersion, filter.MinimumOSVersion) {
			continue
		}

		return true
	}

	return false
}

// GetByIDAndArchitecture returns the latest version of the package with the
// given identifier (case-insensitive), with Installers filtered to only those
// matching arch. An error is returned if no installer for that architecture exists.
func (p *Packages) GetByIDAndArchitecture(ctx context.Context, id, arch string) (*ResourcePackage, error) {
	if id == "" {
		return nil, fmt.Errorf("package identifier must not be empty")
	}

	if arch == "" {
		return nil, fmt.Errorf("architecture must not be empty")
	}

	entry, ok := p.idx.GetByID(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", index.ErrNotFound, id)
	}

	pkg, err := p.loadPackage(ctx, entry, entry.LatestVersion().Version)
	if err != nil {
		return nil, err
	}

	filtered := filterInstallersByArch(pkg.Installers, arch)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("%w: no %s installer found for %s", index.ErrNotFound, arch, id)
	}

	pkg.Installers = filtered

	return pkg, nil
}

// GetByNameAndArchitecture returns packages whose PackageName matches name
// (case-insensitive), with each result's Installers filtered to only those
// matching arch. Packages that have no installer for the requested architecture
// are excluded from the results.
func (p *Packages) GetByNameAndArchitecture(ctx context.Context, name, arch string) ([]*ResourcePackage, error) {
	if name == "" {
		return nil, fmt.Errorf("name must not be empty")
	}

	if arch == "" {
		return nil, fmt.Errorf("architecture must not be empty")
	}

	nameLow := strings.ToLower(name)
	all := p.idx.All()

	var results []*ResourcePackage

	for _, entry := range all {
		latest := entry.LatestVersion()
		if latest.Version == "" {
			continue
		}

		locale, err := p.mfst.GetDefaultLocaleManifest(ctx, entry.PackageIdentifier, latest.Version)
		if err != nil {
			continue
		}

		if strings.ToLower(locale.PackageName) != nameLow {
			continue
		}

		pkg, pkgErr := p.loadPackage(ctx, entry, latest.Version)
		if pkgErr != nil {
			continue
		}

		filtered := filterInstallersByArch(pkg.Installers, arch)
		if len(filtered) == 0 {
			continue // package exists but has no installer for this architecture
		}

		pkg.Installers = filtered
		results = append(results, pkg)
	}

	return results, nil
}

// SearchInstallers returns a flat list of individual resolved installers that
// match the given InstallerFilterOptions. Unlike Search (which returns one
// ResourcePackage per package), SearchInstallers yields one InstallerResult per
// matching installer entry, making it suitable for queries such as "every x64
// MSI installer that registers the 'pwsh' command".
func (p *Packages) SearchInstallers(ctx context.Context, filter *InstallerFilterOptions) (*InstallerSearchResponse, error) {
	if filter == nil {
		filter = &InstallerFilterOptions{}
	}

	var candidates []*index.IndexEntry

	if filter.Publisher != "" {
		candidates = p.idx.GetByPublisher(filter.Publisher)
	} else {
		candidates = p.idx.All()
	}

	var results []*InstallerResult

	for _, entry := range candidates {
		latest := entry.LatestVersion()
		if latest.Version == "" {
			continue
		}

		installer, err := p.mfst.GetInstallerManifest(ctx, entry.PackageIdentifier, latest.Version)
		if err != nil {
			continue
		}

		effective := manifests.FlattenInstallers(installer)

		// Load locale for display fields; non-fatal if absent.
		locale, locErr := p.mfst.GetDefaultLocaleManifest(ctx, entry.PackageIdentifier, latest.Version)

		for _, inst := range effective {
			if !matchesInstallerFilter(inst, filter) {
				continue
			}

			result := &InstallerResult{
				PackageIdentifier: entry.PackageIdentifier,
				PackageVersion:    latest.Version,
				Installer:         inst,
			}

			if locErr == nil {
				result.PackageName = locale.PackageName
				result.Publisher = locale.Publisher
			}

			results = append(results, result)
		}
	}

	total := len(results)

	if filter.Offset > 0 {
		if filter.Offset >= len(results) {
			results = nil
		} else {
			results = results[filter.Offset:]
		}
	}

	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return &InstallerSearchResponse{
		TotalCount: total,
		Results:    results,
	}, nil
}

// filterInstallersByArch returns the subset of installers whose Architecture
// matches arch (case-insensitive).
func filterInstallersByArch(installers []models.EffectiveInstaller, arch string) []models.EffectiveInstaller {
	var out []models.EffectiveInstaller

	for _, inst := range installers {
		if strings.EqualFold(inst.Architecture, arch) {
			out = append(out, inst)
		}
	}

	return out
}

// matchesInstallerFilter returns true if inst satisfies all non-empty fields in filter.
func matchesInstallerFilter(inst models.EffectiveInstaller, filter *InstallerFilterOptions) bool {
	if filter.Architecture != "" && !strings.EqualFold(inst.Architecture, filter.Architecture) {
		return false
	}

	if filter.InstallerType != "" && !strings.EqualFold(inst.InstallerType, filter.InstallerType) {
		return false
	}

	if filter.Scope != "" && !strings.EqualFold(inst.Scope, filter.Scope) {
		return false
	}

	if filter.ProductCode != "" && !strings.EqualFold(inst.ProductCode, filter.ProductCode) {
		return false
	}

	if filter.PackageFamilyName != "" && !strings.EqualFold(inst.PackageFamilyName, filter.PackageFamilyName) {
		return false
	}

	if len(filter.CommandsAny) > 0 && !hasAnyCommand(inst.Commands, filter.CommandsAny) {
		return false
	}

	if filter.MinimumOSVersion != "" && !strings.Contains(inst.MinimumOSVersion, filter.MinimumOSVersion) {
		return false
	}

	return true
}

// hasAnyCommand returns true if commands contains at least one element from
// want (case-insensitive).
func hasAnyCommand(commands []string, want []string) bool {
	for _, w := range want {
		wLow := strings.ToLower(w)

		for _, c := range commands {
			if strings.ToLower(c) == wLow {
				return true
			}
		}
	}

	return false
}

// versionStrings extracts the version strings from a slice of VersionEntry.
func versionStrings(versions []index.VersionEntry) []string {
	out := make([]string, len(versions))

	for i, v := range versions {
		out[i] = v.Version
	}

	return out
}
