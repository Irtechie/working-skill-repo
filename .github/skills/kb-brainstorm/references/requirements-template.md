# KB Brainstorm Requirements Template

Use this when Phase 8 creates or updates a requirements document.

```markdown
---
date: YYYY-MM-DD
topic: <kebab-case-topic>
brainstorm_style: kb-brainstorm
---

# <Topic Title>

## Problem Frame
[Who is affected, what is changing, and why it matters]

## Research Summary

**Findings that shaped requirements:**
- [Finding] - [which requirements/decisions it affected] - [link or note]

**Confidence:** High / Medium / Low - [one-line justification]

## Requirements

**[Group Header]**
- R1. [Concrete requirement]
- R2. [Concrete requirement]

**[Group Header]**
- R3. [Concrete requirement]

## Success Criteria
- [How we will know this solved the right problem]

## Scope Boundaries
- [Deliberate non-goal or exclusion]

## Key Decisions
- [Decision]: [Rationale] - Evidence: [research citation or "assumption"]

## Dependencies / Assumptions
- [Only include if material]

## Alternatives Considered
- [Approach]: [why not chosen] - [research citation]

## Slice Candidates (advisory for /kb-plan)
- [Increment title] - [what user-visible behavior it delivers]
- [Increment title] - [what user-visible behavior it delivers]
<!-- Keep advisory. Do not assign blockers, ordering, or dependencies. /kb-plan owns sequencing. -->

## Outstanding Questions

### Resolve Before Planning
- [Affects R1][User decision] [Question that must be answered before planning can proceed]

### Deferred to Planning
- [Affects R2][Technical] [Question that should be answered during planning]
- [Affects R2][Needs research] [Question that likely requires deeper research during planning]

## Next Steps
[If `Resolve Before Planning` is empty: `-> /kb-plan`]
[If `Resolve Before Planning` is not empty: `-> Resume /kb-brainstorm`]
```

## Visual Aids

Include a visual aid only when it makes the requirements faster to understand:

| Requirements describe... | Visual aid |
|---|---|
| Multi-step user workflow | Mermaid flow or annotated ASCII flow |
| 3+ modes, variants, or states | Markdown comparison table |
| 3+ interacting participants | Mermaid or ASCII relationship diagram |
| Competing approaches | Comparison table |
| Landscape examples | Comparison table |

Skip visuals when prose is clearer, the diagram only restates bullets, or it would describe implementation architecture better left to `kb-plan`.

Use Mermaid for simple flows, ASCII for annotated flows, and tables for comparisons. Prose is authoritative if a visual conflicts with surrounding text.
