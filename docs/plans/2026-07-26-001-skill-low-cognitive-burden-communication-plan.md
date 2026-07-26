---
kb_id: kb-2026-07-26-low-cognitive-burden-communication
slice_id: slice-001
title: "Add low-cognitive-burden communication contract"
blockers: []
verification: verification-only
test_level: integration
functional_risk: narrow
model_tier: medium
model_tier_reason: "The policy spans ambient instructions, question gates, explicit response repair, deterministic proof, and synchronized installs."
model_requirements: ["precise policy editing", "responsibility-boundary judgment", "deterministic Go test coverage", "cross-install hash verification"]
escalation_triggers: ["hard/soft/no-response classes conflict with an existing approval boundary", "local-release fails outside known dirty-worktree scope", "installed copy contains newer useful drift"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-compact", "skill:kb-gate", "file:AGENTS.md", "file:README.md", "file:cmd/kbcheck/communication_contract_test.go"]
shared_resources: ["git:integration-owner", "sync:global-skills"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run TestLowCognitiveBurdenCommunicationContract"
  expect: 0
hitl: false
expected_files:
  - path: AGENTS.md
    op: edit
    scope: "Add the ambient response-shape defaults and exceptions."
  - path: .github/skills/kb-compact/SKILL.md
    op: edit
    scope: "Organize responses by human attention and responsibility instead of word count."
  - path: .github/skills/kb-gate/SKILL.md
    op: edit
    scope: "Translate internal gate classes into plain required, optional, or no-response-needed asks."
  - path: README.md
    op: edit
    scope: "Describe the low-cognitive-burden behavior and credit both upstream sources."
  - path: cmd/kbcheck/communication_contract_test.go
    op: create
    scope: "Protect the ambient, compact, and gate responsibility classes."
  - path: docs/context/research/2026-07-26-low-cognitive-burden-agent-communication.md
    op: create
    scope: "Record prior-art findings and rejected universal terseness rules."
  - path: docs/context/research/README.md
    op: edit
    scope: "Index the reusable research note."
status: done
owner: agent
---

# Add Low-Cognitive-Burden Communication Contract

## Acceptance

- Optimize for comprehension and decision effort, not minimum word count.
- Use plain language and lead with the outcome or the exact human action.
- Classify every apparent question as `must respond`, `may respond - agent can
  handle`, or `no response needed`.
- For a hard stop, explain why only the user can answer, what remains blocked,
  and the recommended choice when one exists.
- For a soft preference, state the agent-owned default and continue without
  blocking unless the user overrides.
- Preserve proof, risks, blockers, exact commands, and safety warnings under
  progressive disclosure.
- Do not invent time estimates or force user action after completed work.
- Keep one existing `kb-compact` skill, synchronize required installs, and
  credit both upstream sources.

## Scope Boundary

No new style skill, plugin, diagnosis claim, universal terseness requirement,
or weakening of real approval and human-only gates.

## Result

Focused proof, core, local-release, and required global synchronization pass.
