package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Irtechie/working-skill-repo/internal/reconcile"
)

func TestNoKBRepoDryRun(t *testing.T) {
	root := initPlainRepo(t)
	var stdout, stderr bytes.Buffer
	code := run([]string{
		"dry-run", "--repo", root, "--cutoff", "2026-08-01T16:30:00Z", "--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var result commandResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Mode != reconcile.ModeDryRun || result.Ledger == nil || result.Plan == nil {
		t.Fatalf("unexpected result: %#v", result)
	}
	if _, err := os.Stat(filepath.Join(root, ".kb")); !os.IsNotExist(err) {
		t.Fatal("dry-run wrote repository state")
	}
}

func TestNoKBRepoBinaryDryRun(t *testing.T) {
	root := initPlainRepo(t)
	binary := filepath.Join(t.TempDir(), "kbreconcile")
	if os.PathSeparator == '\\' {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build standalone CLI: %v\n%s", err, output)
	}
	command := exec.Command(binary,
		"dry-run", "--repo", root, "--cutoff", "2026-08-01T16:30:00Z", "--json",
	)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("standalone dry-run: %v\n%s", err, output)
	}
	var result commandResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "ok" || result.Mode != reconcile.ModeDryRun {
		t.Fatalf("unexpected standalone result: %#v", result)
	}
}

func TestPlanWritesDurableStableJSON(t *testing.T) {
	root := initPlainRepo(t)
	output := filepath.Join(t.TempDir(), "plan.json")
	args := []string{
		"plan", "--repo", root, "--output", output,
		"--cutoff", "2026-08-01T16:30:00Z", "--json",
	}
	var firstOut, firstErr bytes.Buffer
	if code := run(args, &firstOut, &firstErr); code != 0 {
		t.Fatalf("first code=%d stderr=%s", code, firstErr.String())
	}
	firstFile, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	var secondOut, secondErr bytes.Buffer
	if code := run(args, &secondOut, &secondErr); code != 0 {
		t.Fatalf("second code=%d stderr=%s", code, secondErr.String())
	}
	secondFile, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstFile, secondFile) || !bytes.Equal(firstOut.Bytes(), secondOut.Bytes()) {
		t.Fatal("same cutoff and repository did not produce stable JSON")
	}
	var result commandResult
	if err := json.Unmarshal(firstFile, &result); err != nil {
		t.Fatal(err)
	}
	if result.Plan.PolicyVersion == "" || result.Plan.LedgerFingerprint == "" {
		t.Fatalf("plan is not bound to durable inputs: %#v", result.Plan)
	}
}

func TestInventoryRejectsNonRepository(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: missing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"dry-run", "--repo", root, "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("non-repository inventory succeeded")
	}
	var result commandResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != "error" || result.Error == nil {
		t.Fatalf("missing fail-closed JSON error: %#v stderr=%s", result, stderr.String())
	}
}

func TestPlanRequiresOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"plan", "--repo", "."}, &stdout, &stderr); code != 2 {
		t.Fatalf("code=%d, want usage failure", code)
	}
}

func TestApplyVerifyStableJSONAndFailClosedInput(t *testing.T) {
	root := initPlainRepo(t)
	directory := t.TempDir()
	planPath := filepath.Join(directory, "plan.json")
	receiptPath := filepath.Join(directory, "receipt.json")
	planArgs := []string{
		"plan", "--repo", root, "--output", planPath,
		"--cutoff", "2026-08-01T16:30:00Z", "--json",
	}
	var planOut, planErr bytes.Buffer
	if code := run(planArgs, &planOut, &planErr); code != 0 {
		t.Fatalf("plan code=%d stderr=%s", code, planErr.String())
	}

	applyArgs := []string{"apply", "--input", planPath, "--receipt", receiptPath, "--session-id", "apply-test", "--json"}
	var firstOut, firstErr bytes.Buffer
	if code := run(applyArgs, &firstOut, &firstErr); code != 0 {
		t.Fatalf("apply code=%d stderr=%s", code, firstErr.String())
	}
	firstReceipt, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	var secondOut, secondErr bytes.Buffer
	if code := run(applyArgs, &secondOut, &secondErr); code != 0 {
		t.Fatalf("repeat apply code=%d stderr=%s", code, secondErr.String())
	}
	secondReceipt, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstOut.Bytes(), secondOut.Bytes()) || !bytes.Equal(firstReceipt, secondReceipt) {
		t.Fatal("repeated apply was not deterministic")
	}

	var verifyOut, verifyErr bytes.Buffer
	if code := run([]string{"verify", "--input", planPath, "--receipt", receiptPath, "--session-id", "apply-test", "--json"}, &verifyOut, &verifyErr); code != 0 {
		t.Fatalf("verify code=%d stderr=%s", code, verifyErr.String())
	}
	var verified commandResult
	if err := json.Unmarshal(verifyOut.Bytes(), &verified); err != nil {
		t.Fatal(err)
	}
	if verified.Verification == nil || verified.Receipt == nil ||
		verified.Verification.Status != "verified" {
		t.Fatalf("unexpected verify result: %#v", verified)
	}

	tampered := bytes.Replace(planOut.Bytes(), []byte(reconcile.DefaultPolicyVersion), []byte("tampered-policy/v1"), 1)
	tamperedPath := filepath.Join(directory, "tampered.json")
	if err := os.WriteFile(tamperedPath, tampered, 0o600); err != nil {
		t.Fatal(err)
	}
	var rejectedOut, rejectedErr bytes.Buffer
	if code := run([]string{"apply", "--input", tamperedPath, "--receipt", filepath.Join(directory, "rejected.json"), "--json"}, &rejectedOut, &rejectedErr); code == 0 {
		t.Fatal("tampered policy input was accepted")
	}
	var rejected commandResult
	if err := json.Unmarshal(rejectedOut.Bytes(), &rejected); err != nil {
		t.Fatal(err)
	}
	if rejected.Error == nil {
		t.Fatalf("tampered input did not return fail-closed JSON: %s", rejectedErr.String())
	}
}

func TestSemanticClaimCapabilityAndConformanceJSON(t *testing.T) {
	for _, mode := range []string{"claim-capability", "claim-conformance"} {
		var stdout, stderr bytes.Buffer
		if code := run([]string{mode, "--json"}, &stdout, &stderr); code != 0 {
			t.Fatalf("%s code=%d stderr=%s", mode, code, stderr.String())
		}
		var result commandResult
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.ClaimCapability == nil || result.ClaimCapability.SchemaVersion != 1 {
			t.Fatalf("%s missing stable capability: %#v", mode, result)
		}
		if result.ClaimCapability.ProtectedMutationAvailable || result.ClaimCapability.LiveProviderSupported {
			t.Fatalf("%s claimed unsupported live authority: %#v", mode, result.ClaimCapability)
		}
		if mode == "claim-conformance" && (result.Conformance == nil || result.Conformance.Status != "passed") {
			t.Fatalf("reference conformance failed: %#v", result)
		}
	}
}

func initPlainRepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGit(t, root, "init")
	runGit(t, root, "config", "user.email", "test@example.com")
	runGit(t, root, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(root, "file.txt"), []byte("plain\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", "file.txt")
	runGit(t, root, "commit", "-m", "initial")
	return root
}

func runGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", root}, args...)...)
	command.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func init() {
	now = func() time.Time {
		return time.Date(2026, 8, 1, 16, 30, 0, 0, time.UTC)
	}
}
