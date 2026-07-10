package modelrouting

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const CatalogSchemaVersion = 1

type Tier string

const (
	TierTiny   Tier = "tiny"
	TierSmall  Tier = "small"
	TierMedium Tier = "medium"
	TierLarge  Tier = "large"
)

type CapabilityClass string

const (
	ClassSmall   CapabilityClass = "small"
	ClassMedium  CapabilityClass = "medium"
	ClassLarge   CapabilityClass = "large"
	ClassPlanner CapabilityClass = "planner"
)

type Readiness string

const (
	ReadinessDiscovered     Readiness = "discovered"
	ReadinessConfigured     Readiness = "configured"
	ReadinessSelectable     Readiness = "selectable"
	ReadinessDispatchProven Readiness = "dispatch-proven"
)

var readinessOrder = []Readiness{
	ReadinessDiscovered,
	ReadinessConfigured,
	ReadinessSelectable,
	ReadinessDispatchProven,
}

type EvidenceSource string

const (
	EvidenceDeclared        EvidenceSource = "declared"
	EvidenceAdapterPrior    EvidenceSource = "adapter-prior"
	EvidenceKBReceipt       EvidenceSource = "kb-receipt"
	EvidenceRepositoryClaim EvidenceSource = "repository-claim"
	EvidenceModelSelfReport EvidenceSource = "model-self-report"
)

type RetentionClass string

const (
	RetentionNone    RetentionClass = "none"
	RetentionSession RetentionClass = "session"
	RetentionLimited RetentionClass = "limited"
	RetentionUnknown RetentionClass = "unknown"
)

type RiskLevel string

const (
	RiskNormal RiskLevel = "normal"
	RiskBroad  RiskLevel = "broad"
)

type TrustBoundary string

const (
	BoundaryHosted  TrustBoundary = "hosted"
	BoundaryPrivate TrustBoundary = "private"
)

type TrainingUse string

const (
	TrainingNo      TrainingUse = "no"
	TrainingYes     TrainingUse = "yes"
	TrainingUnknown TrainingUse = "unknown"
)

type SupportCohort string

const (
	CohortUnspecified  SupportCohort = "unspecified"
	CohortInitialPilot SupportCohort = "initial-pilot"
)

type SurfaceFingerprint struct {
	Surface    string `json:"surface"`
	Provider   string `json:"provider"`
	Revision   string `json:"revision"`
	ConfigHash string `json:"config_hash"`
	AgentHash  string `json:"agent_hash,omitempty"`
}

type CatalogSource string

const (
	CatalogSourceUser   CatalogSource = "user-local"
	CatalogSourceNative CatalogSource = "native-adapter"
	CatalogSourceRun    CatalogSource = "kb-run"
)

// ValidatedCatalog is the only catalog accepted by the selector. Its contents
// are copied and cannot be replaced with raw repository or adapter data after
// endpoint, provenance, schema, and fingerprint validation.
type ValidatedCatalog struct {
	catalog Catalog
}

type RouteRejection struct {
	Alias  string
	Reason string
}

type Catalog struct {
	SchemaVersion int                  `json:"schema_version"`
	Fingerprint   string               `json:"fingerprint,omitempty"`
	Cohort        SupportCohort        `json:"support_cohort,omitempty"`
	Surfaces      []SurfaceFingerprint `json:"surfaces,omitempty"`
	Current       CurrentModel         `json:"current"`
	Routes        []Route              `json:"routes"`
}

type CurrentModel struct {
	ModelID string `json:"model_id"`
	Surface string `json:"surface,omitempty"`
	Route   *Route `json:"route,omitempty"`
}

type Route struct {
	RouteID         string             `json:"route_id,omitempty"`
	Alias           string             `json:"alias"`
	DisplayModelID  string             `json:"display_model_id"`
	Adapter         string             `json:"adapter"`
	AdapterRevision string             `json:"adapter_revision,omitempty"`
	DispatchMethod  string             `json:"dispatch_method"`
	Profile         string             `json:"profile,omitempty"`
	ProfileRevision string             `json:"profile_revision,omitempty"`
	Destination     string             `json:"destination"`
	Endpoint        string             `json:"endpoint,omitempty"`
	AuthEnv         string             `json:"auth_env,omitempty"`
	Boundary        TrustBoundary      `json:"boundary"`
	Retention       RetentionClass     `json:"retention"`
	TrainingUse     TrainingUse        `json:"training_use"`
	Residency       string             `json:"residency"`
	TrustProvenance string             `json:"trust_provenance"`
	SourceRouteID   string             `json:"source_route_id,omitempty"`
	Readiness       []Readiness        `json:"readiness"`
	Capability      CapabilityEvidence `json:"capability"`
}

