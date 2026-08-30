package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/reconcile"
)

// This file is the protected oracle for slice-003. Every test here is a
// negative security case except the granted happy path: the point is to prove
// that self-supplied proof, absent adapters, unenumerated pairings, protected
// paths, and unowned refs can never reach a merge.

const rehabTestRunID = "run-slice-003"

func rehabTestCutoff() time.Time {
	return time.Date(2026, 8, 26, 12, 0, 0, 0, time.UTC)
}

// rehabAllowingPolicy is the only way merge becomes structurally reachable.
// The shipped DefaultPolicy sets ActionMerge Allowed=false with a zero budget,
// so a grant alone can never authorize a merge.
func rehabAllowingPolicy() reconcile.Policy {
	policy := reconcile.DefaultPolicy()
	for index := range policy.Actions {
		if policy.Actions[index].Class == reconcile.ActionMerge {
			policy.Actions[index].Allowed = true
		}
	}
	policy.RiskBudget.PerRun[reconcile.ActionMerge] = 2
	return policy
}

func rehabTestReport(pairings ...workRealityPairing) workRealityReport {
	return workRealityReport{
		Status:          workRealityStatusOK,
		Action:          workRealityActionReport,
		Cutoff:          rehabTestCutoff().Format(time.RFC3339),
		RemoteAuthority: workRealityRemoteAuthority{State: "authoritative", SHA: "deadbeef", DefaultBranch: "main"},
		Pairings:        pairings,
	}
}

func rehabUnshippedPairing(id, ref, sha string, protected ...string) workRealityPairing {
	return workRealityPairing{
		ID:             id,
		DeclaredID:     "todo-001",
		DeclaredSource: "todo.md",
		Ref:            ref,
		SHA:            sha,
		State:          workRealityStateUnshipped,
		Contained:      "false",
		ProtectedPaths: protected,
	}
}

func rehabValidGrant(pairings ...rehabGrantPairing) *rehabGrant {
	issued := rehabTestCutoff()
	return &rehabGrant{
		SchemaVersion:   1,
		RunID:           rehabTestRunID,
		Operator:        "operator@example.invalid",
		IssuedAt:        issued.Format(time.RFC3339),
		ExpiresAt:       issued.Add(5 * time.Minute).Format(time.RFC3339),
		EvidenceCutoff:  issued.Format(time.RFC3339),
		Caps:            map[string]int{"merge": 2},
		OwnerIdentities: []string{"operator@example.invalid"},
		Pairings:        pairings,
	}
}

// rehabDeliveryRoot builds a repository whose authoritative default tree holds
// the real gate, while the checked-out branch ships a hostile one.
func rehabDeliveryRoot(t *testing.T, branchGate string) (string, string) {
	t.Helper()
	fixture := newWorkRealityFixture(t)
	root := fixture.Root
	defaultSHA := runGitForWorkReality(t, root, "rev-parse", "HEAD")

	if branchGate != "" {
		runGitForWorkReality(t, root, "checkout", "--quiet", "-b", "codex/hostile")
		policy := strings.Replace(
			readFileForRehab(t, filepath.Join(root, "config", "rehab-policy.json")),
			"go run ./cmd/kbcheck local-release", branchGate, 1)
		writeWorkRealityFile(t, root, "config/rehab-policy.json", policy)
		runGitForWorkReality(t, root, "add", "-A")
		runGitForWorkReality(t, root, "commit", "--quiet", "-m", "hostile gate")
	}
	return root, defaultSHA
}

