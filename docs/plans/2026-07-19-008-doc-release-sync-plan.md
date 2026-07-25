---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-008
title: "Document, synchronize, and release the optional routing contract"
blockers: [slice-007]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: medium
model_tier_reason: "Architecture is complete; this slice needs disciplined drift merge, documentation consistency, hash proof, and release execution."
model_requirements: ["cross-install drift review", "documentation consistency", "hash verification", "core and local-release gates"]
escalation_triggers: ["global copies contain newer useful work", "required hashes differ after sync", "core/local-release fails", "docs imply mandatory providers or forced cleanup"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-008.json
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
hitl: false
expected_files:
  - path: README.md
    op: edit
    scope: "Visible graph-routing, multi-session, and worktree workflow."
  - path: docs/context/PROJECT.md
    op: edit
    scope: "Route pointers and current optional-provider/concurrency truth."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Final provider, packet, lease, workspace, and integration architecture."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Canonical contributor and release proof commands."
  - path: docs/context/eval-map.md
    op: edit
    scope: "Graph-routing eval and promotion surfaces."
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "Map/provider freshness and graphify-size-check maintenance result."
  - path: config/skill-quality.json
    op: edit
    scope: "Register new deterministic checks only when required by release policy."
  - path: todo.md
    op: edit
    scope: "Remove completed active row before archival."
  - path: todo-done.md
    op: edit
    scope: "Archive verified completion summary."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Complete; final release gate passed."
human_action: ""
can_continue_other_slices: false
---

# Slice 008: Documentation, Sync, And Release

## What To Build

Finish the user-visible and durable contract, merge any useful drift from
Codex/Copilot/shared-agent copies back into the working source, sync approved
skills, refresh project memory, and run contributor/release gates.

## Why This Slice Exists

This repository is a portable skill bundle. A locally correct implementation is
not released while global copies drift or docs imply a stronger guarantee than
the deterministic checks prove.

## Acceptance Criteria

- README explains when worktrees are required and why leases still matter.
- Architecture docs distinguish planning intent from work-time mechanics.
- Docs state the local Git-common-dir boundary and do not claim cross-machine locking.
- Optional providers remain optional; file-native routing and core checks pass without them.
- Each changed skill is diffed against working, Codex, Copilot, and shared-agent
  copies before overwrite; useful global-only work is merged first.
- Final approved skill hashes match all required install roots.
- `go run ./cmd/kbcheck core` and `go run ./cmd/kbcheck local-release` exit 0.
- `git diff --check` passes for the working repo.
- ATV repositories are not inspected, modified, synced, committed, or gated.
- Project map, eval map, testing docs, todo, and completion archive are current.

## Test Scenarios

- Fresh clone/core path succeeds without optional provider artifacts.
- Required skill hash report has zero drift after sync.
- Docs and CLI help agree on workspace modes, claim scope, and cleanup refusal.
- Local-release fails when a required copy is deliberately drifted in a fixture.

## Proof Check

`go run ./cmd/kbcheck local-release`

## Scope Boundary

No commit, push, PR, merge, provider installation, ATV work, or cleanup of
unrelated user changes unless separately requested.

## Dependencies

Slice 007 must prove the claims before they are documented and propagated.

## Concurrency

Serial integration-owner slice. No other session may sync the same skill roots
while release proof is running.

## Completion Notes

- Reviewed Copilot/shared-agent drift for `ce-review`, `kb-map`, `kb-plan`,
  `kb-regression-snapshot`, `kb-review`, and `kb-work`; repo copies contained
  the approved graph-routing/workspace updates and no useful global-only work was
  found.
- Repaired missing/stale required global skill copies with `kbcheck doctor --fix`
  after setting reviewed-drift markers for the stale Copilot/shared-agent copies.
- Proof passed: `go run ./cmd/kbcheck skill-sync-report --root E:/Dev/Tools/working-skill-repo --config config/skill-quality.json`.
- Proof passed: `go run ./cmd/kbcheck core`.
- Proof passed: `go run ./cmd/kbcheck local-release`.
