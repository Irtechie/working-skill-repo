---
type: kb-manifest
kb_id: kb-2026-06-10-skill-bundle-cleanup
brainstorm_path: docs/brainstorms/2026-06-10-skill-bundle-cleanup-audit.md
created: 2026-06-10
status: reviewed
workflow_shape: "multi-stream-epic"
scope-verified-files:
  - .github/copilot-instructions.md
  - .github/skills/learn/SKILL.md
  - AGENTS.md
  - README.md
  - bin/kb-install.mjs
  - cmd/kbcheck/checks.go
  - cmd/kbcheck/checks_test.go
  - cmd/kbcheck/marketplace_promotion.go
  - cmd/kbcheck/report_validators.go
  - cmd/kbcheck/skill_eval.go
  - cmd/kbcheck/skill_repo_contract_test.go
  - cmd/kbcheck/skill_validators.go
  - docs/brainstorms/2026-06-10-skill-bundle-cleanup-audit.md
  - docs/context/PROJECT.md
  - docs/context/eval-map.md
  - docs/context/memory-maintenance.md
  - docs/context/operations/testing.md
  - docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md
  - docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md
  - docs/plans/2026-06-10-001-verification-audit-refresh-plan.md
  - docs/plans/2026-06-10-002-gate-toolchain-onramp-plan.md
  - docs/plans/2026-06-10-003-core-install-closure-plan.md
  - docs/plans/2026-06-10-004-surface-minimality-review-dedupe-plan.md
  - docs/plans/2026-06-10-005-review-skill-canonicalization-plan.md
  - docs/plans/2026-06-10-006-docs-output-platform-claims-plan.md
  - go.mod
  - todo-done.md
  - todo.md
  - .atv/instincts/project.yaml
  - .atv/kb-completions.txt
  - docs/solutions/workflow-issues/contributor-core-vs-release-sync-gates-2026-06-10.md
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-brainstorm
    status: passed
    required_evidence:
      - "requirements path exists"
      - "Question Gate classification exists"
      - "Resolve Before Planning is empty"
      - "safe assumptions, deferred planning questions, and parked items are recorded"
    proof:
      - docs/brainstorms/2026-06-10-skill-bundle-cleanup-audit.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-plan docs/brainstorms/2026-06-10-skill-bundle-cleanup-audit.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk"
      - "HITL classification is explicit for every slice"
    proof:
      - docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md
      - docs/plans/2026-06-10-001-verification-audit-refresh-plan.md
      - docs/plans/2026-06-10-002-gate-toolchain-onramp-plan.md
      - docs/plans/2026-06-10-003-core-install-closure-plan.md
      - docs/plans/2026-06-10-004-surface-minimality-review-dedupe-plan.md
      - docs/plans/2026-06-10-005-review-skill-canonicalization-plan.md
      - docs/plans/2026-06-10-006-docs-output-platform-claims-plan.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "every slice is done or intentionally skipped"
      - "final verification command/result is recorded"
      - "board/manifest are synced"
      - "scope-verified-files is populated"
      - "required sync drift is zero"
    proof:
      - docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md
      - docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md
      - docs/context/operations/testing.md
      - todo-done.md
      - go.mod
    proof_commands:
      - "go run ./cmd/kbcheck core"
      - "go run ./cmd/kbcheck local-release"
      - "go run ./cmd/kbcheck skill-sync-report --verbose-optional"
      - "git diff --check"
      - "git -C E:\\all-the-vibes diff --check"
      - "node ./bin/kb-install.mjs --target all --profile core --install-root ./.tmp/node-install-core-proof --dry-run"
      - "node ./bin/kb-install.mjs --target all --profile full --install-root ./.tmp/node-install-full-proof --dry-run"
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-complete docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "audit findings classified with evidence"
      - "stale findings retired"
      - "durable memory signals updated"
    proof:
      - docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md
      - docs/context/memory-maintenance.md
      - docs/plans/2026-06-10-001-verification-audit-refresh-plan.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "contributor core gate no longer includes release-only sync drift"
      - "Go directive lowered"
      - "docs updated"
    proof:
      - go.mod
      - cmd/kbcheck/checks.go
      - cmd/kbcheck/checks_test.go
      - cmd/kbcheck/skill_repo_contract_test.go
      - README.md
      - .github/copilot-instructions.md
      - AGENTS.md
      - docs/context/operations/testing.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "core installer profile dependency closure updated"
      - "core/full installer dry-runs passed"
      - "install docs updated"
    proof:
      - bin/kb-install.mjs
      - README.md
      - docs/context/memory-maintenance.md
      - docs/plans/2026-06-10-003-core-install-closure-plan.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "minimality evidence recorded"
      - "no deletion performed without human approval"
    proof:
      - docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md
      - docs/context/memory-maintenance.md
      - docs/plans/2026-06-10-004-surface-minimality-review-dedupe-plan.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "review-skill fork status classified"
      - "no unsafe collapse performed"
    proof:
      - docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md
      - docs/plans/2026-06-10-005-review-skill-canonicalization-plan.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: slice-slice-006-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "misleading selftest wording fixed"
      - "unshipped hook claim fixed"
      - "docs/platform claims aligned"
    proof:
      - cmd/kbcheck/marketplace_promotion.go
      - cmd/kbcheck/report_validators.go
      - cmd/kbcheck/skill_eval.go
      - cmd/kbcheck/skill_validators.go
      - .github/skills/learn/SKILL.md
      - README.md
      - docs/context/PROJECT.md
    blockers: []
    passed_at: "2026-06-10"
    allowed_next_action: "kb-work docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
  - gate_id: complete-to-ship
    status: passed
    passed_at: "2026-06-10"
    allowed_next_action: "kb-ship docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
    required_evidence:
      - "kb-check final command/result recorded"
      - "functional-test skip reason recorded"
      - "kb-review mode and finding counts recorded"
      - "P0/P1 resolved or no P0/P1"
      - "follow-up-resolution summary recorded"
      - "proof/demo evidence recorded"
      - "compound/learn/evolve result recorded"
      - "project-memory refresh proof recorded"
      - "memory-maintenance update recorded"
      - "cleanup result recorded"
      - "alerts list recorded"
    proof:
      - "docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md"
      - "docs/context/research/2026-06-10-skill-bundle-cleanup-audit-refresh.md"
      - "docs/context/operations/testing.md"
      - "docs/context/memory-maintenance.md"
      - "docs/solutions/workflow-issues/contributor-core-vs-release-sync-gates-2026-06-10.md"
      - ".atv/instincts/project.yaml"
      - ".atv/kb-completions.txt"
      - "todo-done.md"
      - "README.md"
      - "bin/kb-install.mjs"
      - "cmd/kbcheck/checks.go"
    proof_commands:
      - "go run ./cmd/kbcheck local-release"
      - "go run ./cmd/kbcheck skill-sync-report"
      - "git diff --check"
      - "git -C E:\\all-the-vibes diff --check"
      - "node ./bin/kb-install.mjs --target all --profile core --install-root ./.tmp/node-install-core-proof --dry-run"
      - "node ./bin/kb-install.mjs --target all --profile full --install-root ./.tmp/node-install-full-proof --dry-run"
    blockers: []
