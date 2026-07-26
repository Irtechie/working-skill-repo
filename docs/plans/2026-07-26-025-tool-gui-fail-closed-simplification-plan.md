---
kb_id: kb-2026-07-26-change-aware-proof-governor
slice_id: slice-005
title: "Remove repo-owned GUI approvals and fail automatic GUI execution closed"
blockers: [slice-003]
verification: tdd
test_level: functional-cli
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "This correction removes a misleading security boundary from the public CLI while preserving the process-launch safety invariant across code, manifests, skills, and installed copies."
model_requirements: ["Go CLI contract repair", "process-launch safety", "cross-skill policy consistency", "dirty-worktree preservation"]
escalation_triggers: ["a visible-browser or native-gui check can reach the runner", "removing approval code weakens headless execution or receipt reuse", "a target skill contains useful target-only drift", "focused proof requires a real GUI"]
context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-005.json
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'ProofExecution|ProofGovernorCLI|ManifestContract' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/proof_execution_test.go
    op: edit
    scope: "Replace approval-token scenarios with a protected oracle proving visible-browser and native-gui checks are blocked before runner invocation."
  - path: cmd/kbcheck/proof_execution.go
    op: edit
    scope: "Delete approval minting, loading, validation, and consumption; deny GUI-capable execution classes before process launch."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Remove proof-approve and its approval-specific flags from the public CLI."
  - path: cmd/kbcheck/proof_governor_selftest.go
    op: edit
    scope: "Use the simplified execution API and retain zero-GUI-launch proof."
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Stop requiring approval_required and keep execution_class as the GUI safety classifier."
  - path: cmd/kbcheck/swarm.go
    op: edit
    scope: "Remove approval_required from parsed manifest slice state."
  - path: config/skill-quality.json
    op: edit
    scope: "Describe visible-browser and native-gui as blocked automatic execution classes."
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Plan execution class without repo-owned approval fields."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Fail GUI-capable automatic proof closed and leave any attended launch to the host/user outside proof-run."
  - path: .github/skills/kb-functional-test/SKILL.md
    op: edit
    scope: "Prefer headless proof and classify GUI-only verification as externally attended rather than token-approved."
  - path: .github/skills/kb-qa/SKILL.md
    op: edit
    scope: "Remove the approval-token contract while retaining pre-launch denial."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document automatic GUI launch prevention and the external attended boundary."
  - path: README.md
    op: edit
    scope: "Rename proof-governor GUI behavior without expanding workflow ceremony."
  - path: docs/plans/2026-07-26-020-kb-change-aware-proof-governor-manifest.md
    op: edit
    scope: "Record this superseding slice and remove current approval_required/attended-approval claims."
  - path: docs/results/2026-07-26-proof-governor-slice-005.md
    op: create
    scope: "Record focused proof, zero GUI launches, sync status, and the unchanged final-release blocker."
protected_oracles:
  - path: cmd/kbcheck/proof_execution_test.go
    role: "automatic GUI execution denial before process launch"
    sha256: "80672702ed84bf6c7ffffdf706af8966df11c7b6d2031c966bdd6779673dc03e"
    oracle_update_reason: "Explicit corrective amendment supersedes the slice-003 approval-token oracle because the token was not an authenticated trust boundary and added unwanted HITL ceremony. RED proved both GUI classes still returned the superseded attended-approval reason while keeping runner count zero; the accepted oracle then added public CLI removal assertions before final protection."
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 005: Automatic GUI Fail-Closed Simplification

## What to Build

Keep the change-aware proof governor's receipts, invalidation, reuse, replay
budget, failure replay protection, timeout handling, and audit trail. Remove the
repo-owned approval document and command. `proof-run` must never start
`visible-browser` or `native-gui` checks; it returns `BLOCK` before calling the
process runner.

An explicitly attended GUI test, if ever needed, belongs to the user/host
session outside this portable proof runner. This repo records the blocker but
does not create another approval workflow.

## Acceptance Criteria

1. `proof-run` blocks both GUI-capable execution classes before runner
   invocation and emits `automatic-gui-execution-disabled:<check-id>`.
2. `proof-approve`, `--approval`, `--approval-ttl`, and approval JSON/marker
   handling no longer exist in the CLI or execution package.
3. CLI, headless-browser, proof reuse, changed-input invalidation, timeout, and
   repeated-failure behavior remain unchanged.
4. Manifest slices use `execution_class` without `approval_required`; the
   contract still requires `functional-native-gui` to declare `native-gui`.
5. Current skills and testing docs describe automatic GUI denial and do not
   claim an authenticated attended-approval mechanism.
6. Focused tests use fake runners and launch no browser, WebView, Tauri, WDIO,
   or desktop process.
7. Only changed target skills are synchronized after reviewing all three
   installed copies for target-only drift.
8. The known `cmd/kbrouter` and unrelated skill-sync blockers remain recorded;
   no identical full `core` or `local-release` replay is performed.

## Test Scenarios

- A `visible-browser` registry check returns `BLOCK`, reports the check ID, and
  leaves runner count at zero.
- A `native-gui` registry check does the same.
- Passing CLI checks still produce receipts and later reuse them.
- Approval-looking CLI flags are rejected as unknown, and `proof-approve` is no
  longer a known command.
- A `functional-native-gui` slice with a non-native execution class is still
  rejected by the manifest contract.

## Scope Boundary

This changes only the portable proof runner and its current policy surfaces. It
does not alter independent model-route approvals, paid-call approvals,
deployment gates, or other domain-specific HITL controls.

## Proof

`go test ./cmd/kbcheck -run 'ProofExecution|ProofGovernorCLI|ManifestContract' -count=1`
