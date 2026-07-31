---
kb_id: kb-2026-07-30-current-agent-workflow-cleanup
slice_id: slice-005
title: "Propagate and prove the cleaned bundle"
blockers: [slice-004]
verification: verification-only
test_level: full
functional_risk: broad
model_tier: medium
model_tier_reason: "Established release mechanics apply, but changed hashes and exact stale-folder removal must be verified across three required targets."
model_requirements: ["release gates", "exact-path sync", "hash comparison", "drift diagnosis"]
escalation_triggers: ["global target contains newer useful work", "deletion scope is unresolved", "local-release reports unexplained drift"]
workspace_mode: shared-serial
conflict_domains: ["sync:global-skills", "docs:release-state"]
shared_resources: ["git:integration-owner", "sync:global-skills", "release:local-release"]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
hitl: false
expected_files:
  - path: README.md
    op: edit
    scope: "Final visible workflow and installed-skill inventory"
  - path: docs/context/architecture/skills.md
    op: edit
    scope: "Observed final skill count and ownership"
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Final proportional lifecycle"
  - path: todo.md
    op: edit
    scope: "Move active goal out after terminal proof"
  - path: todo-done.md
    op: edit
    scope: "Record compact completion summary"
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Run core, inspect target drift, sync exact folders/deletions, and run local-release."
human_action: ""
can_continue_other_slices: true
---

# Propagate and Prove the Cleaned Bundle

Close the goal only after source and every required global target agree.

## Acceptance Criteria

- Targeted tests and `core` pass before propagation.
- Newer useful global-only drift is merged back before overwrite.
- Changed skills are copied to Codex, Copilot, and shared-agent targets.
- Removed skills are deleted by exact resolved folder path from those targets.
- Hashes match and stale folders are absent.
- `local-release` passes after propagation.

## Test Scenarios

- Compare every changed `SKILL.md` hash across all required targets.
- Assert every removed skill folder is absent.
- Run `git diff --check`, `core`, and `local-release`.

## Scope Boundary

Do not inspect or modify ATV repositories.

## Completion Evidence

- Reviewed all 14 source/global drift sets; the three required global targets
  had identical older copies and no target-only files.
- Synchronized 43 required skills to Codex, Copilot, and shared-agent globals.
- Removed `ce-review`, `kb-finish`, and `klfg` from all three targets by exact
  resolved path.
- `skill-sync-report`: 129 comparisons, 0 required issues.
- `local-release`: passed all required checks.
- Proof: `docs/results/proof/current-agent-workflow-cleanup-slice-005.txt`.