func readFileForRehab(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func rehabDecision(t *testing.T, receipt rehabDeliveryReceipt, id string) rehabDeliveryDecision {
	t.Helper()
	for _, decision := range receipt.Decisions {
		if decision.PairingID == id {
			return decision
		}
	}
	t.Fatalf("decision %q not found in %+v", id, receipt.Decisions)
	return rehabDeliveryDecision{}
}

func rehabFullAdapters() rehabAdapters {
	return rehabAdapters{Forge: "github", Reviews: "github-reviews"}
}

func rehabInput(t *testing.T, root string, report workRealityReport, grant *rehabGrant) rehabDeliveryInput {
	t.Helper()
	report.RemoteAuthority.SHA = runGitForWorkReality(t, root, "rev-parse", "main")
	return rehabDeliveryInput{
		Root:        root,
		Report:      report,
		Grant:       grant,
		RunID:       rehabTestRunID,
		Now:         rehabTestCutoff().Add(time.Minute),
		PlanTTL:     10 * time.Minute,
		Adapters:    rehabFullAdapters(),
		Ownership:   map[string]string{"refs/heads/codex/owned": "operator@example.invalid"},
		MergePolicy: rehabAllowingPolicy(),
		Execute:     func(string) error { return nil },
	}
}

func TestRehabDeliveryConsultsExactlyTheShippedMergePredicateSet(t *testing.T) {
	policy := reconcile.DefaultPolicy()
	consulted := rehabMandatoryMergePredicates(policy)
	if len(consulted) == 0 {
		t.Fatalf("expected the shipped ActionMerge mandatory set")
	}

	var action reconcile.ActionPolicy
	for _, candidate := range policy.Actions {
		if candidate.Class == reconcile.ActionMerge {
			action = candidate
		}
	}
	if len(consulted) != len(action.Mandatory) {
		t.Fatalf("consulted %d predicates but ActionMerge declares %d; policy.go gained a predicate this path does not consult",
			len(consulted), len(action.Mandatory))
	}

	declared := map[string]bool{}
	for _, predicate := range action.Mandatory {
		declared[predicate.Name] = true
	}
	for _, name := range consulted {
		if !declared[name] {
			t.Fatalf("consulted predicate %q is not in the shipped manifest; this path restated the list", name)
		}
	}

	// Every consulted predicate must have an evaluation branch. An unknown
	// name falls through to unsatisfied, which would silently block forever.
	input := rehabDeliveryInput{
		Report:   rehabTestReport(),
		Adapters: rehabAdapters{Forge: "f", Checks: "c", Reviews: "r"},
	}
	receipt := rehabDeliveryReceipt{
		GrantState: rehabGrantValid, MergeCap: 1,
		ConsultedPredicates: consulted,
		ProofCommandSource:  "authoritative-default-tree", ResolvedProofCommand: "true",
	}
	pairing := rehabUnshippedPairing("p", "refs/heads/x", "abc")
	for _, name := range consulted {
		if !rehabPredicateHolds(name, input, receipt, pairing) {
			t.Fatalf("mandatory predicate %q has no satisfiable evaluation branch", name)
		}
	}
}

func TestRehabDeliveryShippedPolicyMakesMergeStructurallyUnreachable(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	report := rehabTestReport(rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc"))
	input := rehabInput(t, root, report, rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1",
	}))
	input.MergePolicy = reconcile.DefaultPolicy()

	receipt := evaluateRehabDelivery(input)
	if receipt.MergeCapCeiling != 0 {
		t.Fatalf("the shipped policy forbids merge; ceiling must be 0, got %d", receipt.MergeCapCeiling)
	}
	if receipt.MergeCap != 0 {
		t.Fatalf("a grant may never raise a cap above the shipped ceiling, got %d", receipt.MergeCap)
	}
	decision := rehabDecision(t, receipt, "branch:codex/owned")
	if decision.Action == rehabActionMergeEligible {
		t.Fatalf("a grant must never enable a disallowed action: %+v", decision)
	}
}

func TestRehabDeliveryNeverExecutesABranchSuppliedProofCommand(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "exit 0")
	report := rehabTestReport(rehabUnshippedPairing("branch:codex/hostile", "refs/heads/codex/hostile", "abc"))

	executed := []string{}
	input := rehabInput(t, root, report, rehabValidGrant())
	input.Execute = func(command string) error {
		executed = append(executed, command)
		return nil
	}

	receipt := evaluateRehabDelivery(input)
	for _, command := range executed {
		if strings.Contains(command, "exit 0") {
			t.Fatalf("a branch-supplied proof command was executed: %q", command)
		}
	}
	if strings.Contains(receipt.ResolvedProofCommand, "exit 0") {
		t.Fatalf("the resolved command came from the candidate branch: %q", receipt.ResolvedProofCommand)
	}
	if receipt.ProofCommandSource != "authoritative-default-tree" {
		t.Fatalf("proof must resolve from the authoritative default tree, got %q", receipt.ProofCommandSource)
	}
}

