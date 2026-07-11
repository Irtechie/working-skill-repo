package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestModelsSelectUsesValidatedRunCatalogAndRunOnlyOverride(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "select-cli")
	route := fixture.route("codex.medium", "medium-model", modelrouting.ClassMedium)
	route.Capability.Source = modelrouting.EvidenceAdapterPrior
	route.Capability.DispatchQualified = true
	route.Capability.ExpiresAt = time.Now().Add(time.Hour)
	fixture.installCatalog(route)
	prepared, err := prepareRunRoot(fixture.projectRoot, fixture.runRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := saveDispatchTrustedState(fixture.userRoot, prepared, loadRunCatalogForTest(t, fixture.runRoot)); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runForTest("models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--override", "use", "--alias", route.Alias, "--json")
	if code != 0 {
		t.Fatalf("select failed code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var out selectOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != modelrouting.SelectionRouted || len(out.Aliases) == 0 || out.Aliases[0] != route.Alias {
		t.Fatalf("unexpected selection: %#v", out)
	}
}

func TestModelsSelectReportsExplicitAttemptAndPlannedCorrectionTiers(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "select-attempt-cli")
	route := fixture.route("codex.small", "small-model", modelrouting.ClassSmall)
	route.Capability.Source = modelrouting.EvidenceAdapterPrior
	route.Capability.DispatchQualified = true
	route.Capability.ExpiresAt = time.Now().Add(time.Hour)
	fixture.installCatalog(route)
	prepared, err := prepareRunRoot(fixture.projectRoot, fixture.runRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := saveDispatchTrustedState(fixture.userRoot, prepared, loadRunCatalogForTest(t, fixture.runRoot)); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runForTest("models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--attempt-tier", "small", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--prefer", "native", "--json")
	if code != 0 {
		t.Fatalf("select failed code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var out selectOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != modelrouting.SelectionRouted || len(out.Aliases) == 0 || out.Aliases[0] != route.Alias {
		t.Fatalf("unexpected selection: %#v", out)
	}
	if out.PlannedTier != modelrouting.TierMedium || out.AttemptTier != modelrouting.TierSmall {
		t.Fatalf("selection lost correction metadata: %#v", out)
	}
	if out.Preference != modelrouting.PreferenceNativeFirst {
		t.Fatalf("selection lost run preference: %#v", out)
	}
}

func TestModelsSelectLoadsSavedProjectPriorityUnlessRunPreferenceOverrides(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "select-saved-priority")
	route := fixture.route("codex.medium", "medium-model", modelrouting.ClassMedium)
	route.Readiness = append(route.Readiness, modelrouting.ReadinessDispatchProven)
	route.Capability.Source = modelrouting.EvidenceKBReceipt
	route.Capability.DispatchQualified = true
	route.Capability.DispatchProven = true
	route.Capability.ExpiresAt = time.Now().Add(time.Hour)
	fixture.installCatalog(route)
	prepared, err := prepareRunRoot(fixture.projectRoot, fixture.runRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := saveDispatchTrustedState(fixture.userRoot, prepared, loadRunCatalogForTest(t, fixture.runRoot)); err != nil {
		t.Fatal(err)
	}
	projectID, err := modelrouting.CanonicalProjectIdentity(fixture.projectRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := storeProjectPriority(fixture.userRoot, projectID, modelrouting.PreferenceNativeFirst, false); err != nil {
		t.Fatal(err)
	}
	base := []string{"models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--json"}
	code, stdout, stderr := runForTest(base...)
	if code != 0 {
		t.Fatalf("saved select code=%d stderr=%s", code, stderr)
	}
	var out selectOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Preference != modelrouting.PreferenceNativeFirst {
		t.Fatalf("saved preference not used: %#v", out)
	}
	code, stdout, stderr = runForTest(append(base[:len(base)-1], "--prefer", "self-hosted", "--json")...)
	if code != 0 {
		t.Fatalf("override select code=%d stderr=%s", code, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Preference != modelrouting.PreferenceSelfHostedFirst {
		t.Fatalf("run preference did not override saved: %#v", out)
	}
	code, stdout, stderr = runForTest(append(base[:len(base)-1], "--override", "use", "--alias", route.Alias, "--json")...)
	if code != 0 {
		t.Fatalf("use override select code=%d stderr=%s", code, stderr)
	}
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Preference != modelrouting.PreferenceAutomatic {
		t.Fatalf("saved preference leaked into run override: %#v", out)
	}
}

func TestModelsSelectIgnoreBypassesCorruptSavedPriority(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "select-ignore-corrupt-priority")
	route := fixture.route("codex.medium", "medium-model", modelrouting.ClassMedium)
	fixture.installCatalog(route)
	prepared, err := prepareRunRoot(fixture.projectRoot, fixture.runRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := saveDispatchTrustedState(fixture.userRoot, prepared, loadRunCatalogForTest(t, fixture.runRoot)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fixture.userRoot, userProjectPrioritiesFile), []byte(`{"schema_version":99}`), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runForTest("models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--override", "ignore", "--json")
	if code != 0 {
		t.Fatalf("ignore override was blocked by corrupt saved priority code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var out selectOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != modelrouting.SelectionIgnored || out.Preference != modelrouting.PreferenceAutomatic {
		t.Fatalf("unexpected ignored selection: %#v", out)
	}
}

func TestDispatchSchemaNameIsSliceScoped(t *testing.T) {
	a := "worker-output-schema-" + sha256Text("slice-a")[:16] + ".json"
	b := "worker-output-schema-" + sha256Text("slice-b")[:16] + ".json"
	if a == b || a == "worker-output-schema.json" || b == "worker-output-schema.json" {
		t.Fatalf("schema names collide: %s %s", a, b)
	}
}
