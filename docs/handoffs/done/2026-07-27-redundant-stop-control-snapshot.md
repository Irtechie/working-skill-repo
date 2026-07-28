# Redundant Stop-Control Snapshot

Status: complete
Created: 2026-07-27
Repository: `Irtechie/working-skill-repo`
Branch: `deaderestpool-fix-goal-stop-controls`
Branch ref: `831e794fe2273131c75d34d797393d6aa890e92d`

## Conclusion

The stop-control and Windows process-containment product work is already
integrated on `origin/main` by:

- `a495444` - `fix: make goal stops preemptive`
- `825c2b1` - `docs: consolidate workflow routing guide`

The owning session compared its pre-cleanup snapshot with current
`origin/main`. The related Copilot checkpoint is
`93acfe1c2f7f57790cef3e9f058579232d85de50`. These runtime and adapter paths
were reported byte-identical:

- `.github/skills/kb-goal/agents/openai.yaml`
- `cmd/kbrouter/dispatch_c1_windows_test.go`
- `cmd/kbrouter/dispatch_process_windows.go`

The remaining dirty skill, README, test, and workflow-doc differences are older
or weaker versions of the integrated stop protocol. They must not be committed
over current `main`.

## Reported Dirty Cleanup Boundary

- `.github/skills/kb-goal/SKILL.md`
- `.github/skills/kb-goal/agents/openai.yaml`
- `README.md`
- `cmd/kbcheck/communication_contract_test.go`
- `cmd/kbcheck/workflow_governor.go`
- `cmd/kbrouter/dispatch_c1_windows_test.go`
- `cmd/kbrouter/dispatch_process_windows.go`
- `docs/context/architecture/kb-workflow.md`

Observed before cleanup: 8 files, 97 insertions, 22 deletions.

Reproduction aid while the local checkpoint ref remains available:

```powershell
git diff --name-status 831e794 93acfe1
git diff --exit-code origin/main 93acfe1 -- .github/skills/kb-goal/agents/openai.yaml cmd/kbrouter/dispatch_c1_windows_test.go cmd/kbrouter/dispatch_process_windows.go
```

The first command includes three separately routed EDDR paths, so it is not a
standalone proof of the final eight-path worktree state. The owning session is
the source of the exact pre-cleanup count.

## Cleanup Result

The owning session discarded the eight classified duplicate paths and reported
a clean worktree. No product commit, push, or PR was needed. The branch remains
preserved.

Unrelated EDDR state is tracked separately at
`docs/handoffs/parked/2026-07-27-eddr-experimental-state.md`.
