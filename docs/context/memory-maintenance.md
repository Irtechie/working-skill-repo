# Memory Maintenance

## Active Signals

| Date | Type | Area | Issue | Suggested Action | Status |
|---|---|---|---|---|---|
| 2026-05-29 | drift-risk | ATV propagation | ATV scaffold/plugin copies differ or omit KB skills. The report now classifies them as optional, but the shipping policy still needs a human decision. | Decide whether ATV scaffold/plugin should ship the full KB surface. | open |
| 2026-05-29 | bloat-risk | hot-path skills | Several hot-path skills exceed 400 lines. | Move non-routing templates/examples into lazy references or allowlist with reason. | open |

## Closed Signals

| Date | Type | Area | Resolution |
|---|---|---|---|
| 2026-05-29 | repeated-rediscovery | skill repo testing | Added `scripts/skill-lint.ps1`, `scripts/route-complexity-eval.ps1`, and `kb-check -All` integration. |
| 2026-05-29 | stale-doc | portable memory contract | Added repo-local memory and documented it as skill-bundle maintenance only. |
