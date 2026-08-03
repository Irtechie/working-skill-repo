package main

import "testing"

func TestGHCPFollowOnNoPaidEvidenceStaysNotPromoted(t *testing.T) {
	t.Parallel()
	evidence := ghcpFollowOnEvidence{
		SchemaVersion:         1,
		Cohort:                "ghcp-follow-on",
		EvidenceMode:          "deterministic-no-paid",
		PaidCalls:             0,
		InitialCohortDecision: "not-promoted",
	}
	decision, err := validateGHCPFollowOn(evidence)
	if err != nil {
		t.Fatal(err)
	}
	if decision != "not-promoted" {
		t.Fatalf("decision=%q", decision)
	}
}

func TestGHCPFollowOnRefusesPromotionWithoutIndependentVerifier(t *testing.T) {
	t.Parallel()
	evidence := validGHCPFollowOnEvidence()
	decision, err := validateGHCPFollowOn(evidence)
	if err == nil || decision != "not-promoted" {
		t.Fatalf("decision=%q err=%v", decision, err)
	}
	evidence.Families[1].MedianAIUReduction = 0.19
	if decision, err := validateGHCPFollowOn(evidence); err == nil || decision == "promoted" {
		t.Fatalf("weak family promoted: decision=%q err=%v", decision, err)
	}
}

func TestGHCPFollowOnCannotRewriteInitialCohort(t *testing.T) {
	t.Parallel()
	evidence := validGHCPFollowOnEvidence()
	evidence.InitialCohortDecision = "promoted"
	if _, err := validateGHCPFollowOn(evidence); err == nil {
		t.Fatal("follow-on evidence rewrote initial cohort")
	}
}

func validGHCPFollowOnEvidence() ghcpFollowOnEvidence {
	family := ghcpFamilyEvidence{
		Family: "go-local-logic", Samples: 20, ZeroRightToWrong: true,
		CorrectnessNonInferior: true, AggregateAIULower: true,
		MedianAIUReduction: 0.25, ConfidenceLower: 0.10,
		InterventionRegression: false,
	}
	second := family
	second.Family = "go-cross-file"
	return ghcpFollowOnEvidence{
		SchemaVersion: 1, Cohort: "ghcp-follow-on", EvidenceMode: "attended-live",
		PaidCalls: 80, InitialCohortDecision: "not-promoted",
		Families: []ghcpFamilyEvidence{family, second},
	}
}
