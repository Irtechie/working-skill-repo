package main

import (
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// A nested build tree is removed only after it has been quiet this long, so a
// build that is still writing into it is never pulled out from under itself.
const cargoStrayActivityFloor = 2 * time.Hour

var cargoProjectKeyPattern = regexp.MustCompile(`^[0-9a-f]{24}$`)

// cargoStrayTarget is build output that sits beside or inside the stable
// target but is not part of it.
type cargoStrayTarget struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Bytes  int64  `json:"bytes"`
	Action string `json:"action"`
}

func cargoProfileDirName(name string) bool {
	return name == "debug" || name == "release"
}

// cargoOutputDir reports whether Cargo itself wrote a profile into path.
func cargoOutputDir(path string) bool {
	info, err := os.Lstat(path)
	if err != nil || !info.IsDir() {
		return false
	}
	for _, marker := range []string{".fingerprint", ".cargo-lock"} {
		if _, err := os.Lstat(filepath.Join(path, marker)); err == nil {
			return true
		}
	}
	return false
}

// cargoCacheRootFor climbs out of a configured path that is not a cache root:
// a Cargo profile folder that already holds build output, or another
// project's keyed target. Keying inside either would nest a second full build
// tree where no cleanup looks for it.
func cargoCacheRootFor(configured, key string) string {
	root := configured
	for {
		base := filepath.Base(root)
		parent := filepath.Dir(root)
		if parent == root || base == key {
			return root
		}
		foreignKey := cargoProjectKeyPattern.MatchString(base)
		writtenProfile := cargoProfileDirName(base) && cargoOutputDir(root)
		if !foreignKey && !writtenProfile {
			return root
		}
		root = parent
	}
}

func cargoDirQuiet(path string, now time.Time) bool {
	info, err := os.Lstat(path)
	if err != nil || now.Sub(info.ModTime()) < cargoStrayActivityFloor {
		return false
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		child, err := entry.Info()
		if err != nil || now.Sub(child.ModTime()) < cargoStrayActivityFloor {
			return false
		}
	}
	return true
}

// sweepCargoStrays runs on every resolve. Inside the stable target, which this
// resolver owns, it removes a profile-in-profile tree (the result of pointing
// CARGO_TARGET_DIR at a profile folder) once that tree is quiet. Everything
// else is reported and left alone: Cargo output written straight into the
// shared cache root, and another project's keyed folder inside this target.
// Those may belong to a tool or person this run knows nothing about.
func sweepCargoStrays(target string, now time.Time) []cargoStrayTarget {
	var strays []cargoStrayTarget
	record := func(path, kind, action string, bytes int64) {
		strays = append(strays, cargoStrayTarget{Path: path, Kind: kind, Bytes: bytes, Action: action})
	}
	for _, outer := range []string{"debug", "release"} {
		for _, inner := range []string{"debug", "release"} {
			nested := filepath.Join(target, outer, inner)
			if !cargoOutputDir(nested) {
				continue
			}
			bytes, _ := pathSize(nested)
			switch {
			case !cargoDirQuiet(nested, now):
				record(nested, "nested-profile", "kept-recent", bytes)
			case os.RemoveAll(nested) != nil:
				record(nested, "nested-profile", "remove-failed", bytes)
			default:
				record(nested, "nested-profile", "removed", bytes)
			}
		}
	}
	if entries, err := os.ReadDir(target); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && cargoProjectKeyPattern.MatchString(entry.Name()) {
				path := filepath.Join(target, entry.Name())
				bytes, _ := pathSize(path)
				record(path, "foreign-key-inside-target", "reported", bytes)
			}
		}
	}
	cacheRoot := filepath.Dir(target)
	for _, profile := range []string{"debug", "release"} {
		path := filepath.Join(cacheRoot, profile)
		if cargoOutputDir(path) {
			bytes, _ := pathSize(path)
			record(path, "cache-root-output", "reported", bytes)
		}
	}
	return strays
}
