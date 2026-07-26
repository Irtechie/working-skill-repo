---
type: kb-manifest
kb_id: kb-2026-07-26-change-aware-proof-governor
brainstorm_path: direct-chat
created: 2026-07-26
status: active
workflow_shape: pipeline-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck proof-governor-selftest"
  expect: 0
  why: "Proves changed behavior is never skipped, passing superset proof is reused only against identical relevant inputs, redundant unchanged replays are blocked, and automatic visible/native GUI execution is denied before launch."
model_tier_contract:
  allowed: [small, medium, large]
  default: large
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
proof_governor_contract:
  bias: soundness-first
  unknown_impact_decision: run
  reuse_requires: [passing-receipt, requested-coverage-subset, exact-relevant-input-fingerprint, compatible-environment, unexpired-policy]
  coverage_source: sealed-check-registry
  working_tree_aware: true
  full_suite_replay_limit_per_fingerprint: 1
  automatic_visible_or_native_gui_execution: blocked
  decision_values: [run, reuse, block]
execution_preconditions:
  - "Do not execute against the current overlapping dirty baseline."
  - "Contain or complete the existing changes to kb-work, todo.md, workflow documentation, and cmd/kbcheck before mutation."
  - "Restore bounded cmd/kbcheck package-load execution or preserve the final release slice as blocked."
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "The direct conversation defines sound reuse, changed-input invalidation, redundant replay refusal, and attended native-GUI behavior."
      - "Current repo rules prove unconditional all-snapshot and all-check replay paths exist."
      - "No unresolved product decision is required to decompose the work."
    proof:
      - "Direct user decision in the active task"
      - .github/skills/kb-regression-snapshot/SKILL.md
      - .github/skills/kb-repair/SKILL.md
      - .github/skills/kb-functional-test/SKILL.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-plan change-aware proof reuse and GUI execution governance"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "Manifest and five slice plans exist."
      - "All five context packets validate."
      - "The DAG has no missing blockers or cycles."
      - "Every slice declares acceptance criteria, expected files, verification, test level, functional risk, model tier, and objective proof."
      - "The deterministic manifest contract passes."
      - "The current overlapping dirty baseline is contained before implementation."
    proof:
      - docs/plans/2026-07-26-020-kb-change-aware-proof-governor-manifest.md
      - docs/plans/2026-07-26-021-tool-proof-coverage-contract-plan.md
      - docs/plans/2026-07-26-022-tool-impact-aware-proof-selection-plan.md
      - docs/plans/2026-07-26-023-tool-proof-execution-and-gui-guard-plan.md
      - docs/plans/2026-07-26-024-eval-proof-governor-release-plan.md
      - docs/plans/2026-07-26-025-tool-gui-fail-closed-simplification-plan.md
      - docs/plans/2026-07-26-proof-governor-context/slice-001.json
      - docs/plans/2026-07-26-proof-governor-context/slice-002.json
      - docs/plans/2026-07-26-proof-governor-context/slice-003.json
      - docs/plans/2026-07-26-proof-governor-context/slice-004.json
      - docs/plans/2026-07-26-proof-governor-context/slice-005.json
      - docs/results/2026-07-26-proof-governor-plan-to-work.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-020-kb-change-aware-proof-governor-manifest.md"
  - gate_id: plan-amendment-gui-fail-closed
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "The user explicitly rejected adding more HITL ceremony."
      - "The existing approval document is not an authenticated trust boundary."
      - "Slice-005 preserves automatic pre-launch denial and all non-GUI proof behavior."
      - "The slice plan and context packet validate without requiring full-suite replay."
    proof:
      - docs/plans/2026-07-26-025-tool-gui-fail-closed-simplification-plan.md
      - docs/plans/2026-07-26-proof-governor-context/slice-005.json
      - cmd/kbcheck/proof_execution.go
      - cmd/kbcheck/proof_execution_test.go
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-020-kb-change-aware-proof-governor-manifest.md slice-005"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Implementation and public CLI path exist."
      - "Protected oracle recorded RED then GREEN and retained its accepted hash."
      - "Focused deterministic proof passed."
      - "Scope, QA, snapshot exception, and memory impact are recorded."
    proof:
      - docs/results/2026-07-26-proof-governor-slice-001.md
      - cmd/kbcheck/proof_governor.go
      - cmd/kbcheck/proof_governor_test.go
      - config/proof-governor.schema.json
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "slice-002"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Read-only selector and strict registry exist."
      - "Protected oracle recorded RED then GREEN and retained its accepted hash."
      - "Focused selector, impact, composite, release, and public CLI proof passed."
      - "Scope, QA, snapshot exception, and memory impact are recorded."
    proof:
      - docs/results/2026-07-26-proof-governor-slice-002.md
      - cmd/kbcheck/proof_selection.go
      - cmd/kbcheck/proof_selection_test.go
      - cmd/kbcheck/release.go
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "slice-003"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Historical governed execution, replay budget, audit, timeout, and GUI denial proof exist; slice-005 supersedes its approval mechanism."
      - "Protected oracle recorded RED then GREEN and retained its accepted hash."
      - "Snapshot and workflow contracts select impacted proof and reserve full replay for milestones."
      - "No visible or native GUI process was launched during verification."
    proof:
      - docs/results/2026-07-26-proof-governor-slice-003.md
      - cmd/kbcheck/proof_execution.go
      - cmd/kbcheck/proof_execution_test.go
      - .github/skills/kb-regression-snapshot/scripts/kb-regression-snapshot.ps1
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "slice-004"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: blocked
    required_evidence:
      - "Protected fixture corpus and no-GUI selftest pass."
      - "Core and local-release each pass once on the final tree."
      - "Approved skill copies are hash-identical across required targets."
    proof:
      - docs/results/2026-07-26-proof-governor-slice-004.md
      - evals/proof-governor/fixtures.json
      - cmd/kbcheck/proof_governor_selftest.go
    blockers:
      - "core failed once on twelve unchanged cmd/kbrouter canonical-project-path tests; local-release was not launched because it composes the same core failure."
      - "skill-sync-report has unrelated required drift for Copilot kb-gate and safe-shell-quoting in all three global roots; the nine proof-governor skills are hash-identical."
    passed_at: ""
    allowed_next_action: "repair cmd/kbrouter fixtures and unrelated required sync drift, then run one fresh core and one local-release"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Protected oracle recorded the old reason RED and the fail-closed GUI behavior GREEN."
      - "Approval CLI, types, validation, and marker handling are absent from current code."
      - "Focused CLI, manifest, selftest, lint, hash-sync, and diff proof pass."
      - "No visible or native GUI process launched."
      - "The unchanged slice-004 full-release blocker was not replayed."
    proof:
      - docs/results/2026-07-26-proof-governor-slice-005.md
      - cmd/kbcheck/proof_execution.go
      - cmd/kbcheck/proof_execution_test.go
      - docs/context/operations/testing.md
      - .github/skills/kb-work/SKILL.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "repair slice-004 external blockers, then run one fresh core and one local-release"
