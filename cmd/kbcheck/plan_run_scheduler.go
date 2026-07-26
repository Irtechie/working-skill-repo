package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const planRunLeaseSchemaVersion = 1

type planRunLeaseState struct {
	SchemaVersion int                      `json:"schema_version"`
	RepoIdentity  string                   `json:"repo_identity"`
	Leases        map[string]planRunLease  `json:"leases"`
	Events        []planRunLeaseStateEvent `json:"events,omitempty"`
}

type planRunLeaseStateEvent struct {
	At         string `json:"at"`
	Action     string `json:"action"`
	RunID      string `json:"run_id"`
	Generation int64  `json:"generation"`
}

type planRunLease struct {
	RunID            string        `json:"run_id"`
	ManifestPath     string        `json:"manifest_path"`
	OwnerToken       string        `json:"owner_token"`
	OwnerFingerprint string        `json:"owner_fingerprint,omitempty"`
	Worktree         string        `json:"worktree,omitempty"`
	Branch           string        `json:"branch,omitempty"`
	BaseSHA          string        `json:"base_sha,omitempty"`
	IntegrationHead  string        `json:"integration_head,omitempty"`
	RepoIdentity     string        `json:"repo_identity"`
	Status           string        `json:"status"`
	Generation       int64         `json:"generation"`
	Claims           []leaseClaim  `json:"claims"`
	AcquiredAt       string        `json:"acquired_at"`
	HeartbeatAt      string        `json:"heartbeat_at"`
	ExpiresAt        string        `json:"expires_at"`
	LastUpdatedAt    string        `json:"last_updated_at"`
	LeaseDuration    string        `json:"lease_duration"`
	Limitations      []string      `json:"limitations"`
	Metadata         leaseMetadata `json:"metadata"`
}

type planRunLeaseCommandOptions struct {
	Action       string
	StateRoot    string
	RunID        string
	ManifestPath string
	OwnerToken   string
	Generation   int64
	TTL          time.Duration
	Files        []string
	Prefixes     []string
	Domains      []string
	Resources    []string
	RepoRoot     string
	RepoID       string
	Now          time.Time
}

type planRunCollision struct {
	RunID         string     `json:"run_id"`
	ManifestPath  string     `json:"manifest_path,omitempty"`
	Claim         leaseClaim `json:"claim"`
	ExistingClaim leaseClaim `json:"existing_claim"`
	Reason        string     `json:"reason"`
}

type planRunLeaseResult struct {
	OK         bool               `json:"ok"`
	Action     string             `json:"action"`
	Issue      string             `json:"issue,omitempty"`
	Lease      *planRunLease      `json:"lease,omitempty"`
	Leases     []planRunLease     `json:"leases,omitempty"`
	Collisions []planRunCollision `json:"collisions,omitempty"`
	Requeued   bool               `json:"requeued,omitempty"`
	StateRoot  string             `json:"state_root,omitempty"`
}

func runPlanRunLeaseCommand(root string, opts options, stdout, stderr io.Writer) int {
	result, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action:       opts.sliceLeaseAction,
		StateRoot:    opts.sliceLeaseStateRoot,
		RunID:        opts.runID,
		ManifestPath: opts.manifest,
		OwnerToken:   opts.ownerToken,
		Generation:   opts.leaseGeneration,
		TTL:          opts.leaseTTL,
		Files:        opts.leaseFiles,
		Prefixes:     opts.leasePrefixes,
		Domains:      opts.leaseDomains,
		Resources:    opts.leaseResources,
		RepoRoot:     root,
		RepoID:       opts.repoIdentity,
		Now:          time.Now().UTC(),
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK {
		switch {
		case result.Lease != nil:
			fmt.Fprintf(stdout, "plan-run lease: %s run=%s generation=%d\n", result.Action, result.Lease.RunID, result.Lease.Generation)
		default:
			fmt.Fprintf(stdout, "plan-run lease: %s ok leases=%d\n", result.Action, len(result.Leases))
		}
	} else {
		fmt.Fprintf(stdout, "plan-run lease: %s blocked: %s\n", result.Action, result.Issue)
	}
	if !result.OK {
		return 2
	}
	return 0
}

