package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	proofGovernorSchemaVersion = 1
	proofGovernorRun           = "run"
	proofGovernorReuse         = "reuse"
	proofGovernorBlock         = "block"
	maxProofGovernorBytes      = 1 << 20
)

type proofGovernorNamespace struct {
	Goal  string `json:"goal"`
	Slice string `json:"slice"`
	Run   string `json:"run"`
}

type proofGovernorCheckSpec struct {
	ID             string                  `json:"id"`
	Namespace      proofGovernorNamespace  `json:"namespace"`
	Command        []string                `json:"command"`
	Covers         []string                `json:"covers"`
	Inputs         []string                `json:"inputs"`
	OracleFiles    []string                `json:"oracle_files,omitempty"`
	WorkingDir     string                  `json:"working_dir"`
	TimeoutMS      int                     `json:"timeout_ms"`
	ExpectedExit   int                     `json:"expected_exit"`
	Environment    map[string]string       `json:"environment,omitempty"`
	ExecutionClass string                  `json:"execution_class"`
	MaxAgeSeconds  int                     `json:"max_age_seconds"`
	External       []proofGovernorExternal `json:"external_evidence,omitempty"`
}

type proofGovernorExternal struct {
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

type proofGovernorInput struct {
	Path    string `json:"path"`
	Role    string `json:"role"`
	SHA256  string `json:"sha256,omitempty"`
	Missing bool   `json:"missing,omitempty"`
}

type proofGovernorExecutionResult struct {
	Status      string    `json:"status"`
	ExitCode    int       `json:"exit_code"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at"`
	DurationMS  int64     `json:"duration_ms"`
}

type proofGovernorReceipt struct {
	SchemaVersion        int                          `json:"schema_version"`
	ReceiptID            string                       `json:"receipt_id"`
	Spec                 proofGovernorCheckSpec       `json:"spec"`
	RegistryPath         string                       `json:"registry_path"`
	RegistrySHA256       string                       `json:"registry_sha256"`
	CheckSemanticsSHA256 string                       `json:"check_semantics_sha256"`
	RelevantInputsSHA256 string                       `json:"relevant_inputs_sha256"`
	EnvironmentSHA256    string                       `json:"environment_sha256"`
	CoveredChecks        []string                     `json:"covered_checks"`
	Inputs               []proofGovernorInput         `json:"inputs"`
	Result               proofGovernorExecutionResult `json:"result"`
	IntegritySHA256      string                       `json:"integrity_sha256"`
}

type proofGovernorDecision struct {
	Decision  string   `json:"decision"`
	ReceiptID string   `json:"receipt_id,omitempty"`
	RunChecks []string `json:"run_checks,omitempty"`
	Reused    []string `json:"reused_checks,omitempty"`
	Reasons   []string `json:"reasons"`
}

func runProofGovernorReceiptValidateCommand(root string, opts options, stdout, stderr io.Writer) int {
	path := resolveInputPath(root, opts.receiptPath)
	issues := validateProofGovernorReceiptFile(root, path, time.Now().UTC())
	result := map[string]any{"ok": len(issues) == 0, "receipt": filepath.ToSlash(opts.receiptPath), "issues": issues}
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	} else if len(issues) == 0 {
		fmt.Fprintf(stdout, "proof receipt: PASS %s\n", filepath.ToSlash(opts.receiptPath))
	} else {
		for _, issue := range issues {
			fmt.Fprintln(stderr, issue)
		}
	}
	if len(issues) > 0 {
		return 2
	}
	return 0
}

func captureProofGovernorReceipt(root string, spec proofGovernorCheckSpec, registryPath string, result proofGovernorExecutionResult) (proofGovernorReceipt, error) {
	normalized, issues := normalizeProofGovernorSpec(spec)
	if len(issues) > 0 {
		return proofGovernorReceipt{}, fmt.Errorf("invalid proof spec: %s", strings.Join(issues, "; "))
	}
	registryRelative, registryHash, err := hashProofGovernorFile(root, registryPath)
	if err != nil {
		return proofGovernorReceipt{}, fmt.Errorf("registry: %w", err)
	}
	inputs, inputHash, inputIssues := fingerprintProofGovernorInputs(root, normalized)
	if len(inputIssues) > 0 {
		return proofGovernorReceipt{}, fmt.Errorf("fingerprint inputs: %s", strings.Join(inputIssues, "; "))
	}
	semanticsHash, err := hashProofGovernorJSON(normalized)
	if err != nil {
		return proofGovernorReceipt{}, err
	}
	environmentHash, err := hashProofGovernorJSON(normalized.Environment)
	if err != nil {
		return proofGovernorReceipt{}, err
	}
	if result.DurationMS == 0 && !result.StartedAt.IsZero() && !result.CompletedAt.IsZero() {
		result.DurationMS = result.CompletedAt.Sub(result.StartedAt).Milliseconds()
	}
	receipt := proofGovernorReceipt{
		SchemaVersion:        proofGovernorSchemaVersion,
		Spec:                 normalized,
		RegistryPath:         registryRelative,
		RegistrySHA256:       registryHash,
		CheckSemanticsSHA256: semanticsHash,
		RelevantInputsSHA256: inputHash,
		EnvironmentSHA256:    environmentHash,
		CoveredChecks:        effectiveProofGovernorCoverage(normalized),
		Inputs:               inputs,
		Result:               result,
	}
	integrity, err := proofGovernorReceiptIntegrity(receipt)
	if err != nil {
		return proofGovernorReceipt{}, err
	}
	receipt.ReceiptID = integrity
	receipt.IntegritySHA256 = integrity
	return receipt, nil
}

