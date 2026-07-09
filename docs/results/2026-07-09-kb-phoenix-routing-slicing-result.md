# KB Phoenix Routing/Slicing Absorption Result

Date: 2026-07-09

Environment:

- Repo: `E:\working-skill-repo`
- OS: Windows
- Go toolchain: `GOTOOLCHAIN=go1.25.4+auto`
- Scope: deterministic local harness only

## Summary

KB kept its routing and vertical-slicing workflow, then absorbed the useful
ATV-Phoenix self-healing ideas as local proof contracts:

- objective manifest `done_check`;
- per-slice `proof_check` or explicit `no_check_reason`;
- model-route validation;
- route-history loop detection;
- conservative install-drift doctor;
- dishonest-completion rejection fixtures.

Credit for the self-healing proof-spine concepts belongs to
[ATV-Phoenix](https://github.com/All-The-Vibes/ATV-Phoenix). KB does not claim
Phoenix's published metrics.

## Fixture Corpus

`evals/dishonest-completion/fixtures.json` contains 4 deterministic rejection
fixtures:

| Fixture | Expected Rejection |
| --- | --- |
| `vacuous-done-check` | already-green proof without RED-before-GREEN evidence |
| `missing-proof-check` | done slice without required `proof_check` |
| `invalid-model-route` | slice route outside `model_tier_contract.routes` |
| `route-history-oscillation` | repeated route oscillation without progress |

## Commands Run

| Command | Result |
| --- | --- |
| `go run ./cmd/kbcheck dishonest-completion-selftest` | pass, 4/4 fixtures rejected |
| `go run ./cmd/kbcheck manifest-contract --manifest docs\plans\2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md` | pass |
| `go run ./cmd/kbcheck run-state-selftest` | pass |
| `go run ./cmd/kbcheck doctor-selftest` | pass |
| `go test ./cmd/kbcheck -count=1 -timeout=120s` | pass |
| `go run ./cmd/kbcheck skill-lint` | pass, 0 errors, warnings only |
| `go run ./cmd/kbcheck skill-sync-report` | pass, 234 comparisons, 0 required issues |
| `go run ./cmd/kbcheck doctor --fix` | pass, required issues 0 |
| `go run ./cmd/kbcheck local-release --json` | pass, required failures 0, optional failures 0 |
| `git diff --check` | pass |
| `git -C E:\all-the-vibes diff --check` | pass |
| `kb-regression-snapshot.ps1 capture -SliceId slice-005` | pass |
| `kb-regression-snapshot.ps1 verify` | pass, 5/5 snapshots |

## Release Gate Notes

`local-release` passed after syncing required deployed copies from this repo:

- Codex global skills;
- Copilot global skills;
- shared agents global skills;
- `E:\all-the-vibes\.github\skills`.

ATV scaffold/plugin skill roots still report warning-only optional differences.
That is intentional for this slice because those roots are packaging surfaces,
not required runtime installs, and this change did not ship that surface.

## Limits

- This is a deterministic fixture result, not a live-model benchmark.
- The corpus has 4 negative fixtures; it proves known false-completion shapes
  are rejected, not that all future false completion is impossible.
- `manifest-contract` validates the objective/proof/model-route contract; it
  does not yet execute every recorded `proof_check` command itself.
- Route-history guards have selftests and a JSONL validator, but need more real
  run histories over time.
- Phoenix's reported Copilot/SWE-bench gains are Phoenix's results, not KB's.
