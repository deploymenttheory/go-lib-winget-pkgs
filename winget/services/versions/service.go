package versions

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/index"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
)

// Versions provides access to the version history of WinGet packages.
type Versions struct {
	exec executor.Executor
	idx  *index.Index
	mfst *manifests.Manifests
}

// NewVersions constructs a Versions service.
func NewVersions(exec executor.Executor, idx *index.Index, mfst *manifests.Manifests) *Versions {
	return &Versions{
		exec: exec,
		idx:  idx,
		mfst: mfst,
	}
}

// ListByID returns all available versions for the given package identifier,
// sorted newest first.
func (v *Versions) ListByID(_ context.Context, id string) (*ListVersionsResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("package identifier must not be empty")
	}

	entry, ok := v.idx.GetByID(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", index.ErrNotFound, id)
	}

	versionStrings := make([]string, 0, len(entry.Versions))
	for _, ve := range entry.Versions {
		versionStrings = append(versionStrings, ve.Version)
	}

	return &ListVersionsResponse{
		PackageIdentifier: entry.PackageIdentifier,
		Versions:          versionStrings,
	}, nil
}

// GetByIDAndVersion returns the fully-resolved metadata for a specific version
// of the given package.
func (v *Versions) GetByIDAndVersion(ctx context.Context, id, version string) (*ResourceVersion, error) {
	if id == "" {
		return nil, fmt.Errorf("package identifier must not be empty")
	}

	if version == "" {
		return nil, fmt.Errorf("version must not be empty")
	}

	entry, ok := v.idx.GetByID(id)
	if !ok {
		return nil, fmt.Errorf("%w: %s", index.ErrNotFound, id)
	}

	// Confirm the requested version exists in the index.
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

	return v.loadVersion(ctx, entry.PackageIdentifier, version)
}

// loadVersion reads and merges the manifests for one specific version.
func (v *Versions) loadVersion(ctx context.Context, id, version string) (*ResourceVersion, error) {
	locale, err := v.mfst.GetDefaultLocaleManifest(ctx, id, version)
	if err != nil {
		return nil, fmt.Errorf("loading defaultLocale manifest for %s@%s: %w", id, version, err)
	}

	installer, err := v.mfst.GetInstallerManifest(ctx, id, version)
	if err != nil {
		return nil, fmt.Errorf("loading installer manifest for %s@%s: %w", id, version, err)
	}

	versionManifest, err := v.mfst.GetVersionManifest(ctx, id, version)
	if err != nil {
		return nil, fmt.Errorf("loading version manifest for %s@%s: %w", id, version, err)
	}

	localeCodes, err := v.mfst.ListLocales(ctx, id, version)
	if err != nil {
		return nil, fmt.Errorf("listing locales for %s@%s: %w", id, version, err)
	}

	additionalLocales := make([]manifests.LocaleManifest, 0)

	for _, code := range localeCodes {
		if code == locale.PackageLocale {
			continue // skip the default locale
		}

		lm, localeErr := v.mfst.GetLocaleManifest(ctx, id, version, code)
		if localeErr != nil {
			// Non-fatal: skip unreadable locale files.
			continue
		}

		additionalLocales = append(additionalLocales, *lm)
	}

	return &ResourceVersion{
		PackageIdentifier:   id,
		PackageVersion:      version,
		DefaultLocale:       versionManifest.DefaultLocale,
		Publisher:           locale.Publisher,
		PublisherURL:        locale.PublisherURL,
		PublisherSupportURL: locale.PublisherSupportURL,
		PrivacyURL:          locale.PrivacyURL,
		Author:              locale.Author,
		PackageName:         locale.PackageName,
		PackageURL:          locale.PackageURL,
		License:             locale.License,
		LicenseURL:          locale.LicenseURL,
		Copyright:           locale.Copyright,
		CopyrightURL:        locale.CopyrightURL,
		ShortDescription:    locale.ShortDescription,
		Description:         locale.Description,
		Moniker:             locale.Moniker,
		Tags:                locale.Tags,
		Agreements:          locale.Agreements,
		ReleaseNotes:        locale.ReleaseNotes,
		ReleaseNotesURL:     locale.ReleaseNotesURL,
		PurchaseURL:         locale.PurchaseURL,
		InstallationNotes:   locale.InstallationNotes,
		Documentations:      locale.Documentations,
		Icons:               locale.Icons,
		Installers:          manifests.FlattenInstallers(installer),
		Locales:             additionalLocales,
	}, nil
}
