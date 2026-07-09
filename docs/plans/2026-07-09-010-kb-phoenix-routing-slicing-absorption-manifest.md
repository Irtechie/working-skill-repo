---
type: kb-manifest
kb_id: kb-2026-07-09-phoenix-routing-slicing-absorption
created: 2026-07-09
status: completed
workflow_shape: "skill-bundle-change"
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "proves the finished skill bundle passes repo-local checks and required sync drift gates"
scope_verified_files:
  - .github/skills/kb-complete/SKILL.md
  - .github/skills/kb-gate/SKILL.md
  - .github/skills/kb-goal/SKILL.md
  - .github/skills/kb-plan/SKILL.md
  - .github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1
  - .github/skills/kb-start/SKILL.md
  - .github/skills/kb-work/SKILL.md
  - README.md
  - cmd/kbcheck/checks.go
  - cmd/kbcheck/checks_test.go
  - cmd/kbcheck/dishonest_completion.go
  - cmd/kbcheck/main.go
  - cmd/kbcheck/manifest_contract.go
  - cmd/kbcheck/manifest_contract_test.go
  - cmd/kbcheck/run_state.go
  - cmd/kbcheck/run_state_test.go
  - cmd/kbcheck/skill_repo_contract_test.go
  - cmd/kbcheck/skill_validators.go
  - cmd/kbcheck/skill_validators_test.go
  - cmd/kbcheck/swarm.go
  - docs/context/architecture/kb-workflow.md
  - docs/context/eval-map.md
  - docs/context/operations/skill-bundle-maintenance.md
  - docs/context/operations/testing.md
  - docs/results/2026-07-09-kb-phoenix-routing-slicing-result.md
  - evals/dishonest-completion/fixtures.json
source_material:
  - docs/context/research/2026-07-05-atv-phoenix-self-heal-comparison.md
  - docs/plans/2026-07-05-000-kb-phoenix-proof-spine-merge-manifest.md
  - docs/plans/2026-07-05-010-kb-model-agnostic-planner-economy-manifest.md
scope_summary: "Do not replace KB with Phoenix. Keep KB routing/slicing as the core and absorb Phoenix's remaining objective done-check, per-slice proof-check, bounded run-state, and route-history guardrails."
already_done:
  - "kbcheck sense / trace-verify / accept proof spine exists"
  - "kbcheck learning-adoption measured promotion gate exists"
  - "kb-plan has model_tier and model_route guidance"
  - "kb-work, kb-repair, kb-troubleshoot, kb-complete reference RED-before-GREEN proof"
  - "README warns against installing Phoenix as a competing lifecycle"
outstanding:
  - "top-level done_check is not a manifest/goal contract"
  - "per-slice proof_check/check_spec is not part of slice schema or validator"
  - "manifest validator does not enforce model_route"
  - "KB-native .kb/runs/<goal>/ state and route-history are not defined"
  - "dynamic routing has no oscillation/confidence guard"
  - "kb-regression-snapshot script still defaults to .atv/snapshots"
  - "skill-sync-report is read-only; there is no first-class kb-doctor repair command"
  - "KB has only a small deterministic false-completion corpus; broader live-run corpus still needs growth"
  - "manifest-contract validates recorded proof_check fields but does not yet execute every proof_check command itself"
model_tier_contract:
  allowed: [tiny, small, medium, large]
  default: medium
  routes:
    tiny: ["local-5090-coder", "local-5090-classifier"]
    small: ["local-5090", "local-5090-coder"]
    medium: ["hosted-sonnet", "hosted-gpt-medium"]
    large: ["hosted-large", "hosted-opus-class", "hosted-gpt-large"]