func evaluateProofGovernorReceipt(root string, spec proofGovernorCheckSpec, requested []string, registryPath string, receipt proofGovernorReceipt, now time.Time) proofGovernorDecision {
	decision := proofGovernorDecision{Decision: proofGovernorReuse, ReceiptID: receipt.ReceiptID}
	addReason := func(reason string) {
		for _, existing := range decision.Reasons {
			if existing == reason {
				return
			}
		}
		decision.Reasons = append(decision.Reasons, reason)
	}

	if receipt.SchemaVersion != proofGovernorSchemaVersion {
		addReason("unsupported-receipt-schema")
	}
	if receipt.Result.Status != "pass" || receipt.Result.ExitCode != receipt.Spec.ExpectedExit {
		addReason("receipt-not-passing")
	}
	integrity, err := proofGovernorReceiptIntegrity(receipt)
	if err != nil || receipt.IntegritySHA256 == "" || receipt.ReceiptID != receipt.IntegritySHA256 || receipt.IntegritySHA256 != integrity {
		addReason("receipt-integrity-mismatch")
	}
	normalized, specIssues := normalizeProofGovernorSpec(spec)
	if len(specIssues) > 0 {
		for _, issue := range specIssues {
			addReason("invalid-current-spec:" + issue)
		}
	} else {
		semanticsHash, hashErr := hashProofGovernorJSON(normalized)
		if hashErr != nil || semanticsHash != receipt.CheckSemanticsSHA256 {
			addReason("check-semantics-changed")
		}
		environmentHash, environmentErr := hashProofGovernorJSON(normalized.Environment)
		if environmentErr != nil || environmentHash != receipt.EnvironmentSHA256 {
			addReason("environment-changed")
		}
	}

	requested = normalizeProofGovernorIDs(requested)
	coverage := stringSet(receipt.CoveredChecks)
	for _, id := range requested {
		if !coverage[id] {
			addReason("coverage-miss:" + id)
		}
	}

	registryRelative, registryHash, registryErr := hashProofGovernorFile(root, registryPath)
	if registryErr != nil {
		addReason("registry-unavailable")
	} else if registryRelative != receipt.RegistryPath || registryHash != receipt.RegistrySHA256 {
		addReason("registry-changed")
	}

	currentInputs, currentInputHash, inputIssues := fingerprintProofGovernorInputs(root, normalized)
	for _, issue := range inputIssues {
		addReason(issue)
	}
	currentByKey := map[string]proofGovernorInput{}
	for _, input := range currentInputs {
		currentByKey[input.Role+"\x00"+input.Path] = input
	}
	for _, previous := range receipt.Inputs {
		current, ok := currentByKey[previous.Role+"\x00"+previous.Path]
		switch {
		case !ok:
			addReason(previous.Role + "-input-removed:" + previous.Path)
		case current.Missing:
			addReason(previous.Role + "-input-missing:" + previous.Path)
		case previous.Missing:
			addReason(previous.Role + "-input-was-missing:" + previous.Path)
		case current.SHA256 != previous.SHA256:
			addReason(previous.Role + "-input-changed:" + previous.Path)
		}
	}
	if currentInputHash != "" && currentInputHash != receipt.RelevantInputsSHA256 && len(decision.Reasons) == 0 {
		addReason("relevant-input-set-changed")
	}

	maxAge := time.Duration(normalized.MaxAgeSeconds) * time.Second
	if maxAge <= 0 || receipt.Result.CompletedAt.IsZero() || now.After(receipt.Result.CompletedAt.Add(maxAge)) {
		addReason("receipt-expired")
	}

	if len(decision.Reasons) > 0 {
		decision.Decision = proofGovernorRun
		decision.RunChecks = requested
		decision.ReceiptID = receipt.ReceiptID
		return decision
	}
	decision.Reused = requested
	decision.Reasons = []string{"passing-superset-receipt-with-identical-relevant-inputs"}
	return decision
}

