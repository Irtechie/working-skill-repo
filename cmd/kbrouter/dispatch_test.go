package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

var fakeCodexID atomic.Int64

func TestDispatchCodexExecArgvProfileModelAndProofUnknown(t *testing.T) {
	fixture := newDispatchFixture(t, "argv")
	profileName := "localprof"
	fixture.writeCodexProfile(profileName)
	route := fixture.route("profile.large", "large-model", modelrouting.ClassLarge)
	route.DispatchMethod = "exec-profile"
	route.Profile = profileName
	route.ProfileRevision = fixture.profileRevision(profileName, "", "")
	fixture.installCatalog(route)
	sessionID := "session-argv"
	fixture.trustRoutes(route)
	fixture.trustSession(sessionID, "large-model")
	fake := fixture.fakeCodex(fakeCodexSpec{sessionID: sessionID})
	fixture.withFakeCodex(fake)

	result := fixture.run("--route-alias", route.Alias, "--sandbox", "workspace-write", "--approval-policy", "never", "--network", "none", "--allowed-root", fixture.projectRoot, "--allowed-tool", "codex-harness")
	if result.code != 0 {
		t.Fatalf("dispatch exit=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	args := strings.Fields(readFileForTest(t, fake.argsPath))
	wantContainsInOrder(t, args, []string{"exec", "--model", "large-model", "--profile", profileName, "--sandbox", "workspace-write", "-C", fixture.projectRoot})
	wantContains(t, args, "--add-dir", fixture.projectRoot)
	approvalArg := `approval_policy="never"`
	if runtime.GOOS == "windows" {
		approvalArg = `approval_policy=\"never\"`
	}
	wantContains(t, args, "-c", approvalArg)
	wantContains(t, args, "-c", `sandbox_workspace_write.network_access=false`)
	wantContains(t, args, "--output-schema")
	wantContains(t, args, "--json", "-")
	report := decodeDispatchReport(t, result.stdout)
	if report.Status != "observation-only" || report.Attribution != "exact" {
		t.Fatalf("dispatch must not self-credit even with exact host evidence: %#v", report)
	}
	receipt := decodeReceipt(t, fixture.receiptPath)
	if receipt.WorkProof.Result != modelrouting.ProofUnknown || receipt.RouteEvidence.ProviderReportedModel != "large-model" || receipt.RouteEvidence.SessionID != sessionID {
		t.Fatalf("receipt should bind host evidence but keep proof unknown: %#v", receipt)
	}
}

func TestDispatchPreservesAttemptAndPlannedCorrectionTiers(t *testing.T) {
	fixture := newDispatchFixture(t, "attempt-tier")
	fixture.writePacket(dispatchPacketForTest{
		SchemaVersion: 1, PacketID: "packet-attempt-tier", TaskID: "task-attempt-tier",
		RunID: filepath.Base(fixture.runRoot), SliceID: "slice-004", ModelTier: "medium", AttemptTier: "small",
		TaskFamily: "code", ContextSize: 8192, Risk: "broad", AllowedTools: []string{"codex-harness"},
		ProofTargets: []string{"go test ./cmd/kbrouter"}, Redaction: map[string]any{"bounded": true}, BoundedContext: true,
	})
	route := fixture.route("codex.small", "small-model", modelrouting.ClassSmall)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fixture.trustSession("session-small", "small-model")
	fixture.withFakeCodex(fixture.fakeCodex(fakeCodexSpec{sessionID: "session-small"}))

	result := fixture.run("--route-alias", route.Alias)
	if result.code != 0 {
		t.Fatalf("dispatch exit=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	report := decodeDispatchReport(t, result.stdout)
	if report.PlannedTier != modelrouting.TierMedium || report.AttemptTier != modelrouting.TierSmall {
		t.Fatalf("dispatch lost planned/attempt distinction: %#v", report)
	}

	for name, tiers := range map[string][2]string{
		"equal":       {"medium", "medium"},
		"skips tier":  {"large", "small"},
		"small tries": {"small", "small"},
		"above":       {"small", "medium"},
	} {
		t.Run(name, func(t *testing.T) {
			invalid := dispatchPacketForTest{
				SchemaVersion: 1, PacketID: "packet-invalid-attempt", TaskID: "task-invalid-attempt",
				RunID: filepath.Base(fixture.runRoot), SliceID: "slice-004", ModelTier: tiers[0], AttemptTier: tiers[1],
				TaskFamily: "code", ContextSize: 8192, Risk: "broad", AllowedTools: []string{"codex-harness"},
				ProofTargets: []string{"go test ./cmd/kbrouter"}, Redaction: map[string]any{"bounded": true}, BoundedContext: true,
			}
			data, err := json.Marshal(invalid)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decodeDispatchPacket(data, invalid.RunID, invalid.SliceID); err == nil || !strings.Contains(err.Error(), "attempt tier") {
				t.Fatalf("invalid attempt tier accepted: %v", err)
			}
		})
	}
}

func TestDispatchCorrectionPacketRefusesBeforeWorkerLaunchOrReceipt(t *testing.T) {
	fixture := newDispatchFixture(t, "correction-link")
	projectID, err := modelrouting.CanonicalProjectIdentity(fixture.projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	attempt := modelrouting.RoutingReceipt{
		RouteEvidence: modelrouting.RouteDispatchEvidence{
			RunID: filepath.Base(fixture.runRoot), SliceID: "slice-004", ProjectID: projectID, RouteAlias: "codex.small",
			RouteFingerprint: modelrouting.SHA256Bytes([]byte("small-route")), Adapter: "codex", AdapterRevision: "v1",
			DispatchMethod: "exec-model", RequestedModelID: "small-model", ProviderReportedModel: "small-model", SessionID: "attempt-session",
			TaskFamily: "code", ContextPacketID: "attempt-packet", ContextPacketHash: modelrouting.SHA256Bytes([]byte("attempt-packet")),
			CapabilityEnvelopeHash: "capability-sha256:" + strings.Repeat("a", 64), Attempt: 1,
		},
		WorkProof: modelrouting.WorkProof{Command: "go test ./cmd/kbrouter", ArtifactHash: modelrouting.SHA256Bytes([]byte("failed-proof")), Result: modelrouting.ProofFail},
	}
	attemptHash, err := modelrouting.HashRoutingReceipt(attempt)
	if err != nil {
		t.Fatal(err)
	}
	if err := modelrouting.SaveAtomicJSON(fixture.runRoot, "attempt-receipt.json", attempt, maxCatalogBytes); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"current.diff": "diff", "worker-result.redacted": "worker-result", "worker.log": "worker-log"} {
		if err := os.WriteFile(filepath.Join(fixture.runRoot, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fix := modelrouting.HunkBoundary{ID: "fix", File: "cmd/kbrouter/dispatch.go", StartLine: 1, EndLine: 2}
	keep := modelrouting.HunkBoundary{ID: "keep", File: "cmd/kbrouter/dispatch_test.go", StartLine: 1, EndLine: 2}
	correction := modelrouting.CorrectionPacket{
		SchemaVersion: modelrouting.CorrectionSchemaVersion, PacketID: "correction-packet", RunID: filepath.Base(fixture.runRoot), SliceID: "slice-004",
		AttemptReceipt: modelrouting.AttemptReceiptReference{Path: "attempt-receipt.json", Hash: attemptHash, RouteAlias: "codex.small", Attempt: 1},
		PlannedTier:    modelrouting.TierMedium, CorrectionTier: modelrouting.TierMedium,
		Authority: modelrouting.CorrectionAuthority{
			Owner: modelrouting.AuthorityDriver, OriginalScopeHash: modelrouting.SHA256Bytes([]byte("scope")), PreCorrectionBaselineHash: modelrouting.SHA256Bytes([]byte("baseline")), AllowedChanges: []modelrouting.HunkBoundary{fix},
			Invariants: []string{"routing policy unchanged"}, RelevantInterfaces: []string{"cmd/kbrouter/dispatch.go"}, ExactProof: []string{"go test ./cmd/kbrouter"},
		},
		Failure:       modelrouting.FailureEvidence{CriterionID: "proof-1", Localizable: true, Location: &fix},
		AttemptLedger: []modelrouting.CorrectionAttempt{{Attempt: 1, RouteAlias: "codex.small", ReceiptHash: attemptHash, OutcomeCode: modelrouting.AttemptOutcomeProofFailed}},
		CurrentDiff:   modelrouting.BoundedArtifact{Path: "current.diff", Hash: modelrouting.SHA256Bytes([]byte("diff")), Bytes: int64(len("diff")), Redacted: true},
		WorkerResult:  modelrouting.BoundedArtifact{Path: "worker-result.redacted", Hash: modelrouting.SHA256Bytes([]byte("worker-result")), Bytes: int64(len("worker-result")), Redacted: true},
		WorkerLog:     modelrouting.BoundedArtifact{Path: "worker.log", Hash: modelrouting.SHA256Bytes([]byte("worker-log")), Bytes: int64(len("worker-log")), Redacted: true},
		AcceptedHunks: []modelrouting.AcceptedHunk{{Boundary: keep, ContentHash: modelrouting.SHA256Bytes([]byte("keep")), Oracle: modelrouting.HunkOracle{
			Owner: modelrouting.AuthorityIndependentOracle, Scope: modelrouting.OracleScopeHunkLocal, Command: "go test ./cmd/kbrouter -run Keep",
			ArtifactHash: modelrouting.SHA256Bytes([]byte("keep-proof")), Result: modelrouting.ProofPass,
		}}},
	}
	fixture.writePacket(dispatchPacketForTest{
		SchemaVersion: 1, PacketID: "packet-correction-link", TaskID: "task-correction-link", RunID: filepath.Base(fixture.runRoot), SliceID: "slice-004",
		ModelTier: "medium", TaskFamily: "code", ContextSize: 8192, Risk: "broad", AllowedTools: []string{"codex-harness"},
		ProofTargets: []string{"go test ./cmd/kbrouter"}, Redaction: map[string]any{"bounded": true}, BoundedContext: true, Correction: &correction,
	})
	route := fixture.route("codex.medium", "medium-model", modelrouting.ClassMedium)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fixture.trustSession("correction-session", "medium-model")
	fake := fixture.fakeCodex(fakeCodexSpec{sessionID: "correction-session"})
	fixture.withFakeCodex(fake)

	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "isolated workspace") {
		t.Fatalf("correction dispatch was not fail-closed: exit=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	if _, err := os.Stat(fake.argsPath); !os.IsNotExist(err) {
		t.Fatalf("correction worker launched before isolation: %v", err)
	}
	if _, err := os.Stat(fixture.receiptPath); !os.IsNotExist(err) {
		t.Fatalf("correction receipt minted before validation: %v", err)
	}
}

func TestDispatchRequiresMarkedRunCatalogAndRejectsArbitraryExec(t *testing.T) {
	fixture := newDispatchFixture(t, "required-catalog")
	fake := fixture.fakeCodex(fakeCodexSpec{})
	fixture.withFakeCodex(fake)
	result := fixture.run("--route-alias", "missing")
	if result.code == 0 || !strings.Contains(result.stderr, "run catalog is required") {
		t.Fatalf("missing catalog accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	result = fixture.run("--route-alias", "missing", "--exec", fake.path)
	if result.code == 0 || !strings.Contains(result.stderr, "flag provided but not defined") {
		t.Fatalf("public arbitrary executable was not rejected, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}

func TestDispatchRejectsProjectPathCodexExecutable(t *testing.T) {
	fixture := newDispatchFixture(t, "project-exec")
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fake := fixture.fakeCodexAt(filepath.Join(fixture.projectRoot, "codex"+scriptExt()), fakeCodexSpec{})
	fixture.withFakeCodex(fake)
	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "trusted codex executable") {
		t.Fatalf("project-local codex executable accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}

func TestDispatchRequiresPrivateRouteState(t *testing.T) {
	fixture := newDispatchFixture(t, "missing-private-state")
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fake := fixture.fakeCodex(fakeCodexSpec{})
	fixture.withFakeCodexResolverOnly(fake)
	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "load trusted dispatch state") {
		t.Fatalf("missing private dispatch state accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}

func TestModelsAddProfileRejectsProjectScope(t *testing.T) {
	fixture := newDispatchFixture(t, "profile-project")
	fixture.writeCodexProfile("localprof")
	code, stdout, stderr := runForTest("models", "add",
		"--user-root", fixture.userRoot, "--project-root", fixture.projectRoot,
		"--scope", "project", "--alias", "profile.large", "--profile", "localprof",
	)
	if code == 0 || !strings.Contains(stderr, "project scope cannot store connection details or profiles") {
		t.Fatalf("project profile accepted, code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
}

func TestDispatchRejectsForgedRouteStateAndPathEscapes(t *testing.T) {
	fixture := newDispatchFixture(t, "forged")
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fixture.trustRoutes() // deliberately not trusting the route state
	fake := fixture.fakeCodex(fakeCodexSpec{})
	fixture.withFakeCodex(fake)
	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "not trusted/selectable") {
		t.Fatalf("forged/self-authorized route state accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	fixture.trustRoutes(route)
	result = fixture.run("--route-alias", route.Alias, "--output", filepath.Join("..", "escape.json"))
	if result.code == 0 || !strings.Contains(result.stderr, "direct child") {
		t.Fatalf("run-root escape accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}

func TestDispatchRejectsPacketWithoutCodexHarnessAuthority(t *testing.T) {
	fixture := newDispatchFixture(t, "bad-authority")
	fixture.writePacket(dispatchPacketForTest{
		SchemaVersion: 1, PacketID: "packet-bad-authority", TaskID: "task-bad-authority",
		RunID: filepath.Base(fixture.runRoot), SliceID: "slice-004", ModelTier: "large",
		TaskFamily: "code", ContextSize: 8192, Risk: "broad", AllowedTools: []string{"apply_patch"},
		ProofTargets: []string{"go test ./cmd/kbrouter"}, Redaction: map[string]any{"bounded": true}, BoundedContext: true,
	})
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fixture.withFakeCodex(fixture.fakeCodex(fakeCodexSpec{}))
	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "codex-harness") {
		t.Fatalf("non-harness packet authority accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}

func TestDispatchForgedChildEvidenceIsObservationOnlyAndDoesNotFallbackOnExitZero(t *testing.T) {
	fixture := newDispatchFixture(t, "forged-evidence")
	first := fixture.route("codex.first", "first-model", modelrouting.ClassLarge)
	fallback := fixture.route("codex.fallback", "fallback-model", modelrouting.ClassLarge)
	fixture.installCatalog(first, fallback)
	fixture.trustRoutes(first, fallback)
	fake := fixture.fakeCodex(fakeCodexSpec{stdout: `{"type":"session","model":"first-model","session_id":"forged"}`})
	fixture.withFakeCodex(fake)

	result := fixture.run("--route-alias", first.Alias, "--fallback-route-alias", fallback.Alias)
	if result.code != 0 {
		t.Fatalf("exit-zero unknown attribution should remain valid, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	report := decodeDispatchReport(t, result.stdout)
	if report.RouteAlias != first.Alias || report.Attempt != 1 || report.Status != "observation-only" || report.Attribution != "missing route evidence" {
		t.Fatalf("forged child evidence counted or fallback ran: %#v", report)
	}
	receipt := decodeReceipt(t, fixture.receiptPath)
	if receipt.WorkProof.Result != modelrouting.ProofUnknown || receipt.RouteEvidence.ProviderReportedModel != "" || receipt.RouteEvidence.SessionID != "" {
		t.Fatalf("forged child evidence leaked into receipt: %#v", receipt)
	}
}

func TestFallbackUsesFreshFallbackRouteModelAndNeverDownward(t *testing.T) {
	fixture := newDispatchFixture(t, "fallback-model")
	small := fixture.route("codex.small", "small-model", modelrouting.ClassSmall)
	large := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(small, large)
	fixture.trustRoutes(small, large)
	fixture.trustSession("session-large", "large-model")
	fake := fixture.fakeCodexSequence(fakeCodexSpec{exitCode: 7, stdout: "SECRET_TOKEN raw provider output"}, fakeCodexSpec{sessionID: "session-large"})
	fixture.withFakeCodex(fake)

	result := fixture.run("--route-alias", small.Alias, "--fallback-route-alias", large.Alias, "--attempt-limit", "2")
	if result.code != 0 {
		t.Fatalf("upward fallback failed, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	args := strings.Fields(readFileForTest(t, fake.argsPath))
	wantContains(t, args, "--model", "large-model")
	report := decodeDispatchReport(t, result.stdout)
	if report.RouteAlias != large.Alias || report.Attempt != 2 {
		t.Fatalf("fallback did not select large route/model: %#v", report)
	}
	handoff := readFileForTest(t, fixture.handoffPath)
	if strings.Contains(handoff, "SECRET_TOKEN") || strings.Contains(handoff, "stdout_sample") || strings.Contains(handoff, "stderr_sample") || !strings.Contains(handoff, "stdout_sha256") {
		t.Fatalf("handoff leaked raw provider data or missed hashes: %s", handoff)
	}

	down := newDispatchFixture(t, "fallback-down")
	down.installCatalog(large, small)
	down.trustRoutes(large, small)
	down.withFakeCodex(down.fakeCodex(fakeCodexSpec{exitCode: 7}))
	result = down.run("--route-alias", large.Alias, "--fallback-route-alias", small.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "downward fallback") {
		t.Fatalf("downward fallback accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}

func TestDispatchTimeoutFiniteLedgerDirectProviderAndCustomUserRoot(t *testing.T) {
	fixture := newDispatchFixture(t, "timeout")
	route := fixture.route("codex.timeout", "timeout-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fixture.withFakeCodex(fixture.fakeCodex(fakeCodexSpec{sleepMillis: 200}))
	result := fixture.run("--route-alias", route.Alias, "--timeout", "25ms")
	if result.code == 0 || !strings.Contains(result.stderr, "timeout") {
		t.Fatalf("expected timeout failure, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	if strings.Contains(readFileForTest(t, fixture.handoffPath), "SECRET") {
		t.Fatalf("timeout handoff leaked secret-like output: %s", readFileForTest(t, fixture.handoffPath))
	}

	repeat := newDispatchFixture(t, "finite")
	repeat.installCatalog(route)
	repeat.trustRoutes(route)
	repeat.withFakeCodex(repeat.fakeCodex(fakeCodexSpec{exitCode: 9}))
	result = repeat.run("--route-alias", route.Alias, "--fallback-route-alias", route.Alias, "--attempt-limit", "2")
	if result.code == 0 || !strings.Contains(result.stderr, "already attempted") {
		t.Fatalf("repeat attempt accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}

	direct := newDispatchFixture(t, "direct-provider")
	provider := route
	provider.Alias, provider.Adapter, provider.DispatchMethod, provider.Destination = "direct.provider", "openai-compatible", "chat-completions", "hosted"
	provider.Capability.RouteAlias = provider.Alias
	direct.installCatalog(provider)
	direct.trustRoutes(provider)
	direct.withFakeCodex(direct.fakeCodex(fakeCodexSpec{}))
	result = direct.run("--route-alias", provider.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "direct provider dispatch is not supported") {
		t.Fatalf("direct provider dispatch accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}

	previous := allowCustomUserRootForTests
	allowCustomUserRootForTests = false
	defer func() { allowCustomUserRootForTests = previous }()
	var stdout, stderr strings.Builder
	code := run([]string{"dispatch", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--slice-id", "slice-004", "--packet", "packet.json", "--route-alias", route.Alias}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "fixed user-local trust root") {
		t.Fatalf("production custom user root accepted, code=%d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
}

func TestDispatchEnvKeepsOnlyOSCodexHomeAndRouteAuth(t *testing.T) {
	fixture := newDispatchFixture(t, "env")
	route := fixture.route("codex.env", "env-model", modelrouting.ClassLarge)
	route.AuthEnv = "ROUTE_TOKEN"
	route.Endpoint = "https://1.1.1.1/v1"
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fixture.trustAuth(route)
	t.Setenv("OPENAI_API_KEY", "must-not-forward")
	t.Setenv("LITELLM_API_KEY", "must-not-forward")
	t.Setenv("CODEX_EXTRA_SECRET", "must-not-forward")
	t.Setenv("ROUTE_TOKEN", "route-secret")
	fake := fixture.fakeCodex(fakeCodexSpec{envPath: filepath.Join(fixture.runRoot, "env.txt")})
	fixture.withFakeCodex(fake)
	result := fixture.run("--route-alias", route.Alias)
	if result.code != 0 {
		t.Fatalf("dispatch env exit=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	env := readFileForTest(t, fake.envPath)
	for _, forbidden := range []string{"OPENAI_API_KEY=must-not-forward", "LITELLM_API_KEY=must-not-forward", "CODEX_EXTRA_SECRET=must-not-forward"} {
		if strings.Contains(env, forbidden) {
			t.Fatalf("wildcard provider env forwarded: %s", env)
		}
	}
	if !strings.Contains(env, "CODEX_HOME="+fixture.codexHome) || !strings.Contains(env, "ROUTE_TOKEN=route-secret") {
		t.Fatalf("required env missing: %s", env)
	}
}

type dispatchFixture struct {
	t                 *testing.T
	root              string
	projectRoot       string
	userRoot          string
	codexHome         string
	runRoot           string
	packetPath        string
	outputPath        string
	receiptPath       string
	handoffPath       string
	workerRequestPath string
	hostState         dispatchTrustedState
}

type dispatchCommandResult struct {
	code           int
	stdout, stderr string
}

type fakeCodexSpec struct {
	sessionID   string
	exitCode    int
	stdout      string
	sleepMillis int
	envPath     string
}

type fakeCodex struct {
	path     string
	argsPath string
	envPath  string
}

type dispatchPacketForTest struct {
	SchemaVersion  int                            `json:"schema_version"`
	PacketID       string                         `json:"packet_id"`
	TaskID         string                         `json:"task_id"`
	RunID          string                         `json:"run_id"`
	SliceID        string                         `json:"slice_id"`
	ModelTier      string                         `json:"model_tier"`
	AttemptTier    string                         `json:"attempt_tier,omitempty"`
	TaskFamily     string                         `json:"task_family"`
	ContextSize    int                            `json:"context_size"`
	Risk           string                         `json:"risk"`
	AllowedTools   []string                       `json:"allowed_tools"`
	ProofTargets   []string                       `json:"proof_targets"`
	Redaction      map[string]any                 `json:"redaction"`
	BoundedContext bool                           `json:"bounded_context"`
	Correction     *modelrouting.CorrectionPacket `json:"correction,omitempty"`
}

func newDispatchFixture(t *testing.T, name string) dispatchFixture {
	t.Helper()
	root := t.TempDir()
	projectRoot := filepath.Join(root, "project")
	userRoot := filepath.Join(root, "user")
	codexHome := filepath.Join(root, "codex-home")
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-"+name)
	for _, dir := range []string{projectRoot, userRoot, codexHome} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := prepareRunRoot(projectRoot, runRoot); err != nil {
		t.Fatal(err)
	}
	packetPath := filepath.Join(runRoot, "packet.json")
	fixture := dispatchFixture{
		t: t, root: root, projectRoot: projectRoot, userRoot: userRoot, codexHome: codexHome, runRoot: runRoot,
		packetPath: packetPath, outputPath: filepath.Join(runRoot, "output.json"),
		receiptPath: filepath.Join(runRoot, "receipt.json"), handoffPath: filepath.Join(runRoot, "handoff.json"),
		workerRequestPath: filepath.Join(runRoot, "worker-request.json"),
		hostState:         dispatchTrustedState{SchemaVersion: 1, RouteStates: map[string]string{}},
	}
	fixture.writePacket(dispatchPacketForTest{
		SchemaVersion: 1, PacketID: "packet-" + name, TaskID: "task-" + name,
		RunID: filepath.Base(runRoot), SliceID: "slice-004", ModelTier: "large",
		TaskFamily: "code", ContextSize: 8192, Risk: "broad", AllowedTools: []string{"codex-harness"},
		ProofTargets: []string{"go test ./cmd/kbrouter"}, Redaction: map[string]any{"bounded": true}, BoundedContext: true,
	})
	return fixture
}

func (f dispatchFixture) writePacket(packet dispatchPacketForTest) {
	f.t.Helper()
	data, err := json.Marshal(packet)
	if err != nil {
		f.t.Fatal(err)
	}
	if err := os.WriteFile(f.packetPath, data, 0o600); err != nil {
		f.t.Fatal(err)
	}
}

func (f dispatchFixture) run(extra ...string) dispatchCommandResult {
	f.t.Helper()
	args := []string{"dispatch",
		"--user-root", f.userRoot,
		"--project-root", f.projectRoot,
		"--run-root", f.runRoot,
		"--run-id", filepath.Base(f.runRoot),
		"--slice-id", "slice-004",
		"--packet", "packet.json",
		"--output", "output.json",
		"--receipt", "receipt.json",
		"--handoff", "handoff.json",
		"--worker-request", "worker-request.json",
		"--json",
	}
	args = append(args, extra...)
	code, stdout, stderr := runForTest(args...)
	return dispatchCommandResult{code: code, stdout: stdout, stderr: stderr}
}

func (f dispatchFixture) withFakeCodex(fakes ...fakeCodex) {
	f.t.Helper()
	previousResolver := dispatchExecutableResolver
	previousState := dispatchTrustedStateProvider
	previousHome := dispatchCodexHome
	index := 0
	dispatchExecutableResolver = func() (string, error) {
		if index >= len(fakes) {
			return fakes[len(fakes)-1].path, nil
		}
		path := fakes[index].path
		index++
		return path, nil
	}
	state := f.hostState
	dispatchTrustedStateProvider = func(string, string) (dispatchTrustedState, error) {
		return state, nil
	}
	dispatchCodexHome = func() (string, error) { return f.codexHome, nil }
	f.t.Cleanup(func() {
		dispatchExecutableResolver = previousResolver
		dispatchTrustedStateProvider = previousState
		dispatchCodexHome = previousHome
	})
}

func (f dispatchFixture) withFakeCodexResolverOnly(fake fakeCodex) {
	f.t.Helper()
	previousResolver := dispatchExecutableResolver
	previousHome := dispatchCodexHome
	dispatchExecutableResolver = func() (string, error) {
		return fake.path, nil
	}
	dispatchCodexHome = func() (string, error) { return f.codexHome, nil }
	f.t.Cleanup(func() {
		dispatchExecutableResolver = previousResolver
		dispatchCodexHome = previousHome
	})
}

func (f dispatchFixture) fakeCodex(spec fakeCodexSpec) fakeCodex {
	f.t.Helper()
	id := strconv.FormatInt(fakeCodexID.Add(1), 10)
	path := filepath.Join(f.root, "fake-codex-"+id+scriptExt())
	return f.fakeCodexAt(path, spec)
}

func (f dispatchFixture) fakeCodexAt(path string, spec fakeCodexSpec) fakeCodex {
	f.t.Helper()
	id := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	argsPath := filepath.Join(f.runRoot, "argv-"+id+".txt")
	envPath := spec.envPath
	if envPath == "" {
		envPath = filepath.Join(f.runRoot, "env-"+id+".txt")
	}
	lines := []string{}
	if runtime.GOOS == "windows" {
		lines = append(lines, "@echo off", "echo %* > "+quoteCmd(argsPath), "set > "+quoteCmd(envPath))
		if spec.sleepMillis > 0 {
			lines = append(lines, "powershell -NoProfile -Command Start-Sleep -Milliseconds "+strconv.Itoa(spec.sleepMillis))
		}
	} else {
		lines = append(lines, "#!/bin/sh", "printf '%s\\n' \"$*\" > "+quoteShell(argsPath), "env > "+quoteShell(envPath))
		if spec.sleepMillis > 0 {
			lines = append(lines, "sleep "+strconv.FormatFloat(float64(spec.sleepMillis)/1000, 'f', 3, 64))
		}
	}
	if spec.sessionID != "" {
		lines = append(lines, echoLine(`{"type":"thread.started","thread_id":"`+spec.sessionID+`"}`, false))
	}
	if spec.stdout != "" {
		lines = append(lines, echoLine(spec.stdout, false))
	}
	lines = append(lines, exitLine(spec.exitCode))
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o755); err != nil {
		f.t.Fatal(err)
	}
	return fakeCodex{path: path, argsPath: argsPath, envPath: envPath}
}

func (f dispatchFixture) fakeCodexSequence(first, second fakeCodexSpec) fakeCodex {
	f.t.Helper()
	id := strconv.FormatInt(fakeCodexID.Add(1), 10)
	path := filepath.Join(f.root, "fake-codex-seq-"+id+scriptExt())
	argsPath := filepath.Join(f.runRoot, "argv-seq-"+id+".txt")
	envPath := filepath.Join(f.runRoot, "env-seq-"+id+".txt")
	countPath := filepath.Join(f.runRoot, "count-seq-"+id+".txt")
	lines := []string{}
	if runtime.GOOS == "windows" {
		lines = append(lines, "@echo off", "set N=0", "if exist "+quoteCmd(countPath)+" set /p N=<"+quoteCmd(countPath), "set /a N=N+1", "echo %N% > "+quoteCmd(countPath), "echo %* > "+quoteCmd(argsPath), "set > "+quoteCmd(envPath), "if \"%N%\"==\"1\" (")
		lines = append(lines, fakeCodexBodyLines(first)...)
		lines = append(lines, ")", "if \"%N%\"==\"2\" (")
		lines = append(lines, fakeCodexBodyLines(second)...)
		lines = append(lines, ")", "exit /b 99")
	} else {
		lines = append(lines, "#!/bin/sh", "N=0", "[ -f "+quoteShell(countPath)+" ] && N=$(cat "+quoteShell(countPath)+")", "N=$((N+1))", "printf '%s\\n' \"$N\" > "+quoteShell(countPath), "printf '%s\\n' \"$*\" > "+quoteShell(argsPath), "env > "+quoteShell(envPath), "if [ \"$N\" = 1 ]; then")
		lines = append(lines, fakeCodexBodyLines(first)...)
		lines = append(lines, "fi", "if [ \"$N\" = 2 ]; then")
		lines = append(lines, fakeCodexBodyLines(second)...)
		lines = append(lines, "fi", "exit 99")
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o755); err != nil {
		f.t.Fatal(err)
	}
	return fakeCodex{path: path, argsPath: argsPath, envPath: envPath}
}

func fakeCodexBodyLines(spec fakeCodexSpec) []string {
	lines := []string{}
	if spec.sessionID != "" {
		lines = append(lines, echoLine(`{"type":"thread.started","thread_id":"`+spec.sessionID+`"}`, false))
	}
	if spec.stdout != "" {
		lines = append(lines, echoLine(spec.stdout, false))
	}
	lines = append(lines, exitLine(spec.exitCode))
	return lines
}

func (f dispatchFixture) route(alias, model string, class modelrouting.CapabilityClass) modelrouting.Route {
	return modelrouting.Route{
		Alias: alias, DisplayModelID: model, Adapter: "codex", AdapterRevision: "v1", DispatchMethod: "exec-model", Destination: "codex",
		Boundary: modelrouting.BoundaryHosted, Retention: modelrouting.RetentionSession, TrainingUse: modelrouting.TrainingUnknown,
		Residency: "unknown", TrustProvenance: "test host fixture",
		Readiness:  []modelrouting.Readiness{modelrouting.ReadinessDiscovered, modelrouting.ReadinessConfigured, modelrouting.ReadinessSelectable},
		Capability: modelrouting.CapabilityEvidence{Class: class, Source: modelrouting.EvidenceDeclared, RouteAlias: alias, ModelID: model, TaskFamily: "code", Tools: []string{"codex-harness"}, ContextSize: 8192, Risk: modelrouting.RiskBroad},
	}
}

func (f dispatchFixture) installCatalog(routes ...modelrouting.Route) {
	f.t.Helper()
	catalog := modelrouting.Catalog{SchemaVersion: modelrouting.CatalogSchemaVersion, Routes: routes}
	fingerprintAndSort(&catalog)
	if err := modelrouting.SaveCatalog(f.runRoot, "catalog.json", catalog, modelrouting.StorageOptions{MaxBytes: maxCatalogBytes, Source: modelrouting.CatalogSourceRun}); err != nil {
		f.t.Fatalf("save run catalog: %v", err)
	}
}

func (f *dispatchFixture) trustRoutes(routes ...modelrouting.Route) {
	f.t.Helper()
	f.hostState.RouteStates = map[string]string{}
	for _, route := range routes {
		state, err := modelrouting.ComputeRouteStateFingerprint(route)
		if err != nil {
			f.t.Fatal(err)
		}
		f.hostState.RouteStates[route.Alias] = state
	}
}

func (f *dispatchFixture) trustSession(sessionID, model string) {
	f.t.Helper()
	dir := filepath.Join(f.codexHome, "sessions")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		f.t.Fatal(err)
	}
	lines := []string{
		`{"type":"session_meta","payload":{"id":"` + sessionID + `","model_provider":"codex","cwd":"` + filepath.ToSlash(f.projectRoot) + `"}}`,
		`{"type":"turn_context","payload":{"model":"` + model + `","cwd":"` + filepath.ToSlash(f.projectRoot) + `","approval_policy":"never","sandbox_policy":{"type":"workspace-write"}}}`,
	}
	if err := os.WriteFile(filepath.Join(dir, sessionID+".jsonl"), []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		f.t.Fatal(err)
	}
}

func (f dispatchFixture) trustAuth(route modelrouting.Route) {
	f.t.Helper()
	projectID, err := modelrouting.CanonicalProjectIdentity(f.projectRoot)
	if err != nil {
		f.t.Fatal(err)
	}
	trust := userTrustFileData{SchemaVersion: 1, Projects: []userProjectTrust{{
		ProjectID: projectID,
		AuthBindings: []modelrouting.AuthBinding{{
			Env: route.AuthEnv, Adapter: route.Adapter, Origin: originForEndpoint(route.Endpoint), ExpiresAt: time.Now().Add(time.Hour),
		}},
	}}}
	if err := saveTrustFile(f.userRoot, trust); err != nil {
		f.t.Fatal(err)
	}
}

func (f dispatchFixture) writeCodexProfile(name string) {
	f.t.Helper()
	if err := os.WriteFile(filepath.Join(f.codexHome, name+".config.toml"), []byte("model_provider = \"test\"\n"), 0o600); err != nil {
		f.t.Fatal(err)
	}
}

func (f dispatchFixture) profileRevision(name, endpoint, authEnv string) string {
	f.t.Helper()
	revision, err := trustedCodexProfileRevision(f.codexHome, name, endpoint, authEnv)
	if err != nil {
		f.t.Fatal(err)
	}
	return revision
}

func quoteShell(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "'\"'\"'") + "'"
}

func quoteCmd(path string) string {
	return `"` + path + `"`
}

func scriptExt() string {
	if runtime.GOOS == "windows" {
		return ".cmd"
	}
	return ".sh"
}

func echoLine(line string, stderr bool) string {
	if runtime.GOOS == "windows" {
		if stderr {
			return "echo " + line + " 1>&2"
		}
		return "echo " + line
	}
	quoted := "'" + strings.ReplaceAll(line, "'", "'\"'\"'") + "'"
	if stderr {
		return "printf '%s\\n' " + quoted + " >&2"
	}
	return "printf '%s\\n' " + quoted
}

func exitLine(code int) string {
	if runtime.GOOS == "windows" {
		return "exit /b " + strconv.Itoa(code)
	}
	return "exit " + strconv.Itoa(code)
}

func decodeDispatchReport(t *testing.T, data string) dispatchReport {
	t.Helper()
	var report dispatchReport
	if err := json.Unmarshal([]byte(data), &report); err != nil {
		t.Fatalf("decode dispatch report %q: %v", data, err)
	}
	return report
}

func decodeReceipt(t *testing.T, path string) modelrouting.RoutingReceipt {
	t.Helper()
	var receipt modelrouting.RoutingReceipt
	if err := json.Unmarshal([]byte(readFileForTest(t, path)), &receipt); err != nil {
		t.Fatalf("decode receipt: %v", err)
	}
	return receipt
}

func wantContainsInOrder(t *testing.T, args, want []string) {
	t.Helper()
	position := 0
	for _, expected := range want {
		found := false
		for position < len(args) {
			if args[position] == expected {
				found = true
				position++
				break
			}
			position++
		}
		if !found {
			t.Fatalf("argv missing ordered token %q in %#v", expected, args)
		}
	}
}

func wantContains(t *testing.T, args []string, want ...string) {
	t.Helper()
	for start := 0; start+len(want) <= len(args); start++ {
		match := true
		for offset, token := range want {
			if args[start+offset] != token {
				match = false
				break
			}
		}
		if match {
			return
		}
	}
	t.Fatalf("argv missing sequence %#v in %#v", want, args)
}
