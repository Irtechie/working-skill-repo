---
name: kb-repair
description: "Surgical fix loop for QA and lint failures. Progress-based retry with stuck detection and a 5-iteration ceiling. Called by kb-qa when checks fail — not invoked directly by users."
argument-hint: "[failure report from kb-qa]"
---

# KB Repair — Surgical Fix Loop

When `kb-qa` reports failures (browser checks, lint, or both), this skill attempts targeted fixes without handing off context. The executing agent keeps its full context — no handoff, no new agent.

## When to Run

Called by `kb-qa` (Step 8: Failure Handoff) when any check fails. Never invoked directly by users or `kb-work`.

## Input

Receives from `kb-qa`:

- **Failure report** — which checks failed, expected vs observed, lint errors with file:line
- **Slice context** — `expected_files`, slice plan path, verification mode
- **Screenshots** — for any browser failures (paths to `.atv/qa-screenshots/`)
- **Previous iteration results** — if retrying (empty on first call)

## Repair Protocol

### 1. Read the Failure

Parse the failure report. For each failed check, identify:

- What specifically failed (which element, which lint rule, which line)
- The file(s) most likely responsible
- The minimal change that would fix it

### 2. Make a Surgical Fix

**Constraints — every one is mandatory:**

- Change ONLY the lines causing the failure.
- Do NOT rewrite components, refactor layouts, or restructure code.
- Do NOT add new features, improve adjacent code, or make "while I'm here" changes.
- Respect scope lock — edits MUST be within the slice's `expected_files`. If the fix requires a file outside scope, STOP and escalate to the user.
- For `op: edit` files, read the current file state first. Do not regenerate.

**What surgical means:**

| ✅ Surgical | ❌ Not surgical |
|------------|----------------|
| Fix the button color on line 47 | Rewrite the component |
| Adjust the margin from 8px to 16px | Restructure the layout |
| Fix the lint error on line 42 | Reformat the entire file |
| Add a missing `aria-label` | Refactor for accessibility |
| Fix a typo in the class name | Rename all classes to follow a convention |

### 2.5. Commit the Fix

Each fix gets its own atomic commit:

```bash
git add <changed files>
git commit -m "fix(qa): <specific issue fixed>"
```

One commit per fix. Never batch multiple fixes into one commit. Atomic commits enable atomic reverts — if a fix causes regression in Step 3, revert just that commit:

```bash
git revert --no-edit HEAD
```

Then try a different approach for the same failure. The revert counts as a stuck signal (see Step 4).

### 3. Re-verify

After each fix, re-run ALL checks — not just the one that failed:

- **Lint** (always)
- **Browser checks** (if frontend slice)

A fix for one failure might introduce another. Catch it immediately. Run the same `kb-qa` Steps 0–7 flow on the affected checks.

### 4. Assess Progress

Compare this iteration's results to the previous iteration:

| Result | Meaning | Action |
|--------|---------|--------|
| All checks pass | Fixed | Return success to `kb-qa` |
| Fewer failures | Progress | Continue to next iteration |
| Different failures | Progress (side-effect) | Continue, address new failures |
| Same failure(s) as last iteration | Stuck | Stop the loop |

"Same failure" means the identical check fails with the identical observed behavior. If the check fails but the observed value changed, that's progress (different failure), not stuck.

**Additional stuck signals** — even with "progress," these indicate flailing:

| Signal | What it means | Action |
|--------|---------------|--------|
| Fix was reverted (caused regression) | Approach is wrong, not just imprecise | Counts as stuck if it happens twice |
| Fix touched 3+ files | Not surgical — too broad for a repair | STOP. Ask the user before continuing |
| Same file reverted and re-fixed | Going in circles | Stuck — stop the loop |

### 5. Hard Ceiling

**5 iterations maximum**, even with continuous progress. Prevents infinite loops on flaky rendering, conflicting lint rules, or cascading side-effects.

If a revert was needed, that iteration still counts toward the ceiling.

### 6. On Exhaustion

If stuck or ceiling hit:

1. Log to `todo.md` under the slice:

   ```text
   repair: stuck — N iterations, M unresolved failures
     - <failure 1 description>
     - <failure 2 description>
   ```

2. Attach failure screenshots (if browser failures remain).
3. Slice stays `in_progress`, not marked `done`.
4. Return failure to `kb-qa` — the agent MUST NOT proceed to the next slice.
5. The user decides: fix manually, skip the slice, or abort.

## Principles

- Surgical means surgical. The smallest change that addresses the specific failure.
- Never add scope. Never improve. Only fix what broke.
- Re-verify everything after each fix — side-effects are real.
- If a fix needs a file outside `expected_files`, that's a scope problem, not a repair problem. Escalate.
- Progress is the signal, not iteration count. But even progress has a ceiling.

## Integration

- **Called from:** `kb-qa` (Step 8, on any failure)
- **Returns to:** `kb-qa` (pass or stuck)
- **Scope constraint:** respects `expected_files` from the slice plan
- **Context:** same agent, no handoff — repair keeps the full execution context
