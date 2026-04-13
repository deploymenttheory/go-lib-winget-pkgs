package winget

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/executor"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/index"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/manifests"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/packages"
	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/services/versions"
)

// Client is the top-level entry point for the WinGet library. It manages the
// local repository clone, the in-memory index, and all service objects.
//
// Construct one with NewClient; call Refresh to pull the latest package data.
type Client struct {
	exec executor.Executor
	idx  *index.Index
	cfg  *config.Config

	// Packages exposes GetByID, GetByIDAndVersion, GetByName, ListAll,
	// ListByPublisher, and Search.
	Packages *packages.Packages

	// Versions exposes ListByID and GetByIDAndVersion.
	Versions *versions.Versions

	// Manifests exposes low-level access to individual manifest files.
	Manifests *manifests.Manifests
}

// NewClient creates and initialises a WinGet client. It clones the repository
// (if not already present) and builds the in-memory package index before
// returning.
//
// Pass functional options to override configuration defaults:
//
//	client, err := winget.NewClient(&config.Config{},
//	    winget.WithCacheDir("~/.cache/winget"),
//	    winget.WithAutoRefresh(24*time.Hour),
//	)
func NewClient(cfg *config.Config, opts ...ClientOption) (*Client, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	exec := executor.NewGoGitExecutor(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := exec.EnsureRepo(ctx); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCloneFailed, err)
	}

	client, err := buildClient(cfg, exec)
	if err != nil {
		return nil, err
	}

	exec.GetLogger().Info("WinGet client created",
		zap.String("repo_url", cfg.RepoURL),
		zap.String("cache_dir", cfg.CacheDir),
		zap.Int("package_count", client.idx.Count()),
	)

	// Auto-refresh if the interval is configured.
	if cfg.AutoRefresh > 0 {
		go client.autoRefreshLoop(cfg.AutoRefresh)
	}

	return client, nil
}

// NewClientWithExecutor creates a client using a custom Executor implementation.
// This is primarily used in tests where a mock executor is injected.
func NewClientWithExecutor(cfg *config.Config, exec executor.Executor) (*Client, error) {
	if cfg == nil {
		cfg = &config.Config{}
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return buildClient(cfg, exec)
}

// Refresh pulls the latest changes from origin and rebuilds the in-memory index.
func (c *Client) Refresh(ctx context.Context) error {
	if err := c.exec.Pull(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrPullFailed, err)
	}

	idx, err := index.Build(ctx, c.exec, c.cfg.WorkerCount)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrIndexBuildFailed, err)
	}

	// Wire up services with the new index.
	mfst := manifests.NewManifests(c.exec)
	ver := versions.NewVersions(c.exec, idx, mfst)
	pkgs := packages.NewPackages(c.exec, idx, ver, mfst)

	c.idx = idx
	c.Manifests = mfst
	c.Versions = ver
	c.Packages = pkgs

	return nil
}

// Close is a no-op at present but is provided for forwards-compatibility. It
// should always be deferred after a successful NewClient call.
func (c *Client) Close() error {
	return nil
}

// buildClient constructs the index and service objects from an already-verified
// executor.
func buildClient(cfg *config.Config, exec executor.Executor) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	idx, err := index.Build(ctx, exec, cfg.WorkerCount)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrIndexBuildFailed, err)
	}

	mfst := manifests.NewManifests(exec)
	ver := versions.NewVersions(exec, idx, mfst)
	pkgs := packages.NewPackages(exec, idx, ver, mfst)

	return &Client{
		exec:      exec,
		idx:       idx,
		cfg:       cfg,
		Packages:  pkgs,
		Versions:  ver,
		Manifests: mfst,
	}, nil
}

// autoRefreshLoop runs in a goroutine and refreshes the index at the configured
// interval until the process exits.
func (c *Client) autoRefreshLoop(interval time.Duration) {
	logger := c.exec.GetLogger()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), c.cfg.Timeout)

		if err := c.Refresh(ctx); err != nil {
			logger.Warn("WinGet auto-refresh failed",
				zap.Duration("interval", interval),
				zap.Error(err),
			)
		}

		cancel()
	}
}
