package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestTerminalCleanupDefersCurrentActiveAndDirtyWorktree(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "local-guards")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "terminal-worktree-cleanup", SessionID: "session-1",
		Branch: branch, Worktree: worktree, Status: "in_progress",
	})

	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "terminal-worktree-cleanup", SessionID: "session-1",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK || registered.Receipt == nil {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}

	active, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || active.OK || !strings.Contains(active.Issue, "active queue claim") {
		t.Fatalf("active claim was not protected: result=%#v err=%v", active, err)
	}
	if !pathExists(worktree) {
		t.Fatal("active worktree was removed")
	}

	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "terminal-worktree-cleanup", SessionID: "session-1",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	currentSubdir := filepath.Join(worktree, "nested")
	if err := os.MkdirAll(currentSubdir, 0o755); err != nil {
		t.Fatal(err)
	}
	current, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: currentSubdir, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || current.OK || !strings.Contains(current.Issue, "executing worktree") {
		t.Fatalf("current worktree was not protected: result=%#v err=%v", current, err)
	}

	writeFile(t, filepath.Join(worktree, "dirty.txt"), "preserve me\n")
	dirty, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || dirty.OK || !strings.Contains(dirty.Issue, "dirty") {
		t.Fatalf("dirty worktree was not protected: result=%#v err=%v", dirty, err)
	}
	if !pathExists(filepath.Join(worktree, "dirty.txt")) {
		t.Fatal("dirty file was removed")
	}

	if err := os.Remove(filepath.Join(worktree, "dirty.txt")); err != nil {
		t.Fatal(err)
	}
	clean, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || !clean.OK || clean.Receipt == nil || clean.Receipt.Status != "released" {
		t.Fatalf("clean local worktree was not removed: result=%#v err=%v", clean, err)
	}
	if pathExists(worktree) {
		t.Fatal("clean terminal worktree still exists")
	}
	if got := gitOutput(root, "rev-parse", "refs/heads/"+branch); got != commit {
		t.Fatalf("local-only durable branch was not retained: got=%s want=%s", got, commit)
	}
}

func TestTerminalCleanupPreservesIgnoredFiles(t *testing.T) {
	root := initWorktreeRepo(t)
	writeFile(t, filepath.Join(root, ".gitignore"), "private.env\n")
	runGitForSliceLease(t, root, "add", ".gitignore")
	runGitForSliceLease(t, root, "commit", "-m", "ignore private state")
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "ignored")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "ignored-cleanup", SessionID: "session-ignored",
		Branch: branch, Worktree: worktree, Status: "in_progress",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "ignored-cleanup", SessionID: "session-ignored",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}
	ignored := filepath.Join(worktree, "private.env")
	writeFile(t, ignored, "secret local state\n")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "ignored-cleanup", SessionID: "session-ignored",
		Branch: branch, Worktree: worktree, Status: "done",
	})

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "ignored files") {
		t.Fatalf("ignored file was not protected: result=%#v err=%v", result, err)
	}
	if !pathExists(ignored) {
		t.Fatal("ignored file was deleted")
	}
}

func TestTerminalCleanupPreservesCurrentSessionByIdentity(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "same-session")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "same-session-cleanup", SessionID: "session-current",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "same-session-cleanup", SessionID: "session-current",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root,
		CurrentSession: "session-current", Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "own worktree") {
		t.Fatalf("current session identity did not block cleanup: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) {
		t.Fatal("current session worktree was removed")
	}
}

func TestTerminalCleanupSweepTextReportsPartialMutation(t *testing.T) {
	var output strings.Builder
	writeTerminalCleanupSweepText(&output, terminalCleanupResult{
		OK: false, Action: "sweep", Issue: "active queue claim",
		Scanned: 2, Cleaned: 1,
		Receipt: &terminalCleanupReceipt{SessionID: "session-blocked", Worktree: `C:\worktrees\blocked`},
	})
	text := output.String()
	for _, want := range []string{"blocked", "scanned=2", "cleaned=1", "session=session-blocked", `path=C:\worktrees\blocked`, "active queue claim"} {
		if !strings.Contains(text, want) {
			t.Fatalf("sweep output %q missing %q", text, want)
		}
	}
}

