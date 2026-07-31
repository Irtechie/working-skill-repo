# Session Portfolio Convergence

Status: active
Created: 2026-07-30
Last updated: 2026-07-30

## Objective

Converge the three named working-skill-repo donor sessions without losing unique valuable work, then deliver the reviewed result to the synchronized remote default branch.

## Done Criteria

- Every named donor has an exact retained, duplicate, superseded, or generated-residue disposition.
- Accepted donor work is integrated on a fresh topic branch without editing donor or default checkouts.
- Repo-local contributor and release gates pass.
- Required Codex, Copilot, and shared-agents skill copies match the approved repo source.
- The delivered commit is reachable from `origin/main`.

## Terminal Proof

- `go run ./cmd/kbcheck core`
- `go run ./cmd/kbcheck local-release`
- Structured KB code review with no unresolved P0/P1 findings.
- PR delivery and merge evidence.
- `git fetch origin; git merge-base --is-ancestor <delivered-commit> origin/main`

## Done Check

- Type: gate
- Check: `go run ./cmd/kbcheck local-release` followed by remote-default ancestor proof for the delivered commit
- Expected result: both commands exit 0
- Why sufficient: proves repo quality and required skill propagation, then proves the reviewed tree reached the authoritative default branch

## Current State

- Current artifact: donor inventory for sessions `25da36bb`, `655154e3`, and `2172ad69`
- Next allowed action: `kb-plan docs/context/goals/session-portfolio-convergence.md`
- Last proof: stable shared claim `session-portfolio-convergence` owned by project session `ab858389-7f91-404b-84f0-805fa548d629`

## Work Units

| Unit | Route | Artifact | Status | Proof |
|---|---|---|---|---|
| Donor inventory and classification | `kb-plan` | three donor commits and PR history | active | Git ancestry, diff, and review evidence |
| Accepted patch integration | `kb-work` | convergence manifest | pending | focused tests |
| Final review, propagation, and delivery | `kb-complete` | convergence manifest and PR | pending | `local-release`, review, merge, ancestor proof |

## Blockers

| Blocker | Type | Owner | Resume Condition |
|---|---|---|---|
| None | - | - | - |

## Notes

- Stable work ID: `session-portfolio-convergence`
- Donor worktrees are read-only evidence sources.
