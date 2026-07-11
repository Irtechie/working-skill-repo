package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const modelRoutingInitialPilotEvidence = "docs/results/2026-07-10-session-model-routing-initial-pilot.json"
const modelRoutingFeatureMarker = "internal/modelrouting/selector.go"

type ReleaseResult struct {
	OK               bool              `json:"ok"`
	Profile          string            `json:"profile"`
	GeneratedAt      string            `json:"generated_at"`
	ResultCount      int               `json:"result_count"`
	RequiredFailures int               `json:"required_failures"`
	OptionalFailures int               `json:"optional_failures"`
	Results          []ReleaseCheckRun `json:"results"`
}

type ReleaseCheckRun struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Confidence string `json:"confidence"`
	Required   bool   `json:"required"`
	Command    string `json:"command"`
	ExitCode   *int   `json:"exit_code"`
	DurationMS int64  `json:"duration_ms,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

func runRelease(root string, opts options, stdout, stderr io.Writer, runner processRunner) int {
	checks, err := releaseChecks(root, opts.command, runner)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.dryRun {
		printChecks(stdout, checks)
		return 0
	}

	results := make([]ReleaseCheckRun, 0, len(checks))
	requiredFailures := 0
	optionalFailures := 0
	for _, check := range checks {
		if !opts.json {
			requiredLabel := "optional"
			if check.Required {
				requiredLabel = "required"
			}
			fmt.Fprintf(stdout, "running [%s/%s] %s: %s\n", requiredLabel, check.Confidence, check.Name, check.CommandString())
		}
		run := invokeReleaseCheck(root, check, runner)
		results = append(results, run)
		if run.Required && run.Status != "passed" {
			requiredFailures++
		}
		if !run.Required && run.Status == "failed" {
			optionalFailures++
		}
	}

	summary := ReleaseResult{
		OK:               requiredFailures == 0 && optionalFailures == 0,
		Profile:          opts.command,
		GeneratedAt:      time.Now().Format(time.RFC3339Nano),
		ResultCount:      len(results),
		RequiredFailures: requiredFailures,
		OptionalFailures: optionalFailures,
		Results:          results,
	}

	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(summary)
	} else {
		fmt.Fprintf(stdout, "KB release gate: profile=%s ok=%v\n", summary.Profile, summary.OK)
		for _, result := range summary.Results {
			requiredLabel := "optional"
			if result.Required {
				requiredLabel = "required"
			}
			fmt.Fprintf(stdout, "%s [%s/%s] %s: %s\n", result.Status, requiredLabel, result.Confidence, result.Name, result.Command)
			if result.Status != "passed" && result.Notes != "" {
				fmt.Fprintf(stdout, "  %s\n", result.Notes)
			}
		}
	}
	if !summary.OK {
		return 1
	}
	return 0
}

func releaseChecks(root, profile string, runner processRunner) ([]Check, error) {
	checks := []Check{
		{
			Name:       "kb-check-all",
			Args:       []string{"kbcheck", "core"},
			Reason:     "native core gate",
			Required:   true,
			Confidence: "deterministic-local",
			Run: func(root string) CheckResult {
				var out, err bytes.Buffer
				code := runCore(root, options{command: "core", root: root}, &out, &err, runner)
				return CheckResult{ExitCode: code, Stdout: out.String(), Stderr: err.String()}
			},
		},
		{Name: "git-diff-check", Args: []string{"git", "diff", "--check"}, Reason: "whitespace/conflict guard", Required: true, Confidence: "deterministic-local"},
	}

	if exists(root, ".github/skills") && exists(root, "config/skill-quality.json") {
		checks = append(checks, Check{
			Name: "skill-sync-report", Args: []string{"kbcheck", "skill-sync-report"},
			Reason: "required skill sync target config", Required: true, Confidence: "deterministic-local",
			Run: func(root string) CheckResult { return runNativeCommand(root, []string{"skill-sync-report"}) },
		})
	} else {
		checks = append(checks, Check{Name: "skill-sync-report", Args: []string{"kbcheck", "skill-sync-report"}, Required: false, Confidence: "deterministic-local", Available: func(string) bool { return false }, SkipReason: "skill sync target config unavailable"})
	}
	if exists(root, ".github/skills") {
		checks = append(checks, Check{
			Name: "skill-surface-minimality", Args: []string{"kbcheck", "minimality"},
			Required: false, Confidence: "static-report",
			Run: func(root string) CheckResult { return runNativeCommand(root, []string{"minimality"}) },
		})
	} else {
		checks = append(checks, Check{Name: "skill-surface-minimality", Args: []string{"kbcheck", "minimality"}, Required: false, Confidence: "static-report", Available: func(string) bool { return false }, SkipReason: "skill surface unavailable"})
	}
	// Once the routing feature exists, its canonical evidence is mandatory. Do
	// not let deleting or renaming the evidence silently remove the release gate.
	if exists(root, modelRoutingFeatureMarker) || exists(root, modelRoutingInitialPilotEvidence) {
		checks = append(checks, Check{
			Name: "model-routing-initial-pilot",
			Args: []string{
				"kbcheck", "model-routing-release",
				"--cohort", "initial-pilot",
				"--evidence", modelRoutingInitialPilotEvidence,
			},
			Reason:     "canonical model-routing pilot evidence detected",
			Required:   true,
			Confidence: "deterministic-local",
			Run: func(root string) CheckResult {
				return runNativeCommand(root, []string{
					"model-routing-release",
					"--cohort", "initial-pilot",
					"--evidence", modelRoutingInitialPilotEvidence,
				})
			},
		})
	}
	if profile == "live-release" {
		if exists(root, "evals/route-complexity") {
			checks = append(checks, Check{
				Name: "live-codex-ghcp-corpus", Args: []string{"kbcheck", "eval-run-live-corpus", "--runtime", "codex,ghcp"},
				Required: false, Confidence: "live-model-explicit", Reason: "explicit live model corpus",
				Run: func(root string) CheckResult {
					return runNativeCommand(root, []string{"eval-run-live-corpus", "--runtime", "codex,ghcp"})
				},
			})
		} else {
			checks = append(checks, Check{Name: "live-codex-ghcp-corpus", Args: []string{"kbcheck", "eval-run-live-corpus", "--runtime", "codex,ghcp"}, Required: false, Confidence: "live-model-explicit", Available: func(string) bool { return false }, SkipReason: "route eval fixtures unavailable"})
		}
	}
	return checks, nil
}

func invokeReleaseCheck(root string, check Check, runner processRunner) ReleaseCheckRun {
	if check.Available != nil && !check.Available(root) {
		return ReleaseCheckRun{Name: check.Name, Status: "skipped-explicit", Confidence: check.Confidence, Required: check.Required, Command: check.CommandString(), Notes: check.SkipReason}
	}
	started := time.Now()
	result := runner(root, check)
	duration := time.Since(started).Milliseconds()
	exitCode := result.ExitCode
	status := "passed"
	if result.ExitCode != 0 {
		status = "failed"
	}
	notes := strings.TrimSpace(result.Stdout + "\n" + result.Stderr)
	if len(notes) > 800 {
		notes = notes[:800]
	}
	return ReleaseCheckRun{Name: check.Name, Status: status, Confidence: check.Confidence, Required: check.Required, Command: check.CommandString(), ExitCode: &exitCode, DurationMS: duration, Notes: notes}
}

func pathExists(path string) bool {
	if _, err := os.Stat(filepath.FromSlash(path)); err == nil {
		return true
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
