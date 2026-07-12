---
kb_id: kb-2026-07-11-ghcp-aic-falsification
slice_id: slice-001
title: "Normalize complete GHCP leaf-call accounting"
blockers: []
verification: tdd
test_level: unit
functional_risk: broad
model_tier: medium
context_packet_path: docs/plans/2026-07-11-ghcp-aic-context/slice-001.json
proof_check:
  kind: command_exit
  command: "go test ./internal/ghcpotel ./cmd/amrbench ./cmd/kbcheck -run 'OTel|AIU|Span|Telemetry|Mismatch' -count=1"
  expect: 0
hitl: false
expected_files:
  - path: internal/ghcpotel/parser.go
    op: create
    scope: "Strict bounded JSONL parsing, trace-span identity, leaf accounting, schema/version provenance, and completeness reconciliation."
  - path: internal/ghcpotel/parser_test.go
    op: create
    scope: "RED/GREEN coverage for duplicates, partial export, aiu-nano conflict, overflow, lineage, malformed content, and route mismatch."
  - path: internal/ghcpotel/redact.go
    op: create
    scope: "Convert local raw probes into bounded key-type schema fixtures while rejecting or removing content-bearing values."
  - path: internal/ghcpotel/redact_test.go
    op: create
    scope: "Prove raw prompts, tool payloads, identifiers, usernames, and absolute paths never enter committed fixtures."
  - path: internal/ghcpotel/testdata/schema-probe-redacted.jsonl
    op: create
    scope: "Bounded deterministic schema fixture generated outside the model reasoning packet from the local probe."
  - path: cmd/amrbench/main.go
    op: edit
    scope: "Consume normalized exact accounting and classify mismatches as invalid rather than model failures."
  - path: cmd/amrbench/main_test.go
    op: edit
    scope: "Replace float-only happy-path coverage with adapter integration and observed-route attribution."
  - path: cmd/kbcheck/execution_telemetry.go
    op: edit
    scope: "Reuse the normalized GHCP accounting result without weakening receipt and proof authority."
  - path: cmd/kbcheck/execution_telemetry_test.go
    op: edit
    scope: "Prove raw accounting remains exact, versioned, optional, and never substitutes for proof."
protected_oracles:
  - path: internal/ghcpotel/parser_test.go
    role: "strict complete leaf-call accounting and schema-drift oracle"
    sha256: "filled after RED before implementation"
    update_policy: "requires explicit plan update"
status: pending
owner: agent
can_continue_other_slices: false
---

# Slice 001 - GHCP OTel Accounting

## Acceptance

- Unique cost identity is `(trace_id, span_id)`; `invoke_agent` never adds cost.
- Every credited leaf has one recognized exact cost field and one phase/arm.
- Both documented `aiu` and legacy integer `nano_aiu` are supported without
  float authority; conflicts, partial export, malformed rows, forbidden content,
  and missing closure are invalid.
- Requested and observed model/route evidence stay distinct. Mismatch cost is
  operational waste and never requested-model qualification.
- The raw `.kb` probe is never a worker source file or prompt input. Only the
  redacted bounded test fixture is readable by the implementation worker.
- No model call is part of this slice.

## Scope Boundary

No paid run, sandbox implementation, context experiment, or promotion decision.
