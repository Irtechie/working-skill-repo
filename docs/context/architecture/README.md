# Architecture Index

Use this file as the router, not the deep dive.

## First Reads

| Need | Read This | Then Read |
|---|---|---|
| Full workflow shape | `docs/context/architecture/kb-workflow.md` | `docs/context/architecture/skills.md` |
| Deterministic checks / release gates | `docs/context/architecture/kbcheck.md` | `cmd/kbcheck/main.go` |
| Model routing / private routes | `docs/context/architecture/kbrouter.md` | `cmd/kbrouter/main.go` |
| Graph packets / graphify / traversal recipes | `docs/context/architecture/graph-routing.md` | `config/graph-route.schema.json` |
| Skill inventory / triggers | `docs/context/architecture/skills.md` | `.github/skills/<skill>/SKILL.md` |
| Learning scope and instincts | `docs/context/architecture/kb-learning-model.md` | `docs/context/kb/instincts/` |
| Marketplace policy | `docs/context/architecture/private-skill-marketplace.md` | `config/skill-marketplace.json` |
| AMR benchmark / GHCP AIC harness | `docs/context/architecture/ghcp-aic-benchmark.md` | `cmd/amrbench/main.go` |

## Subsystems

| Subsystem | Purpose | Source of Truth | Related Ops / Evals |
|---|---|---|---|
| Skills | User-facing KB/CE/todo/learning workflows and aliases | `.github/skills/**/SKILL.md`; `docs/context/architecture/skills.md` | `cmd/kbcheck skill-lint`; `route-eval`; `skill-eval` |
| `kbcheck` | Deterministic gate, release, proof, eval, and sync CLI | `cmd/kbcheck/`; `config/skill-quality.json`; `docs/context/architecture/kbcheck.md` | `core`; `local-release`; `graph-routing-eval`; `skill-eval` |
| `kbrouter` | Model route discovery, selection, approval, and preference storage | `cmd/kbrouter/`; `internal/modelrouting/`; `docs/context/architecture/kbrouter.md` | `go test ./cmd/kbrouter -run Catalog|Doctor|Policy`; `model-routing-release` |
| Graph routing | Provider-neutral impact packets, fallback policy, optional exact-symbol/graphify adapters | `internal/graphrouting/`; `config/graph-route.schema.json`; `docs/context/architecture/graph-routing.md` | `graph-route`; `graph-routing-lifecycle-selftest`; `graph-routing-eval --require-ready` |
| Installer / sync | Node installer, PowerShell installer, global target propagation, router binary lifecycle | `bin/kb-install.mjs`; `scripts/install-kb.ps1`; `package.json`; `docs/context/operations/skill-bundle-maintenance.md` | `npm run test`; `doctor`; `skill-sync-report` |
| Benchmark / AIC harness | No-paid AMR readiness and paired grading | `cmd/amrbench/`; `evals/amr-model-benchmark/`; `docs/context/architecture/ghcp-aic-benchmark.md` | `amrbench conformance`; `amrbench grade-paired` |
| Marketplace | Approved-vs-quarantine reusable skill flow | `config/skill-marketplace.json`; `docs/context/architecture/private-skill-marketplace.md` | `marketplace-firebreak`; `marketplace-promote` |
| Learning model | Durable instincts and scoped promotion | `docs/context/architecture/kb-learning-model.md`; `docs/context/kb/` | `learning-adoption` |

## Main Workflow Lanes

| Lane | Entry Skill | Notes |
|---|---|---|
| Durable objective | `kb-goal` | Keeps long-running goal state, terminal proof, blockers, and next actions while delegating work to KB lanes |
| Startup / routing | `kb-start` | Calls `kb-map`, then routes by task shape |
| First-principles task | `kb-task` | Uses map, frames assumptions, delegates to the smallest correct lane |
| Project memory | `kb-map`, `kb-map-bootstrap`, `kb-memory-review` | Creates and maintains repo-local memory |
| Requirements / planning | `kb-brainstorm`, `kb-plan`, `kb-gate` | Converts unclear intent into requirements and vertical slices |
| Execution / repair | `kb-work`, `kb-fix`, `kb-troubleshoot`, `kb-repair` | Executes slices or narrow repair loops with proof gates |
| Verification setup | `kb-eval-map` | Maps repo-native eval surfaces |
| Verification | `kb-check`, `kb-functional-test`, `kb-qa`, `kb-regression-snapshot` | Chooses and runs proof |
| Completion | `kb-complete`, `kb-review`, `ce-compound`, `learn`, `evolve` | Review, memory, learning, cleanup |
| Delivery | `kb-ship`, `kb-land`, `kb-finish`, `klfg` | PR or direct-integration surfaces, depending policy |

## Distribution Targets

- Working source: `<working-skill-repo>\.github\skills\<skill>\`
- Required globals: `~/.codex/skills\<skill>\`, `~/.copilot/skills\<skill>\`,
  `~/.agents/skills\<skill>\`
- Approved catalog: `<agent-marketplace>\skills\<skill>\`,
  `<agent-marketplace>\pipelines\<pipeline>.json`

## Current Coverage Notes

- `.github/workflows\` is empty, so local CLI and docs are the checked-in proof
  surface for install/release behavior.
- `docs/context/eval-map.md` and `docs/context/operations/testing.md` are the
  canonical eval/test maps for this repo.
