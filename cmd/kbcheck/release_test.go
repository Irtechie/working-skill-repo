package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReleaseChecksUseNativeCoreNotPSGate(t *testing.T) {
	t.Parallel()
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

func TestReleaseDiffCheckIncludesStagedChanges(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Release Gate Test"},
	} {
		if code, output := runGitCommand(root, args...); code != 0 {
			t.Fatalf("git %s failed: %s", strings.Join(args, " "), output)
		}
	}
	path := filepath.Join(root, "README.md")
	writeFile(t, path, "clean\n")
	if code, output := runGitCommand(root, "add", "README.md"); code != 0 {
		t.Fatalf("git add baseline failed: %s", output)
	}
	if code, output := runGitCommand(root, "commit", "-m", "baseline"); code != 0 {
		t.Fatalf("git commit baseline failed: %s", output)
	}

	writeFile(t, path, "trailing whitespace  \n")
	if code, output := runGitCommand(root, "add", "README.md"); code != 0 {
		t.Fatalf("git add candidate failed: %s", output)
	}
	writeFile(t, path, "clean\n")
	checks, err := releaseChecks(root, "local-release", runProcessCheck)
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range checks {
		if check.Name != "git-diff-check" {
			continue
		}
		if check.CommandString() != "git diff --cached --check" {
			t.Fatalf("unexpected diff check: %s", check.CommandString())
		}
		result := runProcessCheck(root, check)
		if result.ExitCode == 0 || !strings.Contains(result.Stdout+result.Stderr, "trailing whitespace") {
			t.Fatalf("staged whitespace was not rejected: %+v", result)
		}
		return
	}
	t.Fatal("missing git-diff-check")
}

func TestReleaseCandidateCoherenceRejectsStagedContractReversal(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Release Gate Test"},
	} {
		if code, output := runGitCommand(root, args...); code != 0 {
			t.Fatalf("git %s failed: %s", strings.Join(args, " "), output)
		}
	}
	path := filepath.Join(root, "README.md")
	writeFile(t, path, "![valid](docs/assets/kb-memory-loop.png)\n")
	if code, output := runGitCommand(root, "add", "README.md"); code != 0 {
		t.Fatalf("git add baseline failed: %s", output)
	}
	if code, output := runGitCommand(root, "commit", "-m", "baseline"); code != 0 {
		t.Fatalf("git commit baseline failed: %s", output)
	}

	writeFile(t, path, "![broken](docs/assets/missing.png)\n")
	if code, output := runGitCommand(root, "add", "README.md"); code != 0 {
		t.Fatalf("git add candidate failed: %s", output)
	}
	writeFile(t, path, "![valid](docs/assets/kb-memory-loop.png)\n")

	result := runCandidateCoherence(root, runProcessCheck)
	if result.ExitCode == 0 || !strings.Contains(result.Stderr, "unstaged tracked changes") {
		t.Fatalf("staged contract reversal was not rejected: %+v", result)
	}
}

func TestReleaseCandidateCoherenceCoversCandidateStates(t *testing.T) {
	t.Parallel()
	t.Run("staged only passes", func(t *testing.T) {
		root, path := releaseGitFixture(t)
		writeFile(t, path, "staged\n")
		if code, output := runGitCommand(root, "add", "README.md"); code != 0 {
			t.Fatalf("git add candidate failed: %s", output)
		}
		if result := runCandidateCoherence(root, runProcessCheck); result.ExitCode != 0 {
			t.Fatalf("staged-only candidate failed: %+v", result)
		}
	})
	t.Run("unstaged only fails", func(t *testing.T) {
		root, path := releaseGitFixture(t)
		writeFile(t, path, "unstaged\n")
		if result := runCandidateCoherence(root, runProcessCheck); result.ExitCode == 0 || !strings.Contains(result.Stderr, "unstaged tracked changes") {
			t.Fatalf("unstaged-only candidate was not rejected: %+v", result)
		}
	})
	t.Run("untracked fails", func(t *testing.T) {
		root, _ := releaseGitFixture(t)
		writeFile(t, filepath.Join(root, "new.txt"), "untracked\n")
		if result := runCandidateCoherence(root, runProcessCheck); result.ExitCode == 0 || !strings.Contains(result.Stderr, "untracked non-ignored files") {
			t.Fatalf("untracked candidate was not rejected: %+v", result)
		}
	})
	t.Run("git error propagates", func(t *testing.T) {
		result := runCandidateCoherence(t.TempDir(), func(_ string, check Check) CheckResult {
			return CheckResult{ExitCode: 2, Stderr: check.CommandString() + " failed"}
		})
		if result.ExitCode != 2 || !strings.Contains(result.Stderr, "git diff --quiet failed") {
			t.Fatalf("git command error was not propagated: %+v", result)
		}
	})
}

func releaseGitFixture(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Release Gate Test"},
	} {
		if code, output := runGitCommand(root, args...); code != 0 {
			t.Fatalf("git %s failed: %s", strings.Join(args, " "), output)
		}
	}
	path := filepath.Join(root, "README.md")
	writeFile(t, path, "baseline\n")
	if code, output := runGitCommand(root, "add", "README.md"); code != 0 {
		t.Fatalf("git add baseline failed: %s", output)
	}
	if code, output := runGitCommand(root, "commit", "-m", "baseline"); code != 0 {
		t.Fatalf("git commit baseline failed: %s", output)
	}
	return root, path
}

func TestReleaseReportsCheckStartBeforeRunnerReturns(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
