---
name: kb-rehab
description: Clean house across a repository and check the result in. Pair every declared work item against real refs, settle or correct what is provably done, route the rest to a recorded disposition, and commit the reconciliation so the tree is clean before new work starts. Use when the user asks to clean house, check in, reconcile outstanding work, mark dead or superseded plans, or make the repository current. Not for ordinary code cleanup, code review, or delivering one known feature.
argument-hint: "[optional scope hint, or blank for the whole repository]"
---

# KB Rehab

Clean house, then check in. Pair every declared work item against real refs,
settle or correct what is provably done, drive the rest to a recorded
disposition, and commit the reconciliation so the next session starts from a
clean tree instead of a mess.

This lane writes no implementation code. It does commit the reconciliation
edits it makes, because a clean-house run that leaves a dirty tree has not
cleaned anything.

Two failure modes this lane must not exhibit:

- ending a run by handing back a fresh survey as if it were progress;
- leaving its own edits uncommitted for the next run to rediscover.

See `Completion` for the termination condition.

## Scope

In scope: `todo.md` rows, `docs/plans/*-manifest.md`, `docs/context/goals/`,
`docs/handoffs/active/`, every local branch and worktree in this repository's
Git common directory, and committing the reconciliation edits this lane makes.

Out of scope: writing features, reviewing code, and delivering work. Delivery
belongs to `kb-complete`. Ref and worktree reaping belongs to `kbreconcile`.
Simplifying already-committed code belongs to `kb-simplify`.

## Sequence

1. Invoke `kb-map lookup outstanding work` to resolve the project root and the
   memory files. Do not crawl the repository directly.
2. Run the read-only survey:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --json
   ```

3. Read `status`. A `fail-closed` report has withheld conclusions. Stop, report
   the limitation, and do not mark or deliver anything.
4. Read `packet`. Present the decisions the user must own, exactly as emitted.
   Do not summarize away the irreversible consequence. Grouping is by decision
   shape, never by count: two items belong in one packet item only when a single
   answer disposes of both. Never group heterogeneous artifacts to fit a ceiling.
5. Mark only what the report already proved terminal:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --action mark --json
   ```

6. Remove a declared row only when the user answered its packet item:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --action remove --json
   ```

7. Route every `orphan-work` pairing to one of four dispositions before the run
   ends. This is agent-owned triage, not a user question:

   - **settle** - the declaration's own status is terminal, so `--action mark`
     already claimed it. Nothing further.
   - **outstanding** - the declaration is `active`, `blocked`, or `draft` and its
     gate ledger agrees. Correct as-is; record it and stop touching it.
   - **stale-ledger** - the work is provably on the authoritative default branch
     but the manifest never recorded `ship-to-land`. Correct the declared status
     and cite the landing commit. Never synthesize a gate-ledger entry for a gate
     that did not run; record the landing evidence instead.
   - **user-decision** - none of the above. Only these reach the packet.

   Report the counts for all four. An `orphan-work` item left unrouted is a
   defect in this run, not a backlog item for the next one.
8. Check in the reconciliation. Commit the edits this lane made — settled
   statuses, corrected stale ledgers, removed rows — in one commit that names
   the evidence. Do not fold unrelated working-tree changes into it. See
   `Check In`.
9. For anything in `unshipped`, invoke `kb-complete` with that manifest. Never
   deliver from this lane.
10. For a ref the user authorized reaping, invoke `kbreconcile` with its own
    `apply` gate, passing `--refresh-authority` so the containment proof is
    reachable. Without that flag `kbreconcile` returns
    `authoritative-containment-unavailable` on every default invocation. Never
    delete a ref here.
11. Re-run the survey and confirm each delegation actually changed state. A
    delegation that was invoked but changed nothing is an open item, not a
    closed one. Report it as such.

`kbreconcile` proves containment only against the resolved default branch. It
cannot prove that a commit is preserved on some other ref, so a reap resting on
that argument always escalates to the user. Say so rather than routing around it.

## Check In

Committing is agent-owned durability. Pushing, opening a PR, merging, and global
skill sync are not: each needs explicit authorization every run.

Commit when this lane changed declared state. Stage only the reconciliation:
`todo.md`, manifests, goals, and handoff files this run touched. If the tree
held unrelated dirt before the run, leave it dirty and say so — never sweep a
user's in-flight edits into a rehab commit.

Verify with `git status --short` that the tree is clean of this lane's own edits
before reporting completion. A run that edited declared state and did not check
it in is incomplete.

Never commit on the resolved default branch. If HEAD is the default branch, stop
and ask which branch should carry the reconciliation.

## Completion

State the termination condition explicitly at the end of every run. Rehab is
complete when all four hold:

1. every pairing has a recorded disposition — settled or removed, routed by
   step 7 to `outstanding` or `stale-ledger`, delegated to `kb-complete` or
   `kbreconcile`, or presented to the user and still unanswered;
2. the reconciliation is checked in and `git status --short` shows none of this
   lane's edits outstanding;
3. every delegation was re-surveyed and its effect confirmed; and
4. the opening and closing pairing counts are reported side by side.

A run that closed nothing must say that plainly and name what blocked it. Do
not end a run by handing back a fresh survey as if it were progress, and do not
ask the user a question the triage in step 7 was supposed to answer.

## Delivery eligibility

This lane decides eligibility. It commits its own reconciliation and nothing
else: it never merges, pushes, or deletes without explicit authorization.

Every `unshipped` pairing resolves to exactly one of `report-only`,
`deliver-pr`, or `merge-eligible`. `deliver-pr` and `merge-eligible` both
delegate to `kb-complete`; the difference is only what `kb-complete` is told it
may do.

Default to no grant. Ask for one only when the user has already asked for
merges and the packet shows pairings that could plausibly qualify.

- A grant is bound to one run, one evidence cutoff, and an enumerated ref and
  tip. It expires within the policy `PlanTTL`.
- A grant never raises a ceiling the shipped policy sets, never authorizes a
  protected-path merge, and never authorizes global skill sync.
- Report each refusal using the receipt's reason verbatim.

Say plainly that under this repository's shipped policy `merge-eligible` is
unreachable, so granted delivery still ends at a PR a human reviews. Do not
imply a grant will produce a merge here.

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
- more than five items survive step 7 as `user-decision`; or
- the survey reports a protected path and the user has not named that path.

Escalating more than five `user-decision` items means asking the user to
prioritize them, not compressing them into one unanswerable group.

## Lazy References

- `references/classification.md` - lifecycle states, evidence rules, and the
  fail-closed triggers.
- `references/grant.md` - grant record fields, predicate inheritance, and the
  refusal vocabulary.
