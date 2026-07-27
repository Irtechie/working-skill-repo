//go:build windows

package modelrouting

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCanonicalProjectIdentityStableAcrossWindowsJunction(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(project, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	junction := filepath.Join(root, "project-junction")
	if output, err := exec.Command("cmd", "/c", "mklink", "/J", junction, project).CombinedOutput(); err != nil {
		t.Fatalf("create junction: %v output=%s", err, output)
	}

	direct, err := CanonicalProjectIdentity(project)
	if err != nil {
		t.Fatalf("canonical direct project identity: %v", err)
	}
	aliased, err := CanonicalProjectIdentity(junction)
	if err != nil {
		t.Fatalf("canonical junction project identity: %v", err)
	}
	if aliased != direct {
		t.Fatalf("junction identity=%q want direct identity %q", aliased, direct)
	}
}

func TestCanonicalProjectIdentityRejectsMissingWindowsPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing-project")
	if _, err := CanonicalProjectIdentity(missing); err == nil {
		t.Fatal("missing project path received a canonical identity")
	}
}
