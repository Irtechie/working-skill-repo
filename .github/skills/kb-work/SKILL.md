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
3. Confirm execution once unless the user already asked to run/execute/work the manifest.
4. Execute ready slices in topological order without asking between slices.
5. Update the manifest after each slice so the workflow is resumable.

## Input

<input> #$ARGUMENTS </input>

**If input is empty:** Scan `docs/plans/` for the most recent `*-kb-*-manifest.md` file. If found, use it. Otherwise ask: "Which KB manifest should I execute?"

**If input is a path:** Read the manifest at that path.

**If input is a handoff:** Do not execute the handoff directly. If it links a `docs/plans/*-kb-*-manifest.md`, use that manifest. If it contains only phases, workstreams, bullets, or broad next steps, stop and invoke `kb-plan` to create vertical slices first.

## Pre-Flight

1. **Read the manifest** - parse the YAML frontmatter to get the ordered slice list.
2. **Validate DAG** - confirm no cycles in blockers, all referenced slice IDs exist, and all slice files exist.
3. **Validate slice contracts** - each slice plan must have `expected_files`, `verification`, `blockers`, `status`, and acceptance criteria. New slice plans should also have `test_level` and `functional_risk`. If core fields are missing, stop and route to `kb-plan`; do not infer a manifest from a phase list. If only `test_level` or `functional_risk` is missing on an older plan, invoke `kb-functional-test` to classify them before execution.
4. **Check status** - skip any slices already marked `done`. Resume from the first runnable `pending` slice.
5. **Check worktree** - note dirty or untracked files before executing so unrelated user changes are not staged or reverted.
6. **Sync with board** - read `todo.md` and confirm its status table matches the manifest. If they diverge, the board wins — another agent may have updated it. Reconcile the manifest from the board before proceeding.
7. **Confirm once only when needed:** If the user did not explicitly ask to run/execute/work the manifest, ask: "Ready to execute N remaining slices in order. Proceed?" If the user already asked to execute, continue without this prompt.

After initial execution starts, do not ask before moving from one runnable slice to the next.

Treat statuses as:

| Status | Action |
|--------|--------|
| `pending` | Eligible once blockers are `done` or `skipped` |
| `done` | Skip |
| `blocked` | Stop and ask whether to retry, skip, or abort |
| `human-required` | Waiting on human action; continue unrelated runnable slices if possible |
| `parked` | Intentionally out of bounds today; only a human promotes back to active |
| `skipped` | Skip but keep visible in the summary |

## Board Sync Protocol

`todo.md` is the live execution board. Update it at every status transition:

| Event | Board Update |
|-------|-------------|
| Starting a slice | Set status to 🔧 in_progress |
| Slice completes | Set status to ✅ done |
| Slice blocked | Set status to 🔒 blocked + reason in notes |
| Slice needs human action | Set status to 🛑 human-required + exact ask |
| Slice parked by human | Move to 🧊 Parked / Cold Storage with reason |
| Slice skipped | Set status to ⊘ skipped |
| All slices done | Prepend compact summary to `todo-done.md`, then remove completed feature section and routine completion logs from `todo.md` |

Active handoff files under `docs/handoffs/active/` are restart packets. Create or update one whenever work stops, blocks, or changes the next recommended action. Move completed handoffs to `docs/handoffs/done/`.

**Multi-agent rules:**
- Before claiming a slice, re-read `todo.md`. If another agent set it to 🔧, do not claim it.
- The board is the source of truth — not chat history, not the manifest. If the board says done, it's done.
- Update the board BEFORE starting work (claim) and AFTER completing work (release). This prevents two agents from working the same slice.
- Also update the manifest to stay in sync, but if they conflict, the board wins.
- Do not use root **Work Log** as a permanent archive. During execution, add notes only when they help a later agent resume: blockers, verification commands, durable memory impacts, or non-obvious decisions. Routine "slice complete" and verification-success notes belong in `todo-done.md` at feature completion, not in `todo.md`.
- Blocked is not parked. Use `🔒 blocked` for dependencies, another-agent waits, tool failures, or missing inputs. Use `🧊 Parked / Cold Storage` only for work a human intentionally deferred out of scope.

## Dependency Ordering

