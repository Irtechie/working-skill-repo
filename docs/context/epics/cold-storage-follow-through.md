# Cold Storage Follow Through

Status: completed
Created: 2026-06-01
Last refreshed: 2026-06-01

## Intent

Turn the remaining cold-storage items into either runnable plans, explicit
parked work, or blocked decisions without weakening the release-ready global
skill bundle.

## Success Criteria

- Remaining deletion/trim candidates have runtime-usage proof, a focused trim
  plan, or an explicit keep/park decision.
- Any Go work is scoped honestly: wrapper hardening, partial native parity, or
  full non-PowerShell rewrite.
- Cross-model benchmark prompts are planned as evidence-gathering assets, not
  subjective model impressions.
- Path-specific Copilot instructions are added only for file classes where they
  reduce repeated mistakes.
- Every runnable workstream has a manifest; every non-runnable workstream has a
  named blocker or parked rationale.

## Architecture Decisions

- Do not delete skills or agents from static analysis alone.
- Do not claim a full cross-platform harness until checks run without
  PowerShell.
- Keep the safe path aligned with the normal release gate:
  `go run .\cmd\kbcheck local-release`.
- Do not add path-specific instructions for generic advice already present in
  `AGENTS.md`, skills, or README.

## Research

- Local context loaded: `todo.md`, `docs/context/PROJECT.md`,
  `docs/context/operations/testing.md`, and
  `scripts/skill-surface-minimality.ps1` output.
- External research skipped for startup: these are repo-local workflow/tooling
  decisions.

## Workstreams

| Workstream | Brainstorm | Manifest | Status | Notes |
|---|---|---|---|---|
| Minimality runtime proof | skipped-clear | `docs/plans/2026-06-01-090-kb-cold-storage-follow-through-manifest.md` | done | Slice 091 adds evidence classes before deletion decisions. |
| Go native core gate rewrite | skipped-clear | `docs/plans/2026-06-01-100-kb-go-native-core-gate-rewrite-manifest.md` | done | Windows parity passed; top-level PS wrappers removed after proof. |
| Cross-model benchmark prompts | skipped-clear | `docs/plans/2026-06-01-090-kb-cold-storage-follow-through-manifest.md` | done | Slice 092 creates prompt fixtures and deterministic validator. |
| Copilot path instructions | skipped-clear | `docs/plans/2026-06-01-090-kb-cold-storage-follow-through-manifest.md` | done | Slice 093 adds path-scoped repo-local instructions. |

## Dependency Map

```text
Minimality runtime proof
  -> deletion/trim decisions

Go native core gate rewrite
  -> PS1 removal only after parity proof

Cross-model benchmark prompts
  -> benchmark execution later

Copilot path instructions
  -> instruction files only if high-risk file classes are identified
```

## Dark Factory Queue

Completed:

- `docs/plans/2026-06-01-090-kb-cold-storage-follow-through-manifest.md`
  slices 091, 092, and 093
- `docs/plans/2026-06-01-100-kb-go-native-core-gate-rewrite-manifest.md`
  slices 101 through 105

## Human Checkpoints

1. Answered 2026-06-01: full non-PowerShell rewrite if it works for Windows+;
   remove PS1 only after parity proof.

## Parked / Blocked

- No deletion is approved yet. Static cold-storage candidates remain candidates,
  now with evidence classes.

## Completion Criteria

- Every workstream is done.
- `todo.md` and `todo-done.md` reflect completion.
- Final proof uses the Go-native gate.