func executePlanRunLease(opts planRunLeaseCommandOptions) (planRunLeaseResult, error) {
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action == "" {
		return planRunLeaseResult{}, fmt.Errorf("plan-run-lease requires --action")
	}
	switch action {
	case "acquire", "status", "renew", "expand", "release", "recover":
	default:
		return planRunLeaseResult{}, fmt.Errorf("unsupported plan-run-lease action %q", opts.Action)
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	if action == "acquire" && strings.TrimSpace(opts.RepoRoot) != "" {
		hydrated, err := hydratePlanRunForecast(opts)
		if err != nil {
			return blockedPlanRunLease(action, err.Error(), nil, ""), nil
		}
		opts = hydrated
	}
	stateRoot, err := resolvePlanRunLeaseStateRoot(opts)
	if err != nil {
		return planRunLeaseResult{}, err
	}
	if opts.RepoID == "" {
		opts.RepoID = repoIdentity(opts.RepoRoot, stateRoot)
	}
	result := planRunLeaseResult{Action: action, StateRoot: stateRoot}
	err = withSliceLeaseStateLock(stateRoot, func() error {
		state, err := loadPlanRunLeaseState(stateRoot)
		if err != nil {
			return err
		}
		if state.SchemaVersion == 0 {
			state.SchemaVersion = planRunLeaseSchemaVersion
		}
		if state.Leases == nil {
			state.Leases = map[string]planRunLease{}
		}
		if state.RepoIdentity == "" {
			state.RepoIdentity = opts.RepoID
		}
		if state.RepoIdentity != opts.RepoID {
			return fmt.Errorf("plan-run lease repository identity mismatch")
		}
		slices, err := loadSliceLeaseState(stateRoot)
		if err != nil {
			return err
		}
		if slices.RepoIdentity != "" && slices.RepoIdentity != opts.RepoID {
			return fmt.Errorf("plan-run and slice lease repository identities do not match")
		}

		switch action {
		case "acquire":
			result = acquirePlanRunLease(state, slices, opts, stateRoot)
		case "status":
			result = statusPlanRunLeases(state, opts, stateRoot)
		case "renew":
			result = renewPlanRunLease(state, opts, stateRoot)
		case "expand":
			result = expandPlanRunLease(state, slices, opts, stateRoot)
		case "release":
			result = releasePlanRunLease(state, slices, opts, stateRoot)
		case "recover":
			result = recoverPlanRunLease(state, slices, opts, stateRoot)
		}
		if result.OK && action != "status" {
			appendPlanRunLeaseEvent(&state, opts.Now, action, result.Lease)
			return savePlanRunLeaseState(stateRoot, state)
		}
		return nil
	})
	redactPlanRunLeaseResult(&result)
	return result, err
}

func redactPlanRunLeaseResult(result *planRunLeaseResult) {
	if result.Lease != nil {
		redacted := publicPlanRunLease(*result.Lease)
		result.Lease = &redacted
	}
	for index := range result.Leases {
		result.Leases[index] = publicPlanRunLease(result.Leases[index])
	}
}

func publicPlanRunLease(lease planRunLease) planRunLease {
	if lease.OwnerToken != "" {
		sum := sha256.Sum256([]byte(lease.OwnerToken))
		lease.OwnerFingerprint = fmt.Sprintf("%x", sum[:6])
	}
	lease.OwnerToken = ""
	return lease
}

func hydratePlanRunForecast(opts planRunLeaseCommandOptions) (planRunLeaseCommandOptions, error) {
	manifestPath := resolveInputPath(opts.RepoRoot, opts.ManifestPath)
	manifestPath, err := filepath.Abs(filepath.Clean(manifestPath))
	if err != nil {
		return opts, err
	}
	root, err := filepath.Abs(filepath.Clean(opts.RepoRoot))
	if err != nil {
		return opts, err
	}
	relativeManifest, err := filepath.Rel(root, manifestPath)
	if err != nil {
		return opts, err
	}
	if relativeManifest == ".." || strings.HasPrefix(relativeManifest, ".."+string(filepath.Separator)) {
		return opts, fmt.Errorf("manifest must stay inside the repository")
	}
	slices, err := parseManifestSlices(manifestPath)
	if err != nil {
		return opts, err
	}
	files := append([]string{}, opts.Files...)
	files = append(files, filepath.ToSlash(relativeManifest))
	domains := append([]string{}, opts.Domains...)
	resources := append([]string{}, opts.Resources...)
	for _, slice := range slices {
		domains = append(domains, slice.ConflictDomains...)
		resources = append(resources, slice.SharedResources...)
		if strings.TrimSpace(slice.Path) == "" {
			continue
		}
		slicePath := resolveInputPath(root, slice.Path)
		slicePath, err = filepath.Abs(filepath.Clean(slicePath))
		if err != nil {
			return opts, err
		}
		relativeSlice, err := filepath.Rel(root, slicePath)
		if err != nil || relativeSlice == ".." || strings.HasPrefix(relativeSlice, ".."+string(filepath.Separator)) {
			return opts, fmt.Errorf("slice plan must stay inside the repository: %s", slice.Path)
		}
		files = append(files, filepath.ToSlash(relativeSlice))
		expected, err := readExpectedFiles(slicePath)
		if err != nil {
			return opts, err
		}
		files = append(files, expected...)
	}
	opts.ManifestPath = manifestPath
	opts.Files = files
	opts.Domains = domains
	opts.Resources = resources
	return opts, nil
}

func readExpectedFiles(slicePath string) ([]string, error) {
	content, err := os.ReadFile(slicePath)
	if err != nil {
		return nil, fmt.Errorf("read slice plan %s: %w", slicePath, err)
	}
	expected := []string{}
	inExpected := false
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "expected_files:" {
			inExpected = true
			continue
		}
		if !inExpected {
			continue
		}
		if line != "" && len(line)-len(strings.TrimLeft(line, " \t")) == 0 {
			break
		}
		if strings.HasPrefix(trimmed, "- path:") {
			value := cleanYAMLScalar(strings.TrimSpace(strings.TrimPrefix(trimmed, "- path:")))
			if value != "" {
				expected = append(expected, value)
			}
		}
	}
	return expected, nil
}

