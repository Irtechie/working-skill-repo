package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const fanInSchemaVersion = 1

// Fan-in debt states, ordered most to least severe.
const (
	fanInStateUntracked   = "untracked"
	fanInStateUncommitted = "uncommitted"
	fanInStateUnmerged    = "unmerged"
	fanInStatePrunable    = "prunable"
	fanInStateSettled     = "settled"
)

// fanInOrphanScanLimit bounds the file count walk for an untracked directory so
// a stray node_modules cannot stall the report.
const fanInOrphanScanLimit = 5000

var fanInSkipDirs = map[string]bool{
	".git":          true,
	"node_modules":  true,
	"target":        true,
	"dist":          true,
	"build":         true,
	"vendor":        true,
	".venv":         true,
	"__pycache__":   true,
	".pytest_cache": true,
	".mypy_cache":   true,
	"bin":           true,
	"obj":           true,
}

type fanInUnit struct {
	State       string `json:"state"`
	Branch      string `json:"branch,omitempty"`
	Worktree    string `json:"worktree,omitempty"`
	Ahead       int    `json:"ahead,omitempty"`
	DirtyFiles  int    `json:"dirty_files,omitempty"`
	OrphanFiles int    `json:"orphan_files,omitempty"`
	Detail      string `json:"detail,omitempty"`
}

type fanInReport struct {
	SchemaVersion int         `json:"schema_version"`
	OK            bool        `json:"ok"`
	Repo          string      `json:"repo"`
	DefaultBranch string      `json:"default_branch,omitempty"`
	Scanned       int         `json:"scanned"`
	Debt          int         `json:"debt"`
	Settled       int         `json:"settled"`
	Units         []fanInUnit `json:"units"`
	Limitations   []string    `json:"limitations,omitempty"`
}

type fanInWorktree struct {
	Path     string
	Branch   string
	Detached bool
	Prunable bool
	Bare     bool
}

func runFanInCommand(root string, opts options, stdout, stderr io.Writer) int {
	report, err := buildFanInReport(root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, report)
	} else {
		writeFanInText(stdout, report)
	}
	if opts.requireClear && report.Debt > 0 {
		fmt.Fprintf(stderr, "fan-in: %d unit(s) of work never came back\n", report.Debt)
		return 2
	}
	return 0
}

// buildFanInReport answers one question: which work was started in this repo and
// never reached the default branch? It never mutates anything.
func buildFanInReport(root string) (fanInReport, error) {
	primary, err := terminalCleanupPrimaryCheckout(root)
	if err != nil {
		return fanInReport{}, err
	}
	report := fanInReport{SchemaVersion: fanInSchemaVersion, Repo: primary}

	worktrees, err := listFanInWorktrees(primary)
	if err != nil {
		return fanInReport{}, err
	}

	report.DefaultBranch = resolveFanInDefaultBranch(primary)
	if report.DefaultBranch == "" {
		report.Limitations = append(report.Limitations,
			"default branch could not be resolved; merge analysis skipped")
	}

	worktreeByBranch := map[string]fanInWorktree{}
	known := map[string]bool{}
	for _, worktree := range worktrees {
		if worktree.Path != "" {
			known[strings.ToLower(filepath.Clean(worktree.Path))] = true
		}
		if worktree.Branch != "" {
			worktreeByBranch[worktree.Branch] = worktree
		}
	}

	for _, worktree := range worktrees {
		if worktree.Prunable {
			report.Units = append(report.Units, fanInUnit{
				State:    fanInStatePrunable,
				Branch:   worktree.Branch,
				Worktree: worktree.Path,
				Detail:   "git worktree prune removes this record",
			})
		}
	}

	// A branch is the unit of work. It carries commits whether or not a worktree
	// still exists for it, which is why branches drive this report and worktrees
	// only decorate it.
	for _, branch := range listFanInBranches(primary) {
		worktree := worktreeByBranch[branch]
		unit := fanInUnit{Branch: branch, Worktree: worktree.Path}

		dirty := 0
		if worktree.Path != "" {
			dirty = countFanInDirtyFiles(worktree.Path)
		}

		ahead := -1
		if report.DefaultBranch != "" {
			ahead = countFanInAhead(primary, report.DefaultBranch, branch)
		}
		unit.Ahead = max(ahead, 0)

		switch {
		case dirty > 0:
			unit.State = fanInStateUncommitted
			unit.DirtyFiles = dirty
			unit.Detail = "changes are not committed anywhere"
		case ahead > 0:
			unit.State = fanInStateUnmerged
			unit.Detail = fmt.Sprintf("%d commit(s) not in %s", ahead, report.DefaultBranch)
		case branch == report.DefaultBranch:
			continue
		default:
			unit.State = fanInStateSettled
		}
		report.Units = append(report.Units, unit)
	}

	orphans := findFanInOrphanDirs(worktrees, known, primary)
	report.Units = append(report.Units, orphans...)
	if len(orphans) > 0 {
		// An untracked directory can be work someone abandoned or a session the
		// harness is still provisioning. This report cannot tell them apart, and
		// saying otherwise invites deleting live work.
		report.Limitations = append(report.Limitations,
			"untracked directories may belong to a live session; confirm the owner before removing any")
	}

	sortFanInUnits(report.Units)
	for _, unit := range report.Units {
		report.Scanned++
		if unit.State == fanInStateSettled {
			report.Settled++
			continue
		}
		report.Debt++
	}
	report.OK = report.Debt == 0
	return report, nil
}

