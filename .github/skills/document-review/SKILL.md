---
name: document-review
description: "Optionally review one requirements or plan document with one best-fit specialist when the main-agent self-check leaves material uncertainty."
argument-hint: "[mode:headless] [path/to/document.md]"
---

# Document Review - One Best-Fit Specialist

This is an optional uncertainty reducer, not a mandatory planning phase.
`kb-brainstorm` or `kb-plan` first performs its own requirements check. Invoke
this skill only when one unresolved concern could materially change intent,
scope, feasibility, flow, trust, or decomposition.

## Input and Mode

Parse `mode:headless` as a flag and the remaining token as the document path.
Headless mode requires a path, never asks questions, and returns structured
findings plus a reusable receipt.

Classify the document as:

- `requirements` for product intent, behavior, constraints, and success;
- `plan` for decomposition, dependencies, ownership, and verification.

## Main-Agent Self-Check

Before dispatch, verify:

1. The goal and non-goals are explicit.
2. Requirements do not contradict each other.
3. Acceptance criteria are observable and testable.
4. Dependencies and constraints are supported by repo evidence or labeled
   assumptions.
5. Risks with multiple reasonable answers are visible rather than silently
   chosen.

If this resolves the uncertainty, do not dispatch. Return
`review-status: not-required` with the specific reason.

## Selection

Select exactly one reviewer whose lens best matches the dominant unresolved
uncertainty. Never run always-on reviewers and never stack personas.

| Dominant uncertainty | Reviewer |
|---|---|
| Internal contradiction or terminology drift | `coherence-reviewer` |
| Implementability, dependency, rollout, or architecture feasibility | `feasibility-reviewer` |
| Product premise, priority, value, or opportunity cost | `product-lens-reviewer` |
| Information architecture, interaction, accessibility, or visual behavior | `design-lens-reviewer` |
| Multi-step, branching, retry, cancellation, or recovery flow | `spec-flow-analyzer` |
| Auth, privacy, data exposure, credentials, or trust boundary | `security-lens-reviewer` |
| Scope growth, weak goal alignment, or unjustified complexity | `scope-guardian-reviewer` |
| Broad assumptions or a plan that needs adversarial stress | `adversarial-document-reviewer` |

When two concerns exist, choose the one most likely to change the document and
put the secondary concern into the same prompt. One reviewer remains
accountable for the whole document.

## Universal Reviewer Contract

The selected reviewer reads the full document and must check:

- consistency and authoritative intent;
- feasibility against known constraints;
- requirement and flow completeness;
- testability of acceptance criteria;
- its specialist risk.

Use `references/subagent-template.md` and
`references/findings-schema.json`. The reviewer is read-only and returns one
structured result.

## Findings

Drop malformed or unevidenced findings. Suppress confidence below 0.50.
Deduplication is local because only one reviewer runs.

| Autofix class | Action |
|---|---|
| `auto` | Apply one mechanically correct reconciliation in one edit pass |
| `present` | Return the tradeoff or unresolved judgment to the caller |

P0/P1 must be resolved before slicing. P2/P3 are constraints or improvements,
not automatic blockers. Headless mode never asks questions.

## Receipt

After auto-fixes, bind the receipt to the final document SHA-256 and write:

`docs/results/document-reviews/<document-slug>-<source-sha12>.json`

Include:

- `review_id`, source path/hash, timestamp, document type, and mode;
- one selected/completed persona, or none for `not-required`;
- the specific selection or not-required reason;
- finding counts, unresolved P0/P1, residual findings, and constraints;
- failed persona information when dispatch failed.

The receipt authorizes planning only when the selected reviewer completed and
unresolved P0/P1 counts are zero.

## Boundaries

- Maximum one reviewer call for the requirements boundary.
- Never review one slice at a time.
- A later plan review is a separate boundary and still limited to one reviewer.
- Requirements edits invalidate the receipt.
- Do not rewrite the whole document, invent scope, or implement code.

## Lazy References

- `references/subagent-template.md` - single-reviewer prompt.
- `references/findings-schema.json` - finding contract.
- `references/review-output-template.md` - interactive presentation.
