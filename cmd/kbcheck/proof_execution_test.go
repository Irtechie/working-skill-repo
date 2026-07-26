package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestProofExecutionBudgetRunsOnceThenReuses(t *testing.T) {
	root := t.TempDir()
	registry := proofExecutionRegistry(t, root, "focused", "cli")
	receiptDir := filepath.Join(root, "receipts")
	runs := 0
	runner := func(root string, check Check) CheckResult {
		runs++
		return CheckResult{ExitCode: 0, Stdout: "pass"}
	}
	now := time.Now().UTC()

	first := executeProofGovernorPlan(root, registry, []string{"focused"}, receiptDir, runner, now)
	if first.Decision != proofGovernorRun || !first.OK || runs != 1 {
		t.Fatalf("first execution failed: %#v runs=%d", first, runs)
	}
	second := executeProofGovernorPlan(root, registry, []string{"focused"}, receiptDir, runner, now.Add(time.Second))
	if second.Decision != proofGovernorReuse || !second.OK || runs != 1 {
		t.Fatalf("unchanged execution was not reused: %#v runs=%d", second, runs)
	}
	audit, err := os.ReadFile(filepath.Join(receiptDir, ".proof-audit.jsonl"))
	if err != nil || strings.Count(strings.TrimSpace(string(audit)), "\n") != 1 {
		t.Fatalf("run and reuse decisions were not audited: err=%v audit=%q", err, audit)
	}
}

func TestProofExecutionBudgetBlocksIdenticalFailedReplayUntilInputChanges(t *testing.T) {
	root := t.TempDir()
	registry := proofExecutionRegistry(t, root, "full", "cli")
	receiptDir := filepath.Join(root, "receipts")
	runs := 0
	runner := func(root string, check Check) CheckResult {
		runs++
		return CheckResult{ExitCode: 1, Stderr: "failed"}
	}
	now := time.Now().UTC()

	first := executeProofGovernorPlan(root, registry, []string{"full"}, receiptDir, runner, now)
	if first.OK || runs != 1 {
		t.Fatalf("expected one failed attempt: %#v runs=%d", first, runs)
	}
	second := executeProofGovernorPlan(root, registry, []string{"full"}, receiptDir, runner, now.Add(time.Second))
	if second.Decision != proofGovernorBlock || runs != 1 || !containsProofGovernorReason(second.Reasons, "unchanged-attempt-already-recorded:full") {
		t.Fatalf("identical failure replay was not blocked: %#v runs=%d", second, runs)
	}

	writeProofGovernorFixture(t, root, "input.txt", "changed\n")
	third := executeProofGovernorPlan(root, registry, []string{"full"}, receiptDir, runner, now.Add(2*time.Second))
	if third.Decision != proofGovernorRun || runs != 2 {
		t.Fatalf("changed input should permit affected rerun: %#v runs=%d", third, runs)
	}
}

func TestProofExecutionBlocksAutomaticGUIClassesBeforeSpawn(t *testing.T) {
	for _, executionClass := range []string{"visible-browser", "native-gui"} {
		t.Run(executionClass, func(t *testing.T) {
			root := t.TempDir()
			registry := proofExecutionRegistry(t, root, "desktop", executionClass)
			receiptDir := filepath.Join(root, "receipts")
			runs := 0
			runner := func(root string, check Check) CheckResult {
				runs++
				return CheckResult{ExitCode: 0}
			}

			denied := executeProofGovernorPlan(root, registry, []string{"desktop"}, receiptDir, runner, time.Now().UTC())
			if denied.Decision != proofGovernorBlock || runs != 0 || !containsProofGovernorReason(denied.Reasons, "automatic-gui-execution-disabled:desktop") {
				t.Fatalf("automatic GUI execution was not blocked before spawn: %#v runs=%d", denied, runs)
			}
		})
	}
}

