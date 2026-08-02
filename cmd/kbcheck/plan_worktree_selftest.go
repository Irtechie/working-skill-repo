package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type planWorktreeSelftestOptions struct {
	RealRepoRoot  string
	TempParent    string
	KeepArtifacts bool
}

type planWorktreeSelftestRun struct {
	RunID        string   `json:"run_id"`
	WorktreePath string   `json:"worktree_path"`
	Branch       string   `json:"branch"`
	SliceCommits []string `json:"slice_commits"`
}

type planWorktreeSelftestResult struct {
	ArtifactRoot                 string                    `json:"artifact_root"`
	SourceHeadBefore             string                    `json:"source_head_before"`
	SourceHeadAfter              string                    `json:"source_head_after"`
	SourceStatusBefore           string                    `json:"source_status_before"`
	SourceStatusAfter            string                    `json:"source_status_after"`
	Runs                         []planWorktreeSelftestRun `json:"runs"`
	CollisionOwnerEvidence       []string                  `json:"collision_owner_evidence"`
	DirtyBlocked                 bool                      `json:"dirty_blocked"`
	StaleHeadBlocked             bool                      `json:"stale_head_blocked"`
	WrongWorktreeBlocked         bool                      `json:"wrong_worktree_blocked"`
	DefaultPolicyStopBeforeMerge bool                      `json:"default_policy_stop_before_merge"`
	PRManualStopBeforeMerge      bool                      `json:"pr_manual_stop_before_merge"`
	RealRepoRejected             bool                      `json:"real_repo_rejected"`
	DisposableDefaultBranch      string                    `json:"disposable_default_branch"`
}

type planWorktreeSelftestRunSpec struct {
	RunID       string
	OwnerToken  string
	ManifestRel string
	Branch      string
	Worktree    string
	Files       []string
}

func runPlanWorktreeSelftest(root string, stdout, stderr io.Writer) int {
	result, err := executePlanWorktreeSelftest(planWorktreeSelftestOptions{RealRepoRoot: root})
	if err != nil {
		fmt.Fprintf(stderr, "plan-worktree lifecycle selftest failed: %v\n", err)
		if result.ArtifactRoot != "" {
			fmt.Fprintf(stderr, "failure artifacts preserved: %s\n", result.ArtifactRoot)
		}
		return 1
	}
	fmt.Fprintf(
		stdout,
		"plan-worktree lifecycle selftest: passed (runs=%d commits=%d collisions=%d source-unchanged=true delivery-stopped=true)\n",
		len(result.Runs),
		countPlanWorktreeSelftestCommits(result.Runs),
		len(result.CollisionOwnerEvidence),
	)
	return 0
}

