package reconcile

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSemanticClaimCanonicalKeysAndAliases(t *testing.T) {
	registry := NewResourceRegistry()
	first, err := registry.Register(ResourceIdentity{
		Provider: " GitHub ", Tenant: "Acme", Type: "Publisher", ID: "Widget/Pro",
	}, "publisher:Widget/Pro")
	if err != nil {
		t.Fatal(err)
	}
	second, err := registry.Register(ResourceIdentity{
		Provider: "github", Tenant: "acme", Type: "publisher", ID: "Widget/Pro",
	}, "PUBLISHER:Widget/Pro")
	if err != nil {
		t.Fatal(err)
	}
	if first.Canonical != second.Canonical || first.Display != "publisher:Widget/Pro" {
		t.Fatalf("canonicalization drift: %#v %#v", first, second)
	}
	if _, err := registry.Register(ResourceIdentity{
		Provider: "github", Tenant: "other", Type: "publisher", ID: "Widget/Pro",
	}, "publisher:widget/pro"); err == nil {
		t.Fatal("alias collision was accepted")
	}
	if _, err := registry.Register(ResourceIdentity{
		Provider: "github", Tenant: "acme", Type: "publisher", ID: "e\u0301",
	}); err == nil {
		t.Fatal("non-canonical Unicode was accepted")
	}
	accountOnly, err := CanonicalResourceKey(ResourceIdentity{
		Provider: "aws", Account: "123456789012", Type: "deploy", ID: "production",
	})
	if err != nil || accountOnly.Account != "123456789012" || accountOnly.Tenant != "" {
		t.Fatalf("account-scoped resource failed: key=%#v err=%v", accountOnly, err)
	}
}

func TestSemanticClaimCASTakeoverAndRollbackFence(t *testing.T) {
	now := time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC)
	key := mustResource(t, "publisher", "widget")
	clock := now
	store := NewMemoryClaimStoreWithClock(func() time.Time { return clock })
	if err := store.SetController("controller-a", 7); err != nil {
		t.Fatal(err)
	}
	first, err := store.Acquire(ClaimRequest{
		Key: key, Holder: "session-a", WorkID: "release-widget", SourceRevision: "abc",
		AdapterID: "memory-reference", WorkloadIdentity: "worker-a", ControllerID: "controller-a",
		ControllerEpoch: 7, Now: now, TTL: time.Minute,
	})
	if err != nil {
		t.Fatal(err)
	}
	staleSnapshot := store.Snapshot()
	if _, err := store.Takeover(ClaimRequest{
		Key: key, Holder: "session-b", WorkID: "release-widget", SourceRevision: "def",
		AdapterID: "memory-reference", WorkloadIdentity: "worker-b", ControllerID: "controller-a",
		ControllerEpoch: 7, Now: now.Add(2 * time.Minute), TTL: time.Minute,
	}, ""); err == nil {
		t.Fatal("expiry-only takeover was accepted")
	}
	if _, err := store.Takeover(ClaimRequest{
		Key: key, Holder: "session-b", WorkID: "release-widget", SourceRevision: "def",
		AdapterID: "memory-reference", WorkloadIdentity: "worker-b", ControllerID: "controller-a",
		ControllerEpoch: 7, Now: now.Add(2 * time.Minute), TTL: time.Minute,
	}, first.ClaimRevision); err == nil {
		t.Fatal("caller-controlled time authorized an early takeover")
	}
	clock = now.Add(2 * time.Minute)
	second, err := store.Takeover(ClaimRequest{
		Key: key, Holder: "session-b", WorkID: "release-widget", SourceRevision: "def",
		AdapterID: "memory-reference", WorkloadIdentity: "worker-b", ControllerID: "controller-a",
		ControllerEpoch: 7, Now: now.Add(2 * time.Minute), TTL: time.Minute,
	}, first.ClaimRevision)
	if err != nil {
		t.Fatal(err)
	}
	if second.Generation != first.Generation+1 || second.ClaimRevision == first.ClaimRevision {
		t.Fatalf("takeover did not advance both fences: %#v %#v", first, second)
	}
	if err := store.ValidateCurrent(first); err == nil {
		t.Fatal("stale worker remained current")
	}
	if err := store.Restore(staleSnapshot); err == nil {
		t.Fatal("same-epoch generation and revision rollback was accepted")
	}
	forgedCurrent := second
	forgedCurrent.Holder = "different-session"
	if err := store.ValidateCurrent(forgedCurrent); err == nil {
		t.Fatal("claim fields could be forged while retaining revision and generation")
	}
	malformedKey := mustResource(t, "publisher", "other")
	malformedKey.ID = "different"
	if _, err := store.Acquire(ClaimRequest{
		Key: malformedKey, Holder: "session-c", WorkID: "other", SourceRevision: "ghi",
		AdapterID: "memory-reference", WorkloadIdentity: "worker-c", ControllerID: "controller-a",
		ControllerEpoch: 7, Now: now, TTL: time.Minute,
	}); err == nil {
		t.Fatal("non-canonical resource key was accepted")
	}
	snapshot := store.Snapshot()
	if err := store.SetController("controller-b", 8); err != nil {
		t.Fatal(err)
	}
	if err := store.Restore(snapshot); err == nil {
		t.Fatal("rollback-era snapshot was accepted")
	}
	if err := store.SetController("controller-a", 7); err == nil {
		t.Fatal("old controller incarnation was accepted")
	}
}

