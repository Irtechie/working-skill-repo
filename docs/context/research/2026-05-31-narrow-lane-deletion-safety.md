# Narrow Lane Deletion Safety

Checked: 2026-05-31

## Verdict

Keep the narrow lanes as lazy helpers for now. They are not base-layer skills,
but deleting them would remove distinct proof behavior that is not fully covered
by the main pipeline.

## Lane Decisions

| Lane | Decision | Distinct Value | Deletion Blocker |
|---|---|---|---|
| `kb-fix` | keep | bounded known-bug loop with reproduce/fix/verify ceiling | `kb-start` routes small bugs here |
| `kb-troubleshoot` | keep | evidence-first autonomous diagnosis for unclear broken behavior | route fixtures and README treat it as distinct from `kb-fix` |
| `kb-functional-test` | keep lazy | test-level classifier and mocked-theater audit | `kb-work` delegates classification and UI proof rules here |
| `kb-regression-snapshot` | keep lazy | cross-slice deterministic replay via DOM/API/CLI/file checks | eval baselines cover harness output, not app workflow replay |

## Cleanup

- Do not load these lanes by default.
- Route to them only from `kb-start`, `kb-plan`, `kb-work`, or `kb-complete`
  when their distinct trigger is present.
- Revisit deletion only after route fixtures and reference scans show no caller
  depends on the lane.
