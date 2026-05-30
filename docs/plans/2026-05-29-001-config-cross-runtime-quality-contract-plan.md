---
kb_id: kb-2026-05-29-cross-runtime-skill-quality
slice_id: slice-001
title: "Define cross-runtime quality contract"
blockers: []
verification: verification-only
test_level: none
functional_risk: none
hitl: false
expected_files:
  - path: config/skill-quality.json
    op: create
    scope: "runtime compatibility matrix, lint budgets, required sync targets, optional sync targets, and eval suite settings"
  - path: docs/context/operations/testing.md
    op: edit
    scope: "point to the new quality contract as the source of truth for skill repo checks"
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: ""
human_action: ""
can_continue_other_slices: true
---

# Slice 001: Define Cross-Runtime Quality Contract

## What To Build

Create the configuration contract that later slices consume. It should describe Codex and GHCP support explicitly, including supported, simulated, warning-only, and unsupported surfaces.

## Acceptance Criteria

- `config/skill-quality.json` exists and is valid JSON.
- The config lists runtime entries for `codex` and `ghcp`.
- Each runtime entry lists instruction surfaces, skill directories, agent support, script support, and eval support.
- The config lists skill lint budgets and an allowlist mechanism for justified long hot-path skills.
- The config lists sync targets with required vs optional/intentional omission classification.
- `docs/context/operations/testing.md` points to the config as the quality contract.

## Expected Files

- `config/skill-quality.json`
- `docs/context/operations/testing.md`

## Test Scenarios

- Parse `config/skill-quality.json` with PowerShell `ConvertFrom-Json`.
- Assert both `codex` and `ghcp` runtime entries exist.
- Assert every sync target has a classification.

## Scope Boundary

This slice creates the contract only. It does not implement lint, eval execution, or drift reporting.

## Dependencies

None.