func TestFenceGatewayAuthorizationIdempotencyAndBypass(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC)
	key := mustResource(t, "deploy", "production")
	claim := Claim{
		Key: key, Holder: "session-a", WorkID: "deploy-widget", ClaimRevision: "r1",
		Generation: 4, ControllerID: "controller-a", ControllerEpoch: 9,
		WorkloadIdentity: "worker-a", ExpiresAt: now.Add(time.Minute),
	}
	verifier := exactVerifier{token: "opaque-authority"}
	var calls int32
	gateway := newTestGateway(now, GatewayCapability{
		AdapterID: "reference", AtomicConditionalCommit: true,
		SolePath: SolePathCapability{GatewayPrincipalOnly: true, DirectPathDenied: true, AlternatePathDenied: true},
	}, verifier, func(_ context.Context, _ MutationRequest) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "deployment-42", nil
	})
	request := MutationRequest{
		Claim: claim, Audience: "deploy-gateway", Operation: "deploy",
		RequestDigest: "sha256:payload", IdempotencyKey: "deploy-widget-42", Now: now,
		Authority: ScopedAuthorization{
			Token: "opaque-authority", Audience: "deploy-gateway", ControllerID: "controller-a",
			ControllerEpoch: 9, WorkloadIdentity: "worker-a", WorkID: "deploy-widget",
			Operation: "deploy", Key: key.Canonical, Generation: 4,
			RequestDigest: "sha256:payload", IdempotencyKey: "deploy-widget-42",
			ExpiresAt: now.Add(time.Minute),
		},
	}
	if gateway.Capability().ProtectedMutationAvailable {
		t.Fatal("cold gateway advertised protected mutation before authoritative bootstrap")
	}
	bootstrapGateway(t, gateway, request.Claim)
	unseeded := validMutation(mustResource(t, "deploy", "staging"), now, "unseeded")
	if _, err := gateway.Commit(ctx, unseeded); !errors.Is(err, ErrProtectedMutationUnavailable) {
		t.Fatalf("unseeded resource was admitted by warm gateway: %v", err)
	}
	first, err := gateway.Commit(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := gateway.Commit(ctx, request)
	if err != nil {
		t.Fatal(err)
	}
	if first.Result != replay.Result || !replay.Replayed || calls != 1 {
		t.Fatalf("idempotent replay repeated side effect: first=%#v replay=%#v calls=%d", first, replay, calls)
	}
	mismatch := request
	mismatch.RequestDigest = "sha256:different"
	mismatch.Authority.RequestDigest = mismatch.RequestDigest
	if _, err := gateway.Commit(ctx, mismatch); !errors.Is(err, ErrIdempotencyMismatch) {
		t.Fatalf("payload mismatch error=%v", err)
	}
	wrongAudience := request
	wrongAudience.Audience = "other-gateway"
	if _, err := gateway.Commit(ctx, wrongAudience); err == nil {
		t.Fatal("wrong audience was accepted")
	}
	expired := request
	expired.IdempotencyKey = "expired-authority"
	expired.Authority.IdempotencyKey = expired.IdempotencyKey
	expired.Authority.ExpiresAt = now.Add(-time.Second)
	expired.Now = now.Add(-time.Hour)
	if _, err := gateway.Commit(ctx, expired); err == nil {
		t.Fatal("caller-controlled request time bypassed authority expiry")
	}
	forged := request
	forged.IdempotencyKey = "forged"
	forged.Authority.IdempotencyKey = "forged"
	forged.Authority.Token = "generation-4-is-not-a-credential"
	if _, err := gateway.Commit(ctx, forged); err == nil {
		t.Fatal("forged authority was accepted")
	}
	unavailable := newTestGateway(now, GatewayCapability{}, nil, nil)
	if unavailable.Capability().ProtectedMutationAvailable {
		t.Fatal("missing verifier/atomic/sole-path proof advertised protected mutation")
	}
	if _, err := unavailable.Commit(ctx, request); !errors.Is(err, ErrProtectedMutationUnavailable) {
		t.Fatalf("unavailable gateway error=%v", err)
	}
	weakVerifierGateway := newTestGateway(now, GatewayCapability{
		AtomicConditionalCommit: true,
		SolePath: SolePathCapability{
			GatewayPrincipalOnly: true, DirectPathDenied: true, AlternatePathDenied: true,
		},
	}, weakVerifier{}, func(context.Context, MutationRequest) (string, error) { return "unsafe", nil })
	if weakVerifierGateway.Capability().ProtectedMutationAvailable {
		t.Fatal("binding-only verifier advertised protected mutation")
	}
	stale := request
	stale.IdempotencyKey = "stale-generation"
	stale.Authority.IdempotencyKey = stale.IdempotencyKey
	stale.Claim.Generation = 3
	stale.Authority.Generation = 3
	if _, err := gateway.Commit(ctx, stale); !errors.Is(err, ErrStaleGeneration) {
		t.Fatalf("stale gateway generation error=%v", err)
	}
	rollbackEra := request
	rollbackEra.IdempotencyKey = "rollback-era"
	rollbackEra.Authority.IdempotencyKey = rollbackEra.IdempotencyKey
	rollbackEra.Claim.ControllerEpoch = 8
	rollbackEra.Authority.ControllerEpoch = 8
	rollbackEra.Claim.Generation = 5
	rollbackEra.Authority.Generation = 5
	if _, err := gateway.Commit(ctx, rollbackEra); !errors.Is(err, ErrStaleGeneration) {
		t.Fatalf("rollback-era gateway request error=%v", err)
	}
	revisionCollision := request
	revisionCollision.IdempotencyKey = "same-generation-other-claim"
	revisionCollision.Authority.IdempotencyKey = revisionCollision.IdempotencyKey
	revisionCollision.Claim.ClaimRevision = "r-other"
	if _, err := gateway.Commit(ctx, revisionCollision); !errors.Is(err, ErrStaleGeneration) {
		t.Fatalf("same-generation claim collision error=%v", err)
	}
	racy := newTestGateway(now, GatewayCapability{
		AdapterID: "non-atomic",
		SolePath: SolePathCapability{
			GatewayPrincipalOnly: true, DirectPathDenied: true, AlternatePathDenied: true,
		},
	}, verifier, func(context.Context, MutationRequest) (string, error) { return "unsafe", nil })
	if racy.Capability().ProtectedMutationAvailable {
		t.Fatal("validation-to-commit race was advertised as supported")
	}
	for name, solePath := range map[string]SolePathCapability{
		"direct": {
			GatewayPrincipalOnly: true, DirectPathDenied: false, AlternatePathDenied: true,
		},
		"alternate": {
			GatewayPrincipalOnly: true, DirectPathDenied: true, AlternatePathDenied: false,
		},
	} {
		bypass := newTestGateway(now, GatewayCapability{
			AtomicConditionalCommit: true, SolePath: solePath,
		}, verifier, func(context.Context, MutationRequest) (string, error) { return "unsafe", nil })
		if bypass.Capability().ProtectedMutationAvailable {
			t.Fatalf("%s bypass did not disable protected mutation", name)
		}
	}
}