func acquirePlanRunLease(state planRunLeaseState, slices sliceLeaseState, opts planRunLeaseCommandOptions, stateRoot string) planRunLeaseResult {
	lease, err := newPlanRunLease(state, opts)
	if err != nil {
		return blockedPlanRunLease("acquire", err.Error(), nil, stateRoot)
	}
	if manifestHasPlanRunDefault(opts.ManifestPath) {
		kbID, err := planRunManifestID(opts.ManifestPath)
		if err != nil {
			return blockedPlanRunLease("acquire", err.Error(), nil, stateRoot)
		}
		receipt, err := loadPlanRunWorkspaceReceipt(planRunWorkspaceOptions{RepoRoot: opts.RepoRoot}, kbID)
		if err != nil {
			return blockedPlanRunLease("acquire", "manifest-owned plan-worktree receipt is required: "+err.Error(), nil, stateRoot)
		}
		if receipt.RunID != opts.RunID || receipt.OwnerToken != opts.OwnerToken ||
			!samePath(receipt.Worktree, opts.RepoRoot) ||
			gitOutput(opts.RepoRoot, "branch", "--show-current") != receipt.IntegrationRef {
			return blockedPlanRunLease("acquire", "plan-run lease must be acquired from the exact manifest-owned worktree/ref and owner lineage", nil, stateRoot)
		}
		lease.Worktree = receipt.Worktree
		lease.Branch = receipt.IntegrationRef
		lease.BaseSHA = receipt.BaseSHA
		lease.IntegrationHead = receipt.IntegrationHead
	}
	if existing, ok := state.Leases[lease.RunID]; ok && existing.Status == "active" {
		return blockedPlanRunLease("acquire", "run already has an active or stale lease; use renew, release, or recover", &existing, stateRoot)
	}
	collisions := activePlanRunCollisions(state, lease, opts.Now, "")
	collisions = append(collisions, planRunAgainstSliceCollisions(slices, lease, opts.Now)...)
	if len(collisions) > 0 {
		return planRunLeaseResult{
			OK: false, Action: "acquire", Issue: "active plan-run or slice claim collision",
			Collisions: collisions, StateRoot: stateRoot,
		}
	}
	state.Leases[lease.RunID] = lease
	return planRunLeaseResult{OK: true, Action: "acquire", Lease: &lease, StateRoot: stateRoot}
}

