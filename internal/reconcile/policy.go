package reconcile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const DefaultPolicyVersion = "reconcile-predicates/v1"

type Policy struct {
	SchemaVersion   int            `json:"schema_version"`
	PolicyVersion   string         `json:"policy_version"`
	PlanTTL         string         `json:"plan_ttl"`
	Thresholds      Thresholds     `json:"thresholds"`
	DecisionPacket  DecisionPolicy `json:"decision_packet"`
	RiskBudget      RiskBudget     `json:"risk_budget"`
	ProtectedPaths  []string       `json:"protected_path_classes"`
	DedupAlgorithms []string       `json:"dedup_proof_algorithms"`
	Actions         []ActionPolicy `json:"actions"`
}

type Thresholds struct {
	Destructive      float64 `json:"destructive"`
	Integration      float64 `json:"integration"`
	ReversibleRemote float64 `json:"reversible_remote"`
	AdditiveSalvage  float64 `json:"additive_salvage"`
}

type DecisionPolicy struct {
	MaxGroups int `json:"max_groups"`
}

type RiskBudget struct {
	PerRun        map[string]int `json:"per_run"`
	PerRepository map[string]int `json:"per_repository"`
}

type ActionPolicy struct {
	Class       string                `json:"class"`
	Threshold   float64               `json:"threshold"`
	Destructive bool                  `json:"destructive"`
	Allowed     bool                  `json:"allowed"`
	Downgrade   string                `json:"downgrade"`
	Mandatory   []PredicateDefinition `json:"mandatory_predicates"`
	Optional    []PredicateDefinition `json:"optional_enrichment,omitempty"`
}

type PredicateDefinition struct {
	Name      string   `json:"name"`
	Adapters  []string `json:"authoritative_adapters"`
	Freshness string   `json:"freshness"`
	Downgrade string   `json:"downgrade"`
}

func DefaultPolicy() Policy {
	worktreePredicates := []PredicateDefinition{
		predicateForName("different-executor", "5m"),
		predicateForName("clean-tracked", "5m"),
		predicateForName("clean-untracked", "5m"),
		predicateForName("clean-ignored", "5m"),
		predicateForName("terminal-or-suspended-claim", "5m"),
		predicateForName("exact-worktree-generation", "5m"),
		predicateForName("git-admin-round-trip", "5m"),
		predicateForName("durable-endpoint", "5m"),
		predicateForName("not-current", "5m"),
		predicateForName("not-primary", "5m"),
		predicateForName("not-default", "5m"),
		predicateForName("not-locked", "5m"),
		predicateForName("not-moved", "5m"),
		predicateForName("not-post-cutoff", "5m"),
		predicateForName("remote-monotonic", "5m"),
		predicateForName("non-force-only", "24h"),
		predicateForName("empty-residual-only", "24h"),
	}
	return Policy{
		SchemaVersion: 1,
		PolicyVersion: DefaultPolicyVersion,
		PlanTTL:       "10m",
		Thresholds: Thresholds{
			Destructive: 1, Integration: 1, ReversibleRemote: .98, AdditiveSalvage: .90,
		},
		DecisionPacket: DecisionPolicy{MaxGroups: 5},
		RiskBudget: RiskBudget{
			PerRun: map[string]int{
				ActionMerge: 0, ActionPRClose: 0, ActionLocalRefRetire: 5,
				ActionRemoteRefRetire: 0, ActionWorktreeRetire: 5,
				ActionSessionRetire: 0, ActionSalvage: 2,
			},
			PerRepository: map[string]int{
				ActionMerge: 0, ActionPRClose: 0, ActionLocalRefRetire: 3,
				ActionRemoteRefRetire: 0, ActionWorktreeRetire: 3,
				ActionSessionRetire: 0, ActionSalvage: 1,
			},
		},
		ProtectedPaths: []string{
			"credential", "model-runtime", "learning-memory", "database",
			"socket-lock", "ignored-live-state",
		},
		DedupAlgorithms: []string{
			DedupIdenticalBlob, DedupIdenticalTree, DedupSameBasePatch,
			DedupCommitAncestry, DedupRemoteTopicContainment, DedupRemoteDefaultAncestry,
			DedupProviderMergeTree, DedupProviderMergePatch,
		},
		Actions: []ActionPolicy{
			actionPolicy(ActionMerge, 1, true, false, ClassificationQuarantine,
				"explicit-merge-authority", "exact-pr-head-base", "required-checks-green",
				"required-reviews-green", "fresh-mergeability", "remote-default-authority",
				"not-post-cutoff", "exact-final-tree"),
			actionPolicy(ActionPRClose, .98, false, false, ClassificationQuarantine,
				"exact-containment", "no-unresolved-provider-blocker", "reopen-supported",
				"explicit-close-policy", "not-post-cutoff"),
			actionPolicy(ActionLocalRefRetire, 1, true, true, ClassificationProtected,
				"integrated-endpoint", "exact-ref-sha", "not-default", "not-checked-out",
				"provider-merge-identity-for-rewrite", "exact-tree-or-same-base-patch",
				"not-post-cutoff"),
			actionPolicy(ActionRemoteRefRetire, 1, true, false, ClassificationProtected,
				"provider-exact-ref-cas", "explicit-remote-ref-policy", "not-post-cutoff"),
			{
				Class: ActionWorktreeRetire, Threshold: 1, Destructive: true, Allowed: true,
				Downgrade: ClassificationQuarantine, Mandatory: worktreePredicates,
			},
			actionPolicy(ActionSessionRetire, 1, true, false, ClassificationProtected,
				"host-record-cas", "no-active-process", "durable-resume-or-delivery",
				"physical-state-independent", "retention-policy", "not-post-cutoff"),
			actionPolicy(ActionSalvage, .90, false, true, ClassificationQuarantine,
				"unique-work", "provable-base-and-scope", "protected-path-check-pass",
				"additive-only", "salvage-authority"),
		},
	}
}

