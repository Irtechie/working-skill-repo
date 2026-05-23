---
name: kb-research
description: Reusable research lane for KB workflows. Use when external docs, prior art, framework behavior, market/product landscape, or known failure modes could change direction, or when a brainstorm/plan/fix needs research captured for future sessions.
argument-hint: "[research topic or decision]"
---

# KB Research

Research once, save the useful parts, and make future sessions cheaper.

## When To Research

Research when findings can change:

- Product direction or user experience.
- Architecture, protocol, auth, persistence, streaming/chat, tool routing, or data model.
- Test strategy or failure-mode coverage.
- Whether a shortcut is actually safe.
- A repeated bug or framework/platform uncertainty.

Do not research when local code and existing docs already answer the question well.

## First Check Local Memory

Before browsing or broad searching:

1. Read `docs/context/PROJECT.md`.
2. Check `docs/context/research/README.md`.
3. Search existing research notes and `docs/solutions/`.
4. Check whether a note is stale.

If an existing note is still valid, summarize it and avoid repeating research.

## Output Location

Write or update:

```text
docs/context/research/<topic>.md
```

Use lowercase kebab-case.

## Note Template

```markdown
# <Research Topic>

Checked: YYYY-MM-DD
Budget mode: lean|standard|deep

## Question

## Findings

## Sources

## Applies When

## Stale When

## Rejected Approaches

## Impact On Current Project
```

Update `docs/context/research/README.md` with a compact index row.

## Research Discipline

- Prefer primary sources for technical claims.
- Mark assumptions when evidence is weak.
- Capture rejected approaches when they are tempting and likely to be retried.
- Keep the note decision-focused; do not paste articles or huge docs.
- If research changes active work, update `todo.md` or the relevant brainstorm/plan/handoff pointer.
