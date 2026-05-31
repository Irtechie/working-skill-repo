---
name: ce-compound-refresh
description: Refresh stale or drifting learnings and pattern docs in docs/solutions/ by reviewing, updating, consolidating, replacing, or deleting them against the current codebase. Use after refactors, migrations, dependency upgrades, or when a retrieved learning feels outdated or wrong. Also use when reviewing docs/solutions/ for accuracy, when a recently solved problem contradicts an existing learning, when pattern docs no longer reflect current code, or when multiple docs seem to cover the same topic and might benefit from consolidation.
argument-hint: "[docs/solutions path, topic, or drift reason]"
disable-model-invocation: true
---

# Compound Refresh

Maintain the quality of `docs/solutions/` over time. This workflow reviews existing learnings against the current codebase, then refreshes any derived pattern docs that depend on them.

## Mode Detection

Check if `$ARGUMENTS` contains `mode:autofix`. If present, strip it from arguments (use the remainder as a scope hint) and run in **autofix mode**.

| Mode | When | Behavior |
|------|------|----------|
| **Interactive** (default) | User is present and can answer questions | Ask for decisions on ambiguous cases, confirm actions |
| **Autofix** | `mode:autofix` in arguments | No user interaction. Apply all unambiguous actions (Keep, Update, Consolidate, auto-Delete, Replace with sufficient evidence). Mark ambiguous cases as stale. Generate a summary report at the end. |

### Autofix mode rules

- **Skip all user questions.** Never pause for input.
- **Process all docs in scope.** No scope narrowing questions — if no scope hint was provided, process everything.
- **Attempt all safe actions:** Keep (no-op), Update (fix references), Consolidate (merge and delete subsumed doc), auto-Delete (unambiguous criteria met), Replace (when evidence is sufficient). If a write succeeds, record it as **applied**. If a write fails (e.g., permission denied), record the action as **recommended** in the report and continue — do not stop or ask for permissions.
- **Mark as stale when uncertain.** If classification is genuinely ambiguous (Update vs Replace vs Consolidate vs Delete) or Replace evidence is insufficient, mark as stale with `status: stale`, `stale_reason`, and `stale_date` in the frontmatter. If even the stale-marking write fails, include it as a recommendation.
- **Use conservative confidence.** In interactive mode, borderline cases get a user question. In autofix mode, borderline cases get marked stale. Err toward stale-marking over incorrect action.
- **Always generate a report.** The report is the primary deliverable. It has two sections: **Applied** (actions that were successfully written) and **Recommended** (actions that could not be written, with full rationale so a human can apply them or run the skill interactively). The report structure is the same regardless of what permissions were granted — the only difference is which section each action lands in.

## Interaction Principles

**These principles apply to interactive mode only. In autofix mode, skip all user questions and apply the autofix mode rules above.**

Follow the same interaction style as `kb-brainstorm`:

- Ask questions **one at a time** — use the platform's blocking question tool when available (`ask_user` in Copilot CLI). Otherwise, present numbered options in plain text and wait for the user's reply before continuing
- Prefer **multiple choice** when natural options exist
- Start with **scope and intent**, then narrow only when needed
- Do **not** ask the user to make decisions before you have evidence
- Lead with a recommendation and explain it briefly

The goal is not to force the user through a checklist. The goal is to help them make a good maintenance decision with the smallest amount of friction.

## Refresh Order

Refresh in this order:

1. Review the relevant individual learning docs first
2. Note which learnings stayed valid, were updated, were consolidated, were replaced, or were deleted
3. Then review any pattern docs that depend on those learnings

Why this order:

- learning docs are the primary evidence
- pattern docs are derived from one or more learnings
- stale learnings can make a pattern look more valid than it really is

If the user starts by naming a pattern doc, you may begin there to understand the concern, but inspect the supporting learning docs before changing the pattern.

## Maintenance Model

For each candidate artifact, classify it into one of five outcomes:

| Outcome | Meaning | Default action |
|---------|---------|----------------|
| **Keep** | Still accurate and still useful | No file edit by default; report that it was reviewed and remains trustworthy |
| **Update** | Core solution is still correct, but references drifted | Apply evidence-backed in-place edits |
| **Consolidate** | Two or more docs overlap heavily but are both correct | Merge unique content into the canonical doc, delete the subsumed doc |
| **Replace** | The old artifact is now misleading, but there is a known better replacement | Create a trustworthy successor, then delete the old artifact |
| **Delete** | No longer useful, applicable, or distinct | Delete the file — git history preserves it if anyone needs to recover it later |

## Core Rules

