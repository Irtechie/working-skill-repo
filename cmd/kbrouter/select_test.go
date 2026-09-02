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
	code, stdout, stderr := runForTest("models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--execution-owner", "delegated", "--owner-reason", "bounded-delegation", "--tier-reason", "fixture medium capability", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--override", "use", "--alias", route.Alias, "--json")
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
	if out.ExecutionOwner != modelrouting.ExecutionOwnerDelegated || out.OwnerReason != "bounded-delegation" ||
		out.TierReason != "fixture medium capability" || out.Alias != route.Alias || len(out.Aliases) != 1 {
		t.Fatalf("selection lost singular ownership receipt: %#v", out)
	}
}

func TestModelsSelectRequiresExplicitOwnershipDecision(t *testing.T) {
	code, _, stderr := runForTest(
		"models", "select",
		"--run-root", "fixture", "--run-id", "fixture", "--tier", "medium",
		"--task-family", "code", "--tool", "apply_patch", "--context-size", "4096", "--risk", "normal",
		"--json",
	)
	if code != 2 || !strings.Contains(stderr, "ownership") {
		t.Fatalf("missing ownership was not rejected: code=%d stderr=%s", code, stderr)
	}
}

func TestModelsSelectRejectsVagueCurrentOwnerReason(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "select-invalid-current-reason")
	fixture.installCatalog()
	prepared, err := prepareRunRoot(fixture.projectRoot, fixture.runRoot)
	if err != nil {
		t.Fatal(err)
	}
	if err := saveDispatchTrustedState(fixture.userRoot, prepared, loadRunCatalogForTest(t, fixture.runRoot)); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr := runForTest("models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--execution-owner", "current", "--owner-reason", "this is complex", "--tier-reason", "fixture medium capability", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--json")
	if code != 1 || stderr != "" {
		t.Fatalf("vague current reason was not returned as invalid work: code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var out selectOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != modelrouting.SelectionUnavailable || out.ErrorClass != "invalid-work-request" {
		t.Fatalf("unexpected invalid current reason output: %#v", out)
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
	base := []string{"models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--execution-owner", "delegated", "--owner-reason", "bounded-delegation", "--tier-reason", "fixture medium capability", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--json"}
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
	code, stdout, stderr := runForTest("models", "select", "--user-root", fixture.userRoot, "--project-root", fixture.projectRoot, "--run-root", fixture.runRoot, "--run-id", filepath.Base(fixture.runRoot), "--tier", "medium", "--execution-owner", "current", "--owner-reason", "user-required", "--tier-reason", "fixture current capability", "--task-family", "code", "--tool", "codex-harness", "--context-size", "4096", "--risk", "normal", "--override", "ignore", "--json")
	if code != 0 || strings.Contains(stderr, "project priority") {
		t.Fatalf("ignore override did not bypass corrupt priority before capability validation code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	var out selectOutput
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatal(err)
	}
	if out.Status != modelrouting.SelectionUnavailable || out.Preference != modelrouting.PreferenceAutomatic ||
		out.ErrorClass != "" {
		t.Fatalf("unqualified current route was accepted: %#v", out)
	}
}

func TestDispatchSchemaNameIsSliceScoped(t *testing.T) {
	a := "worker-output-schema-" + sha256Text("slice-a")[:16] + ".json"
	b := "worker-output-schema-" + sha256Text("slice-b")[:16] + ".json"
	if a == b || a == "worker-output-schema.json" || b == "worker-output-schema.json" {
		t.Fatalf("schema names collide: %s %s", a, b)
	}
}