Execute with a topological sort:

1. Build a map of `slice_id -> slice`.
2. For each pending slice, check all `blockers`.
3. Run the first pending slice whose blockers are complete.
4. If pending slices remain but none are runnable, mark the manifest blocked and report the dependency problem.

## Execution Loop

For each slice in dependency order:

### Continuous Execution Rule

Slices should run continuously once execution has started.

Do **not** ask "Proceed to execute slice-N?" between slices. Move to the next runnable slice automatically after:

- slice status is updated;
- board and manifest are synced;
- required deterministic checks pass;
- QA/repair gates pass or are not applicable.

Pause only when a real gate requires it:

- HITL decision or missing value that cannot be generated safely;
- blocked/human-required/parked slice with no unrelated runnable work;
- destructive command approval;
- out-of-scope file edit or diff-scope failure;
- QA/repair exhaustion or stuck loop;
- dependency deadlock;
- user explicitly asked to pause or stop.

### Step 1: Check HITL Flag

If `hitl: true`:

- Present the slice title, description, and the specific question/decision needed.
- Classify the HITL item before stopping:
  - `critical-path` — later slices depend on this decision/access/input.
  - `parallel-blocker` — this slice is blocked, but unrelated slices can continue.
  - `final-validation` — human judgment is useful before release, but not needed for development.
  - `agent-runnable-with-inputs` — human only needs to provide values; the agent can run the check.
- Stop only the dependent path. If unrelated slices are runnable, mark this slice `blocked` or `human-required`, update `todo.md` and the manifest, then continue those slices.
- When marking a slice `human-required`, `parked`, or `blocked`, persist: `owner`, `blocked_reason`, `resume_when`, `next_agent_action`, `human_action`, `can_continue_other_slices`, and `parked_at`.
- Record the user's decision in the slice plan.
- Update manifest status to `done` for this slice only if the decision completes the slice.
- Continue to the next runnable slice.

Missing test inputs are not a reason to ask the user to manually test. If `test_inputs` are missing:

- Ask for the specific missing value.
- Use safe generated or fixture values when acceptable.
- If the input blocks only this slice, mark this slice `human-required` or `blocked` and continue unrelated runnable slices.
- Resume the slice after the value is available and run the verification yourself.

### Step 2: Deepen If Thin

If the slice plan has fewer than 3 acceptance criteria or no test scenarios:

- Run a lightweight deepening pass on this single slice.
- Add concrete test scenarios and likely file paths.
- Keep the pass bounded; do not re-plan the whole feature.

### Step 2.5: Test-Level Classification

Before editing, ensure the slice has a recorded test obligation:

- `test_level`: `none`, `unit`, `integration`, `functional-api`, `functional-cli`, `functional-browser`, or `full`
- `functional_risk`: `none`, `narrow`, `broad`, or `full`

If either field is missing, stale, or contradicted by the acceptance criteria or `expected_files`, invoke `kb-functional-test` with the slice plan. Record the result in the slice frontmatter or notes before implementation.

Use small/mini model subagents for this classification when the platform supports model-tiered agents. Keep the task bounded: classify the slice, audit existing tests for mocked theater, and suggest the narrowest deterministic proof. Escalate to the main model for complex architecture/auth/security/flaky async decisions or repeated test failures.

Do not use `unit` just because it is cheaper. Use `unit` only when unit-level proof can fail for the user-facing bug or behavior. If a unit test could pass while the workflow is broken, require integration or functional proof.

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

Use `references/execution-prompt.md` as the per-slice execution prompt/checklist. Load it only when starting a slice.
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

4. **Log results** in the KB manifest under the slice's `notes` field:

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
   - If no and repair/progress is still possible: run `kb-repair` or a bounded fix loop, then retry verification.
   - If no progress remains: update manifest `status: blocked` or `parked`, add failure notes and resume criteria, then continue unrelated runnable slices.

2. **Sync board** — update `todo.md` status for this slice (done or blocked). Append validation note.

