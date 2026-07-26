package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDefaultBranchBoundaryRejectsLocalAndRemoteDefaultInternalTargets(t *testing.T) {
	root := initWorktreeRepo(t)
	current := gitOutput(root, "branch", "--show-current")
	if current == "" {
		t.Fatal("test repository has no current branch")
	}
	runGitForSliceLease(t, root, "branch", "main")
	if !isResolvedDefaultBranch(root, current) || !isResolvedDefaultBranch(root, "main") {
		t.Fatalf("local defaults not resolved: current=%s", current)
	}

	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "branch", "release")
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	runGitForSliceLease(t, root, "push", "origin", "release")
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/release")
	runGitForSliceLease(t, root, "fetch", "origin")
	runGitForSliceLease(t, root, "remote", "set-head", "origin", "-a")
	if !isResolvedDefaultBranch(root, "release") {
		t.Fatal("remote default branch was not resolved")
	}

	for _, target := range []string{current, "main", "release"} {
		manifest := writePlanRunTestManifest(t, root, "kb-default-"+target)
		worktree := filepath.Join(t.TempDir(), "wt-"+target)
		result, err := executePlanRunWorkspace(planRunWorkspaceOptions{
			Action: "prepare", ManifestPath: manifest, OwnerToken: "owner", CommitAuthorized: true,
			CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:delivery-boundary",
			RunID: "run-default", BaseSHA: gitOutput(root, "rev-parse", "HEAD"),
			Worktree: worktree, IntegrationRef: target, RepoRoot: root, Now: time.Now().UTC(),
		})
		if err != nil || result.OK || !strings.Contains(result.Issue, "default branch") {
			t.Fatalf("default target %s was accepted: result=%#v err=%v", target, result, err)
		}
		if pathExists(worktree) {
			t.Fatalf("default target %s created a worktree before refusal", target)
		}
	}
}

func TestPlanRunPrepareBlocksWhenRemoteDefaultAuthorityIsUnresolved(t *testing.T) {
	root := initWorktreeRepo(t)
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	manifest := writePlanRunTestManifest(t, root, "kb-unresolved-remote-default")
	worktree := filepath.Join(t.TempDir(), "unresolved-default")

	result, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner", CommitAuthorized: true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:unresolved-default",
		RunID: "run-unresolved", BaseSHA: gitOutput(root, "rev-parse", "HEAD"),
		Worktree: worktree, IntegrationRef: "codex/unresolved-default", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "remote default branch authority is unresolved") {
		t.Fatalf("unresolved remote default authority was accepted: result=%#v err=%v", result, err)
	}
	if pathExists(worktree) {
		t.Fatal("unresolved remote authority created a worktree")
	}
}

func TestDirtyBaseAuthorityBlocksRelevantWIPAndPreservesUnrelatedDirt(t *testing.T) {
	requiredRoot, requiredManifest := writeDirtyAuthorityRepo(t, "required")
	requiredPath := filepath.Join(requiredRoot, "src", "required.txt")
	writeFile(t, requiredPath, "user required wip\n")
	beforeHead := gitOutput(requiredRoot, "rev-parse", "HEAD")
	requiredWorktree := filepath.Join(t.TempDir(), "required-wt")
	blocked, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: requiredManifest, OwnerToken: "owner", CommitAuthorized: true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:dirty-required",
		RunID: "run-required", BaseSHA: beforeHead, Worktree: requiredWorktree,
		IntegrationRef: "codex/required", RepoRoot: requiredRoot, Now: time.Now().UTC(),
	})
	if err != nil || blocked.OK || !strings.Contains(blocked.Issue, "relevant dirty") ||
		!strings.Contains(blocked.Issue, "explicit") {
		t.Fatalf("relevant dirty WIP did not block: result=%#v err=%v", blocked, err)
	}
	if gitOutput(requiredRoot, "rev-parse", "HEAD") != beforeHead ||
		string(mustReadFile(t, requiredPath)) != "user required wip\n" ||
		pathExists(requiredWorktree) {
		t.Fatal("relevant dirty refusal committed, changed, or omitted user WIP")
	}

	unrelatedRoot, unrelatedManifest := writeDirtyAuthorityRepo(t, "unrelated")
	unrelatedPath := filepath.Join(unrelatedRoot, "notes.txt")
	writeFile(t, unrelatedPath, "unrelated user dirt\n")
	unrelatedWorktree := filepath.Join(t.TempDir(), "unrelated-wt")
	allowed, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: unrelatedManifest, OwnerToken: "owner", CommitAuthorized: true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:dirty-unrelated",
		RunID: "run-unrelated", BaseSHA: gitOutput(unrelatedRoot, "rev-parse", "HEAD"),
		Worktree: unrelatedWorktree, IntegrationRef: "codex/unrelated",
		RepoRoot: unrelatedRoot, Now: time.Now().UTC(),
	})
	if err != nil || !allowed.OK || allowed.Receipt == nil || !allowed.Receipt.SourceDirty {
		t.Fatalf("unrelated dirt blocked isolated run: result=%#v err=%v", allowed, err)
	}
	if string(mustReadFile(t, unrelatedPath)) != "unrelated user dirt\n" {
		t.Fatal("unrelated dirt changed during plan-run preparation")
	}
}

