# Project Map

Bootstrap: 2026-07-21
Bootstrap confidence: mixed

## Legend

- `[verified]` directly checked in repo files or commands during this refresh
- `[inferred]` strongly implied by adjacent repo evidence but not executed end-to-end
- `[unknown]` visible gap or unmapped surface that still needs confirmation

## What This Is

[verified] Portable KB workflow bundle plus its maintainer harness:

- `.github\skills\*\SKILL.md` and `.github\agents\*.agent.md` define the KB,
  CE, review, todo, and learning lanes.
- `cmd\kbcheck`, `cmd\kbrouter`, and `cmd\amrbench` provide native Go proof,
  model-routing, and AMR benchmark CLIs.
- `bin\kb-install.mjs`, `package.json`, and `scripts\install-kb.ps1` provide
  installer and sync surfaces.
- `evals\` contains deterministic fixture corpora for route selection, skill
  scoring, graph routing, model routing, dishonest completion, and AMR
  conformance.

There is no product app runtime here. The repo proves skill behavior, routing,
install/sync hygiene, and benchmark claims.

## How To Run

[verified] Use one of the supported install paths:

```powershell
npx github:Irtechie/working-skill-repo --target all --profile core
npx github:Irtechie/working-skill-repo --target all --profile full
npx github:Irtechie/working-skill-repo --target repo --repo <path-to-project> --profile core
powershell -ExecutionPolicy Bypass -File scripts\install-kb.ps1 -Target all
```

[verified] Contributor-maintainer entrypoints:

```powershell
go run ./cmd/kbcheck core --list
go run ./cmd/kbcheck core
go run ./cmd/kbcheck plan-worktree-selftest
go test ./cmd/kbcheck -run TerminalCleanup -count=1
go test ./cmd/kbcheck -run 'CargoStorage|CargoBuildStorage' -count=1
go run ./cmd/kbcheck local-release
go run ./cmd/kbrouter --help
go run ./cmd/amrbench conformance --config evals/amr-model-benchmark/config.json --no-paid --require-ready --json
```

## How To Test

[verified] Canonical proof lives in `docs/context/operations/testing.md`. Fast
anchors:

```powershell
npm run test
go build ./...
go vet ./...
go run ./cmd/kbcheck plan-worktree-selftest
go run ./cmd/kbcheck core
go run ./cmd/kbcheck local-release
git diff --check
```

## Current Architecture

[verified] Start at `docs/context/architecture/README.md`. The most common next
docs are:

- `docs/context/architecture/skills.md`
- `docs/context/architecture/kbcheck.md`
- `docs/context/architecture/kbrouter.md`
- `docs/context/architecture/graph-routing.md`

## Subsystem Index

| Area | Read This | Next File | Use When | Confidence |
|---|---|---|---|---|
| Skill system and user-facing lanes | `docs/context/architecture/skills.md` | `.github/skills/<skill>/SKILL.md` | You need a skill purpose, trigger, alias, or workflow owner | [verified] |
| Deterministic proof and release gates | `docs/context/architecture/kbcheck.md` | `cmd/kbcheck/main.go` | You need canonical checks, selftests, or release-blocking commands | [verified] |
| Model routing and optional private routes | `docs/context/architecture/kbrouter.md` | `cmd/kbrouter/main.go` | You need route discovery, selection, approval, or project/user preference behavior | [verified] |
| Graph routing packet/eval surface | `docs/context/architecture/graph-routing.md` | `config/graph-route.schema.json` | You need graph packets, graphify rules, lifecycle/eval fixtures, or traversal recipes | [verified] |
| Eval harness inventory | `docs/context/eval-map.md` | `docs/context/operations/testing.md` | You need existing proof surfaces and exact harness commands | [verified] |
| Installer and sync flow | `docs/context/operations/skill-bundle-maintenance.md` | `bin/kb-install.mjs` | You need install targets, router install behavior, or sync/drift rules | [verified] |
| Workflow operating model | `docs/context/architecture/kb-workflow.md` | `AGENTS.md` | You need the fresh-session loop, phase ownership, or artifact lifecycle | [verified] |
| Learning model | `docs/context/architecture/kb-learning-model.md` | `docs/context/kb/instincts/` | You need instinct scope, promotion, or `.kb` vs tracked state rules | [verified] |
| AMR benchmark / GHCP AIC harness | `docs/context/architecture/ghcp-aic-benchmark.md` | `cmd/amrbench/main.go` | You need no-paid conformance, paired grading, or approval boundary details | [verified] |
| Marketplace and promotion policy | `docs/context/architecture/private-skill-marketplace.md` | `config/skill-marketplace.json` | You need approved-vs-quarantine rules or promotion prerequisites | [verified] |
| Active work and blockers | `todo.md` | linked plans/goals/handoffs | You need current work, blockers, or queued improvements | [verified] |
| Plan / brainstorm / handoff lifecycle | `todo.md` | `docs/brainstorms/`, `docs/plans/`, `docs/handoffs/` | You need where active requirements, manifests, or restart packets live | [verified] |
| Root JS asset purpose | `README.md` | root `avatar-*.js`, `custom-avatars-query-*.js` | You need to decide whether the checked-in JS artifacts are intentional runtime assets or leftover build output | [unknown] |

## Current Work Pointers

- Active board: `todo.md`
- Completed board: `todo-done.md`
- Active handoffs: `docs/handoffs/active/`
- Active plans: `docs/plans/`
- Goals: `docs/context/goals/`
- Maintenance signals: `docs/context/memory-maintenance.md`

## Known Sharp Edges

- [verified] `core` is contributor-safe and repo-local; sync drift blocks only in
  `local-release`.
- [verified] `kbcheck` owns the maintainer gate surface now; PowerShell here is
  installer/helper scope, not the main quality gate.
- [verified] Graph routing is optional-provider safe. `graph-route` validates
  provider-neutral packets; stale or unavailable providers must fall back.
- [verified] `kbrouter` stores route preference and optional private routes
  outside repo-authored manifests; planning records tiers, not concrete models.
- [verified] Each active KB manifest group owns one non-default plan worktree;
  explicit local check-in authority is recorded before mutation, slices commit
  there serially, and disjoint manifests may run concurrently.
  `plan-worktree-selftest` proves the local collision and recovery boundary
  without touching the source checkout or performing delivery.
- [verified] Terminal cleanup is delivery-evidence gated. The current/primary,
  dirty, locked, actively claimed, moved/recreated, default, rewritten, and
  uncontained targets are preserved; unresolved remote-default authority also
  blocks cleanup. A later session verifies the Git-admin
  generation and observed remote SHAs, removes the Git worktree, and deletes
  only an exact merged local feature ref with compare-and-swap. Squash/rebase
  proof, host UI records, and remote-ref deletion remain host-owned. Sweep uses
  the primary checkout as a stable Git context even when `--root` is the target;
  guarded retry can reconcile only an exact empty residual after registration
  is gone and all saved branch, receipt, claim, and delivery identities match.
- [verified] `cmd/amrbench run` remains non-dry blocked until a trusted
  human-approval verifier exists.
- [verified] Cargo checks share the absolute target returned by `kbcheck
  cargo-storage resolve`; guarded finalization removes only marker-owned
  temporary targets and preserves the shared cache.
- [verified] `.github/workflows\` is currently empty; proof is local CLI-driven
  plus repo docs, not checked-in GitHub Actions.
- [inferred] Some long hot-path skills still depend on lazy reference discipline
  to keep startup cost acceptable; keep using the rent audit before trimming.

## Research Index

Use `docs/context/research/README.md`, then jump to:

- `docs/context/research/2026-05-29-skill-repo-gap-audit.md`
- `docs/context/research/2026-05-30-agent-skills-git-distribution.md`
- `docs/context/research/2026-07-09-project-model-routing-surfaces.md`

## Do Not Repeat

- Do not bootstrap consuming-project memory inside this repo.
- Do not sync over global copies without reviewing drift first.
- Do not treat optional providers or dry-run previews as proof of live support.

## Maintenance Notes

Use `docs/context/memory-maintenance.md` for graphify preflight history, stale
map signals, drift-risk notes, and unresolved documentation gaps.

## Learning Model

[verified] Durable KB learning is tracked under `docs/context/kb/`; ephemeral
run artifacts live under `.kb/`.

| Path | Tier | Tracked |
|---|---|---|
| `docs/context/kb/instincts/project.yaml` | project + global instincts | yes |
| `docs/context/kb/instincts/scoped/<scope>.yaml` | workflow/domain instincts | yes |
| `docs/context/kb/kb-completions.txt` | completion counter | yes |
| `.kb/observations.jsonl` | passive observation feed | no |
| `.kb/snapshots/` | regression snapshots | no |

Canonical reference: `docs/context/architecture/kb-learning-model.md`.
