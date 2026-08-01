package main

import (
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
	SchemaVersion int               `json:"schema_version"`
	Status        string            `json:"status"`
	Mode          string            `json:"mode,omitempty"`
	Cutoff        time.Time         `json:"cutoff,omitempty"`
	Output        string            `json:"output,omitempty"`
	Ledger        *reconcile.Ledger `json:"ledger,omitempty"`
	Plan          *reconcile.Plan   `json:"plan,omitempty"`
	Error         *commandError     `json:"error,omitempty"`
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
	if mode != reconcile.ModeDryRun && mode != reconcile.ModePlan {
		return writeCommandError(stdout, stderr, hasJSON(args), 2, "unsupported-mode", "command must be dry-run or plan")
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
