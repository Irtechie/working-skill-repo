---
name: kb-complete
description: "Post-work quality and learning pipeline. Runs ce-review -> resolution gate -> compound -> learn -> evolve -> cleanup after kb-work finishes all slices. Use when the user says 'kb complete', 'complete the work', 'run review and learning', 'finish the KB pipeline', or after kb-work reports all slices done."
argument-hint: "[path to KB manifest, or blank to find latest]"
---

# KB Complete - Post-Work Quality & Learning Pipeline

After `kb-work` finishes executing all slices, this skill runs the quality review, knowledge capture, and cleanup steps. Separated from kb-work so the user can choose when to run it — `klfg` prompts automatically, standalone users invoke it manually.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Scan `docs/plans/` for the most recent `*-kb-*-manifest.md` file with `status: completed`. If none is found, ask: "Which KB manifest should I complete?"

**If input is a path:** Read the manifest at that path.

## Pre-Flight

1. **Read the manifest** — confirm `status: completed` (all slices done/skipped). If slices are still `pending` or `in_progress`, stop: "This manifest has unfinished slices. Run `kb-work` first."
2. **Collect scope context** — scan each slice's `notes` field for `scope-check:` entries. Build the combined list of scope-verified files across all slices. This becomes the review scope.
3. **Collect memory impact** — scan slice notes for `memory-impact:` and `kb-map-refresh:` entries.
4. **Identify the branch baseline** — run `git merge-base HEAD main` to establish the diff range.

If the manifest has no scope-check notes (older format), fall back to `git diff --name-only $(git merge-base HEAD main)..HEAD` for the file list.

## Step 1: Code Review

Before code review, run `kb-check` against the completed manifest scope. If deterministic checks fail, route to `kb-repair` or `kb-fix` before `ce-review`. LLM review does not replace executable verification.

If the manifest contains user-visible, API/CLI, persistence, auth, streaming, or integration changes, run `kb-functional-test` before `ce-review` to confirm the functional coverage is real and not mock-only.

**Invoke `ce-review`** — full multi-agent code review on the feature diff.

This is mandatory. Do not skip, defer, or make it optional.

- **Pass scope from prior gates:** use the collected scope-verified file list from Pre-Flight. Pass this as the scoped file list so ce-review skips its own scope discovery (Stage 1). The scope gates already verified these are the correct files — no need to re-derive from git diff.
- Pass context: the full `git diff` of the feature branch against baseline, scoped to the verified file list
- Capture the output: each finding has a severity (P0/P1/P2/P3) and confidence score
- Store findings for the resolution gate (Step 2)
- **Note:** if scope-verified files are unavailable (older manifest, standalone run), let ce-review do its own scope discovery.

## Step 2: Resolution Gate

Review findings from `ce-review` determine whether completion is allowed:

| Severity | Action |
|----------|--------|
| P0 (critical) | STOP. Fix before proceeding. Re-run affected tests after fix. |
| P1 (important) | STOP. Fix before proceeding. |
| P2 (suggestion) | Log in manifest `notes`. Do not block. |
| P3 (nit) | Log in manifest `notes`. Do not block. |

This gate is mandatory. The agent MUST NOT proceed to Step 3 while unresolved P0/P1 findings exist.

For any P2/P3 findings, invoke `kb-gate` with the rectify prompt. Do not silently leave fixable P2/P3 issues when the user would prefer a clean finish.

After resolving all P0/P1s, update the manifest notes with a summary:
`review: P0=0 P1=2(resolved) P2=3(logged) P3=1(logged)`

**Feed learnings to the observation log:**

For each resolved P0/P1 finding, append one line to `.atv/observations.jsonl`:

```json
{"ts":"<ISO-8601>","hook":"ce-review","tool":"kb-complete","args":{"finding_type":"<category>","severity":"P0|P1","resolution":"<what was fixed>"},"cwd":"<repo-root>","result":"resolved"}
```

This connects the review → learn pipeline. Only P0/P1 findings are worth learning from — P2/P3 are style preferences, not systemic patterns.

Create `.atv/observations.jsonl` if it doesn't exist. Append, never overwrite.

## Step 3: Compound & Learn

After the resolution gate passes, document what this feature taught the system:

1. **Invoke `ce-compound`** with context: a one-sentence summary of what was built and any surprising patterns discovered during implementation.
2. ce-compound writes to `docs/solutions/` with YAML frontmatter — let it run without modification.
3. If the implementation was pure boilerplate (no novel patterns, no gotchas, no decisions worth preserving), skip with a manifest note: `compound: skipped — standard implementation, no novel patterns`
4. Per-slice micro-learnings from slice notes feed into the compound context. Reference them when invoking ce-compound.
5. **Invoke `/learn`** — Extract instincts from this session's work.
   - Run after compound completes (observations from Step 2 are now available)
   - `/learn` reads: observations.jsonl, recent git history, docs/solutions/, existing instincts
   - Record result in manifest notes: `learn: N new instincts, M updated` or `learn: no new patterns`
   - This is automatic — do not ask the user whether to run it
