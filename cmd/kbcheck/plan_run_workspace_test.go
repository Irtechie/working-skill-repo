package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPlanRunWorkspaceRequiresExplicitLocalCommitAuthorization(t *testing.T) {
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, "kb-plan-run-authorization")
	worktree := filepath.Join(t.TempDir(), "plan-run-authorization")

	result, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner-authorization",
		BaseSHA: gitOutput(root, "rev-parse", "HEAD"), Worktree: worktree,
		IntegrationRef: "codex/plan-run-authorization", RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "commit authorization") {
		t.Fatalf("prepare without commit authorization was not blocked: result=%#v err=%v", result, err)
	}
	if pathExists(worktree) || gitOutput(root, "show-ref", "--verify", "refs/heads/codex/plan-run-authorization") != "" {
		t.Fatal("unauthorized prepare created a worktree or branch")
	}
}

func TestPlanRunWorkspacePrepareIsIdempotentAndPreservesDirtySource(t *testing.T) {
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, "kb-plan-run-one")
	baseSHA := gitOutput(root, "rev-parse", "HEAD")
	baseRef := gitOutput(root, "branch", "--show-current")
	sourceFile := filepath.Join(root, "README.md")
	if err := os.WriteFile(sourceFile, []byte("dirty source\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	worktree := filepath.Join(t.TempDir(), "plan-run-one")
	opts := planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner-one",
		CommitAuthorized:   true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:plan-run-one",
		BaseSHA: baseSHA, Worktree: worktree, IntegrationRef: "codex/plan-run-one",
		RepoRoot: root, Now: time.Now().UTC(),
	}

	first, err := executePlanRunWorkspace(opts)
	if err != nil || !first.OK || first.Receipt == nil {
		t.Fatalf("prepare failed: result=%#v err=%v", first, err)
	}
	receipt := first.Receipt
	if receipt.KBID != "kb-plan-run-one" || receipt.ManifestPath != manifest ||
		receipt.BaseRef != baseRef || receipt.BaseSHA != baseSHA ||
		receipt.IntegrationRef != "codex/plan-run-one" ||
		receipt.IntegrationHead != baseSHA || receipt.SourceCheckout != root ||
		receipt.Worktree != worktree || receipt.Status != "prepared" {
		t.Fatalf("incomplete plan-run receipt: %#v", receipt)
	}
	if got := string(mustReadFile(t, sourceFile)); got != "dirty source\n" {
		t.Fatalf("source dirt was changed: %q", got)
	}
	worktreeSourceFile := filepath.Join(worktree, "README.md")
	if pathExists(worktreeSourceFile) {
		if got := string(mustReadFile(t, worktreeSourceFile)); got == "dirty source\n" {
			t.Fatal("dirty source content was silently copied into plan-run worktree")
		}
	}

	second, err := executePlanRunWorkspace(opts)
	if err != nil || !second.OK || second.Receipt == nil ||
		second.Receipt.Worktree != receipt.Worktree ||
		second.Receipt.IntegrationRef != receipt.IntegrationRef {
		t.Fatalf("idempotent prepare failed: result=%#v err=%v", second, err)
	}
}

func TestPlanRunWorkspaceRejectsDefaultBranchOwnerMismatchAndUnsafeRelease(t *testing.T) {
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, "kb-plan-run-two")
	baseSHA := gitOutput(root, "rev-parse", "HEAD")
	baseRef := gitOutput(root, "branch", "--show-current")

	defaultTarget, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner-one",
		CommitAuthorized:   true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:default-boundary",
		BaseSHA: baseSHA, Worktree: filepath.Join(t.TempDir(), "default-target"),
		IntegrationRef: baseRef, RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || defaultTarget.OK || !strings.Contains(defaultTarget.Issue, "default branch") {
		t.Fatalf("default branch was accepted: result=%#v err=%v", defaultTarget, err)
	}

	opts := planRunWorkspaceOptions{
		Action: "prepare", ManifestPath: manifest, OwnerToken: "owner-one",
		CommitAuthorized:   true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:plan-run-two",
		BaseSHA: baseSHA, Worktree: filepath.Join(t.TempDir(), "owned"),
		IntegrationRef: "codex/plan-run-two", RepoRoot: root, Now: time.Now().UTC(),
	}
	prepared, err := executePlanRunWorkspace(opts)
	if err != nil || !prepared.OK {
		t.Fatalf("prepare failed: result=%#v err=%v", prepared, err)
	}
	wrongOwner := opts
	wrongOwner.Action = "status"
	wrongOwner.OwnerToken = "owner-two"
	status, err := executePlanRunWorkspace(wrongOwner)
	if err != nil || status.OK || !strings.Contains(status.Issue, "owner") {
		t.Fatalf("owner mismatch was accepted: result=%#v err=%v", status, err)
	}
	opts.Action = "release"
	released, err := executePlanRunWorkspace(opts)
	if err != nil || released.OK || !strings.Contains(released.Issue, "incomplete") {
		t.Fatalf("active incomplete workspace was released: result=%#v err=%v", released, err)
	}
}

func writePlanRunTestManifest(t *testing.T, root, kbID string) string {
	t.Helper()
	path := filepath.Join(root, "docs", "plans", kbID+".md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\ntype: kb-manifest\nkb_id: " + kbID + "\n---\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOK(t, root, "add", filepath.ToSlash(strings.TrimPrefix(path, root+string(filepath.Separator))))
	gitOK(t, root, "commit", "-m", "add plan manifest")
	return path
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
