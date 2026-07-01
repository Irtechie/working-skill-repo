---
kb_id: kb-2026-07-01-native-scoped-learning
slice_id: slice-015
title: "Update kbcheck Go harness + gitignore + installer (no reintroduction)"
blockers: [slice-012, slice-013, slice-014]
verification: tdd
test_level: unit
functional_risk: narrow
hitl: false
expected_files:
  - path: cmd/kbcheck/atv_delta.go
    op: edit
    scope: "recognize kb-native roots; treat .atv as legacy/absent without failing"
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "path-contract checks reference kb-native roots"
  - path: cmd/kbcheck/skill_validators.go
    op: edit
    scope: "skill path validation allows/expects kb-native roots"
  - path: cmd/kbcheck/report_validators.go
    op: edit
    scope: "report validators reference kb-native roots"
  - path: cmd/kbcheck/checks_test.go
    op: edit
    scope: "update/extend tests to assert kb-native roots and that .atv is not required"
  - path: cmd/kbcheck/skill_validators_test.go
    op: edit
    scope: "test kb-native path acceptance"
  - path: .gitignore
    op: edit
    scope: "replace .atv/eval-runs + .atv/pipeline-runs ignores with .kb/ ephemeral roots"
  - path: bin/kb-install.mjs
    op: edit
    scope: "install writes/expects kb-native roots; never reintroduce .atv or the drifted evolve path"
protected_oracles:
  - path: cmd/kbcheck/checks_test.go
    role: "harness path-contract oracle"
    sha256: "D3F2D509E30394E2E5485BDDFCC96ABA798A768B1A248FC21A80B04A1F2649BC"
    update_policy: "requires explicit plan update"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Update the 8 kbcheck .go touchpoints + tests, gitignore, and installer to kb-native roots; run go test ./cmd/kbcheck/..."
human_action: ""
can_continue_other_slices: true
---

# Slice 015 — kbcheck Go harness + gitignore + installer

## What to build

Make the deterministic verification harness and installer agree with the kb-native
roots, so the refactor is enforced (not just documented) and installs cannot
reintroduce atv or the evolve drift.

1. **Harness** (`cmd/kbcheck/*.go`, ~14 touchpoints across 8 files, notably
   `atv_delta.go`): update path constants/checks to the kb-native roots. Where a
   check specifically asserted `.atv/` existence, change it to assert the kb-native
   root; treat legacy `.atv/` as absent/legacy without hard-failing an installed
   repo that has not migrated (so existing consumers do not break). Rename
   `atv_delta.go`'s user-facing strings if they reference atv, keeping behavior.
2. **Tests**: update `checks_test.go`, `skill_validators_test.go`, and any test
   asserting `.atv/` to assert kb-native roots. Add a test proving the harness does
   NOT require `.atv/` and DOES recognize `docs/context/kb/` + `.kb/`. This test is
   the protected oracle — prove RED first (fails against current .atv assumption),
   then implement.
3. **gitignore**: replace `.atv/eval-runs/` + `.atv/pipeline-runs/` with the `.kb/`
   ephemeral roots (`.kb/eval-runs/`, `.kb/pipeline-runs/`, `.kb/snapshots/`,
   `.kb/qa-screenshots/`, `.kb/observations.jsonl`). Keep durable
   `docs/context/kb/` TRACKED (not ignored).
4. **Installer** (`bin/kb-install.mjs`): ensure the install path writes/creates the
   kb-native roots and NEVER rewrites skills to `.atv/` or to the drifted
   `docs/context/kb/instincts` evolve-only path (root cause of the install drift
   found in planning). If the installer previously rewrote paths per target
   (`.copilot` vs `.agents`), unify to the single kb-native root.

## Acceptance criteria

- `go test ./cmd/kbcheck/...` passes.
- New/updated test asserts kb-native roots and that `.atv/` is not required.
- `.gitignore` ignores `.kb/` ephemeral artifacts and does not ignore
  `docs/context/kb/`.
- Installer produces identical kb-native paths for every target (no per-target
  drift); a fresh install of learn+evolve references the SAME instinct root.

## Test scenarios

- RED: current tests/harness assume `.atv/`; the new oracle test fails before edits.
- GREEN: after edits, harness recognizes kb-native roots; oracle passes; full
  `go test ./cmd/kbcheck/...` green.
- Regression: `kbcheck core` still passes for an unmigrated sample repo (legacy
  `.atv/` treated as absent, not a hard error).

## Scope boundary

- Does not move the live `.atv/instincts/project.yaml` content (slice 016).

## Verification

- tdd (unit): `go test ./cmd/kbcheck/...` is the gate; protected oracle =
  `checks_test.go` kb-native path assertion.