slices:
  - id: slice-001
    title: "Reproduce and classify audit findings"
    path: docs/plans/2026-06-10-001-verification-audit-refresh-plan.md
    blockers: []
    verification: verification-only
    test_level: functional-cli
    functional_risk: broad
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Build the current evidence table and retire stale findings before code edits."
    human_action: ""
    can_continue_other_slices: false
    notes: "Downstream slices must implement only findings still true after this audit refresh."
    protected_oracles: []
  - id: slice-002
    title: "Make the fresh-clone gate and Go toolchain contributor-safe"
    path: docs/plans/2026-06-10-002-gate-toolchain-onramp-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Adjust only confirmed gate/toolchain gaps from slice 001."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
  - id: slice-003
    title: "Close or document the core install dependency closure"
    path: docs/plans/2026-06-10-003-core-install-closure-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Make core install behavior self-consistent for installed skills."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
  - id: slice-004
    title: "Reduce or justify loaded agent and skill surface"
    path: docs/plans/2026-06-10-004-surface-minimality-review-dedupe-plan.md
    blockers: [slice-001]
    verification: verification-only
    test_level: functional-cli
    functional_risk: narrow
    hitl: true
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Prepare evidence-backed delete, merge, or retain recommendations."
    human_action: "Approve any agent or skill deletion before execution removes files."
    can_continue_other_slices: true
    notes: "Deletion is human-gated; reports and allowlist updates are agent-owned."
    protected_oracles: []
  - id: slice-005
    title: "Canonicalize review-skill shared doctrine"
    path: docs/plans/2026-06-10-005-review-skill-canonicalization-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Eliminate accidental drift or document the canonical split."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
  - id: slice-006
    title: "Align docs, selftest wording, hook references, history, and platform claims"
    path: docs/plans/2026-06-10-006-docs-output-platform-claims-plan.md
    blockers: [slice-001]
    verification: verification-only
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Patch only claims and outputs still inaccurate after slice 001."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
---

