package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDisabledRunnerCannotStartAndTouchesNoProfilesOrSecrets(t *testing.T) {
	profileLoads := 0
	secretReads := 0
	runner := newDisabledRunner(func() { profileLoads++ }, func(string) string {
		secretReads++
		return "secret"
	})
	if _, err := runner.Run(context.Background(), runnerRequest{}); !errors.Is(err, errPaidRunnerDisabled) {
		t.Fatalf("err=%v", err)
	}
	if runner.Started() != 0 || profileLoads != 0 || secretReads != 0 {
		t.Fatalf("started=%d profiles=%d secrets=%d", runner.Started(), profileLoads, secretReads)
	}
}

func TestCreditBudgetStopsBeforeCorrection(t *testing.T) {
	budget := newCreditBudget(5, 8, 10)
	if err := budget.Reserve("attempt", 5); err != nil {
		t.Fatal(err)
	}
	if err := budget.Reserve("correction", 5); err == nil {
		t.Fatal("arm ceiling did not stop correction")
	}
	if budget.Reserved() != 5 {
		t.Fatalf("reserved=%d", budget.Reserved())
	}
}

func TestRuntimeModelResolutionKeepsRoutesOutOfConfig(t *testing.T) {
	routes := runtimeRouteCatalog{SchemaVersion: 1, Routes: []runtimeRoute{
		{ModelID: "runtime-model", Runner: "ghcp", Tier: "medium", Available: true},
		{ModelID: "recognized-model", Runner: "byok", Profile: "qwen-local", Tier: "small", Available: true},
	}}
	model, err := resolveRuntimeModel(config{}, routes, "runtime-model", "ghcp", "", "medium")
	if err != nil {
		t.Fatal(err)
	}
	if model.Model != "runtime-model" || model.Runner != "ghcp" || model.Tier != "medium" {
		t.Fatalf("model=%+v", model)
	}
	local, err := resolveRuntimeModel(config{}, routes, "recognized-model", "byok", "qwen-local", "small")
	if err != nil {
		t.Fatal(err)
	}
	if local.Profile != "qwen-local" || local.Runner != "byok" {
		t.Fatalf("local=%+v", local)
	}
	if _, err := resolveRuntimeModel(config{}, routes, "runtime-model", "ghcp", "", "small"); err == nil {
		t.Fatal("tier-mismatched runtime model passed")
	}
}

func TestDraftApplicationRejectsOutsideMutableAllowlistAndGitMetadata(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"allowed.go", "other.go"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("before"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".git", "config"), []byte("[core]"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := applyDraftResponseAllowed(root, `{"files":[{"path":"other.go","content":"after"}]}`, []string{"allowed.go"}); err == nil {
		t.Fatal("outside-allowlist write passed")
	}
	if err := applyDraftResponseAllowed(root, `{"files":[{"path":".git/config","content":"malicious"}]}`, []string{".git/config"}); err == nil {
		t.Fatal("Git metadata write passed")
	}
}

func TestInvalidPhaseNeverEligibleForCorrection(t *testing.T) {
	for _, phase := range []phaseResult{
		{Valid: false, Proof: proofResult{Passed: false}},
		{Valid: false, Proof: proofResult{Passed: true}},
		{Valid: true, Proof: proofResult{Passed: true}},
	} {
		if phaseCanCorrect(phase) {
			t.Fatalf("invalid or passing phase became correction-eligible: %+v", phase)
		}

	}
	if !phaseCanCorrect(phaseResult{Valid: true, Proof: proofResult{Passed: false}}) {
		t.Fatal("valid model failure was not correction-eligible")
	}
}

func TestRunRefusesPaidExecutionBeforeConfigProfileOrProviderAccess(t *testing.T) {
	err := runBenchmark([]string{
		"--mode", "direct", "--task", "task", "--model", "model",
		"--experiment-id", "experiment", "--routes", "missing.json",
	}, os.Stdout)
	if err == nil || err.Error() != "attended execution is disabled until a trusted human-approval verifier is implemented; use --dry-run" {
		t.Fatalf("err=%v", err)
	}
}
