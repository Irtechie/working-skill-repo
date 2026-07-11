package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestDispatchAttestationRequiresExactHostBuiltReceiptBytes(t *testing.T) {
	receipt := modelrouting.RoutingReceipt{
		RouteEvidence: modelrouting.RouteDispatchEvidence{RunID: "run-1", SessionID: "session-1"},
		WorkProof:     modelrouting.WorkProof{Command: "kbrouter dispatch", Result: modelrouting.ProofUnknown},
	}
	exact, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	exact = append(exact, '\n')
	if err := validateHostBuiltReceiptBytes(receipt, exact); err != nil {
		t.Fatal(err)
	}
	if err := validateHostBuiltReceiptBytes(receipt, []byte(`{"route_evidence":{"session_id":"forged"}}`)); err == nil {
		t.Fatal("replacement receipt bytes were accepted for dispatcher attestation")
	}
}

func TestTrustedCodexExecutableRejectsSymlinkedProjectContainment(t *testing.T) {
	root := t.TempDir()
	physicalProject := filepath.Join(root, "physical-project")
	runRoot := filepath.Join(physicalProject, ".kb", "runs", "run-1")
	binRoot := filepath.Join(physicalProject, "bin")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(binRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(binRoot, "codex-contained"+scriptExt())
	if err := os.WriteFile(executable, []byte("contained"), 0o755); err != nil {
		t.Fatal(err)
	}
	projectAlias := filepath.Join(root, "project-alias")
	if err := os.Symlink(physicalProject, projectAlias); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}
	previous := dispatchExecutableResolver
	dispatchExecutableResolver = func() (string, error) { return executable, nil }
	t.Cleanup(func() { dispatchExecutableResolver = previous })
	_, err := resolveTrustedCodexExecutable(projectAlias, filepath.Join(projectAlias, ".kb", "runs", "run-1"))
	if err == nil || !strings.Contains(err.Error(), "cannot live under project or run root") {
		t.Fatalf("symlinked project containment bypassed: %v", err)
	}
}

func TestC2DispatchRevalidatesAdapterPriorExecutableRevisionBeforeLaunch(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "c2-exec-revision")
	route := fixture.route("codex.large", "large-model", modelrouting.ClassLarge)
	route.AdapterRevision = "codex-cli-v1:0000000000000000000000000000000000000000000000000000000000000000:stale"
	route.Readiness = []modelrouting.Readiness{
		modelrouting.ReadinessDiscovered,
		modelrouting.ReadinessConfigured,
		modelrouting.ReadinessSelectable,
	}
	route.Capability.Source = modelrouting.EvidenceAdapterPrior
	route.Capability.DispatchQualified = true
	route.Capability.DispatchProven = false
	route.Capability.ExpiresAt = time.Now().Add(time.Hour)
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fake := fixture.fakeCodex(fakeCodexSpec{})
	fixture.withFakeCodex(fake)

	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "trusted codex executable revision changed") {
		t.Fatalf("stale adapter prior launched or failed unclearly, code=%d stdout=%s stderr=%s", result.code, result.stdout, result.stderr)
	}
	if data, err := os.ReadFile(fake.argsPath); (err != nil && !os.IsNotExist(err)) || strings.Contains(string(data), "exec") {
		t.Fatalf("worker exec started despite stale adapter prior revision: data=%q err=%v", data, err)
	}
}

func TestC2DispatchRejectsVisibleCurrentRouteBeforeLaunch(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "c2-current-visible")
	route := currentRoute("app-only-current-model")
	fixture.installCatalog(route)
	fixture.trustRoutes(route)
	fake := fixture.fakeCodex(fakeCodexSpec{})
	fixture.withFakeCodex(fake)

	result := fixture.run("--route-alias", route.Alias)
	if result.code == 0 || !strings.Contains(result.stderr, "current-model route is visible-only") {
		t.Fatalf("visible current route launched or failed unclearly, code=%d stdout=%s stderr=%s", result.code, result.stdout, result.stderr)
	}
	if _, err := os.Stat(fake.argsPath); !os.IsNotExist(err) {
		t.Fatalf("worker started for visible current route: %v", err)
	}
}
