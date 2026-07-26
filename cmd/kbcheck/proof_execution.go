package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type proofGovernorExecutionSummary struct {
	Decision     string   `json:"decision"`
	OK           bool     `json:"ok"`
	ExitCode     int      `json:"exit_code"`
	RunChecks    []string `json:"run_checks,omitempty"`
	Reused       []string `json:"reused_checks,omitempty"`
	ReceiptPaths []string `json:"receipt_paths,omitempty"`
	Reasons      []string `json:"reasons"`
}

type proofGovernorAuditRow struct {
	At        time.Time `json:"at"`
	Decision  string    `json:"decision"`
	Requested []string  `json:"requested"`
	RunChecks []string  `json:"run_checks,omitempty"`
	Reused    []string  `json:"reused_checks,omitempty"`
	ExitCode  int       `json:"exit_code"`
	Reasons   []string  `json:"reasons"`
}

const maxProofGovernorAuditBytes = 8 << 20

func runProofGovernorExecuteCommand(root string, opts options, stdout, stderr io.Writer) int {
	registryPath := resolveInputPath(root, opts.proofRegistryPath)
	registry, err := loadProofGovernorRegistry(registryPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	registryRelative, err := filepath.Rel(root, registryPath)
	if err != nil || !boundedProofGovernorPath(registryRelative) {
		fmt.Fprintln(stderr, "proof registry must be inside the project root")
		return 1
	}
	result := executeProofGovernorPlanAtRegistry(
		root,
		registry,
		strings.Split(opts.proofRequest, ","),
		resolveInputPath(root, opts.receiptDir),
		filepath.ToSlash(registryRelative),
		runProcessCheck,
		time.Now().UTC(),
	)
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	} else {
		fmt.Fprintf(stdout, "proof-run: %s\n", strings.ToUpper(result.Decision))
		for _, reason := range result.Reasons {
			fmt.Fprintf(stdout, "reason: %s\n", reason)
		}
	}
	if result.OK {
		return 0
	}
	if result.Decision == proofGovernorBlock {
		return 2
	}
	if result.ExitCode != 0 {
		return result.ExitCode
	}
	return 1
}

func executeProofGovernorPlan(root string, registry proofGovernorRegistry, requested []string, receiptDir string, runner processRunner, now time.Time) proofGovernorExecutionSummary {
	return executeProofGovernorPlanAtRegistry(root, registry, requested, receiptDir, "registry.json", runner, now)
}

