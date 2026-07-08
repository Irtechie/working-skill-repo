---
kb_id: kb-2026-07-05-model-agnostic-planner-economy
slice_id: slice-005
title: "Tighten custom-instruction segmentation and first adapter boundary"
blockers: [slice-002, slice-003]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: large
model_tier_reason: "This decides where custom instructions, commands, skills, agents, subagents, tools, and host adapters are allowed to own behavior."
hitl: false
expected_files:
  - path: docs/context/decisions/2026-07-05-surface-segmentation-and-adapter-boundary.md
    op: create
    scope: "record custom-instruction/command/skill/agent/subagent/tool ownership and the first adapter boundary"
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "route ambiguous work without making commands or agents own orchestration"
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "define subagent packet execution and escalation boundaries"
  - path: cmd/kbcheck/surface_report.go
    op: edit
    scope: "lint or report obvious ownership leaks where practical"
  - path: cmd/kbcheck/surface_report_test.go
    op: edit
    scope: "cover custom instructions, thin commands, workflow skills, specialist agents, and deterministic tools"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: "slice-003 done"
next_agent_action: "Document the segmentation contract and implement a small static inventory/report."
human_action: ""
can_continue_other_slices: true
---

# Slice 005 - Custom Instructions, Segmentation, and Adapter Boundary

## What To Build

Make ownership crisp enough that cheap workers get packets, not broad planning
authority.

Use this split:

| Surface | Owns |
|---|---|
| Custom instruction | ambient host/project policy, invariants, and always-loaded constraints |
| Command | host/user entrypoint and arguments |
| Skill | workflow policy, gates, artifacts, escalation, proof |
| Agent | narrow capability/persona and evidence rules |
| Subagent | one runtime invocation with one packet and one result |
| Tool | deterministic side effect or query with compact output |
| Adapter | host/runtime mechanics outside the core planning contract |

## Acceptance Criteria

- A decision note states the ownership contract.
- `kb-start` and `kb-work` point to the same segmentation model.
- Custom instructions such as `AGENTS.md`, `CLAUDE.md`, or Copilot instructions
  carry stable repo/runtime invariants, not live task state or full workflow
  bodies.
- The first adapter boundary is for the runtime the user actually runs daily.
- The core packet schema does not require a Claude-, Codex-, or HumanLayer-only
  field.
- `kbcheck` can at least report obvious surface inventory risks, even if full
  enforcement stays future work.

## Scope Boundary

Do not build every adapter now. Build the boundary for one daily runtime and
make the second adapter a later proof of generality.

## Verification

Run:

```shell
go test ./cmd/kbcheck/...
go run ./cmd/kbcheck core
```
