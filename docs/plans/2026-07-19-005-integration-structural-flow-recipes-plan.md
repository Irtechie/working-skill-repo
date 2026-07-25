---
kb_id: kb-2026-07-19-evidence-graph-routing
slice_id: slice-005
title: "Add budgeted structural and flow traversal recipes"
blockers: [slice-003]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "Task-specific traversal and flow uncertainty affect architecture decisions and can create high-confidence false impact sets if underspecified."
model_requirements: ["AST/CPG/data-flow reasoning", "budgeted graph traversal", "Graphify/Tree-sitter compatibility", "uncertainty and source-verification design"]
escalation_triggers: ["whole-program flow becomes default", "inferred edges become exact", "dynamic limitations disappear from output", "provider data leaks into durable plan state"]
context_packet_path: docs/plans/2026-07-19-evidence-graph-routing-context/slice-005.json
proof_check:
  kind: command_exit
  command: "go test ./internal/graphrouting ./cmd/kbcheck -run 'TraversalRecipe|FlowBudget|Graphify'"
  expect: 0
hitl: false
expected_files:
  - path: internal/graphrouting/recipe.go
    op: create
    scope: "Intent-specific edge order, direction, depth, budget, stop, and fallback rules."
  - path: internal/graphrouting/recipe_test.go
    op: create
    scope: "API, bug, deletion, security, and UI traversal fixtures."
  - path: internal/graphrouting/graphify.go
    op: create
    scope: "Optional Graphify structural adapter with provenance and revision checks."
  - path: internal/graphrouting/graphify_test.go
    op: create
    scope: "Typed-edge, multigraph, stale, missing, and inferred-edge fixtures."
  - path: evals/graph-routing/traversal-recipes.json
    op: create
    scope: "Protected recipe and budget expectations."
  - path: .github/skills/kb-map/SKILL.md
    op: edit
    scope: "Choose a recipe from task intent and return a bounded packet."
  - path: .github/skills/kb-map/references/graph-routing.md
    op: edit
    scope: "Graphify/Tree-sitter/flow adapter boundaries and source-verification fallbacks."
  - path: cmd/kbcheck/graph_route.go
    op: edit
    scope: "Validate recipe selection, budgets, and provider limitations."
protected_oracles:
  - path: evals/graph-routing/traversal-recipes.json
    role: "intent-specific traversal and budget oracle"
    sha256: "02F0DB4ED6188F1E8C8E048D258E3BD4C089B6436AEB47554BFEE8988B325C6A"
    update_policy: "requires explicit plan amendment"
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Proceed to slice 006 KB lifecycle integration."
human_action: ""
can_continue_other_slices: true
---

# Slice 005: Structural And Flow Recipes

## What To Build

Add task-shaped recipes for API changes, bugs, deletion, security flow, and UI
behavior. Each recipe names seed types, typed edges, directions, depth/token
budgets, stop conditions, conflict domains, limitations, and fallback proof.
Support existing Graphify output as an optional structural provider; allow
Tree-sitter/CPG/CodeQL-style evidence through the same contract without making
any one runtime mandatory.

## Why This Slice Exists

A generic BFS/vector query cannot know which relationships matter. AST/call
edges also miss data/control dependencies and dynamic registrations. Bounded
recipes gain useful semantics without paying for global flow on every lookup.

## Acceptance Criteria

- Each required intent selects a deterministic recipe or explicit file-native fallback.
- Local flow is preferred before expensive global flow; source/sink queries are bounded.
- `CALLS_STATIC`, `CALLS_OBSERVED`, `REFERENCES`, `IMPLEMENTS`, `OVERRIDES`,
  `READS_CONFIG`, `GENERATES`, `BUILDS`, `TESTS`, `DOCUMENTS`, and inferred
  edges remain distinguishable.
- LLM/community labels may rank candidates but never mint exact edges.
- Reflection, DI, generated code, aliases, dynamic dispatch, and missing library
  models remain visible limitations.
- Graphify multigraph collapse risk is diagnosed or downgraded before impact claims.
- Query output respects edge/depth/token budgets and cites verification spans.

## Test Scenarios

- API recipe finds implementation, serializer, consumer, test, and doc edges.
- Deletion recipe catches reverse refs and config registration.
- Security recipe follows a bounded source/guard/sink path.
- Dynamic-only fixture reports uncertainty instead of a false complete set.
- Over-budget traversal truncates with continuation/fallback metadata.

## Proof Check

`go test ./internal/graphrouting ./cmd/kbcheck -run 'TraversalRecipe|FlowBudget|Graphify'`

## Scope Boundary

No global graph daemon, mandatory CPG/CodeQL install, semantic community naming
in required checks, or claim that static analysis proves all runtime behavior.

## Dependencies

Slice 003 provides safe parallel execution. Slice 001 provides the output
contract; its behavior is inherited through the dependency chain.

## Concurrency

May run with slice 004 in a separate claimed worktree. Both may edit the graph
contract package, so path-prefix claims decide whether to narrow or serialize.

Execution note: completed under one coordinator in the canonical checkout after
slice 004 was complete and its lease was released. No other mutating slice ran
concurrently.
