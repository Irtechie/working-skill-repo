---
type: kb-manifest
kb_id: kb-2026-07-05-phoenix-proof-spine-merge
brainstorm_path: docs/context/research/2026-07-05-atv-phoenix-self-heal-comparison.md
created: 2026-07-05
status: completed
workflow_shape: "pipeline-change"
model_examples_checked:
  checked_at: "2026-07-05T09:09:26-04:00"
  sources:
    - "https://developers.openai.com/codex/models"
    - "https://developers.openai.com/codex/subagents"
    - "https://platform.claude.com/docs/en/about-claude/models/overview"
    - "https://platform.claude.com/docs/en/about-claude/models/choosing-a-model"
safe_assumptions:
  - "KB remains the primary workflow; Phoenix is mined for proof primitives."
  - "The first implementation target is cmd/kbcheck, not a new Phoenix skill bundle."
  - "Trace artifacts are ephemeral under .kb/trace.jsonl unless a slice promotes summary proof into a manifest."
  - "Model routing is stored as capability tiers, with provider names kept as dated examples."
model_tier_contract:
  tiny: "Deterministic transforms, grep/path inventories, table maintenance, hash comparison, status summaries."
  small: "Bounded classification, fixture/log/doc work, boilerplate, narrow patching when the oracle is fixed."
  medium: "Ordinary implementation slices with clear acceptance criteria, known files, and runnable proof."
  large: "Architecture, unclear decomposition, security/destructive risk, long-context conflicts, failed-loop diagnosis, final high-risk judgment."
  proof_rule: "No model tier is proof. Proof is executable evidence: tests, commands, browser/API/CLI probes, traces, or kbcheck accept."
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-research
    status: passed
    required_evidence:
      - "latest ATV-Phoenix repo inspected at commit fc6e3a4e537bf025be18eb5ac7ae9b98488da207"
      - "Phoenix replacement vs merge decision recorded"
      - "KB strengths and Phoenix proof primitives recorded"
      - "model-tier decomposition gap recorded"
      - "no unresolved ask-now or research-first blockers remain"
    proof:
      - docs/context/research/2026-07-05-atv-phoenix-self-heal-comparison.md
      - docs/context/architecture/kb-workflow.md
      - docs/context/architecture/kb-learning-model.md
      - .github/skills/kb-functional-test/SKILL.md
      - docs/plans/2026-07-05-000-kb-phoenix-proof-spine-merge-manifest.md
    evidence_notes:
      - "Phoenix's useful spine is sense, snapshot, heal, trace, accept, and measured adoption gate."
      - "KB is stronger in scoped learning, repo memory, route selection, and terminal completion gates."
      - "Current KB plans classify verification risk but do not yet classify model execution tier."
    blockers: []
    passed_at: "2026-07-05T09:09:26-04:00"
    allowed_next_action: "kb-plan docs/context/research/2026-07-05-atv-phoenix-self-heal-comparison.md"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest path exists"
      - "all 6 slice plan paths exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk"
      - "each slice records model_tier and escalation boundaries"
      - "HITL classification recorded"
    proof:
      - docs/plans/2026-07-05-000-kb-phoenix-proof-spine-merge-manifest.md
      - docs/plans/2026-07-05-001-harness-kbcheck-sense-trace-accept-plan.md
      - docs/plans/2026-07-05-002-skill-repair-troubleshoot-acceptance-plan.md
      - docs/plans/2026-07-05-003-goal-work-completion-proof-ledger-plan.md
      - docs/plans/2026-07-05-004-learning-measured-adoption-gate-plan.md
      - docs/plans/2026-07-05-005-model-tier-decomposition-contract-plan.md
      - docs/plans/2026-07-05-006-docs-eval-sync-cleanup-plan.md
    dag_validation: "001 unlocks 002 and 003; 004 can run after 001; 005 can run independently; 006 depends on 002,003,004,005; no missing blockers or cycles."
    blockers: []
    passed_at: "2026-07-05T09:09:26-04:00"
    allowed_next_action: "kb-work docs/plans/2026-07-05-000-kb-phoenix-proof-spine-merge-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "proof spine implementation exists"
      - "proof spine tests cover red-green, vacuous-green, tamper, digest, and timeout behavior"
      - "CLI commands are registered"
      - "manual red-green smoke passed"
      - "focused Go tests passed"
    proof:
      - cmd/kbcheck/proof_spine.go
      - cmd/kbcheck/proof_spine_test.go
      - cmd/kbcheck/main.go
      - "proof-spine-smoke ok red-green trace"
      - "go-test-cmd-kbcheck passed"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "kb-repair records RED before repair when a proof check exists"
      - "kb-troubleshoot records RED before editing and requires accept after fix when practical"
      - "kb-check documents sense, accept, trace-verify, and learning-adoption commands"
      - "skill lint passed through core"
      - "focused Go tests passed"
    proof:
      - .github/skills/kb-repair/SKILL.md
      - .github/skills/kb-troubleshoot/SKILL.md
      - .github/skills/kb-check/SKILL.md
      - "core skill-lint passed"
      - "go-test-cmd-kbcheck passed"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "kb-goal terminal proof references kbcheck accept when objective checks exist"
      - "kb-work records accept evidence for fixed known failures"
      - "kb-complete proof gate accepts RED-before-GREEN proof and rejects latest-green-only repairs"
      - "kb-gate prefers accept evidence at phase boundaries"
      - "architecture workflow docs describe proof spine and .kb snapshot path"
      - "core passed"
    proof:
      - .github/skills/kb-goal/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-complete/SKILL.md
      - .github/skills/kb-gate/SKILL.md
      - docs/context/architecture/kb-workflow.md
      - "core passed"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "learning-adoption scorer exists"
      - "tests cover positive adoption, right-to-wrong rejection, low sample rejection, and holdout leakage"
      - "learn/evolve require measured adoption before shared/global promotion"
      - "learning model docs preserve scoped-local default"
      - "manual ADOPT_ELIGIBLE smoke passed"
    proof:
      - cmd/kbcheck/learning_adoption.go
      - cmd/kbcheck/learning_adoption_test.go
      - .github/skills/learn/SKILL.md
      - .github/skills/evolve/SKILL.md
      - docs/context/architecture/kb-learning-model.md
      - "learning-adoption-smoke ADOPT_ELIGIBLE"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "kb-plan requires model_tier in new manifests"
      - "kb-work consumes model_tier without lowering proof"
      - "kb-functional-test keeps mini-model use bounded"
      - "manifest contract validates opt-in model tiers"
      - "architecture workflow docs describe model tier as delegation guidance"
      - "manifest contract passed"
    proof:
      - .github/skills/kb-plan/SKILL.md
      - .github/skills/kb-work/SKILL.md
      - .github/skills/kb-functional-test/SKILL.md
      - cmd/kbcheck/manifest_contract.go
      - cmd/kbcheck/manifest_contract_test.go
      - docs/context/architecture/kb-workflow.md
      - "manifest-contract passed"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-work continue"
  - gate_id: slice-slice-006-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "README documents visible proof-spine and adoption-gate behavior"
      - "testing docs and eval map include new commands"
      - "Phoenix research index is updated"
      - "required skill roots are synced"
      - "working repo release gate passed"
      - "ATV .github skills diff check passed"
    proof:
      - README.md
      - docs/context/operations/testing.md
      - docs/context/eval-map.md
      - docs/context/research/README.md
      - "skill-sync-report zero required issues"
      - "local-release passed"
      - "atv-diff-check passed"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-work finalize"
  - gate_id: work-to-complete
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "every non-skipped slice has a passing slice-to-done gate"
      - "final local-release passed"
      - "working git diff check passed"
      - "ATV git diff check passed"
      - "required skill roots match source"
      - "manifest and todo are synced"
    proof:
      - docs/plans/2026-07-05-000-kb-phoenix-proof-spine-merge-manifest.md
      - "local-release passed"
      - "git-diff-check passed"
      - "atv-diff-check passed"
      - "skill-sync-report zero required issues"
      - "todo updated"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-complete docs/plans/2026-07-05-000-kb-phoenix-proof-spine-merge-manifest.md"
  - gate_id: complete-to-ship
    owner_skill: kb-complete
    status: passed
    required_evidence:
      - "final kb-check/release proof passed"
      - "functional proof level is classified or skipped with reason"
      - "review mode and findings are recorded"
      - "P0/P1 findings are resolved"
      - "follow-up resolution is recorded"
      - "proof/demo evidence is recorded"
      - "compound, learn, and evolve results are recorded"
      - "project memory and maintenance results are recorded"
      - "cleanup result is recorded"
      - "alerts are recorded"
    proof:
      - "local-release passed"
      - "functional-test skipped non-ui-tooling-change"
      - "review-mode local-fallback P0=0 P1=1-resolved P2=0 P3=0"
      - "follow-up-resolution resolved 1 logged 0 blocked 0"
      - "proof evidence local-release manifest-contract proof-spine-smoke learning-adoption-smoke"
      - docs/solutions/logic-errors/proof-spine-digest-check-semantics-2026-07-05.md
      - docs/context/kb/instincts/scoped/kbcheck-proof-spine.yaml
      - docs/context/kb/kb-completions.txt
      - docs/context/memory-maintenance.md
      - "cleanup screenshots-none observations-appended alerts-none"
    blockers: []
    passed_at: "2026-07-05T09:48:41-04:00"
    allowed_next_action: "kb-ship docs/plans/2026-07-05-000-kb-phoenix-proof-spine-merge-manifest.md"
