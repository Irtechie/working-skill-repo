package reconcile

import (
	"bytes"
	"testing"
)

// The rehab pass must lift caution and never proof. If a future change relaxes
// a predicate here instead of a budget, this test is the thing that catches it.
func TestRehabPolicyInheritsEveryPredicateUnchanged(t *testing.T) {
	base := DefaultPolicy()
	rehab := RehabPolicy()

	if len(rehab.Actions) != len(base.Actions) {
		t.Fatalf("action count changed: base %d, rehab %d", len(base.Actions), len(rehab.Actions))
	}
	for index, baseAction := range base.Actions {
		rehabAction := rehab.Actions[index]
		if rehabAction.Class != baseAction.Class {
			t.Fatalf("action %d reordered: base %q, rehab %q", index, baseAction.Class, rehabAction.Class)
		}
		if rehabAction.Allowed != baseAction.Allowed {
			t.Errorf("%s allowed changed to %t; rehab must not widen the mutation surface",
				baseAction.Class, rehabAction.Allowed)
		}
		if rehabAction.Threshold != baseAction.Threshold {
			t.Errorf("%s threshold changed to %v", baseAction.Class, rehabAction.Threshold)
		}
		baseEncoded, err := MarshalStable(baseAction.Mandatory)
		if err != nil {
			t.Fatalf("encode base predicates: %v", err)
		}
		rehabEncoded, err := MarshalStable(rehabAction.Mandatory)
		if err != nil {
			t.Fatalf("encode rehab predicates: %v", err)
		}
		if !bytes.Equal(baseEncoded, rehabEncoded) {
			t.Errorf("%s mandatory predicates differ; rehab gets no pass on proof", baseAction.Class)
		}
	}
}

func TestRehabPolicyRaisesOnlyTheRetireBudgetsItOwns(t *testing.T) {
	base := DefaultPolicy()
	rehab := RehabPolicy()

	raised := map[string]bool{
		ActionLocalRefRetire: true, ActionWorktreeRetire: true, ActionSalvage: true,
	}
	for action, baseCap := range base.RiskBudget.PerRun {
		rehabCap := rehab.RiskBudget.PerRun[action]
		if raised[action] {
			if rehabCap != RehabRetireBudget {
				t.Errorf("per-run %s: want %d, got %d", action, RehabRetireBudget, rehabCap)
			}
			continue
		}
		if rehabCap != baseCap {
			t.Errorf("per-run %s changed from %d to %d; rehab owns only the retire classes",
				action, baseCap, rehabCap)
		}
	}
	for action, baseCap := range base.RiskBudget.PerRepository {
		rehabCap := rehab.RiskBudget.PerRepository[action]
		if raised[action] {
			if rehabCap != RehabRetireBudget {
				t.Errorf("per-repository %s: want %d, got %d", action, RehabRetireBudget, rehabCap)
			}
			continue
		}
		if rehabCap != baseCap {
			t.Errorf("per-repository %s changed from %d to %d", action, baseCap, rehabCap)
		}
	}
}

// The default ceilings are what made rehab need several runs to converge: the
// effective cap is min(PerRun, PerRepository), so three stale worktrees per run
// was the real limit regardless of the per-run value.
func TestRehabPolicyConvergesInOneRunWhereDefaultCannot(t *testing.T) {
	base := DefaultPolicy()
	rehab := RehabPolicy()

	for _, action := range []string{ActionLocalRefRetire, ActionWorktreeRetire} {
		baseEffective := effectiveCap(base, action)
		rehabEffective := effectiveCap(rehab, action)
		if baseEffective >= rehabEffective {
			t.Fatalf("%s: rehab cap %d did not exceed default cap %d",
				action, rehabEffective, baseEffective)
		}
		if baseEffective > 5 {
			t.Fatalf("%s: default cap %d is unexpectedly high; test premise is stale",
				action, baseEffective)
		}
	}
}

// A plan built with the rehab pass must not be applied under default ceilings.
// apply.go rejects a policy-version mismatch, so the distinct version is what
// makes that failure closed rather than silent.
func TestRehabPolicyVersionDiffersSoMismatchedApplyFailsClosed(t *testing.T) {
	if RehabPolicy().PolicyVersion == DefaultPolicy().PolicyVersion {
		t.Fatal("rehab policy version matches the default; a mismatched apply would silently downgrade")
	}
	if RehabPolicy().PolicyVersion != RehabPolicyVersion {
		t.Fatalf("rehab policy version %q does not match the exported constant", RehabPolicy().PolicyVersion)
	}
}

// DefaultPolicy hands out fresh maps, but if that ever changes, RehabPolicy
// would mutate shared state and quietly raise ceilings for every other caller.
func TestRehabPolicyDoesNotMutateTheDefaultPolicy(t *testing.T) {
	before := DefaultPolicy().RiskBudget.PerRun[ActionWorktreeRetire]
	_ = RehabPolicy()
	after := DefaultPolicy().RiskBudget.PerRun[ActionWorktreeRetire]
	if before != after {
		t.Fatalf("DefaultPolicy budget changed from %d to %d after building the rehab policy", before, after)
	}
}

func effectiveCap(policy Policy, action string) int {
	cap := policy.RiskBudget.PerRun[action]
	if perRepo := policy.RiskBudget.PerRepository[action]; perRepo < cap || cap == 0 {
		cap = perRepo
	}
	return cap
}
