---
type: kb-manifest
kb_id: kb-2026-05-30-eval-map
brainstorm_path: docs/brainstorms/2026-05-30-kb-eval-map-requirements.md
created: 2026-05-30
status: completed
slices:
  - id: slice-001
    title: "Add kb-eval-map skill"
    path: docs/plans/2026-05-30-001-skill-kb-eval-map-plan.md
    blockers: []
    verification: verification-only
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
    notes: "scope-check: forecast=1 changed=1 discovered=1 unexplained=0; scope-discovery: todo.md - board sync required by kb-work; proof: scripts/skill-lint.ps1 exit=0 warnings=19"
  - id: slice-002
    title: "Wire eval mapping into bootstrap and docs"
    path: docs/plans/2026-05-30-002-bootstrap-eval-map-plan.md
    blockers: [slice-001]
    verification: verification-only
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
    notes: "scope-check: forecast=4 changed=8 discovered=4 unexplained=0; scope-discovery: docs/context/eval-map.md - eval map for this skill repo; scope-discovery: docs/context/PROJECT.md - map refresh; scope-discovery: AGENTS.md and .github/copilot-instructions.md - memory file list refresh; proof pending final kb-check -All and git diff --check"
  - id: slice-003
    title: "Propagate and verify eval-map skill bundle"
    path: docs/plans/2026-05-30-003-sync-eval-map-plan.md
    blockers: [slice-001, slice-002]
    verification: verification-only
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
    notes: "scope-check: forecast=4 changed=4 discovered=0 unexplained=0; proof: copied kb-eval-map and kb-map-bootstrap to required Codex/Copilot/agents/ATV targets; hashes matched; kb-check -All exit=0; git diff --check exit=0 in working repo and ATV"
---

# KB: Eval Map

## Origin

Brainstorm: `docs/brainstorms/2026-05-30-kb-eval-map-requirements.md`

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Add kb-eval-map skill | - | verification-only | no | done |
| 2 | Wire eval mapping into bootstrap and docs | slice-001 | verification-only | no | done |
| 3 | Propagate and verify eval-map skill bundle | slice-001, slice-002 | verification-only | no | done |

## Dependency DAG

```text
slice-001 -> slice-002 -> slice-003
```

## Verification Strategy

Run the skill repo deterministic quality gate:

```powershell
.\.github\skills\kb-check\scripts\kb-check.ps1 -All
git diff --check
```

Also verify required sync targets after propagation with
`scripts\skill-sync-report.ps1`.