func executePlanWorktreeSelftest(opts planWorktreeSelftestOptions) (result planWorktreeSelftestResult, returnedErr error) {
	realRoot, err := canonicalSelftestPath(opts.RealRepoRoot)
	if err != nil {
		return result, fmt.Errorf("resolve real repository root: %w", err)
	}
	parent := opts.TempParent
	if strings.TrimSpace(parent) == "" {
		parent, err = defaultPlanWorktreeSelftestTempParent(realRoot)
		if err != nil {
			return result, err
		}
	}
	parent, err = canonicalSelftestPath(parent)
	if err != nil {
		return result, fmt.Errorf("resolve selftest temp parent: %w", err)
	}
	artifactRoot, err := os.MkdirTemp(parent, "kb-plan-worktree-selftest-")
	if err != nil {
		return result, fmt.Errorf("create disposable selftest root: %w", err)
	}
	result.ArtifactRoot = artifactRoot
	if err := validatePlanWorktreeSelftestTarget(realRoot, artifactRoot, false); err != nil {
		return result, err
	}
	defer func() {
		if returnedErr != nil {
			_ = writePlanWorktreeSelftestArtifact(result, returnedErr)
			return
		}
		if !opts.KeepArtifacts {
			_ = os.RemoveAll(artifactRoot)
		}
	}()

	result.RealRepoRejected = validatePlanWorktreeSelftestTarget(realRoot, realRoot, true) != nil
	if !result.RealRepoRejected {
		return result, fmt.Errorf("real-repository firebreak accepted its own repository")
	}

	sourceRoot := filepath.Join(artifactRoot, "source")
	if err := initializePlanWorktreeSelftestRepo(sourceRoot); err != nil {
		return result, err
	}
	defaultBranch := gitOutput(sourceRoot, "branch", "--show-current")
	if defaultBranch == "" {
		return result, fmt.Errorf("disposable repository has no default branch")
	}
	result.DisposableDefaultBranch = defaultBranch

	if err := os.WriteFile(filepath.Join(sourceRoot, "user-owned.txt"), []byte("dirty user state\n"), 0o644); err != nil {
		return result, fmt.Errorf("create source dirt: %w", err)
	}
	result.SourceHeadBefore = gitOutput(sourceRoot, "rev-parse", "HEAD")
	result.SourceStatusBefore = gitOutput(sourceRoot, "status", "--porcelain", "--untracked-files=all")
	if result.SourceStatusBefore == "" {
		return result, fmt.Errorf("selftest setup did not create source dirt")
	}

	specs := []planWorktreeSelftestRunSpec{
		{
			RunID: "run-a", OwnerToken: "owner-a",
			ManifestRel: "docs/plans/kb-plan-run-a.md",
			Branch:      "codex/selftest-run-a",
			Worktree:    filepath.Join(artifactRoot, "worktrees", "run-a"),
			Files:       []string{"src/run-a-001.txt", "src/run-a-002.txt"},
		},
		{
			RunID: "run-b", OwnerToken: "owner-b",
			ManifestRel: "docs/plans/kb-plan-run-b.md",
			Branch:      "codex/selftest-run-b",
			Worktree:    filepath.Join(artifactRoot, "worktrees", "run-b"),
			Files:       []string{"src/run-b-001.txt", "src/run-b-002.txt"},
		},
	}

	workspaceReceipts := make([]*planRunWorkspaceReceipt, 0, len(specs))
	for _, spec := range specs {
		workspaceOpts := planRunWorkspaceOptions{
			Action:             "prepare",
			CommitAuthorized:   true,
			CommitAuthorizedBy: "selftest-user",
			CommitApprovalRef:  "selftest:" + spec.RunID,
			ManifestPath:       filepath.Join(sourceRoot, filepath.FromSlash(spec.ManifestRel)),
			OwnerToken:         spec.OwnerToken,
			BaseSHA:            result.SourceHeadBefore,
			Worktree:           spec.Worktree,
			IntegrationRef:     spec.Branch,
			RunID:              spec.RunID,
			RepoRoot:           sourceRoot,
			Now:                time.Now().UTC(),
		}
		prepared, err := executePlanRunWorkspace(workspaceOpts)
		if err != nil || !prepared.OK || prepared.Receipt == nil {
			return result, fmt.Errorf("prepare %s failed: result=%#v err=%v", spec.RunID, prepared, err)
		}
		workspaceReceipts = append(workspaceReceipts, prepared.Receipt)
	}

	leaseA, err := acquireSelftestPlanRunLease(specs[0], workspaceReceipts[0], "browser:4110", "")
	if err != nil {
		return result, err
	}
	if _, err := acquireSelftestPlanRunLease(specs[1], workspaceReceipts[1], "cache:run-b", leaseA.Lease.RepoIdentity); err != nil {
		return result, err
	}
	pathCollision, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action:       "acquire",
		StateRoot:    leaseA.StateRoot,
		RunID:        "collision-path",
		ManifestPath: "collision-path.md",
		OwnerToken:   "collision-owner",
		Files:        []string{specs[0].Files[0]},
		RepoID:       leaseA.Lease.RepoIdentity,
		Now:          time.Now().UTC(),
	})
	if err != nil || pathCollision.OK || len(pathCollision.Collisions) == 0 {
		return result, fmt.Errorf("path collision did not block with owner evidence: result=%#v err=%v", pathCollision, err)
	}
	result.CollisionOwnerEvidence = append(result.CollisionOwnerEvidence, formatSelftestCollisions(pathCollision.Collisions)...)

	resourceCollision, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action:       "acquire",
		StateRoot:    leaseA.StateRoot,
		RunID:        "collision-resource",
		ManifestPath: "collision-resource.md",
		OwnerToken:   "collision-owner",
		Resources:    []string{"browser:4110"},
		RepoID:       leaseA.Lease.RepoIdentity,
		Now:          time.Now().UTC(),
	})
	if err != nil || resourceCollision.OK || len(resourceCollision.Collisions) == 0 {
		return result, fmt.Errorf("resource collision did not block with owner evidence: result=%#v err=%v", resourceCollision, err)
	}
	result.CollisionOwnerEvidence = append(result.CollisionOwnerEvidence, formatSelftestCollisions(resourceCollision.Collisions)...)

	for index := range specs {
		runSummary, err := executeSelftestPlanRun(
			artifactRoot,
			sourceRoot,
			specs[index],
			workspaceReceipts[index],
			index == 0,
			&result,
		)
		if err != nil {
			return result, err
		}
		result.Runs = append(result.Runs, runSummary)
	}

	result.SourceHeadAfter = gitOutput(sourceRoot, "rev-parse", "HEAD")
	result.SourceStatusAfter = gitOutput(sourceRoot, "status", "--porcelain", "--untracked-files=all")
	if result.SourceHeadBefore != result.SourceHeadAfter || result.SourceStatusBefore != result.SourceStatusAfter {
		return result, fmt.Errorf("source default checkout changed during isolated plan runs")
	}
	if got, err := os.ReadFile(filepath.Join(sourceRoot, "user-owned.txt")); err != nil || string(got) != "dirty user state\n" {
		return result, fmt.Errorf("source dirty bytes changed: content=%q err=%v", string(got), err)
	}

	// Absent policy is PR/manual. It must still stop before the default branch:
	// PR-ready is automatic, merging never is.
	result.DefaultPolicyStopBeforeMerge = true
	for index, receipt := range workspaceReceipts {
		if receipt.DeliveryMode != "pr" ||
			receipt.DeliveryMerge != "manual" ||
			specs[index].Branch == defaultBranch ||
			gitOutput(sourceRoot, "rev-parse", "refs/heads/"+defaultBranch) != result.SourceHeadBefore {
			result.DefaultPolicyStopBeforeMerge = false
		}
	}
	if !result.DefaultPolicyStopBeforeMerge {
		return result, fmt.Errorf("absent-policy PR/manual delivery did not stop before default-branch merge")
	}

	prPolicyRoot := filepath.Join(artifactRoot, "pr-policy")
	if err := writeSelftestFile(
		prPolicyRoot,
		"docs/context/operations/kb-routing.yaml",
		"delivery:\n  mode: pr\n  merge: manual\n  post_merge_sync: false\n",
	); err != nil {
		return result, err
	}
	prPolicy, err := resolveKBDeliveryPolicy(prPolicyRoot)
	if err != nil {
		return result, fmt.Errorf("resolve PR/manual policy: %w", err)
	}
	result.PRManualStopBeforeMerge = prPolicy.Mode == "pr" &&
		prPolicy.Merge == "manual" &&
		specs[1].Branch != defaultBranch &&
		gitOutput(sourceRoot, "rev-parse", "refs/heads/"+defaultBranch) == result.SourceHeadBefore
	if !result.PRManualStopBeforeMerge {
		return result, fmt.Errorf("PR/manual delivery simulation did not stop before merge")
	}

	if err := writePlanWorktreeSelftestArtifact(result, nil); err != nil {
		return result, err
	}
	return result, nil
}

