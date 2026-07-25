package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFixtureQualificationRejectsMutableOracleAndWeakBaseline(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "source.go", "package fixture")
	writeFixtureFile(t, root, "source_test.go", "package fixture")
	hash, err := fileSHA256(filepath.Join(root, "source_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	spec := fixtureSpec{
		Root: root, MutablePaths: []string{"source.go", "source_test.go"},
		ProofFiles: []string{"source_test.go"}, ExpectedProofHashes: map[string]string{"source_test.go": hash},
	}
	if _, err := qualifyFixture(spec); err == nil || !strings.Contains(err.Error(), "proof file") {
		t.Fatalf("mutable oracle err=%v", err)
	}
	spec.MutablePaths = []string{"source.go"}
	if _, err := qualifyFixture(spec); err == nil || !strings.Contains(err.Error(), "known solution") {
		t.Fatalf("missing solution err=%v", err)
	}
}

func TestFixtureQualificationBindsProofClosureAndSensitivity(t *testing.T) {
	root := t.TempDir()
	writeFixtureFile(t, root, "source.go", "package fixture")
	writeFixtureFile(t, root, "source_test.go", "package fixture")
	hash, err := fileSHA256(filepath.Join(root, "source_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyProofClosure(root, map[string]string{"source_test.go": hash}); err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, root, "extra_test.go", "package fixture")
	if err := verifyProofClosure(root, map[string]string{"source_test.go": hash}); err == nil {
		t.Fatal("undeclared proof file passed")
	}
}

func TestCopyDirRejectsFixtureSymlink(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(target, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "leak.txt")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if err := copyDir(root, filepath.Join(t.TempDir(), "copy")); err == nil {
		t.Fatal("fixture symlink passed")
	}
}

func writeFixtureFile(t *testing.T, root, name, content string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
