package main

import (
	"path/filepath"
	"testing"
)

func TestParityContractForSkillRepoCheckNames(t *testing.T) {
	t.Setenv("KBCHECK_POWERSHELL", "pwsh")
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-check", "SKILL.md"), "---\nname: kb-check\ndescription: test\n---\n")
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")
	for _, script := range []string{
		"scripts/skill-lint.ps1",
		"scripts/route-complexity-eval.ps1",
		"scripts/skill-eval.ps1",
		"scripts/skill-sync-report.ps1",
	} {
		writeFile(t, filepath.Join(root, filepath.FromSlash(script)), "exit 0")
	}

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"route-complexity-eval", "skill-eval", "skill-lint", "skill-sync-report"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("checks=%v want prefix=%v", got, want)
		}
	}
}
