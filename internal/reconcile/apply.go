package reconcile

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const WorktreeSafetyContractVersion = "reconcile-predicates/v2"

type ApplyOptions struct {
	Bundle          PlanBundle
	Policy          Policy
	ExistingReceipt *ApplyReceipt
	CurrentWorktree string
	CurrentSession  string
	Now             time.Time
	LockTimeout     time.Duration
}

func WorktreeSafetyPredicates(policy Policy) []string {
	action, ok := policyForAction(policy, ActionWorktreeRetire)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(action.Mandatory))
	for _, predicate := range action.Mandatory {
		names = append(names, predicate.Name)
	}
	sort.Strings(names)
	return names
}

func ValidatePlanBundle(bundle PlanBundle, policy Policy, now time.Time, requireUnexpired bool) (string, error) {
	if bundle.Ledger.SchemaVersion != LedgerSchemaVersion ||
		bundle.Plan.SchemaVersion != PlanSchemaVersion {
		return "", fmt.Errorf("unsupported plan or ledger schema")
	}
	if err := ValidatePolicy(policy); err != nil {
		return "", err
	}
	if policy.PolicyVersion != bundle.Plan.PolicyVersion {
		return "", fmt.Errorf("plan policy version does not match the active policy")
	}
	ttl, _ := time.ParseDuration(policy.PlanTTL)
	if !bundle.Plan.ExpiresAt.Equal(bundle.Plan.GeneratedAt.Add(ttl)) {
		return "", fmt.Errorf("plan expiry does not match the active policy")
	}
	if !bundle.Ledger.Cutoff.Equal(bundle.Plan.Cutoff) {
		return "", fmt.Errorf("plan cutoff does not match its evidence ledger")
	}
	if requireUnexpired && (now.Before(bundle.Plan.GeneratedAt) || !now.Before(bundle.Plan.ExpiresAt)) {
		return "", fmt.Errorf("cutoff-bound plan is not currently valid")
	}
	ledgerFingerprint, err := FingerprintLedger(bundle.Ledger)
	if err != nil {
		return "", err
	}
	if ledgerFingerprint != bundle.Plan.LedgerFingerprint {
		return "", fmt.Errorf("plan evidence fingerprint does not match its ledger")
	}
	seen := map[string]bool{}
	for _, action := range bundle.Plan.Actions {
		if action.ID == "" || seen[action.ID] {
			return "", fmt.Errorf("plan has missing or duplicate action identity")
		}
		seen[action.ID] = true
		if !action.Cutoff.Equal(bundle.Plan.Cutoff) || !action.ExpiresAt.Equal(bundle.Plan.ExpiresAt) {
			return "", fmt.Errorf("action %s is not bound to the plan cutoff and expiry", action.ID)
		}
		repository, artifact, err := targetForAction(bundle.Ledger, action)
		if err != nil {
			return "", err
		}
		if repository.ID != action.RepositoryID || len(action.ArtifactIDs) != 1 ||
			artifact.ID != action.ArtifactIDs[0] {
			return "", fmt.Errorf("action %s target identity is ambiguous", action.ID)
		}
		actionPolicy, ok := policyForAction(policy, action.ActionClass)
		if !ok {
			return "", fmt.Errorf("action %s has no active policy", action.ID)
		}
		if err := validatePlannedPredicates(action, actionPolicy); err != nil {
			return "", fmt.Errorf("action %s: %w", action.ID, err)
		}
		if !samePredicates(action.Preconditions, artifact.Predicates) {
			return "", fmt.Errorf("action %s predicate evidence does not match its target", action.ID)
		}
		if !sameProofs(action.DedupProofs, artifact.DedupProofs) {
			return "", fmt.Errorf("action %s dedup evidence does not match its target", action.ID)
		}
		if action.MutationAllowed != localMutationAuthorized(action.ActionClass, action.Classification, policy) {
			return "", fmt.Errorf("action %s execution authorization does not match the active local mutation policy", action.ID)
		}
	}
	return FingerprintPlan(bundle.Plan)
}

