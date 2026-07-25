# Evidence Graph Routing and Multi-Session Worktrees

## Purpose

Implement the approved evidence-backed routing plan without making any vector,
graph, AST, or MCP provider mandatory. Add safe local multi-session execution so
independent slices may use Git worktrees without duplicate claims or unsafe
integration.

## Suggested Route

`kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md`

## Suggested Skills

- `kb-map` — recover repo-local context and confirm this handoff and manifest are
  still current before execution.
- `kb-work` — execute the manifest dependency DAG and enforce its temporary
  single-coordinator rule.
- `kb-check` — run deterministic slice proof and release gates.
- `kb-functional-test` — preserve the planned critical concurrency and routing
  test classifications when implementing fixtures.

## Current State

- Repository: `E:/Dev/Tools/working-skill-repo`
- Branch: `main`
- Worktree already contains unrelated user changes; preserve them.
- Requirements, manifest, eight slice plans, and eight context packets are
  written but uncommitted.
- Slices 001 through 008 are implemented and verified.
- `todo.md` no longer records this initiative as active; `todo-done.md` archives
  the completion summary.
- Repo-local implementation for slices 001 through 008 has occurred. Required
  global skill roots hash-match the repo source. No commit or push has occurred.

## Files to Read First

1. `docs/brainstorms/2026-07-19-evidence-graph-routing-requirements.md`
2. `docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md`
3. `docs/plans/2026-07-19-001-tool-impact-packet-contract-plan.md`
4. `docs/plans/2026-07-19-002-tool-atomic-slice-lease-plan.md`
5. `docs/plans/2026-07-19-003-tool-worktree-isolation-plan.md`
6. `todo.md`

The remaining slice plans are `004` through `008` beside the manifest. Context
packets are in `docs/plans/2026-07-19-evidence-graph-routing-context/`.

## Decisions Already Made

- Ordering is the default; worktrees are an execution optimization for proven
  independent mutating slices.
- `kb-plan` records durable isolation intent, conflict domains, shared
  resources, and integration dependencies. It must not record live worktree
  paths, branch names, or session IDs.
- `kb-work` owns atomic claims, runtime workspace choice, worktree receipts,
  serialized integration, post-integration proof, and conservative cleanup.
- One coordinator owns canonical `todo.md`, manifest status, handoffs, and
  integration order. Workers return commit/diff/proof receipts.
- Local sessions and sibling worktrees coordinate through the Git common
  directory. Separate clones or machines are outside the local lease contract.
- Never silently stash, force-reset, force-delete, or clean an unintegrated or
  dirty worktree.
- Provider adapters remain optional; file-native `rg`, `kb-map`, and `kbcheck`
  paths must continue to work.

## Work Completed

- Audited current `kb-work` ready-set, board claim, scope-lease, and worktree
  behavior.
- Confirmed the current board claim is non-atomic and the scope-lease command
  validates a supplied ledger rather than acquiring ownership.
- Planned eight vertical slices:
  1. provider-neutral impact packet;
  2. atomic local slice lease;
  3. worktree isolation and serialized integration;
  4. exact-symbol adapter;
  5. structural-flow recipes;
  6. KB skill lifecycle integration;
  7. routing and concurrency evals;
  8. documentation, global sync, and release proof.
- Dependency order: `001 -> 002 -> 003 -> {004,005} -> 006 -> 007 -> 008`.

## Execution Safety

Run slices `001` through `003` with one coordinator and one mutating session.
Do not launch this manifest concurrently before slices `002` and `003` are
integrated and proved. Slices `004` and `005` are the first planned parallel
pair using the new lease/worktree path.

## Verification State

- All eight context packet JSON files parsed successfully.
- Manual plan structure check passed: `slices=8 plans=8 packets=8 blockers=8`.
- Plan-to-work gate ledger check passed with the manifest as the allowed next
  action.
- New plan files passed the targeted trailing-whitespace scan; `todo.md` passed
  `git diff --check`.
- 2026-07-19 resume preflight: `python .github/skills/kb-gate/scripts/check_gate_ledger.py docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md --gate plan-to-work --allowed-next "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"` passed.
- 2026-07-19 follow-up diagnosis: the suspected Go hang was a slow cold
  Go 1.25 package/tool cache. A tiny `%TEMP%\codex-go-smoke2` module completed
  `go test` successfully after about 128 seconds.
