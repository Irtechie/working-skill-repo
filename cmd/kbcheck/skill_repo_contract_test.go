package main

import (
	"path/filepath"
	"testing"
)

func TestSkillRepoContractForNativeCheckNames(t *testing.T) {
	t.Parallel()
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

func TestWebUIProofRemainsAgentOwned(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	requireDocContract(t, root, "web UI proof contract", map[string][]contractMatcher{
		".github/skills/kb-qa/SKILL.md": {
			docConcept("local/public UI functional checks"),
			docConcept("Do not ask the user whether to run browser proof"),
			docConcept("browser proof headless automatically"),
			docConcept("proof receipt containing the route"),
			docConcept("do not report partial completion"),
			docConcept("automation gap into manual verification"),
		},
		".github/skills/kb-functional-test/SKILL.md": {
			docConcept("Do not ask the user whether to run browser proof"),
			docConcept("browser proof headless automatically"),
			docConcept("receipt names the route"),
			docConcept("agent-owned test work"),
		},
	})
}

func TestPlanRunWorktreeAndBranchShareFunnyTaskName(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	requireDocContract(t, root, "worktree naming contract", map[string][]contractMatcher{
		".github/skills/kb-work/SKILL.md": {
			docConcept("same codename for the worktree directory and plan-run branch"),
			docConcept("relate recognizably to the task"),
		},
		".github/skills/kb-work/references/worktree-isolation.md": {
			docAnchor("`codex/the-reviewers-have-unionized`"),
			docConcept("share that exact codename"),
		},
	})
}

func TestDeliveryOwnerSkillContracts(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	requireDocContract(t, root, "delivery boundary", map[string][]contractMatcher{
		".github/skills/kb-start/SKILL.md": {
			docAnchor("terminal-cleanup --action sweep"),
			docConcept("current executing session"),
		},
		".github/skills/kb-work/SKILL.md": {
			docConcept("never merges or pushes a resolved default branch"),
			docConcept("never delivers under any policy"),
			docConcept("local leases are not team locks"),
		},
		".github/skills/kb-complete/SKILL.md": {
			docAnchor("terminal-cleanup --action register"),
			docConcept("only delivery candidate"),
			docConcept("Absent policy is PR/manual"),
			docConcept("no policy authorizes merge"),
			docConcept("Wait for review is the safe default"),
			docConcept("Do not ask a terminal integration question"),
			docConcept("release the shared work claim"),
		},
		".github/skills/kb-ship/SKILL.md": {
			docConcept("only shipping candidate"),
			docConcept("correctly based open PR"),
			docConcept("`kb-ship` never merges it"),
			docConcept("remote topic contains the delivered commit"),
		},
		".github/skills/kb-land/SKILL.md": {
			docConcept("authorized to integrate the resolved remote default branch"),
			docConcept("Absence of delivery policy never authorizes landing"),
			docConcept("remote default contains the delivered commit"),
		},
		".github/skills/kb-configure/SKILL.md": {
			docAnchor("mode: pr"),
			docConcept("PR/manual is the default"),
			docConcept("accepting a PR never is"),
			docConcept("not cross-machine team locks"),
		},
	})
}

func TestCargoBuildStorageContract(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	requireDocContract(t, root, "Cargo build-storage contract", map[string][]contractMatcher{
		".github/skills/kb-check/SKILL.md": {
			docAnchor("cargo-storage --action resolve"),
			docAnchor("cargo-storage --action validate-ready"),
			docAnchor("cargo-storage --action not-applicable"),
			docConcept("`cmd/kbcheck help` output includes `cargo-storage`"),
			docConcept("predates `cargo-storage`"),
			docConcept("24 lowercase hex characters of SHA-256"),
			docConcept("already ends with that exact project key"),
			docConcept("prohibited in fallback mode"),
		},
		".github/skills/kb-fix/SKILL.md": {
			docAnchor("target-repro"),
			docConcept("receipt validated by `validate-ready`"),
			docConcept("Preserve the exact `CARGO_TARGET_DIR`"),
		},
		".github/skills/kb-troubleshoot/SKILL.md": {
			docAnchor("target-repro"),
			docConcept("receipt accepted by `validate-ready`"),
			docConcept("Reuse its exact `CARGO_TARGET_DIR`"),
		},
		".github/skills/kb-repair/SKILL.md": {
			docAnchor("target-repair"),
			docConcept("fail-closed portable fallback contract"),
			docConcept("apply its exact `CARGO_TARGET_DIR`"),
		},
		".github/skills/kb-work/SKILL.md": {
			docConcept("accepted by `cargo-storage --action validate-ready`"),
			docConcept("exact `CARGO_TARGET_DIR` across every slice"),
			docConcept("temporary target must be created first in native mode"),
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			docAnchor("Build-storage contract:"),
			docConcept("portable-fallback record"),
			docConcept("run-specific target"),
		},
		".github/skills/kb-finalize/SKILL.md": {
			docAnchor("cargo-storage --action finalize"),
			docAnchor("retained_bytes"),
			docAnchor("removed_bytes"),
			docConcept("temporary targets and deletion were prohibited"),
		},
		".github/skills/kb-complete/SKILL.md": {
			docAnchor("cargo-storage --action validate"),
			docAnchor("done-portable-fallback"),
			docAnchor("not-applicable"),
			docConcept("canonical Git common-directory identity"),
			docConcept("`removed_bytes` to be zero"),
		},
	})
}

func TestPlanWideSpecialistReviewContract(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	requireDocContract(t, root, "plan-wide specialist-review contract", map[string][]contractMatcher{
		".github/skills/kb-plan/SKILL.md": {
			docAnchor("document-review mode:headless"),
			docAnchor("pre_slice_review_contract: true"),
			docAnchor("pre_slice_review"),
			docAnchor("source_sha256"),
			docAnchor("review_artifact_sha256"),
			docAnchor("persona_evidence_json"),
			docAnchor("not_required_reason"),
			docConcept("main-agent requirements check"),
			docConcept("exactly one best-fit reviewer"),
		},
		".github/skills/document-review/SKILL.md": {
			docAnchor("`spec-flow-analyzer`"),
			docAnchor("docs/results/document-reviews/"),
			docConcept("optional uncertainty reducer"),
			docConcept("Never run always-on reviewers and never stack personas"),
			docConcept("Never review one slice at a time"),
			docConcept("reviewer is read-only"),
		},
		".github/skills/kb-brainstorm/SKILL.md": {
			docAnchor("document-review mode:headless"),
			docConcept("requirements self-check"),
			docConcept("exactly one best-fit reviewer"),
			docConcept("source SHA-256 still matches"),
			docConcept("do not dispatch placeholder personas"),
		},
		".github/skills/kb-work/SKILL.md": {
			docAnchor("pre_slice_review"),
			docConcept("not slice implementation owners"),
			docConcept("one new requirements-wide review"),
			docConcept("Do not rerun plan-wide specialist personas per slice"),
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			docAnchor("Pre-slice review receipt:"),
			docConcept("implementation owner, not a document-review persona"),
			docConcept("Do not rerun plan-wide specialist review"),
		},
	})
}
