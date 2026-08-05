package modelrouting

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// linkDirectory creates a directory alias, falling back to a Windows junction
// when unprivileged symlink creation is unavailable. It reports whether an
// alias could be created at all.
func linkDirectory(t *testing.T, target, alias string) bool {
	t.Helper()
	err := os.Symlink(target, alias)
	if err == nil {
		return true
	}
	if runtime.GOOS != "windows" {
		return false
	}
	if output, junctionErr := exec.Command("cmd", "/c", "mklink", "/J", alias, target).CombinedOutput(); junctionErr != nil {
		t.Logf("cannot create directory alias: symlink=%v junction=%v output=%s", err, junctionErr, output)
		return false
	}
	return true
}

// A project root reached through a symlink or junction must load normally.
// Containment is proven lexically by safeStoragePath, so the root's own link
// status cannot enable an escape. Regression guard: rejectSymlinkAncestors
// previously stat-ed the root itself and failed every gate run performed from
// an aliased checkout path.
func TestAliasedProjectRootLoadsAndNestedSymlinkStillRejected(t *testing.T) {
	realRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(realRoot, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realRoot, "config", "doc.json"), []byte(`{"schema_version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var target struct {
		SchemaVersion int `json:"schema_version"`
	}
	relDoc := filepath.Join("config", "doc.json")

	alias := filepath.Join(t.TempDir(), "root-alias")
	if !linkDirectory(t, realRoot, alias) {
		t.Skip("directory aliases are unavailable in this environment")
	}
	if err := LoadStrictProjectJSON(alias, relDoc, &target, 1<<20); err != nil {
		t.Fatalf("aliased project root rejected: %v", err)
	}
	if target.SchemaVersion != 1 {
		t.Fatalf("unexpected payload: %#v", target)
	}

	// The relaxation must not extend below the root: a symlinked intermediate
	// directory can still redirect resolution outside the project.
	linked := filepath.Join(realRoot, "linked")
	if !linkDirectory(t, filepath.Join(realRoot, "config"), linked) {
		return
	}
	err := LoadStrictProjectJSON(realRoot, filepath.Join("linked", "doc.json"), &target, 1<<20)
	if !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("nested symlinked directory accepted: %v", err)
	}
}
