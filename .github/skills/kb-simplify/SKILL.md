---
name: kb-simplify
description: Explicitly invoked maintenance pass over already-committed code. Ranks at most six simplification targets by change frequency and code health, then executes one at a time in a confirm loop. Use when the user asks to simplify, deslop, de-duplicate, remove dead surface, or pay down accumulated maintenance debt. Never automatic and never part of a build, work, or completion loop.
argument-hint: "[path, module, or subsystem to sweep]"
---

# KB Simplify

Find complexity that accumulated because every individual change looked fine on
its own, then remove it one reviewable change at a time.

This is a maintenance lane. It is invoked by the user and never fires on its own.
No KB skill may call it. `kb-work`, `kb-complete`, `kb-qa`, and `kb-finalize`
must not trigger it.

## Use When

- The user says simplify, deslop, clean up, de-duplicate, or pay down debt.
- A subsystem has been edited many times and nobody has stepped back.
- Repeated patterns or near-copies are suspected across files.
- An existing abstraction has grown flags and conditionals.

## Do Not Use For

- A specific diff. That is `kb-review`, whose profiles already cover code health
  and avoidable complexity.
- Whether the architecture should change. That is `kb-architecture-deepening`.
- A failing test or reported bug. That is `kb-fix` or `kb-troubleshoot`.
- Doc, memory, or response compaction. That is `kb-cognitive`.

## Step 1 — Rank By Churn First, Not Ugliness

Stable code is not a target no matter how ugly. Rank by recent change frequency,
then judge health only within that set.

```bash
git log --since="90 days ago" --name-only --format="" -- <path> \
  | grep . | sort | uniq -c | sort -rn | head -40
```

```powershell
git log --since="90 days ago" --name-only --format="" -- <path> |
  Where-Object { $_ } | Group-Object | Sort-Object Count -Descending |
  Select-Object -First 40 Count, Name
```

Read the top files. Skip anything with zero commits in the window and say so.
If the user names a path, honor it and still rank within it.

## Step 2 — Classify Each Candidate

Use exactly these categories:

| Category | Signal |
|---|---|
| `duplication` | Same knowledge expressed in 3+ places |
| `wrong-abstraction` | A shared helper carrying behavior-selecting flags |
| `over-engineering` | Indirection, config, or generality with one caller |
| `dead-surface` | Unreferenced exports, agents, flags, branches, or files |
| `inconsistency` | The same job done differently in nearby code |

## Step 3 — Apply The Abstraction Test Before Proposing Extraction

Ask all three. A `no` on any one means do not extract.

1. **Same knowledge?** DRY governs knowledge, not text. Two blocks that look
   alike but serve different concerns are not duplication.
2. **Third instance?** Two occurrences is a watchlist note, not an action.
3. **Nameable without a behavior flag?** If naming it requires "sometimes X,
   sometimes Y," the abstraction is wrong before it exists. Prefer duplication.

Apply the inverse with equal weight. An existing helper whose callers pass
behavior-selecting booleans or mode strings is a candidate for **Inline
Function**, not for more parameters. Re-inlining into callers and deleting the
branches each caller does not need is a legitimate, often better, outcome.

Reuse the deletion test from `kb-architecture-deepening`: if a change does not
let callers delete coordination, branches, mocks, or setup, it is not a
simplification.

## Step 4 — Present At Most Six

Never present more than six. Fewer is better. If nothing qualifies, say so and
stop rather than manufacturing targets.

```markdown
## Simplification Targets

| # | Target | Category | Refactoring | Risk | What it deletes |
|---|---|---|---|---|---|
| 1 | `path:line` | wrong-abstraction | Inline Function | HIGH | 3 flags, 2 branches |

### 1. <name>
- Evidence: <files, occurrence count, commits in window>
- Proposed change: <named catalog refactoring>
- What this deletes:
- Risk: HIGH (no covering tests) | MEDIUM | LOW
- Leave-alone case: <the honest argument against doing this>

### Considered And Rejected
- <candidate>: <which test it failed>
```

Name the refactoring from Fowler's catalog: Extract Function, Inline Function,
Extract Class, Pull Up Method, Replace Conditional with Polymorphism,
Consolidate Conditional Expression, Remove Dead Code.

Risk is `HIGH` whenever the area has no covering test, regardless of how small
the change looks.

## Step 5 — Execute Exactly One

Wait for the user to choose. Do not batch, and do not start work while
presenting.

1. Confirm tests covering the target pass **before** touching anything. If none
   exist, write characterization tests first that record actual current
   behavior, not desired behavior, and get user agreement before proceeding.
2. Make the structural change only. No behavior changes, no renames of unrelated
   things, no drive-by fixes.
3. Run the same tests. They must pass unchanged.
4. Commit the structural change on its own, separate from any behavioral commit.
5. Report what was deleted, not what was added.

If a second problem is spotted mid-change, write it down and finish the current
one. Do not follow it.

## Step 6 — Re-rank And Ask Again

Re-rank after each completed target, because finishing one often dissolves
others. Present the updated list and ask whether to continue.

Loop until the user stops. Stop immediately and without argument when they do.
Never ask more than once per completed target.

## Do Not

- Run without explicit user invocation.
- Present more than six targets.
- Apply multiple targets in one diff.
- Extract on the second occurrence.
- Refactor code with zero recent commits.
- Rewrite working code into a different style and call it simplification.
- Mix behavioral change into a structural commit.
- Manufacture targets to fill the list.

## Background

Ranking, abstraction, and execution rules are evidence-backed in
`docs/context/research/2026-08-02-simplification-maintenance-pass.md`. Read it
before changing this skill's heuristics.
