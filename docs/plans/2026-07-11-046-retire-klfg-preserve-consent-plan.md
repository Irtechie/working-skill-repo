---
kb_id: kb-2026-07-11-invisible-workflow-ux
slice_id: slice-001
title: "Consolidate completion while preserving explicit standalone execution consent"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: medium
model_tier_reason: "The slice changes public routing, compatibility references, skill inventory, and route fixtures together."
model_requirements: ["skill routing", "cross-file consistency", "deterministic fixture interpretation"]
escalate_when: ["legacy KLFG usage has no safe kb-complete/kb-goal mapping", "installer/runtime closure depends on the physical KLFG skill"]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck route-eval"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/klfg/SKILL.md
    op: delete
    scope: "remove the redundant compatibility alias"
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "route legacy KLFG wording directly to kb-complete without advertising KLFG"
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "remove KLFG caller references while preserving the standalone plan-to-work question"
  - path: .github/skills/kb-goal/SKILL.md
    op: edit
    scope: "describe durable multi-run ownership without contrasting against KLFG"
  - path: .github/skills/kb-task/SKILL.md
    op: edit
    scope: "remove KLFG from user-facing route vocabulary"
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "absorb post-work review/proof/learning/cleanup and retain the single state-aware completion contract"
  - path: .github/skills/kb-finalize/SKILL.md
    op: delete
    scope: "remove the redundant uncommitted internal completion phase after merging its unique behavior into kb-complete"
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "stop auto-invoking a separate finalize phase; return work-to-complete state to explicit kb-complete orchestration"
  - path: .github/skills/kb-brainstorm/SKILL.md
    op: edit
    scope: "remove stale KLFG orchestration references without weakening question gates"
  - path: README.md
    op: edit
    scope: "remove KLFG from public workflow, installed-skill inventory, and historical recommendations"
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "document kb-complete for single runs and kb-goal for durable multi-run objectives"
  - path: config/skill-quality.json
    op: edit
    scope: "remove KLFG from supported route inventory and sync expectations"
  - path: config/atv-upstream-delta.json
    op: edit
    scope: "record intentional exclusion of obsolete one-shot aliases"
  - path: cmd/kbcheck/workflow_governor.go
    op: edit
    scope: "remove KLFG-specific governor claims while retaining phase-order proof"
  - path: cmd/kbcheck/skill_validators.go
    op: edit
    scope: "remove KLFG from any protected or required runtime closure"
  - path: evals/route-complexity/legacy-klfg-alias.json
    op: create
    scope: "prove legacy wording resolves directly to kb-complete"
  - path: evals/route-complexity/standalone-plan-consent.json
    op: create
    scope: "prove standalone planning stops for one explicit plan-to-work confirmation"
protected_oracles: []
status: pending
owner: agent
can_continue_other_slices: false
---

# Slice 001 - Consolidate Completion and Preserve Consent

## What To Build

Remove `klfg` and `kb-finalize` as physical skills. Route legacy full-pipeline
language directly to `kb-complete`, and merge the current post-work quality
pipeline into that command. Preserve `kb-goal` for durable objectives.

Keep this standalone planning behavior:

```text
Plan is ready: <manifest-path>
Continue with `kb-work <manifest-path>` now?
```

Do not ask when execution was already explicitly requested or delegated by an
authorized orchestrator.

## Acceptance Criteria

- `klfg` is absent from runtime skills and current public documentation.
- `kb-finalize` is absent; its review, repair, proof, learning, memory, and
  cleanup behavior is owned by `kb-complete`.
- Standalone `kb-work` cannot publish or trigger configured delivery.
- Legacy “KLFG/full pipeline” wording deterministically routes to `kb-complete`.
- `kb-goal` remains the durable multi-session governor, not the replacement for
  a single `kb-complete` run.
- Standalone `kb-plan` asks once before work.
- Explicit execution intent still chains plan to work without a redundant ask.
- Route fixtures and workflow-governor checks pass.

## Test Scenarios

- Plain “create a plan” stops after planning with one exact continuation prompt.
- “Plan and execute this” writes the manifest and continues to `kb-work`.
- “Run KLFG” routes to `kb-complete` without requiring a KLFG skill file.
- A durable “continue for days until objective X” request routes to `kb-goal`.

## Scope Boundary

Do not hide skill directories, manifests, or gate ledgers. Do not rename current
public KB commands beyond removing `klfg`.
