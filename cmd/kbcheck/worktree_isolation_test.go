package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWorktreePreparePreservesDirtySourceAndWritesReceipt(t *testing.T) {
	t.Parallel()
	root := initWorktreeRepo(t)
	writeFile(t, filepath.Join(root, "dirty.txt"), "source-only\n")
	base := gitOutput(root, "rev-parse", "HEAD")
	owner := "owner-prepare"
	acquireWorktreeTestLease(t, root, "slice-003", owner, base, []string{"src/feature.txt"})

	worktree := filepath.Join(t.TempDir(), "slice-worktree")
	result, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "prepare", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		BaseSHA: base, Worktree: worktree, Branch: "codex/test/slice-003", RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK {
		t.Fatalf("prepare failed result=%#v err=%v", result, err)
	}
	if got := readFile(t, filepath.Join(root, "dirty.txt")); got != "source-only\n" {
		t.Fatalf("dirty source checkout changed: %q", got)
	}
	if result.Receipt == nil || result.Receipt.Status != "prepared" || result.Receipt.Worktree != worktree {
		t.Fatalf("bad receipt: %#v", result.Receipt)
	}
	if !pathExists(filepath.Join(worktree, "src", "base.txt")) {
		t.Fatalf("worktree was not created with repository files")
	}
}

func TestWorktreeIntegrateRequiresOwnerAndStableBase(t *testing.T) {
	t.Parallel()
	root := initWorktreeRepo(t)
	gitOK(t, root, "switch", "-c", "codex/test-coordinator")
	base := gitOutput(root, "rev-parse", "HEAD")
	owner := "owner-integrate"
	acquireWorktreeTestLease(t, root, "slice-003", owner, base, []string{"src/feature.txt"})
	worktree := filepath.Join(t.TempDir(), "slice-worktree")
	prepared, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "prepare", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		BaseSHA: base, Worktree: worktree, Branch: "codex/test/integrate", RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || !prepared.OK {
		t.Fatalf("prepare failed result=%#v err=%v", prepared, err)
	}
	writeFile(t, filepath.Join(worktree, "src", "feature.txt"), "feature\n")
	gitOK(t, worktree, "add", "src/feature.txt")
	gitOK(t, worktree, "commit", "-m", "slice work")

	wrongOwner, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "integrate", SliceID: "slice-003", RunID: "run-1", OwnerToken: "wrong",
		Worktree: worktree, RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || wrongOwner.OK || !strings.Contains(wrongOwner.Issue, "owner token") {
		t.Fatalf("wrong owner was not rejected result=%#v err=%v", wrongOwner, err)
	}

	writeFile(t, filepath.Join(root, "src", "base.txt"), "base drift\n")
	gitOK(t, root, "add", "src/base.txt")
	gitOK(t, root, "commit", "-m", "base drift")
	drift, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "integrate", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		Worktree: worktree, RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || drift.OK || !strings.Contains(drift.Issue, "base revision changed") {
		t.Fatalf("base drift was not rejected result=%#v err=%v", drift, err)
	}
}