1. **Evidence informs judgment.** The signals below are inputs, not a mechanical scorecard. Use engineering judgment to decide whether the artifact is still trustworthy.
2. **Prefer no-write Keep.** Do not update a doc just to leave a review breadcrumb.
3. **Match docs to reality, not the reverse.** When current code differs from a learning, update the learning to reflect the current code. The skill's job is doc accuracy, not code review — do not ask the user whether code changes were "intentional" or "a regression." If the code changed, the doc should match. If the user thinks the code is wrong, that is a separate concern outside this workflow.
4. **Be decisive, minimize questions.** When evidence is clear (file renamed, class moved, reference broken), apply the update. In interactive mode, only ask the user when the right action is genuinely ambiguous. In autofix mode, mark ambiguous cases as stale instead of asking. The goal is automated maintenance with human oversight on judgment calls, not a question for every finding.
5. **Avoid low-value churn.** Do not edit a doc just to fix a typo, polish wording, or make cosmetic changes that do not materially improve accuracy or usability.
6. **Use Update only for meaningful, evidence-backed drift.** Paths, module names, related links, category metadata, code snippets, and clearly stale wording are fair game when fixing them materially improves accuracy.
7. **Use Replace only when there is a real replacement.** That means either:
   - the current conversation contains a recently solved, verified replacement fix, or
   - the user has provided enough concrete replacement context to document the successor honestly, or
   - the codebase investigation found the current approach and can document it as the successor, or
   - newer docs, pattern docs, PRs, or issues provide strong successor evidence.
8. **Delete when the code is gone.** If the referenced code, controller, or workflow no longer exists in the codebase and no successor can be found, delete the file — don't default to Keep just because the general advice is still "sound." A learning about a deleted feature misleads readers into thinking that feature still exists. When in doubt between Keep and Delete, ask the user (in interactive mode) or mark as stale (in autofix mode). But missing referenced files with no matching code is **not** a doubt case — it is strong, unambiguous Delete evidence. Auto-delete it.
9. **Evaluate document-set design, not just accuracy.** In addition to checking whether each doc is accurate, evaluate whether it is still the right unit of knowledge. If two or more docs overlap heavily, determine whether they should remain separate, be cross-scoped more clearly, or be consolidated into one canonical document. Redundant docs are dangerous because they drift silently — two docs saying the same thing will eventually say different things.
10. **Delete, don't archive.** There is no `_archived/` directory. When a doc is no longer useful, delete it. Git history preserves every deleted file — that is the archive. A dedicated archive directory creates problems: archived docs accumulate, pollute search results, and nobody reads them. If someone needs a deleted doc, `git log --diff-filter=D -- docs/solutions/` will find it.

## Scope Selection

Start by discovering learnings and pattern docs under `docs/solutions/`.

Exclude:

- `README.md`
- `docs/solutions/_archived/` (legacy — if this directory exists, flag it for cleanup in the report)

Find all `.md` files under `docs/solutions/`, excluding `README.md` files and anything under `_archived/`. If an `_archived/` directory exists, note it in the report as a legacy artifact that should be cleaned up (files either restored or deleted).

If `$ARGUMENTS` is provided, use it to narrow scope before proceeding. Try these matching strategies in order, stopping at the first that produces results:

1. **Directory match** — check if the argument matches a subdirectory name under `docs/solutions/` (e.g., `performance-issues`, `database-issues`)
2. **Frontmatter match** — search `module`, `component`, or `tags` fields in learning frontmatter for the argument
3. **Filename match** — match against filenames (partial matches are fine)
4. **Content search** — search file contents for the argument as a keyword (useful for feature names or feature areas)

If no matches are found, report that and ask the user to clarify. In autofix mode, report the miss and stop — do not guess at scope.

If no candidate docs are found, report:

```text
No candidate docs found in docs/solutions/.
Run `ce-compound` after solving problems to start building your knowledge base.
```

## Phase 0: Assess and Route

Before asking the user to classify anything:

1. Discover candidate artifacts
2. Estimate scope
3. Choose the lightest interaction path that fits

### Route by Scope

| Scope | When to use it | Interaction style |
|-------|----------------|-------------------|
| **Focused** | 1-2 likely files or user named a specific doc | Investigate directly, then present a recommendation |
| **Batch** | Up to ~8 mostly independent docs | Investigate first, then present grouped recommendations |
| **Broad** | 9+ docs, ambiguous, or repo-wide stale-doc sweep | Triage first, then investigate in batches |

### Broad Scope Triage

When scope is broad (9+ candidate docs), do a lightweight triage before deep investigation:

1. **Inventory** — read frontmatter of all candidate docs, group by module/component/category
2. **Impact clustering** — identify areas with the densest clusters of learnings + pattern docs. A cluster of 5 learnings and 2 patterns covering the same module is higher-impact than 5 isolated single-doc areas, because staleness in one doc is likely to affect the others.
3. **Spot-check drift** — for each cluster, check whether the primary referenced files still exist. Missing references in a high-impact cluster = strongest signal for where to start.
4. **Recommend a starting area** — present the highest-impact cluster with a brief rationale and ask the user to confirm or redirect. In autofix mode, skip the question and process all clusters in impact order.

