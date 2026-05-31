# Memory Maintenance

## Active Signals

| Date | Type | Area | Issue | Suggested Action | Status |
|---|---|---|---|---|---|
| 2026-05-29 | drift-risk | ATV propagation | ATV scaffold/plugin copies differ or omit KB skills. They are intentionally optional thin bundles by default; required targets remain Codex/Copilot/agents/ATV `.github`. | Keep sync report warning-only for optional targets unless a skill is intentionally shipped there. | closed |
| 2026-05-29 | bloat-risk | hot-path skills | Several hot-path skills exceed 400 lines. | Move non-routing templates/examples into lazy references or allowlist with reason. | open |
| 2026-05-30 | distribution-decision | eval harness exporters | Local JSON/Markdown reports are now the source of truth; external dashboard exporters are optional but undecided. | Add an exporter only after a consuming workflow needs Langfuse, Braintrust, LangSmith, Promptfoo, or DeepEval. | open |

## Closed Signals

| Date | Type | Area | Resolution |
|---|---|---|---|
| 2026-05-29 | repeated-rediscovery | skill repo testing | Added `scripts/skill-lint.ps1`, `scripts/route-complexity-eval.ps1`, and `kb-check -All` integration. |
| 2026-05-29 | stale-doc | portable memory contract | Added repo-local memory and documented it as skill-bundle maintenance only. |
| 2026-05-31 | drift-risk | ATV propagation | Codified optional thin ATV scaffold/plugin policy in `AGENTS.md`, `README.md`, and `config/skill-quality.json`; sync report passes with 0 required issues. |
| 2026-05-31 | tooling-gap | OSV security proof | Installed `osv-scanner` locally through the official Go install path; `osv-scanner --version` reports 2.3.8. |