func statusPlanRunLeases(state planRunLeaseState, opts planRunLeaseCommandOptions, stateRoot string) planRunLeaseResult {
	if strings.TrimSpace(opts.RunID) != "" {
		lease, ok := state.Leases[opts.RunID]
		if !ok {
			return blockedPlanRunLease("status", "plan-run lease not found", nil, stateRoot)
		}
		if opts.OwnerToken != "" && opts.OwnerToken != lease.OwnerToken {
			return blockedPlanRunLease("status", "owner token mismatch", &lease, stateRoot)
		}
		lease.Status = effectivePlanRunLeaseStatus(lease, opts.Now)
		return planRunLeaseResult{OK: true, Action: "status", Lease: &lease, StateRoot: stateRoot}
	}
	leases := make([]planRunLease, 0, len(state.Leases))
	for _, lease := range state.Leases {
		lease.Status = effectivePlanRunLeaseStatus(lease, opts.Now)
		leases = append(leases, lease)
	}
	sort.Slice(leases, func(i, j int) bool { return leases[i].RunID < leases[j].RunID })
	return planRunLeaseResult{OK: true, Action: "status", Leases: leases, StateRoot: stateRoot}
}

func renewPlanRunLease(state planRunLeaseState, opts planRunLeaseCommandOptions, stateRoot string) planRunLeaseResult {
	lease, ok := state.Leases[opts.RunID]
	if !ok {
		return blockedPlanRunLease("renew", "plan-run lease not found", nil, stateRoot)
	}
	if issue := requirePlanRunOwnerAndGeneration(lease, opts); issue != "" {
		return blockedPlanRunLease("renew", issue, &lease, stateRoot)
	}
	if effectivePlanRunLeaseStatus(lease, opts.Now) != "active" {
		return blockedPlanRunLease("renew", "lease is expired; use recover", &lease, stateRoot)
	}
	lease.Generation++
	refreshPlanRunLeaseTimes(&lease, opts.Now, opts.TTL)
	state.Leases[lease.RunID] = lease
	return planRunLeaseResult{OK: true, Action: "renew", Lease: &lease, StateRoot: stateRoot}
}

func expandPlanRunLease(state planRunLeaseState, slices sliceLeaseState, opts planRunLeaseCommandOptions, stateRoot string) planRunLeaseResult {
	lease, ok := state.Leases[opts.RunID]
	if !ok {
		return blockedPlanRunLease("expand", "plan-run lease not found", nil, stateRoot)
	}
	if issue := requirePlanRunOwnerAndGeneration(lease, opts); issue != "" {
		return blockedPlanRunLease("expand", issue, &lease, stateRoot)
	}
	if effectivePlanRunLeaseStatus(lease, opts.Now) != "active" {
		return blockedPlanRunLease("expand", "lease is expired; use recover", &lease, stateRoot)
	}
	additions, err := normalizePlanRunClaims(opts.Files, opts.Prefixes, opts.Domains, opts.Resources)
	if err != nil {
		return blockedPlanRunLease("expand", err.Error(), &lease, stateRoot)
	}
	if len(additions) == 0 {
		return blockedPlanRunLease("expand", "at least one observed claim is required", &lease, stateRoot)
	}
	contender := lease
	contender.Claims = mergeLeaseClaims(lease.Claims, additions)
	collisions := activePlanRunCollisions(state, contender, opts.Now, lease.RunID)
	collisions = append(collisions, planRunAgainstSliceCollisions(slices, contender, opts.Now)...)
	if len(collisions) > 0 {
		return planRunLeaseResult{
			OK: false, Action: "expand", Issue: "observed claim collides; requeue before write",
			Lease: &lease, Collisions: collisions, Requeued: true, StateRoot: stateRoot,
		}
	}
	contender.Generation++
	refreshPlanRunLeaseTimes(&contender, opts.Now, opts.TTL)
	state.Leases[contender.RunID] = contender
	return planRunLeaseResult{OK: true, Action: "expand", Lease: &contender, StateRoot: stateRoot}
}

