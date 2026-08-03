package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSliceLeaseTwoProcessSameSliceRace(t *testing.T) {
	t.Parallel()
	if os.Getenv("KBCHECK_SLICE_LEASE_HELPER") == "1" {
		var out, errOut strings.Builder
		code := run([]string{
			"slice-lease", "--action", "acquire", "--state-root", os.Getenv("KBCHECK_SLICE_LEASE_ROOT"),
			"--slice-id", "slice-002", "--run-id", os.Getenv("KBCHECK_SLICE_LEASE_RUN"),
			"--owner-token", os.Getenv("KBCHECK_SLICE_LEASE_OWNER"), "--file", "cmd/kbcheck/slice_lease.go",
		}, &out, &errOut)
		os.Exit(code)
	}

	root := t.TempDir()
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for _, owner := range []string{"owner-a", "owner-b"} {
		wg.Add(1)
		go func(owner string) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=^TestSliceLeaseTwoProcessSameSliceRace$")
			cmd.Env = append(os.Environ(),
				"KBCHECK_SLICE_LEASE_HELPER=1",
				"KBCHECK_SLICE_LEASE_ROOT="+root,
				"KBCHECK_SLICE_LEASE_RUN=run-"+owner,
				"KBCHECK_SLICE_LEASE_OWNER="+owner,
			)
			err := cmd.Run()
			if err == nil {
				results <- 0
				return
			}
			if exit, ok := err.(*exec.ExitError); ok {
				results <- exit.ExitCode()
				return
			}
			t.Errorf("helper failed without exit code: %v", err)
			results <- 99
		}(owner)
	}
	wg.Wait()
	close(results)

	counts := map[int]int{}
	for code := range results {
		counts[code]++
	}
	if counts[0] != 1 || counts[2] != 1 {
		t.Fatalf("expected one winner and one collision, got %#v", counts)
	}
}

func TestSliceLeaseOwnerTokenRenewReleaseAndRecovery(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	acquired, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: root, SliceID: "slice-002", RunID: "run-1", OwnerToken: "owner-1",
		Files: []string{"cmd/kbcheck/slice_lease.go"}, TTL: time.Second, Now: now,
	})
	if err != nil || !acquired.OK || acquired.Lease == nil {
		t.Fatalf("acquire failed: result=%#v err=%v", acquired, err)
	}

	wrongRenew, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "renew", StateRoot: root, SliceID: "slice-002", OwnerToken: "owner-2", Generation: acquired.Lease.Generation, Now: now,
	})
	if err != nil || wrongRenew.OK || !strings.Contains(wrongRenew.Issue, "owner token") {
		t.Fatalf("wrong owner renew passed: result=%#v err=%v", wrongRenew, err)
	}

	liveRecover, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "recover", StateRoot: root, SliceID: "slice-002", OwnerToken: "owner-1", Generation: acquired.Lease.Generation, Now: now,
	})
	if err != nil || liveRecover.OK || !strings.Contains(liveRecover.Issue, "still active") {
		t.Fatalf("live recover passed: result=%#v err=%v", liveRecover, err)
	}

	expiredAt := now.Add(2 * time.Second)
	wrongGeneration, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "recover", StateRoot: root, SliceID: "slice-002", OwnerToken: "owner-1", Generation: acquired.Lease.Generation + 1, Now: expiredAt,
	})
	if err != nil || wrongGeneration.OK || !strings.Contains(wrongGeneration.Issue, "generation") {
		t.Fatalf("wrong generation recover passed: result=%#v err=%v", wrongGeneration, err)
	}

	wrongOwner, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "recover", StateRoot: root, SliceID: "slice-002", OwnerToken: "owner-2", Generation: acquired.Lease.Generation, Now: expiredAt,
	})
	if err != nil || wrongOwner.OK || !strings.Contains(wrongOwner.Issue, "owner token") {
		t.Fatalf("wrong owner recover passed: result=%#v err=%v", wrongOwner, err)
	}

	recovered, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "recover", StateRoot: root, SliceID: "slice-002", OwnerToken: "owner-1", Generation: acquired.Lease.Generation,
		TTL: time.Minute, Now: expiredAt,
	})
	if err != nil || !recovered.OK || recovered.Lease.Generation != 2 {
		t.Fatalf("recover failed: result=%#v err=%v", recovered, err)
	}

	wrongRelease, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "release", StateRoot: root, SliceID: "slice-002", OwnerToken: "owner-2", Generation: recovered.Lease.Generation, Now: expiredAt,
	})
	if err != nil || wrongRelease.OK || !strings.Contains(wrongRelease.Issue, "owner token") {
		t.Fatalf("wrong owner release passed: result=%#v err=%v", wrongRelease, err)
	}

	released, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "release", StateRoot: root, SliceID: "slice-002", OwnerToken: "owner-1", Generation: recovered.Lease.Generation, Now: expiredAt,
	})
	if err != nil || !released.OK {
		t.Fatalf("release failed: result=%#v err=%v", released, err)
	}
}

