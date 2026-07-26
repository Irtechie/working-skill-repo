package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlanRunAdvanceAcceptsSequentialSliceCommitsWithIntegrationHeadCAS(t *testing.T) {
	root, manifest, worktree, opts, receipt := prepareAdvanceTestRun(t, "kb-advance-sequential")
	previous := receipt.IntegrationHead
	archives := []string{}
	archiveBytes := [][]byte{}
	manifestRelative, err := filepath.Rel(root, manifest)
	if err != nil {
		t.Fatal(err)
	}

	for index, sliceID := range []string{"slice-001", "slice-002"} {
		writeFile(t, filepath.Join(worktree, sliceID+".txt"), "slice\n")
		writeFile(t, filepath.Join(worktree, manifestRelative), advanceTestManifest(receipt.KBID, index+1))
		runGitForSliceLease(t, worktree, "add", sliceID+".txt", filepath.ToSlash(manifestRelative))
		runGitForSliceLease(t, worktree, "commit", "-m", "commit "+sliceID+" with lifecycle")
		commit := gitOutput(worktree, "rev-parse", "HEAD")
		observed := []string{sliceID + ".txt", filepath.ToSlash(manifestRelative)}
		sliceLease := acquireAdvanceTestSliceLease(t, receipt, sliceID, observed...)
		proof := writeAdvanceProofReceipt(t, planRunProofReceipt{
			SchemaVersion: 1, KBID: receipt.KBID, RunID: "run-sequential", SliceID: sliceID,
			CommitSHA: commit, ObservedWrites: observed,
			SliceProof:     planRunProofCommand{Args: []string{"git", "rev-parse", "--verify", "HEAD"}, Expect: 0},
			AggregateProof: &planRunProofCommand{Args: []string{"git", "status", "--porcelain"}, Expect: 0, ExpectOutput: ""},
		})
		result, err := executePlanRunWorkspace(planRunWorkspaceOptions{
			Action: "advance", ManifestPath: manifest, OwnerToken: "owner-sequential",
			RunID: "run-sequential", SliceID: sliceID, ExpectedIntegrationHead: previous,
			CommitSHA: commit, Worktree: worktree, IntegrationRef: receipt.IntegrationRef,
			ProofReceipt: proof, RepoRoot: root, Now: time.Now().UTC(),
		})
		if err != nil || !result.OK || result.Receipt == nil {
			t.Fatalf("advance %d failed: result=%#v err=%v", index, result, err)
		}
		if result.Receipt.IntegrationHead != commit || result.Receipt.LastSliceID != sliceID {
			t.Fatalf("advance %d did not record commit: %#v", index, result.Receipt)
		}
		archived, err := os.ReadFile(result.Receipt.LastProofArchive)
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(archived)) != result.Receipt.LastProofSHA256 {
			t.Fatalf("advance %d did not preserve immutable proof evidence: receipt=%#v err=%v", index, result.Receipt, err)
		}
		archives = append(archives, result.Receipt.LastProofArchive)
		archiveBytes = append(archiveBytes, archived)
		writeFile(t, proof, "{}\n")
		afterMutation, err := os.ReadFile(result.Receipt.LastProofArchive)
		if err != nil || !bytes.Equal(afterMutation, archived) {
			t.Fatalf("mutable proof path changed archived evidence: err=%v", err)
		}
		previous = commit
		receipt = result.Receipt
		if index == 1 {
			expiredOpts := opts
			expiredOpts.Action = "complete"
			expiredOpts.Now = time.Now().UTC().Add(time.Hour)
			blocked, err := executePlanRunWorkspace(expiredOpts)
			if err != nil || blocked.OK || !strings.Contains(blocked.Issue, "explicitly released") ||
				!strings.Contains(blocked.Issue, "expired") {
				t.Fatalf("expired unreleased slice was accepted: result=%#v err=%v", blocked, err)
			}
		}
		releaseAdvanceTestSliceLease(t, receipt, sliceLease)
	}

	statusOpts := opts
	statusOpts.Action = "status"
	status, err := executePlanRunWorkspace(statusOpts)
	if err != nil || !status.OK || status.Receipt.IntegrationHead != previous {
		t.Fatalf("status lost sequential integration head: result=%#v err=%v", status, err)
	}
	planStatus, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "status", RunID: receipt.RunID, OwnerToken: receipt.OwnerToken,
		RepoRoot: worktree, Now: time.Now().UTC(),
	})
	if err != nil || !planStatus.OK || planStatus.Lease == nil {
		t.Fatalf("plan lease status failed: result=%#v err=%v", planStatus, err)
	}
	planReleased, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "release", RunID: receipt.RunID, OwnerToken: receipt.OwnerToken,
		Generation: planStatus.Lease.Generation, RepoRoot: worktree, Now: time.Now().UTC(),
	})
	if err != nil || planReleased.OK || !strings.Contains(planReleased.Issue, "atomically") {
		t.Fatalf("manifest-owned plan lease bypassed atomic completion: result=%#v err=%v", planReleased, err)
	}
	unrelatedStatus, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "status", RepoRoot: worktree, Now: time.Now().UTC(),
	})
	if err != nil || !unrelatedStatus.OK {
		t.Fatalf("unrelated lease action failed: result=%#v err=%v", unrelatedStatus, err)
	}
	writeFile(t, archives[0], "{}\n")
	completeOpts := opts
	completeOpts.Action = "complete"
	blocked, err := executePlanRunWorkspace(completeOpts)
	if err != nil || blocked.OK || !strings.Contains(blocked.Issue, "digest mismatch") {
		t.Fatalf("tampered immutable proof was accepted: result=%#v err=%v", blocked, err)
	}
	if err := os.WriteFile(archives[0], archiveBytes[0], 0o644); err != nil {
		t.Fatal(err)
	}
	completed, err := executePlanRunWorkspace(completeOpts)
	if err != nil || !completed.OK || completed.Receipt.Status != "completed" {
		t.Fatalf("plan-run completion transition failed: result=%#v err=%v", completed, err)
	}
	retried, err := executePlanRunWorkspace(completeOpts)
	if err != nil || !retried.OK || retried.Receipt.Status != "completed" {
		t.Fatalf("completion retry did not consume durable complete-release journal: result=%#v err=%v", retried, err)
	}
	planStatus, err = executePlanRunLease(planRunLeaseCommandOptions{
		Action: "status", RunID: receipt.RunID, OwnerToken: receipt.OwnerToken,
		RepoRoot: worktree, Now: time.Now().UTC(),
	})
	if err != nil || !planStatus.OK || planStatus.Lease == nil || planStatus.Lease.Status != "released" {
		t.Fatalf("completion did not atomically release plan lease: result=%#v err=%v", planStatus, err)
	}
	releaseOpts := opts
	releaseOpts.Action = "release"
	released, err := executePlanRunWorkspace(releaseOpts)
	if err != nil || !released.OK || released.Receipt.Status != "released" || pathExists(worktree) {
		t.Fatalf("completed plan-run release failed: result=%#v err=%v", released, err)
	}
}

