---
kb_id: kb-2026-07-11-invisible-workflow-ux
slice_id: slice-002
title: "Report route, proof, and measured AIC without unsupported savings claims"
blockers: [slice-001]
verification: integration
test_level: functional-cli
functional_risk: broad
model_tier: large
model_tier_reason: "The slice joins runtime routing identity, OTel AIC evidence, deterministic proof, and user-visible cost claims."
model_requirements: ["telemetry provenance", "cost accounting", "routing receipts", "UX contract"]
escalate_when: ["production OTel adapter cannot prove leaf-call completeness", "requested and actual models mismatch", "baseline comparison is not equivalent"]
proof_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck ./cmd/kbrouter ./internal/modelrouting/..."
  expect: 0
hitl: false
expected_files:
  - path: .github/skills/kb-start/SKILL.md
    op: edit
    scope: "retain a concise route announcement as reassurance rather than hide it"
  - path: .github/skills/kb-work/SKILL.md
    op: edit
    scope: "define one compact completion receipt covering route, result, proof, and measured cost"
  - path: .github/skills/kb-work/references/execution-prompt.md
    op: edit
    scope: "require route/cost evidence fields without asking workers to self-report them"
  - path: .github/skills/kb-models/SKILL.md
    op: edit
    scope: "explain exact cost, unavailable cost, and model-mismatch reporting"
  - path: cmd/kbcheck/execution_telemetry.go
    op: edit
    scope: "ingest exact AIC only from complete trusted leaf-call evidence and preserve unavailable/mismatch states"
  - path: cmd/kbcheck/execution_telemetry_test.go
    op: edit
    scope: "cover exact AIC, partial AIC, duplicate aggregate spans, and model mismatch"
  - path: internal/modelrouting/receipt.go
    op: edit
    scope: "link measured cost evidence to route identity without treating cost as work proof"
  - path: cmd/kbrouter/main.go
    op: edit
    scope: "render or expose the compact receipt from trusted route and telemetry evidence"
  - path: cmd/kbrouter/dispatch_test.go
    op: edit
    scope: "prove receipt identity and no unsupported savings claim"
  - path: evals/model-routing/route-proof-cost-receipt.json
    op: create
    scope: "known-answer fixture for exact, unavailable, mismatch, and baseline comparison states"
  - path: README.md
    op: edit
    scope: "state the refined UX principle and show the compact route/proof/cost receipt"
  - path: docs/context/architecture/kb-workflow.md
    op: edit
    scope: "document inspectable internals with concise automatic reporting"
protected_oracles: []
status: pending
owner: agent
can_continue_other_slices: false
---

# Slice 002 - Route, Proof, and Cost Receipt

## What To Build

After routed work, report a compact factual receipt:

```text
Route: Small attempt -> Medium correction fallback
Result: passed on Small
Proof: go test ./... (exit 0)
Cost: 2.40 AIC
```

If a measured comparable baseline exists:

```text
Cost: 2.40 AIC vs 7.10 AIC direct baseline
```

Never claim savings from model pricing, requested model, incomplete OTel, or an
unpaired historical task.

## Acceptance Criteria

- The selected and actual model/route are distinguished.
- Exact AIC is the sum of unique leaf chat spans only.
- Parent aggregate spans are not double-counted.
- Partial AIC, missing response-model evidence, and mismatches are reported as
  unavailable or mismatched rather than estimated.
- Routing telemetry never substitutes for deterministic proof.
- Savings appear only for a comparable direct baseline with matching task,
  fixture/version, proof, and measurement boundary.
- The receipt is concise and automatic; the user does not have to ask whether
  routing saved tokens or credits.
- Inspectable full receipts remain available for power users.

## Test Scenarios

- Exact routed success reports route, proof, and AIC.
- Internal model substitution reports mismatch and does not qualify the
  requested model.
- One unpriced leaf span makes total AIC unavailable.
- A direct baseline from a different fixture/version cannot support a savings
  claim.
- AMR attempt plus correction reports total end-to-end AIC.
- Proof failure remains failure even when routing cost is low.

## Scope Boundary

Do not duplicate the GHCP AIC benchmark harness or AMR correction implementation
owned by the active `2026-07-11-040` manifest. Consume its proven telemetry
contract after that work stabilizes.
