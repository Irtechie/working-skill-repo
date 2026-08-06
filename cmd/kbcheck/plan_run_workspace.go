package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const planRunWorkspaceSchemaVersion = 1

const (
	planRunWorkspaceOwnerKB      = "kb"
	planRunWorkspaceOwnerHarness = "harness"
)

// Parallel worktrees are not the throughput control; the claim DAG already
// serializes colliding slices. Extra trees buy merge cost and disk, so the
// ceiling counts harness-created trees too.
const defaultMaxLinkedWorktrees = 2

type planRunWorkspaceOptions struct {
	Action                  string
	ManifestPath            string
	OwnerToken              string
	BaseSHA                 string
	Worktree                string
	IntegrationRef          string
	RunID                   string
	SliceID                 string
	ExpectedIntegrationHead string
	CommitSHA               string
	ProofReceipt            string
	CommitAuthorized        bool
	CommitAuthorizedBy      string
	CommitApprovalRef       string
	RepoRoot                string
	Now                     time.Time
}

type planRunWorkspaceReceipt struct {
	SchemaVersion        int                    `json:"schema_version"`
	KBID                 string                 `json:"kb_id"`
	RunID                string                 `json:"run_id,omitempty"`
	ManifestPath         string                 `json:"manifest_path"`
	OwnerToken           string                 `json:"owner_token"`
	OwnerFingerprint     string                 `json:"owner_fingerprint,omitempty"`
	SourceCheckout       string                 `json:"source_checkout"`
	SourceDirty          bool                   `json:"source_dirty"`
	Worktree             string                 `json:"worktree"`
	WorkspaceOwner       string                 `json:"workspace_owner,omitempty"`
	BaseRef              string                 `json:"base_ref"`
	BaseSHA              string                 `json:"base_sha"`
	IntegrationRef       string                 `json:"integration_ref"`
	IntegrationHead      string                 `json:"integration_head"`
	LastSliceID          string                 `json:"last_slice_id,omitempty"`
	LastSliceCommit      string                 `json:"last_slice_commit,omitempty"`
	LastProofReceipt     string                 `json:"last_proof_receipt,omitempty"`
	LastProofSHA256      string                 `json:"last_proof_sha256,omitempty"`
	LastProofArchive     string                 `json:"last_proof_archive,omitempty"`
	AcceptedProofs       []planRunAcceptedProof `json:"accepted_proofs,omitempty"`
	CommitAuthorized     bool                   `json:"commit_authorized"`
	CommitAuthorizedBy   string                 `json:"commit_authorized_by,omitempty"`
	CommitApprovalRef    string                 `json:"commit_approval_ref,omitempty"`
	CommitAuthorizedAt   string                 `json:"commit_authorized_at,omitempty"`
	DeliveryMode         string                 `json:"delivery_mode"`
	DeliveryMerge        string                 `json:"delivery_merge"`
	DeliveryPolicySource string                 `json:"delivery_policy_source"`
	Status               string                 `json:"status"`
	CleanupState         string                 `json:"cleanup_state"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
	ReleasedAt           string                 `json:"released_at,omitempty"`
	Limitations          []string               `json:"limitations"`
}

type planRunAcceptedProof struct {
	SliceID    string `json:"slice_id"`
	CommitSHA  string `json:"commit_sha"`
	Archive    string `json:"archive"`
	SHA256     string `json:"sha256"`
	AcceptedAt string `json:"accepted_at"`
}

type planRunWorkspaceResult struct {
	OK      bool                     `json:"ok"`
	Action  string                   `json:"action"`
	Issue   string                   `json:"issue,omitempty"`
	Receipt *planRunWorkspaceReceipt `json:"receipt,omitempty"`
}

type kbDeliveryPolicy struct {
	Mode          string
	Merge         string
	PostMergeSync bool
	Source        string
}

type planRunProofCommand struct {
	Args         []string `json:"args"`
	Expect       int      `json:"expect"`
	ExpectOutput string   `json:"expect_output,omitempty"`
}

type planRunProofReceipt struct {
	SchemaVersion  int                  `json:"schema_version"`
	KBID           string               `json:"kb_id"`
	RunID          string               `json:"run_id"`
	SliceID        string               `json:"slice_id"`
	CommitSHA      string               `json:"commit_sha"`
	ObservedWrites []string             `json:"observed_writes"`
	SliceProof     planRunProofCommand  `json:"slice_proof"`
	AggregateProof *planRunProofCommand `json:"aggregate_proof"`
}

func runPlanRunWorkspaceCommand(root string, opts options, stdout, stderr io.Writer) int {
	result, err := executePlanRunWorkspace(planRunWorkspaceOptions{
		Action:                  opts.sliceLeaseAction,
		ManifestPath:            opts.manifest,
		OwnerToken:              opts.ownerToken,
		BaseSHA:                 opts.baseSHA,
		Worktree:                opts.worktreePath,
		IntegrationRef:          opts.branchName,
		RunID:                   opts.runID,
		SliceID:                 opts.sliceID,
		ExpectedIntegrationHead: opts.expectedIntegrationHead,
		CommitSHA:               opts.commitSHA,
		ProofReceipt:            opts.proofReceipt,
		CommitAuthorized:        opts.commitAuthorized,
		CommitAuthorizedBy:      opts.commitAuthorizedBy,
		CommitApprovalRef:       opts.commitApprovalRef,
		RepoRoot:                root,
		Now:                     time.Now().UTC(),
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	redactPlanRunWorkspaceResult(&result)
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK && result.Receipt != nil {
		fmt.Fprintf(stdout, "plan-worktree: %s kb=%s status=%s path=%s\n", result.Action, result.Receipt.KBID, result.Receipt.Status, result.Receipt.Worktree)
	} else {
		fmt.Fprintf(stdout, "plan-worktree: %s blocked: %s\n", result.Action, result.Issue)
	}
	if !result.OK {
		return 2
	}
	return 0
}

func redactPlanRunWorkspaceResult(result *planRunWorkspaceResult) {
	if result.Receipt == nil {
		return
	}
	copy := *result.Receipt
	public := publicPlanRunLease(planRunLease{OwnerToken: copy.OwnerToken})
	copy.OwnerToken = ""
	copy.OwnerFingerprint = public.OwnerFingerprint
	result.Receipt = &copy
}

func executePlanRunWorkspace(opts planRunWorkspaceOptions) (planRunWorkspaceResult, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action == "" {
		return planRunWorkspaceResult{}, fmt.Errorf("plan-worktree requires --action")
	}
	if action != "prepare" && action != "adopt" && action != "status" && action != "advance" && action != "complete" && action != "release" {
		return planRunWorkspaceResult{}, fmt.Errorf("unsupported plan-worktree action %q", opts.Action)
	}
	if strings.TrimSpace(opts.OwnerToken) == "" {
		return blockedPlanRunWorkspace(action, "owner-token is required", nil), nil
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	root, err := filepath.Abs(filepath.Clean(opts.RepoRoot))
	if err != nil {
		return planRunWorkspaceResult{}, err
	}
	opts.RepoRoot = root
	manifest, err := filepath.Abs(filepath.Clean(opts.ManifestPath))
	if err != nil {
		return planRunWorkspaceResult{}, err
	}
	opts.ManifestPath = manifest
	kbID, err := planRunManifestID(manifest)
	if err != nil {
		return blockedPlanRunWorkspace(action, err.Error(), nil), nil
	}

	existing, loadErr := loadPlanRunWorkspaceReceipt(opts, kbID)
	if action == "status" {
		if loadErr != nil {
			return blockedPlanRunWorkspace(action, loadErr.Error(), nil), nil
		}
		if issue := validatePlanRunOwnerAndIdentity(opts, existing); issue != "" {
			return blockedPlanRunWorkspace(action, issue, &existing), nil
		}
		return planRunWorkspaceResult{OK: true, Action: action, Receipt: &existing}, nil
	}
	if action == "release" {
		if loadErr != nil {
			return blockedPlanRunWorkspace(action, loadErr.Error(), nil), nil
		}
		return releasePlanRunWorkspace(opts, existing)
	}
	if action == "complete" {
		if loadErr != nil {
			return blockedPlanRunWorkspace(action, loadErr.Error(), nil), nil
		}
		return completePlanRunWorkspace(opts, existing)
	}
	if action == "advance" {
		if loadErr != nil {
			return blockedPlanRunWorkspace(action, loadErr.Error(), nil), nil
		}
		return advancePlanRunWorkspace(opts, existing)
	}
	if loadErr == nil {
		if issue := validatePlanRunOwnerAndIdentity(opts, existing); issue != "" {
			return blockedPlanRunWorkspace(action, issue, &existing), nil
		}
		if action == "prepare" || action == "adopt" {
			if !existing.CommitAuthorized {
				return blockedPlanRunWorkspace(action, "existing receipt has no durable commit approval provenance; recreate through an explicitly authorized plan-run preparation", &existing), nil
			}
			if requested := requestedPlanRunWorkspaceOwner(action); requested != planRunWorkspaceOwner(existing) {
				return blockedPlanRunWorkspace(action, fmt.Sprintf("existing plan-run receipt is %s-owned; %s cannot change workspace ownership", planRunWorkspaceOwner(existing), action), &existing), nil
			}
		}
		return planRunWorkspaceResult{OK: true, Action: action, Receipt: &existing}, nil
	}
	if !errors.Is(loadErr, os.ErrNotExist) {
		return blockedPlanRunWorkspace(action, loadErr.Error(), nil), nil
	}
	if !opts.CommitAuthorized || strings.TrimSpace(opts.CommitAuthorizedBy) == "" || strings.TrimSpace(opts.CommitApprovalRef) == "" {
		return blockedPlanRunWorkspace(action, "explicit local plan-run commit authorization requires --commit-authorized, --commit-authorized-by, and --commit-approval-ref before mutation", nil), nil
	}
	if action == "adopt" {
		return adoptPlanRunWorkspace(opts, kbID)
	}
	return preparePlanRunWorkspace(opts, kbID)
}

// planRunWorkspaceOwner reports which system owns the worktree lifecycle.
// Receipts written before adoption existed carry no value and remain KB-owned.
func planRunWorkspaceOwner(receipt planRunWorkspaceReceipt) string {
	if strings.TrimSpace(receipt.WorkspaceOwner) == planRunWorkspaceOwnerHarness {
		return planRunWorkspaceOwnerHarness
	}
	return planRunWorkspaceOwnerKB
}

func requestedPlanRunWorkspaceOwner(action string) string {
	if action == "adopt" {
		return planRunWorkspaceOwnerHarness
	}
	return planRunWorkspaceOwnerKB
}

// adoptPlanRunWorkspace binds a plan run to the harness-provided worktree the
// session already occupies. It creates no worktree and no branch, so a coding
// harness that already isolates each session does not get a nested second
// worktree for the same logical thread.
func adoptPlanRunWorkspace(opts planRunWorkspaceOptions, kbID string) (planRunWorkspaceResult, error) {
	primary, err := terminalCleanupPrimaryCheckout(opts.RepoRoot)
	if err != nil {
		return blockedPlanRunWorkspace("adopt", err.Error(), nil), nil
	}
	if samePath(primary, opts.RepoRoot) {
		return blockedPlanRunWorkspace("adopt", "adopt requires an existing harness-owned linked worktree; run prepare from the primary checkout", nil), nil
	}
	branch := gitOutput(opts.RepoRoot, "branch", "--show-current")
	if branch == "" {
		return blockedPlanRunWorkspace("adopt", "adopted worktree must be on a named branch", nil), nil
	}
	if isResolvedDefaultBranch(opts.RepoRoot, branch) {
		return blockedPlanRunWorkspace("adopt", "integration ref must not be the default branch", nil), nil
	}
	if opts.IntegrationRef != "" && opts.IntegrationRef != branch {
		return blockedPlanRunWorkspace("adopt", fmt.Sprintf("requested integration ref does not match the adopted worktree branch: got %s want %s", opts.IntegrationRef, branch), nil), nil
	}
	if strings.TrimSpace(opts.Worktree) != "" {
		requested, err := filepath.Abs(filepath.Clean(opts.Worktree))
		if err != nil {
			return planRunWorkspaceResult{}, err
		}
		if !samePath(requested, opts.RepoRoot) {
			return blockedPlanRunWorkspace("adopt", "adopt records the current worktree; requested worktree path does not match", nil), nil
		}
	}
	if issue := unresolvedRemoteDefaultAuthority(opts.RepoRoot); issue != "" {
		return blockedPlanRunWorkspace("adopt", issue, nil), nil
	}
	head := gitOutput(opts.RepoRoot, "rev-parse", "HEAD")
	if head == "" {
		return blockedPlanRunWorkspace("adopt", "adopted worktree head is unavailable", nil), nil
	}
	if strings.TrimSpace(opts.BaseSHA) != "" {
		requested := gitOutput(opts.RepoRoot, "rev-parse", opts.BaseSHA+"^{commit}")
		if requested != head {
			return blockedPlanRunWorkspace("adopt", "adoption records the current head as the immutable base; requested base does not match", nil), nil
		}
	}
	dirtyPaths, err := planRunDirtyPaths(opts.RepoRoot)
	if err != nil {
		return blockedPlanRunWorkspace("adopt", err.Error(), nil), nil
	}
	if len(dirtyPaths) > 0 {
		return blockedPlanRunWorkspace("adopt", "adopted worktree must be clean before slice commits; preserved paths: "+strings.Join(dirtyPaths, ", "), nil), nil
	}
	delivery, err := resolveKBDeliveryPolicy(opts.RepoRoot)
	if err != nil {
		return blockedPlanRunWorkspace("adopt", err.Error(), nil), nil
	}
	now := opts.Now.Format(time.RFC3339Nano)
	receipt := planRunWorkspaceReceipt{
		SchemaVersion:        planRunWorkspaceSchemaVersion,
		KBID:                 kbID,
		RunID:                defaultPlanRunID(opts.RunID, kbID),
		ManifestPath:         opts.ManifestPath,
		OwnerToken:           opts.OwnerToken,
		SourceCheckout:       opts.RepoRoot,
		SourceDirty:          false,
		Worktree:             opts.RepoRoot,
		WorkspaceOwner:       planRunWorkspaceOwnerHarness,
		BaseRef:              branch,
		BaseSHA:              head,
		IntegrationRef:       branch,
		IntegrationHead:      head,
		CommitAuthorized:     true,
		CommitAuthorizedBy:   strings.TrimSpace(opts.CommitAuthorizedBy),
		CommitApprovalRef:    strings.TrimSpace(opts.CommitApprovalRef),
		CommitAuthorizedAt:   now,
		DeliveryMode:         delivery.Mode,
		DeliveryMerge:        delivery.Merge,
		DeliveryPolicySource: delivery.Source,
		Status:               "prepared",
		CleanupState:         "active",
		CreatedAt:            now,
		UpdatedAt:            now,
		Limitations: []string{
			"worktree and branch lifecycle belong to the coding harness; release never removes them",
			"plan-run ownership coordinates only worktrees sharing this Git common directory",
			"default-branch delivery requires a separate authorized delivery phase",
			"slice commits are authorized only for this adopted harness-owned branch",
			"cleanup is non-force and requires clean integrated state",
		},
	}
	if err := savePlanRunWorkspaceReceipt(opts, receipt); err != nil {
		return planRunWorkspaceResult{}, err
	}
	return planRunWorkspaceResult{OK: true, Action: "adopt", Receipt: &receipt}, nil
}

func completePlanRunWorkspace(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) (planRunWorkspaceResult, error) {
	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return planRunWorkspaceResult{}, err
	}
	result := planRunWorkspaceResult{Action: "complete"}
	err = withSliceLeaseStateLock(stateRoot, func() error {
		current, err := loadPlanRunWorkspaceReceipt(opts, receipt.KBID)
		if err != nil {
			return err
		}
		if issue := validatePlanRunOwnerAndIdentity(opts, current); issue != "" {
			result = blockedPlanRunWorkspace("complete", issue, &current)
			return nil
		}
		if current.IntegrationHead == current.BaseSHA || current.LastSliceCommit != current.IntegrationHead {
			result = blockedPlanRunWorkspace("complete", "plan-run has no fully accepted slice head", &current)
			return nil
		}
		if issue := validatePlanRunHeadCAS(current); issue != "" {
			result = blockedPlanRunWorkspace("complete", issue, &current)
			return nil
		}
		slices, err := loadSliceLeaseState(stateRoot)
		if err != nil {
			return err
		}
		manifestPath, err := planRunManifestPathInWorktree(current)
		if err != nil {
			return err
		}
		if issue := validateTerminalPlanRunManifest(manifestPath); issue != "" {
			result = blockedPlanRunWorkspace("complete", issue, &current)
			return nil
		}
		if issue := validateReleasedPlanRunSlices(manifestPath, current, slices, opts.Now); issue != "" {
			result = blockedPlanRunWorkspace("complete", issue, &current)
			return nil
		}
		if issue := validateAcceptedPlanRunProofs(opts, current, manifestPath); issue != "" {
			result = blockedPlanRunWorkspace("complete", issue, &current)
			return nil
		}
		plans, err := loadPlanRunLeaseState(stateRoot)
		if err != nil {
			return err
		}
		plan, ok := plans.Leases[current.RunID]
		if !ok {
			result = blockedPlanRunWorkspace("complete", "active plan-run lease is required for atomic completion", &current)
			return nil
		}
		if plan.OwnerToken != current.OwnerToken {
			result = blockedPlanRunWorkspace("complete", "plan-run lease owner does not match workspace owner", &current)
			return nil
		}
		switch {
		case effectivePlanRunLeaseStatus(plan, opts.Now) == "active":
			plan.Generation++
			plan.Status = "released"
			plan.LastUpdatedAt = opts.Now.Format(time.RFC3339Nano)
			plans.Leases[current.RunID] = plan
			appendPlanRunLeaseEvent(&plans, opts.Now, "complete-release", &plan)
			if err := savePlanRunLeaseState(stateRoot, plans); err != nil {
				return err
			}
		case hasPlanRunCompletionReleaseJournal(plans, plan):
			// A prior attempt durably released the lease but failed before the
			// workspace receipt write. The matching journal makes retry safe.
		default:
			result = blockedPlanRunWorkspace("complete", "plan-run lease must be active for atomic completion; recover an expired lease first", &current)
			return nil
		}
		current.Status = "completed"
		current.UpdatedAt = opts.Now.Format(time.RFC3339Nano)
		if err := savePlanRunWorkspaceReceipt(opts, current); err != nil {
			return err
		}
		result = planRunWorkspaceResult{OK: true, Action: "complete", Receipt: &current}
		return nil
	})
	return result, err
}

func hasPlanRunCompletionReleaseJournal(state planRunLeaseState, lease planRunLease) bool {
	if lease.Status != "released" {
		return false
	}
	for index := len(state.Events) - 1; index >= 0; index-- {
		event := state.Events[index]
		if event.RunID != lease.RunID {
			continue
		}
		return event.Action == "complete-release" && event.Generation == lease.Generation
	}
	return false
}

func preparePlanRunWorkspace(opts planRunWorkspaceOptions, kbID string) (planRunWorkspaceResult, error) {
	if opts.BaseSHA == "" {
		opts.BaseSHA = gitOutput(opts.RepoRoot, "rev-parse", "HEAD")
	}
	baseSHA := gitOutput(opts.RepoRoot, "rev-parse", opts.BaseSHA+"^{commit}")
	if baseSHA == "" {
		return blockedPlanRunWorkspace("prepare", "base revision is unavailable", nil), nil
	}
	baseRef := resolvePlanRunBaseRef(opts.RepoRoot, baseSHA)
	if opts.IntegrationRef == "" {
		opts.IntegrationRef = "codex/" + safePathPart(kbID)
	}
	if issue := unresolvedRemoteDefaultAuthority(opts.RepoRoot); issue != "" {
		return blockedPlanRunWorkspace("prepare", issue, nil), nil
	}
	if isDefaultIntegrationRef(opts.RepoRoot, baseRef, opts.IntegrationRef) {
		return blockedPlanRunWorkspace("prepare", "integration ref must not be the default branch", nil), nil
	}
	limit, err := resolvePlanRunWorktreeLimit(opts.RepoRoot)
	if err != nil {
		return blockedPlanRunWorkspace("prepare", err.Error(), nil), nil
	}
	if live := activePlanRunWorktrees(opts, kbID); len(live) >= limit {
		return blockedPlanRunWorkspace("prepare", fmt.Sprintf(
			"plan-run worktree limit %d reached; queue behind or release a live run before preparing another (live: %s)",
			limit, strings.Join(live, ", ")), nil), nil
	}
	if opts.Worktree == "" {
		opts.Worktree = defaultPlanRunWorktreePath(opts.RepoRoot, kbID)
	}
	worktree, err := filepath.Abs(filepath.Clean(opts.Worktree))
	if err != nil {
		return planRunWorkspaceResult{}, err
	}
	opts.Worktree = worktree
	if samePath(opts.RepoRoot, worktree) {
		return blockedPlanRunWorkspace("prepare", "plan-run worktree must not be the source checkout", nil), nil
	}
	if pathExists(worktree) {
		return blockedPlanRunWorkspace("prepare", "plan-run worktree path already exists without a matching receipt", nil), nil
	}
	if gitOutput(opts.RepoRoot, "show-ref", "--verify", "refs/heads/"+opts.IntegrationRef) != "" {
		return blockedPlanRunWorkspace("prepare", "integration ref already exists without a matching receipt", nil), nil
	}
	delivery, err := resolveKBDeliveryPolicy(opts.RepoRoot)
	if err != nil {
		return blockedPlanRunWorkspace("prepare", err.Error(), nil), nil
	}
	dirtyPaths, err := planRunDirtyPaths(opts.RepoRoot)
	if err != nil {
		return blockedPlanRunWorkspace("prepare", err.Error(), nil), nil
	}
	sourceDirty := len(dirtyPaths) > 0
	if sourceDirty {
		relevant, err := relevantPlanRunDirtyPaths(opts, dirtyPaths)
		if err != nil {
			return blockedPlanRunWorkspace("prepare", err.Error(), nil), nil
		}
		if len(relevant) > 0 {
			return blockedPlanRunWorkspace(
				"prepare",
				"relevant dirty WIP requires an explicit checkpoint/patch or reviewed execution decision; preserved paths: "+strings.Join(relevant, ", "),
				nil,
			), nil
		}
	}
	if code, out := runGitCommand(opts.RepoRoot, "worktree", "add", "-b", opts.IntegrationRef, worktree, baseSHA); code != 0 {
		return blockedPlanRunWorkspace("prepare", strings.TrimSpace(out), nil), nil
	}
	now := opts.Now.Format(time.RFC3339Nano)
	receipt := planRunWorkspaceReceipt{
		SchemaVersion:        planRunWorkspaceSchemaVersion,
		KBID:                 kbID,
		RunID:                defaultPlanRunID(opts.RunID, kbID),
		ManifestPath:         opts.ManifestPath,
		OwnerToken:           opts.OwnerToken,
		SourceCheckout:       opts.RepoRoot,
		SourceDirty:          sourceDirty,
		Worktree:             worktree,
		WorkspaceOwner:       planRunWorkspaceOwnerKB,
		BaseRef:              baseRef,
		BaseSHA:              baseSHA,
		IntegrationRef:       opts.IntegrationRef,
		IntegrationHead:      gitOutput(worktree, "rev-parse", "HEAD"),
		CommitAuthorized:     true,
		CommitAuthorizedBy:   strings.TrimSpace(opts.CommitAuthorizedBy),
		CommitApprovalRef:    strings.TrimSpace(opts.CommitApprovalRef),
		CommitAuthorizedAt:   now,
		DeliveryMode:         delivery.Mode,
		DeliveryMerge:        delivery.Merge,
		DeliveryPolicySource: delivery.Source,
		Status:               "prepared",
		CleanupState:         "active",
		CreatedAt:            now,
		UpdatedAt:            now,
		Limitations: []string{
			"dirty source changes are preserved but excluded from the plan-run worktree",
			"plan-run ownership coordinates only worktrees sharing this Git common directory",
			"default-branch delivery requires a separate authorized delivery phase",
			"slice commits are authorized only for this local manifest-owned plan-run branch",
			"cleanup is non-force and requires clean integrated state",
		},
	}
	if err := savePlanRunWorkspaceReceipt(opts, receipt); err != nil {
		return planRunWorkspaceResult{}, err
	}
	return planRunWorkspaceResult{OK: true, Action: "prepare", Receipt: &receipt}, nil
}

// activePlanRunWorktrees lists KB-owned plan-run worktrees still live on disk.
// Harness session trees are deliberately excluded: KB neither creates nor
// removes them, so gating on them would fail closed on state KB cannot
// remediate.
func activePlanRunWorktrees(opts planRunWorkspaceOptions, excludeKBID string) []string {
	dir, err := planRunWorkspaceReceiptDir(opts)
	if err != nil {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	live := []string{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		content, readErr := os.ReadFile(filepath.Join(dir, entry.Name()))
		if readErr != nil {
			continue
		}
		var receipt planRunWorkspaceReceipt
		if json.Unmarshal(content, &receipt) != nil {
			continue
		}
		if receipt.KBID == excludeKBID ||
			receipt.WorkspaceOwner != planRunWorkspaceOwnerKB ||
			receipt.CleanupState != "active" ||
			!pathExists(receipt.Worktree) {
			continue
		}
		live = append(live, receipt.Worktree)
	}
	return sortedWorktreeStrings(live)
}

func resolvePlanRunWorktreeLimit(root string) (int, error) {
	limit := defaultMaxLinkedWorktrees
	path := filepath.Join(root, "docs", "context", "operations", "kb-routing.yaml")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return limit, nil
	}
	if err != nil {
		return 0, fmt.Errorf("read execution policy: %w", err)
	}
	inExecution := false
	for _, line := range strings.Split(string(content), "\n") {
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if indent == 0 {
			inExecution = trimmed == "execution:"
			continue
		}
		if !inExecution {
			continue
		}
		key, value, ok := splitYAMLKeyValue(trimmed)
		if !ok || key != "max_plan_run_worktrees" {
			continue
		}
		parsed, convErr := strconv.Atoi(cleanYAMLScalar(value))
		if convErr != nil || parsed < 1 {
			return 0, fmt.Errorf("execution.max_plan_run_worktrees must be a positive integer")
		}
		limit = parsed
	}
	return limit, nil
}

func resolveKBDeliveryPolicy(root string) (kbDeliveryPolicy, error) {
	// Reviewed work defaults to a pushed PR, not a stranded local branch.
	// Stopping at "local" is what leaves finished work invisible on disk.
	// Merge stays manual: PR-ready is automatic, PR acceptance never is.
	policy := kbDeliveryPolicy{
		Mode: "pr", Merge: "manual", PostMergeSync: false, Source: "default-absent",
	}
	path := filepath.Join(root, "docs", "context", "operations", "kb-routing.yaml")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return policy, nil
	}
	if err != nil {
		return kbDeliveryPolicy{}, fmt.Errorf("read delivery policy: %w", err)
	}
	policy.Source = "project"
	inDelivery := false
	for _, line := range strings.Split(string(content), "\n") {
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if indent == 0 {
			inDelivery = trimmed == "delivery:"
			continue
		}
		if !inDelivery {
			continue
		}
		key, value, ok := splitYAMLKeyValue(trimmed)
		if !ok {
			continue
		}
		value = cleanYAMLScalar(value)
		switch key {
		case "mode":
			policy.Mode = strings.ToLower(value)
		case "merge":
			policy.Merge = strings.ToLower(value)
		case "post_merge_sync":
			policy.PostMergeSync = parseBool(value)
		}
	}
	if policy.Mode != "local" && policy.Mode != "pr" && policy.Mode != "direct" {
		return kbDeliveryPolicy{}, fmt.Errorf("delivery mode must be local, pr, or direct")
	}
	if policy.Merge != "manual" && policy.Merge != "auto-after-checks" {
		return kbDeliveryPolicy{}, fmt.Errorf("delivery merge must be manual or auto-after-checks")
	}
	return policy, nil
}

func planRunDirtyPaths(root string) ([]string, error) {
	code, output := runGitCommand(root, "status", "--porcelain", "--untracked-files=all")
	if code != 0 {
		return nil, fmt.Errorf("read source dirty state: %s", strings.TrimSpace(output))
	}
	paths := []string{}
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 4 {
			continue
		}
		value := strings.TrimSpace(line[3:])
		if parts := strings.Split(value, " -> "); len(parts) == 2 {
			paths = append(paths, strings.Trim(parts[0], `"`), strings.Trim(parts[1], `"`))
			continue
		}
		paths = append(paths, strings.Trim(value, `"`))
	}
	return paths, nil
}

