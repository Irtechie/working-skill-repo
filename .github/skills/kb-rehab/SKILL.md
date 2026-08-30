---
name: kb-rehab
description: Clean house across a repository and check everything in. Invoking this lane is the authorization to commit salvageable uncommitted work, discard provable junk, settle or correct what is provably done, drive every stale branch and worktree to a terminal state, and open, merge, and clean up the pull request that carries it, ending on the default branch with a clean tree. Use when the user asks to clean house, check in, reconcile outstanding work, mark dead or superseded plans, close out branches and worktrees, or get the repository back to a clean starting point. Not for ordinary code cleanup, code review, or delivering one known feature.
argument-hint: "[optional scope hint, or blank for the whole repository]"
---

# KB Rehab

Clean house, then check in. Pair every declared work item against real refs,
settle or correct what is provably done, drive the rest to a recorded
disposition, and commit the reconciliation so the next session starts from a
clean tree instead of a mess.

Clean house, then check in. Pair every declared work item against real refs,
settle or correct what is provably done, drive the rest to a recorded
disposition, and land the reconciliation so the next session starts from a
clean tree on the default branch instead of a mess.

This lane writes no implementation code. It does commit, push, merge, and delete,
because a clean-house run that leaves a dirty tree, an unmerged branch, or a
stale worktree has not cleaned anything.

Failure modes this lane must not exhibit:

- ending a run by handing back a fresh survey as if it were progress;
- leaving its own edits uncommitted for the next run to rediscover;
- leaving the working tree dirty because the dirt looked like someone else's;
- leaving branches or worktrees unmerged because closing them was not asked for;
- asking for permission this lane was invoked to exercise.

See `Authorization` for what the invocation already grants and `Completion` for
the termination condition.

## Authorization

**Running this lane is the permission.** A user who invoked `kb-rehab` has
already authorized, for this run and this repository:

- committing any salvageable uncommitted work;
- deleting provable junk;
- creating the branch that carries the reconciliation;
- pushing that branch;
- opening a pull request;
- merging that pull request;
- deleting merged branches, local and remote, and the worktrees holding them;
- reaping refs whose commits the default branch provably contains; and
- returning to the default branch.

Do not ask for these one at a time. Do not ask for them at all. Do not end a run
holding work back pending an approval the invocation already gave. Re-asking is
the specific defect this section exists to prevent, and it is the most common way
this lane fails its user.

This standing grant is scoped, and four things stay outside it:

1. **Unrecoverable ref deletion.** Deleting a ref holding commits the
   authoritative default branch does not contain. Escalate; never guess.
2. **History rewriting.** Force-pushing, rebasing published refs, resetting a
   shared branch, or anything that invalidates a SHA someone else may hold.
3. **Live work inside the grace window.** See `In-Flight Work`.
4. **Anything outside this repository.** Global skill sync, other remotes, and
   other checkouts still need their own explicit authorization.

Nothing else is a reason to stop and ask. In particular, "the tree was already
dirty when I started," "that branch might belong to another session," and "the
user did not say to merge" are not blocks. They are the conditions this lane
exists to resolve.

## Scope

In scope: `todo.md` rows, `docs/plans/*-manifest.md`, `docs/context/goals/`,
`docs/handoffs/active/`, every local branch, every remote-only branch, every
worktree in this repository's Git common directory, **every uncommitted change in
the working tree**, and landing the reconciliation this lane produces.

Out of scope: writing features and reviewing code. Simplifying already-committed
code belongs to `kb-simplify`. Feature delivery is delegated to `kb-complete`,
which inherits this lane's authorization rather than pausing for a new one.

## In-Flight Work

Report in-flight work first, before any survey detail. If the repository already
holds unmerged branches, extra worktrees, or a dirty tree, the opening line of
the run says how many and how old. The user should never have to discover their
own backlog by asking.

**Grace window.** A branch or worktree whose newest commit *and* newest
uncommitted file are both under **4 hours** old may belong to a session running
right now. Leave it untouched and name it as skipped-because-live. This is the
only recency exemption.

Everything older is in scope, regardless of which session created it. A branch
that has sat unmerged since yesterday is exactly what this lane closes out.
"Another session might be using it" stops being true once the work is a day old,
and treating it as true is how branches accumulate for weeks.

If a worktree older than the grace window holds uncommitted work, commit that
work on its own branch before closing the worktree. Never discard it to make the
worktree removable.

## Dirty Tree