func TestSliceLeasePathAndResourceConflictNormalization(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	first, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: root, SliceID: "slice-a", RunID: "run-a", OwnerToken: "owner-a",
		Prefixes: []string{"Cmd/Kbcheck"}, Resources: []string{"browser:4110"}, Now: now,
	})
	if err != nil || !first.OK {
		t.Fatalf("first acquire failed: result=%#v err=%v", first, err)
	}
	prefixCollision, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: root, SliceID: "slice-b", RunID: "run-b", OwnerToken: "owner-b",
		Files: []string{"cmd/kbcheck/main.go"}, Now: now,
	})
	if err != nil || prefixCollision.OK || len(prefixCollision.Collisions) == 0 {
		t.Fatalf("prefix collision missed: result=%#v err=%v", prefixCollision, err)
	}
	resourceCollision, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: root, SliceID: "slice-c", RunID: "run-c", OwnerToken: "owner-c",
		Resources: []string{"browser:4110"}, Now: now,
	})
	if err != nil || resourceCollision.OK || len(resourceCollision.Collisions) == 0 {
		t.Fatalf("resource collision missed: result=%#v err=%v", resourceCollision, err)
	}
	disjoint, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: root, SliceID: "slice-d", RunID: "run-d", OwnerToken: "owner-d",
		Files: []string{"docs/readme.md"}, Resources: []string{"database:test"}, Now: now,
	})
	if err != nil || !disjoint.OK {
		t.Fatalf("disjoint acquire failed: result=%#v err=%v", disjoint, err)
	}
	if _, err := normalizeLeaseClaims([]string{"../outside.go"}, nil, nil); err == nil {
		t.Fatal("path escape was accepted")
	}
}

func TestSliceLeaseGitCommonDirCoordinatesWorktreesButNotClones(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git unavailable")
	}
	root := filepath.Join(t.TempDir(), "repo")
	runGitForSliceLease(t, "", "init", root)
	runGitForSliceLease(t, root, "config", "user.email", "test@example.com")
	runGitForSliceLease(t, root, "config", "user.name", "Lease Test")
	writeFile(t, filepath.Join(root, "README.md"), "lease\n")
	runGitForSliceLease(t, root, "add", "README.md")
	runGitForSliceLease(t, root, "commit", "-m", "init")
	worktree := filepath.Join(t.TempDir(), "repo-worktree")
	runGitForSliceLease(t, root, "worktree", "add", worktree)
	clone := filepath.Join(t.TempDir(), "repo-clone")
	runGitForSliceLease(t, "", "clone", root, clone)

	rootState, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: root})
	if err != nil {
		t.Fatalf("root state: %v", err)
	}
	worktreeState, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: worktree})
	if err != nil {
		t.Fatalf("worktree state: %v", err)
	}
	cloneState, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: clone})
	if err != nil {
		t.Fatalf("clone state: %v", err)
	}
	if rootState != worktreeState {
		t.Fatalf("worktree did not share git common-dir state: root=%s worktree=%s", rootState, worktreeState)
	}
	if rootState == cloneState {
		t.Fatalf("separate clone unexpectedly shared state: %s", rootState)
	}
}