type CapabilityEvidence struct {
	Class          CapabilityClass `json:"class"`
	Source         EvidenceSource  `json:"source"`
	RouteAlias     string          `json:"route_alias"`
	ModelID        string          `json:"model_id"`
	TaskFamily     string          `json:"task_family"`
	Tools          []string        `json:"tools,omitempty"`
	ContextSize    int             `json:"context_size,omitempty"`
	Risk           RiskLevel       `json:"risk"`
	DispatchProven bool            `json:"dispatch_proven"`
	ExpiresAt      time.Time       `json:"expires_at,omitempty"`
}

var (
	ErrRequiredRouteUnavailable        = errors.New("required route unavailable")
	ErrRouteAlreadyAttempted           = errors.New("route already attempted")
	ErrAttemptLedgerFull               = errors.New("attempt ledger full")
	ErrInvalidAttempt                  = errors.New("invalid route attempt")
	ErrUnsafePath                      = errors.New("unsafe storage path")
	ErrStorageSizeExceeded             = errors.New("storage size exceeded")
	ErrCredentialValue                 = errors.New("credential value is not allowed")
	ErrUnsafeEndpoint                  = errors.New("unsafe endpoint")
	ErrPrivateEndpointRequiresApproval = errors.New("private endpoint requires local approval")
	ErrPrivateHTTPRequiresApproval     = ErrPrivateEndpointRequiresApproval
	ErrAuthOriginMismatch              = errors.New("auth origin mismatch")
	ErrDuplicateAlias                  = errors.New("duplicate route alias")
	ErrAliasConflict                   = errors.New("conflicting route alias")
	ErrInvalidCatalog                  = errors.New("invalid catalog")
	ErrUnknownAdapter                  = errors.New("unknown adapter dispatch method")
	ErrInvalidWorkRequest              = errors.New("invalid work request")
)

func validateRouteSchema(route Route) error {
	if strings.TrimSpace(route.Alias) == "" || strings.TrimSpace(route.DisplayModelID) == "" ||
		strings.TrimSpace(route.Adapter) == "" || strings.TrimSpace(route.DispatchMethod) == "" ||
		strings.TrimSpace(route.Destination) == "" {
		return fmt.Errorf("%w: route required field missing", ErrInvalidCatalog)
	}
	if !validAlias(route.Alias) || strings.TrimSpace(route.AdapterRevision) == "" || len(route.AdapterRevision) > 128 {
		return fmt.Errorf("%w: invalid alias or missing adapter revision", ErrInvalidCatalog)
	}
	textFields := []struct {
		name  string
		value string
	}{
		{"display_model_id", route.DisplayModelID},
		{"destination", route.Destination},
		{"residency", route.Residency},
		{"trust_provenance", route.TrustProvenance},
		{"task_family", route.Capability.TaskFamily},
	}
	for _, field := range textFields {
		if !validTextField(field.value, 256) {
			return fmt.Errorf("%w: invalid %s", ErrInvalidCatalog, field.name)
		}
	}
	for _, tool := range route.Capability.Tools {
		if !validTextField(tool, 128) {
			return fmt.Errorf("%w: invalid capability tool", ErrInvalidCatalog)
		}
	}
	if route.Profile != "" && !validProfileName(route.Profile) {
		return fmt.Errorf("%w: invalid profile", ErrInvalidCatalog)
	}
	if route.ProfileRevision != "" && !validTextField(route.ProfileRevision, 128) {
		return fmt.Errorf("%w: invalid profile revision", ErrInvalidCatalog)
	}
	if route.AuthEnv != "" && !validEnvironmentName(route.AuthEnv) {
		return fmt.Errorf("%w: invalid auth environment variable name", ErrInvalidCatalog)
	}
	if route.RouteID != "" && !validRouteID(route.RouteID) {
		return fmt.Errorf("%w: invalid route id", ErrInvalidCatalog)
	}
	if route.SourceRouteID != "" && !validRouteID(route.SourceRouteID) {
		return fmt.Errorf("%w: invalid source route id", ErrInvalidCatalog)
	}
	if !supportedAdapterRevision(route.Adapter, route.DispatchMethod, route.AdapterRevision) {
		return fmt.Errorf("%w: %s/%s", ErrUnknownAdapter, route.Adapter, route.DispatchMethod)
	}
	if route.Boundary != BoundaryHosted && route.Boundary != BoundaryPrivate {
		return fmt.Errorf("%w: invalid trust boundary", ErrInvalidCatalog)
	}
	if !validRetention(route.Retention) || !validTrainingUse(route.TrainingUse) || strings.TrimSpace(route.Residency) == "" || strings.TrimSpace(route.TrustProvenance) == "" {
		return fmt.Errorf("%w: incomplete trust metadata", ErrInvalidCatalog)
	}
	if err := validateReadiness(route.Readiness); err != nil {
		return err
	}
	evidence := route.Capability
	if !validClass(evidence.Class) || !validEvidenceSource(evidence.Source) || !validRisk(evidence.Risk) {
		return fmt.Errorf("%w: invalid capability evidence enum", ErrInvalidCatalog)
	}
	if evidence.RouteAlias != route.Alias || evidence.ModelID != route.DisplayModelID {
		return fmt.Errorf("%w: route/evidence identity mismatch", ErrInvalidCatalog)
	}
	if evidence.DispatchProven != hasReadiness(route.Readiness, ReadinessDispatchProven) {
		return fmt.Errorf("%w: dispatch proof/readiness mismatch", ErrInvalidCatalog)
	}
	return nil
}

