# Memory Maintenance

Last deep review: 2026-07-21

## Counters Since Last Review

- Completed KB cycles: 16
- Durable memory refreshes: 4 (current run; earlier history not reconstructed)
- Closed handoffs: 3 (current run; earlier history not reconstructed)
- Contradiction signals: 3
- Overlap signals: 2
- Stale-doc signals: 2
- Bloat signals: 0
- Repeated-rediscovery signals: 2

## Instinct Roots

Durable instincts live in `docs/context/kb/` (git-tracked). Ephemeral run artifacts live in `.kb/` (git-ignored).

| Path | Purpose | Per-scope cap | Decay |
|---|---|---|---|
| `docs/context/kb/instincts/project.yaml` | project-tier + global-tier instincts (tagged by `scope`) | 50 | `0.5^(days/90)` |
| `docs/context/kb/instincts/scoped/<scope>.yaml` | workflow/domain instincts (default home) | 50 per file | same |
| `docs/context/kb/instincts/archive/` | decayed or evolved instincts | — | — |
| `docs/context/kb/kb-completions.txt` | kb-complete counter | — | — |

The cap and decay apply **per scope file** independently: a busy project tier cannot crowd out scoped learning and vice versa.

Scope hierarchy: `workflow/domain → project → global`. Default = narrowest owning scope. Pull = active scope + ancestors; never siblings. Promotion = nearest common ancestor on cross-sibling recurrence. Landmines = instant one-shot at owning scope.

**X pipeline's lessons are not visible to Y pipeline unless promoted to a shared ancestor.**

Canonical reference: `docs/context/architecture/kb-learning-model.md`.

## Graphify Preflight

graphify-size-check: 2026-07-21 code_files=1797 project_md_bytes=9276 decision=use reason=repo is well above the 200-code-file threshold and graphify.exe is installed; keep ordinary kb-map lookup doc-first and use graph-assisted refresh only for structural traversal work

Notes:

