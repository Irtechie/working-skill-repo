---
type: kb-manifest
kb_id: kb-2026-07-11-invisible-workflow-ux
source: user-decisions-2026-07-11
created: 2026-07-11
status: active
workflow_shape: skill-bundle-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck core"
  expect: 0
  why: "Proves routing fixtures, skill structure, and workflow-governor behavior after simplifying the public workflow."
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: blocked
    required_evidence:
      - "manifest and slice plans exist"
      - "DAG has no missing blockers or cycles"
      - "current AMR/AIC work no longer has overlapping writes on routing, telemetry, README, or config surfaces"
      - "manifest 2026-07-11-040 done_check has passed and its telemetry contract is stable"
      - "user explicitly confirms plan-to-work transition"
    proof:
      - docs/plans/2026-07-11-045-kb-invisible-workflow-ux-manifest.md
      - docs/plans/2026-07-11-046-retire-klfg-preserve-consent-plan.md
      - docs/plans/2026-07-11-047-route-proof-cost-receipt-plan.md
      - "scope collision observed against active 2026-07-10 session-model-routing and 2026-07-11 GHCP AIC work"
    blockers:
      - "active work currently modifies kb-plan, kb-start, kb-work, kb-goal, README, config/skill-quality.json, routing code, and telemetry code"
      - "GHCP AIC falsification manifest 2026-07-11-040 has not yet produced its terminal telemetry proof"
      - "await explicit user confirmation after overlap clears"
    passed_at: ""
    allowed_next_action: "recheck active diffs, then ask whether to run kb-work docs/plans/2026-07-11-045-kb-invisible-workflow-ux-manifest.md"
slices:
  - id: slice-001
    title: "Retire KLFG while preserving explicit standalone execution consent"
    path: docs/plans/2026-07-11-046-retire-klfg-preserve-consent-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The change spans routing policy, compatibility cleanup, route fixtures, and installer-visible skill inventory."
    model_requirements: ["skill routing", "cross-file consistency", "deterministic fixture interpretation"]
    escalate_when: ["legacy callers cannot be mapped to kb-complete or kb-goal", "removing KLFG breaks a supported install/runtime surface"]
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck route-eval"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    can_continue_other_slices: false
    notes: "Execution waits for plan-to-work overlap gate."
    protected_oracles: []
  - id: slice-002
    title: "Report route, proof, and measured AIC without unsupported savings claims"
    path: docs/plans/2026-07-11-047-route-proof-cost-receipt-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "This defines the trust boundary between routing telemetry, exact AIC evidence, deterministic proof, and user-visible claims."
    model_requirements: ["telemetry provenance", "cost accounting", "routing receipts", "UX contract"]
    escalate_when: ["production OTel ingestion is unavailable", "AIC attribution is partial or model identity mismatches", "baseline savings cannot be paired honestly"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck ./cmd/kbrouter ./internal/modelrouting/..."
      expect: 0
    hitl: false
    status: pending
    owner: agent
    can_continue_other_slices: false
    notes: "Consumes rather than duplicates the GHCP AIC falsification work from manifest 2026-07-11-040."
    protected_oracles: []
---

# KB: Invisible Workflow UX

## Decisions

- Remove `klfg`; `kb-complete` owns a single state-aware run and `kb-goal`
  owns durable multi-run objectives.
- Absorb the uncommitted `kb-finalize` phase into `kb-complete`; users and
  internal callers should not need a second completion synonym.
- Keep skill files and manifests inspectable. Voluntary inspection is not
  imposed cognitive load.
- Keep concise route reporting because it reassures users that routing,
  verification, and cost accounting occurred.
- Standalone `kb-plan` asks before transitioning to `kb-work`.
- Standalone `kb-work` returns completed work without publishing. Explicit
  `kb-complete` owns post-work quality and configured delivery.
- Auto-transition remains valid only when the original request included
  execution intent or an already-authorized orchestrator called planning.
- Report measured AIC automatically when exact call evidence exists. Report
  savings only when a comparable direct baseline exists.

## User Experience Principle

> KB should produce better results without requiring users to learn its
> workflow, while clearly reporting what it chose, what it proved, and what it
> cost.

## Dependency Map

```text
slice-001 retire KLFG + preserve consent
  -> slice-002 route/proof/AIC receipt
```

## Scope Collision Gate

Do not execute while the active model-routing/AIC work modifies the same skills,
README, configuration, router, or telemetry files. Re-read the actual diff
immediately before passing `plan-to-work`.