func Apply(options ApplyOptions) (ApplyReceipt, error) {
	if options.Now.IsZero() {
		options.Now = time.Now().UTC()
	}
	if options.LockTimeout <= 0 {
		options.LockTimeout = 5 * time.Second
	}
	planFingerprint, err := ValidatePlanBundle(options.Bundle, options.Policy, options.Now.UTC(), true)
	if err != nil {
		return ApplyReceipt{}, err
	}
	receipt := NewApplyReceipt(options.Bundle, planFingerprint, options.Now)
	if options.ExistingReceipt != nil {
		receipt = *options.ExistingReceipt
		if err := ValidateApplyReceipt(receipt, options.Bundle, planFingerprint); err != nil {
			return ApplyReceipt{}, err
		}
	}
	byID := map[string]ActionReceipt{}
	for _, action := range receipt.Actions {
		byID[action.ActionID] = action
	}
	for _, action := range orderedActions(options.Bundle.Plan.Actions) {
		previous, exists := byID[action.ID]
		if exists && actionReceiptTerminal(previous) {
			continue
		}
		repository, artifact, _ := targetForAction(options.Bundle.Ledger, action)
		result := applyOne(options, repository, artifact, action)
		byID[action.ID] = result
	}
	receipt.Actions = receipt.Actions[:0]
	for _, action := range options.Bundle.Plan.Actions {
		if result, ok := byID[action.ID]; ok {
			receipt.Actions = append(receipt.Actions, result)
		}
	}
	normalizeReceipt(&receipt, false)
	return receipt, nil
}

func orderedActions(actions []PlannedAction) []PlannedAction {
	ordered := append([]PlannedAction(nil), actions...)
	rank := func(action string) int {
		switch action {
		case ActionWorktreeRetire:
			return 0
		case ActionLocalRefRetire:
			return 2
		default:
			return 1
		}
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		left, right := rank(ordered[i].ActionClass), rank(ordered[j].ActionClass)
		if left == right {
			return ordered[i].ID < ordered[j].ID
		}
		return left < right
	})
	return ordered
}

func Verify(options ApplyOptions, receipt ApplyReceipt) (Verification, ApplyReceipt, error) {
	if options.Now.IsZero() {
		options.Now = time.Now().UTC()
	}
	if options.LockTimeout <= 0 {
		options.LockTimeout = 5 * time.Second
	}
	planFingerprint, err := ValidatePlanBundle(options.Bundle, options.Policy, options.Now.UTC(), false)
	if err != nil {
		return Verification{}, receipt, err
	}
	if err := ValidateApplyReceipt(receipt, options.Bundle, planFingerprint); err != nil {
		return Verification{}, receipt, err
	}
	actions := make([]ActionReceipt, 0, len(receipt.Actions))
	for _, prior := range receipt.Actions {
		planned, ok := plannedActionByID(options.Bundle.Plan, prior.ActionID)
		if !ok {
			return Verification{}, receipt, fmt.Errorf("receipt action %s is absent from plan", prior.ActionID)
		}
		repository, artifact, _ := targetForAction(options.Bundle.Ledger, planned)
		actions = append(actions, verifyOne(options, repository, artifact, planned, prior))
	}
	receipt.Actions = actions
	receipt.VerifiedAt = options.Now.UTC()
	normalizeReceipt(&receipt, true)
	verification := Verification{
		SchemaVersion: ApplyReceiptSchemaVersion,
		Status:        receipt.Status, VerifiedAt: receipt.VerifiedAt,
		Actions: append([]ActionReceipt(nil), receipt.Actions...),
	}
	return verification, receipt, nil
}

func applyOne(options ApplyOptions, repository Repository, artifact Artifact, action PlannedAction) ActionReceipt {
	result := newActionReceipt(action, artifact, options.Now)
	actionPolicy, _ := policyForAction(options.Policy, action.ActionClass)
	if !action.MutationAllowed || !localMutationAuthorized(action.ActionClass, action.Classification, options.Policy) ||
		!actionPolicy.Allowed {
		result.Result = "unavailable"
		result.Issue = "action is outside the allowlisted local mutation surface; authoritative fenced adapter required"
		return result
	}
	lock, err := acquireRepositoryLock(repository, options.LockTimeout)
	if err != nil {
		result.Result = "contended"
		result.Issue = "compatible repository lock unavailable: " + err.Error()
		return result
	}
	defer lock.Close()
	switch action.ActionClass {
	case ActionWorktreeRetire:
		applyWorktree(options, repository, artifact, &result)
	case ActionLocalRefRetire:
		applyLocalRef(options, repository, artifact, action, &result)
	}
	return result
}

