# Proof Receipt - kb-rehab slice-001

- kb_id: kb-2026-08-26-kb-rehab-outstanding-work
- slice_id: slice-001
- title: Pair declared work against git reality, read-only
- recorded_at: 2026-08-26T22:35:00Z
- tree: working tree of `codex/w2d-unplanned-entry` at slice-001 acceptance
- route: current orchestrator (DDR exception `context-required`)

## Exact Proof

| Check | Command | Result |
|---|---|---|
| Slice proof | `go test ./cmd/kbcheck -run 'WorkReality' -count=1` | PASS (18 tests) |
| Package regression | `go test ./cmd/kbcheck -count=1` | PASS (228.6s) |
| Vet | `go vet ./...` | PASS |
| Build | `go build ./...` | PASS |
| Format | `gofmt -l cmd/kbcheck/work_reality.go cmd/kbcheck/work_reality_test.go` | clean |
| Whitespace | `git diff --check` | clean |

## Protected Oracle

| Path | Role | SHA-256 |
|---|---|---|
| cmd/kbcheck/work_reality_test.go | Mixed fixture proving every lifecycle classification, every fail-closed downgrade, read-only behavior, determinism, redaction, and the preservation floor | d9592f7cfbab222800ba58b09707af1d8ddd0d241c8aea5ded495d6608c1e439 |

## Acceptance Evidence

Run against this repository, `go run ./cmd/kbcheck work-reality --root . --json`
returned `status=ok`, `remote_authority=authoritative` on `origin/main`, 67
declared items, 9 settled, and 65 pairings with no hand-pairing:

| State | Count | Notes |
|---|---|---|
| `dead` | 3 | `codex/kb-runtime-cognitive-routing`, `deaderestpool-cleanup-reconciler-requirements`, `deaderestpool-fictional-lamp`, each proven patch-equivalent to the authoritative default via `git cherry` |
| `live` | 2 | `codex/w2d-unplanned-entry` (current session) and `codex/this-plan-needs-receipts` (peer session holding a non-terminal claim) |
| `orphan-branch` | 3 | uncontained commits with no declared work item; two touch protected roots (`.github/skills`, `cmd`) and are therefore auto-merge-ineligible under R14b |
| `orphan-work` | 57 | declared work with no paired ref |

Every acceptance criterion in the slice plan is met:

- **Read-only.** `TestWorkRealityIsReadOnly` asserts refs, porcelain status,
  worktree list, and HEAD are byte-identical before and after a run.
- **Evidence-bound classification.** Every pairing carries `contained`, the
  predicate names consulted, and the evidence source for each.
- **Adapter absence removes conclusions.** Absent remote, unreachable remote,
  unresolvable advertised default, and absent predicate manifest each yield
  `status=fail-closed` with zero `dead` and zero `superseded`.
- **No terminal state without containment proof.** Containment is proven by
  `git cherry` patch equivalence against a freshly fetched authoritative
  default resolved through `ls-remote --symref`, never a cached tracking ref.
- **Preservation floor.** `TestWorkRealityPreservationPredicatesSupersetOf
  TerminalCleanup` asserts the predicate set strictly exceeds
  `terminalCleanupSafetyPredicates()`.

## Scope Comparison

Forecast four files; actual four files, all forecast, none discovered, none unused:

| Path | Op | Forecast |
|---|---|---|
| cmd/kbcheck/work_reality.go | create | yes |
| cmd/kbcheck/work_reality_test.go | create | yes |
| cmd/kbcheck/main.go | modify | yes |
| config/rehab-policy.json | create | yes |

## In-Slice Correction

The first run against this repository classified nine manifests carrying a
terminal declared status (`complete`, `completed`, `done`, `superseded`) as
`orphan-work`. Finished work is not orphaned, and the false positive would have
pushed a human to re-review plans that already landed. Corrected by reporting
such items in a `settled` list, excluded from outstanding pairings and never
silently dropped. Covered by
`TestWorkRealitySettledManifestIsNotOrphanWork`.

## Memory Impact

Durable. `kbcheck work-reality` is a new read-only surface consumed by
slice-002 and slice-003; `config/rehab-policy.json` is the versioned predicate
manifest (`rehab-1.0.0`). Context refresh is deferred to slice-002, which
introduces the user-facing `kb-rehab` lane.

## Implementation Note

`internal/reconcile`'s git helpers did not need exporting. `reconcile.Inventory`
already returns every artifact field the pairing needs, and remote authority was
satisfiable through plain `git` plumbing without duplicating reconciler logic.
The escalation trigger flagged at plan time did not fire.
