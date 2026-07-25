package main

import (
	"errors"
	"strings"
	"testing"
)

func TestContainmentUnsupportedBlocksRequireReady(t *testing.T) {
	report := inspectContainment(func() (string, error) {
		return "", errors.New("no sandbox runtime")
	})
	if report.Ready || !strings.Contains(strings.Join(report.Issues, " "), "no sandbox runtime") {
		t.Fatalf("report=%+v", report)
	}
	if err := report.RequireReady(); err == nil {
		t.Fatal("unsupported containment passed require-ready")
	}
}

func TestProofEnvironmentExcludesCredentialsAndProviderState(t *testing.T) {
	env := proofEnvironment([]string{
		"PATH=C:\\tools",
		"SYSTEMROOT=C:\\Windows",
		"HOME=C:\\Users\\person",
		"GH_TOKEN=secret",
		"COPILOT_PROVIDER_API_KEY=secret",
	})
	joined := strings.Join(env, "\n")
	if strings.Contains(joined, "secret") || strings.Contains(joined, "GH_TOKEN") ||
		strings.Contains(joined, "COPILOT_") || strings.Contains(joined, "HOME=") {
		t.Fatalf("unsafe proof environment: %s", joined)
	}
	if !strings.Contains(joined, "PATH=") || !strings.Contains(joined, "SYSTEMROOT=") {
		t.Fatalf("required process environment missing: %s", joined)
	}
}
