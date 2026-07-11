package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParseReleaseArgs(t *testing.T) {
	opts, err := parse([]string{"local-release", "--root", "repo", "--json", "--dry-run"})
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if opts.command != "local-release" || opts.root != "repo" || !opts.json || !opts.dryRun {
		t.Fatalf("unexpected options: %+v", opts)
	}
}

func TestParseCoreList(t *testing.T) {
	opts, err := parse([]string{"core", "--root", "repo", "--list"})
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if opts.command != "core" || !opts.list {
		t.Fatalf("unexpected options: %+v", opts)
	}
}

func TestParseCoreVerbose(t *testing.T) {
	opts, err := parse([]string{"core", "--root", "repo", "--verbose"})
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if opts.command != "core" || !opts.verbose {
		t.Fatalf("unexpected options: %+v", opts)
	}
}

func TestParseRejectsJSONForCore(t *testing.T) {
	_, err := parse([]string{"core", "--json"})
	if err == nil {
		t.Fatal("expected --json to be rejected for core")
	}
}

func TestParseContextPacketAndProviderHygiene(t *testing.T) {
	opts, err := parse([]string{"context-packet", "--packet", "packet.json", "--json"})
	if err != nil || opts.packetPath != "packet.json" || !opts.json {
		t.Fatalf("opts=%+v err=%v", opts, err)
	}
	opts, err = parse([]string{"provider-hygiene", "--include-user"})
	if err != nil || !opts.includeUser {
		t.Fatalf("opts=%+v err=%v", opts, err)
	}
	if _, err := parse([]string{"context-packet"}); err == nil {
		t.Fatal("missing --packet passed")
	}
	if _, err := parse([]string{"provider-hygiene", "--packet", "packet.json"}); err == nil {
		t.Fatal("--packet on provider-hygiene passed")
	}
	if _, err := parse([]string{"core", "--include-user"}); err == nil {
		t.Fatal("--include-user on core passed")
	}
	opts, err = parse([]string{"execution-telemetry", "--telemetry", "usage.json", "--receipt", "receipt.json", "--evidence-envelope", "envelope.json"})
	if err != nil || opts.telemetryPath != "usage.json" || opts.receiptPath != "receipt.json" || opts.evidenceEnvelopePath != "envelope.json" {
		t.Fatalf("opts=%+v err=%v", opts, err)
	}
	if _, err := parse([]string{"execution-telemetry"}); err == nil {
		t.Fatal("missing --telemetry passed")
	}
	if _, err := parse([]string{"execution-telemetry", "--telemetry", "usage.json", "--receipt", "receipt.json"}); err == nil {
		t.Fatal("partial receipt evidence passed")
	}
	if _, err := parse([]string{"execution-telemetry", "--telemetry", "usage.json", "--receipt", "receipt.json", "--evidence-envelope", "envelope.json", "--host-attest"}); err == nil {
		t.Fatal("removed public host-attest signing oracle was accepted")
	}
}

func TestCoreListPrintsNativeChecks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")

	var out strings.Builder
	code := run([]string{"core", "--root", root, "--list"}, &out, &strings.Builder{})
	if code != 0 {
		t.Fatalf("expected list to pass, got %d", code)
	}
	if !strings.Contains(out.String(), "go-test") || !strings.Contains(out.String(), "go test ./...") || strings.Contains(out.String(), "kb-check.ps1 -All") {
		t.Fatalf("unexpected core list: %q", out.String())
	}
}

func TestCoreRunsDiscoveredCheck(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")
	writeFile(t, filepath.Join(root, "go.sum"), "")

	runner := func(root string, check Check) CheckResult {
		if check.Name != "go-test" {
			t.Fatalf("unexpected check: %s", check.Name)
		}
		return CheckResult{ExitCode: 0, Stdout: "ok\n"}
	}

	var out, errOut strings.Builder
	code := runCore(root, options{command: "core", root: root}, &out, &errOut, runner)
	if code != 0 {
		t.Fatalf("expected core to pass, got %d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "ok   go-test") || !strings.Contains(out.String(), "core: ok checks=1") {
		t.Fatalf("missing check output: %q", out.String())
	}
	if strings.Contains(out.String(), "ok\n") {
		t.Fatalf("default core should suppress passing check stdout: %q", out.String())
	}
}

func TestCoreVerbosePreservesPassingOutput(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")
	writeFile(t, filepath.Join(root, "go.sum"), "")

	runner := func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 0, Stdout: "raw ok\n"}
	}

	var out, errOut strings.Builder
	code := runCore(root, options{command: "core", root: root, verbose: true}, &out, &errOut, runner)
	if code != 0 {
		t.Fatalf("expected core to pass, got %d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "==> go-test") || !strings.Contains(out.String(), "raw ok") {
		t.Fatalf("verbose output did not preserve raw stdout: %q", out.String())
	}
	if strings.Contains(out.String(), "core: ok checks=") {
		t.Fatalf("verbose output should not print compact summary: %q", out.String())
	}
}

