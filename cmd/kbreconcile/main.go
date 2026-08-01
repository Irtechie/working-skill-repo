package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/reconcile"
)

const usage = `kbreconcile inventories and plans portfolio reconciliation without mutation.

Usage:
  kbreconcile dry-run --repo <path> [--repo <path>...] [--policy <path>] [--cutoff <RFC3339>] [--json]
  kbreconcile plan --repo <path> [--repo <path>...] --output <path> [--policy <path>] [--cutoff <RFC3339>] [--json]
  kbreconcile apply --input <plan.json> --receipt <receipt.json> [--policy <path>] [--session-id <id>] [--json]
  kbreconcile verify --input <plan.json> --receipt <receipt.json> [--policy <path>] [--session-id <id>] [--json]
  kbreconcile claim-capability [--json]
  kbreconcile claim-conformance [--json]
`

var now = time.Now

type stringList []string

func (values *stringList) String() string { return strings.Join(*values, ",") }

func (values *stringList) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("repository path cannot be empty")
	}
	*values = append(*values, value)
	return nil
}

type commandResult struct {
	SchemaVersion   int                               `json:"schema_version"`
	Status          string                            `json:"status"`
	Mode            string                            `json:"mode,omitempty"`
	Cutoff          time.Time                         `json:"cutoff,omitempty"`
	Output          string                            `json:"output,omitempty"`
	Ledger          *reconcile.Ledger                 `json:"ledger,omitempty"`
	Plan            *reconcile.Plan                   `json:"plan,omitempty"`
	Receipt         *reconcile.ApplyReceipt           `json:"receipt,omitempty"`
	Verification    *reconcile.Verification           `json:"verification,omitempty"`
	ClaimCapability *reconcile.GatewayCapability      `json:"claim_capability,omitempty"`
	Conformance     *reconcile.ClaimConformanceResult `json:"claim_conformance,omitempty"`
	Error           *commandError                     `json:"error,omitempty"`
}

type commandError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, usage)
		return 0
	}
	mode := args[0]
	if mode == "claim-capability" || mode == "claim-conformance" {
		return runClaimContract(mode, args[1:], stdout, stderr)
	}
	if mode == "apply" || mode == "verify" {
		return runApplyVerify(mode, args[1:], stdout, stderr)
	}
	if mode != reconcile.ModeDryRun && mode != reconcile.ModePlan {
		return writeCommandError(stdout, stderr, hasJSON(args), 2, "unsupported-mode", "command must be dry-run, plan, apply, verify, claim-capability, or claim-conformance")
	}

	flags := flag.NewFlagSet("kbreconcile "+mode, flag.ContinueOnError)
	flags.SetOutput(stderr)
	var roots stringList
	var policyPath, cutoffText, outputPath string
	var jsonMode bool
	flags.Var(&roots, "repo", "repository root; repeat for portfolio inventory")
	flags.StringVar(&policyPath, "policy", "", "versioned reconciliation policy")
	flags.StringVar(&cutoffText, "cutoff", "", "immutable RFC3339 scan cutoff")
	flags.StringVar(&outputPath, "output", "", "durable plan output path")
	flags.BoolVar(&jsonMode, "json", false, "emit stable JSON")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if len(roots) == 0 {
		return writeCommandError(stdout, stderr, jsonMode, 2, "missing-repository", "at least one --repo is required")
	}
	if mode == reconcile.ModePlan && strings.TrimSpace(outputPath) == "" {
		return writeCommandError(stdout, stderr, jsonMode, 2, "missing-output", "plan requires --output")
	}
	if flags.NArg() != 0 {
		return writeCommandError(stdout, stderr, jsonMode, 2, "unexpected-argument", "unexpected positional arguments")
	}
	cutoff := now().UTC()
	if cutoffText != "" {
		parsed, err := time.Parse(time.RFC3339Nano, cutoffText)
		if err != nil {
			return writeCommandError(stdout, stderr, jsonMode, 2, "invalid-cutoff", "cutoff must be RFC3339")
		}
		cutoff = parsed.UTC()
	}
	policy := reconcile.DefaultPolicy()
	if policyPath != "" {
		loaded, err := reconcile.LoadPolicy(policyPath)
		if err != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "policy-unavailable", err.Error())
		}
		policy = loaded
	}
	current, _ := os.Getwd()
	ledger, err := reconcile.Inventory(reconcile.InventoryOptions{
		Roots: roots, Cutoff: cutoff, CurrentWorktree: current,
		CurrentSessionID: os.Getenv("COPILOT_SESSION_ID"), Now: now,
	})
	if err != nil {
		return writeCommandError(stdout, stderr, jsonMode, 1, "inventory-failed", err.Error())
	}
	plan, err := reconcile.BuildPlan(ledger, policy)
	if err != nil {
		return writeCommandError(stdout, stderr, jsonMode, 1, "planning-failed", err.Error())
	}
	result := commandResult{
		SchemaVersion: 1, Status: "ok", Mode: mode, Cutoff: ledger.Cutoff,
		Ledger: &ledger, Plan: &plan,
	}
	if mode == reconcile.ModePlan {
		absolute, err := filepath.Abs(filepath.Clean(outputPath))
		if err != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "invalid-output", err.Error())
		}
		result.Output = absolute
		content, err := reconcile.MarshalStable(result)
		if err != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "encode-failed", err.Error())
		}
		if err := writeAtomic(absolute, content); err != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "write-failed", err.Error())
		}
	}
	if jsonMode {
		content, err := reconcile.MarshalStable(result)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		_, _ = stdout.Write(content)
		return 0
	}

	fmt.Fprintf(stdout, "kbreconcile: %s ok repositories=%d actions=%d decisions=%d\n",
		mode, len(ledger.Repositories), len(plan.Actions), len(plan.DecisionPacket.Items))
	if result.Output != "" {
		fmt.Fprintf(stdout, "plan: %s\n", result.Output)
	}
	return 0
}