func validTextField(value string, maxBytes int) bool {
	if strings.TrimSpace(value) == "" || len(value) > maxBytes || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validAlias(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' || character == '.' {
			continue
		}
		return false
	}
	return true
}

func validProfileName(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' || character == '.' {
			continue
		}
		return false
	}
	return !strings.Contains(value, "..")
}

func validEnvironmentName(value string) bool {
	if len(value) == 0 || len(value) > 128 || !((value[0] >= 'A' && value[0] <= 'Z') || value[0] == '_') {
		return false
	}
	for index := 1; index < len(value); index++ {
		character := value[index]
		if (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' {
			continue
		}
		return false
	}
	return true
}

func validateReadiness(values []Readiness) error {
	if len(values) == 0 || len(values) > len(readinessOrder) {
		return fmt.Errorf("%w: empty or oversized readiness", ErrInvalidCatalog)
	}
	for index, value := range values {
		if value != readinessOrder[index] {
			return fmt.Errorf("%w: readiness must be cumulative", ErrInvalidCatalog)
		}
	}
	return nil
}

func knownAdapterDispatch(adapter, method string) bool {
	allowed := map[string]map[string]bool{
		"codex":             {"named-agent": true, "exec-model": true, "exec-profile": true},
		"litellm":           {"openai-compatible": true},
		"openai-compatible": {"chat-completions": true, "codex-profile": true},
	}
	return allowed[adapter][method]
}

func supportedAdapterRevision(adapter, method, revision string) bool {
	if !knownAdapterDispatch(adapter, method) {
		return false
	}
	if revision == "v1" {
		return true
	}
	if adapter == "codex" && (method == "exec-model" || method == "exec-profile") {
		return validCodexAdapterRevision(revision)
	}
	return false
}

func validCodexAdapterRevision(revision string) bool {
	const prefix = "codex-cli-v1:"
	if !strings.HasPrefix(revision, prefix) {
		return false
	}
	rest := strings.TrimPrefix(revision, prefix)
	parts := strings.Split(rest, ":")
	if len(parts) != 2 || len(parts[0]) != 64 || len(parts[1]) == 0 || len(parts[1]) > 40 {
		return false
	}
	for _, character := range parts[0] {
		if (character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') {
			continue
		}
		return false
	}
	for _, character := range parts[1] {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') ||
			character == '-' || character == '_' || character == '.' {
			continue
		}
		return false
	}
	return true
}

func validClass(value CapabilityClass) bool {
	return value == ClassSmall || value == ClassMedium || value == ClassLarge || value == ClassPlanner
}

func validEvidenceSource(value EvidenceSource) bool {
	return value == EvidenceDeclared || value == EvidenceAdapterPrior || value == EvidenceKBReceipt || value == EvidenceRepositoryClaim || value == EvidenceModelSelfReport
}

func validRisk(value RiskLevel) bool { return value == RiskNormal || value == RiskBroad }

func validRetention(value RetentionClass) bool {
	return value == RetentionNone || value == RetentionSession || value == RetentionLimited || value == RetentionUnknown
}

