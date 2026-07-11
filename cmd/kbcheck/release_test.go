package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const testModelRoutingInitialPilotEvidence = "docs/results/2026-07-10-session-model-routing-initial-pilot.json"

func TestReleaseChecksUseNativeCoreNotPSGate(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")

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

func TestReleaseReportsCheckStartBeforeRunnerReturns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")
	var stdout, stderr strings.Builder
	runner := func(_ string, check Check) CheckResult {
		if !strings.Contains(stdout.String(), "running [required/deterministic-local] kb-check-all") {
			t.Fatalf("release did not expose running check before invoking %s: %q", check.Name, stdout.String())
		}
		return CheckResult{ExitCode: 0}
	}
	if code := runRelease(root, options{command: "local-release", root: root}, &stdout, &stderr, runner); code != 0 {
		t.Fatalf("release failed: code=%d stderr=%s", code, stderr.String())
	}
}

func TestLiveReleaseUsesNativeLiveCorpus(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")
	writeFile(t, filepath.Join(root, "evals", "route-complexity", "fixture.json"), "{}")
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
			if check.CommandString() != "kbcheck eval-run-live-corpus --runtime codex,ghcp" || check.Run == nil {
				t.Fatalf("expected native live corpus check, got %+v", check)
			}
		}
	}
	if !found {
		t.Fatal("missing live corpus check")
	}
}

func TestReleaseSkipsSyncForGenericRepoWithoutSkillConfig(t *testing.T) {
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
				t.Fatalf("unavailable generic-repo check should not run")
				return CheckResult{}
			})
			if run.Status != "skipped-explicit" || run.Required {
				t.Fatalf("expected optional skipped-explicit, got %+v", run)
			}
			return
		}
	}
	t.Fatal("missing skill-sync-report release check")
}

func TestReleaseRequiresNativeSyncForSkillRepo(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "demo", "SKILL.md"), "---\nname: demo\ndescription: demo\n---\n")
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
	  "sync_targets": [
	    {"id":"source","path":".github/skills","classification":"source","required":true},
	    {"id":"required","path":".github/skills","classification":"required","required":true}
	  ]
	}`)

	checks, err := releaseChecks(root, "local-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatalf("releaseChecks returned error: %v", err)
	}
	for _, check := range checks {
		if check.Name == "skill-sync-report" {
			if !check.Required || check.CommandString() != "kbcheck skill-sync-report" || check.Run == nil {
				t.Fatalf("expected required native sync check, got %+v", check)
			}
			return
		}
	}
	t.Fatal("missing skill-sync-report release check")
}

func TestLocalReleaseRequiresNativeModelRoutingGateWhenPilotEvidenceExists(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, filepath.FromSlash(testModelRoutingInitialPilotEvidence)), "{}\n")

	checks, err := releaseChecks(root, "local-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatalf("releaseChecks returned error: %v", err)
	}
	var matches []Check
	for _, check := range checks {
		if check.Name == "model-routing-initial-pilot" {
			matches = append(matches, check)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("expected one model-routing release check, got %d", len(matches))
	}
	check := matches[0]
	want := "kbcheck model-routing-release --cohort initial-pilot --evidence " + testModelRoutingInitialPilotEvidence
	if !check.Required || check.CommandString() != want || check.Run == nil {
		t.Fatalf("expected required native model-routing check %q, got %+v", want, check)
	}
}

func TestModelRoutingReleaseFailureBlocksLocalRelease(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, filepath.FromSlash(testModelRoutingInitialPilotEvidence)), "{}\n")

	runner := func(root string, check Check) CheckResult {
		if check.Name == "model-routing-initial-pilot" {
			if check.Run == nil {
				t.Fatal("model-routing release check must use the native runner")
			}
			return check.Run(root)
		}
		return CheckResult{ExitCode: 0}
	}
	var stdout, stderr strings.Builder
	code := runRelease(root, options{command: "local-release", root: root}, &stdout, &stderr, runner)
	if code == 0 {
		t.Fatalf("invalid model-routing evidence did not block local-release: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "failed [required/deterministic-local] model-routing-initial-pilot") {
		t.Fatalf("release output omitted required model-routing failure: %s", stdout.String())
	}
}

func TestGenericReleaseOmitsModelRoutingGateWhenFeatureIsAbsent(t *testing.T) {
	root := t.TempDir()
	for _, profile := range []string{"local-release", "live-release"} {
		checks, err := releaseChecks(root, profile, func(root string, check Check) CheckResult {
			return CheckResult{ExitCode: 0}
		})
		if err != nil {
			t.Fatalf("%s releaseChecks returned error: %v", profile, err)
		}
		for _, check := range checks {
			if check.Name == "model-routing-initial-pilot" {
				t.Fatalf("%s must remain contributor-safe when canonical pilot evidence is absent", profile)
			}
		}
	}
}

func TestModelRoutingFeatureMarkerRequiresGateWhenEvidenceIsMissing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, filepath.FromSlash(modelRoutingFeatureMarker)), "package modelrouting\n")

	checks, err := releaseChecks(root, "local-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatalf("releaseChecks returned error: %v", err)
	}
	for _, check := range checks {
		if check.Name == "model-routing-initial-pilot" {
			if !check.Required || check.Run == nil {
				t.Fatalf("feature marker did not install a required native evidence gate: %+v", check)
			}
			result := check.Run(root)
			if result.ExitCode == 0 || !strings.Contains(strings.ToLower(result.Stderr), "evidence") {
				t.Fatalf("missing canonical evidence did not fail closed: %+v", result)
			}
			return
		}
	}
	t.Fatal("feature marker failed to require model-routing gate")
}

func TestLiveReleaseIncludesModelRoutingGateExactlyOnce(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, filepath.FromSlash(testModelRoutingInitialPilotEvidence)), "{}\n")

	checks, err := releaseChecks(root, "live-release", func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0}
	})
	if err != nil {
		t.Fatalf("releaseChecks returned error: %v", err)
	}
	count := 0
	for _, check := range checks {
		if check.Name == "model-routing-initial-pilot" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("live-release must inherit, not duplicate, the model-routing check; got %d", count)
	}
}
