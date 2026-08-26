---
type: kb-manifest
manifest_schema: 3
kb_id: kb-2026-08-26-kb-rehab-outstanding-work
brainstorm_path: docs/brainstorms/2026-08-26-kb-rehab-outstanding-work-requirements.md
created: 2026-08-26
status: reviewed
workflow_shape: pipeline-change
scope-verified-files:
  - .github/skills/kb-rehab/SKILL.md
  - .github/skills/kb-start/SKILL.md
  - .github/skills/todo-triage/SKILL.md
  - README.md
  - cmd/kbcheck/main.go
  - cmd/kbcheck/terminal_cleanup.go
  - cmd/kbcheck/work_reality.go
  - cmd/kbcheck/work_reality_test.go
  - cmd/kbcheck/rehab_delivery.go
  - cmd/kbcheck/rehab_delivery_test.go
  - config/rehab-policy.json
  - docs/brainstorms/2026-08-26-kb-rehab-outstanding-work-requirements.md
  - docs/context/PROJECT.md
  - docs/context/architecture/kbcheck.md
  - docs/context/architecture/kb-workflow.md
  - docs/plans/2026-08-26-000-kb-rehab-outstanding-work-manifest.md
  - docs/plans/2026-08-26-001-work-reality-report-plan.md
  - docs/plans/2026-08-26-002-rehab-lane-triage-plan.md
  - docs/plans/2026-08-26-003-rehab-granted-delivery-plan.md
  - docs/results/document-reviews/kb-rehab-outstanding-work-requirements-0d399da7ecfb.json
  - internal/reconcile/git.go
  - internal/reconcile/inventory.go
  - internal/reconcile/policy.go
  - todo.md
objective_contract: true
blocker_lifecycle_contract: true
pre_slice_review_contract: true
model_tier_contract: true
workspace_isolation_contract: true
proof_governor_contract: true
source_requirements_sha256: 0d399da7ecfba8b4f39e9e20be013682d82571304bb208ce0a0ae019664e141d
pre_slice_review:
  status: passed
  source: docs/brainstorms/2026-08-26-kb-rehab-outstanding-work-requirements.md
  source_sha256: 0d399da7ecfba8b4f39e9e20be013682d82571304bb208ce0a0ae019664e141d
  mode: requirements-wide
  review_id: kb-rehab-outstanding-work-requirements-0d399da7ecfb
  reviewed_at: "2026-08-26T19:05:00Z"
  review_artifact: docs/results/document-reviews/kb-rehab-outstanding-work-requirements-0d399da7ecfb.json
  review_artifact_sha256: eef0368d9bedda4984310758b43b559a01d765784732af9272e3112ee9d487d2
  persona_evidence_json: '{"security-lens-reviewer":"security-risk: an explicit user grant to auto-merge to a default branch and auto-delete refs in a repository that propagates to three global agent install roots makes proof-command provenance, adapter-absence polarity, containment authority, delivery-state eligibility, and grant binding a new privileged trust boundary."}'
  selected_personas_json: '["security-lens-reviewer"]'
  completed_personas_json: '["security-lens-reviewer"]'
  failed_personas_json: '[]'
  findings_resolved: 18
  unresolved_p0: 0
  unresolved_p1: 0
  residual_findings: 2
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck core"
  expect: 0
  why: "Proves the new work-reality pairing, the kb-rehab lane contract, granted-delivery predicate inheritance, existing terminal-cleanup safety, and documentation gates together."
