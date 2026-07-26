---
kb_id: kb-2026-07-26-change-aware-proof-governor
slice_id: slice-003
title: "Enforce replay budgets and automatic GUI pre-launch denial"
blockers: [slice-002]
verification: tdd
test_level: functional-cli
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "This slice controls process launches and automatic GUI denial, where advisory wording cannot prevent disruptive loops."
model_requirements: ["process execution policy", "automatic GUI launch prevention", "PowerShell and Go integration", "cross-skill workflow consistency"]
escalation_triggers: ["a visible or native GUI can launch from automatic proof", "the same full suite can rerun unchanged without an override", "raw snapshot verify bypasses selection", "repair still requires every unrelated check after each fix"]
context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-003.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'ProofExecutionBudget|AttendedGUI|SnapshotSelection|RepairPolicy' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/proof_execution.go
    op: create
    scope: "Enforce selection decisions, replay ceilings, timeouts, and attended visible/native GUI receipts."
  - path: cmd/kbcheck/proof_execution_test.go
    op: create
    scope: "Prove redundant full replay and unattended GUI denial without launching a GUI."
  - path: .github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1
    op: edit
    scope: "Replace unconditional all-snapshot replay with governed requested/invalidated selection."
  - path: .github/skills/kb-regression-snapshot/SKILL.md
    op: edit
    scope: "Require change-aware selection and milestone-only full replay."
  - path: .github/skills/kb-repair/SKILL.md
    op: edit
    scope: "Rerun failed and impacted checks rather than every unrelated check after each fix."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Consume governed proof receipts and stop unconditional coordinator reproof."
  - path: .github/skills/kb-finalize/SKILL.md
    op: edit
    scope: "Deduplicate pre/post-review child proof while preserving changed-workflow and aggregate finalization checks."
  - path: .github/skills/kb-ship/SKILL.md
    op: edit
    scope: "Run one fresh release aggregate and reuse identical child proof inside the same release state."
  - path: .github/skills/kb-qa/SKILL.md
    op: edit
    scope: "Require headless defaults and deny automatic visible/native execution before launch."
  - path: .github/skills/kb-check/SKILL.md
    op: edit
    scope: "Route repeatable checks through proof planning and execution."
  - path: .github/skills/kb-functional-test/SKILL.md
    op: edit
    scope: "Add native-GUI classification and align focused, manifest, and ship timing with receipt reuse and changed-input invalidation."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Allow functional-native-gui test classification and require execution-class planning."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Validate functional-native-gui and execution-class fields."
protected_oracles:
  - path: cmd/kbcheck/proof_execution_test.go
    role: "historical redundant replay refusal and GUI denial oracle; approval behavior superseded by slice-005"
    sha256: "de181281c6cc6f50e5c4b1dc085d14f2eb8ba40ad74ea185064a361c2ee95cea"
    oracle_update_reason: "Completed the oracle before final protection with single-use approval and bounded audit assertions required by acceptance criteria 5 and 6."
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: false
---

# Slice 003: Governed Execution and GUI Safety

## What to Build

Make the selector operational in the canonical KB runner. The same full suite
against the same relevant fingerprint may pass once and then becomes reusable;
an unchanged replay request is blocked. Failed focused checks may rerun after a
relevant code change. A same-input diagnostic retry is bounded and explicitly
classified rather than becoming a loop.

## Acceptance Criteria

1. A passing full suite runs at most once per exact relevant fingerprint and
   milestone; later subset requests reuse it.
2. A code change invalidates every affected check and permits those checks to
   run again; unrelated checks retain valid receipts.
3. Repair reruns the failed check and registered impact dependents. It does not
   rerun all lint/browser/snapshot checks merely because one fix occurred.
4. Snapshot verification accepts requested/invalidated IDs and no longer scans
   every prior JSON file unconditionally before each slice.
5. Headless browser proof remains agent-runnable. Automatic visible browser and
   native GUI execution fail closed before process launch; any explicitly
   attended session remains outside the portable proof runner.
6. Every process launch and reuse decision appends a bounded audit row so a
   human can see what changed and why a rerun occurred.
7. Canonical tests prove denial using fake commands; implementation proof does
   not launch WebView, Tauri, WDIO, or another visible desktop process.
8. Each check and aggregate run has a configured timeout, bounded output,
   child-process cleanup, and a terminal partial/failure receipt. Timeout uses a
   stable nonzero code such as 124 and cannot be reported as global proof.
9. A batch of headless DOM assertions reuses one browser process/context where
   isolation permits instead of opening and closing Chromium per assertion.
10. `functional-native-gui` is distinct from browser proof and requires
    executable public-surface assertions plus app-build/session fingerprints;
    screenshots alone never satisfy reuse.
11. Repeated identical failure stops immediately; progress-sensitive repair
    still has a hard attempt ceiling and emits a terminal reason.

## Test Scenarios

- Twenty-five-check suite passes once; twenty-five unchanged replay requests are
  reused or blocked without process creation.
- A Rust source edit reruns registered Rust/CLI dependents once while reusing
  unchanged browser/checksum proof.
- A shared UI dependency invalidates its browser checks, which run headless by
  default.
- Visible/native requests are denied before spawning a process.
- A hanging CLI check times out, kills its process tree, writes partial evidence,
  and prevents a global-pass claim.
- Multiple headless DOM checks use one bounded browser batch.
- The existing “rerun ALL checks” and “all previous snapshots” policy phrases
  fail a workflow contract selftest until replaced.

## Scope Boundary

This governs canonical KB proof paths. Host-global interception of arbitrary
shell commands remains adapter-specific; the portable bundle must not claim it
can prevent a user or unrelated tool from bypassing the repo runner.

## Proof

`go test ./cmd/kbcheck -run 'ProofExecutionBudget|AttendedGUI|SnapshotSelection|RepairPolicy' -count=1`
