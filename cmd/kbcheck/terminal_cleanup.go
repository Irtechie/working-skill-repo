package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
	"github.com/Irtechie/working-skill-repo/internal/reconcile"
)

const terminalCleanupSchemaVersion = 2
const terminalCleanupSafetyContractVersion = reconcile.WorktreeSafetyContractVersion

var terminalCleanupRetryDelays = []time.Duration{0, 100 * time.Millisecond, 300 * time.Millisecond}

type terminalCleanupOptions struct {
	Action             string
	WorkID             string
	SessionID          string
	Worktree           string
	Branch             string
	CommitSHA          string
	DeliveryMode       string
	Remote             string
	RepoRoot           string
	CurrentWorktree    string
	CurrentSession     string
	LockTimeout        time.Duration
	Now                time.Time
	BeforeReceiptSweep func(index int, receipt terminalCleanupReceipt)
}

type terminalCleanupReceipt struct {
	SchemaVersion    int                      `json:"schema_version"`
	WorkID           string                   `json:"work_id"`
	SessionID        string                   `json:"session_id"`
	RepoRoot         string                   `json:"repo_root"`
	Worktree         string                   `json:"worktree"`
	WorktreeRealPath string                   `json:"worktree_real_path"`
	WorktreeGitDir   string                   `json:"worktree_git_dir"`
	Branch           string                   `json:"branch"`
	CommitSHA        string                   `json:"commit_sha"`
	WorktreeToken    string                   `json:"worktree_token"`
	DeliveryMode     string                   `json:"delivery_mode"`
	Remote           string                   `json:"remote,omitempty"`
	EvidenceRecorded bool                     `json:"evidence_recorded"`
	RemoteDefaults   []terminalRemoteEvidence `json:"remote_defaults,omitempty"`
	TopicRef         string                   `json:"topic_ref,omitempty"`
	TopicSHA         string                   `json:"topic_sha,omitempty"`
	Status           string                   `json:"status"`
	RegisteredAt     string                   `json:"registered_at"`
	UpdatedAt        string                   `json:"updated_at"`
	RemovedAt        string                   `json:"removed_at,omitempty"`
	ReleasedAt       string                   `json:"released_at,omitempty"`
	Limitation       string                   `json:"limitation,omitempty"`
	SafetyPolicy     string                   `json:"safety_policy_version"`
}

type terminalCleanupResult struct {
	OK          bool                     `json:"ok"`
	Action      string                   `json:"action"`
	Issue       string                   `json:"issue,omitempty"`
	Receipt     *terminalCleanupReceipt  `json:"receipt,omitempty"`
	Scanned     int                      `json:"scanned,omitempty"`
	Cleaned     int                      `json:"cleaned,omitempty"`
	SweepStatus string                   `json:"sweep_status,omitempty"`
	Outcomes    []terminalCleanupOutcome `json:"outcomes,omitempty"`
}

type terminalDeliveryEvidence struct {
	ContainedOnDefault bool
	DefaultBranch      string
	RemoteDefaults     []terminalRemoteEvidence
	TopicRef           string
	TopicSHA           string
}

type terminalRemoteEvidence struct {
	Remote        string `json:"remote"`
	DefaultBranch string `json:"default_branch"`
	DefaultSHA    string `json:"default_sha"`
}

type terminalCleanupOutcome struct {
	WorkID    string `json:"work_id"`
	SessionID string `json:"session_id"`
	Worktree  string `json:"worktree"`
	State     string `json:"state"`
	Issue     string `json:"issue,omitempty"`
}

