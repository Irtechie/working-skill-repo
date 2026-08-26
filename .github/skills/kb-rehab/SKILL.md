---
name: kb-rehab
description: Reconcile all outstanding work in a repository against git reality before new work starts. Use when the user asks what work is still outstanding, wants dead or superseded plans marked, wants branches paired against declared work, or wants the repository cleaned and current so the next task can begin. Not for ordinary cleanup, code review, or delivering one known feature.
argument-hint: "[optional scope hint, or blank for the whole repository]"
---

# KB Rehab

Pair every declared work item against real refs, mark what is provably dead or
superseded, and hand everything else to a human as a bounded decision. This lane
classifies and reports. It writes no implementation code.

## Scope

In scope: `todo.md` rows, `docs/plans/*-manifest.md`, `docs/context/goals/`,
`docs/handoffs/active/`, and every local branch and worktree in this
repository's Git common directory.

Out of scope: writing features, reviewing code, and delivering work. Delivery
belongs to `kb-complete`. Ref and worktree reaping belongs to `kbreconcile`.

## Sequence

1. Invoke `kb-map lookup outstanding work` to resolve the project root and the
   memory files. Do not crawl the repository directly.
2. Run the read-only survey:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --json
   ```

3. Read `status`. A `fail-closed` report has withheld conclusions. Stop, report
   the limitation, and do not mark or deliver anything.
4. Read `packet`. Present at most five grouped decisions to the user exactly as
   emitted. Do not summarize away the irreversible consequence.
5. Mark only what the report already proved terminal:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --action mark --json
   ```

6. Remove a declared row only when the user answered its packet item:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --action remove --json
   ```

7. For anything in `unshipped`, invoke `kb-complete` with that manifest. Never
   deliver from this lane.
8. For a ref the user authorized reaping, invoke `kbreconcile` with its own
   `apply` gate. Never delete a ref here.

## Preservation

An unanswered packet item leaves the work item and the ref untouched. Silence is
preservation, never consent.

A `human-required`, `orphan-work`, `orphan-branch`, `unshipped`, or `live`
pairing is never marked terminal and never removed.

The report refuses every write when it is `fail-closed`, because a report that
could not prove its conclusions cannot authorize a change.

## Escalation

Return to the user, not to action, when:

- a pairing needs a decision vocabulary `todo-triage` does not define;
- a removal would drop a ref holding commits the authoritative default branch
  does not contain;
- the packet would exceed five grouped items after grouping; or
- the survey reports a protected path and the user has not named that path.

## Lazy References

- `references/classification.md` - lifecycle states, evidence rules, and the
  fail-closed triggers.
