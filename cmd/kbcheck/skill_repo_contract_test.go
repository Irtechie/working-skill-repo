package main

import (
	"os"
	"path/filepath"
	"strings"
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
		"context-packet-selftest",
		"cross-model-benchmark-validate",
		"dishonest-completion-selftest",
		"execution-telemetry-selftest",
		"kb-doctor-selftest",
		"kb-pipeline-selftest",
		"kb-release-gate-selftest",
		"kb-run-state-selftest",
		"kb-work-ready-set-selftest",
		"kb-work-slice-lease-selftest",
		"kb-work-scope-lease-selftest",
		"provider-hygiene",
		"provider-hygiene-selftest",
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

func TestDeliveryOwnerSkillContracts(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	required := map[string][]string{
		".github/skills/kb-work/SKILL.md": {
			"never merges or pushes a resolved default branch",
			"Missing delivery policy is local-only",
			"local leases are not team locks",
		},
		".github/skills/kb-complete/SKILL.md": {
			"reviewed manifest-owned plan-run branch is the only delivery candidate",
			"Absent policy is always local-only",
			"PR/manual is the recommended team policy",
		},
		".github/skills/kb-ship/SKILL.md": {
			"reviewed plan-run topic branch is the only shipping candidate",
			"PR/manual stops with the correctly based open PR",
			"`kb-ship` never merges it",
		},
		".github/skills/kb-land/SKILL.md": {
			"only KB skill authorized to integrate the resolved remote default branch",
			"Absence of delivery policy is local-only",
		},
		".github/skills/kb-configure/SKILL.md": {
			"mode: local",
			"PR/manual is the recommended team policy",
			"not cross-machine team locks",
		},
	}
	for relative, tokens := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatalf("read %s: %v", relative, err)
		}
		text := strings.ToLower(strings.Join(strings.Fields(string(content)), " "))
		for _, token := range tokens {
			if !strings.Contains(text, strings.ToLower(strings.Join(strings.Fields(token), " "))) {
				t.Errorf("%s missing delivery boundary %q", relative, token)
			}
		}
	}
}