func TestRehabDeliveryAdapterAbsenceIsABlockerNotAPass(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	report := rehabTestReport(rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc"))

	input := rehabInput(t, root, report, rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1",
	}))
	input.Adapters = rehabAdapters{}

	receipt := evaluateRehabDelivery(input)
	decision := rehabDecision(t, receipt, "branch:codex/owned")
	if decision.Action == rehabActionMergeEligible {
		t.Fatalf("absent adapters must block merge: %+v", decision)
	}
	for _, name := range []string{"exact-pr-head-base", "fresh-mergeability", "required-reviews-green"} {
		if !rehabContains(decision.Unsatisfied, name) {
			t.Fatalf("expected %q unsatisfied with no forge adapter, got %v", name, decision.Unsatisfied)
		}
	}
}

func TestRehabDeliveryNativeGateSatisfiesChecksOnlyForAnOwnedPairing(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	owned := rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc")
	unowned := rehabUnshippedPairing("branch:codex/unowned", "refs/heads/codex/unowned", "def")
	report := rehabTestReport(owned, unowned)

	executed := []string{}
	input := rehabInput(t, root, report, rehabValidGrant(
		rehabGrantPairing{Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1"},
		rehabGrantPairing{Ref: "refs/heads/codex/unowned", TipSHA: "def", PullRequest: "2"},
	))
	input.Execute = func(command string) error {
		executed = append(executed, command)
		return nil
	}

	receipt := evaluateRehabDelivery(input)
	ownedDecision := rehabDecision(t, receipt, "branch:codex/owned")
	if !strings.HasPrefix(ownedDecision.CheckAdapter, "native:") {
		t.Fatalf("expected the native gate to be named in the receipt, got %q", ownedDecision.CheckAdapter)
	}
	if rehabContains(ownedDecision.Unsatisfied, "required-checks-green") {
		t.Fatalf("the native gate must satisfy required-checks-green for an owned pairing: %v", ownedDecision.Unsatisfied)
	}

	unownedDecision := rehabDecision(t, receipt, "branch:codex/unowned")
	if unownedDecision.Action != rehabActionReportOnly {
		t.Fatalf("an unattributed tip must be reported only, got %+v", unownedDecision)
	}
	if unownedDecision.CheckAdapter != "" {
		t.Fatalf("an unowned ref must reach no predicate evaluation, got %q", unownedDecision.CheckAdapter)
	}
	if len(executed) > 1 {
		t.Fatalf("the gate must run once per run, not once per ref: %v", executed)
	}
}

func TestRehabDeliveryFailedNativeGateBlocksMerge(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	report := rehabTestReport(rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc"))

	input := rehabInput(t, root, report, rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1",
	}))
	input.Execute = func(string) error { return errors.New("gate failed") }

	receipt := evaluateRehabDelivery(input)
	if receipt.NativeGateResult != "failed" {
		t.Fatalf("expected a recorded gate failure, got %q", receipt.NativeGateResult)
	}
	decision := rehabDecision(t, receipt, "branch:codex/owned")
	if decision.Action == rehabActionMergeEligible {
		t.Fatalf("a failed gate must block merge: %+v", decision)
	}
}

func TestRehabDeliveryProtectedPathStopsAtOpenPRUnderAValidGrant(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	report := rehabTestReport(rehabUnshippedPairing(
		"branch:codex/owned", "refs/heads/codex/owned", "abc", ".github/skills"))

	input := rehabInput(t, root, report, rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1",
	}))

	receipt := evaluateRehabDelivery(input)
	decision := rehabDecision(t, receipt, "branch:codex/owned")
	if decision.Action != rehabActionDeliverPR {
		t.Fatalf("a protected path must reach an open PR and nothing more, got %+v", decision)
	}
	if !strings.Contains(decision.Reason, ".github/skills") {
		t.Fatalf("the refusal must name the protected path, got %q", decision.Reason)
	}
}