Example:

```text
Found 24 learnings across 5 areas.

The auth module has 5 learnings and 2 pattern docs that cross-reference
each other — and 3 of those reference files that no longer exist.
I'd start there.

1. Start with auth (recommended)
2. Pick a different area
3. Review everything
```

Do not ask action-selection questions yet. First gather evidence.

## Phase 1: Investigate Candidate Learnings

For each learning in scope, read it, cross-reference its claims against the current codebase, and form a recommendation.

A learning has several dimensions that can independently go stale. Surface-level checks catch the obvious drift, but staleness often hides deeper:

- **References** — do the file paths, class names, and modules it mentions still exist or have they moved?
- **Recommended solution** — does the fix still match how the code actually works today? A renamed file with a completely different implementation pattern is not just a path update.
- **Code examples** — if the learning includes code snippets, do they still reflect the current implementation?
- **Related docs** — are cross-referenced learnings and patterns still present and consistent?
- **Auto memory** — does the auto memory directory contain notes in the same problem domain? Read MEMORY.md from the auto memory directory (the path is known from the system prompt context). If it does not exist or is empty, skip this dimension. A memory note describing a different approach than what the learning recommends is a supplementary drift signal.
- **Overlap** — while investigating, note when another doc in scope covers the same problem domain, references the same files, or recommends a similar solution. For each overlap, record: the two file paths, which dimensions overlap (problem, solution, root cause, files, prevention), and which doc appears broader or more current. These signals feed Phase 1.75 (Document-Set Analysis).

Match investigation depth to the learning's specificity — a learning referencing exact file paths and code snippets needs more verification than one describing a general principle.

### Drift Classification: Update vs Replace

The critical distinction is whether the drift is **cosmetic** (references moved but the solution is the same) or **substantive** (the solution itself changed):

- **Update territory** — file paths moved, classes renamed, links broke, metadata drifted, but the core recommended approach is still how the code works. `ce-compound-refresh` fixes these directly.
- **Replace territory** — the recommended solution conflicts with current code, the architectural approach changed, or the pattern is no longer the preferred way. This means a new learning needs to be written. A replacement subagent writes the successor following `ce-compound`'s document format (frontmatter, problem, root cause, solution, prevention), using the investigation evidence already gathered. The orchestrator does not rewrite learnings inline — it delegates to a subagent for context isolation.

**The boundary:** if you find yourself rewriting the solution section or changing what the learning recommends, stop — that is Replace, not Update.

**Memory-sourced drift signals** are supplementary, not primary. A memory note describing a different approach does not alone justify Replace or Delete. Use memory signals to:
- Corroborate codebase-sourced drift (strengthens the case for Replace)
- Prompt deeper investigation when codebase evidence is borderline
- Add context to the evidence report ("(auto memory) notes suggest approach X may have changed since this learning was written")

In autofix mode, memory-only drift (no codebase corroboration) should result in stale-marking, not action.

### Judgment Guidelines

Three guidelines that are easy to get wrong:

1. **Contradiction = strong Replace signal.** If the learning's recommendation conflicts with current code patterns or a recently verified fix, that is not a minor drift — the learning is actively misleading. Classify as Replace.
2. **Age alone is not a stale signal.** A 2-year-old learning that still matches current code is fine. Only use age as a prompt to inspect more carefully.
3. **Check for successors before deleting.** Before recommending Replace or Delete, look for newer learnings, pattern docs, PRs, or issues covering the same problem space. If successor evidence exists, prefer Replace over Delete so readers are directed to the newer guidance.

## Phase 1.5: Investigate Pattern Docs

After reviewing the underlying learning docs, investigate any relevant pattern docs under `docs/solutions/patterns/`.

Pattern docs are high-leverage — a stale pattern is more dangerous than a stale individual learning because future work may treat it as broadly applicable guidance. Evaluate whether the generalized rule still holds given the refreshed state of the learnings it depends on.

A pattern doc with no clear supporting learnings is a stale signal — investigate carefully before keeping it unchanged.

## Phase 1.75: Document-Set Analysis

Load `references/document-set-analysis.md` when multiple docs may overlap, supersede one another, conflict, or need consolidation. For a focused single-doc refresh with no overlap signals, skip this reference.

## Subagent Strategy

Use subagents for context isolation when investigating multiple artifacts — not just because the task sounds complex. Choose the lightest approach that fits:

| Approach | When to use |
|----------|-------------|
| **Main thread only** | Small scope, short docs |
| **Sequential subagents** | 1-2 artifacts with many supporting files to read |
| **Parallel subagents** | 3+ truly independent artifacts with low overlap |
| **Batched subagents** | Broad sweeps — narrow scope first, then investigate in batches |