func relevantPlanRunDirtyPaths(opts planRunWorkspaceOptions, dirtyPaths []string) ([]string, error) {
	forecast, err := hydratePlanRunForecast(planRunLeaseCommandOptions{
		RepoRoot: opts.RepoRoot, ManifestPath: opts.ManifestPath,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve dirty-WIP forecast: %w", err)
	}
	claims, err := normalizePlanRunClaims(forecast.Files, forecast.Prefixes, forecast.Domains, forecast.Resources)
	if err != nil {
		return nil, err
	}
	relevant := []string{}
	for _, path := range dirtyPaths {
		normalized, err := normalizeLeaseClaimPath(path)
		if err != nil {
			return nil, err
		}
		observed := leaseClaim{Kind: "file", Value: normalized}
		for _, claim := range claims {
			if planRunClaimsConflict(claim, observed) {
				relevant = append(relevant, normalized)
				break
			}
		}
	}
	sort.Strings(relevant)
	return relevant, nil
}

func advancePlanRunWorkspace(opts planRunWorkspaceOptions, snapshot planRunWorkspaceReceipt) (planRunWorkspaceResult, error) {
	if issue := validatePlanRunAdvanceInputs(opts, snapshot); issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	proof, proofBytes, proofDigest, issue := loadAndValidatePlanRunProofReceipt(opts, snapshot)
	if issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	if issue := validatePlanRunAdvanceGit(opts, snapshot); issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	if issue := replayPlanRunProof(snapshot.Worktree, "slice proof", proof.SliceProof); issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	if proof.AggregateProof != nil {
		if issue := replayPlanRunProof(snapshot.Worktree, "aggregate proof", *proof.AggregateProof); issue != "" {
			return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
		}
	}

	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return planRunWorkspaceResult{}, err
	}
	result := planRunWorkspaceResult{Action: "advance"}
	err = withSliceLeaseStateLock(stateRoot, func() error {
		current, err := loadPlanRunWorkspaceReceipt(opts, snapshot.KBID)
		if err != nil {
			return err
		}
		if current.IntegrationHead != opts.ExpectedIntegrationHead {
			result = blockedPlanRunWorkspace("advance", "expected integration head compare-and-swap failed", &current)
			return nil
		}
		if issue := validatePlanRunAdvanceInputs(opts, current); issue != "" {
			result = blockedPlanRunWorkspace("advance", issue, &current)
			return nil
		}
		if issue := validatePlanRunAdvanceGit(opts, current); issue != "" {
			result = blockedPlanRunWorkspace("advance", issue, &current)
			return nil
		}
		if issue := validatePlanRunAdvanceClaims(opts, current, proof, stateRoot); issue != "" {
			result = blockedPlanRunWorkspace("advance", issue, &current)
			return nil
		}
		proofArchive, err := persistPlanRunProofEvidence(opts, current, proofBytes, proofDigest)
		if err != nil {
			return err
		}
		current.IntegrationHead = opts.CommitSHA
		current.LastSliceID = opts.SliceID
		current.LastSliceCommit = opts.CommitSHA
		current.LastProofReceipt = opts.ProofReceipt
		current.LastProofSHA256 = proofDigest
		current.LastProofArchive = proofArchive
		current.AcceptedProofs = append(current.AcceptedProofs, planRunAcceptedProof{
			SliceID: opts.SliceID, CommitSHA: opts.CommitSHA, Archive: proofArchive,
			SHA256: proofDigest, AcceptedAt: opts.Now.Format(time.RFC3339Nano),
		})
		current.UpdatedAt = opts.Now.Format(time.RFC3339Nano)
		if err := savePlanRunWorkspaceReceipt(opts, current); err != nil {
			return err
		}
		result = planRunWorkspaceResult{OK: true, Action: "advance", Receipt: &current}
		return nil
	})
	return result, err
}

