package main

import "testing"

func TestLearningAdoptionAcceptsMeasuredGain(t *testing.T) {
	t.Parallel()
	result := evaluateLearningAdoption(learningAdoptionInput{
		Cases: adoptionCases(20, map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true}, map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true, 5: true, 6: true, 7: true, 8: true, 9: true, 10: true, 11: true}),
	})
	if result.Status != "ADOPT_ELIGIBLE" || !result.OK || result.NetGain != 2 || result.RightToWrong != 0 {
		t.Fatalf("expected adoption eligibility, got %+v", result)
	}
}

func TestLearningAdoptionRejectsRightToWrongRegression(t *testing.T) {
	t.Parallel()
	baseline := map[int]bool{}
	candidate := map[int]bool{}
	for i := 0; i < 10; i++ {
		baseline[i] = true
		candidate[i] = true
	}
	delete(candidate, 0)
	for i := 10; i < 14; i++ {
		candidate[i] = true
	}
	result := evaluateLearningAdoption(learningAdoptionInput{Cases: adoptionCases(20, baseline, candidate)})
	if result.OK || result.RightToWrong != 1 || !containsString(result.Issues, "right_to_wrong_regression") {
		t.Fatalf("expected right-to-wrong rejection, got %+v", result)
	}
}

func TestLearningAdoptionRejectsLowSampleCount(t *testing.T) {
	t.Parallel()
	result := evaluateLearningAdoption(learningAdoptionInput{
		Cases: adoptionCases(19, map[int]bool{}, map[int]bool{0: true, 1: true, 2: true}),
	})
	if result.OK || !containsString(result.Issues, "insufficient_sample_count") {
		t.Fatalf("expected low-n rejection, got %+v", result)
	}
}

func TestLearningAdoptionRejectsMemorizedHoldoutString(t *testing.T) {
	t.Parallel()
	result := evaluateLearningAdoption(learningAdoptionInput{
		CandidateInstruction: "Always answer omega-secret-case with green.",
		HoldoutStrings:       []string{"omega-secret-case"},
		Cases:                adoptionCases(20, map[int]bool{}, map[int]bool{0: true, 1: true, 2: true}),
	})
	if result.OK || !containsString(result.Issues, "holdout_leakage") {
		t.Fatalf("expected holdout leakage rejection, got %+v", result)
	}
}

func adoptionCases(n int, baselineCorrect, candidateCorrect map[int]bool) []learningAdoptionCase {
	cases := make([]learningAdoptionCase, 0, n)
	for i := 0; i < n; i++ {
		cases = append(cases, learningAdoptionCase{
			ID:               string(rune('a' + i)),
			BaselineCorrect:  baselineCorrect[i],
			CandidateCorrect: candidateCorrect[i],
		})
	}
	return cases
}
