---
type: kb-manifest
kb_id: kb-2026-07-01-native-scoped-learning
brainstorm_path: docs/context/goals/kb-native-scoped-learning.md
created: 2026-07-01
status: active
workflow_shape: "skill-bundle-change"
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-goal
    status: passed
    required_evidence:
      - "goal/source identified the two changes: de-atv (kb-native store) and component-scoped learning"
      - "current schema inspected: instinct has coarse `domain` enum + single global project.yaml, NO component scope"
      - "atv coupling quantified: 37 .atv/ skill refs + ~14 Go harness touchpoints"
      - "no unresolved ask-now or research-first items remain"
      - "safe assumptions and scope boundaries recorded"
    proof:
      - docs/context/goals/kb-native-scoped-learning.md
      - .github/skills/learn/SKILL.md
      - .atv/instincts/project.yaml
      - .github/skills/evolve/SKILL.md
      - docs/plans/2026-07-01-010-kb-native-scoped-learning-manifest.md
    evidence_notes:
      - "learn/SKILL.md L111-126 instinct schema: domain enum only, no scope field"
      - ".atv/instincts/project.yaml is a single global bucket; instinct #1 mandates domain-generic"
      - "grep: 37 .atv/ refs across .github/skills/*/SKILL.md; kbcheck .atv touchpoints in 8 .go files"
      - "User input: 'save to kb instincts and pull from there', 'refactor kb to be solid on its own', 'chuck atv', 'x pipeline doesn't need to know about learning from y pipeline unless there are pattern similarities'"
    blockers: []
    passed_at: "2026-07-01T08:22:00-04:00"
    allowed_next_action: "kb-plan docs/context/goals/kb-native-scoped-learning.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all 6 slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk"
      - "HITL classification recorded"
    proof:
      - docs/plans/2026-07-01-010-kb-native-scoped-learning-manifest.md
      - docs/plans/2026-07-01-011-skill-kb-learning-contract-plan.md
      - docs/plans/2026-07-01-012-skill-learn-native-scoped-plan.md
      - docs/plans/2026-07-01-013-skill-evolve-native-scoped-plan.md
      - docs/plans/2026-07-01-014-skill-kb-bundle-deatv-plan.md
      - docs/plans/2026-07-01-015-harness-kbcheck-installer-plan.md
      - docs/plans/2026-07-01-016-docs-state-migration-proof-plan.md
    dag_validation: "012/013/014 depend on 011; 015 depends on 012,013,014; 016 depends on 015; no missing blockers or cycles"
    blockers: []
    passed_at: "2026-07-01T08:22:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-01-010-kb-native-scoped-learning-manifest.md"
slices:
  - id: slice-011
    title: "Define kb-native learning contract (roots + component-scope schema)"
    path: docs/plans/2026-07-01-011-skill-kb-learning-contract-plan.md
    blockers: []
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Write the learning-model contract doc: canonical kb-native durable root (docs/context/kb/) + ephemeral root (.kb/), and the component-scope instinct schema."
    human_action: ""
    can_continue_other_slices: true
    notes: "enabling slice; consumers: slices 012-016"
    protected_oracles: []
  - id: slice-012
    title: "Migrate learn skill to kb-native store + scoped-instinct write path"
    path: docs/plans/2026-07-01-012-skill-learn-native-scoped-plan.md
    blockers: [slice-011]
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Edit learn/SKILL.md: repoint instinct/observation paths to kb-native roots and add the scope field + component-scoped write/pull rules."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
  - id: slice-013
    title: "Migrate evolve skill to kb-native path + scoped promotion (fix drift)"
    path: docs/plans/2026-07-01-013-skill-evolve-native-scoped-plan.md
    blockers: [slice-011]
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Edit evolve/SKILL.md: standardize to the kb-native instinct root (removes the installed .agents vs source drift) and promote component-scoped instincts into component-owned skills/docs, not only global."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
  - id: slice-014
    title: "De-atv the kb-* + klfg skills (snapshots/qa/observations/completions)"
    path: docs/plans/2026-07-01-014-skill-kb-bundle-deatv-plan.md
    blockers: [slice-011]
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Rename all .atv/ references in kb-complete, kb-qa, kb-regression-snapshot, kb-repair, kb-work, kb-functional-test, klfg to the kb-native roots; keep kb-complete's learn/evolve/compound wiring pointed at the kb store."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
  - id: slice-015
    title: "Update kbcheck Go harness + gitignore + installer (no reintroduction)"
    path: docs/plans/2026-07-01-015-harness-kbcheck-installer-plan.md
    blockers: [slice-012, slice-013, slice-014]
    verification: tdd
    test_level: unit
    functional_risk: narrow
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Update atv_delta.go + the 8 .go touchpoints and their tests to the kb-native roots; update .gitignore; update kb-install.mjs so installs never reintroduce .atv or the drifted evolve path. Run go test ./cmd/kbcheck/..."
    human_action: ""
    can_continue_other_slices: true
    notes: "deterministic gate = kbcheck Go tests"
    protected_oracles:
      - path: "cmd/kbcheck/checks_test.go"
        role: "harness path-contract oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan update"
  - id: slice-016
    title: "Migrate existing state + refresh docs/proof surface"
    path: docs/plans/2026-07-01-016-docs-state-migration-proof-plan.md
    blockers: [slice-015]
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Move .atv/instincts/project.yaml + kb-completions to the kb-native root; update README/AGENTS/PROJECT/memory-maintenance to remove atv naming and document the two-tier (global + component-scoped) model; run kbcheck core + git diff --check."
    human_action: ""
    can_continue_other_slices: true
    notes: ""
    protected_oracles: []
