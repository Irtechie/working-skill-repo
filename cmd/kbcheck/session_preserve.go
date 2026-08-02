package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Session-end durability is deliberately not delivery. A preserved commit stays
// on the session's own branch, is never pushed, and never claims completion.
// Delivery still requires explicit user consent.

const sessionPreserveSchemaVersion = 1

// sessionPreserveMaxBytes caps any single preserved file. Anything larger is
// reported as excluded rather than silently committed or silently dropped.
const sessionPreserveMaxBytes int64 = 5 << 20

// sessionPreserveBinaryExts lists unambiguous compiled build outputs. Media and
// data files are intentionally absent: in repositories such as audiobooks they
// are real work product, not artifacts.
var sessionPreserveBinaryExts = map[string]bool{
	".exe":   true,
	".dll":   true,
	".so":    true,
	".dylib": true,
	".a":     true,
	".lib":   true,
	".o":     true,
	".obj":   true,
	".pdb":   true,
	".class": true,
	".pyc":   true,
	".pyo":   true,
	".wasm":  true,
}

type sessionPreserveOptions struct {
	Action    string
	SessionID string
	RepoRoot  string
	Worktree  string
	Branch    string
	Now       time.Time
}

type sessionPreserveExclusion struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
	Bytes  int64  `json:"bytes,omitempty"`
}

type sessionPreserveResult struct {
	SchemaVersion int                        `json:"schema_version"`
	Action        string                     `json:"action"`
	Status        string                     `json:"status"`
	SessionID     string                     `json:"session_id,omitempty"`
	Worktree      string                     `json:"worktree,omitempty"`
	Branch        string                     `json:"branch,omitempty"`
	DefaultBranch string                     `json:"default_branch,omitempty"`
	BaseSHA       string                     `json:"base_sha,omitempty"`
	CommitSHA     string                     `json:"commit_sha,omitempty"`
	Preserved     []string                   `json:"preserved"`
	Excluded      []sessionPreserveExclusion `json:"excluded"`
	Pushed        bool                       `json:"pushed"`
	RefusalReason string                     `json:"refusal_reason,omitempty"`
	ObservedAt    string                     `json:"observed_at"`
}

