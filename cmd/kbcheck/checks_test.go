package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverPackageChecks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"scripts":{"lint":"eslint .","test":"vitest","unused":"noop"}}`)
	writeFile(t, filepath.Join(root, "pnpm-lock.yaml"), "")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"lint", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
	if checks[0].Args[0] != "pnpm" {
		t.Fatalf("expected pnpm runner, got %v", checks[0].Args)
	}
}

func TestDiscoverSkillRepoChecksIncludesNativeValidators(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-check", "SKILL.md"), "---\nname: kb-check\ndescription: test\n---\n")
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{
		"atv-upstream-delta",
		"atv-upstream-delta-selftest",
		"cross-model-benchmark-validate",
		"dishonest-completion-selftest",
		"kb-doctor-selftest",
		"kb-pipeline-selftest",
		"kb-release-gate-selftest",
		"kb-run-state-selftest",
		"kb-work-ready-set-selftest",
		"kb-work-scope-lease-selftest",
		"manifest-contract-selftest",
		"marketplace-promotion-selftest",
		"review-reference-guard",
		"route-complexity-eval",
		"skill-eval",
		"skill-eval-baseline-selftest",
		"skill-eval-codex-dry-run",
		"skill-eval-ghcp-dry-run",
		"skill-eval-manifest-selftest",
		"skill-eval-observed-trace-dry-run",
		"skill-eval-quality",
		"skill-lint",
		"skill-marketplace-firebreak",
		"skill-marketplace-firebreak-selftest",
		"skill-surface-minimality",
		"skill-surface-minimality-selftest",
		"skill-surface-report",
		"workflow-governor-selftest",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
}

func TestDiscoverNestedDotnetProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "App", "App.csproj"), "<Project></Project>")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"dotnet-test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
}

func checkNames(checks []Check) []string {
	names := make([]string, 0, len(checks))
	for _, check := range checks {
		names = append(names, check.Name)
	}
	return names
}

// TestKBNativeRootsRecognized is the protected oracle for slice-015.
// It asserts the harness reads observations from .kb/observations.jsonl (kb-native
// ephemeral root) and does NOT require .atv/ to be present.
// RED: fails against pre-slice-015 code because minimality reads .atv/observations.jsonl.
// GREEN: passes after slice-015 changes minimality to read .kb/observations.jsonl.
func TestKBNativeRootsRecognized(t *testing.T) {
	root := t.TempDir()
	// Write a skill with evidence ONLY in .kb/observations.jsonl (kb-native root).
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-skill", "SKILL.md"),
		"---\nname: kb-skill\ndescription: test kb-native skill\n---\n# KB Skill\n")
	writeFile(t, filepath.Join(root, ".kb", "observations.jsonl"),
		`{"tool":"kb-skill","result":"used"}`+"\n")

	// .atv/ must NOT exist — confirm the harness does not require it.
	if _, err := os.Stat(filepath.Join(root, ".atv")); err == nil {
		t.Fatal("test setup error: .atv/ should not exist in the temp dir")
	}

	report, err := computeMinimality(root, ".github/skills", ".github/agents", 6)
	if err != nil {
		t.Fatalf("computeMinimality returned error: %v", err)
	}

	var found *minimalityRow
	for i := range report.SkillClassifications {
		if report.SkillClassifications[i].Name == "kb-skill" {
			found = &report.SkillClassifications[i]
			break
		}
	}
	if found == nil {
		t.Fatal("kb-skill not found in minimality report")
	}
	// The skill must have runtime evidence sourced from .kb/observations.jsonl.
	if found.EvidenceClass != "runtime" {
		t.Fatalf("expected kb-skill to have runtime evidence from .kb/observations.jsonl, got EvidenceClass=%q (classification=%q); .atv/observations.jsonl must NOT be required",
			found.EvidenceClass, found.Classification)
	}
}
