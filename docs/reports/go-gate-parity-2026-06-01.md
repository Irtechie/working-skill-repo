# Go Gate Parity Report

Generated: 2026-06-01T01:38:25.6443078-04:00
Root: E:\working-skill-repo
Result: PASS

## Commands

| Surface | Command | Exit |
|---|---|---:|
| PS list | C:\Program Files\PowerShell\7\pwsh.exe -NoProfile -File .github/skills/kb-check/scripts/kb-check.ps1 -List | 0 |
| Go list | go run .\cmd\kbcheck core --list | 0 |
| PS core | C:\Program Files\PowerShell\7\pwsh.exe -NoProfile -File .github/skills/kb-check/scripts/kb-check.ps1 -All | 0 |
| Go core | go run .\cmd\kbcheck core | 0 |
| PS local release | C:\Program Files\PowerShell\7\pwsh.exe -NoProfile -File scripts/kb-release-gate.ps1 -Profile local-release -Root E:\working-skill-repo -Json | 0 |
| Go local release | go run .\cmd\kbcheck local-release --json | 0 |

## Check Name Diff

Missing in Go:

- none

Extra in Go:

- none

## PS Check Names

- atv-upstream-delta
- atv-upstream-delta-selftest
- cross-model-benchmark-validate
- go-test
- kb-pipeline-selftest
- kb-release-gate-selftest
- marketplace-promotion-selftest
- route-complexity-eval
- skill-eval
- skill-eval-baseline-selftest
- skill-eval-codex-dry-run
- skill-eval-ghcp-dry-run
- skill-eval-manifest-selftest
- skill-eval-observed-trace-dry-run
- skill-eval-quality
- skill-lint
- skill-marketplace-firebreak
- skill-marketplace-firebreak-selftest
- skill-surface-minimality
- skill-surface-minimality-selftest
- skill-surface-report
- skill-sync-report

## Go Check Names

- atv-upstream-delta
- atv-upstream-delta-selftest
- cross-model-benchmark-validate
- go-test
- kb-pipeline-selftest
- kb-release-gate-selftest
- marketplace-promotion-selftest
- route-complexity-eval
- skill-eval
- skill-eval-baseline-selftest
- skill-eval-codex-dry-run
- skill-eval-ghcp-dry-run
- skill-eval-manifest-selftest
- skill-eval-observed-trace-dry-run
- skill-eval-quality
- skill-lint
- skill-marketplace-firebreak
- skill-marketplace-firebreak-selftest
- skill-surface-minimality
- skill-surface-minimality-selftest
- skill-surface-report
- skill-sync-report

## Removal Gate

PS gate wrapper removal is allowed only when this report says Result: PASS.
This report proves Windows parity on this machine. It does not claim macOS or
Linux runtime proof.
