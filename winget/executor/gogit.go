package executor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gogithttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"go.uber.org/zap"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
)

// GoGitExecutor implements Executor using a local clone of the repository
// managed by go-git.
type GoGitExecutor struct {
	cfg      *config.Config
	repoRoot string
	logger   *zap.Logger
}

// NewGoGitExecutor constructs a GoGitExecutor. It does not perform any I/O;
// call EnsureRepo to clone or verify the repository.
func NewGoGitExecutor(cfg *config.Config) *GoGitExecutor {
	return &GoGitExecutor{
		cfg:      cfg,
		repoRoot: filepath.Join(cfg.CacheDir, "winget-pkgs"),
		logger:   cfg.Logger,
	}
}

// GetLogger returns the zap logger configured for this executor.
func (g *GoGitExecutor) GetLogger() *zap.Logger {
	return g.logger
}

// EnsureRepo clones the repository if the cache directory does not yet contain
// it. If the directory already exists it is opened and verified to be a valid
// git repository.
func (g *GoGitExecutor) EnsureRepo(ctx context.Context) error {
	if _, err := os.Stat(filepath.Join(g.repoRoot, ".git")); err == nil {
		g.logger.Debug("WinGet repository already cloned, verifying",
			zap.String("repo_root", g.repoRoot),
		)

		if _, openErr := git.PlainOpen(g.repoRoot); openErr != nil {
			return fmt.Errorf("opening existing repository at %s: %w", g.repoRoot, openErr)
		}

		g.logger.Info("WinGet repository verified",
			zap.String("repo_root", g.repoRoot),
		)

		return nil
	}

	if err := os.MkdirAll(g.repoRoot, 0o750); err != nil {
		return fmt.Errorf("creating cache directory %s: %w", g.repoRoot, err)
	}

	g.logger.Info("Cloning WinGet repository",
		zap.String("repo_url", g.cfg.RepoURL),
		zap.String("cache_dir", g.repoRoot),
		zap.Int("clone_depth", g.cfg.CloneDepth),
	)

	start := time.Now()

	cloneCtx, cancel := context.WithTimeout(ctx, g.cfg.Timeout)
	defer cancel()

	opts := &git.CloneOptions{
		URL:      g.cfg.RepoURL,
		Progress: io.Discard,
		Auth:     publicHTTPAuth(),
	}

	if g.cfg.CloneDepth > 0 {
		opts.Depth = g.cfg.CloneDepth
	}

	if _, err := git.PlainCloneContext(cloneCtx, g.repoRoot, false, opts); err != nil {
		return fmt.Errorf("cloning %s into %s: %w", g.cfg.RepoURL, g.repoRoot, err)
	}

	g.logger.Info("WinGet repository cloned successfully",
		zap.String("repo_url", g.cfg.RepoURL),
		zap.String("cache_dir", g.repoRoot),
		zap.Duration("duration", time.Since(start)),
	)

	return nil
}

// Pull fetches the latest changes from origin and fast-forwards the working tree.
func (g *GoGitExecutor) Pull(ctx context.Context) error {
	g.logger.Info("Pulling latest changes from WinGet repository",
		zap.String("repo_root", g.repoRoot),
	)

	start := time.Now()

	repo, err := git.PlainOpen(g.repoRoot)
	if err != nil {
		return fmt.Errorf("opening repository at %s: %w", g.repoRoot, err)
	}

	wt, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("getting worktree: %w", err)
	}

	pullCtx, cancel := context.WithTimeout(ctx, g.cfg.Timeout)
	defer cancel()

	err = wt.PullContext(pullCtx, &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName("master"),
		Progress:      io.Discard,
		Auth:          publicHTTPAuth(),
		Force:         true,
	})

	switch {
	case errors.Is(err, git.NoErrAlreadyUpToDate):
		g.logger.Info("WinGet repository already up to date",
			zap.String("repo_root", g.repoRoot),
			zap.Duration("duration", time.Since(start)),
		)
	case err != nil:
		return fmt.Errorf("pulling latest changes: %w", err)
	default:
		g.logger.Info("WinGet repository updated successfully",
			zap.String("repo_root", g.repoRoot),
			zap.Duration("duration", time.Since(start)),
		)
	}

	return nil
}

// WalkManifests walks the manifests/ tree and calls fn exactly once for each
// version directory (a leaf directory containing YAML files). dirPath is
// slash-separated and relative to the repository root.
func (g *GoGitExecutor) WalkManifests(ctx context.Context, fn func(dirPath string) error) error {
	manifestsRoot := filepath.Join(g.repoRoot, "manifests")

	g.logger.Debug("Walking WinGet manifests tree",
		zap.String("manifests_root", manifestsRoot),
	)

	visited := make(map[string]struct{})

	return filepath.WalkDir(manifestsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking %s: %w", path, err)
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("walk cancelled: %w", ctx.Err())
		default:
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".yaml") {
			return nil
		}

		dir := filepath.Dir(path)
		if _, seen := visited[dir]; seen {
			return nil
		}

		visited[dir] = struct{}{}

		relDir, relErr := filepath.Rel(g.repoRoot, dir)
		if relErr != nil {
			return fmt.Errorf("computing relative path for %s: %w", dir, relErr)
		}

		// Normalise to forward slashes for cross-platform consistency.
		return fn(filepath.ToSlash(relDir))
	})
}

// ReadFile returns the raw bytes of the file at the given repository-relative
// slash-separated path.
func (g *GoGitExecutor) ReadFile(_ context.Context, path string) ([]byte, error) {
	g.logger.Debug("Reading manifest file",
		zap.String("file_path", path),
	)

	abs := filepath.Join(g.repoRoot, filepath.FromSlash(path))

	data, err := os.ReadFile(abs) //nolint:gosec // path is derived from walked manifests, not user input
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}

	return data, nil
}

// ListFiles returns the names of all regular files in the given
// repository-relative slash-separated directory path.
func (g *GoGitExecutor) ListFiles(_ context.Context, dirPath string) ([]string, error) {
	g.logger.Debug("Listing files in manifest directory",
		zap.String("dir_path", dirPath),
	)

	abs := filepath.Join(g.repoRoot, filepath.FromSlash(dirPath))

	entries, err := os.ReadDir(abs) //nolint:gosec // path is derived from walked manifests
	if err != nil {
		return nil, fmt.Errorf("listing files in %s: %w", dirPath, err)
	}

	names := make([]string, 0, len(entries))

	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}

	return names, nil
}

// RepoRoot returns the absolute path of the local repository clone.
func (g *GoGitExecutor) RepoRoot() string {
	return g.repoRoot
}

// publicHTTPAuth returns HTTP auth that works for public repositories (empty
// credentials). go-git still requires an auth object even for anonymous access
// in some transport implementations.
func publicHTTPAuth() *gogithttp.BasicAuth {
	return &gogithttp.BasicAuth{}
}