func TestFenceGatewayAmbiguousRetryFailsClosed(t *testing.T) {
	now := time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC)
	key := mustResource(t, "publisher", "widget")
	request := validMutation(key, now, "publish-1")
	var calls int32
	gateway := newTestGateway(now, GatewayCapability{
		AdapterID: "reference", AtomicConditionalCommit: true,
		SolePath: SolePathCapability{GatewayPrincipalOnly: true, DirectPathDenied: true, AlternatePathDenied: true},
	}, exactVerifier{token: "opaque-authority"}, func(context.Context, MutationRequest) (string, error) {
		atomic.AddInt32(&calls, 1)
		return "", ErrCommitOutcomeUnknown
	})
	bootstrapGateway(t, gateway, request.Claim)
	if _, err := gateway.Commit(context.Background(), request); !errors.Is(err, ErrCommitOutcomeUnknown) {
		t.Fatalf("ambiguous commit error=%v", err)
	}
	if _, err := gateway.Commit(context.Background(), request); !errors.Is(err, ErrIdempotencyInFlight) {
		t.Fatalf("ambiguous retry error=%v", err)
	}
	successor := request
	successor.IdempotencyKey = "publish-successor"
	successor.Authority.IdempotencyKey = successor.IdempotencyKey
	successor.Claim.Generation++
	successor.Authority.Generation++
	successor.Claim.ClaimRevision = "r2"
	if _, err := gateway.Commit(context.Background(), successor); !errors.Is(err, ErrIdempotencyInFlight) {
		t.Fatalf("ambiguous predecessor allowed a second publisher: %v", err)
	}
	if record, ok := gateway.IdempotencyStatus(key.Canonical, request.IdempotencyKey); !ok || record.State != IdempotencyInFlight {
		t.Fatalf("ambiguous status=%#v ok=%v", record, ok)
	}
	if _, err := gateway.ResolveAmbiguous(context.Background(), key.Canonical, request.IdempotencyKey, nil); !errors.Is(err, ErrIdempotencyInFlight) {
		t.Fatalf("unauthorized recovery error=%v", err)
	}
	recovered, err := gateway.ResolveAmbiguous(
		context.Background(), key.Canonical, request.IdempotencyKey,
		staticRecovery{decision: IdempotencyRecoveryDecision{
			RequestDigest: request.RequestDigest,
			ClaimRevision: request.Claim.ClaimRevision,
			Generation:    request.Claim.Generation,
			State:         IdempotencyCommitted,
			Result:        "publication-42",
		}},
	)
	if err != nil || recovered.Result != "publication-42" || !recovered.Replayed {
		t.Fatalf("authoritative recovery receipt=%#v err=%v", recovered, err)
	}
	replay, err := gateway.Commit(context.Background(), request)
	if err != nil || replay.Result != recovered.Result || calls != 1 {
		t.Fatalf("recovery replay invoked publisher: replay=%#v calls=%d err=%v", replay, calls, err)
	}
}

