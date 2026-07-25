package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDerivePairedSampleRejectsTamperedAndMismatchedRunArtifacts(t *testing.T) {
	root := t.TempDir()
	cfg := config{Tasks: []taskSpec{{
		ID: "task", Family: "family", PlannedTier: "medium", AttemptTier: "small",
	}}}
	direct := boundRunResult("direct")
	amr := boundRunResult("amr")
	directPath := writeRunResultFixture(t, root, "direct.json", direct)
	amrPath := writeRunResultFixture(t, root, "amr.json", amr)
	reference := pairedRunRef{
		SchemaVersion: 1, TaskID: "task", Seed: 1, Family: "family",
		Direct: pairedFileRef{Path: "direct.json", SHA256: mustFileHash(t, directPath)},
		AMR:    pairedFileRef{Path: "amr.json", SHA256: mustFileHash(t, amrPath)},
	}
	sample, err := derivePairedSample(root, cfg, reference)
	if err != nil || !sample.Direct.TelemetryComplete || !sample.AMR.RouteMatch {
		t.Fatalf("sample=%+v err=%v", sample, err)
	}
	reference.AMR.SHA256 = "forged"
	if _, err := derivePairedSample(root, cfg, reference); err == nil {
		t.Fatal("tampered run artifact passed")
	}
	reference.AMR.SHA256 = mustFileHash(t, amrPath)
	amr.ContextContractHash = "other-context"
	writeRunResultFixture(t, root, "amr.json", amr)
	reference.AMR.SHA256 = mustFileHash(t, amrPath)
	if _, err := derivePairedSample(root, cfg, reference); err == nil {
		t.Fatal("mismatched context passed")
	}
}

func boundRunResult(mode string) runResult {
	phase := phaseResult{
		Valid: true, ModelMatch: true, AICAvailable: true, AIUNano: 10, Calls: 1,
		Proof: proofResult{Passed: true, ExitCode: 0, SandboxImage: "image@sha256:digest"},
	}
	return runResult{
		SchemaVersion: 1, Mode: mode, TaskID: "task", TaskFamily: "family", Seed: 1,
		ContextContractHash: "context", ProofClosureHash: "proof", ExperimentID: "experiment",
		ApprovalHash: mode + "-approval", RouteCatalogHash: "routes", DriverModel: "driver",
		PlannedTier: "medium",
		AttemptTier: "small", FinalProof: phase.Proof, DurationMS: 10, Phases: []phaseResult{phase},
	}
}

func writeRunResultFixture(t *testing.T, root, name string, value runResult) string {
	t.Helper()
	path := filepath.Join(root, name)
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func mustFileHash(t *testing.T, path string) string {
	t.Helper()
	hash, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}
