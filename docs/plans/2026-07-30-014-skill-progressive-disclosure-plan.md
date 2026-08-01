---
kb_id: kb-2026-07-30-current-agent-workflow-cleanup
slice_id: slice-004
title: "Refactor oversized hot-path skills"
blockers: [slice-003]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Prompt compaction can remove safety or ownership context unless moved contracts remain reachable at the exact phase that needs them."
model_requirements: ["progressive disclosure", "Markdown refactoring", "contract reachability", "route regression analysis"]
escalation_triggers: ["safety rules move without load cues", "any SKILL.md remains over 500 lines", "route fixtures change unintentionally"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-work", "skill:kb-plan", "skill:kb-brainstorm", "skill:kb-finalize"]
shared_resources: ["git:integration-owner", "release:core"]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck core"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Retain decisions and move deterministic mechanics to references"
  - path: .github/skills/kb-work/references
    op: create
    scope: "One-level execution, receipt, and contract references"
  - path: .github/skills/kb-plan/SKILL.md
    op: edit
    scope: "Retain slice-design judgment and move schema/templates"
  - path: .github/skills/kb-plan/references
    op: create
    scope: "One-level manifest and slice contract references"
  - path: .github/skills/kb-brainstorm/SKILL.md
    op: edit
    scope: "Retain product questioning and move artifact templates"
  - path: .github/skills/kb-brainstorm/references
    op: create
    scope: "One-level requirements and gate templates"
  - path: .github/skills/kb-finalize/SKILL.md
    op: edit
    scope: "Retain lifecycle decisions and move maintenance/cleanup details"
  - path: .github/skills/kb-finalize/references
    op: create
    scope: "One-level proof, learning, memory, and cleanup contracts"
protected_oracles:
  - path: cmd/kbcheck/skill_guidance_test.go
    role: "size and reference-reachability oracle"
    sha256: "db8b6ac30c163d724a0bed25919bcc807dddcdf5e66279b5d5b72ec9cc5e2a33"
    update_policy: "requires explicit plan update"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Refactor one hot-path skill at a time and run the guidance guard after each."
human_action: ""
can_continue_other_slices: true
---

# Refactor Oversized Hot-Path Skills

Reduce loaded skill bodies without reducing the information needed to select the
right route, stop safely, or load the next contract.

## Acceptance Criteria

- Every `SKILL.md` is below 500 lines.
- Hot-path bodies retain triggers, ownership, decisions, stop/safety rules, and
  explicit reference load cues.
- References are one level deep and navigable.
- Deterministic templates and exhaustive mechanics leave the hot path.
- Route and contract fixtures continue to pass.

## Test Scenarios

- Guidance guard reports no size or reference failures.
- Every moved contract is reachable from a named phase cue.
- `kb-plan`, `kb-work`, `kb-brainstorm`, and `kb-finalize` retain their existing
  input/output ownership in route fixtures.

## Scope Boundary

Do not redesign unrelated behavior during extraction.

## Completion Evidence

- Hot-path line counts: brainstorm 135, plan 168, work 209, finalize 176.
- `skill-guidance`, `skill-lint`, route/model-tier fixtures, and all 39 `core`
  checks pass.
- Planner model tiers now use one deterministic `small`/`medium`/`large`
  classifier; legacy `standard` fixtures were migrated to `medium`.
- Windows `.cmd` worker tests now launch through the command processor, and the
  contributor gate runs Go packages as isolated bounded commands so nested
  process-containment tests remain valid.
- Proof: `docs/results/proof/current-agent-workflow-cleanup-slice-004.txt`.
