---
date: 2026-07-09
status: accepted
scope: planner-economy
---

# Model-Agnostic Core vs Payload

Fable's critique is accepted: KB should not declare victory over HumanLayer-style
runtime machinery before proving it can absorb the missing pieces.

## Decision

Keep KB as the planning, proof, skill, sync, learning payload, and lightweight
runtime core. Do not add a second task-state runtime.

Existing manifests, goal ledgers, route history, proof traces, and `todo.md`
already provide durable lifecycle state. The missing execution boundary is a
vendor-neutral context packet plus optional normalized telemetry, not another
database or daemon.

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

## Evidence

- `kbcheck run-state`, manifests, gate ledgers, and the proof spine cover
  recovery and completion without a new state engine.
- `kbcheck context-packet` validates bounded worker inputs without vendor fields.
- Context packets are wired into `kb-plan` and `kb-work`.
- Optional telemetry normalizes runtime/model, turns, input/output/cache tokens,
  proof result, and packet sufficiency while preserving raw values.
- `kbcheck provider-hygiene` rejects Phoenix activation and permits CCE as an
  optional adapter.
- `surface-report` separates base startup from conditional safety/check skills.

## Revisit Trigger

Reconsider a separate runtime only if real runs prove that manifest/run-state
recovery requires model interpretation, packet execution cannot be measured, or
adding a second host adapter leaks vendor assumptions into the core contract.