func TestSemanticClaimDisjointConcurrencyAndSameResourceSerialization(t *testing.T) {
	now := time.Date(2026, 8, 1, 18, 0, 0, 0, time.UTC)
	var active int32
	var maxActive int32
	started := make(chan struct{}, 4)
	release := make(chan struct{})
	gateway := newTestGateway(now, GatewayCapability{
		AdapterID: "reference", AtomicConditionalCommit: true,
		SolePath: SolePathCapability{GatewayPrincipalOnly: true, DirectPathDenied: true, AlternatePathDenied: true},
	}, exactVerifier{token: "opaque-authority"}, func(context.Context, MutationRequest) (string, error) {
		current := atomic.AddInt32(&active, 1)
		for {
			old := atomic.LoadInt32(&maxActive)
			if current <= old || atomic.CompareAndSwapInt32(&maxActive, old, current) {
				break
			}
		}
		started <- struct{}{}
		<-release
		atomic.AddInt32(&active, -1)
		return "ok", nil
	})
	first := validMutation(mustResource(t, "publisher", "one"), now, "one")
	second := validMutation(mustResource(t, "publisher", "two"), now, "two")
	bootstrapGateway(t, gateway, first.Claim, second.Claim)
	var wg sync.WaitGroup
	for _, request := range []MutationRequest{first, second} {
		wg.Add(1)
		go func(request MutationRequest) {
			defer wg.Done()
			if _, err := gateway.Commit(context.Background(), request); err != nil {
				t.Errorf("disjoint commit: %v", err)
			}
		}(request)
	}
	<-started
	<-started
	close(release)
	wg.Wait()
	if maxActive < 2 {
		t.Fatal("disjoint resources did not run concurrently")
	}

	serialRelease := make(chan struct{}, 2)
	serialStarted := make(chan struct{}, 2)
	active = 0
	maxActive = 0
	serial := newTestGateway(now, gateway.Capability(), exactVerifier{token: "opaque-authority"}, func(context.Context, MutationRequest) (string, error) {
		current := atomic.AddInt32(&active, 1)
		if current > atomic.LoadInt32(&maxActive) {
			atomic.StoreInt32(&maxActive, current)
		}
		serialStarted <- struct{}{}
		<-serialRelease
		atomic.AddInt32(&active, -1)
		return "ok", nil
	})
	sameA := validMutation(first.Claim.Key, now, "same-a")
	sameB := validMutation(first.Claim.Key, now, "same-b")
	bootstrapGateway(t, serial, sameA.Claim)
	for _, request := range []MutationRequest{sameA, sameB} {
		wg.Add(1)
		go func(request MutationRequest) {
			defer wg.Done()
			_, _ = serial.Commit(context.Background(), request)
		}(request)
	}
	<-serialStarted
	select {
	case <-serialStarted:
		t.Fatal("same resource entered concurrently")
	case <-time.After(25 * time.Millisecond):
	}
	serialRelease <- struct{}{}
	<-serialStarted
	serialRelease <- struct{}{}
	wg.Wait()
	if maxActive != 1 {
		t.Fatalf("same resource max concurrency=%d", maxActive)
	}
}

