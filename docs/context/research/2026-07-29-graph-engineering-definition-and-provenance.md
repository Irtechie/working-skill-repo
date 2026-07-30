# Graph Engineering Definition And Provenance

Checked: 2026-07-29
Budget mode: deep

## Question

What does the July 2026 "graph engineering" conversation mean, how much of it
is new, how does KB compare, and can code-owned tracing help an LLM diagnose
failures or determine completion?

## Findings

### The term is new; the mechanics are not

"Graph engineering" crystallized on X between July 18 and July 20, 2026.
Peter Steinberger's seed question is directly recoverable:

> Are we still talking loops or did we shift to graphs yet?

The post is dated July 18 at 00:34 UTC, not July 17 UTC. Within 48 hours the
discussion included state-machine criticism, "org chart" metaphors, commercial
products, definitions, and Harrison Chase asking whether the term was
"basically just langgraph."

The strongest technically grounded response came from XState creator David
Khourshid: loops and graphs were being presented as new marketing terms even
though they are state machines and actor-model ideas. LangGraph had already
provided graph-shaped agent orchestration in 2024. Scientific workflows,
dataflow systems, state machines, Petri nets, actor systems, and workflow
provenance predate LLM agents by decades.

The useful new constraint is not graph topology. It is that graph nodes may now
be stochastic, non-idempotent, context-consuming LLM executions whose failures
often look like plausible prose rather than exceptions.

### There is no settled external definition

The public discussion has an emerging meaning, not a stable specification.
"Work graph + org graph" is not an established two-part definition. The Turing
Post article instead distinguishes control graphs, knowledge graphs, execution
traces, and graphs of improvement loops. Its statement that a loop is a graph
whose path returns to an earlier node is technically sound.

"Org graph" with persistent identities, lateral messaging, shared mutable
state, and dynamic agent spawning is a later synthesis of the "org chart"
metaphor and actor-model properties. Those capabilities are optional
multi-agent architecture choices, not requirements for graph engineering.

### KB operational definition

Use this definition inside KB:

> Graph engineering designs and operates an explicit control graph where nodes
> are bounded semantic work units, edges encode admissible control, data, and
> dependency transitions, the runtime owns execution semantics, and every node
> emits verifiable provenance.

This definition keeps five responsibilities distinct:

| Responsibility | Code owns | LLM owns |
|---|---|---|
| Topology | Nodes, edges, dependencies, cycles, barriers | Suggesting decomposition when judgment is needed |
| Scheduling | Ready sets, fan-out, fan-in, concurrency limits | Work inside a dispatched node |
| State | Versions, leases, checkpoints, retry/idempotency keys | Semantic interpretation of bounded state |
| Failure handling | Timeouts, retry policy, quarantine, dependency propagation | Diagnosis and repair proposals |
| Completion | Required-node accounting and evidence gates | Producing evidence and evaluating subjective criteria |

Loops remain valid: they are cycles inside the control graph. Persistent agent
roles, lateral agent messaging, shared mutable state, and dynamic spawning are
not part of the base definition.

### Code should control predictable routing

Google ADK 2.0 documents graph workflows as deterministic processes combining
code logic and AI reasoning. Its workflow agents follow predefined logic
without asking a model to orchestrate the sequence. The common paraphrase
"code controls predictable routing while models handle judgment" accurately
describes the design, but it is not a verbatim Google quote.

Anthropic distinguishes workflows, where predefined code paths orchestrate
models and tools, from agents, where the model directs its own process. The
claim that code-owned routing saves repeated routing tokens is reasonable, but
it is a Turing Post inference rather than a direct Anthropic claim.

### Tracing helps diagnosis, but does not define done

A trace helps an LLM when it is compact, causal, and grounded in runtime
evidence. It should answer:

- Which node first diverged or failed?
- Which parent nodes and state revision fed it?
- Which context, model route, tools, and attempt produced the output?
- Which downstream nodes consumed that output?
- Was the node retried, superseded, quarantined, or accepted?
- Which proof artifact justified acceptance?

Raw transcripts and unbounded logs increase context cost and can obscure the
first failure. The LLM should receive a small failure or completion packet with
pointers to detailed artifacts. Host/runtime events, repository revisions, and
artifact hashes are authoritative; an agent's self-report is not.

Tracing is evidence for completion, not the completion authority. A run is done
only when:

```text
all required nodes are terminal
+ every required acceptance criterion has valid proof
+ required fan-in/review gates passed
+ no unresolved blocking edge remains
```

