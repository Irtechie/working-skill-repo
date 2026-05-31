---
kb_id: kb-2026-05-31-atv-security-marketplace
slice_id: slice-053
title: "Install the approved ATV security skill into global targets"
blockers: [slice-051, slice-052]
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: "E:/agent-marketplace/skills/atv-security/SKILL.md"
    op: edit
    scope: "Install the final approved atv-security skill body."
  - path: "C:/Users/marowe/.codex/skills/atv-security/SKILL.md"
    op: edit
    scope: "Install the final approved atv-security skill body for Codex."
  - path: "C:/Users/marowe/.copilot/skills/atv-security/SKILL.md"
    op: edit
    scope: "Install the final approved atv-security skill body for Copilot."
  - path: "C:/Users/marowe/.agents/skills/atv-security/SKILL.md"
    op: edit
    scope: "Install the final approved atv-security skill body for shared agents."
  - path: "E:/agent-marketplace/catalog/approved-skills.json"
    op: edit
    scope: "Pin the final installed skill SHA256."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Copy final approved skill to marketplace and global targets, update catalog SHA, then prove hash equality."
human_action: ""
can_continue_other_slices: true
---

# Slice 053: Global ATV Security Install

## What To Build

Install only the approved `atv-security` skill into global skill targets and marketplace approved skill storage, then pin the final hash in the marketplace catalog.

## Acceptance Criteria

- Marketplace, Codex, Copilot, and shared agents `atv-security/SKILL.md` files match the trusted ATV source hash.
- `approved-skills.json` pins the same final hash.
- No quarantine file is linked, approved, or copied into global targets.
- No unrelated marketplace or global skill is changed.

## Test Scenarios

- `Get-FileHash` over source, marketplace, and global targets returns one unique SHA256.
- `scripts/skill-marketplace-firebreak.ps1` passes.
- `scripts/skill-marketplace-firebreak-selftest.ps1` proves the negative case still fails.

## Scope Boundary

Do not bulk-install all ATV skills. Do not make the marketplace folder a global load root. Do not overwrite unrelated global skill drift.

