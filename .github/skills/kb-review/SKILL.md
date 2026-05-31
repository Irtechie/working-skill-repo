---
name: kb-review
description: "Structured KB review using tiered persona agents, confidence-gated findings, thermonuclear structural-quality review, and a merge/dedup pipeline. Use when reviewing KB workflow code changes before completion, before creating a PR, or when kb-complete needs its review gate."
argument-hint: "[diff, branch, manifest, or mode]"
---

# KB Review

Reviews code changes using reviewer personas. Prefer authorized reviewer
subagents; otherwise run structured local review and label it
`review-mode: local-fallback`.

## When to Use

- Before creating a PR
- After completing a task during iterative implementation
- When feedback is needed on any code changes
- Can be invoked standalone
- Can run as a read-only or autofix review step inside larger workflows

## Argument Parsing

Parse `$ARGUMENTS` for the following optional tokens. Strip each recognized token before interpreting the remainder as the PR number, GitHub URL, or branch name.

| Token | Example | Effect |
|-------|---------|--------|
| `mode:autofix` | `mode:autofix` | Select autofix mode (see Mode Detection below) |
| `mode:report-only` | `mode:report-only` | Select report-only mode |
| `mode:headless` | `mode:headless` | Select headless mode for programmatic callers (see Mode Detection below) |
| `base:<sha-or-ref>` | `base:abc1234` or `base:origin/main` | Skip scope detection — use this as the diff base directly |
| `plan:<path>` | `plan:docs/plans/2026-03-25-001-feat-foo-plan.md` | Load this plan for requirements verification |

All tokens are optional. Each one present means one less thing to infer. When absent, fall back to existing behavior for that stage.

**Conflicting mode flags:** If multiple mode tokens appear in arguments, stop and do not dispatch agents. If `mode:headless` is one of the conflicting tokens, emit the headless error envelope: `Review failed (headless mode). Reason: conflicting mode flags — <mode_a> and <mode_b> cannot be combined.` Otherwise emit the generic form: `Review failed. Reason: conflicting mode flags — <mode_a> and <mode_b> cannot be combined.`

## Mode Detection

| Mode | When | Behavior |
|------|------|----------|
| **Interactive** (default) | No mode token present | Review, apply safe_auto fixes automatically, present findings, ask for policy decisions on gated/manual findings, and optionally continue into fix/push/PR next steps |
| **Autofix** | `mode:autofix` in arguments | No user interaction. Review, apply only policy-allowed `safe_auto` fixes, re-review in bounded rounds, write a run artifact, and emit residual downstream work when needed |
| **Report-only** | `mode:report-only` in arguments | Strictly read-only. Review and report only, then stop with no edits, artifacts, todos, commits, pushes, or PR actions |
| **Headless** | `mode:headless` in arguments | Programmatic mode for skill-to-skill invocation. Apply `safe_auto` fixes silently (single pass), return all other findings as structured text output, write run artifacts, skip todos, and return "Review complete" signal. No interactive prompts. |

### Autofix mode rules

- **Skip all user questions.** Never pause for approval or clarification once scope has been established.
- **Apply only `safe_auto -> review-fixer` findings.** Leave `gated_auto`, `manual`, `human`, and `release` work unresolved.
- **Write a run artifact** under `.context/compound-engineering/kb-review/<run-id>/` summarizing findings, applied fixes, residual actionable work, and advisory outputs.
- **Create durable root `todo.md` entries only for unresolved actionable findings** whose final owner is `downstream-resolver`. Load `todo-create` for the KB board format.
- **Never commit, push, or create a PR** from autofix mode. Parent workflows own those decisions.

### Report-only mode rules

- **Skip all user questions.** Infer intent conservatively if the diff metadata is thin.
- **Never edit files or externalize work.** Do not write `.context/compound-engineering/kb-review/<run-id>/`, do not create todo files, and do not commit, push, or create a PR.
- **Safe for parallel read-only verification.** `mode:report-only` is the only mode that is safe to run concurrently with browser testing on the same checkout.
- **Do not switch the shared checkout.** If the caller passes an explicit PR or branch target, `mode:report-only` must run in an isolated checkout/worktree or stop instead of running `gh pr checkout` / `git checkout`.
- **Do not overlap mutating review with browser testing on the same checkout.** If a future orchestrator wants fixes, run the mutating review phase after browser testing or in an isolated checkout/worktree.

