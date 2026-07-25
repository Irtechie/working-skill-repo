package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionDDRContractExcludesAMRAndSeparatesHostSurfaces(t *testing.T) {
	root := ddrTestRepoRoot(t)
	required := map[string][]string{
		".github/skills/kb-work/SKILL.md": {
			"### Step 2.6: Orchestrator Ownership Decision (DDR)",
			"**Native host delegation:**",
			"**CLI or user-local delegation:**",
			"AMR remains an unpromoted experimental benchmark.",
			"App-only aliases with CLI-only aliases.",
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			"Do not invoke AMR or pass `attempt_tier` during normal KB work.",
			"Do not send App targets through",
		},
		".github/skills/kb-configure/references/kb-routing-example.yaml": {
			"experimental_amr:",
			"affects_normal_work: false",
		},
		"docs/context/architecture/kb-workflow.md": {
			"The active host's callable schema is authoritative for native targets.",
			"`kbrouter` is authoritative",
			"for Codex CLI and user-local routes",
			"Normal work never passes `attempt_tier`",
		},
	}
	for path, needles := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		text := string(content)
		for _, needle := range needles {
			if !strings.Contains(text, needle) {
				t.Errorf("%s missing production DDR contract %q", path, needle)
			}
		}
	}

	forbidden := []string{
		"AMR is automatic",
		"planned-tier AMR selection is automatic",
		"may make one explicit lower-tier attempt",
	}
	for _, path := range []string{
		".github/skills/kb-plan/SKILL.md",
		".github/skills/kb-work/SKILL.md",
		".github/skills/kb-functional-test/SKILL.md",
		"docs/context/architecture/kb-workflow.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for _, phrase := range forbidden {
			if strings.Contains(string(content), phrase) {
				t.Errorf("%s contains forbidden normal-path AMR phrase %q", path, phrase)
			}
		}
	}
}

func ddrTestRepoRoot(t *testing.T) string {
	t.Helper()
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(current, ".github", "skills")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			t.Fatal("repository root not found")
		}
		current = parent
	}
}
