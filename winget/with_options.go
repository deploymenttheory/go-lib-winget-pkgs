package winget

import (
	"time"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/config"
	"go.uber.org/zap"
)

// ClientOption is a functional option that modifies the client configuration
// before the client is initialised.
type ClientOption func(*config.Config)

// WithCacheDir sets the local directory used to clone and cache the WinGet
// repository. The directory is created if it does not exist.
func WithCacheDir(dir string) ClientOption {
	return func(c *config.Config) {
		c.CacheDir = dir
	}
}

// WithCloneDepth controls the shallow-clone depth. Use 0 for a full clone or
// 1 (the default) for a shallow clone that fetches only the latest commit.
func WithCloneDepth(depth int) ClientOption {
	return func(c *config.Config) {
		c.CloneDepth = depth
	}
}

// WithAutoRefresh configures the client to automatically pull the latest
// changes when the local clone is older than the given interval. Pass 0 to
// disable auto-refresh (the default).
func WithAutoRefresh(interval time.Duration) ClientOption {
	return func(c *config.Config) {
		c.AutoRefresh = interval
	}
}

// WithWorkerCount sets the number of goroutines used during index construction.
// Defaults to runtime.NumCPU().
func WithWorkerCount(n int) ClientOption {
	return func(c *config.Config) {
		c.WorkerCount = n
	}
}

// WithTimeout sets the maximum duration for clone or pull operations.
// Defaults to 5 minutes.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *config.Config) {
		c.Timeout = d
	}
}

// WithRepoURL overrides the Git URL of the WinGet packages repository.
// Defaults to "https://github.com/microsoft/winget-pkgs".
func WithRepoURL(url string) ClientOption {
	return func(c *config.Config) {
		c.RepoURL = url
	}
}

// WithLogger sets a custom zap logger. When not provided, a production logger
// writing JSON to stderr is created automatically. Pass zap.NewNop() to
// suppress all log output.
func WithLogger(logger *zap.Logger) ClientOption {
	return func(c *config.Config) {
		c.Logger = logger
	}
}
