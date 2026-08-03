package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanWorktreeSelftestExercisesDisposableLifecycle(t *testing.T) {
	t.Parallel()
	realRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve real repository root: %v", err)
	}

	result, err := executePlanWorktreeSelftest(planWorktreeSelftestOptions{
		RealRepoRoot:  realRoot,
		TempParent:    t.TempDir(),
		KeepArtifacts: true,
	})
	if err != nil {
		t.Fatalf("plan worktree selftest failed: %v (artifacts: %s)", err, result.ArtifactRoot)
	}

	if len(result.Runs) != 2 {
		t.Fatalf("expected two disjoint plan runs, got %d", len(result.Runs))
	}
	if samePath(result.Runs[0].WorktreePath, result.Runs[1].WorktreePath) {
		t.Fatalf("plan runs reused one worktree: %s", result.Runs[0].WorktreePath)
	}
	for _, run := range result.Runs {
		if run.RunID == "" || run.Branch == "" || run.WorktreePath == "" {
			t.Fatalf("run receipt is incomplete: %+v", run)
		}
		if len(run.SliceCommits) != 2 {
			t.Fatalf("run %s expected two serialized slice commits, got %d", run.RunID, len(run.SliceCommits))
		}
		if !pathWithin(result.ArtifactRoot, run.WorktreePath) {
			t.Fatalf("run %s escaped disposable artifact root: %s", run.RunID, run.WorktreePath)
		}
	}

	if result.SourceHeadBefore != result.SourceHeadAfter {
		t.Fatalf("source default head changed: before=%s after=%s", result.SourceHeadBefore, result.SourceHeadAfter)
	}
	if result.SourceStatusBefore != result.SourceStatusAfter {
		t.Fatalf("source dirt changed:\nbefore=%q\nafter=%q", result.SourceStatusBefore, result.SourceStatusAfter)
	}

	requiredProofs := map[string]bool{
		"dirty recovery":                   result.DirtyBlocked,
		"stale-head recovery":              result.StaleHeadBlocked,
		"wrong-worktree recovery":          result.WrongWorktreeBlocked,
		"default policy stop before merge": result.DefaultPolicyStopBeforeMerge,
		"PR/manual stop":                   result.PRManualStopBeforeMerge,
		"real-repo firebreak":              result.RealRepoRejected,
	}
	for proof, passed := range requiredProofs {
		if !passed {
			t.Errorf("missing lifecycle proof: %s", proof)
		}
	}

	owners := strings.Join(result.CollisionOwnerEvidence, "\n")
	if !strings.Contains(owners, "run-a") {
		t.Errorf("path collision did not identify owning run: %q", owners)
	}
	if !strings.Contains(owners, "resource:browser:4110") {
		t.Errorf("resource collision did not identify owning resource: %q", owners)
	}
}

func TestPlanWorktreeSelftestRealRepoFirebreakCannotBeForced(t *testing.T) {
	t.Parallel()
	realRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve real repository root: %v", err)
	}

	for _, candidate := range []string{
		realRoot,
		filepath.Join(realRoot, ".kb", "selftest"),
		filepath.Dir(realRoot),
	} {
		if err := validatePlanWorktreeSelftestTarget(realRoot, candidate, true); err == nil {
			t.Errorf("force flag bypassed real-repo firebreak for %s", candidate)
		}
	}

	safe := filepath.Join(t.TempDir(), "disposable")
	if err := validatePlanWorktreeSelftestTarget(realRoot, safe, false); err != nil {
		t.Fatalf("safe disposable target rejected: %v", err)
	}
}
