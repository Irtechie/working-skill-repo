---
kb_id: kb-2026-08-01-global-cleanup-reconciliation
slice_id: slice-003
title: "Make KB lifecycle and global install converge automatically"
blockers: [slice-001, slice-002]
verification: tdd
test_level: integration
functional_risk: broad
execution_class: cli
model_tier: large
model_tier_reason: "This slice connects the checked core to queue ownership, completion states, managed binary distribution, and scheduled/later-session execution across runtimes."
model_requirements: ["cross-skill lifecycle design", "managed binary installation", "queue CAS", "documentation/eval synchronization"]
escalation_triggers: ["installation would become mandatory for skill-only users", "open PRs consume active WIP", "cleanup is conflated with delivery", "global copies contain newer drift"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-start", "skill:kb-complete", "skill:kb-finalize", "installer:managed-binaries", "docs:workflow"]
shared_resources: ["install:global-bin", "state:work-queue", "docs:project-map"]
proof_check:
  kind: command_exit
  command: "go test ./... && node --test ./bin/kb-install.test.mjs && go run ./cmd/kbcheck skill-lint --root ."
  expect: 0
hitl: false
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Integrate lifecycle registration, owner-based WIP, optional install, docs, and end-to-end proof."
human_action: ""
can_continue_other_slices: false
protected_oracles:
  - path: bin/kb-install.test.mjs
    purpose: "Managed binary installation remains checksum-verified, drift-safe, optional, and skill-only compatible."
  - path: cmd/kbcheck/delivery_chain_contract_test.go
    purpose: "Delivery authority stays with kb-complete/kb-ship/kb-land and cleanup remains separately gated."
expected_files:
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "Use the global reconciler when available and preserve repo-native safe fallback."
  - path: .github/skills/kb-finalize/SKILL.md
    op: edit
    scope: "Register lifecycle state without deleting the current worktree."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Register local-durable, awaiting-review, or delivery-integrated endpoint separately from cleanup."
  - path: .github/skills/kb-start/scripts/work_queue.ps1
    op: edit
    scope: "Count active execution owners rather than duplicate claims and support terminal/suspended reconciliation."
  - path: bin/kb-install.mjs
    op: edit
    scope: "Install the checksum-verified reconciler binary optionally without breaking skill-only installs."
  - path: bin/kb-install.test.mjs
    op: edit
    scope: "Prove asset mapping, checksum, drift-safe upgrade/uninstall, and optional fallback."
  - path: README.md
    op: edit
    scope: "Document global plan/apply/verify, open-PR waiting, compact exceptions, and independent cleanup dimensions."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Document lifecycle registration and later/scheduled reconciliation."
  - path: docs/context/architecture/kbcheck.md
    op: edit
    scope: "Document shared terminal safety contract and conformance."
  - path: docs/context/operations/skill-bundle-maintenance.md
    op: edit
    scope: "Document reconciler installation and drift-safe maintenance."
  - path: docs/context/PROJECT.md
    op: edit
    scope: "Route fresh sessions to the global reconciler subsystem."
  - path: cmd/kbcheck/delivery_chain_contract_test.go
    op: edit
    scope: "Require endpoint registration, later cleanup, and separate terminal dimensions."
---

# Slice 003 - Automatic Lifecycle Convergence

## Acceptance Criteria

- Successful KB completion registers `local-durable`, `awaiting-review`, or
  `delivery-integrated` before ownership release and never self-deletes.
- Open PRs may suspend execution and release active WIP only with exact remote
  topic evidence and a complete resume packet; refs remain.
- Active WIP caps count distinct active owners, not duplicate claims.
- Later `kb-start`, scheduled, and on-demand runs use the global reconciler when
  available; repo-native sweep remains a compatible fail-closed fallback.
- Installer handling of the reconciler is checksum-verified, drift-safe,
  optional, and does not make skill-only installs fail.
- Documentation and deterministic contracts keep delivery, physical cleanup,
  ref retirement, and host session retirement independently visible.

## Test Scenarios

1. Two active claims from one session consume one owner slot.
2. Archived/terminal exact session evidence releases an orphaned claim through
   CAS; age alone produces one grouped exception.
3. Manual PR completion enters awaiting-review and retains refs/resume state.
4. Missing reconciler asset continues a skill-only install in automatic mode.
5. End-to-end fixtures exercise dry-run, plan, apply, verify, and later-session
   sweep without weakening existing terminal-cleanup tests.

## Scope Boundary

Do not authorize publication, merge, remote ref deletion, host session deletion,
required daemons, cross-machine locking, or global-copy sync before landed source.
