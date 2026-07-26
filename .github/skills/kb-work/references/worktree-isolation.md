# Worktree Isolation Reference

Use this reference for the manifest-owned plan-run workspace. The manifest group
is the worktree unit; its slices run shared-serial on the plan-run branch. The
source/default checkout is orientation and later delivery state only.

## Plan-Run Workspace

Before any mutating slice, prepare or resume one workspace for the manifest:

```powershell
go run ./cmd/kbcheck plan-worktree --action prepare --manifest <manifest-path> --owner-token <plan-token> --base-sha <reviewed-base-sha> --json
```

The receipt records the immutable base ref/SHA, explicit non-default integration
ref and head, source checkout, manifest path, owner, worktree, status, cleanup
state, and limitations under the Git common directory. Repeating prepare for
the same manifest and owner is idempotent. A different owner, base, integration
ref, or path fails closed.

Dirty source files remain byte-for-byte in the source checkout and are excluded
from the plan-run workspace. If the plan depends on those changes, stop for an
explicit checkpoint/containment decision; never stash, reset, clean, or copy
them implicitly. Plan-run release is non-force and refuses active, dirty, or
unintegrated state.

All slices run directly in this worktree. Do not create per-slice worktrees or
branches.

## Legacy Slice Worktrees

The `kbcheck worktree` command remains readable for legacy manifests only. New
plan-run manifests do not call it. When maintaining a legacy receipt:

1. Preserve the source checkout exactly as found. Do not stash, reset, clean, or
   force-checkout user work.
2. Acquire the slice lease first.
3. Prepare a worktree with:

   ```powershell
   go run ./cmd/kbcheck worktree --action prepare --slice-id <slice-id> --run-id <run-id> --owner-token <token> --base-sha <sha> --worktree <path> --branch <branch> --json
   ```

4. Give the worker the isolated worktree path, branch, context packet, expected
   files, proof command, and escalation triggers.
5. Tell the worker not to edit canonical lifecycle files: `todo.md`, the
   manifest, active handoffs, global skill installs, or integration receipts.

## Worker Receipt

The worker returns:

- commit SHA or patch/diff;
- proof command, exit code, and relevant output/artifact path;
- observed writes and conflict-domain discoveries;
- any skipped cleanup or unresolved conflict.

The worker does not merge itself.

## Integrate

The coordinator integrates one receipt at a time:

```powershell
go run ./cmd/kbcheck worktree --action integrate --slice-id <slice-id> --run-id <run-id> --owner-token <token> --json
```

Legacy integration validates the owner token, receipt, and expected source
head. It fails closed when that head moved unexpectedly, the worktree has
uncommitted changes, or Git reports a merge conflict. Preserve the worktree and
lease for recovery when this happens.

After integration, rerun the slice proof from the plan-run worktree. Plan-run
proof, not the worker receipt alone, marks the slice done.

## Release

Release only after successful integration, source proof, manifest/board update,
and a clean isolated worktree:

```powershell
go run ./cmd/kbcheck worktree --action release --slice-id <slice-id> --run-id <run-id> --owner-token <token> --json
```

Release removes the worktree without force and then releases the matching slice
lease generation. If the worktree is dirty, unintegrated, or owned by another
token, release fails and leaves recovery state in place.

## Boundaries

- Coordinates only worktrees that share the same Git common directory.
- Does not coordinate separate clones, remote machines, tmux sessions, or PR
  delivery.
- Does not make graph indexes, browser ports, databases, generated outputs, or
  global skill sync safe unless those resources are declared and isolated or
  serialized.
