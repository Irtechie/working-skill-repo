---
name: kb-review
description: "Run one evidence-bound semantic code review profile after integrated proof. Use at KB completion, before a PR, or for an explicit standalone review."
argument-hint: "[mode:report-only|mode:autofix|mode:headless] [base:<ref>] [plan:<path>]"
---

# KB Review - One Profile, One Boundary

Review the integrated change once. Deterministic tests prove encoded behavior;
this review checks intent, test validity, correctness, and code health.

## Invariants

1. Run at most one reviewer profile for this review boundary.
2. The broad profile is the default. A specialist replaces it; never stack
   broad plus specialist.
3. Every selected profile must answer all four questions:
   - Does the diff satisfy the authoritative intent or specification?
   - Would the proof detect relevant breakage rather than merely execute code?
   - Is the implementation correct, including failure paths and edge cases?
   - Is the code healthy: clear boundaries, minimal complexity, and no
     avoidable structural debt?
4. Review only after the caller has a passing integrated proof receipt.
5. A code-affecting fix invalidates both the review and affected proof. Rerun
   affected deterministic proof, then run one bounded confirmation review.
6. Never claim multi-agent review. This skill dispatches zero or one reviewer.

## Modes

| Mode | Behavior |
|---|---|
| interactive | Report findings; apply only clearly safe fixes and ask only for genuine policy decisions |
| `mode:report-only` | Read-only; no files, todos, commits, pushes, or PR changes |
| `mode:autofix` | Apply only deterministic `safe_auto` fixes; leave other findings for the caller |
| `mode:headless` | Non-interactive caller mode; return structured findings and a receipt |

Conflicting mode flags fail before scope discovery or dispatch.

## Preflight

1. Determine the reviewed base from `base:<ref>`, caller scope, or the
   fork-safe base resolver in `references/resolve-base.sh`.
2. Read the authoritative requirements, manifest, plan, PR body, or issue.
3. Require a proof receipt bound to the integrated tree. If the caller cannot
   provide one, return `review-blocked: integrated-proof-missing`.
4. Collect the exact changed paths. Exclude unrelated pre-existing work and
   identify untracked files that are outside the review.
5. Preserve `docs/brainstorms/`, `docs/plans/`, and `docs/solutions/` as
   protected workflow artifacts.

Load `references/review-process.md` only while executing these steps.

## Skip Classification

Semantic review may be skipped only when every changed path is:

- documentation-only with no executable contract change;
- generated-only from an already-proven generator; or
- mechanically constrained by deterministic validation that covers the full
  changed surface.

Runtime, behavior, contract, configuration, trust-boundary, persistence, API,
CLI, or UI changes cannot skip. Unknown classification reviews rather than
skips. A skip still requires proof covering every changed path and a receipt
with the skip reason.

## Profile Selection

Choose exactly one profile using evidence from the diff.

| Evidence | Profile |
|---|---|
| Exploitable security or trust-boundary risk | `security-reviewer` |
| Migration, backfill, or persistent data transformation | `data-migrations-reviewer` |
| Runtime scaling or materially expensive I/O/query behavior | `performance-reviewer` |
| Retry, timeout, async, queue, or failure-recovery behavior | `reliability-reviewer` |
| Public API or serialization contract change | `api-contract-reviewer` |
| CLI contract or command-handler change | `cli-readiness-reviewer` |
| Large structural refactor or code-health risk dominates | `thermo-nuclear-code-quality-reviewer` |
| Everything else, including unknown risk | `code-review` broad profile |

A specialist prompt must include the four invariant questions. Domain focus
changes emphasis, not coverage. If the exact specialist is unavailable, use
`code-review` with the specialist instructions instead of adding another
reviewer.

## Dispatch Contract

Dispatch one read-only reviewer with:

- authoritative intent and requirements hash;
- base and integrated tree identifiers;
- proof receipt path and hash;
- risk classification and why the profile was selected;
- exact file list and diff;
- the four invariant questions;
- `references/findings-schema.json`.

Use `references/subagent-template.md` for the prompt contract. If no reviewer
agent is available, perform one local structured pass and record
`review-mode: local-fallback`; do not simulate several personas.

## Findings

Use P0-P3 severity. Keep only actionable, evidenced findings. Verify cited code
before reporting and suppress formatter/linter output.

| Class | Route |
|---|---|
| `safe_auto` | May be fixed in interactive, autofix, or headless mode |
| `gated_auto` | Caller resolves because behavior or contracts may change |
| `manual` | Caller or human owns a non-mechanical change |
| `advisory` | Report residual risk without pretending it is implementation work |

P0/P1 block completion until resolved. P2/P3 do not block by severity alone,
but fix cheap and clearly correct issues.

## Receipt

Write or return one review receipt containing:

- base tree and integrated tree;
- requirements source and SHA-256;
- proof receipt path and SHA-256;
- review-policy version;
- risk classification and selected profile;
- review mode and reviewer provenance;
- finding counts, resolutions, and residual risks;
- changed-path scope and skip reason when review was skipped.

The caller owns receipt storage. `kb-finalize` normally stores it with the
manifest proof artifacts.

## Stop Rules

- Do not dispatch without deterministic integrated proof.
- Do not dispatch a second reviewer because the first found nothing.
- Do not add a specialist after the broad profile.
- Do not rerun review for telemetry or model provenance.
- Do not commit, push, merge, or create a PR.

## Lazy References

- `references/review-process.md` - scope, profile selection, dispatch, and receipt flow.
- `references/subagent-template.md` - single-reviewer prompt contract.
- `references/diff-scope.md` - primary, secondary, and pre-existing scope.
- `references/findings-schema.json` - structured finding contract.
- `references/post-review-flow.md` - bounded fix and confirmation-review behavior.
- `references/review-output-template.md` - human-readable output.
- `references/persona-catalog.md` - replacement profile catalog.
