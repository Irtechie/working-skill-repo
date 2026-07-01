---
kb_id: kb-2026-07-01-native-scoped-learning
slice_id: slice-016
title: "Migrate existing state + refresh docs/proof surface"
blockers: [slice-015]
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: docs/context/kb/instincts/project.yaml
    op: create
    scope: "move existing 3 global instincts from .atv/instincts/project.yaml to kb-native root"
  - path: docs/context/kb/instincts/scoped/.gitkeep
    op: create
    scope: "establish the scoped-instinct directory"
  - path: docs/context/kb/kb-completions.txt
    op: create
    scope: "move .atv/kb-completions.txt content"
  - path: .atv/instincts/project.yaml
    op: delete
    scope: "remove legacy location after migration"
  - path: README.md
    op: edit
    scope: "remove atv naming; document kb-native two-tier scoped-by-default learning + promotion-on-recurrence"
  - path: AGENTS.md
    op: edit
    scope: "point learning/instinct guidance at kb-native roots + scoped-by-default rule"
  - path: docs/context/PROJECT.md
    op: edit
    scope: "update learning-model description to kb-native scoped tiers"
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "reference kb-native instinct roots + per-scope decay/cap"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Move state files to kb-native roots, delete legacy .atv locations, refresh visible docs; run kbcheck core + git diff --check."
human_action: ""
can_continue_other_slices: true
---

# Slice 016 — migrate state + docs/proof surface

## What to build

Move the actual state and update the human-visible surface so the repo IS
kb-native, not just described as such.

1. **State migration**:
   - Move the 3 existing global instincts from `.atv/instincts/project.yaml` to
     `docs/context/kb/instincts/project.yaml` verbatim (preserve confidence,
     observations, dates, evidence). Add `scope: project` to each (they are
     genuinely cross-cutting workflow instincts).
   - Create `docs/context/kb/instincts/scoped/` (with `.gitkeep`).
   - Move `.atv/kb-completions.txt` -> `docs/context/kb/kb-completions.txt`.
   - Delete the legacy `.atv/instincts/project.yaml` and `.atv/kb-completions.txt`
     after the move. Leave `.atv/eval-runs/` + `.atv/pipeline-runs/` historical
     provenance in place if still referenced by past eval records, but they are
     gitignored anyway; note them as legacy.
2. **Docs refresh** (visible surface):
   - README.md + AGENTS.md + PROJECT.md + memory-maintenance.md: remove functional
     `.atv/` naming; document the kb-native roots and the **scoped-by-default**
     learning model with **promotion-on-recurrence** and the **landmine fast-path**.
   - State plainly: "X pipeline's lessons are not visible to Y pipeline unless
     promoted to global via cross-scope recurrence."

## Acceptance criteria

- `docs/context/kb/instincts/project.yaml` holds the migrated instincts; legacy
  `.atv/instincts/project.yaml` is gone.
- `docs/context/kb/instincts/scoped/` exists.
- README/AGENTS/PROJECT/memory-maintenance describe the two-tier scoped-by-default
  model and reference only kb-native roots.
- `kbcheck core` passes; `git diff --check` clean.

## Scope boundary

- Does not implement any downstream component's own calibration/fixtures (separate
  repo task, named in the goal ledger).

## Verification

- verification-only: `go run ./cmd/kbcheck core` passes; `git diff --check` clean;
  grep confirms no functional `.atv/` references remain repo-wide (skills + docs).
