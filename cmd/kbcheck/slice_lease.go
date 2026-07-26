package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
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

const (
	sliceLeaseSchemaVersion = 1
	defaultSliceLeaseTTL    = 30 * time.Minute
	sliceLeaseLockTimeout   = 30 * time.Second
)

type leaseStringList []string

func (list *leaseStringList) String() string {
	return strings.Join(*list, ",")
}

func (list *leaseStringList) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("empty value")
	}
	*list = append(*list, value)
	return nil
}

type sliceLeaseState struct {
	SchemaVersion int                    `json:"schema_version"`
	RepoIdentity  string                 `json:"repo_identity"`
	Leases        map[string]sliceLease  `json:"leases"`
	Events        []sliceLeaseStateEvent `json:"events,omitempty"`
}

type sliceLeaseStateEvent struct {
	At         string `json:"at"`
	Action     string `json:"action"`
	SliceID    string `json:"slice_id"`
	RunID      string `json:"run_id,omitempty"`
	Generation int64  `json:"generation"`
}

type sliceLease struct {
	SliceID          string        `json:"slice_id"`
	RunID            string        `json:"run_id"`
	OwnerToken       string        `json:"owner_token"`
	OwnerFingerprint string        `json:"owner_fingerprint,omitempty"`
	RepoIdentity     string        `json:"repo_identity"`
	BaseSHA          string        `json:"base_sha"`
	Worktree         string        `json:"worktree,omitempty"`
	Branch           string        `json:"branch,omitempty"`
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

type leaseMetadata struct {
	CoordinationScope string `json:"coordination_scope"`
}

type leaseClaim struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type sliceLeaseResult struct {
	OK         bool             `json:"ok"`
	Action     string           `json:"action"`
	Issue      string           `json:"issue,omitempty"`
	Lease      *sliceLease      `json:"lease,omitempty"`
	Leases     []sliceLease     `json:"leases,omitempty"`
	Collisions []leaseCollision `json:"collisions,omitempty"`
	StateRoot  string           `json:"state_root,omitempty"`
}

type leaseCollision struct {
	SliceID          string     `json:"slice_id"`
	RunID            string     `json:"run_id,omitempty"`
	OwnerToken       string     `json:"owner_token,omitempty"`
	OwnerFingerprint string     `json:"owner_fingerprint,omitempty"`
	Claim            leaseClaim `json:"claim"`
	Reason           string     `json:"reason"`
}

type sliceLeaseCommandOptions struct {
	Action      string
	StateRoot   string
	SliceID     string
	RunID       string
	OwnerToken  string
	Generation  int64
	TTL         time.Duration
	Files       []string
	Prefixes    []string
	Resources   []string
	BaseSHA     string
	Worktree    string
	Branch      string
	RepoRoot    string
	RepoID      string
	Now         time.Time
	AllowCreate bool
}

func registerSliceLeaseFlags(fs *flag.FlagSet, opts *options) {
	fs.StringVar(&opts.sliceLeaseAction, "action", "", "slice-lease action: acquire, status, renew, release, recover")
	fs.StringVar(&opts.sliceLeaseStateRoot, "state-root", "", "slice lease state root; defaults under git common dir")
	fs.StringVar(&opts.sliceID, "slice-id", "", "slice id")
	fs.StringVar(&opts.ownerToken, "owner-token", "", "opaque owner token")
	fs.Int64Var(&opts.leaseGeneration, "generation", 0, "expected lease generation")
	fs.DurationVar(&opts.leaseTTL, "ttl", defaultSliceLeaseTTL, "lease ttl")
	fs.Var((*leaseStringList)(&opts.leaseFiles), "file", "exact repository-relative file claim")
	fs.Var((*leaseStringList)(&opts.leasePrefixes), "prefix", "repository-relative path prefix claim")
	fs.Var((*leaseStringList)(&opts.leaseDomains), "domain", "plan-run conflict domain claim such as skill:kb-work")
	fs.Var((*leaseStringList)(&opts.leaseResources), "resource", "resource claim such as browser:4110")
	fs.StringVar(&opts.baseSHA, "base-sha", "", "base revision")
	fs.StringVar(&opts.worktreePath, "worktree", "", "worktree path")
	fs.StringVar(&opts.branchName, "branch", "", "branch name")
	fs.StringVar(&opts.repoIdentity, "repo-id", "", "canonical repository identity")
}

func runSliceLeaseCommand(root string, opts options, stdout, stderr io.Writer) int {
	command := sliceLeaseCommandOptions{
		Action:     opts.sliceLeaseAction,
		StateRoot:  opts.sliceLeaseStateRoot,
		SliceID:    opts.sliceID,
		RunID:      opts.runID,
		OwnerToken: opts.ownerToken,
		Generation: opts.leaseGeneration,
		TTL:        opts.leaseTTL,
		Files:      opts.leaseFiles,
		Prefixes:   opts.leasePrefixes,
		Resources:  opts.leaseResources,
		BaseSHA:    opts.baseSHA,
		Worktree:   opts.worktreePath,
		Branch:     opts.branchName,
		RepoID:     opts.repoIdentity,
		RepoRoot:   root,
		Now:        time.Now().UTC(),
	}
	result, err := executeSliceLease(command)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else if result.OK {
		switch {
		case result.Lease != nil:
			fmt.Fprintf(stdout, "slice lease: %s slice=%s generation=%d\n", result.Action, result.Lease.SliceID, result.Lease.Generation)
		default:
			fmt.Fprintf(stdout, "slice lease: %s ok leases=%d\n", result.Action, len(result.Leases))
		}
	} else {
		fmt.Fprintf(stdout, "slice lease: %s blocked: %s\n", result.Action, result.Issue)
	}
	if !result.OK {
		return 2
	}
	return 0
}

func executeSliceLease(opts sliceLeaseCommandOptions) (sliceLeaseResult, error) {
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}
	action := strings.ToLower(strings.TrimSpace(opts.Action))
	if action == "" {
		return sliceLeaseResult{}, fmt.Errorf("slice-lease requires --action")
	}
	stateRoot, err := resolveSliceLeaseStateRoot(opts)
	if err != nil {
		return sliceLeaseResult{}, err
	}
	if opts.RepoID == "" {
		opts.RepoID = repoIdentity(opts.RepoRoot, stateRoot)
	}
	if opts.BaseSHA == "" {
		opts.BaseSHA = gitOutput(opts.RepoRoot, "rev-parse", "HEAD")
	}
	if opts.Worktree == "" {
		opts.Worktree = opts.RepoRoot
	}
	if opts.Branch == "" {
		opts.Branch = gitOutput(opts.RepoRoot, "branch", "--show-current")
	}
	result := sliceLeaseResult{Action: action, StateRoot: stateRoot}
	err = withSliceLeaseStateLock(stateRoot, func() error {
		state, err := loadSliceLeaseState(stateRoot)
		if err != nil {
			return err
		}
		if state.SchemaVersion == 0 {
			state.SchemaVersion = sliceLeaseSchemaVersion
		}
		if state.Leases == nil {
			state.Leases = map[string]sliceLease{}
		}
		if state.RepoIdentity == "" {
			state.RepoIdentity = opts.RepoID
		}
		if state.RepoIdentity != opts.RepoID {
			return fmt.Errorf("slice lease repository identity mismatch")
		}
		planRuns, err := loadPlanRunLeaseState(stateRoot)
		if err != nil {
			return err
		}
		if planRuns.RepoIdentity != "" && planRuns.RepoIdentity != opts.RepoID {
			return fmt.Errorf("plan-run and slice lease repository identities do not match")
		}

		switch action {
		case "acquire":
			result = acquireSliceLease(state, planRuns, opts, stateRoot)
		case "status":
			result = statusSliceLeases(state, opts, stateRoot)
		case "renew":
			result = renewSliceLease(state, opts, stateRoot)
		case "release":
			result = releaseSliceLease(state, opts, stateRoot)
		case "recover":
			result = recoverSliceLease(state, opts, stateRoot)
		default:
			return fmt.Errorf("unsupported slice-lease action %q", opts.Action)
		}
		if result.OK && action != "status" {
			appendSliceLeaseEvent(&state, opts.Now, action, result.Lease)
			return saveSliceLeaseState(stateRoot, state)
		}
		return nil
	})
	redactSliceLeaseResult(&result)
	return result, err
}

