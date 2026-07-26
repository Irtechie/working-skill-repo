---
kb_id: kb-2026-07-26-ddr-route-announcement
slice_id: slice-001
title: "Announce the evidence-backed DDR route before execution"
blockers: []
verification: tdd
test_level: unit
functional_risk: narrow
model_tier: medium
model_tier_reason: "This changes shared orchestration behavior across installed skill copies and must preserve route provenance and fallback safety."
model_requirements:
  - "Can edit shared skill, execution prompt, architecture docs, and contract tests consistently."
  - "Can distinguish a selected route from a conditional explicit fallback."
  - "Can run focused Go contract tests and sync-hash verification."
escalation_triggers:
  - "The existing selector cannot expose enough evidence to name the selected primary route."
  - "The requested fallback wording would imply automatic cross-owner or downward fallback."
  - "Focused DDR contract tests fail outside the new announcement expectation."
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run TestProductionDDRContractExcludesAMRAndSeparatesHostSurfaces -count=1"
  expect: 0
hitl: false
expected_files:
  - path: cmd/kbcheck/ddr_contract_test.go
    op: edit
    scope: "Protect the required pre-execution route announcement and its provenance-safe fallback wording."
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "Require one compact user-visible DDR route line after selection and before execution."
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "Carry the route announcement into each slice execution prompt."
  - path: README.md
    op: edit
    scope: "Show the user-facing announcement format and safe model-name rule."
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "Document announcement timing and evidence boundaries."
  - path: ~/.codex/skills/kb-work/
    op: edit
    scope: "Sync the final approved kb-work skill."
  - path: ~/.copilot/skills/kb-work/
    op: edit
    scope: "Sync the final approved kb-work skill."
  - path: ~/.agents/skills/kb-work/
    op: edit
    scope: "Sync the final approved kb-work skill."
protected_oracles:
  - path: cmd/kbcheck/ddr_contract_test.go
    role: "DDR production contract oracle"
    sha256: "7e692e447faab2c4d6403c4c415c5e06484669356428cff6819dde801c525c26"
    oracle_update_reason: "Containment review proved substring-only checks allowed worker emission or routing authority, post-mutation timing, lower-tier named fallback, duplicate grammar, and dated-plan coupling; amend the oracle with negative mutations and production-surface boundaries."
    update_policy: "The new announcement expectations are fixed before implementation."
status: done
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Wait for the unrelated required global drift workstreams, then rerun local-release."
human_action: ""
can_continue_other_slices: true
---

# Announce the Evidence-Backed DDR Route Before Execution

## What to Build

Immediately after `kb-work` chooses `current` or `delegated` ownership and
resolves any delegated route, but before the slice starts, emit one compact
user-visible line:

```text
DDR route: <current|subagent> | primary: <current orchestrator|evidence-backed-route> | fallback: <none|explicit same-tier/higher reselection|evidence-backed-route (conditional; explicit reselect)> | tier: <small|medium|large> | proof: <short-proof-target>
```

## Acceptance Criteria

- Every ready slice gets exactly one route line before mutation or worker
  dispatch.
- `subagent` is the user-facing label for internal `delegated` ownership.
- A concrete primary or fallback alias appears only when the active host or
  `kbrouter` proves it callable and selected/eligible.
- The fallback field never implies automatic downward or cross-owner fallback.
  A named backup is explicitly conditional on a fresh eligibility check and a
  new execution attempt with a fresh ownership/selection decision; it is never
  a second owner decision within one attempt. Otherwise say
  `explicit same-tier/higher reselection`.
- Current execution names the current route only when host evidence exposes it;
  otherwise use `current orchestrator`.
- The line includes the required tier and the narrow proof target.
- Existing DDR, AMR-exclusion, host-surface, and proof contracts remain intact.
- Repo and Codex/Copilot/Agents `kb-work` copies are hash-identical.

## Test Scenarios

1. The DDR contract test fails when `kb-work` lacks the required route line.
2. The test passes when the skill requires announcement after selection and
   before execution.
3. The execution prompt carries the same timing and evidence rule.
4. The architecture doc explains that a named fallback is conditional, never
   automatic.
5. Focused DDR tests pass and tracked docs contain no unsafe automatic fallback
   wording.

## Scope Boundary

- Do not hard-code any concrete or personal route into shared plans or skills.
- Do not change route selection order, AMR policy, model catalogs, or trust
  rules.
- Do not implement a runtime hook or new CLI command.
- Do not push. Commit only under subsequent explicit user authorization.

## Proof Check

```powershell
go test ./cmd/kbcheck -run TestProductionDDRContractExcludesAMRAndSeparatesHostSurfaces -count=1
```
