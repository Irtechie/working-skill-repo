---
name: kb-architecture-deepening
description: Lazy architecture exploration lane. Use when the user asks where a codebase should get deeper, simpler to test, or more modular, not for ordinary cleanup or diff review.
argument-hint: "[codebase area, module, or architecture question]"
---

# KB Architecture Deepening

Find places where the codebase needs deeper modules: small public interfaces
hiding meaningful implementation, clearer locality, and better test surfaces.

This is an exploration lane, not a default workflow. Do not load it for normal
bug fixes, generic cleanup, or review of a specific diff.

## Use When

- The question is "where should this codebase get deeper?"
- A subsystem has shallow wrappers, scattered logic, hard-to-test boundaries, or
  repeated call-site coordination.
- The user asks to compare architecture directions before implementation.
- A plan needs a structural option before slicing.

## Do Not Use For

- AI residue cleanup, dead comments, debug prints, placeholder code, or style
  polish. That is cleanup/deslop territory.
- `kb-review` of a specific diff. The thermo-nuclear reviewer handles structural
  findings against an actual change.
- `kb-map` memory coverage. Map records architecture truth; this lane evaluates
  whether the architecture should change.

## Process

1. Read the relevant local architecture docs and source files named by
   `kb-map`; do not crawl the whole repo by default.
2. Identify shallow surfaces: thin wrappers, repeated orchestration, modules
   whose callers know too much, or tests that must inspect internals.
3. Apply the deletion test: if a proposed boundary does not let callers delete
   coordination, branches, mocks, or setup, it probably is not deeper.
4. Prefer interfaces that become the natural test surface.
5. Return at most three architecture candidates, each with proof and cost.

## Output

```markdown
## Architecture Deepening

### Candidate 1: <name>
- Current friction:
- Deeper interface:
- What this deletes:
- Test surface:
- Risk/cost:
- First slice:

### Not Recommended
- <idea>: <why it fails the deletion test>
```

Escalate to `kb-plan` only when the user chooses an option or the next slice is
obvious and low-risk.