### Headless mode rules

- **Skip all user questions.** Never use the platform question tool (`ask_user` in Copilot CLI) or other interactive prompts. Infer intent conservatively if the diff metadata is thin.
- **Require a determinable diff scope.** If headless mode cannot determine a diff scope (no branch, PR, or `base:` ref determinable without user interaction), emit `Review failed (headless mode). Reason: no diff scope detected. Re-invoke with a branch name, PR number, or base:<ref>.` and stop without dispatching agents.
- **Apply only `safe_auto -> review-fixer` findings in a single pass.** No bounded re-review rounds. Leave `gated_auto`, `manual`, `human`, and `release` work unresolved and return them in the structured output.
- **Return all non-auto findings as structured text output.** Use the headless output envelope format from `references/review-process.md`, preserving severity, autofix_class, owner, requires_verification, confidence, evidence[], and pre_existing per finding.
- **Write a run artifact** under `.context/compound-engineering/kb-review/<run-id>/` summarizing findings, applied fixes, and advisory outputs. Include the artifact path in the structured output.
- **Do not create todo files.** The caller receives structured findings and routes downstream work itself.
- **Do not switch the shared checkout.** If the caller passes an explicit PR or branch target, `mode:headless` must run in an isolated checkout/worktree or stop instead of running `gh pr checkout` / `git checkout`. When stopping, emit `Review failed (headless mode). Reason: cannot switch shared checkout. Re-invoke with base:<ref> to review the current checkout, or run from an isolated worktree.`
- **Not safe for concurrent use on a shared checkout.** Unlike `mode:report-only`, headless mutates files (applies `safe_auto` fixes). Callers must not run headless concurrently with other mutating operations on the same checkout.
- **Never commit, push, or create a PR** from headless mode. The caller owns those decisions.
- **End with "Review complete" as the terminal signal** so callers can detect completion. If all reviewers fail or time out, emit `Code review degraded (headless mode). Reason: 0 of N reviewers returned results.` followed by "Review complete".

## Severity Scale

All reviewers use P0-P3:

| Level | Meaning | Action |
|-------|---------|--------|
| **P0** | Critical breakage, exploitable vulnerability, data loss/corruption | Must fix before merge |
| **P1** | High-impact defect likely hit in normal usage, breaking contract | Should fix |
| **P2** | Moderate issue with meaningful downside (edge case, perf regression, maintainability trap) | Fix if straightforward |
| **P3** | Low-impact, narrow scope, minor improvement | User's discretion |

## Action Routing

Severity answers **urgency**. Routing answers **who acts next** and **whether this skill may mutate the checkout**.

| `autofix_class` | Default owner | Meaning |
|-----------------|---------------|---------|
| `safe_auto` | `review-fixer` | Local, deterministic fix suitable for the in-skill fixer when the current mode allows mutation |
| `gated_auto` | `downstream-resolver` or `human` | Concrete fix exists, but it changes behavior, contracts, permissions, or another sensitive boundary that should not be auto-applied by default |
| `manual` | `downstream-resolver` or `human` | Actionable work that should be handed off rather than fixed in-skill |
| `advisory` | `human` or `release` | Report-only output such as learnings, rollout notes, or residual risk |

Routing rules:

- **Synthesis owns the final route.** Persona-provided routing metadata is input, not the last word.
- **Choose the more conservative route on disagreement.** A merged finding may move from `safe_auto` to `gated_auto` or `manual`, but never the other way without stronger evidence.
- **Only `safe_auto -> review-fixer` enters the in-skill fixer queue automatically.**
- **`requires_verification: true` means a fix is not complete without targeted tests, a focused re-review, or operational validation.**

## Reviewers

