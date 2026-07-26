# Plan-Run Worktree Isolation Handoff

Created: 2026-07-26
Last refreshed: 2026-07-26
Status: done
Suggested route: none; delivery remains separately authorized

## Intent

Execute the five-slice plan that gives each mutating manifest its own plan-run
branch/worktree, adds cross-manifest conflict claims, integrates child receipts
only into the plan-run branch, and keeps default-branch delivery outside
`kb-work`.

## Final State

- All five slices and terminal gates passed.
- One manifest group owns one worktree; every slice is shared-serial in it.
  Per-slice worktrees are rejected for active plan runs.
- Manifest and slice leases reject overlapping live path, domain, and resource
  claims before mutation.
- Sequential slice commits advance only the owning plan-run head after exact
  worktree/ref/owner identity, compare-and-swap, claim checks, and proof replay.
- Accepted proof bytes are archived immutably and revalidated at completion.
- Default-branch delivery and dirty-WIP checkpoint authority remain outside
  `kb-work`; absent policy is local-only.
- The disposable two-manifest selftest, full command package, required skill
  sync report, contributor core, and local release gate passed.
- Three independent reviewer lanes finished with no unresolved P0/P1/P2.
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

None. The local topic branch is the completed candidate. Any push, PR, merge,
or default-branch integration remains a separate delivery action.

## Human Required

Default-branch delivery, pushes, destructive cleanup, and force operations
remain separately unauthorized.

## Pointers

- Project map: `docs/context/PROJECT.md`
- Manifest: `docs/plans/2026-07-26-010-kb-plan-run-worktree-isolation-manifest.md`
- DDR owner manifest: `docs/plans/2026-07-26-000-kb-ddr-route-announcement-manifest.md`
- Harness recovery: `docs/plans/2026-07-20-000-kb-harness-validation-recovery-manifest.md`
- Todo: `todo.md`

## Bootstrap Note

This initiative began in a manually prepared worktree before the new
`plan-worktree` receipt existed, so its live bootstrap lease was closed with the
lease commands rather than fabricating a workspace receipt.

## Completion Criteria

Satisfied without touching or delivering the default branch.
