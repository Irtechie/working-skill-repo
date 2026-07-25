---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-001
title: "Return a provider-neutral evidence and impact packet"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Defines the stable cross-skill boundary before any extractor/provider can dictate durable semantics."
model_requirements: ["cross-skill architecture judgment", "Go schema and CLI tests", "file-native fallback", "source provenance and freshness design"]
escalation_triggers: ["mandatory provider or daemon becomes necessary", "repo identity or freshness cannot fail closed", "backward compatibility requires weakening proof"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-001.json
proof_check:
  kind: command_exit
  command: "go test ./internal/graphrouting ./cmd/kbcheck -run 'GraphRoute|ImpactPacket'"
  expect: 0
hitl: false
expected_files:
  - path: internal/graphrouting/contract.go
    op: create
    scope: "Provider-neutral symbols, typed edges, provenance, freshness, uncertainty, and impact packet types."
  - path: internal/graphrouting/contract_test.go
    op: create
    scope: "Accepted and rejected contract fixtures."
  - path: cmd/kbcheck/graph_route.go
    op: create
    scope: "Validate and print compact impact packets through a functional CLI."
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Register graph-route validation and packet commands."
  - path: config/graph-route.schema.json
    op: create
    scope: "Machine-readable provider-neutral route/packet schema."
  - path: evals/graph-routing/impact-packet-valid.json
    op: create
    scope: "Protected accepted packet fixture."
  - path: evals/graph-routing/impact-packet-stale.json
    op: create
    scope: "Rejected stale or mismatched revision fixture."
  - path: .github/skills/kb-map/SKILL.md
    op: edit
    scope: "Return contract and file-native fallback semantics."
  - path: .github/skills/kb-map/references/graph-routing.md
    op: edit
    scope: "Provider-neutral packet fields and evidence precedence."
protected_oracles:
  - path: evals/graph-routing/impact-packet-valid.json
    role: "provider-neutral accepted packet"
    sha256: "40CCC36E0F19494AA08DACC372140E81FDFC086DA051F4378D367C55836FBE66"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Protect the contract fixtures, prove missing/freshness fields fail, then implement the smallest file-native packet path."
human_action: ""
can_continue_other_slices: false
---

# Slice 001: Provider-Neutral Impact Packet

## What To Build

Create a compact impact-packet contract that `kb-map` can produce without any
optional provider. It must carry seed symbols/files, typed edges, direct and
reverse impact, tests/docs, source spans, provenance, confidence, revision,
dirty/worktree fingerprint, limitations, and the exact fallback taken.

## Why This Slice Exists

Without this boundary, Graphify, SCIP, CCE, or another adapter would define KB's
architecture by accident. A stable packet makes providers replaceable and
gives later workflow phases one falsifiable input.

## Acceptance Criteria

- A file-native packet validates without Graphify, SCIP, CCE, MCP, or a daemon.
- Exact, observed, structural, heuristic, and LLM-inferred evidence are distinct.
- Missing repo identity, revision, source location for load-bearing exact edges,
  or freshness state fails validation.
- Stale/unsupported providers downgrade to a named fallback; they never silently
  produce an authoritative packet.
- Output is budgeted and cites source spans rather than dumping the graph.
- Existing `PROJECT.md` child-doc routing still works when no graph route exists.
- Pre-edit drift for `kb-map` is compared across working, Codex, Copilot, and
  shared-agent copies; useful global-only changes are merged first.

## Test Scenarios

- Valid file-native packet exits 0 and prints a compact summary.
- Stale revision, missing provenance, unknown confidence class, and path escape fail.
- Optional-provider-unavailable fixture succeeds only through explicit fallback.
- Dirty worktree fingerprint differs from clean HEAD-only identity.

## Proof Check

`go test ./internal/graphrouting ./cmd/kbcheck -run 'GraphRoute|ImpactPacket'`

## Scope Boundary

No SCIP/Graphify/CPG adapter, worktree creation, vector store, daemon, or global
sync in this slice.

## Dependencies

None. This is the smallest enabling slice consumed immediately by slices 004,
005, and 006.

## Concurrency

Serial-only until atomic slice claims and worktree isolation are implemented.
