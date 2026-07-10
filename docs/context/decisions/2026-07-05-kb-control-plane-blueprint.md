---
date: 2026-07-09
status: implemented-narrow
scope: planner-economy
manifest: docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
approved_at: 2026-07-05T13:59:39-04:00
---

# KB Control Plane Blueprint

## Target Flow

```text
custom instructions
  -> command
  -> skill
  -> manifest / goal / run state
  -> context packet
  -> subagent or tool
  -> proof
  -> telemetry
  -> scoped learning
```

The goal is not to replace KB with HumanLayer, Phoenix, or ACP. Existing KB
artifacts remain the state spine. The implemented addition is the smallest
bounded packet and telemetry contract needed for cheaper execution.

## What We Keep

KB remains the source of truth for:

- decomposition into vertical slices;
- repo-local memory and handoff discipline;
- deterministic proof gates through `kbcheck`;
- scoped learning and promotion rules;
- drift-safe skill sync across working, global, and ATV roots;
- model-tier routing by capability, not vendor name.

## What We Add

The spike adds the missing control-plane pieces:

- context packets that workers consume before acting;
- telemetry for predicted tier, actual tier/model, proof, rework, escalation,
  and packet sufficiency;
- provider hygiene that keeps CCE optional and Phoenix disconnected;
- honest separation of base and conditional loaded skill surfaces.

## Surface Ownership

| Surface | Owns | Must Not Own |
|---|---|---|
| Custom instruction | stable repo/host policy, safety rules, source precedence, canonical commands | live task state, long workflow bodies, persona prompts |
| Command | user/host entrypoint and arguments | orchestration or hidden policy |
| Skill | workflow policy, gates, artifacts, escalation, proof contracts | volatile repo facts or runtime session state |
| Manifest / goal / run state | current phase, status, lineage, packet pointer, proof pointer | prose-only completion claims |
| Context packet | bounded worker input, constraints, allowed tools/files, proof target, escalation triggers | broad repo rediscovery authority |
| Agent | reusable specialist capability and evidence rules | workflow routing or durable state |
| Subagent | one runtime invocation with one packet and one result | independent planning authority |
| Tool | deterministic side effect or compact query result | hidden reasoning or policy |
| Adapter | host/runtime mechanics for Codex, Claude, GHCP, LiteLLM, local models, or future runners | core planning semantics |

## State Boundary

Do not create another task-state schema. Use manifests for slice state, goal
ledgers for durable objectives, `.kb/runs` for ephemeral route state, and proof
traces for objective acceptance.

## Context Packet Minimum

Every non-trivial worker packet should include:

- repo memory files checked;
- files/interfaces already read and why they matter;
- deterministic prefetch outputs such as `rg` inventories or schema summaries;
- constraints and out-of-scope boundaries;
- allowed files/tools or broad-search policy;
- acceptance/proof target;
- predicted `model_tier` and reason;
- escalation triggers.

## Recovery Rules

Do not claim self-healing until recovery is deterministic:

- invalid transitions fail with specific repair hints;
- stale `running`, `waiting_human`, or `interrupted` states are detectable;
- resume/fork rules are fixtures, not prose;
- human input is persisted as state;
- proof can be rerun from task state without reading chat history.

## Model Economy

Use the expensive model where ambiguity is highest:

- `large`: decomposition, design, architecture/security, failed-loop diagnosis,
  final synthesis;
- `medium`: ordinary vertical slices with complete packets;
- `small`: narrow mechanical edits and simple tests with clear packet context;
- `tiny`: inventories, schema/frontmatter fill, summaries, and status updates.

Proof does not get weaker for cheaper models. The worker can be cheap because
the packet is good, not because correctness matters less.

## Absorption Pass Criteria

Keep KB as runtime core if the spike proves:

- existing state validates, resumes, and repairs through `kbcheck`;
- packets compose with `kb-plan` and `kb-work`;
- recovery does not depend on model judgment;
- adapter details stay outside slice plans;
- telemetry can calibrate model-tier routing.

Move KB onto a smaller runtime/state engine if:

- markdown edits become the state database;
- recovery requires model interpretation;
- adapter assumptions leak into planning;
- packet execution cannot be externally measured;
- adding a second adapter would require redesign.

## First Implementation Boundary

Build only enough to prove the architecture:

- no daemon;
- no UI;
- no Kubernetes/CRD runtime;
- no five-adapter matrix;
- no global state migration;
- no sync propagation until the spike is accepted.

Implemented through
`docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md`.
