---
kb_id: kb-2026-07-21-adhd-output-profile
slice_id: slice-001
title: "Add compact ADHD-friendly response profile"
blockers: []
verification: verification-only
test_level: integration
functional_risk: narrow
model_tier: medium
model_tier_reason: "Cross-install policy wording must preserve proof and safety exceptions."
model_requirements: ["precise policy editing", "cross-install hash verification"]
escalation_triggers: ["local-release fails outside known dirty-worktree scope", "installed copy contains newer useful drift"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-compact", "file:AGENTS.md", "file:README.md"]
shared_resources: ["git:integration-owner", "sync:global-skills"]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
hitl: false
expected_files:
  - path: AGENTS.md
    op: edit
    scope: "Add the ambient response-shape defaults and exceptions."
  - path: .github/skills/kb-compact/SKILL.md
    op: edit
    scope: "Add the reusable action/state/details response profile."
  - path: README.md
    op: edit
    scope: "Describe the behavior and credit the upstream inspiration."
status: pending
owner: agent
---

# Add Compact ADHD-Friendly Response Profile

## Acceptance

- Lead with the outcome or next action; never an announcement.
- Use `Done | Next | Blocked` only when it improves orientation.
- Keep the primary surface to five items and place optional depth under a named
  section without hiding proof, risk, blockers, or safety warnings.
- Do not invent time estimates or force user action after completed work.
- Keep one existing `kb-compact` skill, synchronize required installs, and
  credit `ayghri/i-have-adhd`.

## Scope Boundary

No new skill, plugin, diagnosis claim, or universal restatement requirement.