type exactVerifier struct{ token string }

type weakVerifier struct{}

type staticRecovery struct {
	decision IdempotencyRecoveryDecision
	err      error
}

type exactBootstrapVerifier struct{}

func newTestGateway(
	now time.Time,
	capability GatewayCapability,
	verifier AuthorizationVerifier,
	commit SideEffectCommit,
) *ReferenceGateway {
	return NewReferenceGatewayWithClock(capability, verifier, commit, func() time.Time { return now })
}

func bootstrapGateway(t *testing.T, gateway *ReferenceGateway, claims ...Claim) {
	t.Helper()
	if len(claims) == 0 {
		t.Fatal("bootstrap requires current claims")
	}
	state := GatewayBootstrapState{
		SchemaVersion:    SemanticClaimSchemaVersion,
		ReconciliationID: "authoritative-test-bootstrap",
		ControllerID:     claims[0].ControllerID,
		ControllerEpoch:  claims[0].ControllerEpoch,
		CurrentClaims:    claims,
	}
	if err := gateway.Bootstrap(context.Background(), state, exactBootstrapVerifier{}); err != nil {
		t.Fatalf("bootstrap gateway: %v", err)
	}
}

func (exactBootstrapVerifier) Capability() GatewayBootstrapVerifierCapability {
	return GatewayBootstrapVerifierCapability{
		AuthenticatedAuthority: true,
		RollbackProtected:      true,
	}
}

func (exactBootstrapVerifier) VerifyBootstrap(context.Context, GatewayBootstrapState) error {
	return nil
}

func (r staticRecovery) Resolve(context.Context, IdempotencyRecord) (IdempotencyRecoveryDecision, error) {
	return r.decision, r.err
}

func (weakVerifier) Capability() AuthorizationVerifierCapability {
	return AuthorizationVerifierCapability{}
}
func (weakVerifier) Verify(context.Context, ScopedAuthorization, MutationRequest) error {
	return nil
}

func (v exactVerifier) Capability() AuthorizationVerifierCapability {
	return AuthorizationVerifierCapability{
		AuthenticatedAuthority: true,
		AudienceBound:          true,
		WorkloadBound:          true,
		RequestDigestBound:     true,
		ExpiryEnforced:         true,
	}
}

func (v exactVerifier) Verify(_ context.Context, authorization ScopedAuthorization, request MutationRequest) error {
	if authorization.Token != v.token {
		return ErrAuthorityInvalid
	}
	return ValidateAuthorizationBinding(authorization, request)
}

func mustResource(t *testing.T, kind, id string) ResourceKey {
	t.Helper()
	key, err := CanonicalResourceKey(ResourceIdentity{
		Provider: "example", Tenant: "tenant", Type: kind, ID: id,
	})
	if err != nil {
		t.Fatal(err)
	}
	return key
}

func validMutation(key ResourceKey, now time.Time, idempotency string) MutationRequest {
	claim := Claim{
		Key: key, Holder: "session-a", WorkID: "work", ClaimRevision: "r1",
		Generation: 1, ControllerID: "controller-a", ControllerEpoch: 1,
		WorkloadIdentity: "worker-a", ExpiresAt: now.Add(time.Minute),
	}
	return MutationRequest{
		Claim: claim, Audience: "gateway", Operation: "publish", RequestDigest: "sha256:payload",
		IdempotencyKey: idempotency, Now: now,
		Authority: ScopedAuthorization{
			Token: "opaque-authority", Audience: "gateway", ControllerID: "controller-a",
			ControllerEpoch: 1, WorkloadIdentity: "worker-a", WorkID: "work",
			Operation: "publish", Key: key.Canonical, Generation: 1,
			RequestDigest: "sha256:payload", IdempotencyKey: idempotency,
			ExpiresAt: now.Add(time.Minute),
		},
	}
}