func TestDeliveryOwnerDefaultsLocalAndKeepsPolicyOutsidePlanIntegration(t *testing.T) {
	root := initWorktreeRepo(t)
	local, err := resolveKBDeliveryPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if local.Mode != "local" || local.Merge != "manual" || local.Source != "default-absent" {
		t.Fatalf("absent policy did not default local-only: %#v", local)
	}

	config := filepath.Join(root, "docs", "context", "operations", "kb-routing.yaml")
	writeFile(t, config, "delivery:\n  mode: pr\n  merge: manual\n  post_merge_sync: false\n")
	prPolicy, err := resolveKBDeliveryPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	if prPolicy.Mode != "pr" || prPolicy.Merge != "manual" || prPolicy.Source != "project" {
		t.Fatalf("PR/manual policy not resolved: %#v", prPolicy)
	}

	manifest := writePlanRunTestManifest(t, root, "kb-delivery-context")
	result, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner", RunID: "run-delivery", CommitAuthorized: true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:delivery-context",
		BaseSHA: gitOutput(root, "rev-parse", "HEAD"), Worktree: filepath.Join(t.TempDir(), "delivery"),
		IntegrationRef: "codex/delivery-context", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("prepare failed: result=%#v err=%v", result, err)
	}
	if result.Receipt.DeliveryMode != "pr" || result.Receipt.DeliveryMerge != "manual" ||
		result.Receipt.DeliveryPolicySource != "project" {
		t.Fatalf("receipt lost delivery context: %#v", result.Receipt)
	}
	if result.Receipt.IntegrationRef == currentDefaultBranch(root) {
		t.Fatal("delivery context retargeted internal integration to default")
	}
}

func writeDirtyAuthorityRepo(t *testing.T, suffix string) (string, string) {
	t.Helper()
	root := initWorktreeRepo(t)
	required := filepath.Join(root, "src", "required.txt")
	writeFile(t, required, "committed\n")
	runGitForSliceLease(t, root, "add", "src/required.txt")
	runGitForSliceLease(t, root, "commit", "-m", "required baseline")

	slicePath := filepath.Join(root, "docs", "plans", "slice-"+suffix+".md")
	writeFile(t, slicePath, "---\nexpected_files:\n  - path: src/required.txt\n    op: edit\n---\n")
	manifest := filepath.Join(root, "docs", "plans", "manifest-"+suffix+".md")
	writeFile(t, manifest, "---\ntype: kb-manifest\nkb_id: kb-dirty-"+suffix+"\nslices:\n  - id: slice-001\n    path: docs/plans/slice-"+suffix+".md\n    blockers: []\n    status: pending\n---\n")
	runGitForSliceLease(t, root, "add", "docs/plans")
	runGitForSliceLease(t, root, "commit", "-m", "plan")
	return root, manifest
}
