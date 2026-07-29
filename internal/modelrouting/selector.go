package modelrouting

import (
	"sort"
	"strings"
	"time"
)

type WorkRequest struct {
	PlannedTier    Tier
	AttemptTier    Tier
	ExecutionOwner ExecutionOwner
	OwnerReason    string
	TierReason     string
	TaskFamily     string
	Tools          []string
	ContextSize    int
	Risk           RiskLevel
	SensitiveData  bool
	ProjectID      string
}

type ExecutionOwner string

const (
	ExecutionOwnerCurrent   ExecutionOwner = "current"
	ExecutionOwnerDelegated ExecutionOwner = "delegated"
)

type OverrideMode string

const (
	OverrideUse     OverrideMode = "use"
	OverrideRequire OverrideMode = "require"
	OverrideIgnore  OverrideMode = "ignore"
)

type RunOverride struct {
	Mode   OverrideMode
	Alias  string
	Prefer RoutePreference
}

type RoutePreference string

const (
	PreferenceAutomatic       RoutePreference = "automatic"
	PreferenceSelfHostedFirst RoutePreference = "self-hosted-first"
	PreferenceNativeFirst     RoutePreference = "native-first"
)

type SelectionStatus string

const (
	SelectionRouted      SelectionStatus = "routed"
	SelectionCurrent     SelectionStatus = "current"
	SelectionIgnored     SelectionStatus = "ignored"
	SelectionUnavailable SelectionStatus = "unavailable"
	SelectionDegraded    SelectionStatus = "degraded-current"
)

type SelectionDecision struct {
	Status         SelectionStatus
	Routes         []Route
	Current        CurrentModel
	PlannedTier    Tier
	AttemptTier    Tier
	ExecutionOwner ExecutionOwner
	OwnerReason    string
	TierReason     string
	Preference     RoutePreference
}

func SelectRoute(validated ValidatedCatalog, req WorkRequest, policy PolicyContext, override RunOverride, ledger AttemptLedger, now time.Time) (SelectionDecision, error) {
	catalog := cloneCatalog(validated.catalog)
	decision := selectionDecisionForRequest(req)
	decision.Preference = effectiveRoutePreference(override.Prefer, req.Risk)
	if override.Mode == OverrideIgnore {
		decision.ExecutionOwner = ExecutionOwnerCurrent
		decision.OwnerReason = "user-required"
		if !validWorkRequest(req) {
			decision.Status = SelectionUnavailable
			return decision, ErrInvalidWorkRequest
		}
		if !currentEligible(catalog.Current, req, policy, now) {
			decision.Status, decision.Current = SelectionUnavailable, catalog.Current
			return decision, nil
		}
		decision.Status, decision.Current = SelectionIgnored, catalog.Current
		return decision, nil
	}
	if req.ProjectID == "" {
		decision.Status = SelectionUnavailable
		return decision, ErrInvalidWorkRequest
	}
	if !validWorkRequest(req) || !validRoutePreference(override.Prefer) {
		decision.Status = SelectionUnavailable
		return decision, ErrInvalidWorkRequest
	}
	if req.ExecutionOwner == ExecutionOwnerCurrent {
		if override.Mode == OverrideUse || override.Mode == OverrideRequire {
			decision.Status = SelectionUnavailable
			return decision, ErrInvalidWorkRequest
		}
		if currentOwnerReasonCode(req.OwnerReason) == "no-qualified-route" && len(eligibleRoutes(catalog, req, policy, ledger, now)) > 0 {
			decision.Status = SelectionUnavailable
			return decision, ErrInvalidWorkRequest
		}
		if !currentEligible(catalog.Current, req, policy, now) {
			decision.Status, decision.Current = SelectionUnavailable, catalog.Current
			return decision, nil
		}
		decision.Status, decision.Current = SelectionCurrent, catalog.Current
		return decision, nil
	}
	automatic := preferEligibleRoutes(eligibleRoutes(catalog, req, policy, ledger, now), decision.Preference)
	if override.Mode == OverrideRequire {
		route, ok := explicitlyDelegatedSelectable(catalog, override.Alias, req, policy, ledger, now)
		if ok {
			decision.Status, decision.Routes = SelectionRouted, []Route{route}
			return decision, nil
		}
		decision.Status = SelectionUnavailable
		return decision, ErrRequiredRouteUnavailable
	}
	if override.Mode == OverrideUse {
		if preferred, ok := preferredSelectable(catalog, override.Alias, req, policy, ledger, now, true); ok {
			decision.Status, decision.Routes = SelectionRouted, []Route{preferred}
			return decision, nil
		}
	}
	if len(automatic) > 0 {
		decision.Status, decision.Routes = SelectionRouted, automatic[:1]
		return decision, nil
	}
	decision.Status = SelectionUnavailable
	return decision, nil
}

