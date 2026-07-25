---
kb_id: kb-2026-07-20-harness-validation-recovery
slice_id: slice-001
title: "Restore kbcheck package-load and core-list proof"
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
expected_files:
  - path: cmd/kbcheck/main.go
    op: edit
    scope: "Repair any CLI/init/package-load behavior that prevents core --list from returning."
  - path: cmd/kbcheck/checks.go
    op: edit
    scope: "Repair check discovery if a new check blocks list mode."
  - path: cmd/kbcheck/*_test.go
    op: edit
    scope: "Tighten the smallest failing or hanging test introduced by recent kbcheck changes."
  - path: internal/graphrouting/
    op: edit
    scope: "Inspect only if graphrouting package init or tests cause the load hang."
  - path: docs/context/operations/testing.md
    op: edit
    scope: "Record the confirmed root cause and current proof command once fixed."
protected_oracles: []
status: pending
owner: agent
blocked_reason: ""
resume_when: ""
next_agent_action: "Bound go list/go test probes, isolate the package-load hang, repair the smallest owning file, and prove core --list returns."
human_action: ""
can_continue_other_slices: false
---

# Slice 001: Restore kbcheck Package-Load And Core-List Proof

## What To Build

Repair the current silent Go package-load/runtime hang so the repo's cheapest
native proof command returns normally:

```powershell
go run ./cmd/kbcheck core --list
```

## Acceptance Criteria

- `go version` remains on the supported local toolchain.
- `go list ./cmd/kbcheck` and `go run ./cmd/kbcheck core --list` complete.
- The fix does not revert unrelated dirty work.
- If a newly added test, init path, generated binary, or package dependency is
  responsible, the root cause is documented in the slice notes.
- No later slice proceeds with an unbounded Go hang.

## Test Scenarios

- Run `go list ./cmd/kbcheck` with a bounded wrapper.
- Run `go test ./cmd/kbcheck -run <focused-root-cause-test> -count=1`.
- Run `go run ./cmd/kbcheck core --list`.

## Proof Check

`go run ./cmd/kbcheck core --list`

## Scope Boundary

No global skill sync, external harness runner, or broad release fix in this
slice. Do not clean the dirty tree except for files proven to be generated and
owned by this task.

## Dependencies

None. This is the root blocker for all target-side proof.
