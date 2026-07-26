---
type: kb-manifest
kb_id: kb-2026-07-26-low-cognitive-burden-communication
created: 2026-07-26
status: complete
workflow_shape: skill-bundle-change
objective_contract: true
done_check:
  kind: command_exit
  command: "go run ./cmd/kbcheck local-release"
  expect: 0
  why: "proves repo quality and required global skill copies are synchronized"
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "user intent, bro/i-have-adhd/HumanLayer prior art, and resolved Question Gate state are captured"
    proof:
      - docs/context/research/2026-07-26-low-cognitive-burden-agent-communication.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-plan user-provided communication contract"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest and slice plan exist"
      - "slice has acceptance criteria, expected files, responsibility classes, and focused proof"
      - "existing kb-compact copies were hash-identical before editing"
    proof:
      - docs/plans/2026-07-26-000-kb-low-cognitive-burden-communication-manifest.md
      - docs/plans/2026-07-26-001-skill-low-cognitive-burden-communication-plan.md
      - docs/plans/2026-07-26-002-skill-low-burden-review-artifacts-plan.md
      - docs/context/research/2026-07-26-low-cognitive-burden-agent-communication.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "kb-work docs/plans/2026-07-26-000-kb-low-cognitive-burden-communication-manifest.md"
  - gate_id: slice-slice-001-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Core communication policy and deterministic contract test pass."
      - "Skill lint, repo core, local-release, and required global sync pass."
    proof:
      - docs/results/2026-07-26-low-cognitive-burden-communication.md
      - AGENTS.md
      - .github/skills/kb-compact/SKILL.md
      - .github/skills/kb-gate/SKILL.md
      - cmd/kbcheck/communication_contract_test.go
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "slice-002"
  - gate_id: slice-slice-002-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "PR first-screen and companion-document contracts pass focused proof."
      - "Skill lint, repo core, local-release, and required global sync pass."
    proof:
      - docs/results/2026-07-26-low-cognitive-burden-communication.md
      - .github/skills/kb-ship/SKILL.md
      - docs/context/operations/low-burden-review-artifacts.md
      - cmd/kbcheck/communication_contract_test.go
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "slice-003"
  - gate_id: slice-slice-003-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Executive brief skill and deterministic generator exist."
      - "Responsibility, visual-threshold, strict-input, and golden-output tests pass."
      - "Core, local-release, required global sync, and hash verification pass."
    proof:
      - docs/results/2026-07-26-low-cognitive-burden-communication.md
      - .github/skills/kb-executive-brief/SKILL.md
      - cmd/kbbrief/main.go
      - cmd/kbbrief/main_test.go
      - cmd/kbbrief/testdata/executive-brief.golden.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "none - local delivery complete"
  - gate_id: plan-to-work-pr-review-extension
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "The reviewed public PR workbench package and its 16-test proof were inspected."
      - "The working skill bundle is the canonical installed source."
      - "PR visualization is lazy-loaded only after a PR exists."
      - "Public Pages publication remains an explicit external-state gate."
    proof:
      - docs/plans/2026-07-26-004-skill-pr-review-workbench-integration-plan.md
      - docs/plans/2026-07-26-005-validation-pr-review-comprehension-trial-plan.md
      - docs/context/operations/low-burden-review-artifacts.md
      - .github/skills/pr-review-workbench/SKILL.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "integrate and exercise pr-review-workbench"
  - gate_id: slice-slice-004-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "The PR workbench package is repo-owned, locally tested, documented, and synchronized."
      - "kb-ship invokes it only after a PR exists and only when requested or configured."
    proof:
      - .github/skills/pr-review-workbench/SKILL.md
      - cmd/kbcheck/pr_review_workbench_contract_test.go
      - docs/results/2026-07-26-pr-review-workbench-trial.md
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "slice-005"
  - gate_id: slice-slice-005-to-done
    owner_skill: kb-work
    status: passed
    required_evidence:
      - "Two real open PRs produce commit-pinned offline workbenches."
      - "Deterministic renderer proof enforces first-screen state, links, and inert content."
      - "The unavailable in-app file preview is recorded without a browser workaround."
      - "The trial records whether the view lowers comprehension burden."
    proof:
      - docs/results/2026-07-26-pr-review-workbench-trial.md
      - cmd/kbcheck/pr_review_workbench_contract_test.go
      - .kb/pr-review-workbench/agent-marketplace-public-pr-1.html
      - .kb/pr-review-workbench/skill-marketplace-pr-2.html
    blockers: []
    passed_at: "2026-07-26"
    allowed_next_action: "none - local delivery complete"