func TestPlanRunAdvanceRejectsWrongIdentityDirtyStateAndStaleIntegrationHead(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions
		issue  string
	}{
		{
			name: "wrong worktree",
			mutate: func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions {
				commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
				return advanceOptionsForTest(t, root, manifest, filepath.Join(t.TempDir(), "other"), receipt, commit, "slice-001")
			},
			issue: "worktree",
		},
		{
			name: "wrong branch",
			mutate: func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions {
				commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
				opts := advanceOptionsForTest(t, root, manifest, worktree, receipt, commit, "slice-001")
				opts.IntegrationRef = "codex/not-the-owner"
				return opts
			},
			issue: "integration ref",
		},
		{
			name: "wrong owner",
			mutate: func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions {
				commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
				opts := advanceOptionsForTest(t, root, manifest, worktree, receipt, commit, "slice-001")
				opts.OwnerToken = "wrong-owner"
				return opts
			},
			issue: "owner",
		},
		{
			name: "wrong run",
			mutate: func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions {
				commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
				opts := advanceOptionsForTest(t, root, manifest, worktree, receipt, commit, "slice-001")
				opts.RunID = "wrong-run"
				return opts
			},
			issue: "run",
		},
		{
			name: "dirty worktree",
			mutate: func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions {
				commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
				writeFile(t, filepath.Join(worktree, "dirty.txt"), "dirty\n")
				return advanceOptionsForTest(t, root, manifest, worktree, receipt, commit, "slice-001")
			},
			issue: "dirty",
		},
		{
			name: "unexpected head movement",
			mutate: func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions {
				commit := commitPlanRunSlice(t, worktree, "outside.txt", "outside\n", "outside actor")
				opts := advanceOptionsForTest(t, root, manifest, worktree, receipt, receipt.IntegrationHead, "slice-001")
				if commit == receipt.IntegrationHead {
					t.Fatal("outside commit did not move head")
				}
				return opts
			},
			issue: "current head",
		},
		{
			name: "stale compare and swap",
			mutate: func(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt) planRunWorkspaceOptions {
				commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
				opts := advanceOptionsForTest(t, root, manifest, worktree, receipt, commit, "slice-001")
				opts.ExpectedIntegrationHead = strings.Repeat("a", 40)
				return opts
			},
			issue: "expected integration head",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, manifest, worktree, _, receipt := prepareAdvanceTestRun(t, "kb-advance-"+strings.ReplaceAll(test.name, " ", "-"))
			opts := test.mutate(t, root, manifest, worktree, receipt)
			before := receipt.IntegrationHead
			result, err := executePlanRunWorkspace(opts)
			if err != nil || result.OK || !strings.Contains(strings.ToLower(result.Issue), test.issue) {
				t.Fatalf("advance did not fail closed: result=%#v err=%v", result, err)
			}
			stored, err := loadPlanRunWorkspaceReceipt(planRunWorkspaceOptions{RepoRoot: root}, receipt.KBID)
			if err != nil || stored.IntegrationHead != before {
				t.Fatalf("failed advance mutated receipt: receipt=%#v err=%v", stored, err)
			}
		})
	}
}

