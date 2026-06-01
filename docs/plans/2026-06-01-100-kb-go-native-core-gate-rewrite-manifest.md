---
type: kb-manifest
kb_id: kb-2026-06-01-go-native-core-gate-rewrite
brainstorm_path: skipped-clear
created: 2026-06-01
status: completed
workflow_shape: "pipeline-change"
slices:
  - id: slice-101
    title: "Define native Go gate parity contract"
    path: docs/plans/2026-06-01-101-tool-go-gate-parity-contract-plan.md
    blockers: []
    verification: integration
    test_level: functional-cli
    functional_risk: medium
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Extract current PS1 gate behavior into Go-testable parity cases before rewriting behavior."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=4 discovered=1 unexplained=0; scope-discovery: cmd/kbcheck/checks_test.go - parity contract belongs with check discovery tests; protected-oracle-sha cmd/kbcheck/main_test.go=363C17BED332CBBB16E952F8410C37EEF37B250D45FB0D31C95574609F7477B8; proof: go test ./... exit=0; memory-impact: durable; areas=testing"
    protected_oracles:
      - path: "cmd/kbcheck/main_test.go"
        role: "existing wrapper behavior oracle"
        sha256: "363C17BED332CBBB16E952F8410C37EEF37B250D45FB0D31C95574609F7477B8"
        update_policy: "requires explicit plan update"
  - id: slice-102
    title: "Implement native Go kb-check core runner"
    path: docs/plans/2026-06-01-102-tool-go-native-kb-check-plan.md
    blockers: [slice-101]
    verification: integration
    test_level: functional-cli
    functional_risk: medium
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Replace the core command's PowerShell delegation with native Go check discovery/execution."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=6 discovered=1 unexplained=0; scope-discovery: cmd/kbcheck/release.go - shared CheckResult/process plumbing used by core and release; proof: go run .\\cmd\\kbcheck core --list exit=0; proof: go run .\\cmd\\kbcheck core exit=0; memory-impact: durable; areas=testing"
    protected_oracles: []
  - id: slice-103
    title: "Implement native Go release gate"
    path: docs/plans/2026-06-01-103-tool-go-native-release-gate-plan.md
    blockers: [slice-102]
    verification: integration
    test_level: functional-cli
    functional_risk: medium
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Move local-release/live-release orchestration to native Go while preserving explicit live boundaries."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=5 changed=4 discovered=1 unexplained=0; scope-discovery: scripts/kb-release-gate-selftest.ps1 - release selftest now proves Go release path; proof: go test ./... exit=0; proof: go run .\\cmd\\kbcheck local-release --json exit=0; memory-impact: durable; areas=testing,release"
    protected_oracles: []
  - id: slice-104
    title: "Prove PS1 and Go parity on Windows"
    path: docs/plans/2026-06-01-104-tool-go-ps1-parity-proof-plan.md
    blockers: [slice-103]
    verification: integration
    test_level: functional-cli
    functional_risk: medium
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Run PS1 and Go gates side by side and persist a parity report before any PS1 removal."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=3 changed=3 discovered=0 unexplained=0; proof: powershell -NoProfile -ExecutionPolicy Bypass -File scripts\\go-ps1-parity-report.ps1 exit=0; report: docs/reports/go-gate-parity-2026-06-01.md Result=PASS; memory-impact: durable; areas=testing"
    protected_oracles: []
  - id: slice-105
    title: "Remove PS1 gate dependency after proof"
    path: docs/plans/2026-06-01-105-tool-remove-ps1-gate-dependency-plan.md
    blockers: [slice-104]
    verification: integration
    test_level: functional-cli
    functional_risk: broad
    hitl: false
    status: done
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Only after parity proof passes, remove or demote PS1 gate entrypoints and update sync/docs."
    human_action: ""
    can_continue_other_slices: true
    notes: "scope-check: forecast=6 changed=9 discovered=3 unexplained=0; scope-discovery: AGENTS.md - canonical gate changed; scope-discovery: docs/context/PROJECT.md and docs/context/eval-map.md - current memory command references changed; scope-discovery: config/pipelines/skill-bundle-proof-spike.json - protected check-runner path changed; proof: go run .\\cmd\\kbcheck local-release --json exit=0; proof: scripts\\skill-sync-report.ps1 exit=0 with 36 matches per target; memory-impact: durable; areas=testing,release,sync"
    protected_oracles: []
---

# KB: Go Native Core Gate Rewrite

## Origin

Epic: `docs/context/epics/cold-storage-follow-through.md`

Human decision: full non-PowerShell rewrite if it works for Windows+; remove
PS1 after proof.

## Workflow Shape

`pipeline-change` - this changes the core proof gate, release gate, docs, skill
sync behavior, and cross-runtime portability claims.

## Slice Overview

| # | Slice | Blocked By | Verification | HITL | Status |
|---|---|---|---|---|---|
| 1 | Define native Go gate parity contract | - | integration | no | done |
| 2 | Implement native Go kb-check core runner | slice-101 | integration | no | done |
| 3 | Implement native Go release gate | slice-102 | integration | no | done |
| 4 | Prove PS1 and Go parity on Windows | slice-103 | integration | no | done |
| 5 | Remove PS1 gate dependency after proof | slice-104 | integration | no | done |

## Non-Negotiables

- The old top-level PS1 wrappers remain until Go parity proof passes on Windows.
- Go owns top-level `core` and `local-release` orchestration. Individual
  PowerShell validator scripts may remain until each validator is separately
  ported.
- `live-release` keeps live model calls explicit.
- If Go parity fails, stop before PS1 removal and keep the existing PS1 gate.

## Final Verification Target

- `go test ./...`
- Go native `core`
- Go native `local-release`
- Old PS1 `kb-check -All` and `kb-release-gate.ps1 -Profile local-release`
  in the parity report before removal
- Parity report showing matched pass/fail/check lists
- `scripts\skill-sync-report.ps1`
- `git diff --check`