gate_ledger:
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all six slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk, model_tier, and model_route"
      - "README points to the outstanding Phoenix-routing plan"
      - "todo has an active work row for this manifest"
    proof:
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
      - docs/plans/2026-07-09-011-tool-snapshot-gate-health-plan.md
      - docs/plans/2026-07-09-012-tool-manifest-done-proof-schema-plan.md
      - docs/plans/2026-07-09-013-skill-done-proof-wiring-plan.md
      - docs/plans/2026-07-09-014-kb-run-state-route-history-plan.md
      - docs/plans/2026-07-09-015-eval-doc-release-plan.md
      - docs/plans/2026-07-09-016-tool-kb-doctor-install-drift-plan.md
      - README.md
      - todo.md
    blockers: []
    passed_at: "2026-07-09T12:17:02-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "snapshot script default changed from .atv/snapshots to .kb/snapshots"
      - "default snapshot verify passes"
      - "active runtime snapshot surfaces no longer reference .atv/snapshots"
      - "Go module command hang is documented with bounded repro commands and a narrow non-Go proof path"
      - "git diff --check passes"
    proof:
      - .github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1
      - docs/context/operations/testing.md
      - docs/handoffs/active/2026-07-09-phoenix-routing-go-gate-blocker.md
      - todo.md
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
    blockers: []
    passed_at: "2026-07-09T13:02:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "objective_contract requires top-level done_check"
      - "objective_contract requires proof_check or justified no_check_reason"
      - "model_route is validated against model_tier_contract routes"
      - "existing manifest contract tests still pass"
      - "current manifest passes manifest-contract validation"
    proof:
      - cmd/kbcheck/manifest_contract.go
      - cmd/kbcheck/swarm.go
      - cmd/kbcheck/manifest_contract_test.go
      - docs/plans/2026-07-09-012-tool-manifest-done-proof-schema-plan.md
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
    blockers: []
    passed_at: "2026-07-09T13:34:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "kb-plan emits objective_contract, done_check, proof_check, and model_route guidance"
      - "kb-work validates objective-contract manifests before execution and before done"
      - "kb-goal requires a done check or scoped human exception"
      - "kb-complete summarizes done/proof evidence before terminal completion"
      - "kb-gate blocks phase advancement on missing objective/proof fields"
      - "eval-map documents manifest-contract as a P0 proof surface"
      - "current manifest passes manifest-contract validation"
      - "skill-lint and core-list checks pass"
    proof:
      - .github/skills/kb-plan/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-goal/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-gate/SKILL.md
      - docs/context/eval-map.md
      - docs/plans/2026-07-09-013-skill-done-proof-wiring-plan.md
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
    blockers: []
    passed_at: "2026-07-09T15:46:29-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "kbcheck validates route-history JSONL for oscillation, low-confidence, and no-progress loops"
      - "run-state selftest exercises valid, oscillating, low-confidence, no-progress, and malformed histories"
      - "core check discovery includes kb-run-state-selftest"
      - "kb-goal documents .kb/runs/<goal>/ shape and run-state guard"
      - "kb-start validates route history before looping"
      - "kb-work records route-history progress events when run-state is active"
      - "workflow architecture documents KB-native run state and durable-vs-ephemeral boundary"
      - "focused and broad kbcheck tests pass"
      - "regression snapshot captured and replayed"
    proof:
      - cmd/kbcheck/run_state.go
      - cmd/kbcheck/run_state_test.go
      - cmd/kbcheck/main.go
      - cmd/kbcheck/checks.go
      - cmd/kbcheck/checks_test.go
      - cmd/kbcheck/skill_repo_contract_test.go
      - .github/skills/kb-goal/SKILL.md
      - .github/skills/kb-start/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - docs/context/architecture/kb-workflow.md
      - docs/plans/2026-07-09-014-kb-run-state-route-history-plan.md
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
    blockers: []
    passed_at: "2026-07-09T16:04:06-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
  - gate_id: slice-slice-006-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "doctor command reports configured skill install drift"
      - "doctor --fix repairs marked stale required targets from working source"
      - "doctor --fix refuses unknown required drift with merge-back instruction"
      - "doctor selftest is registered in the core check list"
      - "README lists doctor, manifest-contract, and run-state commands"
      - "maintenance docs explain skill-sync-report versus doctor --fix"
      - "focused and broad kbcheck tests pass"
      - "skill-lint passes without nonexistent kb-doctor warning"
      - "regression snapshot captured and replayed"
    proof:
      - cmd/kbcheck/main.go
      - cmd/kbcheck/checks.go
      - cmd/kbcheck/checks_test.go
      - cmd/kbcheck/skill_repo_contract_test.go
      - cmd/kbcheck/skill_validators.go
      - cmd/kbcheck/skill_validators_test.go
      - README.md
      - docs/context/operations/skill-bundle-maintenance.md
      - docs/plans/2026-07-09-016-tool-kb-doctor-install-drift-plan.md
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
    blockers: []
    passed_at: "2026-07-09T16:12:04-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "dishonest-completion fixtures reject vacuous done checks, missing proof checks, invalid model routes, and route oscillation"
      - "RESULT artifact records corpus, commands, environment, limitations, and Phoenix credit without borrowing Phoenix metrics"
      - "README, eval-map, and testing docs list the implemented proof surfaces"
      - "required Codex, Copilot, shared agents, and ATV .github skill roots are synced"
      - "local-release passes with zero required failures and zero optional failures"
    proof:
      - evals/dishonest-completion/fixtures.json
      - cmd/kbcheck/dishonest_completion.go
      - docs/results/2026-07-09-kb-phoenix-routing-slicing-result.md
      - README.md
      - docs/context/eval-map.md
      - docs/context/operations/testing.md
      - docs/plans/2026-07-09-015-eval-doc-release-plan.md
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
    blockers: []
    passed_at: "2026-07-09T16:30:00-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "all manifest slices are done or terminal"
      - "each done slice has a passed slice-to-done gate"
      - "objective contract manifest validates"
      - "false-completion fixture selftest passes"
      - "local-release passes after required skill sync"
      - "Phoenix self-healing credit and KB result artifact are published"
      - "scope_verified_files is populated for touched workflow surfaces"
      - "ATV checkout diff has no whitespace/conflict failures"
    proof:
      - docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md
      - docs/plans/2026-07-09-011-tool-snapshot-gate-health-plan.md
      - docs/plans/2026-07-09-012-tool-manifest-done-proof-schema-plan.md
      - docs/plans/2026-07-09-013-skill-done-proof-wiring-plan.md
      - docs/plans/2026-07-09-014-kb-run-state-route-history-plan.md
      - docs/plans/2026-07-09-015-eval-doc-release-plan.md
      - docs/plans/2026-07-09-016-tool-kb-doctor-install-drift-plan.md
      - docs/results/2026-07-09-kb-phoenix-routing-slicing-result.md
      - evals/dishonest-completion/fixtures.json
      - README.md
    blockers: []
    passed_at: "2026-07-09T16:36:00-04:00"
    allowed_next_action: "kb-complete docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md"
