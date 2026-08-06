# Manifest-Owned Worktree Reference

## Contents

- Choose Adopt or Prepare
- Concurrency Ceiling
- Naming
- Adopt the Harness Workspace
- Prepare a KB-Owned Plan Run
- Slice Commit Receipt
- Accept a Slice Commit
- Continue and Release
- Boundaries

The manifest group is the only worktree unit. Every slice for that workstream
runs and commits on its one plan-run branch. Never create a worktree or branch
per slice.

## Choose Adopt or Prepare

Resolve the workspace before mutation. A coding harness that gives each session
its own linked worktree already provides the isolation unit; creating a second
one nests two worktrees on the same logical thread and leaves strays behind.

| Current checkout | Action | Worktree lifecycle owner |
|---|---|---|
| Harness-provided linked worktree on a non-default branch | `adopt` | harness |
| Shared or primary checkout | `prepare` | KB |

`adopt` creates no worktree and no branch. It fails closed on the primary
checkout, a detached HEAD, a resolved default branch, a dirty tree, or a
requested worktree/branch/base that does not match the current one.

## Concurrency Ceiling

At most two KB-owned plan-run worktrees may be live in one repository. That is
two concurrent manifest groups, not two slices; slices stay serial inside their
group. `prepare` fails closed at the ceiling and names the live runs. The third
workstream gets no tree and no branch — it queues until a slot frees, or it
belongs on an existing plan-run branch as more slices.

Never resolve a ceiling block by checking out a third branch in an existing
tree. That preempts the run in flight instead of parallelizing it.

Harness-created session worktrees are not counted and never removed by KB.
Gating on them would fail closed on state KB cannot remediate.

`adopt` creates nothing and is never capped.

Raise or lower the ceiling only with evidence, in
`docs/context/operations/kb-routing.yaml`:

```yaml
execution:
  max_plan_run_worktrees: 2
```

The claim DAG, not the tree count, is the concurrency control. Extra trees buy
merge cost and disk rather than throughput.

## Naming

A `prepare`d worktree needs a name. An adopted worktree keeps the name the
harness already gave it.

Choose a short, task-specific worktree codename with irreverent, self-aware,
absurdly specific humor that remains safe on a shared screen, terminal, PR, or
audit log. Aim for something a wisecracking antihero would name, not the
antihero's name or fandom references. Prefer names such as
`the-reviewers-have-unionized`, `this-prompt-needs-an-adult`, or
`somehow-another-abstraction` over opaque generated pairs such as
`improved-funicular`. Keep the manifest ID in receipts; the human-facing name
should optimize recognition rather than duplicate it.

Branch and worktree basename must share that exact codename. A namespace prefix
is allowed for the branch, but the funny task name must remain intact:

| Task | Worktree basename | Branch |
|---|---|---|
| Reduce reviewer overuse | `the-reviewers-have-unionized` | `codex/the-reviewers-have-unionized` |
| Simplify an overgrown prompt | `this-prompt-needs-an-adult` | `codex/this-prompt-needs-an-adult` |
| Remove needless abstractions | `somehow-another-abstraction` | `codex/somehow-another-abstraction` |

Do not use a funny name merely because it is funny. It must relate recognizably
to the work so branch lists and worktree paths remain useful routing evidence.

## Adopt the Harness Workspace

```powershell
go run ./cmd/kbcheck plan-worktree --action adopt --manifest <manifest-path> --run-id <run-id> --owner-token <plan-token> --commit-authorized --commit-authorized-by <actor> --commit-approval-ref <reference> --root <current-worktree> --json
```

The receipt records the current worktree, its existing branch as the integration
ref, and its current head as the immutable base. Repeating adopt for the same
identity is idempotent. `prepare` cannot later take over an adopted receipt, and
adopt cannot take over a KB-owned one.

## Prepare a KB-Owned Plan Run

Before mutation from a shared or primary checkout:

```powershell
go run ./cmd/kbcheck plan-worktree --action prepare --manifest <manifest-path> --run-id <run-id> --owner-token <plan-token> --base-sha <reviewed-base-sha> --worktree <parent>\<repo>-<codename> --branch codex/<codename> --json
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

Release removes a KB-owned worktree. Releasing an adopted receipt returns
ownership to the harness and records `cleanup_state: harness-owned` without
removing the worktree or its branch; the harness that created the session owns
that teardown.

## Boundaries

- Coordinates only worktrees sharing one Git common directory.
- Separate clones and machines rely on branch/PR protections.
- No per-slice worktree, per-slice branch, merge, reset, stash, force cleanup,
  remote push, PR action, or default-branch delivery occurs here.
- KB never creates or deletes a harness-owned worktree.
- Ports, databases, generated outputs, and global installs remain unsafe unless
  explicitly claimed and serialized.

## Tooling Availability

`cmd/kbcheck` belongs to this bundle's source repo and does not ship with an
installed skill. Treat every `go run ./cmd/kbcheck ...` command above as
conditional: run it when the repo provides it, otherwise perform the equivalent
git worktree step directly and record what you ran. A missing harness changes
which command you run, never whether the worktree is verified.
