---
kb_id: kb-2026-07-09-phoenix-routing-slicing-absorption
slice_id: slice-003
title: "Wire done/proof checks into KB skills"
blockers: [slice-002]
verification: integration
test_level: integration
functional_risk: none
model_tier: medium
model_route: hosted-sonnet
proof_check:
  kind: command_exit
  command: "rg -n \"objective_contract|done_check|proof_check|model_route\" .github/skills/kb-plan/SKILL.md .github/skills/kb-work/SKILL.md .github/skills/kb-goal/SKILL.md .github/skills/kb-complete/SKILL.md .github/skills/kb-gate/SKILL.md docs/context/eval-map.md"
  expect: 0
hitl: false
expected_files:
  - path: ".github/skills/kb-plan/SKILL.md"
    op: edit
    scope: "Emit done_check and proof_check fields in manifest and slice templates."
  - path: ".github/skills/kb-work/SKILL.md"
    op: edit
    scope: "Require proof_check before slice completion when the contract is active."
  - path: ".github/skills/kb-goal/SKILL.md"
    op: edit
    scope: "Require a top-level objective done_check for long-running goals or a recorded exception."
  - path: ".github/skills/kb-complete/SKILL.md"
    op: edit
    scope: "Collect done_check/proof_check evidence before final completion."
  - path: ".github/skills/kb-gate/SKILL.md"
    op: edit
    scope: "Treat missing objective checks as gate evidence gaps when the contract is active."
  - path: "docs/context/eval-map.md"
    op: edit
    scope: "Document the new proof surfaces and canonical commands."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Update skill contracts after the Go validator can enforce the fields."
human_action: ""
can_continue_other_slices: true
---

# Slice 003: Wire Done/Proof Checks Into KB Skills

## What This Delivers

The new schema becomes part of normal KB planning and completion, not just a
validator feature.

## Acceptance Criteria

- New `kb-plan` manifests include a top-level `done_check` when work is
  goal-like, autonomous, or long-running.
- New slice plans include `proof_check` or an explicit no-check reason.
- `kb-work` blocks done status when an active-contract slice has no proof check
  evidence.
- `kb-goal` does not start a durable objective without a done check or an
  explicit human-approved exception.
- `kb-complete` summarizes done/proof check evidence before declaring work done.

## Test Scenarios

- Run the relevant `kbcheck` validator fixture from slice 002 against a plan
  emitted by updated `kb-plan`.
- Search proves `done_check` and `proof_check` appear in the skill templates and
  completion gates.

## Scope Boundary

No external Phoenix runtime and no Phoenix lifecycle skill names.