slices:
  - id: slice-001
    title: "Add kbcheck sense/trace/accept proof spine"
    path: docs/plans/2026-07-05-001-harness-kbcheck-sense-trace-accept-plan.md
    blockers: []
    verification: tdd
    test_level: integration
    functional_risk: narrow
    model_tier: medium
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Implement the CLI proof spine in cmd/kbcheck with tests for red-green acceptance, vacuous-green rejection, and trace tamper rejection."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=4 discovered=0; proof: go test ./cmd/kbcheck passed; proof-spine smoke passed RED->GREEN accept; memory-impact: durable"
    protected_oracles:
      - path: "cmd/kbcheck/proof_spine_test.go"
        role: "trace-derived acceptance oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan update"
  - id: slice-002
    title: "Wire repair/troubleshoot to failure-first acceptance"
    path: docs/plans/2026-07-05-002-skill-repair-troubleshoot-acceptance-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: integration
    functional_risk: narrow
    model_tier: medium
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Update kb-repair and kb-troubleshoot so reproduced failures become sense checks and successful fixes require kbcheck accept when a failing oracle exists."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=3 discovered=0; proof: core skill-lint passed and go test ./cmd/kbcheck passed; memory-impact: durable"
    protected_oracles:
      - path: "cmd/kbcheck/skill_validators_test.go"
        role: "skill contract regression oracle"
        sha256: "filled by kb-work if edited"
        update_policy: "requires explicit plan update when changed"
  - id: slice-003
    title: "Require proof ledger in goal/work/complete"
    path: docs/plans/2026-07-05-003-goal-work-completion-proof-ledger-plan.md
    blockers: [slice-001]
    verification: integration
    test_level: integration
    functional_risk: broad
    model_tier: large
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Thread trace-derived proof through kb-goal, kb-work, kb-complete, and kb-gate so terminal claims prefer executable accept evidence."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=7 changed=5 discovered=0; proof: core passed; policy now rejects latest-green-only repair claims; memory-impact: durable"
    protected_oracles: []
  - id: slice-004
    title: "Add measured adoption gate to learn/evolve"
    path: docs/plans/2026-07-05-004-learning-measured-adoption-gate-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: integration
    functional_risk: narrow
    model_tier: large
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add a held-out eval gate for shared/global learning promotion while preserving KB's narrow scoped-learning default."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=6 changed=5 discovered=0; proof: go test ./cmd/kbcheck passed and learning-adoption smoke returned ADOPT_ELIGIBLE; memory-impact: durable"
    protected_oracles:
      - path: "cmd/kbcheck/learning_adoption_test.go"
        role: "measured promotion acceptance oracle"
        sha256: "filled by kb-work after RED/protection"
        update_policy: "requires explicit plan update"
  - id: slice-005
    title: "Add model-tier decomposition contract"
    path: docs/plans/2026-07-05-005-model-tier-decomposition-contract-plan.md
    blockers: []
    verification: integration
    test_level: integration
    functional_risk: narrow
    model_tier: large
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Update kb-plan and kb-work so every slice records model_tier, tier_reason, and escalation rules."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=6 changed=6 discovered=0; proof: manifest-contract passed and go test ./cmd/kbcheck passed; final proof stays executable regardless of model tier"
    protected_oracles: []
  - id: slice-006
    title: "Refresh docs, eval map, release proof, and sync targets"
    path: docs/plans/2026-07-05-006-docs-eval-sync-cleanup-plan.md
    blockers: [slice-002, slice-003, slice-004, slice-005]
    verification: verification-only
    test_level: none
    functional_risk: none
    model_tier: small
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Update visible docs/eval map, run local-release, and sync approved skill changes to the required installs/repos per AGENTS.md."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=6 changed=4 discovered=required-sync; proof: local-release passed; skill-sync-report zero required issues; ATV diff-check passed; optional scaffold/plugin drift remains warning-only"
    protected_oracles: []
