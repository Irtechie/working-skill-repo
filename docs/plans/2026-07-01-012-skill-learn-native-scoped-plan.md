---
kb_id: kb-2026-07-01-native-scoped-learning
slice_id: slice-012
title: "Migrate learn skill to kb-native store + scoped-instinct write path"
blockers: [slice-011]
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: .github/skills/learn/SKILL.md
    op: edit
    scope: "repoint .atv/instincts + .atv/observations.jsonl to kb-native roots; add scope field to instinct format; add component-instinct classification route + scoped write/pull rules"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Edit learn/SKILL.md per the contract from slice 011."
human_action: ""
can_continue_other_slices: true
---

# Slice 012 — learn: kb-native store + scoped write

## What to build

Update `.github/skills/learn/SKILL.md` to the slice-011 contract.

Changes:

1. **Paths**: every `.atv/instincts/project.yaml` -> `docs/context/kb/instincts/project.yaml`;
   `.atv/instincts/archive/` -> `docs/context/kb/instincts/archive/`;
   `.atv/observations.jsonl` -> `.kb/observations.jsonl`. (14 instinct + 6
   observation references in this file family.)
2. **Scope field**: add `scope:` to the instinct YAML format block. Default =
   the ACTIVE component scope (not `project`). A component lesson gets
   `scope: <component>` and is written to
   `docs/context/kb/instincts/scoped/<component>.yaml`. `scope: project` (global)
   is used only after cross-scope promotion.
3. **Classification routes**: update the feedback-routing table so ordinary lessons
   default to `component-instinct` (scoped), `instinct-evidence` (global) is reached
   only via promotion-on-recurrence, and `landmine-candidate` is the instant
   one-shot fast-path recorded at the owning scope. Small lessons must never default
   to global.
4. **Promotion-on-recurrence**: add the rule that when the same trigger+behavior
   appears in >= 2 distinct scoped files, learn writes a generalized global instinct
   citing the originating scopes. This is the only path scoped -> global.
5. **Decay + cap**: apply the existing decay and the 50-instinct cap PER scope file
   (global bucket and each scoped file each capped), so scoped learning cannot be
   crowded out by global.
6. **Pull rule**: `/learn` reads the active scope's file + global only; never other
   components' scoped files.

## Acceptance criteria

- No `.atv/` path remains in learn/SKILL.md except an explicit historical/provenance
  note if needed.
- Instinct format block shows `scope:` with default = active component and the
  scoped-file convention; global only via promotion.
- Routing table: ordinary lessons default to `component-instinct`; `instinct-evidence`
  (global) only via promotion-on-recurrence; `landmine-candidate` is instant/scoped.
- Promotion-on-recurrence rule (>= 2 scopes -> global) is documented in learn.
- Observation feed path is `.kb/observations.jsonl` and described as optional.

## Scope boundary

- Does not edit evolve, kb-*, harness, or installer (later slices).

## Verification

- verification-only: `kbcheck skill-lint` on learn passes (0 errors; known size
  warnings acceptable). grep proves 0 `.atv/` hard refs remain in the file.