func redactSliceLeaseResult(result *sliceLeaseResult) {
	if result.Lease != nil {
		redacted := publicSliceLease(*result.Lease)
		result.Lease = &redacted
	}
	for index := range result.Leases {
		result.Leases[index] = publicSliceLease(result.Leases[index])
	}
	for index := range result.Collisions {
		if result.Collisions[index].OwnerToken != "" {
			result.Collisions[index].OwnerFingerprint = publicPlanRunLease(planRunLease{
				OwnerToken: result.Collisions[index].OwnerToken,
			}).OwnerFingerprint
		}
		result.Collisions[index].OwnerToken = ""
	}
}

func publicSliceLease(lease sliceLease) sliceLease {
	if lease.OwnerToken != "" {
		lease.OwnerFingerprint = publicPlanRunLease(planRunLease{OwnerToken: lease.OwnerToken}).OwnerFingerprint
	}
	lease.OwnerToken = ""
	return lease
}

func acquireSliceLease(state sliceLeaseState, planRuns planRunLeaseState, opts sliceLeaseCommandOptions, stateRoot string) sliceLeaseResult {
	base, err := newSliceLease(state, opts, stateRoot, 1)
	if err != nil {
		return blockedSliceLease("acquire", err.Error(), stateRoot)
	}
	key := sliceLeaseKey(base.RunID, base.SliceID)
	if existing, ok := state.Leases[key]; ok && existing.Status == "active" {
		return blockedLease("acquire", "slice already has an active or stale lease; use renew, release, or recover", existing, stateRoot)
	}
	if issue, collisions := validateSlicePlanRunComposition(planRuns, base, opts.Now); issue != "" {
		return sliceLeaseResult{
			OK: false, Action: "acquire", Issue: issue, Collisions: collisions, StateRoot: stateRoot,
		}
	}
	collisions := activeLeaseCollisions(state, base, opts.Now)
	if len(collisions) > 0 {
		return sliceLeaseResult{OK: false, Action: "acquire", Issue: "active lease collision", Collisions: collisions, StateRoot: stateRoot}
	}
	state.Leases[key] = base
	return sliceLeaseResult{OK: true, Action: "acquire", Lease: &base, StateRoot: stateRoot}
}

