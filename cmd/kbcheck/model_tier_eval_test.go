package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type modelTierEvalFixtureCorpus struct {
	SchemaVersion int                        `json:"schema_version"`
	Cases         []modelTierEvalFixtureCase `json:"cases"`
}

type modelTierEvalFixtureCase struct {
	ID                string `json:"id"`
	Mutation          string `json:"mutation"`
	ExpectedDecision  string `json:"expected_decision,omitempty"`
	ExpectedExclusion string `json:"expected_exclusion,omitempty"`
	ExpectedError     bool   `json:"expected_error"`
}

func TestModelTierEvalDeterministicCorpus(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "evals", "model-tier-qualification", "fixtures.json"))
	if err != nil {
		t.Fatalf("read fixture corpus: %v", err)
	}
	var corpus modelTierEvalFixtureCorpus
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&corpus); err != nil {
		t.Fatalf("decode fixture corpus: %v", err)
	}
	if corpus.SchemaVersion != 1 || len(corpus.Cases) < 15 {
		t.Fatalf("fixture corpus is too small or unsupported: %#v", corpus)
	}
	for _, fixture := range corpus.Cases {
		t.Run(fixture.ID, func(t *testing.T) {
			root := t.TempDir()
			evidence := validModelTierEvalEvidence()
			applyModelTierEvalMutation(t, evidence, fixture.Mutation)
			materializeModelTierEvalArtifacts(t, root, evidence)
			applyModelTierEvalPostMaterializationMutation(t, evidence, fixture.Mutation)
			path := filepath.Join(root, "evidence.json")
			writeJSONFixture(t, path, evidence)
			var stdout, stderr strings.Builder
			code := run([]string{"model-tier-eval", "--root", root, "--evidence", "evidence.json", "--json"}, &stdout, &stderr)
			if fixture.ExpectedError {
				if code == 0 {
					t.Fatalf("expected command error, output=%s", stdout.String())
				}
				return
			}
			if code != 0 {
				t.Fatalf("unexpected command error %d: %s", code, stderr.String())
			}
			var result map[string]any
			if err := json.Unmarshal([]byte(stdout.String()), &result); err != nil {
				t.Fatalf("decode result: %v\n%s", err, stdout.String())
			}
			if result["decision"] != fixture.ExpectedDecision {
				t.Fatalf("decision=%v want=%s result=%s", result["decision"], fixture.ExpectedDecision, stdout.String())
			}
			if result["experimental"] != true || result["routing_promotion"] != false {
				t.Fatalf("result lost experimental boundary: %s", stdout.String())
			}
			if fixture.ExpectedExclusion != "" {
				exclusions, ok := result["exclusions"].(map[string]any)
				if !ok || exclusions[fixture.ExpectedExclusion].(float64) < 1 {
					t.Fatalf("missing exclusion %s: %s", fixture.ExpectedExclusion, stdout.String())
				}
			}
		})
	}
}

