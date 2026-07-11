package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestCatalogDiscoverWritesOnlyRunCatalogAndBoundsSlowAdapter(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	userRoot := filepath.Join(root, "user")

	code, stdout, stderr := runForTest("models", "discover",
		"--run-root", runRoot,
		"--user-root", userRoot,
		"--project-root", projectRoot,
		"--current-model", "gpt-5.5",
		"--adapter-timeout", "10ms",
		"--session-timeout", "50ms",
		"--include-slow-fixture",
		"--json")
	if code != 0 {
		t.Fatalf("discover exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if _, err := os.Stat(filepath.Join(runRoot, "catalog.json")); err != nil {
		t.Fatalf("expected run catalog: %v", err)
	}
	assertNotExists(t, filepath.Join(userRoot, "models.json"))
	assertNotExists(t, filepath.Join(projectRoot, "kb-models.json"))
	if !strings.Contains(stdout, `"slow-fixture"`) || !strings.Contains(stdout, `"status":"unavailable"`) {
		t.Fatalf("expected bounded slow adapter status in %s", stdout)
	}
	if !strings.Contains(stdout, `"current"`) || !strings.Contains(stdout, "gpt-5.5") {
		t.Fatalf("current model should survive slow adapter failure: %s", stdout)
	}
}

func TestCatalogAddUsesStrictUserSchemaAndRejectsUnsafeProjectPolicy(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")

	code, _, stderr := runForTest("models", "add",
		"--scope", "user",
		"--user-root", userRoot,
		"--alias", "bad",
		"--model", "Deepseek4",
		"--adapter", "openai-compatible",
		"--dispatch-method", "chat-completions",
		"--destination", "local-lan",
		"--endpoint", "http://127.0.0.1:4000/v1",
		"--auth-env", "sk-secret-value",
		"--boundary", "private",
		"--retention", "none",
		"--training-use", "no",
		"--residency", "local",
		"--trust-provenance", "local",
		"--class", "large",
		"--approve-endpoint")
	if code == 0 || !strings.Contains(stderr, "auth-env must be an environment variable name") {
		t.Fatalf("expected secret auth-env rejection, code=%d stderr=%s", code, stderr)
	}

	code, _, stderr = runForTest("models", "add",
		"--scope", "project",
		"--project-root", projectRoot,
		"--alias", "local",
		"--model", "Deepseek4",
		"--endpoint", "http://127.0.0.1:4000/v1",
		"--auth-env", "LITELLM_API_KEY")
	if code == 0 || !strings.Contains(stderr, "project scope cannot store connection details") {
		t.Fatalf("expected project connection detail rejection, code=%d stderr=%s", code, stderr)
	}

	code, stdout, stderr := runForTest("models", "add",
		"--scope", "user",
		"--user-root", userRoot,
		"--alias", "local.deepseek",
		"--model", "Deepseek4",
		"--adapter", "openai-compatible",
		"--dispatch-method", "chat-completions",
		"--destination", "local-lan",
		"--endpoint", "http://127.0.0.1:4000/v1",
		"--auth-env", "LITELLM_API_KEY",
		"--boundary", "private",
		"--retention", "none",
		"--training-use", "no",
		"--residency", "local",
		"--trust-provenance", "local runbook",
		"--class", "large",
		"--approve-endpoint",
		"--json")
	if code != 0 {
		t.Fatalf("valid user add exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	catalog := loadUserCatalogForTest(t, userRoot)
	if len(catalog.Routes) != 1 || catalog.Routes[0].AuthEnv != "LITELLM_API_KEY" {
		t.Fatalf("unexpected stored catalog: %#v", catalog.Routes)
	}
	if catalog.Routes[0].Capability.DispatchProven {
		t.Fatalf("user-local add must not self-promote dispatch proof")
	}
}

func TestProjectPriorityIsUserLocalCanonicalAndQuickAddIsConservative(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectA := filepath.Join(root, "project-a")
	projectB := filepath.Join(root, "project-b")
	for _, project := range []string{projectA, projectB} {
		if err := os.MkdirAll(project, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(userRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	trustBefore := []byte(`{"schema_version":1,"projects":[]}`)
	if err := os.WriteFile(filepath.Join(userRoot, userTrustFile), trustBefore, 0o600); err != nil {
		t.Fatal(err)
	}

	code, stdout, stderr := runForTest("models", "priority", "--user-root", userRoot, "--project-root", projectA, "--mode", "self-hosted-first", "--json")
	if code != 0 {
		t.Fatalf("priority A code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	code, stdout, stderr = runForTest("models", "priority", "--user-root", userRoot, "--project-root", projectB, "--mode", "native-first", "--json")
	if code != 0 {
		t.Fatalf("priority B code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if _, err := os.Stat(filepath.Join(projectA, "kb-models.json")); !os.IsNotExist(err) {
		t.Fatalf("priority must not write repository state: %v", err)
	}
	preferences, err := loadProjectPriorities(userRoot)
	if err != nil {
		t.Fatal(err)
	}
	idA, _ := modelrouting.CanonicalProjectIdentity(projectA)
	idB, _ := modelrouting.CanonicalProjectIdentity(projectB)
	if got := preferences.priorityFor(idA); got != modelrouting.PreferenceSelfHostedFirst {
		t.Fatalf("project A priority=%q", got)
	}
	if got := preferences.priorityFor(idB); got != modelrouting.PreferenceNativeFirst {
		t.Fatalf("project B priority=%q", got)
	}

	code, stdout, stderr = runForTest("models", "priority", "--user-root", userRoot, "--project-root", projectA, "--clear", "--json")
	if code != 0 {
		t.Fatalf("priority clear code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	code, stdout, stderr = runForTest("models", "priority", "--user-root", userRoot, "--project-root", projectB, "--reset", "--json")
	if code != 0 {
		t.Fatalf("priority reset code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	preferences, err = loadProjectPriorities(userRoot)
	if err != nil {
		t.Fatal(err)
	}
	if got := preferences.priorityFor(idA); got != modelrouting.PreferenceAutomatic {
		t.Fatalf("cleared project A priority=%q", got)
	}
	if got := preferences.priorityFor(idB); got != modelrouting.PreferenceAutomatic {
		t.Fatalf("reset project B priority=%q", got)
	}
	if len(preferences.Projects) != 0 {
		t.Fatalf("clear/reset retained project entries: %#v", preferences.Projects)
	}

	code, stdout, stderr = runForTest("models", "add", "--scope", "user", "--user-root", userRoot, "--alias", "extra.quick", "--model", "model", "--endpoint", "https://models.example.invalid/v1", "--auth-env", "TEST_API_KEY", "--json")
	if code != 0 {
		t.Fatalf("quick add code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	route := loadUserCatalogForTest(t, userRoot).Routes[0]
	if route.Destination != "https://models.example.invalid" || route.ManagementOrigin != modelrouting.OriginExtra || route.Hosting != modelrouting.HostingUnknown ||
		route.Capability.Class != modelrouting.ClassUnknown || route.Capability.TaskFamily != "unknown" || route.Capability.Risk != modelrouting.RiskUnknown ||
		route.Capability.DispatchQualified || route.Capability.DispatchProven {
		t.Fatalf("quick add was not conservative: %#v", route)
	}
	trustAfter, err := os.ReadFile(filepath.Join(userRoot, userTrustFile))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(trustBefore, trustAfter) {
		t.Fatalf("priority/quick add mutated trust state: before=%s after=%s", trustBefore, trustAfter)
	}
	for _, project := range []string{projectA, projectB} {
		entries, err := os.ReadDir(project)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("configuration wrote repository state in %s: %#v", project, entries)
		}
	}

	absentUserRoot := filepath.Join(root, "absent-user")
	absentProjectRoot := filepath.Join(root, "absent-project")
	if err := os.Mkdir(absentProjectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = runForTest("models", "show", "--user-root", absentUserRoot, "--project-root", absentProjectRoot, "--json")
	if code != 0 {
		t.Fatalf("normal read code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	assertNotExists(t, absentUserRoot)
	entries, err := os.ReadDir(absentProjectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("normal read created repository state: %#v", entries)
	}
}

func TestPolicyProjectPreferencesContainNoConnectionDetails(t *testing.T) {
	projectRoot := t.TempDir()
	code, stdout, stderr := runForTest("models", "prefer",
		"--scope", "project",
		"--project-root", projectRoot,
		"--alias", "local.deepseek",
		"--project-id", "project-a",
		"--json")
	if code != 0 {
		t.Fatalf("prefer exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, "kb-models.json"))
	if err != nil {
		t.Fatalf("read project policy: %v", err)
	}
	text := string(data)
	for _, forbidden := range []string{"endpoint", "auth_env", "command", "http://", "LITELLM_API_KEY"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("project policy leaked %q: %s", forbidden, text)
		}
	}
	if !strings.Contains(text, "local.deepseek") || !strings.Contains(stdout, "project") {
		t.Fatalf("preference not recorded: stdout=%s file=%s", stdout, text)
	}
}

func TestDoctorReportsSeparateDimensionsWithoutMutation(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	addTrustedRouteForTest(t, userRoot, "local.deepseek", "Deepseek4", "http://127.0.0.1:4000/v1")
	before := statModTimeForTest(t, filepath.Join(userRoot, "models.json"))

	code, stdout, stderr := runForTest("models", "doctor", "--user-root", userRoot, "--json")
	if code != 0 {
		t.Fatalf("doctor exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	for _, dimension := range []string{"discovery", "configured", "selectable", "dispatch_proven", "auth", "control"} {
		if !strings.Contains(stdout, `"`+dimension+`"`) {
			t.Fatalf("doctor missing %s dimension: %s", dimension, stdout)
		}
	}
	if !strings.Contains(stdout, `"dispatch_proven":{"status":"unavailable"`) {
		t.Fatalf("doctor must not mark dispatch proven from listing alone: %s", stdout)
	}
	after := statModTimeForTest(t, filepath.Join(userRoot, "models.json"))
	if !before.Equal(after) {
		t.Fatalf("doctor mutated user catalog: before=%s after=%s", before, after)
	}
}

func TestCatalogOpenAICompatibleDiscoveryUsesTrustedRouteAndBoundedResponse(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	projectRoot := filepath.Join(root, "project")
	if err := os.MkdirAll(projectRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	os.Setenv("LITELLM_API_KEY", "test-token")
	t.Cleanup(func() { os.Unsetenv("LITELLM_API_KEY") })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Fatalf("missing auth header")
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"qwen-local"},{"id":"deepseek-local"}]}`))
	}))
	t.Cleanup(server.Close)
	addTrustedRouteForTest(t, userRoot, "local.qwen", "qwen-local", server.URL+"/v1")
	code, stdout, stderr := runForTest("models", "approve", "--user-root", userRoot, "--project-root", projectRoot, "--alias", "local.qwen", "--expires-in", "1h", "--json")
	if code != 0 {
		t.Fatalf("approve route for project exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}

	code, stdout, stderr = runForTest("models", "discover",
		"--run-root", runRoot,
		"--user-root", userRoot,
		"--project-root", projectRoot,
		"--current-model", "gpt-5.5",
		"--probe-openai-compatible",
		"--json")
	if code != 0 {
		t.Fatalf("discover openai-compatible exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if !strings.Contains(stdout, `"openai-compatible"`) || !strings.Contains(stdout, "deepseek-local") {
		t.Fatalf("expected configured OpenAI-compatible models in output: %s", stdout)
	}
	runCatalog := loadRunCatalogForTest(t, runRoot)
	if runCatalog.Fingerprint == "" {
		t.Fatalf("run catalog must be fingerprinted")
	}
	preservedPrivateMetadata := false
	for _, route := range runCatalog.Routes {
		if route.Capability.DispatchProven {
			t.Fatalf("discovery must not award dispatch proof: %#v", route)
		}
		if strings.HasPrefix(route.Alias, "local.qwen.") && route.Boundary == modelrouting.BoundaryPrivate && route.Retention == modelrouting.RetentionNone && route.TrainingUse == modelrouting.TrainingNo && route.Residency == "local" {
			if route.RouteID != "" || route.SourceRouteID == "" {
				t.Fatalf("probed child exposed or lost opaque source id: %#v", route)
			}
			preservedPrivateMetadata = true
		}
	}
	if !preservedPrivateMetadata {
		t.Fatalf("probed child route laundered parent trust metadata: %#v", runCatalog.Routes)
	}
}

func TestCatalogCalibrateIsAttendedOnlyNoCapabilityCredit(t *testing.T) {
	root := t.TempDir()
	userRoot := filepath.Join(root, "user")
	addTrustedRouteForTest(t, userRoot, "local.qwen", "qwen-local", "http://127.0.0.1:4000/v1")
	before := readFileForTest(t, filepath.Join(userRoot, "models.json"))

	code, stdout, stderr := runForTest("models", "calibrate", "--user-root", userRoot, "--alias", "local.qwen", "--json")
	if code != 0 {
		t.Fatalf("calibrate exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "attended") || !strings.Contains(stdout, "no inference dispatched") {
		t.Fatalf("calibrate should only describe attended check: %s", stdout)
	}
	after := readFileForTest(t, filepath.Join(userRoot, "models.json"))
	if before != after {
		t.Fatalf("calibrate mutated catalog")
	}
}

func runForTest(args ...string) (int, string, string) {
	previous := allowCustomUserRootForTests
	previousConfirmer := approvalConfirmer
	allowCustomUserRootForTests = true
	approvalConfirmer = func(approvalPrompt, io.Writer) error { return nil }
	defer func() {
		allowCustomUserRootForTests = previous
		approvalConfirmer = previousConfirmer
	}()
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func addTrustedRouteForTest(t *testing.T, userRoot, alias, model, endpoint string) {
	t.Helper()
	code, stdout, stderr := runForTest("models", "add",
		"--scope", "user",
		"--user-root", userRoot,
		"--alias", alias,
		"--model", model,
		"--adapter", "openai-compatible",
		"--dispatch-method", "chat-completions",
		"--destination", "local-lan",
		"--endpoint", endpoint,
		"--auth-env", "LITELLM_API_KEY",
		"--boundary", "private",
		"--retention", "none",
		"--training-use", "no",
		"--residency", "local",
		"--trust-provenance", "test approval",
		"--class", "large",
		"--approve-endpoint",
		"--json")
	if code != 0 {
		t.Fatalf("add trusted route exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
}

func loadUserCatalogForTest(t *testing.T, userRoot string) modelrouting.Catalog {
	t.Helper()
	return decodeCatalogForTest(t, filepath.Join(userRoot, "models.json"))
}

func loadRunCatalogForTest(t *testing.T, runRoot string) modelrouting.Catalog {
	t.Helper()
	return decodeCatalogForTest(t, filepath.Join(runRoot, "catalog.json"))
}

func decodeCatalogForTest(t *testing.T, path string) modelrouting.Catalog {
	t.Helper()
	var catalog modelrouting.Catalog
	if err := json.Unmarshal([]byte(readFileForTest(t, path)), &catalog); err != nil {
		t.Fatalf("decode catalog %s: %v", path, err)
	}
	return catalog
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("unexpected file exists: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func readFileForTest(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func statModTimeForTest(t *testing.T, path string) time.Time {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	return info.ModTime()
}

func TestPolicyJSONHasNoUnknownFields(t *testing.T) {
	projectRoot := t.TempDir()
	code, stdout, stderr := runForTest("models", "ignore-routing", "--scope", "project", "--project-root", projectRoot, "--json")
	if code != 0 {
		t.Fatalf("ignore-routing exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var policy projectModelsPolicy
	data := []byte(readFileForTest(t, filepath.Join(projectRoot, "kb-models.json")))
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatalf("project policy should be strict JSON-compatible: %v", err)
	}
	if !policy.IgnoreRouting {
		t.Fatalf("ignore-routing not recorded: %#v", policy)
	}
}

func TestCatalogCapturedCodexFixtureAndFingerprintRefresh(t *testing.T) {
	userRoot := t.TempDir()
	opts := discoverOptions{
		commonOptions:      commonOptions{userRoot: userRoot, projectRoot: "."},
		currentModel:       "gpt-5.5",
		adapterTimeout:     time.Second,
		sessionTimeout:     2 * time.Second,
		codexModelsFixture: filepath.Join("testdata", "codex-models.json"),
	}
	first, err := discoverCatalog(opts)
	if err != nil {
		t.Fatalf("first fixture discovery: %v", err)
	}
	second, err := discoverCatalog(opts)
	if err != nil {
		t.Fatalf("second fixture discovery: %v", err)
	}
	if first.Catalog.Fingerprint == "" || first.Catalog.Fingerprint != second.Catalog.Fingerprint {
		t.Fatalf("unchanged discovery fingerprint first=%q second=%q", first.Catalog.Fingerprint, second.Catalog.Fingerprint)
	}
	if !containsRouteModel(first.Catalog.Routes, "gpt-5.5") || !containsRouteModel(first.Catalog.Routes, "gpt-5.4-mini") {
		t.Fatalf("captured Codex models missing: %#v", first.Catalog.Routes)
	}

	code, stdout, stderr := runForTest("models", "add",
		"--scope", "user", "--user-root", userRoot,
		"--alias", "hosted.extra", "--model", "hosted-extra",
		"--adapter", "openai-compatible", "--dispatch-method", "chat-completions",
		"--destination", "hosted-extra", "--endpoint", "https://models.example.invalid/v1",
		"--boundary", "hosted", "--retention", "session", "--training-use", "no",
		"--residency", "declared", "--trust-provenance", "user-local declaration", "--class", "medium")
	if code != 0 {
		t.Fatalf("add fingerprint route exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	changed, err := discoverCatalog(opts)
	if err != nil {
		t.Fatalf("changed fixture discovery: %v", err)
	}
	if changed.Catalog.Fingerprint == first.Catalog.Fingerprint {
		t.Fatalf("configured-route change did not refresh fingerprint %q", changed.Catalog.Fingerprint)
	}
}

func TestTrustIsSeparateExplicitProjectBoundAndRevocable(t *testing.T) {
	userRoot := t.TempDir()
	addTrustedRouteForTest(t, userRoot, "local.secure", "secure-local", "http://127.0.0.1:4000/v1")
	projectID, err := modelrouting.CanonicalProjectIdentity(".")
	if err != nil {
		t.Fatalf("canonical project: %v", err)
	}
	trust, err := loadTrustFile(userRoot)
	if err != nil {
		t.Fatalf("load trust: %v", err)
	}
	project := findProjectTrust(trust, projectID)
	if len(project.RouteApprovals) != 1 || len(project.EndpointApprovals) != 1 || len(project.AuthBindings) != 1 {
		t.Fatalf("explicit approval not persisted separately: %#v", project)
	}

	code, stdout, stderr := runForTest("models", "deny", "--user-root", userRoot, "--project-root", ".", "--alias", "local.secure", "--json")
	if code != 0 {
		t.Fatalf("deny exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	policy, err := policyContextForProject(userRoot, ".")
	if err != nil {
		t.Fatalf("policy after deny: %v", err)
	}
	if routeProjectSelectable(loadUserCatalogForTest(t, userRoot).Routes[0], policy) {
		t.Fatalf("denied private route remained project-selectable")
	}

	code, stdout, stderr = runForTest("models", "approve", "--user-root", userRoot, "--project-root", ".", "--alias", "local.secure", "--expires-in", "1h", "--json")
	if code != 0 {
		t.Fatalf("reapprove exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	policy, err = policyContextForProject(userRoot, ".")
	if err != nil || !routeProjectSelectable(loadUserCatalogForTest(t, userRoot).Routes[0], policy) {
		t.Fatalf("explicitly approved route not selectable err=%v", err)
	}

	code, stdout, stderr = runForTest("models", "revoke", "--user-root", userRoot, "--project-root", ".", "--alias", "local.secure", "--json")
	if code != 0 {
		t.Fatalf("revoke exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	trust, err = loadTrustFile(userRoot)
	if err != nil {
		t.Fatalf("load revoked trust: %v", err)
	}
	project = findProjectTrust(trust, projectID)
	if len(project.RouteApprovals) != 0 || len(project.EndpointApprovals) != 0 || len(project.AuthBindings) != 0 {
		t.Fatalf("revoke left endpoint/auth trust active: %#v", project)
	}
}

func TestConcurrentApprovalPreservesInterveningDenial(t *testing.T) {
	userRoot := t.TempDir()
	addTrustedRouteForTest(t, userRoot, "local.one", "model-one", "http://127.0.0.1:4000/v1")
	addTrustedRouteForTest(t, userRoot, "local.two", "model-two", "http://127.0.0.1:4001/v1")
	previousRoot := allowCustomUserRootForTests
	previousConfirmer := approvalConfirmer
	allowCustomUserRootForTests = true
	started := make(chan struct{})
	release := make(chan struct{})
	approvalConfirmer = func(prompt approvalPrompt, _ io.Writer) error {
		if prompt.RouteAlias == "local.one" {
			close(started)
			<-release
		}
		return nil
	}
	defer func() {
		allowCustomUserRootForTests = previousRoot
		approvalConfirmer = previousConfirmer
	}()
	type commandResult struct {
		code           int
		stdout, stderr string
	}
	approved := make(chan commandResult, 1)
	go func() {
		var stdout, stderr bytes.Buffer
		code := run([]string{"models", "approve", "--user-root", userRoot, "--project-root", ".", "--alias", "local.one", "--expires-in", "1h", "--json"}, &stdout, &stderr)
		approved <- commandResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
	}()
	<-started
	var denyOut, denyErr bytes.Buffer
	denyCode := run([]string{"models", "deny", "--user-root", userRoot, "--project-root", ".", "--alias", "local.two", "--json"}, &denyOut, &denyErr)
	if denyCode != 0 {
		t.Fatalf("concurrent deny exit=%d stderr=%s stdout=%s", denyCode, denyErr.String(), denyOut.String())
	}
	close(release)
	result := <-approved
	if result.code != 0 {
		t.Fatalf("approval exit=%d stderr=%s stdout=%s", result.code, result.stderr, result.stdout)
	}
	policy, err := policyContextForProject(userRoot, ".")
	if err != nil {
		t.Fatal(err)
	}
	one, _, oneFingerprint, err := routeApprovalInputs(userRoot, ".", "local.one")
	if err != nil || one.Alias == "" {
		t.Fatal(err)
	}
	_, _, twoFingerprint, err := routeApprovalInputs(userRoot, ".", "local.two")
	if err != nil {
		t.Fatal(err)
	}
	foundApproval, foundDenial := false, false
	for _, approval := range policy.Trusted.RouteApprovals {
		foundApproval = foundApproval || approval.RouteFingerprint == oneFingerprint
	}
	for _, denial := range policy.Trusted.RouteDenials {
		foundDenial = foundDenial || denial.RouteFingerprint == twoFingerprint
	}
	if !foundApproval || !foundDenial {
		t.Fatalf("concurrent state lost: approvals=%#v denials=%#v", policy.Trusted.RouteApprovals, policy.Trusted.RouteDenials)
	}
}

func TestApprovedPrivateRouteSurvivesRedactedRunCatalogSelection(t *testing.T) {
	userRoot := t.TempDir()
	addTrustedRouteForTest(t, userRoot, "local.secure", "secure-local", "http://127.0.0.1:4000/v1")
	original := loadUserCatalogForTest(t, userRoot).Routes[0]
	sourceFingerprint, err := modelrouting.ComputeRouteFingerprint(original)
	if err != nil {
		t.Fatalf("source fingerprint: %v", err)
	}

	report, err := discoverCatalog(discoverOptions{
		commonOptions:      commonOptions{userRoot: userRoot, projectRoot: "."},
		currentModel:       "gpt-5.5",
		adapterTimeout:     time.Second,
		sessionTimeout:     2 * time.Second,
		codexModelsFixture: filepath.Join("testdata", "codex-models.json"),
	})
	if err != nil {
		t.Fatalf("discover redacted catalog: %v", err)
	}
	var redacted modelrouting.Route
	for _, route := range report.Catalog.Routes {
		if route.Alias == original.Alias {
			redacted = route
			break
		}
	}
	if redacted.Alias == "" || redacted.Endpoint != "" || redacted.AuthEnv != "" {
		t.Fatalf("private route was absent or leaked connection details: %#v", redacted)
	}
	if redacted.SourceRouteID != original.RouteID || redacted.RouteID != "" {
		t.Fatalf("redacted opaque source id=%q route_id=%q want source=%q", redacted.SourceRouteID, redacted.RouteID, original.RouteID)
	}
	if redacted.SourceRouteID == sourceFingerprint || strings.Contains(redacted.SourceRouteID, "127.0.0.1") || strings.Contains(redacted.SourceRouteID, "LITELLM") {
		t.Fatalf("redacted source reference is derived from private route fields: %q", redacted.SourceRouteID)
	}

	policy, err := policyContextForProject(userRoot, ".")
	if err != nil {
		t.Fatalf("load project policy: %v", err)
	}
	policy.TrustedCurrentModelID = "gpt-5.5"
	policy.TrustedCurrentSurface = "current"
	policy.TrustedCurrentRouteState, _ = modelrouting.ComputeRouteStateFingerprint(*report.Catalog.Current.Route)
	validated, rejections, err := modelrouting.ValidateCatalogForSelection(report.Catalog, policy, nil, time.Now(), modelrouting.CatalogSourceRun)
	if err != nil {
		t.Fatalf("validate redacted run catalog: %v", err)
	}
	for _, rejection := range rejections {
		if rejection.Alias == redacted.Alias {
			t.Fatalf("approved redacted route rejected: %#v", rejection)
		}
	}
	request := modelrouting.WorkRequest{
		PlannedTier: modelrouting.TierLarge,
		TaskFamily:  "code",
		Tools:       []string{"apply_patch"},
		ContextSize: 1,
		Risk:        modelrouting.RiskNormal,
		ProjectID:   policy.Project.ProjectID,
	}
	decision, err := modelrouting.SelectRoute(validated, request, policy, modelrouting.RunOverride{Mode: modelrouting.OverrideRequire, Alias: redacted.Alias}, modelrouting.AttemptLedger{}, time.Now())
	if err != nil || len(decision.Routes) == 0 || decision.Routes[0].Alias != redacted.Alias {
		t.Fatalf("approved redacted route not selectable: decision=%#v err=%v", decision, err)
	}

	tampered := report.Catalog
	for index := range tampered.Routes {
		if tampered.Routes[index].Alias == redacted.Alias {
			tampered.Routes[index].SourceRouteID = ""
			tampered.Routes[index].Boundary = modelrouting.BoundaryHosted
			tampered.Routes[index].Capability.Source = modelrouting.EvidenceAdapterPrior
		}
	}
	fingerprintAndSort(&tampered)
	_, tamperedRejections, err := modelrouting.ValidateCatalogForSelection(tampered, policy, nil, time.Now(), modelrouting.CatalogSourceRun)
	if err != nil {
		t.Fatalf("validate tampered run catalog envelope: %v", err)
	}
	rejected := false
	for _, rejection := range tamperedRejections {
		if rejection.Alias == redacted.Alias && strings.Contains(rejection.Reason, "missing trusted source identity") {
			rejected = true
		}
	}
	if !rejected {
		t.Fatalf("source-removal tamper was accepted: %#v", tamperedRejections)
	}
}

func TestRedactedRouteRejectsSubstitutedSourceID(t *testing.T) {
	userRoot := t.TempDir()
	addTrustedRouteForTest(t, userRoot, "local.one", "model-one", "http://127.0.0.1:4000/v1")
	addTrustedRouteForTest(t, userRoot, "local.two", "model-two", "http://127.0.0.1:4001/v1")
	report, err := discoverCatalog(discoverOptions{
		commonOptions:      commonOptions{userRoot: userRoot, projectRoot: "."},
		currentModel:       "gpt-5.5",
		adapterTimeout:     time.Second,
		sessionTimeout:     2 * time.Second,
		codexModelsFixture: filepath.Join("testdata", "codex-models.json"),
	})
	if err != nil {
		t.Fatal(err)
	}
	var secondRouteID string
	for _, route := range report.Catalog.Routes {
		if route.Alias == "local.two" {
			secondRouteID = route.SourceRouteID
		}
	}
	for index := range report.Catalog.Routes {
		if report.Catalog.Routes[index].Alias == "local.one" {
			report.Catalog.Routes[index].SourceRouteID = secondRouteID
		}
	}
	fingerprintAndSort(&report.Catalog)
	policy, err := policyContextForProject(userRoot, ".")
	if err != nil {
		t.Fatal(err)
	}
	_, rejections, err := modelrouting.ValidateCatalogForSelection(report.Catalog, policy, nil, time.Now(), modelrouting.CatalogSourceRun)
	if err != nil {
		t.Fatal(err)
	}
	for _, rejection := range rejections {
		if rejection.Alias == "local.one" {
			return
		}
	}
	t.Fatalf("source substitution was accepted: %#v", rejections)
}

func TestCatalogDoesNotMintTrustOrSendCatalogNamedSecret(t *testing.T) {
	userRoot := t.TempDir()
	projectRoot := t.TempDir()
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	t.Setenv("OPENAI_API_KEY", "must-not-leave-process")

	code, stdout, stderr := runForTest("models", "add",
		"--scope", "user", "--user-root", userRoot,
		"--alias", "unapproved.private", "--model", "private-model",
		"--adapter", "openai-compatible", "--dispatch-method", "chat-completions",
		"--destination", "private", "--endpoint", server.URL+"/v1", "--auth-env", "OPENAI_API_KEY",
		"--boundary", "private", "--retention", "none", "--training-use", "no",
		"--residency", "local", "--trust-provenance", "unapproved fixture", "--class", "medium")
	if code != 0 {
		t.Fatalf("store unapproved route exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	assertNotExists(t, filepath.Join(userRoot, userTrustFile))

	code, stdout, stderr = runForTest("models", "discover",
		"--run-root", runRoot, "--user-root", userRoot, "--project-root", projectRoot,
		"--current-model", "gpt-5.5", "--probe-openai-compatible", "--json")
	if code != 0 {
		t.Fatalf("unapproved discovery should degrade, exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if requests.Load() != 0 {
		t.Fatalf("unapproved catalog caused %d outbound requests", requests.Load())
	}
	if !strings.Contains(stdout, "no trusted configured route with models") {
		t.Fatalf("missing explicit trust failure: %s", stdout)
	}
	runCatalog := loadRunCatalogForTest(t, runRoot)
	for _, route := range runCatalog.Routes {
		if route.Endpoint != "" || route.AuthEnv != "" {
			t.Fatalf("run catalog leaked connection details: %#v", route)
		}
	}
}

func TestDiscoverRejectsRepositoryOrAncestorRunRootWithoutChangingPermissions(t *testing.T) {
	projectRoot := t.TempDir()
	parentRoot := filepath.Dir(projectRoot)
	projectBefore, err := os.Stat(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	parentBefore, err := os.Stat(parentRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, unsafeRoot := range []string{projectRoot, parentRoot, filepath.Join(projectRoot, ".kb", "runs")} {
		code, stdout, stderr := runForTest("models", "discover", "--run-root", unsafeRoot, "--project-root", projectRoot, "--current-model", "gpt-5.5", "--json")
		if code == 0 || !strings.Contains(stderr, "dedicated direct child") {
			t.Fatalf("unsafe run root %s exit=%d stderr=%s stdout=%s", unsafeRoot, code, stderr, stdout)
		}
	}
	projectAfter, err := os.Stat(projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	parentAfter, err := os.Stat(parentRoot)
	if err != nil {
		t.Fatal(err)
	}
	if projectBefore.Mode().Perm() != projectAfter.Mode().Perm() || parentBefore.Mode().Perm() != parentAfter.Mode().Perm() {
		t.Fatalf("unsafe run root changed permissions: project %#o->%#o parent %#o->%#o", projectBefore.Mode().Perm(), projectAfter.Mode().Perm(), parentBefore.Mode().Perm(), parentAfter.Mode().Perm())
	}
}

func TestPrepareRunRootRejectsSymlinkedProjectAncestors(t *testing.T) {
	for _, ancestor := range []string{".kb", "runs"} {
		projectRoot := t.TempDir()
		externalRoot := t.TempDir()
		if ancestor == "runs" {
			if err := os.Mkdir(filepath.Join(projectRoot, ".kb"), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		link := filepath.Join(projectRoot, ".kb")
		if ancestor == "runs" {
			link = filepath.Join(link, "runs")
		}
		if err := os.Symlink(externalRoot, link); err != nil {
			t.Logf("skip %s symlink case on this platform: %v", ancestor, err)
			continue
		}
		runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
		if _, err := prepareRunRoot(projectRoot, runRoot); err == nil {
			t.Fatalf("accepted symlinked %s ancestor", ancestor)
		}
		assertNotExists(t, filepath.Join(externalRoot, "run-1"))
	}
}

func TestPrepareRunRootRejectsExistingRunRootSymlink(t *testing.T) {
	projectRoot := t.TempDir()
	base := filepath.Join(projectRoot, ".kb", "runs")
	actual := filepath.Join(base, "actual")
	if err := os.MkdirAll(actual, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(base, "alias")
	if err := os.Symlink(actual, alias); err != nil {
		t.Skipf("directory aliases unavailable: %v", err)
	}
	if _, err := prepareRunRoot(projectRoot, alias); !errors.Is(err, modelrouting.ErrUnsafePath) {
		t.Fatalf("existing run-root symlink error=%v", err)
	}
}

func TestPrepareRunRootRejectsMixedAliasProjectAndRunPath(t *testing.T) {
	projectRoot := t.TempDir()
	aliasRoot := filepath.Join(t.TempDir(), "project-alias")
	if err := os.Symlink(projectRoot, aliasRoot); err != nil {
		t.Skipf("directory aliases unavailable: %v", err)
	}
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	if _, err := prepareRunRoot(aliasRoot, runRoot); err == nil {
		t.Fatalf("mixed alias project/run error=%v", err)
	}
	assertNotExists(t, filepath.Join(projectRoot, ".kb"))
}

func TestPrepareRunRootFailsClosedWhenProjectCannotBeCanonicalized(t *testing.T) {
	parent := t.TempDir()
	projectRoot := filepath.Join(parent, "missing-project")
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	if _, err := prepareRunRoot(projectRoot, runRoot); err == nil || !strings.Contains(err.Error(), "canonicalize project root") {
		t.Fatalf("nonexistent project canonicalization error=%v", err)
	}
	assertNotExists(t, projectRoot)
}

func TestSafeWindowsRunIDRejectsAmbiguousNames(t *testing.T) {
	for _, value := range []string{"run-1", "com0", "lpt10", "context"} {
		if !safeWindowsRunID(value) {
			t.Fatalf("rejected safe run id %q", value)
		}
	}
	for _, value := range []string{
		"run-1.", "run-1 ", "run:stream", "CON", "con.txt", "PRN.json", "AUX", "NUL.log",
		"COM1", "com9.txt", "LPT1", "lpt9.log", "CON .txt",
	} {
		if safeWindowsRunID(value) {
			t.Fatalf("accepted Windows-ambiguous run id %q", value)
		}
	}
}

func TestPreparedRunRootDetectsAncestorReplacement(t *testing.T) {
	projectRoot := t.TempDir()
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	prepared, err := prepareRunRoot(projectRoot, runRoot)
	if err != nil {
		t.Fatal(err)
	}
	oldBase := prepared.basePath + ".old"
	if err := os.Rename(prepared.basePath, oldBase); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(prepared.basePath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := prepared.revalidate(); err == nil {
		t.Fatal("revalidation accepted a replaced .kb/runs ancestor")
	}
}

func TestCanonicalizeProspectivePathResolvesExistingAncestorAlias(t *testing.T) {
	root := t.TempDir()
	aliasParent := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, aliasParent); err != nil {
		t.Skipf("directory aliases unavailable: %v", err)
	}
	got, err := canonicalizeProspectivePath(filepath.Join(aliasParent, "missing", "run-1"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "missing", "run-1")
	gotCanonical, err := filepath.Abs(got)
	if err != nil {
		t.Fatal(err)
	}
	wantCanonical, err := filepath.Abs(want)
	if err != nil {
		t.Fatal(err)
	}
	if !sameFilesystemPath(gotCanonical, wantCanonical) {
		t.Fatalf("prospective path=%q want=%q", gotCanonical, wantCanonical)
	}
}

func TestDeniedHostedRouteIsNeverProbed(t *testing.T) {
	userRoot := t.TempDir()
	code, stdout, stderr := runForTest("models", "add",
		"--scope", "user", "--user-root", userRoot,
		"--alias", "hosted.denied", "--model", "denied-model",
		"--adapter", "openai-compatible", "--dispatch-method", "chat-completions",
		"--destination", "hosted", "--endpoint", "https://models.example.invalid/v1",
		"--boundary", "hosted", "--retention", "session", "--training-use", "no",
		"--residency", "declared", "--trust-provenance", "test route", "--class", "medium")
	if code != 0 {
		t.Fatalf("add hosted route exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	code, stdout, stderr = runForTest("models", "deny", "--user-root", userRoot, "--project-root", ".", "--alias", "hosted.denied", "--json")
	if code != 0 {
		t.Fatalf("deny hosted route exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}

	previous := fetchOpenAICompatibleModels
	var calls atomic.Int32
	fetchOpenAICompatibleModels = func(context.Context, modelrouting.ValidatedEndpoint, modelrouting.Route, string, int64) ([]string, error) {
		calls.Add(1)
		return []string{"should-not-run"}, nil
	}
	defer func() { fetchOpenAICompatibleModels = previous }()

	report, err := discoverCatalog(discoverOptions{
		commonOptions:         commonOptions{userRoot: userRoot, projectRoot: "."},
		currentModel:          "gpt-5.5",
		probeOpenAICompatible: true,
		adapterTimeout:        time.Second,
		sessionTimeout:        2 * time.Second,
		codexModelsFixture:    filepath.Join("testdata", "codex-models.json"),
	})
	if err != nil {
		t.Fatalf("discover denied route: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("denied hosted route was probed %d times: %#v", calls.Load(), report.Adapters)
	}
}

func TestProjectPolicyErrorFailsConfiguredProbeClosed(t *testing.T) {
	userRoot := t.TempDir()
	projectRoot := t.TempDir()
	code, stdout, stderr := runForTest("models", "add",
		"--scope", "user", "--user-root", userRoot,
		"--alias", "hosted.policy-error", "--model", "policy-error-model",
		"--adapter", "openai-compatible", "--dispatch-method", "chat-completions",
		"--destination", "hosted", "--endpoint", "https://models.example.invalid/v1",
		"--boundary", "hosted", "--retention", "session", "--training-use", "no",
		"--residency", "declared", "--trust-provenance", "test route", "--class", "medium")
	if code != 0 {
		t.Fatalf("add hosted route exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, projectPolicyFile), []byte(`{"schema_version":1,"unexpected":true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	previous := fetchOpenAICompatibleModels
	var calls atomic.Int32
	fetchOpenAICompatibleModels = func(context.Context, modelrouting.ValidatedEndpoint, modelrouting.Route, string, int64) ([]string, error) {
		calls.Add(1)
		return []string{"should-not-run"}, nil
	}
	defer func() { fetchOpenAICompatibleModels = previous }()
	report, err := discoverCatalog(discoverOptions{
		commonOptions:         commonOptions{userRoot: userRoot, projectRoot: projectRoot},
		currentModel:          "gpt-5.5",
		probeOpenAICompatible: true,
		adapterTimeout:        time.Second,
		sessionTimeout:        2 * time.Second,
		codexModelsFixture:    filepath.Join("testdata", "codex-models.json"),
	})
	if err != nil {
		t.Fatalf("discover policy-error route: %v", err)
	}
	if calls.Load() != 0 {
		t.Fatalf("policy-error route was probed %d times", calls.Load())
	}
	foundPolicyError := false
	for _, adapter := range report.Adapters {
		if adapter.Name == "project-policy" && adapter.Status == "unavailable" {
			foundPolicyError = true
		}
	}
	if !foundPolicyError {
		t.Fatalf("policy failure was not reported: %#v", report.Adapters)
	}
}

func TestAttendedApprovalFailureDoesNotMutate(t *testing.T) {
	userRoot := t.TempDir()
	previousRoot := allowCustomUserRootForTests
	previousConfirmer := approvalConfirmer
	allowCustomUserRootForTests = true
	approvalConfirmer = func(prompt approvalPrompt, _ io.Writer) error {
		if prompt.ProjectPath == "" || prompt.ProjectID == "" || prompt.RouteFingerprint == "" || prompt.Origin == "" || prompt.AuthEnv != "LITELLM_API_KEY" || prompt.ExpiresAt.IsZero() {
			t.Fatalf("approval prompt was not fully bound: %#v", prompt)
		}
		return errors.New("user declined")
	}
	defer func() {
		allowCustomUserRootForTests = previousRoot
		approvalConfirmer = previousConfirmer
	}()
	var stdout, stderr bytes.Buffer
	code := run([]string{"models", "add",
		"--scope", "user", "--user-root", userRoot,
		"--alias", "local.declined", "--model", "declined-model",
		"--adapter", "openai-compatible", "--dispatch-method", "chat-completions",
		"--destination", "local", "--endpoint", "http://127.0.0.1:4000/v1",
		"--auth-env", "LITELLM_API_KEY", "--boundary", "private", "--retention", "none",
		"--training-use", "no", "--residency", "local", "--trust-provenance", "test",
		"--class", "medium", "--approve-endpoint"}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "user declined") {
		t.Fatalf("declined approval exit=%d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	assertNotExists(t, filepath.Join(userRoot, userCatalogFile))
	assertNotExists(t, filepath.Join(userRoot, userTrustFile))
}

func TestCredentialConsumingCommandsRejectCustomUserRootOutsideTestSeam(t *testing.T) {
	previous := allowCustomUserRootForTests
	allowCustomUserRootForTests = false
	defer func() { allowCustomUserRootForTests = previous }()
	var stdout, stderr bytes.Buffer
	code := run([]string{"models", "discover", "--run-root", t.TempDir(), "--user-root", t.TempDir(), "--probe-openai-compatible"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "fixed user-local trust root") {
		t.Fatalf("custom credential root exit=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"models", "add", "--scope", "user", "--user-root", t.TempDir(), "--alias", "x", "--model", "x", "--destination", "x", "--trust-provenance", "test"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "fixed user-local root") {
		t.Fatalf("custom user mutation root exit=%d stderr=%s", code, stderr.String())
	}
}

func TestOperatingSystemHomeIgnoresCallerHomeOverride(t *testing.T) {
	attackerRoot := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("USERPROFILE", attackerRoot)
	} else {
		t.Setenv("HOME", attackerRoot)
	}
	home, err := operatingSystemUserHome()
	if err != nil {
		t.Fatalf("operating-system home: %v", err)
	}
	if samePathForTest(home, attackerRoot) {
		t.Fatalf("OS user home followed caller-controlled override %q", attackerRoot)
	}
}

func TestDurableSchemasRejectMissingUnsupportedDuplicateOversizeAndSymlink(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		"missing.json":     `{"current":{},"routes":[]}`,
		"unsupported.json": `{"schema_version":2,"current":{},"routes":[]}`,
		"duplicate.json":   `{"schema_version":1,"schema_version":1,"current":{},"routes":[]}`,
	}
	for name, body := range cases {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadCatalogNamedForTest(root, name); err == nil {
			t.Fatalf("%s unexpectedly loaded", name)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "oversize.json"), make([]byte, maxCatalogBytes+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadCatalogNamedForTest(root, "oversize.json"); err == nil {
		t.Fatalf("oversized durable JSON unexpectedly loaded")
	}
	if err := os.Symlink(filepath.Join(root, "missing.json"), filepath.Join(root, userCatalogFile)); err == nil {
		if _, err := loadUserCatalog(root); err == nil {
			t.Fatalf("symlinked user catalog unexpectedly loaded")
		}
	}
}

func TestClearResetsMatchingScope(t *testing.T) {
	userRoot := t.TempDir()
	projectRoot := t.TempDir()
	for _, args := range [][]string{
		{"models", "prefer", "--scope", "user", "--user-root", userRoot, "--alias", "a"},
		{"models", "ignore-routing", "--scope", "user", "--user-root", userRoot},
		{"models", "prefer", "--scope", "project", "--project-root", projectRoot, "--alias", "b"},
		{"models", "ignore-routing", "--scope", "project", "--project-root", projectRoot},
	} {
		if code, stdout, stderr := runForTest(args...); code != 0 {
			t.Fatalf("setup %v exit=%d stderr=%s stdout=%s", args, code, stderr, stdout)
		}
	}
	for _, args := range [][]string{
		{"models", "clear", "--scope", "user", "--user-root", userRoot},
		{"models", "reset", "--scope", "project", "--project-root", projectRoot},
	} {
		if code, stdout, stderr := runForTest(args...); code != 0 {
			t.Fatalf("clear %v exit=%d stderr=%s stdout=%s", args, code, stderr, stdout)
		}
	}
	userPolicy, err := loadUserPreferences(userRoot)
	if err != nil || userPolicy.IgnoreRouting || len(userPolicy.PreferredAliases) != 0 {
		t.Fatalf("user clear failed policy=%#v err=%v", userPolicy, err)
	}
	projectPolicy, err := loadProjectPolicy(projectRoot)
	if err != nil || projectPolicy.IgnoreRouting || len(projectPolicy.PreferredAliases) != 0 {
		t.Fatalf("project reset failed policy=%#v err=%v", projectPolicy, err)
	}
}

func TestDoctorProbeDistinguishesReachabilityAndModelPresence(t *testing.T) {
	userRoot := t.TempDir()
	t.Setenv("LITELLM_API_KEY", "test-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":[{"id":"different-model"}]}`))
	}))
	addTrustedRouteForTest(t, userRoot, "local.expected", "expected-model", server.URL+"/v1")

	code, stdout, stderr := runForTest("models", "doctor", "--user-root", userRoot, "--project-root", ".", "--probe", "--json")
	if code != 0 {
		t.Fatalf("doctor model-absent exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if !strings.Contains(stdout, `"reachability":{"status":"available"`) || !strings.Contains(stdout, `"model_presence":{"status":"unavailable"`) {
		t.Fatalf("doctor did not distinguish reachability/model absence: %s", stdout)
	}
	server.Close()
	code, stdout, stderr = runForTest("models", "doctor", "--user-root", userRoot, "--project-root", ".", "--probe", "--json")
	if code != 0 || !strings.Contains(stdout, `"reachability":{"status":"unavailable"`) {
		t.Fatalf("doctor unreachable exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
}

func TestShowRedactsPrivateConnectionDetails(t *testing.T) {
	userRoot := t.TempDir()
	addTrustedRouteForTest(t, userRoot, "local.redacted", "redacted-model", "http://127.0.0.1:4000/v1")
	code, stdout, stderr := runForTest("models", "show", "--user-root", userRoot, "--json")
	if code != 0 {
		t.Fatalf("show exit=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	for _, forbidden := range []string{"127.0.0.1", "LITELLM_API_KEY", `"auth_env"`, `"endpoint"`} {
		if strings.Contains(stdout, forbidden) {
			t.Fatalf("show leaked %q: %s", forbidden, stdout)
		}
	}
}

func loadCatalogNamedForTest(root, name string) (modelrouting.Catalog, error) {
	var catalog modelrouting.Catalog
	if err := modelrouting.LoadStrictJSON(root, name, &catalog, maxCatalogBytes); err != nil {
		return modelrouting.Catalog{}, err
	}
	if err := modelrouting.ValidateCatalogStatic(catalog, modelrouting.CatalogSourceUser); err != nil {
		return modelrouting.Catalog{}, err
	}
	return catalog, nil
}

func containsRouteModel(routes []modelrouting.Route, model string) bool {
	for _, route := range routes {
		if route.DisplayModelID == model {
			return true
		}
	}
	return false
}

func samePathForTest(left, right string) bool {
	left, _ = filepath.Abs(filepath.Clean(left))
	right, _ = filepath.Abs(filepath.Clean(right))
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