func TestTerminalCleanupRequiresRemoteContainmentAndRetainsPRRefs(t *testing.T) {
	root := initWorktreeRepo(t)
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	defaultBranch := gitOutput(root, "branch", "--show-current")
	runGitForSliceLease(t, root, "push", "-u", "origin", defaultBranch)
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch)
	runGitForSliceLease(t, root, "fetch", "origin")
	runGitForSliceLease(t, root, "remote", "set-head", "origin", "-a")

	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "remote-guards")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "remote-cleanup", SessionID: "session-2",
		Branch: branch, Worktree: worktree, Status: "done",
	})

	uncontained, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "remote-cleanup", SessionID: "session-2",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "pr",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || uncontained.OK || !strings.Contains(uncontained.Issue, "remote topic") {
		t.Fatalf("unpushed PR work registered: result=%#v err=%v", uncontained, err)
	}

	runGitForSliceLease(t, worktree, "push", "-u", "origin", branch)
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "remote-cleanup", SessionID: "session-2",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "pr",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("pushed PR work did not register: result=%#v err=%v", registered, err)
	}
	removed, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || !removed.OK || removed.Receipt == nil || removed.Receipt.Status != "released" {
		t.Fatalf("pushed PR worktree was not removed: result=%#v err=%v", removed, err)
	}
	if got := gitOutput(root, "rev-parse", "refs/heads/"+branch); got != commit {
		t.Fatalf("unmerged PR local ref was deleted: got=%s want=%s", got, commit)
	}
	if got := gitOutput(root, "show-ref", "--verify", "refs/remotes/origin/"+branch); got == "" {
		t.Fatal("remote feature ref was deleted without a race-safe remote CAS")
	}
}

func TestTerminalCleanupDirectIntegrationDeletesOnlyMergedLocalRef(t *testing.T) {
	root := initWorktreeRepo(t)
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	defaultBranch := gitOutput(root, "branch", "--show-current")
	runGitForSliceLease(t, root, "push", "-u", "origin", defaultBranch)
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch)

	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "direct-integrated")
	runGitForSliceLease(t, root, "merge", "--no-ff", "--no-edit", branch)
	runGitForSliceLease(t, root, "push", "origin", defaultBranch)
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "direct-cleanup", SessionID: "session-direct",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "direct-cleanup", SessionID: "session-direct",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "direct",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("direct register failed: result=%#v err=%v", registered, err)
	}
	cleaned, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || !cleaned.OK || cleaned.Receipt == nil || cleaned.Receipt.Status != "released" {
		t.Fatalf("direct cleanup failed: result=%#v err=%v", cleaned, err)
	}
	if pathExists(worktree) {
		t.Fatal("integrated worktree still exists")
	}
	if got := gitOutput(root, "show-ref", "--verify", "refs/heads/"+branch); got != "" {
		t.Fatalf("merged local feature ref still exists: %s", got)
	}
}

func TestTerminalCleanupSerializesAgainstQueueClaims(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "queue-lock")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "queue-lock-cleanup", SessionID: "session-lock",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "queue-lock-cleanup", SessionID: "session-lock",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}

	queuePath, err := terminalCleanupQueuePath(root)
	if err != nil {
		t.Fatal(err)
	}
	lock, err := modelrouting.AcquireSharedProjectLock(filepath.Dir(queuePath), "work-queue.lock", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Close()
	_, err = executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper",
		LockTimeout: 100 * time.Millisecond, Now: time.Now().UTC(),
	})
	if err == nil || !strings.Contains(err.Error(), "shared work queue lock") {
		t.Fatalf("cleanup did not serialize against queue lock: %v", err)
	}
	if !pathExists(worktree) {
		t.Fatal("worktree was removed without the shared queue lock")
	}
}