func statusSliceLeases(state sliceLeaseState, opts sliceLeaseCommandOptions, stateRoot string) sliceLeaseResult {
	leases := make([]sliceLease, 0, len(state.Leases))
	for _, lease := range state.Leases {
		lease.Status = effectiveLeaseStatus(lease, opts.Now)
		leases = append(leases, lease)
	}
	sort.Slice(leases, func(i, j int) bool {
		if leases[i].RunID == leases[j].RunID {
			return leases[i].SliceID < leases[j].SliceID
		}
		return leases[i].RunID < leases[j].RunID
	})
	return sliceLeaseResult{OK: true, Action: "status", Leases: leases, StateRoot: stateRoot}
}

func renewSliceLease(state sliceLeaseState, opts sliceLeaseCommandOptions, stateRoot string) sliceLeaseResult {
	key, lease, ok := findSliceLease(state, opts)
	if !ok {
		return blockedSliceLease("renew", "lease not found", stateRoot)
	}
	if issue := requireLeaseOwnerAndGeneration(lease, opts); issue != "" {
		return blockedLease("renew", issue, lease, stateRoot)
	}
	if effectiveLeaseStatus(lease, opts.Now) != "active" {
		return blockedLease("renew", "lease is expired; use recover", lease, stateRoot)
	}
	lease.Generation++
	refreshLeaseTimes(&lease, opts.Now, opts.TTL)
	state.Leases[key] = lease
	return sliceLeaseResult{OK: true, Action: "renew", Lease: &lease, StateRoot: stateRoot}
}

func releaseSliceLease(state sliceLeaseState, opts sliceLeaseCommandOptions, stateRoot string) sliceLeaseResult {
	key, lease, ok := findSliceLease(state, opts)
	if !ok {
		return blockedSliceLease("release", "lease not found", stateRoot)
	}
	if issue := requireLeaseOwnerAndGeneration(lease, opts); issue != "" {
		return blockedLease("release", issue, lease, stateRoot)
	}
	lease.Generation++
	lease.Status = "released"
	lease.LastUpdatedAt = opts.Now.Format(time.RFC3339Nano)
	state.Leases[key] = lease
	return sliceLeaseResult{OK: true, Action: "release", Lease: &lease, StateRoot: stateRoot}
}

