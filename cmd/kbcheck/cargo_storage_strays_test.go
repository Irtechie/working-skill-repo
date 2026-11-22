package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeCargoProfile(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, ".fingerprint"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "artifact.rlib"), []byte("artifact"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func resolveWithEnvironment(t *testing.T, root, configured, runID string, now time.Time) cargoStorageResult {
	t.Helper()
	t.Setenv("CARGO_TARGET_DIR", configured)
	result, err := executeCargoStorage(cargoStorageOptions{Action: "resolve", RunID: runID, RepoRoot: root, Now: now})
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("resolve failed: result=%#v err=%v", result, err)
	}
	return result
}

func TestCargoStorageClimbsOutOfProfileFolderAndForeignKey(t *testing.T) {
	root := initWorktreeRepo(t)
	cacheRoot, err := canonicalPath(filepath.Join(t.TempDir(), "shared-target"))
	if err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(cacheRoot, "debug")
	writeCargoProfile(t, profile)

	fromProfile := resolveWithEnvironment(t, root, profile, "run-profile", time.Now().UTC())
	if filepath.Dir(fromProfile.Receipt.StableTarget) != cacheRoot {
		t.Fatalf("a profile folder was used as a cache root: %s", fromProfile.Receipt.StableTarget)
	}
	foreign := filepath.Join(cacheRoot, "0123456789abcdef01234567")
	fromForeign := resolveWithEnvironment(t, root, foreign, "run-foreign", time.Now().UTC())
	if fromForeign.Receipt.StableTarget != fromProfile.Receipt.StableTarget {
		t.Fatalf("another project's keyed folder was used as a cache root: %s", fromForeign.Receipt.StableTarget)
	}
	// A profile folder inside this project's own keyed target resolves to that target.
	inside := filepath.Join(fromProfile.Receipt.StableTarget, "debug")
	writeCargoProfile(t, inside)
	fromInside := resolveWithEnvironment(t, root, inside, "run-inside", time.Now().UTC())
	if fromInside.Receipt.StableTarget != fromProfile.Receipt.StableTarget {
		t.Fatalf("own profile folder nested a second target: %s", fromInside.Receipt.StableTarget)
	}
	// An empty folder that merely happens to be called debug is still a cache root.
	plain := filepath.Join(t.TempDir(), "debug")
	fromPlain := resolveWithEnvironment(t, root, plain, "run-plain", time.Now().UTC())
	if canonicalDir, _ := canonicalPath(plain); filepath.Dir(fromPlain.Receipt.StableTarget) != canonicalDir {
		t.Fatalf("an unwritten folder named debug was not kept as the cache root: %s", fromPlain.Receipt.StableTarget)
	}
}

func TestCargoStorageResolveRemovesQuietNestedProfileAndOnlyReportsTheRest(t *testing.T) {
	root := initWorktreeRepo(t)
	cacheRoot, err := canonicalPath(filepath.Join(t.TempDir(), "shared-target"))
	if err != nil {
		t.Fatal(err)
	}
	first := resolveWithEnvironment(t, root, cacheRoot, "run-first", time.Now().UTC())
	target := first.Receipt.StableTarget
	if len(first.Strays) != 0 {
		t.Fatalf("clean target reported strays: %#v", first.Strays)
	}
	nested := filepath.Join(target, "debug", "debug")
	writeCargoProfile(t, filepath.Join(target, "debug"))
	writeCargoProfile(t, nested)
	rootOutput := filepath.Join(cacheRoot, "debug")
	writeCargoProfile(t, rootOutput)
	foreign := filepath.Join(target, "0123456789abcdef01234567")
	writeCargoProfile(t, filepath.Join(foreign, "debug"))

	actions := func(result cargoStorageResult) map[string]string {
		found := map[string]string{}
		for _, stray := range result.Strays {
			found[stray.Path] = stray.Kind + ":" + stray.Action
		}
		return found
	}
	// Just written: a build may still be using it, so nothing is removed.
	recent := actions(resolveWithEnvironment(t, root, cacheRoot, "run-recent", time.Now().UTC()))
	if recent[nested] != "nested-profile:kept-recent" {
		t.Fatalf("a recently written nested profile was not kept: %#v", recent)
	}
	later := time.Now().UTC().Add(3 * time.Hour)
	swept := actions(resolveWithEnvironment(t, root, cacheRoot, "run-later", later))
	if swept[nested] != "nested-profile:removed" {
		t.Fatalf("a quiet nested profile was not removed: %#v", swept)
	}
	if _, err := os.Stat(nested); !os.IsNotExist(err) {
		t.Fatalf("nested profile still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(target, "debug", "artifact.rlib")); err != nil {
		t.Fatalf("the real profile was damaged: %v", err)
	}
	if swept[rootOutput] != "cache-root-output:reported" || swept[foreign] != "foreign-key-inside-target:reported" {
		t.Fatalf("shared or foreign output was not reported: %#v", swept)
	}
	for _, kept := range []string{rootOutput, foreign} {
		if _, err := os.Stat(kept); err != nil {
			t.Fatalf("reported output %s was deleted: %v", kept, err)
		}
	}
}