func validatePlanRunAdvanceClaims(
	opts planRunWorkspaceOptions,
	receipt planRunWorkspaceReceipt,
	proof planRunProofReceipt,
	stateRoot string,
) string {
	actualWrites, issue := planRunCommitWrites(receipt.Worktree, opts.ExpectedIntegrationHead, opts.CommitSHA)
	if issue != "" {
		return issue
	}
	observedWrites, issue := normalizePlanRunObservedWrites(proof.ObservedWrites)
	if issue != "" {
		return issue
	}
	if strings.Join(actualWrites, "\n") != strings.Join(observedWrites, "\n") {
		return fmt.Sprintf("proof observed_writes do not exactly match commit diff: actual=%v observed=%v", actualWrites, observedWrites)
	}

	sliceState, err := loadSliceLeaseState(stateRoot)
	if err != nil {
		return "load active slice lease: " + err.Error()
	}
	_, slice, ok := findSliceLease(sliceState, sliceLeaseCommandOptions{RunID: opts.RunID, SliceID: opts.SliceID})
	if !ok || effectiveLeaseStatus(slice, opts.Now) != "active" {
		return "active slice lease is required before plan-run advance"
	}
	if slice.OwnerToken != opts.OwnerToken || !samePath(slice.Worktree, receipt.Worktree) ||
		slice.Branch != receipt.IntegrationRef {
		return "active slice lease does not match plan-run owner/worktree/ref"
	}

	planState, err := loadPlanRunLeaseState(stateRoot)
	if err != nil {
		return "load active plan-run lease: " + err.Error()
	}
	plan, ok := planState.Leases[opts.RunID]
	if !ok || effectivePlanRunLeaseStatus(plan, opts.Now) != "active" {
		return "active plan-run lease is required before plan-run advance"
	}
	if plan.OwnerToken != opts.OwnerToken {
		return "active plan-run lease owner does not match plan-run receipt"
	}
	if plan.Worktree != "" && (!samePath(plan.Worktree, receipt.Worktree) || plan.Branch != receipt.IntegrationRef) {
		return "active plan-run lease does not match manifest-owned worktree/ref"
	}
	for _, path := range actualWrites {
		claim := leaseClaim{Kind: "file", Value: path}
		if !planRunClaimCovers(slice.Claims, claim) {
			return "commit write is outside active slice lease: " + path
		}
		if !planRunClaimCovers(plan.Claims, claim) {
			return "commit write is outside active plan-run lease: " + path
		}
	}
	return ""
}

