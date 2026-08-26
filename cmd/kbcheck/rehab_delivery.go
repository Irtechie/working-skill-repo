package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
	"github.com/Irtechie/working-skill-repo/internal/reconcile"
)

// This file decides which pairings a grant may deliver. It adds no merge
// engine, no deletion engine, and no inventory engine. It consumes the shipped
// ActionMerge predicate manifest, validates a runtime grant, and delegates.

const (
	rehabDeliveryReceiptVersion = 1

	rehabActionReportOnly    = "report-only"
	rehabActionDeliverPR     = "deliver-pr"
	rehabActionMergeEligible = "merge-eligible"

	rehabGrantAbsent = "absent"
	rehabGrantValid  = "valid"
	rehabGrantVoid   = "void"

	// rehabSyncRefusal is emitted verbatim after every granted decision. A
	// merge grant never authorizes propagation to the global install roots.
	rehabSyncRefusal = "Sync: not-authorized"
)

type rehabGrantPairing struct {
	Ref         string `json:"ref"`
	TipSHA      string `json:"tip_sha"`
	PullRequest string `json:"pull_request"`
}

type rehabGrant struct {
	SchemaVersion   int                 `json:"schema_version"`
	RunID           string              `json:"run_id"`
	Operator        string              `json:"operator"`
	IssuedAt        string              `json:"issued_at"`
	ExpiresAt       string              `json:"expires_at"`
	EvidenceCutoff  string              `json:"evidence_cutoff"`
	Caps            map[string]int      `json:"caps"`
	OwnerIdentities []string            `json:"owner_identities"`
	Pairings        []rehabGrantPairing `json:"pairings"`
}

// rehabAdapters names the reachable evidence sources. An empty field means the
// adapter is absent, which leaves its predicate unsatisfied. Absence is never a
// pass.
type rehabAdapters struct {
	Forge           string `json:"forge"`
	Checks          string `json:"checks"`
	Reviews         string `json:"reviews"`
	NativeCheckGate string `json:"native_check_gate"`
}

type rehabDeliveryInput struct {
	Root        string
	Report      workRealityReport
	Policy      workRealityPolicy
	Grant       *rehabGrant
	RunID       string
	Now         time.Time
	PlanTTL     time.Duration
	Adapters    rehabAdapters
	Ownership   map[string]string
	MergePolicy reconcile.Policy
	Execute     func(command string) error
	LockDir     string
	LockTimeout time.Duration
}

type rehabDeliveryDecision struct {
	PairingID    string   `json:"pairing_id"`
	Ref          string   `json:"ref"`
	State        string   `json:"state"`
	Action       string   `json:"action"`
	Reason       string   `json:"reason"`
	Unsatisfied  []string `json:"unsatisfied_predicates,omitempty"`
	CheckAdapter string   `json:"check_adapter,omitempty"`
	Delegate     string   `json:"delegate,omitempty"`
}

type rehabDeliveryReceipt struct {
	SchemaVersion            int                     `json:"schema_version"`
	RunID                    string                  `json:"run_id"`
	PredicateManifestVersion string                  `json:"predicate_manifest_version"`
	ConsultedPredicates      []string                `json:"consulted_predicates"`
	GrantState               string                  `json:"grant_state"`
	GrantVoidReason          string                  `json:"grant_void_reason,omitempty"`
	MergeCap                 int                     `json:"merge_cap"`
	MergeCapCeiling          int                     `json:"merge_cap_ceiling"`
	Contended                bool                    `json:"contended"`
	Decisions                []rehabDeliveryDecision `json:"decisions"`
	ExecutedCommands         []string                `json:"executed_commands"`
	ResolvedProofCommand     string                  `json:"resolved_proof_command,omitempty"`
	ProofCommandSource       string                  `json:"proof_command_source"`
	NativeGateResult         string                  `json:"native_gate_result,omitempty"`
	Sync                     string                  `json:"sync"`
}

// rehabMandatoryMergePredicates consumes the shipped ActionMerge manifest. It
// never restates the list: if policy.go gains a predicate, this returns it and
// the parity test proves this path consults it.
func rehabMandatoryMergePredicates(policy reconcile.Policy) []string {
	for _, action := range policy.Actions {
		if action.Class != reconcile.ActionMerge {
			continue
		}
		names := make([]string, 0, len(action.Mandatory))
		for _, predicate := range action.Mandatory {
			names = append(names, predicate.Name)
		}
		sort.Strings(names)
		return names
	}
	return nil
}

