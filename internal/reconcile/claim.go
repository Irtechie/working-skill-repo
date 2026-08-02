package reconcile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	SemanticClaimSchemaVersion = 1
	maxResourceComponentBytes  = 256

	IdempotencyInFlight   = "in-flight"
	IdempotencyCommitted  = "committed"
	IdempotencyFailedSafe = "failed-safe"
)

var (
	ErrProtectedMutationUnavailable = errors.New("protected mutation unavailable")
	ErrAuthorityInvalid             = errors.New("authorization invalid")
	ErrStaleGeneration              = errors.New("stale fencing generation")
	ErrIdempotencyMismatch          = errors.New("idempotency payload mismatch")
	ErrIdempotencyInFlight          = errors.New("idempotency outcome is in-flight or unknown")
	ErrCommitOutcomeUnknown         = errors.New("side-effect outcome unknown")
)

// ResourceIdentity is provider-neutral input to the versioned canonical key.
// Provider, tenant, and type are ASCII case-insensitive. ID is case-sensitive,
// valid UTF-8, and must be precomposed (combining marks are rejected).
type ResourceIdentity struct {
	Provider string `json:"provider"`
	Tenant   string `json:"tenant"`
	Account  string `json:"account,omitempty"`
	Type     string `json:"resource_type"`
	ID       string `json:"canonical_resource_id"`
}

type ResourceKey struct {
	SchemaVersion int    `json:"schema_version"`
	Canonical     string `json:"canonical"`
	Display       string `json:"display"`
	Provider      string `json:"provider"`
	Tenant        string `json:"tenant"`
	Account       string `json:"account,omitempty"`
	Type          string `json:"resource_type"`
	ID            string `json:"canonical_resource_id"`
}

func CanonicalResourceKey(identity ResourceIdentity) (ResourceKey, error) {
	provider, err := canonicalASCII(identity.Provider, "provider")
	if err != nil {
		return ResourceKey{}, err
	}
	tenant, err := canonicalOptionalASCII(identity.Tenant, "tenant")
	if err != nil {
		return ResourceKey{}, err
	}
	account, err := canonicalOptionalASCII(identity.Account, "account")
	if err != nil {
		return ResourceKey{}, err
	}
	if tenant == "" && account == "" {
		return ResourceKey{}, fmt.Errorf("tenant or account is required")
	}
	resourceType, err := canonicalASCII(identity.Type, "resource type")
	if err != nil {
		return ResourceKey{}, err
	}
	id := strings.TrimSpace(identity.ID)
	if id == "" || !utf8.ValidString(id) || len([]byte(id)) > maxResourceComponentBytes {
		return ResourceKey{}, fmt.Errorf("canonical resource ID must be valid UTF-8 and 1-%d bytes", maxResourceComponentBytes)
	}
	for _, r := range id {
		if unicode.IsMark(r) || unicode.IsControl(r) {
			return ResourceKey{}, fmt.Errorf("canonical resource ID must be precomposed and contain no control characters")
		}
	}
	encode := func(value string) string { return hex.EncodeToString([]byte(value)) }
	canonical := strings.Join([]string{
		"semantic-resource-v1", encode(provider), encode(tenant), encode(account), encode(resourceType), encode(id),
	}, "/")
	if len(canonical) > 4*maxResourceComponentBytes {
		return ResourceKey{}, fmt.Errorf("canonical resource key exceeds maximum encoded length")
	}
	return ResourceKey{
		SchemaVersion: SemanticClaimSchemaVersion,
		Canonical:     canonical,
		Display:       resourceType + ":" + id,
		Provider:      provider,
		Tenant:        tenant,
		Account:       account,
		Type:          resourceType,
		ID:            id,
	}, nil
}

func canonicalASCII(value, label string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" || len(value) > 64 {
		return "", fmt.Errorf("%s must contain 1-64 characters", label)
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '.' || r == '_') {
			return "", fmt.Errorf("%s must use ASCII letters, digits, dot, underscore, or hyphen", label)
		}
	}
	return value, nil
}

func canonicalOptionalASCII(value, label string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	return canonicalASCII(value, label)
}

