---
kb_id: kb-2026-07-26-low-cognitive-burden-communication
slice_id: slice-005
title: "Trial the workbench on real open pull requests"
blockers: [slice-004]
verification: integration
test_level: functional-browser
functional_risk: narrow
execution_class: headless-browser
model_tier: medium
model_tier_reason: "The trial needs source-anchored PR synthesis plus rendered-browser verification without publishing or mutating GitHub."
model_requirements: ["GitHub PR evidence", "source-diff inspection", "browser DOM assertions", "visual comprehension judgment"]
escalation_triggers: ["a PR head changes during collection", "source evidence cannot be inspected", "the HTML hides a blocker", "the first screen exceeds its cognitive-load budget"]
workspace_mode: shared-serial
conflict_domains: ["artifact:.kb/pr-review-workbench", "file:docs/results/2026-07-26-pr-review-workbench-trial.md"]
shared_resources: ["github:read-only", "browser:local-html"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run TestPRReviewWorkbenchContract -count=1"
  expect: 0
hitl: false
expected_files:
  - path: .kb/pr-review-workbench/
    op: create
    scope: "Store ephemeral source packets, generated HTML, and browser evidence."
  - path: docs/results/2026-07-26-pr-review-workbench-trial.md
    op: create
    scope: "Record PRs, pinned SHAs, evidence gaps, browser proof, and comprehension findings."
status: done
owner: agent
---

# Trial The PR Review Workbench

## Acceptance

- Collect the two open publication PRs without executing PR-controlled code.
- Inspect their changed source and add only source-anchored behavioral claims.
- Render one offline HTML workbench per repository.
- Browser assertions prove one decision state, at most five facts, one next
  action, working evidence tabs, original-PR link, and no external requests.
- Record what became easier to understand and any remaining review burden.

## Scope Boundary

No review submission, approval, requested changes, merge, PR-body edit, Pages
deployment, or claim that the PR is correct.

## Result

Two commit-pinned workbenches were generated and their deterministic HTML
contract passed. The first-screen comparison immediately separated the lean
public package from the larger source/test/release PR. Fresh in-app preview is
unavailable because the Codex browser security policy rejects `file://`
navigation; after the user selected downloadable HTML as sufficient delivery,
that optional preview does not block the trial. No alternate browser workaround
was attempted. See
`docs/results/2026-07-26-pr-review-workbench-trial.md`.