func normalizeProofGovernorSpec(spec proofGovernorCheckSpec) (proofGovernorCheckSpec, []string) {
	issues := []string{}
	spec.ID = strings.TrimSpace(spec.ID)
	spec.Namespace.Goal = strings.TrimSpace(spec.Namespace.Goal)
	spec.Namespace.Slice = strings.TrimSpace(spec.Namespace.Slice)
	spec.Namespace.Run = strings.TrimSpace(spec.Namespace.Run)
	spec.WorkingDir = filepath.ToSlash(filepath.Clean(strings.TrimSpace(spec.WorkingDir)))
	spec.ExecutionClass = strings.ToLower(strings.TrimSpace(spec.ExecutionClass))
	spec.Command = trimNonEmptyProofGovernorStrings(spec.Command)
	spec.Covers = normalizeProofGovernorIDs(spec.Covers)
	spec.Inputs = normalizeProofGovernorPaths(spec.Inputs)
	spec.OracleFiles = normalizeProofGovernorPaths(spec.OracleFiles)
	sort.Slice(spec.External, func(i, j int) bool {
		if spec.External[i].ID == spec.External[j].ID {
			return spec.External[i].Digest < spec.External[j].Digest
		}
		return spec.External[i].ID < spec.External[j].ID
	})

	require := func(ok bool, message string) {
		if !ok {
			issues = append(issues, message)
		}
	}
	require(spec.ID != "", "id is required")
	require(spec.Namespace.Goal != "" && spec.Namespace.Slice != "" && spec.Namespace.Run != "", "goal slice and run namespace are required")
	require(len(spec.Command) > 0, "command is required")
	require(spec.WorkingDir != "" && boundedProofGovernorPath(spec.WorkingDir), "working_dir must be a bounded relative path")
	require(spec.TimeoutMS > 0, "timeout_ms must be positive")
	require(spec.MaxAgeSeconds > 0, "max_age_seconds must be positive")
	require(len(spec.Inputs) > 0, "inputs must not be empty")
	require(validProofGovernorExecutionClass(spec.ExecutionClass), "execution_class is invalid")
	for _, path := range append(append([]string{}, spec.Inputs...), spec.OracleFiles...) {
		require(boundedProofGovernorPath(path), "input paths must be bounded relative files")
	}
	for _, external := range spec.External {
		require(strings.TrimSpace(external.ID) != "" && validProofGovernorDigest(external.Digest), "external evidence requires id and sha256 digest")
	}
	return spec, issues
}

func fingerprintProofGovernorInputs(root string, spec proofGovernorCheckSpec) ([]proofGovernorInput, string, []string) {
	inputs := make([]proofGovernorInput, 0, len(spec.Inputs)+len(spec.OracleFiles))
	issues := []string{}
	add := func(role string, paths []string) {
		for _, relative := range paths {
			entry := proofGovernorInput{Path: relative, Role: role}
			path, err := resolveProofGovernorPath(root, relative, true)
			if err != nil {
				entry.Missing = true
				issues = append(issues, role+"-input-invalid:"+relative)
				inputs = append(inputs, entry)
				continue
			}
			info, err := os.Stat(path)
			if os.IsNotExist(err) {
				entry.Missing = true
				issues = append(issues, role+"-input-missing:"+relative)
				inputs = append(inputs, entry)
				continue
			}
			if err != nil || info.IsDir() {
				entry.Missing = true
				issues = append(issues, role+"-input-unreadable:"+relative)
				inputs = append(inputs, entry)
				continue
			}
			entry.SHA256, err = sha256File(path)
			if err != nil {
				entry.Missing = true
				issues = append(issues, role+"-input-unreadable:"+relative)
			}
			inputs = append(inputs, entry)
		}
	}
	add("relevant", spec.Inputs)
	add("oracle", spec.OracleFiles)
	sort.Slice(inputs, func(i, j int) bool {
		if inputs[i].Role == inputs[j].Role {
			return inputs[i].Path < inputs[j].Path
		}
		return inputs[i].Role < inputs[j].Role
	})
	hash, err := hashProofGovernorJSON(inputs)
	if err != nil {
		issues = append(issues, "input-fingerprint-error")
		return inputs, "", issues
	}
	return inputs, hash, issues
}