A run that ends with a dirty tree has failed, and "the dirt was already there"
is not a defense. Classify every uncommitted path into exactly one of three:

- **salvageable** - anything tracked, plus untracked source, docs, config, or
  tests a human plausibly wrote. Commit it. Prefer this whenever classification
  is uncertain: committing is reversible and a wrong commit costs one `git
  revert`, while a wrong delete can cost the work outright.
- **junk** - gitignored paths, build and dist output, coverage and profiling
  artifacts, editor swap and backup files, crash dumps, and zero-byte files.
  Delete it and report each deletion by path.
- **untouchable** - anything matching the policy's `credential_path_patterns`.
  Never commit it and never delete it. Report it and leave it exactly as found.

There is no fourth category and no "left dirty, not mine." If salvageable work is
clearly unrelated to the reconciliation, commit it separately with its own
message saying what it appears to be, so it is durable and reviewable rather than
swept in or abandoned.

## Sequence

1. Invoke `kb-map lookup outstanding work` to resolve the project root and the
   memory files. Do not crawl the repository directly.
2. Run the read-only survey:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --json
   ```

3. Read `status`. A `fail-closed` report has withheld conclusions. Stop, report
   the limitation, and do not mark or deliver anything.
4. Read `packet`. Present only the decisions that survive `Authorization` - the
   uncontained deletions, protected paths, and genuine ambiguities the standing
   grant does not cover. Emit those verbatim without summarizing away the
   irreversible consequence. Grouping is by decision shape, never by count: two
   items belong in one packet item only when a single answer disposes of both.
   Never group heterogeneous artifacts to fit a ceiling. A packet item the grant
   already covers is not a decision; act on it instead of surfacing it.
5. Mark only what the report already proved terminal:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --action mark --json
   ```

6. Remove declared rows the survey already proved dead or superseded, and rows
   whose packet item the user answered:

   ```shell
   go run ./cmd/kbcheck work-reality --root . --action remove --json
   ```

   Dead work comes off disk so the next run does not rediscover it. A row the
   survey could not prove terminal stays put and goes to the packet.

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
8. Land the reconciliation end to end: resolve the dirty tree, branch, commit,
   gate, push, PR, merge, return to the default branch, and delete what merged.
   See `Check In`.
9. For anything in `unshipped`, invoke `kb-complete` with that manifest and this
   run's authorization, so the delegate does not stall waiting for an approval
   the user already gave. Delivery is executed by `kb-complete`, not here.
10. Reap every ref and worktree the survey proved contained:

    ```shell
    go run ./cmd/kbreconcile plan --repo . --rehab --refresh-authority --output <plan>
    go run ./cmd/kbreconcile apply --repo . --rehab --input <plan> --receipt <receipt>
    ```

    `--refresh-authority` makes the containment proof reachable; without it
    `kbreconcile` returns `authoritative-containment-unavailable` on every
    default invocation. `--rehab` raises the retire risk budgets from the
    defaults of 5 per run and 3 per repository to 100, so one run converges
    instead of leaving a backlog for the next. It raises no proof: every
    mandatory predicate is inherited unchanged. Pass `--rehab` to **both** plan
    and apply, or the policy versions differ and apply refuses the bundle.

    Reaping a contained ref needs no further approval; reaping an uncontained
    one is never allowed here.

    `kbreconcile` mutates local refs and worktrees only - `localMutationAuthorized`
    permits no other class, whatever the policy says. Delete merged **remote**
    branches directly with `git push origin --delete`, under `Authorization`,
    once containment is proven. Do not wait for `kbreconcile` to do it.
11. Re-run the survey and confirm each delegation actually changed state. A
    delegation that was invoked but changed nothing is an open item, not a
    closed one. Report it as such.

`kbreconcile` proves containment only against the resolved default branch. It
cannot prove that a commit is preserved on some other ref, so a reap resting on
that argument always escalates to the user. Say so rather than routing around it.

## Check In

Commit, push, open the PR, merge it, and return to the default branch. All of it,
every run, without asking. `Authorization` already granted this.

1. Resolve the dirty tree first, per `Dirty Tree`. Nothing below runs against a
   tree holding unclassified changes.
2. If HEAD is the default branch, create the reconciliation branch and switch to
   it. Name it and say so; never stop to ask which branch should carry the work.
   Never commit on the default branch itself.
3. Commit the reconciliation - settled statuses, corrected stale ledgers, removed
   rows - in one commit that names the evidence. Commit unrelated salvageable
   work separately rather than folding it in.