slices:
  - id: slice-001
    title: "Define sealed proof coverage and working-tree-aware receipts"
    path: docs/plans/2026-07-26-021-tool-proof-coverage-contract-plan.md
    blockers: []
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "Incorrect proof identity or invalidation can silently skip a changed behavior, so the contract needs conservative dependency and dirty-working-tree semantics."
    model_requirements: ["Go schema and CLI contract design", "content-addressed evidence", "working-tree and untracked-file hashing", "negative-test discipline"]
    escalation_triggers: ["coverage can be self-asserted without a sealed check registry", "unknown inputs can produce reuse", "dirty relevant files are omitted from fingerprints", "a suite exit code claims unenumerated child coverage"]
    context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-001.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'ProofCoverage|ProofReceipt|RelevantInputFingerprint' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/kbcheck/proof_governor_test.go
        role: "false-reuse rejection and working-tree fingerprint oracle"
        sha256: "5aeef634af709f68ccfda6d27556a287eef7c9a79240f5753c25b7f0597811a4"
        oracle_update_reason: "Explicit plan amendment after first green to add the missing public CLI-dispatch acceptance oracle."
        update_policy: "requires explicit plan amendment"
    notes: "DDR route: current orchestrator; router-unavailable=binary-not-found; proof=go test ./cmd/kbcheck -run Proof -count=1 -timeout 45s PASS; protected-oracle-sha256=5aeef634af709f68ccfda6d27556a287eef7c9a79240f5753c25b7f0597811a4; oracle amendment added the missing public CLI-dispatch test after first green; scope-check forecast=5 changed=5 discovered=5 unexplained=0; qa-browser skipped no UI-reachable behavior; snapshot deferred until governed selector exists because legacy all-snapshot replay is the defect under repair; memory-impact=durable refresh=pending"
  - id: slice-002
    title: "Select only invalidated proof and reuse passing supersets"
    path: docs/plans/2026-07-26-022-tool-impact-aware-proof-selection-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "Coverage subsumption and change impact are the core soundness boundary; a false reuse is worse than a conservative rerun."
    model_requirements: ["set-subsumption reasoning", "change-impact explanation", "deterministic CLI decisions", "conservative invalidation fallbacks"]
    escalation_triggers: ["one changed input does not invalidate every dependent check", "the decision lacks exact invalidating paths", "overlapping checks run twice against one fingerprint", "unknown dependency metadata is treated as unchanged"]
    context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-002.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'ProofSelection|CoverageSubsumption|ImpactInvalidation' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/kbcheck/proof_selection_test.go
        role: "changed-input invalidation and passing-superset reuse oracle"
        sha256: "79fe938cefa3070f9252e7e29573ac24a30e0023e0abfbb01ea2435d888ad439"
        update_policy: "requires explicit plan amendment"
    notes: "DDR route: current orchestrator; router-unavailable=binary-not-found; selector and public CLI proof PASS; protected-oracle-sha256=79fe938cefa3070f9252e7e29573ac24a30e0023e0abfbb01ea2435d888ad439; scope-check forecast=4 changed=4 discovered=4 unexplained=0; adjacent Windows path normalization recorded; qa-browser skipped no UI-reachable behavior; snapshot deferred until governed execution exists; memory-impact=durable refresh=pending"
  - id: slice-003
    title: "Enforce replay budgets and automatic GUI pre-launch denial"
    path: docs/plans/2026-07-26-023-tool-proof-execution-and-gui-guard-plan.md
    blockers: [slice-002]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "This slice controls process launches and automatic GUI denial, where advisory wording cannot prevent disruptive loops."
    model_requirements: ["process execution policy", "automatic GUI launch prevention", "PowerShell and Go integration", "cross-skill workflow consistency"]
    escalation_triggers: ["a visible or native GUI can launch from automatic proof", "the same full suite can rerun unchanged without an override", "raw snapshot verify bypasses selection", "repair still requires every unrelated check after each fix"]
    context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-003.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'ProofExecutionBudget|AttendedGUI|SnapshotSelection|RepairPolicy' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: false
    protected_oracles:
      - path: cmd/kbcheck/proof_execution_test.go
        role: "historical redundant replay refusal and GUI denial oracle; approval behavior superseded by slice-005"
        sha256: "de181281c6cc6f50e5c4b1dc085d14f2eb8ba40ad74ea185064a361c2ee95cea"
        oracle_update_reason: "Completed the oracle before final protection with single-use approval and bounded audit assertions required by acceptance criteria 5 and 6."
        update_policy: "requires explicit plan amendment"
    notes: "DDR route: current orchestrator; router-unavailable=binary-not-found; focused execution, GUI denial, workflow, and manifest proof PASS; empty snapshot milestone PASS then REUSE; protected-oracle-sha256=de181281c6cc6f50e5c4b1dc085d14f2eb8ba40ad74ea185064a361c2ee95cea; no visible/native GUI launched; scope-check forecast=12 changed=12 adjacent public CLI/parser and workflow state explained; memory-impact=durable refresh=pending"
  - id: slice-004
    title: "Prove, document, and release the proof governor"
    path: docs/plans/2026-07-26-024-eval-proof-governor-release-plan.md
    blockers: [slice-003]
    verification: integration
    test_level: full
    functional_risk: full
    execution_class: cli
    model_tier: large
    model_tier_reason: "The completion claim spans CLI behavior, snapshot execution, workflow skills, browser/native safety, documentation, and global skill propagation."
    model_requirements: ["end-to-end deterministic fixtures", "Windows process checks", "skill contract review", "release and sync verification"]
    escalation_triggers: ["the selftest launches a real GUI", "core or local-release remains unresponsive", "global skill drift contains useful target-only changes", "a fixture can pass after skipping changed coverage"]
    context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-004.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck proof-governor-selftest"
      expect: 0
    hitl: false
    status: in_progress
    owner: agent
    blocked_reason: "Core exposed twelve unchanged cmd/kbrouter canonical-project-path fixture failures; full sync report also has unrelated kb-gate/safe-shell-quoting drift. Local-release was not launched because it composes the same failing gates."
    resume_when: "The cmd/kbrouter fixture failure and unrelated required global sync drift are repaired with focused proof."
    next_agent_action: "After that external blocker changes, run one fresh core and then one local-release. If both pass, close the slice gate and continue to kb-finalize."
    human_action: ""
    can_continue_other_slices: true
    protected_oracles:
      - path: evals/proof-governor/fixtures.json
        role: "end-to-end changed-input, superset-reuse, loop, and GUI refusal corpus"
        sha256: "7a5df23f89ec255c7d75fdd05d8750e75292a4366a466e93bdeb7c01dbdb9ece"
        update_policy: "requires explicit plan amendment"
    notes: "DDR route: current orchestrator; router-unavailable=binary-not-found; protected corpus RED unknown-command then GREEN 12 scenarios gui_launches=0; focused Proof/ManifestContract/ReleaseProfile and discovery proof PASS; skill-lint 0 errors; drift reviewed and 9 proof-governor skills synced hash-identically to 3 targets; full sync report retains unrelated kb-gate/safe-shell-quoting drift; git diff --check PASS; final core ran once and failed on one owned discovery expectation plus twelve pre-existing cmd/kbrouter canonical-project-path failures; owned discovery failure repaired with focused proof; core not replayed and local-release not launched to avoid identical reproof; no visible/native GUI launched; memory-impact=durable refresh=pending"
  - id: slice-005
    title: "Remove repo-owned GUI approvals and fail automatic GUI execution closed"
    path: docs/plans/2026-07-26-025-tool-gui-fail-closed-simplification-plan.md
    blockers: [slice-003]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    execution_class: cli
    model_tier: large
    model_tier_reason: "This correction removes a misleading security boundary from the public CLI while preserving the process-launch safety invariant across code, manifests, skills, and installed copies."
    model_requirements: ["Go CLI contract repair", "process-launch safety", "cross-skill policy consistency", "dirty-worktree preservation"]
    escalation_triggers: ["a visible-browser or native-gui check can reach the runner", "removing approval code weakens headless execution or receipt reuse", "a target skill contains useful target-only drift", "focused proof requires a real GUI"]
    context_packet_path: docs/plans/2026-07-26-proof-governor-context/slice-005.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'ProofExecution|ProofGovernorCLI|ManifestContract' -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: ""
    human_action: ""
    can_continue_other_slices: true
    protected_oracles:
      - path: cmd/kbcheck/proof_execution_test.go
        role: "automatic GUI execution denial before process launch"
        sha256: "80672702ed84bf6c7ffffdf706af8966df11c7b6d2031c966bdd6779673dc03e"
        oracle_update_reason: "Explicit corrective amendment supersedes the slice-003 approval-token oracle because the token was not an authenticated trust boundary and added unwanted HITL ceremony. RED proved both GUI classes still returned the superseded attended-approval reason while keeping runner count zero; the accepted oracle then added public CLI removal assertions before final protection."
        update_policy: "requires explicit plan amendment"
    notes: "DDR route: current orchestrator; router-unavailable=binary-not-found; protected RED returned attended-approval-required for both GUI classes with runner count zero; focused ProofExecution/ProofGovernorCLI/ManifestContract GREEN; proof-governor selftest PASS scenarios=12 gui_launches=0; current approval surface absent; skill-lint 0 errors; 4 edited skills hash-identical across repo plus 3 required roots; git diff --check PASS; core/local-release not replayed because the unchanged slice-004 blockers remain; no GUI launched; no stage/commit/push."
