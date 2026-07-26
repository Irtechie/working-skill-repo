package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProofCoverageReceiptReusesPassingSuperset(t *testing.T) {
	root := t.TempDir()
	writeProofGovernorFixture(t, root, "src/rust.go", "package demo\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["full","rust","cli","checksum","browser"]}`)

	spec := proofGovernorCheckSpec{
		ID:             "full",
		Namespace:      proofGovernorNamespace{Goal: "proof-governor", Slice: "slice-001", Run: "run-1"},
		Command:        []string{"go", "test", "./..."},
		Covers:         []string{"rust", "cli", "checksum", "browser"},
		Inputs:         []string{"src/rust.go"},
		OracleFiles:    []string{"registry.json"},
		WorkingDir:     ".",
		TimeoutMS:      30_000,
		ExpectedExit:   0,
		Environment:    map[string]string{"GOOS": "windows", "transport": "headless"},
		ExecutionClass: "cli",
		MaxAgeSeconds:  600,
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	receipt, err := captureProofGovernorReceipt(root, spec, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now.Add(-time.Second), CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}

	decision := evaluateProofGovernorReceipt(root, spec, []string{"rust"}, "registry.json", receipt, now.Add(time.Minute))
	if decision.Decision != proofGovernorReuse {
		t.Fatalf("expected reuse, got %#v", decision)
	}
	if decision.ReceiptID == "" || len(decision.Reasons) == 0 {
		t.Fatalf("reuse must identify its receipt and reason: %#v", decision)
	}
}

func TestRelevantInputFingerprintRejectsChangedDirtyAndUntrackedContent(t *testing.T) {
	root := t.TempDir()
	writeProofGovernorFixture(t, root, "tracked.txt", "before\n")
	writeProofGovernorFixture(t, root, "new-untracked.txt", "candidate\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["focused"]}`)

	spec := proofGovernorCheckSpec{
		ID: "focused", Namespace: proofGovernorNamespace{Goal: "goal", Slice: "slice", Run: "run"},
		Command: []string{"tool", "check"}, Covers: []string{"focused"},
		Inputs: []string{"tracked.txt", "new-untracked.txt"}, OracleFiles: []string{"registry.json"},
		WorkingDir: ".", TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 300,
	}
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	receipt, err := captureProofGovernorReceipt(root, spec, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}

	writeProofGovernorFixture(t, root, "new-untracked.txt", "changed\n")
	decision := evaluateProofGovernorReceipt(root, spec, []string{"focused"}, "registry.json", receipt, now.Add(time.Minute))
	if decision.Decision != proofGovernorRun || !containsProofGovernorReason(decision.Reasons, "relevant-input-changed:new-untracked.txt") {
		t.Fatalf("changed untracked content must invalidate receipt: %#v", decision)
	}

	writeProofGovernorFixture(t, root, "new-untracked.txt", "candidate\n")
	writeProofGovernorFixture(t, root, "tracked.txt", "after\n")
	decision = evaluateProofGovernorReceipt(root, spec, []string{"focused"}, "registry.json", receipt, now.Add(time.Minute))
	if decision.Decision != proofGovernorRun || !containsProofGovernorReason(decision.Reasons, "relevant-input-changed:tracked.txt") {
		t.Fatalf("changed tracked content must invalidate receipt: %#v", decision)
	}
}

func TestProofReceiptRejectsFailedPartialUnknownCoverageAndSemanticDrift(t *testing.T) {
	root := t.TempDir()
	writeProofGovernorFixture(t, root, "input.txt", "stable\n")
	writeProofGovernorFixture(t, root, "oracle.txt", "assertion\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["full","rust"]}`)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	base := proofGovernorCheckSpec{
		ID: "full", Namespace: proofGovernorNamespace{Goal: "goal", Slice: "slice", Run: "run"},
		Command: []string{"tool", "--check"}, Covers: []string{"rust"},
		Inputs: []string{"input.txt"}, OracleFiles: []string{"oracle.txt"},
		WorkingDir: ".", TimeoutMS: 2_000, ExpectedExit: 0,
		Environment: map[string]string{"runtime": "v1"}, ExecutionClass: "cli", MaxAgeSeconds: 300,
	}

	passing, err := captureProofGovernorReceipt(root, base, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		request []string
		spec    proofGovernorCheckSpec
		mutate  func(*proofGovernorReceipt)
		reason  string
	}{
		{name: "unknown coverage", request: []string{"browser"}, spec: base, reason: "coverage-miss:browser"},
		{name: "failed", request: []string{"rust"}, spec: base, mutate: func(r *proofGovernorReceipt) { r.Result.Status = "fail" }, reason: "receipt-not-passing"},
		{name: "partial", request: []string{"rust"}, spec: base, mutate: func(r *proofGovernorReceipt) { r.Result.Status = "partial" }, reason: "receipt-not-passing"},
		{name: "command", request: []string{"rust"}, spec: mutateProofGovernorSpec(base, func(s *proofGovernorCheckSpec) { s.Command = []string{"tool", "--other"} }), reason: "check-semantics-changed"},
		{name: "cwd", request: []string{"rust"}, spec: mutateProofGovernorSpec(base, func(s *proofGovernorCheckSpec) { s.WorkingDir = "subdir" }), reason: "check-semantics-changed"},
		{name: "timeout", request: []string{"rust"}, spec: mutateProofGovernorSpec(base, func(s *proofGovernorCheckSpec) { s.TimeoutMS++ }), reason: "check-semantics-changed"},
		{name: "expected", request: []string{"rust"}, spec: mutateProofGovernorSpec(base, func(s *proofGovernorCheckSpec) { s.ExpectedExit = 1 }), reason: "check-semantics-changed"},
		{name: "environment", request: []string{"rust"}, spec: mutateProofGovernorSpec(base, func(s *proofGovernorCheckSpec) { s.Environment = map[string]string{"runtime": "v2"} }), reason: "check-semantics-changed"},
		{name: "namespace", request: []string{"rust"}, spec: mutateProofGovernorSpec(base, func(s *proofGovernorCheckSpec) { s.Namespace.Run = "run-2" }), reason: "check-semantics-changed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			receipt := passing
			if tt.mutate != nil {
				tt.mutate(&receipt)
			}
			decision := evaluateProofGovernorReceipt(root, tt.spec, tt.request, "registry.json", receipt, now.Add(time.Minute))
			if decision.Decision != proofGovernorRun || !containsProofGovernorReason(decision.Reasons, tt.reason) {
				t.Fatalf("expected RUN with %q, got %#v", tt.reason, decision)
			}
		})
	}
}

func TestProofReceiptRejectsMissingInputsRegistryDriftAndExpiry(t *testing.T) {
	root := t.TempDir()
	writeProofGovernorFixture(t, root, "input.txt", "stable\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["focused"]}`)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	spec := proofGovernorCheckSpec{
		ID: "focused", Namespace: proofGovernorNamespace{Goal: "goal", Slice: "slice", Run: "run"},
		Command: []string{"tool"}, Covers: []string{"focused"}, Inputs: []string{"input.txt"},
		WorkingDir: ".", TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 30,
	}
	receipt, err := captureProofGovernorReceipt(root, spec, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Remove(filepath.Join(root, "input.txt")); err != nil {
		t.Fatal(err)
	}
	decision := evaluateProofGovernorReceipt(root, spec, []string{"focused"}, "registry.json", receipt, now.Add(time.Second))
	if decision.Decision != proofGovernorRun || !containsProofGovernorReason(decision.Reasons, "relevant-input-missing:input.txt") {
		t.Fatalf("missing input must invalidate: %#v", decision)
	}

	writeProofGovernorFixture(t, root, "input.txt", "stable\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["focused","new"]}`)
	decision = evaluateProofGovernorReceipt(root, spec, []string{"focused"}, "registry.json", receipt, now.Add(time.Second))
	if decision.Decision != proofGovernorRun || !containsProofGovernorReason(decision.Reasons, "registry-changed") {
		t.Fatalf("registry drift must invalidate: %#v", decision)
	}

	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["focused"]}`)
	decision = evaluateProofGovernorReceipt(root, spec, []string{"focused"}, "registry.json", receipt, now.Add(time.Minute))
	if decision.Decision != proofGovernorRun || !containsProofGovernorReason(decision.Reasons, "receipt-expired") {
		t.Fatalf("expired receipt must invalidate: %#v", decision)
	}
}

func TestProofReceiptFileValidationRejectsTampering(t *testing.T) {
	root := t.TempDir()
	writeProofGovernorFixture(t, root, "input.txt", "stable\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["focused"]}`)
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	spec := proofGovernorCheckSpec{
		ID: "focused", Namespace: proofGovernorNamespace{Goal: "goal", Slice: "slice", Run: "run"},
		Command: []string{"tool"}, Covers: []string{"focused"}, Inputs: []string{"input.txt"},
		WorkingDir: ".", TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 300,
	}
	receipt, err := captureProofGovernorReceipt(root, spec, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "focused.proof.json")
	if err := writeProofGovernorReceipt(path, receipt); err != nil {
		t.Fatal(err)
	}
	if issues := validateProofGovernorReceiptFile(root, path, now.Add(time.Second)); len(issues) != 0 {
		t.Fatalf("valid receipt rejected: %v", issues)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tampered := strings.Replace(string(content), `"status": "pass"`, `"status": "fail"`, 1)
	if err := os.WriteFile(path, []byte(tampered), 0o644); err != nil {
		t.Fatal(err)
	}
	if issues := validateProofGovernorReceiptFile(root, path, now.Add(time.Second)); !containsProofGovernorReason(issues, "receipt-integrity-mismatch") {
		t.Fatalf("tampering was not rejected: %v", issues)
	}
}

func TestManifestContractValidatesProofReceiptContents(t *testing.T) {
	root := t.TempDir()
	writeProofGovernorFixture(t, root, "go.mod", "module example.test/proof\n")
	writeProofGovernorFixture(t, root, "input.txt", "stable\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["focused"]}`)
	now := time.Now().UTC()
	spec := proofGovernorCheckSpec{
		ID: "focused", Namespace: proofGovernorNamespace{Goal: "goal", Slice: "slice", Run: "run"},
		Command: []string{"tool"}, Covers: []string{"focused"}, Inputs: []string{"input.txt"},
		WorkingDir: ".", TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 300,
	}
	receipt, err := captureProofGovernorReceipt(root, spec, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	receiptPath := filepath.Join(root, "focused.proof.json")
	if err := writeProofGovernorReceipt(receiptPath, receipt); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(root, "docs", "plans", "manifest.md")
	writeProofGovernorFixture(t, root, "docs/plans/manifest.md", "---\n---\n")
	gate := manifestGate{
		GateID: "slice-to-done", Status: "passed", RequiredEvidence: []string{"receipt"},
		Proof: []string{receiptPath}, PassedAt: now.Format(time.RFC3339),
	}
	if issues := validateAdvanceableGate(manifestPath, gate, false); len(issues) != 0 {
		t.Fatalf("valid proof receipt rejected by manifest gate: %#v", issues)
	}

	writeProofGovernorFixture(t, root, "input.txt", "changed\n")
	issues := validateAdvanceableGate(manifestPath, gate, false)
	if !hasManifestIssue(issues, "invalid-proof-receipt") {
		t.Fatalf("stale proof receipt passed manifest gate: %#v", issues)
	}
}

func TestProofReceiptValidateCommand(t *testing.T) {
	root := t.TempDir()
	writeProofGovernorFixture(t, root, "input.txt", "stable\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"checks":["focused"]}`)
	now := time.Now().UTC()
	spec := proofGovernorCheckSpec{
		ID: "focused", Namespace: proofGovernorNamespace{Goal: "goal", Slice: "slice", Run: "run"},
		Command: []string{"tool"}, Covers: []string{"focused"}, Inputs: []string{"input.txt"},
		WorkingDir: ".", TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 300,
	}
	receipt, err := captureProofGovernorReceipt(root, spec, "registry.json", proofGovernorExecutionResult{
		Status: "pass", ExitCode: 0, StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := writeProofGovernorReceipt(filepath.Join(root, "focused.proof.json"), receipt); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"proof-receipt-validate", "--root", root, "--receipt", "focused.proof.json", "--json"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), `"ok": true`) {
		t.Fatalf("receipt CLI failed code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func mutateProofGovernorSpec(spec proofGovernorCheckSpec, mutate func(*proofGovernorCheckSpec)) proofGovernorCheckSpec {
	copySpec := spec
	copySpec.Command = append([]string{}, spec.Command...)
	copySpec.Covers = append([]string{}, spec.Covers...)
	copySpec.Inputs = append([]string{}, spec.Inputs...)
	copySpec.OracleFiles = append([]string{}, spec.OracleFiles...)
	copySpec.Environment = map[string]string{}
	for key, value := range spec.Environment {
		copySpec.Environment[key] = value
	}
	mutate(&copySpec)
	return copySpec
}

func containsProofGovernorReason(reasons []string, expected string) bool {
	for _, reason := range reasons {
		if reason == expected || strings.HasPrefix(reason, expected) {
			return true
		}
	}
	return false
}

func writeProofGovernorFixture(t *testing.T, root, relative, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