func listFanInWorktrees(root string) ([]fanInWorktree, error) {
	output := gitOutput(root, "worktree", "list", "--porcelain")
	if strings.TrimSpace(output) == "" {
		return nil, fmt.Errorf("git worktree list produced no output for %s", root)
	}
	var worktrees []fanInWorktree
	var current fanInWorktree
	flush := func() {
		if current.Path != "" {
			worktrees = append(worktrees, current)
		}
		current = fanInWorktree{}
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			current.Path = filepath.Clean(strings.TrimPrefix(line, "worktree "))
		case strings.HasPrefix(line, "branch "):
			current.Branch = strings.TrimPrefix(strings.TrimPrefix(line, "branch "), "refs/heads/")
		case line == "detached":
			current.Detached = true
		case line == "bare":
			current.Bare = true
		case strings.HasPrefix(line, "prunable"):
			current.Prunable = true
		}
	}
	flush()
	return worktrees, nil
}

func listFanInBranches(root string) []string {
	output := gitOutput(root, "for-each-ref", "--format=%(refname:short)", "refs/heads")
	var branches []string
	for _, line := range strings.Split(output, "\n") {
		if name := strings.TrimSpace(line); name != "" {
			branches = append(branches, name)
		}
	}
	return branches
}

func resolveFanInDefaultBranch(root string) string {
	if ref := gitOutput(root, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD"); ref != "" {
		if name := strings.TrimPrefix(strings.TrimSpace(ref), "origin/"); name != "" {
			return name
		}
	}
	for _, candidate := range []string{"main", "master"} {
		if gitOutput(root, "rev-parse", "--verify", "--quiet", "refs/heads/"+candidate) != "" {
			return candidate
		}
	}
	return ""
}

func countFanInAhead(root, base, branch string) int {
	if base == branch {
		return 0
	}
	value := gitOutput(root, "rev-list", "--count", base+".."+branch)
	count, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return -1
	}
	return count
}

func countFanInDirtyFiles(worktree string) int {
	if _, err := os.Stat(worktree); err != nil {
		return 0
	}
	output := gitOutput(worktree, "status", "--porcelain")
	count := 0
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}
	return count
}

