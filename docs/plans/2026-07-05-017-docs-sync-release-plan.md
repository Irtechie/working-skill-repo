---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-007
title: "Update docs, sync surfaces, and release gate"
blockers: [slice-006]
verification: verification-only
test_level: functional-cli
functional_risk: narrow
model_tier: medium
model_tier_reason: "This is mostly propagation and release hygiene after the architecture decision is accepted."
hitl: false
expected_files:
  - path: README.md
    op: edit
    scope: "document visible planner-economy, context-packet, task-state, or adapter changes if shipped"
  - path: docs/context/PROJECT.md
    op: edit
    scope: "refresh the route map if workflow surfaces changed"
  - path: docs/context/research/README.md
    op: edit
    scope: "index the updated HumanLayer/Dex research"
  - path: todo.md
    op: edit
    scope: "close or redirect the active workstream"
  - path: .github/skills
    op: sync
    scope: "propagate approved skill changes to required roots after drift review"
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: "slice-006 done"
next_agent_action: "Update docs, sync final skill copies, and run release gates."
human_action: ""
can_continue_other_slices: true
notes: "Working docs and ATV-facing README updated; kb-plan/kb-work synced to required Codex, Copilot, shared agents, and ATV .github roots."
---

# Slice 007 - Docs, Sync, and Release

## What To Build

Propagate only the architecture changes accepted after the absorption spike.

## Acceptance Criteria

- User-facing docs match the accepted core-vs-payload decision.
- Research index and project map point to the updated decision.
- Required skill roots are compared before overwriting and synced only after
  useful drift is merged.
- ATV-facing README is updated only if ATV-shipped behavior changes.
- Release gates pass.

## Scope Boundary

Do not sync speculative or rejected spike changes.

## Verification

Run:

```shell
go run ./cmd/kbcheck local-release
git diff --check
```