func applyWorktree(options ApplyOptions, repository Repository, artifact Artifact, result *ActionReceipt) {
	worktrees, err := listFreshWorktrees(repository.Root)
	if err != nil {
		blockAction(result, err)
		return
	}
	planned := plannedWorktree(repository, artifact.Path)
	target, found := freshWorktreeByPath(worktrees, artifact.Path)
	if !found {
		for _, worktree := range worktrees {
			if worktree.Branch == strings.TrimPrefix(artifact.Ref, "refs/heads/") {
				blockAction(result, fmt.Errorf("target worktree moved or was recreated under a different identity"))
				return
			}
		}
		empty, residualErr := exactEmptyResidual(artifact.Path, artifact.Path)
		if residualErr != nil {
			result.Result = "blocked"
			result.Issue = residualErr.Error()
			result.PhysicalCleanupState = StateBlocked
			return
		}
		if empty {
			if planned.GitDir == "" || pathExistsLocal(planned.GitDir) {
				blockAction(result, fmt.Errorf("partial residual Git-admin identity is not proven retired"))
				return
			}
			branch := strings.TrimPrefix(artifact.Ref, "refs/heads/")
			ref, _ := gitOutput(repository.Root, "rev-parse", "refs/heads/"+branch+"^{commit}")
			if ref != artifact.SHA {
				blockAction(result, fmt.Errorf("local branch identity changed; partial reconciliation refused"))
				return
			}
			if err := os.Remove(artifact.Path); err != nil {
				blockAction(result, fmt.Errorf("remove exact empty residual: %w", err))
				return
			}
			result.Result = "applied"
		} else {
			result.Result = "already-applied"
		}
		result.PhysicalCleanupState = StateVerifiedRetired
		result.ObservedAfter = "worktree registration and target path absent"
		return
	}
	result.ObservedBefore = target.Branch + "@" + target.Head
	if err := validateFreshWorktree(options, repository, artifact, planned, target, worktrees); err != nil {
		blockAction(result, err)
		return
	}
	delivery, err := freshDeliveryState(repository.Root, repository, target.Branch, target.Head, false)
	if err != nil {
		blockAction(result, err)
		return
	}
	result.DeliveryState = delivery
	if err := removeNonForceWorktree(repository.Root, target.Path); err != nil {
		blockAction(result, err)
		return
	}
	after, err := listFreshWorktrees(repository.Root)
	if err != nil {
		blockAction(result, err)
		return
	}
	if _, exists := freshWorktreeByPath(after, target.Path); exists || pathExistsLocal(target.Path) {
		result.Result = "blocked"
		result.Issue = "worktree removal postcondition failed"
		result.PhysicalCleanupState = StatePartialRepairable
		return
	}
	result.Result = "applied"
	result.PhysicalCleanupState = StateVerifiedRetired
	result.RefRetirementState = StatePreserved
	result.ObservedAfter = "worktree registration and target path absent"
}

func validateFreshWorktree(
	options ApplyOptions,
	repository Repository,
	artifact Artifact,
	planned Worktree,
	target freshWorktree,
	all []freshWorktree,
) error {
	if !sameCanonicalPath(target.Path, artifact.Path) || target.Head != artifact.SHA ||
		target.Branch != strings.TrimPrefix(artifact.Ref, "refs/heads/") {
		return fmt.Errorf("fresh target identity does not match the cutoff-bound plan")
	}
	if planned.GitDir == "" || !sameCanonicalPath(planned.GitDir, target.GitDir) {
		return fmt.Errorf("worktree generation or Git-admin identity changed after planning")
	}
	if target.Locked {
		return fmt.Errorf("target worktree is locked")
	}
	if len(target.Dirt.Tracked) > 0 || len(target.Dirt.Untracked) > 0 || len(target.Dirt.Ignored) > 0 {
		return fmt.Errorf("target worktree has tracked, untracked, or ignored dirt")
	}
	current := canonicalPath(options.CurrentWorktree)
	if current == "" {
		current = canonicalPath(repository.CurrentWorktree)
	}
	if sameCanonicalPath(current, target.Path) {
		return fmt.Errorf("current executing worktree cannot retire itself")
	}
	if len(all) > 0 && sameCanonicalPath(all[0].Path, target.Path) {
		return fmt.Errorf("primary worktree cannot be retired")
	}
	if target.Branch == repository.DefaultBranch || planned.IsDefault {
		return fmt.Errorf("default worktree cannot be retired")
	}
	common := repository.CommonDir
	ok, gitDir := gitAdminRoundTrip(target.Path, common)
	if !ok || !sameCanonicalPath(gitDir, planned.GitDir) {
		return fmt.Errorf("Git-admin round-trip or generation check failed")
	}
	if err := exactTerminalClaim(repository, artifact, options.Bundle.Plan.Cutoff, options.CurrentSession); err != nil {
		return err
	}
	return nil
}