17 reviewer personas in layered conditionals, plus shared learning/runtime agents. KB review replaces the standard maintainability reviewer with the thermonuclear code-quality reviewer. Load `references/persona-catalog.md` only when the full catalog is needed.

## Review Mode

Record one mode: `review-mode: multi-agent` only when reviewer agents actually
ran; otherwise `review-mode: local-fallback` with the reason and local lenses
used. Local fallback is valid, but must not be described as multi-agent review.
If fallback coverage is materially weaker, report that as residual risk.

### Runtime Agent Types

`kb-review` is this skill/orchestrator, not an Agent tool `agent_type`.

Never call the Agent tool with `agent_type: kb-review`. When this skill needs subagents, use runtime-valid reviewer agent types such as `correctness-reviewer`, `testing-reviewer`, `thermo-nuclear-code-quality-reviewer`, `project-standards-reviewer`, `security-reviewer`, `performance-reviewer`, `api-contract-reviewer`, `data-migrations-reviewer`, `reliability-reviewer`, `adversarial-reviewer`, `cli-readiness-reviewer`, `previous-comments-reviewer`, `dhh-rails-reviewer`, `kieran-rails-reviewer`, `kieran-python-reviewer`, `kieran-typescript-reviewer`, `julik-frontend-races-reviewer`, `schema-drift-detector`, `deployment-verification-agent`, `agent-native-reviewer`, `learnings-researcher`, or `repo-research-analyst`.

If a desired persona does not exist as a runtime agent type, use `code-review` with the persona instructions in the task prompt rather than inventing an agent type. A broader valid reviewer is better than a failed dispatch.

**Always-on (every review):**

| Agent | Focus |
|-------|-------|
| `correctness-reviewer` | Logic errors, edge cases, state bugs, error propagation |
| `testing-reviewer` | Coverage gaps, weak assertions, brittle tests |
| `thermo-nuclear-code-quality-reviewer` | Structural simplification, code-judo opportunities, file sprawl, spaghetti branching, abstraction and boundary debt |
| `project-standards-reviewer` | AGENTS.md compliance -- frontmatter, references, naming, portability |
| `agent-native-reviewer` | Verify new features are agent-accessible |
| `learnings-researcher` | Search docs/solutions/ for past issues related to this PR |

**Cross-cutting conditional (selected per diff):**

| Agent | Select when diff touches... |
|-------|---------------------------|
| `security-reviewer` | Auth, public endpoints, user input, permissions |
| `performance-reviewer` | DB queries, data transforms, caching, async |
| `api-contract-reviewer` | Routes, serializers, type signatures, versioning |
| `data-migrations-reviewer` | Migrations, schema changes, backfills |
| `reliability-reviewer` | Error handling, retries, timeouts, background jobs |
| `adversarial-reviewer` | Diff >=50 changed non-test/non-generated/non-lockfile lines, or auth, payments, data mutations, external APIs |
| `cli-readiness-reviewer` | CLI command definitions, argument parsing, CLI framework usage, command handler implementations |
| `previous-comments-reviewer` | Reviewing a PR that has existing review comments or threads |

**Stack-specific conditional (selected per diff):**

| Agent | Select when diff touches... |
|-------|---------------------------|
| `dhh-rails-reviewer` | Rails architecture, service objects, session/auth choices, or Hotwire-vs-SPA boundaries |
| `kieran-rails-reviewer` | Rails application code where conventions, naming, and maintainability are in play |
| `kieran-python-reviewer` | Python modules, endpoints, scripts, or services |
| `kieran-typescript-reviewer` | TypeScript components, services, hooks, utilities, or shared types |
| `julik-frontend-races-reviewer` | Stimulus/Turbo controllers, DOM events, timers, animations, or async UI flows |

**Shared conditional (migration-specific):**

| Agent | Select when diff includes migration files |
|-------|------------------------------------------|
| `schema-drift-detector` | Cross-references schema changes against included migrations |
| `deployment-verification-agent` | Produces deployment checklist with verification queries |

## Review Scope

