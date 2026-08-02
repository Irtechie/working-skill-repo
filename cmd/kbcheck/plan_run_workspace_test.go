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

func TestPlanRunWorkspaceAdoptsHarnessWorktreeWithoutNesting(t *testing.T) {
	root := initWorktreeRepo(t)
	writePlanRunTestManifest(t, root, "kb-plan-run-adopt")
	session := adoptTestHarnessWorktree(t, root, "codex/harness-session")
	sessionManifest := filepath.Join(session, "docs", "plans", "kb-plan-run-adopt.md")
	head := gitOutput(session, "rev-parse", "HEAD")
	before := strings.Count(gitOutput(root, "worktree", "list"), "\n")

	opts := planRunWorkspaceOptions{
		Action: "adopt", ManifestPath: sessionManifest, OwnerToken: "owner-adopt",
		CommitAuthorized:   true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:adopt",
		RepoRoot: session, Now: time.Now().UTC(),
	}
	result, err := executePlanRunWorkspace(opts)
	if err != nil || !result.OK || result.Receipt == nil {
		t.Fatalf("adopt failed: result=%#v err=%v", result, err)
	}
	receipt := result.Receipt
	if receipt.WorkspaceOwner != planRunWorkspaceOwnerHarness {
		t.Fatalf("adopted receipt is not harness-owned: %#v", receipt)
	}
	if !samePath(receipt.Worktree, session) || !samePath(receipt.SourceCheckout, session) {
		t.Fatalf("adopted receipt does not bind the current worktree: %#v", receipt)
	}
	if receipt.IntegrationRef != "codex/harness-session" || receipt.IntegrationHead != head || receipt.BaseSHA != head {
		t.Fatalf("adopted receipt lineage is wrong: %#v", receipt)
	}
	if after := strings.Count(gitOutput(root, "worktree", "list"), "\n"); after != before {
		t.Fatalf("adopt created a nested worktree: before=%d after=%d", before, after)
	}

	second, err := executePlanRunWorkspace(opts)
	if err != nil || !second.OK || second.Receipt == nil || !samePath(second.Receipt.Worktree, session) {
		t.Fatalf("idempotent adopt failed: result=%#v err=%v", second, err)
	}

	prepare := opts
	prepare.Action = "prepare"
	conflict, err := executePlanRunWorkspace(prepare)
	if err != nil || conflict.OK || !strings.Contains(conflict.Issue, "workspace ownership") {
		t.Fatalf("prepare silently replaced a harness-owned receipt: result=%#v err=%v", conflict, err)
	}
}

func TestPlanRunWorkspaceAdoptRejectsPrimaryDefaultAndDirtyCheckouts(t *testing.T) {
	root := initWorktreeRepo(t)
	manifest := writePlanRunTestManifest(t, root, "kb-plan-run-adopt-guard")
	base := planRunWorkspaceOptions{
		Action: "adopt", ManifestPath: manifest, OwnerToken: "owner-guard",
		CommitAuthorized:   true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:adopt-guard",
		RepoRoot: root, Now: time.Now().UTC(),
	}

	primary, err := executePlanRunWorkspace(base)
	if err != nil || primary.OK || !strings.Contains(primary.Issue, "linked worktree") {
		t.Fatalf("adopt accepted the primary checkout: result=%#v err=%v", primary, err)
	}

	defaultBranch := gitOutput(root, "branch", "--show-current")
	// The primary checkout already holds one resolved default branch, so park a
	// linked worktree on the other conventional default name. A harness worktree
	// sitting on any resolved default must never become a slice commit target.
	otherDefault := "main"
	if defaultBranch == "main" {
		otherDefault = "master"
	}
	onDefaultRoot := adoptTestHarnessWorktreeAt(t, root, otherDefault, "default-parked")
	onDefault := base
	onDefault.RepoRoot = onDefaultRoot
	onDefault.ManifestPath = filepath.Join(onDefaultRoot, "docs", "plans", "kb-plan-run-adopt-guard.md")
	blockedDefault, err := executePlanRunWorkspace(onDefault)
	if err != nil || blockedDefault.OK || !strings.Contains(blockedDefault.Issue, "default branch") {
		t.Fatalf("adopt accepted a default-branch worktree: result=%#v err=%v", blockedDefault, err)
	}

	dirtySession := adoptTestHarnessWorktree(t, root, "codex/guard-dirty")
	writeFile(t, filepath.Join(dirtySession, "src", "base.txt"), "session wip\n")
	dirty := base
	dirty.RepoRoot = dirtySession
	dirty.ManifestPath = filepath.Join(dirtySession, "docs", "plans", "kb-plan-run-adopt-guard.md")
	blockedDirty, err := executePlanRunWorkspace(dirty)
	if err != nil || blockedDirty.OK || !strings.Contains(blockedDirty.Issue, "must be clean") {
		t.Fatalf("adopt accepted a dirty worktree: result=%#v err=%v", blockedDirty, err)
	}
	if got := string(mustReadFile(t, filepath.Join(dirtySession, "src", "base.txt"))); got != "session wip\n" {
		t.Fatalf("blocked adopt mutated session dirt: %q", got)
	}
}

func TestPlanRunWorkspaceReleaseKeepsHarnessOwnedWorktree(t *testing.T) {
	root := initWorktreeRepo(t)
	writePlanRunTestManifest(t, root, "kb-plan-run-adopt-release")
	session := adoptTestHarnessWorktree(t, root, "codex/release-session")
	sessionManifest := filepath.Join(session, "docs", "plans", "kb-plan-run-adopt-release.md")

	opts := planRunWorkspaceOptions{
		Action: "adopt", ManifestPath: sessionManifest, OwnerToken: "owner-release",
		CommitAuthorized:   true,
		CommitAuthorizedBy: "test-user", CommitApprovalRef: "test:adopt-release",
		RepoRoot: session, Now: time.Now().UTC(),
	}
	adopted, err := executePlanRunWorkspace(opts)
	if err != nil || !adopted.OK {
		t.Fatalf("adopt failed: result=%#v err=%v", adopted, err)
	}

	receipt := *adopted.Receipt
	receipt.Status = "completed"
	receipt.OwnerToken = "owner-release"
	if err := savePlanRunWorkspaceReceipt(opts, receipt); err != nil {
		t.Fatal(err)
	}

	opts.Action = "release"
	released, err := executePlanRunWorkspace(opts)
	if err != nil || !released.OK || released.Receipt == nil {
		t.Fatalf("release of adopted workspace failed: result=%#v err=%v", released, err)
	}
	if released.Receipt.CleanupState != "harness-owned" {
		t.Fatalf("adopted release did not return ownership to the harness: %#v", released.Receipt)
	}
	if !pathExists(session) || gitOutput(session, "rev-parse", "HEAD") == "" {
		t.Fatal("release deleted a harness-owned worktree")
	}
	if gitOutput(root, "show-ref", "--verify", "refs/heads/codex/release-session") == "" {
		t.Fatal("release deleted the harness-owned branch")
	}
}

func adoptTestHarnessWorktree(t *testing.T, root, branch string) string {
	t.Helper()
	return adoptTestHarnessWorktreeAt(t, root, branch, safePathPart(branch))
}

func adoptTestHarnessWorktreeAt(t *testing.T, root, branch, name string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if gitOutput(root, "show-ref", "--verify", "refs/heads/"+branch) == "" {
		gitOK(t, root, "worktree", "add", "-b", branch, path, "HEAD")
	} else {
		gitOK(t, root, "worktree", "add", "--detach", path, "HEAD")
		gitOK(t, path, "checkout", branch)
	}
	return path
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