// explicitlyDelegatedSelectable is the bounded first-dispatch compatibility
// path. A trusted/selectable route may not have a prior dispatch receipt yet,
// but an explicit delegated owner still must satisfy the complete task, tier,
// tool, context, risk, trust, and route-binding envelope.
func explicitlyDelegatedSelectable(catalog Catalog, alias string, req WorkRequest, policy PolicyContext, ledger AttemptLedger, now time.Time) (Route, bool) {
	if alias == "" || ledger.Attempted(alias) {
		return Route{}, false
	}
	floor := tierFloor(selectionAttemptTier(req))
	for _, route := range catalog.Routes {
		if route.Alias != alias || validateRouteSchema(route) != nil ||
			!routeAllowedByPolicy(route, req, policy, now) ||
			!readinessCumulativeThrough(route.Readiness, ReadinessSelectable) ||
			!capabilityEnvelopeEligible(route, req, floor) {
			continue
		}
		return route, true
	}
	return Route{}, false
}

func validWorkRequest(req WorkRequest) bool {
	if !validTierRequest(req) {
		return false
	}
	if !validExecutionOwner(req.ExecutionOwner) || !validOwnerReason(req.ExecutionOwner, req.OwnerReason) || strings.TrimSpace(req.TierReason) == "" {
		return false
	}
	if req.ProjectID == "" || req.TaskFamily == "" || len(req.Tools) == 0 || req.ContextSize <= 0 || !validRisk(req.Risk) {
		return false
	}
	seen := make(map[string]struct{}, len(req.Tools))
	for _, tool := range req.Tools {
		if tool == "" {
			return false
		}
		if _, exists := seen[tool]; exists {
			return false
		}
		seen[tool] = struct{}{}
	}
	return true
}

func validExecutionOwner(owner ExecutionOwner) bool {
	return owner == ExecutionOwnerCurrent || owner == ExecutionOwnerDelegated
}

func validOwnerReason(owner ExecutionOwner, reason string) bool {
	if strings.TrimSpace(reason) == "" {
		return false
	}
	if owner == ExecutionOwnerDelegated {
		// Delegated ownership remains extensible: callers may record the bounded
		// worker-selection rationale that is useful to their workflow.
		return true
	}
	code, explanation, hasExplanation := strings.Cut(strings.TrimSpace(reason), ":")
	code = strings.TrimSpace(code)
	if hasExplanation && strings.TrimSpace(explanation) == "" {
		return false
	}
	switch code {
	case "reasoning-required", "context-required", "tool-required", "authority-required", "trust-required", "user-required", "no-qualified-route":
		return true
	default:
		return false
	}
}

func currentOwnerReasonCode(reason string) string {
	code, _, _ := strings.Cut(strings.TrimSpace(reason), ":")
	return strings.TrimSpace(code)
}

func validTierRequest(req WorkRequest) bool {
	if !validTier(req.PlannedTier) {
		return false
	}
	if req.AttemptTier == "" {
		return true
	}
	return (req.PlannedTier == TierMedium && req.AttemptTier == TierSmall) ||
		(req.PlannedTier == TierLarge && req.AttemptTier == TierMedium)
}

func validRoutePreference(preference RoutePreference) bool {
	return preference == "" || preference == PreferenceAutomatic || preference == PreferenceSelfHostedFirst || preference == PreferenceNativeFirst
}

func normalizedRoutePreference(preference RoutePreference) RoutePreference {
	if preference == "" {
		return PreferenceAutomatic
	}
	return preference
}

func effectiveRoutePreference(preference RoutePreference, risk RiskLevel) RoutePreference {
	preference = normalizedRoutePreference(preference)
	if risk == RiskBroad && preference == PreferenceSelfHostedFirst {
		return PreferenceAutomatic
	}
	return preference
}

func preferredSelectable(catalog Catalog, alias string, req WorkRequest, policy PolicyContext, ledger AttemptLedger, now time.Time, completeEnvelope bool) (Route, bool) {
	if !completeEnvelope || alias == "" || ledger.Attempted(alias) {
		return Route{}, false
	}
	floor := tierFloor(selectionAttemptTier(req))
	for _, route := range catalog.Routes {
		if route.Alias != alias || validateRouteSchema(route) != nil || !routeAllowedByPolicy(route, req, policy, now) || !automaticEligible(route, req, floor, now) {
			continue
		}
		return route, true
	}
	return Route{}, false
}