func TestModelTierEvalTextOutputIsScoped(t *testing.T) {
	root := t.TempDir()
	evidence := validModelTierEvalEvidence()
	materializeModelTierEvalArtifacts(t, root, evidence)
	writeJSONFixture(t, filepath.Join(root, "evidence.json"), evidence)
	var stdout, stderr strings.Builder
	code := run([]string{"model-tier-eval", "--root", root, "--evidence", "evidence.json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("command failed %d: %s", code, stderr.String())
	}
	for _, want := range []string{"qualified", "medium", "30 admitted", "experimental", "does not promote routing"} {
		if !strings.Contains(strings.ToLower(stdout.String()), want) {
			t.Fatalf("text output missing %q: %s", want, stdout.String())
		}
	}
}

func TestModelTierEvalRejectsEscapedSymlinkAndOversizeInput(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.json")
	writeJSONFixture(t, outside, validModelTierEvalEvidence())
	for name, path := range map[string]string{
		"absolute":  outside,
		"traversal": filepath.Join("..", filepath.Base(filepath.Dir(outside)), filepath.Base(outside)),
	} {
		t.Run(name, func(t *testing.T) {
			if code := run([]string{"model-tier-eval", "--root", root, "--evidence", path, "--json"}, &strings.Builder{}, &strings.Builder{}); code == 0 {
				t.Fatalf("%s evidence path passed", name)
			}
		})
	}
	symlink := filepath.Join(root, "evidence.json")
	if err := os.Symlink(outside, symlink); err == nil {
		if code := run([]string{"model-tier-eval", "--root", root, "--evidence", "evidence.json", "--json"}, &strings.Builder{}, &strings.Builder{}); code == 0 {
			t.Fatal("symlink evidence passed")
		}
	}
	if err := os.WriteFile(filepath.Join(root, "large.json"), bytes.Repeat([]byte("x"), int(maxModelRoutingReleaseBytes+1)), 0o644); err != nil {
		t.Fatalf("write oversized input: %v", err)
	}
	if code := run([]string{"model-tier-eval", "--root", root, "--evidence", "large.json", "--json"}, &strings.Builder{}, &strings.Builder{}); code == 0 {
		t.Fatal("oversized evidence passed")
	}
}

func validModelTierEvalEvidence() map[string]any {
	families := []any{}
	for index := 0; index < 5; index++ {
		families = append(families, map[string]any{
			"id":      fmt.Sprintf("family-%d", index+1),
			"holdout": index == 4,
			"dimensions": map[string]any{
				"task_structure":       fmt.Sprintf("structure-%d", index+1),
				"repository_mechanism": fmt.Sprintf("mechanism-%d", index+1),
				"oracle_type":          fmt.Sprintf("oracle-%d", index+1),
				"primary_failure_mode": fmt.Sprintf("failure-%d", index+1),
			},
		})
	}
	fingerprint := map[string]any{
		"route_fingerprint":        strings.Repeat("1", 64),
		"provider_model_revision":  "provider-revision-1",
		"route_config_revision":    "route-config-1",
		"system_instructions_hash": strings.Repeat("2", 64),
		"tools_hash":               strings.Repeat("3", 64),
		"context_risk_policy_hash": strings.Repeat("4", 64),
		"planner_revision":         "kb-plan-v1",
		"scorer_revision":          "model-tier-eval-v1",
		"oracle_revision":          "oracle-v1",
	}
	fingerprintHash := hashJSONFixture(fingerprint)
	attempts := []any{}
	expected := []any{}
	for index := 0; index < 30; index++ {
		id := fmt.Sprintf("attempt-%02d", index+1)
		expected = append(expected, id)
		attempts = append(attempts, map[string]any{
			"id":                           id,
			"fixture_id":                   fmt.Sprintf("fixture-%02d", index+1),
			"fixture_revision":             "v1",
			"family_id":                    fmt.Sprintf("family-%d", index/6+1),
			"target_tier":                  "medium",
			"execution_fingerprint_sha256": fingerprintHash,
			"plan_sha256":                  hashStringFixture("plan-" + id),
			"request_sha256":               hashStringFixture("request-" + id),
			"response_sha256":              hashStringFixture("response-" + id),
			"proof_sha256":                 hashStringFixture("proof-" + id),
			"attempted_at":                 "2026-06-02T00:00:00Z",
			"outcome":                      "model-pass",
			"failure_kind":                 "",
			"response_produced":            true,
			"one_attempt_identity_valid":   true,
			"proof_bound":                  true,
			"trust": map[string]any{
				"plan":      "reproduced",
				"route":     "reproduced",
				"execution": "reproduced",
				"oracle":    "reproduced",
				"proof":     "reproduced",
			},
			"artifact_paths": map[string]any{
				"cohort_manifest":      "",
				"fixture":              "",
				"frozen_plan":          "",
				"request":              "",
				"response":             "",
				"plan_sufficiency":     "",
				"executor_package":     "",
				"route_execution":      "",
				"oracle_qualification": "",
				"proof_result":         "",
			},
			"artifact_hashes": map[string]any{
				"cohort_manifest":      hashStringFixture("cohort"),
				"fixture":              hashStringFixture("fixture-" + id),
				"frozen_plan":          hashStringFixture("plan-" + id),
				"plan_sufficiency":     hashStringFixture("sufficiency-" + id),
				"executor_package":     hashStringFixture("executor"),
				"route_execution":      hashStringFixture("route-" + id),
				"oracle_qualification": hashStringFixture("oracle"),
				"proof_result":         hashStringFixture("proof-" + id),
			},
		})
	}
	return map[string]any{
		"schema_version":     1,
		"target_tier":        "medium",
		"threshold_revision": "medium-v1",
		"evidence_date":      "2026-06-02T00:00:00Z",
		"scope": map[string]any{
			"task_families":         []any{"family-1", "family-2", "family-3", "family-4", "family-5"},
			"tools":                 []any{"git", "go"},
			"context_risk_envelope": "bounded-normal-risk",
			"route_revision":        "route-config-1",
		},
		"cohort": map[string]any{
			"id":                         "medium-cohort-1",
			"manifest_path":              filepath.ToSlash(filepath.Join("artifacts", "cohort-manifest.json")),
			"manifest_sha256":            hashStringFixture("cohort"),
			"manifest_signer_id":         "model-tier-test",
			"manifest_signature_ed25519": strings.Repeat("0", ed25519.SignatureSize*2),
			"preregistered_at":           "2026-06-01T00:00:00Z",
			"first_attempt_at":           "2026-06-02T00:00:00Z",
			"provenance_kind":            "ed25519-signature",
			"expected_attempt_ids":       expected,
			"families":                   families,
		},
		"execution_fingerprint":                fingerprint,
		"current_execution_fingerprint_sha256": fingerprintHash,
		"attempts":                             attempts,
	}
}

func applyModelTierEvalMutation(t *testing.T, evidence map[string]any, mutation string) {
	t.Helper()
	attempts := evidence["attempts"].([]any)
	cohort := evidence["cohort"].(map[string]any)
	switch mutation {
	case "none":
	case "proof-failure", "output-schema-failure":
		attempt := attempts[0].(map[string]any)
		attempt["outcome"] = "model-failure"
		attempt["failure_kind"] = strings.TrimSuffix(mutation, "-failure")
	case "underpowered":
		evidence["attempts"] = attempts[:29]
		cohort["expected_attempt_ids"] = cohort["expected_attempt_ids"].([]any)[:29]
	case "four-families":
		evidence["attempts"] = attempts[:24]
		cohort["expected_attempt_ids"] = cohort["expected_attempt_ids"].([]any)[:24]
		cohort["families"] = cohort["families"].([]any)[:4]
	case "missing-holdout":
		for _, family := range cohort["families"].([]any) {
			family.(map[string]any)["holdout"] = false
		}
	case "family-concentration":
		for index := 0; index < 7; index++ {
			attempts[index].(map[string]any)["family_id"] = "family-1"
		}
	case "stale-fingerprint":
		evidence["current_execution_fingerprint_sha256"] = strings.Repeat("0", 64)
	case "unsupported-trust":
		attempts[0].(map[string]any)["trust"].(map[string]any)["proof"] = "unsupported"
	case "unobservable-expired":
		evidence["execution_fingerprint"].(map[string]any)["provider_model_revision"] = ""
		evidence["evidence_date"] = "2000-01-01T00:00:00Z"
		refreshModelTierEvalFingerprint(evidence)
	case "omitted-attempt":
		evidence["attempts"] = attempts[:29]
	case "omitted-attempt-and-ledger", "artifact-hash-mismatch", "cohort-manifest-hash-mismatch", "forged-provenance":
	case "replayed-attempt":
		evidence["attempts"] = append(attempts, attempts[0])
	case "unknown-field":
		evidence["unknown"] = true
	case "mixed-fingerprint":
		attempts[0].(map[string]any)["execution_fingerprint_sha256"] = strings.Repeat("0", 64)
	case "unsupported-target":
		evidence["target_tier"] = "large"
	case "sensitive-value":
		evidence["scope"].(map[string]any)["route_revision"] = "https://user:secret@example.test/v1"
	case "correlated-families":
		families := cohort["families"].([]any)
		families[1].(map[string]any)["dimensions"] = families[0].(map[string]any)["dimensions"]
	default:
		exclusions := map[string]string{
			"exclude-plan":          "plan-insufficient",
			"exclude-oracle":        "oracle-invalid",
			"exclude-preflight":     "preflight-schema",
			"exclude-route":         "route-infrastructure-not-run",
			"exclude-no-response":   "execution-failed-no-response",
			"exclude-indeterminate": "execution-indeterminate",
			"exclude-proof":         "proof-infrastructure",
		}
		outcome, ok := exclusions[mutation]
		if !ok {
			t.Fatalf("unknown fixture mutation %q", mutation)
		}
		attempt := attempts[0].(map[string]any)
		attempt["outcome"] = outcome
		attempt["response_produced"] = false
		attempt["proof_bound"] = false
	}
}

func applyModelTierEvalPostMaterializationMutation(t *testing.T, evidence map[string]any, mutation string) {
	t.Helper()
	switch mutation {
	case "omitted-attempt-and-ledger":
		attempts := evidence["attempts"].([]any)
		evidence["attempts"] = attempts[:29]
		cohort := evidence["cohort"].(map[string]any)
		expected := cohort["expected_attempt_ids"].([]any)
		cohort["expected_attempt_ids"] = expected[:29]
	case "artifact-hash-mismatch":
		attempt := evidence["attempts"].([]any)[0].(map[string]any)
		attempt["artifact_hashes"].(map[string]any)["frozen_plan"] = strings.Repeat("0", 64)
	case "cohort-manifest-hash-mismatch":
		evidence["cohort"].(map[string]any)["manifest_sha256"] = strings.Repeat("0", 64)
	case "forged-provenance":
		evidence["cohort"].(map[string]any)["manifest_signature_ed25519"] = strings.Repeat("0", ed25519.SignatureSize*2)
	}
}

func materializeModelTierEvalArtifacts(t *testing.T, root string, evidence map[string]any) {
	t.Helper()
	cohort := evidence["cohort"].(map[string]any)
	fingerprintHash := hashJSONFixture(evidence["execution_fingerprint"])
	manifest := map[string]any{
		"schema_version":               1,
		"id":                           cohort["id"],
		"target_tier":                  evidence["target_tier"],
		"threshold_revision":           evidence["threshold_revision"],
		"scope":                        evidence["scope"],
		"preregistered_at":             cohort["preregistered_at"],
		"first_attempt_at":             cohort["first_attempt_at"],
		"provenance_kind":              cohort["provenance_kind"],
		"expected_attempt_ids":         cohort["expected_attempt_ids"],
		"families":                     cohort["families"],
		"execution_fingerprint_sha256": fingerprintHash,
	}
	manifestRelative := cohort["manifest_path"].(string)
	manifestPath := filepath.Join(root, filepath.FromSlash(manifestRelative))
	writeJSONFixture(t, manifestPath, manifest)
	manifestContent, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read cohort manifest: %v", err)
	}
	manifestHash := hashBytesFixture(manifestContent)
	cohort["manifest_sha256"] = manifestHash
	privateKey := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{7}, ed25519.SeedSize))
	publicKey := privateKey.Public().(ed25519.PublicKey)
	cohort["manifest_signature_ed25519"] = hex.EncodeToString(ed25519.Sign(privateKey, manifestContent))
	writeJSONFixture(t, filepath.Join(root, "config", "model-tier-trust.json"), map[string]any{
		"schema_version": 1,
		"signers": []any{map[string]any{
			"id":                 cohort["manifest_signer_id"],
			"public_key_ed25519": hex.EncodeToString(publicKey),
		}},
	})

	for _, raw := range evidence["attempts"].([]any) {
		attempt := raw.(map[string]any)
		id := attempt["id"].(string)
		base := filepath.ToSlash(filepath.Join("artifacts", id))
		contents := map[string]string{
			"fixture":              "fixture-" + id,
			"frozen_plan":          "plan-" + id,
			"request":              "request-" + id,
			"response":             "response-" + id,
			"plan_sufficiency":     "sufficiency-" + id,
			"executor_package":     "executor",
			"route_execution":      "route-" + id,
			"oracle_qualification": "oracle",
			"proof_result":         "proof-" + id,
		}
		paths := map[string]any{"cohort_manifest": manifestRelative}
		hashes := map[string]any{"cohort_manifest": manifestHash}
		for name, content := range contents {
			relative := filepath.ToSlash(filepath.Join(base, name+".txt"))
			path := filepath.Join(root, filepath.FromSlash(relative))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatalf("create artifact directory: %v", err)
			}
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				t.Fatalf("write artifact: %v", err)
			}
			paths[name] = relative
			if name != "request" && name != "response" {
				hashes[name] = hashStringFixture(content)
			}
		}
		attempt["artifact_paths"] = paths
		attempt["artifact_hashes"] = hashes
		attempt["plan_sha256"] = hashes["frozen_plan"]
		attempt["request_sha256"] = hashStringFixture(contents["request"])
		attempt["response_sha256"] = hashStringFixture(contents["response"])
		attempt["proof_sha256"] = hashes["proof_result"]
	}
}

func refreshModelTierEvalFingerprint(evidence map[string]any) {
	hash := hashJSONFixture(evidence["execution_fingerprint"])
	evidence["current_execution_fingerprint_sha256"] = hash
	for _, raw := range evidence["attempts"].([]any) {
		raw.(map[string]any)["execution_fingerprint_sha256"] = hash
	}
}

func hashJSONFixture(value any) string {
	content, _ := json.Marshal(value)
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func hashStringFixture(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hashBytesFixture(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func writeJSONFixture(t *testing.T, path string, value any) {
	t.Helper()
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create fixture directory: %v", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}
