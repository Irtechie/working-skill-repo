---
kb_id: kb-2026-07-26-blocker-lifecycle-contract
slice_id: slice-031
title: "Enforce blocker lifecycle metadata and stale rechecks"
blockers: []
verification: tdd
test_level: integration
functional_risk: narrow
model_tier: medium
model_tier_reason: "The parser and validator affect every opt-in KB manifest and must remain backward compatible."
model_requirements: ["Go parser changes", "negative contract fixtures", "portable Python parity", "backward compatibility"]
escalation_triggers: ["legacy manifests become invalid without opting in", "release-only scope can block implementation", "stale timestamps pass", "Python and Go disagree"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'TestBlockerLifecycle|TestManifestContract' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: "Parse and validate opt-in gate scope, responsibility, recheck, freshness, and propagation metadata."
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: "Protect valid pause/human/external states and reject stale or over-propagated blockers."
  - path: .github/skills/kb-gate/scripts/check_gate_ledger.py
    op: edit
    scope: "Keep the portable gate checker aligned with the Go lifecycle contract."
  - path: .github/skills/kb-gate/references/gate-ledger.md
    op: edit
    scope: "Document the opt-in gate fields and examples."
  - path: .github/skills/kb-gate/scripts/check_gate_ledger_test.py
    op: add
    scope: "Protect portable parser parity, duplicate rejection, strict timestamps, and quoted list values."
status: done
owner: agent
---

# Enforce Blocker Lifecycle Metadata and Stale Rechecks

## Acceptance Criteria

- Preserve legacy manifests unless they opt into
  `blocker_lifecycle_contract: true`.
- Reject `paused` as a gate result. Preserve the current technical gate state
  and use `paused` only for execution or durable-goal state.
- Require each opt-in gate to declare its scope.
- Require blocked and needs-human gates to identify responsibility, affected
  scope, exact resume condition, recheck sensor, last check time, and
  propagation.
- Reject `needs-human` owned by the agent and `blocked` owned by the human.
- Reject release, deployment, or optional-capability blockers that propagate
  beyond their current gate.
- Reject stale nonterminal gate claims after 72 hours.
- Reject duplicate gate IDs, malformed lifecycle booleans, stale `passed_at`,
  underspecified quarantine, and ambiguous date-only blocker checks.
- Preserve quoted commas and list entries containing colons in both parsers.
- Keep the Go and portable Python checkers behaviorally aligned.

## Scope Boundary

Do not weaken proof, safety, credential, deployment, signing, destructive
action, or subjective acceptance gates. This slice changes classification and
scope, not whether required evidence exists.