func TestWorktreeIntegrateAndReleaseRequiresCleanIntegratedWorktree(t *testing.T) {
	t.Parallel()
	root := initWorktreeRepo(t)
	gitOK(t, root, "switch", "-c", "codex/test-coordinator")
	base := gitOutput(root, "rev-parse", "HEAD")
	owner := "owner-release"
	acquireWorktreeTestLease(t, root, "slice-003", owner, base, []string{"src/feature.txt"})
	worktree := filepath.Join(t.TempDir(), "slice-worktree")
	prepared, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "prepare", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		BaseSHA: base, Worktree: worktree, Branch: "codex/test/release", RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || !prepared.OK {
		t.Fatalf("prepare failed result=%#v err=%v", prepared, err)
	}
	beforeIntegrate, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "release", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		Worktree: worktree, RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || beforeIntegrate.OK || !strings.Contains(beforeIntegrate.Issue, "integrated") {
		t.Fatalf("unintegrated release was not rejected result=%#v err=%v", beforeIntegrate, err)
	}

	writeFile(t, filepath.Join(worktree, "src", "feature.txt"), "feature\n")
	gitOK(t, worktree, "add", "src/feature.txt")
	gitOK(t, worktree, "commit", "-m", "slice work")
	integrated, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "integrate", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		Worktree: worktree, RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || !integrated.OK {
		t.Fatalf("integrate failed result=%#v err=%v", integrated, err)
	}
	if got := strings.ReplaceAll(readFile(t, filepath.Join(root, "src", "feature.txt")), "\r\n", "\n"); got != "feature\n" {
		t.Fatalf("integrated file mismatch: %q", got)
	}
	writeFile(t, filepath.Join(worktree, "scratch.txt"), "dirty\n")
	dirtyRelease, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "release", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		Worktree: worktree, RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || dirtyRelease.OK || !strings.Contains(dirtyRelease.Issue, "dirty") {
		t.Fatalf("dirty release was not rejected result=%#v err=%v", dirtyRelease, err)
	}
	if !pathExists(worktree) {
		t.Fatalf("dirty worktree was removed")
	}
	_ = os.Remove(filepath.Join(worktree, "scratch.txt"))
	released, err := executeWorktreeCommand(worktreeCommandOptions{
		Action: "release", SliceID: "slice-003", RunID: "run-1", OwnerToken: owner,
		Worktree: worktree, RepoRoot: root, LegacyCompatibility: true, Now: time.Now().UTC(),
	})
	if err != nil || !released.OK || released.Receipt.Status != "released" {
		t.Fatalf("release failed result=%#v err=%v", released, err)
	}
	if pathExists(worktree) {
		t.Fatalf("clean integrated worktree was not removed")
	}
}

func TestWorktreeCommandStatusJSON(t *testing.T) {
	t.Parallel()
	root := initWorktreeRepo(t)
	base := gitOutput(root, "rev-parse", "HEAD")
	owner := "owner-command"
	acquireWorktreeTestLease(t, root, "slice-003", owner, base, []string{"src/feature.txt"})
	worktree := filepath.Join(t.TempDir(), "slice-worktree")
	var out, errOut strings.Builder
	code := run([]string{
		"worktree", "--legacy-slice-worktree", "--action", "prepare", "--slice-id", "slice-003", "--run-id", "run-1",
		"--owner-token", owner, "--base-sha", base, "--worktree", worktree, "--branch", "codex/test/command",
		"--root", root, "--json",
	}, &out, &errOut)
	if code != 0 {
		t.Fatalf("prepare command failed code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	out.Reset()
	errOut.Reset()
	code = run([]string{"worktree", "--legacy-slice-worktree", "--action", "status", "--slice-id", "slice-003", "--run-id", "run-1", "--root", root, "--json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"status": "prepared"`) {
		t.Fatalf("status command failed code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func initWorktreeRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	gitOK(t, root, "init")
	// A user-level core.fsmonitor setting can launch a detached daemon for each
	// disposable repository. On Windows that daemon inherits CombinedOutput's
	// pipe handles, so Git appears to hang after the command itself exits.
	gitOK(t, root, "config", "core.fsmonitor", "false")
	gitOK(t, root, "config", "user.email", "kb@example.invalid")
	gitOK(t, root, "config", "user.name", "KB Test")
	writeFile(t, filepath.Join(root, "src", "base.txt"), "base\n")
	gitOK(t, root, "add", ".")
	gitOK(t, root, "commit", "-m", "base")
	return root
}

func acquireWorktreeTestLease(t *testing.T, root, sliceID, owner, base string, files []string) {
	t.Helper()
	result, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", SliceID: sliceID, RunID: "run-1", OwnerToken: owner,
		BaseSHA: base, Files: files, Resources: []string{"git:worktree", "git:integration-owner"},
		RepoRoot: root, Now: time.Now().UTC(),
	})
	if err != nil || !result.OK {
		t.Fatalf("acquire lease failed result=%#v err=%v", result, err)
	}
}

func gitOK(t *testing.T, root string, args ...string) {
	t.Helper()
	if code, out := runGitCommand(root, args...); code != 0 {
		t.Fatalf("git %v failed: %s", args, out)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