func TestTerminalCleanupLockIsCompatibleWithPowerShell(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell FileShare lock compatibility is Windows-specific")
	}
	powerShell, err := exec.LookPath("powershell")
	if err != nil {
		t.Skip("Windows PowerShell is unavailable")
	}
	lockDir := t.TempDir()
	lockPath := filepath.Join(lockDir, "work-queue.lock")
	readyPath := filepath.Join(lockDir, "ready")
	quote := func(value string) string { return strings.ReplaceAll(value, "'", "''") }
	script := "$lock=[System.IO.File]::Open('" + quote(lockPath) +
		"',[System.IO.FileMode]::OpenOrCreate,[System.IO.FileAccess]::ReadWrite,[System.IO.FileShare]::None);" +
		"try {[System.IO.File]::WriteAllText('" + quote(readyPath) +
		"','ready'); Start-Sleep -Milliseconds 1200} finally {$lock.Dispose()}"
	command := exec.Command(powerShell, "-NoProfile", "-NonInteractive", "-Command", script)
	var output strings.Builder
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if command.ProcessState == nil {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	}()
	deadline := time.Now().Add(3 * time.Second)
	for !pathExists(readyPath) && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if !pathExists(readyPath) {
		t.Fatalf("PowerShell lock did not become ready: %s", output.String())
	}
	if lock, err := modelrouting.AcquireSharedProjectLock(lockDir, "work-queue.lock", 100*time.Millisecond); err == nil {
		_ = lock.Close()
		t.Fatal("Go lock acquired while PowerShell held FileShare.None")
	}
	if err := command.Wait(); err != nil {
		t.Fatalf("PowerShell lock process failed: %v: %s", err, output.String())
	}
	lock, err := modelrouting.AcquireSharedProjectLock(lockDir, "work-queue.lock", time.Second)
	if err != nil {
		t.Fatalf("Go lock did not acquire after PowerShell released it: %v", err)
	}
	_ = lock.Close()
}

func TestTerminalCleanupUsesAuthoritativeRemoteDefault(t *testing.T) {
	root := initWorktreeRepo(t)
	base := gitOutput(root, "rev-parse", "HEAD")
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	defaultBranch := gitOutput(root, "branch", "--show-current")
	runGitForSliceLease(t, root, "push", "-u", "origin", defaultBranch)
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch)
	runGitForSliceLease(t, root, "fetch", "origin")
	runGitForSliceLease(t, root, "remote", "set-head", "origin", "-a")

	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "stale-default")
	runGitForSliceLease(t, root, "merge", "--no-ff", "--no-edit", branch)
	runGitForSliceLease(t, root, "push", "origin", defaultBranch)
	runGitForSliceLease(t, root, "branch", "release", base)
	runGitForSliceLease(t, root, "push", "origin", "release")
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/release")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "stale-default-cleanup", SessionID: "session-stale",
		Branch: branch, Worktree: worktree, Status: "done",
	})

	staleLocalHead, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "stale-default-cleanup", SessionID: "session-stale",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "direct",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || staleLocalHead.OK || !strings.Contains(staleLocalHead.Issue, "remote default does not contain") {
		t.Fatalf("authoritative remote default was not honored: result=%#v err=%v", staleLocalHead, err)
	}
	if got := gitOutput(root, "rev-parse", "refs/heads/"+branch); got != commit {
		t.Fatalf("stale local remote HEAD authorized ref deletion: got=%s want=%s", got, commit)
	}
	if !pathExists(worktree) {
		t.Fatal("stale local remote HEAD authorized worktree removal")
	}
}

func TestTerminalCleanupRejectsNewAuthoritativeRemoteDefaultTarget(t *testing.T) {
	root := initWorktreeRepo(t)
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	oldDefault := gitOutput(root, "branch", "--show-current")
	runGitForSliceLease(t, root, "push", "-u", "origin", oldDefault)
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/"+oldDefault)
	runGitForSliceLease(t, root, "fetch", "origin")
	runGitForSliceLease(t, root, "remote", "set-head", "origin", "-a")

	worktree := filepath.Join(t.TempDir(), "release-worktree")
	runGitForSliceLease(t, root, "worktree", "add", "-b", "release", worktree)
	writeFile(t, filepath.Join(worktree, "release.txt"), "new default\n")
	runGitForSliceLease(t, worktree, "add", "release.txt")
	runGitForSliceLease(t, worktree, "commit", "-m", "prepare release")
	commit := gitOutput(worktree, "rev-parse", "HEAD")
	runGitForSliceLease(t, worktree, "push", "origin", "release")
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/release")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "new-default-cleanup", SessionID: "session-new-default",
		Branch: "release", Worktree: worktree, Status: "done",
	})

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "new-default-cleanup", SessionID: "session-new-default",
		Worktree: worktree, Branch: "release", CommitSHA: commit, DeliveryMode: "direct",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "authoritative remote default") {
		t.Fatalf("new authoritative default target was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) || gitOutput(root, "rev-parse", "refs/heads/release") != commit {
		t.Fatal("new authoritative default target was mutated")
	}
}

