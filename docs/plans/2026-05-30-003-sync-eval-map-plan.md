---
kb_id: kb-2026-05-30-eval-map
slice_id: slice-003
title: "Propagate and verify eval-map skill bundle"
blockers: [slice-001, slice-002]
verification: verification-only
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: C:/Users/marowe/.codex/skills/kb-eval-map/SKILL.md
    op: create
    scope: "sync approved skill to Codex global install"
  - path: C:/Users/marowe/.copilot/skills/kb-eval-map/SKILL.md
    op: create
    scope: "sync approved skill to Copilot global install"
  - path: C:/Users/marowe/.agents/skills/kb-eval-map/SKILL.md
    op: create
    scope: "sync approved skill to shared agents global install"
  - path: E:/all-the-vibes/.github/skills/kb-eval-map/SKILL.md
    op: create
    scope: "sync approved skill to ATV .github skills"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 003: Propagate And Verify Eval-Map Skill Bundle

## What To Build

Propagate the approved `kb-eval-map` skill to required runtime targets and verify
the skill repo remains clean.

## Acceptance Criteria

- Required Codex, Copilot, agents, and ATV `.github` skill targets contain the
  same `kb-eval-map/SKILL.md` hash as the working repo.
- `scripts/skill-sync-report.ps1` reports zero required issues.
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` exits 0.
- `git diff --check` exits 0 in touched repos.

## Verification

- Run `scripts/skill-sync-report.ps1`.
- Run `.\.github\skills\kb-check\scripts\kb-check.ps1 -All`.
- Run `git diff --check` in `E:\working-skill-repo` and `E:\all-the-vibes`.

Result:

- Required target hashes match for `kb-eval-map` and `kb-map-bootstrap`.
- `scripts/skill-sync-report.ps1` reported 0 required issues.
- `.\.github\skills\kb-check\scripts\kb-check.ps1 -All` exited 0.
- `git diff --check` exited 0 in `E:\working-skill-repo` and
  `E:\all-the-vibes`.
