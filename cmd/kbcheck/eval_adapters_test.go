package main

import (
	"strings"
	"testing"
)

func TestEvalPromptWithholdsPrivateFixtureFields(t *testing.T) {
	t.Parallel()
	fixture := map[string]any{
		"id": "canary-case", "user_prompt": "route this request", "repo_state": map[string]any{"branch": "main"},
		"expected": map[string]any{"route": "SECRET-EXPECTED-ROUTE"},
		"guards": []any{"SECRET-GUARD"}, "scoring_metadata": map[string]any{"canary": "SECRET-SCORER"},
	}
	prompt := evalPrompt(fixture, "codex", "run-1")
	for _, private := range []string{"SECRET-EXPECTED-ROUTE", "SECRET-GUARD", "SECRET-SCORER", "expected_result \"pass\""} {
		if strings.Contains(prompt, private) {
			t.Errorf("private or self-scored fixture content leaked into prompt: %q", private)
		}
	}
	for _, public := range []string{"canary-case", "route this request", "\"branch\": \"main\"", "scorer, not you, determines pass or fail"} {
		if !strings.Contains(prompt, public) {
			t.Errorf("public prompt content missing: %q", public)
		}
	}
}

func TestEvalAdapterMarksSyntheticEvidence(t *testing.T) {
	t.Parallel()
	fixture := map[string]any{"id": "fixture", "expected": map[string]any{"route": "kb-fix"}}
	if got := stringValue(dryRunResult(fixture, "codex", "run")["evidence_kind"]); got != "synthetic" {
		t.Fatalf("dry run evidence kind=%q", got)
	}
	manifest := newRunManifest(".", "run", "codex", "dry-run", fixture)
	if got := stringValue(manifest["evidence_kind"]); got != "synthetic" {
		t.Fatalf("dry-run manifest evidence kind=%q", got)
	}
	if got := stringValue(newRunManifest(".", "run", "codex", "live", fixture)["evidence_kind"]); got != "live" {
		t.Fatalf("live manifest evidence kind=%q", got)
	}
}