func rehabMergeCeiling(policy reconcile.Policy) int {
	action, allowed := rehabMergeAction(policy)
	if !allowed || !action.Allowed {
		return 0
	}
	if budget, ok := policy.RiskBudget.PerRun[reconcile.ActionMerge]; ok {
		return budget
	}
	return 0
}

func rehabMergeAction(policy reconcile.Policy) (reconcile.ActionPolicy, bool) {
	for _, action := range policy.Actions {
		if action.Class == reconcile.ActionMerge {
			return action, true
		}
	}
	return reconcile.ActionPolicy{}, false
}

func loadRehabGrant(path string) (*rehabGrant, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	grant := &rehabGrant{}
	if err := decoder.Decode(grant); err != nil {
		return nil, fmt.Errorf("parse grant: %w", err)
	}
	return grant, nil
}

// validateRehabGrant voids a grant that cannot be bound to this exact run and
// this exact observed state. A grant is a runtime input, never a stored one.
func validateRehabGrant(grant *rehabGrant, input rehabDeliveryInput) (string, string) {
	if grant == nil {
		return rehabGrantAbsent, ""
	}
	if grant.SchemaVersion != 1 {
		return rehabGrantVoid, "unsupported grant schema version"
	}
	if strings.TrimSpace(grant.Operator) == "" {
		return rehabGrantVoid, "grant names no operator identity"
	}
	if strings.TrimSpace(grant.RunID) == "" || grant.RunID != input.RunID {
		return rehabGrantVoid, "grant run id does not match this run; a replayed grant is void"
	}
	issued, err := time.Parse(time.RFC3339, grant.IssuedAt)
	if err != nil {
		return rehabGrantVoid, "grant issue time is unparseable"
	}
	expires, err := time.Parse(time.RFC3339, grant.ExpiresAt)
	if err != nil {
		return rehabGrantVoid, "grant expiry is unparseable"
	}
	if !input.Now.Before(expires) {
		return rehabGrantVoid, "grant expired at " + grant.ExpiresAt
	}
	ttl := input.PlanTTL
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	if expires.Sub(issued) > ttl {
		return rehabGrantVoid, "grant lifetime exceeds the plan TTL"
	}
	cutoff, err := time.Parse(time.RFC3339, grant.EvidenceCutoff)
	if err != nil {
		return rehabGrantVoid, "grant evidence cutoff is unparseable"
	}
	reportCutoff, err := time.Parse(time.RFC3339, input.Report.Cutoff)
	if err != nil {
		return rehabGrantVoid, "report cutoff is unparseable"
	}
	if !cutoff.Equal(reportCutoff) {
		return rehabGrantVoid, "grant evidence cutoff does not match the report it claims to authorize"
	}
	if len(grant.Pairings) == 0 {
		return rehabGrantVoid, "grant enumerates no pairing"
	}
	return rehabGrantValid, ""
}

// resolveRehabProofCommand reads the gate from the authoritative default tree.
// A candidate branch can ship any policy it likes; this never reads it, so a
// branch can never supply the command that judges it.
func resolveRehabProofCommand(root, defaultSHA string) (string, string) {
	if strings.TrimSpace(defaultSHA) == "" {
		return "", "unresolved: no authoritative default tree"
	}
	raw, err := gitCapture(root, "show", defaultSHA+":config/rehab-policy.json")
	if err != nil {
		return "", "unresolved: config/rehab-policy.json absent from the authoritative default tree"
	}
	policy := workRealityPolicy{}
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return "", "unresolved: authoritative default policy is unparseable"
	}
	if strings.TrimSpace(policy.NativeCheckGate) == "" {
		return "", "unresolved: authoritative default policy declares no native check gate"
	}
	return policy.NativeCheckGate, "authoritative-default-tree"
}

