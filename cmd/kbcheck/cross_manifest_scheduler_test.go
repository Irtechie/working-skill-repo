package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCrossManifestSchedulerSerializesNormalizedClaimsAndAdmitsDisjointRuns(t *testing.T) {
	t.Parallel()
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	first := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: "docs/plans/a.md",
		OwnerToken: "owner-a", Files: []string{"Cmd/Kbcheck/Main.go"},
		Prefixes: []string{"generated/output"}, Domains: []string{"skill:KB-WORK"},
		Resources: []string{"browser:4110", "git:integration-owner"}, Now: now,
	})
	if !first.OK || first.Lease == nil {
		t.Fatalf("first acquire failed: %#v", first)
	}

	tests := []struct {
		name string
		opts planRunLeaseCommandOptions
		kind string
	}{
		{
			name: "same file normalized",
			opts: planRunLeaseCommandOptions{Files: []string{"cmd\\kbcheck\\main.go"}},
			kind: "file",
		},
		{
			name: "file under prefix",
			opts: planRunLeaseCommandOptions{Files: []string{"generated/output/frame.png"}},
			kind: "file",
		},
		{
			name: "domain",
			opts: planRunLeaseCommandOptions{Domains: []string{"skill:kb-work"}},
			kind: "domain",
		},
		{
			name: "resource",
			opts: planRunLeaseCommandOptions{Resources: []string{"BROWSER:4110"}},
			kind: "resource",
		},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := test.opts
			opts.Action = "acquire"
			opts.StateRoot = stateRoot
			opts.RunID = "run-conflict-" + string(rune('a'+index))
			opts.ManifestPath = "docs/plans/conflict.md"
			opts.OwnerToken = "owner-conflict"
			opts.Now = now
			result := mustExecutePlanRunLease(t, opts)
			if result.OK || len(result.Collisions) == 0 {
				t.Fatalf("conflicting run was admitted: %#v", result)
			}
			collision := result.Collisions[0]
			if collision.RunID != "run-a" || collision.Claim.Kind != test.kind ||
				!strings.Contains(collision.Reason, "run-a") {
				t.Fatalf("collision did not identify owner and claim: %#v", collision)
			}
		})
	}

	disjoint := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-b", ManifestPath: "docs/plans/b.md",
		OwnerToken: "owner-b", Files: []string{"docs/readme.md"},
		Domains: []string{"skill:kb-plan"}, Resources: []string{"database:test-b"}, Now: now,
	})
	if !disjoint.OK {
		t.Fatalf("disjoint run was blocked: %#v", disjoint)
	}
}

func TestCrossManifestSchedulerObservedExpansionRequeuesBeforeStateMutation(t *testing.T) {
	t.Parallel()
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	first := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: "docs/plans/a.md",
		OwnerToken: "owner-a", Files: []string{"src/a.go"}, Now: now,
	})
	second := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-b", ManifestPath: "docs/plans/b.md",
		OwnerToken: "owner-b", Files: []string{"src/b.go"}, Now: now,
	})
	if !first.OK || !second.OK {
		t.Fatalf("setup acquire failed: first=%#v second=%#v", first, second)
	}

	blocked := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "expand", StateRoot: stateRoot, RunID: "run-b", OwnerToken: "owner-b",
		Generation: second.Lease.Generation, Files: []string{"src/a.go"}, Now: now,
	})
	if blocked.OK || !blocked.Requeued || len(blocked.Collisions) == 0 {
		t.Fatalf("observed collision did not requeue: %#v", blocked)
	}

	status := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "status", StateRoot: stateRoot, RunID: "run-b", OwnerToken: "owner-b", Now: now,
	})
	if !status.OK || status.Lease == nil || status.Lease.Generation != second.Lease.Generation {
		t.Fatalf("failed expansion mutated generation: %#v", status)
	}
	for _, claim := range status.Lease.Claims {
		if claim.Kind == "file" && claim.Value == normalizePathForOracle("src/a.go") {
			t.Fatalf("failed expansion mutated live claims: %#v", status.Lease.Claims)
		}
	}
}

