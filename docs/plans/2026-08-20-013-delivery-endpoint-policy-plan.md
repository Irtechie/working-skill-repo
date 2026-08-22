---
kb_id: kb-2026-08-20-routing-cognitive-delivery
slice_id: routing-003
title: Make solo and collaborative delivery endpoints explicit
blockers: [routing-002]
verification: tdd
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: Delivery changes commit, push, PR, merge, and protection authority.
model_requirements: [Git delivery policy, authorization boundaries, lifecycle testing]
escalation_triggers: [merge bypass, direct default push, PR state ambiguity]
token_budget: 6000
workspace_mode: shared-serial
conflict_domains: [skill:p2d, skill:w2d, skill:kb-complete, skill:kb-ship, skill:kb-land]
proof_check: {kind: command_exit, command: "go run ./cmd/kbcheck workflow-governor-selftest", expect: 0}
hitl: false
status: pending
---

# Make Delivery Endpoints Explicit

## Acceptance Criteria

- Solo P2D/W2D intent opens and accepts its PR without a second merge question;
  protections still prevent unsafe merge and yield `awaiting-review` with the
  exact unmet condition instead.
- Collaborative completion creates a reviewer-friendly PR and stops at
  `awaiting-review` absent separate merge authority.
- PRs derive concise outcome, change, proof, risk, and follow-up sections from
  structured state.
- Every PR includes `Should we merge?` with a direct recommendation, exact
  satisfied/unsatisfied conditions, reviewer-owned attention, and agent-handled
  items. UI-visible changes attach selected Playwright or host-browser evidence
  images; non-UI changes link the relevant CLI/API/test receipt instead.
- A successful or blocked P2D/W2D merge response includes the PR URL, merged
  commit or unmet condition, proof, and those selected UI evidence images.
- Solo delivery permits one unresolved delivery candidate per repository. A
  completed implementation is not reported as done until the declared delivery
  endpoint is reached; subsequent branch/worktree creation is blocked until
  then unless the user explicitly authorizes a second candidate.