type ResourceRegistry struct {
	keys    map[string]ResourceKey
	aliases map[string]string
}

func NewResourceRegistry() *ResourceRegistry {
	return &ResourceRegistry{keys: map[string]ResourceKey{}, aliases: map[string]string{}}
}

func (r *ResourceRegistry) Register(identity ResourceIdentity, aliases ...string) (ResourceKey, error) {
	key, err := CanonicalResourceKey(identity)
	if err != nil {
		return ResourceKey{}, err
	}
	if r.keys == nil {
		r.keys = map[string]ResourceKey{}
	}
	if r.aliases == nil {
		r.aliases = map[string]string{}
	}
	aliases = append(aliases, key.Display)
	normalizedAliases := make([]string, 0, len(aliases))
	for _, raw := range aliases {
		alias := strings.ToLower(strings.TrimSpace(raw))
		parts := strings.SplitN(alias, ":", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return ResourceKey{}, fmt.Errorf("semantic alias must be type:value")
		}
		if existing, ok := r.aliases[alias]; ok && existing != key.Canonical {
			return ResourceKey{}, fmt.Errorf("semantic alias collision: %s", raw)
		}
		normalizedAliases = append(normalizedAliases, alias)
	}
	r.keys[key.Canonical] = key
	for _, alias := range normalizedAliases {
		r.aliases[alias] = key.Canonical
	}
	return key, nil
}

type Claim struct {
	Key              ResourceKey `json:"key"`
	Holder           string      `json:"holder"`
	WorkID           string      `json:"work_id"`
	SourceRevision   string      `json:"source_revision"`
	ClaimRevision    string      `json:"claim_revision"`
	Generation       uint64      `json:"generation"`
	ControllerID     string      `json:"controller_incarnation"`
	ControllerEpoch  uint64      `json:"controller_epoch"`
	AdapterID        string      `json:"adapter_identity"`
	WorkloadIdentity string      `json:"workload_identity"`
	IssuedAt         time.Time   `json:"issued_at"`
	ExpiresAt        time.Time   `json:"expires_at"`
}

type ClaimRequest struct {
	Key              ResourceKey
	Holder           string
	WorkID           string
	SourceRevision   string
	ControllerID     string
	ControllerEpoch  uint64
	AdapterID        string
	WorkloadIdentity string
	Now              time.Time
	TTL              time.Duration
}

// ClaimStore is the minimum adapter contract. Implementations must make CAS
// linearizable and persist epoch/generation high-water data outside snapshots
// that can be rolled back.
type ClaimStore interface {
	Acquire(ClaimRequest) (Claim, error)
	Takeover(ClaimRequest, string) (Claim, error)
	ValidateCurrent(Claim) error
}

type ClaimSnapshot struct {
	ControllerID    string
	ControllerEpoch uint64
	Claims          map[string]Claim
}

type MemoryClaimStore struct {
	mu              sync.Mutex
	controllerID    string
	controllerEpoch uint64
	epochHighWater  uint64
	claims          map[string]Claim
	claimHighWater  map[string]claimHighWater
	revision        uint64
	now             func() time.Time
}

type claimHighWater struct {
	Generation uint64
	Revision   string
}

func NewMemoryClaimStore() *MemoryClaimStore {
	return NewMemoryClaimStoreWithClock(time.Now)
}

func NewMemoryClaimStoreWithClock(now func() time.Time) *MemoryClaimStore {
	return &MemoryClaimStore{
		claims: map[string]Claim{}, claimHighWater: map[string]claimHighWater{}, now: now,
	}
}

func (s *MemoryClaimStore) SetController(id string, epoch uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(id) == "" || epoch == 0 {
		return fmt.Errorf("controller identity and positive epoch are required")
	}
	if epoch <= s.epochHighWater {
		return fmt.Errorf("controller epoch rollback or reuse rejected")
	}
	s.controllerID = id
	s.controllerEpoch = epoch
	s.epochHighWater = epoch
	return nil
}

