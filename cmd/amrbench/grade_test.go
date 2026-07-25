package main

import (
	"strings"
	"testing"
)

func TestPairedGradeRejectsInvalidEvidence(t *testing.T) {
	valid := pairedSample{
		TaskID: "task", Seed: 1, Family: "go-local-logic",
		Direct: pairedArm{Correct: true, AIUNano: 100, LatencyMS: 100, TelemetryComplete: true, RouteMatch: true, OracleIntact: true, IsolationReady: true, ProofPresent: true},
		AMR:    pairedArm{Correct: true, AIUNano: 70, LatencyMS: 120, TelemetryComplete: true, RouteMatch: true, OracleIntact: true, IsolationReady: true, ProofPresent: true},
	}
	tests := map[string]func(*pairedSample){
		"partial telemetry": func(sample *pairedSample) { sample.AMR.TelemetryComplete = false },
		"route mismatch":    func(sample *pairedSample) { sample.AMR.RouteMatch = false },
		"oracle mutation":   func(sample *pairedSample) { sample.AMR.OracleIntact = false },
		"isolation failure": func(sample *pairedSample) { sample.AMR.IsolationReady = false },
		"missing proof":     func(sample *pairedSample) { sample.AMR.ProofPresent = false },
		"zero direct cost":  func(sample *pairedSample) { sample.Direct.AIUNano = 0 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			sample := valid
			mutate(&sample)
			if _, err := gradePairedSamples([]pairedSample{sample}); err == nil {
				t.Fatal("invalid paired evidence passed")
			}
		})
	}
	duplicate := valid
	if _, err := gradePairedSamples([]pairedSample{valid, duplicate}); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate pair err=%v", err)
	}
}

func TestPairedGradeIsFamilySpecificAndRequiresCorrectnessAggregateAndMedian(t *testing.T) {
	samples := make([]pairedSample, 0, 40)
	for index := 0; index < 20; index++ {
		samples = append(samples, validPairedSample("go-local-logic", "retry", index, 100, 70))
		samples = append(samples, validPairedSample("go-cross-file", "cache", index, 100, 95))
	}
	report, err := gradePairedSamples(samples)
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Families) != 2 {
		t.Fatalf("families=%+v", report.Families)
	}
	if !report.Families["go-local-logic"].PromotionEligible {
		t.Fatalf("strong family=%+v", report.Families["go-local-logic"])
	}
	if report.Families["go-cross-file"].PromotionEligible {
		t.Fatalf("weak family=%+v", report.Families["go-cross-file"])
	}

	regression := append([]pairedSample(nil), samples[:20]...)
	regression[0].AMR.Correct = false
	report, err = gradePairedSamples(regression)
	if err != nil {
		t.Fatal(err)
	}
	if report.Families["go-local-logic"].PromotionEligible {
		t.Fatal("right-to-wrong family was eligible")
	}
}

func TestPairedGradeReportsFullFallbackTailAndInterventions(t *testing.T) {
	samples := make([]pairedSample, 0, 20)
	for index := 0; index < 20; index++ {
		sample := validPairedSample("go-local-logic", "retry", index, 100, 70)
		sample.AMR.FallbackUsed = index%2 == 0
		sample.AMR.RepairTailAIUNano = int64(index)
		sample.Direct.Interventions = 1
		sample.AMR.Interventions = 1
		samples = append(samples, sample)
	}
	report, err := gradePairedSamples(samples)
	if err != nil {
		t.Fatal(err)
	}
	family := report.Families["go-local-logic"]
	if family.FallbackCount != 10 || family.RepairTailP90Nano <= 0 || family.AMRInterventions != family.DirectInterventions {
		t.Fatalf("family=%+v", family)
	}
}

func validPairedSample(family, task string, seed int, directAIU, amrAIU int64) pairedSample {
	return pairedSample{
		TaskID: task, Seed: seed, Family: family,
		Direct: pairedArm{Correct: true, AIUNano: directAIU, LatencyMS: 100, TelemetryComplete: true, RouteMatch: true, OracleIntact: true, IsolationReady: true, ProofPresent: true},
		AMR:    pairedArm{Correct: true, AIUNano: amrAIU, LatencyMS: 120, TelemetryComplete: true, RouteMatch: true, OracleIntact: true, IsolationReady: true, ProofPresent: true},
	}
}
