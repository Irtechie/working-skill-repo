---
kb_id: kb-2026-07-01-native-scoped-learning
slice_id: slice-013
title: "Migrate evolve skill to kb-native path + scoped promotion (fix drift)"
blockers: [slice-011]
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: .github/skills/evolve/SKILL.md
    op: edit
    scope: "standardize instinct root to docs/context/kb/instincts (fix source vs installed .agents drift); add scoped-instinct promotion into component-owned skills/docs"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Edit evolve/SKILL.md per the contract from slice 011."
human_action: ""
can_continue_other_slices: true
---

# Slice 013 — evolve: kb-native path + scoped promotion

## What to build

Update `.github/skills/evolve/SKILL.md` to the slice-011 contract and resolve the
drift discovered during planning:

- Source evolve reads `.atv/instincts/project.yaml`.
- Installed `.agents` evolve reads `docs/context/kb/instincts/project.yaml`.

The kb-native target (`docs/context/kb/instincts/`) is canonical, so making source
match it removes the drift permanently.

Changes:

1. **Paths**: `.atv/instincts/project.yaml` -> `docs/context/kb/instincts/project.yaml`;
   `.atv/instincts/archive/` -> `docs/context/kb/instincts/archive/`. Also read
   `docs/context/kb/instincts/scoped/<scope>.yaml`.
2. **Scoped promotion**: when a mature instinct is scoped to a component, `/evolve`
   promotes it into that component's owned skill/doc surface (e.g. a component
   SKILL.md or the component's config/calibration doc), NOT only a global
   `.github/skills/<x>`. Document the decision rule: global instinct -> global
   skill; component instinct -> component-owned surface.
3. **Candidate filter**: unchanged thresholds (confidence > 0.85, obs > 5, active
   within 90 days), but evaluated per scope file too.

## Acceptance criteria

- No `.atv/` path remains in evolve/SKILL.md.
- evolve reads the same kb-native root that learn writes (round-trip consistent).
- Scoped promotion path is documented and distinct from global promotion.

## Scope boundary

- Does not migrate the actual project.yaml file (slice 016) or touch the harness.

## Verification

- verification-only: `kbcheck skill-lint` on evolve passes. grep proves learn and
  evolve now reference the identical instinct root.
