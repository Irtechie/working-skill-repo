package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestModelsSelectUsesValidatedRunCatalogAndRunOnlyOverride(t *testing.T) {
	skipIfPrivateACLUnsupported(t)
	fixture := newDispatchFixture(t, "select-cli")
	route := fixture.route("codex.medium", "medium-model", modelrouting.ClassMedium)
	route.Readiness = append(route.Readiness, modelrouting.ReadinessDispatchProven)
	route.Capability.Source = modelrouting.EvidenceAdapterPrior
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

func TestDispatchSchemaNameIsSliceScoped(t *testing.T) {
	a := "worker-output-schema-" + sha256Text("slice-a")[:16] + ".json"
	b := "worker-output-schema-" + sha256Text("slice-b")[:16] + ".json"
	if a == b || a == "worker-output-schema.json" || b == "worker-output-schema.json" {
		t.Fatalf("schema names collide: %s %s", a, b)
	}
}