# KB: Skill Bundle Cleanup Audit Follow-up

## Origin

Brainstorm: `docs/brainstorms/2026-06-10-skill-bundle-cleanup-audit.md`

## Workflow Shape

`multi-stream-epic` because the audit spans contributor setup, install
profiles, validation output, skill dependencies, reviewer-skill doctrine,
minimality, repo docs, and platform claims. Slice 001 prevents stale-audit
cleanup from mutating already-fixed surfaces.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Reproduce and classify audit findings | - | verification-only | no | done |
| 2 | Make the fresh-clone gate and Go toolchain contributor-safe | slice-001 | integration | no | done |
| 3 | Close or document the core install dependency closure | slice-001 | integration | no | done |
| 4 | Reduce or justify loaded agent and skill surface | slice-001 | verification-only | yes | done |
| 5 | Canonicalize review-skill shared doctrine | slice-001 | integration | no | done |
| 6 | Align docs, selftest wording, hook references, history, and platform claims | slice-001 | verification-only | no | done |

## DAG

- `slice-001` has no blockers.
- `slice-002` through `slice-006` each depend on `slice-001`.
- No cycles exist.

## Execution Rule

Do not implement audit claims directly from the pasted text. Execute slice 001
first, then update downstream scope if a finding is stale, already fixed, or
materially different in the current repo.

## Next Command

`kb-complete docs/plans/2026-06-10-000-kb-skill-bundle-cleanup-manifest.md`

## kb-complete closure - 2026-06-10

- review: local-fallback; P0=0 P1=0 P2=1(resolved) P3=0.
- follow-up-resolution: resolved 1, logged 0, blocked 0.
- proof/demo evidence: `go run ./cmd/kbcheck local-release`, `go run ./cmd/kbcheck skill-sync-report`, `git diff --check`, `git -C E:ll-the-vibes diff --check`, and installer dry-runs passed; browser verification skipped because this is a non-UI skill/tooling change.
- compound: `docs/solutions/workflow-issues/contributor-core-vs-release-sync-gates-2026-06-10.md`.
- learn: added 1 project instinct and decayed 1 stale instinct.
- evolve: skipped; completion counter is 7, not divisible by 5.
- project memory: refreshed `docs/context/PROJECT.md`, `docs/context/operations/testing.md`, `docs/context/eval-map.md`, and `docs/context/memory-maintenance.md`.
- memory-maintenance: recorded 3 signals.
- cleanup: `.atv/qa-screenshots` and `.atv/observations.jsonl` were absent, so artifact cleanup was a no-op; todo hygiene complete.
- alerts: `docs/solutions/` discoverability is not surfaced in `AGENTS.md`; optional ATV scaffold/plugin drift remains warning-only unless intentionally shipping those surfaces.