func releasePlanRunLease(state planRunLeaseState, slices sliceLeaseState, opts planRunLeaseCommandOptions, stateRoot string) planRunLeaseResult {
	lease, ok := state.Leases[opts.RunID]
	if !ok {
		return blockedPlanRunLease("release", "plan-run lease not found", nil, stateRoot)
	}
	if issue := requirePlanRunOwnerAndGeneration(lease, opts); issue != "" {
		return blockedPlanRunLease("release", issue, &lease, stateRoot)
	}
	if strings.TrimSpace(lease.Worktree) != "" {
		return blockedPlanRunLease("release", "manifest-owned plan-run leases are released atomically by plan-worktree complete", &lease, stateRoot)
	}
	for _, slice := range slices.Leases {
		if slice.RunID == lease.RunID && effectiveLeaseStatus(slice, opts.Now) == "active" {
			return blockedPlanRunLease("release", fmt.Sprintf("run still owns active slice lease %s; release slices first", slice.SliceID), &lease, stateRoot)
		}
	}
	lease.Generation++
	lease.Status = "released"
	lease.LastUpdatedAt = opts.Now.Format(time.RFC3339Nano)
	state.Leases[lease.RunID] = lease
	return planRunLeaseResult{OK: true, Action: "release", Lease: &lease, StateRoot: stateRoot}
}

func recoverPlanRunLease(state planRunLeaseState, slices sliceLeaseState, opts planRunLeaseCommandOptions, stateRoot string) planRunLeaseResult {
	lease, ok := state.Leases[opts.RunID]
	if !ok {
		return blockedPlanRunLease("recover", "plan-run lease not found", nil, stateRoot)
	}
	if issue := requirePlanRunOwnerAndGeneration(lease, opts); issue != "" {
		return blockedPlanRunLease("recover", issue, &lease, stateRoot)
	}
	if effectivePlanRunLeaseStatus(lease, opts.Now) == "active" {
		return blockedPlanRunLease("recover", "lease is still active", &lease, stateRoot)
	}
	collisions := activePlanRunCollisions(state, lease, opts.Now, lease.RunID)
	collisions = append(collisions, planRunAgainstSliceCollisions(slices, lease, opts.Now)...)
	if len(collisions) > 0 {
		return planRunLeaseResult{
			OK: false, Action: "recover", Issue: "lease recovery collides; requeue",
			Lease: &lease, Collisions: collisions, Requeued: true, StateRoot: stateRoot,
		}
	}
	lease.Generation++
	lease.Status = "active"
	refreshPlanRunLeaseTimes(&lease, opts.Now, opts.TTL)
	state.Leases[lease.RunID] = lease
	return planRunLeaseResult{OK: true, Action: "recover", Lease: &lease, StateRoot: stateRoot}
}

func newPlanRunLease(state planRunLeaseState, opts planRunLeaseCommandOptions) (planRunLease, error) {
	if strings.TrimSpace(opts.RunID) == "" {
		return planRunLease{}, fmt.Errorf("run-id is required")
	}
	if strings.TrimSpace(opts.ManifestPath) == "" {
		return planRunLease{}, fmt.Errorf("manifest is required")
	}
	owner := strings.TrimSpace(opts.OwnerToken)
	if owner == "" {
		return planRunLease{}, fmt.Errorf("owner-token is required")
	}
	claims, err := normalizePlanRunClaims(opts.Files, opts.Prefixes, opts.Domains, opts.Resources)
	if err != nil {
		return planRunLease{}, err
	}
	if len(claims) == 0 {
		return planRunLease{}, fmt.Errorf("at least one file, prefix, domain, or resource claim is required")
	}
	now := opts.Now
	lease := planRunLease{
		RunID:        opts.RunID,
		ManifestPath: filepath.ToSlash(filepath.Clean(opts.ManifestPath)),
		OwnerToken:   owner,
		RepoIdentity: state.RepoIdentity,
		Status:       "active",
		Generation:   1,
		Claims:       claims,
		AcquiredAt:   now.Format(time.RFC3339Nano),
		Limitations: []string{
			"coordinates only sessions sharing this Git common directory",
			"separate clones and machines are not coordinated; rely on branch/PR protections",
			"does not provide a remote lock, daemon, conflict resolver, or default-branch delivery",
		},
		Metadata: leaseMetadata{CoordinationScope: "git-common-dir"},
	}
	refreshPlanRunLeaseTimes(&lease, now, opts.TTL)
	return lease, nil
}