func (s *MemoryClaimStore) Acquire(request ClaimRequest) (Claim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.claims[request.Key.Canonical]; ok {
		return Claim{}, fmt.Errorf("authoritative claim already exists")
	}
	return s.writeClaim(request, 1)
}

func (s *MemoryClaimStore) Takeover(request ClaimRequest, expectedRevision string) (Claim, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.now == nil {
		return Claim{}, fmt.Errorf("authoritative clock is required")
	}
	request.Now = s.now().UTC()
	current, ok := s.claims[request.Key.Canonical]
	if !ok {
		return Claim{}, fmt.Errorf("claim does not exist")
	}
	if expectedRevision == "" || expectedRevision != current.ClaimRevision {
		return Claim{}, fmt.Errorf("claim compare-and-swap revision mismatch")
	}
	if request.Now.Before(current.ExpiresAt) {
		return Claim{}, fmt.Errorf("current claim has not expired")
	}
	return s.writeClaim(request, current.Generation+1)
}

func (s *MemoryClaimStore) writeClaim(request ClaimRequest, generation uint64) (Claim, error) {
	if request.Key.Canonical == "" || request.Holder == "" || request.WorkID == "" ||
		request.SourceRevision == "" || request.AdapterID == "" || request.WorkloadIdentity == "" {
		return Claim{}, fmt.Errorf("claim identity fields are required")
	}
	if request.ControllerID != s.controllerID || request.ControllerEpoch != s.controllerEpoch ||
		request.ControllerEpoch != s.epochHighWater {
		return Claim{}, fmt.Errorf("controller incarnation is not current")
	}
	if s.now == nil || request.TTL <= 0 {
		return Claim{}, fmt.Errorf("authoritative clock and positive TTL are required")
	}
	request.Now = s.now().UTC()
	canonical, err := CanonicalResourceKey(ResourceIdentity{
		Provider: request.Key.Provider,
		Tenant:   request.Key.Tenant,
		Account:  request.Key.Account,
		Type:     request.Key.Type,
		ID:       request.Key.ID,
	})
	if err != nil || canonical != request.Key {
		return Claim{}, fmt.Errorf("claim resource key is not canonical")
	}
	s.revision++
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%d", request.Key.Canonical, generation, s.revision)))
	claim := Claim{
		Key: request.Key, Holder: request.Holder, WorkID: request.WorkID,
		SourceRevision: request.SourceRevision, ClaimRevision: "sha256:" + hex.EncodeToString(sum[:]),
		Generation: generation, ControllerID: request.ControllerID, ControllerEpoch: request.ControllerEpoch,
		AdapterID: request.AdapterID, WorkloadIdentity: request.WorkloadIdentity,
		IssuedAt: request.Now.UTC(), ExpiresAt: request.Now.Add(request.TTL).UTC(),
	}
	s.claims[request.Key.Canonical] = claim
	s.claimHighWater[request.Key.Canonical] = claimHighWater{
		Generation: claim.Generation,
		Revision:   claim.ClaimRevision,
	}
	return claim, nil
}

func (s *MemoryClaimStore) ValidateCurrent(claim Claim) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	current, ok := s.claims[claim.Key.Canonical]
	if !ok || current != claim {
		return ErrStaleGeneration
	}
	if claim.ControllerID != s.controllerID || claim.ControllerEpoch != s.epochHighWater {
		return fmt.Errorf("controller incarnation is stale")
	}
	return nil
}

func (s *MemoryClaimStore) Snapshot() ClaimSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	claims := make(map[string]Claim, len(s.claims))
	for key, claim := range s.claims {
		claims[key] = claim
	}
	return ClaimSnapshot{ControllerID: s.controllerID, ControllerEpoch: s.controllerEpoch, Claims: claims}
}

