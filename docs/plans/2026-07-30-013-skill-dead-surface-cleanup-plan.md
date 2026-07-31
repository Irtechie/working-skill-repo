---
kb_id: kb-2026-07-30-current-agent-workflow-cleanup
slice_id: slice-003
title: "Remove dead completion aliases and stale surfaces"
blockers: [slice-002]
verification: integration
test_level: functional-cli
functional_risk: narrow
model_tier: medium
model_tier_reason: "The deletions are clear but every current routing, install, documentation, and guard surface must migrate coherently."
model_requirements: ["call inventory", "route fixtures", "installer profiles", "documentation consistency"]
escalation_triggers: ["an alias owns unique behavior", "a current caller has no replacement", "target deletion is not exact-path bounded"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-finish", "skill:klfg", "docs:skill-inventory", "config:routes"]
shared_resources: ["git:integration-owner", "config:skill-quality"]
proof_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck core"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-finish
    op: delete
    scope: "Remove deprecated compatibility alias"
  - path: .github/skills/klfg
    op: delete
    scope: "Remove deprecated compatibility alias"
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "Route old completion language directly to kb-complete if still accepted"
  - path: .github/skills/kb-task/SKILL.md
    op: edit
    scope: "Remove obsolete lane references"
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Own all completion behavior formerly reached through aliases"
  - path: config/skill-quality.json
    op: edit
    scope: "Remove alias routes and stale profile entries"
  - path: README.md
    op: edit
    scope: "Update installed skills and visible workflow"
  - path: AGENTS.md
    op: edit
    scope: "Remove obsolete protection and caller rules"
  - path: docs/context/architecture/skills.md
    op: edit
    scope: "Regenerate inventory and owner map"
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Build the parity matrix, delete aliases, then clear current references."
human_action: ""
can_continue_other_slices: true
---

# Remove Dead Completion Aliases and Stale Surfaces

Delete compatibility-only skill folders after `kb-complete` and `kb-review`
own every retained behavior.

## Acceptance Criteria

- Capability matrix documents the replacement owner and proof for every named
  deletion.
- `kb-finish`, `klfg`, and `ce-review` do not exist in source or required global
  targets.
- Current route fixtures, installer profiles, configs, ambient instructions,
  and architecture docs do not advertise them.
- Historical plans, solutions, and research may retain factual references.
- Explicit optional skills with distinct jobs remain.

## Test Scenarios

- Current operational search returns no removed route.
- Deprecated invocation fixtures route to `kb-complete` only if compatibility
  language remains supported by the router.
- Installer expected inventory excludes removed folders.

## Scope Boundary

Do not rewrite historical artifacts to erase past decisions.
