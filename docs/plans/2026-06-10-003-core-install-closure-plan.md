---
kb_id: kb-2026-06-10-skill-bundle-cleanup
slice_id: slice-003
title: "Close or document the core install dependency closure"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: broad
hitl: false
expected_files:
  - path: README.md
    op: edit
    scope: "Document the exact core and full profile contract after dependency closure is settled."
  - path: config/skill-quality.json
    op: edit
    scope: "Add or update install-profile and dependency-closure checks if the contract belongs in the gate."
  - path: cmd/kbcheck
    op: edit
    scope: "Add deterministic profile/dependency validation if current checks do not cover it."
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "Add graceful fallback wording only if core intentionally excludes downstream skills."
  - path: .github/skills/kb-map/SKILL.md
    op: edit
    scope: "Add graceful fallback wording only if core intentionally excludes kb-map-bootstrap."
  - path: .github/skills/kb-complete/SKILL.md
    op: edit
    scope: "Add graceful fallback wording only if core intentionally excludes completion dependencies."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Choose one contract: core is dependency-closed, or core skills degrade explicitly when optional skills are absent."
human_action: ""
can_continue_other_slices: true
---

# Slice 003: Close or document the core install dependency closure

## What To Build

Make the `core` install profile internally honest. A core-only install must not
tell an agent to invoke missing skills without either installing the dependency
closure or providing explicit fallback behavior.

## Acceptance Criteria

- The installed core profile is either dependency-closed or deliberately
  documented as a thin starter profile with safe fallbacks.
- Deterministic checks catch future profile drift.
- README install examples match the implemented profile behavior.
- `kb-start`, `kb-map`, and `kb-complete` do not assume missing skills in a
  supported core-only install.

## Expected Files

- `README.md`
- `config/skill-quality.json`
- `cmd/kbcheck`
- `.github/skills/kb-start/SKILL.md`
- `.github/skills/kb-map/SKILL.md`
- `.github/skills/kb-complete/SKILL.md`

## Test Scenarios

- Installer smoke for the core profile into a temporary destination.
- Dependency-closure check proves referenced required skills are present or have
  documented fallback behavior.
- `go run ./cmd/kbcheck core` includes the profile check if implemented.

## Scope Boundary

Do not broaden `core` just because a skill mentions an optional advanced lane.
Only mandatory startup/completion dependencies belong in a closure.

## Dependencies

Blocked by `slice-001`.
