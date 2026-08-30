# kb-rehab slice-003 proof receipt

- slice: `slice-003` (granted delivery and delegated reaping)
- manifest: `docs/plans/2026-08-26-000-kb-rehab-outstanding-work-manifest.md`
- plan: `docs/plans/2026-08-26-003-rehab-granted-delivery-plan.md`
- date: 2026-08-26
- execution_class: `cli`
- test_level: `unit`

## Protected oracle

| File | SHA256 |
|---|---|
| `cmd/kbcheck/rehab_delivery_test.go` | `46dae7fa7889f954cb1b043003a85344fd934fb9432f8a1d2c8c2b42a026497b` |

`cmd/kbcheck/work_reality_test.go` is unchanged in this slice and retains the
slice-002 hash `29d5fd11517756e463750e36342b8cb49f42680f135373846fd4d36c041be914`.

## Commands

| Command | Result |
|---|---|
| `go build ./...` | pass |
| `go vet ./...` | pass |
| `go test ./cmd/kbcheck ./internal/reconcile -run 'RehabDelivery\|RehabGrant\|WorkReality\|TerminalCleanup\|ReconcileContract' -count=1` | pass (116.9s) |
| `go test ./cmd/kbcheck -count=1` | pass |
| `go run ./cmd/kbcheck skill-lint` | 0 errors; `kb-rehab` not flagged |
| `go run ./cmd/kbcheck manifest-contract` | ok |
| `git diff --check` | clean |

All 14 slice-003 tests and 5 subtests pass. Every test except the granted happy
path is a negative security case.

## Mutation check

The anti-self-proving mechanic is the slice's P0. It was verified by mutation
rather than by assertion alone.

`resolveRehabProofCommand` was temporarily changed to read
`config/rehab-policy.json` from the working tree instead of
`git show <authoritative-default-sha>:config/rehab-policy.json`. Result:

```text
--- FAIL: TestRehabDeliveryNeverExecutesABranchSuppliedProofCommand
    rehab_delivery_test.go:220: a branch-supplied proof command was executed: "exit 0"
```

The mutation was reverted and the test returns to pass. The oracle fails for the
intended reason and is not vacuous.

## Findings

`merge-eligible` is structurally unreachable under the shipped
`reconcile.DefaultPolicy()`. `ActionMerge.Allowed` is `false` and
`RiskBudget.PerRun[ActionMerge]` is `0`, so the ceiling resolves to `0` and a
grant requesting more is clamped to `0`, leaving `explicit-merge-authority`
unsatisfied. `TestRehabDeliveryShippedPolicyMakesMergeStructurallyUnreachable`
pins this. Tests needing a granted merge construct an allowing policy fixture.

This is the honest outcome for this repository: there is no forge check adapter,
and most branches touch a protected path. Granted delivery ends at a PR.

The mandatory predicate set is consumed from `internal/reconcile/policy.go`, not
restated. `TestRehabDeliveryConsultsExactlyTheShippedMergePredicateSet` asserts
set equality with the shipped `ActionMerge` manifest and that every consulted
name has a satisfiable evaluation branch, so a new mandatory predicate blocks
this lane rather than silently passing.

## Scope

| Path | Kind |
|---|---|
| `cmd/kbcheck/rehab_delivery.go` | created |
| `cmd/kbcheck/rehab_delivery_test.go` | created (protected oracle) |
| `.github/skills/kb-rehab/references/grant.md` | created |
| `.github/skills/kb-rehab/SKILL.md` | modified (delivery eligibility section, reference entry) |
| `config/rehab-policy.json` | modified (`native_check_gate`, manifest version `rehab-1.1.0`) |
| `cmd/kbcheck/work_reality.go` | modified (`native_check_gate` policy field) |

No discovered path fell outside the forecast except
`cmd/kbcheck/work_reality.go`, which the plan's step 7 anticipated as the
carrier of the policy field.

## Delivery authority

Local commit only. This slice was not pushed, not opened as a PR, not merged,
and not synced to any global skill root.
