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

func TestWebUIProofRemainsAgentOwned(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	required := map[string][]string{
		".github/skills/kb-qa/SKILL.md": {
			"default for local/public UI functional checks",
			"Do not ask the user whether to run browser proof",
			"Run browser proof headless automatically unless the user explicitly requests",
			"record one proof receipt containing the route",
			"do not report partial completion",
			"automation gap into manual verification",
		},
		".github/skills/kb-functional-test/SKILL.md": {
			"Do not ask the user whether to run browser proof",
			"Run browser proof headless automatically unless the user explicitly requests",
			"A web UI check is not complete until its receipt names the route",
			"evidence is agent-owned test work",
		},
	}
	for relative, phrases := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, phrase := range phrases {
			if !strings.Contains(string(content), phrase) {
				t.Errorf("%s missing web UI proof contract %q", relative, phrase)
			}
		}
	}
}

func TestPlanRunWorktreeAndBranchShareFunnyTaskName(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	required := map[string][]string{
		".github/skills/kb-work/SKILL.md": {
			"Use the same codename for the worktree directory and plan-run branch",
			"must relate recognizably to the task",
		},
		".github/skills/kb-work/references/worktree-isolation.md": {
			"Branch and worktree basename must share that exact codename",
			"`codex/the-reviewers-have-unionized`",
		},
	}
	for relative, phrases := range required {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, phrase := range phrases {
			if !strings.Contains(string(content), phrase) {
				t.Errorf("%s missing worktree naming contract %q", relative, phrase)
			}
		}
	}
}

func TestDeliveryOwnerSkillContracts(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	required := map[string][]string{
		".github/skills/kb-start/SKILL.md": {
			"terminal-cleanup --action sweep",
			"current executing session",
		},
		".github/skills/kb-work/SKILL.md": {
			"never merges or pushes a resolved default branch",
			"never delivers under any policy",
			"local leases are not team locks",
		},
		".github/skills/kb-complete/SKILL.md": {
			"reviewed manifest-owned plan-run branch is the only delivery candidate",
			"Absent policy is PR/manual",
			"no policy authorizes merge",
			"Sync now, or wait for PR review?",
			"terminal-cleanup --action register",
			"release the shared work claim",
		},
		".github/skills/kb-ship/SKILL.md": {
			"reviewed plan-run topic branch is the only shipping candidate",
			"PR/manual stops with the correctly based open PR",
			"`kb-ship` never merges it",
			"remote topic contains the delivered commit",
		},
		".github/skills/kb-land/SKILL.md": {
			"only KB skill authorized to integrate the resolved remote default branch",
			"Absence of delivery policy never authorizes landing",
			"remote default contains the delivered commit",
		},
		".github/skills/kb-configure/SKILL.md": {
			"mode: pr",
			"PR/manual is the default",
			"accepting a PR never is",
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

func TestCargoBuildStorageContract(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	required := map[string][]string{
		".github/skills/kb-check/SKILL.md": {
			"cargo-storage --action resolve",
			"cargo-storage --action validate-ready",
			"cargo-storage --action not-applicable",
			"capability probe confirms the project's `cmd/kbcheck help` output includes `cargo-storage`",
			"absent or predates `cargo-storage`",
			"treat it as a cache root, append the first 24 lowercase hex characters of SHA-256",
			"If the configured path already ends with that exact project key, reuse it",
			"Temporary targets and automated deletion are prohibited in fallback mode",
		},
		".github/skills/kb-fix/SKILL.md": {
			"native `cargo-storage resolve` receipt validated by `validate-ready`",
			"Preserve the exact `CARGO_TARGET_DIR` through every fix and verification attempt",
			"target-repro",
		},
		".github/skills/kb-troubleshoot/SKILL.md": {
			"native `cargo-storage resolve` receipt accepted by `validate-ready`",
			"Reuse its exact `CARGO_TARGET_DIR` for diagnosis, fixes, retries, probes, and final verification",
			"target-repro",
		},
		".github/skills/kb-repair/SKILL.md": {
			"native validated receipt or fail-closed portable fallback contract",
			"apply its exact `CARGO_TARGET_DIR`",
			"target-repair",
		},
		".github/skills/kb-work/SKILL.md": {
			"receipt accepted by `cargo-storage --action validate-ready`",
			"Apply the returned exact `CARGO_TARGET_DIR` across every slice, worker, repair, proof batch, session, and worktree",
			"A temporary target must be created first in native mode",
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			"Build-storage contract:",
			"require a native receipt accepted by `cargo-storage --action validate-ready` or a `kb-check` portable-fallback record",
			"Do not replace it with a phase-, worker-, or run-specific target",
		},
		".github/skills/kb-finalize/SKILL.md": {
			"cargo-storage --action finalize",
			"retained_bytes",
			"removed_bytes",
			"temporary targets and deletion were prohibited",
		},
		".github/skills/kb-complete/SKILL.md": {
			"cargo-storage --action validate",
			"done-portable-fallback",
			"Recompute the canonical Git common-directory identity and its 24-hex project key",
			"`removed_bytes` to be zero, and no temporary target entries",
			"not-applicable",
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
				t.Errorf("%s missing Cargo build-storage contract %q", relative, token)
			}
		}
	}
}

func TestPlanWideSpecialistReviewContract(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	required := map[string][]string{
		".github/skills/kb-plan/SKILL.md": {
			"document-review mode:headless",
			"perform the main-agent requirements check",
			"selects exactly one best-fit reviewer",
			"pre_slice_review_contract: true",
			"pre_slice_review",
			"source_sha256",
			"review_artifact_sha256",
			"persona_evidence_json",
			"not_required_reason",
		},
		".github/skills/document-review/SKILL.md": {
			"optional uncertainty reducer",
			"Never run always-on reviewers and never stack personas",
			"Never review one slice at a time",
			"`spec-flow-analyzer`",
			"The reviewer is read-only",
			"docs/results/document-reviews/",
		},
		".github/skills/kb-brainstorm/SKILL.md": {
			"First perform the requirements self-check",
			"selects exactly one best-fit reviewer",
			"document-review mode:headless",
			"source SHA-256 still matches",
			"do not dispatch placeholder personas",
		},
		".github/skills/kb-work/SKILL.md": {
			"pre_slice_review",
			"document-review personas are not slice implementation owners",
			"return to `kb-plan` for one new requirements-wide review",
			"Do not rerun plan-wide specialist personas per slice",
		},
		".github/skills/kb-work/references/execution-prompt.md": {
			"Pre-slice review receipt:",
			"You are an implementation owner, not a document-review persona",
			"Do not rerun plan-wide specialist review",
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
				t.Errorf("%s missing plan-wide specialist-review contract %q", relative, token)
			}
		}
	}
}
