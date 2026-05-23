---
name: kb-work
description: "Sequential executor for kb-plan output. Runs all vertical slices in dependency order with fresh context per task, TDD enforcement, and HITL pauses. Use when the user says 'kb work', 'work the plan', 'execute the plan', 'run the KB pipeline', 'execute all slices', or wants guided execution of a planned feature."
argument-hint: "[path to KB manifest, or blank to find latest]"
---

# KB Work - Sequential Slice Executor

Run all vertical slices from a `kb-plan` manifest in dependency order. Keep each slice scoped, enforce the requested verification mode, and pause on HITL tasks.

## Quick Start

1. Read the KB manifest.
2. Validate the dependency DAG and statuses.
3. Confirm execution with the user unless they requested non-interactive execution.
4. Execute ready slices in topological order.
5. Update the manifest after each slice so the workflow is resumable.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Scan `docs/plans/` for the most recent `*-kanban-*-manifest.md` file. If found, use it. Otherwise ask: "Which KB manifest should I execute?"

**If input is a path:** Read the manifest at that path.

## Pre-Flight

1. **Read the manifest** - parse the YAML frontmatter to get the ordered slice list.
2. **Validate DAG** - confirm no cycles in blockers, all referenced slice IDs exist, and all slice files exist.
3. **Check status** - skip any slices already marked `done`. Resume from the first runnable `pending` slice.
4. **Check worktree** - note dirty or untracked files before executing so unrelated user changes are not staged or reverted.
5. **Sync with board** - read `docs/kanban.md` and confirm its status table matches the manifest. If they diverge, the board wins — another agent may have updated it. Reconcile the manifest from the board before proceeding.
6. **Confirm with user:** "Ready to execute N remaining slices in order. Proceed?"

Treat statuses as:

| Status | Action |
|--------|--------|
| `pending` | Eligible once blockers are `done` or `skipped` |
| `done` | Skip |
| `blocked` | Stop and ask whether to retry, skip, or abort |
| `skipped` | Skip but keep visible in the summary |

## Board Sync Protocol

`docs/kanban.md` is the multi-agent handoff file. Update it at every status transition:

| Event | Board Update |
|-------|-------------|
| Starting a slice | Set status to 🔧 in_progress |
| Slice completes | Set status to ✅ done |
| Slice blocked | Set status to 🔒 blocked + reason in notes |
| Slice skipped | Set status to ⊘ skipped |
| All slices done | Move entire feature section to `docs/kanban-done.md` |

**Multi-agent rules:**
- Before claiming a slice, re-read `docs/kanban.md`. If another agent set it to 🔧, do not claim it.
- The board is the source of truth — not chat history, not the manifest. If the board says done, it's done.
- Update the board BEFORE starting work (claim) and AFTER completing work (release). This prevents two agents from working the same slice.
- Also update the manifest to stay in sync, but if they conflict, the board wins.
- Append a short entry to the **📝 Work Log** section after each slice: `- YYYY-MM-DD: <slice title> — done (<agent/session>)`

## Dependency Ordering

Execute with a topological sort:

1. Build a map of `slice_id -> slice`.
2. For each pending slice, check all `blockers`.
3. Run the first pending slice whose blockers are complete.
4. If pending slices remain but none are runnable, mark the manifest blocked and report the dependency problem.

## Execution Loop

For each slice in dependency order:

### Step 1: Check HITL Flag

If `hitl: true`:

- Present the slice title, description, and the specific question/decision needed.
- Stop and wait for user input.
- Record the user's decision in the slice plan.
- Update manifest status to `done` for this slice only if the decision completes the slice.
- Continue to the next runnable slice.

### Step 2: Deepen If Thin

If the slice plan has fewer than 3 acceptance criteria or no test scenarios:

- Run a lightweight deepening pass on this single slice.
- Add concrete test scenarios and likely file paths.
- Keep the pass bounded; do not re-plan the whole feature.

### Step 3.0: Scope Lock

Before executing the slice, load the declared scope and enforce it proactively — prevent out-of-scope edits before they happen.

1. **Read `expected_files`** from the slice plan's frontmatter.
2. **If `expected_files` is empty or missing**, the gate fails. Stop and require the field to be populated before execution begins.
3. **Expand scope with convention-matched test files.** For each entry in `expected_files`, automatically allow its corresponding test file based on project naming conventions:

   | Source file | Auto-allowed test file(s) |
   |-------------|--------------------------|
   | `src/foo.py` | `tests/test_foo.py`, `test/test_foo.py` |
   | `src/Foo.tsx` | `src/Foo.test.tsx`, `src/__tests__/Foo.tsx` |
   | `lib/foo.rb` | `spec/foo_spec.rb`, `test/foo_test.rb` |
   | `pkg/foo.go` | `pkg/foo_test.go` |

   Test files that don't correspond to any `expected_files` entry are still out of scope.

