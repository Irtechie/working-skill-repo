package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestModelRoutingReleaseAcceptsHonestNoPaidArtifactWithoutPromotion(t *testing.T) {
	root, evidence := writeValidModelRoutingReleaseFixture(t)
	code, stdout, stderr := runModelRoutingReleaseForTest(root, evidence)
	if code != 0 {
		t.Fatalf("honest no-paid artifact failed: code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "decision=not-promoted") || !strings.Contains(stdout, "supported_cohorts=0") {
		t.Fatalf("success output implied promotion or omitted support count: %s", stdout)
	}
}

func TestModelRoutingReleaseRejectsSelfAuthoredLiveSupportWithoutExternalVerifier(t *testing.T) {
	root, evidencePath := writeValidModelRoutingReleaseFixture(t)
	receiptPath := filepath.Join(root, "docs", "results", "live-receipt.json")
	installPath := filepath.Join(root, "docs", "results", "install-proof.json")
	writeReleaseJSON(t, receiptPath, map[string]any{
		"schema_version": 1, "evidence_class": "route-bound-live-receipt", "cohort": "codex-cli",
		"adapter": "codex", "runtime": "codex-cli", "selection_policy": "planned-tier", "proof_harness": "v1", "route_revision": "v1",
		"route_fingerprint": strings.Repeat("1", 64), "context_packet_hash": strings.Repeat("2", 64), "work_proof_hash": strings.Repeat("3", 64),
		"proof_status": "pass", "live": true,
	})
	writeReleaseJSON(t, installPath, map[string]any{
		"schema_version": 1, "evidence_class": "install-proof", "cohort": "codex-cli", "adapter_revision": "v1",
		"installed_hash": strings.Repeat("4", 64), "platforms": []any{"windows"},
	})
	evidence := readJSONMap(t, evidencePath)
	evidence["evidence_mode"] = "attended-live"
	evidence["model_provenance"] = "route-bound"
	evidence["paid_calls"] = float64(2)
	evidence["live_support_status"] = "qualified"
	evidence["router"] = map[string]any{"status": "available", "reason": "attended-live-probe"}
	evidence["supported_cohorts"] = []any{"codex-cli"}
	evidence["support_claims"] = []any{map[string]any{
		"cohort": "codex-cli", "status": "supported", "adapter": "codex", "runtime": "codex-cli",
		"selection_policy": "planned-tier", "proof_harness": "v1", "route_revision": "v1", "evidence_source": "live-route-bound",
		"live_receipt":  map[string]any{"path": "docs/results/live-receipt.json", "sha256": releaseFileHash(t, receiptPath)},
		"install_proof": map[string]any{"path": "docs/results/install-proof.json", "sha256": releaseFileHash(t, installPath)},
	}}
	writeReleaseJSON(t, evidencePath, evidence)
	code, stdout, stderr := runModelRoutingReleaseForTest(root, evidencePath)
	if code == 0 || !strings.Contains(stderr, "external verifier") {
		t.Fatalf("self-authored live support was not rejected: code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestModelRoutingReleaseRejectsDishonestOrUnsafeEvidence(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(t *testing.T, root string, evidence map[string]any)
		want   string
	}{
		{
			name: "unknown evidence field",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				evidence["self_reported_savings"] = true
			},
			want: "unknown field",
		},
		{
			name: "fixture path escape",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				fixtureRefs(evidence)[0]["path"] = "../outside.json"
			},
			want: "repository-relative",
		},
		{
			name: "fixture hash mismatch",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				fixtureRefs(evidence)[0]["sha256"] = strings.Repeat("0", 64)
			},
			want: "hash mismatch",
		},
		{
			name: "deterministic fixture counted as efficiency",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				fixtureRefs(evidence)[0]["efficiency_evidence"] = true
			},
			want: "cannot count as efficiency",
		},
		{
			name: "correction fixture counted as support",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				fixtureRefs(evidence)[1]["support_evidence"] = true
			},
			want: "cannot count as support",
		},
		{
			name: "fixture claims live execution",
			mutate: func(t *testing.T, root string, evidence map[string]any) {
				path := filepath.Join(root, "evals", "model-routing", "initial-pilot.json")
				fixture := readJSONMap(t, path)
				fixture["live"] = true
				writeReleaseJSON(t, path, fixture)
				fixtureRefs(evidence)[0]["sha256"] = releaseFileHash(t, path)
			},
			want: "fixture must be non-live",
		},
		{
			name: "fixture unknown field",
			mutate: func(t *testing.T, root string, evidence map[string]any) {
				path := filepath.Join(root, "evals", "model-routing", "initial-pilot.json")
				fixture := readJSONMap(t, path)
				fixture["claimed_support"] = true
				writeReleaseJSON(t, path, fixture)
				fixtureRefs(evidence)[0]["sha256"] = releaseFileHash(t, path)
			},
			want: "unknown field",
		},
		{
			name: "unsupported GHCP support claim",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				evidence["supported_cohorts"] = []any{"ghcp"}
				evidence["support_claims"] = []any{map[string]any{
					"cohort": "ghcp", "status": "supported", "adapter": "ghcp", "runtime": "ghcp",
					"selection_policy": "planned-tier", "proof_harness": "v1", "route_revision": "v1",
					"evidence_source": "live-route-bound",
				}}
			},
			want: "unsupported cohort",
		},
		{
			name: "supported cohort lacks live receipt and install proof",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				evidence["supported_cohorts"] = []any{"codex-cli"}
				evidence["support_claims"] = []any{map[string]any{
					"cohort": "codex-cli", "status": "supported", "adapter": "codex", "runtime": "codex-cli",
					"selection_policy": "planned-tier", "proof_harness": "v1", "route_revision": "v1",
					"evidence_source": "live-route-bound",
				}}
			},
			want: "live receipt",
		},
		{
			name: "promotion without preregistered gates",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				evidence["release_decision"] = "promoted"
				evidence["next_lower_attempts"] = "enabled"
			},
			want: "not-promoted",
		},
		{
			name: "unavailable metric encoded as zero",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				metrics := evidence["metrics"].(map[string]any)
				cost := metrics["total_billed_cost"].(map[string]any)
				cost["value"] = float64(0)
			},
			want: "unavailable metric",
		},
		{
			name: "claimed support sourced from deterministic fixture",
			mutate: func(_ *testing.T, _ string, evidence map[string]any) {
				evidence["supported_cohorts"] = []any{"codex-cli"}
				evidence["support_claims"] = []any{map[string]any{
					"cohort": "codex-cli", "status": "supported", "adapter": "codex", "runtime": "codex-cli",
					"selection_policy": "planned-tier", "proof_harness": "v1", "route_revision": "v1",
					"evidence_source": "deterministic-fixture",
				}}
			},
			want: "route-bound live evidence",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			root, evidencePath := writeValidModelRoutingReleaseFixture(t)
			evidence := readJSONMap(t, evidencePath)
			tc.mutate(t, root, evidence)
			writeReleaseJSON(t, evidencePath, evidence)
			code, stdout, stderr := runModelRoutingReleaseForTest(root, evidencePath)
			if code == 0 {
				t.Fatalf("dishonest artifact passed: stdout=%s", stdout)
			}
			if !strings.Contains(strings.ToLower(stderr), strings.ToLower(tc.want)) {
				t.Fatalf("expected error containing %q, got stdout=%s stderr=%s", tc.want, stdout, stderr)
			}
		})
	}
}