func defaultPlanWorktreeSelftestTempParent(realRoot string) (string, error) {
	candidates := []string{os.TempDir()}
	if cacheRoot, err := os.UserCacheDir(); err == nil {
		candidates = append(candidates, filepath.Join(cacheRoot, "kbcheck-selftests"))
	}
	for _, candidate := range candidates {
		probe := filepath.Join(candidate, "kb-plan-worktree-selftest-probe")
		if validatePlanWorktreeSelftestTarget(realRoot, probe, false) != nil {
			continue
		}
		if err := os.MkdirAll(candidate, 0o755); err != nil {
			continue
		}
		return candidate, nil
	}
	return "", fmt.Errorf("no disposable selftest temp parent is disjoint from the real repository")
}

func acquireSelftestPlanRunLease(spec planWorktreeSelftestRunSpec, receipt *planRunWorkspaceReceipt, resource, repoID string) (planRunLeaseResult, error) {
	result, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action:       "acquire",
		RunID:        spec.RunID,
		ManifestPath: filepath.Join(receipt.Worktree, filepath.FromSlash(spec.ManifestRel)),
		OwnerToken:   spec.OwnerToken,
		Resources:    []string{resource},
		RepoRoot:     receipt.Worktree,
		RepoID:       repoID,
		Now:          time.Now().UTC(),
	})
	if err != nil || !result.OK || result.Lease == nil {
		return result, fmt.Errorf("acquire %s lease failed: result=%#v err=%v", spec.RunID, result, err)
	}
	return result, nil
}