// findFanInOrphanDirs reports directories sitting among this repo's linked
// worktrees that git has no record of. These are the worst case: git cannot
// recover them, so they are reported for human triage and never touched.
//
// Only parents that are predominantly worktrees are scanned. A lone worktree
// placed in a general-purpose directory (a temp dir, a dev folder, the primary
// checkout's parent) does not make that directory's siblings abandoned work.
func findFanInOrphanDirs(worktrees []fanInWorktree, known map[string]bool, primary string) []fanInUnit {
	primaryParent := strings.ToLower(filepath.Dir(filepath.Clean(primary)))
	parents := map[string]bool{}
	for _, worktree := range worktrees {
		if worktree.Path == "" || worktree.Bare {
			continue
		}
		parent := filepath.Dir(worktree.Path)
		if strings.EqualFold(parent, primaryParent) {
			continue
		}
		parents[parent] = true
	}
	var units []fanInUnit
	seen := map[string]bool{}
	for parent := range parents {
		entries, err := os.ReadDir(parent)
		if err != nil {
			continue
		}
		if !fanInParentIsWorktreeManaged(parent, entries, known) {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			path := filepath.Join(parent, entry.Name())
			key := strings.ToLower(filepath.Clean(path))
			if known[key] || seen[key] || fanInSkipDirs[entry.Name()] {
				continue
			}
			if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
				continue
			}
			seen[key] = true
			units = append(units, fanInUnit{
				State:       fanInStateUntracked,
				Worktree:    path,
				OrphanFiles: countFanInOrphanFiles(path),
				Detail:      "no git record; verify owner before touching",
			})
		}
	}
	return units
}

// fanInParentIsWorktreeManaged reports whether a directory exists to hold
// worktrees, evidenced by holding more than one of them. A directory with a
// single worktree is just where someone put a worktree, so its siblings are
// unrelated files rather than abandoned work.
//
// This deliberately counts worktrees rather than requiring them to outnumber
// everything else: heavy sprawl means orphans can outnumber live worktrees, and
// the report must not go quiet at exactly the moment it matters.
func fanInParentIsWorktreeManaged(parent string, entries []os.DirEntry, known map[string]bool) bool {
	worktrees := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if known[strings.ToLower(filepath.Clean(filepath.Join(parent, entry.Name())))] {
			worktrees++
		}
		if worktrees > 1 {
			return true
		}
	}
	return false
}

func countFanInOrphanFiles(root string) int {
	count := 0
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			if path != root && fanInSkipDirs[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		count++
		if count >= fanInOrphanScanLimit {
			return filepath.SkipAll
		}
		return nil
	})
	return count
}

func sortFanInUnits(units []fanInUnit) {
	severity := map[string]int{
		fanInStateUntracked:   0,
		fanInStateUncommitted: 1,
		fanInStateUnmerged:    2,
		fanInStatePrunable:    3,
		fanInStateSettled:     4,
	}
	sort.SliceStable(units, func(i, j int) bool {
		left, right := units[i], units[j]
		if severity[left.State] != severity[right.State] {
			return severity[left.State] < severity[right.State]
		}
		if left.Ahead != right.Ahead {
			return left.Ahead > right.Ahead
		}
		if left.Branch != right.Branch {
			return left.Branch < right.Branch
		}
		return left.Worktree < right.Worktree
	})
}

func writeFanInText(stdout io.Writer, report fanInReport) {
	fmt.Fprintf(stdout, "fan-in: debt=%d settled=%d scanned=%d default=%s\n",
		report.Debt, report.Settled, report.Scanned, fanInDisplay(report.DefaultBranch))
	for _, unit := range report.Units {
		if unit.State == fanInStateSettled {
			continue
		}
		fmt.Fprintf(stdout, "  %-14s %-34s %s\n",
			unit.State, fanInDisplay(unit.Branch), fanInMeasure(unit))
		if unit.Worktree != "" {
			fmt.Fprintf(stdout, "  %-14s %s\n", "", unit.Worktree)
		}
	}
	for _, limitation := range report.Limitations {
		fmt.Fprintf(stdout, "  note: %s\n", limitation)
	}
}

func fanInMeasure(unit fanInUnit) string {
	switch unit.State {
	case fanInStateUntracked:
		return fmt.Sprintf("%d file(s), no git record", unit.OrphanFiles)
	case fanInStateUncommitted:
		return fmt.Sprintf("%d uncommitted file(s)", unit.DirtyFiles)
	case fanInStateUnmerged:
		return fmt.Sprintf("ahead %d", unit.Ahead)
	default:
		return unit.Detail
	}
}

func fanInDisplay(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}
