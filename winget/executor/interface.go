// Package executor abstracts git repository operations and filesystem access.
package executor

import (
	"context"

	"go.uber.org/zap"
)

// Executor abstracts the operations needed to access a cloned WinGet repository.
// It is satisfied by the go-git implementation in production and by the in-memory
// mock in tests.
type Executor interface {
	// EnsureRepo clones the repository if it is not already present on disk, or
	// verifies it is accessible if it already exists.
	EnsureRepo(ctx context.Context) error

	// Pull fetches the latest changes from origin and fast-forwards the local
	// working tree to match.
	Pull(ctx context.Context) error

	// WalkManifests calls fn exactly once for each version directory found under
	// the repository's manifests/ tree. The dirPath argument is a slash-separated
	// path relative to the repository root, e.g.
	// "manifests/m/Microsoft/PowerShell/7.5.0.0".
	// Walking stops and returns the first non-nil error returned by fn, or any
	// filesystem error encountered during traversal.
	WalkManifests(ctx context.Context, fn func(dirPath string) error) error

	// ReadFile returns the raw contents of the file at the given repository-relative
	// path (forward-slash separated).
	ReadFile(ctx context.Context, path string) ([]byte, error)

	// ListFiles returns the file names (not full paths) of all regular files in the
	// given repository-relative directory.
	ListFiles(ctx context.Context, dirPath string) ([]string, error)

	// RepoRoot returns the absolute local filesystem path to the cloned repository.
	RepoRoot() string

	// GetLogger returns the zap logger configured for this executor. Services and
	// index builders use this to obtain a consistent logger without holding their
	// own reference.
	GetLogger() *zap.Logger
}
