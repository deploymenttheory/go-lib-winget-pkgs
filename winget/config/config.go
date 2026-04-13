// Package config defines configuration for the WinGet library client.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"
)

const (
	defaultRepoURL     = "https://github.com/microsoft/winget-pkgs"
	defaultCloneDepth  = 1
	defaultTimeout     = 5 * time.Minute
	defaultCacheDirName = "go-lib-winget-pkgs"
)

// Config holds all settings used to initialise the WinGet client.
type Config struct {
	// RepoURL is the Git URL of the WinGet packages repository.
	// Defaults to "https://github.com/microsoft/winget-pkgs".
	RepoURL string

	// CacheDir is the local directory where the repository will be cloned.
	// Defaults to ~/.cache/go-lib-winget-pkgs on Linux/macOS and
	// %LOCALAPPDATA%/go-lib-winget-pkgs on Windows.
	CacheDir string

	// CloneDepth controls the shallow-clone depth passed to git.
	// 0 means a full clone. 1 (the default) fetches only the latest commit,
	// which is much faster for large repositories.
	CloneDepth int

	// AutoRefresh, when non-zero, causes the client to pull the latest changes
	// from origin if the local clone is older than this duration.
	// When zero (the default), the client never auto-refreshes.
	AutoRefresh time.Duration

	// WorkerCount is the number of goroutines used when building the in-memory
	// index. Defaults to runtime.NumCPU().
	WorkerCount int

	// Timeout is the maximum time allowed for clone or pull operations.
	// Defaults to 5 minutes.
	Timeout time.Duration

	// Logger is the zap logger used throughout the library. When nil, a
	// production logger (JSON format, Info level, writing to stderr) is
	// created automatically. Pass zap.NewNop() to suppress all output.
	Logger *zap.Logger
}

// Validate checks the configuration for missing or invalid values and fills in
// defaults where appropriate.
func (c *Config) Validate() error {
	if c.RepoURL == "" {
		c.RepoURL = defaultRepoURL
	}

	if c.CacheDir == "" {
		dir, err := defaultCacheDir()
		if err != nil {
			return fmt.Errorf("determining default cache directory: %w", err)
		}

		c.CacheDir = dir
	}

	if c.CloneDepth < 0 {
		return fmt.Errorf("CloneDepth must be >= 0, got %d", c.CloneDepth)
	}

	if c.CloneDepth == 0 {
		c.CloneDepth = defaultCloneDepth
	}

	if c.WorkerCount <= 0 {
		c.WorkerCount = runtime.NumCPU()
	}

	if c.Timeout <= 0 {
		c.Timeout = defaultTimeout
	}

	if c.Logger == nil {
		logger, err := zap.NewProduction()
		if err != nil {
			return fmt.Errorf("creating default zap logger: %w", err)
		}

		c.Logger = logger
	}

	return nil
}

func defaultCacheDir() (string, error) {
	// Use $XDG_CACHE_HOME if set, otherwise ~/.cache on Unix and
	// %LOCALAPPDATA% on Windows.
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, defaultCacheDirName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("looking up user home directory: %w", err)
	}

	if runtime.GOOS == "windows" {
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			local = filepath.Join(home, "AppData", "Local")
		}

		return filepath.Join(local, defaultCacheDirName), nil
	}

	return filepath.Join(home, ".cache", defaultCacheDirName), nil
}
