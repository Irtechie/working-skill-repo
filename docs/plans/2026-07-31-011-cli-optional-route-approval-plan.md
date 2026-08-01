---
kb_id: kb-2026-07-31-optional-route-approval
slice_id: slice-001
title: "Let users disable attended route approval without disabling router safety"
blockers: []
verification: tdd
test_level: functional-cli
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "This changes a trust boundary across catalog persistence, policy evaluation, endpoint/auth validation, dispatch fallback, and DDR replay checks."
model_requirements: ["security-boundary reasoning", "Go CLI integration", "backward-compatible state migration", "cross-surface documentation"]
escalation_triggers: ["disabled mode bypasses explicit denials", "static endpoint safety is weakened", "required mode no longer enforces attended approval", "DDR can redispatch or skip proof"]
workspace_mode: shared-serial
conflict_domains: ["cmd:kbrouter", "internal:modelrouting", "skill:kb-models", "docs:model-routing", "install:user-router"]
shared_resources: ["user-state:~/.kb/models.json", "binary:~/.kb/bin/kbrouter.exe", "install:global-kb-models"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbrouter ./internal/modelrouting -run 'ApprovalMode|OptionalRouteApproval' -count=1"
  expect: 0
hitl: false
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "No implementation work remains; delivery is owned by the parent manifest."
human_action: ""
can_continue_other_slices: true
protected_oracles:
  - path: cmd/kbrouter/catalog_test.go
    purpose: "Prove the public approval-mode command, no-prompt default, required-mode opt-in, selectable route behavior, and explicit-denial precedence."
  - path: cmd/kbrouter/ddr_test.go
    purpose: "Prove DDR accepts an unapproved configured route only in disabled mode and still retains bounded attempt/proof behavior."
expected_files:
  - path: internal/modelrouting/catalog.go
    op: edit
    scope: "Persist the user-local approval mode in the existing catalog schema."
  - path: internal/modelrouting/policy.go
    op: edit
    scope: "Centralize approval-required semantics while preserving denial and non-approval policy checks."
  - path: internal/modelrouting/storage.go
    op: edit
    scope: "Reject unknown approval modes and keep non-user catalogs free of local preference state."
  - path: cmd/kbrouter/main.go
    op: edit
    scope: "Add the noninteractive models approval-mode command."
  - path: cmd/kbrouter/catalog.go
    op: edit
    scope: "Load the mode into project policy and apply it to selection/auth checks."
  - path: cmd/kbrouter/ddr.go
    op: edit
    scope: "Apply disabled mode to initial and just-in-time DDR trust checks."
  - path: cmd/kbrouter/dispatch.go
    op: edit
    scope: "Apply disabled mode to less-trusted fallback approval checks."
  - path: cmd/kbrouter/catalog_test.go
    op: edit
    scope: "Protect CLI persistence, default behavior, denial precedence, and policy bypass."
  - path: cmd/kbrouter/ddr_test.go
    op: edit
    scope: "Protect real DDR eligibility under disabled mode."
  - path: .github/skills/kb-models/SKILL.md
    op: edit
    scope: "Stop treating attended approval as universally mandatory and document the user-owned mode."
  - path: docs/context/architecture/kbrouter.md
    op: edit
    scope: "Document mode semantics and retained controls."
  - path: LOCAL_MODELS.example.md
    op: edit
    scope: "Show required and disabled setup paths."
  - path: README.md
    op: edit
    scope: "Expose the visible approval-mode workflow."
  - path: todo.md
    op: edit
    scope: "Track the active repair and replace the stale blocker after installation."
---

# Slice 001 - Optional Route Approval

## Acceptance Criteria

- `kbrouter models approval-mode --mode required|disabled` writes only canonical
  user-local route state and requires no attended confirmation.
- Missing mode means `disabled`, so ordinary bounded routing does not stop for
  endpoint/auth permission.
- `required` remains available as an explicit opt-in and retains the attended,
  project-bound fingerprint confirmation.
- `disabled` permits configured routes without route, endpoint, or auth approval
  receipts, including DDR's just-in-time revalidation.
- Explicit route denials, endpoint scheme/metadata/private-boundary checks,
  destination/retention/sensitive-data policy, DDR reservation, one-attempt
  behavior, and deterministic proof remain enforced.
- `models show` and `models doctor` expose the effective redacted mode.
- The installed router uses `disabled` for this user and the existing
  `deepseek-local` route is eligible without `trust.json`.

## Test Scenarios

1. A catalog with no mode permits an otherwise eligible unapproved extra route.
2. The CLI persists `disabled`, preserves routes, and makes the unapproved route
   selectable without creating or mutating `trust.json`.
3. A persisted explicit denial still blocks the same route in disabled mode.
4. The CLI persists `required`, and the same unapproved route is blocked until
   attended approval exists.
5. Unknown modes fail validation and do not corrupt the catalog.
6. DDR reaches its bounded endpoint attempt path in disabled mode without a
   trust receipt; required mode still returns the existing untrusted result.
7. Targeted tests, `kbcheck core`, `kbcheck local-release`, install, catalog
   inspection, `git diff --check`, PR delivery, merge containment, and global
   Codex/Copilot/shared-agent skill synchronization pass.

## Scope Boundary

Do not remove attended approval, disable explicit denials, weaken endpoint
static safety, relax data policy, permit multiple DDR attempts, or skip
deterministic proof.