func recoverSliceLease(state sliceLeaseState, opts sliceLeaseCommandOptions, stateRoot string) sliceLeaseResult {
	key, lease, ok := findSliceLease(state, opts)
	if !ok {
		return blockedSliceLease("recover", "lease not found", stateRoot)
	}
	if issue := requireLeaseOwnerAndGeneration(lease, opts); issue != "" {
		return blockedLease("recover", issue, lease, stateRoot)
	}
	if effectiveLeaseStatus(lease, opts.Now) == "active" {
		return blockedLease("recover", "lease is still active", lease, stateRoot)
	}
	lease.Generation++
	lease.Status = "active"
	refreshLeaseTimes(&lease, opts.Now, opts.TTL)
	state.Leases[key] = lease
	return sliceLeaseResult{OK: true, Action: "recover", Lease: &lease, StateRoot: stateRoot}
}

func newSliceLease(state sliceLeaseState, opts sliceLeaseCommandOptions, stateRoot string, generation int64) (sliceLease, error) {
	if strings.TrimSpace(opts.SliceID) == "" || strings.TrimSpace(opts.RunID) == "" {
		return sliceLease{}, fmt.Errorf("slice-id and run-id are required")
	}
	owner := strings.TrimSpace(opts.OwnerToken)
	if owner == "" {
		generated, err := generateOwnerToken()
		if err != nil {
			return sliceLease{}, err
		}
		owner = generated
	}
	claims, err := normalizeLeaseClaims(opts.Files, opts.Prefixes, opts.Resources)
	if err != nil {
		return sliceLease{}, err
	}
	if len(claims) == 0 {
		return sliceLease{}, fmt.Errorf("at least one --file, --prefix, or --resource claim is required")
	}
	now := opts.Now
	ttl := opts.TTL
	if ttl <= 0 {
		ttl = defaultSliceLeaseTTL
	}
	lease := sliceLease{
		SliceID:      opts.SliceID,
		RunID:        opts.RunID,
		OwnerToken:   owner,
		RepoIdentity: state.RepoIdentity,
		BaseSHA:      opts.BaseSHA,
		Worktree:     opts.Worktree,
		Branch:       opts.Branch,
		Status:       "active",
		Generation:   generation,
		Claims:       claims,
		AcquiredAt:   now.Format(time.RFC3339Nano),
		Limitations: []string{
			"coordinates only sessions sharing this Git common directory",
			"does not coordinate separate clones, machines, or remote schedulers",
		},
		Metadata: leaseMetadata{CoordinationScope: "git-common-dir"},
	}
	refreshLeaseTimes(&lease, now, ttl)
	return lease, nil
}

func refreshLeaseTimes(lease *sliceLease, now time.Time, ttl time.Duration) {
	if ttl <= 0 {
		ttl = defaultSliceLeaseTTL
	}
	lease.HeartbeatAt = now.Format(time.RFC3339Nano)
	lease.ExpiresAt = now.Add(ttl).Format(time.RFC3339Nano)
	lease.LastUpdatedAt = lease.HeartbeatAt
	lease.LeaseDuration = ttl.String()
}

func requireLeaseOwnerAndGeneration(lease sliceLease, opts sliceLeaseCommandOptions) string {
	if opts.SliceID == "" {
		return "slice-id is required"
	}
	if opts.OwnerToken == "" || opts.OwnerToken != lease.OwnerToken {
		return "owner token mismatch"
	}
	if opts.Generation <= 0 || opts.Generation != lease.Generation {
		return "generation mismatch"
	}
	return ""
}

func activeLeaseCollisions(state sliceLeaseState, contender sliceLease, now time.Time) []leaseCollision {
	collisions := []leaseCollision{}
	for _, existing := range state.Leases {
		if existing.Status != "active" || effectiveLeaseStatus(existing, now) != "active" {
			continue
		}
		for _, existingClaim := range existing.Claims {
			for _, claim := range contender.Claims {
				if leaseClaimsConflict(existingClaim, claim) {
					collisions = append(collisions, leaseCollision{
						SliceID: existing.SliceID, RunID: existing.RunID, OwnerToken: existing.OwnerToken, Claim: claim,
						Reason: fmt.Sprintf("conflicts with run %s slice %s %s claim %s", existing.RunID, existing.SliceID, existingClaim.Kind, existingClaim.Value),
					})
				}
			}
		}
	}
	return collisions
}