6. **Check evolution cadence:**
   - Read `.atv/kb-completions.txt` (create with `0` if missing)
   - Increment by 1
   - Write the new value back
   - If the new value is divisible by 5:
     - Invoke `/evolve` to check for promotable instincts
     - Log result in manifest notes: `evolve: promoted N instincts` or `evolve: no candidates ready`
   - If not divisible by 5: skip silently
   - Commit the counter file with the manifest update

## Step 3.5: Project Memory Refresh Gate

Before cleanup or final "complete", make sure a fresh session can resume without a lesson from the user.

Run `kb-map refresh` when any of these are true:

- The manifest contains `memory-impact: durable`.
- `kb-work` left any `refresh=pending` note.
- Review fixes changed behavior, architecture, run/test commands, integrations, or known sharp edges.
- `docs/context/PROJECT.md` points to stale subsystem docs after the feature diff.

Skip with a manifest note only when changes are clearly cosmetic, copy-only, formatting-only, lint-only, or isolated tests with no durable behavior change:

```text
kb-map-refresh: skipped - cosmetic/no durable architecture change
```

When refresh runs, update affected docs only:

- `docs/context/PROJECT.md` for route-map, command, or subsystem index changes.
- `docs/context/architecture/*` for durable subsystem behavior.
- `docs/context/operations/*` for run/test/deploy/QA changes.
- `docs/context/research/*` for reusable research outcomes.
- `docs/context/decisions/*` for accepted/rejected approaches.
- `todo.md` and handoff files for current operational state.

Then add:

```text
kb-map-refresh: done - <docs updated>
```

## Step 4: Cleanup

Prune ephemeral artifacts. Heavy KB usage generates file sprawl — clean it up per-feature, not manually.

1. **QA screenshots** — delete `.atv/qa-screenshots/` contents for this feature's slices. Screenshots should already be referenced in commits or PR bodies. Safe to remove.

2. **Observations log** — trim `.atv/observations.jsonl` entries older than 90 days. Matches the recency decay half-life in `/learn`. Append-only logs grow indefinitely without this.

   ```bash
   # Keep entries from the last 90 days
   python -c "
   import json, sys
   from datetime import datetime, timedelta
   cutoff = (datetime.utcnow() - timedelta(days=90)).isoformat()
   lines = open('.atv/observations.jsonl').readlines()
   kept = [l for l in lines if json.loads(l).get('ts','') >= cutoff]
   open('.atv/observations.jsonl','w').writelines(kept)
   print(f'observations: kept {len(kept)}/{len(lines)}')
   "
   ```

   If Python is unavailable or the file doesn't exist, skip with a note.

3. **Plan files** — leave manifests and slice plans in `docs/plans/`. Lightweight reference material, useful for tracing decisions.

4. **Log cleanup** in the manifest notes:

   ```text
   cleanup: screenshots pruned, observations trimmed to 90d
   ```

5. **Todo hygiene** — verify `todo.md` contains only active, blocked, parked, manual, or handoff-pointer work. If the completed feature or routine slice completion logs remain there, move a compact summary to `todo-done.md` and remove those entries from `todo.md`. `todo.md` must keep its `## Rules` section at the top; do not depend on a separate `todo-rules.md`.

## Step 5: Done

Update the manifest `status: reviewed` and report:

```text
KB <name> complete.
- Review: P0=N P1=N(resolved) P2=N P3=N
- Compound: <written | skipped>
- Learn: <N new, M updated | no new patterns>
- Evolve: <promoted N | skipped | no candidates>
- Project memory: <refreshed | skipped with reason>
- Cleanup: done

Ready to ship. Run /land when you're ready to push and open a PR.
```

## Failure Handling

| Situation | Action |
|-----------|--------|
| ce-review fails to run | Log error, ask user whether to retry or skip review |
| P0/P1 fix breaks tests | Re-run tests, treat as new failure, fix before proceeding |
| compound/learn/evolve fails | Log error, continue — these are non-blocking |
| Manifest not found | Ask user for path |
| Manifest has unfinished slices | Stop, tell user to run kb-work first |

## Integration

- **Input from:** `kb-work` (completed manifest)
- **Review engine:** `ce-review` with scope passthrough
- **Documentation:** `ce-compound` → `docs/solutions/`
- **Project memory:** `kb-map refresh` → `docs/context/*`, `todo.md`, handoffs
- **Learning:** `/learn` → `.atv/instincts/project.yaml`
- **Evolution:** `/evolve` → `.github/skills/learned-*/`
- **Shipping:** `/land` (separate, deliberate act — not part of this skill)