func planRunCommitWrites(worktree, fromSHA, toSHA string) ([]string, string) {
	code, output := runGitCommand(worktree, "diff", "--name-only", "-z", "--diff-filter=ACDMRTUXB", fromSHA, toSHA, "--")
	if code != 0 {
		return nil, "read plan-run commit diff: " + strings.TrimSpace(output)
	}
	values := []string{}
	for _, value := range strings.Split(output, "\x00") {
		if value == "" {
			continue
		}
		normalized, err := normalizeLeaseClaimPath(value)
		if err != nil {
			return nil, "normalize commit write: " + err.Error()
		}
		values = append(values, normalized)
	}
	sort.Strings(values)
	if len(values) == 0 {
		return nil, "slice commit diff has no writes"
	}
	return values, ""
}

func normalizePlanRunObservedWrites(paths []string) ([]string, string) {
	seen := map[string]bool{}
	values := []string{}
	for _, value := range paths {
		normalized, err := normalizeLeaseClaimPath(value)
		if err != nil {
			return nil, "normalize proof observed write: " + err.Error()
		}
		if seen[normalized] {
			return nil, "proof observed_writes contains duplicate path: " + normalized
		}
		seen[normalized] = true
		values = append(values, normalized)
	}
	sort.Strings(values)
	return values, ""
}

