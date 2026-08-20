# KB Runtime, Cognitive Routing, and Surface Reduction

Status: active
Created: 2026-08-20
Last refreshed: 2026-08-20

## Intent

Move hot-path orchestration from conversational instructions to deterministic
state and tool contracts while keeping portable, human-readable phase adapters.

## Success Criteria

See `docs/brainstorms/2026-08-20-kb-runtime-cognitive-routing-requirements.md`.

## Architecture Decisions

- Phase ownership stays separate; runtime state unifies transition mechanics.
- Serial execution is the default; parallelism is explicit and evidence-bound.
- Live model catalogs are host-local; plans declare capability contracts only.
  Execution prefers the exact tier, then a qualified higher tier; it never
  automatically routes downward.
- Markdown is a boundary projection from structured state.

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|
| Runtime state and structured tool contract | skipped-clear | `docs/plans/2026-08-20-000-kb-runtime-state-contract-manifest.md` | planned | Foundation for all hot-path extraction |
| DDR selection enforcement | skipped-clear | `docs/plans/2026-08-20-010-kb-routing-cognitive-delivery-manifest.md` | planned | Depends on runtime receipts |
| Cognitive route and delivery UX | skipped-clear | `docs/plans/2026-08-20-010-kb-routing-cognitive-delivery-manifest.md` | planned | Depends on runtime projection contract |
| Serial execution and route classifier | skipped-clear | `docs/plans/2026-08-20-020-kb-surface-retirement-manifest.md` | planned | Runs after delivery contract |
| Skill-surface audit and AMR retirement | skipped-clear | `docs/plans/2026-08-20-020-kb-surface-retirement-manifest.md` | planned | Evidence-first; no skill deletion is pre-authorized |

## Dependency Map

`runtime-state` -> `DDR-enforcement` -> `cognitive-delivery` ->
`serial-classifier` -> `legacy-retirement`

## Execution Queue

Run one serial delivery train. No per-slice or multi-manifest worktree fan-out
is authorized. The train cannot start another delivery candidate until the
current candidate reaches `delivery-integrated` or the user explicitly changes
the delivery policy. The serial-policy workstream may remove the plan-run
worktree default for later work only after its migration proof passes.

## Human Checkpoints

None. The user has specified the product direction and acceptance boundary.

## Parked / Blocked

Embedding-based routing is parked pending evidence that the deterministic
classifier cannot meet route-fixture accuracy and cost requirements.

## Completion Criteria

Every workstream has terminal proof; the active skill and global install
surfaces match; no active AMR route remains; delivery follows the configured
solo/collaborative policy. A skill may be removed only through a separate,
evidence-backed and user-approved decision after the audit.