slices:
  - id: slice-001
    title: "Fix snapshot path drift and gate health"
    path: docs/plans/2026-07-09-011-tool-snapshot-gate-health-plan.md
    blockers: []
    verification: integration
    test_level: integration
    functional_risk: none
    model_tier: small
    model_route: local-5090-coder
    proof_check:
      kind: command_exit
      command: "powershell -NoProfile -ExecutionPolicy Bypass -File .github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1 verify"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Fix .atv snapshot default, diagnose current Go gate hangs, and prove the maintainer gate can list/run again before adding more harness code."
    human_action: ""
    can_continue_other_slices: false
    notes: "scope-forecast: loaded 3 expected files + 0 convention-matched tests; scope-forecast-unused: cmd/kbcheck - Go command hang reproduced but code not touched; scope-check: forecast=3 changed=3 discovered=1 unexplained=0; scope-discovery: docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md - manifest status/gate evidence required by kb-work; proof: snapshot verify default exit=0 PASS 0/0; proof: rg active snapshot surfaces found no .atv/snapshots runtime refs; proof: git diff --check exit=0; gate-health: blocked for Go harness expansion because go list/test/run module commands timeout; memory-impact: operational; docs=docs/context/operations/testing.md"
  - id: slice-002
    title: "Add manifest done_check and proof_check schema enforcement"
    path: docs/plans/2026-07-09-012-tool-manifest-done-proof-schema-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: unit
    functional_risk: none
    model_tier: medium
    model_route: hosted-sonnet
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck manifest-contract-selftest"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Write failing validator tests first, then extend the parser/validator until the new schema passes."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 3 expected files + 0 convention-matched tests; scope-discovery: docs/plans/2026-07-09-012-tool-manifest-done-proof-schema-plan.md - protected oracle SHA required by kb-work; scope-discovery: docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md - slice status/gate evidence required by kb-work; scope-check: forecast=3 changed=5 discovered=2 unexplained=0; protected-oracle: cmd/kbcheck/manifest_contract_test.go sha256=2e28ab6a67da6e0ae9d553ad1387a0f7d9bed477da90554fee703c30e58f543a; proof: go test ./cmd/kbcheck -run TestManifestContract -count=1 -timeout=30s exit=0; proof: go test ./cmd/kbcheck -count=1 -timeout=120s exit=0; proof: go run ./cmd/kbcheck manifest-contract --manifest docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md exit=0; memory-impact: durable; areas=kbcheck manifest contract"
  - id: slice-003
    title: "Wire done/proof checks into KB skills"
    path: docs/plans/2026-07-09-013-skill-done-proof-wiring-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: integration
    functional_risk: none
    model_tier: medium
    model_route: hosted-sonnet
    proof_check:
      kind: command_exit
      command: "rg -n \"objective_contract|done_check|proof_check|model_route\" .github/skills/kb-plan/SKILL.md .github/skills/kb-work/SKILL.md .github/skills/kb-goal/SKILL.md .github/skills/kb-complete/SKILL.md .github/skills/kb-gate/SKILL.md docs/context/eval-map.md"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Update kb-plan, kb-work, kb-goal, kb-complete, kb-gate, and eval-map so the new checks are emitted and enforced."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 6 expected files + 0 convention-matched tests; scope-discovery: docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md - objective contract fields, gate evidence, and status required by kb-work; scope-discovery: docs/plans/2026-07-09-013-skill-done-proof-wiring-plan.md - slice-local proof_check/status required by kb-work; scope-discovery: README.md - user requested explicit Phoenix self-healing credit; scope-discovery: todo.md - board status sync required by kb-work; scope-check: forecast=6 changed=10 discovered=4 unexplained=0; proof: rg objective_contract/done_check/proof_check/model_route surfaces exit=0; proof: go run ./cmd/kbcheck manifest-contract --manifest docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md exit=0; proof: go run ./cmd/kbcheck skill-lint exit=0 warnings-only; proof: go run ./cmd/kbcheck core --list exit=0; snapshot: capture PASS slice-003 and verify PASS 1/1; memory-impact: durable; areas=KB objective contract skills, eval-map"
  - id: slice-004
    title: "Define KB-native run state and route-history guards"
    path: docs/plans/2026-07-09-014-kb-run-state-route-history-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: integration
    functional_risk: none
    model_tier: large
    model_route: hosted-large
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck run-state-selftest"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Define .kb/runs/<goal>/ state, route-history JSONL, oscillation guard, and confidence fallback without importing Phoenix runtime."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 5 expected file groups + 0 convention-matched tests; scope-discovery: docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md - gate evidence and status required by kb-work; scope-discovery: docs/plans/2026-07-09-014-kb-run-state-route-history-plan.md - slice-local proof_check/status required by kb-work; scope-discovery: todo.md - board status sync required by kb-work; scope-check: forecast=5 changed=13 discovered=3 unexplained=0; proof: go test ./cmd/kbcheck -run \"TestRunState|TestDiscoverSkillRepoChecksIncludesNativeValidators|TestSkillRepoContractForNativeCheckNames\" -count=1 -timeout=60s exit=0; proof: go run ./cmd/kbcheck run-state-selftest exit=0; proof: go test ./cmd/kbcheck -count=1 -timeout=120s exit=0; proof: go run ./cmd/kbcheck skill-lint exit=0 warnings-only; proof: go run ./cmd/kbcheck core --list includes kb-run-state-selftest; snapshot: capture PASS slice-004 and verify PASS 2/2; memory-impact: durable; areas=KB run state, route history guard, workflow architecture"
  - id: slice-005
    title: "Add measured eval result docs and release sync"
    path: docs/plans/2026-07-09-015-eval-doc-release-plan.md
    blockers: [slice-003, slice-004, slice-006]
    verification: verification-only
    test_level: full
    functional_risk: none
    model_tier: medium
    model_route: hosted-sonnet
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck local-release"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add false-done/vacuous-check fixtures, publish RESULT-style KB measurements with honest limits, update README/testing docs, run local-release, and sync skill roots."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 6 expected file groups + 0 convention-matched tests; scope-discovery: cmd/kbcheck/dishonest_completion.go, cmd/kbcheck/main.go, cmd/kbcheck/checks.go, cmd/kbcheck/checks_test.go, cmd/kbcheck/skill_repo_contract_test.go - fixture selftest required implementation and core registration; scope-discovery: docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md - gate evidence and status required by kb-work; scope-check: forecast=6 changed=12 discovered=6 unexplained=0; proof: go run ./cmd/kbcheck dishonest-completion-selftest exit=0 rejected 4/4 fixtures; proof: go test ./cmd/kbcheck -run \"TestManifestContract|TestRunState|TestDiscoverSkillRepoChecksIncludesNativeValidators|TestSkillRepoContractForNativeCheckNames\" -count=1 -timeout=60s exit=0; proof: go run ./cmd/kbcheck skill-sync-report exit=0 required issues=0; proof: go run ./cmd/kbcheck doctor --fix exit=0 required_issues=0; proof: go run ./cmd/kbcheck local-release --json exit=0 required_failures=0 optional_failures=0; snapshot: capture PASS slice-005 and verify PASS 5/5; memory-impact: durable; docs=docs/results/2026-07-09-kb-phoenix-routing-slicing-result.md"
  - id: slice-006
    title: "Add kb-doctor install drift repair"
    path: docs/plans/2026-07-09-016-tool-kb-doctor-install-drift-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: integration
    functional_risk: none
    model_tier: small
    model_route: local-5090-coder
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck doctor-selftest"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Implement doctor around skill-sync-report."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-forecast: loaded 4 expected file groups + 0 convention-matched tests; scope-forecast-unused: config/skill-quality.json - inspected and reused without duplication; scope-discovery: docs/plans/2026-07-09-010-kb-phoenix-routing-slicing-absorption-manifest.md - proof_check, gate evidence, and status required by kb-work; scope-discovery: docs/plans/2026-07-09-016-tool-kb-doctor-install-drift-plan.md - slice-local proof_check/status required by kb-work; scope-discovery: todo.md - board status sync required by kb-work; scope-check: forecast=4 changed=11 discovered=3 unexplained=0; proof: go test ./cmd/kbcheck -run \"TestDoctor|TestSkillSyncReportFindsRequiredDrift|TestDiscoverSkillRepoChecksIncludesNativeValidators|TestSkillRepoContractForNativeCheckNames\" -count=1 -timeout=60s exit=0; proof: go run ./cmd/kbcheck doctor-selftest exit=0; proof: go test ./cmd/kbcheck -count=1 -timeout=120s exit=0; proof: go run ./cmd/kbcheck skill-lint exit=0 warnings-only; proof: go run ./cmd/kbcheck core --list includes kb-doctor-selftest; snapshot: capture PASS slice-006 and verify PASS 3/3; memory-impact: durable; areas=skill sync doctor, install drift repair docs"
---