func TestProofGovernorCLIRemovesApprovalSurface(t *testing.T) {
	if _, err := parse([]string{"proof-approve"}); err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Fatalf("proof-approve should be removed, got %v", err)
	}
	_, err := parse([]string{
		"proof-run",
		"--registry", "registry.json",
		"--receipt-dir", "receipts",
		"--request", "desktop",
		"--approval", "approval.json",
	})
	if err == nil || !strings.Contains(err.Error(), "flag provided but not defined: -approval") {
		t.Fatalf("--approval should be removed, got %v", err)
	}
}

func TestProofExecutionWritesTimeoutReceiptAndNoGlobalPass(t *testing.T) {
	root := t.TempDir()
	registry := proofExecutionRegistry(t, root, "slow", "cli")
	receiptDir := filepath.Join(root, "receipts")
	runner := func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 124, Stderr: "timed out and process tree killed"}
	}

	result := executeProofGovernorPlan(root, registry, []string{"slow"}, receiptDir, runner, time.Now().UTC())
	if result.OK || result.ExitCode != 124 || !containsProofGovernorReason(result.Reasons, "check-timeout:slow") {
		t.Fatalf("timeout was not terminal partial evidence: %#v", result)
	}
	receipts, issues := loadProofGovernorReceipts(receiptDir)
	if len(issues) != 0 || len(receipts) != 1 || receipts[0].Result.Status != "timeout" || receipts[0].Result.ExitCode != 124 {
		t.Fatalf("timeout receipt missing: receipts=%#v issues=%#v", receipts, issues)
	}
}

func TestRepairPolicyRejectsUnconditionalReplayLanguage(t *testing.T) {
	root := proofExecutionRepoRoot(t)
	files := map[string][]string{
		".github/skills/kb-repair/SKILL.md": {
			"After each fix, re-run ALL checks",
		},
		".github/skills/kb-regression-snapshot/SKILL.md": {
			"against all previous snapshots",
		},
		".github/skills/kb-work/SKILL.md": {
			"Verify all previous snapshots under `.kb/snapshots/`",
		},
	}
	for relative, forbidden := range files {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, phrase := range forbidden {
			if strings.Contains(string(content), phrase) {
				t.Errorf("%s retains unconditional replay phrase %q", relative, phrase)
			}
		}
	}
}

func TestSnapshotSelectionRequiresScopeAndReusesMilestone(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("PowerShell snapshot contract is Windows-specific")
	}
	if _, err := os.Stat(filepath.Join(proofExecutionRepoRoot(t), ".github", "skills", "kb-regression-snapshot", "scripts", "kb-regression-snapshot.ps1")); err != nil {
		t.Fatal(err)
	}
	// The script contract is covered textually here; its functional no-scope and
	// milestone-reuse paths run in proof-governor-selftest after fixtures exist.
	content, err := os.ReadFile(filepath.Join(proofExecutionRepoRoot(t), ".github", "skills", "kb-regression-snapshot", "scripts", "kb-regression-snapshot.ps1"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	for _, required := range []string{"SnapshotId", "MilestoneId", "snapshot-verify: REUSE", "headless: true"} {
		if !strings.Contains(text, required) {
			t.Errorf("snapshot runner missing governed contract %q", required)
		}
	}
}

func proofExecutionRegistry(t *testing.T, root, id, class string) proofGovernorRegistry {
	t.Helper()
	writeProofGovernorFixture(t, root, "input.txt", "stable\n")
	writeProofGovernorFixture(t, root, "registry.json", `{"schema_version":1,"checks":[]}`)
	return proofGovernorRegistry{
		SchemaVersion: 1,
		Checks: []proofGovernorCheckSpec{{
			ID: id, Namespace: proofGovernorNamespace{Goal: "goal", Slice: "integrated", Run: "plan-run"},
			Command: []string{"fake-check", id}, Covers: []string{id}, Inputs: []string{"input.txt"},
			WorkingDir: ".", TimeoutMS: 1_000, ExpectedExit: 0,
			Environment: map[string]string{"runtime": "v1"}, ExecutionClass: class, MaxAgeSeconds: 300,
		}},
	}
}

func proofExecutionRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}
