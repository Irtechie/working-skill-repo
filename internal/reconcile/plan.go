package reconcile

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type Plan struct {
	SchemaVersion     int               `json:"schema_version"`
	PolicyVersion     string            `json:"policy_version"`
	LedgerFingerprint string            `json:"ledger_fingerprint"`
	Cutoff            time.Time         `json:"cutoff"`
	GeneratedAt       time.Time         `json:"generated_at"`
	ExpiresAt         time.Time         `json:"expires_at"`
	Actor             string            `json:"actor"`
	SessionID         string            `json:"session_id,omitempty"`
	Outcomes          []Outcome         `json:"outcomes"`
	Actions           []PlannedAction   `json:"actions"`
	DecisionPacket    DecisionPacket    `json:"decision_packet"`
	Health            ConvergenceHealth `json:"convergence_health"`
	Limitations       []string          `json:"limitations,omitempty"`
}

type Outcome struct {
	ArtifactID        string   `json:"artifact_id"`
	RepositoryID      string   `json:"repository_id"`
	Classification    string   `json:"classification"`
	ActionClass       string   `json:"action_class,omitempty"`
	Confidence        float64  `json:"confidence"`
	Evidence          []string `json:"evidence,omitempty"`
	MissingPredicates []string `json:"missing_predicates,omitempty"`
	ProtectionReasons []string `json:"protection_reasons,omitempty"`
	DeferredReason    string   `json:"deferred_reason,omitempty"`
	DecisionID        string   `json:"decision_id,omitempty"`
	ResumeSensor      string   `json:"resume_sensor,omitempty"`
}

type PlannedAction struct {
	ID              string              `json:"id"`
	ActionClass     string              `json:"action_class"`
	Classification  string              `json:"classification"`
	RepositoryID    string              `json:"repository_id"`
	ArtifactIDs     []string            `json:"artifact_ids"`
	Confidence      float64             `json:"confidence"`
	Preconditions   []PredicateEvidence `json:"preconditions"`
	DedupProofs     []DedupProof        `json:"dedup_proofs,omitempty"`
	Cutoff          time.Time           `json:"cutoff"`
	ExpiresAt       time.Time           `json:"expires_at"`
	MutationAllowed bool                `json:"mutation_allowed"`
}

type DecisionPacket struct {
	MaxGroups int            `json:"max_groups"`
	Items     []DecisionItem `json:"items"`
}

type DecisionItem struct {
	ID                      string   `json:"id"`
	Reason                  string   `json:"reason"`
	ArtifactIDs             []string `json:"artifact_ids"`
	RecommendedChoice       string   `json:"recommended_choice"`
	EvidenceAndUncertainty  string   `json:"evidence_and_uncertainty"`
	IrreversibleConsequence string   `json:"irreversible_consequence"`
	SafeDefault             string   `json:"safe_default"`
	ExpiryRecheckSensor     string   `json:"expiry_recheck_sensor"`
}

type ConvergenceHealth struct {
	EligibleTotal     int             `json:"eligible_total"`
	DeferredByAction  []DeferredCount `json:"deferred_by_action"`
	OldestDeferredAge string          `json:"oldest_deferred_age,omitempty"`
	NetBacklogChange  string          `json:"net_backlog_change"`
	RunsToConvergence int             `json:"runs_to_convergence"`
	ProjectionBasis   string          `json:"projection_basis"`
}

type DeferredCount struct {
	ActionClass string `json:"action_class"`
	Count       int    `json:"count"`
}