func TestSliceCommitRequiresCoordinatorProofReplayAndAggregateSuccess(t *testing.T) {
	root, manifest, worktree, _, receipt := prepareAdvanceTestRun(t, "kb-advance-proof")
	commit := commitPlanRunSlice(t, worktree, "slice.txt", "slice\n", "slice")
	acquireAdvanceTestSliceLease(t, receipt, "slice-001", "slice.txt")
	base := advanceOptionsForTest(t, root, manifest, worktree, receipt, commit, "slice-001")

	missing := base
	missing.ProofReceipt = ""
	result, err := executePlanRunWorkspace(missing)
	if err != nil || result.OK || !strings.Contains(result.Issue, "proof receipt") {
		t.Fatalf("missing proof receipt passed: result=%#v err=%v", result, err)
	}

	failingProof := writeAdvanceProofReceipt(t, planRunProofReceipt{
		SchemaVersion: 1, KBID: receipt.KBID, RunID: "run-proof", SliceID: "slice-001",
		CommitSHA: commit, ObservedWrites: []string{"slice.txt"},
		SliceProof: planRunProofCommand{Args: []string{"git", "rev-parse", "--verify", "HEAD"}, Expect: 0},
		AggregateProof: &planRunProofCommand{
			Args: []string{"git", "diff", "--quiet", receipt.IntegrationHead, commit}, Expect: 0,
		},
	})
	failing := base
	failing.ProofReceipt = failingProof
	result, err = executePlanRunWorkspace(failing)
	if err != nil || result.OK || !strings.Contains(result.Issue, "aggregate proof") {
		t.Fatalf("aggregate proof failure passed: result=%#v err=%v", result, err)
	}
	stored, err := loadPlanRunWorkspaceReceipt(planRunWorkspaceOptions{RepoRoot: root}, receipt.KBID)
	if err != nil || stored.IntegrationHead != receipt.IntegrationHead {
		t.Fatalf("failed aggregate proof advanced receipt: receipt=%#v err=%v", stored, err)
	}

	mismatch := writeAdvanceProofReceipt(t, planRunProofReceipt{
		SchemaVersion: 1, KBID: receipt.KBID, RunID: "run-proof", SliceID: "slice-001",
		CommitSHA: receipt.IntegrationHead, ObservedWrites: []string{"slice.txt"},
		SliceProof: planRunProofCommand{Args: []string{"git", "rev-parse", "--verify", "HEAD"}, Expect: 0},
	})
	base.ProofReceipt = mismatch
	result, err = executePlanRunWorkspace(base)
	if err != nil || result.OK || !strings.Contains(result.Issue, "commit") {
		t.Fatalf("mismatched worker receipt passed: result=%#v err=%v", result, err)
	}
}