func TestPlanRunLeaseOwnerGenerationLifecycleFailsClosed(t *testing.T) {
	t.Parallel()
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	acquired := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: "docs/plans/a.md",
		OwnerToken: "owner-a", Files: []string{"src/a.go"}, TTL: time.Second, Now: now,
	})
	if !acquired.OK || acquired.Lease == nil {
		t.Fatalf("acquire failed: %#v", acquired)
	}

	for _, attempt := range []planRunLeaseCommandOptions{
		{Action: "renew", StateRoot: stateRoot, RunID: "run-a", OwnerToken: "wrong", Generation: acquired.Lease.Generation, Now: now},
		{Action: "release", StateRoot: stateRoot, RunID: "run-a", OwnerToken: "owner-a", Generation: acquired.Lease.Generation + 1, Now: now},
		{Action: "recover", StateRoot: stateRoot, RunID: "run-a", OwnerToken: "owner-a", Generation: acquired.Lease.Generation, Now: now},
	} {
		result := mustExecutePlanRunLease(t, attempt)
		if result.OK {
			t.Fatalf("%s unexpectedly succeeded: %#v", attempt.Action, result)
		}
	}

	expiredAt := now.Add(2 * time.Second)
	recovered := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "recover", StateRoot: stateRoot, RunID: "run-a", OwnerToken: "owner-a",
		Generation: acquired.Lease.Generation, TTL: time.Minute, Now: expiredAt,
	})
	if !recovered.OK || recovered.Lease.Generation != acquired.Lease.Generation+1 {
		t.Fatalf("expired owner recovery failed: %#v", recovered)
	}
	wrongRelease := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "release", StateRoot: stateRoot, RunID: "run-a", OwnerToken: "owner-a",
		Generation: acquired.Lease.Generation, Now: expiredAt,
	})
	if wrongRelease.OK {
		t.Fatalf("stale generation released lease: %#v", wrongRelease)
	}
	released := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "release", StateRoot: stateRoot, RunID: "run-a", OwnerToken: "owner-a",
		Generation: recovered.Lease.Generation, Now: expiredAt,
	})
	if !released.OK {
		t.Fatalf("release failed: %#v", released)
	}
}

func TestPlanRunLeaseComposesWithSliceClaimsAndStatesLocalLimitation(t *testing.T) {
	t.Parallel()
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	plan := mustExecutePlanRunLease(t, planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: "docs/plans/a.md",
		OwnerToken: "owner-a", Prefixes: []string{"cmd/kbcheck"},
		Resources: []string{"git:integration-owner"}, Now: now,
	})
	if !plan.OK || plan.Lease == nil {
		t.Fatalf("plan acquire failed: %#v", plan)
	}
	limitations := strings.Join(plan.Lease.Limitations, " ")
	if !strings.Contains(limitations, "separate clones") ||
		!strings.Contains(limitations, "machines") ||
		!strings.Contains(limitations, "branch/PR protections") {
		t.Fatalf("local-only coordination limitation is incomplete: %v", plan.Lease.Limitations)
	}

	slice, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, SliceID: "slice-001", RunID: "run-a",
		OwnerToken: "owner-a", Files: []string{"cmd/kbcheck/main.go"},
		Resources: []string{"git:integration-owner"}, Now: now,
	})
	if err != nil || !slice.OK {
		t.Fatalf("same-lineage slice was blocked: result=%#v err=%v", slice, err)
	}

	wrongOwner, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, SliceID: "slice-002", RunID: "run-a",
		OwnerToken: "owner-b", Files: []string{"cmd/kbcheck/swarm.go"}, Now: now,
	})
	if err != nil || wrongOwner.OK || !strings.Contains(wrongOwner.Issue, "plan-run owner") {
		t.Fatalf("wrong-owner slice composed with plan: result=%#v err=%v", wrongOwner, err)
	}

	unforecast, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, SliceID: "slice-003", RunID: "run-a",
		OwnerToken: "owner-a", Files: []string{"docs/outside.md"}, Now: now,
	})
	if err != nil || unforecast.OK || !strings.Contains(unforecast.Issue, "expand") {
		t.Fatalf("unforecast write acquired before expansion: result=%#v err=%v", unforecast, err)
	}

	otherRun, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, SliceID: "slice-001", RunID: "run-b",
		OwnerToken: "owner-b", Files: []string{"cmd/kbcheck/other.go"}, Now: now,
	})
	if err != nil || otherRun.OK || len(otherRun.Collisions) == 0 {
		t.Fatalf("other run bypassed manifest claim: result=%#v err=%v", otherRun, err)
	}
}

func TestPlanRunLeaseStateRootSharesWorktreesButNotSeparateClone(t *testing.T) {
	t.Parallel()
	root := initWorktreeRepo(t)
	worktree := filepath.Join(t.TempDir(), "sibling")
	runGitForSliceLease(t, root, "worktree", "add", worktree)
	clone := filepath.Join(t.TempDir(), "clone")
	runGitForSliceLease(t, "", "clone", root, clone)

	rootState, err := resolvePlanRunLeaseStateRoot(planRunLeaseCommandOptions{RepoRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	worktreeState, err := resolvePlanRunLeaseStateRoot(planRunLeaseCommandOptions{RepoRoot: worktree})
	if err != nil {
		t.Fatal(err)
	}
	cloneState, err := resolvePlanRunLeaseStateRoot(planRunLeaseCommandOptions{RepoRoot: clone})
	if err != nil {
		t.Fatal(err)
	}
	if rootState != worktreeState {
		t.Fatalf("sibling worktrees do not share scheduler state: root=%s sibling=%s", rootState, worktreeState)
	}
	if rootState == cloneState {
		t.Fatalf("separate clone falsely shares scheduler state: %s", rootState)
	}
}

func mustExecutePlanRunLease(t *testing.T, opts planRunLeaseCommandOptions) planRunLeaseResult {
	t.Helper()
	result, err := executePlanRunLease(opts)
	if err != nil {
		t.Fatalf("executePlanRunLease: %v", err)
	}
	return result
}

func normalizePathForOracle(value string) string {
	normalized, err := normalizeLeaseClaimPath(value)
	if err != nil {
		panic(err)
	}
	return normalized
}
