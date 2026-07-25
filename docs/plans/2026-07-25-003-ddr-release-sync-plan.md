---
type: kb-slice-plan
kb_id: kb-2026-07-25-orchestrator-directed-ddr
slice_id: ddr-003
title: DDR release and sync
status: done
blockers: [ddr-001, ddr-002]
verification: verification-only
test_level: full
functional_risk: broad
model_tier: large
model_tier_reason: "Release reconciliation requires repository ownership, dirty-baseline judgment, and exact sync control."
model_requirements: ["Git integration authority", "release gate interpretation", "exact-path sync"]
escalation_triggers: ["release gate times out", "global drift conflicts", "push is rejected"]
workspace_mode: shared-serial
conflict_domains: ["git:main", "global skill installs"]
shared_resources: ["git:integration-owner", "sync:global-skills"]
context_packet: docs/plans/2026-07-25-orchestrator-directed-ddr-context/ddr-003.json
expected_files:
  - path: todo.md
    op: edit
    scope: project active-work lifecycle
  - path: README.md
    op: edit
    scope: update only if release-visible commands change
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
---

# DDR-003: DDR release and sync

## Goal

Integrate the reviewed change without losing newer global-skill drift.

## Steps

1. Run focused proof and `go run ./cmd/kbcheck core`.
2. Compare each changed skill with Codex, Copilot, and shared-agent installs.
3. Merge useful global-only drift back into the working bundle.
4. Run `go run ./cmd/kbcheck local-release`.
5. Sync exact approved skill directories and verify hashes.
6. Commit, merge the isolated branch to `main`, and push.

## Block policy

A command timeout is a failed proof boundary. Do not describe the release as
green or overwrite global installs if the release gate cannot complete.

## Acceptance criteria

- The final commit contains only the bounded DDR change.
- Required global skill targets are compared before overwrite.
- Synced copies hash-match the working bundle.
- `main` push is non-force and preserves unrelated files.