- `graphify.exe` is available at the user Python Scripts path. [verified]
- `uv` is not installed on this workstation. [verified]
- Existing graph-routing proof remains fixture-first through
  `cmd\kbcheck` and `evals\graph-routing\`; no checked-in live provider packet is
  treated as authoritative. [verified]

## Active Signals

| Date | Type | Area | Issue | Suggested Action | Status |
|---|---|---|---|---|---|
| 2026-05-29 | drift-risk | ATV propagation | ATV scaffold/plugin copies previously differed or omitted KB skills. Current policy expects all tracked roots to match unless a packaging exception is recorded. | Keep sync report visible for all targets; treat required drift as blocking. | closed |
| 2026-05-29 | bloat-risk | hot-path skills | Several hot-path skills exceed 400 lines, but the threshold is a review signal rather than an automatic trim target. | Closed by the rent table in `docs/plans/2026-06-10-011-kb-skill-bundle-hardening-manifest.md`; keep always-needed routing/gate content inline and extract only genuinely optional material. | closed |
| 2026-05-30 | distribution-decision | eval harness exporters | Local JSON/Markdown reports are now the source of truth; external dashboard exporters are optional but undecided. | Add an exporter only after a consuming workflow needs Langfuse, Braintrust, LangSmith, Promptfoo, or DeepEval. | open |
| 2026-06-01 | platform-proof | Go gate portability | Windows parity smoke proof exists for the native Go top-level gate, but macOS/Linux have not been exercised on real machines. GitHub workflow mutation was intentionally omitted from the 2026-06-10 push. | Run `go run ./cmd/kbcheck local-release` on macOS/Linux before claiming full OS parity. | open |
| 2026-06-10 | contributor-onramp | core gate | `core` included required sync drift checks, causing fresh or partially installed environments to fail before repo-local quality could be distinguished from deployment drift. | Keep `core` repo-local and enforce required sync drift through `local-release` / `skill-sync-report`. | closed |
| 2026-06-10 | drift-risk | review skills | `ce-review` and `kb-review` intentionally duplicate some reference files for portability, but slice 005 had no deterministic guard for shared-reference drift or documented owner/reason metadata for forks. | `config/skill-quality.json` declares shared review-reference pairs and intentional forks; `kbcheck review-reference-guard` runs in `core`. | closed |
| 2026-06-10 | drift-risk | review reference roots | The first review-reference guard was enumeration-based and did not sweep for newly duplicated common filenames or classify `document-review` overlaps. | Guard now sweeps configured review reference roots, fails unclassified common filenames, and classifies `document-review` overlaps. | closed |
| 2026-06-10 | install-profile | core closure | The core profile previously installed only six skills while those skills invoked a larger runtime graph. | Core now installs every runtime skill plus baseline review/document agents; full adds every specialist agent. | closed |
| 2026-06-10 | minimality | agent surface | `kbcheck minimality` now reports 11 unproven agents after the approved `cli-agent-readiness-reviewer` merge; remaining candidates are static-only and not deletion approvals. | Keep remaining candidates parked until explicit human approval and runtime proof. | open |
| 2026-06-10 | archive-policy | plans directory | Root `docs/plans/` had 100 files, making current work harder to find. | Archived 89 historical plan files under `docs/plans/archive/YYYY-MM/`; keep root focused on current-day active/recent plans. | closed |
| 2026-07-01 | workflow-contract | live steering | KB lacked an in-flight steering layer between one-off PR feedback and post-work `learn`/`evolve`. | Added optional `kb-goal` live steering, feedback classification, docs, and solution note. | closed |
| 2026-07-01 | state-migration | learning roots | Instincts and kb-completions counter lived under `.atv/instincts/` (legacy ATV root), coupling durable learning to the ATV install. | Migrated to `docs/context/kb/instincts/project.yaml` and `docs/context/kb/kb-completions.txt`; added scoped instinct directory `docs/context/kb/instincts/scoped/`; deleted legacy `.atv/` copies (slice-016). | closed |
| 2026-07-05 | workflow-contract | proof spine | KB adopted Phoenix-style failure-first proof, measured learning adoption, and model-tier decomposition contracts. | After real usage, review `cmd/kbcheck` proof-spine checks and skill instructions for over/under-strict acceptance rules. | open |
| 2026-07-09 | provider-hygiene | optional providers | Repo-local and user-global provider state were conflated; substring scanning also misclassified disabled providers. | Core is repo-local; `provider-hygiene --include-user` performs semantic machine inspection and active Phoenix entries fail. | closed |
| 2026-07-09 | contradiction | completion state | A completed scoped-learning goal still pointed to `kb-work`, its manifest remained active/pending, and a resolved handoff remained active. | Reconciled goal/manifest statuses, moved the handoff to done, and reduced `todo.md` to active work only. | closed |
| 2026-07-09 | overlap | provider gate guidance | Provider hygiene extends the existing contributor-core vs release/environment gate rule. | Added a focused provider solution and refreshed the canonical gate solution with a cross-link. | closed |
| 2026-07-09 | repeated-rediscovery | optional provider boundary | Provider optionality had been described in docs but was not mechanically distinguished in local config. | Added semantic provider-hygiene checks, fixtures, and a scoped instinct. | closed |
| 2026-07-12 | contradiction | GHCP AIC handoff state | The active no-paid handoff still described slice 002 blocked and slices 003-004 pending after the manifest recorded all slices done. | Refreshed the handoff to finalization/approval state and retained only the temporary Podman cleanup pointer. | closed |
| 2026-07-12 | stale-doc | AMR benchmark README | The benchmark README advertised hosted model aliases, legacy unpaired qualification, and executable attended runs that no longer matched the fail-closed implementation. | Rewrote the README around no-paid conformance, artifact-derived paired grading, dry-run preview, and disabled provider execution. | closed |
| 2026-07-12 | overlap | evaluation proof semantics | GHCP AIC artifact binding overlapped the existing proof-spine digest solution across problem, root cause, solution, and prevention. | Updated the existing proof-spine solution instead of creating a duplicate. | closed |
| 2026-07-12 | repeated-rediscovery | digest labels versus evidence | The proof-spine had already established semantic digests, but the first benchmark draft again treated self-authored hashes and booleans as authority. | Generalized the solution guidance and added scoped model-routing/evaluation instincts. | closed |
| 2026-07-19 | provider-boundary | evidence graph routing | Graph routing now has optional exact-symbol and structural/flow adapters plus deterministic fixtures. The durable maintenance risk is stale provider evidence being mistaken for proof or distributed locking. | Keep `graph-routing-lifecycle-selftest` and `graph-routing-eval --require-ready` green; refresh packet/provider docs when a real provider is promoted beyond fixtures. | closed |
| 2026-07-21 | stale-doc | repo memory layout | `PROJECT.md` and the architecture index lacked dedicated route docs for `kbcheck`, `kbrouter`, the full skill inventory, and the graph-routing surface. | Refreshed the route map, eval map, testing ops, and added focused architecture docs. | closed |
| 2026-07-21 | unknown-surface | CI/release automation | Release/install proof exists in local Node and Go commands, but `.github/workflows/` is empty and the durable owner of external CI is not documented here. | Decide whether local-only proof is intentional or document/add the checked-in workflow source of truth. | open |
| 2026-07-26 | contradiction | plan-run worktree ownership | The July 19 hybrid guidance allowed independent slices to receive their own worktrees, but real multi-plan collision pressure showed that this was the wrong isolation unit. | Superseded the old slice plan and made one manifest group own one worktree with shared-serial slices. | closed |

## Closed Signals

| Date | Type | Area | Resolution |
|---|---|---|---|
| 2026-05-29 | repeated-rediscovery | skill repo testing | Added `scripts/skill-lint.ps1`, `scripts/route-complexity-eval.ps1`, and `kb-check -All` integration. |
| 2026-05-29 | stale-doc | portable memory contract | Added repo-local memory and documented it as skill-bundle maintenance only. |
| 2026-05-31 | drift-risk | ATV propagation | Codified optional thin ATV scaffold/plugin policy in `AGENTS.md`, `README.md`, and `config/skill-quality.json`; sync report passes with 0 required issues. |
| 2026-05-31 | tooling-gap | OSV security proof | Installed `osv-scanner` locally through the official Go install path; `osv-scanner --version` reports 2.3.8. |
| 2026-05-31 | source-of-truth | ATV upstream resync | Recorded original `All-The-Vibes/ATV-StarterKit` `upstream/main` as authoritative for ATV-native imports, while preserving this repo as the KB overlay. |
| 2026-06-01 | test-flake | eval baseline selftest | Fixed shared temp-directory collision by giving each baseline selftest run a unique `.atv/eval-baseline-selftest-*` directory. |
| 2026-06-01 | test-flake | pipeline selftest | Fixed concurrent pipeline run collision by adding milliseconds plus a GUID suffix to generated pipeline run IDs. |
| 2026-06-01 | stale-doc | gate commands | Replaced current docs and agent instructions that pointed at retired `kb-check.ps1` / `kb-release-gate.ps1` wrappers with `cmd/kbcheck` commands. |
