---
type: kb-slice-plan
kb_id: kb-2026-07-26-delegation-first-ddr
slice_id: ddr-001
title: "Enforce delegation-first ownership with bounded current exceptions"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: narrow
execution_class: cli
model_tier: medium
model_tier_reason: "The behavior is bounded to an existing Go selector and documented CLI contract, with deterministic tests."
model_requirements: ["Go implementation", "CLI contract reasoning", "focused deterministic tests"]
escalation_triggers: ["compatibility callers require unbounded reason text", "native-host and CLI catalogs cannot preserve distinct authority"]
proof_check:
  kind: command_exit
  command: "go test ./internal/modelrouting ./cmd/kbrouter ./cmd/kbcheck"
  expect: 0
hitl: false
expected_files:
  - path: internal/modelrouting/selector.go
    op: edit
    scope: "validate current-owner exception reasons and prove no-qualified-route against eligible routes"
  - path: internal/modelrouting/selector_test.go
    op: edit
    scope: "protect recognized-reason and no-qualified-route behavior"
  - path: internal/modelrouting/policy.go
    op: edit
    scope: "use OS-specific existing-path canonicalization for project identity"
  - path: internal/modelrouting/identity_windows.go
    op: edit
    scope: "resolve Windows junctions through a filesystem handle"
  - path: internal/modelrouting/identity_unix.go
    op: edit
    scope: "retain EvalSymlinks behavior behind the shared OS helper"
  - path: internal/modelrouting/identity_windows_test.go
    op: create
    scope: "prove junction parity and missing-path rejection"
  - path: cmd/kbrouter/select.go
    op: edit
    scope: "make the current-owner reason gate visible in CLI help"
  - path: cmd/kbrouter/select_test.go
    op: edit
    scope: "prove the CLI rejects invalid current-owner claims"
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "make qualified subagent execution the default and define current-owner exceptions"
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "separate orchestrator planning from normal subagent execution"
  - path: cmd/kbcheck/ddr_contract_test.go
    op: edit
    scope: "make delegation-first and exception-gate language executable"
  - path: README.md
    op: edit
    scope: "describe the user-visible routing contract"
  - path: docs/context/architecture/kbrouter.md
    op: edit
    scope: "document selector enforcement and native-host boundary"
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "document the delegation-first flow"
protected_oracles:
  - path: internal/modelrouting/selector_test.go
    role: "owner-selection behavior oracle"
    update_policy: "add cases before implementation; do not weaken existing capability checks"
  - path: cmd/kbcheck/ddr_contract_test.go
    role: "portable skill-policy oracle"
    update_policy: "require the new policy phrases and retain existing AMR exclusions"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Enforce delegation-first ownership

## Build

Keep the orchestrator responsible for decomposition, minimum-tier judgment,
worker selection, supervision, proof, and synthesis. Make one qualified
subagent the normal execution owner after the work is bounded. This is one
owner per slice, not one worker per plan: dispatch all safe independent ready
slices in parallel when their dependencies, writes, and shared resources are
isolated. Keep the plan portable: the pickup host maps each slice's complexity
tier to an exact live qualified model and records that choice only in the
runtime receipt.

Permit current execution only for a recognized reason:

- `reasoning-required`
- `context-required`
- `tool-required`
- `authority-required`
- `trust-required`
- `user-required`
- `no-qualified-route`

Allow a short explanation after the reason code. Require
`no-qualified-route` to inspect the live eligible CLI routes, while the skill
also requires the orchestrator to inspect the active host's callable surface.

## Test scenarios

1. A bounded delegated request selects exactly one qualified worker.
2. A current request with an unknown or vague reason is rejected.
3. A recognized reasoning/context/authority exception may retain a qualified
   current orchestrator.
4. `no-qualified-route` is rejected when an eligible route exists.
5. `no-qualified-route` may retain a qualified current orchestrator when no
   eligible route exists.
6. Downward routing and silent cross-owner fallback remain forbidden.
7. Multiple independent ready slices may select and run separate tier-qualified
   subagents in parallel.
8. Project identity is stable across a Windows junction and canonical path;
   nonexistent roots still fail closed.

## Scope boundary

Do not promote AMR, merge native-host identities into the CLI catalog, or make a
route receipt substitute for deterministic work proof.
