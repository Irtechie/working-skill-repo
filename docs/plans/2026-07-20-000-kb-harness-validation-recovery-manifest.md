---
type: kb-manifest
kb_id: kb-2026-07-20-harness-validation-recovery
brainstorm_path: "direct user request: /kb-plan lets get all of these things fixed"
created: 2026-07-20
status: active
workflow_shape: pipeline-change
objective_contract: true
done_check:
  kind: command_exit
  command: "powershell -NoProfile -ExecutionPolicy Bypass -File scripts/harness-engineering-review.ps1 -TargetRoot E:/working-skill-repo -HarnessRoot E:/Data/Codex/tmp/harness-engineering-lf-20260720"
  expect: 0
  why: "Proves the working skill repo native gates and the lopopolo/harness-engineering corpus checks through the repo-owned runner."
model_tier_contract:
  allowed: [small, medium, large]
  default: large
model_selection_contract:
  timing: work-time
  catalog: host-native-plus-user
  fallback: same-tier-then-higher-then-current
  automatic_downgrade: false
gate_ledger:
  - gate_id: brainstorm-to-plan
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "clear direct source exists"
      - "observed failures are enumerated"
      - "no unresolved ask-now or research-first blockers remain"
      - "safe assumptions are recorded"
    proof:
      - "2026-07-20 observed failures: kbcheck core timed out after 90s; go test ./... timed out after 90s; default Python 3.11 cannot parse harness type-alias syntax; uv unavailable; CRLF clone caused harness SHA-256 mismatch"
      - "LF-stable harness clone verified: py -3.12 sources/scripts/validate_manifest.py passed; py -3.12 sources/scripts/test_manifest.py passed"
      - "safe assumption: fix the working skill repo harness and local external-runner workflow; do not patch lopopolo/harness-engineering upstream unless a later task explicitly targets that repo"
    blockers: []
    passed_at: "2026-07-20T00:00:00Z"
    allowed_next_action: "kb-plan direct failure recovery"
  - gate_id: plan-to-work
    owner_skill: kb-plan
    status: passed
    required_evidence:
      - "manifest and four slice plans exist"
      - "DAG has no missing blockers or cycles"
      - "each slice has acceptance criteria, expected_files, verification, test_level, functional_risk, model_tier, model requirements, and escalation triggers"
      - "objective done_check exists"
      - "context packets exist with required execution fields"
    proof:
      - docs/plans/2026-07-20-000-kb-harness-validation-recovery-manifest.md
      - docs/plans/2026-07-20-001-fix-kbcheck-package-load-hang-plan.md
      - docs/plans/2026-07-20-002-add-bounded-proof-diagnostics-plan.md
      - docs/plans/2026-07-20-003-add-harness-engineering-runner-plan.md
      - docs/plans/2026-07-20-004-rerun-release-and-harness-review-plan.md
      - docs/plans/2026-07-20-harness-validation-recovery-context/slice-001.json
      - docs/plans/2026-07-20-harness-validation-recovery-context/slice-002.json
      - docs/plans/2026-07-20-harness-validation-recovery-context/slice-003.json
      - docs/plans/2026-07-20-harness-validation-recovery-context/slice-004.json
      - "context-packet validator is part of the currently hanging Go surface; packets were direct-field checked and slice 001 restores validator availability"
    blockers: []
    passed_at: "2026-07-20T00:00:00Z"
    allowed_next_action: "kb-work docs/plans/2026-07-20-000-kb-harness-validation-recovery-manifest.md"
