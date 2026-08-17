---
name: w2d
description: "Work to done. Post-plan command that executes a validated manifest, finalizes it, opens a PR, and merges that PR when permissions and required checks allow. Use after kb-plan, or when the user says w2d, 'work to done', 'take it all the way', or 'land it'."
argument-hint: "[manifest path, or blank to resume the active manifest]"
---

# W2D

Carry a planned manifest to an accepted PR in one command.

```text
validated manifest
  -> kb-work
  -> kb-finalize
  -> kb-ship        (push topic branch, open PR)
  -> kb-land        (merge when permitted)
  -> report exactly one delivery state
```

`w2d` is the post-plan endpoint. `kb-plan` chains here once `plan-to-work`
passes. It orchestrates existing phases and never reimplements them.

## Authorization

Explicit `w2d` invocation is the run-scoped merge authorization. The user asked
for done, not for an open PR.

That authorization is intent, never capability. After `kb-ship` opens the PR,
merge only when all of these hold:

- the authenticated account can merge this PR on this base branch;
- branch protection, required checks, and required approvals are satisfied;
- the PR base is the resolved remote default and the head is the exact
  `kb-ship` commit.

If any fails, stop at the open PR and report `awaiting-review` with the exact
missing condition. A blocked merge is a successful `w2d` run with an honest
endpoint, not a failure to retry around.

Never use force push, admin bypass, hook bypass, or auto-merge that would
bypass an unsatisfied protection rule.

## Preconditions

1. Run `kb-map lookup <request>` and resolve the active project root.
2. Claim or resume the objective in the shared `kb-start` work queue before any
   mutating phase. Stop successor creation if another live session owns it.
3. Require a validated manifest with `plan-to-work: passed`. With no manifest,
   route to `kb-plan` first; `w2d` never executes unplanned work.
4. If input is blank, resume the single active manifest from `todo.md`. Ask only
   when multiple active manifests are genuinely plausible.

## Phase Loop

Re-read the manifest after every delegated phase. Durable state, not chat
memory, chooses the next action.

| Current state | Action |
|---|---|
| runnable slices remain | `kb-work <manifest>` |
| `work-to-complete: passed` | `kb-finalize <manifest>` |
| `complete-to-ship: passed\|quarantined` | `kb-ship <manifest>` |
| open PR, merge conditions met | `kb-land <manifest>` |
| open PR, merge conditions unmet | stop at `awaiting-review` |
| paused | stop without changing technical gate status |
| blocked on critical path | recheck the recorded sensor, then stop only affected work |

Heartbeat the work claim after every delegated phase. Publish `done`,
`blocked`, or `superseded` before terminal output.

Treat a passing exact-tree proof receipt as durable phase evidence. Do not
rerun proof merely because orchestration advanced a phase. Rerun only checks
invalidated by changed files, dependencies, test configuration, environment
identity, or delivery tree.

## Scope Rules

- Delegated phases keep their own gates and authority. `w2d` selects the phase
  owner; it never grants a phase a permission that phase does not hold.
- `kb-land` remains the only skill that integrates the remote default branch.
- Do not stage, commit, revert, or overwrite unrelated dirty work.
- The reviewed manifest-owned plan-run branch is the only delivery candidate.
  Never reconstruct delivery from a dirty source or default checkout.
- Do not roll a narrow gate up to the whole product. A release-only failure
  blocks merge while local implementation stays complete.

## Relationship to kb-complete

`kb-complete` is policy-driven: it reads stored `delivery` policy and stops at
an open PR unless that policy or a separate authorization permits merge.

`w2d` is intent-driven: invoking it supplies that authorization for this run.

Use `kb-complete` when stored project policy should decide the endpoint. Use
`w2d` when the user wants planned work accepted now.

`w2d` still honors a stored `delivery.mode: local`. That is an explicit opt-out
from publishing, so report the reviewed manifest and stop rather than pushing.
`delivery.mode: direct` stays owned by `kb-complete` and `kb-land`; `w2d`
always delivers through a PR.

## Stop Rules

- Do not merge without a satisfied protection, check, and approval state.
- Do not merge a PR whose head is not the exact audited `kb-ship` commit.
- Do not execute a manifest that has not passed `plan-to-work`.
- Do not retry a merge that failed an authorization or protection condition;
  report it and stop.
- Do not run post-merge sync or propagation unless policy explicitly enables it.
- Do not delete the current session worktree.

## Output

Report exactly one delivery state and the evidence behind it:

```text
W2D: <manifest>
State: local-durable | awaiting-review | delivery-integrated
PR: <url or none>
Merge: merged <sha> | blocked <exact missing condition> | not-applicable
Proof: <phase receipts>
```
