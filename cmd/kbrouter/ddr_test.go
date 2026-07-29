package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

type ddrServerSpec struct {
	modelStatus int
	models      []string
	chatStatus  int
	chatBody    string
	modelDelay  time.Duration
	chatDelay   time.Duration
}

type ddrServerCounts struct {
	models        atomic.Int32
	chat          atomic.Int32
	authorization atomic.Value
	chatRequest   atomic.Value
}

func TestDDRAttemptFunctionalOutcomesAreBoundedAndNeverRetry(t *testing.T) {
	cases := []struct {
		name        string
		spec        ddrServerSpec
		approved    bool
		proof       string
		wantStatus  string
		wantFailure string
		wantPhase   string
		wantModels  int32
		wantChat    int32
		maxElapsed  time.Duration
		authEnv     string
		authValue   string
	}{
		{name: "available", spec: availableDDRServer(), approved: true, proof: "pass", wantStatus: "completed", wantModels: 1, wantChat: 1, maxElapsed: time.Second, authEnv: "LOCAL_DDR_TEST_KEY", authValue: "fixture-token"},
		{name: "untrusted", spec: availableDDRServer(), approved: false, proof: "pass", wantStatus: "parent-return", wantFailure: "untrusted", wantPhase: "eligibility", wantModels: 0, wantChat: 0, maxElapsed: time.Second},
		{name: "probe-unauthorized", spec: ddrServerSpec{modelStatus: http.StatusUnauthorized}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "unauthorized", wantPhase: "probe", wantModels: 1, wantChat: 0, maxElapsed: time.Second, authEnv: "LOCAL_DDR_TEST_KEY", authValue: "wrong-token"},
		{name: "model-missing", spec: ddrServerSpec{modelStatus: http.StatusOK, models: []string{"different-model"}}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "model-missing", wantPhase: "probe", wantModels: 1, wantChat: 0, maxElapsed: time.Second},
		{name: "probe-timeout", spec: ddrServerSpec{modelStatus: http.StatusOK, models: []string{"operator-model"}, modelDelay: 150 * time.Millisecond}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "timeout", wantPhase: "probe", wantModels: 1, wantChat: 0, maxElapsed: 500 * time.Millisecond},
		{name: "probe-server-error", spec: ddrServerSpec{modelStatus: http.StatusServiceUnavailable}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "server-error", wantPhase: "probe", wantModels: 1, wantChat: 0, maxElapsed: time.Second},
		{name: "dispatch-unauthorized", spec: ddrServerSpec{modelStatus: http.StatusOK, models: []string{"operator-model"}, chatStatus: http.StatusUnauthorized}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "unauthorized", wantPhase: "dispatch", wantModels: 1, wantChat: 1, maxElapsed: time.Second},
		{name: "dispatch-timeout", spec: ddrServerSpec{modelStatus: http.StatusOK, models: []string{"operator-model"}, chatDelay: 150 * time.Millisecond}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "timeout", wantPhase: "dispatch", wantModels: 1, wantChat: 1, maxElapsed: 500 * time.Millisecond},
		{name: "server-error", spec: ddrServerSpec{modelStatus: http.StatusOK, models: []string{"operator-model"}, chatStatus: http.StatusServiceUnavailable}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "server-error", wantPhase: "dispatch", wantModels: 1, wantChat: 1, maxElapsed: time.Second},
		{name: "dispatch-failed", spec: ddrServerSpec{modelStatus: http.StatusOK, models: []string{"operator-model"}, chatStatus: http.StatusOK, chatBody: `{"choices":[]}`}, approved: true, proof: "pass", wantStatus: "parent-return", wantFailure: "dispatch-failed", wantPhase: "dispatch", wantModels: 1, wantChat: 1, maxElapsed: time.Second},
		{name: "proof-failed", spec: availableDDRServer(), approved: true, proof: "fail", wantStatus: "parent-return", wantFailure: "proof-failed", wantPhase: "proof", wantModels: 1, wantChat: 1, maxElapsed: time.Second},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server, counts := newDDRServer(t, tc.spec)
			defer server.Close()
			fixture := newDDRAttemptFixture(t, server.URL+"/v1", tc.authEnv, tc.approved)
			if tc.authEnv != "" {
				t.Setenv(tc.authEnv, tc.authValue)
			}
			start := time.Now()
			code, stdout, stderr := fixture.run()
			elapsed := time.Since(start)
			report := decodeDDRAttemptReport(t, stdout)
			if tc.name == "available" || tc.name == "proof-failed" {
				if code != 0 || report.Status != "awaiting-proof" || report.Action != "run-proof" ||
					report.Response != "bounded result" || report.ResultDelivery != "first-response-only" {
					t.Fatalf("attempt did not return one result for proof: code=%d report=%#v stderr=%s", code, report, stderr)
				}
				receipt := fixture.persistedReceipt()
				if strings.Contains(receipt, "bounded result") || (tc.authValue != "" && strings.Contains(receipt, tc.authValue)) {
					t.Fatalf("receipt persisted response or auth secret: %s", receipt)
				}
				code, stdout, stderr = fixture.resolve(tc.proof)
				report = decodeDDRAttemptReport(t, stdout)
			}
			if tc.wantStatus == "completed" && code != 0 {
				t.Fatalf("resolve exit=%d stderr=%s stdout=%s", code, stderr, stdout)
			}
			if tc.wantStatus == "parent-return" && code != ddrParentReturnExitCode {
				t.Fatalf("parent return exit=%d stderr=%s stdout=%s", code, stderr, stdout)
			}
			if report.Status != tc.wantStatus || report.FailureClass != tc.wantFailure {
				t.Fatalf("report=%#v", report)
			}
			if report.FailurePhase != tc.wantPhase {
				t.Fatalf("failure phase=%q want=%q report=%#v", report.FailurePhase, tc.wantPhase, report)
			}
			if report.Attempt > 1 || report.ParentSelection != "active-parent-or-host-native" {
				t.Fatalf("unbounded or provider-specific parent return: %#v", report)
			}
			if elapsed > tc.maxElapsed {
				t.Fatalf("elapsed=%s bound=%s report=%#v", elapsed, tc.maxElapsed, report)
			}
			if got := counts.models.Load(); got != tc.wantModels {
				t.Fatalf("model probes=%d want=%d", got, tc.wantModels)
			}
			if got := counts.chat.Load(); got != tc.wantChat {
				t.Fatalf("chat attempts=%d want=%d", got, tc.wantChat)
			}
			if tc.name == "available" {
				if got, _ := counts.authorization.Load().(string); got != "Bearer "+tc.authValue {
					t.Fatalf("authorization header=%q", got)
				}
				payload, _ := counts.chatRequest.Load().(string)
				if strings.Contains(payload, tc.authValue) || !strings.Contains(payload, `"model":"operator-model"`) ||
					!strings.Contains(payload, `"stream":false`) {
					t.Fatalf("unexpected dispatch payload: %s", payload)
				}
			}
			if report.TotalLatencyMS < 0 || report.ProbeLatencyMS < 0 || report.DispatchLatencyMS < 0 {
				t.Fatalf("latency missing: %#v", report)
			}

			receipt := fixture.persistedReceipt()
			if strings.Contains(receipt, "bounded result") || (tc.authValue != "" && strings.Contains(receipt, tc.authValue)) {
				t.Fatalf("receipt persisted response or auth secret: %s", receipt)
			}

			secondCode, secondStdout, secondStderr := fixture.run()
			second := decodeDDRAttemptReport(t, secondStdout)
			if secondCode != code {
				t.Fatalf("replay exit changed: first=%d second=%d stderr=%s", code, secondCode, secondStderr)
			}
			if second.RunID != report.RunID || second.SliceID != report.SliceID || counts.models.Load() != tc.wantModels || counts.chat.Load() != tc.wantChat {
				t.Fatalf("second local attempt occurred: first=%#v second=%#v models=%d chat=%d", report, second, counts.models.Load(), counts.chat.Load())
			}
		})
	}
}