func runClaimContract(mode string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("kbreconcile "+mode, flag.ContinueOnError)
	flags.SetOutput(stderr)
	var jsonMode bool
	flags.BoolVar(&jsonMode, "json", false, "emit stable JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		return writeCommandError(stdout, stderr, jsonMode, 2, "unexpected-argument", "unexpected positional arguments")
	}
	capability := reconcile.ReferenceClaimCapability()
	result := commandResult{
		SchemaVersion: 1, Status: "ok", Mode: mode, ClaimCapability: &capability,
	}
	if mode == "claim-conformance" {
		conformance := reconcile.ReferenceClaimConformance()
		result.Conformance = &conformance
	}
	if jsonMode {
		content, err := reconcile.MarshalStable(result)
		if err != nil {
			return writeCommandError(stdout, stderr, true, 1, "encode-failed", err.Error())
		}
		_, _ = stdout.Write(content)
		return 0
	}
	fmt.Fprintf(stdout, "kbreconcile: %s ok protected-mutation=%t live-provider=%t\n",
		mode, capability.ProtectedMutationAvailable, capability.LiveProviderSupported)
	return 0
}

func runApplyVerify(mode string, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("kbreconcile "+mode, flag.ContinueOnError)
	flags.SetOutput(stderr)
	var inputPath, receiptPath, policyPath, sessionID string
	var jsonMode bool
	flags.StringVar(&inputPath, "input", "", "cutoff-bound plan bundle")
	flags.StringVar(&receiptPath, "receipt", "", "durable apply receipt")
	flags.StringVar(&policyPath, "policy", "", "versioned reconciliation policy")
	flags.StringVar(&sessionID, "session-id", os.Getenv("COPILOT_SESSION_ID"), "current executor session identity")
	flags.BoolVar(&jsonMode, "json", false, "emit stable JSON")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		return writeCommandError(stdout, stderr, jsonMode, 2, "unexpected-argument", "unexpected positional arguments")
	}
	if strings.TrimSpace(inputPath) == "" || strings.TrimSpace(receiptPath) == "" {
		return writeCommandError(stdout, stderr, jsonMode, 2, "missing-input", "apply and verify require --input and --receipt")
	}
	policy := reconcile.DefaultPolicy()
	if policyPath != "" {
		loaded, err := reconcile.LoadPolicy(policyPath)
		if err != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "policy-unavailable", err.Error())
		}
		policy = loaded
	}
	bundle, err := loadPlanBundle(inputPath)
	if err != nil {
		return writeCommandError(stdout, stderr, jsonMode, 1, "plan-unavailable", err.Error())
	}
	absoluteReceipt, err := filepath.Abs(filepath.Clean(receiptPath))
	if err != nil {
		return writeCommandError(stdout, stderr, jsonMode, 1, "invalid-receipt", err.Error())
	}
	current, _ := os.Getwd()
	options := reconcile.ApplyOptions{
		Bundle: bundle, Policy: policy, CurrentWorktree: current,
		CurrentSession: sessionID, Now: now(),
	}
	result := commandResult{SchemaVersion: 1, Status: "ok", Mode: mode, Output: absoluteReceipt}
	switch mode {
	case "apply":
		existing, loadErr := loadApplyReceiptIfPresent(absoluteReceipt)
		if loadErr != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "receipt-unavailable", loadErr.Error())
		}
		options.ExistingReceipt = existing
		receipt, applyErr := reconcile.Apply(options)
		if applyErr != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "apply-refused", applyErr.Error())
		}
		result.Receipt = &receipt
	case "verify":
		existing, loadErr := loadApplyReceipt(absoluteReceipt)
		if loadErr != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "receipt-unavailable", loadErr.Error())
		}
		verification, receipt, verifyErr := reconcile.Verify(options, existing)
		if verifyErr != nil {
			return writeCommandError(stdout, stderr, jsonMode, 1, "verify-refused", verifyErr.Error())
		}
		result.Receipt = &receipt
		result.Verification = &verification
	}
	content, err := reconcile.MarshalStable(*result.Receipt)
	if err != nil {
		return writeCommandError(stdout, stderr, jsonMode, 1, "encode-failed", err.Error())
	}
	if err := writeAtomic(absoluteReceipt, content); err != nil {
		return writeCommandError(stdout, stderr, jsonMode, 1, "write-failed", err.Error())
	}
	if jsonMode {
		content, err := reconcile.MarshalStable(result)
		if err != nil {
			return writeCommandError(stdout, stderr, true, 1, "encode-failed", err.Error())
		}
		_, _ = stdout.Write(content)
	} else {
		fmt.Fprintf(stdout, "kbreconcile: %s %s receipt=%s\n", mode, result.Receipt.Status, absoluteReceipt)
	}
	return 0
}

