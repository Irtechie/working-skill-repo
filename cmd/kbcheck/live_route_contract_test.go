package main

import (
	"strings"
	"testing"
)

func TestLiveRouteAlternativesRetainProofRequirements(t *testing.T) {
	fixture := map[string]any{"expected": map[string]any{
		"route": "kb-fix", "accepted_routes": []any{"bounded direct edit"},
		"max_user_questions": 0, "artifacts": []any{"changed file"}, "proof": []any{"git diff --check"},
	}}
	for _, test := range []struct {
		name, route string
		proof       []string
		wantPass    bool
	}{
		{"canonical", "kb-fix", []string{"git diff --check"}, true},
		{"documented alternative", "bounded direct edit", []string{"git diff --check"}, true},
		{"wrong route", "kb-epic", []string{"git diff --check"}, false},
		{"missing proof", "bounded direct edit", nil, false},
		{"unrelated proof", "bounded direct edit", []string{"looks fine"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			proof := []any{}
			for _, item := range test.proof {
				proof = append(proof, item)
			}
			result := map[string]any{"id": "run", "fixture_id": "case", "actual": map[string]any{
				"route": test.route, "user_questions": 0, "artifacts": []any{"changed file"}, "proof": proof,
			}, "trace": map[string]any{"files_read": []string{}, "commands": []string{}}, "claim_checks": []any{}}
			issues, _, _ := scoreSkillEvalResult(t.TempDir(), result, map[string]map[string]any{"case": fixture})
			if (len(issues) == 0) != test.wantPass {
				t.Fatalf("issues=%+v", issues)
			}
		})
	}
}

func TestLiveResponseVocabularyIsNotFixtureDerived(t *testing.T) {
	first := evalPrompt(map[string]any{"id": "case", "prompt": "public task", "expected": map[string]any{"route": "SECRET-ROUTE", "artifacts": []string{"SECRET-ARTIFACT"}, "proof": []string{"SECRET-PROOF"}}}, "codex", "run")
	second := evalPrompt(map[string]any{"id": "case", "prompt": "public task", "expected": map[string]any{"route": "DIFFERENT-ANSWER"}}, "codex", "run")
	if first != second || strings.Contains(first, "SECRET-") {
		t.Fatal("private expectation affected prompt")
	}
	for _, contract := range []string{"Artifact categories:", "Proof categories:", "not a checklist", "claim_checks as []", "ONLY the JSON object"} {
		if !strings.Contains(first, contract) {
			t.Fatalf("missing public contract: %s", contract)
		}
	}
}