func hashProofGovernorFile(root, relative string) (string, string, error) {
	relative = filepath.ToSlash(filepath.Clean(strings.TrimSpace(relative)))
	path, err := resolveProofGovernorPath(root, relative, false)
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", "", err
	}
	if info.IsDir() {
		return "", "", fmt.Errorf("path is a directory")
	}
	hash, err := sha256File(path)
	return relative, hash, err
}

func resolveProofGovernorPath(root, relative string, allowMissing bool) (string, error) {
	if !boundedProofGovernorPath(relative) {
		return "", fmt.Errorf("path must be bounded and relative: %s", relative)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rootCanonical := rootAbs
	if evaluated, evalErr := filepath.EvalSymlinks(rootAbs); evalErr == nil {
		rootCanonical = evaluated
	}
	candidate := filepath.Join(rootAbs, filepath.FromSlash(relative))
	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	if evaluated, evalErr := filepath.EvalSymlinks(candidateAbs); evalErr == nil {
		candidateAbs = evaluated
	} else if !allowMissing && !os.IsNotExist(evalErr) {
		return "", evalErr
	}
	rel, err := filepath.Rel(rootCanonical, candidateAbs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes project root: %s", relative)
	}
	return candidateAbs, nil
}

func boundedProofGovernorPath(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || filepath.IsAbs(value) || strings.ContainsAny(value, "*?[]") {
		return false
	}
	clean := filepath.Clean(filepath.FromSlash(value))
	return clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator))
}

func proofGovernorReceiptIntegrity(receipt proofGovernorReceipt) (string, error) {
	copyReceipt := receipt
	copyReceipt.ReceiptID = ""
	copyReceipt.IntegritySHA256 = ""
	return hashProofGovernorJSON(copyReceipt)
}

func writeProofGovernorReceipt(path string, receipt proofGovernorReceipt) error {
	integrity, err := proofGovernorReceiptIntegrity(receipt)
	if err != nil {
		return err
	}
	receipt.ReceiptID = integrity
	receipt.IntegritySHA256 = integrity
	content, err := json.MarshalIndent(receipt, "", "  ")
	if err != nil {
		return err
	}
	if len(content) > maxProofGovernorBytes {
		return fmt.Errorf("proof receipt exceeds size limit")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".proof-receipt-*.tmp")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	defer os.Remove(tempName)
	if _, err := temp.Write(append(content, '\n')); err != nil {
		_ = temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempName, path)
}

func validateProofGovernorReceiptFile(root, path string, now time.Time) []string {
	content, err := os.ReadFile(path)
	if err != nil {
		return []string{"receipt-read-error"}
	}
	if len(content) > maxProofGovernorBytes {
		return []string{"receipt-size-limit"}
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var receipt proofGovernorReceipt
	if err := decoder.Decode(&receipt); err != nil {
		return []string{"receipt-parse-error"}
	}
	decision := evaluateProofGovernorReceipt(root, receipt.Spec, receipt.CoveredChecks, receipt.RegistryPath, receipt, now)
	if decision.Decision != proofGovernorReuse {
		return decision.Reasons
	}
	return nil
}

func hashProofGovernorJSON(value any) (string, error) {
	content, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func effectiveProofGovernorCoverage(spec proofGovernorCheckSpec) []string {
	return normalizeProofGovernorIDs(append(append([]string{}, spec.Covers...), spec.ID))
}

func normalizeProofGovernorIDs(values []string) []string {
	seen := map[string]bool{}
	normalized := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		normalized = append(normalized, value)
	}
	sort.Strings(normalized)
	return normalized
}

func normalizeProofGovernorPaths(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = filepath.ToSlash(filepath.Clean(strings.TrimSpace(value)))
		if runtime.GOOS == "windows" {
			value = strings.ToLower(value)
		}
		if value != "" && value != "." {
			normalized = append(normalized, value)
		}
	}
	sort.Strings(normalized)
	return dedupeProofGovernorStrings(normalized)
}

func trimNonEmptyProofGovernorStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return result
}

func dedupeProofGovernorStrings(values []string) []string {
	if len(values) == 0 {
		return values
	}
	result := []string{values[0]}
	for _, value := range values[1:] {
		if value != result[len(result)-1] {
			result = append(result, value)
		}
	}
	return result
}

func stringSet(values []string) map[string]bool {
	result := map[string]bool{}
	for _, value := range values {
		result[value] = true
	}
	return result
}

func validProofGovernorExecutionClass(value string) bool {
	switch value {
	case "cli", "headless-browser", "visible-browser", "native-gui":
		return true
	default:
		return false
	}
}

func validProofGovernorDigest(value string) bool {
	value = strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "sha256:")
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
