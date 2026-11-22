package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// isolateCargoStorageEnvironment points the Cargo cache variables at a
// disposable folder for the whole test process, so no test can resolve into a
// real build cache. Child processes inherit the redirected values.
func isolateCargoStorageEnvironment() (func(), error) {
	dir, err := os.MkdirTemp("", "kbcheck-cargo-isolation-")
	if err != nil {
		return nil, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	for name, sub := range map[string]string{
		"CARGO_TARGET_DIR":     "cargo-target",
		"KB_CARGO_TARGET_ROOT": "cargo-root",
	} {
		if err := os.Setenv(name, filepath.Join(dir, sub)); err != nil {
			cleanup()
			return nil, err
		}
	}
	return cleanup, nil
}

// runIsolatedTests runs the package tests under the isolated environment.
func runIsolatedTests(m *testing.M) int {
	cleanup, err := isolateCargoStorageEnvironment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cargo storage isolation: %v\n", err)
		return 1
	}
	defer cleanup()
	return m.Run()
}

func TestCargoStorageTestProcessIsIsolatedFromRealCache(t *testing.T) {
	configured := os.Getenv("CARGO_TARGET_DIR")
	if !filepath.IsAbs(configured) {
		t.Fatalf("CARGO_TARGET_DIR must be absolute in tests, got %q", configured)
	}
	tempArea, err := filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	resolvedConfigured := configured
	if parent, err := filepath.EvalSymlinks(filepath.Dir(configured)); err == nil {
		resolvedConfigured = filepath.Join(parent, filepath.Base(configured))
	}
	if !cargoIsolationWithin(tempArea, resolvedConfigured) {
		t.Fatalf("CARGO_TARGET_DIR %q is outside the temporary area %q", configured, tempArea)
	}
	if root := os.Getenv("KB_CARGO_TARGET_ROOT"); !strings.HasPrefix(filepath.Clean(root), filepath.Dir(filepath.Clean(configured))) {
		t.Fatalf("KB_CARGO_TARGET_ROOT %q is not beside the redirected target", root)
	}

	repo := initWorktreeRepo(t)
	result, err := executeCargoStorage(cargoStorageOptions{
		Action: "resolve", RunID: "isolation-proof", RepoRoot: repo, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("resolve failed: result=%#v err=%v", result, err)
	}
	target := result.Receipt.StableTarget
	if !cargoIsolationWithin(resolvedConfigured, target) {
		// The resolver canonicalizes; compare against the canonical form too.
		canonical, cerr := canonicalPath(configured)
		if cerr != nil || !cargoIsolationWithin(canonical, target) {
			t.Fatalf("resolved target %q is not under the redirected cache %q", target, configured)
		}
	}
}

func cargoIsolationWithin(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
