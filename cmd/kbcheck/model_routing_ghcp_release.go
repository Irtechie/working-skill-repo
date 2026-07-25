package main

import (
	"fmt"
	"io"
)

type ghcpFollowOnEvidence struct {
	SchemaVersion         int                  `json:"schema_version"`
	Cohort                string               `json:"cohort"`
	EvidenceMode          string               `json:"evidence_mode"`
	PaidCalls             int                  `json:"paid_calls"`
	InitialCohortDecision string               `json:"initial_cohort_decision"`
	Families              []ghcpFamilyEvidence `json:"families"`
}

type ghcpFamilyEvidence struct {
	Family                 string  `json:"family"`
	Samples                int     `json:"samples"`
	ZeroRightToWrong       bool    `json:"zero_right_to_wrong"`
	CorrectnessNonInferior bool    `json:"correctness_non_inferior"`
	AggregateAIULower      bool    `json:"aggregate_aiu_lower"`
	MedianAIUReduction     float64 `json:"median_aiu_reduction"`
	ConfidenceLower        float64 `json:"confidence_lower"`
	InterventionRegression bool    `json:"intervention_regression"`
}

func validateGHCPFollowOn(evidence ghcpFollowOnEvidence) (string, error) {
	if evidence.SchemaVersion != 1 || evidence.Cohort != "ghcp-follow-on" {
		return "", fmt.Errorf("unsupported GHCP follow-on schema or cohort")
	}

	if evidence.InitialCohortDecision != "not-promoted" {
		return "", fmt.Errorf("GHCP follow-on must preserve the landed initial cohort decision")
	}
	if evidence.EvidenceMode == "deterministic-no-paid" {
		if evidence.PaidCalls != 0 || len(evidence.Families) != 0 {
			return "", fmt.Errorf("no-paid GHCP evidence cannot contain live family results")
		}
		return "not-promoted", nil
	}
	if evidence.EvidenceMode != "attended-live" || evidence.PaidCalls <= 0 || len(evidence.Families) == 0 {
		return "", fmt.Errorf("GHCP promotion requires attended live paired evidence")
	}
	seen := map[string]bool{}
	for _, family := range evidence.Families {
		if family.Family == "" || seen[family.Family] {
			return "", fmt.Errorf("GHCP family evidence must be unique")
		}
		seen[family.Family] = true
		if family.Samples < 20 || !family.ZeroRightToWrong || !family.CorrectnessNonInferior ||
			!family.AggregateAIULower || family.MedianAIUReduction < 0.20 ||
			family.ConfidenceLower <= 0 || family.InterventionRegression {
			return "not-promoted", fmt.Errorf("GHCP family %q did not satisfy paired promotion gates", family.Family)
		}
	}
	return "not-promoted", fmt.Errorf("GHCP attended evidence cannot promote until an independent live verifier is implemented")
}

func runGHCPFollowOnReleaseCommand(root string, opts options, stdout, stderr io.Writer) int {
	path, err := resolveModelRoutingReleaseFile(root, opts.evidencePath)
	if err != nil {
		fmt.Fprintf(stderr, "GHCP follow-on evidence: %v\n", err)
		return 1
	}
	var evidence ghcpFollowOnEvidence
	if err := readStrictModelRoutingJSON(path, &evidence); err != nil {
		fmt.Fprintf(stderr, "GHCP follow-on evidence: %v\n", err)
		return 1
	}
	decision, err := validateGHCPFollowOn(evidence)
	if err != nil {
		if opts.json {
			writeJSON(stdout, map[string]any{"ok": false, "cohort": evidence.Cohort, "decision": decision, "issues": []string{err.Error()}})
			return 1
		}
		fmt.Fprintf(stderr, "GHCP follow-on evidence: %v\n", err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, map[string]any{
			"ok": true, "cohort": evidence.Cohort, "decision": decision,
			"paid_calls": evidence.PaidCalls, "families": len(evidence.Families),
		})
		return 0
	}
	fmt.Fprintf(stdout, "model-routing-release: honest cohort=%s decision=%s paid_calls=%d families=%d\n", evidence.Cohort, decision, evidence.PaidCalls, len(evidence.Families))
	return 0
}
