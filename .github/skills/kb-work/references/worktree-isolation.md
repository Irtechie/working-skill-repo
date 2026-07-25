# Worktree Isolation Reference

Use this reference only after a slice has an atomic slice lease and needs an
isolated Git worktree.

## Prepare

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

Integration validates the owner token, receipt, and base revision. It fails
closed when the source branch moved, the worktree has uncommitted changes, or Git
reports a merge conflict. Preserve the worktree and lease for recovery when this
happens.

After integration, rerun the slice proof from the source checkout. Source proof,
not the worker receipt alone, marks the slice done.

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
