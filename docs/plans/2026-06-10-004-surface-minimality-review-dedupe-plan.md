---
kb_id: kb-2026-06-10-skill-bundle-cleanup
slice_id: slice-004
title: "Reduce or justify loaded agent and skill surface"
blockers: [slice-001]
verification: verification-only
test_level: functional-cli
functional_risk: narrow
hitl: true
expected_files:
  - path: cmd/kbcheck
    op: edit
    scope: "Improve minimality reporting or protected classifications if static evidence is insufficient."
  - path: config/skill-quality.json
    op: edit
    scope: "Record retained-surface allowlist entries with evidence-backed reasons."
  - path: .github/agents
    op: edit
    scope: "Delete, merge, or retain duplicate/orphan agents only after approval and evidence."
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "Record durable minimality decisions and any parked deletion candidates."
  - path: README.md
    op: edit
    scope: "Update visible installed-surface claims if the shipped agent/skill set changes."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Produce an evidence-backed minimality report and request deletion approval before removing files."
human_action: "Approve deletion or merging of any agent/skill files."
can_continue_other_slices: true
test_inputs:
  - name: deletion_approval
    source: user
    required_for: "Removing or merging agent/skill files"
    value: "No deletion requested or performed; candidates remain parked for future human approval."
---

# Slice 004: Reduce or justify loaded agent and skill surface

## What To Build

Turn the orphan/duplicate-agent concern into an evidence-backed minimality
decision: delete, merge, protect, or document retained surfaces.

## Acceptance Criteria

- Static minimality output identifies referenced, unreferenced, protected, and
  duplicate-like surfaces.
- Any deletion candidate includes proof that no skill, config, persona catalog,
  doc, or expected workflow references it.
- Protected skills and agents are retained with explicit reasons.
- Human approval is obtained before destructive deletion.
- The repo's minimality checks and docs match the final surface.

## Expected Files

- `cmd/kbcheck`
- `config/skill-quality.json`
- `.github/agents`
- `docs/context/memory-maintenance.md`
- `README.md`

## Test Scenarios

- `go run ./cmd/kbcheck minimality`
- `go run ./cmd/kbcheck core`
- Targeted reference checks for every deletion candidate.

## Scope Boundary

Do not remove `kb-review`, `ce-review`, `ce-compound`, or
`ce-compound-refresh` unless invoking skills are rewritten first.

## Dependencies

Blocked by `slice-001`.

## HITL

Deletion approval is required. Reporting, allowlists, and non-destructive docs
updates are agent-owned.
