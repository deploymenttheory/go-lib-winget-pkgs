package index

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor"
)

var (
	// ErrNotFound is returned when a package is not present in the index.
	ErrNotFound = errors.New("package not found in index")
)

// Index is a thread-safe in-memory index of all WinGet packages discovered in
// the repository. It is built once at startup and rebuilt when Refresh is called.
type Index struct {
	mu          sync.RWMutex
	byID        map[string]*IndexEntry // lowercase(PackageIdentifier) → entry
	byPublisher map[string][]*IndexEntry // lowercase(Publisher) → entries
	all         []*IndexEntry
}

// Build walks the repository's manifests/ tree and constructs a new Index.
// It discovers packages by inspecting directory paths and YAML filenames —
// no manifest YAML is parsed at this stage.
func Build(ctx context.Context, exec executor.Executor, _ int) (*Index, error) {
	logger := exec.GetLogger()

	logger.Info("Building WinGet package index")

	start := time.Now()

	idx := &Index{
		byID:        make(map[string]*IndexEntry),
		byPublisher: make(map[string][]*IndexEntry),
	}

	// Track entries being built (keyed by package ID).
	building := make(map[string]*IndexEntry)

	err := exec.WalkManifests(ctx, func(dirPath string) error {
		// dirPath is e.g. "manifests/m/Microsoft/PowerShell/7.5.0.0"
		files, listErr := exec.ListFiles(ctx, dirPath)
		if listErr != nil {
			logger.Warn("Skipping unreadable manifest directory",
				zap.String("dir_path", dirPath),
				zap.Error(listErr),
			)
			// Skip unreadable directories rather than aborting the whole walk.
			return nil
		}

		pkgID, version := extractIDAndVersion(dirPath, files)
		if pkgID == "" || version == "" {
			return nil
		}

		idLow := strings.ToLower(pkgID)

		entry, exists := building[idLow]
		if !exists {
			publisher := publisherOf(pkgID)
			entry = &IndexEntry{
				PackageIdentifier: pkgID,
				Publisher:         publisher,
			}

			building[idLow] = entry
		}

		entry.Versions = append(entry.Versions, VersionEntry{
			Version: version,
			DirPath: dirPath,
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking manifests: %w", err)
	}

	// Sort versions and populate lookup maps.
	for idLow, entry := range building {
		sortVersionsDesc(entry.Versions)
		idx.all = append(idx.all, entry)
		idx.byID[idLow] = entry

		pubLow := strings.ToLower(entry.Publisher)
		idx.byPublisher[pubLow] = append(idx.byPublisher[pubLow], entry)
	}

	logger.Info("WinGet package index built",
		zap.Int("package_count", len(idx.byID)),
		zap.Duration("duration", time.Since(start)),
	)

	return idx, nil
}

// GetByID returns the index entry for the given package identifier (case-insensitive).
func (idx *Index) GetByID(id string) (*IndexEntry, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	entry, ok := idx.byID[strings.ToLower(id)]

	return entry, ok
}

// GetByPublisher returns all index entries whose publisher matches the given
// string (case-insensitive, exact match on the publisher component).
func (idx *Index) GetByPublisher(publisher string) []*IndexEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.byPublisher[strings.ToLower(publisher)]
}

// All returns every index entry in an unspecified order.
func (idx *Index) All() []*IndexEntry {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.all
}

// Count returns the total number of packages in the index.
func (idx *Index) Count() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return len(idx.byID)
}

// extractIDAndVersion infers the PackageIdentifier and version from a version
// directory path and its YAML filenames. The version manifest is identified by
// its filename: {PackageIdentifier}.yaml (no ".installer." or ".locale." segment).
func extractIDAndVersion(dirPath string, files []string) (pkgID, version string) {
	// The version is the last path component.
	slash := strings.LastIndex(dirPath, "/")
	if slash < 0 {
		return "", ""
	}

	version = dirPath[slash+1:]

	// Find the version manifest filename: ends in .yaml but not .installer.yaml
	// or .locale.*.yaml.
	for _, f := range files {
		if !strings.HasSuffix(f, ".yaml") {
			continue
		}

		if strings.Contains(f, ".installer.") || strings.Contains(f, ".locale.") {
			continue
		}

		// Filename is "{PackageIdentifier}.yaml"
		pkgID = strings.TrimSuffix(f, ".yaml")
		return pkgID, version
	}

	return "", ""
}

// publisherOf returns the first dot-separated component of a package identifier.
func publisherOf(id string) string {
	dot := strings.IndexByte(id, '.')
	if dot < 0 {
		return id
	}

	return id[:dot]
}

// sortVersionsDesc sorts version entries with the highest version first using a
// simple numeric component comparison.
func sortVersionsDesc(versions []VersionEntry) {
	// Insertion sort — entry counts per package are small (typically < 20).
	for i := 1; i < len(versions); i++ {
		key := versions[i]

		j := i - 1
		for j >= 0 && compareVersions(versions[j].Version, key.Version) < 0 {
			versions[j+1] = versions[j]
			j--
		}

		versions[j+1] = key
	}
}

// compareVersions compares two version strings by their numeric components.
// Returns >0 if a > b, <0 if a < b, 0 if equal.
func compareVersions(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := range maxLen {
		aNum := parseVersionSegment(aParts, i)
		bNum := parseVersionSegment(bParts, i)

		if aNum != bNum {
			return aNum - bNum
		}
	}

	return 0
}

func parseVersionSegment(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}

	n := 0
	for _, ch := range parts[i] {
		if ch < '0' || ch > '9' {
			break
		}

		n = n*10 + int(ch-'0')
	}

	return n
}
