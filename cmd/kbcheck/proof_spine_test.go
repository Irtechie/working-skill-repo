package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProofAcceptsRedThenGreenTrace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	tracePath := filepath.Join(root, ".kb", "trace.jsonl")
	statePath := filepath.Join(root, "state.txt")
	check := helperStateCheck(statePath)

	if err := os.WriteFile(statePath, []byte("fail"), 0o644); err != nil {
		t.Fatalf("write red state: %v", err)
	}
	digest, err := proofCheckDigest(root, check)
	if err != nil {
		t.Fatalf("digest check: %v", err)
	}

	redResult := proofSense(root, check)
	if redResult.OK {
		t.Fatal("expected red helper check to fail")
	}
	if _, err := appendProofTrace(tracePath, "sense", digest, redResult.OK, redResult.Signal, redResult.Evidence); err != nil {
		t.Fatalf("append red: %v", err)
	}
	if err := os.WriteFile(statePath, []byte("pass"), 0o644); err != nil {
		t.Fatalf("write green state: %v", err)
	}
	greenResult := proofSense(root, check)
	if !greenResult.OK {
		t.Fatalf("expected green helper check to pass: %+v", greenResult)
	}
	if _, err := appendProofTrace(tracePath, "sense", digest, greenResult.OK, greenResult.Signal, greenResult.Evidence); err != nil {
		t.Fatalf("append green: %v", err)
	}

	gate, err := proofAccept(root, tracePath, check)
	if err != nil {
		t.Fatalf("accept returned error: %v", err)
	}
	if !gate.OK || !gate.TraceIntact || !gate.SawRed || !gate.GreenAfterRed || !gate.CurrentlyGreen {
		t.Fatalf("expected failure-first accept, got %+v", gate)
	}
}

func TestProofRejectsVacuousGreenTrace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	tracePath := filepath.Join(root, ".kb", "trace.jsonl")
	check := helperCommandCheck("pass")
	digest, err := proofCheckDigest(root, check)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	result := proofSense(root, check)
	if !result.OK {
		t.Fatalf("expected green helper check to pass: %+v", result)
	}
	if _, err := appendProofTrace(tracePath, "sense", digest, result.OK, result.Signal, result.Evidence); err != nil {
		t.Fatalf("append green: %v", err)
	}

	gate, err := proofAccept(root, tracePath, check)
	if err != nil {
		t.Fatalf("accept returned error: %v", err)
	}
	if gate.OK || gate.SawRed {
		t.Fatalf("vacuous green must be rejected, got %+v", gate)
	}
}

func TestProofRejectsTamperedTrace(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	tracePath := filepath.Join(root, ".kb", "trace.jsonl")
	check := helperCommandCheck("pass")
	digest, err := proofCheckDigest(root, check)
	if err != nil {
		t.Fatalf("digest: %v", err)
	}
	if _, err := appendProofTrace(tracePath, "sense", digest, false, "exit 1", "red"); err != nil {
		t.Fatalf("append red: %v", err)
	}
	if _, err := appendProofTrace(tracePath, "sense", digest, true, "exit 0", "green"); err != nil {
		t.Fatalf("append green: %v", err)
	}
	content, err := os.ReadFile(tracePath)
	if err != nil {
		t.Fatalf("read trace: %v", err)
	}
	tampered := strings.Replace(string(content), `"ok":false`, `"ok":true`, 1)
	if tampered == string(content) {
		t.Fatalf("test did not tamper trace row: %s", content)
	}
	if err := os.WriteFile(tracePath, []byte(tampered), 0o644); err != nil {
		t.Fatalf("write tampered trace: %v", err)
	}

	verify, err := verifyProofTrace(tracePath)
	if err != nil {
		t.Fatalf("verify returned error: %v", err)
	}
	if verify.OK || verify.BrokenAt == nil || *verify.BrokenAt != 0 {
		t.Fatalf("expected tamper at row 0, got %+v", verify)
	}
	gate, err := proofAccept(root, tracePath, check)
	if err != nil {
		t.Fatalf("accept returned error: %v", err)
	}
	if gate.OK || gate.TraceIntact {
		t.Fatalf("tampered trace must reject accept, got %+v", gate)
	}
}

func TestProofDigestChangesWhenCheckerScriptChanges(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	checker := filepath.Join(root, "checker.txt")
	if err := os.WriteFile(checker, []byte("v1"), 0o644); err != nil {
		t.Fatalf("write checker: %v", err)
	}
	check := ProofCheck{Kind: "command_exit", Target: []string{checker}, Expect: 0}
	first, err := proofCheckDigest(root, check)
	if err != nil {
		t.Fatalf("first digest: %v", err)
	}
	if err := os.WriteFile(checker, []byte("v2"), 0o644); err != nil {
		t.Fatalf("rewrite checker: %v", err)
	}
	second, err := proofCheckDigest(root, check)
	if err != nil {
		t.Fatalf("second digest: %v", err)
	}
	if first == second {
		t.Fatalf("checker digest should change when checker file changes: %s", first)
	}
}

func TestProofDigestChangesWhenTimeoutChanges(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	check := helperCommandCheck("pass")
	first, err := proofCheckDigest(root, check)
	if err != nil {
		t.Fatalf("first digest: %v", err)
	}
	check.TimeoutMS = 100
	second, err := proofCheckDigest(root, check)
	if err != nil {
		t.Fatalf("second digest: %v", err)
	}
	if first == second {
		t.Fatalf("check digest should change when timeout changes: %s", first)
	}
}

func TestProofRejectsFractionalExpectedExit(t *testing.T) {
	t.Parallel()
	if _, err := proofExpectedExit(float64(1.5)); err == nil {
		t.Fatal("expected fractional exit code to be rejected")
	}
}

func TestProofSenseTimeoutFailsCleanly(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	check := helperCommandCheck("sleep")
	check.TimeoutMS = 25
	result := proofSense(root, check)
	if result.OK || !strings.Contains(result.Signal, "timeout") {
		t.Fatalf("expected timeout failure, got %+v", result)
	}
}

func helperCommandCheck(mode string) ProofCheck {
	return ProofCheck{
		Kind:   "command_exit",
		Target: []string{os.Args[0], "-test.run=TestProofSpineHelperProcess", "--", mode},
		Expect: 0,
	}
}

func helperStateCheck(path string) ProofCheck {
	return ProofCheck{
		Kind:   "command_exit",
		Target: []string{os.Args[0], "-test.run=TestProofSpineHelperProcess", "--", "state", path},
		Expect: 0,
	}
}

func TestProofSpineHelperProcess(t *testing.T) {
	t.Parallel()
	mode := ""
	modeArg := ""
	for i, arg := range os.Args {
		if arg == "--" && i+1 < len(os.Args) {
			mode = os.Args[i+1]
			if i+2 < len(os.Args) {
				modeArg = os.Args[i+2]
			}
			break
		}
	}
	switch mode {
	case "":
		return
	case "pass":
		os.Exit(0)
	case "fail":
		os.Exit(1)
	case "sleep":
		time.Sleep(500 * time.Millisecond)
		os.Exit(0)
	case "state":
		content, err := os.ReadFile(modeArg)
		if err != nil || strings.TrimSpace(string(content)) != "pass" {
			os.Exit(1)
		}
		os.Exit(0)
	default:
		os.Exit(2)
	}
}
