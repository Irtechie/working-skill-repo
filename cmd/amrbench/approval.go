package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var experimentIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type approvalReceipt struct {
	SchemaVersion        int    `json:"schema_version"`
	ExperimentID         string `json:"experiment_id"`
	ApprovedAt           string `json:"approved_at"`
	ExpiresAt            string `json:"expires_at"`
	ConfigSHA256         string `json:"config_sha256"`
	Mode                 string `json:"mode"`
	TaskID               string `json:"task_id"`
	ContextContractHash  string `json:"context_contract_hash"`
	RouteCatalogHash     string `json:"route_catalog_hash"`
	DirectModel          string `json:"direct_model,omitempty"`
	DirectRunner         string `json:"direct_runner,omitempty"`
	DirectProfile        string `json:"direct_profile,omitempty"`
	AttemptModel         string `json:"attempt_model,omitempty"`
	AttemptRunner        string `json:"attempt_runner,omitempty"`
	AttemptProfile       string `json:"attempt_profile,omitempty"`
	DriverModel          string `json:"driver_model,omitempty"`
	DriverRunner         string `json:"driver_runner,omitempty"`
	DriverProfile        string `json:"driver_profile,omitempty"`
	Repeat               int    `json:"repeat"`
	MaxAICreditsPerCall  int    `json:"max_ai_credits_per_call"`
	MaxAICreditsPerArm   int    `json:"max_ai_credits_per_arm"`
	MaxExperimentCredits int    `json:"max_experiment_ai_credits"`
}

type budgetLedger struct {
	SchemaVersion  int      `json:"schema_version"`
	ExperimentID   string   `json:"experiment_id"`
	ApprovalHashes []string `json:"approval_hashes"`
	Reserved       int64    `json:"reserved_ai_credits"`
	Maximum        int64    `json:"maximum_ai_credits"`
}

type runtimeRouteCatalog struct {
	SchemaVersion int            `json:"schema_version"`
	Routes        []runtimeRoute `json:"routes"`
}

type runtimeRoute struct {
	ModelID        string `json:"model_id"`
	Runner         string `json:"runner"`
	Profile        string `json:"profile,omitempty"`
	Tier           string `json:"tier"`
	Available      bool   `json:"available"`
	EvidencePath   string `json:"evidence_path"`
	EvidenceSHA256 string `json:"evidence_sha256"`
}

type routeProbeEvidence struct {
	SchemaVersion int    `json:"schema_version"`
	ModelID       string `json:"model_id"`
	Runner        string `json:"runner"`
	Profile       string `json:"profile,omitempty"`
	Tier          string `json:"tier"`
	Available     bool   `json:"available"`
	ObservedAt    string `json:"observed_at"`
	ExpiresAt     string `json:"expires_at"`
}

func validateApproval(path, configPath string, expected approvalReceipt, now time.Time) (approvalReceipt, string, error) {
	if strings.TrimSpace(path) == "" {
		return approvalReceipt{}, "", fmt.Errorf("paid execution requires --approval bound to the exact attended matrix")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return approvalReceipt{}, "", err
	}
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	var actual approvalReceipt
	if err := decoder.Decode(&actual); err != nil {
		return approvalReceipt{}, "", fmt.Errorf("decode approval receipt: %w", err)
	}
	configHash, err := fileSHA256(configPath)
	if err != nil {
		return approvalReceipt{}, "", err
	}
	expected.SchemaVersion = 1
	expected.ConfigSHA256 = configHash
	if actual.SchemaVersion != expected.SchemaVersion ||
		actual.ExperimentID != expected.ExperimentID ||
		actual.ConfigSHA256 != expected.ConfigSHA256 ||
		actual.Mode != expected.Mode || actual.TaskID != expected.TaskID ||
		actual.ContextContractHash != expected.ContextContractHash ||
		actual.RouteCatalogHash != expected.RouteCatalogHash ||
		actual.DirectModel != expected.DirectModel || actual.DirectRunner != expected.DirectRunner || actual.DirectProfile != expected.DirectProfile ||
		actual.AttemptModel != expected.AttemptModel || actual.AttemptRunner != expected.AttemptRunner || actual.AttemptProfile != expected.AttemptProfile ||
		actual.DriverModel != expected.DriverModel || actual.DriverRunner != expected.DriverRunner || actual.DriverProfile != expected.DriverProfile ||
		actual.Repeat != expected.Repeat ||
		actual.MaxAICreditsPerCall != expected.MaxAICreditsPerCall ||
		actual.MaxAICreditsPerArm != expected.MaxAICreditsPerArm ||
		actual.MaxExperimentCredits != expected.MaxExperimentCredits {
		return approvalReceipt{}, "", fmt.Errorf("approval receipt does not match the exact attended matrix")
	}

	approvedAt, err := time.Parse(time.RFC3339, actual.ApprovedAt)
	if err != nil || approvedAt.After(now) {
		return approvalReceipt{}, "", fmt.Errorf("approval approved_at is invalid")
	}

	expiresAt, err := time.Parse(time.RFC3339, actual.ExpiresAt)
	if err != nil || !expiresAt.After(now) || expiresAt.Sub(approvedAt) > 24*time.Hour {
		return approvalReceipt{}, "", fmt.Errorf("approval receipt is expired or exceeds 24 hours")
	}
	sum := sha256.Sum256(content)
	return actual, hex.EncodeToString(sum[:]), nil
}