func TestRehabDeliveryVoidsExpiredReplayedAndDriftedGrants(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	report := rehabTestReport(rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc"))

	cases := []struct {
		name    string
		mutate  func(*rehabGrant)
		runID   string
		fragmnt string
	}{
		{"expired", func(g *rehabGrant) {
			g.ExpiresAt = rehabTestCutoff().Add(-time.Minute).Format(time.RFC3339)
		}, rehabTestRunID, "expired"},
		{"replayed", func(g *rehabGrant) { g.RunID = "run-earlier" }, rehabTestRunID, "replayed"},
		{"ttl-exceeded", func(g *rehabGrant) {
			g.ExpiresAt = rehabTestCutoff().Add(90 * time.Minute).Format(time.RFC3339)
		}, rehabTestRunID, "plan TTL"},
		{"cutoff-drift", func(g *rehabGrant) {
			g.EvidenceCutoff = rehabTestCutoff().Add(time.Hour).Format(time.RFC3339)
		}, rehabTestRunID, "evidence cutoff"},
		{"no-pairings", func(g *rehabGrant) { g.Pairings = nil }, rehabTestRunID, "enumerates no pairing"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			grant := rehabValidGrant(rehabGrantPairing{
				Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1",
			})
			testCase.mutate(grant)
			receipt := evaluateRehabDelivery(rehabInput(t, root, report, grant))
			if receipt.GrantState != rehabGrantVoid {
				t.Fatalf("expected a void grant, got %q", receipt.GrantState)
			}
			if !strings.Contains(receipt.GrantVoidReason, testCase.fragmnt) {
				t.Fatalf("expected a reason mentioning %q, got %q", testCase.fragmnt, receipt.GrantVoidReason)
			}
			for _, decision := range receipt.Decisions {
				if decision.Action == rehabActionMergeEligible {
					t.Fatalf("a void grant must produce zero merges: %+v", decision)
				}
			}
		})
	}
}

func TestRehabDeliveryRefusesADriftedTipAndAnUnenumeratedPairing(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	drifted := rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc")
	report := rehabTestReport(drifted)

	moved := rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/owned", TipSHA: "moved-since-the-grant", PullRequest: "1",
	})
	receipt := evaluateRehabDelivery(rehabInput(t, root, report, moved))
	decision := rehabDecision(t, receipt, "branch:codex/owned")
	if decision.Action == rehabActionMergeEligible {
		t.Fatalf("a drifted tip must never merge: %+v", decision)
	}
	if !strings.Contains(decision.Reason, "tip SHA") {
		t.Fatalf("the refusal must name the tip drift, got %q", decision.Reason)
	}

	other := rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/somewhere-else", TipSHA: "abc", PullRequest: "9",
	})
	receipt = evaluateRehabDelivery(rehabInput(t, root, report, other))
	decision = rehabDecision(t, receipt, "branch:codex/owned")
	if decision.Action == rehabActionMergeEligible {
		t.Fatalf("an unenumerated pairing must never merge: %+v", decision)
	}
	if !strings.Contains(decision.Reason, "does not enumerate") {
		t.Fatalf("the refusal must say the grant does not enumerate the ref, got %q", decision.Reason)
	}
}

func TestRehabDeliveryNeverDeliversANonUnshippedState(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	states := []string{
		workRealityStateDead, workRealityStateSuperseded, workRealityStateLive,
		workRealityStateOrphanWork, workRealityStateOrphanBranch, workRealityStateHumanRequired,
	}
	pairings := []workRealityPairing{}
	grantPairings := []rehabGrantPairing{}
	for index, state := range states {
		ref := "refs/heads/codex/state-" + state
		pairings = append(pairings, workRealityPairing{
			ID: "branch:" + state, DeclaredID: "todo-001", DeclaredSource: "todo.md",
			Ref: ref, SHA: "sha" + string(rune('a'+index)), State: state, Contained: "unknown",
		})
		grantPairings = append(grantPairings, rehabGrantPairing{
			Ref: ref, TipSHA: "sha" + string(rune('a'+index)), PullRequest: "1",
		})
	}

	receipt := evaluateRehabDelivery(rehabInput(t, root, rehabTestReport(pairings...), rehabValidGrant(grantPairings...)))
	for _, decision := range receipt.Decisions {
		if decision.Action != rehabActionReportOnly {
			t.Fatalf("state %s must be reported only, got %+v", decision.State, decision)
		}
		if !strings.Contains(decision.Reason, "never delivered under any grant") {
			t.Fatalf("the refusal must be state-based, got %q", decision.Reason)
		}
	}
}

