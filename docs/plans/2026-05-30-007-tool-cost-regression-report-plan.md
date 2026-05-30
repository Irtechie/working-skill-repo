---
kb_id: kb-2026-05-30-live-cross-runtime-skill-eval-harness
slice_id: slice-007
title: "Add cost and regression reporting"
blockers: [slice-003]
verification: functional
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: scripts/skill-eval-regression-report.ps1
    op: create
    scope: "summarize live run metrics and compare against selected baseline"
  - path: evals/skill-eval/baselines/README.md
    op: create
    scope: "document baseline storage and retention policy"
  - path: evals/skill-eval/README.md
    op: edit
    scope: "document cost proxies and regression comparison commands"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 007: Add Cost And Regression Reporting

## What To Build

Add a local report that answers whether live eval quality improved, regressed, or
shifted spend. It should use available cost proxies: wall-clock time, model,
tool count when parseable, retries, prompt/result sizes, and transcript/log
sizes.

## Acceptance Criteria

- Report can summarize a current run directory.
- Report can compare current results to a selected baseline.
- Missing token/cost data is represented as unavailable, not zero.
- Output is local JSON and Markdown; external dashboards remain exporter-only.

## Verification

- `powershell -ExecutionPolicy Bypass -File scripts\skill-eval-regression-report.ps1 -RunRoot .atv/eval-runs`
- Synthetic baseline comparison with a small checked-in fixture or generated temp fixture.
- `git diff --check`