func eligibleRoutes(catalog Catalog, req WorkRequest, policy PolicyContext, ledger AttemptLedger, now time.Time) []Route {
	floor := tierFloor(selectionAttemptTier(req))
	floorRank := classRank(floor)
	sameClass := make([]Route, 0, len(catalog.Routes))
	higher := make([]Route, 0, len(catalog.Routes))
	for _, route := range catalog.Routes {
		if route.Capability.Class == ClassPlanner || ledger.Attempted(route.Alias) || validateRouteSchema(route) != nil || !routeAllowedByPolicy(route, req, policy, now) || !automaticEligible(route, req, floor, now) {
			continue
		}
		if classRank(route.Capability.Class) == floorRank {
			sameClass = append(sameClass, route)
		} else {
			higher = append(higher, route)
		}
	}
	sortRoutesByEvidence(sameClass, req, now)
	sortRoutesByEvidence(higher, req, now)
	if len(sameClass) > 0 {
		threshold := evidenceStrength(sameClass[0], req, now)
		qualified := higher[:0]
		for _, route := range higher {
			if evidenceStrength(route, req, now) > threshold {
				qualified = append(qualified, route)
			}
		}
		higher = qualified
	}
	return append(sameClass, higher...)
}

func preferEligibleRoutes(routes []Route, preference RoutePreference) []Route {
	if preference == "" || preference == PreferenceAutomatic || len(routes) < 2 {
		return routes
	}
	byRank := make(map[int][]Route)
	for _, route := range routes {
		rank := classRank(route.Capability.Class)
		byRank[rank] = append(byRank[rank], route)
	}
	for rank, group := range byRank {
		preferred := make([]Route, 0, len(group))
		for _, route := range group {
			if routeMatchesPreference(route, preference) {
				preferred = append(preferred, route)
			}
		}
		for _, route := range group {
			if !routeMatchesPreference(route, preference) {
				preferred = append(preferred, route)
			}
		}
		byRank[rank] = preferred
	}
	ordered := make([]Route, len(routes))
	offsets := make(map[int]int)
	for index, route := range routes {
		rank := classRank(route.Capability.Class)
		ordered[index] = byRank[rank][offsets[rank]]
		offsets[rank]++
	}
	return ordered
}

func routeMatchesPreference(route Route, preference RoutePreference) bool {
	if preference == PreferenceSelfHostedFirst {
		return route.Hosting == HostingSelfHosted
	}
	return preference == PreferenceNativeFirst && route.ManagementOrigin == OriginNative
}

func selectionDecisionForRequest(req WorkRequest) SelectionDecision {
	return SelectionDecision{
		PlannedTier: req.PlannedTier, AttemptTier: selectionAttemptTier(req),
		ExecutionOwner: req.ExecutionOwner, OwnerReason: req.OwnerReason, TierReason: req.TierReason,
	}
}

func selectionAttemptTier(req WorkRequest) Tier {
	if req.AttemptTier != "" {
		return req.AttemptTier
	}
	return req.PlannedTier
}

func validTier(tier Tier) bool {
	return tier == TierTiny || tier == TierSmall || tier == TierMedium || tier == TierLarge
}

func sortRoutesByEvidence(routes []Route, req WorkRequest, now time.Time) {
	sort.SliceStable(routes, func(i, j int) bool {
		left, right := evidenceStrength(routes[i], req, now), evidenceStrength(routes[j], req, now)
		if left != right {
			return left > right
		}
		leftRank, rightRank := classRank(routes[i].Capability.Class), classRank(routes[j].Capability.Class)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if !routes[i].Capability.ExpiresAt.Equal(routes[j].Capability.ExpiresAt) {
			return routes[i].Capability.ExpiresAt.After(routes[j].Capability.ExpiresAt)
		}
		return routes[i].Alias < routes[j].Alias
	})
}

func evidenceStrength(route Route, req WorkRequest, now time.Time) int64 {
	evidence := route.Capability
	var score int64
	switch evidence.Source {
	case EvidenceKBReceipt:
		score = 200
	case EvidenceAdapterPrior:
		score = 100
	}
	if evidence.TaskFamily == req.TaskFamily && req.TaskFamily != "" {
		score += 30
	}
	if req.ContextSize > 0 && evidence.ContextSize >= req.ContextSize {
		score += 20
	}
	if riskCovers(evidence.Risk, req.Risk) {
		score += 20
	}
	for _, tool := range req.Tools {
		if containsString(evidence.Tools, tool) {
			score += 5
		}
	}
	if evidence.ExpiresAt.After(now) {
		freshness := int64(evidence.ExpiresAt.Sub(now) / time.Hour)
		if freshness > 10 {
			freshness = 10
		}
		score += freshness
	}
	return score
}

