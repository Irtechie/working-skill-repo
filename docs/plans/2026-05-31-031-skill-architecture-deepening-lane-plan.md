---
kb_id: kb-2026-05-31-lazy-lane-consolidation
slice_id: slice-001
title: "Create architecture deepening lazy lane"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-architecture-deepening/SKILL.md
    op: add
    scope: "Compact lazy lane if route fixtures prove distinct value."
  - path: evals/route-complexity/
    op: edit
    scope: "Add fixture distinguishing architecture deepening from deslop/review."
  - path: .github/agents/thermo-nuclear-code-quality-reviewer.agent.md
    op: edit
    scope: "Optionally add deletion-test language only if still accurate."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Add route fixture first, then compact skill if it earns a lane."
human_action: ""
can_continue_other_slices: true
---

# Create Architecture Deepening Lazy Lane

## What To Build

Add a compact architecture-deepening lane only if route fixtures prove it is
distinct from cleanup and review. It should help identify where a codebase needs
deeper modules, smaller interfaces, or better locality.

## Acceptance Criteria

- Route fixture sends architecture exploration to the new lane.
- Cleanup/deslop prompts do not route to architecture deepening.
- Review prompts do not route to architecture deepening.
- Skill stays compact and lazy.
- Thermo-nuclear review only borrows deletion-test language if useful.

## Test Scenarios

- Run route-complexity eval.
- Run skill lint.
- Run `kb-check -All`.

## Scope Boundary

Do not put this in `kb-memory-review`. Do not vendor large third-party text.