func TestSliceLeaseCommandStatusJSON(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	var out, errOut strings.Builder
	code := run([]string{
		"slice-lease", "--action", "acquire", "--state-root", root, "--slice-id", "slice-002", "--run-id", "run-1",
		"--owner-token", "owner-1", "--file", "cmd/kbcheck/slice_lease.go", "--json",
	}, &out, &errOut)
	if code != 0 {
		t.Fatalf("acquire command failed: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	var acquired sliceLeaseResult
	if err := json.Unmarshal([]byte(out.String()), &acquired); err != nil || acquired.Lease == nil {
		t.Fatalf("parse acquire: result=%#v err=%v", acquired, err)
	}
	out.Reset()
	errOut.Reset()
	code = run([]string{"slice-lease", "--action", "status", "--state-root", root, "--json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"slice-002"`) {
		t.Fatalf("status command failed: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestPlanRunLeaseAllowsRunQualifiedSiblingSliceIDsInsideOneWorkstreamEach(t *testing.T) {
	t.Parallel()
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	for _, plan := range []planRunLeaseCommandOptions{
		{
			Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: "docs/plans/a.md",
			OwnerToken: "owner-a", Prefixes: []string{"src/a"}, Now: now,
		},
		{
			Action: "acquire", StateRoot: stateRoot, RunID: "run-b", ManifestPath: "docs/plans/b.md",
			OwnerToken: "owner-b", Prefixes: []string{"src/b"}, Now: now,
		},
	} {
		result, err := executePlanRunLease(plan)
		if err != nil || !result.OK {
			t.Fatalf("plan acquire failed: result=%#v err=%v", result, err)
		}
	}
	for _, slice := range []sliceLeaseCommandOptions{
		{
			Action: "acquire", StateRoot: stateRoot, RunID: "run-a", SliceID: "slice-001",
			OwnerToken: "owner-a", Files: []string{"src/a/one.go"}, Now: now,
		},
		{
			Action: "acquire", StateRoot: stateRoot, RunID: "run-b", SliceID: "slice-001",
			OwnerToken: "owner-b", Files: []string{"src/b/one.go"}, Now: now,
		},
	} {
		result, err := executeSliceLease(slice)
		if err != nil || !result.OK {
			t.Fatalf("run-qualified slice acquire failed: result=%#v err=%v", result, err)
		}
	}
	status, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "status", StateRoot: stateRoot, Now: now,
	})
	if err != nil || !status.OK || len(status.Leases) != 2 {
		t.Fatalf("run-qualified slice status failed: result=%#v err=%v", status, err)
	}
}

func TestPlanRunLeaseRecoveryRechecksContentionAndReleaseWaitsForSlices(t *testing.T) {
	t.Parallel()
	stateRoot := t.TempDir()
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	first, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: "docs/plans/a.md",
		OwnerToken: "owner-a", Files: []string{"src/a.go"}, TTL: time.Second, Now: now,
	})
	if err != nil || !first.OK {
		t.Fatalf("first acquire failed: result=%#v err=%v", first, err)
	}
	later := now.Add(2 * time.Second)
	second, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-b", ManifestPath: "docs/plans/b.md",
		OwnerToken: "owner-b", Files: []string{"src/a.go"}, TTL: time.Minute, Now: later,
	})
	if err != nil || !second.OK {
		t.Fatalf("second acquire after expiry failed: result=%#v err=%v", second, err)
	}
	recovered, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "recover", StateRoot: stateRoot, RunID: "run-a", OwnerToken: "owner-a",
		Generation: first.Lease.Generation, TTL: time.Minute, Now: later,
	})
	if err != nil || recovered.OK || !recovered.Requeued || len(recovered.Collisions) == 0 {
		t.Fatalf("recovery bypassed new owner: result=%#v err=%v", recovered, err)
	}

	disjointRoot := t.TempDir()
	plan, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: disjointRoot, RunID: "run-c", ManifestPath: "docs/plans/c.md",
		OwnerToken: "owner-c", Files: []string{"src/c.go"}, Now: now,
	})
	if err != nil || !plan.OK {
		t.Fatalf("disjoint plan acquire failed: result=%#v err=%v", plan, err)
	}
	slice, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: disjointRoot, RunID: "run-c", SliceID: "slice-001",
		OwnerToken: "owner-c", Files: []string{"src/c.go"}, Now: now,
	})
	if err != nil || !slice.OK {
		t.Fatalf("slice acquire failed: result=%#v err=%v", slice, err)
	}
	release, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "release", StateRoot: disjointRoot, RunID: "run-c", OwnerToken: "owner-c",
		Generation: plan.Lease.Generation, Now: now,
	})
	if err != nil || release.OK || !strings.Contains(release.Issue, "release slices first") {
		t.Fatalf("plan released with active slice: result=%#v err=%v", release, err)
	}
}

