package main

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/Irtechie/working-skill-repo/internal/reconcile"
)

func TestReconcileTerminalCleanupPredicateContract(t *testing.T) {
	t.Parallel()
	policy, err := reconcile.LoadPolicy(filepath.Join("..", "..", "config", "reconcile-predicates.json"))
	if err != nil {
		t.Fatal(err)
	}
	if policy.PolicyVersion != terminalCleanupSafetyContractVersion {
		t.Fatalf("terminal cleanup policy=%s global policy=%s", terminalCleanupSafetyContractVersion, policy.PolicyVersion)
	}
	global := reconcile.WorktreeSafetyPredicates(policy)
	terminal := terminalCleanupSafetyPredicates()
	if !reflect.DeepEqual(global, terminal) {
		t.Fatalf("global and terminal cleanup predicates diverged:\nglobal=%v\nterminal=%v", global, terminal)
	}
}

// TestWorkRealityRemoteAuthorityContract pins the rehab survey to the same
// authority type kbreconcile plans against. These two once disagreed about
// which branches were dead; an alias makes a silent redefinition a build
// failure instead of a wrong deletion.
func TestWorkRealityRemoteAuthorityContract(t *testing.T) {
	t.Parallel()
	if reflect.TypeOf(workRealityRemoteAuthority{}) != reflect.TypeOf(reconcile.RemoteAuthority{}) {
		t.Fatalf("work-reality authority type diverged from reconcile.RemoteAuthority")
	}
	authoritative := workRealityRemoteAuthority{
		State: reconcile.RemoteAuthorityAuthoritative, DefaultBranch: "main", SHA: "deadbeef",
	}
	if !authoritative.Authoritative() {
		t.Fatalf("a resolved default branch must be authoritative")
	}
	// A state string alone must never authorize an irreversible decision.
	incomplete := workRealityRemoteAuthority{State: reconcile.RemoteAuthorityAuthoritative}
	if incomplete.Authoritative() {
		t.Fatalf("authority without a branch and sha must not be authoritative")
	}
	if (workRealityRemoteAuthority{State: reconcile.RemoteAuthorityUnavailable}).Authoritative() {
		t.Fatalf("unavailable authority must not be authoritative")
	}
}

// TestParseSymrefAdvertisementReadsAdvertisedDefault covers the parser both the
// survey and refreshRemotes now share.
func TestParseSymrefAdvertisementReadsAdvertisedDefault(t *testing.T) {
	t.Parallel()
	branch, sha := reconcile.ParseSymrefAdvertisement(
		"ref: refs/heads/main\tHEAD\n1111111111111111111111111111111111111111\tHEAD\n")
	if branch != "main" || sha != "1111111111111111111111111111111111111111" {
		t.Fatalf("branch=%q sha=%q", branch, sha)
	}
	if branch, sha := reconcile.ParseSymrefAdvertisement("garbage\n"); branch != "" || sha != "" {
		t.Fatalf("unparseable advertisement must yield nothing, got branch=%q sha=%q", branch, sha)
	}
}
