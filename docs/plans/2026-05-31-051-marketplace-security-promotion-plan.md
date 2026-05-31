---
kb_id: kb-2026-05-31-atv-security-marketplace
slice_id: slice-051
title: "Promote ATV security into the approved marketplace catalog"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: "E:/agent-marketplace/catalog/approved-skills.json"
    op: edit
    scope: "Add approved atv-security entry with source, approval reason, final sha256, and proof commands."
  - path: "E:/agent-marketplace/catalog/harness-index.json"
    op: edit
    scope: "Register dependency-vulnerability-osv harness."
  - path: "E:/agent-marketplace/harnesses/dependency-vulnerability-osv.json"
    op: create
    scope: "Describe the OSV dependency-vulnerability proof command and pass criteria."
  - path: "E:/agent-marketplace/README.md"
    op: edit
    scope: "Document the approved atv-security promotion and quarantine boundary."
  - path: "E:/agent-marketplace/skills/README.md"
    op: edit
    scope: "List approved marketplace skills and atv-security install intent."
  - path: "E:/agent-marketplace/harnesses/README.md"
    op: edit
    scope: "Document dependency-vulnerability-osv harness use."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Validate marketplace JSON and catalog hash against the final promoted skill."
human_action: ""
can_continue_other_slices: true
---

# Slice 051: Promote ATV Security Into Marketplace

## What To Build

Promote `atv-security` as a trusted, explicitly approved security skill in `E:/agent-marketplace`, with a matching OSV harness entry and clear docs that this is approved catalog content, not quarantine.

## Acceptance Criteria

- `catalog/approved-skills.json` contains exactly one approved `atv-security` entry with source path, approval reason, final SHA256, and proof commands.
- `catalog/harness-index.json` lists `dependency-vulnerability-osv`.
- `harnesses/dependency-vulnerability-osv.json` is valid JSON and describes command, inputs, pass criteria, and skip behavior.
- Marketplace docs make clear that quarantine is not loadable and `atv-security` is approved by exception as trusted ATV security tooling.

## Test Scenarios

- Parse all edited/new JSON files with PowerShell `ConvertFrom-Json`.
- Compare catalog SHA256 to the final marketplace `skills/atv-security/SKILL.md` SHA256.
- Run the skill-repo firebreak check to prove no approved catalog entry resolves into quarantine.

## Scope Boundary

Do not promote unrelated ATV skills. Do not convert marketplace into a global runtime root. Do not install public or quarantined skills.