model_tier_contract:
  allowed: [medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  default_owner: delegated
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  automatic_cross_owner_fallback: false
  amr_required: false
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
plan_run_worktree:
  branch: codex/w2d-unplanned-entry
  workspace_mode: shared-serial
  commit_authorized: true
  commit_authorized_by: user
  commit_approval_ref: "2026-08-26T18:02 explicit grant, scoped to this rehab skill manifest only"
delivery_authority:
  source: project-policy-absent
  mode: local
  merge: manual
  post_merge_sync: false
  authorized_actions: [local-commit]
  forbidden_actions: [create-plan-worktree, push-topic, create-pr, merge, push-default, remote-ref-delete, host-session-delete]
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "The requirements define the declared-work-to-git pairing, a containment-proven lifecycle, grant-bound delivery, and delegated reaping without duplicating kbreconcile's artifact scope."
      - "The requirements-wide adversarial review has no unresolved P0/P1 findings; both residuals are P2/P3 and carry binding constraints."
      - "Proof-command provenance, adapter-absence polarity, containment authority, delivery-state eligibility, and blast radius were each corrected before decomposition."
      - "The user's explicit maximum-autonomy grant is bound to deterministic proof, so ambiguity never defaults to action."
    proof:
      - docs/brainstorms/2026-08-26-kb-rehab-outstanding-work-requirements.md
      - docs/results/document-reviews/kb-rehab-outstanding-work-requirements-0d399da7ecfb.json
      - docs/brainstorms/2026-08-01-global-cleanup-reconciliation-requirements.md
      - internal/reconcile/policy.go
    blockers: []
    passed_at: "2026-08-26T19:08:00Z"
    allowed_next_action: "kb-plan docs/brainstorms/2026-08-26-kb-rehab-outstanding-work-requirements.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    gate_scope: implementation
    status: passed
    required_evidence:
      - "Three vertical slices cover read-only pairing, the kb-rehab triage lane, and grant-bound delivery with delegated reaping."
      - "Every requirement R1-R19 including R1a, R1b, R4a, R5a, R13a, R13b, R14a, R14b, R14c, and R15a maps to exactly one owning slice."
      - "The dependency graph is linear and acyclic, and every slice serializes the shared Git common directory, refs, and skill tree."
      - "Every destructive path inherits the shipped ActionMerge mandatory predicates and meets or exceeds terminalCleanupSafetyPredicates()."
      - "No slice adds a new dependency, and each slice names the cheaper tier it ruled out."
    proof:
      - docs/plans/2026-08-26-001-work-reality-report-plan.md
      - docs/plans/2026-08-26-002-rehab-lane-triage-plan.md
      - docs/plans/2026-08-26-003-rehab-granted-delivery-plan.md
      - docs/brainstorms/2026-08-26-kb-rehab-outstanding-work-requirements.md
      - internal/reconcile/policy.go
    blockers: []
    passed_at: "2026-08-26T19:10:00Z"
    allowed_next_action: "kb-work docs/plans/2026-08-26-000-kb-rehab-outstanding-work-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "kbcheck work-reality pairs declared work against git reality read-only and mutates no ref, worktree, or working tree."
      - "Every terminal classification is backed by git cherry patch equivalence against a freshly fetched authoritative default resolved through ls-remote --symref."
      - "Absent remote, unreachable remote, unresolvable advertised default, and absent predicate manifest each fail closed with zero dead and zero superseded."
      - "The preservation predicate set strictly exceeds terminalCleanupSafetyPredicates()."
      - "Run against this repository the command classified three branches dead with containment proof, three orphan-branch, and two live, with no hand-pairing."
    proof:
      - docs/results/proofs/kb-rehab-20260826-slice-001.md
      - cmd/kbcheck/work_reality.go
      - cmd/kbcheck/work_reality_test.go
      - config/rehab-policy.json
      - docs/plans/2026-08-26-001-work-reality-report-plan.md
    blockers: []
    passed_at: "2026-08-26T22:35:00Z"
    allowed_next_action: "kb-work docs/plans/2026-08-26-000-kb-rehab-outstanding-work-manifest.md"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    gate_scope: implementation
    status: passed
    required_evidence:
      - "kb-rehab exists as a routable lane at kb-start rank 3a and names kb-map, kbcheck work-reality, kb-complete, and kbreconcile explicitly."
      - "Only a pairing the report classified dead or superseded with containment proof may be marked, and a fail-closed report refuses every write."
      - "A removal is permitted only when a superseding or completing artifact already resolves in the authoritative default tree and the ref holds no uncontained commits."
      - "The decision packet never exceeds five grouped items, never omits a mandated field, and accounts for every ambiguous pairing."
      - "An unanswered packet item leaves both the work item and the ref untouched."
    proof:
      - docs/results/proofs/kb-rehab-20260826-slice-002.md
      - .github/skills/kb-rehab/SKILL.md
      - .github/skills/kb-rehab/references/classification.md
      - cmd/kbcheck/work_reality.go
      - docs/plans/2026-08-26-002-rehab-lane-triage-plan.md
    blockers: []
    passed_at: "2026-08-26T23:40:00Z"
    allowed_next_action: "kb-work docs/plans/2026-08-26-000-kb-rehab-outstanding-work-manifest.md"
slices:
  - id: slice-001
    title: "Pair declared work against git reality, read-only"
    path: docs/plans/2026-08-26-001-work-reality-report-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: medium
    model_tier_reason: "Bounded additive work over the exported reconcile.Inventory API and existing remote-authority helpers, with fully enumerated acceptance criteria and no open product or architecture choice."
    model_requirements: ["Go against an existing internal package", "Git ancestry and patch-equivalence reasoning", "deterministic fixture repositories", "fail-closed classification design"]
    escalation_triggers: ["a classification would be emitted without the evidence its lifecycle row names", "containment would be reimplemented instead of reusing internal/reconcile", "a missing adapter would add rather than withhold a conclusion", "the preservation set would fall short of terminalCleanupSafetyPredicates()"]
    workspace_mode: shared-serial
    conflict_domains: ["go:kbcheck-commands", "cli:kbcheck", "config:rehab-policy"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'WorkReality' -count=1"
      expect: 0
    status: done
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/kbcheck/work_reality_test.go
        role: "mixed-fixture lifecycle, fail-closed, read-only, determinism, redaction, and preservation-floor oracle"
        sha256: "d9592f7cfbab222800ba58b09707af1d8ddd0d241c8aea5ded495d6608c1e439"
        update_policy: "requires an explicit slice-plan amendment"
    notes: "cost_tier=2 reusing reconcile.Inventory, fetchAuthoritativeRemoteDefault, and terminalCleanupSafetyPredicates; ruled out tier 6 because a new git-walking package would duplicate the shipped inventory ledger. DDR route=current orchestrator, exception context-required; exact slice proof PASS 18 tests; full cmd/kbcheck package PASS; go vet/go build/gofmt/git diff --check PASS; qa-browser skipped, no UI-reachable behavior; forecast implementation=4 actual implementation=4 discovered=0 unused=0; predicate-manifest-version=rehab-1.0.0; in-slice correction: terminal-status manifests now report as settled instead of the false-positive orphan-work; internal/reconcile git helpers did not need exporting, so the plan-time escalation trigger did not fire."
  - id: slice-002
    title: "kb-rehab lane: triage, markers, and decision packet"
    path: docs/plans/2026-08-26-002-rehab-lane-triage-plan.md
    blockers: [slice-001]
    verification: verification
    test_level: integration
    functional_risk: moderate
    execution_class: cli
    model_tier: medium
    model_tier_reason: "Authoring one thin skill lane and a bounded marker writer over a proven report, inheriting the todo-triage decision taxonomy and the reconciler packet contract without introducing new vocabulary."
    model_requirements: ["KB skill authoring against the repo skill contract", "Markdown table editing that preserves existing status markers", "integration testing against a fixture todo.md"]
    escalation_triggers: ["a marker or removal would be written without containment proof", "a decision vocabulary todo-triage does not define would be required", "the packet would exceed five items or omit a mandated field", "ambiguity would default to action"]
    workspace_mode: shared-serial
    conflict_domains: ["skills:kb-rehab", "skills:kb-start", "docs:readme", "go:kbcheck-commands"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'WorkReality|SkillRepoContract' -count=1"
      expect: 0
    status: done
    owner: agent
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/kbcheck/work_reality_test.go
        role: "lifecycle oracle extended with marker, removal-gate, preservation, and packet-bound assertions"
        sha256: "29d5fd11517756e463750e36342b8cb49f42680f135373846fd4d36c041be914"
        update_policy: "additive only; a slice-001 assertion may not be weakened or deleted without an explicit plan amendment"
    notes: "cost_tier=2 reusing todo-triage's decision taxonomy and kbreconcile's decision-packet contract; ruled out tier 6 because a new classification vocabulary would fork triage semantics already versioned in this bundle. DDR route=current orchestrator, exception context-required; exact slice proof PASS; full cmd/kbcheck package PASS; go vet/go build/skill-lint/git diff --check PASS; qa-browser skipped, no UI-reachable behavior; forecast implementation=6 actual implementation=6 discovered=0 unused=0, plus the slice-001-forecast cmd/kbcheck/main.go command-surface edit; packet grouped by state/protection-scope folds surplus groups into omitted_pairings instead of dropping them; marking reuses todo.md's existing skipped vocabulary rather than forking triage semantics."
  - id: slice-003
    title: "Granted delivery and delegated reaping"
    path: docs/plans/2026-08-26-003-rehab-granted-delivery-plan.md
    blockers: [slice-002]
    verification: tdd
    test_level: functional-cli
    functional_risk: destructive
    execution_class: cli
    model_tier: large
    model_tier_reason: "An explicit user grant authorizes merges to a default branch that propagates into three global agent install roots, so proof provenance, adapter-absence polarity, predicate inheritance, grant binding, and TOCTOU fencing must hold together and one weak predicate is an irreversible supply-chain write."
    model_requirements: ["security-boundary reasoning about self-attesting proof and adapter absence", "compare-and-swap and TOCTOU fencing", "faithful inheritance of a versioned predicate manifest", "adversarial negative fixtures"]
    escalation_triggers: ["a predicate would be satisfied by adapter absence", "a proof command would be resolved from the candidate branch", "a grant would authorize a pairing it does not enumerate", "a grant would be persisted or inherited across runs", "a protected-path pairing would auto-merge", "post-merge sync would be attempted under a grant"]
    workspace_mode: shared-serial
    conflict_domains: ["go:kbcheck-commands", "go:reconcile-policy", "cli:kbcheck", "git:refs"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck ./internal/reconcile -run 'RehabDelivery|RehabGrant|TerminalCleanup' -count=1"
      expect: 0
    status: pending
    owner: agent
    can_continue_other_slices: false
    notes: "cost_tier=2 inheriting the eight ActionMerge mandatory predicates from internal/reconcile/policy.go and delegating ref and worktree reaping to kbreconcile apply; ruled out tier 6 because a second deletion engine would fork the safety predicates this bundle already ships. R15a substitutes 'go run ./cmd/kbcheck local-release' resolved from the authoritative default tree as the check adapter; R14b still stops every protected-path pairing at PR."
---

# kb-rehab: Outstanding Work Reconciliation

## Objective

Pair every piece of *declared* work in this repository against *git reality*,
classify each pairing with containment-proven evidence, deliver what is
provably complete, mark what is provably dead or superseded, and hand back a
bounded packet for everything that is genuinely ambiguous.

`kbreconcile` already owns artifact residue: worktrees, sessions, refs,
receipts, and queue claims. `todo-triage` already owns declared-work
classification but is git-blind. Nothing pairs the two. That pairing is the
whole product.

## Non-Goals

- Rebuilding inventory, ref deletion, or worktree reaping. Slice 003 delegates
  to `kbreconcile apply`.
- Forking the triage decision vocabulary. Slice 002 inherits `todo-triage`.
- Any post-merge sync to global install roots. R14a forbids it under a grant;
  it stays an explicit, separate human authorization.
- Touching ATV or any sibling checkout.

## Authority Posture

The user granted maximum runtime autonomy on 2026-08-26: auto-merge and
auto-delete anything proving terminal, ask only on genuine ambiguity. That
grant is bound to deterministic proof, not to confidence.

The grant governs what `kb-rehab` may do **when it runs**. A separate explicit
grant on 2026-08-26T18:02, scoped to this manifest only, authorizes local
commits for the work that builds it. Publishing, PR creation, merging, and
post-merge sync remain unauthorized and stay separate human decisions.

## Blast Radius

Merging to `main` in this repository syncs `.github/skills/**` into
`~/.codex/skills/`, `~/.copilot/skills/`, and `~/.agents/skills/`. That is a
supply-chain write into every future agent session on this host, not a
repo-local change. R14b therefore makes any pairing touching `.github/skills`,
`.github/agents`, `.github/instructions`, `cmd`, `internal`, `scripts`, or
`config/skill-quality.json` auto-merge-ineligible; those stop at PR for human
review regardless of the grant.

## Slice Sequence

```mermaid
flowchart LR
  S1["slice-001<br/>work-reality report<br/>read-only"] --> S2["slice-002<br/>kb-rehab lane<br/>markers + packet"] --> S3["slice-003<br/>granted delivery<br/>delegated reaping"]
```

The graph is deliberately linear. Slice 002 cannot classify without slice 001's
evidence, and slice 003 must not delete or merge without slice 002's
containment-proven classification.