func evaluateRehabDelivery(input rehabDeliveryInput) rehabDeliveryReceipt {
	if input.Now.IsZero() {
		input.Now = time.Now().UTC()
	}
	mergePolicy := input.MergePolicy
	if len(mergePolicy.Actions) == 0 {
		mergePolicy = reconcile.DefaultPolicy()
	}

	receipt := rehabDeliveryReceipt{
		SchemaVersion:            rehabDeliveryReceiptVersion,
		RunID:                    input.RunID,
		PredicateManifestVersion: mergePolicy.PolicyVersion,
		ConsultedPredicates:      rehabMandatoryMergePredicates(mergePolicy),
		MergeCapCeiling:          rehabMergeCeiling(mergePolicy),
		Decisions:                []rehabDeliveryDecision{},
		ExecutedCommands:         []string{},
		Sync:                     rehabSyncRefusal,
	}

	state, voidReason := validateRehabGrant(input.Grant, input)
	receipt.GrantState = state
	receipt.GrantVoidReason = voidReason
	receipt.MergeCap = rehabResolveMergeCap(input.Grant, state, receipt.MergeCapCeiling)

	command, source := resolveRehabProofCommand(input.Root, input.Report.RemoteAuthority.SHA)
	receipt.ResolvedProofCommand = command
	receipt.ProofCommandSource = source

	// The gate runs once, and only the command resolved from the authoritative
	// default tree ever runs. A command a candidate branch introduced is never
	// passed to the executor.
	if command != "" && source == "authoritative-default-tree" && input.Execute != nil {
		receipt.ExecutedCommands = append(receipt.ExecutedCommands, command)
		if err := input.Execute(command); err != nil {
			receipt.NativeGateResult = "failed"
		} else {
			receipt.NativeGateResult = "passed"
		}
	}

	if lock, contended := rehabAcquireLock(input); contended {
		receipt.Contended = true
		return receipt
	} else if lock != nil {
		defer func() { _ = lock.Close() }()
	}

	enumerated := map[string]rehabGrantPairing{}
	if state == rehabGrantValid {
		for _, pairing := range input.Grant.Pairings {
			enumerated[pairing.Ref] = pairing
		}
	}

	merged := 0
	for _, pairing := range input.Report.Pairings {
		decision := rehabDecidePairing(input, receipt, pairing, enumerated, state, &merged)
		receipt.Decisions = append(receipt.Decisions, decision)
	}
	sort.Slice(receipt.Decisions, func(i, j int) bool {
		return receipt.Decisions[i].PairingID < receipt.Decisions[j].PairingID
	})
	return receipt
}

func rehabResolveMergeCap(grant *rehabGrant, state string, ceiling int) int {
	if state != rehabGrantValid || grant == nil {
		return 0
	}
	requested, ok := grant.Caps["merge"]
	if !ok || requested <= 0 {
		return 0
	}
	if requested > ceiling {
		return ceiling
	}
	return requested
}

func rehabAcquireLock(input rehabDeliveryInput) (*modelrouting.PrivateStateLock, bool) {
	if strings.TrimSpace(input.LockDir) == "" {
		return nil, false
	}
	timeout := input.LockTimeout
	if timeout <= 0 {
		timeout = 2 * time.Second
	}
	lock, err := modelrouting.AcquireSharedProjectLock(input.LockDir, "work-queue.lock", timeout)
	if err != nil {
		return nil, true
	}
	return lock, false
}

// rehabDecidePairing filters by state first, then by attribution, then by
// predicate. Nothing reaches a merge decision without surviving every earlier
// filter, and a protected path stops at an open PR under any grant.
func rehabDecidePairing(
	input rehabDeliveryInput,
	receipt rehabDeliveryReceipt,
	pairing workRealityPairing,
	enumerated map[string]rehabGrantPairing,
	grantState string,
	merged *int,
) rehabDeliveryDecision {
	decision := rehabDeliveryDecision{
		PairingID: pairing.ID,
		Ref:       pairing.Ref,
		State:     pairing.State,
		Action:    rehabActionReportOnly,
	}

	if pairing.State != workRealityStateUnshipped {
		decision.Reason = "state " + pairing.State + " is never delivered under any grant"
		return decision
	}
	if pairing.DeclaredID == "" {
		decision.Reason = "no present, parsed, attributable declared work item"
		return decision
	}

	identity := input.Ownership[pairing.Ref]
	if !rehabIdentityAuthorized(identity, input.Grant, grantState) {
		decision.Reason = "tip identity " + rehabIdentityLabel(identity) + " is not named by the grant; reported only"
		return decision
	}

	unsatisfied := rehabUnsatisfiedPredicates(input, receipt, pairing)
	decision.Unsatisfied = unsatisfied
	decision.CheckAdapter = rehabCheckAdapter(input, receipt)
	decision.Delegate = "kb-complete"
	decision.Action = rehabActionDeliverPR

	if len(unsatisfied) > 0 {
		decision.Reason = "delivered to PR only: unsatisfied mandatory predicates"
		return decision
	}
	if len(pairing.ProtectedPaths) > 0 {
		decision.Reason = "delivered to PR only: touches protected paths " +
			strings.Join(pairing.ProtectedPaths, ", ") + "; never auto-merge-eligible under any grant"
		return decision
	}
	if grantState != rehabGrantValid {
		decision.Reason = "delivered to PR only: grant is " + grantState
		return decision
	}
	granted, listed := enumerated[pairing.Ref]
	if !listed {
		decision.Reason = "delivered to PR only: the grant does not enumerate this ref"
		return decision
	}
	if granted.TipSHA == "" || granted.TipSHA != pairing.SHA {
		decision.Reason = "delivered to PR only: the enumerated tip SHA no longer matches the observed tip"
		return decision
	}
	if *merged >= receipt.MergeCap {
		decision.Reason = "delivered to PR only: merge cap " + fmt.Sprint(receipt.MergeCap) + " reached"
		return decision
	}

	*merged++
	decision.Action = rehabActionMergeEligible
	decision.Reason = "every mandatory predicate holds, the grant enumerates this ref and tip, and no protected path is touched"
	return decision
}