func TestRehabDeliveryRecordsContendedAndMutatesNothing(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	report := rehabTestReport(rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc"))
	lockDir := filepath.Join(t.TempDir(), "locks")

	input := rehabInput(t, root, report, rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1",
	}))
	input.LockDir = lockDir
	input.LockTimeout = 200 * time.Millisecond

	first := evaluateRehabDeliveryHoldingLock(t, input)
	defer first()

	receipt := evaluateRehabDelivery(input)
	if !receipt.Contended {
		t.Fatalf("a second actor must record contended, got %+v", receipt)
	}
	if len(receipt.Decisions) != 0 {
		t.Fatalf("a contended run must decide nothing, got %+v", receipt.Decisions)
	}
}

func evaluateRehabDeliveryHoldingLock(t *testing.T, input rehabDeliveryInput) func() {
	t.Helper()
	lock, contended := rehabAcquireLock(input)
	if contended || lock == nil {
		t.Fatalf("the first actor must acquire the lock")
	}
	return func() { _ = lock.Close() }
}

func TestRehabDeliveryGrantedMergeRefusesSyncAndDelegatesDelivery(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	report := rehabTestReport(rehabUnshippedPairing("branch:codex/owned", "refs/heads/codex/owned", "abc"))

	input := rehabInput(t, root, report, rehabValidGrant(rehabGrantPairing{
		Ref: "refs/heads/codex/owned", TipSHA: "abc", PullRequest: "1",
	}))
	input.Adapters = rehabAdapters{Forge: "github", Checks: "github-actions", Reviews: "github-reviews"}

	receipt := evaluateRehabDelivery(input)
	decision := rehabDecision(t, receipt, "branch:codex/owned")
	if decision.Action != rehabActionMergeEligible {
		t.Fatalf("expected merge eligibility on the granted happy path, got %+v", decision)
	}
	if decision.Delegate != "kb-complete" {
		t.Fatalf("delivery must delegate to kb-complete, got %q", decision.Delegate)
	}
	if receipt.Sync != "Sync: not-authorized" {
		t.Fatalf("a granted merge must still refuse sync, got %q", receipt.Sync)
	}
}

func TestRehabDeliveryHonoursTheMergeCap(t *testing.T) {
	root, _ := rehabDeliveryRoot(t, "")
	pairings := []workRealityPairing{}
	grantPairings := []rehabGrantPairing{}
	for index := 0; index < 4; index++ {
		ref := "refs/heads/codex/owned-" + string(rune('a'+index))
		sha := "sha" + string(rune('a'+index))
		pairings = append(pairings, rehabUnshippedPairing("branch:"+ref, ref, sha))
		grantPairings = append(grantPairings, rehabGrantPairing{Ref: ref, TipSHA: sha, PullRequest: "1"})
	}

	grant := rehabValidGrant(grantPairings...)
	grant.Caps = map[string]int{"merge": 99}

	input := rehabInput(t, root, rehabTestReport(pairings...), grant)
	input.Adapters = rehabAdapters{Forge: "github", Checks: "github-actions", Reviews: "github-reviews"}
	input.Ownership = map[string]string{}
	for _, pairing := range grantPairings {
		input.Ownership[pairing.Ref] = "operator@example.invalid"
	}

	receipt := evaluateRehabDelivery(input)
	if receipt.MergeCap != 2 {
		t.Fatalf("a grant may raise a cap only to the shipped ceiling of 2, got %d", receipt.MergeCap)
	}
	merged := 0
	for _, decision := range receipt.Decisions {
		if decision.Action == rehabActionMergeEligible {
			merged++
		}
	}
	if merged != 2 {
		t.Fatalf("expected exactly 2 merge-eligible decisions under the cap, got %d", merged)
	}
}

func TestRehabDeliveryRejectsUnknownGrantFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "grant.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":1,"run_id":"x","smuggled":true}`), 0o600); err != nil {
		t.Fatalf("write grant: %v", err)
	}
	if _, err := loadRehabGrant(path); err == nil {
		t.Fatalf("an unknown grant field must be rejected, not ignored")
	}
}

func rehabContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