---

# KB: Phoenix Proof Spine Merge

## Origin

Research: `docs/context/research/2026-07-05-atv-phoenix-self-heal-comparison.md`

The latest Phoenix repo was useful, but not as a replacement. The plan keeps KB's
workflow and learning model, then ports the Phoenix-style executable proof spine
into `cmd/kbcheck`.

## What Mine Is Better At Overall

KB is better as the project operating system:

- app-local scoped learning with ancestor-only pull and recurrence-based promotion;
- richer route selection across fix, troubleshoot, brainstorm, plan, work,
  complete, ship, epic, and durable goals;
- repo-local memory: project maps, handoffs, eval maps, todo ledgers, solutions,
  and scoped instincts;
- completion discipline: review, QA, proof/demo evidence, compound, learn,
  evolve, memory refresh, and sync gates.

Phoenix is better at one proof primitive: computed failure-first acceptance from
trace state. That is what this plan adopts.

## Decomposition Answer

KB decomposition already looks at slices "to the letter" for verticality,
dependencies, verification mode, `test_level`, `functional_risk`, HITL, and
expected files. It does not yet encode which model tier can safely run each
slice. Slice 005 adds that.

Model tiers are capability tiers, not permanent provider names:

| Tier | Use For | Current Examples Checked 2026-07-05 |
|---|---|---|
| `large` | unclear decomposition, architecture, security/destructive risk, long-context conflict resolution, final P0/P1 judgment | `gpt-5.5`, Claude Fable/Opus-class |
| `medium` | ordinary implementation with clear acceptance criteria and runnable proof | Claude Sonnet-class, strong coding model equivalents |
| `small` | bounded classification, fixtures, boilerplate, docs, log summaries, narrow patches with fixed oracle | `gpt-5.4-mini`, Claude Haiku-class |
| `tiny` | deterministic transforms, grep inventories, table updates, hash/status summaries | fastest local/tiny model or no LLM |

