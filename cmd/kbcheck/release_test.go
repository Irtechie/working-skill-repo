package main

import (
	"path/filepath"
	"testing"
)

func TestReleaseChecksUseNativeCoreNotPSGate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")
	writeFile(t, filepath.Join(root, "scripts", "skill-sync-report.ps1"), "exit 0")

	checks, err := releaseChecks(root, "local-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatalf("releaseChecks returned error: %v", err)
	}
	if checks[0].Name != "kb-check-all" || checks[0].CommandString() != "kbcheck core" {
		t.Fatalf("expected native core release check, got %+v", checks[0])
	}
	for _, check := range checks {
		if check.Name == "kb-release-gate" || check.CommandString() == "scripts/kb-release-gate.ps1" {
			t.Fatalf("release gate must not delegate to kb-release-gate.ps1: %+v", check)
		}
	}
}

func TestLiveReleaseLabelsUnavailableCorpusExplicit(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")
	checks, err := releaseChecks(root, "live-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatalf("releaseChecks returned error: %v", err)
	}
	found := false
	for _, check := range checks {
		if check.Name == "live-codex-ghcp-corpus" {
			found = true
			run := invokeReleaseCheck(root, check, func(root string, check Check) CheckResult {
				t.Fatalf("unavailable check should not run")
				return CheckResult{}
			})
			if run.Status != "skipped-explicit" {
				t.Fatalf("expected skipped-explicit, got %+v", run)
			}
		}
	}
	if !found {
		t.Fatal("missing live corpus check")
	}
}

func TestReleaseFailsWhenRequiredSyncScriptMissing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")

	checks, err := releaseChecks(root, "local-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatalf("releaseChecks returned error: %v", err)
	}
	for _, check := range checks {
		if check.Name == "skill-sync-report" {
			run := invokeReleaseCheck(root, check, func(root string, check Check) CheckResult {
				t.Fatalf("unavailable required check should not run")
				return CheckResult{}
			})
			if run.Status != "skipped-explicit" || !run.Required {
				t.Fatalf("expected required skipped-explicit, got %+v", run)
			}
			return
		}
	}
	t.Fatal("missing skill-sync-report release check")
}