func executeSelftestPlanRun(
	artifactRoot string,
	sourceRoot string,
	spec planWorktreeSelftestRunSpec,
	receipt *planRunWorkspaceReceipt,
	exerciseRecovery bool,
	result *planWorktreeSelftestResult,
) (planWorktreeSelftestRun, error) {
	summary := planWorktreeSelftestRun{
		RunID: spec.RunID, WorktreePath: receipt.Worktree, Branch: receipt.IntegrationRef,
	}
	for label, wrongWorktree := range map[string]string{
		"source checkout": sourceRoot,
		"second worktree": filepath.Join(artifactRoot, "worktrees", spec.RunID+"-slice"),
	} {
		blocked, err := executeSliceLease(sliceLeaseCommandOptions{
			Action: "acquire", SliceID: spec.RunID + "-wrong-worktree", RunID: spec.RunID,
			OwnerToken: spec.OwnerToken, Files: []string{spec.Files[0]},
			Worktree: wrongWorktree, Branch: receipt.IntegrationRef,
			RepoRoot: receipt.Worktree, Now: time.Now().UTC(),
		})
		if err != nil || blocked.OK || !strings.Contains(blocked.Issue, "manifest-owned plan worktree") {
			return summary, fmt.Errorf("%s slice authority was not rejected: result=%#v err=%v", label, blocked, err)
		}
	}
	current := *receipt
	for index, relativePath := range spec.Files {
		sliceID := fmt.Sprintf("%s-slice-%03d", spec.RunID, index+1)
		sliceLeaseResult, err := executeSliceLease(sliceLeaseCommandOptions{
			Action: "acquire", SliceID: sliceID, RunID: spec.RunID, OwnerToken: spec.OwnerToken,
			Files: []string{relativePath}, Worktree: receipt.Worktree,
			Branch: receipt.IntegrationRef, RepoRoot: receipt.Worktree, Now: time.Now().UTC(),
		})
		if err != nil || !sliceLeaseResult.OK || sliceLeaseResult.Lease == nil {
			return summary, fmt.Errorf("acquire %s slice lease failed: result=%#v err=%v", sliceID, sliceLeaseResult, err)
		}
		if err := writeSelftestFile(receipt.Worktree, relativePath, sliceID+"\n"); err != nil {
			return summary, err
		}
		if err := selftestGit(receipt.Worktree, "add", "--", filepath.FromSlash(relativePath)); err != nil {
			return summary, err
		}
		if err := selftestGit(receipt.Worktree, "commit", "-m", "selftest "+sliceID); err != nil {
			return summary, err
		}
		commit := gitOutput(receipt.Worktree, "rev-parse", "HEAD")
		if commit == "" {
			return summary, fmt.Errorf("%s did not create a slice commit", sliceID)
		}
		proofPath, err := writeSelftestProofReceipt(
			filepath.Join(artifactRoot, "proofs"),
			current,
			spec.RunID,
			sliceID,
			commit,
			relativePath,
		)
		if err != nil {
			return summary, err
		}
		advance := planRunWorkspaceOptions{
			Action:                  "advance",
			ManifestPath:            filepath.Join(sourceRoot, filepath.FromSlash(spec.ManifestRel)),
			OwnerToken:              spec.OwnerToken,
			RunID:                   spec.RunID,
			SliceID:                 sliceID,
			ExpectedIntegrationHead: current.IntegrationHead,
			CommitSHA:               commit,
			ProofReceipt:            proofPath,
			Worktree:                receipt.Worktree,
			IntegrationRef:          receipt.IntegrationRef,
			RepoRoot:                sourceRoot,
			Now:                     time.Now().UTC(),
		}

		if exerciseRecovery && index == 0 {
			wrongWorktree := advance
			wrongWorktree.Worktree = filepath.Join(artifactRoot, "not-the-owned-worktree")
			blocked, err := executePlanRunWorkspace(wrongWorktree)
			if err != nil || blocked.OK || !strings.Contains(strings.ToLower(blocked.Issue), "worktree") ||
				blocked.Receipt == nil || blocked.Receipt.IntegrationHead != current.IntegrationHead {
				return summary, fmt.Errorf("wrong-worktree recovery failed closed: result=%#v err=%v", blocked, err)
			}
			result.WrongWorktreeBlocked = true

			stale := advance
			stale.ExpectedIntegrationHead = strings.Repeat("0", 40)
			blocked, err = executePlanRunWorkspace(stale)
			if err != nil || blocked.OK || !strings.Contains(strings.ToLower(blocked.Issue), "integration head") ||
				blocked.Receipt == nil || blocked.Receipt.IntegrationHead != current.IntegrationHead {
				return summary, fmt.Errorf("stale-head recovery failed closed: result=%#v err=%v", blocked, err)
			}
			result.StaleHeadBlocked = true

			dirtyPath := filepath.Join(receipt.Worktree, "selftest-dirty-preserved.txt")
			if err := os.WriteFile(dirtyPath, []byte("preserve until explicit recovery\n"), 0o644); err != nil {
				return summary, err
			}
			blocked, err = executePlanRunWorkspace(advance)
			if err != nil || blocked.OK || !strings.Contains(strings.ToLower(blocked.Issue), "dirty") ||
				blocked.Receipt == nil || blocked.Receipt.IntegrationHead != current.IntegrationHead ||
				!pathExists(dirtyPath) {
				return summary, fmt.Errorf("dirty recovery did not preserve state: result=%#v err=%v", blocked, err)
			}
			result.DirtyBlocked = true
			if err := os.Remove(dirtyPath); err != nil {
				return summary, fmt.Errorf("remove owned selftest dirt: %w", err)
			}
		}

		advanced, err := executePlanRunWorkspace(advance)
		if err != nil || !advanced.OK || advanced.Receipt == nil {
			return summary, fmt.Errorf("advance %s failed: result=%#v err=%v", sliceID, advanced, err)
		}
		if advanced.Receipt.IntegrationHead != commit {
			return summary, fmt.Errorf("advance %s recorded wrong integration head", sliceID)
		}
		current = *advanced.Receipt
		summary.SliceCommits = append(summary.SliceCommits, commit)
		released, err := executeSliceLease(sliceLeaseCommandOptions{
			Action: "release", SliceID: sliceID, RunID: spec.RunID, OwnerToken: spec.OwnerToken,
			Generation: sliceLeaseResult.Lease.Generation, RepoRoot: receipt.Worktree, Now: time.Now().UTC(),
		})
		if err != nil || !released.OK {
			return summary, fmt.Errorf("release %s slice lease failed: result=%#v err=%v", sliceID, released, err)
		}
	}
	return summary, nil
}

