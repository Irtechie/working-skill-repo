---
kb_id: kb-2026-06-10-skill-bundle-cleanup
slice_id: slice-005
title: "Canonicalize review-skill shared doctrine"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: .github/skills/kb-review/SKILL.md
    op: edit
    scope: "Delegate or reference shared review doctrine where duplication is accidental."
  - path: .github/skills/kb-review/references
    op: edit
    scope: "Deduplicate or document KB-specific reference differences."
  - path: .github/skills/ce-review/SKILL.md
    op: edit
    scope: "Delegate or reference shared review doctrine where duplication is accidental."
  - path: .github/skills/ce-review/references
    op: edit
    scope: "Deduplicate or document CE-specific reference differences."
  - path: config/skill-quality.json
    op: edit
    scope: "Add drift checks or allowlists for intentional review-skill splits."
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "Record the canonical review-skill source-of-truth decision."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Compare current review-skill references, then remove accidental forks or document intentional ones."
human_action: ""
can_continue_other_slices: true
---

# Slice 005: Canonicalize review-skill shared doctrine

## What To Build

Make the relationship between `kb-review` and `ce-review` explicit enough that
future edits do not create silent drift across duplicated reference trees.

## Acceptance Criteria

- Byte-identical shared references are either centralized or protected by a
  drift check.
- Intentional differences are documented with owner and reason.
- `kb-review` remains the KB completion review lane.
- `ce-review` remains the generalized CE review lane.
- The quality gate catches future accidental divergence where appropriate.

## Expected Files

- `.github/skills/kb-review/SKILL.md`
- `.github/skills/kb-review/references`
- `.github/skills/ce-review/SKILL.md`
- `.github/skills/ce-review/references`
- `config/skill-quality.json`
- `docs/context/memory-maintenance.md`

## Test Scenarios

- Targeted hash comparison for shared review references.
- `go run ./cmd/kbcheck skill-lint`
- `go run ./cmd/kbcheck core`

## Scope Boundary

Do not collapse the two public skill entry points unless all callers and docs
are updated in the same slice.

## Dependencies

Blocked by `slice-001`.