4. **Before opening any file for writing**, check its path against `expected_files` + auto-allowed test files:

   | Finding | Action |
   |---------|--------|
   | File is listed in `expected_files` or is a convention-matched test | Proceed with the edit. |
   | File is NOT listed and not a matching test | **STOP.** Do not edit. Ask the user: "This file isn't in the slice scope. Add it to `expected_files`, or skip this edit?" |

5. **Log the lock** in the manifest notes: `scope-lock: loaded N expected files + M auto-allowed test files`

This gate pairs with Step 3.6 (Diff-Scope Verification). Scope Lock prevents out-of-scope edits before they happen. Diff-Scope Verification catches anything that slipped through after the fact. Both are mandatory. Neither can be skipped, overridden, or deferred.

### Step 3: Execute

Use a fresh sub-agent when the platform supports delegated execution and the user has permitted it. Otherwise execute the slice locally while keeping the scope limited to this slice.

The execution prompt or local checklist must include:

```text
You are executing a single vertical slice. Complete it fully.

Kanban: <kanban_id>
Slice: <slice_id> - <title>
Verification mode: <tdd|integration|verification-only>

Plan contents:
<full slice plan content>

Instructions:
1. Read the plan completely.
2. Set up on the current branch.
3. For files marked `op: edit` in expected_files:
   - Read the CURRENT file content FIRST — the file as it exists on disk, not as the plan describes it.
   - Make only the change described in the `scope` field. Do not regenerate or rewrite the file.
   - Preserve all existing behavior not mentioned in the scope.
   - If the file has been modified by prior slices or manual edits, those changes are authoritative. The plan spec is not.
4. For files marked `op: create`: write the file from the plan spec.
5. Apply the verification mode:
   - tdd: write one failing test first, confirm it fails for the right reason, implement minimal code to pass, then refactor.
   - integration: write an integration test proving the path works, then implement the wiring.
   - verification-only: implement the change, then verify builds pass and no regressions.
6. Run relevant tests first, then the full test suite when practical.
7. Stage only files changed for this slice.
8. Commit only if the user asked for commits, with message: "feat(<slice-id>): <title>"

Do not modify other slices' files unless required for this slice.
Do not add scope beyond what the plan specifies.
Do not stage unrelated dirty or untracked files.
```

### Step 3.5: System-Wide Test Check

Before marking a slice done, pause and ask these questions — vertical slices cut through all layers, so side-effects matter:

| Question | What to do |
|----------|------------|
| **What fires when this runs?** Callbacks, middleware, observers, event handlers — trace two levels out from your change. | Read the actual code for callbacks on models you touch, middleware in the request chain, `after_*` hooks. |
| **Do my tests exercise the real chain?** If every dependency is mocked, the test proves logic in isolation — not interaction. | Write at least one integration test that uses real objects through the full callback/middleware chain. |
| **Can failure leave orphaned state?** If your code persists state before calling an external service, what happens when the service fails? | Trace the failure path. If state is created before the risky call, test that failure cleans up or that retry is idempotent. |
| **What other interfaces expose this?** Mixins, DSLs, alternative entry points. | Grep for the method/behavior in related classes. If parity is needed, add it now. |

**When to skip:** Leaf-node changes with no callbacks, no state persistence, no parallel interfaces. Purely additive changes (new helper, new partial) take 10 seconds to confirm "nothing fires, skip."

### Step 3.6: Diff-Scope Verification

After a slice completes, verify that the files actually changed match the slice's declared `expected_files`. This is a hard gate — the agent does not self-report, the actual git diff is checked.

1. **Get the actual diff:**

   ```bash
   git diff --name-only $(git merge-base HEAD main)..HEAD
   ```

   This produces the list of files modified by this slice relative to the branch baseline.

2. **Load the declared scope** from the slice plan's `expected_files` frontmatter field.

3. **Compare and enforce:**

   Apply the same convention-matched test file expansion as Step 3.0 — test files that correspond to an `expected_files` entry are automatically in scope.

   | Finding | Action |
   |---------|--------|
   | Files changed that are NOT in `expected_files` and not convention-matched tests | **STOP.** Flag each out-of-scope file. Do not proceed to the next slice. Ask the user whether to amend the plan, revert the change, or accept the scope expansion. |
   | Files in `expected_files` that were NOT changed | Flag as potentially incomplete. Ask the user whether the slice is truly done or if work was missed. |
   | Perfect match (including auto-allowed tests) | Proceed. |

4. **Log results** in the kanban manifest under the slice's `notes` field:

   ```text
   notes: "scope-check: 5/5 expected files changed, 0 out-of-scope"
   ```

5. **If the slice plan has no `expected_files` field**, the gate fails. Stop and require the field to be added before proceeding. Do not infer or guess the expected files — the plan must declare them explicitly.

