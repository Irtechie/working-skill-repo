---
kb_id: kb-2026-05-31-atv-security-marketplace
slice_id: slice-052
title: "Add OSV dependency-vulnerability proof to ATV security"
blockers: []
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: "E:/all-the-vibes/plugins/atv-skill-atv-security/skills/atv-security/SKILL.md"
    op: edit
    scope: "Add deterministic OSV Scanner guidance to A06 dependency vulnerability checks."
  - path: "E:/all-the-vibes/pkg/scaffold/templates/skills/atv-security/SKILL.md"
    op: edit
    scope: "Sync the final ATV security skill body."
  - path: "E:/all-the-vibes/plugins/atv-pack-security/skills/atv-security/SKILL.md"
    op: edit
    scope: "Sync the final ATV security skill body."
  - path: "E:/all-the-vibes/plugins/atv-everything/skills/atv-security/SKILL.md"
    op: edit
    scope: "Sync the final ATV security skill body."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Copy final source to ATV shipped copies and verify SHA equality."
human_action: ""
can_continue_other_slices: true
---

# Slice 052: ATV Security OSV Propagation

## What To Build

Update the trusted ATV `atv-security` skill so A06 dependency checks prefer OSV Scanner machine evidence when available, then propagate the same skill body to ATV shipped copies.

## Acceptance Criteria

- The A06 section requires `osv-scanner scan source -r <repo-or-scope-path> --format json --output-file docs/security/osv-YYYY-MM-DD.json` when manifests or lockfiles exist and `osv-scanner` is available.
- The skill records command, exit code, JSON report path, and severity counts.
- Critical/high known vulnerabilities become findings; version-age-only claims do not become vulnerability findings.
- Absence of `osv-scanner` is recorded as `skipped-unavailable` with the install command.
- All targeted ATV copies have identical `SKILL.md` hashes.

## Test Scenarios

- `Select-String` confirms the final skill names OSV, skip behavior, and no-fix behavior.
- `Get-FileHash` confirms all targeted ATV `SKILL.md` copies match the trusted source.
- `git -C E:/all-the-vibes diff --check` passes for touched files.

## Scope Boundary

Do not run `osv-scanner fix`. Do not rewrite the broader ATV security taxonomy. Do not touch unrelated dirty ATV KB skill sync changes.

