package modelrouting

import (
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSelectRouteRespectsPlannedAndAttemptTiersOverridesAndEvidenceStrength(t *testing.T) {
	now := fixedNow()
	mediumA := provenRoute("medium-a", ClassMedium, "openai", "codex", "named-agent", "gpt-medium", "code", now.Add(time.Hour))
	mediumB := provenRoute("medium-b", ClassMedium, "openai", "codex", "named-agent", "gpt-medium-b", "code", now.Add(2*time.Hour))
	mediumB = adapterPriorQualifiedRoute(mediumB)
	large := provenRoute("large-a", ClassLarge, "openai", "codex", "named-agent", "gpt-large", "code", now.Add(time.Hour))
	large = adapterPriorQualifiedRoute(large) // not stronger than the best Medium route
	catalog := catalogWithCurrent(now, []Route{
		mediumA,
		mediumB,
		large,
		provenRoute("small-a", ClassSmall, "openai", "codex", "named-agent", "gpt-small", "code", now.Add(time.Hour)),
	})
	req := broadRequest(TierMedium)
	policy := publicPolicy()

	decision, err := selectForTest(t, catalog, req, policy, RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	assertAliases(t, decision, []string{"medium-a", "medium-b"})
	if decision.PlannedTier != TierMedium || decision.AttemptTier != TierMedium {
		t.Fatalf("decision tiers planned=%q attempt=%q", decision.PlannedTier, decision.AttemptTier)
	}

	// A lower tier is considered only when the caller explicitly marks this
	// bounded packet for that attempt. Task family alone never lowers the floor.
	withAttempt := req
	withAttempt.AttemptTier = TierSmall
	decision, err = selectForTest(t, catalog, withAttempt, policy, RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("small attempt: %v", err)
	}
	assertAliases(t, decision, []string{"small-a"})
	if decision.PlannedTier != TierMedium || decision.AttemptTier != TierSmall {
		t.Fatalf("attempt decision tiers planned=%q attempt=%q", decision.PlannedTier, decision.AttemptTier)
	}
	decision, err = selectForTest(t, catalog, withAttempt, policy, RunOverride{Mode: OverrideUse, Alias: "small-a"}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("use eligible small attempt: %v", err)
	}
	assertAliases(t, decision, []string{"small-a"})

	// `use` is only a preference within the fully eligible automatic set. A
	// declared route cannot bypass proof, tools, context, risk, or the tier floor.
	attended := declaredRoute("attended", ClassSmall, "openai", "codex", "named-agent", "manual-model", "code")
	catalog.Routes = append(catalog.Routes, attended)
	decision, err = selectForTest(t, catalog, req, policy, RunOverride{Mode: OverrideUse, Alias: "attended"}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("use attended: %v", err)
	}
	assertAliases(t, decision, []string{"medium-a", "medium-b"})

	// `require` remains the exact attended pin and may select a trusted route
	// without automatic capability evidence.
	decision, err = selectForTest(t, catalog, req, policy, RunOverride{Mode: OverrideRequire, Alias: "attended"}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("require attended: %v", err)
	}
	assertAliases(t, decision, []string{"attended"})

	decision, err = selectForTest(t, catalog, req, policy, RunOverride{Mode: OverrideRequire, Alias: "missing"}, AttemptLedger{}, now)
	if !errors.Is(err, ErrRequiredRouteUnavailable) || decision.Status != SelectionUnavailable {
		t.Fatalf("missing require decision=%#v err=%v", decision, err)
	}

	decision, err = selectForTest(t, catalog, req, policy, RunOverride{Mode: OverrideIgnore}, AttemptLedger{}, now)
	if err != nil || decision.Status != SelectionIgnored || decision.Current.ModelID != "current-gpt" || len(decision.Routes) != 0 {
		t.Fatalf("ignore decision=%#v err=%v", decision, err)
	}

	// A stronger exact-match Large receipt is a qualified upward fallback.
	large.Capability.Source = EvidenceKBReceipt
	large.Capability.DispatchProven = true
	large.Readiness = append(large.Readiness, ReadinessDispatchProven)
	large.Capability.ContextSize = 16384
	mediumA = adapterPriorQualifiedRoute(mediumA)
	catalog.Routes[0], catalog.Routes[2] = mediumA, large
	decision, err = selectForTest(t, catalog, req, policy, RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("qualified escalation: %v", err)
	}
	assertAliases(t, decision, []string{"medium-b", "medium-a", "large-a"})
}

func TestSelectRouteAcceptsOnlyTheExactNextLowerAttemptTier(t *testing.T) {
	now := fixedNow()
	for name, tiers := range map[string][2]Tier{
		"equal":       {TierMedium, TierMedium},
		"skips tier":  {TierLarge, TierSmall},
		"small tries": {TierSmall, TierTiny},
		"above":       {TierSmall, TierMedium},
	} {
		t.Run(name, func(t *testing.T) {
			req := broadRequest(tiers[0])
			req.AttemptTier = tiers[1]
			decision, err := selectForTest(t, catalogWithCurrent(now, nil), req, publicPolicy(), RunOverride{}, AttemptLedger{}, now)
			if !errors.Is(err, ErrInvalidWorkRequest) || decision.Status != SelectionUnavailable {
				t.Fatalf("decision=%#v err=%v", decision, err)
			}
			if decision.PlannedTier != tiers[0] || decision.AttemptTier != tiers[1] {
				t.Fatalf("invalid decision lost tier metadata: %#v", decision)
			}
		})
	}
}

func TestSelectRoutePreferenceReordersOnlyAlreadyEligibleSameTierRoutes(t *testing.T) {
	now := fixedNow()
	local := provenRoute("local", ClassMedium, "local-lan", "codex", "named-agent", "local-model", "code", now.Add(time.Hour))
	local.Boundary = BoundaryPrivate
	hosted := provenRoute("hosted", ClassMedium, "openai", "codex", "named-agent", "hosted-model", "code", now.Add(2*time.Hour))
	policy := publicPolicy()
	local.ManagementOrigin = OriginExtra
	local.Hosting = HostingSelfHosted
	hosted.ManagementOrigin = OriginNative
	hosted.Hosting = HostingProviderHosted
	fingerprint, err := ComputeRouteFingerprint(local)
	if err != nil {
		t.Fatal(err)
	}
	policy.Trusted.RouteApprovals = []RouteApproval{{ProjectID: "project-a", RouteFingerprint: fingerprint, ExpiresAt: now.Add(time.Hour)}}
	decision, err := selectForTest(t, catalogWithCurrent(now, []Route{hosted, local}), broadRequest(TierMedium), policy, RunOverride{Prefer: PreferenceSelfHostedFirst}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"local", "hosted"})
	decision, err = selectForTest(t, catalogWithCurrent(now, []Route{hosted, local}), broadRequest(TierMedium), policy, RunOverride{Prefer: PreferenceNativeFirst}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"hosted", "local"})
}

