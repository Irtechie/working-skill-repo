---
kb_id: kb-2026-06-10-skill-bundle-cleanup
slice_id: slice-006
title: "Align docs, selftest wording, hook references, history, and platform claims"
blockers: [slice-001]
verification: verification-only
test_level: functional-cli
functional_risk: narrow
hitl: false
expected_files:
  - path: README.md
    op: edit
    scope: "Correct visible workflow, install, repo-hygiene, and platform-support claims."
  - path: AGENTS.md
    op: edit
    scope: "Align agent instructions with current gate, sync, and repo-hygiene reality."
  - path: .github/copilot-instructions.md
    op: edit
    scope: "Align Copilot instructions with current gate and platform support."
  - path: .github/skills/learn/SKILL.md
    op: edit
    scope: "Remove or caveat references to hook files that are not shipped, if still present."
  - path: cmd/kbcheck
    op: edit
    scope: "Rename successful negative selftest output from failed-looking wording to correctly-rejected wording."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Update expected output and platform caveats."
  - path: docs/context/memory-maintenance.md
    op: edit
    scope: "Record any remaining process-history or platform-proof gaps."
protected_oracles: []
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Patch inaccurate wording only after slice 001 confirms the claim still exists."
human_action: ""
can_continue_other_slices: true
---

# Slice 006: Align docs, selftest wording, hook references, history, and platform claims

## What To Build

Remove misleading docs and output friction that makes the repo look broken when
it is behaving correctly, or overclaims portability where proof is incomplete.

## Acceptance Criteria

- Negative selftests describe expected rejection as `correctly rejected` or an
  equivalent non-failure phrase.
- Missing hook references are either vendored, removed, or clearly marked as
  optional external infrastructure.
- README, `AGENTS.md`, and Copilot instructions agree on core/local-release
  expectations.
- Cross-platform statements match current Go and script reality.
- Historical process docs are either intentionally retained with a repo-local
  policy or moved behind a clearly documented archival boundary.

## Expected Files

- `README.md`
- `AGENTS.md`
- `.github/copilot-instructions.md`
- `.github/skills/learn/SKILL.md`
- `cmd/kbcheck`
- `docs/context/operations/testing.md`
- `docs/context/memory-maintenance.md`

## Test Scenarios

- `go run ./cmd/kbcheck core`
- `go run ./cmd/kbcheck local-release` when release-surface wording changes.
- Targeted search confirms no user-facing "failed" wording remains for
  expected negative selftest passes.
- Targeted search confirms hook and platform claims match shipped files.

## Scope Boundary

Do not delete completed plans or research notes without a retention policy and
human-visible archival decision.

## Dependencies

Blocked by `slice-001`.