func validTrainingUse(value TrainingUse) bool {
	return value == TrainingNo || value == TrainingYes || value == TrainingUnknown
}

func ComputeRouteFingerprint(route Route) (string, error) {
	identity := struct {
		RouteID         string         `json:"route_id"`
		Alias           string         `json:"alias"`
		DisplayModelID  string         `json:"display_model_id"`
		Adapter         string         `json:"adapter"`
		AdapterRevision string         `json:"adapter_revision"`
		DispatchMethod  string         `json:"dispatch_method"`
		Profile         string         `json:"profile,omitempty"`
		ProfileRevision string         `json:"profile_revision,omitempty"`
		Destination     string         `json:"destination"`
		Endpoint        string         `json:"endpoint"`
		AuthEnv         string         `json:"auth_env"`
		Boundary        TrustBoundary  `json:"boundary"`
		Retention       RetentionClass `json:"retention"`
		TrainingUse     TrainingUse    `json:"training_use"`
		Residency       string         `json:"residency"`
		TrustProvenance string         `json:"trust_provenance"`
		SourceRouteID   string         `json:"source_route_id"`
	}{route.RouteID, route.Alias, route.DisplayModelID, route.Adapter, route.AdapterRevision, route.DispatchMethod, route.Profile, route.ProfileRevision,
		route.Destination, route.Endpoint, route.AuthEnv, route.Boundary, route.Retention,
		route.TrainingUse, route.Residency, route.TrustProvenance, route.SourceRouteID}
	data, err := json.Marshal(identity)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "route-sha256:" + hex.EncodeToString(sum[:]), nil
}

// ApprovalRouteFingerprint resolves an opaque run reference against trusted
// user-local configuration before computing the private approval fingerprint.
func ApprovalRouteFingerprint(route Route, sources map[string]Route) (string, error) {
	if route.SourceRouteID != "" {
		source, ok := sources[route.SourceRouteID]
		if !ok || source.RouteID != route.SourceRouteID {
			return "", fmt.Errorf("%w: unknown source route id", ErrInvalidCatalog)
		}
		return ComputeRouteFingerprint(source)
	}
	return ComputeRouteFingerprint(route)
}

func validRouteID(value string) bool {
	const prefix = "route-id:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+32 {
		return false
	}
	for _, character := range value[len(prefix):] {
		if (character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') {
			continue
		}
		return false
	}
	return true
}

func validRouteFingerprint(value string) bool {
	const prefix = "route-sha256:"
	if !strings.HasPrefix(value, prefix) || len(value) != len(prefix)+64 {
		return false
	}
	for _, character := range value[len(prefix):] {
		if (character >= '0' && character <= '9') || (character >= 'a' && character <= 'f') {
			continue
		}
		return false
	}
	return true
}

func ComputeRouteStateFingerprint(route Route) (string, error) {
	configuration, err := ComputeRouteFingerprint(route)
	if err != nil {
		return "", err
	}
	tools := append([]string(nil), route.Capability.Tools...)
	sort.Strings(tools)
	state := struct {
		Configuration  string          `json:"configuration"`
		Readiness      []Readiness     `json:"readiness"`
		Class          CapabilityClass `json:"class"`
		Source         EvidenceSource  `json:"source"`
		RouteAlias     string          `json:"route_alias"`
		ModelID        string          `json:"model_id"`
		TaskFamily     string          `json:"task_family"`
		Tools          []string        `json:"tools"`
		ContextSize    int             `json:"context_size"`
		Risk           RiskLevel       `json:"risk"`
		DispatchProven bool            `json:"dispatch_proven"`
		ExpiresAt      time.Time       `json:"expires_at"`
	}{configuration, append([]Readiness(nil), route.Readiness...), route.Capability.Class,
		route.Capability.Source, route.Capability.RouteAlias, route.Capability.ModelID,
		route.Capability.TaskFamily, tools, route.Capability.ContextSize, route.Capability.Risk,
		route.Capability.DispatchProven, route.Capability.ExpiresAt}
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "route-state-sha256:" + hex.EncodeToString(sum[:]), nil
}

