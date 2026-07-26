---
kb_id: kb-2026-07-26-low-cognitive-burden-communication
slice_id: slice-003
title: "Generate executive briefs and useful visuals from source-owned data"
blockers: []
verification: verification-only
test_level: functional-cli
functional_risk: narrow
execution_class: cli
model_tier: medium
model_tier_reason: "The generator must preserve responsibility and proof boundaries while deciding deterministically when a visual lowers cognitive burden."
model_requirements: ["skill design", "strict JSON validation", "deterministic Markdown and Mermaid generation", "cross-install hash verification"]
escalation_triggers: ["generated output hides a hard response or blocker", "visual input is not source-owned", "visual threshold produces decorative noise", "required global sync fails"]
workspace_mode: shared-serial
conflict_domains: ["skill:kb-executive-brief", "tool:kbbrief", "file:README.md", "file:cmd/kbcheck/communication_contract_test.go"]
shared_resources: ["sync:global-skills"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbbrief ./cmd/kbcheck -run 'TestExecutiveBrief|TestResponsibilityContracts|TestVisualGateAndReferences|TestStrictJSONAndOutput|TestLowCognitiveBurdenCommunicationContract'"
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-executive-brief/SKILL.md
    op: create
    scope: "Define the responsibility-first executive brief workflow and visual gate."
  - path: cmd/kbbrief/main.go
    op: create
    scope: "Strictly validate source-owned JSON and render deterministic Markdown and Mermaid."
  - path: cmd/kbbrief/main_test.go
    op: create
    scope: "Protect responsibility classes, visual thresholds, strict input, and repeatable output."
  - path: cmd/kbbrief/testdata/executive-brief.json
    op: create
    scope: "Provide a compact source-owned example."
  - path: cmd/kbbrief/testdata/executive-brief.golden.md
    op: create
    scope: "Protect the complete generated first-screen shape."
  - path: docs/context/operations/low-burden-review-artifacts.md
    op: edit
    scope: "Document generated briefs as projections of source-owned JSON."
  - path: README.md
    op: edit
    scope: "Document the skill and generator command."
  - path: cmd/kbcheck/communication_contract_test.go
    op: edit
    scope: "Protect the new skill and source/projection boundary."
status: done
owner: agent
---

# Generate Executive Briefs And Useful Visuals

## Acceptance

- Generate an executive first screen from strict schema-versioned JSON.
- Lead with whether a human response is hard-required, optional, or unnecessary.
- Preserve outcome, no more than five key points, agent-handled work, proof,
  risks/later, and the companion source.
- Emit Mermaid automatically only with at least three meaningful nodes and two
  valid relationships.
- Reject unknown JSON fields, broken visual references, blank items, and
  incomplete responsibility contracts.
- Regeneration safely replaces an existing output file.
- Synchronize the skill to Codex, Copilot, and shared-agent roots.

## Scope Boundary

No charting framework, HTML dashboard, LLM summarization runtime, invented
metrics, PR creation, or replacement of source-owned plans/results/proof.

## Result

Implemented `kb-executive-brief` and the deterministic `kbbrief` generator.
Focused tests protect strict input, responsibility classes, visual thresholds,
valid Mermaid references, golden output, and safe regeneration. The skill is
hash-identical across the repository, Codex, Copilot, and shared-agent roots.
