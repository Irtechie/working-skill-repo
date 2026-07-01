---
kb_id: kb-2026-07-01-native-scoped-learning
slice_id: slice-011
title: "Define kb-native learning contract (roots + component-scope schema)"
blockers: []
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: docs/context/architecture/kb-learning-model.md
    op: create
    scope: "the contract: canonical roots + two-tier (global + component-scoped) learning model + scope schema + pull rule"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Write the learning-model contract doc that every downstream slice implements."
human_action: ""
can_continue_other_slices: true
---

# Slice 011 — kb-native learning contract

## What to build

A single authoritative contract doc, `docs/context/architecture/kb-learning-model.md`,
that downstream slices implement. It fixes the ambiguity that caused both the atv
coupling and the too-generic learning.

Define:

1. **Canonical roots** (replacing `.atv/`):
   - Durable, git-tracked: `docs/context/kb/` — `instincts/project.yaml`,
     `instincts/scoped/<scope>.yaml`, `instincts/archive/`, `kb-completions.txt`.
   - Ephemeral, gitignored: `.kb/` — `snapshots/`, `qa-screenshots/`,
     `observations.jsonl`.
2. **Scope hierarchy (not a flat two-tier)**:
   Learning lives at the NARROWEST scope that owns it, and climbs only on recurrence:
   `workflow/domain -> project -> global`.
   - **workflow/domain** (e.g. `audio`, `image`, `video`, `motion`): the DEFAULT
     home. "If audio fails, it's an audio issue" (Mark, 2026-07-01). Most lessons
     stop here.
   - **project**: only for genuinely project-wide conventions that span multiple
     workflows within one project.
   - **global**: the RARE exception — genuinely universal process lessons (e.g.
     "read the manifest, don't hand-glob") that are not tied to any domain. Never a
     default; reached only by cross-DOMAIN recurrence.
   - A component/surface can be a sub-scope of its workflow (e.g.
     `audio/voice-eval`), inheriting the workflow scope above it.
   - **Promotion is to the NEAREST COMMON ANCESTOR, not straight to global.** Same
     pattern in `audio/tts` and `audio/sfx` -> promote to `audio`, NOT global. Same
     pattern across `audio` and `image` -> promote to `project`. Only a pattern that
     is domain-neutral AND recurs across projects reaches `global`.
3. **Scope schema**: add `scope:` to the instinct format. Default = the active
   NARROWEST scope (workflow/domain or its sub-component), NOT `project`. Scope is a
   hierarchical path (e.g. `audio`, `audio/voice-eval`, `image/comparer`). Global is
   `scope: global` (or a reserved top marker) reached only via cross-domain
   promotion. Files: `docs/context/kb/instincts/scoped/<scope-path>.yaml`, with the
   global bucket at `docs/context/kb/instincts/project.yaml` (kept for the rare
   universal + project tiers, tagged by `scope`).
4. **Pull rule**: when working in a scope, load that scope's file, its ANCESTOR
   scopes up the hierarchy (workflow, then project, then global), and nothing from
   sibling scopes. `audio/voice-eval` loads `audio/voice-eval` + `audio` + project +
   global — never `image` or `video`.
5. **Promotion-on-recurrence rule**: when the same trigger+behavior appears in >= 2
   sibling scopes, promote a generalized copy to their NEAREST COMMON ANCESTOR
   (not straight to global), citing the originating scopes as evidence. This is the
   only path a lesson climbs the hierarchy.
6. **Landmine fast-path**: a verified high-severity trap is an instant one-shot
   learn (one observation, per the existing landmine rule), recorded immediately at
   its owning scope. Small ordinary lessons do NOT get the fast-path and do NOT
   default upward.
7. **Classification update**: extend the learn feedback-routing table so ordinary
   lessons default to `scoped-instinct` at the narrowest scope; `instinct-evidence`
   (an ancestor/global tier) is reached only via promotion; `landmine-candidate` is
   the instant fast-path at the owning scope.

## Acceptance criteria

- Doc exists and names every canonical root path exactly as later slices will use.
- Doc defines the scope HIERARCHY `workflow/domain -> project -> global` and states
  the default is the narrowest owning scope (workflow/domain), not `project`.
- Doc gives the pull rule: active scope + its ancestors only; never sibling scopes.
- Doc defines promotion-to-nearest-common-ancestor on recurrence as the only climb
  path; global reached only by cross-domain recurrence.
- Doc defines the landmine fast-path (instant, at owning scope) vs ordinary lessons.
- Doc explicitly says the ATV observer hook is optional and not required.

## Scope boundary

- No skill edits in this slice — contract only. Slices 012-016 implement it.

## Verification

- verification-only: `kbcheck core` still passes (doc addition, no code change);
  the doc is the reviewable artifact.