func TestTerminalCleanupRejectsMovedWorktree(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "moved")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "moved-cleanup", SessionID: "session-moved",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "moved-cleanup", SessionID: "session-moved",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}
	moved := filepath.Join(filepath.Dir(worktree), "moved-worktree")
	runGitForSliceLease(t, root, "worktree", "move", worktree, moved)
	runGitForSliceLease(t, moved, "switch", "--detach")

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "moved to a different path") {
		t.Fatalf("moved worktree was not preserved: result=%#v err=%v", result, err)
	}
	if !pathExists(moved) || gitOutput(moved, "rev-parse", "HEAD") != commit {
		t.Fatal("moved worktree was mutated")
	}
}

func TestTerminalCleanupRejectsReplacedWorktreeGeneration(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "replaced")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "replaced-cleanup", SessionID: "session-replaced",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "replaced-cleanup", SessionID: "session-replaced",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}
	runGitForSliceLease(t, root, "worktree", "remove", worktree)
	runGitForSliceLease(t, root, "worktree", "add", worktree, branch)

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "generation") {
		t.Fatalf("replacement worktree generation was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) || gitOutput(worktree, "rev-parse", "HEAD") != commit {
		t.Fatal("replacement worktree was mutated")
	}
}

func TestTerminalCleanupRejectsBrokenAdminRoundTrip(t *testing.T) {
	root := initWorktreeRepo(t)
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "broken-roundtrip")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "roundtrip-cleanup", SessionID: "session-roundtrip",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "roundtrip-cleanup", SessionID: "session-roundtrip",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK || registered.Receipt == nil {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}
	writeFile(t, filepath.Join(registered.Receipt.WorktreeGitDir, "gitdir"), filepath.Join(t.TempDir(), ".git")+"\n")

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || result.OK ||
		(!strings.Contains(result.Issue, "round-trip") &&
			!strings.Contains(result.Issue, "not a registered Git worktree")) {
		t.Fatalf("broken admin round-trip was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) {
		t.Fatal("worktree with broken admin round-trip was removed")
	}
}

func TestTerminalCleanupRejectsLockedAndMissingWorktrees(t *testing.T) {
	for _, test := range []struct {
		name      string
		mutate    func(*testing.T, string, string)
		wantIssue string
	}{
		{
			name: "locked",
			mutate: func(t *testing.T, root, worktree string) {
				runGitForSliceLease(t, root, "worktree", "lock", worktree)
			},
			wantIssue: "locked",
		},
		{
			name: "missing",
			mutate: func(t *testing.T, _, worktree string) {
				if err := os.RemoveAll(worktree); err != nil {
					t.Fatal(err)
				}
			},
			wantIssue: "scoped prune or manual recovery",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := initWorktreeRepo(t)
			worktree, branch, commit := createTerminalCleanupWorktree(t, root, test.name)
			workID := test.name + "-cleanup"
			sessionID := "session-" + test.name
			writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
				WorkID: workID, SessionID: sessionID, Branch: branch, Worktree: worktree, Status: "done",
			})
			registered, err := executeTerminalCleanup(terminalCleanupOptions{
				Action: "register", WorkID: workID, SessionID: sessionID,
				Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
				RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
			})
			if err != nil || !registered.OK {
				t.Fatalf("register failed: result=%#v err=%v", registered, err)
			}
			test.mutate(t, root, worktree)

			result, err := executeTerminalCleanup(terminalCleanupOptions{
				Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
			})
			if err != nil || result.OK || !strings.Contains(result.Issue, test.wantIssue) {
				t.Fatalf("%s worktree was accepted: result=%#v err=%v", test.name, result, err)
			}
		})
	}
}

func TestTerminalCleanupRejectsRewrittenRemoteDefaultEvidence(t *testing.T) {
	root := initWorktreeRepo(t)
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	defaultBranch := gitOutput(root, "branch", "--show-current")
	runGitForSliceLease(t, root, "push", "-u", "origin", defaultBranch)
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch)

	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "rewritten-default")
	runGitForSliceLease(t, root, "merge", "--no-ff", "--no-edit", branch)
	runGitForSliceLease(t, root, "push", "origin", defaultBranch)
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "rewrite-cleanup", SessionID: "session-rewrite",
		Branch: branch, Worktree: worktree, Status: "done",
	})
	registered, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "rewrite-cleanup", SessionID: "session-rewrite",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "direct",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || !registered.OK {
		t.Fatalf("register failed: result=%#v err=%v", registered, err)
	}

	tree := gitOutput(root, "rev-parse", "HEAD^{tree}")
	rewritten := gitOutput(root, "commit-tree", tree, "-m", "rewrite remote default")
	runGitForSliceLease(t, root, "push", "--force", "origin", rewritten+":refs/heads/"+defaultBranch)
	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "history was rewritten") {
		t.Fatalf("rewritten remote evidence was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) || gitOutput(root, "rev-parse", "refs/heads/"+branch) != commit {
		t.Fatal("rewritten remote evidence mutated recoverable work")
	}
}

