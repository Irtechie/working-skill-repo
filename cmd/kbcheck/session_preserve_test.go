package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newPreserveRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitOK(t, root, "init", "--initial-branch=main")
	gitOK(t, root, "config", "user.email", "kb@example.test")
	gitOK(t, root, "config", "user.name", "KB Test")
	writePreserveFile(t, root, "README.md", "base\n")
	gitOK(t, root, "add", "README.md")
	gitOK(t, root, "commit", "-m", "base")
	return root
}

func writePreserveFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func preserveOpts(root, action string) sessionPreserveOptions {
	return sessionPreserveOptions{
		Action:    action,
		SessionID: "session-under-test",
		RepoRoot:  root,
		Worktree:  root,
		Now:       time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC),
	}
}

func TestSessionPreserveRefusesOnDefaultBranch(t *testing.T) {
	root := newPreserveRepo(t)
	writePreserveFile(t, root, "src/feature.go", "package src\n")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err == nil {
		t.Fatal("expected refusal on the shared default branch")
	}
	if result.Status != "refused" || result.RefusalReason != "default-branch" {
		t.Fatalf("expected default-branch refusal, got %#v", result)
	}
	if result.CommitSHA != "" {
		t.Fatal("refusal must not create a commit")
	}
	if head := gitOutput(root, "log", "--oneline"); strings.Count(head, "\n") != 0 {
		t.Fatalf("refusal must leave history untouched: %q", head)
	}
}

func TestSessionPreserveCommitsOnSessionBranchWithoutPushing(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	writePreserveFile(t, root, "src/feature.go", "package src\n")
	writePreserveFile(t, root, "README.md", "base\nedited\n")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err != nil {
		t.Fatalf("preserve failed: %v result=%#v", err, result)
	}
	if result.Status != "preserved" {
		t.Fatalf("expected preserved, got %q", result.Status)
	}
	if result.Pushed {
		t.Fatal("session-preserve must never push")
	}
	if len(result.Preserved) != 2 {
		t.Fatalf("expected both files preserved, got %v", result.Preserved)
	}
	if result.CommitSHA == "" || result.CommitSHA == result.BaseSHA {
		t.Fatal("expected a new commit distinct from the base")
	}
	if status := gitOutput(root, "status", "--porcelain"); status != "" {
		t.Fatalf("worktree should be clean after preserve, got %q", status)
	}
	if branch := gitOutput(root, "rev-parse", "--abbrev-ref", "HEAD"); branch != "session-work" {
		t.Fatalf("preserve must not change branches, got %q", branch)
	}
	message := gitOutput(root, "log", "-1", "--pretty=%B")
	if !strings.Contains(message, "not a delivery") {
		t.Fatalf("commit message must disclaim delivery, got %q", message)
	}
}

func TestSessionPreserveExcludesBuildArtifactsButKeepsSource(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	writePreserveFile(t, root, "src/feature.go", "package src\n")
	writePreserveFile(t, root, "kbcheck.exe", "MZ binary payload\n")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err != nil {
		t.Fatalf("preserve failed: %v", err)
	}
	for _, preserved := range result.Preserved {
		if strings.HasSuffix(preserved, ".exe") {
			t.Fatalf("build artifact must not be committed: %v", result.Preserved)
		}
	}
	if len(result.Excluded) != 1 || result.Excluded[0].Reason != "build-artifact" {
		t.Fatalf("expected one build-artifact exclusion, got %#v", result.Excluded)
	}
	// The excluded artifact must remain on disk and must remain untracked.
	if _, err := os.Stat(filepath.Join(root, "kbcheck.exe")); err != nil {
		t.Fatal("excluded artifact must stay on disk")
	}
	if tracked := gitOutput(root, "ls-files", "kbcheck.exe"); tracked != "" {
		t.Fatal("excluded artifact must remain untracked")
	}
}

func TestSessionPreserveExcludesOversizedFiles(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	writePreserveFile(t, root, "huge.log", strings.Repeat("x", int(sessionPreserveMaxBytes)+1))
	writePreserveFile(t, root, "src/feature.go", "package src\n")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err != nil {
		t.Fatalf("preserve failed: %v", err)
	}
	if len(result.Excluded) != 1 || result.Excluded[0].Reason != "oversized" {
		t.Fatalf("expected one oversized exclusion, got %#v", result.Excluded)
	}
	if result.Excluded[0].Bytes <= sessionPreserveMaxBytes {
		t.Fatalf("oversized exclusion must report real size, got %d", result.Excluded[0].Bytes)
	}
	if len(result.Preserved) != 1 || result.Preserved[0] != "src/feature.go" {
		t.Fatalf("source must still be preserved, got %v", result.Preserved)
	}
}

