package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProductionDDRContractSeparatesHostSurfaces(t *testing.T) {
	t.Parallel()
	root := ddrTestRepoRoot(t)
	// One policy, one requirement. kb-work and kb-workflow phrase the
	// single-local-route rule differently, so the contract tracks the policy term.
	noSecondLocalRoute := docConcept("second local route")

	requireDocContract(t, root, "production DDR contract", map[string][]contractMatcher{
		"README.md": {
			docConcept("delegation-first DDR"),
			docConcept("Direct current-owner execution remains an evidence-bound exception"),
		},
		".github/skills/kb-work/SKILL.md": {
			docAnchor("### Step 2.6: Orchestrator Ownership Decision (DDR)"),
			docAnchor("**Native host delegation:**"),
			docAnchor("**CLI or user-local delegation:**"),
			docAnchor("DDR route: <current|subagent> | primary:"),
			docAnchor("parent-on-first-local-failure"),
			docConcept("delegation-first DDR"),
			docConcept("current-owner exception gate"),
			docConcept("no-qualified-route"),
			docConcept("execution_owner", "delegated"),
			docConcept("per slice, not per plan"),
			docConcept("subagents in parallel"),
			docConcept("portable tier"),
			docConcept("App-only aliases with CLI-only aliases"),
			docConcept("emit exactly"),
			docConcept("compact user-visible line before mutation"),
			docConcept("sole emitter"),
			docConcept("current orchestrator"),
			docConcept("per-slice DDR route line"),
			noSecondLocalRoute,
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			docAnchor("Route announcement:"),
			docAnchor("Router receipt:"),
			docConcept("immutable orchestration receipts"),
			docConcept("re-decide ownership"),
			docConcept("evidence-backed"),
			docConcept("emit or repeat"),
		},
		"docs/context/architecture/kb-workflow.md": {
			docAnchor("parent-on-first-local-failure"),
			docConcept("same-tier-or-higher subagent"),
			docConcept("one owner per slice, not one worker per plan"),
			docConcept("portable across hosts"),
			docConcept("recognized reason gate"),
			docConcept("callable schema is authoritative"),
			docConcept("kbrouter` is authoritative"),
			docConcept("CLI and user-local routes"),
			docConcept("route announcement is evidence-bound"),
			noSecondLocalRoute,
		},
	})

	canonicalAnnouncement := "DDR route: <current|subagent> | primary: <current orchestrator|evidence-backed-route> | return: <none|parent-on-first-local-failure|required-alias-block> | tier: <small|medium|large> | proof: <short-proof-target>"
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
			skill:  mustMutateContract(t, skillText, "sole emitter", "an emitter alongside the delegated worker"),
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
			skill:  mustMutateContract(t, skillText, "line before mutation", "line after mutation"),
			prompt: executionPrompt,
		},
		"authoritative grammar is duplicated": {
			skill:  skillText + "\n" + canonicalAnnouncement + "\n",
			prompt: executionPrompt,
		},
		"parent return permits a second local route": {
			skill:  mustMutateContract(t, skillText, "second local route", "second remote route"),
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

	path := writeManifest(t, `
---
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: true
  automatic_cross_owner_fallback: false
slices: []
gate_ledger: []
---
`)
	result, err := validateManifestContract(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || !hasManifestIssue(result.Issues, "invalid-model-selection-contract-field") {
		t.Fatalf("automatic downward routing was not mechanically rejected: %#v", result)
	}
}

func TestRetiredModelRoutingSurfacesAreAbsent(t *testing.T) {
	t.Parallel()
	root := ddrTestRepoRoot(t)
	for path := range trackedTextFiles(t, root) {
		slashPath := filepath.ToSlash(path)
		if strings.HasPrefix(slashPath, "cmd/"+strings.Join([]string{"amr", "bench"}, "")) ||
			strings.HasPrefix(slashPath, "evals/"+strings.Join([]string{"amr", "model", "benchmark"}, "-")) {
			t.Errorf("retired model-routing path remains tracked: %s", path)
		}
	}
	forbidden := []string{
		strings.Join([]string{"amr", "bench"}, ""),
		strings.Join([]string{"attempt", "tier"}, "-"),
		strings.Join([]string{"attempt", "tier"}, "_"),
		strings.Join([]string{"support", "cohort"}, "_"),
		strings.Join([]string{"initial", "pilot"}, "-"),
		strings.Join([]string{"model", "routing", "release"}, "_"),
	}
	for path, content := range trackedTextFiles(t, root) {
		text := strings.ToLower(string(content))
		for _, token := range forbidden {
			if strings.Contains(text, token) {
				t.Errorf("%s contains retired model-routing token %q", path, token)
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
	normalizedSkill := normalizeContractText(skillText)
	for _, invariant := range []contractMatcher{
		docConcept("emit exactly"),
		docConcept("compact user-visible line before mutation"),
		docConcept("sole emitter"),
		docConcept("second local route"),
	} {
		if what, absent := invariant.missing(normalizedSkill); absent {
			return fmt.Errorf("kb-work skill missing DDR emission invariant: %s", what)
		}
	}
	if !strings.Contains(normalizeContractText(executionPrompt), normalizeContractText("The route announcement above was already emitted by the orchestrator before")) {
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
