package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProofSelectionReusesPassingSuperset(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := proofSelectionFixtureRegistry(t, root)
	now := time.Now().UTC()
	full := registry.Checks[0]
	receipt := captureSelectionReceipt(t, root, registry, full, now)

	plan := selectProofGovernorPlan(root, registry, []string{"rust", "cli"}, []proofGovernorReceipt{receipt}, now.Add(time.Second))
	if plan.Decision != proofGovernorReuse || len(plan.RunChecks) != 0 {
		t.Fatalf("passing superset should satisfy subset without a run: %#v", plan)
	}
	if strings.Join(plan.Reused, ",") != "cli,rust" {
		t.Fatalf("unexpected reused checks: %#v", plan.Reused)
	}
}

func TestProofSelectionRunsOnlyInvalidatedChecksAndExplainsPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := proofSelectionFixtureRegistry(t, root)
	now := time.Now().UTC()
	receipts := make([]proofGovernorReceipt, 0, len(registry.Checks))
	for _, spec := range registry.Checks {
		receipts = append(receipts, captureSelectionReceipt(t, root, registry, spec, now))
	}

	writeProofGovernorFixture(t, root, "src/rust.go", "package demo\n// changed\n")
	plan := selectProofGovernorPlan(root, registry, []string{"rust", "browser"}, receipts, now.Add(time.Second))
	if plan.Decision != proofGovernorRun {
		t.Fatalf("changed Rust input must require a run: %#v", plan)
	}
	if strings.Join(plan.RunChecks, ",") != "rust" || strings.Join(plan.Reused, ",") != "browser" {
		t.Fatalf("expected rust run and browser reuse: %#v", plan)
	}
	if !containsProofGovernorReason(plan.Reasons, "rust:relevant-input-changed:src/rust.go") {
		t.Fatalf("missing exact invalidation path: %#v", plan.Reasons)
	}
}

func TestImpactInvalidationFansOutAcrossSharedDependency(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := proofSelectionFixtureRegistry(t, root)
	now := time.Now().UTC()
	receipts := make([]proofGovernorReceipt, 0, len(registry.Checks))
	for _, spec := range registry.Checks {
		receipts = append(receipts, captureSelectionReceipt(t, root, registry, spec, now))
	}

	writeProofGovernorFixture(t, root, "src/shared.go", "package demo\n// changed shared dependency\n")
	plan := selectProofGovernorPlan(root, registry, []string{"rust", "cli", "browser"}, receipts, now.Add(time.Second))
	if plan.Decision != proofGovernorRun || strings.Join(plan.RunChecks, ",") != "cli,rust" || strings.Join(plan.Reused, ",") != "browser" {
		t.Fatalf("shared dependency did not invalidate declared dependents: %#v", plan)
	}
	if !containsProofGovernorReason(plan.Reasons, "rust:relevant-input-changed:src/shared.go") ||
		!containsProofGovernorReason(plan.Reasons, "cli:relevant-input-changed:src/shared.go") {
		t.Fatalf("fan-out reasons missing: %#v", plan.Reasons)
	}
}

func TestCoverageSubsumptionCollapsesCompositeAndDuplicates(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := proofSelectionFixtureRegistry(t, root)

	plan := selectProofGovernorPlan(root, registry, []string{"full", "rust", "rust", "cli"}, nil, time.Now().UTC())
	if plan.Decision != proofGovernorRun || strings.Join(plan.RunChecks, ",") != "full" {
		t.Fatalf("composite request should run full exactly once: %#v", plan)
	}
	if !containsProofGovernorReason(plan.Reasons, "full:no-passing-receipt") {
		t.Fatalf("missing no-receipt reason: %#v", plan.Reasons)
	}
}

func TestProofSelectionBlocksUnknownCheck(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := proofSelectionFixtureRegistry(t, root)

	plan := selectProofGovernorPlan(root, registry, []string{"unknown"}, nil, time.Now().UTC())
	if plan.Decision != proofGovernorBlock || !containsProofGovernorReason(plan.Reasons, "unknown-check:unknown") {
		t.Fatalf("unknown check must block before execution: %#v", plan)
	}
}

func TestProofSelectionRejectsPreIntegrationNamespaceReceipt(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := proofSelectionFixtureRegistry(t, root)
	now := time.Now().UTC()
	rust := findProofSelectionSpec(t, registry, "rust")
	worker := rust
	worker.Namespace.Run = "worker-worktree"
	receipt := captureSelectionReceipt(t, root, registry, worker, now)

	plan := selectProofGovernorPlan(root, registry, []string{"rust"}, []proofGovernorReceipt{receipt}, now.Add(time.Second))
	if plan.Decision != proofGovernorRun || !containsProofGovernorReason(plan.Reasons, "rust:check-semantics-changed") {
		t.Fatalf("worker namespace receipt must not replace integrated proof: %#v", plan)
	}
}

