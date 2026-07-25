package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestApprovalBindsExactMatrixAndDurableExperimentBudget(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.json")
	if err := os.WriteFile(configPath, []byte(`{"schema_version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	configHash, err := fileSHA256(configPath)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 12, 8, 0, 0, 0, time.UTC)
	expected := approvalReceipt{
		ExperimentID: "exp-1", Mode: "amr", TaskID: "retry", AttemptModel: "small",
		ContextContractHash: "context-hash",
		RouteCatalogHash:    "routes-hash",
		AttemptRunner:       "ghcp", DriverModel: "large", DriverRunner: "ghcp",
		Repeat: 2, MaxAICreditsPerCall: 5, MaxAICreditsPerArm: 10, MaxExperimentCredits: 40,
	}
	actual := expected
	actual.SchemaVersion = 1
	actual.ConfigSHA256 = configHash
	actual.ApprovedAt = now.Add(-time.Minute).Format(time.RFC3339)
	actual.ExpiresAt = now.Add(time.Hour).Format(time.RFC3339)
	approvalPath := filepath.Join(root, "approval.json")
	content, _ := json.Marshal(actual)
	if err := os.WriteFile(approvalPath, content, 0o600); err != nil {
		t.Fatal(err)
	}
	approval, hash, err := validateApproval(approvalPath, configPath, expected, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := reserveDurableBudget(root, approval, hash); err != nil {
		t.Fatal(err)
	}
	if err := reserveDurableBudget(root, approval, hash); err == nil {
		t.Fatal("identical matrix was reserved twice")
	}
}

func TestApprovalRejectsChangedMatrixBeforeBudgetReservation(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "config.json")
	if err := os.WriteFile(configPath, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	approvalPath := filepath.Join(root, "approval.json")
	if err := os.WriteFile(approvalPath, []byte(`{"schema_version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := validateApproval(approvalPath, configPath, approvalReceipt{ExperimentID: "exp"}, time.Now().UTC()); err == nil {
		t.Fatal("incomplete approval passed")
	}
}