func TestOriginHostingAndQualifiedProofRemainIndependent(t *testing.T) {
	now := fixedNow()
	privateProvider := provenRoute("private-provider", ClassMedium, "openai", "codex", "named-agent", "provider-model", "code", now.Add(2*time.Hour))
	privateProvider.Boundary = BoundaryPrivate
	privateProvider.ManagementOrigin = OriginExtra
	privateProvider.Hosting = HostingProviderHosted
	selfHosted := provenRoute("self-hosted", ClassMedium, "local-lan", "codex", "named-agent", "local-model", "code", now.Add(time.Hour))
	selfHosted.ManagementOrigin = OriginExtra
	selfHosted.Hosting = HostingSelfHosted
	native := provenRoute("native", ClassMedium, "openai", "codex", "named-agent", "native-model", "code", now.Add(3*time.Hour))
	native.ManagementOrigin = OriginNative
	native.Hosting = HostingProviderHosted

	policy := publicPolicy()
	for _, route := range []Route{privateProvider, selfHosted} {
		fingerprint, err := ComputeRouteFingerprint(route)
		if err != nil {
			t.Fatal(err)
		}
		policy.Trusted.RouteApprovals = append(policy.Trusted.RouteApprovals, RouteApproval{ProjectID: "project-a", RouteFingerprint: fingerprint, ExpiresAt: now.Add(time.Hour)})
	}
	catalog := catalogWithCurrent(now, []Route{privateProvider, selfHosted, native})
	decision, err := selectForTest(t, catalog, broadRequest(TierMedium), policy, RunOverride{Prefer: PreferenceSelfHostedFirst}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"self-hosted", "native", "private-provider"})
	decision, err = selectForTest(t, catalog, broadRequest(TierMedium), policy, RunOverride{Prefer: PreferenceNativeFirst}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"native", "private-provider", "self-hosted"})

	qualified := declaredRoute("qualified", ClassMedium, "openai", "codex", "named-agent", "qualified-model", "code")
	qualified.ManagementOrigin = OriginNative
	qualified.Hosting = HostingProviderHosted
	qualified.Capability.Source = EvidenceAdapterPrior
	qualified.Capability.DispatchQualified = true
	qualified.Capability.ExpiresAt = now.Add(time.Hour)
	decision, err = selectForTest(t, catalogWithCurrent(now, []Route{qualified}), broadRequest(TierMedium), publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"qualified"})
	if decision.Routes[0].Capability.DispatchProven {
		t.Fatal("adapter qualification must not claim exact route-bound dispatch proof")
	}
}

func TestPlannerClassRequiresExplicitUseOrRequire(t *testing.T) {
	now := fixedNow()
	large := provenRoute("large", ClassLarge, "openai", "codex", "named-agent", "large-model", "code", now.Add(time.Hour))
	planner := provenRoute("planner", ClassPlanner, "openai", "codex", "named-agent", "planner-model", "code", now.Add(2*time.Hour))
	catalog := catalogWithCurrent(now, []Route{planner, large})
	req := broadRequest(TierLarge)
	policy := publicPolicy()

	decision, err := selectForTest(t, catalog, req, policy, RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"large"})
	decision, err = selectForTest(t, catalog, req, policy, RunOverride{Mode: OverrideUse, Alias: "planner"}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"planner", "large"})
	decision, err = selectForTest(t, catalog, req, policy, RunOverride{Mode: OverrideRequire, Alias: "planner"}, AttemptLedger{}, now)
	if err != nil {
		t.Fatal(err)
	}
	assertAliases(t, decision, []string{"planner"})
}

func TestUseCannotBypassTierToolsContextOrRisk(t *testing.T) {
	now := fixedNow()
	good := provenRoute("good", ClassMedium, "openai", "codex", "named-agent", "good-model", "code", now.Add(time.Hour))
	belowTier := provenRoute("below-tier", ClassSmall, "openai", "codex", "named-agent", "small-model", "code", now.Add(time.Hour))
	wrongTools := provenRoute("wrong-tools", ClassMedium, "openai", "codex", "named-agent", "tools-model", "code", now.Add(time.Hour))
	wrongTools.Capability.Tools = []string{"read"}
	shortContext := provenRoute("short-context", ClassMedium, "openai", "codex", "named-agent", "context-model", "code", now.Add(time.Hour))
	shortContext.Capability.ContextSize = 1024
	narrowRisk := provenRoute("narrow-risk", ClassMedium, "openai", "codex", "named-agent", "risk-model", "code", now.Add(time.Hour))
	narrowRisk.Capability.Risk = RiskNormal
	catalog := catalogWithCurrent(now, []Route{belowTier, wrongTools, shortContext, narrowRisk, good})
	req := broadRequest(TierMedium)
	for _, alias := range []string{"below-tier", "wrong-tools", "short-context", "narrow-risk"} {
		decision, err := selectForTest(t, catalog, req, publicPolicy(), RunOverride{Mode: OverrideUse, Alias: alias}, AttemptLedger{}, now)
		if err != nil {
			t.Fatalf("use %s: %v", alias, err)
		}
		assertAliases(t, decision, []string{"good"})
	}
}