func ComputeCatalogFingerprint(catalog Catalog) (string, error) {
	surfaces := append([]SurfaceFingerprint(nil), catalog.Surfaces...)
	sort.Slice(surfaces, func(i, j int) bool {
		left := surfaces[i].Surface + "\x00" + surfaces[i].Provider + "\x00" + surfaces[i].Revision + "\x00" + surfaces[i].ConfigHash + "\x00" + surfaces[i].AgentHash
		right := surfaces[j].Surface + "\x00" + surfaces[j].Provider + "\x00" + surfaces[j].Revision + "\x00" + surfaces[j].ConfigHash + "\x00" + surfaces[j].AgentHash
		return left < right
	})
	routeIDs := make([]string, 0, len(catalog.Routes))
	for _, route := range catalog.Routes {
		fingerprint, err := ComputeRouteStateFingerprint(route)
		if err != nil {
			return "", err
		}
		routeIDs = append(routeIDs, fingerprint)
	}
	sort.Strings(routeIDs)
	currentRoute := ""
	if catalog.Current.Route != nil {
		var err error
		currentRoute, err = ComputeRouteStateFingerprint(*catalog.Current.Route)
		if err != nil {
			return "", err
		}
	}
	payload := struct {
		Cohort       SupportCohort        `json:"cohort"`
		Surfaces     []SurfaceFingerprint `json:"surfaces"`
		Current      string               `json:"current"`
		CurrentRoute string               `json:"current_route"`
		Routes       []string             `json:"routes"`
	}{catalog.Cohort, surfaces, catalog.Current.ModelID, currentRoute, routeIDs}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return "catalog-sha256:" + hex.EncodeToString(sum[:]), nil
}

func MergeCatalog(native, user Catalog) (Catalog, error) {
	if native.SchemaVersion != CatalogSchemaVersion || user.SchemaVersion != CatalogSchemaVersion {
		return Catalog{}, fmt.Errorf("%w: unsupported schema version", ErrInvalidCatalog)
	}
	merged := native
	byAlias := make(map[string]Route, len(native.Routes)+len(user.Routes))
	routes := append([]Route(nil), native.Routes...)
	for _, route := range user.Routes {
		routes = append(routes, normalizeUserRoute(route))
	}
	for _, route := range routes {
		if existing, ok := byAlias[route.Alias]; ok {
			left, _ := ComputeRouteFingerprint(existing)
			right, _ := ComputeRouteFingerprint(route)
			if left != right {
				return Catalog{}, fmt.Errorf("%w: %s", ErrAliasConflict, route.Alias)
			}
			continue
		}
		byAlias[route.Alias] = route
	}
	// Do not reuse native.Routes backing storage: catalog merge must not mutate
	// either input while it produces a deterministic result.
	merged.Routes = nil
	aliases := make([]string, 0, len(byAlias))
	for alias := range byAlias {
		aliases = append(aliases, alias)
	}
	sort.Strings(aliases)
	for _, alias := range aliases {
		merged.Routes = append(merged.Routes, byAlias[alias])
	}
	fingerprint, err := ComputeCatalogFingerprint(merged)
	if err != nil {
		return Catalog{}, err
	}
	merged.Fingerprint = fingerprint
	return merged, nil
}

func normalizeUserRoute(route Route) Route {
	route.SourceRouteID = ""
	route.Capability.Source = EvidenceDeclared
	route.Capability.DispatchProven = false
	if len(route.Readiness) > 3 {
		route.Readiness = append([]Readiness(nil), route.Readiness[:3]...)
	} else {
		route.Readiness = append([]Readiness(nil), route.Readiness...)
	}
	return route
}

func validateRouteProvenance(route Route, source CatalogSource) error {
	if !supportedAdapterRevision(route.Adapter, route.DispatchMethod, route.AdapterRevision) {
		return fmt.Errorf("%w: unsupported adapter revision", ErrUnknownAdapter)
	}
	switch source {
	case CatalogSourceUser:
		if !validRouteID(route.RouteID) || route.SourceRouteID != "" || route.Capability.Source != EvidenceDeclared || route.Capability.DispatchProven || hasReadiness(route.Readiness, ReadinessDispatchProven) {
			return fmt.Errorf("%w: user routes may only declare selectable capability", ErrInvalidCatalog)
		}
	case CatalogSourceNative:
		if route.RouteID != "" || route.SourceRouteID != "" || (route.Capability.Source != EvidenceDeclared && route.Capability.Source != EvidenceAdapterPrior) {
			return fmt.Errorf("%w: native routes cannot claim receipt evidence", ErrInvalidCatalog)
		}
	case CatalogSourceRun:
		if route.Capability.Source != EvidenceDeclared && route.Capability.Source != EvidenceAdapterPrior && route.Capability.Source != EvidenceKBReceipt {
			return fmt.Errorf("%w: untrusted run evidence source", ErrInvalidCatalog)
		}
	default:
		return fmt.Errorf("%w: catalog source is required", ErrInvalidCatalog)
	}
	return nil
}