func TestTerminalCleanupRefreshesRemoteEvidenceBetweenReceipts(t *testing.T) {
	root := initWorktreeRepo(t)
	configureTerminalCleanupRemote(t, root)
	defaultBranch := gitOutput(root, "branch", "--show-current")
	base := gitOutput(root, "rev-parse", "HEAD")
	firstWorktree, firstBranch, firstCommit := createTerminalCleanupWorktree(t, root, "batch-first")
	runGitForSliceLease(t, root, "merge", "--no-ff", "--no-edit", firstBranch)
	secondWorktree, secondBranch, secondCommit := createTerminalCleanupWorktree(t, root, "batch-second")
	runGitForSliceLease(t, root, "merge", "--no-ff", "--no-edit", secondBranch)
	runGitForSliceLease(t, root, "push", "origin", defaultBranch)
	queue := []terminalCleanupQueueEntry{
		{
			WorkID: "batch-a", SessionID: "session-a", Branch: firstBranch,
			Worktree: firstWorktree, Status: "done",
		},
		{
			WorkID: "batch-b", SessionID: "session-b", Branch: secondBranch,
			Worktree: secondWorktree, Status: "done",
		},
	}
	writeTerminalCleanupQueue(t, root, queue...)
	for _, target := range []struct {
		workID, sessionID, worktree, branch, commit string
	}{
		{"batch-a", "session-a", firstWorktree, firstBranch, firstCommit},
		{"batch-b", "session-b", secondWorktree, secondBranch, secondCommit},
	} {
		result, err := executeTerminalCleanup(terminalCleanupOptions{
			Action: "register", WorkID: target.workID, SessionID: target.sessionID,
			Worktree: target.worktree, Branch: target.branch, CommitSHA: target.commit,
			DeliveryMode: "direct", Remote: "origin", RepoRoot: root,
			CurrentWorktree: root, Now: time.Now().UTC(),
		})
		if err != nil || !result.OK {
			t.Fatalf("register %s failed: result=%#v err=%v", target.workID, result, err)
		}
	}

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper",
		Now: time.Now().UTC(),
		BeforeReceiptSweep: func(index int, _ terminalCleanupReceipt) {
			if index == 1 {
				runGitForSliceLease(t, root, "push", "--force", "origin", base+":refs/heads/"+defaultBranch)
			}
		},
	})
	if err != nil || result.OK || result.SweepStatus != "partial" ||
		!strings.Contains(result.Issue, "history was rewritten") {
		t.Fatalf("between-receipt rewrite was not blocked: result=%#v err=%v", result, err)
	}
	if pathExists(firstWorktree) || !pathExists(secondWorktree) {
		t.Fatal("batch sweep did not remove only the receipt verified before the rewrite")
	}
	if got := gitOutput(root, "rev-parse", "refs/heads/"+secondBranch); got != secondCommit {
		t.Fatalf("second feature ref changed after remote rewrite: got=%s want=%s", got, secondCommit)
	}
}

