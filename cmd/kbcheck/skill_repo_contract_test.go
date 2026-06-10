package main

import (
	"path/filepath"
	"testing"
)

func TestSkillRepoContractForNativeCheckNames(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-check", "SKILL.md"), "---\nname: kb-check\ndescription: test\n---\n")
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{
		"cross-model-benchmark-validate",
		"kb-pipeline-selftest",
		"kb-release-gate-selftest",
		"kb-work-ready-set-selftest",
		"kb-work-scope-lease-selftest",
		"route-complexity-eval",
		"skill-eval",
		"skill-lint",
		"skill-marketplace-firebreak",
		"skill-marketplace-firebreak-selftest",
		"skill-surface-minimality",
		"skill-surface-minimality-selftest",
		"skill-surface-report",
		"workflow-governor-selftest",
	}
	if len(got) < len(want) {
		t.Fatalf("checks=%v want at least %v", got, want)
	}
	for _, name := range want {
		if !contains(got, name) {
			t.Fatalf("checks=%v missing %s", got, name)
		}
	}
}