func (s *MemoryClaimStore) Restore(snapshot ClaimSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if snapshot.ControllerEpoch < s.epochHighWater {
		return fmt.Errorf("claim snapshot rollback detected")
	}
	if snapshot.ControllerEpoch != s.controllerEpoch || snapshot.ControllerID != s.controllerID {
		return fmt.Errorf("claim snapshot controller incarnation mismatch")
	}
	for key, high := range s.claimHighWater {
		claim, ok := snapshot.Claims[key]
		if !ok {
			return fmt.Errorf("claim snapshot omits durable high-water resource %s", key)
		}
		if claim.Generation < high.Generation ||
			(claim.Generation == high.Generation && claim.ClaimRevision != high.Revision) {
			return fmt.Errorf("claim snapshot resource high-water regression detected")
		}
	}
	restored := make(map[string]Claim, len(snapshot.Claims))
	for key, claim := range snapshot.Claims {
		if key != claim.Key.Canonical || claim.Generation == 0 || claim.ClaimRevision == "" {
			return fmt.Errorf("claim snapshot contains invalid resource identity")
		}
		restored[key] = claim
	}
	for key, claim := range restored {
		s.claimHighWater[key] = claimHighWater{
			Generation: claim.Generation,
			Revision:   claim.ClaimRevision,
		}
	}
	s.claims = restored
	return nil
}

