---
kb_id: kb-2026-08-23-epistemic-investigation-gate
slice_id: epistemic-005
title: Enforce planning assurance after behavioral promotion
blockers: [epistemic-004]
verification: tdd
test_level: integration
functional_risk: broad
model_tier: large
model_tier_reason: This slice changes the manifest phase boundary and must preserve legacy resumability while rejecting stale or inconclusive new-plan assurance.
model_requirements: [Go contract design, backward compatibility, skill workflow design, hash-bound receipt validation]
escalation_triggers: [matched replay is not promote, legacy manifests fail, self-referential hashing appears, ordinary supported plans require user interaction]
workspace_mode: shared-serial
conflict_domains: [skill:kb-plan, skill:kb-work, namespace:manifest-contract, path:.github/skills/kb-plan/references]
shared_resources: [git:integration-owner, gate:plan-to-work]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run 'ManifestContract.*PlanningAssurance' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: Promote the successful loop into the new-plan output contract and require a hash-bound assurance sidecar before plan-to-work passes.
  - path: .github/skills/kb-plan/references/manifest-template.md
    op: edit
    scope: Document the new manifest schema and planning-assurance receipt binding.
  - path: .github/skills/kb-plan/references/planning-assurance.md
    op: create
    scope: Define the compact sidecar schema, evidence-delta rule, terminal states, and invalidation behavior.
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: Reject new-schema manifests with missing, stale, or inconclusive planning assurance while preserving legacy resume behavior.
  - path: cmd/kbcheck/manifest_contract.go
    op: edit
    scope: Validate the new-schema assurance sidecar, plan/source hashes, progress deltas, terminal state, and plan-to-work compatibility.
  - path: cmd/kbcheck/manifest_contract_test.go
    op: edit
    scope: Protect valid fast-path, investigated-and-revised, stalled-inconclusive, stale-hash, malformed, and legacy-manifest behavior.
protected_oracles:
  - path: cmd/kbcheck/manifest_contract_test.go
    role: Planning-assurance enforcement and backward-compatibility oracle.
    sha256: filled by kb-work after RED/protection
    update_policy: Protect after RED; later semantic changes require explicit plan amendment.
status: pending
owner: agent
blocked_reason: ""
resume_when: epistemic-004 returns promote with protected ease and epistemic metrics
next_agent_action: Protect manifest-contract tests, prove RED, add the smallest new-schema sidecar validator, then update kb-plan and kb-work contracts.
human_action: ""
can_continue_other_slices: false
---

# Planning Assurance Enforcement

## Deliverable

New KB plans cannot pass `plan-to-work` without a compact, hash-bound challenge
receipt. The receipt proves the planning loop ran against the exact requirements
and plan contents; it does not claim that a structural validator can judge the
semantic quality of every cited source.

## Acceptance Criteria

- The matched replay result is `promote`; otherwise this slice does not run.
- New-schema manifests bind one repo-relative assurance JSON sidecar by SHA-256.
- The sidecar binds the requirements source, manifest, and every slice plan by
  repo-relative path and content hash without self-referential manifest hashing.
- Each material premise records materiality, initial detection state, inspected
  evidence references, resolution, and whether the plan changed.
- Every repeated iteration records an evidence or plan delta. A repeated pass
  with no delta must terminate `inconclusive`.
- `plan-to-work: passed` requires assurance `passed`; `inconclusive` preserves
  the artifact but blocks work.
- Editing any bound source or plan invalidates the receipt.
- A supported fast-path plan may pass with zero investigation iterations when
  its load-bearing premises already cite adequate current evidence.
- Legacy manifests remain valid and resumable under their original schema.
- `kb-work` performs no duplicate semantic review; it validates the receipt and
  phase state, then executes normally.
- Deterministic tests cover false-positive investigation, contradicted-premise
  revision, insufficient evidence, stale hashes, and legacy compatibility.

## Scope Boundary

No live model call, eval-oracle change, global skill sync, push, PR, merge, or
claim that structural validation proves evidence relevance.
