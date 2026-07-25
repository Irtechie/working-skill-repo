package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type graphRoutingLifecycleTrace struct {
	SchemaVersion int                    `json:"schema_version"`
	Packet        lifecyclePacketSummary `json:"packet"`
	Plan          lifecyclePlanSummary   `json:"plan"`
	Work          lifecycleWorkSummary   `json:"work"`
	Review        lifecycleReviewSummary `json:"review"`
}

type lifecyclePacketSummary struct {
	PacketID       string         `json:"packet_id"`
	Freshness      string         `json:"freshness"`
	FallbackMode   string         `json:"fallback_mode"`
	EvidenceCounts map[string]int `json:"evidence_counts"`
	Impact         []string       `json:"impact"`
}

type lifecyclePlanSummary struct {
	ImpactForecast                []string `json:"impact_forecast"`
	ExpectedFiles                 []string `json:"expected_files"`
	ExpectedFilesAreForecast      bool     `json:"expected_files_are_forecast"`
	WorkspaceMode                 string   `json:"workspace_mode"`
	ConflictDomains               []string `json:"conflict_domains"`
	CoordinatorOwnsLifecycleFiles bool     `json:"coordinator_owns_lifecycle_files"`
}

type lifecycleWorkSummary struct {
	LeaseAcquired           bool                      `json:"lease_acquired"`
	WorkerCanMarkDone       bool                      `json:"worker_can_mark_done"`
	FunctionalProofRequired bool                      `json:"functional_proof_required"`
	ReceiptCanReplaceProof  bool                      `json:"receipt_can_replace_proof"`
	ObservedFiles           []lifecycleObservedImpact `json:"observed_files"`
}

type lifecycleObservedImpact struct {
	Path       string `json:"path"`
	Provenance string `json:"provenance"`
}

type lifecycleReviewSummary struct {
	MissedImpact       []string `json:"missed_impact"`
	UnexplainedGrowth  []string `json:"unexplained_growth"`
	ReceiptOnlyVerdict bool     `json:"receipt_only_verdict"`
}

func runGraphRoutingLifecycleSelftest(root string, stdout, stderr io.Writer) int {
	path := filepath.Join(root, "evals", "skill-eval", "selftest", "pass-evidence-graph-routing.json")
	trace, err := loadGraphRoutingLifecycleTrace(path)
	if err != nil {
		fmt.Fprintf(stderr, "graph-routing lifecycle selftest: %v\n", err)
		return 1
	}
	if issues := validateGraphRoutingLifecycleTrace(trace); len(issues) > 0 {
		fmt.Fprintf(stderr, "graph-routing lifecycle selftest valid fixture failed: %s\n", strings.Join(issues, "; "))
		return 1
	}

	stale := trace
	stale.Packet.Freshness = "stale"
	stale.Packet.FallbackMode = "authoritative-provider"
	if !hasLifecycleIssue(validateGraphRoutingLifecycleTrace(stale), "stale packet must use file-native fallback") {
		fmt.Fprintln(stderr, "graph-routing lifecycle selftest: stale authoritative packet was accepted")
		return 1
	}

	workerDone := trace
	workerDone.Work.WorkerCanMarkDone = true
	if !hasLifecycleIssue(validateGraphRoutingLifecycleTrace(workerDone), "worker receipt cannot mark done") {
		fmt.Fprintln(stderr, "graph-routing lifecycle selftest: worker done receipt was accepted")
		return 1
	}

	scopeGrowth := trace
	scopeGrowth.Review.UnexplainedGrowth = []string{"src/unplanned.go"}
	if !hasLifecycleIssue(validateGraphRoutingLifecycleTrace(scopeGrowth), "unexplained impact growth") {
		fmt.Fprintln(stderr, "graph-routing lifecycle selftest: unexplained scope growth was accepted")
		return 1
	}

	fmt.Fprintln(stdout, "graph-routing lifecycle selftest: passed")
	return 0
}

func loadGraphRoutingLifecycleTrace(path string) (graphRoutingLifecycleTrace, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return graphRoutingLifecycleTrace{}, err
	}
	var trace graphRoutingLifecycleTrace
	if err := json.Unmarshal(content, &trace); err == nil && trace.SchemaVersion == 1 {
		return trace, nil
	}
	var wrapped struct {
		Lifecycle graphRoutingLifecycleTrace `json:"lifecycle"`
	}
	if err := json.Unmarshal(content, &wrapped); err != nil {
		return graphRoutingLifecycleTrace{}, err
	}
	if wrapped.Lifecycle.SchemaVersion != 1 {
		return graphRoutingLifecycleTrace{}, fmt.Errorf("lifecycle payload is missing or invalid")
	}
	return wrapped.Lifecycle, nil
}

func validateGraphRoutingLifecycleTrace(trace graphRoutingLifecycleTrace) []string {
	issues := []string{}
	if trace.SchemaVersion != 1 {
		issues = append(issues, "schema_version must be 1")
	}
	if strings.TrimSpace(trace.Packet.PacketID) == "" {
		issues = append(issues, "packet_id is required")
	}
	if trace.Packet.Freshness == "stale" && trace.Packet.FallbackMode != "file-native" {
		issues = append(issues, "stale packet must use file-native fallback")
	}
	if trace.Packet.Freshness == "" || trace.Packet.FallbackMode == "" {
		issues = append(issues, "packet freshness and fallback are required")
	}
	if len(trace.Packet.Impact) == 0 {
		issues = append(issues, "impact forecast needs at least one cited path")
	}
	if !trace.Plan.ExpectedFilesAreForecast {
		issues = append(issues, "expected_files must remain a forecast")
	}
	if !validWorkspaceMode(trace.Plan.WorkspaceMode) {
		issues = append(issues, "workspace mode must be shared-serial or worktree-required")
	}
	if len(trace.Plan.ConflictDomains) == 0 {
		issues = append(issues, "conflict domains are required")
	}
	if !trace.Plan.CoordinatorOwnsLifecycleFiles {
		issues = append(issues, "coordinator must own lifecycle files")
	}
	if !trace.Work.LeaseAcquired {
		issues = append(issues, "slice lease must be acquired before mutation")
	}
	if trace.Work.WorkerCanMarkDone {
		issues = append(issues, "worker receipt cannot mark done")
	}
	if !trace.Work.FunctionalProofRequired || trace.Work.ReceiptCanReplaceProof {
		issues = append(issues, "receipt cannot replace functional proof")
	}
	forecast := map[string]bool{}
	for _, path := range append(append([]string{}, trace.Packet.Impact...), trace.Plan.ExpectedFiles...) {
		forecast[path] = true
	}
	for _, observed := range trace.Work.ObservedFiles {
		if strings.TrimSpace(observed.Path) == "" || strings.TrimSpace(observed.Provenance) == "" {
			issues = append(issues, "observed impact requires path and provenance")
			continue
		}
		if !forecast[observed.Path] && !strings.Contains(observed.Provenance, "impact") {
			issues = append(issues, "observed file outside forecast requires impact provenance")
		}
	}
	if len(trace.Review.MissedImpact) > 0 {
		issues = append(issues, "missed impact must be resolved before completion")
	}
	if len(trace.Review.UnexplainedGrowth) > 0 {
		issues = append(issues, "unexplained impact growth")
	}
	if trace.Review.ReceiptOnlyVerdict {
		issues = append(issues, "review cannot accept receipt-only verdict")
	}
	return issues
}

func hasLifecycleIssue(issues []string, want string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, want) {
			return true
		}
	}
	return false
}
