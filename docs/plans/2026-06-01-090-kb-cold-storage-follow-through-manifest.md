---
type: kb-manifest
kb_id: kb-2026-06-01-cold-storage-follow-through
brainstorm_path: skipped-clear
created: 2026-06-01
status: completed
workflow_shape: "multi-stream-epic"
slices:
  - id: slice-091
    title: "Prove minimality candidates before deletion"
    path: docs/plans/2026-06-01-091-tool-minimality-runtime-proof-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Build a proof report for remaining cold-storage candidates before any trim/delete action."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=4 changed=3 discovered=0 unexplained=0; protected-oracle-sha scripts/skill-surface-minimality-selftest.ps1=885B40D16961928B7AE2A6E988F84DA68F7CB8F021B632EB606DFD1F48858FF7; proof: scripts\\skill-surface-minimality-selftest.ps1 exit=0; proof: scripts\\skill-surface-minimality.ps1 -Json exit=0; memory-impact: durable; areas=minimality"
    protected_oracles:
      - path: "scripts/skill-surface-minimality-selftest.ps1"
        role: "minimality classification oracle"
        sha256: "885B40D16961928B7AE2A6E988F84DA68F7CB8F021B632EB606DFD1F48858FF7"
        update_policy: "requires explicit plan update"
  - id: slice-092
    title: "Add cross-model benchmark prompt pack"
    path: docs/plans/2026-06-01-092-eval-cross-model-benchmark-prompts-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Create benchmark prompt fixtures and scoring expectations without running live models by default."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=6 changed=6 discovered=0 unexplained=0; proof: scripts\\cross-model-benchmark-validate.ps1 exit=0; proof: go run .\\cmd\\kbcheck core --list includes cross-model-benchmark-validate; memory-impact: durable; areas=evals"
    protected_oracles: []
  - id: slice-093
    title: "Add high-value Copilot path instructions"
    path: docs/plans/2026-06-01-093-doc-copilot-path-instructions-plan.md
    blockers: []
    verification: integration
    test_level: none
    functional_risk: none
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Inventory file classes and add path instructions only where repo-local rules beat global guidance."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=4 discovered=0 unexplained=0; proof: instruction applyTo selectors inspected; proof: go run .\\cmd\\kbcheck local-release --json exit=0; memory-impact: durable; areas=instructions,docs"
    protected_oracles: []
  - id: slice-094
    title: "Decide Go harness rewrite scope"
    path: docs/plans/2026-06-01-094-hitl-go-harness-rewrite-scope-plan.md
    blockers: []
    verification: hitl
    test_level: none
    functional_risk: none
    hitl: true
    status: completed
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Follow the dedicated Go rewrite manifest."
    human_action: ""
    can_continue_other_slices: true
    notes: "Answered 2026-06-01: full non-PowerShell rewrite if it works for Windows+; remove PS1 only after parity proof. Follow-up manifest: docs/plans/2026-06-01-100-kb-go-native-core-gate-rewrite-manifest.md."
    protected_oracles: []
---

# KB: Cold Storage Follow Through

## Origin

Epic: `docs/context/epics/cold-storage-follow-through.md`

## Workflow Shape

`multi-stream-epic` - remaining work spans deletion proof, benchmark assets,
Copilot instruction routing, and a separate cross-platform rewrite decision.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Prove minimality candidates before deletion | - | integration | no | done |
| 2 | Add cross-model benchmark prompt pack | - | integration | no | done |
| 3 | Add high-value Copilot path instructions | - | integration | no | done |
| 4 | Decide Go harness rewrite scope | - | hitl | yes | completed |

## Assumptions

- No deletion is approved by this manifest.
- Live model benchmark execution remains explicit and outside the local Go gate.
- Go-native gate parity passed on Windows; old top-level PS wrappers were
  removed after proof.

## Final Verification Target

- `go test ./...`
- `go run .\cmd\kbcheck core`
- `go run .\cmd\kbcheck local-release`
- `scripts\skill-sync-report.ps1`
- `git diff --check`