func effectiveLeaseStatus(lease sliceLease, now time.Time) string {
	if lease.Status != "active" {
		return lease.Status
	}
	expiresAt, err := time.Parse(time.RFC3339Nano, lease.ExpiresAt)
	if err != nil || !now.Before(expiresAt) {
		return "expired"
	}
	return "active"
}

func normalizeLeaseClaims(files, prefixes, resources []string) ([]leaseClaim, error) {
	claims := []leaseClaim{}
	addPath := func(kind, value string) error {
		normalized, err := normalizeLeaseClaimPath(value)
		if err != nil {
			return err
		}
		claims = append(claims, leaseClaim{Kind: kind, Value: normalized})
		return nil
	}
	for _, file := range files {
		if err := addPath("file", file); err != nil {
			return nil, err
		}
	}
	for _, prefix := range prefixes {
		normalized, err := normalizeLeaseClaimPath(prefix)
		if err != nil {
			return nil, err
		}
		claims = append(claims, leaseClaim{Kind: "prefix", Value: strings.TrimSuffix(normalized, "/") + "/"})
	}
	for _, resource := range resources {
		resource = strings.ToLower(strings.TrimSpace(resource))
		parts := strings.SplitN(resource, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" || strings.ContainsAny(resource, " \t\r\n") {
			return nil, fmt.Errorf("resource claims must be kind:value without whitespace")
		}
		claims = append(claims, leaseClaim{Kind: "resource", Value: resource})
	}
	sort.Slice(claims, func(i, j int) bool {
		if claims[i].Kind == claims[j].Kind {
			return claims[i].Value < claims[j].Value
		}
		return claims[i].Kind < claims[j].Kind
	})
	return claims, nil
}

func normalizeLeaseClaimPath(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || strings.ContainsAny(trimmed, "*?[]") {
		return "", fmt.Errorf("lease paths must be non-empty repository-relative paths without globs")
	}
	portable := strings.ReplaceAll(trimmed, "\\", "/")
	if len(portable) >= 2 && portable[1] == ':' {
		return "", fmt.Errorf("lease path must stay inside the repository: %s", value)
	}
	clean := filepath.Clean(filepath.FromSlash(portable))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("lease path must stay inside the repository: %s", value)
	}
	normalized := filepath.ToSlash(clean)
	return strings.ToLower(normalized), nil
}

func leaseClaimsConflict(left, right leaseClaim) bool {
	if left.Kind == "resource" || right.Kind == "resource" {
		return left.Kind == right.Kind && left.Value == right.Value
	}
	if left.Kind == "file" && right.Kind == "file" {
		return left.Value == right.Value
	}
	if left.Kind == "prefix" && right.Kind == "prefix" {
		return strings.HasPrefix(left.Value, right.Value) || strings.HasPrefix(right.Value, left.Value)
	}
	if left.Kind == "prefix" && right.Kind == "file" {
		return strings.HasPrefix(right.Value, left.Value)
	}
	if left.Kind == "file" && right.Kind == "prefix" {
		return strings.HasPrefix(left.Value, right.Value)
	}
	return false
}

func blockedLease(action, issue string, lease sliceLease, stateRoot string) sliceLeaseResult {
	lease.Status = effectiveLeaseStatus(lease, time.Now().UTC())
	return sliceLeaseResult{OK: false, Action: action, Issue: issue, Lease: &lease, StateRoot: stateRoot}
}

func blockedSliceLease(action, issue, stateRoot string) sliceLeaseResult {
	return sliceLeaseResult{OK: false, Action: action, Issue: issue, StateRoot: stateRoot}
}

func appendSliceLeaseEvent(state *sliceLeaseState, now time.Time, action string, lease *sliceLease) {
	if lease == nil {
		return
	}
	state.Events = append(state.Events, sliceLeaseStateEvent{
		At: now.Format(time.RFC3339Nano), Action: action, SliceID: lease.SliceID, RunID: lease.RunID, Generation: lease.Generation,
	})
	if len(state.Events) > 200 {
		state.Events = state.Events[len(state.Events)-200:]
	}
}

