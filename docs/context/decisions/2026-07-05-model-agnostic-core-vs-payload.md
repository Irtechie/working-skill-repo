---
date: 2026-07-05
status: approved-for-spike
scope: planner-economy
---

# Model-Agnostic Core vs Payload

Fable's critique is accepted: KB should not declare victory over HumanLayer-style
runtime machinery before proving it can absorb the missing pieces.

## Provisional Decision

Use KB as the planning, proof, skill, sync, and learning payload. Run a bounded
absorption spike to decide whether KB should also remain the durable runtime
core.

The user approved the absorption spike scope on 2026-07-05T13:59:39-04:00. The
runtime-core decision remains pending until the spike evidence exists.

Detailed blueprint: `docs/context/decisions/2026-07-05-kb-control-plane-blueprint.md`.

## Corrected Claims

- HumanLayer/CodeLayer already has durable sessions, approvals as state, HITL
  persistence, parent session lineage, status transitions, token/cost fields,
  and event history.
- The public `humanlayer/humanlayer` repo is not a clean fork target; its README
  points to a rebuilt product and calls much of the public code deprecated.
- The reported stuck-state issue is evidence that durable state needs tested
  recovery invariants, not that state machines are automatically safer.
- KB's plausible advantage is packetized decomposition, repo-local memory,
  deterministic proof, skill sync discipline, custom-instruction hygiene, and
  scoped learning.

## Decision Test

Keep KB as runtime core only if the spike proves:

- structured task state validates and recovers deterministically;
- context packets compose with `kb-plan`, `kb-work`, and `kbcheck`;
- model-tier telemetry can calibrate what tiny/small/medium/large workers can
  actually handle;
- one daily-runtime adapter can be added without leaking runtime assumptions into
  the core contract.

If those fail, preserve KB as payload and design or adopt a small runtime/state
engine under it.
