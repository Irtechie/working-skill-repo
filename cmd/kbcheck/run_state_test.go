package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStateDetectsRouteOscillation(t *testing.T) {
	path := writeRouteHistory(t, `
{"route":"kb-plan","confidence":0.80}
{"route":"kb-work","confidence":0.80}
{"route":"kb-plan","confidence":0.80}
{"route":"kb-work","confidence":0.80}
`)
	result, err := validateRouteHistory(path)
	if err != nil {
		t.Fatalf("validateRouteHistory returned error: %v", err)
	}
	if result.OK || !hasRunStateIssue(result.Issues, "route-oscillation") {
		t.Fatalf("expected route oscillation issue, got %#v", result)
	}
}

func TestRunStateDetectsLowConfidenceNoProgress(t *testing.T) {
	path := writeRouteHistory(t, `
{"route":"kb-start","confidence":0.40}
{"route":"kb-brainstorm","confidence":0.42}
{"route":"kb-plan","confidence":0.44}
`)
	result, err := validateRouteHistory(path)
	if err != nil {
		t.Fatalf("validateRouteHistory returned error: %v", err)
	}
	if result.OK || !hasRunStateIssue(result.Issues, "low-confidence-no-progress") {
		t.Fatalf("expected low-confidence issue, got %#v", result)
	}
}

func TestRunStateAllowsProgressBetweenRoutes(t *testing.T) {
	path := writeRouteHistory(t, `
{"route":"kb-start","confidence":0.42}
{"route":"kb-plan","confidence":0.43,"progress_key":"manifest-created"}
{"route":"kb-work","confidence":0.44}
`)
	result, err := validateRouteHistory(path)
	if err != nil {
		t.Fatalf("validateRouteHistory returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected progress to reset low-confidence run, got %#v", result)
	}
}

func TestRunStateCommandRequiresHistory(t *testing.T) {
	var out, errOut strings.Builder
	code := run([]string{"run-state"}, &out, &errOut)
	if code == 0 || !strings.Contains(errOut.String(), "run-state requires --history") {
		t.Fatalf("expected missing history error, code=%d stderr=%s", code, errOut.String())
	}
}

func writeRouteHistory(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "route-history.jsonl")
	writeFile(t, path, strings.TrimLeft(body, "\n"))
	return path
}