# KB: Phoenix Routing/Slicing Absorption

## Origin

The useful Phoenix proof spine has already been absorbed. The remaining value is
not Phoenix's lifecycle vocabulary; it is the stricter contract that a goal has
an objective done check, each slice has an objective proof check or explicit
exception, and autonomous routing stops when it oscillates or loses confidence.

The latest critique adds three valid gaps:

- install drift has a report, but no Phoenix-style doctor/repair command;
- deterministic eval fixtures exist, but there is no public RESULT-style record
  proving KB workflow outcomes over a measured corpus;
- public polish should distinguish implemented proof from aspiration.

One critique item is stale locally: this repo already has `LICENSE`.

## Keep / Reject

Keep KB as the routing and slicing core. Do not fork Phoenix as the main app and
do not install the Phoenix skill pack beside KB. Absorb only the remaining
objective-contract mechanics that improve KB's planner and executor.

Reject:

- required Rust/MCP runtime;
- Phoenix lifecycle names as KB replacements;
- TokenMasterX or graph tooling as default;
- Phoenix's published metrics as claims for KB.

## Slice Overview

| # | Slice | Blocked By | Verification | Model | HITL | Status |
|---|---|---|---|---|---|---|
| 1 | Fix snapshot path drift and gate health | - | integration | small / local-5090-coder | no | done |
| 2 | Add manifest done_check and proof_check schema enforcement | slice-001 | tdd | medium / hosted-sonnet | no | done |
| 3 | Wire done/proof checks into KB skills | slice-002 | integration | medium / hosted-sonnet | no | done |
| 4 | Define KB-native run state and route-history guards | slice-002 | integration | large / hosted-large | no | done |
| 5 | Add measured eval result docs and release sync | slice-003, slice-004, slice-006 | verification-only | medium / hosted-sonnet | no | done |
| 6 | Add kb-doctor install drift repair | slice-001 | integration | small / local-5090-coder | no | done |

## Done Criteria

- `kbcheck` can validate manifests with `done_check`, `proof_check`, and
  `model_route`.
- `kb-plan` emits these fields for new manifests and slice plans.
- `kb-work` and `kb-complete` block false completion when required objective
  checks are missing or not accepted.
- Long-running goals have a documented `.kb/runs/<goal>/` state shape and
  route-history/stuck guard.
- `kbcheck doctor` or equivalent reports drift, repairs stale installed copies
  only from the working source, and blocks/flags possible global-only useful
  changes instead of overwriting them silently.
- KB publishes a small RESULT-style measured outcome record with fixtures,
  commands, dates, limitations, and no borrowed Phoenix metrics.
- README and eval-map document what is implemented, what is optional, and what
  Phoenix ideas remain rejected.