slices:
  - id: slice-001
    title: "Restore kbcheck package-load and core-list proof"
    path: docs/plans/2026-07-20-001-fix-kbcheck-package-load-hang-plan.md
    blockers: []
    verification: functional
    test_level: functional-cli
    functional_risk: broad
    model_tier: large
    model_tier_reason: "The failure is a silent package-load/runtime hang across the repo's canonical Go proof boundary, so diagnosis can cross test, init, release, and newly added graph-routing code."
    model_requirements: ["Go package-load debugging", "Windows process inspection", "non-destructive dirty-worktree handling", "ability to bisect recent kbcheck and graphrouting changes"]
    escalation_triggers: ["go list ./cmd/kbcheck still hangs after isolating init/test causes", "a generated binary or background process is holding module files", "fix requires reverting unrelated user changes", "hang reproduces outside this repo with the same toolchain"]
    context_packet_path: docs/plans/2026-07-20-harness-validation-recovery-context/slice-001.json
    proof_check:
      kind: command_exit
      command: "go run ./cmd/kbcheck core --list"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Bound go list/go test probes, isolate the package-load hang, repair the smallest owning file, and prove core --list returns."
    human_action: ""
    can_continue_other_slices: false
    notes: "Do not claim any Go proof green until this slice passes."
  - id: slice-002
    title: "Fail stalled proof commands with useful diagnostics"
    path: docs/plans/2026-07-20-002-add-bounded-proof-diagnostics-plan.md
    blockers: [slice-001]
    verification: tdd
    test_level: functional-cli
    functional_risk: broad
    model_tier: medium
    model_tier_reason: "The core behavior is bounded command execution and actionable timeout reporting after the package-load surface is available."
    model_requirements: ["Go command runner tests", "Windows process-tree cleanup", "release-gate output design", "backward-compatible CLI flags"]
    escalation_triggers: ["timeout leaves child processes running", "core/local-release output truncates the failing command", "diagnostics require a daemon or persistent monitor", "existing release tests become flaky"]
    context_packet_path: docs/plans/2026-07-20-harness-validation-recovery-context/slice-002.json
    proof_check:
      kind: command_exit
      command: "go test ./cmd/kbcheck -run 'Timeout|Release|Core'"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Add or tighten timeout diagnostics around native proof commands so future hangs fail closed with command, duration, and next probe."
    human_action: ""
    can_continue_other_slices: false
    notes: "Preserve raw verbose output; add compact summaries only where they reduce rework."
  - id: slice-003
    title: "Make harness-engineering validation reproducible on Windows"
    path: docs/plans/2026-07-20-003-add-harness-engineering-runner-plan.md
    blockers: [slice-002]
    verification: integration
    test_level: functional-cli
    functional_risk: narrow
    model_tier: medium
    model_tier_reason: "The external failure mode is now known: Python 3.12 syntax plus CRLF-sensitive immutable hashes, both fixable with a bounded runner."
    model_requirements: ["PowerShell-safe path handling", "Git clone config", "Python launcher discovery", "clear uv fallback semantics"]
    escalation_triggers: ["Python 3.12 is unavailable", "LF-stable clone still fails manifest validation", "runner would mutate the external harness repo", "network clone fails after retries"]
    context_packet_path: docs/plans/2026-07-20-harness-validation-recovery-context/slice-003.json
    proof_check:
      kind: command_exit
      command: "powershell -NoProfile -ExecutionPolicy Bypass -File scripts/harness-engineering-review.ps1 -HarnessOnly -HarnessRoot E:/Data/Codex/tmp/harness-engineering-lf-20260720"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Create a repo-owned runner that requires or discovers Python 3.12, clones with core.autocrlf=false when needed, uses uv only when available, and records the exact harness commit."
    human_action: ""
    can_continue_other_slices: false
    notes: "The runner should report CRLF clone mismatch as a setup failure, not as an upstream harness defect."
  - id: slice-004
    title: "Rerun combined release and harness review proof"
    path: docs/plans/2026-07-20-004-rerun-release-and-harness-review-plan.md
    blockers: [slice-003]
    verification: functional
    test_level: functional-cli
    functional_risk: broad
    model_tier: medium
    model_tier_reason: "This is final integration proof across the repaired Go gate, bounded diagnostics, and external harness runner."
    model_requirements: ["release-gate interpretation", "external harness evidence capture", "project memory refresh", "dirty worktree scope accounting"]
    escalation_triggers: ["local-release fails for unrelated active work", "global skill drift requires merge before sync", "harness runner passes but repository-review claim lacks target-boundary evidence", "proof output cannot distinguish target failure from harness setup failure"]
    context_packet_path: docs/plans/2026-07-20-harness-validation-recovery-context/slice-004.json
    proof_check:
      kind: command_exit
      command: "powershell -NoProfile -ExecutionPolicy Bypass -File scripts/harness-engineering-review.ps1 -TargetRoot E:/working-skill-repo -HarnessRoot E:/Data/Codex/tmp/harness-engineering-lf-20260720"
      expect: 0
    hitl: false
    status: pending
    owner: agent
    blocked_reason: ""
    resume_when: ""
    next_agent_action: "Run the combined proof, update testing docs/results with exact commands and outcomes, and leave the manifest/todo state honest."
    human_action: ""
    can_continue_other_slices: false
    notes: "If local-release is blocked by unrelated dirty work, record the precise failing check and the smallest safe continuation."
---

# KB: Harness Validation Recovery

## Origin

Direct user request after testing `lopopolo/harness-engineering` against the
working skill repo exposed five failures:

- `go run ./cmd/kbcheck core --list` did not complete.
- `go run ./cmd/kbcheck core` timed out after 90 seconds.
- `go test ./...` timed out after 90 seconds.
- default `python` was 3.11.9 and could not parse harness Python 3.12 syntax.
- an ordinary Windows clone of `lopopolo/harness-engineering` failed immutable
  source hash validation because line endings changed.

## Workflow Shape

`pipeline-change` - the fix touches proof command behavior, release diagnostics,
external harness invocation, testing docs, and final evidence.

## Slice Overview

| # | Slice | Blocked By | Verification | Test level | Status |
|---|---|---|---|---|---|
| 1 | Restore kbcheck package-load and core-list proof | - | functional | functional-cli | pending |
| 2 | Fail stalled proof commands with useful diagnostics | slice-001 | tdd | functional-cli | pending |
| 3 | Make harness-engineering validation reproducible on Windows | slice-002 | integration | functional-cli | pending |
| 4 | Rerun combined release and harness review proof | slice-003 | functional | functional-cli | pending |

## Non-Negotiables

- Do not revert or overwrite unrelated dirty work in this checkout.
- Do not treat `harness-engineering` as a mutable upstream target.
- Clone the external harness with `core.autocrlf=false` or prove the existing
  checkout is LF-stable before trusting immutable source hashes.
- Use Python 3.12 for the external harness scripts.
- `uv` may be used when available, but the runner must have an explicit
  Python 3.12 fallback because `uv` was unavailable on this machine.
- A timed-out proof must identify the command, timeout, and next diagnostic
  step; silence is a failing behavior.

## Done

- `go run ./cmd/kbcheck core --list` returns normally.
- `go run ./cmd/kbcheck core` and `go test ./...` either pass or fail with
  bounded, actionable diagnostics rather than hanging silently.
- `scripts/harness-engineering-review.ps1` validates an LF-stable
  `lopopolo/harness-engineering` checkout under Python 3.12.
- The combined runner exits 0 against `E:/working-skill-repo` and the
  LF-stable harness clone, or records a precise remaining target failure that
  is outside this recovery scope.