func validatePlanRunAdvanceInputs(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) string {
	if issue := validatePlanRunOwnerAndIdentity(opts, receipt); issue != "" {
		return issue
	}
	if !receipt.CommitAuthorized {
		return "plan-run receipt does not record explicit local commit authorization"
	}
	if strings.TrimSpace(opts.RunID) == "" || opts.RunID != effectivePlanRunID(receipt) {
		return "run lineage does not match plan-run receipt"
	}
	if strings.TrimSpace(opts.SliceID) == "" {
		return "slice-id is required"
	}
	if strings.TrimSpace(opts.ExpectedIntegrationHead) == "" || opts.ExpectedIntegrationHead != receipt.IntegrationHead {
		return "expected integration head does not match plan-run receipt"
	}
	if strings.TrimSpace(opts.CommitSHA) == "" {
		return "commit-sha is required"
	}
	if strings.TrimSpace(opts.Worktree) == "" || !samePath(opts.Worktree, receipt.Worktree) {
		return "exact plan-run worktree is required"
	}
	if strings.TrimSpace(opts.IntegrationRef) == "" || opts.IntegrationRef != receipt.IntegrationRef {
		return "exact integration ref is required"
	}
	if strings.TrimSpace(opts.ProofReceipt) == "" {
		return "proof receipt is required"
	}
	return ""
}

