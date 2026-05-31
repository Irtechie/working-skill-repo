# Architecture Deepening Lane Requirements

Status: draft
Created: 2026-05-31
Epic: `docs/context/epics/skill-minimalism-and-proof-harness.md`
Research: `docs/context/research/2026-05-31-architecture-deepening-vs-deslop-thermo.md`

## Problem

The bundle has strict diff review through thermo-nuclear code quality and can
use cleanup-oriented deslop tools, but it does not have a dedicated lane for
codebase-wide architecture improvement. Architecture exploration should not be
forced into review comments or generic cleanup.

## Decision

Add architecture deepening as a lazy architecture-audit lane, not a base-layer
skill. It should be invoked only when the user asks where a subsystem/codebase
should get deeper, more testable, or easier for agents to navigate.

Preferred direction: create a compact lazy `kb-architecture-deepening` skill
only if route fixtures prove it is distinct from cleanup and review. Do not put
this inside `kb-memory-review`; architecture exploration and memory maintenance
have different jobs, and combining them would make both fuzzier.

## Requirements

- Use the vocabulary from Matt Pocock's architecture-deepening skill:
  module, interface, implementation, depth, seam, adapter, leverage, locality.
- Preserve the deletion test: if deleting a module makes complexity vanish, it
  was pass-through bloat; if complexity reappears across callers, it earned its
  place.
- Treat the interface as the test surface. If tests must reach past the
  interface, the module may be the wrong shape.
- Read repo-local architecture and decision docs before proposing candidates.
- Present deepening candidates first. Do not design interfaces before the user
  picks a candidate.
- Capture rejected candidates only when the reason would prevent future
  re-suggestion.
- Keep deslop separate as hygiene cleanup.
- Keep thermo-nuclear separate as diff review.

## Candidate Integration Points

| Option | Pros | Risks |
|---|---|---|
| New `kb-architecture-deepening` skill | Clear lazy lane; easiest to route/eval | Adds skill surface |
| Extend `kb-memory-review` | Uses existing memory/audit path | Could blur memory maintenance with architecture design |
| Extend `kb-review`/thermo-nuclear | Reuses reviewer | Wrong scope; review gate becomes exploratory |

Preferred starting point: new lazy skill gated by route fixtures, with
thermo-nuclear borrowing only the deletion-test language.

## Resolve Before Planning

- Confirm whether a compact new lazy skill is acceptable once route fixtures
  prove it routes separately from deslop and thermo-nuclear review.
- Decide whether to actually vendor/adapt Matt Pocock's vocabulary files or
  write a compact original reference based on the same concepts.
- Decide whether route fixtures should include architecture-deepening prompts
  before the skill exists.

## Slice Candidates

- Add architecture-deepening route fixture coverage.
- Add deletion-test/interface-test-surface language to thermo-nuclear review.
- Create a lazy architecture-deepening skill or reference note.
- Add eval fixture proving deslop, thermo-nuclear, and architecture-deepening
  route to different lanes.