func TestPlanRunLeaseTwoProcessClaimRaceHasOneMutationAuthority(t *testing.T) {
	t.Parallel()
	if os.Getenv("KBCHECK_PLAN_RUN_LEASE_HELPER") == "1" {
		var out, errOut strings.Builder
		code := run([]string{
			"plan-run-lease", "--action", "acquire",
			"--state-root", os.Getenv("KBCHECK_PLAN_RUN_LEASE_ROOT"),
			"--root", os.Getenv("KBCHECK_PLAN_RUN_REPO_ROOT"),
			"--run-id", os.Getenv("KBCHECK_PLAN_RUN_LEASE_RUN"),
			"--manifest", os.Getenv("KBCHECK_PLAN_RUN_MANIFEST"),
			"--owner-token", os.Getenv("KBCHECK_PLAN_RUN_LEASE_OWNER"),
			"--file", "src/shared.go",
		}, &out, &errOut)
		os.Exit(code)
	}

	stateRoot := t.TempDir()
	manifest := filepath.Join(stateRoot, "race.md")
	writeFile(t, manifest, "---\ntype: kb-manifest\nslices: []\n---\n")
	results := make(chan int, 2)
	var wg sync.WaitGroup
	for _, suffix := range []string{"a", "b"} {
		wg.Add(1)
		go func(suffix string) {
			defer wg.Done()
			cmd := exec.Command(os.Args[0], "-test.run=^TestPlanRunLeaseTwoProcessClaimRaceHasOneMutationAuthority$")
			cmd.Env = append(os.Environ(),
				"KBCHECK_PLAN_RUN_LEASE_HELPER=1",
				"KBCHECK_PLAN_RUN_LEASE_ROOT="+stateRoot,
				"KBCHECK_PLAN_RUN_REPO_ROOT="+stateRoot,
				"KBCHECK_PLAN_RUN_MANIFEST="+manifest,
				"KBCHECK_PLAN_RUN_LEASE_RUN=run-"+suffix,
				"KBCHECK_PLAN_RUN_LEASE_OWNER=owner-"+suffix,
			)
			err := cmd.Run()
			if err == nil {
				results <- 0
				return
			}
			if exit, ok := err.(*exec.ExitError); ok {
				results <- exit.ExitCode()
				return
			}
			t.Errorf("plan-run helper failed without exit code: %v", err)
			results <- 99
		}(suffix)
	}
	wg.Wait()
	close(results)

	counts := map[int]int{}
	for code := range results {
		counts[code]++
	}
	if counts[0] != 1 || counts[2] != 1 {
		t.Fatalf("expected one plan-run winner and one collision, got %#v", counts)
	}
}

func runGitForSliceLease(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, output)
	}
}