func validatePlanRunAdvanceGit(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) string {
	if !samePath(opts.Worktree, receipt.Worktree) {
		return "worktree does not match plan-run receipt"
	}
	if branch := gitOutput(receipt.Worktree, "branch", "--show-current"); branch != receipt.IntegrationRef {
		return fmt.Sprintf("plan-run worktree branch mismatch: got %s want %s", branch, receipt.IntegrationRef)
	}
	currentHead := gitOutput(receipt.Worktree, "rev-parse", "HEAD")
	if currentHead != opts.CommitSHA {
		return fmt.Sprintf("current head does not match slice commit: got %s want %s", currentHead, opts.CommitSHA)
	}
	refHead := gitOutput(receipt.Worktree, "rev-parse", "refs/heads/"+receipt.IntegrationRef)
	if refHead != opts.CommitSHA {
		return "integration ref does not point at slice commit"
	}
	if opts.CommitSHA == opts.ExpectedIntegrationHead {
		return "slice commit must advance integration head"
	}
	if code, _ := runGitCommand(receipt.Worktree, "merge-base", "--is-ancestor", opts.ExpectedIntegrationHead, opts.CommitSHA); code != 0 {
		return "slice commit is not a descendant of expected integration head"
	}
	if strings.TrimSpace(gitOutput(receipt.Worktree, "status", "--porcelain")) != "" {
		return "plan-run worktree is dirty"
	}
	return ""
}

func loadAndValidatePlanRunProofReceipt(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) (planRunProofReceipt, []byte, string, string) {
	content, err := os.ReadFile(opts.ProofReceipt)
	if err != nil {
		return planRunProofReceipt{}, nil, "", fmt.Sprintf("read proof receipt: %v", err)
	}
	digest := fmt.Sprintf("%x", sha256.Sum256(content))
	var proof planRunProofReceipt
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proof); err != nil {
		return planRunProofReceipt{}, nil, "", fmt.Sprintf("parse proof receipt: %v", err)
	}
	if proof.SchemaVersion != 1 {
		return planRunProofReceipt{}, nil, "", fmt.Sprintf("unsupported proof receipt schema_version %d", proof.SchemaVersion)
	}
	if proof.KBID != receipt.KBID || proof.RunID != opts.RunID || proof.SliceID != opts.SliceID {
		return planRunProofReceipt{}, nil, "", "proof receipt owner/run/slice lineage mismatch"
	}
	if proof.CommitSHA != opts.CommitSHA {
		return planRunProofReceipt{}, nil, "", "proof receipt commit does not match slice commit"
	}
	if len(proof.ObservedWrites) == 0 {
		return planRunProofReceipt{}, nil, "", "proof receipt requires observed writes"
	}
	if issue := validatePlanRunProofCommand(proof.SliceProof); issue != "" {
		return planRunProofReceipt{}, nil, "", "slice proof " + issue
	}
	if proof.AggregateProof != nil {
		if issue := validatePlanRunProofCommand(*proof.AggregateProof); issue != "" {
			return planRunProofReceipt{}, nil, "", "aggregate proof " + issue
		}
	}
	return proof, content, digest, ""
}