func normalizePlanRunClaims(files, prefixes, domains, resources []string) ([]leaseClaim, error) {
	claims, err := normalizeLeaseClaims(files, prefixes, resources)
	if err != nil {
		return nil, err
	}
	for _, raw := range domains {
		value := strings.ToLower(strings.TrimSpace(raw))
		parts := strings.SplitN(value, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(value, " \t\r\n") {
			return nil, fmt.Errorf("domain claims must be kind:value without whitespace")
		}
		switch parts[0] {
		case "file":
			normalized, err := normalizeLeaseClaimPath(parts[1])
			if err != nil {
				return nil, err
			}
			claims = append(claims, leaseClaim{Kind: "file", Value: normalized})
		case "prefix":
			normalized, err := normalizeLeaseClaimPath(parts[1])
			if err != nil {
				return nil, err
			}
			claims = append(claims, leaseClaim{Kind: "prefix", Value: strings.TrimSuffix(normalized, "/") + "/"})
		default:
			claims = append(claims, leaseClaim{Kind: "domain", Value: value})
		}
	}
	return dedupeAndSortLeaseClaims(claims), nil
}

func mergeLeaseClaims(existing, additions []leaseClaim) []leaseClaim {
	return dedupeAndSortLeaseClaims(append(append([]leaseClaim{}, existing...), additions...))
}

func dedupeAndSortLeaseClaims(claims []leaseClaim) []leaseClaim {
	seen := map[string]bool{}
	result := make([]leaseClaim, 0, len(claims))
	for _, claim := range claims {
		key := claim.Kind + "\x00" + claim.Value
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, claim)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Kind == result[j].Kind {
			return result[i].Value < result[j].Value
		}
		return result[i].Kind < result[j].Kind
	})
	return result
}

func planRunClaimsConflict(left, right leaseClaim) bool {
	if (left.Kind == "domain" || left.Kind == "resource") &&
		(right.Kind == "domain" || right.Kind == "resource") {
		return left.Value == right.Value
	}
	if left.Kind == "domain" || right.Kind == "domain" {
		return left.Kind == right.Kind && left.Value == right.Value
	}
	return leaseClaimsConflict(left, right)
}

func activePlanRunCollisions(state planRunLeaseState, contender planRunLease, now time.Time, ignoreRunID string) []planRunCollision {
	collisions := []planRunCollision{}
	for _, existing := range state.Leases {
		if existing.RunID == ignoreRunID || effectivePlanRunLeaseStatus(existing, now) != "active" {
			continue
		}
		for _, existingClaim := range existing.Claims {
			for _, claim := range contender.Claims {
				if planRunClaimsConflict(existingClaim, claim) {
					collisions = append(collisions, planRunCollision{
						RunID: existing.RunID, ManifestPath: existing.ManifestPath,
						Claim: claim, ExistingClaim: existingClaim,
						Reason: fmt.Sprintf("run %s owns %s claim %s", existing.RunID, existingClaim.Kind, existingClaim.Value),
					})
				}
			}
		}
	}
	return collisions
}

func planRunAgainstSliceCollisions(state sliceLeaseState, contender planRunLease, now time.Time) []planRunCollision {
	collisions := []planRunCollision{}
	for _, slice := range state.Leases {
		if effectiveLeaseStatus(slice, now) != "active" {
			continue
		}
		if slice.RunID == contender.RunID {
			if slice.OwnerToken != contender.OwnerToken {
				collisions = append(collisions, planRunCollision{
					RunID:  slice.RunID,
					Reason: fmt.Sprintf("run %s slice %s has a different owner token", slice.RunID, slice.SliceID),
				})
				continue
			}
			for _, claim := range slice.Claims {
				if !planRunClaimCovers(contender.Claims, claim) {
					collisions = append(collisions, planRunCollision{
						RunID: slice.RunID, Claim: claim, ExistingClaim: claim,
						Reason: fmt.Sprintf("run %s slice %s claim %s:%s is outside the manifest claim", slice.RunID, slice.SliceID, claim.Kind, claim.Value),
					})
				}
			}
			continue
		}
		for _, existingClaim := range slice.Claims {
			for _, claim := range contender.Claims {
				if planRunClaimsConflict(existingClaim, claim) {
					collisions = append(collisions, planRunCollision{
						RunID: slice.RunID, Claim: claim, ExistingClaim: existingClaim,
						Reason: fmt.Sprintf("run %s slice %s owns %s claim %s", slice.RunID, slice.SliceID, existingClaim.Kind, existingClaim.Value),
					})
				}
			}
		}
	}
	return collisions
}

