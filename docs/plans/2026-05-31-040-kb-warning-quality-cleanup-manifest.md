---
type: kb-manifest
kb_id: kb-2026-05-31-warning-quality-cleanup
brainstorm_path: inline user request 2026-05-31
created: 2026-05-31
status: reviewed
workflow_shape: "skill-bundle-change"
slices:
  - id: slice-001
    title: "Add missing argument hints"
    path: docs/plans/2026-05-31-041-skill-argument-hints-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add low-risk argument-hint frontmatter to older skills that lint currently warns about."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=7 changed=7 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; proof-result: missing argument-hint warnings removed; memory-impact: none"
    protected_oracles: []
  - id: slice-002
    title: "Codify local review fallback"
    path: docs/plans/2026-05-31-042-review-local-fallback-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Make kb-review/kb-complete truthful when reviewer subagents are unavailable."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=2 changed=2 discovered=0 unexplained=0; proof: rg -n local-fallback/review-mode .github/skills/kb-review/SKILL.md .github/skills/kb-complete/SKILL.md; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-lint.ps1; memory-impact: durable"
    protected_oracles: []
  - id: slice-003
    title: "Compact optional sync warnings"
    path: docs/plans/2026-05-31-043-sync-report-optional-summary-plan.md
    blockers: []
    verification: integration
    test_level: cli
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Keep optional ATV drift visible but stop flooding default sync output."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=2 changed=2 discovered=0 unexplained=0; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-sync-report.ps1; proof: powershell -ExecutionPolicy Bypass -File scripts/skill-sync-report.ps1 -VerboseOptional; proof-result: default optional warnings summarized, verbose switch prints per-skill rows; memory-impact: operational; docs=docs/context/operations/testing.md"
    protected_oracles: []
---

# KB: Warning Quality Cleanup

## Origin

User requested a surgical `kb-plan -> kb-work -> kb-complete` pass for the
remaining accepted warnings.

## Workflow Shape

`skill-bundle-change` - edits skill contracts, lint output, sync reporting, and
docs; no new harness architecture.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Add missing argument hints | - | integration | no | done |
| 2 | Codify local review fallback | - | integration | no | done |
| 3 | Compact optional sync warnings | - | integration | no | done |

## Done Criteria

- Missing `argument-hint` warnings are removed without changing workflow behavior.
- `kb-review` and `kb-complete` no longer imply multi-agent review when the
  runtime cannot spawn reviewers.
- `skill-sync-report` reports optional ATV scaffold/plugin differences compactly
  by default while preserving detailed output behind an explicit switch.
- Canonical gate, diff check, and required sync report pass.

## Completion Review

- Review: P0=0 P1=0 P2=0 P3=0.
- Review mode: local-fallback. No subagent reviewers were spawned.
- Follow-up resolution: resolved 0, logged 0, blocked 0.
- Proof: `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`,
  `git diff --check`, and `scripts\skill-sync-report.ps1` passed.
- Learn: skipped - warning cleanup only; no repo-specific landmine or reusable
  implementation pattern.
- Evolve: skipped - completion count not divisible by 5.
- Project memory: refreshed operational testing note.
- Memory maintenance: no new signals.
- Compact: skipped - long-skill warnings are a separate measured trim pass.