func predicateForName(name, freshness string) PredicateDefinition {
	return PredicateDefinition{
		Name: name, Adapters: adaptersForPredicate(name), Freshness: freshness,
		Downgrade: ClassificationQuarantine,
	}
}

func actionPolicy(class string, threshold float64, destructive, allowed bool, downgrade string, names ...string) ActionPolicy {
	predicates := make([]PredicateDefinition, 0, len(names))
	for _, name := range names {
		predicate := predicateForName(name, "5m")
		predicate.Downgrade = downgrade
		predicates = append(predicates, predicate)
	}
	return ActionPolicy{
		Class: class, Threshold: threshold, Destructive: destructive,
		Allowed: allowed, Downgrade: downgrade, Mandatory: predicates,
	}
}

func adaptersForPredicate(name string) []string {
	switch name {
	case "explicit-merge-authority", "explicit-close-policy", "explicit-remote-ref-policy",
		"retention-policy", "additive-only", "salvage-authority",
		"non-force-only", "empty-residual-only":
		return []string{"policy"}
	case "exact-pr-head-base", "required-checks-green", "required-reviews-green",
		"fresh-mergeability", "remote-default-authority",
		"provider-merge-identity-for-rewrite", "provider-exact-ref-cas",
		"no-unresolved-provider-blocker", "reopen-supported":
		return []string{"provider"}
	case "host-record-cas", "no-active-process", "different-executor":
		return []string{"host-session"}
	case "not-post-cutoff":
		return []string{"ledger"}
	case "exact-containment", "integrated-endpoint", "durable-endpoint", "not-default":
		return []string{"git", "provider"}
	case "not-current":
		return []string{"git", "host-session"}
	case "remote-monotonic":
		return []string{"provider", "git-fetch"}
	case "durable-resume-or-delivery":
		return []string{"host-session", "provider"}
	case "physical-state-independent":
		return []string{"git", "host-session"}
	case "terminal-or-suspended-claim":
		return []string{"queue"}
	case "exact-worktree-generation", "git-admin-round-trip", "not-moved":
		return []string{"git-admin"}
	case "protected-path-check-pass":
		return []string{"metadata-classifier"}
	default:
		return []string{"git"}
	}
}

func LoadPolicy(path string) (Policy, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var policy Policy
	if err := decoder.Decode(&policy); err != nil {
		return Policy{}, fmt.Errorf("parse policy: %w", err)
	}
	if err := ValidatePolicy(policy); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func ValidatePolicy(policy Policy) error {
	if policy.SchemaVersion != 1 || policy.PolicyVersion == "" {
		return fmt.Errorf("unsupported or missing policy version")
	}
	if policy.Thresholds.Destructive < 1 || policy.Thresholds.Integration < 1 {
		return fmt.Errorf("destructive and integration thresholds cannot be lower than 1.00")
	}
	if policy.Thresholds.ReversibleRemote < .98 || policy.Thresholds.AdditiveSalvage < .90 {
		return fmt.Errorf("policy thresholds weaken fail-closed minimums")
	}
	if policy.DecisionPacket.MaxGroups < 1 || policy.DecisionPacket.MaxGroups > 5 {
		return fmt.Errorf("decision packet max_groups must be between 1 and 5")
	}
	if _, err := time.ParseDuration(policy.PlanTTL); err != nil {
		return fmt.Errorf("invalid plan_ttl: %w", err)
	}
	seen := map[string]bool{}
	for _, action := range policy.Actions {
		if action.Class == "" || seen[action.Class] {
			return fmt.Errorf("missing or duplicate action class %q", action.Class)
		}
		seen[action.Class] = true
		if action.Destructive && action.Threshold < 1 {
			return fmt.Errorf("%s destructive threshold cannot be lower than 1.00", action.Class)
		}
		for _, item := range action.Mandatory {
			if item.Name == "" || len(item.Adapters) == 0 || item.Downgrade == "" {
				return fmt.Errorf("%s has incomplete mandatory predicate", action.Class)
			}
			if _, err := time.ParseDuration(item.Freshness); err != nil {
				return fmt.Errorf("%s predicate %s has invalid freshness", action.Class, item.Name)
			}
		}
	}
	for _, required := range []string{
		ActionMerge, ActionPRClose, ActionLocalRefRetire, ActionRemoteRefRetire,
		ActionWorktreeRetire, ActionSessionRetire, ActionSalvage,
	} {
		if !seen[required] {
			return fmt.Errorf("policy is missing action class %s", required)
		}
	}
	algorithms := map[string]bool{}
	for _, algorithm := range policy.DedupAlgorithms {
		algorithms[algorithm] = true
	}
	for _, required := range []string{
		DedupIdenticalBlob, DedupIdenticalTree, DedupSameBasePatch,
		DedupCommitAncestry, DedupRemoteTopicContainment, DedupRemoteDefaultAncestry,
		DedupProviderMergeTree, DedupProviderMergePatch,
	} {
		if !algorithms[required] {
			return fmt.Errorf("policy is missing exact dedup algorithm %s", required)
		}
	}
	for action, cap := range policy.RiskBudget.PerRun {
		if cap < 0 {
			return fmt.Errorf("negative per-run budget for %s", action)
		}
	}
	for action, cap := range policy.RiskBudget.PerRepository {
		if cap < 0 {
			return fmt.Errorf("negative per-repository budget for %s", action)
		}
	}
	return nil
}

func policyForAction(policy Policy, class string) (ActionPolicy, bool) {
	for _, action := range policy.Actions {
		if action.Class == class {
			return action, true
		}
	}
	return ActionPolicy{}, false
}