func runSessionPreserveCommand(root string, opts options, stdout, stderr io.Writer) int {
	current, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "resolve current worktree: %v\n", err)
		return 1
	}
	worktree := current
	if strings.TrimSpace(opts.worktreePath) != "" {
		worktree = opts.worktreePath
	}

	result, err := executeSessionPreserve(sessionPreserveOptions{
		Action:    opts.sliceLeaseAction,
		SessionID: opts.sessionID,
		RepoRoot:  root,
		Worktree:  worktree,
		Branch:    opts.branchName,
		Now:       time.Now().UTC(),
	})

	if opts.json {
		writeJSON(stdout, result)
	} else {
		writeSessionPreserveText(stdout, result)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func executeSessionPreserve(opts sessionPreserveOptions) (sessionPreserveResult, error) {
	result := sessionPreserveResult{
		SchemaVersion: sessionPreserveSchemaVersion,
		Action:        opts.Action,
		SessionID:     opts.SessionID,
		Preserved:     []string{},
		Excluded:      []sessionPreserveExclusion{},
		Pushed:        false,
		ObservedAt:    opts.Now.Format(time.RFC3339Nano),
	}

	if opts.Action != "plan" && opts.Action != "apply" {
		return refuseSessionPreserve(result, "invalid-action",
			fmt.Errorf("session-preserve action must be plan or apply"))
	}
	if strings.TrimSpace(opts.SessionID) == "" {
		return refuseSessionPreserve(result, "missing-session-id",
			fmt.Errorf("session-preserve requires --session-id"))
	}

	worktree, err := filepath.Abs(opts.Worktree)
	if err != nil {
		return refuseSessionPreserve(result, "unresolvable-worktree", err)
	}
	result.Worktree = worktree

	if out, err := gitChecked(worktree, "rev-parse", "--is-inside-work-tree"); err != nil || out != "true" {
		return refuseSessionPreserve(result, "not-a-worktree",
			fmt.Errorf("session-preserve requires a git worktree: %s", worktree))
	}

	// An in-progress merge, rebase, cherry-pick, or bisect means HEAD is mid-flight.
	// Committing on top of that would corrupt the operation.
	if reason := inProgressGitOperation(worktree); reason != "" {
		return refuseSessionPreserve(result, reason,
			fmt.Errorf("session-preserve refuses during an in-progress %s", reason))
	}

	branch, err := gitChecked(worktree, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return refuseSessionPreserve(result, "unresolvable-head", err)
	}
	if branch == "HEAD" {
		return refuseSessionPreserve(result, "detached-head",
			fmt.Errorf("session-preserve refuses on a detached HEAD"))
	}
	result.Branch = branch

	if requested := strings.TrimSpace(opts.Branch); requested != "" && requested != branch {
		return refuseSessionPreserve(result, "branch-mismatch",
			fmt.Errorf("session-preserve requested branch %q but worktree is on %q", requested, branch))
	}

	defaultBranch := resolveSessionDefaultBranch(worktree)
	result.DefaultBranch = defaultBranch
	if defaultBranch != "" && branch == defaultBranch {
		return refuseSessionPreserve(result, "default-branch",
			fmt.Errorf("session-preserve refuses on the shared default branch %q; durability applies to session-owned branches only", branch))
	}

	head, err := gitChecked(worktree, "rev-parse", "HEAD")
	if err != nil {
		return refuseSessionPreserve(result, "unresolvable-base", err)
	}
	result.BaseSHA = head

	changed, err := sessionPreserveChangedPaths(worktree)
	if err != nil {
		return refuseSessionPreserve(result, "unreadable-status", err)
	}

	for _, path := range changed {
		if reason, size := sessionPreserveExclusionFor(worktree, path); reason != "" {
			result.Excluded = append(result.Excluded, sessionPreserveExclusion{
				Path:   path,
				Reason: reason,
				Bytes:  size,
			})
			continue
		}
		result.Preserved = append(result.Preserved, path)
	}

	if len(result.Preserved) == 0 {
		result.Status = "nothing-to-preserve"
		return result, nil
	}

	if opts.Action == "plan" {
		result.Status = "would-preserve"
		return result, nil
	}

	for _, path := range result.Preserved {
		if _, err := gitChecked(worktree, "add", "--", path); err != nil {
			return refuseSessionPreserve(result, "stage-failed",
				fmt.Errorf("stage %s: %w", path, err))
		}
	}

	message := sessionPreserveCommitMessage(opts.SessionID, len(result.Preserved), len(result.Excluded))
	if _, err := gitChecked(worktree, "commit", "--no-verify", "-m", message); err != nil {
		return refuseSessionPreserve(result, "commit-failed", err)
	}

	commit, err := gitChecked(worktree, "rev-parse", "HEAD")
	if err != nil {
		return refuseSessionPreserve(result, "unresolvable-commit", err)
	}
	result.CommitSHA = commit
	result.Status = "preserved"
	return result, nil
}

func refuseSessionPreserve(result sessionPreserveResult, reason string, err error) (sessionPreserveResult, error) {
	result.Status = "refused"
	result.RefusalReason = reason
	result.Pushed = false
	return result, err
}

func inProgressGitOperation(worktree string) string {
	gitDir, err := gitChecked(worktree, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return ""
	}
	markers := []struct {
		path   string
		reason string
	}{
		{"MERGE_HEAD", "merge"},
		{"CHERRY_PICK_HEAD", "cherry-pick"},
		{"REVERT_HEAD", "revert"},
		{"BISECT_LOG", "bisect"},
		{"rebase-merge", "rebase"},
		{"rebase-apply", "rebase"},
	}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(gitDir, marker.path)); err == nil {
			return marker.reason
		}
	}
	return ""
}