func loadRuntimeRouteCatalog(path string) (runtimeRouteCatalog, string, error) {
	var catalog runtimeRouteCatalog
	if strings.TrimSpace(path) == "" {
		return catalog, "", fmt.Errorf("paid execution requires --routes with observed availability and tier evidence")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return catalog, "", err
	}
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return catalog, "", err
	}
	if catalog.SchemaVersion != 1 || len(catalog.Routes) == 0 {
		return catalog, "", fmt.Errorf("runtime route catalog is empty or unsupported")
	}
	seen := map[string]bool{}
	base := filepath.Dir(path)
	resolvedBase, err := filepath.EvalSymlinks(base)
	if err != nil {
		return catalog, "", err
	}
	for _, route := range catalog.Routes {
		key := route.ModelID + "\x00" + route.Runner + "\x00" + route.Profile
		if route.ModelID == "" || !stringIn(route.Runner, "ghcp", "byok") ||
			!stringIn(route.Tier, "small", "medium", "large") ||
			!route.Available || seen[key] {
			return catalog, "", fmt.Errorf("runtime route catalog contains an invalid or duplicate route")
		}

		evidencePath := filepath.Clean(filepath.FromSlash(route.EvidencePath))
		if evidencePath == "." || evidencePath == ".." || filepath.IsAbs(evidencePath) ||
			strings.HasPrefix(evidencePath, ".."+string(filepath.Separator)) {
			return catalog, "", fmt.Errorf("runtime route evidence path escapes catalog root")
		}
		evidenceFile := filepath.Join(base, evidencePath)
		info, err := os.Lstat(evidenceFile)
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return catalog, "", fmt.Errorf("runtime route evidence must be a regular non-symlink file")
		}
		evidenceHash, err := fileSHA256(evidenceFile)
		if err != nil || evidenceHash != route.EvidenceSHA256 {
			return catalog, "", fmt.Errorf("runtime route evidence hash mismatch")
		}
		resolvedEvidence, err := filepath.EvalSymlinks(evidenceFile)
		if err != nil {
			return catalog, "", err
		}
		relativeResolved, err := filepath.Rel(resolvedBase, resolvedEvidence)
		if err != nil || relativeResolved == ".." || strings.HasPrefix(relativeResolved, ".."+string(filepath.Separator)) {
			return catalog, "", fmt.Errorf("runtime route evidence escapes catalog root through a linked parent")
		}
		var evidence routeProbeEvidence
		evidenceContent, err := os.ReadFile(resolvedEvidence)
		if err != nil {
			return catalog, "", err
		}
		evidenceDecoder := json.NewDecoder(strings.NewReader(string(evidenceContent)))
		evidenceDecoder.DisallowUnknownFields()
		if err := evidenceDecoder.Decode(&evidence); err != nil {
			return catalog, "", fmt.Errorf("decode runtime route evidence: %w", err)
		}
		observedAt, observedErr := time.Parse(time.RFC3339, evidence.ObservedAt)
		expiresAt, expiresErr := time.Parse(time.RFC3339, evidence.ExpiresAt)
		now := time.Now().UTC()
		if evidence.SchemaVersion != 1 || evidence.ModelID != route.ModelID ||
			evidence.Runner != route.Runner || evidence.Profile != route.Profile ||
			evidence.Tier != route.Tier || evidence.Available != route.Available ||
			observedErr != nil || expiresErr != nil || observedAt.After(now) ||
			!expiresAt.After(now) || expiresAt.Sub(observedAt) > 24*time.Hour {
			return catalog, "", fmt.Errorf("runtime route evidence does not attest the catalog route or is stale")
		}
		seen[key] = true
	}
	sum := sha256.Sum256(content)
	return catalog, hex.EncodeToString(sum[:]), nil
}