func TestTerminalCleanupChecksEveryAuthoritativeRemoteDefault(t *testing.T) {
	root := initWorktreeRepo(t)
	origin := filepath.Join(t.TempDir(), "origin.git")
	upstream := filepath.Join(t.TempDir(), "upstream.git")
	runGitForSliceLease(t, "", "init", "--bare", origin)
	runGitForSliceLease(t, "", "init", "--bare", upstream)
	runGitForSliceLease(t, root, "remote", "add", "origin", origin)
	runGitForSliceLease(t, root, "remote", "add", "upstream", upstream)
	defaultBranch := gitOutput(root, "branch", "--show-current")
	runGitForSliceLease(t, root, "push", "origin", defaultBranch)
	runGitForSliceLease(t, "", "--git-dir", origin, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch)

	worktree := filepath.Join(t.TempDir(), "upstream-default")
	runGitForSliceLease(t, root, "worktree", "add", "-b", "release", worktree)
	writeFile(t, filepath.Join(worktree, "release.txt"), "upstream default\n")
	runGitForSliceLease(t, worktree, "add", "release.txt")
	runGitForSliceLease(t, worktree, "commit", "-m", "release default")
	commit := gitOutput(worktree, "rev-parse", "HEAD")
	runGitForSliceLease(t, worktree, "push", "upstream", "release")
	runGitForSliceLease(t, "", "--git-dir", upstream, "symbolic-ref", "HEAD", "refs/heads/release")
	runGitForSliceLease(t, root, "merge", "--no-ff", "--no-edit", "release")
	runGitForSliceLease(t, root, "push", "origin", defaultBranch)
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "all-remotes-cleanup", SessionID: "session-all-remotes",
		Branch: "release", Worktree: worktree, Status: "done",
	})

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "all-remotes-cleanup", SessionID: "session-all-remotes",
		Worktree: worktree, Branch: "release", CommitSHA: commit, DeliveryMode: "direct",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "authoritative remote default") {
		t.Fatalf("other remote default target was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) || gitOutput(root, "rev-parse", "refs/heads/release") != commit {
		t.Fatal("other remote default target was mutated")
	}
}

func TestTerminalCleanupKeepsBlockedReceiptAssociation(t *testing.T) {
	root := initWorktreeRepo(t)
	blockedWorktree, blockedBranch, blockedCommit := createTerminalCleanupWorktree(t, root, "blocked-first")
	cleanWorktree, cleanBranch, cleanCommit := createTerminalCleanupWorktree(t, root, "clean-second")
	queue := []terminalCleanupQueueEntry{
		{
			WorkID: "work-a", SessionID: "session-a", Branch: blockedBranch,
			Worktree: blockedWorktree, Status: "in_progress",
		},
		{
			WorkID: "work-b", SessionID: "session-b", Branch: cleanBranch,
			Worktree: cleanWorktree, Status: "done",
		},
	}
	writeTerminalCleanupQueue(t, root, queue...)
	for _, target := range []struct {
		workID, sessionID, worktree, branch, commit string
	}{
		{"work-a", "session-a", blockedWorktree, blockedBranch, blockedCommit},
		{"work-b", "session-b", cleanWorktree, cleanBranch, cleanCommit},
	} {
		result, err := executeTerminalCleanup(terminalCleanupOptions{
			Action: "register", WorkID: target.workID, SessionID: target.sessionID,
			Worktree: target.worktree, Branch: target.branch, CommitSHA: target.commit,
			DeliveryMode: "local", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
		})
		if err != nil || !result.OK {
			t.Fatalf("register %s failed: result=%#v err=%v", target.workID, result, err)
		}
	}

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "sweep", RepoRoot: root, CurrentWorktree: root, CurrentSession: "sweeper", Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "active queue claim") ||
		result.Receipt == nil || result.Receipt.WorkID != "work-a" {
		t.Fatalf("blocked issue lost its receipt association: result=%#v err=%v", result, err)
	}
	if !pathExists(blockedWorktree) || pathExists(cleanWorktree) {
		t.Fatal("mixed sweep did not preserve blocked and remove eligible worktree")
	}
	if result.SweepStatus != "partial" || len(result.Outcomes) != 2 ||
		result.Outcomes[0].State != "blocked" || result.Outcomes[1].State != "removed" {
		t.Fatalf("mixed sweep ledger is incomplete: %#v", result)
	}
}

func TestTerminalCleanupFailsClosedWhenRemoteDefaultIsUnresolved(t *testing.T) {
	root := initWorktreeRepo(t)
	remote := filepath.Join(t.TempDir(), "remote.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	worktree, branch, commit := createTerminalCleanupWorktree(t, root, "missing-default")
	runGitForSliceLease(t, worktree, "push", "origin", branch)
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "missing-default-cleanup", SessionID: "session-missing",
		Branch: branch, Worktree: worktree, Status: "done",
	})

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "missing-default-cleanup", SessionID: "session-missing",
		Worktree: worktree, Branch: branch, CommitSHA: commit, DeliveryMode: "direct",
		Remote: "origin", RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "authority is unresolved") {
		t.Fatalf("unresolved remote default was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) {
		t.Fatal("unresolved remote default removed worktree")
	}
}