func writeSelftestProofReceipt(root string, receipt planRunWorkspaceReceipt, runID, sliceID, commit, observed string) (string, error) {
	proof := planRunProofReceipt{
		SchemaVersion: 1,
		KBID:          receipt.KBID,
		RunID:         runID,
		SliceID:       sliceID,
		CommitSHA:     commit,
		ObservedWrites: []string{
			filepath.ToSlash(observed),
		},
		SliceProof: planRunProofCommand{
			Args:   []string{"git", "rev-parse", "--verify", "HEAD"},
			Expect: 0,
		},
		AggregateProof: &planRunProofCommand{
			Args:         []string{"git", "status", "--porcelain"},
			Expect:       0,
			ExpectOutput: "",
		},
	}
	content, err := json.MarshalIndent(proof, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(root, sliceID+".json")
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func initializePlanWorktreeSelftestRepo(root string) error {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := selftestGit(root, "init"); err != nil {
		return err
	}
	if err := selftestGit(root, "config", "user.email", "kb-selftest@example.invalid"); err != nil {
		return err
	}
	if err := selftestGit(root, "config", "user.name", "KB Lifecycle Selftest"); err != nil {
		return err
	}
	files := map[string]string{
		"README.md":                     "# Disposable lifecycle selftest\n",
		"user-owned.txt":                "clean baseline\n",
		"src/run-a-001.txt":             "baseline\n",
		"src/run-a-002.txt":             "baseline\n",
		"src/run-b-001.txt":             "baseline\n",
		"src/run-b-002.txt":             "baseline\n",
		"docs/plans/run-a-slice-001.md": selftestSlicePlan("src/run-a-001.txt"),
		"docs/plans/run-a-slice-002.md": selftestSlicePlan("src/run-a-002.txt"),
		"docs/plans/run-b-slice-001.md": selftestSlicePlan("src/run-b-001.txt"),
		"docs/plans/run-b-slice-002.md": selftestSlicePlan("src/run-b-002.txt"),
		"docs/plans/kb-plan-run-a.md":   selftestManifest("kb-plan-run-a", "run-a", "browser:4110"),
		"docs/plans/kb-plan-run-b.md":   selftestManifest("kb-plan-run-b", "run-b", "cache:run-b"),
	}
	for relativePath, content := range files {
		if err := writeSelftestFile(root, relativePath, content); err != nil {
			return err
		}
	}
	if err := selftestGit(root, "add", "--", "."); err != nil {
		return err
	}
	if err := selftestGit(root, "commit", "-m", "selftest baseline"); err != nil {
		return err
	}
	if err := selftestGit(root, "branch", "-M", "main"); err != nil {
		return err
	}
	return nil
}

func selftestManifest(kbID, prefix, resource string) string {
	return fmt.Sprintf(`---
type: kb-manifest
kb_id: %s
workspace_isolation_contract:
  coordinator_owned_lifecycle: true
  plan_run_worktree_default: true
  internal_integration_target: plan-run-branch
  default_branch_delivery_owner: kb-complete
  allowed_modes: [shared-serial]
slices:
  - id: %s-slice-001
    path: docs/plans/%s-slice-001.md
    blockers: []
    status: pending
    conflict_domains: [file:src/%s-001.txt]
    shared_resources: [%s]
  - id: %s-slice-002
    path: docs/plans/%s-slice-002.md
    blockers: [%s-slice-001]
    status: pending
    conflict_domains: [file:src/%s-002.txt]
    shared_resources: [%s]
---
`, kbID, prefix, prefix, prefix, resource, prefix, prefix, prefix, prefix, resource)
}

func selftestSlicePlan(expected string) string {
	return fmt.Sprintf(`---
type: kb-slice
expected_files:
  - path: %s
    op: edit
verification: integration
blockers: []
status: pending
---
`, expected)
}

func selftestGit(root string, args ...string) error {
	if code, output := runGitCommand(root, args...); code != 0 {
		return fmt.Errorf("git %s failed in %s: %s", strings.Join(args, " "), root, strings.TrimSpace(output))
	}
	return nil
}

func writeSelftestFile(root, relativePath, content string) error {
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func formatSelftestCollisions(collisions []planRunCollision) []string {
	evidence := make([]string, 0, len(collisions))
	for _, collision := range collisions {
		evidence = append(
			evidence,
			fmt.Sprintf(
				"%s owns %s:%s (%s)",
				collision.RunID,
				collision.ExistingClaim.Kind,
				collision.ExistingClaim.Value,
				collision.Reason,
			),
		)
	}
	return evidence
}

func countPlanWorktreeSelftestCommits(runs []planWorktreeSelftestRun) int {
	total := 0
	for _, run := range runs {
		total += len(run.SliceCommits)
	}
	return total
}

func writePlanWorktreeSelftestArtifact(result planWorktreeSelftestResult, runErr error) error {
	payload := struct {
		Result planWorktreeSelftestResult `json:"result"`
		Error  string                     `json:"error,omitempty"`
	}{Result: result}
	if runErr != nil {
		payload.Error = runErr.Error()
	}
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(result.ArtifactRoot, "lifecycle-result.json")
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		return fmt.Errorf("write selftest artifact: %w", err)
	}
	return nil
}

func validatePlanWorktreeSelftestTarget(realRoot, candidate string, force bool) error {
	if force {
		return fmt.Errorf("plan-worktree selftest has no force mode")
	}
	real, err := canonicalSelftestPath(realRoot)
	if err != nil {
		return err
	}
	target, err := canonicalSelftestPath(candidate)
	if err != nil {
		return err
	}
	if samePath(real, target) || pathWithin(real, target) || pathWithin(target, real) {
		return fmt.Errorf("selftest target must be disposable and disjoint from real repository: %s", target)
	}
	return nil
}

func canonicalSelftestPath(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	existing := absolute
	suffix := []string{}
	for {
		resolved, err := filepath.EvalSymlinks(existing)
		if err == nil {
			for index := len(suffix) - 1; index >= 0; index-- {
				resolved = filepath.Join(resolved, suffix[index])
			}
			return filepath.Clean(resolved), nil
		}
		parent := filepath.Dir(existing)
		if samePath(parent, existing) {
			return absolute, nil
		}
		suffix = append(suffix, filepath.Base(existing))
		existing = parent
	}
}

func pathWithin(parent, child string) bool {
	parentAbs, err := filepath.Abs(filepath.Clean(parent))
	if err != nil {
		return false
	}
	childAbs, err := filepath.Abs(filepath.Clean(child))
	if err != nil || samePath(parentAbs, childAbs) {
		return false
	}
	relative, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return false
	}
	return relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator)) &&
		!filepath.IsAbs(relative)
}
