package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestModelsDiscoverDoesNotExposeFixtureFlagsInProduction(t *testing.T) {
	previous := allowCustomUserRootForTests
	allowCustomUserRootForTests = false
	defer func() { allowCustomUserRootForTests = previous }()

	var stdout, stderr bytes.Buffer
	code := run([]string{"models", "discover", "--codex-models-fixture", "forged.json"}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("fixture flag remained public: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestCodexProviderMetadataClassifiesWithoutModelNameInference(t *testing.T) {
	entries, err := parseCodexModels([]byte(`{"models":[
		{"slug":"opaque-a","description":"Small, fast, and cost-efficient model for simpler coding tasks.","context_window":272000},
		{"slug":"opaque-b","description":"Strong model for everyday coding.","context_window":272000},
		{"slug":"opaque-c","description":"Frontier model for complex coding, research, and real-world work.","context_window":272000},
		{"slug":"gpt-looks-large","description":"Unknown description.","context_window":272000}
	]}`))
	if err != nil {
		t.Fatal(err)
	}
	classes := map[string]modelrouting.CapabilityClass{}
	for _, entry := range entries {
		if entry.Capability != nil {
			classes[entry.ID] = entry.Capability.Class
		}
	}
	if classes["opaque-a"] != modelrouting.ClassSmall || classes["opaque-b"] != modelrouting.ClassMedium || classes["opaque-c"] != modelrouting.ClassLarge {
		t.Fatalf("provider metadata classes=%v", classes)
	}
	if _, ok := classes["gpt-looks-large"]; ok {
		t.Fatalf("model name earned capability without provider metadata: %v", classes)
	}
}

func TestCodexFixtureDiscoveryNeverPromotesAdapterPrior(t *testing.T) {
	previous := dispatchProcessTreeContainment
	previousNow := catalogNow
	dispatchProcessTreeContainment = func() error { return nil }
	catalogNow = func() time.Time { return time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC) }
	defer func() {
		dispatchProcessTreeContainment = previous
		catalogNow = previousNow
	}()

	opts := discoverOptions{
		commonOptions:      commonOptions{userRoot: t.TempDir(), projectRoot: "."},
		currentModel:       "app-current-only",
		adapterTimeout:     time.Second,
		sessionTimeout:     2 * time.Second,
		codexModelsFixture: filepath.Join("testdata", "codex-models.json"),
	}
	first, err := discoverCatalog(opts)
	if err != nil {
		t.Fatalf("first discovery: %v", err)
	}
	second, err := discoverCatalog(opts)
	if err != nil {
		t.Fatalf("second discovery: %v", err)
	}
	if first.Catalog.Fingerprint != second.Catalog.Fingerprint {
		t.Fatalf("fingerprint was not stable within the deterministic freshness window: %q != %q", first.Catalog.Fingerprint, second.Catalog.Fingerprint)
	}

	var codexRoute, currentRoute *modelrouting.Route
	for index := range first.Catalog.Routes {
		route := &first.Catalog.Routes[index]
		switch route.Alias {
		case "codex.gpt-5.5":
			codexRoute = route
		case "current":
			currentRoute = route
		}
	}
	if codexRoute == nil {
		t.Fatalf("missing codex CLI route: %#v", first.Catalog.Routes)
	}
	if currentRoute == nil {
		t.Fatalf("missing current fallback route: %#v", first.Catalog.Routes)
	}
	if got, want := codexRoute.Capability.Source, modelrouting.EvidenceDeclared; got != want {
		t.Fatalf("fixture-backed codex source=%q want declared-only", got)
	}
	if codexRoute.Capability.DispatchProven || hasReadiness(codexRoute.Readiness, modelrouting.ReadinessDispatchProven) {
		t.Fatalf("fixture-backed codex route gained adapter prior: %#v", codexRoute)
	}
	if codexRoute.AdapterRevision != "v1" {
		t.Fatalf("fixture-backed codex adapter revision=%q want v1", codexRoute.AdapterRevision)
	}
	miniRoute := routeByAlias(first.Catalog.Routes, "codex.gpt-5.4-mini")
	if miniRoute == nil {
		t.Fatalf("missing mini route: %#v", first.Catalog.Routes)
	}
	if miniRoute.Capability.Class == modelrouting.ClassLarge || miniRoute.Capability.DispatchProven {
		t.Fatalf("listing/name-only mini route was promoted: %#v", miniRoute)
	}
	if currentRoute.Capability.DispatchProven || currentRoute.Capability.Source != modelrouting.EvidenceDeclared {
		t.Fatalf("current fallback self-promoted: %#v", currentRoute)
	}
}

func TestCodexDiscoveryPromotesOnlyExactWindowsExecutableIdentity(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "project")
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := writeDiscoveryCodexExecutable(t, root, "codex-exact", true)
	previousResolver := dispatchExecutableResolver
	previousContainment := dispatchProcessTreeContainment
	previousNow := catalogNow
	dispatchExecutableResolver = func() (string, error) { return fake.path, nil }
	dispatchProcessTreeContainment = func() error { return nil }
	catalogNow = func() time.Time { return time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC) }
	t.Cleanup(func() {
		dispatchExecutableResolver = previousResolver
		dispatchProcessTreeContainment = previousContainment
		catalogNow = previousNow
	})

	report, err := discoverCatalog(discoverOptions{
		commonOptions:  commonOptions{userRoot: filepath.Join(root, "user"), projectRoot: projectRoot},
		runRoot:        runRoot,
		currentModel:   "app-current-only",
		adapterTimeout: 15 * time.Second,
		sessionTimeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	route := routeByAlias(report.Catalog.Routes, "codex.gpt-5.5")
	if route == nil {
		t.Fatalf("missing codex route: %#v", report.Catalog.Routes)
	}
	if runtime.GOOS != "windows" {
		if route.Capability.DispatchProven || route.Capability.Source != modelrouting.EvidenceDeclared {
			t.Fatalf("non-Windows codex route gained adapter prior: %#v", route)
		}
		return
	}
	if route.Capability.Source != modelrouting.EvidenceAdapterPrior || !route.Capability.DispatchQualified || route.Capability.DispatchProven || hasReadiness(route.Readiness, modelrouting.ReadinessDispatchProven) {
		t.Fatalf("exact executable did not earn qualified-only adapter prior: %#v", route)
	}
	miniRoute := routeByAlias(report.Catalog.Routes, "codex.gpt-5.4-mini")
	if miniRoute == nil {
		t.Fatalf("missing mini route: %#v", report.Catalog.Routes)
	}
	if miniRoute.Capability.Class == modelrouting.ClassLarge || miniRoute.Capability.DispatchQualified || miniRoute.Capability.DispatchProven {
		t.Fatalf("listing/name-only mini route was promoted by exact executable: %#v", miniRoute)
	}
	if !strings.HasPrefix(route.AdapterRevision, "codex-cli-v1:") || route.AdapterRevision == "v1" {
		t.Fatalf("adapter revision did not bind executable identity: %q", route.AdapterRevision)
	}
	if !strings.Contains(route.TrustProvenance, "codex-cli 9.8.7") {
		t.Fatalf("trust provenance does not bind CLI version: %q", route.TrustProvenance)
	}
	invocations, err := os.ReadFile(fake.invocations)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(invocations), "debug models") {
		t.Fatalf("model enumeration did not use exact executable: %s", invocations)
	}
}

