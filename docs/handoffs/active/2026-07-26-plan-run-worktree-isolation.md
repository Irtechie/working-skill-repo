# Plan-Run Worktree Isolation Handoff

Created: 2026-07-26
Last refreshed: 2026-07-26
Status: active
Suggested route: kb-work

## Intent

Execute the five-slice plan that gives each mutating manifest its own plan-run
branch/worktree, adds cross-manifest conflict claims, integrates child receipts
only into the plan-run branch, and keeps default-branch delivery outside
`kb-work`.

## Current State

- Manifest contract passes with zero issues.
- All five context packets validate.
- `plan-to-work` passed after the overlapping DDR scope was contained.
- No implementation file has yet been changed for this workstream.
- The shared `main` checkout has overlapping uncommitted DDR changes in:
  - `.github/skills/kb-work/SKILL.md`
  - `.github/skills/kb-work/references/execution-prompt.md`
  - `cmd/kbcheck/ddr_contract_test.go`
  - `docs/context/architecture/kb-workflow.md`
- The coherent seven-file DDR scope was independently reviewed and committed on
  `codex/ddr-route-announcement-containment` at `3f1d916`; its worktree is clean.
- A separate change-aware proof-governor plan appeared in the shared checkout
  during preflight; preserve its files and board row as another owner's work.

## Next Agent Action

1. Start slice-001 from the passed manifest gate.
2. Create the manifest-owned plan-run branch/worktree without mutating or
   integrating into the dirty default checkout.
3. Keep bootstrap execution shared-serial until the new scheduler and plan-run
   integration path prove safe concurrency.

## Human Required

None for slice-001. Default-branch delivery, pushes, destructive cleanup, and
force operations remain separately unauthorized.

## Pointers

- Project map: `docs/context/PROJECT.md`
- Manifest: `docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md`
- DDR owner manifest: `docs/plans/2026-07-26-000-kb-ddr-route-announcement-manifest.md`
- Harness recovery: `docs/plans/2026-07-20-000-kb-harness-validation-recovery-manifest.md`
- Todo: `todo.md`

## Staleness Check

Recheck the exact Git status, registered worktrees, active board, and containment
branches before resuming.

## Completion Criteria

All five slices have passing terminal gates, the disposable multi-plan lifecycle
passes, project/global skill drift is reviewed and resolved, and `core` plus
`local-release` pass without touching or delivering the default branch.