func loadPlanBundle(path string) (reconcile.PlanBundle, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return reconcile.PlanBundle{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var result commandResult
	if err := decoder.Decode(&result); err != nil {
		return reconcile.PlanBundle{}, fmt.Errorf("parse plan bundle: %w", err)
	}
	if err := requireJSONEnd(decoder); err != nil {
		return reconcile.PlanBundle{}, fmt.Errorf("parse plan bundle: %w", err)
	}
	if result.Ledger == nil || result.Plan == nil || result.Status != "ok" || result.Mode != reconcile.ModePlan {
		return reconcile.PlanBundle{}, fmt.Errorf("input is not a successful durable plan bundle")
	}
	return reconcile.PlanBundle{Ledger: *result.Ledger, Plan: *result.Plan}, nil
}

func loadApplyReceiptIfPresent(path string) (*reconcile.ApplyReceipt, error) {
	receipt, err := loadApplyReceipt(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func loadApplyReceipt(path string) (reconcile.ApplyReceipt, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return reconcile.ApplyReceipt{}, err
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var receipt reconcile.ApplyReceipt
	if err := decoder.Decode(&receipt); err != nil {
		return reconcile.ApplyReceipt{}, fmt.Errorf("parse apply receipt: %w", err)
	}
	if err := requireJSONEnd(decoder); err != nil {
		return reconcile.ApplyReceipt{}, fmt.Errorf("parse apply receipt: %w", err)
	}
	return receipt, nil
}

func requireJSONEnd(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("unexpected trailing JSON value")
		}
		return err
	}
	return nil
}

func writeCommandError(stdout, stderr io.Writer, jsonMode bool, exitCode int, code, message string) int {
	result := commandResult{
		SchemaVersion: 1, Status: "error",
		Error: &commandError{Code: code, Message: message},
	}
	if jsonMode {
		content, err := reconcile.MarshalStable(result)
		if err == nil {
			_, _ = stdout.Write(content)
		} else {
			fmt.Fprintln(stderr, err)
			return 1
		}
	} else {
		fmt.Fprintln(stderr, message)
	}
	return exitCode
}

func hasJSON(args []string) bool {
	for _, argument := range args {
		if argument == "--json" {
			return true
		}
	}
	return false
}

func writeAtomic(path string, content []byte) error {
	parent := filepath.Dir(path)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return err
	}
	staging := path + ".new"
	if err := os.WriteFile(staging, content, 0o600); err != nil {
		return err
	}
	if err := os.Rename(staging, path); err != nil {
		_ = os.Remove(staging)
		return err
	}
	return nil
}