func applyLocalRef(options ApplyOptions, repository Repository, artifact Artifact, action PlannedAction, result *ActionReceipt) {
	ref := artifact.Ref
	if !strings.HasPrefix(ref, "refs/heads/") || artifact.SHA == "" {
		blockAction(result, fmt.Errorf("local ref target identity is incomplete"))
		return
	}
	current, err := gitOutput(repository.Root, "rev-parse", ref+"^{commit}")
	if err != nil || current == "" {
		result.Result = "already-applied"
		result.RefRetirementState = StateVerifiedRetired
		result.PhysicalCleanupState = StateVerifiedRetired
		result.ObservedAfter = "local ref absent"
		return
	}
	if current != artifact.SHA {
		blockAction(result, fmt.Errorf("local ref changed after planning; exact-SHA deletion refused"))
		return
	}
	branch := strings.TrimPrefix(ref, "refs/heads/")
	worktrees, err := listFreshWorktrees(repository.Root)
	if err != nil {
		blockAction(result, err)
		return
	}
	for _, worktree := range worktrees {
		if worktree.Branch == branch {
			blockAction(result, fmt.Errorf("local ref is still checked out in a worktree"))
			return
		}
	}
	result.PhysicalCleanupState = StateVerifiedRetired
	delivery, err := freshDeliveryState(repository.Root, repository, branch, artifact.SHA, true)
	if err != nil {
		blockAction(result, err)
		return
	}
	if delivery != StateIntegratedDefault || !hasFreshAncestryProof(action.DedupProofs) {
		blockAction(result, fmt.Errorf("integrated delivery and exact ancestry proof are required"))
		return
	}
	result.DeliveryState = delivery
	for _, remote := range repository.Remotes {
		fresh, refreshErr := refreshRemotes(repository.Root, branch)
		if refreshErr != nil {
			blockAction(result, refreshErr)
			return
		}
		for _, item := range fresh {
			if item.Name == remote.Name && item.DefaultBranch == branch {
				blockAction(result, fmt.Errorf("authoritative remote default ref cannot be retired"))
				return
			}
		}
	}
	if repository.DefaultBranch == branch {
		blockAction(result, fmt.Errorf("default local ref cannot be retired"))
		return
	}
	result.ObservedBefore = ref + "@" + current
	if err := deleteExactLocalRef(repository.Root, ref, artifact.SHA); err != nil {
		blockAction(result, err)
		return
	}
	if value, _ := gitOutput(repository.Root, "show-ref", "--verify", ref); value != "" {
		blockAction(result, fmt.Errorf("local ref deletion postcondition failed"))
		return
	}
	result.Result = "applied"
	result.RefRetirementState = StateVerifiedRetired
	result.ObservedAfter = "local ref absent"
}

func verifyOne(options ApplyOptions, repository Repository, artifact Artifact, action PlannedAction, prior ActionReceipt) ActionReceipt {
	result := prior
	result.VerifiedAt = options.Now.UTC()
	lock, err := acquireRepositoryLock(repository, options.LockTimeout)
	if err != nil {
		result.Result = "contended"
		result.Issue = "compatible repository lock unavailable: " + err.Error()
		return result
	}
	defer lock.Close()
	switch action.ActionClass {
	case ActionWorktreeRetire:
		worktrees, listErr := listFreshWorktrees(repository.Root)
		if listErr != nil {
			blockAction(&result, listErr)
			return result
		}
		if _, found := freshWorktreeByPath(worktrees, artifact.Path); found {
			result.PhysicalCleanupState = StatePreserved
			return result
		}
		for _, worktree := range worktrees {
			if worktree.Branch == strings.TrimPrefix(artifact.Ref, "refs/heads/") {
				result.PhysicalCleanupState = StateBlocked
				result.Issue = "target worktree moved or was recreated under a different identity"
				return result
			}
		}
		if pathExistsLocal(artifact.Path) {
			empty, residualErr := exactEmptyResidual(artifact.Path, artifact.Path)
			if residualErr != nil {
				result.PhysicalCleanupState = StateBlocked
				result.Issue = residualErr.Error()
			} else if empty {
				result.PhysicalCleanupState = StatePartialRepairable
				result.Issue = "exact empty residual is repairable by a repeated apply"
			}
			return result
		}
		result.PhysicalCleanupState = StateVerifiedRetired
	case ActionLocalRefRetire:
		if current, _ := gitOutput(repository.Root, "show-ref", "--verify", artifact.Ref); current == "" {
			result.RefRetirementState = StateVerifiedRetired
		} else {
			result.RefRetirementState = StatePreserved
		}
	default:
		result.Result = "unavailable"
		result.SessionRecordState = StateUnavailable
	}
	return result
}