func TestSessionPreserveIgnoresGitignoredFiles(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	writePreserveFile(t, root, ".gitignore", "secrets/\n")
	gitOK(t, root, "add", ".gitignore")
	gitOK(t, root, "commit", "-m", "ignore secrets")
	writePreserveFile(t, root, "secrets/token.txt", "do-not-commit\n")
	writePreserveFile(t, root, "src/feature.go", "package src\n")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err != nil {
		t.Fatalf("preserve failed: %v", err)
	}
	for _, preserved := range result.Preserved {
		if strings.HasPrefix(preserved, "secrets/") {
			t.Fatalf("gitignored path must never be preserved, got %v", result.Preserved)
		}
	}
}

func TestSessionPreserveIsNoOpOnCleanTree(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err != nil {
		t.Fatalf("clean tree must not error: %v", err)
	}
	if result.Status != "nothing-to-preserve" {
		t.Fatalf("expected nothing-to-preserve, got %q", result.Status)
	}
	if result.CommitSHA != "" {
		t.Fatal("clean tree must not produce an empty commit")
	}
}

func TestSessionPreservePlanDoesNotMutate(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	writePreserveFile(t, root, "src/feature.go", "package src\n")
	before := gitOutput(root, "rev-parse", "HEAD")

	result, err := executeSessionPreserve(preserveOpts(root, "plan"))
	if err != nil {
		t.Fatalf("plan failed: %v", err)
	}
	if result.Status != "would-preserve" || len(result.Preserved) != 1 {
		t.Fatalf("expected a would-preserve forecast, got %#v", result)
	}
	if after := gitOutput(root, "rev-parse", "HEAD"); after != before {
		t.Fatal("plan must not move HEAD")
	}
	if status := gitOutput(root, "status", "--porcelain"); status == "" {
		t.Fatal("plan must leave the dirty tree intact")
	}
}

func TestSessionPreserveRefusesDetachedHead(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "--detach", "HEAD")
	writePreserveFile(t, root, "src/feature.go", "package src\n")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err == nil {
		t.Fatal("expected refusal on detached HEAD")
	}
	if result.RefusalReason != "detached-head" {
		t.Fatalf("expected detached-head refusal, got %q", result.RefusalReason)
	}
}

func TestSessionPreserveRefusesBranchMismatch(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	writePreserveFile(t, root, "src/feature.go", "package src\n")

	opts := preserveOpts(root, "apply")
	opts.Branch = "some-other-branch"
	result, err := executeSessionPreserve(opts)
	if err == nil {
		t.Fatal("expected refusal when the requested branch does not match")
	}
	if result.RefusalReason != "branch-mismatch" {
		t.Fatalf("expected branch-mismatch refusal, got %q", result.RefusalReason)
	}
}

func TestSessionPreserveRefusesMidMerge(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	writePreserveFile(t, root, "conflict.txt", "session side\n")
	gitOK(t, root, "add", "conflict.txt")
	gitOK(t, root, "commit", "-m", "session change")
	gitOK(t, root, "checkout", "main")
	writePreserveFile(t, root, "conflict.txt", "main side\n")
	gitOK(t, root, "add", "conflict.txt")
	gitOK(t, root, "commit", "-m", "main change")
	gitOK(t, root, "checkout", "session-work")
	// Deliberately conflicting merge leaves MERGE_HEAD in place.
	runGitCommand(root, "merge", "main")

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err == nil {
		t.Fatal("expected refusal during an in-progress merge")
	}
	if result.RefusalReason != "merge" {
		t.Fatalf("expected merge refusal, got %q", result.RefusalReason)
	}
}

func TestSessionPreserveRequiresSessionIdentity(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")

	opts := preserveOpts(root, "apply")
	opts.SessionID = ""
	result, err := executeSessionPreserve(opts)
	if err == nil {
		t.Fatal("expected refusal without a session id")
	}
	if result.RefusalReason != "missing-session-id" {
		t.Fatalf("expected missing-session-id refusal, got %q", result.RefusalReason)
	}
}

func TestSessionPreserveHandlesDeletions(t *testing.T) {
	root := newPreserveRepo(t)
	gitOK(t, root, "checkout", "-b", "session-work")
	if err := os.Remove(filepath.Join(root, "README.md")); err != nil {
		t.Fatal(err)
	}

	result, err := executeSessionPreserve(preserveOpts(root, "apply"))
	if err != nil {
		t.Fatalf("preserve failed: %v result=%#v", err, result)
	}
	if result.Status != "preserved" {
		t.Fatalf("expected deletion to be preserved, got %q", result.Status)
	}
	if tracked := gitOutput(root, "ls-files", "README.md"); tracked != "" {
		t.Fatal("deletion should be recorded in the preserved commit")
	}
}