func BuildPlan(ledger Ledger, policy Policy) (Plan, error) {
	if ledger.SchemaVersion != LedgerSchemaVersion {
		return Plan{}, fmt.Errorf("unsupported ledger schema %d", ledger.SchemaVersion)
	}
	if err := ValidatePolicy(policy); err != nil {
		return Plan{}, err
	}
	ledger = normalizeLedger(ledger)
	fingerprint, err := FingerprintLedger(ledger)
	if err != nil {
		return Plan{}, err
	}
	ttl, _ := time.ParseDuration(policy.PlanTTL)
	plan := Plan{
		SchemaVersion: PlanSchemaVersion, PolicyVersion: policy.PolicyVersion,
		LedgerFingerprint: fingerprint, Cutoff: ledger.Cutoff,
		GeneratedAt: ledger.GeneratedAt, ExpiresAt: ledger.GeneratedAt.Add(ttl),
		Actor: ledger.Actor, SessionID: ledger.SessionID,
		DecisionPacket: DecisionPacket{MaxGroups: policy.DecisionPacket.MaxGroups},
		Health: ConvergenceHealth{
			NetBacklogChange: "unavailable:no-previous-verified-run",
			ProjectionBasis:  "configured action caps; no authority or budget is increased",
		},
		Limitations: append([]string(nil), ledger.Limitations...),
	}

	runUsed := map[string]int{}
	repoUsed := map[string]map[string]int{}
	eligibleByAction := map[string]int{}
	deferredByAction := map[string]int{}
	var oldestDeferred time.Time
	ambiguities := map[string][]int{}

	for _, repository := range ledger.Repositories {
		if repoUsed[repository.ID] == nil {
			repoUsed[repository.ID] = map[string]int{}
		}
		for _, artifact := range repository.Artifacts {
			outcome := classifyArtifact(artifact, ledger.Cutoff, policy)
			outcome.RepositoryID = repository.ID
			plan.Outcomes = append(plan.Outcomes, outcome)
			outcomeIndex := len(plan.Outcomes) - 1

			if outcome.Classification == ClassificationHumanRequired {
				ambiguities[artifact.Ambiguity] = append(ambiguities[artifact.Ambiguity], outcomeIndex)
				continue
			}
			if outcome.ActionClass == "" ||
				(outcome.Classification != ClassificationRoutineRetire &&
					outcome.Classification != ClassificationSafeSupersede &&
					outcome.Classification != ClassificationSalvage) {
				continue
			}
			eligibleByAction[outcome.ActionClass]++
			runCap := policy.RiskBudget.PerRun[outcome.ActionClass]
			repoCap := policy.RiskBudget.PerRepository[outcome.ActionClass]
			actionPolicy, _ := policyForAction(policy, outcome.ActionClass)
			if !actionPolicy.Allowed || runCap == 0 || repoCap == 0 ||
				runUsed[outcome.ActionClass] >= runCap ||
				repoUsed[repository.ID][outcome.ActionClass] >= repoCap {
				plan.Outcomes[outcomeIndex].DeferredReason = "risk-budget-exhausted-or-action-disabled"
				deferredByAction[outcome.ActionClass]++
				if oldestDeferred.IsZero() || artifact.ObservedAt.Before(oldestDeferred) {
					oldestDeferred = artifact.ObservedAt
				}
				continue
			}
			runUsed[outcome.ActionClass]++
			repoUsed[repository.ID][outcome.ActionClass]++
			plan.Actions = append(plan.Actions, PlannedAction{
				ID:          "action:" + outcome.ActionClass + ":" + artifact.ID,
				ActionClass: outcome.ActionClass, Classification: outcome.Classification,
				RepositoryID: repository.ID, ArtifactIDs: []string{artifact.ID},
				Confidence: outcome.Confidence, Preconditions: append([]PredicateEvidence(nil), artifact.Predicates...),
				DedupProofs: append([]DedupProof(nil), artifact.DedupProofs...),
				Cutoff:      ledger.Cutoff, ExpiresAt: plan.ExpiresAt,
				MutationAllowed: false,
			})
		}
	}

	plan.DecisionPacket.Items = buildDecisionItems(plan.Outcomes, ambiguities, policy.DecisionPacket.MaxGroups)
	plan.Health.EligibleTotal = 0
	for _, count := range eligibleByAction {
		plan.Health.EligibleTotal += count
	}
	actionNames := make([]string, 0, len(deferredByAction))
	for action := range deferredByAction {
		actionNames = append(actionNames, action)
	}
	sort.Strings(actionNames)
	for _, action := range actionNames {
		plan.Health.DeferredByAction = append(plan.Health.DeferredByAction, DeferredCount{
			ActionClass: action, Count: deferredByAction[action],
		})
	}
	if !oldestDeferred.IsZero() {
		age := ledger.GeneratedAt.Sub(oldestDeferred)
		if age < 0 {
			age = 0
		}
		plan.Health.OldestDeferredAge = age.Round(time.Second).String()
	}
	for action, total := range eligibleByAction {
		cap := policy.RiskBudget.PerRun[action]
		if perRepo := policy.RiskBudget.PerRepository[action]; perRepo < cap || cap == 0 {
			cap = perRepo
		}
		if cap <= 0 {
			continue
		}
		runs := int(math.Ceil(float64(total) / float64(cap)))
		if runs > plan.Health.RunsToConvergence {
			plan.Health.RunsToConvergence = runs
		}
	}
	sort.Slice(plan.Outcomes, func(i, j int) bool {
		return plan.Outcomes[i].ArtifactID < plan.Outcomes[j].ArtifactID
	})
	sort.Slice(plan.Actions, func(i, j int) bool { return plan.Actions[i].ID < plan.Actions[j].ID })
	sort.Strings(plan.Limitations)
	return plan, nil
}

