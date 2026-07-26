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
			"DDR route: <current|subagent> | primary:",
			"after route selection and before mutation or worker dispatch",
			"The orchestrator is the sole emitter",
			"otherwise use `current orchestrator`",
			"(conditional; explicit reselect)",
			"This preview rule never suppresses the mandatory per-slice DDR route line.",
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			"Do not invoke AMR or pass `attempt_tier` during normal KB work.",
			"Do not send App targets through",
			"Route announcement:",
			"Never name a model or alias",
			"Do not emit or repeat the route announcement.",
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
			"route announcement is evidence-bound",
		},
		"README.md": {
			"DDR route: <current|subagent> | primary:",
			"A named fallback is conditional",
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

	canonicalAnnouncement := "DDR route: <current|subagent> | primary: <current orchestrator|evidence-backed-route> | fallback: <none|explicit same-tier/higher reselection|evidence-backed-route (conditional; explicit reselect)> | tier: <small|medium|large> | proof: <short-proof-target>"
	for _, path := range []string{
		".github/skills/kb-work/SKILL.md",
		".github/skills/kb-work/references/execution-prompt.md",
		"README.md",
		"docs/context/architecture/kb-workflow.md",
		"docs/plans/2026-07-26-001-tool-ddr-route-announcement-plan.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if !strings.Contains(string(content), canonicalAnnouncement) {
			t.Errorf("%s missing canonical DDR announcement grammar", path)
		}
	}

	executionPrompt, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(".github/skills/kb-work/references/execution-prompt.md")))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(executionPrompt), "Emit exactly one compact user-visible line before execution") {
		t.Error("delegated execution prompt must not become a second route-announcement emitter")
	}

	slicePlan, err := os.ReadFile(filepath.Join(root, filepath.FromSlash("docs/plans/2026-07-26-001-tool-ddr-route-announcement-plan.md")))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbiddenAlias := range []string{"local Qwen", "Luna"} {
		if strings.Contains(string(slicePlan), forbiddenAlias) {
			t.Errorf("portable route-announcement plan hard-codes model alias %q", forbiddenAlias)
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