func TestTerminalCleanupRefusesPrimaryAndDefaultTargets(t *testing.T) {
	root := initWorktreeRepo(t)
	branch := gitOutput(root, "branch", "--show-current")
	commit := gitOutput(root, "rev-parse", "HEAD")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "primary-cleanup", SessionID: "session-3",
		Branch: branch, Worktree: root, Status: "done",
	})

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "primary-cleanup", SessionID: "session-3",
		Worktree: root, Branch: branch, CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: filepath.Join(t.TempDir(), "other"), Now: time.Now().UTC(),
	})
	if err != nil || result.OK ||
		(!strings.Contains(result.Issue, "primary checkout") && !strings.Contains(result.Issue, "default branch")) {
		t.Fatalf("primary/default target was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(root) {
		t.Fatal("primary checkout was removed")
	}
}

func TestTerminalCleanupFailsClosedWithoutAuthoritativeDefault(t *testing.T) {
	root := t.TempDir()
	runGitForSliceLease(t, root, "init", "-b", "trunk")
	runGitForSliceLease(t, root, "config", "core.fsmonitor", "false")
	runGitForSliceLease(t, root, "config", "user.email", "kb@example.invalid")
	runGitForSliceLease(t, root, "config", "user.name", "KB Test")
	writeFile(t, filepath.Join(root, "base.txt"), "base\n")
	runGitForSliceLease(t, root, "add", "base.txt")
	runGitForSliceLease(t, root, "commit", "-m", "base")
	runGitForSliceLease(t, root, "switch", "-c", "side")
	worktree := filepath.Join(t.TempDir(), "trunk-worktree")
	runGitForSliceLease(t, root, "worktree", "add", worktree, "trunk")
	commit := gitOutput(worktree, "rev-parse", "HEAD")
	writeTerminalCleanupQueue(t, root, terminalCleanupQueueEntry{
		WorkID: "no-authority-cleanup", SessionID: "session-no-authority",
		Branch: "trunk", Worktree: worktree, Status: "done",
	})

	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action: "register", WorkID: "no-authority-cleanup", SessionID: "session-no-authority",
		Worktree: worktree, Branch: "trunk", CommitSHA: commit, DeliveryMode: "local",
		RepoRoot: root, CurrentWorktree: root, Now: time.Now().UTC(),
	})
	if err != nil || result.OK || !strings.Contains(result.Issue, "no configured remotes") {
		t.Fatalf("local nonstandard default without authority was accepted: result=%#v err=%v", result, err)
	}
	if !pathExists(worktree) || gitOutput(root, "rev-parse", "refs/heads/trunk") != commit {
		t.Fatal("unresolved local default authority mutated recovery state")
	}
}

func createTerminalCleanupWorktree(t *testing.T, root, suffix string) (string, string, string) {
	t.Helper()
	configureTerminalCleanupRemote(t, root)
	branch := "feature/" + suffix
	worktree := filepath.Join(t.TempDir(), "worktree")
	runGitForSliceLease(t, root, "worktree", "add", "-b", branch, worktree)
	writeFile(t, filepath.Join(worktree, "delivered.txt"), suffix+"\n")
	runGitForSliceLease(t, worktree, "add", "delivered.txt")
	runGitForSliceLease(t, worktree, "commit", "-m", "deliver "+suffix)
	return worktree, branch, gitOutput(worktree, "rev-parse", "HEAD")
}

func configureTerminalCleanupRemote(t *testing.T, root string) {
	t.Helper()
	if strings.TrimSpace(gitOutput(root, "remote")) != "" {
		return
	}
	remote := filepath.Join(t.TempDir(), "origin.git")
	runGitForSliceLease(t, "", "init", "--bare", remote)
	runGitForSliceLease(t, root, "remote", "add", "origin", remote)
	defaultBranch := gitOutput(root, "branch", "--show-current")
	runGitForSliceLease(t, root, "push", "-u", "origin", defaultBranch)
	runGitForSliceLease(t, "", "--git-dir", remote, "symbolic-ref", "HEAD", "refs/heads/"+defaultBranch)
}

func writeTerminalCleanupQueue(t *testing.T, root string, entries ...terminalCleanupQueueEntry) {
	t.Helper()
	path, err := terminalCleanupQueuePath(root)
	if err != nil {
		t.Fatal(err)
	}
	content, err := json.Marshal(entries)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