func classifyArtifact(artifact Artifact, cutoff time.Time, policy Policy) Outcome {
	outcome := Outcome{
		ArtifactID: artifact.ID, ProtectionReasons: append([]string(nil), artifact.ProtectionReasons...),
		ResumeSensor: "re-inventory artifact from authoritative adapters after the recorded cutoff",
	}
	sort.Strings(outcome.ProtectionReasons)
	if hasAnyReason(artifact.ProtectionReasons, "current", "primary", "active", "active-claim", "live-process") {
		outcome.Classification = ClassificationPreserveActive
		outcome.Evidence = []string{"non-overridable active/current protection"}
		return outcome
	}
	if len(artifact.ProtectionReasons) > 0 || artifact.UpdatedAt.After(cutoff) {
		outcome.Classification = ClassificationProtected
		if artifact.UpdatedAt.After(cutoff) && !containsString(outcome.ProtectionReasons, "post-cutoff") {
			outcome.ProtectionReasons = append(outcome.ProtectionReasons, "post-cutoff")
			sort.Strings(outcome.ProtectionReasons)
		}
		outcome.Evidence = []string{"non-overridable protection predicate"}
		return outcome
	}
	if artifact.UniqueWork {
		if !artifact.SalvageSafe {
			outcome.Classification = ClassificationQuarantine
			outcome.Evidence = []string{"unique work is not safe for additive salvage"}
			return outcome
		}
		confidence, missing := confidenceForAction(artifact, cutoff, policy, ActionSalvage)
		outcome.Confidence, outcome.MissingPredicates = confidence, missing
		if confidence >= policy.Thresholds.AdditiveSalvage && len(missing) == 0 {
			outcome.Classification, outcome.ActionClass = ClassificationSalvage, ActionSalvage
			outcome.Evidence = []string{"unique coherent work; additive salvage predicates complete"}
		} else {
			outcome.Classification = ClassificationQuarantine
		}
		return outcome
	}
	if artifact.Ambiguity != "" {
		outcome.Classification = ClassificationHumanRequired
		outcome.Evidence = []string{"irreversible decision lacks authoritative evidence: " + artifact.Ambiguity}
		return outcome
	}
	actionClass := ActionLocalRefRetire
	classification := ClassificationSafeSupersede
	if artifact.Kind == ArtifactWorktree {
		actionClass = ActionWorktreeRetire
		classification = ClassificationRoutineRetire
	}
	if !hasExactDedupProof(artifact.DedupProofs, policy.DedupAlgorithms) {
		outcome.Classification = ClassificationQuarantine
		outcome.Evidence = []string{"exact deterministic containment proof unavailable"}
		return outcome
	}
	confidence, missing := confidenceForAction(artifact, cutoff, policy, actionClass)
	outcome.Confidence, outcome.MissingPredicates = confidence, missing
	actionPolicy, ok := policyForAction(policy, actionClass)
	if !ok || len(missing) > 0 || confidence < actionPolicy.Threshold {
		outcome.Classification = ClassificationQuarantine
		outcome.Evidence = []string{"mandatory action predicates incomplete, stale, or non-authoritative"}
		return outcome
	}
	outcome.Classification, outcome.ActionClass = classification, actionClass
	outcome.Evidence = []string{"exact deterministic containment and all mandatory predicates"}
	return outcome
}