func persistPlanRunProofEvidence(
	opts planRunWorkspaceOptions,
	receipt planRunWorkspaceReceipt,
	content []byte,
	digest string,
) (string, error) {
	receiptPath, err := planRunWorkspaceReceiptPath(opts, receipt.KBID)
	if err != nil {
		return "", err
	}
	root := filepath.Join(filepath.Dir(filepath.Dir(receiptPath)), "plan-run-proofs", safePathPart(receipt.KBID))
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	name := safePathPart(opts.SliceID) + "-" + safePathPart(opts.CommitSHA) + ".json"
	path := filepath.Join(root, name)
	if existing, err := os.ReadFile(path); err == nil {
		if fmt.Sprintf("%x", sha256.Sum256(existing)) != digest {
			return "", fmt.Errorf("immutable proof archive collision at %s", path)
		}
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	temp := path + ".tmp"
	if err := os.WriteFile(temp, content, 0o644); err != nil {
		return "", err
	}
	if err := os.Rename(temp, path); err != nil {
		_ = os.Remove(temp)
		return "", err
	}
	return path, nil
}

func validateReleasedPlanRunSlices(
	manifestPath string,
	receipt planRunWorkspaceReceipt,
	state sliceLeaseState,
	now time.Time,
) string {
	slices, err := parseManifestSlices(manifestPath)
	if err != nil {
		return err.Error()
	}
	for _, manifestSlice := range slices {
		_, lease, ok := findSliceLease(state, sliceLeaseCommandOptions{
			RunID: receipt.RunID, SliceID: manifestSlice.ID,
		})
		if manifestSlice.Status == "done" && !ok {
			return fmt.Sprintf("done slice %s has no durable lease release evidence", manifestSlice.ID)
		}
		if !ok {
			continue
		}
		if lease.OwnerToken != receipt.OwnerToken {
			return fmt.Sprintf("slice %s lease owner does not match plan-run owner", manifestSlice.ID)
		}
		if lease.Status != "released" {
			status := effectiveLeaseStatus(lease, now)
			return fmt.Sprintf("slice %s lease must be explicitly released before completion; status=%s", manifestSlice.ID, status)
		}
	}
	return ""
}

func validateAcceptedPlanRunProofs(
	opts planRunWorkspaceOptions,
	receipt planRunWorkspaceReceipt,
	manifestPath string,
) string {
	if len(receipt.AcceptedProofs) == 0 {
		return "accepted proof ledger is required before plan-run completion"
	}
	receiptPath, err := planRunWorkspaceReceiptPath(opts, receipt.KBID)
	if err != nil {
		return err.Error()
	}
	expectedRoot := filepath.Join(
		filepath.Dir(filepath.Dir(receiptPath)),
		"plan-run-proofs",
		safePathPart(receipt.KBID),
	)
	acceptedSlices := map[string]bool{}
	finalAggregateSeen := false
	for _, accepted := range receipt.AcceptedProofs {
		archive, err := filepath.Abs(filepath.Clean(accepted.Archive))
		if err != nil {
			return fmt.Sprintf("resolve accepted proof archive for %s: %v", accepted.SliceID, err)
		}
		relative, err := filepath.Rel(expectedRoot, archive)
		if err != nil || filepath.IsAbs(relative) || relative == ".." ||
			strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Sprintf("accepted proof archive for %s is outside the plan-run proof root", accepted.SliceID)
		}
		content, err := os.ReadFile(archive)
		if err != nil {
			return fmt.Sprintf("accepted proof archive for %s is unavailable: %v", accepted.SliceID, err)
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(content))
		if digest != accepted.SHA256 {
			return fmt.Sprintf("accepted proof archive digest mismatch for %s", accepted.SliceID)
		}
		var proof planRunProofReceipt
		if err := json.Unmarshal(content, &proof); err != nil {
			return fmt.Sprintf("accepted proof archive for %s is invalid: %v", accepted.SliceID, err)
		}
		if proof.KBID != receipt.KBID || proof.RunID != receipt.RunID ||
			proof.SliceID != accepted.SliceID || proof.CommitSHA != accepted.CommitSHA {
			return fmt.Sprintf("accepted proof archive lineage mismatch for %s", accepted.SliceID)
		}
		finalAggregateSeen = proof.AggregateProof != nil
		acceptedSlices[accepted.SliceID] = true
	}
	slices, err := parseManifestSlices(manifestPath)
	if err != nil {
		return err.Error()
	}
	for _, slice := range slices {
		if slice.Status == "done" && !acceptedSlices[slice.ID] {
			return fmt.Sprintf("done slice %s has no accepted immutable proof", slice.ID)
		}
	}
	last := receipt.AcceptedProofs[len(receipt.AcceptedProofs)-1]
	if last.SliceID != receipt.LastSliceID || last.CommitSHA != receipt.LastSliceCommit ||
		last.Archive != receipt.LastProofArchive || last.SHA256 != receipt.LastProofSHA256 {
		return "last accepted proof summary does not match the immutable proof ledger"
	}
	if !finalAggregateSeen {
		return "final accepted proof requires aggregate proof"
	}
	return ""
}

func validatePlanRunProofCommand(proof planRunProofCommand) string {
	if len(proof.Args) == 0 || strings.TrimSpace(proof.Args[0]) == "" {
		return "command args are required"
	}
	if proof.Expect < 0 || proof.Expect > 255 {
		return "expected exit code must be between 0 and 255"
	}
	return ""
}

func replayPlanRunProof(worktree, label string, proof planRunProofCommand) string {
	command := exec.Command(proof.Args[0], proof.Args[1:]...)
	command.Dir = worktree
	output, err := command.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			exitCode = exit.ExitCode()
		} else {
			return fmt.Sprintf("%s could not run: %v", label, err)
		}
	}
	if exitCode != proof.Expect {
		return fmt.Sprintf("%s failed: exit=%d want=%d output=%s", label, exitCode, proof.Expect, strings.TrimSpace(string(output)))
	}
	if proof.ExpectOutput != "" || len(output) == 0 {
		if strings.TrimSpace(string(output)) != strings.TrimSpace(proof.ExpectOutput) {
			return fmt.Sprintf("%s output mismatch", label)
		}
	}
	return ""
}

func defaultPlanRunID(runID, kbID string) string {
	if strings.TrimSpace(runID) != "" {
		return strings.TrimSpace(runID)
	}
	return kbID
}

func effectivePlanRunID(receipt planRunWorkspaceReceipt) string {
	return defaultPlanRunID(receipt.RunID, receipt.KBID)
}

func releasePlanRunWorkspace(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) (planRunWorkspaceResult, error) {
	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return planRunWorkspaceResult{}, err
	}
	result := planRunWorkspaceResult{Action: "release"}
	err = withSliceLeaseStateLock(stateRoot, func() error {
		current, err := loadPlanRunWorkspaceReceipt(opts, receipt.KBID)
		if err != nil {
			return err
		}
		if issue := validatePlanRunOwnerAndIdentity(opts, current); issue != "" {
			result = blockedPlanRunWorkspace("release", issue, &current)
			return nil
		}
		if current.Status != "completed" && current.Status != "delivered" {
			result = blockedPlanRunWorkspace("release", "active incomplete plan-run workspace cannot be released", &current)
			return nil
		}
		if issue := validatePlanRunHeadCAS(current); issue != "" {
			result = blockedPlanRunWorkspace("release", issue, &current)
			return nil
		}
		if planRunWorkspaceOwner(current) == planRunWorkspaceOwnerHarness {
			// The coding harness created this worktree and owns its teardown.
			// Releasing returns ownership; it never removes the checkout the
			// session is still using.
			current.CleanupState = "harness-owned"
		} else {
			if code, out := runGitCommand(opts.RepoRoot, "worktree", "remove", current.Worktree); code != 0 {
				result = blockedPlanRunWorkspace("release", strings.TrimSpace(out), &current)
				return nil
			}
			current.CleanupState = "released"
		}
		current.Status = "released"
		current.ReleasedAt = opts.Now.Format(time.RFC3339Nano)
		current.UpdatedAt = current.ReleasedAt
		if err := savePlanRunWorkspaceReceipt(opts, current); err != nil {
			return err
		}
		result = planRunWorkspaceResult{OK: true, Action: "release", Receipt: &current}
		return nil
	})
	return result, err
}

func validatePlanRunHeadCAS(receipt planRunWorkspaceReceipt) string {
	if branch := gitOutput(receipt.Worktree, "branch", "--show-current"); branch != receipt.IntegrationRef {
		return fmt.Sprintf("plan-run worktree branch mismatch: got %s want %s", branch, receipt.IntegrationRef)
	}
	if head := gitOutput(receipt.Worktree, "rev-parse", "HEAD"); head != receipt.IntegrationHead {
		return fmt.Sprintf("plan-run worktree head moved after acceptance: got %s want %s", head, receipt.IntegrationHead)
	}
	if ref := gitOutput(receipt.Worktree, "rev-parse", "refs/heads/"+receipt.IntegrationRef); ref != receipt.IntegrationHead {
		return "plan-run integration ref moved after acceptance"
	}
	if strings.TrimSpace(gitOutput(receipt.Worktree, "status", "--porcelain")) != "" {
		return "plan-run worktree is dirty; non-force lifecycle transition refused"
	}
	return ""
}

func planRunManifestPathInWorktree(receipt planRunWorkspaceReceipt) (string, error) {
	relative, err := filepath.Rel(receipt.SourceCheckout, receipt.ManifestPath)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("manifest path cannot be mapped into the plan-run worktree")
	}
	return filepath.Join(receipt.Worktree, relative), nil
}

