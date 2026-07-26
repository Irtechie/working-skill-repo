---
kb_id: kb-2026-07-26-low-cognitive-burden-communication
slice_id: slice-004
title: "Integrate the PR review workbench into lazy PR delivery"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
execution_class: cli
model_tier: medium
model_tier_reason: "The package crosses GitHub evidence, untrusted HTML, review mutation boundaries, skill discovery, and global synchronization."
model_requirements: ["source-package review", "untrusted-content safety", "deterministic CLI proof", "skill sync"]
escalation_triggers: ["the marketplace package differs materially between sources", "the renderer can execute remote content", "lazy loading changes ordinary shipping", "global sync refuses drift"]
workspace_mode: shared-serial
conflict_domains: ["skill:pr-review-workbench", "skill:kb-ship", "skill:kb-executive-brief", "file:README.md"]
shared_resources: ["sync:global-skills"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run TestPRReviewWorkbenchContract -count=1"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/pr-review-workbench/SKILL.md
    op: create
    scope: "Own the lazy PR evidence and offline review workflow."
  - path: .github/skills/pr-review-workbench/agents/openai.yaml
    op: create
    scope: "Expose portable skill metadata."
  - path: .github/skills/pr-review-workbench/references/evidence-contract.md
    op: create
    scope: "Preserve proof, readiness, safety, and first-screen rules."
  - path: .github/skills/pr-review-workbench/scripts/pr_review_workbench.py
    op: create
    scope: "Collect bounded PR evidence and render inert HTML."
  - path: .github/skills/kb-ship/SKILL.md
    op: edit
    scope: "Lazy-load the workbench only after PR creation."
  - path: .github/skills/kb-executive-brief/SKILL.md
    op: edit
    scope: "Route complex PR walkthroughs to the HTML workbench."
  - path: cmd/kbcheck/pr_review_workbench_contract_test.go
    op: create
    scope: "Protect package ownership, lazy routing, safety, and CLI output."
  - path: cmd/kbcheck/testdata/pr-review-workbench.json
    op: create
    scope: "Provide deterministic commit-pinned PR evidence."
  - path: README.md
    op: edit
    scope: "Document the installed skill and lazy PR route."
status: done
owner: agent
---

# Integrate The PR Review Workbench

## Acceptance

- The working skill repository owns a complete runnable package.
- `kb-ship` loads it only after an open PR exists and only when requested or
  enabled by repo policy.
- Ordinary shipping does not load or execute the renderer.
- The HTML stays self-contained, inert, escaped, and commit-pinned.
- Public Pages remains opt-in; private PR evidence is never published by
  default.
- Focused proof, repo gates, and all required global skill copies pass.

## Scope Boundary

No GitHub review submission, merge, automatic Pages deployment, private-data
publication, or modification of the marketplace source repositories.

## Result

The package is repo-owned, lazy-loaded after PR creation, and supports a local
artifact plus a separate downloadable artifact branch. Focused contract proof
passes. Global sync and the final release gate are recorded in the parent
manifest and result.