func confidenceForAction(artifact Artifact, cutoff time.Time, policy Policy, actionClass string) (float64, []string) {
	action, ok := policyForAction(policy, actionClass)
	if !ok {
		return 0, []string{"policy:" + actionClass}
	}
	evidence := map[string]PredicateEvidence{}
	for _, predicate := range artifact.Predicates {
		evidence[predicate.Name] = predicate
	}
	var missing []string
	for _, required := range action.Mandatory {
		item, exists := evidence[required.Name]
		freshness, _ := time.ParseDuration(required.Freshness)
		if !exists || item.State != PredicatePass || !item.Authoritative ||
			item.ObservedAt.After(cutoff) || cutoff.Sub(item.ObservedAt) > freshness {
			missing = append(missing, required.Name)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		completeness := float64(len(action.Mandatory)-len(missing)) / float64(len(action.Mandatory))
		if completeness < 0 {
			completeness = 0
		}
		return math.Floor(completeness*100) / 100, missing
	}
	confidence := 1.0
	for _, optional := range action.Optional {
		if item, exists := evidence[optional.Name]; !exists || item.State != PredicatePass {
			confidence -= .01
		}
	}
	if confidence < 0 {
		confidence = 0
	}
	return confidence, nil
}

func hasExactDedupProof(proofs []DedupProof, allowed []string) bool {
	allowedSet := map[string]bool{}
	for _, algorithm := range allowed {
		allowedSet[algorithm] = true
	}
	for _, proof := range proofs {
		if allowedSet[proof.Algorithm] && proof.Authoritative && proof.Identity != "" {
			return true
		}
	}
	return false
}

func buildDecisionItems(outcomes []Outcome, groups map[string][]int, cap int) []DecisionItem {
	reasons := make([]string, 0, len(groups))
	for reason := range groups {
		reasons = append(reasons, reason)
	}
	sort.Strings(reasons)
	var items []DecisionItem
	for index, reason := range reasons {
		indices := groups[reason]
		if index >= cap {
			for _, outcomeIndex := range indices {
				outcomes[outcomeIndex].Classification = ClassificationQuarantine
			}
			continue
		}
		id := fmt.Sprintf("decision-%02d", index+1)
		artifactIDs := make([]string, 0, len(indices))
		for _, outcomeIndex := range indices {
			outcomes[outcomeIndex].DecisionID = id
			artifactIDs = append(artifactIDs, outcomes[outcomeIndex].ArtifactID)
		}
		sort.Strings(artifactIDs)
		items = append(items, DecisionItem{
			ID: id, Reason: reason, ArtifactIDs: artifactIDs,
			RecommendedChoice:       "preserve until authoritative evidence is available",
			EvidenceAndUncertainty:  "available evidence cannot prove " + reason,
			IrreversibleConsequence: "retirement could discard unique or unintegrated work",
			SafeDefault:             "quarantine",
			ExpiryRecheckSensor:     "refresh the named provider/host/Git authority and rebuild the cutoff-bound plan",
		})
	}
	return items
}

func hasAnyReason(reasons []string, needles ...string) bool {
	for _, reason := range reasons {
		for _, needle := range needles {
			if reason == needle || strings.HasPrefix(reason, needle+"-") {
				return true
			}
		}
	}
	return false
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