The runtime should compute that predicate and expose a compact explanation such
as `graph-run inspect --failed` or `graph-run inspect --why-not-done`.

### KB was already graph-shaped

KB had a dependency DAG in `kb-plan` on May 23, 2026, bounded parallel
ready-set execution in `kb-work` on June 1, and review fan-out plus merge/dedup
in May. Provider-neutral source-code graph routing arrived July 25.

The existing execution graph already covers explicit dependencies, safe
parallelism, barriers, leases, gate ledgers, and independent proof. The main
gap is not an "org graph" or a new `kb-graph` lane. It is a unified immutable
node receipt joining data that currently lives across execution telemetry,
slice leases, gate ledgers, and review artifacts.

Minimum useful receipt:

```json
{
  "run_id": "run-42",
  "node_id": "slice-003.attempt-2",
  "slice_id": "slice-003",
  "depends_on": ["slice-001"],
  "session_id": "session-abc",
  "context_packet_hash": "sha256:...",
  "state_version": "abc123",
  "base_sha": "abc123",
  "head_sha": "def456",
  "actual_route": "medium",
  "actual_model": "model-id",
  "started_at": "RFC3339",
  "completed_at": "RFC3339",
  "status": "passed",
  "retry_count": 1,
  "proof_artifact_hash": "sha256:..."
}
```

The receipt should be emitted by code at node termination, linked from the
gate ledger, and validated for dependency order, state/revision freshness,
artifact hashes, and proof presence. Review fan-in should list the reviewer
receipt IDs that contributed to each merged finding.

## Sources

- Peter Steinberger, X, 2026-07-18:
  https://x.com/steipete/status/2078277297791189132
- David Khourshid, X, 2026-07-18:
  https://x.com/DavidKPiano/status/2078452098631454949
- Rohit Garg, X, 2026-07-19:
  https://x.com/rohit4verse/status/2078776077745623498
- Shann Holmberg, X, 2026-07-20:
  https://x.com/shannholmberg/status/2079096565344739643
- Harrison Chase, X, 2026-07-20:
  https://x.com/hwchase17/status/2079219804951683380
- Turing Post, 2026-07-24:
  https://www.turingpost.com/p/is-graph-engineering-real-why-everyone-is-talking-about-it
- Google Agent Development Kit graph workflows:
  https://adk.dev/graphs/
- Google ADK workflow agents:
  https://adk.dev/agents/workflow-agents/
- Anthropic, Building Effective Agents:
  https://www.anthropic.com/engineering/building-effective-agents
- LangGraph overview:
  https://docs.langchain.com/oss/python/langgraph/overview
- LangGraph announcement, 2024:
  https://www.langchain.com/blog/langgraph
- W3C PROV overview:
  https://www.w3.org/TR/prov-overview/
- OpenTelemetry generative AI semantic conventions:
  https://opentelemetry.io/docs/specs/semconv/gen-ai/

## Applies When

- Deciding whether a task needs a single loop, a DAG, or a cyclic state machine.
- Designing `kb-work` scheduling, retries, barriers, or completion reporting.
- Adding node-level execution telemetry, provenance, or debugging commands.
- Evaluating claims that persistent multi-agent organizations are required.

## Stale When

- A recognized standards body or major framework publishes a materially
  different stable definition of graph engineering.
- KB replaces its current manifest, lease, telemetry, or gate-ledger contracts.
- OpenTelemetry finalizes incompatible GenAI agent-span conventions.

## Rejected Approaches

- Add a user-facing `kb-graph` lane: duplicates `kb-plan`, `kb-work`,
  `kb-review`, and `kb-map` ownership.
- Treat an "org graph" as the missing half: the term is not established and its
  actor-model features add distributed-state cost without proving user value.
- Feed raw traces to the model: expensive, noisy, and weaker than a bounded
  causal packet.
- Let the executing model self-certify completion: provenance supports a gate;
  it does not replace one.

## Impact On Current Project

Keep the current lanes. If implementation is authorized, make provenance an
additive contract:

1. Extend the execution telemetry envelope with node identity, attempt,
   dependencies, state/base/head revisions, timestamps, status, and lease
   generation.
2. Emit one immutable receipt per node attempt.
3. Link its path and hash from the slice gate ledger.
4. Add a deterministic validator and compact failed/why-not-done projection.
5. Add reviewer receipt IDs to review merge provenance.