func TestDDRAttemptConcurrentReservationAllowsOneNetworkAttempt(t *testing.T) {
	spec := availableDDRServer()
	spec.chatDelay = 80 * time.Millisecond
	server, counts := newDDRServer(t, spec)
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	type result struct {
		code   int
		stdout string
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			code, stdout, _ := fixture.run()
			results <- result{code: code, stdout: stdout}
		}()
	}
	close(start)
	workers.Wait()
	close(results)
	codes := map[int]int{}
	for item := range results {
		codes[item.code]++
		_ = decodeDDRAttemptReport(t, item.stdout)
	}
	if counts.models.Load() != 1 || counts.chat.Load() != 1 || codes[0] != 1 || codes[ddrParentReturnExitCode] != 1 {
		t.Fatalf("reservation ordering failed: models=%d chat=%d codes=%v", counts.models.Load(), counts.chat.Load(), codes)
	}
}

func TestDDRAttemptRequirePinBlocksInsteadOfReturningToParent(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", false)
	code, stdout, stderr := fixture.runWith("--require")
	if code == 0 || code == ddrParentReturnExitCode {
		t.Fatalf("require pin did not block: exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	report := decodeDDRAttemptReport(t, stdout)
	if report.Status != "blocked" || report.Action != "block-required-route" || report.FailureClass != "untrusted" {
		t.Fatalf("require report=%#v", report)
	}
	if counts.models.Load() != 0 || counts.chat.Load() != 0 {
		t.Fatalf("untrusted required route contacted endpoint")
	}
}

func TestDDRAttemptRequiresRouteApprovalForHostedBoundary(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", false)
	catalog := loadUserCatalogForTest(t, fixture.userRoot)
	catalog.Routes[0].Boundary = modelrouting.BoundaryHosted
	catalog.Routes[0].Endpoint = "https://example.com/v1"
	if err := saveUserCatalog(fixture.userRoot, catalog); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := fixture.run()
	report := decodeDDRAttemptReport(t, stdout)
	if code != ddrParentReturnExitCode || report.FailureClass != "untrusted" {
		t.Fatalf("hosted unapproved route exit=%d report=%#v stderr=%s", code, report, stderr)
	}
	if counts.models.Load() != 0 || counts.chat.Load() != 0 {
		t.Fatal("hosted unapproved route contacted endpoint")
	}
}

func TestDDRAttemptHonorsProjectAliasPolicyBeforeNetwork(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	writeJSONForTest(t, filepath.Join(fixture.projectRoot, "kb-models.json"), map[string]any{
		"schema_version":  1,
		"allowed_aliases": []string{"different.route"},
	})
	code, stdout, stderr := fixture.run()
	report := decodeDDRAttemptReport(t, stdout)
	if code != ddrParentReturnExitCode || report.FailureClass != "untrusted" {
		t.Fatalf("policy-rejected route exit=%d report=%#v stderr=%s", code, report, stderr)
	}
	if counts.models.Load() != 0 || counts.chat.Load() != 0 {
		t.Fatal("policy-rejected route contacted endpoint")
	}
}

func TestDDRRequiredProofFailureBlocksWithExactExit(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	code, stdout, stderr := fixture.runWith("--require")
	attempt := decodeDDRAttemptReport(t, stdout)
	if code != 0 || attempt.Status != "awaiting-proof" {
		t.Fatalf("required attempt exit=%d report=%#v stderr=%s", code, attempt, stderr)
	}
	code, stdout, stderr = fixture.resolve("fail")
	resolved := decodeDDRAttemptReport(t, stdout)
	if code != ddrBlockedExitCode || resolved.Status != "blocked" || resolved.FailureClass != "proof-failed" {
		t.Fatalf("required proof failure exit=%d report=%#v stderr=%s", code, resolved, stderr)
	}
	if counts.models.Load() != 1 || counts.chat.Load() != 1 {
		t.Fatalf("required proof resolution retried network")
	}
}

func TestDDRRejectsConflictingAttemptAndProofReplays(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	code, _, stderr := fixture.run()
	if code != 0 {
		t.Fatalf("attempt exit=%d stderr=%s", code, stderr)
	}
	code, _, stderr = fixture.resolve("pass")
	if code != 0 {
		t.Fatalf("resolve exit=%d stderr=%s", code, stderr)
	}
	writeJSONForTest(t, fixture.requestPath, map[string]any{
		"schema_version": 1,
		"messages":       []map[string]string{{"role": "user", "content": "Changed request."}},
		"max_tokens":     64,
	})
	code, stdout, stderr := fixture.run()
	if code == 0 || !strings.Contains(stdout+stderr, "binding conflict") {
		t.Fatalf("conflicting attempt replay exit=%d output=%s%s", code, stdout, stderr)
	}
	code, stdout, stderr = fixture.resolve("fail")
	if code == 0 || !strings.Contains(stdout+stderr, "proof replay conflict") {
		t.Fatalf("conflicting proof replay exit=%d output=%s%s", code, stdout, stderr)
	}
	if counts.models.Load() != 1 || counts.chat.Load() != 1 {
		t.Fatalf("conflicting replay retried network")
	}
}

func TestDDRAttemptRejectsInvalidTierAndImplicitSensitivity(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	code, stdout, stderr := fixture.runWith("--tier", "smal")
	if code != 2 || !strings.Contains(stdout+stderr, "tier must be") {
		t.Fatalf("invalid tier exit=%d output=%s%s", code, stdout, stderr)
	}
	args := []string{
		"ddr", "attempt", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot,
		"--run-id", "run-2", "--slice-id", "slice-2", "--alias", "local.coder",
		"--tier", "small", "--tier-reason", "bounded", "--task-family", "code",
		"--tool", "text", "--context-size", "1024", "--risk", "normal",
		"--request", fixture.requestPath, "--json",
	}
	code, stdout, stderr = runForTest(args...)
	if code != 2 || !strings.Contains(stdout+stderr, "sensitive-data must be explicitly set") {
		t.Fatalf("implicit sensitivity exit=%d output=%s%s", code, stdout, stderr)
	}
	if counts.models.Load() != 0 || counts.chat.Load() != 0 {
		t.Fatalf("invalid input contacted endpoint")
	}
}

func TestDDRCommandsExposeFocusedHelpAndStructuredUsageErrors(t *testing.T) {
	for _, args := range [][]string{
		{"ddr", "attempt", "--help"},
		{"ddr", "resolve", "--help"},
		{"models", "import", "--help"},
	} {
		code, stdout, stderr := runForTest(args...)
		if code != 0 || !strings.Contains(stdout, "Usage:") || stderr != "" {
			t.Fatalf("help args=%v exit=%d stdout=%s stderr=%s", args, code, stdout, stderr)
		}
	}
	code, stdout, stderr := runForTest("ddr", "attempt", "--json", "--bogus")
	if code != 2 || stderr != "" {
		t.Fatalf("structured usage exit=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	var envelope struct {
		SchemaVersion int `json:"schema_version"`
		Status        string
		Error         struct {
			Code    string
			Message string
		}
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("usage envelope=%q err=%v", stdout, err)
	}
	if envelope.SchemaVersion != 1 || envelope.Status != "error" ||
		envelope.Error.Code != "invalid-arguments" || envelope.Error.Message == "" {
		t.Fatalf("usage envelope=%#v", envelope)
	}
}

func TestCheckedInDDRRequestExampleIsARejectedPlaceholder(t *testing.T) {
	projectRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	examplePath := filepath.Join(projectRoot, "config", "kbrouter-ddr-request.example.json")
	_, err = loadDDRRequest(projectRoot, examplePath)
	if err == nil || !strings.Contains(err.Error(), "invalid DDR request message") {
		t.Fatalf("checked-in request placeholder accepted: %v", err)
	}
}

func TestDDRAttemptPendingProofReplayReturnsToParentWithoutRetry(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	code, stdout, stderr := fixture.run()
	first := decodeDDRAttemptReport(t, stdout)
	if code != 0 || first.Status != "awaiting-proof" {
		t.Fatalf("first attempt exit=%d report=%#v stderr=%s", code, first, stderr)
	}
	code, stdout, stderr = fixture.run()
	replay := decodeDDRAttemptReport(t, stdout)
	if code != ddrParentReturnExitCode || replay.Status != "parent-return" || replay.FailureClass != "result-not-retained" {
		t.Fatalf("pending replay exit=%d report=%#v stderr=%s", code, replay, stderr)
	}
	if counts.models.Load() != 1 || counts.chat.Load() != 1 {
		t.Fatalf("pending replay retried network: models=%d chat=%d", counts.models.Load(), counts.chat.Load())
	}
	code, stdout, stderr = fixture.resolve("pass")
	resolved := decodeDDRAttemptReport(t, stdout)
	if code != ddrParentReturnExitCode || resolved.Status != "parent-return" {
		t.Fatalf("terminal parent return changed during resolve: exit=%d report=%#v stderr=%s", code, resolved, stderr)
	}
}

func TestDDRAttemptReservationPreventsRetryAfterUncertainDispatch(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	root, name := fixture.receiptLocation()
	report := fixture.attemptingReport(false)
	if err := modelrouting.SaveAtomicJSON(root, name, report, maxCatalogBytes); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := fixture.run()
	report = decodeDDRAttemptReport(t, stdout)
	if code != ddrParentReturnExitCode || report.Status != "parent-return" || report.FailureClass != "attempt-state-uncertain" {
		t.Fatalf("uncertain replay exit=%d report=%#v stderr=%s", code, report, stderr)
	}
	if counts.models.Load() != 0 || counts.chat.Load() != 0 {
		t.Fatalf("uncertain attempt retried network")
	}
}

func TestDDRAttemptRequiredReservationReplayBlocksWithoutRetry(t *testing.T) {
	server, counts := newDDRServer(t, availableDDRServer())
	defer server.Close()
	fixture := newDDRAttemptFixture(t, server.URL+"/v1", "", true)
	root, name := fixture.receiptLocation()
	report := fixture.attemptingReport(true)
	if err := modelrouting.SaveAtomicJSON(root, name, report, maxCatalogBytes); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := fixture.runWith("--require")
	replay := decodeDDRAttemptReport(t, stdout)
	if code != ddrBlockedExitCode || replay.Status != "blocked" || replay.FailureClass != "attempt-state-uncertain" {
		t.Fatalf("required replay exit=%d report=%#v stderr=%s", code, replay, stderr)
	}
	if counts.models.Load() != 0 || counts.chat.Load() != 0 {
		t.Fatalf("required uncertain attempt retried network")
	}
}

func availableDDRServer() ddrServerSpec {
	return ddrServerSpec{
		modelStatus: http.StatusOK, models: []string{"operator-model"},
		chatStatus: http.StatusOK,
		chatBody:   `{"model":"operator-model","choices":[{"message":{"role":"assistant","content":"bounded result"}}]}`,
	}
}

func newDDRServer(t *testing.T, spec ddrServerSpec) (*httptest.Server, *ddrServerCounts) {
	t.Helper()
	counts := &ddrServerCounts{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counts.authorization.Store(r.Header.Get("Authorization"))
		switch r.URL.Path {
		case "/v1/models":
			counts.models.Add(1)
			if spec.modelDelay > 0 {
				time.Sleep(spec.modelDelay)
			}
			status := spec.modelStatus
			if status == 0 {
				status = http.StatusOK
			}
			w.WriteHeader(status)
			if status == http.StatusOK {
				data := make([]map[string]string, 0, len(spec.models))
				for _, model := range spec.models {
					data = append(data, map[string]string{"id": model})
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
			}
		case "/v1/chat/completions":
			counts.chat.Add(1)
			if spec.chatDelay > 0 {
				time.Sleep(spec.chatDelay)
			}
			body, _ := io.ReadAll(r.Body)
			counts.chatRequest.Store(string(body))
			status := spec.chatStatus
			if status == 0 {
				status = http.StatusOK
			}
			w.WriteHeader(status)
			if spec.chatBody != "" {
				_, _ = w.Write([]byte(spec.chatBody))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	return server, counts
}

type ddrAttemptFixture struct {
	t           *testing.T
	userRoot    string
	projectRoot string
	requestPath string
}

func newDDRAttemptFixture(t *testing.T, endpoint, authEnv string, approved bool) ddrAttemptFixture {
	t.Helper()
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	importPath := filepath.Join(root, "routes.json")
	writeJSONForTest(t, importPath, map[string]any{
		"schema_version": 1,
		"routes": []map[string]any{{
			"alias": "local.coder", "model_id": "operator-model", "endpoint": endpoint,
			"auth_env": authEnv, "boundary": "private", "hosting": "self-hosted", "retention": "none",
			"training_use": "no", "residency": "local", "trust_provenance": "test operator",
			"capability": map[string]any{
				"class": "small", "task_family": "code", "tools": []string{"text"},
				"context_size": 4096, "risk": "normal",
			},
		}},
	})
	code, stdout, stderr := runForTest("models", "import", "--user-root", userRoot, "--project-root", projectRoot, "--file", importPath, "--json")
	if code != 0 {
		t.Fatalf("fixture import exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if approved {
		route := loadUserCatalogForTest(t, userRoot).Routes[0]
		projectID, err := modelrouting.CanonicalProjectIdentity(projectRoot)
		if err != nil {
			t.Fatal(err)
		}
		fingerprint, err := modelrouting.ApprovalRouteFingerprint(route, nil)
		if err != nil {
			t.Fatal(err)
		}
		trust := approveRouteTrust(userTrustFileData{SchemaVersion: 1}, projectID, route, fingerprint, time.Now().Add(time.Hour))
		if err := saveTrustFile(userRoot, trust); err != nil {
			t.Fatal(err)
		}
	}
	requestPath := filepath.Join(projectRoot, "request.json")
	writeJSONForTest(t, requestPath, map[string]any{
		"schema_version": 1,
		"messages":       []map[string]string{{"role": "user", "content": "Return a bounded result."}},
		"max_tokens":     64,
	})
	return ddrAttemptFixture{t: t, userRoot: userRoot, projectRoot: projectRoot, requestPath: requestPath}
}

func (f ddrAttemptFixture) run() (int, string, string) {
	return f.runWith()
}

func (f ddrAttemptFixture) runWith(extra ...string) (int, string, string) {
	f.t.Helper()
	args := []string{
		"ddr", "attempt", "--user-root", f.userRoot, "--project-root", f.projectRoot,
		"--run-id", "run-1", "--slice-id", "slice-1", "--alias", "local.coder",
		"--tier", "small", "--tier-reason", "bounded text task", "--task-family", "code",
		"--tool", "text", "--context-size", "1024", "--risk", "normal",
		"--sensitive-data=false",
		"--request", f.requestPath, "--probe-timeout", "40ms", "--timeout", "100ms",
		"--json",
	}
	args = append(args, extra...)
	return runForTest(args...)
}

func (f ddrAttemptFixture) resolve(proof string) (int, string, string) {
	f.t.Helper()
	return runForTest(
		"ddr", "resolve", "--user-root", f.userRoot, "--project-root", f.projectRoot,
		"--run-id", "run-1", "--slice-id", "slice-1",
		"--proof-result", proof, "--proof-command", "go test ./...",
		"--proof-artifact-hash", modelrouting.SHA256Bytes([]byte("proof")), "--json",
	)
}

func (f ddrAttemptFixture) persistedReceipt() string {
	f.t.Helper()
	root, name := f.receiptLocation()
	return readFileForTest(f.t, filepath.Join(root, name))
}

func (f ddrAttemptFixture) receiptLocation() (string, string) {
	f.t.Helper()
	projectID, err := modelrouting.CanonicalProjectIdentity(f.projectRoot)
	if err != nil {
		f.t.Fatal(err)
	}
	return ddrAttemptReceiptLocation(f.userRoot, projectID, "run-1", "slice-1")
}

func (f ddrAttemptFixture) attemptingReport(require bool) ddrAttemptReport {
	f.t.Helper()
	opts := ddrAttemptOptions{
		commonOptions: commonOptions{userRoot: f.userRoot, projectRoot: f.projectRoot},
		runID:         "run-1", sliceID: "slice-1", alias: "local.coder",
		tier: "small", tierReason: "bounded text task", taskFamily: "code",
		tools: repeatFlag{"text"}, contextSize: 1024, risk: "normal",
		requestPath: f.requestPath, sensitiveSet: true, require: require,
		probeTimeout: 40 * time.Millisecond, timeout: 100 * time.Millisecond,
	}
	projectID, err := modelrouting.CanonicalProjectIdentity(f.projectRoot)
	if err != nil {
		f.t.Fatal(err)
	}
	request, err := loadDDRRequest(f.projectRoot, f.requestPath)
	if err != nil {
		f.t.Fatal(err)
	}
	requestBytes, err := json.Marshal(request)
	if err != nil {
		f.t.Fatal(err)
	}
	requestHash := modelrouting.SHA256Bytes(requestBytes)
	route := loadUserCatalogForTest(f.t, f.userRoot).Routes[0]
	policy, err := policyContextForProject(f.userRoot, f.projectRoot)
	if err != nil {
		f.t.Fatal(err)
	}
	fingerprint, err := modelrouting.ApprovalRouteFingerprint(route, policy.RouteSources)
	if err != nil {
		f.t.Fatal(err)
	}
	binding, err := ddrAttemptBindingHash(opts, projectID, requestHash, fingerprint)
	if err != nil {
		f.t.Fatal(err)
	}
	return ddrAttemptReport{
		SchemaVersion: ddrAttemptSchemaVersion, Status: "attempting", Action: "await-local-attempt",
		ProjectID: projectID, RunID: "run-1", SliceID: "slice-1", RouteAlias: "local.coder",
		RouteFingerprint: fingerprint, RequestHash: requestHash, AttemptBinding: binding,
		Attempt: 1, RequirePin: require, ObservedAt: time.Now().UTC(),
	}
}

func decodeDDRAttemptReport(t *testing.T, value string) ddrAttemptReport {
	t.Helper()
	var report ddrAttemptReport
	if err := json.Unmarshal([]byte(strings.TrimSpace(value)), &report); err != nil {
		t.Fatalf("decode report %q: %v", value, err)
	}
	if report.SchemaVersion != 1 {
		t.Fatalf("report schema=%d report=%#v", report.SchemaVersion, report)
	}
	return report
}

func (r ddrAttemptReport) String() string {
	return fmt.Sprintf("%s/%s attempt=%d failure=%s", r.RunID, r.SliceID, r.Attempt, r.FailureClass)
}