- `go run ./cmd/kbcheck manifest-contract --manifest docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md` now passes after repairing the under-proved `brainstorm-to-plan` gate proof count.
- `go run ./cmd/kbcheck context-packet --packet docs/plans/2026-07-19-evidence-graph-routing-context/slice-001.json` passes.
- Slice 001 proof passed: `go test ./internal/graphrouting ./cmd/kbcheck -run 'GraphRoute|ImpactPacket' -count=1 -v`.
- Slice 001 CLI proof passed for the valid packet and failed closed for the stale packet.
- Slice 002 proof passed: `go test ./cmd/kbcheck -run 'SliceLease|ScopeLease' -count=1 -v`.
- Slice 002 CLI selftest passed: `go run ./cmd/kbcheck slice-lease-selftest`.
- Slice 002 protected oracle hash: `cmd/kbcheck/slice_lease_test.go` SHA256 `9371669B79F75CF9F43D5A5CB0949933EEB25458E9B63BCC5E6939154B64C8A1`.
- Slice 003 pre-slice snapshot verify passed from canonical root `E:/Dev/Tools/working-skill-repo`: `snapshot-verify: PASS 15/15 snapshots`.
- Slice 003 proof passed: `go test ./cmd/kbcheck -run 'Worktree|IntegrationOwner' -count=1 -v`.
- Slice 003 manifest contract passed after adding `workspace_isolation_contract`.
- Slice 003 snapshot capture passed: `.kb/snapshots/evidence-graph-routing-slice-003.json`.
- Slice 003 protected oracle hash: `cmd/kbcheck/worktree_isolation_test.go` SHA256 `F52A314AC9E21D6EEAB78728189C7953FFB6E89FA79351B4E798F8DF30542D30`.
- Slice 004 proof passed: `go test ./internal/graphrouting -run 'ExactSymbol|SCIP|Fallback' -count=1 -v`.
- Slice 004 protected oracle hash: `evals/graph-routing/exact-symbol-index.json` SHA256 `8D3296B66A3E045803E0B0C337611FAF88EADFB351F374DD81A794B8CEA2729D`.
- Slice 004 used shared-serial fallback because the slice 001-003 baseline is
  uncommitted in the canonical checkout and would not be present in a fresh
  worktree from `HEAD`; no other mutating slice ran concurrently.
- Slice 005 proof passed: `go test ./internal/graphrouting ./cmd/kbcheck -run 'TraversalRecipe|FlowBudget|Graphify' -count=1 -v`.
- Slice 005 protected oracle hash: `evals/graph-routing/traversal-recipes.json` SHA256 `02F0DB4ED6188F1E8C8E048D258E3BD4C089B6436AEB47554BFEE8988B325C6A`.
- Slice 006 proof passed: `go run ./cmd/kbcheck graph-routing-lifecycle-selftest`.
- Slice 006 focused tests passed: `go test ./cmd/kbcheck -run 'ManifestContract|GraphRoutingLifecycle' -count=1 -v`.
- Slice 006 protected oracle hash: `evals/skill-eval/selftest/pass-evidence-graph-routing.json` SHA256 `80B306772B62E1035BD39DB36D4239A498E37198830F39AAFCAE63A872ACF786`.
- Slice 007 proof passed: `go run ./cmd/kbcheck graph-routing-eval --require-ready`.
- Slice 007 focused tests passed: `go test ./cmd/kbcheck -run 'GraphRoutingEval' -count=1 -v`.
- Slice 007 protected oracle hash: `evals/graph-routing/expected-results.json` SHA256 `0A01D0543B363E41EF801A2E823AC908286A6F674A10EDF52DE426F719519CBF`.
- Slice 008 proof passed: `go run ./cmd/kbcheck skill-sync-report --root E:/Dev/Tools/working-skill-repo --config config/skill-quality.json`.
- Slice 008 proof passed: `go run ./cmd/kbcheck core`.
- Slice 008 proof passed: `go run ./cmd/kbcheck local-release`.

## Blockers / HITL

No human decision is currently required. The earlier Go preflight blocker is
resolved. Continue under the one-coordinator rule.

## Next Action

Resume with:

```text
kb-start docs/handoffs/active/2026-07-19-evidence-graph-routing.md
```

Rerun:

```powershell
python .github/skills/kb-gate/scripts/check_gate_ledger.py docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md --gate plan-to-work --allowed-next "kb-work docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md"
go run ./cmd/kbcheck manifest-contract --manifest docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md
go run ./cmd/kbcheck context-packet --packet docs/plans/2026-07-19-evidence-graph-routing-context/slice-004.json
go run ./cmd/kbcheck context-packet --packet docs/plans/2026-07-19-evidence-graph-routing-context/slice-005.json
```

All slices are complete. If this handoff is read later, use
`docs/plans/2026-07-19-000-kb-evidence-graph-routing-manifest.md` and
`todo-done.md` as the durable completion record.