func stringIn(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func reserveDurableBudget(root string, approval approvalReceipt, approvalHash string) error {
	if !experimentIDPattern.MatchString(approval.ExperimentID) {
		return fmt.Errorf("invalid experiment ID")
	}
	phases := int64(1)
	if approval.Mode == "amr" {
		phases = 2
	}
	perArm, err := checkedMultiply(phases, int64(approval.MaxAICreditsPerCall))
	if err != nil {
		return err
	}
	if perArm > int64(approval.MaxAICreditsPerArm) {
		return fmt.Errorf("requested arm reserves %d credits, above ceiling %d", perArm, approval.MaxAICreditsPerArm)
	}
	requested, err := checkedMultiply(perArm, int64(approval.Repeat))
	if err != nil {
		return err
	}
	experimentRoot := filepath.Join(root, "experiments", approval.ExperimentID)
	if err := os.MkdirAll(experimentRoot, 0o700); err != nil {
		return err
	}
	lockPath := filepath.Join(experimentRoot, "budget.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("experiment budget is locked or indeterminate: %w", err)
	}
	_, _ = fmt.Fprintf(lock, "pid=%d created=%s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
	_ = lock.Close()
	defer os.Remove(lockPath)

	ledgerPath := filepath.Join(experimentRoot, "budget.json")
	ledger := budgetLedger{
		SchemaVersion: 1, ExperimentID: approval.ExperimentID, ApprovalHashes: []string{approvalHash},
		Maximum: int64(approval.MaxExperimentCredits),
	}
	if content, readErr := os.ReadFile(ledgerPath); readErr == nil {
		decoder := json.NewDecoder(strings.NewReader(string(content)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&ledger); err != nil {
			return fmt.Errorf("decode durable budget ledger: %w", err)
		}
		if ledger.SchemaVersion != 1 || ledger.ExperimentID != approval.ExperimentID ||
			ledger.Maximum != int64(approval.MaxExperimentCredits) {
			return fmt.Errorf("durable budget ledger does not match approval")
		}
		if stringIn(approvalHash, ledger.ApprovalHashes...) {
			return fmt.Errorf("this exact approved matrix is already reserved; use the existing result or a new human approval")
		}
		ledger.ApprovalHashes = append(ledger.ApprovalHashes, approvalHash)
	} else if !os.IsNotExist(readErr) {
		return readErr
	}

	if requested < 0 || ledger.Reserved > ledger.Maximum-requested {
		return fmt.Errorf("experiment budget exhausted: reserved=%d requested=%d maximum=%d", ledger.Reserved, requested, ledger.Maximum)
	}
	ledger.Reserved += requested
	return writeAtomicJSONFile(ledgerPath, ledger)
}

func budgetLedgerRoot() (string, error) {
	account, err := user.Current()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(account.HomeDir) == "" {
		return "", fmt.Errorf("user home is unavailable")
	}
	return filepath.Join(account.HomeDir, ".kb", "amr-bench-budgets"), nil
}

func checkedMultiply(left, right int64) (int64, error) {
	if left < 0 || right < 0 || (left != 0 && right > (1<<63-1)/left) {
		return 0, fmt.Errorf("credit budget multiplication overflow")
	}
	return left * right, nil
}

func writeAtomicJSONFile(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	temp, err := os.CreateTemp(filepath.Dir(path), ".budget-*.tmp")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		_ = temp.Close()
		return err
	}
	if _, err := temp.Write(content); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func fileSHA256(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func hashStringMap(value map[string]string) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}
