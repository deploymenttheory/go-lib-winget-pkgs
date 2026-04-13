// Package winget provides a client for programmatic access to WinGet package
// metadata from the microsoft/winget-pkgs repository.
package winget

import (
	"errors"

	"github.com/deploymenttheory/go-lib-winget-pkgs/winget/index"
)

// IsNotFound reports whether err (or any error in its chain) indicates that the
// requested package or version was not found in the index.
func IsNotFound(err error) bool {
	return errors.Is(err, index.ErrNotFound)
}

// IsCloneFailed reports whether err (or any error in its chain) indicates a
// failure during initial repository cloning.
func IsCloneFailed(err error) bool {
	return errors.Is(err, ErrCloneFailed)
}

// IsPullFailed reports whether err (or any error in its chain) indicates a
// failure during repository pull / refresh.
func IsPullFailed(err error) bool {
	return errors.Is(err, ErrPullFailed)
}

// IsIndexBuildFailed reports whether err (or any error in its chain) indicates a
// failure during in-memory index construction.
func IsIndexBuildFailed(err error) bool {
	return errors.Is(err, ErrIndexBuildFailed)
}

// Sentinel errors for client-level failures.
var (
	// ErrCloneFailed is returned when the initial repository clone fails.
	ErrCloneFailed = errors.New("repository clone failed")

	// ErrPullFailed is returned when a git pull / refresh operation fails.
	ErrPullFailed = errors.New("repository pull failed")

	// ErrIndexBuildFailed is returned when building the in-memory index fails.
	ErrIndexBuildFailed = errors.New("index build failed")
)