This gate is mandatory. It cannot be skipped, overridden, or deferred.

### Step 3.7: Destructive Command Guard

Before executing any shell command during a slice, check it against this blocklist:

| Blocked Pattern | Why |
|-----------------|-----|
| `rm -rf`, `rm` with recursive/force flags | Irreversible file deletion |
| `git push --force` / `git push -f` | Rewrites remote history |
| `git reset --hard` | Discards uncommitted work |
| `DROP TABLE` / `DROP DATABASE` / `TRUNCATE` | Irreversible data loss |
| `git clean -fd` | Deletes untracked files permanently |
| Bulk delete operations on files or data | Mass irreversible removal |
| Overwriting production config files | Environment-breaking changes |

**When a blocked command is detected:**

1. **STOP.** Do not execute.
2. Show the user the exact command and explain why it's blocked.
3. Wait for explicit HITL approval before proceeding.
4. If running in autonomous mode (no HITL available), skip the command and log in the manifest notes: `destructive-guard: blocked <command> — no HITL available`

This is enforcement, not a warning. The agent MUST NOT execute destructive commands without explicit human confirmation. This gate cannot be skipped, overridden, or deferred.

### Step 3.8: KB QA (all slices)

Invoke `kb-qa` with the current slice context. QA runs:

- **Lint check** on files in `expected_files` (every slice)
- **Browser verification** against acceptance criteria (frontend slices only)

If any check fails, `kb-qa` invokes `kb-repair` for surgical fixes (progress-based, 5-iteration cap). If repair exhausts or gets stuck, STOP — do not proceed to the next slice.

Backend-only slices skip browser checks but still run lint.

This gate is mandatory. It cannot be skipped or deferred.

### Step 3.9: Figma Design Sync (UI slices only)

If the slice involves UI changes and Figma designs exist:

1. Implement components following design specs
2. Use the **figma-design-sync** agent iteratively to compare
3. Fix visual differences identified
4. Repeat until implementation matches design

Skip this step entirely for non-UI slices.

### Step 4: Verify and Update

After the slice completes:

1. **Check result**
   - If yes: update manifest `status: done` for this slice and update the body table.
   - If no: update manifest `status: blocked`, add failure notes, and ask user how to proceed.

2. **Sync board** — update `docs/kanban.md` status for this slice (✅ or 🔒). Append validation note.

3. **Run verification**
   - Run the relevant test command for the repo.
   - If a full suite is too expensive or unavailable, explain the narrower verification used.

4. **Optional commit**
   - If the user asked for commits, stage only the manifest file for status updates and commit it separately.

5. Continue to the next runnable slice.

### Step 5: Completion

When all slices are `done` or intentionally `skipped`:

1. Update manifest `status: completed`.
2. Run final verification.
3. **Archive to board** — move the feature section from `docs/kanban.md` to `docs/kanban-done.md`. Prepend at the top of the archive file with a completion date header.
4. Report summary:

```text
KB <name> — all slices complete.
- N slices executed
- S slices skipped
- M tests added
- K files changed
Verification: <command/result>

Next: run kb-complete for review, documentation, and learning.
```

5. **Persist scope context** — collect the scope-verified file lists from each slice's `notes` field (the `scope-check:` entries from Step 3.6). Write the combined list to the manifest frontmatter as `scope-verified-files` so `kb-complete` can pass it to ce-review without re-deriving.

**Post-work steps (ce-review, compound, learn, evolve, cleanup) are handled by `kb-complete`.** If running via `klfg`, the pipeline prompts to continue automatically. If running standalone, invoke `kb-complete <manifest-path>` manually when ready.

## Failure Handling

| Situation | Action |
|-----------|--------|
| Slice execution fails | Mark blocked, show error, ask user |
| Test suite fails after a slice | Mark blocked, show which tests broke, ask user |
| HITL pause | Present context, wait for user, record decision |
| User says "abort" | Mark remaining slices as `pending`, stop |
| User says "skip" | Mark slice `skipped`, continue to next runnable slice |

## Resume Support

KB work is resumable:

- Manifest tracks status per slice.
- Re-running `kb-work` on the same manifest picks up where it left off.
- Already done or skipped slices are not rerun.

## Success Criteria

- No slice runs before its blockers are complete.
- Manifest frontmatter and body table reflect actual slice status.
- Each completed slice has verification evidence recorded in the final response or failure notes.
- No unrelated files are staged, committed, reverted, or overwritten.

## Integration

- **Input from:** `kb-plan`
- **Deepening:** Optionally runs `deepen-plan` per slice before execution
- **Execution engine:** Fresh sub-agents when available, local execution otherwise
- **Verification:** Invokes `tdd` skill principles per slice when verification mode requires it
- **Post-completion:** Invoke `kb-complete` for review, documentation, and learning
