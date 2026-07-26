---
type: kb-manifest
kb_id: kb-2026-07-26-ddr-route-announcement
brainstorm_path: direct-chat
created: 2026-07-26
status: active
workflow_shape: skill-bundle-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go test ./cmd/kbcheck -run TestProductionDDRContractExcludesAMRAndSeparatesHostSurfaces -count=1"
  expect: 0
  why: "Proves the shared DDR contract requires an evidence-backed route announcement before execution."
model_tier_contract:
  allowed: [small, medium, large]
  default: medium
model_selection_contract:
  timing: work-time
  decision_owner: orchestrator
  owner_choice: current-or-delegated
  max_owner_decisions_per_slice: 1
  catalog: active-host-plus-user-local
  delegated_fallback: same-tier-then-higher
  automatic_downward_routing: false
  amr_required: false
  automatic_cross_owner_fallback: false
  model_requirements:
    - "Use the active host schema and user-local catalog as the only route-name evidence."
    - "Preserve one owner decision, same-tier-or-higher delegation, and deterministic proof."
  escalation_triggers:
    - "No eligible route can be proven for the required tier."
    - "A proposed named fallback would bypass a fresh eligibility check or change owners silently."
scope-verified-files:
  - cmd/kbcheck/ddr_contract_test.go
  - .github/skills/kb-work/SKILL.md
  - .github/skills/kb-work/references/execution-prompt.md
  - README.md
  - docs/context/architecture/kb-workflow.md
  - docs/plans/2026-07-26-000-kb-ddr-route-announcement-manifest.md
  - docs/plans/2026-07-26-001-tool-ddr-route-announcement-plan.md
  - todo.md
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "The direct user request defines the desired pre-execution route narration."
      - "No unresolved product, safety, or architecture decision remains."
    proof:
      - "User request in the active task"
      - README.md
      - .github/skills/kb-work/SKILL.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-plan DDR route announcement"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "Manifest and slice plan exist."
      - "The one-slice DAG has no missing blockers or cycles."
      - "The slice declares acceptance criteria, expected files, verification, risk, model tier, and proof."
    proof:
      - docs/plans/2026-07-26-000-kb-ddr-route-announcement-manifest.md
      - docs/plans/2026-07-26-001-tool-ddr-route-announcement-plan.md
      - "The one-slice DAG has no blockers and therefore no missing references or cycle."
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-000-kb-ddr-route-announcement-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "The announcement behavior is implemented on every shared execution surface."
      - "The protected DDR oracle showed RED before implementation and remained unchanged through GREEN."
      - "Focused deterministic proof passed."
      - "QA lint/diff checks passed and browser QA was correctly classified."
      - "Regression snapshot capture and full replay passed."
      - "Repo/global skill sync was verified."
      - "Scope verification found no unexplained files."
      - "Memory impact was classified and durable architecture docs were updated."
    proof:
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-work/references/execution-prompt.md
      - cmd/kbcheck/ddr_contract_test.go
      - README.md
      - docs/context/architecture/kb-workflow.md
      - .kb/snapshots/ddr-route-announcement.json
      - "Focused DDR contract showed RED then GREEN and the protected oracle hash was preserved."
      - "Formatting and diff checks passed; browser QA was skipped because no UI behavior changed."
      - "Pre-sync drift review found the same two expected announcement files changed in Codex, Copilot, and Agents, with no target-only files or useful global-only work."
      - "Codex, Copilot, and Agents kb-work directory diffs against source: 0 lines."
      - "Regression snapshot replay passed all 17 checks after review repair and recapture."
      - "scope-check: forecast=8 changed=8 discovered=3 unexplained=0."
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "hold at work-to-complete until the repo release gates are responsive"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: blocked
    required_evidence:
      - "Every non-skipped slice has a passing terminal gate."
      - "Final deterministic verification passed."
      - "No unresolved P0 or P1 issue was introduced by this slice."
      - "Board and manifest are synchronized."
      - "Scope-verified files are recorded."
      - "Durable workflow documentation was refreshed."
    proof:
      - docs/plans/2026-07-26-000-kb-ddr-route-announcement-manifest.md
      - cmd/kbcheck/ddr_contract_test.go
      - .kb/snapshots/ddr-route-announcement.json
      - todo.md
      - "scope-check: forecast=8 changed=8 discovered=3 unexplained=0."
      - "README and workflow architecture describe the evidence-bound announcement."
    blockers:
      - "go run ./cmd/kbcheck core timed out after 59 seconds with no output."
      - "go run ./cmd/kbcheck local-release timed out after 59 seconds with no output."
    allowed_next_action: "restore the existing harness-validation workstream, then rerun core and local-release"
slices:
  - id: slice-001
    title: "Announce the evidence-backed DDR route before execution"
    path: docs/plans/2026-07-26-001-tool-ddr-route-announcement-plan.md
    blockers: []
    verification: tdd
    test_level: unit
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "This changes shared orchestration behavior across installed skill copies and must preserve route provenance and fallback safety."
    model_requirements: ["shared skill and contract-test editing", "route provenance and fallback-safety reasoning", "focused Go and sync-hash verification"]
    escalation_triggers: ["selector cannot expose evidence for a primary route", "fallback wording implies automatic cross-owner or downward fallback", "focused DDR tests fail outside the new expectation"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run TestProductionDDRContractExcludesAMRAndSeparatesHostSurfaces -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Wait for the repo validation harness recovery, then rerun core and local-release."
    human_action: ""
    can_continue_other_slices: true
    notes: "execution_owner=current; owner_reason=reasoner-required due overlapping dirty skill/docs context; scope-forecast: loaded 8 expected files + 0 convention-matched tests; scope-discovery: manifest, slice plan, and todo.md are required KB lifecycle artifacts; scope-check: forecast=8 changed=8 discovered=3 unexplained=0; RED/GREEN round 1 established the announcement contract; kb-review found P1 duplicate emission plus P2 current fallback, preview, grammar, and portable-alias issues; protected oracle explicitly amended and RED/GREEN round 2 passed; protected-oracle SHA256 1422b18191baebc8f77750f3a923414aba3e6391e24fd7f47a2ff506aff314e6 preserved; follow-up correctness review found no actionable defects; pre-sync-drift: Codex, Copilot, and Agents each differed only in kb-work/SKILL.md and references/execution-prompt.md, with identical reviewed diffs and no target-only files or useful global-only work; post-sync diffs: 0; qa-lint: PASS gofmt and git diff --check; qa-browser: skipped - no UI-reachable behavior changed; snapshot: PASS 17/17 after review repair and recapture; memory-impact: durable; kb-map-refresh: done through README and workflow architecture update; core and local-release each timed out after 59 seconds with no output, so final release completion remains blocked."
    protected_oracles:
      - path: cmd/kbcheck/ddr_contract_test.go
        role: "DDR production contract oracle"
        sha256: "1422b18191baebc8f77750f3a923414aba3e6391e24fd7f47a2ff506aff314e6"
        oracle_update_reason: "Review P1 proved the first oracle allowed both orchestrator and worker emission; amend it to require one emitting authority and the current-owner fallback."
        update_policy: "The new announcement expectations are fixed before implementation."
---

# KB: DDR Route Announcement

## Origin

Direct user request: make plan execution say whether it is using the current
model or a subagent, including a primary route and safe fallback description.

## Workflow Shape

`skill-bundle-change` — one narrow behavior crosses the shared `kb-work` skill,
its execution prompt, public operating docs, deterministic contract proof, and
global skill copies.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Announce the evidence-backed DDR route before execution | - | tdd | no | done |
