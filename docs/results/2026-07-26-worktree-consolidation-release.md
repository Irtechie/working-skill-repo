# Worktree Consolidation Release

Status: verified candidate tree; direct `main` delivery authorized by the user.

## Reconciliation

Eight registered worktrees were inspected before commit or cleanup.

| Worktree | Classification | Evidence |
|---|---|---|
| `working-skill-repo` | authoritative active tree | `main`; contains the combined source, docs, tests, and skill changes |
| `working-skill-repo-ddr-containment` | clean and superseded | zero commits not already in `main` |
| `working-skill-repo-plan-run-isolation` | superseded implementation branch | 31 of 48 branch-touched blobs match the active tree exactly; the remaining 17 have newer active-tree versions; plan-worktree and delivery-boundary tests pass |
| `wsr-head-proof` | clean detached proof checkout | no dirty state or unique commit |
| `wsr-graph-proof-20260725` | dirty but fully contained donor | 78 changed paths: 47 match the active tree, 31 exact blobs occur in `main` history, zero unseen blobs |
| `wsr-index-proof-20260725` | dirty but fully contained donor | 69 changed paths: 48 match the active tree, 21 exact blobs occur in `main` history, zero unseen blobs |
| `wsr-orchestrator-ddr` | clean and merged | zero commits not already in `main` |
| `working-skill-repo-routing-owner` | superseded design branch | replaced by the later orchestrator-directed DDR release on `main`; current routing/model tests pass |

No donor content was copied blindly and no unrelated worktree was reset.
Worktrees may be removed only after the consolidated commit is published and
remote `main` containment is verified.

## Requested README State

- DDR remains documented as the normal orchestrator-owned routing model.
- AMR is absent from `README.md`.
- AYGHRI's `i-have-adhd` and Plannotator's `bro` skill remain credited.

## Proof

- focused plan-worktree, delivery-boundary, PR-workbench, communication, router,
  and model-routing tests passed;
- `go run ./cmd/kbcheck core` passed 38 checks;
- `go run ./cmd/kbcheck local-release` passed all required components;
- `skill-sync-report` passed 138 comparisons with zero required issues;
- `git diff --check` passed.
