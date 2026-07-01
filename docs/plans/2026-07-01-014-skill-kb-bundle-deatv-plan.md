---
kb_id: kb-2026-07-01-native-scoped-learning
slice_id: slice-014
title: "De-atv the kb-* + klfg skills (snapshots/qa/observations/completions)"
blockers: [slice-011]
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: ".atv/{instincts,observations.jsonl,snapshots,kb-completions,qa-screenshots} -> kb-native roots; keep learn/evolve/compound wiring pointed at kb store"
  - path: .github/skills/kb-regression-snapshot/SKILL.md
    op: edit
    scope: ".atv/snapshots/ -> .kb/snapshots/"
  - path: .github/skills/kb-qa/SKILL.md
    op: edit
    scope: ".atv/qa-screenshots/ -> .kb/qa-screenshots/"
  - path: .github/skills/kb-repair/SKILL.md
    op: edit
    scope: ".atv/qa-screenshots + .atv/snapshots references -> .kb/ roots"
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: ".atv/snapshots references -> .kb/snapshots/"
  - path: .github/skills/kb-functional-test/SKILL.md
    op: edit
    scope: ".atv/qa-screenshots -> .kb/qa-screenshots/"
  - path: .github/skills/klfg/SKILL.md
    op: edit
    scope: "observations.jsonl reference described via kb-native path"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Rename every .atv/ reference in the kb-* + klfg skills to the kb-native roots from slice 011."
human_action: ""
can_continue_other_slices: true
---

# Slice 014 — de-atv kb-* + klfg skills

## What to build

Rename all `.atv/` references in the remaining bundle skills to the slice-011
canonical roots. This is the bulk of the "chuck atv" work (the non-learning skills).

Mapping (from slice 011):

- `.atv/instincts/...` -> `docs/context/kb/instincts/...`
- `.atv/kb-completions.txt` -> `docs/context/kb/kb-completions.txt`
- `.atv/observations.jsonl` -> `.kb/observations.jsonl`
- `.atv/snapshots/` -> `.kb/snapshots/`
- `.atv/qa-screenshots/` -> `.kb/qa-screenshots/`
- `.atv/pipeline-runs/`, `.atv/eval-runs/` -> `.kb/pipeline-runs/`, `.kb/eval-runs/`
  (if referenced)

Preserve kb-complete's semantics: it still runs learn/evolve/compound and feeds
resolved P0/P1 to the observations log — just at the kb-native paths. Where
kb-complete mentions ATV by name (e.g. "based on ATV" in kb-map-bootstrap
starter-kit-deltas), keep factual provenance references but do not keep functional
`.atv/` runtime paths.

## Acceptance criteria

- grep for `\.atv/` across `.github/skills/*/SKILL.md` returns zero FUNCTIONAL
  path references (only allowed: explicit provenance/history notes, e.g.
  starter-kit-deltas or "originally from ATV").
- kb-complete still names the learn->instincts, evolve, compound->docs/solutions
  chain, now at kb-native roots.
- No skill references a path that another skill does not also agree on (consistent
  root set).

## Scope boundary

- Does not touch the Go harness, installer, gitignore (slice 015) or migrate the
  actual state files (slice 016).

## Verification

- verification-only: `kbcheck skill-lint` passes on each edited skill; grep proves
  0 functional `.atv/` refs remain in the skills tree.
