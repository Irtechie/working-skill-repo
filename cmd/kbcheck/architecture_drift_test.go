package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeArchitectureFixture(t *testing.T, root string, cfg architectureConfig, components []string, docs map[string]string) {
	t.Helper()

	for _, name := range components {
		if err := os.MkdirAll(filepath.Join(root, "cmd", name), 0o755); err != nil {
			t.Fatalf("mkdir component %s: %v", name, err)
		}
	}

	docsDir := cfg.DocsDir
	if docsDir == "" {
		docsDir = "docs/context/architecture"
	}
	full := filepath.Join(root, filepath.FromSlash(docsDir))
	if err := os.MkdirAll(full, 0o755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	for name, body := range docs {
		if err := os.WriteFile(filepath.Join(full, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write doc %s: %v", name, err)
		}
	}

	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	cfgPath := filepath.Join(root, filepath.FromSlash(architectureConfigRelPath))
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(cfgPath, raw, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func TestArchitectureDriftNotConfiguredIsNotAFailure(t *testing.T) {
	root := t.TempDir()
	report, err := computeArchitectureDrift(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Status != "not-configured" {
		t.Fatalf("status = %q, want not-configured", report.Status)
	}
}

func TestArchitectureDriftCleanRepo(t *testing.T) {
	root := t.TempDir()
	writeArchitectureFixture(t, root,
		architectureConfig{Roots: []string{"cmd"}, Components: []string{"alpha", "beta"}},
		[]string{"alpha", "beta"},
		map[string]string{"overview.md": "alpha handles ingest. beta handles routing."},
	)

	report, err := computeArchitectureDrift(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.Status != "ok" {
		t.Fatalf("status = %q, want ok (undeclared=%v phantom=%v undocumented=%v)",
			report.Status, report.Undeclared, report.Phantom, report.Undocumented)
	}
}

// A new component appearing on disk without a declaration is the ratchet this
// check exists to stop.
func TestArchitectureDriftDetectsUndeclaredComponent(t *testing.T) {
	root := t.TempDir()
	writeArchitectureFixture(t, root,
		architectureConfig{Roots: []string{"cmd"}, Components: []string{"alpha"}},
		[]string{"alpha", "sneaky-helper"},
		map[string]string{"overview.md": "alpha handles ingest."},
	)

	report, err := computeArchitectureDrift(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Undeclared) != 1 || report.Undeclared[0].Component != "sneaky-helper" {
		t.Fatalf("undeclared = %+v, want exactly sneaky-helper", report.Undeclared)
	}
	if report.Status != "drift" {
		t.Fatalf("status = %q, want drift", report.Status)
	}
}

func TestArchitectureDriftDetectsPhantomComponent(t *testing.T) {
	root := t.TempDir()
	writeArchitectureFixture(t, root,
		architectureConfig{Roots: []string{"cmd"}, Components: []string{"alpha", "ghost"}},
		[]string{"alpha"},
		map[string]string{"overview.md": "alpha handles ingest. ghost handles nothing."},
	)

	report, err := computeArchitectureDrift(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Phantom) != 1 || report.Phantom[0].Component != "ghost" {
		t.Fatalf("phantom = %+v, want exactly ghost", report.Phantom)
	}
}

func TestArchitectureDriftDetectsUndocumentedComponent(t *testing.T) {
	root := t.TempDir()
	writeArchitectureFixture(t, root,
		architectureConfig{Roots: []string{"cmd"}, Components: []string{"alpha", "beta"}},
		[]string{"alpha", "beta"},
		map[string]string{"overview.md": "alpha handles ingest."},
	)

	report, err := computeArchitectureDrift(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Undocumented) != 1 || report.Undocumented[0].Component != "beta" {
		t.Fatalf("undocumented = %+v, want exactly beta", report.Undocumented)
	}
}

// "memory" must not be considered documented by a mention of "memory-eval".
func TestArchitectureDriftDocMatchIsTokenBounded(t *testing.T) {
	root := t.TempDir()
	writeArchitectureFixture(t, root,
		architectureConfig{Roots: []string{"cmd"}, Components: []string{"memory"}},
		[]string{"memory"},
		map[string]string{"overview.md": "memory-eval harness notes."},
	)

	report, err := computeArchitectureDrift(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Undocumented) != 1 || report.Undocumented[0].Component != "memory" {
		t.Fatalf("undocumented = %+v, want memory flagged despite memory-eval mention", report.Undocumented)
	}
}

func TestArchitectureDriftExemptDirectoriesAreIgnored(t *testing.T) {
	root := t.TempDir()
	writeArchitectureFixture(t, root,
		architectureConfig{Roots: []string{"cmd"}, Components: []string{"alpha"}, Exempt: []string{"testdata"}},
		[]string{"alpha", "testdata"},
		map[string]string{"overview.md": "alpha handles ingest."},
	)

	report, err := computeArchitectureDrift(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Undeclared) != 0 {
		t.Fatalf("undeclared = %+v, want none (testdata is exempt)", report.Undeclared)
	}
}

func TestArchitectureDriftMissingRootIsAnError(t *testing.T) {
	root := t.TempDir()
	writeArchitectureFixture(t, root,
		architectureConfig{Roots: []string{"nonexistent"}, Components: []string{"alpha"}},
		[]string{"alpha"},
		map[string]string{"overview.md": "alpha handles ingest."},
	)

	if _, err := computeArchitectureDrift(root); err == nil {
		t.Fatal("expected error for a declared root that does not exist")
	}
}

func TestArchitectureDriftInitWritesDeclaration(t *testing.T) {
	root := t.TempDir()
	mkdirAllT(t, filepath.Join(root, "apps", "alpha"))
	mkdirAllT(t, filepath.Join(root, "apps", "beta"))
	mkdirAllT(t, filepath.Join(root, "apps", ".hidden"))

	var out, errOut bytes.Buffer
	opts := options{sliceLeaseAction: "init", architectureRoots: "apps"}
	if code := runArchitectureDriftCommand(root, opts, &out, &errOut); code != 0 {
		t.Fatalf("init exit=%d stderr=%s", code, errOut.String())
	}

	raw, err := os.ReadFile(filepath.Join(root, architectureConfigRelPath))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var cfg architectureConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if got := strings.Join(cfg.Components, ","); got != "alpha,beta" {
		t.Fatalf("components = %q, want alpha,beta (hidden dirs excluded)", got)
	}
	if cfg.DocsDir != "docs/context/architecture" {
		t.Fatalf("docs_dir = %q", cfg.DocsDir)
	}
}

func TestArchitectureDriftInitRefusesToOverwrite(t *testing.T) {
	root := t.TempDir()
	mkdirAllT(t, filepath.Join(root, "apps", "alpha"))
	mkdirAllT(t, filepath.Join(root, "config"))
	existing := []byte(`{"roots":["apps"],"components":["alpha"]}`)
	if err := os.WriteFile(filepath.Join(root, architectureConfigRelPath), existing, 0o644); err != nil {
		t.Fatal(err)
	}

	var out, errOut bytes.Buffer
	opts := options{sliceLeaseAction: "init", architectureRoots: "apps"}
	if code := runArchitectureDriftCommand(root, opts, &out, &errOut); code == 0 {
		t.Fatal("init overwrote an existing declaration; friction is the mechanism")
	}

	after, err := os.ReadFile(filepath.Join(root, architectureConfigRelPath))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(existing) {
		t.Fatalf("config mutated: %s", after)
	}
}

func mkdirAllT(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
