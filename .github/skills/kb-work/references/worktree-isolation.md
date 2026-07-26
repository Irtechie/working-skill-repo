# Manifest-Owned Worktree Reference

The manifest group is the only worktree unit. Every slice for that workstream
runs and commits on its one plan-run branch. Never create a worktree or branch
per slice.

## Prepare the Plan Run

Before mutation, prepare or resume the manifest-owned workspace:

```powershell
go run ./cmd/kbcheck plan-worktree --action prepare --manifest <manifest-path> --run-id <run-id> --owner-token <plan-token> --base-sha <reviewed-base-sha> --json
```

The receipt records the immutable base, explicit non-default integration ref
and head, source checkout, manifest, run, owner, worktree, status, cleanup
state, and local-only limitations under the Git common directory. Repeating
prepare for the same identity is idempotent. A different owner, base, ref, or
path fails closed.

Dirty source files remain byte-for-byte in the source checkout and are excluded
from the plan-run worktree. If the plan depends on them, stop for an explicit
containment decision. Never stash, reset, clean, or copy them implicitly.

## Slice Commit Receipt

Each slice runs shared-serial in the exact receipt worktree/ref. When commits
are authorized, the worker returns a machine-readable receipt containing:

- manifest ID, run ID, slice ID, and current commit SHA;
- observed writes and newly discovered claims;
- the exact slice-proof command and expected exit;
- the aggregate-proof command and expected exit.

The worker does not change branches, create worktrees, merge, reset, stash,
clean, or edit `todo.md`, manifests, active handoffs, global installs, or the
plan-run receipt.

## Accept a Slice Commit

The coordinator accepts one commit at a time:

```powershell
go run ./cmd/kbcheck plan-worktree --action advance --manifest <manifest-path> --run-id <run-id> --slice-id <slice-id> --owner-token <plan-token> --expected-integration-head <prior-head> --commit-sha <slice-commit> --proof-receipt <receipt.json> --worktree <exact-plan-worktree> --branch <exact-plan-run-ref> --json
```

Acceptance requires:

- exact manifest, run, owner, worktree, and integration-ref lineage;
- a clean plan-run checkout;
- the recorded prior integration head as a compare-and-swap precondition;
- the slice commit as both current `HEAD` and integration-ref target;
- the slice commit to be a strict descendant of the prior head;
- coordinator replay of slice proof and aggregate proof in the plan-run
  worktree.

Only after every check passes does the coordinator atomically advance the
receipt's `integration_head`. A mismatch or failed proof leaves the receipt
unchanged and preserves the checkout for recovery. `advance` performs no Git
mutation.

## Continue and Release

The next serialized slice starts from the newly recorded integration head.
Release the slice lease only after acceptance and lifecycle projection.
Plan-run workspace release remains a separate, non-force final action and
refuses active, dirty, or unintegrated state.

## Boundaries

- Coordinates only worktrees sharing one Git common directory.
- Separate clones and machines rely on branch/PR protections.
- No per-slice worktree, per-slice branch, merge, reset, stash, force cleanup,
  remote push, PR action, or default-branch delivery occurs here.
- Ports, databases, generated outputs, and global installs remain unsafe unless
  explicitly claimed and serialized.