func validatePlannedPredicates(action PlannedAction, policy ActionPolicy) error {
	evidence := map[string]PredicateEvidence{}
	for _, predicate := range action.Preconditions {
		if _, duplicate := evidence[predicate.Name]; duplicate {
			return fmt.Errorf("duplicate mandatory predicate %s", predicate.Name)
		}
		evidence[predicate.Name] = predicate
	}
	for _, required := range policy.Mandatory {
		item, ok := evidence[required.Name]
		if !ok || item.State != PredicatePass || !item.Authoritative {
			return fmt.Errorf("mandatory predicate %s is missing or failed", required.Name)
		}
		if !containsString(required.Adapters, item.Source) {
			return fmt.Errorf("mandatory predicate %s has a non-authoritative source", required.Name)
		}
		if item.ObservedAt.After(action.Cutoff) {
			return fmt.Errorf("mandatory predicate %s is post-cutoff", required.Name)
		}
		freshness, _ := time.ParseDuration(required.Freshness)
		if action.Cutoff.Sub(item.ObservedAt) > freshness {
			return fmt.Errorf("mandatory predicate %s is stale", required.Name)
		}
	}
	return nil
}

func targetForAction(ledger Ledger, action PlannedAction) (Repository, Artifact, error) {
	if len(action.ArtifactIDs) != 1 {
		return Repository{}, Artifact{}, fmt.Errorf("action %s must target exactly one artifact", action.ID)
	}
	for _, repository := range ledger.Repositories {
		if repository.ID != action.RepositoryID {
			continue
		}
		for _, artifact := range repository.Artifacts {
			if artifact.ID == action.ArtifactIDs[0] {
				return repository, artifact, nil
			}
		}
		return Repository{}, Artifact{}, fmt.Errorf("action %s target artifact is unavailable", action.ID)
	}
	return Repository{}, Artifact{}, fmt.Errorf("action %s target repository is unavailable", action.ID)
}

func plannedWorktree(repository Repository, path string) Worktree {
	for _, worktree := range repository.Worktrees {
		if sameCanonicalPath(worktree.Path, path) {
			return worktree
		}
	}
	return Worktree{}
}

func freshWorktreeByPath(worktrees []freshWorktree, path string) (freshWorktree, bool) {
	for _, worktree := range worktrees {
		if sameCanonicalPath(worktree.Path, path) {
			return worktree, true
		}
	}
	return freshWorktree{}, false
}

func newActionReceipt(action PlannedAction, artifact Artifact, now time.Time) ActionReceipt {
	artifactID := ""
	if len(action.ArtifactIDs) == 1 {
		artifactID = action.ArtifactIDs[0]
	}
	return ActionReceipt{
		ActionID: action.ID, ActionClass: action.ActionClass,
		RepositoryID: action.RepositoryID, ArtifactID: artifactID,
		TargetPath: artifact.Path, TargetRef: artifact.Ref, TargetSHA: artifact.SHA,
		Result: "preserved", DeliveryState: StatePreserved,
		PhysicalCleanupState: StatePreserved, RefRetirementState: StatePreserved,
		SessionRecordState: StateUnavailable, AppliedAt: now.UTC(),
	}
}

func blockAction(result *ActionReceipt, err error) {
	result.Result = "blocked"
	result.Issue = err.Error()
	if result.PhysicalCleanupState != StateVerifiedRetired {
		result.PhysicalCleanupState = StateBlocked
	}
}

func actionReceiptTerminal(receipt ActionReceipt) bool {
	return receipt.Result == "applied" || receipt.Result == "already-applied" ||
		receipt.Result == "unavailable" || receipt.Result == "quarantined"
}

func plannedActionByID(plan Plan, id string) (PlannedAction, bool) {
	for _, action := range plan.Actions {
		if action.ID == id {
			return action, true
		}
	}
	return PlannedAction{}, false
}

func sameProofs(left, right []DedupProof) bool {
	leftJSON, _ := MarshalStable(left)
	rightJSON, _ := MarshalStable(right)
	return string(leftJSON) == string(rightJSON)
}

func samePredicates(left, right []PredicateEvidence) bool {
	leftJSON, _ := MarshalStable(left)
	rightJSON, _ := MarshalStable(right)
	return string(leftJSON) == string(rightJSON)
}

func hasFreshAncestryProof(proofs []DedupProof) bool {
	for _, proof := range proofs {
		if proof.Authoritative && proof.Identity != "" &&
			(proof.Algorithm == DedupCommitAncestry || proof.Algorithm == DedupRemoteDefaultAncestry) {
			return true
		}
	}
	return false
}

func pathExistsLocal(path string) bool {
	_, err := os.Lstat(filepath.Clean(path))
	return err == nil
}