func validateSourceRouteReference(route Route, policy PolicyContext, trustedCurrent bool) error {
	if route.SourceRouteID == "" {
		if !trustedCurrent && (hasReadiness(route.Readiness, ReadinessConfigured) || hasReadiness(route.Readiness, ReadinessSelectable)) {
			state, err := ComputeRouteStateFingerprint(route)
			if err != nil || policy.TrustedRouteStates[route.Alias] != state {
				return fmt.Errorf("%w: configured run route is missing trusted source identity", ErrInvalidCatalog)
			}
		}
		return nil
	}
	source, ok := policy.RouteSources[route.SourceRouteID]
	if !ok || source.SourceRouteID != "" || source.RouteID != route.SourceRouteID {
		return fmt.Errorf("%w: source route is not present in trusted user-local configuration", ErrInvalidCatalog)
	}
	if route.RouteID != "" {
		return fmt.Errorf("%w: redacted route exposed a user-local route id field", ErrInvalidCatalog)
	}
	if route.Endpoint != "" || route.AuthEnv != "" ||
		route.Adapter != source.Adapter || route.AdapterRevision != source.AdapterRevision ||
		route.DispatchMethod != source.DispatchMethod || route.Profile != source.Profile || route.ProfileRevision != source.ProfileRevision || route.Destination != source.Destination ||
		route.Boundary != source.Boundary || route.Retention != source.Retention ||
		route.TrainingUse != source.TrainingUse || route.Residency != source.Residency ||
		route.TrustProvenance != source.TrustProvenance ||
		route.Capability.Class != source.Capability.Class ||
		route.Capability.Source != EvidenceDeclared || route.Capability.DispatchProven ||
		route.Capability.TaskFamily != source.Capability.TaskFamily ||
		route.Capability.ContextSize != source.Capability.ContextSize ||
		route.Capability.Risk != source.Capability.Risk ||
		!slices.Equal(route.Capability.Tools, source.Capability.Tools) {
		return fmt.Errorf("%w: redacted route does not match trusted source", ErrInvalidCatalog)
	}
	if route.Alias == source.Alias {
		if route.DisplayModelID != source.DisplayModelID || !slices.Equal(route.Readiness, source.Readiness) {
			return fmt.Errorf("%w: redacted route identity does not match trusted source", ErrInvalidCatalog)
		}
		return nil
	}
	if !strings.HasPrefix(route.Alias, source.Alias+".") || len(route.Readiness) != 1 || route.Readiness[0] != ReadinessDiscovered {
		return fmt.Errorf("%w: discovered child route is not safely derived from trusted source", ErrInvalidCatalog)
	}
	return nil
}

func ValidateCatalogForSelection(catalog Catalog, policy PolicyContext, resolver Resolver, now time.Time, source CatalogSource) (ValidatedCatalog, []RouteRejection, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ValidateCatalogForSelectionContext(ctx, catalog, policy, resolver, now, source)
}