func executeProofGovernorPlanAtRegistry(root string, registry proofGovernorRegistry, requested []string, receiptDir, registryPath string, runner processRunner, now time.Time) (summary proofGovernorExecutionSummary) {
	requested = normalizeProofGovernorIDs(requested)
	defer func() {
		row := proofGovernorAuditRow{
			At: now, Decision: summary.Decision, Requested: requested,
			RunChecks: summary.RunChecks, Reused: summary.Reused,
			ExitCode: summary.ExitCode, Reasons: summary.Reasons,
		}
		if err := appendProofGovernorAudit(receiptDir, row); err != nil {
			summary.OK = false
			summary.Decision = proofGovernorBlock
			summary.Reasons = append(summary.Reasons, "audit-write-error")
		}
	}()
	receipts, receiptIssues := loadProofGovernorReceipts(receiptDir)
	plan := selectProofGovernorPlan(root, registry, requested, receipts, now)
	summary = proofGovernorExecutionSummary{
		Decision:  plan.Decision,
		OK:        plan.Decision == proofGovernorReuse,
		RunChecks: append([]string{}, plan.RunChecks...),
		Reused:    append([]string{}, plan.Reused...),
		Reasons:   append([]string{}, plan.Reasons...),
	}
	if len(receiptIssues) > 0 {
		summary.Decision = proofGovernorBlock
		summary.OK = false
		for _, issue := range receiptIssues {
			summary.Reasons = append(summary.Reasons, "receipt-load:"+issue)
		}
		return summary
	}
	if plan.Decision != proofGovernorRun {
		return summary
	}
	specs, issues := indexProofGovernorRegistry(registry)
	if len(issues) > 0 {
		summary.Decision = proofGovernorBlock
		summary.Reasons = append(summary.Reasons, issues...)
		return summary
	}

	for _, checkID := range plan.RunChecks {
		spec := specs[checkID]
		prospective, err := captureProofGovernorReceipt(root, spec, registryPath, proofGovernorExecutionResult{
			Status: "pending", StartedAt: now, CompletedAt: now,
		})
		if err != nil {
			summary.Decision = proofGovernorBlock
			summary.Reasons = append(summary.Reasons, "fingerprint-error:"+checkID+":"+err.Error())
			return summary
		}
		if unchangedProofGovernorAttemptRecorded(prospective, receipts) {
			summary.Decision = proofGovernorBlock
			summary.Reasons = append(summary.Reasons, "unchanged-attempt-already-recorded:"+checkID)
			return summary
		}
		if spec.ExecutionClass == "visible-browser" || spec.ExecutionClass == "native-gui" {
			summary.Decision = proofGovernorBlock
			summary.Reasons = append(summary.Reasons, "automatic-gui-execution-disabled:"+checkID)
			return summary
		}

		started := time.Now().UTC()
		workingDir, err := resolveProofGovernorPath(root, spec.WorkingDir, false)
		if err != nil {
			summary.Decision = proofGovernorBlock
			summary.Reasons = append(summary.Reasons, "working-dir-invalid:"+checkID)
			return summary
		}
		check := Check{Name: spec.ID, Args: append([]string{}, spec.Command...), Timeout: time.Duration(spec.TimeoutMS) * time.Millisecond}
		result := runner(workingDir, check)
		completed := time.Now().UTC()
		status := "fail"
		switch {
		case result.ExitCode == spec.ExpectedExit:
			status = "pass"
		case result.ExitCode == 124:
			status = "timeout"
		case result.ExitCode == 125:
			status = "partial"
		}
		receipt, err := captureProofGovernorReceipt(root, spec, registryPath, proofGovernorExecutionResult{
			Status: status, ExitCode: result.ExitCode, StartedAt: started, CompletedAt: completed,
		})
		if err != nil {
			summary.Decision = proofGovernorBlock
			summary.Reasons = append(summary.Reasons, "receipt-capture-error:"+checkID)
			return summary
		}
		receiptPath := filepath.Join(receiptDir, checkID+"-"+receipt.ReceiptID[:16]+".proof.json")
		if err := writeProofGovernorReceipt(receiptPath, receipt); err != nil {
			summary.Decision = proofGovernorBlock
			summary.Reasons = append(summary.Reasons, "receipt-write-error:"+checkID)
			return summary
		}
		summary.ReceiptPaths = append(summary.ReceiptPaths, filepath.ToSlash(receiptPath))
		if status != "pass" {
			summary.OK = false
			summary.ExitCode = result.ExitCode
			if status == "timeout" {
				summary.Reasons = append(summary.Reasons, "check-timeout:"+checkID)
			} else if status == "partial" {
				summary.Reasons = append(summary.Reasons, "check-partial:"+checkID)
			} else {
				summary.Reasons = append(summary.Reasons, "check-failed:"+checkID)
			}
			return summary
		}
		receipts = append(receipts, receipt)
	}
	summary.OK = true
	summary.ExitCode = 0
	summary.Reasons = append(summary.Reasons, "selected-checks-passed")
	return summary
}

func appendProofGovernorAudit(receiptDir string, row proofGovernorAuditRow) error {
	content, err := json.Marshal(row)
	if err != nil {
		return err
	}
	if len(content) > maxProofGovernorBytes {
		return fmt.Errorf("proof audit row exceeds size limit")
	}
	if err := os.MkdirAll(receiptDir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(receiptDir, ".proof-audit.jsonl")
	if info, statErr := os.Stat(path); statErr == nil && info.Size()+int64(len(content)+1) > maxProofGovernorAuditBytes {
		return fmt.Errorf("proof audit ledger exceeds size limit")
	} else if statErr != nil && !os.IsNotExist(statErr) {
		return statErr
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(content, '\n'))
	return err
}

func unchangedProofGovernorAttemptRecorded(prospective proofGovernorReceipt, receipts []proofGovernorReceipt) bool {
	for _, receipt := range receipts {
		if receipt.Spec.ID != prospective.Spec.ID || receipt.Result.Status == "pass" {
			continue
		}
		if receipt.RegistryPath == prospective.RegistryPath &&
			receipt.RegistrySHA256 == prospective.RegistrySHA256 &&
			receipt.CheckSemanticsSHA256 == prospective.CheckSemanticsSHA256 &&
			receipt.RelevantInputsSHA256 == prospective.RelevantInputsSHA256 &&
			receipt.EnvironmentSHA256 == prospective.EnvironmentSHA256 {
			return true
		}
	}
	return false
}
