# Memory Maintenance

## Active Signals

| Date | Type | Area | Issue | Suggested Action | Status |
|---|---|---|---|---|---|
| 2026-05-29 | drift-risk | ATV propagation | ATV scaffold/plugin copies previously differed or omitted KB skills. Current policy expects all tracked roots to match unless a packaging exception is recorded. | Keep sync report visible for all targets; treat required drift as blocking. | closed |
| 2026-05-29 | bloat-risk | hot-path skills | Several hot-path skills exceed 400 lines. | Move non-routing templates/examples into lazy references or allowlist with reason. | open |
| 2026-05-30 | distribution-decision | eval harness exporters | Local JSON/Markdown reports are now the source of truth; external dashboard exporters are optional but undecided. | Add an exporter only after a consuming workflow needs Langfuse, Braintrust, LangSmith, Promptfoo, or DeepEval. | open |
| 2026-06-01 | platform-proof | Go gate portability | Windows parity smoke proof exists for the native Go top-level gate, but macOS/Linux have not been exercised on real machines. | Run `go run ./cmd/kbcheck local-release` on macOS/Linux before claiming full OS parity. | open |

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