// ValidateCatalogForSelectionContext applies one caller-owned deadline across
// every route, including DNS. Dispatchers should pass their work/session context.
func ValidateCatalogForSelectionContext(ctx context.Context, catalog Catalog, policy PolicyContext, resolver Resolver, now time.Time, source CatalogSource) (ValidatedCatalog, []RouteRejection, error) {
	if catalog.SchemaVersion != CatalogSchemaVersion {
		return ValidatedCatalog{}, nil, fmt.Errorf("%w: unsupported schema version", ErrInvalidCatalog)
	}
	if source == CatalogSourceRun && catalog.Fingerprint == "" {
		return ValidatedCatalog{}, nil, fmt.Errorf("%w: run catalog fingerprint is required", ErrInvalidCatalog)
	}
	if catalog.Fingerprint != "" {
		expected, err := ComputeCatalogFingerprint(catalog)
		if err != nil || expected != catalog.Fingerprint {
			return ValidatedCatalog{}, nil, fmt.Errorf("%w: catalog fingerprint mismatch", ErrInvalidCatalog)
		}
	}
	validated := cloneCatalog(catalog)
	validated.Routes = nil
	rejections := make([]RouteRejection, 0)
	aliases := make(map[string]string, len(catalog.Routes))
	for _, route := range catalog.Routes {
		stateFingerprint, fingerprintErr := ComputeRouteStateFingerprint(route)
		if fingerprintErr != nil {
			return ValidatedCatalog{}, nil, fingerprintErr
		}
		if existing, exists := aliases[route.Alias]; exists {
			if existing == stateFingerprint {
				return ValidatedCatalog{}, nil, fmt.Errorf("%w: %s", ErrDuplicateAlias, route.Alias)
			}
			return ValidatedCatalog{}, nil, fmt.Errorf("%w: %s", ErrAliasConflict, route.Alias)
		}
		aliases[route.Alias] = stateFingerprint
		if err := validateRouteProvenance(route, source); err != nil {
			rejections = append(rejections, RouteRejection{Alias: route.Alias, Reason: err.Error()})
			continue
		}
		trustedCurrent := isTrustedCurrentRoute(route, catalog, policy)
		if err := validateSourceRouteReference(route, policy, trustedCurrent); err != nil {
			rejections = append(rejections, RouteRejection{Alias: route.Alias, Reason: err.Error()})
			continue
		}
		if err := ValidateRouteContext(ctx, route, policy, resolver, now); err != nil {
			rejections = append(rejections, RouteRejection{Alias: route.Alias, Reason: err.Error()})
			continue
		}
		validated.Routes = append(validated.Routes, cloneRoute(route))
	}
	if catalog.Current.Route != nil {
		if err := validateRouteProvenance(*catalog.Current.Route, source); err != nil {
			validated.Current.Route = nil
			rejections = append(rejections, RouteRejection{Alias: catalog.Current.Route.Alias, Reason: err.Error()})
		} else if err := validateSourceRouteReference(*catalog.Current.Route, policy, isTrustedCurrentRoute(*catalog.Current.Route, catalog, policy)); err != nil {
			validated.Current.Route = nil
			rejections = append(rejections, RouteRejection{Alias: catalog.Current.Route.Alias, Reason: err.Error()})
		} else if err := ValidateRouteContext(ctx, *catalog.Current.Route, policy, resolver, now); err != nil {
			validated.Current.Route = nil
			rejections = append(rejections, RouteRejection{Alias: catalog.Current.Route.Alias, Reason: err.Error()})
		}
	}
	return ValidatedCatalog{catalog: validated}, rejections, nil
}

func isTrustedCurrentRoute(route Route, catalog Catalog, policy PolicyContext) bool {
	if policy.TrustedCurrentModelID == "" || policy.TrustedCurrentSurface == "" || policy.TrustedCurrentRouteState == "" || catalog.Current.Route == nil ||
		catalog.Current.ModelID != policy.TrustedCurrentModelID || catalog.Current.Surface != policy.TrustedCurrentSurface ||
		route.RouteID != "" || route.SourceRouteID != "" || route.Alias != "current" ||
		route.DisplayModelID != policy.TrustedCurrentModelID || route.Adapter != "codex" ||
		route.AdapterRevision != "v1" || route.DispatchMethod != "exec-model" || route.Destination != "current" ||
		route.Endpoint != "" || route.AuthEnv != "" || route.Boundary != BoundaryHosted ||
		route.TrustProvenance != "active orchestrator" || route.Capability.Source != EvidenceDeclared || route.Capability.DispatchProven {
		return false
	}
	left, leftErr := ComputeRouteStateFingerprint(route)
	right, rightErr := ComputeRouteStateFingerprint(*catalog.Current.Route)
	return leftErr == nil && rightErr == nil && left == right && left == policy.TrustedCurrentRouteState
}

func cloneCatalog(catalog Catalog) Catalog {
	copy := catalog
	copy.Surfaces = append([]SurfaceFingerprint(nil), catalog.Surfaces...)
	copy.Routes = make([]Route, 0, len(catalog.Routes))
	for _, route := range catalog.Routes {
		copy.Routes = append(copy.Routes, cloneRoute(route))
	}
	if catalog.Current.Route != nil {
		route := cloneRoute(*catalog.Current.Route)
		copy.Current.Route = &route
	}
	return copy
}

func cloneRoute(route Route) Route {
	copy := route
	copy.Readiness = append([]Readiness(nil), route.Readiness...)
	copy.Capability.Tools = append([]string(nil), route.Capability.Tools...)
	return copy
}