func sliceLeaseKey(runID, sliceID string) string {
	return strings.TrimSpace(runID) + "::" + strings.TrimSpace(sliceID)
}

func findSliceLease(state sliceLeaseState, opts sliceLeaseCommandOptions) (string, sliceLease, bool) {
	if opts.RunID != "" {
		key := sliceLeaseKey(opts.RunID, opts.SliceID)
		lease, ok := state.Leases[key]
		if ok {
			return key, lease, true
		}
	}
	var foundKey string
	var found sliceLease
	for key, lease := range state.Leases {
		if lease.SliceID != opts.SliceID {
			continue
		}
		if opts.RunID != "" && lease.RunID != opts.RunID {
			continue
		}
		if foundKey != "" {
			return "", sliceLease{}, false
		}
		foundKey, found = key, lease
	}
	return foundKey, found, foundKey != ""
}

func validateSlicePlanRunComposition(state planRunLeaseState, contender sliceLease, now time.Time) (string, []leaseCollision) {
	active := 0
	var owner *planRunLease
	collisions := []leaseCollision{}
	for _, run := range state.Leases {
		if effectivePlanRunLeaseStatus(run, now) != "active" {
			continue
		}
		active++
		if run.RunID == contender.RunID {
			runCopy := run
			owner = &runCopy
			continue
		}
		for _, existingClaim := range run.Claims {
			for _, claim := range contender.Claims {
				if planRunClaimsConflict(existingClaim, claim) {
					collisions = append(collisions, leaseCollision{
						RunID: run.RunID, OwnerToken: run.OwnerToken, Claim: claim,
						Reason: fmt.Sprintf("conflicts with run %s manifest %s %s claim %s", run.RunID, run.ManifestPath, existingClaim.Kind, existingClaim.Value),
					})
				}
			}
		}
	}
	if len(collisions) > 0 {
		return "active plan-run claim collision", collisions
	}
	if active == 0 {
		return "", nil
	}
	if owner == nil {
		return "active plan-run lease required before slice acquisition", nil
	}
	if owner.OwnerToken != contender.OwnerToken {
		return "plan-run owner does not match slice owner", nil
	}
	if owner.Worktree != "" {
		if !samePath(owner.Worktree, contender.Worktree) {
			return "slice lease must be acquired from the exact manifest-owned plan worktree", nil
		}
		if owner.Branch != contender.Branch {
			return "slice lease branch does not match the manifest-owned plan branch", nil
		}
		kbID, err := planRunManifestID(owner.ManifestPath)
		if err != nil {
			return "resolve manifest-owned plan receipt: " + err.Error(), nil
		}
		receipt, err := loadPlanRunWorkspaceReceipt(planRunWorkspaceOptions{RepoRoot: owner.Worktree}, kbID)
		if err != nil {
			return "load manifest-owned plan receipt: " + err.Error(), nil
		}
		expectedHead := receipt.IntegrationHead
		if expectedHead != contender.BaseSHA {
			return "slice lease base does not match the current manifest-owned integration head", nil
		}
	}
	for _, claim := range contender.Claims {
		if !planRunClaimCovers(owner.Claims, claim) {
			return fmt.Sprintf("slice claim %s:%s was not forecast; expand the plan-run lease before write", claim.Kind, claim.Value), nil
		}
	}
	return "", nil
}

func planRunClaimCovers(forecast []leaseClaim, observed leaseClaim) bool {
	for _, claim := range forecast {
		if claim.Kind == observed.Kind && claim.Value == observed.Value {
			return true
		}
		switch {
		case claim.Kind == "prefix" && observed.Kind == "file":
			if strings.HasPrefix(observed.Value, claim.Value) {
				return true
			}
		case claim.Kind == "prefix" && observed.Kind == "prefix":
			if strings.HasPrefix(observed.Value, claim.Value) {
				return true
			}
		}
	}
	return false
}

