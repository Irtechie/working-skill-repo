package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFanInClassifiesUnmergedUncommittedAndSettledWork(t *testing.T) {
	root := initWorktreeRepo(t)
	gitOK(t, root, "branch", "-M", "main")

	// Settled: a branch merged back into main leaves no debt.
	gitOK(t, root, "checkout", "-b", "settled-work")
	writeFile(t, filepath.Join(root, "src", "settled.txt"), "settled\n")
	gitOK(t, root, "add", ".")
	gitOK(t, root, "commit", "-m", "settled work")
	gitOK(t, root, "checkout", "main")
	gitOK(t, root, "merge", "--no-ff", "-m", "merge settled", "settled-work")

	// Unmerged: commits that never reached main.
	gitOK(t, root, "checkout", "-b", "unmerged-work")
	writeFile(t, filepath.Join(root, "src", "unmerged.txt"), "unmerged\n")
	gitOK(t, root, "add", ".")
	gitOK(t, root, "commit", "-m", "unmerged work")
	gitOK(t, root, "checkout", "main")

	report, err := buildFanInReport(root)
	if err != nil {
		t.Fatalf("build fan-in report: %v", err)
	}
	if report.DefaultBranch != "main" {
		t.Fatalf("expected default branch main, got %q", report.DefaultBranch)
	}

	states := fanInStatesByBranch(report)
	if states["unmerged-work"] != fanInStateUnmerged {
		t.Fatalf("expected unmerged-work to be unmerged, got %q", states["unmerged-work"])
	}
	if states["settled-work"] != fanInStateSettled {
		t.Fatalf("expected settled-work to be settled, got %q", states["settled-work"])
	}
	if report.Debt != 1 {
		t.Fatalf("expected debt=1, got %d (units %+v)", report.Debt, report.Units)
	}
	if report.OK {
		t.Fatal("expected OK=false while debt remains")
	}

	unit := fanInUnitForBranch(t, report, "unmerged-work")
	if unit.Ahead != 1 {
		t.Fatalf("expected unmerged-work ahead=1, got %d", unit.Ahead)
	}
}

func TestFanInReportsUncommittedWorkAheadOfUnmerged(t *testing.T) {
	root := initWorktreeRepo(t)
	gitOK(t, root, "branch", "-M", "main")
	writeFile(t, filepath.Join(root, "src", "dirty.txt"), "uncommitted\n")

	report, err := buildFanInReport(root)
	if err != nil {
		t.Fatalf("build fan-in report: %v", err)
	}
	unit := fanInUnitForBranch(t, report, "main")
	if unit.State != fanInStateUncommitted {
		t.Fatalf("expected main to be uncommitted, got %q", unit.State)
	}
	if unit.DirtyFiles != 1 {
		t.Fatalf("expected 1 dirty file, got %d", unit.DirtyFiles)
	}
	// Uncommitted work is the most recoverable-by-nobody state after an orphan
	// directory, so it must outrank unmerged commits in the report ordering.
	if report.Units[0].State != fanInStateUncommitted {
		t.Fatalf("expected uncommitted first, got %q", report.Units[0].State)
	}
}