If a provider calls a model `ds4` or similar, the contract should classify it by
measured capability in this repo, not by marketing name.

## Slice Overview

| # | Slice | Blocked By | Verification | Model | HITL | Status |
|---|---|---|---|---|---|---|
| 001 | kbcheck sense/trace/accept proof spine | - | tdd/integration | medium | no | done |
| 002 | repair/troubleshoot failure-first acceptance | 001 | integration | medium | no | done |
| 003 | goal/work/complete proof ledger | 001 | integration | large | no | done |
| 004 | measured adoption gate for learn/evolve | 001 | tdd/integration | large | no | done |
| 005 | model-tier decomposition contract | - | integration | large | no | done |
| 006 | docs/eval/sync cleanup | 002,003,004,005 | verification-only | small | no | done |

## Work Boundary

Do not import Phoenix's full skill pack. Do not replace KB's scoped learning
store. Do not add an automatic global-learning path. Shared/global promotion must
earn it through recurrence or measured adoption proof.

## Completion Evidence

- Review: `review-mode: local-fallback`; P0=0, P1=1 resolved, P2=0, P3=0.
- Follow-up resolution: resolved 1, logged 0, blocked 0.
- Proof/demo evidence: `go test ./cmd/kbcheck`, proof-spine RED->GREEN smoke,
  learning-adoption ADOPT_ELIGIBLE smoke, `go run ./cmd/kbcheck
  manifest-contract`, `go run ./cmd/kbcheck local-release`, working
  `git diff --check`, and ATV `git diff --check` passed.
- Compound: wrote
  `docs/solutions/logic-errors/proof-spine-digest-check-semantics-2026-07-05.md`.
- Learn: added scoped instinct
  `docs/context/kb/instincts/scoped/kbcheck-proof-spine.yaml`.
- Evolve: skipped; completion counter moved from 11 to 12.
- Project memory: refreshed `docs/context/operations/testing.md`,
  `docs/context/eval-map.md`, `docs/context/architecture/kb-workflow.md`,
  `docs/context/architecture/kb-learning-model.md`, and
  `docs/context/memory-maintenance.md`.
- Cleanup: no QA screenshots; `.kb/observations.jsonl` appended with the
  resolved P1 review finding.
- Alerts: none.