func resolveSessionDefaultBranch(worktree string) string {
	if ref, err := gitChecked(worktree, "symbolic-ref", "--short", "refs/remotes/origin/HEAD"); err == nil && ref != "" {
		if _, name, ok := strings.Cut(ref, "/"); ok && name != "" {
			return name
		}
	}
	for _, candidate := range []string{"main", "master"} {
		if _, err := gitChecked(worktree, "rev-parse", "--verify", "refs/heads/"+candidate); err == nil {
			return candidate
		}
	}
	return ""
}

// sessionPreserveChangedPaths returns tracked modifications, deletions, and
// untracked files. Ignored files are excluded by git itself.
func sessionPreserveChangedPaths(worktree string) ([]string, error) {
	raw, err := gitRaw(worktree, "status", "--porcelain=v1", "-z", "--untracked-files=all")
	if err != nil {
		return nil, err
	}

	fields := strings.Split(raw, "\x00")
	paths := make([]string, 0, len(fields))
	seen := map[string]bool{}

	for i := 0; i < len(fields); i++ {
		entry := fields[i]
		if len(entry) < 4 {
			continue
		}
		status := entry[:2]
		path := entry[3:]
		// A rename entry is followed by its origin path in the next field.
		if status[0] == 'R' || status[1] == 'R' {
			if i+1 < len(fields) {
				origin := fields[i+1]
				i++
				if origin != "" && !seen[origin] {
					seen[origin] = true
					paths = append(paths, origin)
				}
			}
		}
		if path != "" && !seen[path] {
			seen[path] = true
			paths = append(paths, path)
		}
	}
	return paths, nil
}

func sessionPreserveExclusionFor(worktree, path string) (string, int64) {
	if sessionPreserveBinaryExts[strings.ToLower(filepath.Ext(path))] {
		return "build-artifact", 0
	}
	info, err := os.Lstat(filepath.Join(worktree, filepath.FromSlash(path)))
	if err != nil {
		// A deletion has no file on disk and is safe to preserve.
		return "", 0
	}
	if info.IsDir() {
		return "", 0
	}
	if info.Size() > sessionPreserveMaxBytes {
		return "oversized", info.Size()
	}
	return "", 0
}

func sessionPreserveCommitMessage(sessionID string, preserved, excluded int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "WIP: preserve session work in progress\n\n")
	fmt.Fprintf(&b, "Automatic session-end durability commit. This is not a delivery\n")
	fmt.Fprintf(&b, "claim: the work is unreviewed, unproven, and was never pushed.\n\n")
	fmt.Fprintf(&b, "Preserved files: %d\n", preserved)
	if excluded > 0 {
		fmt.Fprintf(&b, "Excluded files: %d (build artifacts or oversized; still on disk)\n", excluded)
	}
	fmt.Fprintf(&b, "Session: %s\n", sessionID)
	return b.String()
}

func writeSessionPreserveText(w io.Writer, result sessionPreserveResult) {
	switch result.Status {
	case "preserved":
		fmt.Fprintf(w, "session-preserve: preserved %d file(s) on %s as %s\n",
			len(result.Preserved), result.Branch, shortSHA(result.CommitSHA))
	case "would-preserve":
		fmt.Fprintf(w, "session-preserve: would preserve %d file(s) on %s\n",
			len(result.Preserved), result.Branch)
	case "nothing-to-preserve":
		fmt.Fprintf(w, "session-preserve: nothing to preserve on %s\n", result.Branch)
	default:
		fmt.Fprintf(w, "session-preserve: refused (%s)\n", result.RefusalReason)
	}
	for _, excluded := range result.Excluded {
		fmt.Fprintf(w, "  excluded %s (%s)\n", excluded.Path, excluded.Reason)
	}
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

func gitRaw(worktree string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", worktree}, args...)
	output, err := exec.Command("git", cmdArgs...).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func gitChecked(worktree string, args ...string) (string, error) {
	out, err := gitRaw(worktree, args...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}
