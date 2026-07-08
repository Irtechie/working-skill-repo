---
date: 2026-07-05
status: approved-for-spike
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
  -> task state
  -> context packet
  -> subagent or tool
  -> proof
  -> telemetry
  -> scoped learning
```

The goal is not to replace KB with HumanLayer, Phoenix, or ACP. The goal is to
keep KB's planning/proof/learning strengths and add the smallest durable state
spine needed for cheaper model execution and deterministic recovery.

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

- structured task state;
- context packets that workers consume before acting;
- parent/child lineage for delegated work;
- explicit blocked/waiting/interrupted/failed/done states;
- deterministic recovery hints for stale or invalid states;
- telemetry for predicted tier, actual tier/model, proof, rework, escalation,
  and packet sufficiency;
- one daily-runtime adapter boundary before attempting a second adapter.

## Surface Ownership

| Surface | Owns | Must Not Own |
|---|---|---|
| Custom instruction | stable repo/host policy, safety rules, source precedence, canonical commands | live task state, long workflow bodies, persona prompts |
| Command | user/host entrypoint and arguments | orchestration or hidden policy |
| Skill | workflow policy, gates, artifacts, escalation, proof contracts | volatile repo facts or runtime session state |
| Task state | current task phase, status, lineage, packet pointer, proof pointer, telemetry | prose-only planning or model reasoning |
| Context packet | bounded worker input, constraints, allowed tools/files, proof target, escalation triggers | broad repo rediscovery authority |
| Agent | reusable specialist capability and evidence rules | workflow routing or durable state |
| Subagent | one runtime invocation with one packet and one result | independent planning authority |
| Tool | deterministic side effect or compact query result | hidden reasoning or policy |
| Adapter | host/runtime mechanics for Codex, Claude, GHCP, LiteLLM, local models, or future runners | core planning semantics |

## Minimal Task State

The first schema should be boring:

- `id`, `slice_id`, `parent_id`;
- `phase`, `status`;
- `owner`, `created_at`, `updated_at`;
- `context_packet`;
- `predicted_model_tier`;
- `actual_runtime`, `actual_model`, `actual_tier` when available;
- `proof_command`, `proof_result`;
- `rework_count`, `escalation_reason`, `packet_sufficient`;
- `recovery_hint` for invalid, stale, or blocked states.

Statuses must be externally checkable. At minimum:

```text
pending, ready, running, waiting_human, blocked, interrupted, failed, done
```

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

- task state validates, resumes, and repairs through `kbcheck`;
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

Start with slice-002 in
`docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md`.
