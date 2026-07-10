package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/modelrouting"
)

func TestNormalizeExecutionTelemetryDowngradesForgedFieldsWithoutReceipt(t *testing.T) {
	usage, err := normalizeExecutionTelemetry(map[string]any{
		"packet_id": "p1", "run_id": "run-1", "project_id": "project-a", "slice_id": "slice-1", "context_packet_hash": "sha256:packet", "session_id": "forged-session",
		"runtime": "ghcp", "requested_route": "medium-a", "actual_route": "forged-route",
		"requested_model": "model-a", "actual_model": "forged-model", "receipt_status": "credited",
		"predicted_tier": "small", "actual_tier": "small", "turns": 2, "input_tokens": 100,
		"output_tokens": 20, "cache_read_tokens": 50, "rework_count": 1,
		"proof_result": "pass", "packet_sufficiency": "sufficient",
		"effective_token_model": "raw-v1",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if usage.InputTokens != 100 || usage.ActualRoute != "" || usage.ActualModel != "" || usage.SessionID != "" || usage.ReceiptStatus != "missing" || usage.ProofResult != "unknown" {
		t.Fatalf("usage=%#v", usage)
	}
}

func TestNormalizeExecutionTelemetryWithCreditedReceiptOverridesForgedAssertions(t *testing.T) {
	evidence, err := evaluateExecutionTelemetryEvidence(exactExecutionTelemetryReceipt(time.Now().UTC()))
	if err != nil || evidence == nil {
		t.Fatalf("evidence=%#v err=%v", evidence, err)
	}
	usage, err := normalizeExecutionTelemetry(map[string]any{
		"packet_id": "p1", "run_id": "run-1", "project_id": "project-a", "slice_id": "slice-1", "context_packet_hash": "sha256:packet", "session_id": "forged-session",
		"runtime": "ghcp", "requested_route": "medium-a", "actual_route": "forged-route",
		"requested_model": "model-a", "actual_model": "forged-model", "receipt_status": "credited",
		"predicted_tier": "small", "actual_tier": "small", "turns": 2, "input_tokens": 100,
		"proof_result": "unknown", "packet_sufficiency": "sufficient",
		"effective_token_model": "raw-v1",
	}, evidence)
	if err != nil {
		t.Fatal(err)
	}
	if usage.ActualRoute != "medium-a" || usage.ActualModel != "model-a" || usage.SessionID != "session-1" || usage.ReceiptStatus != "credited" || usage.ProofResult != "pass" {
		t.Fatalf("usage=%#v", usage)
	}
}

func TestNormalizeExecutionTelemetryObservationOnlyForProofUnknown(t *testing.T) {
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	receipt.WorkProof.Result = modelrouting.ProofUnknown
	evidence, err := evaluateExecutionTelemetryEvidence(receipt, envelope)
	if err != nil || evidence == nil {
		t.Fatalf("evidence=%#v err=%v", evidence, err)
	}
	if evidence.ReceiptStatus != "observation-only" || evidence.ProofResult != "unknown" || evidence.ActualRoute != "medium-a" {
		t.Fatalf("evidence=%#v", evidence)
	}
}

func TestNormalizeExecutionTelemetryDoesNotTreatLegacyModelAsActual(t *testing.T) {
	usage, err := normalizeExecutionTelemetry(map[string]any{
		"packet_id": "p1", "runtime": "codex", "model": "legacy-advisory-model",
		"requested_route": "codex.large", "receipt_status": "unavailable",
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if usage.RequestedModel != "legacy-advisory-model" {
		t.Fatalf("legacy model should remain requested/advisory: %#v", usage)
	}
	if usage.ActualModel != "" {
		t.Fatalf("legacy model was copied into actual_model without receipt evidence: %#v", usage)
	}
}

func TestNormalizeExecutionTelemetryRejectsInvalidCounters(t *testing.T) {
	if _, err := normalizeExecutionTelemetry(map[string]any{"input_tokens": -1}, nil); err == nil {
		t.Fatal("negative telemetry passed")
	}
}

func TestNormalizeExecutionTelemetryRequiresPacketID(t *testing.T) {
	if _, err := normalizeExecutionTelemetry(map[string]any{"input_tokens": 1}, nil); err == nil {
		t.Fatal("telemetry without packet_id passed")
	}
}

func TestExecutionTelemetryCommandUsesReceiptEnvelopeEvidence(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	attestationRoot := filepath.Join(t.TempDir(), "host-attestations")
	useTestExecutionAttestationRoot(t, attestationRoot)
	projectID := canonicalTelemetryProjectID(t, root)
	writeExecutionTelemetryFixture(t, root, "telemetry.json", map[string]any{
		"packet_id": "p1", "run_id": "run-1", "project_id": projectID, "slice_id": "slice-1", "context_packet_hash": "sha256:packet", "session_id": "forged-session",
		"runtime": "ghcp", "requested_route": "medium-a", "actual_route": "forged-route",
		"requested_model": "model-a", "actual_model": "forged-model", "receipt_status": "credited",
		"predicted_tier": "medium", "actual_tier": "medium", "turns": 2, "input_tokens": 100,
		"proof_result": "pass", "packet_sufficiency": "sufficient", "effective_token_model": "raw-v1",
	})
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	receipt.WorkProof.Command = "kbrouter dispatch"
	receipt.WorkProof.Result = modelrouting.ProofUnknown
	envelope.ProjectID = projectID
	receipt.RouteEvidence.ProjectID = projectID
	receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))
	recordDispatchAttestationForTelemetryTest(t, attestationRoot, root, paths.receipt)
	var out, errOut strings.Builder
	code := run([]string{
		"execution-telemetry",
		"--root", root,
		"--telemetry", "telemetry.json",
		"--receipt", paths.receipt,
		"--evidence-envelope", paths.envelope,
		"--json",
	}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), `"receipt_status": "observation-only"`) || !strings.Contains(out.String(), `"actual_route": "medium-a"`) || !strings.Contains(out.String(), `"actual_model": "model-a"`) || !strings.Contains(out.String(), `"session_id": "session-1"`) {
		t.Fatalf("stdout=%s", out.String())
	}
	if strings.Contains(out.String(), "forged-route") || strings.Contains(out.String(), "forged-model") || strings.Contains(out.String(), "forged-session") {
		t.Fatalf("forged assertion survived: %s", out.String())
	}
	out.Reset()
	errOut.Reset()
	code = run([]string{
		"execution-telemetry", "--root", root, "--telemetry", "telemetry.json",
		"--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json",
	}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"receipt_status": "observation-only"`) {
		t.Fatalf("saved host attestation was not reusable: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	runRoot := filepath.Join(root, filepath.FromSlash(filepath.Dir(paths.receipt)))
	var tampered modelrouting.RoutingReceipt
	if err := modelrouting.LoadStrictJSON(runRoot, filepath.Base(paths.receipt), &tampered, maxExecutionTelemetryBytes); err != nil {
		t.Fatal(err)
	}
	tampered.RouteEvidence.SessionID = "forged-session"
	if err := modelrouting.SaveAtomicJSON(runRoot, filepath.Base(paths.receipt), tampered, maxExecutionTelemetryBytes); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	errOut.Reset()
	code = run([]string{
		"execution-telemetry", "--root", root, "--telemetry", "telemetry.json",
		"--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json",
	}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"receipt_status": "mismatch"`) {
		t.Fatalf("tampered host attestation was accepted: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func recordDispatchAttestationForTelemetryTest(t *testing.T, attestationRoot, projectRoot, receiptPath string) {
	t.Helper()
	runRoot := filepath.Join(projectRoot, filepath.FromSlash(filepath.Dir(receiptPath)))
	var receipt modelrouting.RoutingReceipt
	receiptBytes, err := modelrouting.LoadStrictJSONBytes(runRoot, filepath.Base(receiptPath), &receipt, maxExecutionTelemetryBytes)
	if err != nil {
		t.Fatal(err)
	}
	if err := modelrouting.RecordDispatchReceiptAttestation(attestationRoot, receipt, modelrouting.SHA256Bytes(receiptBytes), time.Now().UTC()); err != nil {
		skipIfTelemetryPrivateACLUnsupported(t, err)
		t.Fatal(err)
	}
}

func TestExecutionTelemetryMatchingProjectEvidenceNeedsHostAttestation(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	useTestExecutionAttestationRoot(t, filepath.Join(t.TempDir(), "empty-host-attestations"))
	projectID := canonicalTelemetryProjectID(t, root)
	writeExecutionTelemetryFixture(t, root, "telemetry.json", exactTelemetryPayload(projectID))
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	envelope.ProjectID = projectID
	receipt.RouteEvidence.ProjectID = projectID
	receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))

	var out, errOut strings.Builder
	code := run([]string{
		"execution-telemetry", "--root", root, "--telemetry", "telemetry.json",
		"--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json",
	}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"receipt_status": "mismatch"`) || strings.Contains(out.String(), `"actual_model"`) {
		t.Fatalf("matching project evidence credited without host attestation: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestDispatchAttestationRootRejectsSymlinkedProjectContainment(t *testing.T) {
	root := t.TempDir()
	physicalProject := filepath.Join(root, "physical-project")
	if err := os.MkdirAll(physicalProject, 0o755); err != nil {
		t.Fatal(err)
	}
	projectAlias := filepath.Join(root, "project-alias")
	if err := os.Symlink(physicalProject, projectAlias); err != nil {
		t.Skipf("directory symlinks unavailable: %v", err)
	}
	useTestExecutionAttestationRoot(t, filepath.Join(physicalProject, ".kb", "host-state"))
	receipt, _ := exactExecutionTelemetryReceipt(time.Now().UTC())
	receipt.WorkProof = modelrouting.WorkProof{Command: "kbrouter dispatch", ArtifactHash: "sha256:proof", Result: modelrouting.ProofUnknown}
	verified, err := verifyDispatchReceiptAttestation(projectAlias, receipt, modelrouting.SHA256Bytes([]byte("receipt")))
	if err == nil || verified || !strings.Contains(err.Error(), "must remain outside the project") {
		t.Fatalf("symlinked project containment bypassed: verified=%v err=%v", verified, err)
	}
}

func TestExecutionTelemetryCommandDowngradesMismatchedEnvelope(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	projectID := canonicalTelemetryProjectID(t, root)
	writeExecutionTelemetryFixture(t, root, "telemetry.json", map[string]any{
		"packet_id": "p1", "run_id": "run-1", "project_id": projectID, "slice_id": "slice-1", "context_packet_hash": "sha256:packet",
		"runtime": "ghcp", "requested_route": "medium-a",
		"actual_route": "forged-route", "requested_model": "model-a", "actual_model": "forged-model",
		"receipt_status": "credited", "proof_result": "pass", "effective_token_model": "raw-v1",
	})
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	envelope.ProjectID = projectID
	receipt.RouteEvidence.ProjectID = projectID
	envelope.RouteAlias = "other-route"
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))

	var out, errOut strings.Builder
	code := run([]string{
		"execution-telemetry",
		"--root", root,
		"--telemetry", "telemetry.json",
		"--receipt", paths.receipt,
		"--evidence-envelope", paths.envelope,
		"--json",
	}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
	if !strings.Contains(out.String(), `"receipt_status": "mismatch"`) || strings.Contains(out.String(), `"actual_route": "medium-a"`) {
		t.Fatalf("stdout=%s", out.String())
	}
}

func TestExecutionTelemetryRejectsForgedProjectEvidenceUnderPrivateRun(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	projectID := canonicalTelemetryProjectID(t, root)
	writeExecutionTelemetryFixture(t, root, "telemetry.json", exactTelemetryPayload(projectID))
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	envelope.ProjectID = "forged-project"
	receipt.RouteEvidence.ProjectID = "forged-project"
	receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))

	var out, errOut strings.Builder
	code := run([]string{"execution-telemetry", "--root", root, "--telemetry", "telemetry.json", "--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"receipt_status": "mismatch"`) {
		t.Fatalf("forged project evidence was credited or failed unclearly: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestExecutionTelemetryRejectsCrossRunAndPacketReplay(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	projectID := canonicalTelemetryProjectID(t, root)
	writeExecutionTelemetryFixture(t, root, "telemetry.json", exactTelemetryPayload(projectID))
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	envelope.ProjectID = projectID
	receipt.RouteEvidence.ProjectID = projectID
	envelope.RunID = "run-2"
	receipt.RouteEvidence.RunID = "run-2"
	envelope.ContextPacketHash = "sha256:other-packet"
	receipt.RouteEvidence.ContextPacketHash = "sha256:other-packet"
	receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))

	var out, errOut strings.Builder
	code := run([]string{"execution-telemetry", "--root", root, "--telemetry", "telemetry.json", "--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"receipt_status": "mismatch"`) {
		t.Fatalf("cross-run/hash replay was credited or failed unclearly: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestExecutionTelemetryRejectsOmittedRunBindingWithEvidence(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	projectID := canonicalTelemetryProjectID(t, root)
	payload := exactTelemetryPayload(projectID)
	delete(payload, "run_id")
	writeExecutionTelemetryFixture(t, root, "telemetry.json", payload)
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	envelope.ProjectID = projectID
	receipt.RouteEvidence.ProjectID = projectID
	receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))

	var out, errOut strings.Builder
	code := run([]string{"execution-telemetry", "--root", root, "--telemetry", "telemetry.json", "--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"receipt_status": "mismatch"`) {
		t.Fatalf("omitted run binding was credited or failed unclearly: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestExecutionTelemetryRejectsCallerMaxAgeExtensionByStrictSchema(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	projectID := canonicalTelemetryProjectID(t, root)
	writeExecutionTelemetryFixture(t, root, "telemetry.json", exactTelemetryPayload(projectID))
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC().Add(-2 * time.Hour))
	envelope.ProjectID = projectID
	receipt.RouteEvidence.ProjectID = projectID
	receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))
	envelopePath := filepath.Join(root, paths.envelope)
	var raw map[string]any
	if err := json.Unmarshal(readFileForTelemetryTest(t, envelopePath), &raw); err != nil {
		t.Fatal(err)
	}
	raw["max_age"] = "24h"
	content, _ := json.MarshalIndent(raw, "", "  ")
	if err := os.WriteFile(envelopePath, append(content, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	code := run([]string{"execution-telemetry", "--root", root, "--telemetry", "telemetry.json", "--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json"}, &out, &errOut)
	if code != 2 || !strings.Contains(out.String(), "max_age") {
		t.Fatalf("caller max_age extension was not rejected by strict schema: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func TestExecutionTelemetryRejectsProofArtifactMismatch(t *testing.T) {
	root := privateTelemetryProjectRoot(t)
	projectID := canonicalTelemetryProjectID(t, root)
	writeExecutionTelemetryFixture(t, root, "telemetry.json", exactTelemetryPayload(projectID))
	receipt, envelope := exactExecutionTelemetryReceipt(time.Now().UTC())
	envelope.ProjectID = projectID
	receipt.RouteEvidence.ProjectID = projectID
	receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	paths := writePrivateExecutionEvidenceFiles(t, root, receipt, envelope, []byte(`{"ok":true}`))
	if err := os.WriteFile(filepath.Join(root, paths.proof), []byte(`{"tampered":true}`), 0o600); err != nil {
		t.Fatal(err)
	}

	var out, errOut strings.Builder
	code := run([]string{"execution-telemetry", "--root", root, "--telemetry", "telemetry.json", "--receipt", paths.receipt, "--evidence-envelope", paths.envelope, "--json"}, &out, &errOut)
	if code != 0 || !strings.Contains(out.String(), `"receipt_status": "mismatch"`) {
		t.Fatalf("proof artifact mismatch was credited or failed unclearly: code=%d stdout=%s stderr=%s", code, out.String(), errOut.String())
	}
}

func writeExecutionTelemetryFixture(t *testing.T, root, name string, payload map[string]any) {
	t.Helper()
	content, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, name), append(content, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeExecutionReceiptFiles(t *testing.T, root string, receipt modelrouting.RoutingReceipt, envelope modelrouting.EvidenceEnvelope) {
	t.Helper()
	receiptBytes, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "receipt.json"), append(receiptBytes, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	fileEnvelope := executionTelemetryEnvelopeFile{
		RunID: envelope.RunID, SliceID: envelope.SliceID, ProjectID: envelope.ProjectID, RouteAlias: envelope.RouteAlias,
		RouteFingerprint: envelope.RouteFingerprint, Adapter: envelope.Adapter, AdapterRevision: envelope.AdapterRevision,
		DispatchMethod: envelope.DispatchMethod, ModelID: envelope.ModelID, TaskFamily: envelope.TaskFamily,
		ContextPacketID: envelope.ContextPacketID, ContextPacketHash: envelope.ContextPacketHash, ProofArtifactHash: envelope.ProofArtifactHash,
		ProofArtifactPath: "proof.json", Surface: envelope.Surface, Provider: envelope.Provider, Tools: envelope.Tools, ContextSize: envelope.ContextSize, Risk: string(envelope.Risk),
	}
	envelopeBytes, err := json.MarshalIndent(fileEnvelope, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "envelope.json"), append(envelopeBytes, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}

type privateEvidencePaths struct {
	receipt  string
	envelope string
	proof    string
}

func privateTelemetryProjectRoot(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "project")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func canonicalTelemetryProjectID(t *testing.T, root string) string {
	t.Helper()
	projectID, err := modelrouting.CanonicalProjectIdentity(root)
	if err != nil {
		t.Fatal(err)
	}
	return projectID
}

func exactTelemetryPayload(projectID string) map[string]any {
	return map[string]any{
		"packet_id": "p1", "run_id": "run-1", "project_id": projectID, "slice_id": "slice-1", "context_packet_hash": "sha256:packet",
		"runtime": "codex", "requested_route": "medium-a", "requested_model": "model-a",
		"receipt_status": "credited", "proof_result": "pass", "effective_token_model": "raw-v1",
	}
}

func writePrivateExecutionEvidenceFiles(t *testing.T, root string, receipt modelrouting.RoutingReceipt, envelope modelrouting.EvidenceEnvelope, proofContent []byte) privateEvidencePaths {
	t.Helper()
	runRoot := filepath.Join(root, ".kb", "runs", "run-1")
	marker := struct {
		SchemaVersion int    `json:"schema_version"`
		ProjectID     string `json:"project_id"`
		RunID         string `json:"run_id"`
	}{SchemaVersion: 1, ProjectID: canonicalTelemetryProjectID(t, root), RunID: "run-1"}
	if err := modelrouting.SaveAtomicJSON(runRoot, ".kb-run-root.json", marker, maxExecutionTelemetryBytes); err != nil {
		skipIfTelemetryPrivateACLUnsupported(t, err)
		t.Fatal(err)
	}
	proofName := "proof.json"
	if err := modelrouting.SaveAtomicJSON(runRoot, proofName, json.RawMessage(proofContent), maxExecutionTelemetryBytes); err != nil {
		skipIfTelemetryPrivateACLUnsupported(t, err)
		t.Fatal(err)
	}
	storedProof, err := os.ReadFile(filepath.Join(runRoot, proofName))
	if err != nil {
		t.Fatal(err)
	}
	proofHash := modelrouting.SHA256Bytes(storedProof)
	envelope.ProofArtifactHash = proofHash
	receipt.WorkProof.ArtifactHash = proofHash
	fileEnvelope := executionTelemetryEnvelopeFile{
		RunID: envelope.RunID, SliceID: envelope.SliceID, ProjectID: envelope.ProjectID, RouteAlias: envelope.RouteAlias,
		RouteFingerprint: envelope.RouteFingerprint, Adapter: envelope.Adapter, AdapterRevision: envelope.AdapterRevision,
		DispatchMethod: envelope.DispatchMethod, ModelID: envelope.ModelID, TaskFamily: envelope.TaskFamily,
		ContextPacketID: envelope.ContextPacketID, ContextPacketHash: envelope.ContextPacketHash, ProofArtifactHash: envelope.ProofArtifactHash,
		ProofArtifactPath: proofName, Surface: envelope.Surface, Provider: envelope.Provider, Tools: envelope.Tools,
		ContextSize: envelope.ContextSize, Risk: string(envelope.Risk),
	}
	if receipt.RouteEvidence.CapabilityEnvelopeHash == "" {
		receipt.RouteEvidence.CapabilityEnvelopeHash = mustCapabilityHash(t, envelope)
	}
	if err := modelrouting.SaveAtomicJSON(runRoot, "receipt.json", receipt, maxExecutionTelemetryBytes); err != nil {
		skipIfTelemetryPrivateACLUnsupported(t, err)
		t.Fatal(err)
	}
	if err := modelrouting.SaveAtomicJSON(runRoot, "envelope.json", fileEnvelope, maxExecutionTelemetryBytes); err != nil {
		skipIfTelemetryPrivateACLUnsupported(t, err)
		t.Fatal(err)
	}
	return privateEvidencePaths{
		receipt:  filepath.ToSlash(filepath.Join(".kb", "runs", "run-1", "receipt.json")),
		envelope: filepath.ToSlash(filepath.Join(".kb", "runs", "run-1", "envelope.json")),
		proof:    filepath.ToSlash(filepath.Join(".kb", "runs", "run-1", proofName)),
	}
}

func mustCapabilityHash(t *testing.T, envelope modelrouting.EvidenceEnvelope) string {
	t.Helper()
	hash, err := modelrouting.ComputeCapabilityEnvelopeHash(envelope)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func readFileForTelemetryTest(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func skipIfTelemetryPrivateACLUnsupported(t *testing.T, err error) {
	t.Helper()
	if runtime.GOOS == "windows" && (errors.Is(err, os.ErrPermission) || strings.Contains(err.Error(), "Access is denied")) {
		t.Skipf("workspace sandbox denies private Windows ACL setup: %v", err)
	}
}

func useTestExecutionAttestationRoot(t *testing.T, root string) {
	t.Helper()
	previous := executionAttestationRoot
	executionAttestationRoot = func() (string, error) { return root, nil }
	t.Cleanup(func() { executionAttestationRoot = previous })
}
