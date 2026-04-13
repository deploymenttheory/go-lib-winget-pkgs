// Package mock provides an in-memory Executor implementation for unit tests.
package mock

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// Executor is an in-memory implementation of executor.Executor suitable for
// unit tests. Populate Files with repo-relative paths mapped to raw byte
// content before calling any methods.
type Executor struct {
	// Files maps repo-relative slash-separated file paths to their contents.
	// e.g. "manifests/m/Microsoft/PowerShell/7.5.0.0/Microsoft.PowerShell.yaml"
	Files map[string][]byte

	// RepoRootPath is the value returned by RepoRoot.
	RepoRootPath string
}

// New constructs a mock Executor with an empty file store.
func New() *Executor {
	return &Executor{
		Files:        make(map[string][]byte),
		RepoRootPath: "/mock-repo",
	}
}

// AddFile registers a file at the given repo-relative path with the provided
// content.
func (m *Executor) AddFile(path string, content []byte) {
	m.Files[path] = content
}

// EnsureRepo is a no-op in the mock — no network or disk access required.
func (m *Executor) EnsureRepo(_ context.Context) error {
	return nil
}

// Pull is a no-op in the mock.
func (m *Executor) Pull(_ context.Context) error {
	return nil
}

// WalkManifests iterates over Files and calls fn once per unique directory
// that contains YAML files under manifests/.
func (m *Executor) WalkManifests(_ context.Context, fn func(dirPath string) error) error {
	visited := make(map[string]struct{})

	for path := range m.Files {
		if !strings.HasPrefix(path, "manifests/") {
			continue
		}

		if !strings.HasSuffix(path, ".yaml") {
			continue
		}

		slash := strings.LastIndex(path, "/")
		if slash < 0 {
			continue
		}

		dir := path[:slash]
		if _, seen := visited[dir]; seen {
			continue
		}

		visited[dir] = struct{}{}

		if err := fn(dir); err != nil {
			return err
		}
	}

	return nil
}

// ReadFile returns the content registered for the given path.
func (m *Executor) ReadFile(_ context.Context, path string) ([]byte, error) {
	data, ok := m.Files[path]
	if !ok {
		return nil, fmt.Errorf("file not found in mock: %s", path)
	}

	return data, nil
}

// ListFiles returns the names of files registered under dirPath.
func (m *Executor) ListFiles(_ context.Context, dirPath string) ([]string, error) {
	prefix := dirPath + "/"
	seen := make(map[string]struct{})

	for path := range m.Files {
		if !strings.HasPrefix(path, prefix) {
			continue
		}

		rest := path[len(prefix):]
		if strings.Contains(rest, "/") {
			continue // sub-directory, not a direct child
		}

		seen[rest] = struct{}{}
	}

	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}

	return names, nil
}

// RepoRoot returns the mock repository root path.
func (m *Executor) RepoRoot() string {
	return m.RepoRootPath
}

// GetLogger returns a no-op logger suitable for tests.
func (m *Executor) GetLogger() *zap.Logger {
	return zap.NewNop()
}