func TestPlanRunAdvanceRejectsMissingOrFalseClaimEvidence(t *testing.T) {
	t.Run("no active slice lease", func(t *testing.T) {
		root, manifest, worktree, _, receipt := prepareAdvanceTestRun(t, "kb-advance-no-slice-lease")
		commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
		opts := validAdvanceOptions(t, root, manifest, receipt, commit, "slice-001", []string{"a.txt"})
		result, err := executePlanRunWorkspace(opts)
		if err != nil || result.OK || !strings.Contains(result.Issue, "active slice lease") {
			t.Fatalf("advance without slice lease passed: result=%#v err=%v", result, err)
		}
	})

	t.Run("observed writes lie", func(t *testing.T) {
		root, manifest, worktree, _, receipt := prepareAdvanceTestRun(t, "kb-advance-observed-lie")
		commit := commitPlanRunSlice(t, worktree, "a.txt", "a\n", "slice")
		acquireAdvanceTestSliceLease(t, receipt, "slice-001", "a.txt")
		opts := validAdvanceOptions(t, root, manifest, receipt, commit, "slice-001", []string{"outside.txt"})
		result, err := executePlanRunWorkspace(opts)
		if err != nil || result.OK || !strings.Contains(result.Issue, "exactly match commit diff") {
			t.Fatalf("false observed writes passed: result=%#v err=%v", result, err)
		}
	})

	t.Run("write outside slice claim", func(t *testing.T) {
		root, manifest, worktree, _, receipt := prepareAdvanceTestRun(t, "kb-advance-outside-slice-claim")
		commit := commitPlanRunSlice(t, worktree, "outside.txt", "outside\n", "slice")
		acquireAdvanceTestSliceLease(t, receipt, "slice-001", "a.txt")
		opts := validAdvanceOptions(t, root, manifest, receipt, commit, "slice-001", []string{"outside.txt"})
		result, err := executePlanRunWorkspace(opts)
		if err != nil || result.OK || !strings.Contains(result.Issue, "outside active slice lease") {
			t.Fatalf("out-of-claim write passed: result=%#v err=%v", result, err)
		}
	})
}

func prepareAdvanceTestRun(t *testing.T, kbID string) (string, string, string, planRunWorkspaceOptions, *planRunWorkspaceReceipt) {
	t.Helper()
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, kbID)
	writeFile(t, manifest, advanceTestManifest(kbID, 0))
	runGitForSliceLease(t, root, "add", filepath.ToSlash(strings.TrimPrefix(manifest, root+string(filepath.Separator))))
	runGitForSliceLease(t, root, "commit", "-m", "add advance lifecycle contract")
	worktree := filepath.Join(t.TempDir(), kbID)
	opts := planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner-" + strings.TrimPrefix(kbID, "kb-advance-"),
		CommitAuthorized:   true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:" + kbID,
		RunID:   "run-" + strings.TrimPrefix(kbID, "kb-advance-"),
		BaseSHA: gitOutput(root, "rev-parse", "HEAD"), Worktree: worktree,
		IntegrationRef: "codex/" + kbID, RepoRoot: root, Now: time.Now().UTC(),
	}
	result, err := executePlanRunWorkspace(opts)
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("prepare failed: result=%#v err=%v", result, err)
	}
	relativeManifest, err := filepath.Rel(root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	planLease, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", RunID: result.Receipt.RunID, ManifestPath: filepath.Join(worktree, relativeManifest),
		OwnerToken: result.Receipt.OwnerToken,
		Files: []string{
			"slice-001.txt", "slice-002.txt", "slice.txt", "a.txt", "outside.txt",
			filepath.ToSlash(relativeManifest),
		},
		RepoRoot: worktree, Now: time.Now().UTC(),
	})
	if err != nil || !planLease.OK {
		t.Fatalf("plan lease acquire failed: result=%#v err=%v", planLease, err)
	}
	return root, manifest, worktree, opts, result.Receipt
}

func advanceTestManifest(kbID string, stage int) string {
	status := "active"
	sliceOne := "pending"
	sliceTwo := "pending"
	sliceOneGate := "pending"
	sliceTwoGate := "pending"
	workGate := "pending"
	if stage >= 1 {
		sliceOne = "done"
		sliceOneGate = "passed"
	}
	if stage >= 2 {
		status = "completed"
		sliceTwo = "done"
		sliceTwoGate = "passed"
		workGate = "passed"
	}
	return fmt.Sprintf(`---
type: kb-manifest
kb_id: %s
status: %s
workspace_isolation_contract:
  plan_run_worktree_default: true
gate_ledger:
  - gate_id: slice-slice-001-to-done
    status: %s
  - gate_id: slice-slice-002-to-done
    status: %s
  - gate_id: work-to-complete
    status: %s
slices:
  - id: slice-001
    blockers: []
    status: %s
  - id: slice-002
    blockers: [slice-001]
    status: %s
---
`, kbID, status, sliceOneGate, sliceTwoGate, workGate, sliceOne, sliceTwo)
}

