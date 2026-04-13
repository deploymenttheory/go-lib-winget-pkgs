// Package index provides the in-memory index of WinGet packages.
package index

// VersionEntry records a single available version of a package and the
// repository-relative path to its manifest directory.
type VersionEntry struct {
	// Version is the package version string as it appears in the manifest, e.g.
	// "7.5.0.0" or "3.1.23.0".
	Version string

	// DirPath is the slash-separated repository-relative path to the version
	// directory, e.g. "manifests/m/Microsoft/PowerShell/7.5.0.0".
	DirPath string
}

// IndexEntry holds all known versions of a single package.
type IndexEntry struct {
	// PackageIdentifier is the canonical dot-separated package ID as declared in
	// the manifest files, e.g. "Microsoft.PowerShell".
	PackageIdentifier string

	// Publisher is the first component of PackageIdentifier, e.g. "Microsoft".
	Publisher string

	// Versions lists every available version, sorted descending (newest first).
	Versions []VersionEntry
}

// LatestVersion returns the newest version entry, or the zero value if there
// are no versions.
func (e *IndexEntry) LatestVersion() VersionEntry {
	if len(e.Versions) == 0 {
		return VersionEntry{}
	}

	return e.Versions[0]
}