func validateTerminalPlanRunManifest(path string) string {
	frontmatter, err := loadManifestFrontmatter(path)
	if err != nil {
		return err.Error()
	}
	status := ""
	for _, line := range strings.Split(frontmatter, "\n") {
		if countIndent(line) != 0 {
			continue
		}
		key, value, ok := splitYAMLKeyValue(strings.TrimSpace(line))
		if ok && key == "status" {
			status = value
			break
		}
	}
	if status != "completed" {
		return "manifest status must be completed before plan-run completion"
	}
	slices, err := parseManifestSlices(path)
	if err != nil {
		return err.Error()
	}
	gates, err := parseManifestGates(path)
	if err != nil {
		return err.Error()
	}
	gateStatus := map[string]string{}
	for _, gate := range gates {
		gateStatus[gate.GateID] = gate.Status
	}
	for _, slice := range slices {
		if slice.Status != "done" && slice.Status != "skipped" {
			return "all manifest slices must be terminal before plan-run completion"
		}
		if slice.Status == "done" && gateStatus["slice-"+slice.ID+"-to-done"] != "passed" {
			return "each done slice requires a passing terminal gate before plan-run completion"
		}
	}
	if gateStatus["work-to-complete"] != "passed" {
		return "work-to-complete gate must pass before plan-run completion"
	}
	return ""
}

func validatePlanRunOwnerAndIdentity(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) string {
	if receipt.OwnerToken != opts.OwnerToken {
		return "owner token does not match plan-run receipt"
	}
	if !samePath(receipt.ManifestPath, opts.ManifestPath) {
		return "manifest path does not match plan-run receipt"
	}
	if opts.BaseSHA != "" {
		requested := gitOutput(opts.RepoRoot, "rev-parse", opts.BaseSHA+"^{commit}")
		if requested == "" || requested != receipt.BaseSHA {
			return "immutable base does not match plan-run receipt"
		}
	}
	if opts.IntegrationRef != "" && opts.IntegrationRef != receipt.IntegrationRef {
		return "integration ref does not match plan-run receipt"
	}
	if opts.Worktree != "" && !samePath(opts.Worktree, receipt.Worktree) {
		return "worktree path does not match plan-run receipt"
	}
	return ""
}

func planRunManifestID(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read manifest: %w", err)
	}
	for _, line := range strings.Split(string(content), "\n") {
		key, value, ok := splitYAMLKeyValue(strings.TrimSpace(line))
		if ok && key == "kb_id" {
			value = strings.Trim(value, `"'`)
			if value != "" {
				return value, nil
			}
		}
	}
	return "", fmt.Errorf("manifest requires kb_id")
}

func planRunWorkspaceReceiptDir(opts planRunWorkspaceOptions) (string, error) {
	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(stateRoot), "plan-runs"), nil
}

func planRunWorkspaceReceiptPath(opts planRunWorkspaceOptions, kbID string) (string, error) {
	dir, err := planRunWorkspaceReceiptDir(opts)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, safePathPart(kbID)+".json"), nil
}

func loadPlanRunWorkspaceReceipt(opts planRunWorkspaceOptions, kbID string) (planRunWorkspaceReceipt, error) {
	path, err := planRunWorkspaceReceiptPath(opts, kbID)
	if err != nil {
		return planRunWorkspaceReceipt{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return planRunWorkspaceReceipt{}, fmt.Errorf("%w: plan-run receipt not found; slice-only receipts require explicit migration", err)
		}
		return planRunWorkspaceReceipt{}, err
	}
	var receipt planRunWorkspaceReceipt
	if err := json.Unmarshal(content, &receipt); err != nil {
		return planRunWorkspaceReceipt{}, err
	}
	if receipt.SchemaVersion != planRunWorkspaceSchemaVersion {
		return planRunWorkspaceReceipt{}, fmt.Errorf("unsupported plan-run receipt schema_version %d; migrate explicitly", receipt.SchemaVersion)
	}
	return receipt, nil
}

func savePlanRunWorkspaceReceipt(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) error {
	path, err := planRunWorkspaceReceiptPath(opts, receipt.KBID)
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

func isDefaultIntegrationRef(root, baseRef, integrationRef string) bool {
	normalized := strings.TrimPrefix(strings.TrimSpace(integrationRef), "refs/heads/")
	if normalized == strings.TrimPrefix(strings.TrimSpace(baseRef), "refs/heads/") {
		return true
	}
	return isResolvedDefaultBranch(root, normalized)
}

func isResolvedDefaultBranch(root, branch string) bool {
	normalized := strings.TrimPrefix(strings.TrimSpace(branch), "refs/heads/")
	if normalized == "" {
		return false
	}
	defaults := resolvedDefaultBranches(root)
	return defaults[normalized]
}

func currentDefaultBranch(root string) string {
	defaults := resolvedDefaultBranches(root)
	for _, preferred := range []string{"main", "master"} {
		if defaults[preferred] {
			return preferred
		}
	}
	names := make([]string, 0, len(defaults))
	for name := range defaults {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) > 0 {
		return names[0]
	}
	return ""
}

func resolvedDefaultBranches(root string) map[string]bool {
	defaults := map[string]bool{}
	for _, candidate := range []string{"main", "master"} {
		if gitOutput(root, "show-ref", "--verify", "refs/heads/"+candidate) != "" {
			defaults[candidate] = true
		}
	}
	remotes := strings.Fields(gitOutput(root, "remote"))
	for _, remote := range remotes {
		ref := gitOutput(root, "symbolic-ref", "--quiet", "--short", "refs/remotes/"+remote+"/HEAD")
		if ref != "" {
			defaults[strings.TrimPrefix(ref, remote+"/")] = true
		}
	}
	if len(defaults) == 0 {
		if current := gitOutput(root, "branch", "--show-current"); current != "" {
			defaults[current] = true
		}
	}
	return defaults
}

func unresolvedRemoteDefaultAuthority(root string) string {
	for _, remote := range strings.Fields(gitOutput(root, "remote")) {
		ref := gitOutput(root, "symbolic-ref", "--quiet", "--short", "refs/remotes/"+remote+"/HEAD")
		if ref == "" {
			return fmt.Sprintf("remote default branch authority is unresolved for %s; fetch and set %s/HEAD before plan-worktree preparation", remote, remote)
		}
	}
	return ""
}

func resolvePlanRunBaseRef(root, baseSHA string) string {
	headSHA := gitOutput(root, "rev-parse", "HEAD")
	if headSHA == baseSHA {
		if current := gitOutput(root, "branch", "--show-current"); current != "" {
			return current
		}
	}
	refs := strings.Fields(gitOutput(root, "for-each-ref", "--format=%(refname:short)", "--points-at", baseSHA, "refs/heads"))
	if len(refs) > 0 {
		return refs[0]
	}
	return "detached:" + baseSHA
}

func defaultPlanRunWorktreePath(root, kbID string) string {
	return filepath.Join(filepath.Dir(root), filepath.Base(root)+"-"+safePathPart(kbID))
}

func blockedPlanRunWorkspace(action, issue string, receipt *planRunWorkspaceReceipt) planRunWorkspaceResult {
	if strings.TrimSpace(issue) == "" {
		issue = "blocked"
	}
	return planRunWorkspaceResult{OK: false, Action: action, Issue: issue, Receipt: receipt}
}
