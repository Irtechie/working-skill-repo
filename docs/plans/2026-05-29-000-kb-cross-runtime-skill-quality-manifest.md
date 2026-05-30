---
type: kb-manifest
kb_id: kb-2026-05-29-cross-runtime-skill-quality
brainstorm_path: docs/brainstorms/2026-05-29-cross-runtime-skill-quality-requirements.md
created: 2026-05-29
status: reviewed
scope-verified-files:
  - config/skill-quality.json
  - scripts/skill-lint.ps1
  - scripts/route-complexity-eval.ps1
  - scripts/skill-sync-report.ps1
  - .github/skills/kb-check/scripts/kb-check.ps1
  - evals/route-complexity/README.md
  - evals/route-complexity/tiny-typo-fix.json
  - evals/route-complexity/known-failing-test.json
  - evals/route-complexity/unclear-broken-behavior.json
  - evals/route-complexity/bounded-feature-fast.json
  - evals/route-complexity/stale-handoff.json
  - evals/route-complexity/broad-migration.json
  - evals/route-complexity/release-ship-flow.json
  - evals/route-complexity/cross-runtime-instruction-update.json
  - AGENTS.md
  - .github/copilot-instructions.md
  - README.md
  - docs/context/PROJECT.md
  - docs/context/operations/testing.md
  - docs/context/memory-maintenance.md
  - docs/context/research/2026-05-29-skill-repo-gap-audit.md
  - .atv/kb-completions.txt
slices:
  - id: slice-001
    title: "Define cross-runtime quality contract"
    path: docs/plans/2026-05-29-001-config-cross-runtime-quality-contract-plan.md
    blockers: []
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=2 changed=2 discovered=0 unexplained=0; proof: ConvertFrom-Json config/skill-quality.json + git diff --check; memory-impact: durable; areas=testing,quality-contract; docs=docs/context/operations/testing.md"
  - id: slice-002
    title: "Add deterministic skill lint"
    path: docs/plans/2026-05-29-002-tool-skill-lint-plan.md
    blockers: [slice-001]
    verification: functional-cli
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=2 changed=2 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1 exit=0 warnings=19; memory-impact: operational"
  - id: slice-003
    title: "Add route and complexity eval fixtures"
    path: docs/plans/2026-05-29-003-eval-route-complexity-plan.md
    blockers: [slice-001]
    verification: functional-cli
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=11 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/route-complexity-eval.ps1 exit=0 fixtures=8; memory-impact: operational"
  - id: slice-004
    title: "Wire skill quality into kb-check"
    path: docs/plans/2026-05-29-004-tool-kb-check-skill-quality-plan.md
    blockers: [slice-002, slice-003]
    verification: functional-cli
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=2 changed=2 discovered=0 unexplained=0; proof: .github/skills/kb-check/scripts/kb-check.ps1 -List and -All exit=0; memory-impact: durable; areas=testing; docs=docs/context/operations/testing.md"
  - id: slice-005
    title: "Add read-only sync drift report"
    path: docs/plans/2026-05-29-005-tool-sync-drift-report-plan.md
    blockers: [slice-001]
    verification: functional-cli
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=3 discovered=0 unexplained=0; proof: scripts/skill-sync-report.ps1 exit=0 required_issues=0 and kb-check -All exit=0; memory-impact: durable; areas=testing,sync; docs=docs/context/operations/testing.md"
  - id: slice-006
    title: "Document canonical quality workflow"
    path: docs/plans/2026-05-29-006-docs-quality-workflow-plan.md
    blockers: [slice-004, slice-005]
    verification: verification-only
    test_level: none
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=7 discovered=2 unexplained=0; scope-discovery: docs/context/memory-maintenance.md - stale maintenance signals closed; scope-discovery: docs/context/research/2026-05-29-skill-repo-gap-audit.md - historical no-harness finding updated; proof: kb-check -All exit=0 + git diff --check exit=0; memory-impact: durable; areas=docs,testing,repo-memory; kb-map-refresh: done"
---

# KB: Cross-Runtime Skill Quality

## Origin

Brainstorm: `docs/brainstorms/2026-05-29-cross-runtime-skill-quality-requirements.md`

## Slice Overview

| # | Slice | Blocked By | Verification | Test Level | HITL | Status |
|---|---|---|---|---|---|---|
| 1 | Define cross-runtime quality contract | - | verification-only | none | no | done |
| 2 | Add deterministic skill lint | slice-001 | functional-cli | functional-cli | no | done |
| 3 | Add route and complexity eval fixtures | slice-001 | functional-cli | functional-cli | no | done |
| 4 | Wire skill quality into kb-check | slice-002, slice-003 | functional-cli | functional-cli | no | done |
| 5 | Add read-only sync drift report | slice-001 | functional-cli | functional-cli | no | done |
| 6 | Document canonical quality workflow | slice-004, slice-005 | verification-only | none | no | done |

## Assumptions

- PowerShell is the first implementation target because the existing `kb-check` helper is PowerShell and the repo is currently Windows-oriented.
- Route eval fixtures can be deterministic metadata checks first; model-scored cross-model evals are out of scope for this manifest.
- GHCP compatibility means the artifacts must be readable and useful from GitHub Copilot/Copilot Chat contexts even when Codex-only tool APIs are unavailable.

## Completion Criteria

- All slices are done or intentionally skipped.
- `kb-check` discovers and can run skill-quality checks in this repo.
- The route-complexity fixture suite includes Codex/GHCP applicability metadata.
- The sync report is read-only and clearly separates required targets from optional omissions.

## Completion Review

- Review: P0=0 P1=0 P2=0 P3=0 from local completion review. Subagent review was not run because this session does not have explicit user authorization for parallel agents.
- Follow-up resolution: resolved 1 runner resilience issue in `scripts/route-complexity-eval.ps1`.
- Proof: `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`, PowerShell parser check for all scripts, JSON summary checks for all three new runners, and `git diff --check`.
- Compound: skipped - the durable workflow is already captured in `README.md`, `docs/context/operations/testing.md`, and this manifest.
- Learn: no new project instinct written; existing cross-project instinct still applies.
- Evolve: skipped; completion count is 1, not divisible by 5.
- Project memory: refreshed in `docs/context/PROJECT.md`, `docs/context/operations/testing.md`, `docs/context/memory-maintenance.md`, and `todo.md`.
- Memory maintenance: closed repeated-rediscovery and stale-doc signals; ATV optional target decision remains open.
- Cleanup: active handoff moved to `docs/handoffs/done/`.