func TestModelRoutingReleaseRejectsSymlinkFixture(t *testing.T) {
	root, evidencePath := writeValidModelRoutingReleaseFixture(t)
	target := filepath.Join(root, "evals", "model-routing", "initial-pilot-target.json")
	link := filepath.Join(root, "evals", "model-routing", "initial-pilot-link.json")
	writeReleaseFile(t, target, []byte(`{"schema_version":1}`))
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink creation unavailable: %v", err)
	}
	evidence := readJSONMap(t, evidencePath)
	fixtureRefs(evidence)[0]["path"] = "evals/model-routing/initial-pilot-link.json"
	fixtureRefs(evidence)[0]["sha256"] = releaseFileHash(t, target)
	writeReleaseJSON(t, evidencePath, evidence)
	code, stdout, stderr := runModelRoutingReleaseForTest(root, evidencePath)
	if code == 0 || !strings.Contains(strings.ToLower(stderr), "symlink") {
		t.Fatalf("symlink fixture was not rejected: code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
}

func TestModelRoutingReleaseProofRunnerIsFixedNoPaidAndFailurePropagates(t *testing.T) {
	root, evidencePath := writeValidModelRoutingReleaseFixture(t)
	calls := [][]string{}
	runner := func(_ context.Context, _ string, command []string) modelRoutingProofResult {
		calls = append(calls, append([]string(nil), command...))
		if len(calls) == 2 {
			return modelRoutingProofResult{ExitCode: 1, Output: "FAIL seeded refusal proof"}
		}
		return modelRoutingProofResult{ExitCode: 0, Output: "ok deterministic proof"}
	}
	var stdout, stderr bytes.Buffer
	code := runModelRoutingReleaseCommandWithRunner(root, options{cohort: "initial-pilot", evidencePath: evidencePath}, &stdout, &stderr, runner)
	if code == 0 || !strings.Contains(stderr.String(), "deterministic proof command failed") {
		t.Fatalf("proof failure did not fail release validation: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if len(calls) != 2 {
		t.Fatalf("expected two fixed proof calls, got %d", len(calls))
	}
	for _, command := range calls {
		joined := strings.Join(command, " ")
		if len(command) < 3 || command[0] != "go" || command[1] != "test" || strings.Contains(joined, "codex") || strings.Contains(joined, "copilot") || strings.Contains(joined, "dispatch --") {
			t.Fatalf("proof runner received a paid, model, network, or configurable command: %q", joined)
		}
	}
}

func TestModelRoutingPromotionRequiresMeasuredNonzeroPrimaryMetric(t *testing.T) {
	confidence := 1.0
	improvement := 20.0
	evidence := modelRoutingReleaseEvidence{
		SupportClaims: []modelRoutingSupportClaim{{Cohort: "codex-cli"}},
		Promotion: modelRoutingPromotionStatus{
			PreregisteredGatesMet: true, IndependentHoldout: true, PowerJustified: true,
			LiveSampleSize: 20, TaskFamilies: 3, ZeroRightToWrong: true,
			ConfidenceBoundMaxRegressionPercent: &confidence, PrimaryMetric: "total_billed_cost",
			MedianImprovementPercent: &improvement, CorrectionSafety: "qualified",
		},
		Metrics: modelRoutingReleaseMetrics{
			TotalBilledCost: modelRoutingMetric{Status: "unavailable"},
			TotalTokens:     modelRoutingMetric{Status: "unavailable"},
			WallClockMS:     modelRoutingMetric{Status: "unavailable"},
		},
	}
	if err := validateModelRoutingPromotion(evidence); err == nil || !strings.Contains(err.Error(), "measured") {
		t.Fatalf("unavailable primary metric justified promotion: %v", err)
	}
	zero := 0.0
	evidence.Metrics.TotalBilledCost = modelRoutingMetric{Status: "measured", Value: &zero, Unit: "AIC"}
	if err := validateModelRoutingPromotion(evidence); err == nil || !strings.Contains(err.Error(), "nonzero") {
		t.Fatalf("zero primary metric justified promotion: %v", err)
	}
}

func writeValidModelRoutingReleaseFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	initialPath := filepath.Join(root, "evals", "model-routing", "initial-pilot.json")
	correctionPath := filepath.Join(root, "evals", "model-routing", "correction-pilot.json")
	writeReleaseJSON(t, initialPath, map[string]any{
		"schema_version": 1, "suite": "model-routing-initial-pilot", "evidence_class": "deterministic-conformance",
		"execution_capability": "non-live-routing-conformance", "live": false, "efficiency_evidence": false,
		"cases": []any{map[string]any{"id": "small-planned-tier", "task_family": "code-fix", "expected": "planned-tier-or-handoff"}},
	})
	writeReleaseJSON(t, correctionPath, map[string]any{
		"schema_version": 1, "suite": "model-routing-correction-pilot", "evidence_class": "seeded-correction-safety",
		"execution_capability": "handoff-validation-and-refusal-only", "live": false, "efficiency_evidence": false,
		"cases": []any{map[string]any{"id": "localized-failure", "task_family": "code-fix", "expected": "fail-closed-ordinary-fallback"}},
	})
	evidence := map[string]any{
		"schema_version":      1,
		"cohort":              "initial-pilot",
		"evidence_mode":       "deterministic-no-paid",
		"model_provenance":    "unavailable",
		"paid_calls":          0,
		"baseline":            "planned-tier-host-native",
		"release_decision":    "not-promoted",
		"next_lower_attempts": "disabled",
		"live_support_status": "not-qualified",
		"router":              map[string]any{"status": "unavailable", "reason": "private-acl-access-denied"},
		"supported_cohorts":   []any{},
		"support_claims":      []any{},
		"parked_surfaces":     []any{"ghcp", "codex-app-exact-attribution", "tinyboss-control", "mcp-dispatch", "direct-chat-completions"},
		"fixtures": []any{
			map[string]any{"class": "deterministic-conformance", "path": "evals/model-routing/initial-pilot.json", "sha256": releaseFileHash(t, initialPath), "live": false, "efficiency_evidence": false, "support_evidence": false},
			map[string]any{"class": "seeded-correction-safety", "path": "evals/model-routing/correction-pilot.json", "sha256": releaseFileHash(t, correctionPath), "live": false, "efficiency_evidence": false, "support_evidence": false},
		},
		"metrics": map[string]any{
			"total_billed_cost": map[string]any{"status": "unavailable"},
			"total_tokens":      map[string]any{"status": "unavailable"},
			"wall_clock_ms":     map[string]any{"status": "unavailable"},
		},
		"promotion": map[string]any{
			"preregistered_gates_met": false, "independent_holdout": false, "power_justified": false,
			"live_sample_size": 0, "task_families": 0, "zero_right_to_wrong": false,
			"primary_metric": "unavailable", "correction_safety": "not-run",
		},
	}
	evidencePath := filepath.Join(root, "docs", "results", "evidence.json")
	writeReleaseJSON(t, evidencePath, evidence)
	return root, evidencePath
}

func runModelRoutingReleaseForTest(root, evidence string) (int, string, string) {
	var stdout, stderr bytes.Buffer
	code := runModelRoutingReleaseCommandWithRunner(root, options{cohort: "initial-pilot", evidencePath: evidence}, &stdout, &stderr, modelRoutingDisabledProofRunner)
	return code, stdout.String(), stderr.String()
}

func fixtureRefs(evidence map[string]any) []map[string]any {
	raw := evidence["fixtures"].([]any)
	refs := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		refs = append(refs, item.(map[string]any))
	}
	return refs
}

func readJSONMap(t *testing.T, path string) map[string]any {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err := json.Unmarshal(content, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func writeReleaseJSON(t *testing.T, path string, value any) {
	t.Helper()
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	writeReleaseFile(t, path, append(content, '\n'))
}

func writeReleaseFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}

func releaseFileHash(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