func automaticEligible(route Route, req WorkRequest, floor CapabilityClass, now time.Time) bool {
	if !readinessCumulativeThrough(route.Readiness, ReadinessSelectable) {
		return false
	}
	evidence := route.Capability
	if route.Hosting == HostingSelfHosted && route.Adapter == "openai-compatible" &&
		route.DispatchMethod == "chat-completions" && req.Risk == RiskNormal &&
		evidence.Source == EvidenceDeclared && !evidence.DispatchProven {
		return capabilityEnvelopeEligible(route, req, floor)
	}
	if !(evidence.DispatchQualified || evidence.DispatchProven) || (evidence.Source != EvidenceKBReceipt && evidence.Source != EvidenceAdapterPrior) {
		return false
	}
	if evidence.ExpiresAt.IsZero() || !now.Before(evidence.ExpiresAt) {
		return false
	}
	if !capabilityEnvelopeEligible(route, req, floor) {
		return false
	}
	return true
}

func capabilityEnvelopeEligible(route Route, req WorkRequest, floor CapabilityClass) bool {
	evidence := route.Capability
	if evidence.RouteAlias != route.Alias || evidence.ModelID == "" || route.DisplayModelID == "" || evidence.ModelID != route.DisplayModelID {
		return false
	}
	if classRank(evidence.Class) < classRank(floor) {
		return false
	}
	if evidence.TaskFamily != req.TaskFamily {
		return false
	}
	if evidence.ContextSize <= 0 || evidence.ContextSize < req.ContextSize {
		return false
	}
	if !validRisk(evidence.Risk) || !riskCovers(evidence.Risk, req.Risk) {
		return false
	}
	for _, tool := range req.Tools {
		if !containsString(evidence.Tools, tool) {
			return false
		}
	}
	return true
}

func currentEligible(current CurrentModel, req WorkRequest, policy PolicyContext, now time.Time) bool {
	if policy.Project.DenyCurrentFallback || current.ModelID == "" || current.Route == nil || current.Route.DisplayModelID != current.ModelID {
		return false
	}
	route := *current.Route
	if validateRouteSchema(route) != nil || !routeAllowedByPolicy(route, req, policy, now) {
		return false
	}
	evidence := route.Capability
	// The current route is already executing inside the active host, so it does
	// not need delegated dispatch readiness or proof. Validate every capability
	// the host actually reports, but do not invent App-specific tool or context
	// claims merely to retain work. Unknown dimensions remain bounded by the
	// orchestrator's explicit current-owner decision.
	if evidence.Class != ClassUnknown && classRank(evidence.Class) < classRank(tierFloor(req.PlannedTier)) {
		return false
	}
	if evidence.TaskFamily != "" && evidence.TaskFamily != "unknown" && evidence.TaskFamily != req.TaskFamily {
		return false
	}
	if evidence.ContextSize > 0 && evidence.ContextSize < req.ContextSize {
		return false
	}
	if evidence.Risk != RiskUnknown && (!validRisk(evidence.Risk) || !riskCovers(evidence.Risk, req.Risk)) {
		return false
	}
	if len(evidence.Tools) > 0 {
		for _, tool := range req.Tools {
			if !containsString(evidence.Tools, tool) {
				return false
			}
		}
	}
	if evidence.Class == ClassUnknown && (evidence.TaskFamily == "" || evidence.TaskFamily == "unknown") &&
		evidence.ContextSize == 0 && evidence.Risk == RiskUnknown && len(evidence.Tools) == 0 {
		if route.TrustProvenance != "active orchestrator" || route.Destination != "current" {
			return false
		}
	}
	return true
}

func tierFloor(tier Tier) CapabilityClass {
	switch tier {
	case TierTiny, TierSmall:
		return ClassSmall
	case TierMedium:
		return ClassMedium
	case TierLarge:
		return ClassLarge
	default:
		return ClassLarge
	}
}

func classRank(class CapabilityClass) int {
	switch class {
	case ClassSmall:
		return 1
	case ClassMedium:
		return 2
	case ClassLarge, ClassPlanner:
		return 3
	default:
		return 0
	}
}

func riskCovers(evidence, requested RiskLevel) bool {
	if requested == RiskBroad {
		return evidence == RiskBroad
	}
	return evidence == RiskNormal || evidence == RiskBroad
}

func readinessCumulativeThrough(values []Readiness, target Readiness) bool {
	targetIndex := -1
	for index, value := range readinessOrder {
		if value == target {
			targetIndex = index
			break
		}
	}
	if targetIndex < 0 || len(values) <= targetIndex {
		return false
	}
	for index := 0; index <= targetIndex; index++ {
		if values[index] != readinessOrder[index] {
			return false
		}
	}
	return true
}

func hasReadiness(values []Readiness, target Readiness) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