type terminalCleanupQueueEntry struct {
	WorkID      string `json:"work_id"`
	Status      string `json:"status"`
	SessionID   string `json:"session_id"`
	Branch      string `json:"branch"`
	Worktree    string `json:"worktree"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	CompletedAt string `json:"completed_at,omitempty"`
}

type terminalGitWorktree struct {
	Path   string
	Branch string
	Head   string
	Locked bool
}

func terminalCleanupSafetyPredicates() []string {
	return []string{
		"clean-ignored",
		"clean-tracked",
		"clean-untracked",
		"different-executor",
		"durable-endpoint",
		"empty-residual-only",
		"exact-worktree-generation",
		"git-admin-round-trip",
		"non-force-only",
		"not-current",
		"not-default",
		"not-locked",
		"not-moved",
		"not-post-cutoff",
		"not-primary",
		"remote-monotonic",
		"terminal-or-suspended-claim",
	}
}

func runTerminalCleanupCommand(root string, opts options, stdout, stderr io.Writer) int {
	current, _ := os.Getwd()
	result, err := executeTerminalCleanup(terminalCleanupOptions{
		Action:          opts.sliceLeaseAction,
		WorkID:          opts.workID,
		SessionID:       opts.sessionID,
		Worktree:        opts.worktreePath,
		Branch:          opts.branchName,
		CommitSHA:       opts.commitSHA,
		DeliveryMode:    opts.deliveryMode,
		Remote:          opts.remote,
		RepoRoot:        root,
		CurrentWorktree: current,
		CurrentSession:  opts.sessionID,
		Now:             time.Now().UTC(),
	})
	if err != nil {
		if opts.sliceLeaseAction == "sweep" {
			if opts.json {
				writeJSON(stdout, result)
			} else {
				writeTerminalCleanupSweepText(stdout, result)
			}
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else if result.Action == "sweep" {
		writeTerminalCleanupSweepText(stdout, result)
	} else if result.OK && result.Receipt != nil {
		fmt.Fprintf(stdout, "terminal-cleanup: %s session=%s status=%s path=%s\n",
			result.Action, result.Receipt.SessionID, result.Receipt.Status, result.Receipt.Worktree)
	} else if result.OK {
		fmt.Fprintf(stdout, "terminal-cleanup: %s ok scanned=%d cleaned=%d\n", result.Action, result.Scanned, result.Cleaned)
	} else {
		fmt.Fprintf(stdout, "terminal-cleanup: %s blocked: %s\n", result.Action, result.Issue)
	}
	if !result.OK {
		return 2
	}
	return 0
}

func executeTerminalCleanup(opts terminalCleanupOptions) (terminalCleanupResult, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action != "register" && action != "sweep" {
		return terminalCleanupResult{}, fmt.Errorf("terminal-cleanup action must be register or sweep")
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	root, err := filepath.Abs(filepath.Clean(opts.RepoRoot))
	if err != nil {
		return terminalCleanupResult{}, err
	}
	opts.RepoRoot = root
	if opts.CurrentWorktree != "" {
		opts.CurrentWorktree, err = filepath.Abs(filepath.Clean(opts.CurrentWorktree))
		if err != nil {
			return terminalCleanupResult{}, err
		}
		if currentRoot := gitOutput(opts.CurrentWorktree, "rev-parse", "--show-toplevel"); currentRoot != "" {
			opts.CurrentWorktree, err = filepath.Abs(filepath.Clean(currentRoot))
			if err != nil {
				return terminalCleanupResult{}, err
			}
		}
	}
	if action == "register" {
		return registerTerminalCleanup(opts)
	}
	opts.RepoRoot, err = terminalCleanupPrimaryCheckout(opts.RepoRoot)
	if err != nil {
		return blockedTerminalCleanup("sweep", err.Error(), nil), nil
	}
	if strings.TrimSpace(opts.CurrentSession) == "" {
		return blockedTerminalCleanup("sweep", "current session-id is required", nil), nil
	}
	return sweepTerminalCleanup(opts)
}

func writeTerminalCleanupSweepText(stdout io.Writer, result terminalCleanupResult) {
	if result.OK {
		fmt.Fprintf(stdout, "terminal-cleanup: sweep ok scanned=%d cleaned=%d\n", result.Scanned, result.Cleaned)
		return
	}
	session := ""
	path := ""
	if result.Receipt != nil {
		session = result.Receipt.SessionID
		path = result.Receipt.Worktree
	}
	fmt.Fprintf(stdout,
		"terminal-cleanup: sweep blocked scanned=%d cleaned=%d session=%s path=%s issue=%s\n",
		result.Scanned, result.Cleaned, session, path, result.Issue,
	)
}

func registerTerminalCleanup(opts terminalCleanupOptions) (terminalCleanupResult, error) {
	if strings.TrimSpace(opts.WorkID) == "" || strings.TrimSpace(opts.SessionID) == "" {
		return blockedTerminalCleanup("register", "work-id and session-id are required", nil), nil
	}
	mode := strings.ToLower(strings.TrimSpace(opts.DeliveryMode))
	if mode != "local" && mode != "pr" && mode != "direct" {
		return blockedTerminalCleanup("register", "delivery-mode must be local, pr, or direct", nil), nil
	}
	worktree, err := filepath.Abs(filepath.Clean(opts.Worktree))
	if err != nil {
		return terminalCleanupResult{}, err
	}
	primary, err := terminalCleanupPrimaryCheckout(opts.RepoRoot)
	if err != nil {
		return blockedTerminalCleanup("register", err.Error(), nil), nil
	}
	receipt := terminalCleanupReceipt{
		SchemaVersion: terminalCleanupSchemaVersion,
		WorkID:        strings.TrimSpace(opts.WorkID),
		SessionID:     strings.TrimSpace(opts.SessionID),
		RepoRoot:      primary,
		Worktree:      worktree,
		Branch:        strings.TrimPrefix(strings.TrimSpace(opts.Branch), "refs/heads/"),
		CommitSHA:     strings.TrimSpace(opts.CommitSHA),
		DeliveryMode:  mode,
		Remote:        strings.TrimSpace(opts.Remote),
		Status:        "registered",
		RegisteredAt:  opts.Now.Format(time.RFC3339Nano),
		UpdatedAt:     opts.Now.Format(time.RFC3339Nano),
		Limitation:    "host UI session records remain host-owned; this receipt retires the Git worktree and exact merged local ref",
		SafetyPolicy:  terminalCleanupSafetyContractVersion,
	}
	if receipt.Branch == "" || receipt.CommitSHA == "" {
		return blockedTerminalCleanup("register", "worktree, branch, and commit-sha are required", &receipt), nil
	}
	receipt.WorktreeRealPath, err = filepath.EvalSymlinks(receipt.Worktree)
	if err != nil {
		return blockedTerminalCleanup("register", "worktree real path cannot be resolved: "+err.Error(), &receipt), nil
	}
	receipt.WorktreeRealPath, err = filepath.Abs(filepath.Clean(receipt.WorktreeRealPath))
	if err != nil {
		return terminalCleanupResult{}, err
	}
	queue, err := loadTerminalCleanupQueue(opts.RepoRoot)
	if err != nil {
		return blockedTerminalCleanup("register", err.Error(), &receipt), nil
	}
	claim, issue := matchingTerminalCleanupClaim(queue, receipt)
	if issue != "" {
		return blockedTerminalCleanup("register", issue, &receipt), nil
	}
	if claim.Status == "blocked" || claim.Status == "superseded" {
		return blockedTerminalCleanup("register", "only active or durably done work can register terminal cleanup", &receipt), nil
	}
	if issue := validateTerminalCleanupTarget(opts.RepoRoot, receipt, ""); issue != "" {
		return blockedTerminalCleanup("register", issue, &receipt), nil
	}
	delivery, issue := validateTerminalDelivery(opts.RepoRoot, receipt, nil)
	if issue != "" {
		return blockedTerminalCleanup("register", issue, &receipt), nil
	}
	receipt.EvidenceRecorded = true
	receipt.RemoteDefaults = delivery.RemoteDefaults
	receipt.TopicRef = delivery.TopicRef
	receipt.TopicSHA = delivery.TopicSHA
	token, markerPath, gitDir, err := createTerminalWorktreeMarker(receipt.Worktree)
	if err != nil {
		return blockedTerminalCleanup("register", err.Error(), &receipt), nil
	}
	receipt.WorktreeToken = token
	receipt.WorktreeGitDir = gitDir
	if err := saveTerminalCleanupReceipt(opts.RepoRoot, receipt); err != nil {
		_ = os.Remove(markerPath)
		return terminalCleanupResult{}, err
	}
	return terminalCleanupResult{OK: true, Action: "register", Receipt: &receipt}, nil
}

func sweepTerminalCleanup(opts terminalCleanupOptions) (terminalCleanupResult, error) {
	receipts, err := loadTerminalCleanupReceipts(opts.RepoRoot)
	if err != nil {
		return terminalCleanupResult{}, err
	}
	result := terminalCleanupResult{OK: true, Action: "sweep", Scanned: len(receipts)}
	var firstErr error
	for index, receipt := range receipts {
		if receipt.Status == "released" {
			result.Outcomes = append(result.Outcomes, terminalCleanupOutcome{
				WorkID: receipt.WorkID, SessionID: receipt.SessionID, Worktree: receipt.Worktree, State: "removed",
			})
			continue
		}
		if opts.BeforeReceiptSweep != nil {
			opts.BeforeReceiptSweep(index, receipt)
		}
		item, err := sweepOneTerminalCleanup(opts, receipt)
		if err != nil {
			result.OK = false
			issue := err.Error()
			result.Outcomes = append(result.Outcomes, terminalCleanupOutcome{
				WorkID: receipt.WorkID, SessionID: receipt.SessionID, Worktree: receipt.Worktree,
				State: "failed", Issue: issue,
			})
			if result.Issue == "" {
				result.Issue = issue
				result.Receipt = &receipt
			}
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if item.OK {
			if item.Receipt != nil && item.Receipt.Status == "released" {
				result.Cleaned++
			}
			result.Outcomes = append(result.Outcomes, terminalCleanupOutcome{
				WorkID: receipt.WorkID, SessionID: receipt.SessionID, Worktree: receipt.Worktree, State: "removed",
			})
			if result.Receipt == nil {
				result.Receipt = item.Receipt
			}
			continue
		}
		result.OK = false
		result.Outcomes = append(result.Outcomes, terminalCleanupOutcome{
			WorkID: receipt.WorkID, SessionID: receipt.SessionID, Worktree: receipt.Worktree,
			State: "blocked", Issue: item.Issue,
		})
		if result.Issue == "" {
			result.Issue = item.Issue
			result.Receipt = item.Receipt
		}
	}
	switch {
	case result.OK:
		result.SweepStatus = "complete"
	case result.Cleaned > 0:
		result.SweepStatus = "partial"
	default:
		result.SweepStatus = "aborted"
	}
	return result, firstErr
}

func sweepOneTerminalCleanup(opts terminalCleanupOptions, receipt terminalCleanupReceipt) (terminalCleanupResult, error) {
	currentReceipt, err := loadTerminalCleanupReceipt(opts.RepoRoot, receipt)
	if err != nil {
		return blockedTerminalCleanup("sweep", err.Error(), &receipt), nil
	}
	if currentReceipt.Status == "released" {
		return terminalCleanupResult{OK: true, Action: "sweep", Receipt: &currentReceipt}, nil
	}
	queuePath, err := terminalCleanupQueuePath(opts.RepoRoot)
	if err != nil {
		return terminalCleanupResult{}, err
	}
	timeout := opts.LockTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	lock, err := modelrouting.AcquireSharedProjectLock(filepath.Dir(queuePath), "work-queue.lock", timeout)
	if err != nil {
		return terminalCleanupResult{}, fmt.Errorf("acquire shared work queue lock: %w", err)
	}
	defer lock.Close()
	return sweepOneTerminalCleanupLocked(opts, currentReceipt)
}

func sweepOneTerminalCleanupLocked(
	opts terminalCleanupOptions,
	receipt terminalCleanupReceipt,
) (terminalCleanupResult, error) {
	observedUpdatedAt := receipt.UpdatedAt
	currentReceipt, err := loadTerminalCleanupReceipt(opts.RepoRoot, receipt)
	if err != nil {
		return blockedTerminalCleanup("sweep", err.Error(), &receipt), nil
	}
	receipt = currentReceipt
	if receipt.Status == "released" {
		return terminalCleanupResult{OK: true, Action: "sweep", Receipt: &receipt}, nil
	}
	if receipt.SafetyPolicy != terminalCleanupSafetyContractVersion {
		return blockedTerminalCleanup("sweep", "terminal cleanup safety policy version is missing or incompatible", &receipt), nil
	}

	if receipt.UpdatedAt != observedUpdatedAt {
		return blockedTerminalCleanup("sweep", "cleanup receipt changed while sweep evidence was collected", &receipt), nil
	}

	queue, err := loadTerminalCleanupQueue(opts.RepoRoot)
	if err != nil {
		return blockedTerminalCleanup("sweep", err.Error(), &receipt), nil
	}
	claim, issue := matchingTerminalCleanupClaim(queue, receipt)
	if issue != "" {
		return blockedTerminalCleanup("sweep", issue, &receipt), nil
	}
	for _, entry := range queue {
		if entry.Status != "queued" && entry.Status != "in_progress" {
			continue
		}
		if samePath(entry.Worktree, receipt.Worktree) || entry.Branch == receipt.Branch {
			return blockedTerminalCleanup("sweep", "active queue claim still owns the worktree or branch", &receipt), nil
		}
	}
	if claim.Status != "done" {
		return blockedTerminalCleanup("sweep", "cleanup requires a durably done queue claim", &receipt), nil
	}
	if receipt.SessionID == opts.CurrentSession {
		return blockedTerminalCleanup("sweep", "current executing session cannot retire its own worktree", &receipt), nil
	}
	if receipt.Status == "registered" {
		if samePath(opts.CurrentWorktree, receipt.Worktree) {
			return blockedTerminalCleanup("sweep", "current executing worktree cannot remove itself", &receipt), nil
		}
		worktrees, err := listTerminalGitWorktrees(opts.RepoRoot)
		if err != nil {
			return blockedTerminalCleanup("sweep", err.Error(), &receipt), nil
		}
		target, registered := worktrees[terminalPathKey(receipt.Worktree)]
		if registered && !pathExists(receipt.Worktree) {
			return blockedTerminalCleanup("sweep", "registered worktree path is missing; scoped prune or manual recovery is required", &receipt), nil
		}
		if !registered {
			if pathExists(receipt.Worktree) {
				if issue := reconcileTerminalPartialRemoval(opts.RepoRoot, receipt, worktrees); issue != "" {
					return blockedTerminalCleanup("sweep", issue, &receipt), nil
				}
				receipt.Status = "worktree-removed"
				receipt.RemovedAt = opts.Now.Format(time.RFC3339Nano)
				receipt.UpdatedAt = receipt.RemovedAt
				if err := saveTerminalCleanupReceipt(opts.RepoRoot, receipt); err != nil {
					return terminalCleanupResult{}, err
				}
			} else {
				for _, worktree := range worktrees {
					if token, _ := readTerminalWorktreeMarker(worktree.Path); token == receipt.WorktreeToken {
						return blockedTerminalCleanup("sweep", "registered worktree moved to a different path after cleanup registration", &receipt), nil
					}
				}
				return blockedTerminalCleanup("sweep", "registered worktree and admin identity are missing; scoped prune or manual recovery is required", &receipt), nil
			}
		} else {
			if target.Branch != receipt.Branch || target.Head != receipt.CommitSHA {
				return blockedTerminalCleanup("sweep", "registered worktree identity moved after cleanup registration", &receipt), nil
			}
			if issue := validateTerminalCleanupTarget(opts.RepoRoot, receipt, opts.CurrentWorktree); issue != "" {
				return blockedTerminalCleanup("sweep", issue, &receipt), nil
			}
			if _, issue := validateTerminalDelivery(opts.RepoRoot, receipt, nil); issue != "" {
				return blockedTerminalCleanup("sweep", issue, &receipt), nil
			}
			if issue := removeTerminalWorktreeWithRetry(opts.RepoRoot, receipt.Worktree); issue != "" {
				return blockedTerminalCleanup("sweep", issue, &receipt), nil
			}
			receipt.Status = "worktree-removed"
			receipt.RemovedAt = opts.Now.Format(time.RFC3339Nano)
			receipt.UpdatedAt = receipt.RemovedAt
			if err := saveTerminalCleanupReceipt(opts.RepoRoot, receipt); err != nil {
				return terminalCleanupResult{}, err
			}
		}
	}

	if receipt.DeliveryMode == "local" || receipt.DeliveryMode == "pr" {
		receipt.Status = "released"
		receipt.ReleasedAt = opts.Now.Format(time.RFC3339Nano)
		receipt.UpdatedAt = receipt.ReleasedAt
		if err := saveTerminalCleanupReceipt(opts.RepoRoot, receipt); err != nil {
			return terminalCleanupResult{}, err
		}
		return terminalCleanupResult{OK: true, Action: "sweep", Receipt: &receipt}, nil
	}
	delivery, issue := validateTerminalDelivery(opts.RepoRoot, receipt, nil)
	if issue != "" {
		return blockedTerminalCleanup("sweep", issue, &receipt), nil
	}
	if !delivery.ContainedOnDefault {
		return terminalCleanupResult{OK: true, Action: "sweep", Receipt: &receipt}, nil
	}
	if isResolvedDefaultBranch(opts.RepoRoot, receipt.Branch) {
		return blockedTerminalCleanup("sweep", "default branch ref deletion is forbidden", &receipt), nil
	}
	worktrees, err := listTerminalGitWorktrees(opts.RepoRoot)
	if err != nil {
		return blockedTerminalCleanup("sweep", err.Error(), &receipt), nil
	}
	for _, worktree := range worktrees {
		if worktree.Branch == receipt.Branch {
			return blockedTerminalCleanup("sweep", "merged local feature ref is checked out in a Git worktree", &receipt), nil
		}
	}
	localRef := "refs/heads/" + receipt.Branch
	current := gitOutput(opts.RepoRoot, "rev-parse", localRef+"^{commit}")
	if current != "" && current != receipt.CommitSHA {
		return blockedTerminalCleanup("sweep", "local feature ref moved after delivery; exact deletion refused", &receipt), nil
	}
	if current == receipt.CommitSHA {
		if code, out := runGitCommand(receipt.RepoRoot, "update-ref", "-d", localRef, receipt.CommitSHA); code != 0 {
			return blockedTerminalCleanup("sweep", "compare-and-swap local feature ref deletion failed: "+strings.TrimSpace(out), &receipt), nil
		}
		if gitOutput(opts.RepoRoot, "show-ref", "--verify", localRef) != "" {
			return blockedTerminalCleanup("sweep", "local feature ref still exists after checked-out-aware deletion", &receipt), nil
		}
	}
	receipt.Status = "released"
	receipt.ReleasedAt = opts.Now.Format(time.RFC3339Nano)
	receipt.UpdatedAt = receipt.ReleasedAt
	if err := saveTerminalCleanupReceipt(opts.RepoRoot, receipt); err != nil {
		return terminalCleanupResult{}, err
	}
	return terminalCleanupResult{OK: true, Action: "sweep", Receipt: &receipt}, nil
}

func reconcileTerminalPartialRemoval(
	root string,
	receipt terminalCleanupReceipt,
	worktrees map[string]terminalGitWorktree,
) string {
	if receipt.WorktreeToken == "" || receipt.WorktreeGitDir == "" ||
		receipt.WorktreeRealPath == "" || !receipt.EvidenceRecorded {
		return "partial cleanup receipt is missing registered worktree identity evidence"
	}
	for _, worktree := range worktrees {
		if token, _ := readTerminalWorktreeMarker(worktree.Path); token == receipt.WorktreeToken {
			return "registered worktree moved to a different path after cleanup registration"
		}
	}
	common, err := terminalCleanupCommonDir(root)
	if err != nil {
		return err.Error()
	}
	adminParent := filepath.Dir(receipt.WorktreeGitDir)
	expectedAdminParent := filepath.Join(common, "worktrees")
	if !samePath(adminParent, expectedAdminParent) {
		return "partial cleanup receipt admin identity is outside the registered Git common directory"
	}
	if pathExists(adminParent) {
		adminParent, err = filepath.EvalSymlinks(adminParent)
		if err != nil {
			return "partial cleanup receipt admin parent is unreadable"
		}
		expectedAdminParent, err = filepath.EvalSymlinks(expectedAdminParent)
		if err != nil {
			return "registered Git worktree admin parent is unreadable"
		}
		if !samePath(adminParent, expectedAdminParent) {
			return "partial cleanup receipt admin identity resolves outside the registered Git common directory"
		}
	}
	if pathExists(receipt.WorktreeGitDir) {
		return "target path exists but is not a registered Git worktree; its Git admin identity was not removed"
	}
	localRef := "refs/heads/" + receipt.Branch
	if current := gitOutput(root, "rev-parse", localRef+"^{commit}"); current != receipt.CommitSHA {
		return "local branch ref does not match the cleanup receipt; partial reconciliation refused"
	}
	if _, issue := validateTerminalDelivery(root, receipt, nil); issue != "" {
		return issue
	}
	return removeTerminalEmptyResidualWithRetry(receipt)
}

func removeTerminalEmptyResidualWithRetry(receipt terminalCleanupReceipt) string {
	issue := ""
	for _, delay := range terminalCleanupRetryDelays {
		if delay > 0 {
			time.Sleep(delay)
		}
		currentIssue, retryable := inspectTerminalEmptyResidual(receipt)
		if currentIssue != "" {
			issue = currentIssue
			if !retryable {
				return issue
			}
			continue
		}
		if err := os.Remove(receipt.Worktree); err == nil {
			return ""
		} else {
			issue = "remove empty partial cleanup residual directory: " + err.Error()
		}
	}
	return issue
}

func inspectTerminalEmptyResidual(receipt terminalCleanupReceipt) (string, bool) {
	info, err := os.Lstat(receipt.Worktree)
	if err != nil {
		return "partial cleanup residual directory is unreadable: " + err.Error(), true
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "partial cleanup residual path is not an exact directory", false
	}
	realPath, err := filepath.EvalSymlinks(receipt.Worktree)
	if err != nil {
		return "partial cleanup residual real path is unreadable: " + err.Error(), true
	}
	if !samePath(realPath, receipt.WorktreeRealPath) {
		return "partial cleanup residual real path does not match the registered target", false
	}
	entries, err := os.ReadDir(receipt.Worktree)
	if err != nil {
		return "partial cleanup residual directory is unreadable: " + err.Error(), true
	}
	if len(entries) != 0 {
		return "partial cleanup residual directory is not empty; local data is preserved", false
	}
	return "", false
}

func removeTerminalWorktreeWithRetry(root, worktree string) string {
	var output string
	for _, delay := range terminalCleanupRetryDelays {
		if delay > 0 {
			time.Sleep(delay)
		}
		code, current := runGitCommand(root, "worktree", "remove", worktree)
		if code == 0 {
			return ""
		}
		output = strings.TrimSpace(current)
	}
	return "non-force worktree removal failed after bounded retries: " + output
}

func validateTerminalCleanupTarget(root string, receipt terminalCleanupReceipt, currentWorktree string) string {
	if filepath.Dir(receipt.Worktree) == receipt.Worktree {
		return "filesystem root cannot be a cleanup target"
	}
	primary, err := terminalCleanupPrimaryCheckout(root)
	if err != nil {
		return err.Error()
	}
	if samePath(receipt.Worktree, primary) {
		return "primary checkout cannot be a cleanup target"
	}
	if currentWorktree != "" && samePath(receipt.Worktree, currentWorktree) {
		return "current executing worktree cannot remove itself"
	}
	if isResolvedDefaultBranch(root, receipt.Branch) {
		return "default branch worktree cannot be a cleanup target"
	}
	targetCommon, err := terminalCleanupCommonDir(receipt.Worktree)
	if err != nil {
		return "target is not a readable Git worktree"
	}
	rootCommon, err := terminalCleanupCommonDir(root)
	if err != nil {
		return err.Error()
	}
	if !samePath(targetCommon, rootCommon) {
		return "target worktree belongs to a different Git common directory"
	}
	worktrees, err := listTerminalGitWorktrees(root)
	if err != nil {
		return err.Error()
	}
	target, ok := worktrees[terminalPathKey(receipt.Worktree)]
	if !ok {
		return "target is not a registered Git worktree"
	}
	if target.Locked {
		return "target worktree is locked; cleanup will not unlock it"
	}
	if target.Branch != receipt.Branch || target.Head != receipt.CommitSHA {
		return "target worktree branch or HEAD does not match delivered identity"
	}
	if receipt.WorktreeGitDir != "" {
		if issue := validateTerminalWorktreeRoundTrip(root, receipt); issue != "" {
			return issue
		}
	}
	if receipt.WorktreeToken != "" {
		token, err := readTerminalWorktreeMarker(receipt.Worktree)
		if err != nil || token != receipt.WorktreeToken {
			return "target worktree generation does not match cleanup registration"
		}
	}
	if strings.TrimSpace(gitOutput(receipt.Worktree, "status", "--porcelain", "--untracked-files=all")) != "" {
		return "target worktree is dirty; non-force cleanup refused"
	}
	if code, output := runGitCommand(receipt.Worktree, "ls-files", "--others", "--ignored", "--exclude-standard", "--directory", "-z"); code != 0 {
		return "ignored-file inspection failed: " + strings.TrimSpace(output)
	} else if len(output) > 0 {
		return "target worktree contains ignored files; cleanup preserves local ignored data"
	}
	if ref := gitOutput(root, "rev-parse", "refs/heads/"+receipt.Branch+"^{commit}"); ref != receipt.CommitSHA {
		return "local branch ref does not durably own the delivered commit"
	}
	return ""
}

func validateTerminalDelivery(
	root string,
	receipt terminalCleanupReceipt,
	defaults []terminalRemoteEvidence,
) (terminalDeliveryEvidence, string) {
	evidence, issue := collectTerminalDeliveryEvidence(root, receipt, defaults)
	if issue != "" {
		return terminalDeliveryEvidence{}, issue
	}
	if receipt.EvidenceRecorded {
		if issue := validateTerminalDeliveryMonotonic(receipt, evidence); issue != "" {
			return terminalDeliveryEvidence{}, issue
		}
	}
	if receipt.DeliveryMode == "direct" && !evidence.ContainedOnDefault {
		return evidence, "remote default does not contain the direct-delivery commit; squash/rebase integration requires provider-backed merge proof and is retained"
	}
	return evidence, ""
}

func collectTerminalRemoteDefaults(root string, receipt terminalCleanupReceipt) ([]terminalRemoteEvidence, string) {
	var defaults []terminalRemoteEvidence
	remotes := strings.Fields(gitOutput(root, "remote"))
	if len(remotes) == 0 {
		return nil, "authoritative remote default branch is unresolved: repository has no configured remotes"
	}
	for _, remote := range remotes {
		ref, defaultBranch, issue := fetchAuthoritativeRemoteDefault(root, remote)
		if issue != "" {
			return nil, issue
		}
		defaultSHA := gitOutput(root, "rev-parse", ref+"^{commit}")
		if defaultSHA == "" {
			return nil, "fetched authoritative remote default has no commit"
		}
		if receipt.Branch != "" && defaultBranch == receipt.Branch {
			return nil, "authoritative remote default branch cannot be a cleanup target"
		}
		defaults = append(defaults, terminalRemoteEvidence{
			Remote: remote, DefaultBranch: defaultBranch, DefaultSHA: defaultSHA,
		})
	}
	sort.Slice(defaults, func(i, j int) bool {
		return defaults[i].Remote < defaults[j].Remote
	})
	return defaults, ""
}

func collectTerminalDeliveryEvidence(
	root string,
	receipt terminalCleanupReceipt,
	defaults []terminalRemoteEvidence,
) (terminalDeliveryEvidence, string) {
	evidence := terminalDeliveryEvidence{}
	if defaults == nil {
		var issue string
		defaults, issue = collectTerminalRemoteDefaults(root, receipt)
		if issue != "" {
			return terminalDeliveryEvidence{}, issue
		}
	}
	evidence.RemoteDefaults = append(evidence.RemoteDefaults, defaults...)
	for _, item := range evidence.RemoteDefaults {
		if item.DefaultBranch == receipt.Branch {
			return terminalDeliveryEvidence{}, "authoritative remote default branch cannot be a cleanup target"
		}
		if item.Remote == receipt.Remote {
			evidence.DefaultBranch = item.DefaultBranch
		}
	}
	if receipt.DeliveryMode == "local" {
		if ref := gitOutput(root, "rev-parse", "refs/heads/"+receipt.Branch+"^{commit}"); ref != receipt.CommitSHA {
			return terminalDeliveryEvidence{}, "local delivery requires an exact durable branch ref at the delivered commit"
		}
		return evidence, ""
	}
	if receipt.Remote == "" {
		return terminalDeliveryEvidence{}, "remote is required for PR or direct delivery cleanup"
	}
	defaultRef := ""
	for _, item := range evidence.RemoteDefaults {
		if item.Remote == receipt.Remote {
			defaultRef = "refs/remotes/" + item.Remote + "/" + item.DefaultBranch
			break
		}
	}
	if defaultRef == "" {
		return terminalDeliveryEvidence{}, "delivery remote is not configured"
	}
	contained, issue := gitCommitContains(root, defaultRef, receipt.CommitSHA)
	if issue != "" {
		return terminalDeliveryEvidence{}, issue
	}
	evidence.ContainedOnDefault = contained
	if receipt.DeliveryMode == "direct" {
		return evidence, ""
	}
	if contained {
		return evidence, ""
	}
	topicRef, issue := fetchExactRemoteBranch(root, receipt.Remote, receipt.Branch)
	if issue != "" {
		return evidence, "remote topic containment is unavailable: " + issue
	}
	contained, issue = gitCommitContains(root, topicRef, receipt.CommitSHA)
	if issue != "" {
		return evidence, issue
	}
	if !contained {
		return evidence, "remote topic does not contain the PR delivery commit"
	}
	evidence.TopicRef = topicRef
	evidence.TopicSHA = gitOutput(root, "rev-parse", topicRef+"^{commit}")
	return evidence, ""
}

func validateTerminalDeliveryMonotonic(receipt terminalCleanupReceipt, current terminalDeliveryEvidence) string {
	if len(receipt.RemoteDefaults) != len(current.RemoteDefaults) {
		return "configured remote-default evidence set changed after cleanup registration"
	}
	for index, before := range receipt.RemoteDefaults {
		after := current.RemoteDefaults[index]
		if before.Remote != after.Remote || before.DefaultBranch != after.DefaultBranch {
			return "authoritative remote default name changed after cleanup registration"
		}
		contained, issue := gitCommitContains(receipt.RepoRoot, "refs/remotes/"+after.Remote+"/"+after.DefaultBranch, before.DefaultSHA)
		if issue != "" {
			return issue
		}
		if !contained {
			return "authoritative remote default history was rewritten after cleanup registration"
		}
	}
	if receipt.DeliveryMode == "pr" {
		if current.ContainedOnDefault {
			return ""
		}
		if receipt.TopicRef == "" || receipt.TopicSHA == "" ||
			current.TopicRef != receipt.TopicRef {
			return "remote topic evidence changed after cleanup registration"
		}
		contained, issue := gitCommitContains(receipt.RepoRoot, current.TopicRef, receipt.TopicSHA)
		if issue != "" {
			return issue
		}
		if !contained {
			return "remote topic history was rewritten after cleanup registration"
		}
	}
	return ""
}

func fetchAuthoritativeRemoteDefault(root, remote string) (string, string, string) {
	code, output := runGitCommand(root, "ls-remote", "--symref", remote, "HEAD")
	if code != 0 {
		return "", "", "resolve remote default authority: " + strings.TrimSpace(output)
	}
	branch := ""
	headSHA := ""
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "ref:" && fields[2] == "HEAD" {
			branch = strings.TrimPrefix(fields[1], "refs/heads/")
		}
		if len(fields) == 2 && fields[1] == "HEAD" {
			headSHA = fields[0]
		}
	}
	if branch == "" || headSHA == "" {
		return "", "", "remote default branch authority is unresolved for " + remote
	}
	ref := "refs/remotes/" + remote + "/" + branch
	refspec := "+refs/heads/*:refs/remotes/" + remote + "/*"
	if code, fetchOutput := runGitCommand(root, "fetch", "--prune", "--no-tags", remote, refspec); code != 0 {
		return "", "", "fetch authoritative remote snapshot: " + strings.TrimSpace(fetchOutput)
	}
	if fetched := gitOutput(root, "rev-parse", ref+"^{commit}"); fetched != headSHA {
		return "", "", "remote default moved while delivery containment was being verified"
	}
	return ref, branch, ""
}

func fetchExactRemoteBranch(root, remote, branch string) (string, string) {
	remoteRef := "refs/heads/" + branch
	code, output := runGitCommand(root, "ls-remote", "--heads", remote, remoteRef)
	if code != 0 {
		return "", "resolve remote branch: " + strings.TrimSpace(output)
	}
	fields := strings.Fields(output)
	if len(fields) != 2 || fields[1] != remoteRef {
		return "", "remote branch is unavailable: " + remoteRef
	}
	expected := fields[0]
	localRef := "refs/remotes/" + remote + "/" + branch
	if fetched := gitOutput(root, "rev-parse", localRef+"^{commit}"); fetched != expected {
		return "", "remote branch changed after the authoritative remote snapshot was fetched"
	}
	return localRef, ""
}

func gitCommitContains(root, containerRef, commit string) (bool, string) {
	if gitOutput(root, "rev-parse", containerRef+"^{commit}") == "" ||
		gitOutput(root, "rev-parse", commit+"^{commit}") == "" {
		return false, "containment verification references an unavailable commit"
	}
	code, output := runGitCommand(root, "merge-base", "--is-ancestor", commit, containerRef)
	switch code {
	case 0:
		return true, ""
	case 1:
		return false, ""
	default:
		return false, "containment verification failed: " + strings.TrimSpace(output)
	}
}

func matchingTerminalCleanupClaim(queue []terminalCleanupQueueEntry, receipt terminalCleanupReceipt) (terminalCleanupQueueEntry, string) {
	identityFound := false
	for _, entry := range queue {
		if entry.WorkID != receipt.WorkID || entry.SessionID != receipt.SessionID {
			continue
		}
		identityFound = true
		if entry.Branch == receipt.Branch && samePath(entry.Worktree, receipt.Worktree) {
			return entry, ""
		}
	}
	if identityFound {
		return terminalCleanupQueueEntry{}, "queue claim identity does not match cleanup target"
	}
	return terminalCleanupQueueEntry{}, "matching work queue claim is unavailable"
}

func terminalCleanupQueuePath(root string) (string, error) {
	common, err := terminalCleanupCommonDir(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(common, ".copilot-kb", "work-queue.json"), nil
}

func loadTerminalCleanupQueue(root string) ([]terminalCleanupQueueEntry, error) {
	path, err := terminalCleanupQueuePath(root)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read work queue: %w", err)
	}
	content = bytes.TrimPrefix(content, []byte{0xef, 0xbb, 0xbf})
	var queue []terminalCleanupQueueEntry
	if err := json.Unmarshal(content, &queue); err == nil {
		return queue, nil
	}
	var one terminalCleanupQueueEntry
	if err := json.Unmarshal(content, &one); err != nil {
		return nil, fmt.Errorf("parse work queue: %w", err)
	}
	return []terminalCleanupQueueEntry{one}, nil
}

func terminalCleanupReceiptDir(root string) (string, error) {
	common, err := terminalCleanupCommonDir(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(common, ".copilot-kb", "terminal-cleanup"), nil
}

func terminalCleanupReceiptPath(root string, receipt terminalCleanupReceipt) (string, error) {
	dir, err := terminalCleanupReceiptDir(root)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, safePathPart(receipt.SessionID)+"-"+safePathPart(receipt.WorkID)+".json"), nil
}

func saveTerminalCleanupReceipt(root string, receipt terminalCleanupReceipt) error {
	path, err := terminalCleanupReceiptPath(root, receipt)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, append(content, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

func loadTerminalCleanupReceipt(root string, identity terminalCleanupReceipt) (terminalCleanupReceipt, error) {
	path, err := terminalCleanupReceiptPath(root, identity)
	if err != nil {
		return terminalCleanupReceipt{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return terminalCleanupReceipt{}, fmt.Errorf("load terminal cleanup receipt: %w", err)
	}
	var receipt terminalCleanupReceipt
	if err := json.Unmarshal(content, &receipt); err != nil {
		return terminalCleanupReceipt{}, fmt.Errorf("decode terminal cleanup receipt: %w", err)
	}
	if receipt.WorkID != identity.WorkID || receipt.SessionID != identity.SessionID {
		return terminalCleanupReceipt{}, fmt.Errorf("terminal cleanup receipt identity changed")
	}
	return receipt, nil
}

func loadTerminalCleanupReceipts(root string) ([]terminalCleanupReceipt, error) {
	dir, err := terminalCleanupReceiptDir(root)
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	receipts := []terminalCleanupReceipt{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var receipt terminalCleanupReceipt
		if err := json.Unmarshal(content, &receipt); err != nil {
			return nil, fmt.Errorf("parse terminal cleanup receipt %s: %w", entry.Name(), err)
		}
		if receipt.SchemaVersion != terminalCleanupSchemaVersion {
			return nil, fmt.Errorf("unsupported terminal cleanup receipt schema_version %d", receipt.SchemaVersion)
		}
		receipts = append(receipts, receipt)
	}
	return receipts, nil
}

func terminalCleanupCommonDir(root string) (string, error) {
	value := gitOutput(root, "rev-parse", "--git-common-dir")
	if value == "" {
		return "", fmt.Errorf("Git common directory is unavailable")
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(root, value)
	}
	return filepath.Abs(filepath.Clean(value))
}

func terminalCleanupPrimaryCheckout(root string) (string, error) {
	common, err := terminalCleanupCommonDir(root)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(filepath.Base(common), ".git") {
		return "", fmt.Errorf("primary checkout cannot be resolved from Git common directory")
	}
	return filepath.Dir(common), nil
}

func createTerminalWorktreeMarker(worktree string) (string, string, string, error) {
	gitDir, err := terminalWorktreeGitDir(worktree)
	if err != nil {
		return "", "", "", err
	}
	path := filepath.Join(gitDir, "copilot-terminal-cleanup-id")
	var tokenBytes [16]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return "", "", "", fmt.Errorf("generate worktree identity: %w", err)
	}
	token := hex.EncodeToString(tokenBytes[:])
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return "", "", "", fmt.Errorf("worktree already has a terminal cleanup identity marker")
		}
		return "", "", "", fmt.Errorf("create worktree identity marker: %w", err)
	}
	if _, err := file.WriteString(token + "\n"); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", "", "", fmt.Errorf("write worktree identity marker: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return "", "", "", fmt.Errorf("close worktree identity marker: %w", err)
	}
	return token, path, gitDir, nil
}

func readTerminalWorktreeMarker(worktree string) (string, error) {
	gitDir, err := terminalWorktreeGitDir(worktree)
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(filepath.Join(gitDir, "copilot-terminal-cleanup-id"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

func terminalWorktreeGitDir(worktree string) (string, error) {
	value := gitOutput(worktree, "rev-parse", "--absolute-git-dir")
	if value == "" {
		return "", fmt.Errorf("worktree Git directory is unavailable")
	}
	return filepath.Abs(filepath.Clean(value))
}

func validateTerminalWorktreeRoundTrip(root string, receipt terminalCleanupReceipt) string {
	info, err := os.Lstat(receipt.Worktree)
	if err != nil {
		return "target worktree path is unreadable"
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return "target worktree path became a symbolic link after cleanup registration"
	}
	realPath, err := filepath.EvalSymlinks(receipt.Worktree)
	if err != nil {
		return "target worktree real path is unreadable"
	}
	if receipt.WorktreeRealPath == "" || !samePath(realPath, receipt.WorktreeRealPath) {
		return "target worktree real path changed after cleanup registration"
	}
	currentGitDir, err := terminalWorktreeGitDir(receipt.Worktree)
	if err != nil || receipt.WorktreeGitDir == "" || !samePath(currentGitDir, receipt.WorktreeGitDir) {
		return "target worktree admin identity changed after cleanup registration"
	}
	common, err := terminalCleanupCommonDir(root)
	if err != nil {
		return err.Error()
	}
	adminParent := filepath.Join(common, "worktrees")
	if !samePath(filepath.Dir(receipt.WorktreeGitDir), adminParent) {
		return "target worktree admin directory is outside the registered Git common directory"
	}
	worktreeGitFile := filepath.Join(receipt.Worktree, ".git")
	content, err := os.ReadFile(worktreeGitFile)
	if err != nil {
		return "target worktree .git pointer is unreadable"
	}
	pointer := strings.TrimSpace(string(content))
	if !strings.HasPrefix(pointer, "gitdir: ") {
		return "target worktree .git pointer is malformed"
	}
	fromWorktree := strings.TrimSpace(strings.TrimPrefix(pointer, "gitdir: "))
	if !filepath.IsAbs(fromWorktree) {
		fromWorktree = filepath.Join(receipt.Worktree, fromWorktree)
	}
	if !samePath(fromWorktree, receipt.WorktreeGitDir) {
		return "target worktree .git pointer does not round-trip to its admin directory"
	}
	adminPointer, err := os.ReadFile(filepath.Join(receipt.WorktreeGitDir, "gitdir"))
	if err != nil {
		return "target worktree admin pointer is unreadable"
	}
	fromAdmin := strings.TrimSpace(string(adminPointer))
	if !filepath.IsAbs(fromAdmin) {
		fromAdmin = filepath.Join(receipt.WorktreeGitDir, fromAdmin)
	}
	if !samePath(fromAdmin, worktreeGitFile) {
		return "target worktree admin pointer does not round-trip to its worktree"
	}
	return ""
}

func listTerminalGitWorktrees(root string) (map[string]terminalGitWorktree, error) {
	code, output := runGitCommand(root, "worktree", "list", "--porcelain", "-z")
	if code != 0 {
		return nil, fmt.Errorf("list Git worktrees: %s", strings.TrimSpace(output))
	}
	result := map[string]terminalGitWorktree{}
	var current terminalGitWorktree
	flush := func() {
		if current.Path == "" {
			return
		}
		current.Path, _ = filepath.Abs(filepath.Clean(current.Path))
		result[terminalPathKey(current.Path)] = current
		current = terminalGitWorktree{}
	}
	for _, line := range strings.Split(output, "\x00") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			flush()
			current.Path = strings.TrimPrefix(line, "worktree ")
		case strings.HasPrefix(line, "HEAD "):
			current.Head = strings.TrimPrefix(line, "HEAD ")
		case strings.HasPrefix(line, "branch refs/heads/"):
			current.Branch = strings.TrimPrefix(line, "branch refs/heads/")
		case strings.HasPrefix(line, "locked"):
			current.Locked = true
		case line == "":
			flush()
		}
	}
	flush()
	return result, nil
}

func terminalPathKey(path string) string {
	absolute, _ := filepath.Abs(filepath.Clean(path))
	return strings.ToLower(absolute)
}

func blockedTerminalCleanup(action, issue string, receipt *terminalCleanupReceipt) terminalCleanupResult {
	return terminalCleanupResult{OK: false, Action: action, Issue: issue, Receipt: receipt}
}
