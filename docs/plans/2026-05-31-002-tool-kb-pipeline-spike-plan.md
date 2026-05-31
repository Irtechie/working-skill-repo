---
kb_id: kb-2026-05-31-proof-pipeline-spike
slice_id: slice-002
title: "Build tiny coded pipeline spike"
blockers: []
verification: integration
test_level: cli
functional_risk: narrow
hitl: false
expected_files:
  - path: config/pipelines/
    op: add
    scope: "Add one minimal pipeline definition for skill-bundle/proof work."
  - path: scripts/kb-pipeline.ps1
    op: add
    scope: "Create start/status/resume spike commands with run-folder output."
  - path: .atv/pipeline-runs/
    op: generated
    scope: "Runner-generated run folders; should stay ignored if appropriate."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Document spike command and expected artifacts."
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Implement one boring coded pipeline runner before adding more pipeline types."
human_action: ""
can_continue_other_slices: true
---

# Build Tiny Coded Pipeline Spike

## What To Build

Create a minimal coded pipeline runner that selects or starts one predefined
pipeline and writes auditable run artifacts. It should generate phase prompts
and context packets, not auto-spawn agents.

## Acceptance Criteria

- `scripts/kb-pipeline.ps1 -Start <pipeline-id>` creates a run directory.
- The run records the selected pipeline, phases, required artifacts, protected
  files, and proof commands.
- The runner can show status for the latest or named run.
- The spike supports one pipeline only at first.
- The design is easy to delete if real usage proves it wrong.

## Test Scenarios

- Start the pilot pipeline from a clean repo state.
- Confirm run files are created with stable names.
- Confirm status reads the run back without model judgment.
- Confirm bad pipeline IDs fail deterministically.

## Scope Boundary

Do not build a generic agent framework. Do not auto-run fresh agents yet. Do not
add frontend or audiobook pipelines in this slice.