**When spawning any subagent, include this instruction in its task prompt:**

> Use dedicated file search and read tools (Glob, Grep, Read) for all investigation. Do NOT use shell commands (ls, find, cat, grep, test, bash) for file operations. This avoids permission prompts and is more reliable.
>
> Also read MEMORY.md from the auto memory directory if it exists. Check for notes related to the learning's problem domain. Report any memory-sourced drift signals separately from codebase-sourced evidence, tagged with "(auto memory)" in the evidence section. If MEMORY.md does not exist or is empty, skip this check.

There are two subagent roles:

1. **Investigation subagents** — read-only. They must not edit files, create successors, or delete anything. Each returns: file path, evidence, recommended action, confidence, and open questions. These can run in parallel when artifacts are independent.
2. **Replacement subagents** — write a single new learning to replace a stale one. These run **one at a time, sequentially** (each replacement subagent may need to read significant code, and running multiple in parallel risks context exhaustion). The orchestrator handles all deletions and metadata updates after each replacement completes.

The orchestrator merges investigation results, detects contradictions, coordinates replacement subagents, and performs all deletions/metadata edits centrally. In interactive mode, it asks the user questions on ambiguous cases. In autofix mode, it marks ambiguous cases as stale instead. If two artifacts overlap or discuss the same root issue, investigate them together rather than parallelizing.

## Phase 2: Classify the Right Maintenance Action

Choose one action per artifact: Keep, Update, Consolidate, Replace, or Delete.

Load `references/maintenance-actions.md` when action boundaries matter, especially Consolidate vs Delete, Update vs Replace, or problem-domain checks before deletion.

## Pattern Guidance

Apply the same five outcomes (Keep, Update, Consolidate, Replace, Delete) to pattern docs, but evaluate them as **derived guidance** rather than incident-level learnings. Key differences:

- **Keep**: the underlying learnings still support the generalized rule and examples remain representative
- **Update**: the rule holds but examples, links, scope, or supporting references drifted
- **Consolidate**: two pattern docs generalize the same set of learnings or cover the same design concern — merge into one canonical pattern
- **Replace**: the generalized rule is now misleading, or the underlying learnings support a different synthesis. Base the replacement on the refreshed learning set — do not invent new rules from guesswork
- **Delete**: the pattern is no longer valid, no longer recurring, or fully subsumed by a stronger pattern doc with no unique content remaining

## Phase 3: Ask for Decisions

In autofix mode, skip questions and apply safe unambiguous actions. In interactive mode, ask only when the maintenance action is genuinely ambiguous.

Load `references/decision-flow.md` when presenting decisions, batching broad refreshes, or choosing how to ask the user.

## Phase 4: Execute the Chosen Action

Execute the chosen action with the least file churn that restores truth: no-op Keep, evidence-backed Update, merge-and-delete Consolidate, successor-writing Replace, or direct Delete.

Load `references/execution-flows.md` before editing, replacing, consolidating, or deleting docs.

## Output and Commit

Report files reviewed, actions taken, stale/ambiguous items, and legacy cleanup. Commit only when the mode and git context allow it.

Load `references/reporting-and-commit.md` when producing the final report or committing refresh changes.

## Relationship to ce-compound

- `ce-compound` captures a newly solved, verified problem
- `ce-compound-refresh` maintains older learnings as the codebase evolves — both their individual accuracy and their collective design as a document set

Use **Replace** only when the refresh process has enough real evidence to write a trustworthy successor. When evidence is insufficient, mark as stale and recommend `ce-compound` for when the user next encounters that problem area.

Use **Consolidate** proactively when the document set has grown organically and redundancy has crept in. Every `ce-compound` invocation adds a new doc — over time, multiple docs may cover the same problem from slightly different angles. Periodic consolidation keeps the document set lean and authoritative.

## Discoverability Check

After the refresh report is generated, check whether project instructions surface `docs/solutions/` well enough for future agents to discover the knowledge store.

Load `references/discoverability-check.md` when running this check or when a missing instruction-file mention may need an edit/commit.

## Lazy References

Load these only when the named phase needs them:

- `references/document-set-analysis.md` - overlap, supersession, canonical-doc, retrieval-value, and conflict checks.
- `references/maintenance-actions.md` - Keep, Update, Consolidate, Replace, Delete boundaries.
- `references/decision-flow.md` - interactive and batch decision prompts.
- `references/execution-flows.md` - edit, consolidate, replace, delete, and stale-mark execution details.
- `references/reporting-and-commit.md` - output structure and commit behavior.
- `references/discoverability-check.md` - instruction-file awareness check for `docs/solutions/`.