func TestCoreFailurePropagates(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")

	runner := func(root string, check Check) CheckResult {
		return CheckResult{ExitCode: 7, Stderr: "boom"}
	}

	var out, errOut strings.Builder
	code := runCore(root, options{command: "core", root: root}, &out, &errOut, runner)
	if code != 7 {
		t.Fatalf("expected exit 7, got %d", code)
	}
	if !strings.Contains(errOut.String(), "FAIL go-test") || !strings.Contains(errOut.String(), "boom") || !strings.Contains(errOut.String(), "check failed: go-test") {
		t.Fatalf("missing failure output: %q", errOut.String())
	}
}

func TestRunProcessCheckFailsClosedOnTimeout(t *testing.T) {
	if os.Getenv("KBCHECK_TIMEOUT_HELPER") == "1" {
		time.Sleep(5 * time.Second)
		return
	}
	t.Setenv("KBCHECK_TIMEOUT_HELPER", "1")
	result := runProcessCheck(t.TempDir(), Check{
		Name:    "timeout-helper",
		Args:    []string{os.Args[0], "-test.run=^TestRunProcessCheckFailsClosedOnTimeout$"},
		Timeout: 50 * time.Millisecond,
	})
	if result.ExitCode != 124 || !strings.Contains(result.Stderr, "timed out") {
		t.Fatalf("timeout did not fail closed: %+v", result)
	}
}

func TestRunProcessCheckFailsClosedOnOversizedOutput(t *testing.T) {
	if os.Getenv("KBCHECK_OUTPUT_HELPER") == "1" {
		fmt.Print(strings.Repeat("x", maxProcessCheckOutputBytes+1))
		return
	}
	t.Setenv("KBCHECK_OUTPUT_HELPER", "1")
	result := runProcessCheck(t.TempDir(), Check{
		Name: "output-helper",
		Args: []string{os.Args[0], "-test.run=^TestRunProcessCheckFailsClosedOnOversizedOutput$"},
	})
	if result.ExitCode != 125 || !strings.Contains(result.Stderr, "output exceeded") {
		t.Fatalf("oversized output did not fail closed: code=%d stderr=%q", result.ExitCode, result.Stderr)
	}
}

func TestRunProcessCheckTimeoutKillsGrandchild(t *testing.T) {
	if os.Getenv("KBCHECK_GRANDCHILD_CHILD") == "1" {
		time.Sleep(7 * time.Second)
		if err := os.WriteFile(os.Getenv("KBCHECK_GRANDCHILD_SENTINEL"), []byte("survived"), 0o600); err != nil {
			os.Exit(2)
		}
		return
	}
	if os.Getenv("KBCHECK_GRANDCHILD_HELPER") == "parent" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestRunProcessCheckTimeoutKillsGrandchild$")
		cmd.Env = append(os.Environ(), "KBCHECK_GRANDCHILD_CHILD=1")
		if err := cmd.Start(); err != nil {
			os.Exit(3)
		}
		if err := os.WriteFile(os.Getenv("KBCHECK_GRANDCHILD_READY"), []byte("started"), 0o600); err != nil {
			os.Exit(4)
		}
		time.Sleep(20 * time.Second)
		return
	}

	sentinel := filepath.Join(t.TempDir(), "grandchild-survived.txt")
	ready := filepath.Join(filepath.Dir(sentinel), "grandchild-started.txt")
	t.Setenv("KBCHECK_GRANDCHILD_HELPER", "parent")
	t.Setenv("KBCHECK_GRANDCHILD_SENTINEL", sentinel)
	t.Setenv("KBCHECK_GRANDCHILD_READY", ready)
	result := runProcessCheck(t.TempDir(), Check{
		Name:    "grandchild-timeout-helper",
		Args:    []string{os.Args[0], "-test.run=^TestRunProcessCheckTimeoutKillsGrandchild$"},
		Timeout: 5 * time.Second,
	})
	if result.ExitCode != 124 {
		t.Fatalf("grandchild helper did not time out: %+v", result)
	}
	if _, err := os.Stat(ready); err != nil {
		t.Fatalf("grandchild helper never proved it started: %v", err)
	}
	time.Sleep(3 * time.Second)
	if _, err := os.Stat(sentinel); err == nil {
		t.Fatalf("grandchild survived proof-gate containment and wrote %s", sentinel)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat sentinel: %v", err)
	}
}

func TestReleaseJSONReportsRequiredFailure(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module fixture\n")

	runner := func(root string, check Check) CheckResult {
		if check.Name == "kb-check-all" {
			return CheckResult{ExitCode: 3, Stderr: "core failed"}
		}
		return CheckResult{ExitCode: 0}
	}

	var out, errOut strings.Builder
	code := runRelease(root, options{command: "local-release", root: root, json: true}, &out, &errOut, runner)
	if code == 0 {
		t.Fatal("expected release to fail")
	}
	if !strings.Contains(out.String(), `"required_failures": 1`) {
		t.Fatalf("missing JSON failure count: %s", out.String())
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