func TestCodexUnixDiscoveryCanonicalizesLauncherSymlinkButStaysUnproven(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix launcher symlink coverage")
	}
	root := t.TempDir()
	projectRoot := filepath.Join(root, "project")
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := writeDiscoveryCodexExecutable(t, root, "codex-target", true)
	launcher := filepath.Join(root, "bin", "codex-link")
	if err := os.Symlink(fake.path, launcher); err != nil {
		t.Fatal(err)
	}
	previousResolver := dispatchExecutableResolver
	dispatchExecutableResolver = func() (string, error) { return launcher, nil }
	t.Cleanup(func() { dispatchExecutableResolver = previousResolver })

	report, err := discoverCatalog(discoverOptions{
		commonOptions:  commonOptions{userRoot: filepath.Join(root, "user"), projectRoot: projectRoot},
		runRoot:        runRoot,
		currentModel:   "current-only",
		adapterTimeout: 15 * time.Second,
		sessionTimeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	route := routeByAlias(report.Catalog.Routes, "codex.gpt-5.5")
	if route == nil {
		t.Fatalf("symlinked Unix launcher did not enumerate models: %#v", report.Adapters)
	}
	if route.Capability.DispatchProven || route.Capability.Source != modelrouting.EvidenceDeclared {
		t.Fatalf("Unix route gained dispatch proof: %#v", route)
	}
}

func TestCodexDiscoveryLeavesModelsDiscoveredOnlyOnContractFailure(t *testing.T) {
	root := t.TempDir()
	projectRoot := filepath.Join(root, "project")
	runRoot := filepath.Join(projectRoot, ".kb", "runs", "run-1")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	fake := writeDiscoveryCodexExecutable(t, root, "codex-bad-contract", false)
	previousResolver := dispatchExecutableResolver
	previousContainment := dispatchProcessTreeContainment
	dispatchExecutableResolver = func() (string, error) { return fake.path, nil }
	dispatchProcessTreeContainment = func() error { return nil }
	t.Cleanup(func() {
		dispatchExecutableResolver = previousResolver
		dispatchProcessTreeContainment = previousContainment
	})

	report, err := discoverCatalog(discoverOptions{
		commonOptions:  commonOptions{userRoot: filepath.Join(root, "user"), projectRoot: projectRoot},
		runRoot:        runRoot,
		currentModel:   "app-current-only",
		adapterTimeout: 15 * time.Second,
		sessionTimeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	route := routeByAlias(report.Catalog.Routes, "codex.gpt-5.5")
	if route == nil {
		t.Fatalf("missing codex route: %#v", report.Catalog.Routes)
	}
	if route.Capability.Source != modelrouting.EvidenceDeclared || route.Capability.DispatchProven || hasReadiness(route.Readiness, modelrouting.ReadinessDispatchProven) {
		t.Fatalf("contract-failed route was promoted: %#v", route)
	}
}

type discoveryCodexExecutable struct {
	path        string
	invocations string
}

func writeDiscoveryCodexExecutable(t *testing.T, root, name string, completeContract bool) discoveryCodexExecutable {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "bin", name+scriptExt())
	invocations := filepath.Join(root, name+"-invocations.txt")
	contract := "--model --sandbox -C --add-dir -c --output-schema --json"
	if !completeContract {
		contract = "--model --sandbox -C --add-dir -c --json"
	}
	var lines []string
	if runtime.GOOS == "windows" {
		lines = []string{
			"@echo off",
			"echo %*>>" + quoteCmd(invocations),
			"if \"%1\"==\"--version\" (",
			"  echo codex-cli 9.8.7",
			"  exit /b 0",
			")",
			"if \"%1\"==\"exec\" if \"%2\"==\"--help\" (",
			"  echo " + contract,
			"  exit /b 0",
			")",
			"if \"%1\"==\"debug\" if \"%2\"==\"models\" if \"%3\"==\"--help\" (",
			"  echo models help",
			"  exit /b 0",
			")",
			"if \"%1\"==\"debug\" if \"%2\"==\"models\" (",
			"  echo {^\"models^\": [{^\"slug^\": ^\"gpt-5.5^\", ^\"capability^\": {^\"class^\": ^\"large^\", ^\"task_family^\": ^\"code^\", ^\"tools^\": [^\"codex-harness^\"], ^\"context_size^\": 8192, ^\"risk^\": ^\"broad^\"}}, {^\"slug^\": ^\"gpt-5.4-mini^\"}]}",
			"  exit /b 0",
			")",
			"exit /b 9",
		}
	} else {
		lines = []string{
			"#!/bin/sh",
			"printf '%s\\n' \"$*\" >> " + quoteShell(invocations),
			"if [ \"$1\" = \"--version\" ]; then printf '%s\\n' 'codex-cli 9.8.7'; exit 0; fi",
			"if [ \"$1\" = \"exec\" ] && [ \"$2\" = \"--help\" ]; then printf '%s\\n' '" + contract + "'; exit 0; fi",
			"if [ \"$1\" = \"debug\" ] && [ \"$2\" = \"models\" ] && [ \"$3\" = \"--help\" ]; then printf '%s\\n' 'models help'; exit 0; fi",
			"if [ \"$1\" = \"debug\" ] && [ \"$2\" = \"models\" ]; then printf '%s\\n' '{\"models\": [{\"slug\": \"gpt-5.5\", \"capability\": {\"class\": \"large\", \"task_family\": \"code\", \"tools\": [\"codex-harness\"], \"context_size\": 8192, \"risk\": \"broad\"}}, {\"slug\": \"gpt-5.4-mini\"}]}'; exit 0; fi",
			"exit 9",
		}
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o755); err != nil {
		t.Fatal(err)
	}
	return discoveryCodexExecutable{path: path, invocations: invocations}
}

func routeByAlias(routes []modelrouting.Route, alias string) *modelrouting.Route {
	for index := range routes {
		if routes[index].Alias == alias {
			return &routes[index]
		}
	}
	return nil
}

func TestCodexDiscoveryLeavesCLIModelsInformativeWithoutContainmentProof(t *testing.T) {
	previous := dispatchProcessTreeContainment
	dispatchProcessTreeContainment = func() error { return errors.New("no containment proof") }
	defer func() { dispatchProcessTreeContainment = previous }()

	report, err := discoverCatalog(discoverOptions{
		commonOptions:      commonOptions{userRoot: t.TempDir(), projectRoot: "."},
		currentModel:       "app-current-only",
		adapterTimeout:     time.Second,
		sessionTimeout:     2 * time.Second,
		codexModelsFixture: filepath.Join("testdata", "codex-models.json"),
	})
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	for _, route := range report.Catalog.Routes {
		if route.Adapter != "codex" || route.Alias == "current" {
			continue
		}
		if len(route.Readiness) != 1 || route.Readiness[0] != modelrouting.ReadinessDiscovered {
			t.Fatalf("containment-free codex route was promoted: %#v", route)
		}
		if route.Capability.Source != modelrouting.EvidenceDeclared || route.Capability.DispatchProven || !route.Capability.ExpiresAt.IsZero() {
			t.Fatalf("containment-free codex evidence was promoted: %#v", route.Capability)
		}
	}
}