Every review spawns all 4 always-on personas plus the 2 shared learning/runtime agents, then adds whichever cross-cutting and stack-specific conditionals fit the diff. The model naturally right-sizes: a small config change triggers 0 conditionals = 6 reviewers. A Rails auth feature might trigger security + reliability + kieran-rails + dhh-rails = 10 reviewers.

## Protected Artifacts

The following paths are compound-engineering pipeline artifacts and must never be flagged for deletion, removal, or gitignore by any reviewer:

- `docs/brainstorms/*` -- requirements documents created by kb-brainstorm
- `docs/plans/*.md` -- plan files created by kb-plan (living documents with progress checkboxes)
- `docs/solutions/*.md` -- solution documents created during the pipeline

If a reviewer flags any file in these directories for cleanup or removal, discard that finding during synthesis.

## How to Run

Use the review process reference only when executing a review. It contains scope detection, intent discovery, plan discovery, reviewer selection, sub-agent dispatch, merge/dedup, synthesis, and headless output details.

Load `references/review-process.md` when:

- a review is actually being run;
- scope/base/PR handling is needed;
- reviewer selection or dispatch details are needed;
- findings need merging, deduplication, or headless formatting.

Default flow:

1. Determine diff scope and base.
2. Discover intent and relevant plan/requirements context.
3. Select always-on and conditional reviewers.
4. Spawn valid reviewer agents with `references/subagent-template.md` and `references/findings-schema.json`.
5. Merge, confidence-gate, deduplicate, and synthesize findings.
6. Apply the quality gates below before reporting.

## Quality Gates

Before delivering the review, verify:

1. **Every finding is actionable.** Re-read each finding. If it says "consider", "might want to", or "could be improved" without a concrete fix, rewrite it with a specific action. Vague findings waste engineering time.
2. **No false positives from skimming.** For each finding, verify the surrounding code was actually read. Check that the "bug" isn't handled elsewhere in the same function, that the "unused import" isn't used in a type annotation, that the "missing null check" isn't guarded by the caller.
3. **Severity is calibrated.** A style nit is never P0. A SQL injection is never P3. Re-check every severity assignment.
4. **Line numbers are accurate.** Verify each cited line number against the file content. A finding pointing to the wrong line is worse than no finding.
5. **Protected artifacts are respected.** Discard any findings that recommend deleting or gitignoring files in `docs/brainstorms/`, `docs/plans/`, or `docs/solutions/`.
6. **Findings don't duplicate linter output.** Don't flag things the project's linter/formatter would catch (missing semicolons, wrong indentation). Focus on semantic issues.

## Language-Aware Conditionals

This skill uses stack-specific reviewer agents when the diff clearly warrants them. Keep those agents opinionated. They are not generic language checkers; they add a distinct review lens on top of the always-on and cross-cutting personas.

Do not spawn them mechanically from file extensions alone. The trigger is meaningful changed behavior, architecture, or UI state in that stack.

## After Review

Use the post-review flow reference only after findings have been synthesized.

Load `references/post-review-flow.md` when:

- `safe_auto` findings need to be applied;
- interactive, autofix, report-only, or headless post-review behavior matters;
- residual actionable work needs an artifact or todo handoff;
- the user wants to push fixes or create a PR after review.

Default rule: apply only `safe_auto -> review-fixer` automatically when the active mode allows mutation. Do not widen fixes by severity alone. Never commit, push, or create a PR from autofix/headless/report-only modes.

## Fallback

If the platform cannot run reviewer subagents, run a local structured review and
label the output `review-mode: local-fallback`. If sequential valid reviewer
agents are available but parallel dispatch is not, sequential agents may still
count as `review-mode: multi-agent`; local-only review may not.

---

## Lazy References

Load these only when the named phase needs them:

- `references/persona-catalog.md` - reviewer catalog and selection matrix.
- `references/review-process.md` - full review execution flow.
- `references/post-review-flow.md` - fixer, artifact, todo, push, and PR behavior after review.
- `references/subagent-template.md` - prompt contract for reviewer subagents.
- `references/diff-scope.md` - primary/secondary/pre-existing scope rules.
- `references/findings-schema.json` - structured finding contract.
- `references/review-output-template.md` - report formatting.
