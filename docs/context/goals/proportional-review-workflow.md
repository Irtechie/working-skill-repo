# Proportional Review Workflow

Status: active
Created: 2026-07-30
Last updated: 2026-07-30

## Objective

Align the complete KB skill bundle with current agent guidance: proportional
review, single-agent-first routing, progressive disclosure, bounded loops,
conditional memory, and removal of dead compatibility skills.

## Done Criteria

- No reviewer agent runs after an individual slice.
- Routine integrated work launches at most one broad semantic reviewer.
- Pre-slice and specialist reviewers require concrete uncertainty or risk.
- Dead duplicate review skills and all enforced references to them are removed.
- Learning and memory work run only when their trigger is present.
- Every remaining skill is audited against the current guidance rubric.
- No `SKILL.md` exceeds 500 lines.
- Repo and required global skill copies agree.

## Terminal Proof

- Targeted review-policy and reference-guard tests pass.
- `go run ./cmd/kbcheck core` passes.
- `go run ./cmd/kbcheck local-release` passes after required skill propagation.

## Done Check

- Type: command_exit
- Check: `go run ./cmd/kbcheck local-release`
- Expected result: exit code 0
- Why sufficient: It composes repo checks, diff hygiene, static review reports,
  and blocking drift checks against every required global skill target.

## Current State

- Current artifact: `docs/context/epics/current-agent-workflow-cleanup.md`
- Next allowed action: plan every epic workstream
- Last proof: call-graph inspection and external research captured in
  `docs/context/research/2026-07-30-proportional-agent-code-review.md`

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Requirements and plan | `kb-plan` | proportional review brainstorm | active | manifest contract |
| Workflow implementation | `kb-work` | pending manifest | pending | slice proof |
| Finalization and propagation | `kb-complete` | pending manifest | pending | local-release |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|

## Notes

- Git history is the recovery path for intentionally removed skills.
- Review dimensions are mandatory coverage; separate reviewer agents are not.
