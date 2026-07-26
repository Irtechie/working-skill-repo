---
type: kb-brainstorm
created: 2026-07-19
status: ready-for-plan
workflow_shape: pipeline-change
---

# Evidence-Backed Graph Routing And Multi-Session Isolation

## Intent

Improve `kb-map` from a size-gated graph pointer into a provider-neutral,
evidence-backed impact router, then carry its output through planning, work,
review, and verification. Make that workflow safe when several local Codex
sessions operate on the same repository.

## Current Evidence

- `kb-map` currently decides when graph routing may pay, points to one
  `graph_route`, and requires source verification. It does not define typed
  edge semantics, exact-symbol precedence, an impact-packet contract,
  provider freshness, or task-specific traversal recipes.
- File-native `rg`/`kb-map`/`kbcheck` behavior is mandatory. Graphify, SCIP/LSP
  indexes, CPG/data-flow tooling, vectors, MCP, and similar providers
  remain optional adapters and must not become install or runtime requirements.
- `kb-work` correctly limits shared-checkout mutation to one slice and asks
  agents to claim work in `todo.md`, but the claim is a non-atomic read/modify/
  write convention.
- `kbcheck scope-lease` detects collisions in a completed JSON ledger; it does
  not atomically acquire or release a slice claim and cannot distinguish two
  sessions that both claim the same slice ID.
- Worktrees isolate filesystem mutation but do not prevent duplicate claims,
  shared-board divergence, graph-index namespace collisions, or mergeback
  conflicts.

## Decisions And Reasons

### 1. Define a provider-neutral evidence and impact contract first

Reason: choosing a database or extractor first lets that tool's output shape
become accidental architecture. The durable contract must instead define
symbols, typed edges, provenance, confidence, revision/freshness, uncertainty,
source spans, and the compact impact packet consumed by KB.

### 2. Prefer exact symbols before structural heuristics or embeddings

Reason: compiler/language-server indexes can distinguish definitions,
references, implementations, and overrides. Embeddings and structural matches
are useful candidate generators, but cannot safely establish code identity.

### 3. Use task-shaped traversal recipes

Reason: API changes, bug tracing, deletion safety, security flow, and UI changes
need different edge types, directions, depth budgets, and stopping rules. A
generic breadth-first graph dump spends tokens without proving impact.

### 4. Add scoped structural, control-flow, and data-flow expansion

Reason: AST/call edges alone miss values, guards, aliases, registrations, and
runtime behavior. Flow analysis should be bounded to selected sources/sinks or
impact roots because whole-program flow is expensive and may over-approximate.

### 5. Make freshness and uncertainty executable gates

Reason: a precise stale graph is still wrong. Every packet must bind to a
canonical repository identity, revision, dirty/worktree fingerprint, extractor
version, and limitations such as reflection, generated code, or dynamic
dispatch. Stale or unsupported evidence must downgrade to source inspection.

### 6. Carry the impact packet through plan, work, and review

Reason: orientation does not improve implementation unless `kb-plan` records
the forecast, `kb-work` reconciles observed writes, and `kb-review` checks for
missed consumers, tests, docs, and unexplained expansion.

### 7. Replace advisory board claiming with atomic local claims

Reason: two sessions can both read `pending` before either writes
`in_progress`. The shared Git common directory provides a local coordination
root across worktrees without a daemon. Atomic acquire/renew/release with an
owner token closes that same-repository TOCTOU race.

### 8. Put worktree intent in `kb-plan`; put worktree mechanics in `kb-work`

Reason: planning can identify conflict domains, shared resources, integration
order, and whether isolation is required, but it cannot safely freeze a live
path, branch, base revision, or current-session claim. `kb-work` has the live
Git/dirty/session evidence needed to create or reuse an isolated worktree,
serialize integration, rerun proof, and clean up safely.

### 9. Measure correctness before promotion

Reason: token reduction alone can reward confidently incomplete routing. The
gate must measure impacted-symbol recall, false positives per retrieved token,
stale-index detection, missed tests/docs, race prevention, dirty-checkout
preservation, integration conflicts, and review-discovered omissions.

## Multi-Session Contract

- Until the atomic-claim and worktree-isolation slices are integrated, only one
  mutating session may execute this manifest.
- Read-only sessions may share a checkout.
- A mutating session must atomically acquire the slice before editing.
- Shared-checkout mutation remains serial.
- Independent mutating slices may run concurrently only in separate worktrees
  and only when declared conflict domains are disjoint.
- One integration owner updates the canonical manifest, `todo.md`, and
  handoffs. Workers return a commit/diff, proof receipt, observed writes, and
  discovered conflicts instead of independently merging lifecycle files.
- Integration is serialized and proof reruns after integration.
- Worktree cleanup is never forced. It occurs only after integration succeeds,
  the worktree is clean, and the matching owner token releases the lease.
- Atomic claims coordinate worktrees sharing one Git common directory. Separate
  clones or machines remain outside this local contract.

## Proposed Planning Fields

These are feature requirements, not fields the current manifest validator can
already enforce:

```yaml
workspace_mode: auto|shared-serial|worktree-required|read-only-parallel
conflict_domains: [files, generated-output, browser, port, database, index]
shared_state_owner: coordinator
integration_after: [slice-id]
```

`kb-work` records live values in a receipt, not the durable plan:

```yaml
claim_id: <opaque-owner-token>
repo_identity: <canonical-git-common-dir-identity>
base_revision: <sha>
worktree_path: <runtime-path>
branch: codex/<kb-id>/<slice-id>-<short-id>
heartbeat_at: <timestamp>
observed_writes: []
integration_status: pending|integrated|conflict|released
```

## Query Recipe Minimums

| Intent | Required traversal |
|---|---|
| API/contract change | definition -> references -> implementations -> serializers -> consumers -> tests/docs |
| Bug | entrypoint -> callees -> relevant data/control flow -> guards/error paths -> reproducing tests |
| Deletion | reverse references/dependencies -> registrations/config strings -> generated/runtime edges |
| Security | bounded source -> sanitizer/guard -> sink path |
| UI behavior | component/route -> state/action -> API/data contract -> rendered functional test |

## Question Gate

- `ask-now`: none.
- `research-first`: resolved by primary-source review of Tree-sitter queries,
  SCIP exact-symbol indexes, stack-graph incremental resolution, CodeQL data
  flow, and code-property graphs, plus the repo's existing swarm and adapter
  research.
- `safe-assumption`: local concurrent sessions use worktrees belonging to the
  same Git common directory. Proof: multi-worktree integration fixture.
- `defer-to-planning`: exact Go package/file boundaries and whether the first
  exact-symbol adapter consumes SCIP protobuf directly or a deterministic
  exported snapshot.
- `parked`: cross-machine/distributed locking, required daemons/MCP servers,
  automatic provider installation, a mandatory vector database, broad global
  flow analysis, and neighboring ATV changes.

## Resolve Before Planning

None.

## References

- `.github/skills/kb-map/references/graph-routing.md`
- `.github/skills/kb-work/SKILL.md`
- `cmd/kbcheck/swarm.go`
- `docs/plans/archive/2026-06/2026-06-01-110-kb-work-swarm-ready-set-manifest.md`
- `docs/context/research/2026-07-05-dexhorthy-humanlayer-agent-harness-research.md`
- https://github.com/scip-code/scip
- https://tree-sitter.github.io/tree-sitter/using-parsers/queries/index.html
- https://github.blog/open-source/introducing-stack-graphs/
- https://codeql.github.com/docs/writing-codeql-queries/about-data-flow-analysis/
- https://fraunhofer-aisec.github.io/cpg/
