package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestC1RejectsDuplicateAndReservedDispatchArtifacts(t *testing.T) {
	fixture := newDispatchFixture(t, "c1-artifacts")
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fixture.withFakeCodex(fixture.fakeCodex(fakeCodexSpec{}))

	result := fixture.run("--route-alias", route.Alias, "--output", "same.json", "--receipt", "same.json")
	if result.code == 0 || !strings.Contains(result.stderr, "artifact paths must be pairwise distinct") {
		t.Fatalf("duplicate artifacts accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}

	result = fixture.run("--route-alias", route.Alias, "--output", "catalog.json")
	if result.code == 0 || !strings.Contains(result.stderr, "reserved artifact name") {
		t.Fatalf("reserved artifact accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}

	result = fixture.run("--route-alias", route.Alias, "--packet", dispatchRunLockName)
	if result.code == 0 || !strings.Contains(result.stderr, "reserved artifact name") {
		t.Fatalf("dispatch lock artifact accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}

func TestC1RejectsDerivedArtifactNamespaceCollision(t *testing.T) {
	root := t.TempDir()
	prepared := preparedRunRoot{runPath: root}
	packet := filepath.Join(root, "output-attempt-1.json")
	output := filepath.Join(root, "output.json")
	receipt := filepath.Join(root, "receipt.json")
	handoff := filepath.Join(root, "handoff.json")
	schema := filepath.Join(root, "worker-output-schema.json")
	if err := validateDispatchArtifactNamespace(prepared, 2, packet, output, receipt, handoff, "", schema); err == nil || !strings.Contains(err.Error(), "artifact paths must be pairwise distinct") {
		t.Fatalf("derived attempt collision accepted: %v", err)
	}
}

func TestC1RejectsPreExistingDispatchArtifacts(t *testing.T) {
	root := t.TempDir()
	prepared := preparedRunRoot{runPath: root}
	output := filepath.Join(root, "output.json")
	receipt := filepath.Join(root, "receipt.json")
	handoff := filepath.Join(root, "handoff.json")
	schema := filepath.Join(root, "worker-output-schema.json")
	if err := os.WriteFile(filepath.Join(root, "output-attempt-1.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := validateDispatchArtifactNamespace(prepared, 2, filepath.Join(root, "packet.json"), output, receipt, handoff, "", schema)
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite existing dispatch artifact") {
		t.Fatalf("pre-existing attempt output accepted: %v", err)
	}
}

func TestC1FallbackUsesPerAttemptArtifacts(t *testing.T) {
	fixture := newDispatchFixture(t, "c1-attempt-artifacts")
	first := fixture.route("codex.small", "small-model", modelrouting.ClassSmall)
	fallback := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(first, fallback)
	fixture.trustRoutes(first, fallback)
	fixture.trustSession("session-large-c1", "large-model")
	fixture.withFakeCodex(fixture.fakeCodexSequence(fakeCodexSpec{exitCode: 7}, fakeCodexSpec{sessionID: "session-large-c1"}))

	result := fixture.run("--route-alias", first.Alias, "--fallback-route-alias", fallback.Alias, "--attempt-limit", "2")
	if result.code != 0 {
		t.Fatalf("fallback failed, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	report := decodeDispatchReport(t, result.stdout)
	if filepath.Base(report.OutputPath) == "output.json" || filepath.Base(report.ReceiptPath) == "receipt.json" {
		t.Fatalf("final fallback report did not point at per-attempt artifacts: %#v", report)
	}
	if !strings.Contains(filepath.Base(report.OutputPath), "attempt-2") || !strings.Contains(filepath.Base(report.ReceiptPath), "attempt-2") {
		t.Fatalf("final fallback artifact names are not attempt-scoped: %#v", report)
	}
	if _, err := os.Stat(filepath.Join(fixture.runRoot, "output-attempt-1.json")); err != nil {
		t.Fatalf("attempt-1 output not preserved: %v", err)
	}
}

func TestC1ContainmentUnavailableDoesNotStartOrFallback(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "c1-containment-unavailable")
	first := fixture.route("codex.small", "small-model", modelrouting.ClassSmall)
	fallback := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(first, fallback)
	fixture.trustRoutes(first, fallback)
	fake := fixture.fakeCodexSequence(fakeCodexSpec{exitCode: 7}, fakeCodexSpec{})
	fixture.withFakeCodex(fake)

	previous := dispatchProcessTreeContainment
	dispatchProcessTreeContainment = func() error { return errors.New("fixture containment unavailable") }
	t.Cleanup(func() { dispatchProcessTreeContainment = previous })

	result := fixture.run("--route-alias", first.Alias, "--fallback-route-alias", fallback.Alias, "--attempt-limit", "2")
	if result.code == 0 || !strings.Contains(result.stderr, "dispatch unavailable before worker start") {
		t.Fatalf("containment-unavailable dispatch was misreported: code=%d stdout=%s stderr=%s", result.code, result.stdout, result.stderr)
	}
	if _, err := os.Stat(fake.argsPath); !os.IsNotExist(err) {
		t.Fatalf("worker started or argv probe failed: %v", err)
	}
	for _, path := range []string{fixture.outputPath, fixture.receiptPath, fixture.handoffPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("pre-start unavailability wrote execution artifact %s: %v", path, err)
		}
	}
}

func TestC1SessionEvidenceRequiresOptionalProfileAndRouteBindings(t *testing.T) {
	root := t.TempDir()
	codexHome := filepath.Join(root, "codex-home")
	if err := os.MkdirAll(filepath.Join(codexHome, "sessions"), 0o700); err != nil {
		t.Fatal(err)
	}
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	req := modelrouting.DispatchRequest{
		CWD: projectRoot, Model: "large-model", Sandbox: "workspace-write", ApprovalPolicy: "never",
		Profile: "localprof", ProfileRevision: "sha256:good", RouteAlias: "profile.large",
	}
	route := modelrouting.Route{Alias: "profile.large", Destination: "codex", Profile: "localprof", ProfileRevision: "sha256:good"}
	sessionID := "session-c1-profile"
	lines := []string{
		`{"type":"session_meta","payload":{"id":"` + sessionID + `","model_provider":"codex","cwd":"` + filepath.ToSlash(projectRoot) + `"}}`,
		`{"type":"turn_context","payload":{"model":"large-model","cwd":"` + filepath.ToSlash(projectRoot) + `","approval_policy":"never","sandbox_policy":{"type":"workspace-write"},"profile":"localprof","profile_revision":"sha256:evil","route_alias":"profile.large"}}`,
	}
	if err := os.WriteFile(filepath.Join(codexHome, "sessions", sessionID+".jsonl"), []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
	if evidence := loadCodexSessionEvidence(codexHome, sessionID, req, route, time.Now().Add(-time.Minute)); evidence.Model != "" || evidence.SessionID != "" {
		t.Fatalf("mismatched optional profile revision was credited: %#v", evidence)
	}
}

func TestC1SessionEvidenceFindsDatedCodexRollout(t *testing.T) {
	root := t.TempDir()
	codexHome := filepath.Join(root, "codex-home")
	projectRoot := filepath.Join(root, "project")
	sessionID := "thread-abc_123"
	sessionDir := filepath.Join(codexHome, "sessions", "2026", "07", "10")
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	sessionPath := filepath.Join(sessionDir, "rollout-20260710T120000-"+sessionID+".jsonl")
	writeCodexSessionLogForTest(t, sessionPath, sessionID, "codex", projectRoot, "large-model", "never", "workspace-write")
	req := modelrouting.DispatchRequest{CWD: projectRoot, Model: "large-model", Sandbox: "workspace-write", ApprovalPolicy: "never", RouteAlias: "codex.large"}
	route := modelrouting.Route{Alias: "codex.large", Destination: "codex"}
	evidence := loadCodexSessionEvidence(codexHome, sessionID, req, route, time.Now().Add(-time.Minute))
	if evidence.Model != "large-model" || evidence.SessionID != sessionID {
		t.Fatalf("dated rollout session evidence unavailable: %#v", evidence)
	}
}

func TestC1SessionEvidenceAcceptsLargeRolloutLogAfterEvidence(t *testing.T) {
	root := t.TempDir()
	codexHome := filepath.Join(root, "codex-home")
	projectRoot := filepath.Join(root, "project")
	sessionID := "thread-large-log"
	sessionDir := filepath.Join(codexHome, "sessions", "2026", "07", "10")
	if err := os.MkdirAll(sessionDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	sessionPath := filepath.Join(sessionDir, "rollout-20260710T120000-"+sessionID+".jsonl")
	writeCodexSessionLogForTest(t, sessionPath, sessionID, "codex", projectRoot, "large-model", "never", "workspace-write")
	file, err := os.OpenFile(sessionPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString("\n" + strings.Repeat(`{"type":"noise","payload":"padding-padding-padding-padding"}`+"\n", 70000)); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(sessionPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() <= defaultSessionEvidenceLimit {
		t.Fatalf("test fixture is not larger than evidence cap: size=%d cap=%d", info.Size(), defaultSessionEvidenceLimit)
	}
	req := modelrouting.DispatchRequest{CWD: projectRoot, Model: "large-model", Sandbox: "workspace-write", ApprovalPolicy: "never", RouteAlias: "codex.large"}
	route := modelrouting.Route{Alias: "codex.large", Destination: "codex"}
	evidence := loadCodexSessionEvidence(codexHome, sessionID, req, route, time.Now().Add(-time.Minute))
	if evidence.Model != "large-model" || evidence.SessionID != sessionID {
		t.Fatalf("large rollout session evidence unavailable: %#v", evidence)
	}
}

func TestC1SessionEvidenceDoesNotEnumerateHistoricalSessions(t *testing.T) {
	root := t.TempDir()
	codexHome := filepath.Join(root, "codex-home")
	projectRoot := filepath.Join(root, "project")
	oldDir := filepath.Join(codexHome, "sessions", "1999", "01", "01")
	currentDir := filepath.Join(codexHome, "sessions", "2026", "07", "10")
	if err := os.MkdirAll(oldDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(currentDir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 4105; index++ {
		name := filepath.Join(oldDir, "rollout-19990101T000000-noise-"+strconv.Itoa(index)+".jsonl")
		if err := os.WriteFile(name, []byte("{}\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sessionID := "thread-current-after-noise"
	sessionPath := filepath.Join(currentDir, "rollout-20260710T120000-"+sessionID+".jsonl")
	writeCodexSessionLogForTest(t, sessionPath, sessionID, "codex", projectRoot, "large-model", "never", "workspace-write")
	attemptStart := time.Date(2026, 7, 10, 12, 0, 0, 0, time.Local)
	if err := os.Chtimes(sessionPath, attemptStart.Add(time.Second), attemptStart.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	req := modelrouting.DispatchRequest{CWD: projectRoot, Model: "large-model", Sandbox: "workspace-write", ApprovalPolicy: "never", RouteAlias: "codex.large"}
	route := modelrouting.Route{Alias: "codex.large", Destination: "codex"}
	evidence := loadCodexSessionEvidence(codexHome, sessionID, req, route, attemptStart)
	if evidence.Model != "large-model" || evidence.SessionID != sessionID {
		t.Fatalf("historical session noise blocked current evidence: %#v", evidence)
	}
}

func TestC1SessionEvidenceRejectsStaleAmbiguousAndUnsafeSessions(t *testing.T) {
	root := t.TempDir()
	codexHome := filepath.Join(root, "codex-home")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(filepath.Join(codexHome, "sessions", "2026", "07", "10"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(codexHome, "sessions", "2026", "07", "11"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	req := modelrouting.DispatchRequest{CWD: projectRoot, Model: "large-model", Sandbox: "workspace-write", ApprovalPolicy: "never", RouteAlias: "codex.large"}
	route := modelrouting.Route{Alias: "codex.large", Destination: "codex"}

	staleID := "thread-stale"
	stalePath := filepath.Join(codexHome, "sessions", "2026", "07", "10", "rollout-20260710T120000-"+staleID+".jsonl")
	writeCodexSessionLogForTest(t, stalePath, staleID, "codex", projectRoot, "large-model", "never", "workspace-write")
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(stalePath, old, old); err != nil {
		t.Fatal(err)
	}
	if evidence := loadCodexSessionEvidence(codexHome, staleID, req, route, time.Now()); evidence.Model != "" {
		t.Fatalf("stale session evidence credited: %#v", evidence)
	}

	ambiguousID := "thread-ambiguous"
	for _, day := range []string{"10", "11"} {
		path := filepath.Join(codexHome, "sessions", "2026", "07", day, "rollout-202607"+day+"T120000-"+ambiguousID+".jsonl")
		writeCodexSessionLogForTest(t, path, ambiguousID, "codex", projectRoot, "large-model", "never", "workspace-write")
	}
	if evidence := loadCodexSessionEvidence(codexHome, ambiguousID, req, route, time.Now().Add(-time.Minute)); evidence.Model != "" {
		t.Fatalf("ambiguous session evidence credited: %#v", evidence)
	}
	if evidence := loadCodexSessionEvidence(codexHome, `..\escape`, req, route, time.Now().Add(-time.Minute)); evidence.Model != "" {
		t.Fatalf("unsafe session id credited: %#v", evidence)
	}
}

func TestC1ReadRunChildSingleOpenRejectsReplacement(t *testing.T) {
	fixture := newDispatchFixture(t, "c1-single-open")
	path := filepath.Join(fixture.runRoot, "single-open.json")
	if err := os.WriteFile(path, []byte(`{"ok":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	prepared, err := prepareRunRoot(fixture.projectRoot, fixture.runRoot)
	if err != nil {
		t.Fatal(err)
	}
	data, err := readRunChild(prepared, path, maxCatalogBytes)
	if err != nil {
		t.Fatalf("safe child read failed: %v", err)
	}
	var payload map[string]bool
	if err := json.Unmarshal(data, &payload); err != nil || !payload["ok"] {
		t.Fatalf("unexpected child data %q err=%v", data, err)
	}
}

func TestC1FallbackTrustTransitionRequiresApproval(t *testing.T) {
	base := modelrouting.Route{
		Alias: "first", DisplayModelID: "first", Boundary: modelrouting.BoundaryHosted, Destination: "provider-a", Endpoint: "https://a.example/v1",
		Retention: modelrouting.RetentionSession, TrainingUse: modelrouting.TrainingNo, Residency: "us",
		TrustProvenance: "trusted-source-a",
	}
	next := base
	next.Alias, next.DisplayModelID = "next", "next"
	next.Capability.RouteAlias, next.Capability.ModelID = next.Alias, next.DisplayModelID
	policy := modelrouting.PolicyContext{Project: modelrouting.ProjectPolicy{ProjectID: "project-a"}}

	cases := []struct {
		name  string
		setup func(*modelrouting.Route, *modelrouting.Route)
		want  bool
	}{
		{name: "unchanged trust", want: true},
		{name: "adapter revision changes", setup: func(_, r *modelrouting.Route) { r.AdapterRevision = "v2" }},
		{name: "boundary hosted to private", setup: func(_, r *modelrouting.Route) { r.Boundary = modelrouting.BoundaryPrivate }},
		{name: "boundary private to hosted", setup: func(first, _ *modelrouting.Route) { first.Boundary = modelrouting.BoundaryPrivate }},
		{name: "worse retention", setup: func(_, r *modelrouting.Route) { r.Retention = modelrouting.RetentionLimited }},
		{name: "training changes", setup: func(_, r *modelrouting.Route) { r.TrainingUse = modelrouting.TrainingUnknown }},
		{name: "residency changes", setup: func(_, r *modelrouting.Route) { r.Residency = "eu" }},
		{name: "destination changes", setup: func(_, r *modelrouting.Route) { r.Destination = "provider-b" }},
		{name: "endpoint changes", setup: func(_, r *modelrouting.Route) { r.Endpoint = "https://b.example/v1" }},
		{name: "trust provenance changes", setup: func(_, r *modelrouting.Route) { r.TrustProvenance = "trusted-source-b" }},
	}
	for _, tc := range cases {
		first := base
		candidate := next
		if tc.setup != nil {
			tc.setup(&first, &candidate)
		}
		got := fallbackTrustTransitionAllowed(first, candidate, policy)
		if got != tc.want {
			t.Fatalf("%s: got %v want %v", tc.name, got, tc.want)
		}
	}

	approved := next
	approved.Destination = "provider-b"
	sourcePolicy := policy
	fingerprint, err := modelrouting.ApprovalRouteFingerprint(approved, nil)
	if err != nil {
		t.Fatal(err)
	}
	sourcePolicy.Trusted.RouteApprovals = []modelrouting.RouteApproval{{
		ProjectID:        sourcePolicy.Project.ProjectID,
		RouteFingerprint: fingerprint,
		ExpiresAt:        time.Now().Add(time.Hour),
	}}
	if !fallbackTrustTransitionAllowed(base, approved, sourcePolicy) {
		t.Fatalf("approved provider transition was rejected")
	}
}

func TestC1ModelsApproveHostedRouteApprovalEnablesFallbackTransition(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct {
		alias       string
		model       string
		destination string
		endpoint    string
	}{
		{"hosted.a", "model-a", "provider-a", "https://a.example/v1"},
		{"hosted.b", "model-b", "provider-b", "https://b.example/v1"},
	} {
		code, stdout, stderr := runForTest("models", "add",
			"--scope", "user",
			"--user-root", userRoot,
			"--project-root", projectRoot,
			"--alias", route.alias,
			"--model", route.model,
			"--adapter", "codex",
			"--dispatch-method", "exec-model",
			"--destination", route.destination,
			"--endpoint", route.endpoint,
			"--boundary", "hosted",
			"--retention", "session",
			"--training-use", "no",
			"--residency", "us",
			"--trust-provenance", "hosted fixture",
			"--class", "large",
			"--json")
		if code != 0 {
			t.Fatalf("models add %s exit=%d stdout=%s stderr=%s", route.alias, code, stdout, stderr)
		}
	}
	code, stdout, stderr := runForTest("models", "approve",
		"--user-root", userRoot,
		"--project-root", projectRoot,
		"--alias", "hosted.b",
		"--expires-in", "1h",
		"--json")
	if code != 0 {
		t.Fatalf("models approve hosted route exit=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	first, _, _, err := routeApprovalInputs(userRoot, projectRoot, "hosted.a")
	if err != nil {
		t.Fatal(err)
	}
	next, projectID, fingerprint, err := routeApprovalInputs(userRoot, projectRoot, "hosted.b")
	if err != nil {
		t.Fatal(err)
	}
	trust, err := loadTrustFile(userRoot)
	if err != nil {
		t.Fatal(err)
	}
	project := findProjectTrust(trust, projectID)
	if len(project.RouteApprovals) != 1 || project.RouteApprovals[0].RouteFingerprint != fingerprint {
		t.Fatalf("hosted approval did not persist exact route approval: %#v fingerprint=%s", project.RouteApprovals, fingerprint)
	}
	policy := modelrouting.PolicyContext{
		Project: modelrouting.ProjectPolicy{ProjectID: projectID},
		Trusted: modelrouting.UserTrust{ProjectID: projectID, RouteApprovals: project.RouteApprovals},
	}
	if !fallbackTrustTransitionAllowed(first, next, policy) {
		t.Fatalf("production hosted route approval did not enable fallback transition")
	}
}

func TestC1ConcurrentDispatchLockPreservesFailedHandoff(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "c1-concurrent-dispatch")
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)

	fake := fixture.fakeCodex(fakeCodexSpec{exitCode: 7, sleepMillis: 250})
	previousResolver := dispatchExecutableResolver
	previousState := dispatchTrustedStateProvider
	previousHome := dispatchCodexHome
	dispatchExecutableResolver = func() (string, error) { return fake.path, nil }
	state := fixture.hostState
	dispatchTrustedStateProvider = func(string, string) (dispatchTrustedState, error) { return state, nil }
	dispatchCodexHome = func() (string, error) { return fixture.codexHome, nil }
	t.Cleanup(func() {
		dispatchExecutableResolver = previousResolver
		dispatchTrustedStateProvider = previousState
		dispatchCodexHome = previousHome
	})

	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := dispatchCodexWorker(dispatchOptions{
				commonOptions:     commonOptions{userRoot: fixture.userRoot, projectRoot: fixture.projectRoot, json: true},
				runRoot:           fixture.runRoot,
				runID:             filepath.Base(fixture.runRoot),
				sliceID:           "slice-004",
				packetPath:        "packet.json",
				outputPath:        "output.json",
				receiptPath:       "receipt.json",
				handoffPath:       "handoff.json",
				workerRequestPath: "worker-request.json",
				routeAlias:        route.Alias,
				timeout:           10 * time.Second,
				outputLimit:       defaultDispatchOutputLimit,
				attemptLimit:      2,
				sandbox:           "workspace-write",
				approvalPolicy:    "never",
				network:           "none",
			})
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)

	var workerFailures, namespaceRejections int
	for err := range results {
		if err == nil {
			t.Fatalf("unexpected concurrent dispatch success")
		}
		combined := err.Error()
		switch {
		case strings.Contains(combined, "worker exited nonzero"):
			workerFailures++
		case strings.Contains(combined, "refusing to overwrite existing dispatch artifact"):
			namespaceRejections++
		default:
			t.Fatalf("unexpected concurrent dispatch error: %v", err)
		}
	}
	if workerFailures != 1 || namespaceRejections != 1 {
		t.Fatalf("got workerFailures=%d namespaceRejections=%d, want one of each", workerFailures, namespaceRejections)
	}

	handoffData, err := os.ReadFile(fixture.handoffPath)
	if err != nil {
		t.Fatalf("read handoff: %v", err)
	}
	var handoff struct {
		Attempts []json.RawMessage `json:"attempts"`
	}
	if err := json.Unmarshal(handoffData, &handoff); err != nil {
		t.Fatalf("decode handoff: %v", err)
	}
	if len(handoff.Attempts) != 1 {
		t.Fatalf("handoff attempts=%d, want exactly one preserved attempt; handoff=%s", len(handoff.Attempts), handoffData)
	}
}

func TestC1DispatchLockReusesPersistentUnlockedStateFile(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	root := t.TempDir()
	lockRoot := filepath.Join(root, "user", "dispatch-state", "project", "run")
	first, err := modelrouting.AcquirePrivateStateLock(lockRoot, "dispatch.lock", 2*time.Second)
	if err != nil {
		t.Fatalf("first lock: %v", err)
	}
	lockPath := filepath.Join(lockRoot, "dispatch.lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("persistent lock file missing: %v", err)
	}
	if err := first.Close(); err != nil {
		t.Fatalf("close first lock: %v", err)
	}
	second, err := modelrouting.AcquirePrivateStateLock(lockRoot, "dispatch.lock", 2*time.Second)
	if err != nil {
		t.Fatalf("unlocked persistent lock file blocked reacquire: %v", err)
	}
	if err := second.Close(); err != nil {
		t.Fatalf("close second lock: %v", err)
	}
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("lock file should remain stable for reuse: %v", err)
	}
}

func TestC1DispatchLockSerializesConcurrentOwners(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	root := t.TempDir()
	lockRoot := filepath.Join(root, "user", "dispatch-state", "project", "run")
	held, err := modelrouting.AcquirePrivateStateLock(lockRoot, "dispatch.lock", 2*time.Second)
	if err != nil {
		t.Fatalf("held lock: %v", err)
	}
	acquired := make(chan error, 1)
	go func() {
		lock, lockErr := modelrouting.AcquirePrivateStateLock(lockRoot, "dispatch.lock", 50*time.Millisecond)
		if lockErr == nil {
			_ = lock.Close()
		}
		acquired <- lockErr
	}()
	if err := <-acquired; err == nil || !strings.Contains(err.Error(), "timed out waiting for private state lock") {
		t.Fatalf("second lock was not bounded by deadline: %v", err)
	}
	if err := held.Close(); err != nil {
		t.Fatalf("close held lock: %v", err)
	}
	reacquired, err := modelrouting.AcquirePrivateStateLock(lockRoot, "dispatch.lock", 2*time.Second)
	if err != nil {
		t.Fatalf("lock did not release after owner close: %v", err)
	}
	_ = reacquired.Close()
}

func skipIfPrivateACLUnsupported(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		return
	}
	root := filepath.Join(t.TempDir(), "private-acl-probe")
	lock, err := modelrouting.AcquirePrivateStateLock(root, "probe.lock", 100*time.Millisecond)
	if err == nil {
		_ = lock.Close()
		return
	}
	if strings.Contains(err.Error(), "Access is denied") {
		t.Skipf("workspace sandbox denies private Windows ACL setup: %v", err)
	}
	t.Fatalf("private ACL probe failed unexpectedly: %v", err)
}

func writeCodexSessionLogForTest(t *testing.T, path, sessionID, provider, cwd, model, approval, sandbox string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	lines := []string{
		`{"type":"session_meta","payload":{"id":"` + sessionID + `","model_provider":"` + provider + `","cwd":"` + filepath.ToSlash(cwd) + `"}}`,
		`{"type":"turn_context","payload":{"model":"` + model + `","cwd":"` + filepath.ToSlash(cwd) + `","approval_policy":"` + approval + `","sandbox_policy":{"type":"` + sandbox + `"}}}`,
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestC1ProfileRejectsUnsupportedCredentialAmbiguity(t *testing.T) {
	fixture := newDispatchFixture(t, "c1-profile")
	name := "ambiguous"
	if err := os.WriteFile(filepath.Join(fixture.codexHome, name+".config.toml"), []byte("model_provider = \"codex\"\nbase_url = \"https://api.example/v1\"\napi_key_env_var = \"ROUTE_TOKEN\"\napi_key = \"$ROUTE_TOKEN\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := trustedCodexProfileRevision(fixture.codexHome, name, "https://api.example/v1", "ROUTE_TOKEN"); err == nil || !strings.Contains(err.Error(), "unsupported credential") {
		t.Fatalf("ambiguous static credential profile accepted: %v", err)
	}
}

func TestC1TrustedStateRejectsExpiredState(t *testing.T) {
	fixture := newDispatchFixture(t, "c1-expired-state")
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	prepared, err := prepareRunRoot(fixture.projectRoot, fixture.runRoot)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := modelrouting.LoadCatalog(fixture.runRoot, "catalog.json", modelrouting.StorageOptions{MaxBytes: maxCatalogBytes, Source: modelrouting.CatalogSourceRun})
	if err != nil {
		t.Fatal(err)
	}
	if err := saveDispatchTrustedState(fixture.userRoot, prepared, catalog); err != nil {
		t.Fatal(err)
	}
	relDir, file := dispatchStateLocation(prepared.marker.ProjectID, prepared.marker.RunID)
	var state dispatchTrustedState
	if err := modelrouting.LoadStrictJSON(filepath.Join(fixture.userRoot, relDir), file, &state, maxCatalogBytes); err != nil {
		t.Fatal(err)
	}
	state.ExpiresAt = time.Now().Add(-time.Minute)
	if err := modelrouting.SaveAtomicJSON(filepath.Join(fixture.userRoot, relDir), file, state, maxCatalogBytes); err != nil {
		t.Fatal(err)
	}
	fixture.withFakeCodexResolverOnly(fixture.fakeCodex(fakeCodexSpec{}))
	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "expired") {
		t.Fatalf("expired private route state accepted, code=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
}