func resolveSliceLeaseStateRoot(opts sliceLeaseCommandOptions) (string, error) {
	if opts.StateRoot != "" {
		return filepath.Abs(filepath.Clean(opts.StateRoot))
	}
	commonDir, err := gitCommonDir(opts.RepoRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(commonDir, "kb", "slice-leases"), nil
}

func gitCommonDir(root string) (string, error) {
	output, err := exec.Command("git", "-C", root, "rev-parse", "--git-common-dir").Output()
	if err != nil {
		return "", fmt.Errorf("resolve git common dir: %w", err)
	}
	value := strings.TrimSpace(string(output))
	if value == "" {
		return "", fmt.Errorf("resolve git common dir: empty output")
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(root, filepath.FromSlash(value))
	}
	return filepath.Abs(filepath.Clean(value))
}

func withSliceLeaseStateLock(stateRoot string, mutate func() error) error {
	if err := os.MkdirAll(stateRoot, 0o755); err != nil {
		return err
	}
	lockDir := filepath.Join(stateRoot, ".lock")
	deadline := time.Now().Add(sliceLeaseLockTimeout)
	for {
		err := os.Mkdir(lockDir, 0o755)
		if err == nil {
			defer os.Remove(lockDir)
			return mutate()
		}
		if !errors.Is(err, os.ErrExist) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for slice lease state lock")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func loadSliceLeaseState(stateRoot string) (sliceLeaseState, error) {
	path := filepath.Join(stateRoot, "leases.json")
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return sliceLeaseState{SchemaVersion: sliceLeaseSchemaVersion, Leases: map[string]sliceLease{}}, nil
	}
	if err != nil {
		return sliceLeaseState{}, err
	}
	var state sliceLeaseState
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return sliceLeaseState{}, err
	}
	if state.SchemaVersion != sliceLeaseSchemaVersion {
		return sliceLeaseState{}, fmt.Errorf("unsupported slice lease schema_version %d", state.SchemaVersion)
	}
	if state.Leases == nil {
		state.Leases = map[string]sliceLease{}
	}
	return state, nil
}

func saveSliceLeaseState(stateRoot string, state sliceLeaseState) error {
	content, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	temp, err := os.CreateTemp(stateRoot, ".leases-*.tmp")
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
	if err := os.Rename(tempName, filepath.Join(stateRoot, "leases.json")); err != nil {
		return err
	}
	ok = true
	return nil
}

func repoIdentity(root, stateRoot string) string {
	if value := gitOutput(root, "config", "--get", "remote.origin.url"); value != "" {
		return "git:" + value
	}
	if value := gitOutput(root, "rev-parse", "--path-format=absolute", "--git-common-dir"); value != "" {
		return "git-common:file://" + filepath.ToSlash(filepath.Clean(value))
	}
	return "state-root:" + filepath.ToSlash(stateRoot)
}

func gitOutput(root string, args ...string) string {
	if strings.TrimSpace(root) == "" {
		return ""
	}
	cmdArgs := append([]string{"-C", root}, args...)
	output, err := exec.Command("git", cmdArgs...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func generateOwnerToken() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(raw[:]), nil
}

func runSliceLeaseSelftest(stdout, stderr io.Writer) int {
	temp, err := os.MkdirTemp("", "kb-slice-lease-selftest-*")
	if err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(temp)
	now := time.Now().UTC()
	acquired, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: temp, SliceID: "slice-001", RunID: "run-1", OwnerToken: "owner-1", Files: []string{"src/a.go"}, Now: now,
	})
	if err != nil || !acquired.OK || acquired.Lease == nil {
		fmt.Fprintf(stderr, "acquire failed: result=%#v err=%v\n", acquired, err)
		return 1
	}
	blocked, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "acquire", StateRoot: temp, SliceID: "slice-002", RunID: "run-2", OwnerToken: "owner-2", Files: []string{"src/a.go"}, Now: now,
	})
	if err != nil || blocked.OK {
		fmt.Fprintf(stderr, "collision failed: result=%#v err=%v\n", blocked, err)
		return 1
	}
	released, err := executeSliceLease(sliceLeaseCommandOptions{
		Action: "release", StateRoot: temp, SliceID: "slice-001", OwnerToken: "owner-1", Generation: acquired.Lease.Generation, Now: now,
	})
	if err != nil || !released.OK {
		fmt.Fprintf(stderr, "release failed: result=%#v err=%v\n", released, err)
		return 1
	}
	fmt.Fprintln(stdout, "kb-work slice-lease selftest: passed")
	return 0
}

func parseLeaseGeneration(value string) int64 {
	generation, _ := strconv.ParseInt(value, 10, 64)
	return generation
}
