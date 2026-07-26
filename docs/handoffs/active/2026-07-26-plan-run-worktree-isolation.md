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
- Slice-001 is done: the new `plan-worktree` CLI owns idempotent manifest
  receipts, immutable base identity, explicit non-default integration refs,
  dirty-source preservation, and non-force release refusal.
- Focused and full `cmd/kbcheck` proof pass; the protected oracle hash is
  `82645d4bdbce01a665076e4dbb24794e90108b32e8a858bb9a2673c90a26b6db`.
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

1. Start slice-002 from the owning plan-run worktree.
2. Extend common-directory ownership with atomic manifest-level conflict and
   shared-resource claims.
3. Keep execution shared-serial until the scheduler and plan-run integration
   path prove safe concurrency.

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
