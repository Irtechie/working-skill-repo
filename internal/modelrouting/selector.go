package modelrouting

import (
	"sort"
	"time"
)

type WorkRequest struct {
	PlannedTier   Tier
	TaskFamily    string
	Tools         []string
	ContextSize   int
	Risk          RiskLevel
	SensitiveData bool
	ProjectID     string
}

type OverrideMode string

const (
	OverrideUse     OverrideMode = "use"
	OverrideRequire OverrideMode = "require"
	OverrideIgnore  OverrideMode = "ignore"
)

type RunOverride struct {
	Mode  OverrideMode
	Alias string
}

type SelectionStatus string

const (
	SelectionRouted      SelectionStatus = "routed"
	SelectionIgnored     SelectionStatus = "ignored"
	SelectionUnavailable SelectionStatus = "unavailable"
	SelectionDegraded    SelectionStatus = "degraded-current"
)

type SelectionDecision struct {
	Status  SelectionStatus
	Routes  []Route
	Current CurrentModel
}

func SelectRoute(validated ValidatedCatalog, req WorkRequest, policy PolicyContext, override RunOverride, ledger AttemptLedger, now time.Time) (SelectionDecision, error) {
	catalog := cloneCatalog(validated.catalog)
	if override.Mode == OverrideIgnore {
		return SelectionDecision{Status: SelectionIgnored, Current: catalog.Current}, nil
	}
	if req.ProjectID == "" {
		return SelectionDecision{Status: SelectionUnavailable, Current: catalog.Current}, ErrInvalidWorkRequest
	}
	completeEnvelope := validWorkRequest(req)
	automatic := []Route(nil)
	if completeEnvelope {
		automatic = eligibleRoutes(catalog, req, policy, ledger, now)
	}
	if override.Mode == OverrideRequire {
		if route, ok := explicitlySelectable(catalog, override.Alias, req, policy, ledger, now); ok {
			return SelectionDecision{Status: SelectionRouted, Routes: []Route{route}}, nil
		}
		return SelectionDecision{Status: SelectionUnavailable, Current: catalog.Current}, ErrRequiredRouteUnavailable
	}
	if override.Mode == OverrideUse {
		if preferred, ok := explicitlySelectable(catalog, override.Alias, req, policy, ledger, now); ok {
			routes := []Route{preferred}
			for _, route := range automatic {
				if route.Alias != preferred.Alias {
					routes = append(routes, route)
				}
			}
			return SelectionDecision{Status: SelectionRouted, Routes: routes}, nil
		}
	}
	if !completeEnvelope {
		return SelectionDecision{Status: SelectionUnavailable, Current: catalog.Current}, ErrInvalidWorkRequest
	}
	if len(automatic) > 0 {
		return SelectionDecision{Status: SelectionRouted, Routes: automatic}, nil
	}
	if currentFallbackAllowed(catalog.Current, req, policy, now) {
		return SelectionDecision{Status: SelectionDegraded, Current: catalog.Current}, nil
	}
	return SelectionDecision{Status: SelectionUnavailable, Current: catalog.Current}, nil
}

func validWorkRequest(req WorkRequest) bool {
	if req.PlannedTier != TierTiny && req.PlannedTier != TierSmall && req.PlannedTier != TierMedium && req.PlannedTier != TierLarge {
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

func explicitlySelectable(catalog Catalog, alias string, req WorkRequest, policy PolicyContext, ledger AttemptLedger, now time.Time) (Route, bool) {
	if alias == "" || ledger.Attempted(alias) {
		return Route{}, false
	}
	for _, route := range catalog.Routes {
		if route.Alias != alias || validateRouteSchema(route) != nil || !hasReadiness(route.Readiness, ReadinessSelectable) || !routeAllowedByPolicy(route, req, policy, now) {
			continue
		}
		return route, true
	}
	return Route{}, false
}

func eligibleRoutes(catalog Catalog, req WorkRequest, policy PolicyContext, ledger AttemptLedger, now time.Time) []Route {
	floor := tierFloor(req.PlannedTier)
	floorRank := classRank(floor)
	sameClass := make([]Route, 0, len(catalog.Routes))
	higher := make([]Route, 0, len(catalog.Routes))
	for _, route := range catalog.Routes {
		if ledger.Attempted(route.Alias) || validateRouteSchema(route) != nil || !routeAllowedByPolicy(route, req, policy, now) || !automaticEligible(route, req, floor, now) {
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
	if !readinessCumulativeThrough(route.Readiness, ReadinessDispatchProven) {
		return false
	}
	evidence := route.Capability
	if !evidence.DispatchProven || (evidence.Source != EvidenceKBReceipt && evidence.Source != EvidenceAdapterPrior) {
		return false
	}
	if evidence.ExpiresAt.IsZero() || !now.Before(evidence.ExpiresAt) {
		return false
	}
	if evidence.RouteAlias != route.Alias || evidence.ModelID == "" || route.DisplayModelID == "" || evidence.ModelID != route.DisplayModelID {
		return false
	}
	if classRank(evidence.Class) < classRank(floor) {
		return false
	}
	if req.TaskFamily != "" && evidence.TaskFamily != req.TaskFamily {
		return false
	}
	if req.ContextSize > 0 && (evidence.ContextSize <= 0 || evidence.ContextSize < req.ContextSize) {
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

func currentFallbackAllowed(current CurrentModel, req WorkRequest, policy PolicyContext, now time.Time) bool {
	if policy.Project.DenyCurrentFallback || current.ModelID == "" || current.Route == nil || current.Route.DisplayModelID != current.ModelID {
		return false
	}
	return validateRouteSchema(*current.Route) == nil && routeAllowedByPolicy(*current.Route, req, policy, now)
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