3. **Run verification**
   - Invoke `kb-check` for deterministic verification.
   - Prefer existing scripts, lint, typecheck, tests, browser checks, builds, and CI-equivalent commands over LLM inspection.
   - If a full suite is too expensive or unavailable, run the narrowest deterministic check that proves the slice and record why.
   - Invoke `kb-functional-test` whenever `test_level` is `integration`, `functional-api`, `functional-cli`, `functional-browser`, or `full`, or when user-visible/cross-boundary changes appear despite a lower test level.
   - Record `test-level: <value>; functional-risk: <value>; proof: <command/artifact>` in the manifest notes.

4. **Assess memory impact**
   - Classify the slice as `memory-impact: none`, `operational`, or `durable`.
   - `none`: cosmetic, copy, formatting, lint-only, or isolated tests with no behavior/architecture change.
   - `operational`: active state, blockers, verification commands, or handoff instructions changed. Update `todo.md` or the active handoff.
   - `durable`: user-visible behavior, API/data/storage/auth/routing/streaming/tool/action/job/integration behavior, run/test commands, subsystem boundaries, sharp edges, or rejected approaches changed.
   - For durable changes, add a manifest note: `memory-impact: durable; areas=<areas>; docs=<candidate docs>; refresh=pending`.
   - If the affected doc is obvious and small, update it now. Otherwise leave `refresh=pending` for Step 5.

5. **Optional commit**
   - If the user asked for commits, stage only the manifest file for status updates and commit it separately.

6. Continue to the next runnable slice.

### Step 5: Completion

When all slices are `done` or intentionally `skipped`:

1. Update manifest `status: completed`.
2. Run final verification.
3. Run `kb-gate` if verification, QA, repair, or functional-test checks surfaced P0/P1/P2/P3/P4 issues. P0/P1 block completion while unresolved, but safe/actionable blockers should be rectified before asking the user. P2/P3/P4 do not block by severity alone.
4. **Refresh project memory** — if any slice has `memory-impact: durable` or `refresh=pending`, run `kb-map refresh` before archiving. Update affected architecture, operation, decision, research, `todo.md`, and handoff pointers. Add manifest note: `kb-map-refresh: done` or `kb-map-refresh: skipped - <reason>`.
5. **Archive to board** — move the feature summary from `todo.md` to `todo-done.md`. Prepend at the top of the archive file with a completion date header.
6. **Prune active board** — remove the completed feature section from `todo.md`. Also remove routine work-log entries for the completed feature from `todo.md`; keep only still-active, `🔒 blocked`, `🧊 parked`, `🛑 human-required`, or handoff-pointer items.
7. Report summary:

```text
KB <name> — all slices complete.
- N slices executed
- S slices skipped
- M tests added
- K files changed
Verification: <command/result>

Next: kb-complete runs automatically for review, documentation, and learning.
```

8. **Persist scope context** — collect the scope-verified file lists from each slice's `notes` field (the `scope-check:` entries from Step 3.6). Write the combined list to the manifest frontmatter as `scope-verified-files` so `kb-complete` can pass it to ce-review without re-deriving.

**Post-work steps (ce-review, compound, learn, evolve, cleanup) are handled by `kb-complete`.** After all slices are `done` or intentionally `skipped`, invoke `kb-complete <manifest-path>` automatically unless the user explicitly asked to stop after work execution.

## Failure Handling

| Situation | Action |
|-----------|--------|
| Slice execution fails with progress possible | Run `kb-repair` or a bounded fix loop, then retry verification |
| Slice execution fails with no progress | Mark only that slice blocked/parked, write resume packet, continue unrelated runnable slices |
| Test suite fails after a slice | Run `kb-repair`; if stuck, mark affected slice blocked/parked and continue unrelated runnable slices |
| HITL critical-path pause | Present context, wait for user, record decision |
| HITL not on critical path | Park the slice and continue unrelated runnable slices |
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
- **Deepening:** Use `kb-research` only when a slice has a material unresolved uncertainty before execution.
- **Execution engine:** Fresh sub-agents when available, local execution otherwise
- **Verification:** Invokes `tdd` skill principles per slice when verification mode requires it
- **Deterministic checks:** Invokes `kb-check` before a slice is marked done
- **Functional checks:** Invokes `kb-functional-test` for user-visible and cross-boundary behavior
- **Post-completion:** Automatically invoke `kb-complete` after all slices are done or intentionally skipped