slices:
  - id: slice-001
    title: "Add low-cognitive-burden communication contract"
    path: docs/plans/2026-07-26-001-skill-low-cognitive-burden-communication-plan.md
    blockers: []
    verification: verification-only
    test_level: integration
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The policy spans ambient instructions, question gates, explicit response repair, deterministic proof, and synchronized installs."
    model_requirements: ["precise policy editing", "responsibility-boundary judgment", "deterministic Go test coverage", "cross-install hash verification"]
    escalation_triggers: ["hard/soft/no-response classes conflict with an existing approval boundary", "local-release fails outside known dirty-worktree scope", "installed copy contains newer useful drift"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-compact", "skill:kb-gate", "file:AGENTS.md", "file:README.md", "file:cmd/kbcheck/communication_contract_test.go"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run TestLowCognitiveBurdenCommunicationContract"
      expect: 0
    hitl: false
    status: done
    owner: agent
  - id: slice-002
    title: "Apply low-burden structure to PR and companion review artifacts"
    path: docs/plans/2026-07-26-002-skill-low-burden-review-artifacts-plan.md
    blockers: [slice-001]
    verification: verification-only
    test_level: integration
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "Review artifacts must preserve high-leverage human decisions while avoiding duplicated tactical history."
    model_requirements: ["HumanLayer prior-art synthesis", "PR workflow policy editing", "deterministic contract coverage", "cross-install hash verification"]
    escalation_triggers: ["PR structure hides release blockers or proof", "companion document duplicates source-of-truth state", "installed copy contains newer useful drift"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:kb-ship", "file:README.md", "file:cmd/kbcheck/communication_contract_test.go"]
    shared_resources: ["git:integration-owner", "sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run TestLowCognitiveBurdenCommunicationContract"
      expect: 0
    hitl: false
    status: done
    owner: agent
  - id: slice-003
    title: "Generate executive briefs and useful visuals from source-owned data"
    path: docs/plans/2026-07-26-003-skill-executive-brief-generator-plan.md
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
    status: done
    owner: agent
  - id: slice-004
    title: "Integrate the PR review workbench into lazy PR delivery"
    path: docs/plans/2026-07-26-004-skill-pr-review-workbench-integration-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    execution_class: cli
    model_tier: medium
    model_tier_reason: "The package crosses GitHub evidence, untrusted HTML, review mutation boundaries, skill discovery, and global synchronization."
    model_requirements: ["source-package review", "untrusted-content safety", "deterministic CLI proof", "skill sync"]
    escalation_triggers: ["the marketplace package differs materially between sources", "the renderer can execute remote content", "lazy loading changes ordinary shipping", "global sync refuses drift"]
    workspace_mode: shared-serial
    conflict_domains: ["skill:pr-review-workbench", "skill:kb-ship", "skill:kb-executive-brief", "file:README.md"]
    shared_resources: ["sync:global-skills"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run TestPRReviewWorkbenchContract -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
  - id: slice-005
    title: "Trial the workbench on real open pull requests"
    path: docs/plans/2026-07-26-005-validation-pr-review-comprehension-trial-plan.md
    blockers: [slice-004]
    verification: integration
    test_level: functional-browser
    functional_risk: narrow
    execution_class: headless-browser
    model_tier: medium
    model_tier_reason: "The trial needs source-anchored PR synthesis plus rendered-browser verification without publishing or mutating GitHub."
    model_requirements: ["GitHub PR evidence", "source-diff inspection", "browser DOM assertions", "visual comprehension judgment"]
    escalation_triggers: ["a PR head changes during collection", "source evidence cannot be inspected", "the HTML hides a blocker", "the first screen exceeds its cognitive-load budget"]
    workspace_mode: shared-serial
    conflict_domains: ["artifact:.kb/pr-review-workbench", "file:docs/results/2026-07-26-pr-review-workbench-trial.md"]
    shared_resources: ["github:read-only", "browser:local-html"]
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run TestPRReviewWorkbenchContract -count=1"
      expect: 0
    hitl: false
    status: done
    owner: agent
---

# KB: Low-Cognitive-Burden Agent Communication

Combine plain-language and action-first prior art with a responsibility test:
tell the reader whether they must respond, may respond while the agent can
handle it, or do not need to respond. Optimize comprehension and decision
effort rather than word count.

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Add low-cognitive-burden communication contract | - | verification-only | no | done |
| 2 | Apply low-burden structure to PR and companion review artifacts | slice-001 | verification-only | no | done |
| 3 | Generate executive briefs and useful visuals from source-owned data | - | verification-only | no | done |
| 4 | Integrate the PR review workbench into lazy PR delivery | - | integration | no | done |
| 5 | Trial the workbench on real open pull requests | slice-004 | integration | no | done |