4. Run the repository's native check gate from `config/rehab-policy.json`
   (`native_check_gate`). A failing gate blocks the merge, not the commit: the
   work stays committed and durable while you report the failure.
5. Push the branch.
6. Open a pull request naming the evidence and the counts.
7. Merge it.
8. Return to the default branch, fast-forward it, and delete the merged branch
   locally and remotely along with any worktree that held it.
9. Verify and report: `git status --short` clean, HEAD on the default branch,
   level with its remote, and `git cherry <default> HEAD` empty.

A branch protection rule or a required check refusing the merge is a real block.
Report the refusal verbatim, leave the PR open, name what a human must do, and
say the run ended short of terminal. That refusal is the only acceptable reason
to finish with an open PR - not caution, and not a missing approval.

## Completion

State the termination condition explicitly at the end of every run. Rehab is
complete when all of these hold:

1. the working tree is clean, with every path committed, deleted as junk, or
   reported as untouchable;
2. HEAD is on the default branch and level with its remote;
3. every pairing has a recorded disposition - settled or removed, routed by
   step 7 to `outstanding` or `stale-ledger`, delegated to `kb-complete` or
   `kbreconcile`, or presented to the user and still unanswered;
4. every branch and worktree in scope is merged and deleted, reaped on proof,
   or named with the specific reason it survived - inside the grace window,
   uncontained, or protected;
5. the reconciliation PR is merged, or its refusal is reported verbatim;
6. every delegation was re-surveyed and its effect confirmed; and
7. the opening and closing counts are reported side by side.

These are not acceptable outcomes, and each one is a defect in the run rather
than work for the next session:

- "I did not clean up the dirty files."
- "I did not clean up the branches or worktrees."
- "I did not check anything in."
- "I did not open or merge the PR."

A run that closed nothing must say that plainly and name what blocked it. Do
not end a run by handing back a fresh survey as if it were progress, and do not
ask the user a question the triage in step 7 or the grant in `Authorization` was
supposed to answer.

## Delivery eligibility

This lane decides eligibility and acts on it. It lands its own reconciliation
directly, and it delegates feature delivery to `kb-complete` with this run's
authorization attached, so the delegate does not stop to collect an approval the
user already gave.

Every `unshipped` pairing resolves to exactly one of `report-only`,
`deliver-pr`, or `merge-eligible`. `deliver-pr` and `merge-eligible` both
delegate to `kb-complete`; the difference is only what `kb-complete` is told it
may do.

The grant is standing for the run, not something to request. Bound it and report
it rather than negotiating it:

- A grant is bound to one run, one evidence cutoff, and an enumerated ref and
  tip. It expires within the policy `PlanTTL`.
- A grant never raises a ceiling the shipped policy sets, never authorizes a
  protected-path merge, and never authorizes global skill sync.
- Report each refusal using the receipt's reason verbatim.

`report-only` is reserved for pairings the policy itself refuses, not for
pairings you would rather a human confirmed. A protected path under
`protected_paths` still escalates, and so does an uncontained ref. Those
refusals are real; caution is not one.

Keep one distinction honest. This lane merges **its own reconciliation PR**
directly under `Authorization`, using `gh`. Delivering someone else's
`unshipped` feature branch is a different path, and two facts bound it:
`internal/reconcile/policy.go` registers `ActionMerge` with `allowed=false` and
a per-run budget of `0`, and `plan.go`'s `localMutationAuthorized` permits only
`ActionWorktreeRetire` and `ActionLocalRefRetire` to mutate at all. `--rehab`
changes neither. `merge-eligible` is therefore unreachable here, and delegated
feature delivery still ends at a PR a human reviews. Say that plainly rather
than implying the standing grant auto-merges feature work.

## Preservation

These limits bound the standing grant; they do not reintroduce a general "ask
first" default. They apply to what reaches the packet, not to what
`Authorization` already covers.

An unanswered packet item leaves that work item and that ref untouched. Silence
is preservation, never consent - but only an item that legitimately reached the
packet qualifies. Do not manufacture a packet item to avoid acting.

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

Escalation is for these four cases only. It is not a way to hand back work the
grant covers, and a run that escalates instead of acting on granted work has
failed the same way as a run that closed nothing.

## Lazy References

- `references/classification.md` - lifecycle states, evidence rules, and the
  fail-closed triggers.
- `references/grant.md` - grant record fields, predicate inheritance, and the
  refusal vocabulary.