func TestFanInReportsOrphanDirectoriesBesideLinkedWorktreesOnly(t *testing.T) {
	root := initWorktreeRepo(t)
	gitOK(t, root, "branch", "-M", "main")

	// Two linked worktrees mark this parent as a worktree-managed directory.
	linkedParent := t.TempDir()
	linked := filepath.Join(linkedParent, "linked")
	gitOK(t, root, "worktree", "add", "-b", "linked-work", linked)
	second := filepath.Join(linkedParent, "second")
	gitOK(t, root, "worktree", "add", "-b", "second-work", second)

	// An untracked sibling of those worktrees is work with no git record.
	orphan := filepath.Join(linkedParent, "orphan")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatalf("create orphan dir: %v", err)
	}
	writeFile(t, filepath.Join(orphan, "stranded.txt"), "stranded\n")

	// Build caches and tool directories are not abandoned work.
	noise := filepath.Join(linkedParent, "__pycache__")
	if err := os.MkdirAll(noise, 0o755); err != nil {
		t.Fatalf("create cache dir: %v", err)
	}
	writeFile(t, filepath.Join(noise, "cached.pyc"), "cached\n")

	// A sibling of the PRIMARY checkout is an unrelated project, not debt.
	unrelated := filepath.Join(filepath.Dir(root), "unrelated-project")
	if err := os.MkdirAll(unrelated, 0o755); err != nil {
		t.Fatalf("create unrelated dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(unrelated) })

	report, err := buildFanInReport(root)
	if err != nil {
		t.Fatalf("build fan-in report: %v", err)
	}

	var found *fanInUnit
	for index, unit := range report.Units {
		if unit.State != fanInStateUntracked {
			continue
		}
		if filepath.Clean(unit.Worktree) == filepath.Clean(unrelated) {
			t.Fatal("primary checkout sibling must not be reported as orphaned work")
		}
		if filepath.Clean(unit.Worktree) == filepath.Clean(noise) {
			t.Fatal("build cache directory must not be reported as orphaned work")
		}
		if filepath.Clean(unit.Worktree) == filepath.Clean(orphan) {
			found = &report.Units[index]
		}
	}
	if found == nil {
		t.Fatalf("expected orphan directory to be reported, got %+v", report.Units)
	}
	if found.OrphanFiles != 1 {
		t.Fatalf("expected 1 stranded file, got %d", found.OrphanFiles)
	}
	if report.Units[0].State != fanInStateUntracked {
		t.Fatalf("expected untracked work ranked first, got %q", report.Units[0].State)
	}
	// A live session the harness is still provisioning looks exactly like an
	// abandoned directory. The report must say so rather than imply a verdict.
	if !strings.Contains(strings.Join(report.Limitations, " "), "live session") {
		t.Fatalf("untracked findings must carry a live-session caveat, got %v", report.Limitations)
	}
}

func TestFanInIsReadOnlyAndGatesOnlyWhenRequested(t *testing.T) {
	root := initWorktreeRepo(t)
	gitOK(t, root, "branch", "-M", "main")
	gitOK(t, root, "checkout", "-b", "unmerged-work")
	writeFile(t, filepath.Join(root, "src", "unmerged.txt"), "unmerged\n")
	gitOK(t, root, "add", ".")
	gitOK(t, root, "commit", "-m", "unmerged work")
	gitOK(t, root, "checkout", "main")

	before := gitOutput(root, "for-each-ref", "--format=%(refname)", "refs/heads")

	if code := runFanInCommand(root, options{}, io.Discard, io.Discard); code != 0 {
		t.Fatalf("expected report-only exit 0, got %d", code)
	}
	if code := runFanInCommand(root, options{requireClear: true}, io.Discard, io.Discard); code != 2 {
		t.Fatalf("expected --require-clear exit 2 with debt, got %d", code)
	}

	after := gitOutput(root, "for-each-ref", "--format=%(refname)", "refs/heads")
	if before != after {
		t.Fatalf("fan-in mutated refs:\nbefore %q\nafter  %q", before, after)
	}
	if status := gitOutput(root, "status", "--porcelain"); status != "" {
		t.Fatalf("fan-in dirtied the worktree: %q", status)
	}
}

func TestFanInIgnoresOrphansBesideALoneWorktree(t *testing.T) {
	root := initWorktreeRepo(t)
	gitOK(t, root, "branch", "-M", "main")

	// One worktree parked in a general-purpose directory does not make that
	// directory's unrelated siblings abandoned work.
	shared := t.TempDir()
	gitOK(t, root, "worktree", "add", "-b", "lone-work", filepath.Join(shared, "lone"))
	sibling := filepath.Join(shared, "unrelated-temp-data")
	if err := os.MkdirAll(sibling, 0o755); err != nil {
		t.Fatalf("create sibling dir: %v", err)
	}
	writeFile(t, filepath.Join(sibling, "scratch.txt"), "scratch\n")

	report, err := buildFanInReport(root)
	if err != nil {
		t.Fatalf("build fan-in report: %v", err)
	}
	for _, unit := range report.Units {
		if unit.State == fanInStateUntracked {
			t.Fatalf("lone-worktree parent must not produce orphan findings, got %+v", unit)
		}
	}
}

func fanInStatesByBranch(report fanInReport) map[string]string {
	states := map[string]string{}
	for _, unit := range report.Units {
		if unit.Branch != "" {
			states[unit.Branch] = unit.State
		}
	}
	return states
}

func fanInUnitForBranch(t *testing.T, report fanInReport, branch string) fanInUnit {
	t.Helper()
	for _, unit := range report.Units {
		if unit.Branch == branch {
			return unit
		}
	}
	t.Fatalf("branch %q not present in report %+v", branch, report.Units)
	return fanInUnit{}
}
