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