func requirePlanRunOwnerAndGeneration(lease planRunLease, opts planRunLeaseCommandOptions) string {
	if strings.TrimSpace(opts.RunID) == "" {
		return "run-id is required"
	}
	if opts.OwnerToken == "" || opts.OwnerToken != lease.OwnerToken {
		return "owner token mismatch"
	}
	if opts.Generation <= 0 || opts.Generation != lease.Generation {
		return "generation mismatch"
	}
	return ""
}

func effectivePlanRunLeaseStatus(lease planRunLease, now time.Time) string {
	if lease.Status != "active" {
		return lease.Status
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, lease.ExpiresAt)
	if err != nil || !now.Before(expiresAt) {
		return "expired"
	}
	return "active"
}

func refreshPlanRunLeaseTimes(lease *planRunLease, now time.Time, ttl time.Duration) {
	if ttl <= 0 {
		ttl = defaultSliceLeaseTTL
	}
	lease.HeartbeatAt = now.Format(time.RFC3339Nano)
	lease.ExpiresAt = now.Add(ttl).Format(time.RFC3339Nano)
	lease.LastUpdatedAt = lease.HeartbeatAt
	lease.LeaseDuration = ttl.String()
}

func blockedPlanRunLease(action, issue string, lease *planRunLease, stateRoot string) planRunLeaseResult {
	return planRunLeaseResult{OK: false, Action: action, Issue: issue, Lease: lease, StateRoot: stateRoot}
}

func resolvePlanRunLeaseStateRoot(opts planRunLeaseCommandOptions) (string, error) {
	return resolveSliceLeaseStateRoot(sliceLeaseCommandOptions{
		StateRoot: opts.StateRoot,
		RepoRoot:  opts.RepoRoot,
	})
}

func loadPlanRunLeaseState(stateRoot string) (planRunLeaseState, error) {
	path := filepath.Join(stateRoot, "plan-run-leases.json")
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return planRunLeaseState{SchemaVersion: planRunLeaseSchemaVersion, Leases: map[string]planRunLease{}}, nil
	}
	if err != nil {
		return planRunLeaseState{}, err
	}
	var state planRunLeaseState
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return planRunLeaseState{}, err
	}
	if state.SchemaVersion != planRunLeaseSchemaVersion {
		return planRunLeaseState{}, fmt.Errorf("unsupported plan-run lease schema_version %d", state.SchemaVersion)
	}
	if state.Leases == nil {
		state.Leases = map[string]planRunLease{}
	}
	return state, nil
}

func savePlanRunLeaseState(stateRoot string, state planRunLeaseState) error {
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	temp, err := os.CreateTemp(stateRoot, ".plan-run-leases-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	ok := false
	defer func() {
		_ = temp.Close()
		if !ok {
			_ = os.Remove(tempName)
		}
	}()
	if _, err := temp.Write(content); err != nil {
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempName, filepath.Join(stateRoot, "plan-run-leases.json")); err != nil {
		return err
	}
	ok = true
	return nil
}

func appendPlanRunLeaseEvent(state *planRunLeaseState, now time.Time, action string, lease *planRunLease) {
	if lease == nil {
		return
	}
	state.Events = append(state.Events, planRunLeaseStateEvent{
		At: now.Format(time.RFC3339Nano), Action: action, RunID: lease.RunID, Generation: lease.Generation,
	})
	if len(state.Events) > 200 {
		state.Events = state.Events[len(state.Events)-200:]
	}
}

func runPlanRunLeaseSelftest(stdout, stderr io.Writer) int {
	stateRoot, err := os.MkdirTemp("", "kb-plan-run-lease-selftest-*")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(stateRoot)
	now := time.Now().UTC()
	first, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-a", ManifestPath: "a.md",
		OwnerToken: "owner-a", Files: []string{"src/a.go"}, Now: now,
	})
	if err != nil || !first.OK {
		fmt.Fprintf(stderr, "first acquire failed: result=%#v err=%v\n", first, err)
		return 1
	}
	blocked, err := executePlanRunLease(planRunLeaseCommandOptions{
		Action: "acquire", StateRoot: stateRoot, RunID: "run-b", ManifestPath: "b.md",
		OwnerToken: "owner-b", Files: []string{"src/a.go"}, Now: now,
	})
	if err != nil || blocked.OK {
		fmt.Fprintf(stderr, "collision failed: result=%#v err=%v\n", blocked, err)
		return 1
	}
	fmt.Fprintln(stdout, "kb-work plan-run-lease selftest: passed")
	return 0
}
