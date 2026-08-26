---
name: kb-brainstorm
description: "Proportional requirements discovery that orients from repo memory, researches only material uncertainty, and produces a planning-ready requirements source."
argument-hint: "[feature idea or problem]"
---

# KB Brainstorm - Requirements Before Slices

Answer what to build and why. Do not implement code or preselect slice owners.
Use repo-relative paths in every durable artifact.

## Input

<feature_description> #$ARGUMENTS </feature_description>

If input is empty, ask for the feature or problem. If a matching active
requirements document exists, ask whether to resume it rather than duplicating
work.

## Proportionality

Classify the scope as lightweight, standard, or deep.

- Lightweight: local orientation and at most two material questions.
- Standard: local constraints, relevant prior art, and bounded decision work.
- Deep: broader research and more questions only where they prevent rework.

Research is conditional. Invoke `kb-research` only when current external facts,
prior art, framework behavior, or failure modes could change framing. Record
citations or label claims as assumptions.

## Orientation

1. Resolve project root through `kb-map`.
2. Read the smallest relevant memory packet and governing instructions.
3. Inspect one or two nearby implementation or requirements patterns.
4. Verify checkable infrastructure claims against source.
5. Separate product decisions from planning details.

Do not interview the user before this orientation unless the topic itself is
unclear.

## Question Gate

Classify every material unknown:

| Class | Meaning |
|---|---|
| `ask-now` | Only the user can decide intent, scope, acceptance, authority, or irreversible risk |
| `research-first` | Evidence can answer before asking |
| `safe-assumption` | Reversible, low-risk assumption with evidence and a catching proof |
| `defer-to-planning` | Technical detail better resolved while slicing |
| `parked` | Explicitly outside current scope |

Do not hand off with unresolved `ask-now` or `research-first` items.
`safe-assumption` must record why it is reversible, its evidence, and the proof
that will catch an incorrect assumption.

Ask one question at a time through the platform question tool. Questions must
change behavior, scope, priority, risk, acceptance, or verification. Do not ask
the user to choose among agent-owned implementation fixes.

## Cheapest Sufficient Outcome

Establish that the need is real and unmet before writing requirements for it.
The two cheapest answers are requirement-level, not implementation detail:

| Tier | Answer |
|---|---|
| 1 | Nothing. The requirement does not survive contact with the problem - cut it. |
| 2 | Prior art already satisfies it. Point at the artifact instead of restating it. |

Prior art is not only local. When the topic touches a system the user already
operates, check sibling project memory and prior sessions before proposing new
behavior. Requirements written over unread prior art are the most expensive
rework in this workflow, because planning, slicing, and review all inherit the
mistake before anything reveals it.

Record the resolved tier and its evidence in the artifact. New behavior enters
requirements only after tiers 1 and 2 fail, and the ruled-out prior art must be
named specifically - "nothing existing covers this" without naming what was
checked is not evidence.

Trust boundaries, data-loss handling, security controls, and accessibility are
never charged to this ladder. They are funded at every tier.

## Requirements Work

Resolve:

1. Problem and target user/operator/maintainer.
2. Desired outcome and measurable success.
3. Must-have behavior and failure/recovery behavior.
4. Explicit non-goals.
5. Constraints, dependencies, and assumptions.
6. Trust, data, migration, integration, or compatibility boundaries.
7. Observable acceptance criteria.
8. Deferred planning questions and parked work.

Challenge weak premises using `kb-first-principles` behavior. Accept factual
corrections, verify disputed claims, and avoid pendulum-swings.

## Artifact

Write or update:

`docs/brainstorms/YYYY-MM-DD-<topic>-requirements.md`

Use `references/requirements-template.md`. Omit empty decorative sections, but
preserve goals, requirements, non-goals, decisions, assumptions, question-gate
state, acceptance criteria, evidence, and source paths.

## Requirements Self-Check

First perform the requirements self-check before considering reviewer dispatch.

Before handoff:

- confirm goals and non-goals are explicit;
- reconcile contradictions and terminology;
- ensure acceptance criteria are observable and testable;
- label unsupported dependencies as assumptions;
- cover meaningful failure, recovery, and trust behavior;
- ensure planning will not need to invent product behavior.

Fix clear defects directly.

Invoke `document-review mode:headless <requirements-path>` only when this
self-check leaves one material uncertainty that could change product intent,
flow, scope, feasibility, architecture, trust, or decomposition. It selects
exactly one best-fit reviewer; never stack personas or review per slice.
When no review is needed, record a specific not-required reason and do not
dispatch placeholder personas. Reuse the receipt only while the source SHA-256
still matches.

Route review findings through `kb-gate`. Resolve P0/P1 before planning. P2/P3
do not block by severity alone; fix clear issues and record real residual
constraints.

## Handoff

Complete when the requirements source is internally consistent, question-gate
clean, reviewed when warranted, and ready for decomposition.

If execution was requested, invoke `kb-plan <requirements-path>` and continue.
Otherwise return the requirements path and exact next command.

## Stop Rules

- Do not implement code.
- Do not create horizontal phases or slice IDs.
- Do not run external research as ceremony.
- Do not run multiple document reviewers.
- Do not ask quota questions.

## Lazy References

- `references/requirements-template.md` - requirements artifact structure.
- `references/interview-prompts.md` - optional deep-scope question prompts.