func acquireAdvanceTestSliceLease(t *testing.T, receipt *planRunWorkspaceReceipt, sliceID string, files ...string) *sliceLease {
	t.Helper()
	result, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", SliceID: sliceID, RunID: receipt.RunID, OwnerToken: receipt.OwnerToken,
		Files: files, Worktree: receipt.Worktree,
		Branch: receipt.IntegrationRef, BaseSHA: receipt.IntegrationHead,
		RepoRoot: receipt.Worktree, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK || result.Lease == nil {
		t.Fatalf("slice lease acquire failed: result=%#v err=%v", result, err)
	}
	return result.Lease
}

func releaseAdvanceTestSliceLease(t *testing.T, receipt *planRunWorkspaceReceipt, lease *sliceLease) {
	t.Helper()
	result, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "release", SliceID: lease.SliceID, RunID: receipt.RunID,
		OwnerToken: receipt.OwnerToken, Generation: lease.Generation,
		RepoRoot: receipt.Worktree, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK {
		t.Fatalf("slice lease release failed: result=%#v err=%v", result, err)
	}
}

func advanceOptionsForTest(t *testing.T, root, manifest, worktree string, receipt *planRunWorkspaceReceipt, commit, sliceID string) planRunWorkspaceOptions {
	t.Helper()
	proof := writeAdvanceProofReceipt(t, planRunProofReceipt{
		SchemaVersion: 1, KBID: receipt.KBID, RunID: effectivePlanRunID(*receipt), SliceID: sliceID,
		CommitSHA: commit, ObservedWrites: []string{"a.txt"},
		SliceProof: planRunProofCommand{Args: []string{"git", "rev-parse", "--verify", "HEAD"}, Expect: 0},
	})
	return planRunWorkspaceOptions{
		Action: "advance", ManifestPath: manifest, OwnerToken: receipt.OwnerToken,
		RunID: effectivePlanRunID(*receipt), SliceID: sliceID,
		ExpectedIntegrationHead: receipt.IntegrationHead, CommitSHA: commit,
		Worktree: worktree, IntegrationRef: receipt.IntegrationRef,
		ProofReceipt: proof, RepoRoot: root, Now: time.Now().UTC(),
	}
}

func validAdvanceOptions(
	t *testing.T,
	root, manifest string,
	receipt *planRunWorkspaceReceipt,
	commit, sliceID string,
	observed []string,
) planRunWorkspaceOptions {
	t.Helper()
	proof := writeAdvanceProofReceipt(t, planRunProofReceipt{
		SchemaVersion: 1, KBID: receipt.KBID, RunID: receipt.RunID, SliceID: sliceID,
		CommitSHA: commit, ObservedWrites: observed,
		SliceProof:     planRunProofCommand{Args: []string{"git", "rev-parse", "--verify", "HEAD"}, Expect: 0},
		AggregateProof: &planRunProofCommand{Args: []string{"git", "status", "--porcelain"}, Expect: 0, ExpectOutput: ""},
	})
	return planRunWorkspaceOptions{
		Action: "advance", ManifestPath: manifest, OwnerToken: receipt.OwnerToken,
		RunID: receipt.RunID, SliceID: sliceID, ExpectedIntegrationHead: receipt.IntegrationHead,
		CommitSHA: commit, Worktree: receipt.Worktree, IntegrationRef: receipt.IntegrationRef,
		ProofReceipt: proof, RepoRoot: root, Now: time.Now().UTC(),
	}
}

func commitPlanRunSlice(t *testing.T, worktree, name, content, message string) string {
	t.Helper()
	writeFile(t, filepath.Join(worktree, name), content)
	runGitForSliceLease(t, worktree, "add", name)
	runGitForSliceLease(t, worktree, "commit", "-m", message)
	return gitOutput(worktree, "rev-parse", "HEAD")
}

func writeAdvanceProofReceipt(t *testing.T, receipt planRunProofReceipt) string {
	t.Helper()
	content, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "proof.json")
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
