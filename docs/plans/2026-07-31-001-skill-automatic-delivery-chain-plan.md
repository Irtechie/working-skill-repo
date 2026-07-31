---
kb_id: kb-2026-07-31-automatic-delivery-chain
slice_id: slice-001
title: "Make successful KB work continue through authorized delivery"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "The slice must preserve owner-specific publishing and integration boundaries while removing every successful terminal handoff between work, finalization, completion, shipping, landing, and sync."
model_requirements: ["cross-skill contract reasoning", "GitHub delivery safety", "deterministic Go contract tests", "release and sync workflow knowledge"]
escalation_triggers: ["a phase would gain another phase's mutation authority", "required checks or reviews cannot be observed", "the active PR cannot target resolved remote default", "installed drift contains newer useful work"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-work", "skill:kb-finalize", "skill:kb-complete", "skill:kb-ship", "skill:kb-land", "git:default-branch", "docs:kb-workflow"]
shared_resources: ["git:integration-owner", "github:pull-request", "install:global-skill-roots"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'AutomaticDeliveryChain|DeliveryOwnerSkillContracts' -count=1"
  expect: 0
hitl: false
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Run kb-finalize and configured delivery."
human_action: ""
can_continue_other_slices: true
protected_oracles:
  - path: cmd/kbcheck/delivery_chain_contract_test.go
    purpose: "Mechanically requires every automatic transition and every preserved safety boundary."
expected_files:
  - path: cmd/kbcheck/delivery_chain_contract_test.go
    op: create
    scope: "Require the complete automatic chain, policy selection, owner boundaries, main-containment proof, and post-land sync."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Automatically invoke kb-finalize and then return to kb-complete without acquiring default-branch mutation authority."
  - path: .github/skills/kb-finalize/SKILL.md
    op: edit
    scope: "Automatically invoke kb-complete after complete-to-ship passes instead of ending at a manual handoff."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Own the uninterrupted state loop, invoke kb-ship for PR delivery, and invoke authorized kb-land after ship gates."
  - path: .github/skills/kb-ship/SKILL.md
    op: edit
    scope: "Return shipped PR evidence to kb-complete while retaining the never-merge boundary."
  - path: .github/skills/kb-land/SKILL.md
    op: edit
    scope: "Retain sole remote-default integration authority and require remote containment before source-to-installed sync."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Replace the manual terminal handoff with the automatic gated phase sequence."
  - path: docs/context/architecture/skills.md
    op: edit
    scope: "Describe the completion and delivery group as one automatically continued workflow."
  - path: README.md
    op: edit
    scope: "Document the visible source-to-landed workflow and safety boundary."
  - path: evals/route-complexity/finish-plan-flow.json
    op: edit
    scope: "Expect authorized land and sync artifacts for an end-to-end completion request."
  - path: evals/route-complexity/release-ship-flow.json
    op: edit
    scope: "Remove the contradictory claim that kb-complete ended before checked-in delivery."
  - path: evals/skill-eval/selftest/pass-finish-plan-flow.json
    op: edit
    scope: "Trace kb-land and remote-default containment through the successful flow."
  - path: docs/plans/2026-07-09-020-kb-plan-to-pr-finish-manifest.md
    op: edit
    scope: "Mark the stale active plan superseded and remove its obsolete non-shipping terminal rule."
  - path: docs/plans/2026-07-09-022-kb-finish-orchestrator-plan.md
    op: edit
    scope: "Remove the obsolete kb-finish/non-shipping acceptance criterion."
  - path: todo.md
    op: edit
    scope: "Track this active manifest and supersede the old plan-to-PR workstream."
  - path: todo-done.md
    op: edit
    scope: "Record the completed source workflow after exact-tree proof."
---

# Slice 001 - Automatic Authorized Delivery

## Acceptance Criteria

- Successful `kb-work` automatically invokes `kb-finalize`; successful
  finalization automatically invokes or returns control to `kb-complete`.
- `kb-complete` re-reads durable state and applies delivery policy without a
  manual terminal handoff.
- PR delivery invokes `kb-ship`; `kb-ship` commits, pushes, and opens or updates
  the correctly based PR but never merges.
- `kb-complete` invokes `kb-land` only for configured or same-run authorized
  direct/auto-merge delivery after ship proof and required checks/reviews.
- Only `kb-land` integrates remote default. It proves fetched remote-default
  containment before configured post-integration sync.
- Missing policy remains local-only. No phase force-pushes, bypasses branch
  protection, required checks, hooks, or required reviews.
- The stale active plan-to-PR contract no longer claims that ordinary successful
  completion stops before shipping.
- Source and installed global copies are compared before overwrite; final
  installed `kb-work` hashes match the landed source.

## Test Scenarios

1. Static contract proof finds every required transition in its owning skill.
2. Static contract proof fails if `kb-work` gains default-branch integration
   authority, `kb-ship` gains merge authority, or `kb-land` loses sole ownership.
3. Route fixtures model plan-to-land completion through `kb-complete`, including
   PR, remote-default containment, and installed sync evidence.
4. Targeted Go tests, `kbcheck core`, and `kbcheck local-release` pass on the
   exact reviewed tree.
5. After merge, fetched `origin/main` contains the delivered commit and all three
   installed `kb-work/SKILL.md` copies match the landed source hash.

## Scope Boundary

Do not add a new orchestrator skill, grant `kb-work` or `kb-ship` direct default
mutation rights, weaken local-only defaults, bypass repository protections, or
sync installed copies before remote-default integration is proven.
