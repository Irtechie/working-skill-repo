package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlanRunAdvanceAcceptsSequentialSliceCommitsWithIntegrationHeadCAS(t *testing.T) {
	root, manifest, worktree, opts, receipt := prepareAdvanceTestRun(t, "kb-advance-sequential")
	previous := receipt.IntegrationHead

	for index, sliceID := range []string{"slice-001", "slice-002"} {
		commit := commitPlanRunSlice(t, worktree, sliceID+".txt", "slice\n", "commit "+sliceID)
		proof := writeAdvanceProofReceipt(t, planRunProofReceipt{
			SchemaVersion: 1, KBID: receipt.KBID, RunID: "run-sequential", SliceID: sliceID,
			CommitSHA: commit, ObservedWrites: []string{sliceID + ".txt"},
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
		previous = commit
	}

	statusOpts := opts
	statusOpts.Action = "status"
	status, err := executePlanRunWorkspace(statusOpts)
	if err != nil || !status.OK || status.Receipt.IntegrationHead != previous {
		t.Fatalf("status lost sequential integration head: result=%#v err=%v", status, err)
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

func prepareAdvanceTestRun(t *testing.T, kbID string) (string, string, string, planRunWorkspaceOptions, *planRunWorkspaceReceipt) {
	t.Helper()
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, kbID)
	worktree := filepath.Join(t.TempDir(), kbID)
	opts := planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner-" + strings.TrimPrefix(kbID, "kb-advance-"),
		RunID:   "run-" + strings.TrimPrefix(kbID, "kb-advance-"),
		BaseSHA: gitOutput(root, "rev-parse", "HEAD"), Worktree: worktree,
		IntegrationRef: "codex/" + kbID, RepoRoot: root, Now: time.Now().UTC(),
	}
	result, err := executePlanRunWorkspace(opts)
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("prepare failed: result=%#v err=%v", result, err)
	}
	return root, manifest, worktree, opts, result.Receipt
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