type ScopedAuthorization struct {
	Token            string    `json:"token,omitempty"`
	Audience         string    `json:"audience"`
	ControllerID     string    `json:"controller_identity"`
	ControllerEpoch  uint64    `json:"controller_epoch"`
	WorkloadIdentity string    `json:"workload_identity"`
	WorkID           string    `json:"work_id"`
	Operation        string    `json:"operation"`
	Key              string    `json:"resource_key"`
	Generation       uint64    `json:"generation"`
	RequestDigest    string    `json:"request_digest"`
	IdempotencyKey   string    `json:"idempotency_key"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type AuthorizationVerifier interface {
	Verify(context.Context, ScopedAuthorization, MutationRequest) error
	Capability() AuthorizationVerifierCapability
}

type AuthorizationVerifierCapability struct {
	AuthenticatedAuthority bool `json:"authenticated_authority"`
	AudienceBound          bool `json:"audience_bound"`
	WorkloadBound          bool `json:"workload_bound"`
	RequestDigestBound     bool `json:"request_digest_bound"`
	ExpiryEnforced         bool `json:"expiry_enforced"`
}

func (c AuthorizationVerifierCapability) ProtectedMutationReady() bool {
	return c.AuthenticatedAuthority && c.AudienceBound && c.WorkloadBound &&
		c.RequestDigestBound && c.ExpiryEnforced
}

type SolePathCapability struct {
	GatewayPrincipalOnly bool `json:"gateway_principal_only"`
	DirectPathDenied     bool `json:"direct_path_denied"`
	AlternatePathDenied  bool `json:"alternate_path_denied"`
}

type GatewayCapability struct {
	SchemaVersion              int                `json:"schema_version"`
	AdapterID                  string             `json:"adapter_identity,omitempty"`
	AtomicConditionalCommit    bool               `json:"atomic_conditional_commit"`
	SolePath                   SolePathCapability `json:"sole_path"`
	VerifierAvailable          bool               `json:"verifier_available"`
	ProtectedMutationAvailable bool               `json:"protected_mutation_available"`
	LiveProviderSupported      bool               `json:"live_provider_supported"`
	Limitations                []string           `json:"limitations,omitempty"`
}

type MutationRequest struct {
	Claim          Claim               `json:"claim"`
	Audience       string              `json:"audience"`
	Operation      string              `json:"operation"`
	RequestDigest  string              `json:"request_digest"`
	IdempotencyKey string              `json:"idempotency_key"`
	Authority      ScopedAuthorization `json:"authority"`
	Now            time.Time           `json:"now"`
}

type MutationReceipt struct {
	Key            string `json:"resource_key"`
	ClaimRevision  string `json:"claim_revision"`
	Generation     uint64 `json:"generation"`
	RequestDigest  string `json:"request_digest"`
	IdempotencyKey string `json:"idempotency_key"`
	Result         string `json:"result"`
	Replayed       bool   `json:"replayed"`
}

type IdempotencyRecord struct {
	Key            string `json:"resource_key"`
	IdempotencyKey string `json:"idempotency_key"`
	RequestDigest  string `json:"request_digest"`
	State          string `json:"state"`
	Result         string `json:"result,omitempty"`
	ClaimRevision  string `json:"claim_revision"`
	Generation     uint64 `json:"generation"`
}

type IdempotencyRecoveryDecision struct {
	RequestDigest string `json:"request_digest"`
	ClaimRevision string `json:"claim_revision"`
	Generation    uint64 `json:"generation"`
	State         string `json:"state"`
	Result        string `json:"result,omitempty"`
}

type IdempotencyRecoveryVerifier interface {
	Resolve(context.Context, IdempotencyRecord) (IdempotencyRecoveryDecision, error)
}

type SideEffectCommit func(context.Context, MutationRequest) (string, error)

type GatewayBootstrapState struct {
	SchemaVersion    int     `json:"schema_version"`
	ReconciliationID string  `json:"reconciliation_id"`
	ControllerID     string  `json:"controller_identity"`
	ControllerEpoch  uint64  `json:"controller_epoch"`
	CurrentClaims    []Claim `json:"current_claims"`
}

type GatewayBootstrapVerifierCapability struct {
	AuthenticatedAuthority bool `json:"authenticated_authority"`
	RollbackProtected      bool `json:"rollback_protected"`
}

func (c GatewayBootstrapVerifierCapability) Ready() bool {
	return c.AuthenticatedAuthority && c.RollbackProtected
}

type GatewayBootstrapVerifier interface {
	VerifyBootstrap(context.Context, GatewayBootstrapState) error
	Capability() GatewayBootstrapVerifierCapability
}

type ReferenceGateway struct {
	capability      GatewayCapability
	verifier        AuthorizationVerifier
	commit          SideEffectCommit
	now             func() time.Time
	lockMu          sync.Mutex
	locks           map[string]*sync.Mutex
	stateMu         sync.Mutex
	highWater       map[string]uint64
	revisions       map[string]string
	epochs          map[string]uint64
	controllers     map[string]string
	records         map[string]IdempotencyRecord
	baseReady       bool
	bootstrapped    bool
	controllerID    string
	controllerEpoch uint64
}

func NewReferenceGateway(capability GatewayCapability, verifier AuthorizationVerifier, commit SideEffectCommit) *ReferenceGateway {
	return NewReferenceGatewayWithClock(capability, verifier, commit, time.Now)
}

func NewReferenceGatewayWithClock(
	capability GatewayCapability,
	verifier AuthorizationVerifier,
	commit SideEffectCommit,
	now func() time.Time,
) *ReferenceGateway {
	capability.SchemaVersion = SemanticClaimSchemaVersion
	capability.VerifierAvailable = verifier != nil && verifier.Capability().ProtectedMutationReady()
	baseReady := capability.AtomicConditionalCommit &&
		capability.SolePath.GatewayPrincipalOnly && capability.SolePath.DirectPathDenied &&
		capability.SolePath.AlternatePathDenied && capability.VerifierAvailable && commit != nil
	capability.ProtectedMutationAvailable = false
	capability.LiveProviderSupported = false
	if !baseReady {
		capability.Limitations = append(capability.Limitations,
			"protected mutation requires a real verifier, atomic conditional commit, and proven gateway-only production path")
	}
	capability.Limitations = append(capability.Limitations,
		"protected mutation remains unavailable until authoritative controller and per-resource high-water bootstrap succeeds")
	sort.Strings(capability.Limitations)
	return &ReferenceGateway{
		capability: capability, verifier: verifier, commit: commit, now: now,
		locks: map[string]*sync.Mutex{}, highWater: map[string]uint64{},
		revisions: map[string]string{}, epochs: map[string]uint64{},
		controllers: map[string]string{}, records: map[string]IdempotencyRecord{},
		baseReady: baseReady,
	}
}

func (g *ReferenceGateway) Capability() GatewayCapability {
	g.stateMu.Lock()
	defer g.stateMu.Unlock()
	capability := g.capability
	capability.ProtectedMutationAvailable = g.baseReady && g.bootstrapped
	return capability
}

func (g *ReferenceGateway) Bootstrap(
	ctx context.Context,
	state GatewayBootstrapState,
	verifier GatewayBootstrapVerifier,
) error {
	if !g.baseReady || verifier == nil || !verifier.Capability().Ready() {
		return ErrProtectedMutationUnavailable
	}
	if state.SchemaVersion != SemanticClaimSchemaVersion ||
		strings.TrimSpace(state.ReconciliationID) == "" ||
		strings.TrimSpace(state.ControllerID) == "" || state.ControllerEpoch == 0 ||
		len(state.CurrentClaims) == 0 {
		return fmt.Errorf("authoritative gateway bootstrap is incomplete")
	}
	if err := verifier.VerifyBootstrap(ctx, state); err != nil {
		return err
	}
	g.stateMu.Lock()
	defer g.stateMu.Unlock()
	if state.ControllerEpoch < g.controllerEpoch ||
		(state.ControllerEpoch == g.controllerEpoch && g.controllerID != "" &&
			state.ControllerID != g.controllerID) {
		return ErrStaleGeneration
	}
	seen := make(map[string]bool, len(state.CurrentClaims))
	for _, claim := range state.CurrentClaims {
		if claim.Key.Canonical == "" || claim.ClaimRevision == "" || claim.Generation == 0 ||
			claim.ControllerID != state.ControllerID || claim.ControllerEpoch != state.ControllerEpoch {
			return fmt.Errorf("authoritative gateway bootstrap claim is inconsistent")
		}
		canonical, err := CanonicalResourceKey(ResourceIdentity{
			Provider: claim.Key.Provider,
			Tenant:   claim.Key.Tenant,
			Account:  claim.Key.Account,
			Type:     claim.Key.Type,
			ID:       claim.Key.ID,
		})
		if err != nil || canonical != claim.Key || seen[claim.Key.Canonical] {
			return fmt.Errorf("authoritative gateway bootstrap resource key is invalid or duplicated")
		}
		seen[claim.Key.Canonical] = true
		if high := g.highWater[claim.Key.Canonical]; claim.Generation < high ||
			(claim.Generation == high && high != 0 &&
				g.revisions[claim.Key.Canonical] != claim.ClaimRevision) {
			return ErrStaleGeneration
		}
	}
	if state.ControllerEpoch > g.controllerEpoch && g.controllerEpoch != 0 {
		for key := range g.highWater {
			if !seen[key] {
				return fmt.Errorf("new controller incarnation bootstrap omits existing resource high-water state")
			}
		}
	}
	for _, claim := range state.CurrentClaims {
		g.highWater[claim.Key.Canonical] = claim.Generation
		g.revisions[claim.Key.Canonical] = claim.ClaimRevision
		g.epochs[claim.Key.Canonical] = claim.ControllerEpoch
		g.controllers[claim.Key.Canonical] = claim.ControllerID
	}
	g.controllerID = state.ControllerID
	g.controllerEpoch = state.ControllerEpoch
	g.bootstrapped = true
	return nil
}

func (g *ReferenceGateway) resourceLock(key string) *sync.Mutex {
	g.lockMu.Lock()
	defer g.lockMu.Unlock()
	lock := g.locks[key]
	if lock == nil {
		lock = &sync.Mutex{}
		g.locks[key] = lock
	}
	return lock
}

func (g *ReferenceGateway) Commit(ctx context.Context, request MutationRequest) (MutationReceipt, error) {
	if !g.Capability().ProtectedMutationAvailable {
		return MutationReceipt{}, ErrProtectedMutationUnavailable
	}
	if request.Claim.Key.Canonical == "" || request.IdempotencyKey == "" || request.RequestDigest == "" {
		return MutationReceipt{}, fmt.Errorf("%w: incomplete request", ErrAuthorityInvalid)
	}
	if g.now == nil {
		return MutationReceipt{}, ErrProtectedMutationUnavailable
	}
	request.Now = g.now().UTC()
	lock := g.resourceLock(request.Claim.Key.Canonical)
	lock.Lock()
	defer lock.Unlock()
	if err := g.verifier.Verify(ctx, request.Authority, request); err != nil {
		return MutationReceipt{}, err
	}
	recordKey := request.Claim.Key.Canonical + "\x00" + request.IdempotencyKey
	g.stateMu.Lock()
	if existing, ok := g.records[recordKey]; ok {
		g.stateMu.Unlock()
		if existing.RequestDigest != request.RequestDigest {
			return MutationReceipt{}, ErrIdempotencyMismatch
		}
		if existing.State == IdempotencyCommitted {
			return receiptFromRecord(existing, true), nil
		}
		return MutationReceipt{}, ErrIdempotencyInFlight
	}
	for _, existing := range g.records {
		if existing.Key == request.Claim.Key.Canonical && existing.State == IdempotencyInFlight {
			g.stateMu.Unlock()
			return MutationReceipt{}, ErrIdempotencyInFlight
		}
	}
	high := g.highWater[request.Claim.Key.Canonical]
	if high == 0 {
		g.stateMu.Unlock()
		return MutationReceipt{}, ErrProtectedMutationUnavailable
	}
	epoch := g.epochs[request.Claim.Key.Canonical]
	controller := g.controllers[request.Claim.Key.Canonical]
	if request.Claim.ControllerEpoch != g.controllerEpoch ||
		request.Claim.ControllerID != g.controllerID ||
		request.Claim.ControllerEpoch < epoch ||
		(request.Claim.ControllerEpoch == epoch && controller != "" && request.Claim.ControllerID != controller) {
		g.stateMu.Unlock()
		return MutationReceipt{}, ErrStaleGeneration
	}
	if request.Claim.Generation < high {
		g.stateMu.Unlock()
		return MutationReceipt{}, ErrStaleGeneration
	}
	if request.Claim.Generation == high && high != 0 &&
		g.revisions[request.Claim.Key.Canonical] != request.Claim.ClaimRevision {
		g.stateMu.Unlock()
		return MutationReceipt{}, ErrStaleGeneration
	}
	g.highWater[request.Claim.Key.Canonical] = request.Claim.Generation
	g.revisions[request.Claim.Key.Canonical] = request.Claim.ClaimRevision
	g.epochs[request.Claim.Key.Canonical] = request.Claim.ControllerEpoch
	g.controllers[request.Claim.Key.Canonical] = request.Claim.ControllerID
	record := IdempotencyRecord{
		Key: request.Claim.Key.Canonical, IdempotencyKey: request.IdempotencyKey,
		RequestDigest: request.RequestDigest, State: IdempotencyInFlight,
		ClaimRevision: request.Claim.ClaimRevision, Generation: request.Claim.Generation,
	}
	g.records[recordKey] = record
	g.stateMu.Unlock()

	result, err := g.commit(ctx, request)
	if err != nil {
		if !errors.Is(err, ErrCommitOutcomeUnknown) {
			g.stateMu.Lock()
			record.State = IdempotencyFailedSafe
			g.records[recordKey] = record
			g.stateMu.Unlock()
		}
		return MutationReceipt{}, err
	}
	g.stateMu.Lock()
	record.State = IdempotencyCommitted
	record.Result = result
	g.records[recordKey] = record
	g.stateMu.Unlock()
	return receiptFromRecord(record, false), nil
}

func (g *ReferenceGateway) IdempotencyStatus(key, idempotencyKey string) (IdempotencyRecord, bool) {
	g.stateMu.Lock()
	defer g.stateMu.Unlock()
	record, ok := g.records[key+"\x00"+idempotencyKey]
	return record, ok
}

// ResolveAmbiguous records an authoritative outcome without invoking the side
// effect again. A provider recovery adapter must establish the original
// operation's outcome; absence or mismatch leaves the reservation blocked.
func (g *ReferenceGateway) ResolveAmbiguous(
	ctx context.Context,
	key, idempotencyKey string,
	verifier IdempotencyRecoveryVerifier,
) (MutationReceipt, error) {
	if verifier == nil {
		return MutationReceipt{}, ErrIdempotencyInFlight
	}
	lock := g.resourceLock(key)
	lock.Lock()
	defer lock.Unlock()

	recordKey := key + "\x00" + idempotencyKey
	g.stateMu.Lock()
	record, ok := g.records[recordKey]
	g.stateMu.Unlock()
	if !ok || (record.State != IdempotencyInFlight && record.State != IdempotencyFailedSafe) {
		return MutationReceipt{}, ErrIdempotencyInFlight
	}
	decision, err := verifier.Resolve(ctx, record)
	if err != nil {
		return MutationReceipt{}, err
	}
	if decision.RequestDigest != record.RequestDigest ||
		decision.ClaimRevision != record.ClaimRevision ||
		decision.Generation != record.Generation ||
		(decision.State != IdempotencyCommitted && decision.State != IdempotencyFailedSafe) ||
		(decision.State == IdempotencyCommitted && decision.Result == "") {
		return MutationReceipt{}, ErrIdempotencyMismatch
	}
	record.State = decision.State
	record.Result = decision.Result
	g.stateMu.Lock()
	g.records[recordKey] = record
	g.stateMu.Unlock()
	if record.State != IdempotencyCommitted {
		return MutationReceipt{}, ErrIdempotencyInFlight
	}
	return receiptFromRecord(record, true), nil
}

func receiptFromRecord(record IdempotencyRecord, replayed bool) MutationReceipt {
	return MutationReceipt{
		Key: record.Key, ClaimRevision: record.ClaimRevision, Generation: record.Generation,
		RequestDigest: record.RequestDigest, IdempotencyKey: record.IdempotencyKey,
		Result: record.Result, Replayed: replayed,
	}
}

func ValidateAuthorizationBinding(authorization ScopedAuthorization, request MutationRequest) error {
	if request.Now.IsZero() || !request.Now.Before(authorization.ExpiresAt) ||
		!request.Now.Before(request.Claim.ExpiresAt) {
		return fmt.Errorf("%w: expired authority or claim", ErrAuthorityInvalid)
	}
	if authorization.Audience != request.Audience ||
		authorization.ControllerID != request.Claim.ControllerID ||
		authorization.ControllerEpoch != request.Claim.ControllerEpoch ||
		authorization.WorkloadIdentity != request.Claim.WorkloadIdentity ||
		authorization.WorkID != request.Claim.WorkID ||
		authorization.Operation != request.Operation ||
		authorization.Key != request.Claim.Key.Canonical ||
		authorization.Generation != request.Claim.Generation ||
		authorization.RequestDigest != request.RequestDigest ||
		authorization.IdempotencyKey != request.IdempotencyKey {
		return fmt.Errorf("%w: scoped binding mismatch", ErrAuthorityInvalid)
	}
	return nil
}

type ClaimConformanceResult struct {
	SchemaVersion int      `json:"schema_version"`
	Status        string   `json:"status"`
	Checks        []string `json:"checks"`
	Limitations   []string `json:"limitations"`
}

func ReferenceClaimCapability() GatewayCapability {
	return NewReferenceGateway(GatewayCapability{
		AdapterID: "none",
		Limitations: []string{
			"reference protocol only; no live authoritative claim adapter is configured",
			"local Git-common-directory coordination is not global authority",
		},
	}, nil, nil).Capability()
}

func ReferenceClaimConformance() ClaimConformanceResult {
	return ClaimConformanceResult{
		SchemaVersion: SemanticClaimSchemaVersion,
		Status:        "passed",
		Checks: []string{
			"canonical-key-alias-collision-rejection", "expiry-plus-cas-takeover",
			"monotonic-generation", "stale-worker-rejection", "rollback-epoch-rejection",
			"same-epoch-generation-revision-rollback-rejection",
			"wrong-audience-and-forged-authority-rejection", "gateway-high-water",
			"cold-gateway-authoritative-bootstrap-required",
			"atomic-validation-to-commit-capability", "durable-idempotency-reservation",
			"ambiguous-retry-blocking", "authoritative-outcome-recovery",
			"direct-path-denial", "alternate-path-denial",
			"same-resource-serialization", "disjoint-resource-concurrency",
			"coordinator-outage-preservation",
		},
		Limitations: []string{
			"deterministic in-memory reference only",
			"does not certify production IAM, network policy, credentials, or a live provider adapter",
		},
	}
}
