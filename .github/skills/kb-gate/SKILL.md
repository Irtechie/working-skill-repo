---
name: kb-gate
description: Shared phase-gate policy for KB workflows. Use before moving from brainstorm to plan, plan to work, work to complete, or complete to ship when P0/P1/P2/P3 findings, review issues, ambiguity, weak tests, or unresolved risks exist.
argument-hint: "[phase, artifact path, or finding list]"
---

# KB Gate

Do not let known issues drift silently into the next phase.

## Severity

| Severity | Meaning | Default |
|---|---|---|
| P0 | Will likely build the wrong thing, break core behavior, create safety/security/data risk, or make the next phase invalid | Block |
| P1 | Important ambiguity, missing verification, serious design/test gap, or likely rework | Block |
| P2 | Non-blocking but fixable quality, clarity, edge-case, or maintainability issue | Offer rectify-all |
| P3 | Minor polish, wording, naming, or cleanup issue | Offer rectify-all when cheap |

## Who Fixes

P0/P1 block the next phase, but they do not automatically require human input.

The agent should rectify P0/P1 without asking when the fix is safe and evidence-backed:

- contradiction or stale wording in a doc;
- missing acceptance criterion derivable from the source material;
- missing verification mode or expected files;
- broken dependency DAG with one clear correction;
- deterministic test/lint/build failure with a local fix;
- review finding with a concrete safe/gated auto-fix.

Ask the human only when resolution requires:

- product intent or priority judgment;
- accepting or rejecting scope;
- credentials, login, external system access, or real-world approval;
- destructive/risky operation;
- choosing between multiple reasonable architecture/product paths;
- changing the user's stated requirements.

## Phase Gates

- **Brainstorm -> plan:** block on unresolved P0/P1 requirements, contradictions, missing core behavior, unsafe assumptions, or missing verification inputs.
- **Plan -> work:** block on broken DAG, missing acceptance criteria, missing verification mode, missing expected files, weak functional coverage, unsafe HITL, or unresolved architecture/security risk.
- **Work -> complete:** block on failing deterministic checks, failed functional flows, scope violations, unresolved durable memory refresh, or blocked slices not explicitly parked.
- **Complete -> ship:** block on unresolved P0/P1 review findings, failed checks, release risk, or unrecorded human-only blockers.

## Rectify Prompt

When findings exist, first classify them:

- `auto_rectify`: agent can fix safely now;
- `needs_human`: requires one of the human-only conditions above;
- `defer_log`: non-blocking and not worth fixing now.

Fix `auto_rectify` items before asking. Then ask only for remaining human/judgment decisions:

```text
I found <count> issues before <next phase>: P0=<n>, P1=<n>, P2=<n>, P3=<n>.
I can rectify <auto_count> safely now. <human_count> need your decision.
Do you want me to rectify all safe/actionable issues before moving on?
```

If the user says yes, fix all safe/actionable issues, rerun the relevant review/check, and then continue.

If the user says no:

- P0/P1 still block unless resolved, reclassified with evidence, or converted into a scoped parked item outside this work.
- P2/P3 may be logged in the artifact, `todo.md`, or `todo-done.md` with owner/status.

## Output

Report severity counts, actions taken, remaining blockers, deferred items, and whether the next phase is allowed.
