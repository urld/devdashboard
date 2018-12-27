// Original work: Copyright 2017 The Go Authors.
// Modified work: Copyright 2018 David Url.
// Use of this source code is governed by a BSD-style
// license that can be found in the go.LICENSE file.

// Package devdashdata loads the project's corpus into memory to allow easy
// analysis without worrying about APIs and their pagination, quotas, and
// other nuisances and limitations.
package devdashdata

import (
	"context"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/urld/devdashboard"
)

// Get returns the project's corpus, containing all Git commits,
// issue tracker activity and metadata since the beginning of the project.
//
// Use Corpus.Update to keep the corpus up-to-date. If you do this, you must
// hold the read lock if reading and updating concurrently.
//
// The initial call to Get will download all the data into a directory
// "devdashboard" under your operating system's user cache directory.
// Subsequent calls will only download what's changed since the previous call.
//
// For daemons, use Corpus.Update to incrementally update an
// already-loaded Corpus.
//
// See https://godoc.org/github.com/urld/devdashboard#Corpus for how to walk
// the data structure.
func Get(ctx context.Context, targetDir string) (*devdashboard.Corpus, error) {
	if err := os.MkdirAll(targetDir, 0700); err != nil {
		return nil, err
	}
	mutSrc := devdashboard.NewDiskMutationLogger(targetDir)
	corpus := new(devdashboard.Corpus)
	if err := corpus.Initialize(ctx, mutSrc); err != nil {
		return nil, err
	}
	return corpus, nil
}

// DefaultDir returns the directory containing the cached mutation logs.
func DefaultDir() string {
	return filepath.Join(XdgCacheDir(), "devdashboard")
}

// XdgCacheDir returns the XDG Base Directory Specification cache
// directory.
func XdgCacheDir() string {
	cache := os.Getenv("XDG_CACHE_HOME")
	if cache != "" {
		return cache
	}
	home := homeDir()
	// Not XDG but standard for OS X.
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library/Caches")
	}
	return filepath.Join(home, ".cache")
}

func homeDir() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
	}
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}
	u, err := user.Current()
	if err != nil {
		log.Fatalf("failed to get home directory or current user: %v", err)
	}
	return u.HomeDir
}
