package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const planRunWorkspaceSchemaVersion = 1

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
	RepoRoot                string
	Now                     time.Time
}

type planRunWorkspaceReceipt struct {
	SchemaVersion    int      `json:"schema_version"`
	KBID             string   `json:"kb_id"`
	RunID            string   `json:"run_id,omitempty"`
	ManifestPath     string   `json:"manifest_path"`
	OwnerToken       string   `json:"owner_token"`
	SourceCheckout   string   `json:"source_checkout"`
	SourceDirty      bool     `json:"source_dirty"`
	Worktree         string   `json:"worktree"`
	BaseRef          string   `json:"base_ref"`
	BaseSHA          string   `json:"base_sha"`
	IntegrationRef   string   `json:"integration_ref"`
	IntegrationHead  string   `json:"integration_head"`
	LastSliceID      string   `json:"last_slice_id,omitempty"`
	LastSliceCommit  string   `json:"last_slice_commit,omitempty"`
	LastProofReceipt string   `json:"last_proof_receipt,omitempty"`
	Status           string   `json:"status"`
	CleanupState     string   `json:"cleanup_state"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
	ReleasedAt       string   `json:"released_at,omitempty"`
	Limitations      []string `json:"limitations"`
}

type planRunWorkspaceResult struct {
	OK      bool                     `json:"ok"`
	Action  string                   `json:"action"`
	Issue   string                   `json:"issue,omitempty"`
	Receipt *planRunWorkspaceReceipt `json:"receipt,omitempty"`
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
		RepoRoot:                root,
		Now:                     time.Now().UTC(),
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
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

func executePlanRunWorkspace(opts planRunWorkspaceOptions) (planRunWorkspaceResult, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action == "" {
		return planRunWorkspaceResult{}, fmt.Errorf("plan-worktree requires --action")
	}
	if action != "prepare" && action != "status" && action != "advance" && action != "release" {
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
		return planRunWorkspaceResult{OK: true, Action: action, Receipt: &existing}, nil
	}
	if !os.IsNotExist(loadErr) {
		return blockedPlanRunWorkspace(action, loadErr.Error(), nil), nil
	}
	return preparePlanRunWorkspace(opts, kbID)
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
	if isDefaultIntegrationRef(opts.RepoRoot, baseRef, opts.IntegrationRef) {
		return blockedPlanRunWorkspace("prepare", "integration ref must not be the default branch", nil), nil
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
	sourceDirty := strings.TrimSpace(gitOutput(opts.RepoRoot, "status", "--porcelain")) != ""
	if code, out := runGitCommand(opts.RepoRoot, "worktree", "add", "-b", opts.IntegrationRef, worktree, baseSHA); code != 0 {
		return blockedPlanRunWorkspace("prepare", strings.TrimSpace(out), nil), nil
	}
	now := opts.Now.Format(time.RFC3339Nano)
	receipt := planRunWorkspaceReceipt{
		SchemaVersion:   planRunWorkspaceSchemaVersion,
		KBID:            kbID,
		RunID:           defaultPlanRunID(opts.RunID, kbID),
		ManifestPath:    opts.ManifestPath,
		OwnerToken:      opts.OwnerToken,
		SourceCheckout:  opts.RepoRoot,
		SourceDirty:     sourceDirty,
		Worktree:        worktree,
		BaseRef:         baseRef,
		BaseSHA:         baseSHA,
		IntegrationRef:  opts.IntegrationRef,
		IntegrationHead: gitOutput(worktree, "rev-parse", "HEAD"),
		Status:          "prepared",
		CleanupState:    "active",
		CreatedAt:       now,
		UpdatedAt:       now,
		Limitations: []string{
			"dirty source changes are preserved but excluded from the plan-run worktree",
			"plan-run ownership coordinates only worktrees sharing this Git common directory",
			"default-branch delivery requires a separate authorized delivery phase",
			"cleanup is non-force and requires clean integrated state",
		},
	}
	if err := savePlanRunWorkspaceReceipt(opts, receipt); err != nil {
		return planRunWorkspaceResult{}, err
	}
	return planRunWorkspaceResult{OK: true, Action: "prepare", Receipt: &receipt}, nil
}

func advancePlanRunWorkspace(opts planRunWorkspaceOptions, snapshot planRunWorkspaceReceipt) (planRunWorkspaceResult, error) {
	if issue := validatePlanRunAdvanceInputs(opts, snapshot); issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	proof, issue := loadAndValidatePlanRunProofReceipt(opts, snapshot)
	if issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	if issue := validatePlanRunAdvanceGit(opts, snapshot); issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	if issue := replayPlanRunProof(snapshot.Worktree, "slice proof", proof.SliceProof); issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
	}
	if proof.AggregateProof == nil {
		return blockedPlanRunWorkspace("advance", "aggregate proof is required", &snapshot), nil
	}
	if issue := replayPlanRunProof(snapshot.Worktree, "aggregate proof", *proof.AggregateProof); issue != "" {
		return blockedPlanRunWorkspace("advance", issue, &snapshot), nil
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
		current.IntegrationHead = opts.CommitSHA
		current.LastSliceID = opts.SliceID
		current.LastSliceCommit = opts.CommitSHA
		current.LastProofReceipt = opts.ProofReceipt
		current.UpdatedAt = opts.Now.Format(time.RFC3339Nano)
		if err := savePlanRunWorkspaceReceipt(opts, current); err != nil {
			return err
		}
		result = planRunWorkspaceResult{OK: true, Action: "advance", Receipt: &current}
		return nil
	})
	return result, err
}

func validatePlanRunAdvanceInputs(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) string {
	if issue := validatePlanRunOwnerAndIdentity(opts, receipt); issue != "" {
		return issue
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

func loadAndValidatePlanRunProofReceipt(opts planRunWorkspaceOptions, receipt planRunWorkspaceReceipt) (planRunProofReceipt, string) {
	content, err := os.ReadFile(opts.ProofReceipt)
	if err != nil {
		return planRunProofReceipt{}, fmt.Sprintf("read proof receipt: %v", err)
	}
	var proof planRunProofReceipt
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&proof); err != nil {
		return planRunProofReceipt{}, fmt.Sprintf("parse proof receipt: %v", err)
	}
	if proof.SchemaVersion != 1 {
		return planRunProofReceipt{}, fmt.Sprintf("unsupported proof receipt schema_version %d", proof.SchemaVersion)
	}
	if proof.KBID != receipt.KBID || proof.RunID != opts.RunID || proof.SliceID != opts.SliceID {
		return planRunProofReceipt{}, "proof receipt owner/run/slice lineage mismatch"
	}
	if proof.CommitSHA != opts.CommitSHA {
		return planRunProofReceipt{}, "proof receipt commit does not match slice commit"
	}
	if len(proof.ObservedWrites) == 0 {
		return planRunProofReceipt{}, "proof receipt requires observed writes"
	}
	if issue := validatePlanRunProofCommand(proof.SliceProof); issue != "" {
		return planRunProofReceipt{}, "slice proof " + issue
	}
	if proof.AggregateProof != nil {
		if issue := validatePlanRunProofCommand(*proof.AggregateProof); issue != "" {
			return planRunProofReceipt{}, "aggregate proof " + issue
		}
	}
	return proof, ""
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
	if issue := validatePlanRunOwnerAndIdentity(opts, receipt); issue != "" {
		return blockedPlanRunWorkspace("release", issue, &receipt), nil
	}
	if receipt.Status != "integrated" && receipt.Status != "delivered" {
		return blockedPlanRunWorkspace("release", "active unintegrated plan-run workspace cannot be released", &receipt), nil
	}
	if strings.TrimSpace(gitOutput(receipt.Worktree, "status", "--porcelain")) != "" {
		return blockedPlanRunWorkspace("release", "plan-run worktree is dirty; non-force release refused", &receipt), nil
	}
	if code, out := runGitCommand(opts.RepoRoot, "worktree", "remove", receipt.Worktree); code != 0 {
		return blockedPlanRunWorkspace("release", strings.TrimSpace(out), &receipt), nil
	}
	receipt.Status = "released"
	receipt.CleanupState = "released"
	receipt.ReleasedAt = opts.Now.Format(time.RFC3339Nano)
	receipt.UpdatedAt = receipt.ReleasedAt
	if err := savePlanRunWorkspaceReceipt(opts, receipt); err != nil {
		return planRunWorkspaceResult{}, err
	}
	return planRunWorkspaceResult{OK: true, Action: "release", Receipt: &receipt}, nil
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

func planRunWorkspaceReceiptPath(opts planRunWorkspaceOptions, kbID string) (string, error) {
	stateRoot, err := resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{RepoRoot: opts.RepoRoot})
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(stateRoot), "plan-runs", safePathPart(kbID)+".json"), nil
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
	remoteDefault := gitOutput(root, "symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD")
	remoteDefault = strings.TrimPrefix(remoteDefault, "origin/")
	return remoteDefault != "" && normalized == remoteDefault
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
