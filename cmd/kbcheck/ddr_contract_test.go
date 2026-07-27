package main

import (
	"fmt"
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
			"normal work uses delegation-first DDR",
			"The current-owner exception gate accepts only",
			"`no-qualified-route` is valid only after inspecting both",
			"Default `execution_owner` to `delegated`",
			"\"Exactly one\" is per slice, not per plan.",
			"dispatch those subagents in parallel",
			"Resolve that portable tier when the plan is picked up.",
			"AMR remains an unpromoted experimental benchmark.",
			"App-only aliases with CLI-only aliases.",
			"DDR route: <current|subagent> | primary:",
			"After the ownership decision and, when delegated, route selection, emit exactly",
			"one compact user-visible line before mutation or worker dispatch",
			"The orchestrator is the sole emitter",
			"otherwise use `current orchestrator`",
			"(conditional; explicit reselect)",
			"This preview rule never suppresses the mandatory per-slice DDR route line.",
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			"Do not invoke AMR or pass `attempt_tier` during normal KB work.",
			"Route announcement:",
			"Router receipt:",
			"immutable orchestration receipts",
			"Do not re-decide ownership, discover or select a route, dispatch, or delegate",
			"never replace an evidence-backed",
			"Do not emit or repeat it",
		},
		".github/skills/kb-configure/references/kb-routing-example.yaml": {
			"experimental_amr:",
			"affects_normal_work: false",
		},
		"docs/context/architecture/kb-workflow.md": {
			"One qualified same-tier-or-higher subagent normally owns",
			"one owner per slice, not one worker per plan",
			"The tier is portable across hosts.",
			"recognized reason gate",
			"The active host's callable schema is authoritative for native targets.",
			"`kbrouter` is authoritative",
			"for Codex CLI and user-local routes",
			"Normal work never passes `attempt_tier`",
			"route announcement is evidence-bound",
		},
		"README.md": {
			"A planned tier is portable",
			"one qualified subagent normally executes each bounded slice",
			"one owner per slice, not one worker per plan",
			"semi-gated exception",
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
	skillText := readDDRContractFile(t, root, ".github/skills/kb-work/SKILL.md")
	executionPrompt := readDDRContractFile(t, root, ".github/skills/kb-work/references/execution-prompt.md")
	if err := validateDDREmissionContract(skillText, executionPrompt, canonicalAnnouncement); err != nil {
		t.Fatal(err)
	}

	mutations := map[string]struct {
		skill  string
		prompt string
	}{
		"worker becomes an emitter": {
			skill:  strings.Replace(skillText, "The orchestrator is the sole emitter", "The orchestrator or delegated worker may emit", 1),
			prompt: executionPrompt,
		},
		"worker prompt adds an emission instruction": {
			skill:  skillText,
			prompt: executionPrompt + "\nEmit exactly one compact user-visible DDR route line.\n",
		},
		"worker prompt adds routing authority": {
			skill:  skillText,
			prompt: executionPrompt + "\nDiscover or select a route and dispatch it.\n",
		},
		"announcement moves after mutation": {
			skill: strings.Replace(
				skillText,
				"one compact user-visible line before mutation or worker dispatch",
				"one compact user-visible line after mutation or worker dispatch",
				1,
			),
			prompt: executionPrompt,
		},
		"authoritative grammar is duplicated": {
			skill:  skillText + "\n" + canonicalAnnouncement + "\n",
			prompt: executionPrompt,
		},
		"named fallback permits a lower tier": {
			skill: strings.Replace(
				skillText,
				"A named fallback must be proven same-tier-or-higher eligible",
				"A named fallback may be lower-tier when eligible",
				1,
			),
			prompt: executionPrompt,
		},
	}
	for name, mutation := range mutations {
		t.Run(name, func(t *testing.T) {
			if err := validateDDREmissionContract(mutation.skill, mutation.prompt, canonicalAnnouncement); err == nil {
				t.Fatal("mutated contract unexpectedly passed")
			}
		})
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

func validateDDREmissionContract(skillText, executionPrompt, canonicalAnnouncement string) error {
	if count := strings.Count(skillText, canonicalAnnouncement); count != 1 {
		return fmt.Errorf("kb-work skill must contain the authoritative DDR grammar exactly once; got %d", count)
	}
	if count := strings.Count(executionPrompt, canonicalAnnouncement); count != 1 {
		return fmt.Errorf("delegated execution prompt must carry the populated DDR receipt exactly once; got %d", count)
	}
	for _, required := range []string{
		"After the ownership decision and, when delegated, route selection, emit exactly",
		"one compact user-visible line before mutation or worker dispatch",
		"The orchestrator is the sole emitter",
		"A named fallback must be proven same-tier-or-higher eligible",
	} {
		if !strings.Contains(skillText, required) {
			return fmt.Errorf("kb-work skill missing DDR emission invariant %q", required)
		}
	}
	if !strings.Contains(executionPrompt, "The route announcement above was already emitted by the orchestrator before") {
		return fmt.Errorf("delegated execution prompt must identify the orchestrator as the prior emitter")
	}
	lowerPrompt := strings.ToLower(executionPrompt)
	for _, forbidden := range []string{
		"emit exactly one compact",
		"emit the compact ddr route",
		"announce the ddr route",
		"discover or select a route and dispatch it",
		"select exactly one qualified",
		"call the exact native target",
	} {
		if strings.Contains(lowerPrompt, forbidden) {
			return fmt.Errorf("delegated execution prompt contains worker emission instruction %q", forbidden)
		}
	}
	return nil
}

func readDDRContractFile(t *testing.T, root, path string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(content)
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
