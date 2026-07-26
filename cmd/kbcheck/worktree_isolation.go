package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const worktreeIsolationSchemaVersion = 1

type worktreeCommandOptions struct {
	Action              string
	SliceID             string
	RunID               string
	OwnerToken          string
	BaseSHA             string
	Worktree            string
	Branch              string
	RepoRoot            string
	LegacyCompatibility bool
	Now                 time.Time
}

type worktreeReceipt struct {
	SchemaVersion int      `json:"schema_version"`
	SliceID       string   `json:"slice_id"`
	RunID         string   `json:"run_id"`
	OwnerToken    string   `json:"owner_token"`
	RepoRoot      string   `json:"repo_root"`
	Worktree      string   `json:"worktree"`
	Branch        string   `json:"branch"`
	BaseSHA       string   `json:"base_sha"`
	Status        string   `json:"status"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
	IntegratedAt  string   `json:"integrated_at,omitempty"`
	ReleasedAt    string   `json:"released_at,omitempty"`
	Limitations   []string `json:"limitations"`
}

type worktreeResult struct {
	OK      bool             `json:"ok"`
	Action  string           `json:"action"`
	Issue   string           `json:"issue,omitempty"`
	Receipt *worktreeReceipt `json:"receipt,omitempty"`
}

func runWorktreeCommand(root string, opts options, stdout, stderr io.Writer) int {
	command := worktreeCommandOptions{
		Action:              opts.sliceLeaseAction,
		SliceID:             opts.sliceID,
		RunID:               opts.runID,
		OwnerToken:          opts.ownerToken,
		BaseSHA:             opts.baseSHA,
		Worktree:            opts.worktreePath,
		Branch:              opts.branchName,
		RepoRoot:            root,
		LegacyCompatibility: opts.legacySliceWorktree,
		Now:                 time.Now().UTC(),
	}
	result, err := executeWorktreeCommand(command)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK && result.Receipt != nil {
		fmt.Fprintf(stdout, "worktree: %s slice=%s status=%s path=%s\n", result.Action, result.Receipt.SliceID, result.Receipt.Status, result.Receipt.Worktree)
	} else if result.OK {
		fmt.Fprintf(stdout, "worktree: %s ok\n", result.Action)
	} else {
		fmt.Fprintf(stdout, "worktree: %s blocked: %s\n", result.Action, result.Issue)
	}
	if !result.OK {
		return 2
	}
	return 0
}

func executeWorktreeCommand(opts worktreeCommandOptions) (worktreeResult, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action == "" {
		return worktreeResult{}, fmt.Errorf("worktree requires --action")
	}
	if !opts.LegacyCompatibility {
		return blockedWorktree(action, "legacy per-slice worktree command requires explicit --legacy-slice-worktree; plan runs use plan-worktree", nil), nil
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	root, err := filepath.Abs(filepath.Clean(opts.RepoRoot))
	if err != nil {
		return worktreeResult{}, err
	}
	opts.RepoRoot = root
	switch action {
	case "prepare":
		return prepareWorktree(opts)
	case "status":
		return statusWorktree(opts)
	case "integrate":
		return integrateWorktree(opts)
	case "release":
		return releaseWorktree(opts)
	default:
		return worktreeResult{}, fmt.Errorf("unsupported worktree action %q", opts.Action)
	}
}

func prepareWorktree(opts worktreeCommandOptions) (worktreeResult, error) {
	if issue := requireWorktreeIdentity(opts); issue != "" {
		return blockedWorktree("prepare", issue, nil), nil
	}
	if active, issue := activePlanRunForLegacyWorktree(opts); issue != "" {
		return blockedWorktree("prepare", issue, nil), nil
	} else if active {
		return blockedWorktree("prepare", "legacy per-slice worktrees are forbidden for plan runs; use the manifest-owned plan-worktree", nil), nil
	}
	lease, issue := activeWorktreeLease(opts)
	if issue != "" {
		return blockedWorktree("prepare", issue, nil), nil
	}
	if opts.BaseSHA == "" {
		opts.BaseSHA = gitOutput(opts.RepoRoot, "rev-parse", "HEAD")
	}
	if opts.BaseSHA == "" {
		return blockedWorktree("prepare", "base revision is unavailable", nil), nil
	}
	if opts.Worktree == "" {
		opts.Worktree = defaultSliceWorktreePath(opts)
	}
	absWorktree, err := filepath.Abs(filepath.Clean(opts.Worktree))
	if err != nil {
		return worktreeResult{}, err
	}
	opts.Worktree = absWorktree
	if opts.Branch == "" {
		opts.Branch = defaultSliceBranch(opts)
	}
	if samePath(opts.Worktree, opts.RepoRoot) {
		return blockedWorktree("prepare", "worktree path must not be the source checkout", nil), nil
	}
	receipt := worktreeReceipt{
		SchemaVersion: worktreeIsolationSchemaVersion,
		SliceID:       opts.SliceID,
		RunID:         opts.RunID,
		OwnerToken:    opts.OwnerToken,
		RepoRoot:      opts.RepoRoot,
		Worktree:      opts.Worktree,
		Branch:        opts.Branch,
		BaseSHA:       opts.BaseSHA,
		Status:        "prepared",
		CreatedAt:     opts.Now.Format(time.RFC3339Nano),
		UpdatedAt:     opts.Now.Format(time.RFC3339Nano),
		Limitations: []string{
			"worktrees isolate filesystem mutation but not task ownership",
			"integration must be serialized by one coordinator",
			"cleanup never uses force and requires integrated clean state",
		},
	}
	if existing, err := loadWorktreeReceipt(opts); err == nil && existing.Status != "" {
		if existing.OwnerToken != opts.OwnerToken || existing.SliceID != opts.SliceID {
			return blockedWorktree("prepare", "existing receipt belongs to another owner", &existing), nil
		}
		return worktreeResult{OK: true, Action: "prepare", Receipt: &existing}, nil
	}
	if pathExists(opts.Worktree) {
		return blockedWorktree("prepare", "worktree path already exists without a matching receipt", nil), nil
	}
	if code, out := runGitCommand(opts.RepoRoot, "worktree", "add", "-b", opts.Branch, opts.Worktree, opts.BaseSHA); code != 0 {
		return blockedWorktree("prepare", strings.TrimSpace(out), nil), nil
	}
	receipt.BaseSHA = lease.BaseSHA
	if receipt.BaseSHA == "" {
		receipt.BaseSHA = opts.BaseSHA
	}
	if err := saveWorktreeReceipt(opts, receipt); err != nil {
		return worktreeResult{}, err
	}
	return worktreeResult{OK: true, Action: "prepare", Receipt: &receipt}, nil
}

func statusWorktree(opts worktreeCommandOptions) (worktreeResult, error) {
	receipt, err := loadWorktreeReceipt(opts)
	if err != nil {
		return blockedWorktree("status", err.Error(), nil), nil
	}
	return worktreeResult{OK: true, Action: "status", Receipt: &receipt}, nil
}

func integrateWorktree(opts worktreeCommandOptions) (worktreeResult, error) {
	if issue := requireWorktreeIdentity(opts); issue != "" {
		return blockedWorktree("integrate", issue, nil), nil
	}
	receipt, err := loadWorktreeReceipt(opts)
	if err != nil {
		return blockedWorktree("integrate", err.Error(), nil), nil
	}
	if receipt.OwnerToken != opts.OwnerToken {
		return blockedWorktree("integrate", "owner token does not match receipt", &receipt), nil
	}
	if _, issue := activeWorktreeLease(opts); issue != "" {
		return blockedWorktree("integrate", issue, &receipt), nil
	}
	targetBranch := gitOutput(opts.RepoRoot, "branch", "--show-current")
	if targetBranch == "" {
		return blockedWorktree("integrate", "internal integration target branch is unavailable", &receipt), nil
	}
	if isResolvedDefaultBranch(opts.RepoRoot, targetBranch) {
		return blockedWorktree("integrate", "internal integration target resolves to a local or remote default branch", &receipt), nil
	}
	head := gitOutput(opts.RepoRoot, "rev-parse", "HEAD")
	if head == "" || receipt.BaseSHA == "" || head != receipt.BaseSHA {
		return blockedWorktree("integrate", "source base revision changed before integration", &receipt), nil
	}
	if code, out := runGitCommand(opts.Worktree, "status", "--porcelain"); code != 0 {
		return blockedWorktree("integrate", strings.TrimSpace(out), &receipt), nil
	} else if strings.TrimSpace(out) != "" {
		return blockedWorktree("integrate", "worktree has uncommitted changes; commit or discard inside the isolated worktree first", &receipt), nil
	}
	if code, out := runGitCommand(opts.RepoRoot, "merge", "--no-ff", "--no-edit", receipt.Branch); code != 0 {
		return blockedWorktree("integrate", strings.TrimSpace(out), &receipt), nil
	}
	receipt.Status = "integrated"
	receipt.IntegratedAt = opts.Now.Format(time.RFC3339Nano)
	receipt.UpdatedAt = receipt.IntegratedAt
	if err := saveWorktreeReceipt(opts, receipt); err != nil {
		return worktreeResult{}, err
	}
	return worktreeResult{OK: true, Action: "integrate", Receipt: &receipt}, nil
}

func releaseWorktree(opts worktreeCommandOptions) (worktreeResult, error) {
	if issue := requireWorktreeIdentity(opts); issue != "" {
		return blockedWorktree("release", issue, nil), nil
	}
	receipt, err := loadWorktreeReceipt(opts)
	if err != nil {
		return blockedWorktree("release", err.Error(), nil), nil
	}
	if receipt.OwnerToken != opts.OwnerToken {
		return blockedWorktree("release", "owner token does not match receipt", &receipt), nil
	}
	if receipt.Status != "integrated" {
		return blockedWorktree("release", "worktree must be integrated before release", &receipt), nil
	}
	if code, out := runGitCommand(receipt.Worktree, "status", "--porcelain"); code != 0 {
		return blockedWorktree("release", strings.TrimSpace(out), &receipt), nil
	} else if strings.TrimSpace(out) != "" {
		return blockedWorktree("release", "worktree is dirty; cleanup refuses non-force removal", &receipt), nil
	}
	lease, issue := activeWorktreeLease(opts)
	if issue != "" {
		return blockedWorktree("release", issue, &receipt), nil
	}
	if code, out := runGitCommand(opts.RepoRoot, "worktree", "remove", receipt.Worktree); code != 0 {
		return blockedWorktree("release", strings.TrimSpace(out), &receipt), nil
	}
	released, err := executeSliceLease(sliceLeaseCommandOptions{
		Action:     "release",
		SliceID:    opts.SliceID,
		OwnerToken: opts.OwnerToken,
		Generation: lease.Generation,
		RepoRoot:   opts.RepoRoot,
		Now:        opts.Now,
	})
	if err != nil {
		return worktreeResult{}, err
	}
	if !released.OK {
		return blockedWorktree("release", "slice lease release failed: "+released.Issue, &receipt), nil
	}
	receipt.Status = "released"
	receipt.ReleasedAt = opts.Now.Format(time.RFC3339Nano)
	receipt.UpdatedAt = receipt.ReleasedAt
	if err := saveWorktreeReceipt(opts, receipt); err != nil {
		return worktreeResult{}, err
	}
	return worktreeResult{OK: true, Action: "release", Receipt: &receipt}, nil
}

func requireWorktreeIdentity(opts worktreeCommandOptions) string {
	if strings.TrimSpace(opts.SliceID) == "" || strings.TrimSpace(opts.RunID) == "" {
		return "slice-id and run-id are required"
	}
	if strings.TrimSpace(opts.OwnerToken) == "" {
		return "owner-token is required"
	}
	return ""
}

func activeWorktreeLease(opts worktreeCommandOptions) (sliceLease, string) {
	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return sliceLease{}, err.Error()
	}
	state, err := loadSliceLeaseState(stateRoot)
	if err != nil {
		return sliceLease{}, err.Error()
	}
	_, lease, ok := findSliceLease(state, sliceLeaseCommandOptions{
		SliceID:    opts.SliceID,
		RunID:      opts.RunID,
		OwnerToken: opts.OwnerToken,
	})
	if !ok {
		return sliceLease{}, "slice lease not found"
	}
	if lease.OwnerToken != opts.OwnerToken {
		return lease, "owner token does not match active slice lease"
	}
	if effectiveLeaseStatus(lease, opts.Now) != "active" {
		return lease, "slice lease is not active"
	}
	return lease, ""
}

func activePlanRunForLegacyWorktree(opts worktreeCommandOptions) (bool, string) {
	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return false, err.Error()
	}
	state, err := loadPlanRunLeaseState(stateRoot)
	if err != nil {
		return false, err.Error()
	}
	for _, lease := range state.Leases {
		if lease.RunID == opts.RunID && effectivePlanRunLeaseStatus(lease, opts.Now) == "active" {
			return true, ""
		}
	}
	return false, ""
}

func worktreeReceiptPath(opts worktreeCommandOptions) (string, error) {
	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(stateRoot), "worktrees", safePathPart(opts.RunID), safePathPart(opts.SliceID)+".json"), nil
}

func loadWorktreeReceipt(opts worktreeCommandOptions) (worktreeReceipt, error) {
	path, err := worktreeReceiptPath(opts)
	if err != nil {
		return worktreeReceipt{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return worktreeReceipt{}, err
	}
	var receipt worktreeReceipt
	if err := json.Unmarshal(content, &receipt); err != nil {
		return worktreeReceipt{}, err
	}
	if receipt.SchemaVersion != worktreeIsolationSchemaVersion {
		return worktreeReceipt{}, fmt.Errorf("unsupported worktree receipt schema_version %d", receipt.SchemaVersion)
	}
	return receipt, nil
}

func saveWorktreeReceipt(opts worktreeCommandOptions, receipt worktreeReceipt) error {
	path, err := worktreeReceiptPath(opts)
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

func defaultSliceWorktreePath(opts worktreeCommandOptions) string {
	parent := filepath.Dir(opts.RepoRoot)
	name := filepath.Base(opts.RepoRoot)
	return filepath.Join(parent, ".kb-worktrees", safePathPart(name), safePathPart(opts.SliceID))
}

func defaultSliceBranch(opts worktreeCommandOptions) string {
	return "codex/" + safePathPart(opts.RunID) + "/" + safePathPart(opts.SliceID) + "-" + randomSuffix()
}

func safePathPart(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	re := regexp.MustCompile(`[^a-z0-9._-]+`)
	value = re.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-._")
	if value == "" {
		return "slice"
	}
	if len(value) > 64 {
		return value[:64]
	}
	return value
}

func randomSuffix() string {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

func runGitCommand(root string, args ...string) (int, string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode(), string(out)
		}
		return 1, err.Error()
	}
	return 0, string(out)
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(filepath.Clean(a))
	bb, errB := filepath.Abs(filepath.Clean(b))
	if errA != nil || errB != nil {
		return false
	}
	return strings.EqualFold(aa, bb)
}

func blockedWorktree(action, issue string, receipt *worktreeReceipt) worktreeResult {
	if strings.TrimSpace(issue) == "" {
		issue = "blocked"
	}
	return worktreeResult{OK: false, Action: action, Issue: issue, Receipt: receipt}
}

func sortedWorktreeStrings(values []string) []string {
	out := append([]string{}, values...)
	sort.Strings(out)
	return out
}