---

# KB: kb-native, component-scoped learning (de-atv)

## Origin

Goal: `docs/context/goals/kb-native-scoped-learning.md`

Two coupled changes surfaced from live use (the fleet-eval face-ID wrong-reference
incident showed learning was too generic to attach to the failing component):

1. **De-atv / kb-native**: the skill bundle stores learning + run state under an
   `.atv/` root and reads an optional atv hook feed. Make the bundle self-contained
   under kb-owned roots so it does not depend on the ATV install. Keep the mechanics
   adopted from ATV (learn/evolve/compound), but under KB ownership.
2. **Component-scoped learning**: the instinct schema has only a coarse `domain`
   enum and one global `project.yaml`. There is no way to attach a lesson to the
   exact component that needs it. Add a `scope` field + a component-local tier so
   `/learn` can hit the specific part (e.g. an image-comparer's calibration) and
   `/evolve` can promote scoped instincts into component-owned skills/docs.
   **Scoped-by-default** (Mark, 2026-07-01): "x pipeline doesn't need to know about
   learning from y pipeline unless there are pattern similarities." Ordinary lessons
   default to their component scope; they reach global ONLY via cross-scope
   recurrence (same pattern in >= 2 scopes). Verified high-severity landmines are an
   instant one-shot learn at their owning scope; small lessons never default global.

## Workflow Shape

`skill-bundle-change` — edits several portable KB skills + the Go verification
harness + installer + visible docs. No new app runtime.

## Canonical roots (decided; recorded as safe assumption)

- Durable, git-tracked knowledge -> `docs/context/kb/` (instincts, kb-completions,
  scoped instincts). Matches the direction the installed evolve already drifted to.
- Ephemeral, gitignored run artifacts -> `.kb/` (snapshots, qa-screenshots,
  observations.jsonl), replacing `.atv/`.
- Component-scoped instincts -> `docs/context/kb/instincts/scoped/<scope>.yaml`,
  plus an optional `scope:` field on any instinct. **Default scope is the active
  component, not `project`.** Pull rule: when working a component, load that
  component's scoped instincts + global; never other components' scopes. A lesson
  reaches global (`scope: project`) only via promotion-on-recurrence (same pattern
  in >= 2 scopes). Verified high-severity landmines are instant one-shot, scoped.

## Scope boundary

- This plan changes the skill bundle only. Applying component-scoped learning to a
  specific downstream component (e.g. the llmcommune fleet-eval image comparer's
  `calibration.yaml` + fixtures) is a separate, downstream consumer task in that
  repo — named here, not executed here.
- The optional ATV observer hook (`.github/hooks/copilot-hooks.json`) is not
  reintroduced. Passive tool-use capture becomes optional, not required.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|-------|------------|--------------|------|--------|
| 011 | kb-native learning contract (roots + scope schema) | - | verification-only | no | pending |
| 012 | learn: kb-native store + scoped write | 011 | verification-only | no | pending |
| 013 | evolve: kb-native path + scoped promotion (fix drift) | 011 | verification-only | no | pending |
| 014 | de-atv kb-* + klfg (snapshots/qa/observations/completions) | 011 | verification-only | no | pending |
| 015 | kbcheck Go harness + gitignore + installer | 012,013,014 | tdd (unit) | no | pending |
| 016 | migrate state + docs/proof surface | 015 | verification-only | no | pending |

## Goal Link

Update `docs/context/goals/kb-native-scoped-learning.md` as each slice finishes.
Goal stays active until kbcheck core + go test pass and the docs proof is recorded.