func TestProofPlanCommandIsReadOnlyAndMachineReadable(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	registry := proofSelectionFixtureRegistry(t, root)
	registryPath := filepath.Join(root, "registry.json")
	writeProofSelectionRegistry(t, registryPath, registry)
	receiptDir := filepath.Join(root, "receipts")
	if err := os.MkdirAll(receiptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	receipt := captureSelectionReceipt(t, root, registry, findProofSelectionSpec(t, registry, "rust"), now)
	if err := writeProofGovernorReceipt(filepath.Join(receiptDir, "rust.proof.json"), receipt); err != nil {
		t.Fatal(err)
	}

	before := countProofSelectionFiles(t, root)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{
		"proof-plan", "--root", root, "--registry", "registry.json",
		"--receipt-dir", "receipts", "--request", "rust", "--json",
	}, &stdout, &stderr)
	after := countProofSelectionFiles(t, root)
	if code != 0 || !strings.Contains(stdout.String(), `"decision": "reuse"`) || before != after {
		t.Fatalf("proof-plan failed or mutated state code=%d before=%d after=%d stdout=%s stderr=%s", code, before, after, stdout.String(), stderr.String())
	}
}

func TestReleaseProfileDoesNotRepeatCoreChildCheck(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")
	writeFile(t, filepath.Join(root, ".github", "skills", "demo", "SKILL.md"), "---\nname: demo\ndescription: demo\n---\n")

	checks, err := releaseChecks(root, "local-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range checks {
		if check.Name == "skill-surface-minimality" {
			t.Fatalf("local-release repeats minimality already included by core: %#v", checks)
		}
	}
}

func proofSelectionFixtureRegistry(t *testing.T, root string) proofGovernorRegistry {
	t.Helper()
	writeProofGovernorFixture(t, root, "src/rust.go", "package demo\n")
	writeProofGovernorFixture(t, root, "src/shared.go", "package demo\n")
	writeProofGovernorFixture(t, root, "src/cli.go", "package demo\n")
	writeProofGovernorFixture(t, root, "web/app.tsx", "export const App = () => null\n")
	writeProofGovernorFixture(t, root, "artifact.bin", "stable\n")
	namespace := proofGovernorNamespace{Goal: "proof-governor", Slice: "integrated", Run: "plan-run"}
	makeSpec := func(id string, covers, inputs []string, class string) proofGovernorCheckSpec {
		return proofGovernorCheckSpec{
			ID: id, Namespace: namespace, Command: []string{"fake-check", id},
			Covers: covers, Inputs: inputs, WorkingDir: ".", TimeoutMS: 5_000,
			ExpectedExit: 0, Environment: map[string]string{"runtime": "v1"},
			ExecutionClass: class, MaxAgeSeconds: 600,
		}
	}
	registry := proofGovernorRegistry{
		SchemaVersion: 1,
		Checks: []proofGovernorCheckSpec{
			makeSpec("full", []string{"rust", "cli", "checksum", "browser"}, []string{"src/rust.go", "src/shared.go", "src/cli.go", "web/app.tsx", "artifact.bin"}, "cli"),
			makeSpec("rust", []string{"rust"}, []string{"src/rust.go", "src/shared.go"}, "cli"),
			makeSpec("cli", []string{"cli"}, []string{"src/cli.go", "src/shared.go"}, "cli"),
			makeSpec("checksum", []string{"checksum"}, []string{"artifact.bin"}, "cli"),
			makeSpec("browser", []string{"browser"}, []string{"web/app.tsx"}, "headless-browser"),
		},
	}
	writeProofSelectionRegistry(t, filepath.Join(root, "registry.json"), registry)
	return registry
}

func captureSelectionReceipt(t *testing.T, root string, registry proofGovernorRegistry, spec proofGovernorCheckSpec, now time.Time) proofGovernorReceipt {
	t.Helper()
	receipt, err := captureProofGovernorReceipt(root, spec, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now.Add(-time.Second), CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}

func findProofSelectionSpec(t *testing.T, registry proofGovernorRegistry, id string) proofGovernorCheckSpec {
	t.Helper()
	for _, spec := range registry.Checks {
		if spec.ID == id {
			return spec
		}
	}
	t.Fatalf("missing registry check %s", id)
	return proofGovernorCheckSpec{}
}

func writeProofSelectionRegistry(t *testing.T, path string, registry proofGovernorRegistry) {
	t.Helper()
	content, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func countProofSelectionFiles(t *testing.T, root string) int {
	t.Helper()
	count := 0
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return count
}