func TestAutomaticEligibilityFailsClosedOnUnknownEvidenceAndTrust(t *testing.T) {
	now := fixedNow()
	good := provenRoute("good", ClassMedium, "openai", "codex", "named-agent", "good-model", "code", now.Add(time.Hour))
	missingFamily := good
	missingFamily.Alias, missingFamily.DisplayModelID = "missing-family", "missing-family"
	missingFamily.Capability.RouteAlias, missingFamily.Capability.ModelID, missingFamily.Capability.TaskFamily = missingFamily.Alias, missingFamily.DisplayModelID, ""
	missingContext := good
	missingContext.Alias, missingContext.DisplayModelID = "missing-context", "missing-context"
	missingContext.Capability.RouteAlias, missingContext.Capability.ModelID, missingContext.Capability.ContextSize = missingContext.Alias, missingContext.DisplayModelID, 0
	missingRisk := good
	missingRisk.Alias, missingRisk.DisplayModelID = "missing-risk", "missing-risk"
	missingRisk.Capability.RouteAlias, missingRisk.Capability.ModelID, missingRisk.Capability.Risk = missingRisk.Alias, missingRisk.DisplayModelID, ""
	unknownRetention := good
	unknownRetention.Alias, unknownRetention.DisplayModelID, unknownRetention.Retention = "unknown-retention", "unknown-retention", RetentionUnknown
	unknownRetention.Capability.RouteAlias, unknownRetention.Capability.ModelID = unknownRetention.Alias, unknownRetention.DisplayModelID
	notCumulative := good
	notCumulative.Alias, notCumulative.DisplayModelID = "not-cumulative", "not-cumulative"
	notCumulative.Capability.RouteAlias, notCumulative.Capability.ModelID = notCumulative.Alias, notCumulative.DisplayModelID
	notCumulative.Readiness = []Readiness{ReadinessDispatchProven}
	downward := provenRoute("small", ClassSmall, "openai", "codex", "named-agent", "small", "code", now.Add(time.Hour))

	decision, err := selectForTest(t, catalogWithCurrent(now, []Route{missingFamily, missingContext, missingRisk, unknownRetention, notCumulative, downward, good}), broadRequest(TierMedium), publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	assertAliases(t, decision, []string{"good"})

	sensitive := broadRequest(TierMedium)
	sensitive.SensitiveData = true
	missingDestinationApproval := publicPolicy()
	missingDestinationApproval.Project.AllowedDestinations = nil
	decision, _ = selectForTest(t, catalogWithCurrent(now, []Route{good}), sensitive, missingDestinationApproval, RunOverride{}, AttemptLedger{}, now)
	assertAliases(t, decision, nil)
	trainingUnknown := good
	trainingUnknown.TrainingUse = TrainingUnknown
	decision, _ = selectForTest(t, catalogWithCurrent(now, []Route{trainingUnknown}), sensitive, publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	assertAliases(t, decision, nil)
	missingRetentionApproval := publicPolicy()
	missingRetentionApproval.Project.MaxRetention = ""
	decision, _ = selectForTest(t, catalogWithCurrent(now, []Route{good}), sensitive, missingRetentionApproval, RunOverride{}, AttemptLedger{}, now)
	assertAliases(t, decision, nil)
}

func TestSelectionRequiresValidatedCatalogAndRejectsUnsafeEndpointCandidates(t *testing.T) {
	now := fixedNow()
	unsafe := provenRoute("unsafe", ClassMedium, "local-lan", "openai-compatible", "chat-completions", "bad", "code", now.Add(time.Hour))
	unsafe.Endpoint = "https://169.254.169.254/latest/meta-data"
	unsafe.Boundary = BoundaryPrivate
	raw := catalogWithCurrent(now, []Route{unsafe})
	sealCatalogForTest(t, &raw)
	validated, rejections, err := ValidateCatalogForSelection(raw, publicPolicy(), publicResolver(), now, CatalogSourceRun)
	if err != nil || len(rejections) != 1 {
		t.Fatalf("validated=%#v rejections=%#v err=%v", validated, rejections, err)
	}
	decision, err := SelectRoute(validated, broadRequest(TierMedium), publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	if err != nil || len(decision.Routes) != 0 || decision.Status != SelectionDegraded {
		t.Fatalf("unsafe candidate decision=%#v err=%v", decision, err)
	}
}

func TestProjectPolicyNarrowsAndTrustedApprovalCannotBeForgedByProject(t *testing.T) {
	now := fixedNow()
	private := provenRoute("local-qwen", ClassMedium, "local-lan", "litellm", "openai-compatible", "qwen", "code", now.Add(time.Hour))
	private.Boundary = BoundaryPrivate
	public := provenRoute("hosted-medium", ClassMedium, "openai", "codex", "named-agent", "gpt-medium", "code", now.Add(time.Hour))
	catalog := catalogWithCurrent(now, []Route{private, public})
	req := broadRequest(TierMedium)
	policy := PolicyContext{Project: ProjectPolicy{
		ProjectID:           "project-a",
		AllowedDestinations: []string{"local-lan", "openai"},
		AllowedAliases:      []string{"local-qwen"},
		MaxRetention:        RetentionSession,
	}}

	decision, err := selectForTest(t, catalog, req, policy, RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	assertAliases(t, decision, nil) // alias narrowing excludes public; tracked policy cannot approve private

	fingerprint, err := ComputeRouteFingerprint(private)
	if err != nil {
		t.Fatalf("route fingerprint: %v", err)
	}
	policy.Trusted = UserTrust{ProjectID: "project-a", RouteApprovals: []RouteApproval{{ProjectID: "project-a", RouteFingerprint: fingerprint, ExpiresAt: now.Add(time.Hour)}}}
	decision, err = selectForTest(t, catalog, req, policy, RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("approved select: %v", err)
	}
	assertAliases(t, decision, []string{"local-qwen"})

	private.Endpoint = "https://192.168.1.99/v1" // approval must not survive route replacement
	catalog.Routes[0] = private
	decision, _ = selectForTest(t, catalog, req, policy, RunOverride{}, AttemptLedger{}, now)
	assertAliases(t, decision, nil)

	policy.Project.ProjectID = "other-project"
	decision, _ = selectForTest(t, catalog, req, policy, RunOverride{}, AttemptLedger{}, now)
	assertAliases(t, decision, nil)
}

func TestAttemptLedgerIsAlwaysFiniteAndRejectsEmptyOrRepeatedAttempts(t *testing.T) {
	ledger := NewAttemptLedger(2)
	if err := ledger.Record(""); !errors.Is(err, ErrInvalidAttempt) {
		t.Fatalf("empty attempt error=%v", err)
	}
	if err := ledger.Record("a"); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Record("a"); !errors.Is(err, ErrRouteAlreadyAttempted) {
		t.Fatalf("repeat error=%v", err)
	}
	if err := ledger.Record("b"); err != nil {
		t.Fatal(err)
	}
	if err := ledger.Record("c"); !errors.Is(err, ErrAttemptLedgerFull) {
		t.Fatalf("full error=%v", err)
	}
	defaulted := NewAttemptLedger(0)
	for index := 0; index < defaultAttemptLimit; index++ {
		if err := defaulted.Record(string(rune('a' + index))); err != nil {
			t.Fatalf("default attempt %d: %v", index, err)
		}
	}
	if err := defaulted.Record("overflow"); !errors.Is(err, ErrAttemptLedgerFull) {
		t.Fatalf("default ledger was not finite: %v", err)
	}
}

func TestReceiptCreditRequiresExactFreshTaskAndProofBinding(t *testing.T) {
	now := fixedNow()
	receipt, envelope := exactReceipt(now)
	credit, observation := EvaluateReceiptCredit(receipt, envelope, now)
	if !credit || observation.Status != ObservationCredited {
		t.Fatalf("exact credit=%v observation=%#v", credit, observation)
	}

	cases := map[string]func(*RoutingReceipt){
		"model mismatch":   func(r *RoutingReceipt) { r.RouteEvidence.ProviderReportedModel = "other" },
		"task mismatch":    func(r *RoutingReceipt) { r.RouteEvidence.TaskFamily = "docs" },
		"missing session":  func(r *RoutingReceipt) { r.RouteEvidence.SessionID = "" },
		"wrong run":        func(r *RoutingReceipt) { r.RouteEvidence.RunID = "other" },
		"wrong packet":     func(r *RoutingReceipt) { r.RouteEvidence.ContextPacketHash = "other" },
		"wrong capability": func(r *RoutingReceipt) { r.RouteEvidence.CapabilityEnvelopeHash = "other" },
		"wrong proof":      func(r *RoutingReceipt) { r.WorkProof.ArtifactHash = "other" },
		"stale":            func(r *RoutingReceipt) { r.RouteEvidence.ObservedAt = now.Add(-2 * time.Hour) },
		"future":           func(r *RoutingReceipt) { r.RouteEvidence.ObservedAt = now.Add(time.Minute) },
		"failed proof":     func(r *RoutingReceipt) { r.WorkProof.Result = ProofFail },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			candidate := receipt
			mutate(&candidate)
			credit, observation := EvaluateReceiptCredit(candidate, envelope, now)
			if credit || observation.Status != ObservationOnly {
				t.Fatalf("credit=%v observation=%#v", credit, observation)
			}
		})
	}
}

func TestStrictStorageRejectsCredentialsUnknownDuplicateAndUnsafePaths(t *testing.T) {
	root := t.TempDir()
	now := fixedNow()
	catalog := catalogWithCurrent(now, []Route{provenRoute("medium-a", ClassMedium, "openai", "codex", "named-agent", "gpt-medium", "code", now.Add(time.Hour))})
	policy := publicPolicy()
	opts := StorageOptions{MaxBytes: 64 * 1024, Resolver: publicResolver(), Now: now, Policy: policy, Source: CatalogSourceRun}
	sealCatalogForTest(t, &catalog)
	if err := SaveCatalog(root, "catalog.json", catalog, opts); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := os.Chmod(filepath.Join(root, "catalog.json"), 0o600); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	before, _ := os.Stat(filepath.Join(root, "catalog.json"))
	catalog.Current.ModelID = "current-2"
	catalog.Current.Route.DisplayModelID = "current-2"
	catalog.Current.Route.Capability.ModelID = "current-2"
	sealCatalogForTest(t, &catalog)
	if err := SaveCatalog(root, "catalog.json", catalog, opts); err != nil {
		t.Fatalf("resave: %v", err)
	}
	after, _ := os.Stat(filepath.Join(root, "catalog.json"))
	if after.Mode().Perm() != before.Mode().Perm() {
		t.Fatalf("mode after rewrite=%#o want=%#o", after.Mode().Perm(), before.Mode().Perm())
	}
	loaded, err := LoadCatalog(root, "catalog.json", opts)
	if err != nil || loaded.Current.ModelID != "current-2" {
		t.Fatalf("load=%#v err=%v", loaded, err)
	}

	invalid := map[string]string{
		"credential.json":    `{"schema_version":1,"current":{},"routes":[],"auth_value":"secret"}`,
		"duplicate-key.json": `{"schema_version":1,"schema_version":1,"current":{},"routes":[]}`,
		"trailing.json":      `{"schema_version":1,"current":{},"routes":[]} {}`,
	}
	for name, body := range invalid {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(root, name, opts); err == nil {
			t.Fatalf("%s unexpectedly loaded", name)
		}
	}

	duplicate := catalog
	duplicate.Routes = append(duplicate.Routes, duplicate.Routes[0])
	sealCatalogForTest(t, &duplicate)
	if err := SaveCatalog(root, "duplicate-alias.json", duplicate, opts); !errors.Is(err, ErrDuplicateAlias) {
		t.Fatalf("duplicate alias error=%v", err)
	}
	if err := SaveCatalog(root, `..\escape.json`, catalog, opts); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("traversal error=%v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "too-large.json"), make([]byte, 1024), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := secureStorageFileSecurity(filepath.Join(root, "too-large.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(root, "too-large.json", StorageOptions{MaxBytes: 128, Policy: policy}); !errors.Is(err, ErrStorageSizeExceeded) {
		t.Fatalf("size error=%v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "linked"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "catalog.json"), filepath.Join(root, "linked", "catalog.json")); err == nil {
		if err := SaveCatalog(root, filepath.Join("linked", "catalog.json"), catalog, opts); !errors.Is(err, ErrUnsafePath) {
			t.Fatalf("symlink error=%v", err)
		}
	}
}

func TestEndpointValidationReturnsPinnedIPsAndRequiresPrivateApproval(t *testing.T) {
	now := fixedNow()
	base := provenRoute("route", ClassMedium, "openai", "openai-compatible", "chat-completions", "gpt", "code", now.Add(time.Hour))
	base.Endpoint, base.AuthEnv = "https://api.openai.com/v1", "OPENAI_API_KEY"
	policy := publicPolicy()
	policy.Trusted.AuthBindings = []AuthBinding{{Env: "OPENAI_API_KEY", Adapter: "openai-compatible", Origin: "https://api.openai.com", ExpiresAt: now.Add(time.Hour)}}
	validated, err := ValidateEndpoint(base, policy, publicResolver(), now)
	if err != nil || validated.Origin != "https://api.openai.com" || len(validated.PinnedIPs) != 1 {
		t.Fatalf("validated=%#v err=%v", validated, err)
	}

	base.Endpoint = "http://api.openai.com/v1"
	if _, err := ValidateEndpoint(base, policy, publicResolver(), now); !errors.Is(err, ErrUnsafeEndpoint) {
		t.Fatalf("public HTTP error=%v", err)
	}
	base.Endpoint = "https://169.254.169.254/latest/meta-data"
	if _, err := ValidateEndpoint(base, policy, publicResolver(), now); !errors.Is(err, ErrUnsafeEndpoint) {
		t.Fatalf("metadata error=%v", err)
	}
	base.Endpoint, base.Destination, base.AuthEnv = "https://192.168.1.205:4000/v1", "local-lan", "LOCAL_KEY"
	base.Boundary = BoundaryPrivate
	policy.Trusted.AuthBindings = []AuthBinding{{Env: "LOCAL_KEY", Adapter: "openai-compatible", Origin: "https://192.168.1.205:4000", ExpiresAt: now.Add(time.Hour)}}
	if _, err := ValidateEndpoint(base, policy, publicResolver(), now); !errors.Is(err, ErrPrivateEndpointRequiresApproval) {
		t.Fatalf("private TLS without approval error=%v", err)
	}
	policy.Trusted.EndpointApprovals = []EndpointApproval{{Origin: "https://192.168.1.205:4000", ProjectID: "project-a", ExpiresAt: now.Add(time.Hour)}}
	if _, err := ValidateEndpoint(base, policy, publicResolver(), now); err != nil {
		t.Fatalf("approved private TLS: %v", err)
	}
	base.Endpoint = "https://rebinding.example/v1"
	base.AuthEnv = ""
	resolver := StaticResolver(map[string][]net.IP{"rebinding.example": {net.ParseIP("104.18.0.3"), net.ParseIP("169.254.169.254")}})
	if _, err := ValidateEndpoint(base, policy, resolver, now); !errors.Is(err, ErrUnsafeEndpoint) {
		t.Fatalf("mixed DNS error=%v", err)
	}
	base.Endpoint = "https://api.openai.com/v1?token=secret"
	if _, err := ValidateEndpoint(base, policy, publicResolver(), now); !errors.Is(err, ErrUnsafeEndpoint) {
		t.Fatalf("query credential error=%v", err)
	}
}

func TestStorageRoundTripsAuthenticatedAndPrivateRoutesWithTrustedLocalBindings(t *testing.T) {
	root := t.TempDir()
	now := fixedNow()
	public := provenRoute("public-auth", ClassMedium, "openai", "openai-compatible", "chat-completions", "gpt", "code", now.Add(time.Hour))
	public.Endpoint, public.AuthEnv = "https://api.openai.com/v1", "OPENAI_API_KEY"
	private := provenRoute("private-auth", ClassMedium, "local-lan", "openai-compatible", "chat-completions", "qwen", "code", now.Add(time.Hour))
	private.Endpoint, private.AuthEnv, private.Boundary = "http://192.168.1.205:4000/v1", "LOCAL_KEY", BoundaryPrivate
	policy := publicPolicy()
	policy.Trusted.AuthBindings = []AuthBinding{
		{Env: "OPENAI_API_KEY", Adapter: "openai-compatible", Origin: "https://api.openai.com", ExpiresAt: now.Add(time.Hour)},
		{Env: "LOCAL_KEY", Adapter: "openai-compatible", Origin: "http://192.168.1.205:4000", ExpiresAt: now.Add(time.Hour)},
	}
	// This is trusted user-local endpoint consent for persistence. Project route
	// activation remains separately gated by a route fingerprint approval.
	policy.Trusted.EndpointApprovals = []EndpointApproval{{Origin: "http://192.168.1.205:4000", ExpiresAt: now.Add(time.Hour)}}
	opts := StorageOptions{MaxBytes: 64 * 1024, Resolver: publicResolver(), Now: now, Policy: policy, Source: CatalogSourceRun}
	catalog := catalogWithCurrent(now, []Route{public, private})
	sealCatalogForTest(t, &catalog)
	if err := SaveCatalog(root, "catalog.json", catalog, opts); err != nil {
		t.Fatalf("save authenticated routes: %v", err)
	}
	loaded, err := LoadCatalog(root, "catalog.json", opts)
	if err != nil || len(loaded.Routes) != 2 {
		t.Fatalf("load authenticated routes=%#v err=%v", loaded, err)
	}
}

type deadlineResolver struct {
	calls atomic.Int32
}

func (r *deadlineResolver) LookupIP(ctx context.Context, _ string) ([]net.IP, error) {
	r.calls.Add(1)
	<-ctx.Done()
	return nil, ctx.Err()
}

func TestCatalogSelectionUsesOneCallerDeadlineAndStorageDoesNotResolveDNS(t *testing.T) {
	now := fixedNow()
	route := declaredRoute("bounded", ClassMedium, "hosted", "openai-compatible", "chat-completions", "bounded-model", "code")
	route.Endpoint = "https://bounded.example.invalid/v1"
	catalog := catalogWithCurrent(now, []Route{route})
	sealCatalogForTest(t, &catalog)
	policy := trustCatalogRouteStates(publicPolicy(), catalog)
	resolver := &deadlineResolver{}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, rejections, err := ValidateCatalogForSelectionContext(ctx, catalog, policy, resolver, now, CatalogSourceRun)
	if err != nil || len(rejections) == 0 || time.Since(started) > 500*time.Millisecond {
		t.Fatalf("bounded validation elapsed=%s rejections=%#v err=%v", time.Since(started), rejections, err)
	}
	if resolver.calls.Load() != 1 {
		t.Fatalf("resolver calls=%d want=1", resolver.calls.Load())
	}

	root := t.TempDir()
	resolver.calls.Store(0)
	started = time.Now()
	if err := SaveCatalog(root, "catalog.json", catalog, StorageOptions{MaxBytes: 64 * 1024, Resolver: resolver, Now: now, Policy: policy, Source: CatalogSourceRun}); err != nil {
		t.Fatalf("static catalog save: %v", err)
	}
	if resolver.calls.Load() != 0 || time.Since(started) > 500*time.Millisecond {
		t.Fatalf("durable save performed DNS calls=%d elapsed=%s", resolver.calls.Load(), time.Since(started))
	}
}

func TestStrictStorageRejectsUnsafeLocalPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows DACL mutation is covered by the Windows-specific ACL test")
	}
	root := t.TempDir()
	value := map[string]int{"schema_version": 1}
	if err := SaveAtomicJSON(root, "private.json", value, 1024); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "private.json")
	if err := os.Chmod(path, 0o666); err != nil {
		t.Fatal(err)
	}
	var loaded map[string]int
	if err := LoadStrictJSON(root, "private.json", &loaded, 1024); !errors.Is(err, ErrUnsafePath) {
		t.Fatalf("unsafe file permissions error=%v", err)
	}
}

func TestProjectJSONPreservesRepositoryPermissions(t *testing.T) {
	root := t.TempDir()
	before, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	value := map[string]int{"schema_version": 1}
	if err := SaveAtomicProjectJSON(root, "kb-models.json", value, 1024); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if before.Mode().Perm() != after.Mode().Perm() {
		t.Fatalf("project save changed repository root mode from %#o to %#o", before.Mode().Perm(), after.Mode().Perm())
	}
	var loaded map[string]int
	if err := LoadStrictProjectJSON(root, "kb-models.json", &loaded, 1024); err != nil || loaded["schema_version"] != 1 {
		t.Fatalf("load normal project JSON=%#v err=%v", loaded, err)
	}
	if runtime.GOOS != "windows" {
		fileInfo, statErr := os.Stat(filepath.Join(root, "kb-models.json"))
		if statErr != nil {
			t.Fatal(statErr)
		}
		if fileInfo.Mode().Perm() != 0o644 {
			t.Fatalf("project file mode=%#o want=0644", fileInfo.Mode().Perm())
		}
	}
}

func TestUserCatalogCannotSelfPromoteCapabilityEvidence(t *testing.T) {
	root := t.TempDir()
	now := fixedNow()
	claimed := provenRoute("claimed", ClassPlanner, "openai", "codex", "named-agent", "claimed", "code", now.Add(time.Hour))
	userCatalog := Catalog{SchemaVersion: CatalogSchemaVersion, Routes: []Route{claimed}}
	opts := StorageOptions{MaxBytes: 64 * 1024, Resolver: publicResolver(), Now: now, Policy: publicPolicy(), Source: CatalogSourceUser}
	if err := SaveCatalog(root, "models.json", userCatalog, opts); err == nil {
		t.Fatal("user catalog self-promoted KB receipt evidence")
	}
	merged, err := MergeCatalog(Catalog{SchemaVersion: CatalogSchemaVersion}, userCatalog)
	if err != nil {
		t.Fatalf("merge user catalog: %v", err)
	}
	if got := merged.Routes[0]; got.Capability.Source != EvidenceDeclared || got.Capability.DispatchProven || hasReadiness(got.Readiness, ReadinessDispatchProven) {
		t.Fatalf("user route was not capped to declared/selectable: %#v", got)
	}
	forgedReference := declaredRoute("forged-source", ClassMedium, "openai", "codex", "exec-model", "gpt-medium", "code")
	forgedReference.RouteID = "route-id:" + strings.Repeat("a", 32)
	forgedReference.SourceRouteID = "route-id:" + strings.Repeat("b", 32)
	if err := ValidateCatalogStatic(Catalog{SchemaVersion: CatalogSchemaVersion, Routes: []Route{forgedReference}}, CatalogSourceUser); err == nil {
		t.Fatal("user catalog supplied its own trusted source-route reference")
	}
}

func TestCanonicalProjectIdentityStableAcrossLinkedWorktreesButNotClones(t *testing.T) {
	root := t.TempDir()
	common := filepath.Join(root, "repo.git")
	worktreeOneGit := filepath.Join(common, "worktrees", "one")
	worktreeTwoGit := filepath.Join(common, "worktrees", "two")
	for _, dir := range []string{worktreeOneGit, worktreeTwoGit} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "commondir"), []byte("../..\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	checkoutOne, checkoutTwo := filepath.Join(root, "one"), filepath.Join(root, "two")
	for checkout, gitdir := range map[string]string{checkoutOne: worktreeOneGit, checkoutTwo: worktreeTwoGit} {
		if err := os.MkdirAll(filepath.Join(checkout, "nested"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(checkout, ".git"), []byte("gitdir: "+gitdir+"\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	one, err := CanonicalProjectIdentity(filepath.Join(checkoutOne, "nested"))
	if err != nil {
		t.Fatal(err)
	}
	two, err := CanonicalProjectIdentity(checkoutTwo)
	if err != nil {
		t.Fatal(err)
	}
	if one != two {
		t.Fatalf("linked worktree identities differ: %q != %q", one, two)
	}
	alias := filepath.Join(t.TempDir(), "checkout-alias")
	if err := os.Symlink(checkoutOne, alias); err != nil && runtime.GOOS == "windows" {
		if output, junctionErr := exec.Command("cmd", "/c", "mklink", "/J", alias, checkoutOne).CombinedOutput(); junctionErr != nil {
			t.Fatalf("create directory alias: symlink=%v junction=%v output=%s", err, junctionErr, output)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	aliased, err := CanonicalProjectIdentity(filepath.Join(alias, "nested"))
	if err != nil {
		t.Fatal(err)
	}
	if aliased != one {
		t.Fatalf("symlink alias changed identity: %q != %q", aliased, one)
	}
	clone := filepath.Join(root, "clone")
	if err := os.MkdirAll(filepath.Join(clone, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	other, err := CanonicalProjectIdentity(clone)
	if err != nil {
		t.Fatal(err)
	}
	if one == other {
		t.Fatalf("unrelated clone inherited identity %q", one)
	}
	if err := os.Rename(filepath.Join(clone, ".git"), filepath.Join(clone, ".git-old")); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(clone, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	replacement, err := CanonicalProjectIdentity(clone)
	if err != nil {
		t.Fatal(err)
	}
	if replacement == other {
		t.Fatalf("replacement repository inherited identity %q", other)
	}
}

func TestCanonicalProjectIdentityStableAcrossRepositoryMove(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	before, err := CanonicalProjectIdentity(repo)
	if err != nil {
		t.Fatal(err)
	}
	moved := filepath.Join(parent, "renamed-repo")
	if err := os.Rename(repo, moved); err != nil {
		t.Fatal(err)
	}
	after, err := CanonicalProjectIdentity(moved)
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("repository move changed identity: %q != %q", after, before)
	}
}

func TestRouteSchemaRejectsDispatchProvenWithoutQualification(t *testing.T) {
	route := provenRoute("invalid-proof", ClassMedium, "openai", "codex", "named-agent", "gpt-medium", "code", fixedNow().Add(time.Hour))
	route.Capability.DispatchQualified = false
	if err := ValidateCatalogStatic(Catalog{SchemaVersion: CatalogSchemaVersion, Routes: []Route{route}}, CatalogSourceRun); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("dispatch-proven route without qualification error=%v", err)
	}
	route.Capability.DispatchQualified = true
	route.Capability.Source = EvidenceAdapterPrior
	if err := ValidateCatalogStatic(Catalog{SchemaVersion: CatalogSchemaVersion, Routes: []Route{route}}, CatalogSourceRun); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("dispatch-proven route without receipt evidence error=%v", err)
	}
}

func TestDefaultManagementMetadataDoesNotInvalidateLegacyRouteApproval(t *testing.T) {
	legacy := declaredRoute("legacy-extra", ClassMedium, "local-lan", "codex", "named-agent", "local-model", "code")
	legacy.ManagementOrigin = ""
	legacy.Hosting = ""
	legacy.DiscoverySources = nil
	legacyFingerprint, err := ComputeRouteFingerprint(legacy)
	if err != nil {
		t.Fatal(err)
	}

	defaulted := legacy
	defaulted.ManagementOrigin = OriginExtra
	defaulted.Hosting = HostingUnknown
	defaulted.DiscoverySources = []string{}
	defaultedFingerprint, err := ComputeRouteFingerprint(defaulted)
	if err != nil {
		t.Fatal(err)
	}
	if defaultedFingerprint != legacyFingerprint {
		t.Fatalf("selection-only metadata invalidated approval: defaulted=%q legacy=%q", defaultedFingerprint, legacyFingerprint)
	}
}

func TestCatalogMergeFingerprintAndCurrentFallbackPolicy(t *testing.T) {
	now := fixedNow()
	native := catalogWithCurrent(now, []Route{provenRoute("native", ClassMedium, "openai", "codex", "named-agent", "native", "code", now.Add(time.Hour))})
	native.Cohort = CohortInitialPilot
	native.Surfaces = []SurfaceFingerprint{{Surface: "codex-cli", Provider: "openai", Revision: "0.143", ConfigHash: "cfg"}}
	extra := Catalog{SchemaVersion: CatalogSchemaVersion, Routes: []Route{provenRoute("extra", ClassLarge, "openai", "codex", "named-agent", "extra", "code", now.Add(time.Hour))}}
	merged, err := MergeCatalog(native, extra)
	if err != nil || len(merged.Routes) != 2 || merged.Fingerprint == "" {
		t.Fatalf("merged=%#v err=%v", merged, err)
	}
	again, _ := ComputeCatalogFingerprint(merged)
	if again != merged.Fingerprint {
		t.Fatalf("fingerprint=%q recomputed=%q", merged.Fingerprint, again)
	}
	conflict := extra
	conflict.Routes[0].Alias = "native"
	if _, err := MergeCatalog(native, conflict); !errors.Is(err, ErrAliasConflict) {
		t.Fatalf("conflict error=%v", err)
	}

	noRoutes := catalogWithCurrent(now, nil)
	decision, err := selectForTest(t, noRoutes, broadRequest(TierMedium), publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	if err != nil || decision.Status != SelectionDegraded {
		t.Fatalf("current fallback decision=%#v err=%v", decision, err)
	}
	policy := publicPolicy()
	policy.Project.DenyCurrentFallback = true
	decision, _ = selectForTest(t, noRoutes, broadRequest(TierMedium), policy, RunOverride{}, AttemptLedger{}, now)
	if decision.Status != SelectionUnavailable {
		t.Fatalf("denied current fallback decision=%#v", decision)
	}
}

func TestRunFingerprintCoversEvidenceAndValidatedCatalogIsImmutable(t *testing.T) {
	now := fixedNow()
	route := provenRoute("medium", ClassMedium, "openai", "codex", "named-agent", "medium", "code", now.Add(time.Hour))
	catalog := catalogWithCurrent(now, []Route{route})
	if _, _, err := ValidateCatalogForSelection(catalog, publicPolicy(), publicResolver(), now, CatalogSourceRun); err == nil {
		t.Fatal("unfingerprinted run catalog was accepted")
	}
	sealCatalogForTest(t, &catalog)
	original := catalog.Fingerprint
	changed := catalog
	changed.Routes = append([]Route(nil), catalog.Routes...)
	changed.Routes[0].Capability.Class = ClassLarge
	sealCatalogForTest(t, &changed)
	if changed.Fingerprint == original {
		t.Fatal("capability evidence change did not change catalog fingerprint")
	}

	duplicate := catalog
	duplicate.Routes = append(append([]Route(nil), catalog.Routes...), catalog.Routes[0])
	sealCatalogForTest(t, &duplicate)
	if _, _, err := ValidateCatalogForSelection(duplicate, publicPolicy(), publicResolver(), now, CatalogSourceRun); !errors.Is(err, ErrDuplicateAlias) {
		t.Fatalf("duplicate alias error=%v", err)
	}

	trustedPolicy := trustCatalogRouteStates(publicPolicy(), catalog)
	validated, _, err := ValidateCatalogForSelection(catalog, trustedPolicy, publicResolver(), now, CatalogSourceRun)
	if err != nil {
		t.Fatal(err)
	}
	request := broadRequest(TierMedium)
	first, err := SelectRoute(validated, request, publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	if err != nil || len(first.Routes) != 1 {
		t.Fatalf("first decision=%#v err=%v", first, err)
	}
	first.Routes[0].Capability.Tools[0] = "mutated"
	ignored, err := SelectRoute(validated, request, publicPolicy(), RunOverride{Mode: OverrideIgnore}, AttemptLedger{}, now)
	if err != nil || ignored.Current.Route == nil {
		t.Fatalf("ignore decision=%#v err=%v", ignored, err)
	}
	ignored.Current.Route.Capability.Tools[0] = "mutated"
	second, err := SelectRoute(validated, request, publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	ignoredAgain, ignoreErr := SelectRoute(validated, request, publicPolicy(), RunOverride{Mode: OverrideIgnore}, AttemptLedger{}, now)
	if err != nil || ignoreErr != nil || second.Routes[0].Capability.Tools[0] != "apply_patch" || ignoredAgain.Current.Route.Capability.Tools[0] != "apply_patch" {
		t.Fatalf("validated state mutated through decision: %#v err=%v", second, err)
	}

	incomplete := request
	incomplete.TaskFamily, incomplete.Tools, incomplete.ContextSize, incomplete.Risk = "", nil, 0, ""
	decision, err := SelectRoute(validated, incomplete, publicPolicy(), RunOverride{}, AttemptLedger{}, now)
	if !errors.Is(err, ErrInvalidWorkRequest) || decision.Status != SelectionUnavailable {
		t.Fatalf("incomplete automatic request decision=%#v err=%v", decision, err)
	}
	decision, err = SelectRoute(validated, incomplete, publicPolicy(), RunOverride{Mode: OverrideRequire, Alias: "medium"}, AttemptLedger{}, now)
	if err != nil || len(decision.Routes) != 1 || decision.Routes[0].Alias != "medium" {
		t.Fatalf("attended incomplete request decision=%#v err=%v", decision, err)
	}
}

func TestRunCatalogCannotSelfDeclareTrustedCurrent(t *testing.T) {
	now := fixedNow()
	forged := provenRoute("forged-current", ClassPlanner, "local-lan", "codex", "exec-model", "forged-model", "code", now.Add(time.Hour))
	forged.Capability.Source = EvidenceAdapterPrior
	forged.Capability.DispatchProven = false
	forged.Readiness = []Readiness{ReadinessDiscovered, ReadinessConfigured, ReadinessSelectable}
	catalog := Catalog{SchemaVersion: CatalogSchemaVersion, Current: CurrentModel{ModelID: forged.DisplayModelID, Surface: "codex-app", Route: &forged}, Routes: []Route{forged}}
	sealCatalogForTest(t, &catalog)
	validated, rejections, err := ValidateCatalogForSelection(catalog, publicPolicy(), publicResolver(), now, CatalogSourceRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(rejections) != 2 || len(validated.catalog.Routes) != 0 || validated.catalog.Current.Route != nil {
		t.Fatalf("self-declared current survived: validated=%#v rejections=%#v", validated.catalog, rejections)
	}
}

func TestRunCatalogCannotRewriteTrustedCurrentMetadata(t *testing.T) {
	now := fixedNow()
	catalog := catalogWithCurrent(now, nil)
	catalog.Current.Route.Retention = RetentionNone
	catalog.Current.Route.TrainingUse = TrainingNo
	catalog.Current.Route.Residency = "local"
	catalog.Current.Route.Capability.ContextSize = 1 << 20
	sealCatalogForTest(t, &catalog)
	validated, rejections, err := ValidateCatalogForSelection(catalog, publicPolicy(), publicResolver(), now, CatalogSourceRun)
	if err != nil {
		t.Fatal(err)
	}
	if len(rejections) != 1 || validated.catalog.Current.Route != nil {
		t.Fatalf("rewritten current metadata survived: validated=%#v rejections=%#v", validated.catalog.Current, rejections)
	}
}

func exactReceipt(now time.Time) (RoutingReceipt, EvidenceEnvelope) {
	envelope := EvidenceEnvelope{
		RunID: "run-1", SliceID: "slice-1", ProjectID: "project-a", RouteAlias: "medium-a",
		RouteFingerprint: "route-sha256:abc", Adapter: "codex", AdapterRevision: "v1",
		DispatchMethod: "named-agent", ModelID: "gpt-medium", TaskFamily: "code",
		ContextPacketHash: "sha256:packet", ProofArtifactHash: "sha256:proof", Surface: "codex-cli", Provider: "openai",
		Tools: []string{"apply_patch", "go test"}, ContextSize: 4096, Risk: RiskBroad, MaxAge: time.Hour,
	}
	capabilityHash, _ := ComputeCapabilityEnvelopeHash(envelope)
	receipt := RoutingReceipt{RouteEvidence: RouteDispatchEvidence{
		RunID: envelope.RunID, SliceID: envelope.SliceID, ProjectID: envelope.ProjectID,
		RouteAlias: envelope.RouteAlias, RouteFingerprint: envelope.RouteFingerprint,
		Adapter: envelope.Adapter, AdapterRevision: envelope.AdapterRevision, DispatchMethod: envelope.DispatchMethod,
		RequestedModelID: envelope.ModelID, ProviderReportedModel: envelope.ModelID, SessionID: "session-1",
		TaskFamily: envelope.TaskFamily, ContextPacketHash: envelope.ContextPacketHash, CapabilityEnvelopeHash: capabilityHash, Attempt: 1, ObservedAt: now,
	}, WorkProof: WorkProof{Command: "go test ./internal/modelrouting", ArtifactHash: envelope.ProofArtifactHash, Result: ProofPass}}
	return receipt, envelope
}

func fixedNow() time.Time { return time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC) }

func publicResolver() Resolver {
	return StaticResolver(map[string][]net.IP{"api.openai.com": {net.ParseIP("104.18.0.1")}})
}

func publicPolicy() PolicyContext {
	policy := PolicyContext{
		Project:               ProjectPolicy{ProjectID: "project-a", AllowedDestinations: []string{"openai", "local-lan", "current"}, MaxRetention: RetentionSession},
		Trusted:               UserTrust{ProjectID: "project-a"},
		TrustedCurrentModelID: "current-gpt",
		TrustedCurrentSurface: "codex-app",
	}
	current := catalogWithCurrent(fixedNow(), nil).Current.Route
	policy.TrustedCurrentRouteState, _ = ComputeRouteStateFingerprint(*current)
	return policy
}

func broadRequest(tier Tier) WorkRequest {
	return WorkRequest{PlannedTier: tier, TaskFamily: "code", Tools: []string{"apply_patch", "go test"}, ContextSize: 4096, Risk: RiskBroad, ProjectID: "project-a"}
}

func catalogWithCurrent(now time.Time, routes []Route) Catalog {
	currentRoute := Route{
		Alias: "current", DisplayModelID: "current-gpt", Adapter: "codex", AdapterRevision: "v1", DispatchMethod: "exec-model", Destination: "current",
		ManagementOrigin: OriginNative, Hosting: HostingProviderHosted, DiscoverySources: []string{"active-host"},
		Boundary: BoundaryHosted, Retention: RetentionSession, TrainingUse: TrainingUnknown, Residency: "unknown", TrustProvenance: "active orchestrator",
		Readiness: []Readiness{ReadinessDiscovered, ReadinessConfigured, ReadinessSelectable},
		Capability: CapabilityEvidence{Class: ClassPlanner, Source: EvidenceDeclared, RouteAlias: "current", ModelID: "current-gpt", TaskFamily: "code",
			Tools: []string{"apply_patch", "go test"}, ContextSize: 8192, Risk: RiskBroad, DispatchProven: false},
	}
	return Catalog{SchemaVersion: CatalogSchemaVersion, Current: CurrentModel{ModelID: "current-gpt", Surface: "codex-app", Route: &currentRoute}, Routes: routes}
}

func provenRoute(alias string, class CapabilityClass, destination, adapter, dispatchMethod, modelID, family string, expires time.Time) Route {
	return Route{
		Alias: alias, DisplayModelID: modelID, Adapter: adapter, AdapterRevision: "v1", DispatchMethod: dispatchMethod, Destination: destination,
		ManagementOrigin: OriginNative, Hosting: HostingProviderHosted, DiscoverySources: []string{"test-adapter"},
		Boundary: BoundaryHosted, Retention: RetentionSession, TrainingUse: TrainingNo, Residency: "declared", TrustProvenance: "adapter-v1",
		Readiness: []Readiness{ReadinessDiscovered, ReadinessConfigured, ReadinessSelectable, ReadinessDispatchProven},
		Capability: CapabilityEvidence{Class: class, Source: EvidenceKBReceipt, RouteAlias: alias, ModelID: modelID, TaskFamily: family,
			Tools: []string{"apply_patch", "go test"}, ContextSize: 8192, Risk: RiskBroad, DispatchQualified: true, DispatchProven: true, ExpiresAt: expires},
	}
}

func declaredRoute(alias string, class CapabilityClass, destination, adapter, dispatchMethod, modelID, family string) Route {
	route := provenRoute(alias, class, destination, adapter, dispatchMethod, modelID, family, fixedNow().Add(time.Hour))
	route.Readiness = []Readiness{ReadinessDiscovered, ReadinessConfigured, ReadinessSelectable}
	route.Capability.Source, route.Capability.DispatchQualified, route.Capability.DispatchProven = EvidenceDeclared, false, false
	return route
}

func adapterPriorQualifiedRoute(route Route) Route {
	route.Capability.Source = EvidenceAdapterPrior
	route.Capability.DispatchQualified = true
	route.Capability.DispatchProven = false
	readiness := route.Readiness[:0]
	for _, value := range route.Readiness {
		if value != ReadinessDispatchProven {
			readiness = append(readiness, value)
		}
	}
	route.Readiness = readiness
	return route
}

func assertAliases(t *testing.T, decision SelectionDecision, want []string) {
	t.Helper()
	got := make([]string, 0, len(decision.Routes))
	for _, route := range decision.Routes {
		got = append(got, route.Alias)
	}
	if len(want) == 0 && len(got) == 0 {
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("route aliases=%v want=%v decision=%#v", got, want, decision)
	}
}

func selectForTest(t *testing.T, catalog Catalog, req WorkRequest, policy PolicyContext, override RunOverride, ledger AttemptLedger, now time.Time) (SelectionDecision, error) {
	t.Helper()
	sealCatalogForTest(t, &catalog)
	policy = trustCatalogRouteStates(policy, catalog)
	validated, _, err := ValidateCatalogForSelection(catalog, policy, publicResolver(), now, CatalogSourceRun)
	if err != nil {
		t.Fatalf("validate catalog: %v", err)
	}
	return SelectRoute(validated, req, policy, override, ledger, now)
}

func trustCatalogRouteStates(policy PolicyContext, catalog Catalog) PolicyContext {
	policy.TrustedRouteStates = make(map[string]string, len(catalog.Routes))
	for _, route := range catalog.Routes {
		if state, err := ComputeRouteStateFingerprint(route); err == nil {
			policy.TrustedRouteStates[route.Alias] = state
		}
	}
	return policy
}

func sealCatalogForTest(t *testing.T, catalog *Catalog) {
	t.Helper()
	fingerprint, err := ComputeCatalogFingerprint(*catalog)
	if err != nil {
		t.Fatalf("compute catalog fingerprint: %v", err)
	}
	catalog.Fingerprint = fingerprint
}

func TestNoCredentialFieldRemainsInRouteSchema(t *testing.T) {
	typeOfRoute := reflect.TypeOf(Route{})
	for index := 0; index < typeOfRoute.NumField(); index++ {
		field := typeOfRoute.Field(index)
		if field.Name == "AuthValue" || strings.Contains(string(field.Tag), "auth_value") || strings.Contains(string(field.Tag), "credential") {
			t.Fatalf("Route still exposes credential field %s", field.Name)
		}
	}
}

func TestRouteDenialOverridesOtherwiseEligibleHostedRoute(t *testing.T) {
	now := fixedNow()
	route := provenRoute("hosted-denied", ClassMedium, "openai", "codex", "named-agent", "gpt-medium", "code", now.Add(time.Hour))
	fingerprint, err := ComputeRouteFingerprint(route)
	if err != nil {
		t.Fatal(err)
	}
	policy := publicPolicy()
	policy.Trusted.RouteDenials = []RouteDenial{{ProjectID: "project-a", RouteFingerprint: fingerprint, CreatedAt: now}}
	decision, err := selectForTest(t, catalogWithCurrent(now, []Route{route}), broadRequest(TierMedium), policy, RunOverride{}, AttemptLedger{}, now)
	if err != nil {
		t.Fatalf("select denied route: %v", err)
	}
	if decision.Status != SelectionDegraded || len(decision.Routes) != 0 {
		t.Fatalf("denied route was selected: %#v", decision)
	}
}