func rehabIdentityLabel(identity string) string {
	if strings.TrimSpace(identity) == "" {
		return "<unknown>"
	}
	return identity
}

func rehabIdentityAuthorized(identity string, grant *rehabGrant, state string) bool {
	if state != rehabGrantValid || grant == nil {
		return false
	}
	identity = strings.TrimSpace(identity)
	if identity == "" {
		return false
	}
	for _, owner := range grant.OwnerIdentities {
		if strings.EqualFold(strings.TrimSpace(owner), identity) {
			return true
		}
	}
	return false
}

// rehabUnsatisfiedPredicates evaluates every mandatory ActionMerge predicate.
// Each predicate needs a reachable, authoritative adapter; an absent adapter
// leaves its predicate unsatisfied.
func rehabUnsatisfiedPredicates(input rehabDeliveryInput, receipt rehabDeliveryReceipt, pairing workRealityPairing) []string {
	unsatisfied := []string{}
	for _, name := range receipt.ConsultedPredicates {
		if rehabPredicateHolds(name, input, receipt, pairing) {
			continue
		}
		unsatisfied = append(unsatisfied, name)
	}
	sort.Strings(unsatisfied)
	return unsatisfied
}

func rehabPredicateHolds(name string, input rehabDeliveryInput, receipt rehabDeliveryReceipt, pairing workRealityPairing) bool {
	switch name {
	case "explicit-merge-authority":
		return receipt.GrantState == rehabGrantValid && receipt.MergeCap > 0
	case "exact-pr-head-base", "fresh-mergeability":
		return strings.TrimSpace(input.Adapters.Forge) != ""
	case "required-reviews-green":
		return strings.TrimSpace(input.Adapters.Reviews) != ""
	case "required-checks-green":
		return receipt.NativeGateResult != "failed" && rehabCheckAdapter(input, receipt) != ""
	case "remote-default-authority":
		return input.Report.RemoteAuthority.State == "authoritative"
	case "not-post-cutoff":
		return input.Report.Status == workRealityStatusOK
	case "exact-final-tree":
		return strings.TrimSpace(pairing.SHA) != ""
	default:
		// An unknown predicate is an unconsulted predicate. Treat it as
		// unsatisfied so a new mandatory predicate in policy.go blocks rather
		// than silently passing.
		return false
	}
}

// rehabCheckAdapter returns the adapter that may satisfy required-checks-green.
// A repository-native gate counts only when it was resolved from the
// authoritative default tree.
func rehabCheckAdapter(input rehabDeliveryInput, receipt rehabDeliveryReceipt) string {
	if forge := strings.TrimSpace(input.Adapters.Checks); forge != "" {
		return forge
	}
	if receipt.ProofCommandSource != "authoritative-default-tree" {
		return ""
	}
	if strings.TrimSpace(receipt.ResolvedProofCommand) == "" {
		return ""
	}
	return "native:" + receipt.ResolvedProofCommand
}

func writeRehabDeliveryReceipt(path string, receipt rehabDeliveryReceipt) error {
	encoded, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o600)
}