---

# KB: Change-Aware Proof Governor

## Origin

Direct conversation: retain complete testing for changed behavior, stop replaying
equivalent proof against unchanged inputs, explain why any prior result became
stale, and prevent unattended visible/native GUI test launches.

## Workflow Shape

`pipeline-change` — the correction spans proof schemas, a deterministic selector,
the snapshot runner, repair and workflow policies, execution/approval safety,
fixtures, documentation, and skill propagation.

## Decisions Carried Forward

- Soundness wins over speed: missing or ambiguous dependency metadata means run.
- A receipt can be reused only when its passing coverage is a true superset of
  the requested check IDs and all relevant content/environment fingerprints
  still match.
- Coverage comes from a sealed check registry or enumerated suite children, not
  an agent-authored claim that a command “covers everything.”
- Fingerprints include tracked, dirty, and relevant untracked inputs; commit SHA
  alone is insufficient.
- Proof identity also binds the goal/slice/run namespace, oracle and verifier
  files, command/argv, working directory, timeout, expected result, environment
  contract, and relevant external evidence. A change to any semantic field
  invalidates reuse.
- Every decision prints `RUN`, `REUSE`, or `BLOCK` plus the exact check IDs,
  receipt, relevant hashes, and changed paths responsible.
- Focused proof runs during a slice; changed-workflow smoke runs after a
  manifest; one full suite runs at a genuine integration/release milestone.
- Composition still receives one fresh aggregate proof after integration.
  Reuse removes duplicate child executions inside that proof state; it does not
  turn worker receipts into completion authority.
- A passing full-suite receipt can satisfy later subset requests until relevant
  inputs change.
- Automatic visible browser and native GUI execution is denied before process
  launch. Any explicitly attended session remains outside `proof-run`.
- Every check and aggregate run has a ceiling, child cleanup, and a terminal
  partial/failure receipt. An unbounded “verify all” result is never accepted.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Define sealed proof coverage and working-tree-aware receipts | execution precondition | tdd / functional-cli | no | done |
| 2 | Select only invalidated proof and reuse passing supersets | slice-001 | tdd / functional-cli | no | done |
| 3 | Enforce replay budgets and attended visible/native GUI launches | slice-002 | tdd / functional-cli | no | in_progress |
| 4 | Prove, document, and release the proof governor | slice-003 + external release gate | integration / full | no | pending |

## Execution Gate

Execution is authorized and serial. The pre-existing DDR announcement hunks and
provider-hygiene changes are user-owned and must remain intact. Focused
`cmd/kbcheck` package proof is responsive; the final full-release slice remains
independently blocked if `core` or `local-release` cannot finish within its
bounded diagnostic window.
