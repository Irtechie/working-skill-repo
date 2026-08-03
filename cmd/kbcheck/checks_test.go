package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDiscoverPackageChecks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"scripts":{"lint":"eslint .","test":"vitest","unused":"noop"}}`)
	writeFile(t, filepath.Join(root, "pnpm-lock.yaml"), "")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"lint", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
	if checks[0].Args[0] != "pnpm" {
		t.Fatalf("expected pnpm runner, got %v", checks[0].Args)
	}
}

func TestGoTestsIsolateChildProcessOwnersAndPropagateFailure(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module example.test/fixture\n\ngo 1.23\n")
	writeFile(t, filepath.Join(root, "pkg", "regular.go"), "package regular\n")
	writeFile(t, filepath.Join(root, "pkg", "regular_test.go"), "package regular\n\nimport \"testing\"\n\nfunc TestRegular(t *testing.T) {}\n")
	checkTest := filepath.Join(root, "cmd", "kbcheck", "check_test.go")
	writeFile(t, checkTest, "package kbcheck\n\nimport \"testing\"\n\nfunc TestCheck(t *testing.T) {}\n")
	routerTest := filepath.Join(root, "cmd", "kbrouter", "router_test.go")
	writeFile(t, routerTest, "package kbrouter\n\nimport \"testing\"\n\nfunc TestRouter(t *testing.T) { t.Fatal(\"router failure\") }\n")

	if result := runGoTestsWithProcessIsolation(root); result.ExitCode == 0 {
		t.Fatal("router package failure was not propagated")
	}

	writeFile(t, routerTest, "package kbrouter\n\nimport \"testing\"\n\nfunc TestRouter(t *testing.T) {}\n")
	writeFile(t, checkTest, "package kbcheck\n\nimport \"testing\"\n\nfunc TestCheck(t *testing.T) { t.Fatal(\"kbcheck failure\") }\n")
	if result := runGoTestsWithProcessIsolation(root); result.ExitCode == 0 {
		t.Fatal("kbcheck package failure was not propagated")
	}

	writeFile(t, checkTest, "package kbcheck\n\nimport \"testing\"\n\nfunc TestCheck(t *testing.T) {}\n")
	if result := runGoTestsWithProcessIsolation(root); result.ExitCode != 0 {
		t.Fatalf("isolated Go tests failed after repair: exit=%d stdout=%s stderr=%s", result.ExitCode, result.Stdout, result.Stderr)
	}
}

func TestGoTestPackagePartitionIsolatesOnlyChildProcessOwners(t *testing.T) {
	packages := []string{
		"example.test/fixture/cmd/kbcheck",
		"example.test/fixture/cmd/kbrouter",
		"example.test/fixture/cmd/other",
		"example.test/fixture/internal/modelrouting",
	}
	regular, isolated := partitionGoTestPackages(packages)
	if want := []string{
		"example.test/fixture/cmd/other",
		"example.test/fixture/internal/modelrouting",
	}; !reflect.DeepEqual(regular, want) {
		t.Fatalf("regular=%v want=%v", regular, want)
	}
	if want := []string{
		"example.test/fixture/cmd/kbcheck",
		"example.test/fixture/cmd/kbrouter",
	}; !reflect.DeepEqual(isolated, want) {
		t.Fatalf("isolated=%v want=%v", isolated, want)
	}
}

func TestIsolatedGoTestArgsBoundThePackageTestBinary(t *testing.T) {
	args := isolatedGoTestArgs("example.test/fixture/cmd/kbcheck")
	if len(args) != 5 || args[0] != "test" || args[1] != "-buildvcs=false" ||
		args[2] != "-timeout="+(defaultProcessCheckTimeout-processCheckTerminationWait).String() ||
		args[3] != goTestParallelFlag() ||
		args[4] != "example.test/fixture/cmd/kbcheck" {
		t.Fatalf("isolated args=%v", args)
	}
}

func TestGoTestParallelismTracksMemoryHeadroomNotCPUCount(t *testing.T) {
	// Go defaults -parallel to GOMAXPROCS. These suites fork git subprocesses
	// per test, so concurrency must follow memory available to new processes.
	got := goTestParallelism()
	cpus := runtime.NumCPU()
	if got < 1 || got > cpus {
		t.Fatalf("parallelism %d outside [1,%d]", got, cpus)
	}

	available, known := availableProcessMemoryBytes()
	if !known {
		if want := min(cpus, goTestFallbackParallelism); got != want {
			t.Fatalf("unknown headroom should fall back to %d, got %d", want, got)
		}
	} else {
		affordable := int(available / goTestMemoryBudgetPerTest)
		if affordable < 1 {
			affordable = 1
		}
		if want := min(cpus, affordable); got != want {
			t.Fatalf("headroom %d bytes implies %d, got %d", available, want, got)
		}
	}

	if flag := goTestParallelFlag(); flag != "-parallel="+strconv.Itoa(got) {
		t.Fatalf("flag %q does not carry parallelism %d", flag, got)
	}
}

func TestAvailableProcessMemoryIsPlausibleWhenReported(t *testing.T) {
	available, known := availableProcessMemoryBytes()
	if !known {
		t.Skip("platform does not report process memory headroom")
	}
	// A host that reports headroom but claims less than 16 MiB is returning
	// junk, which would silently pin the suite to -parallel=1 forever.
	if available < 16<<20 {
		t.Fatalf("implausible headroom %d bytes", available)
	}
}

func TestUncontainedRunnerBoundsTimeoutOverflowAndInheritedPipes(t *testing.T) {
	switch os.Getenv("KBCHECK_UNCONTAINED_HELPER") {
	case "timeout":
		time.Sleep(5 * time.Second)
		return
	case "overflow":
		fmt.Print(strings.Repeat("x", maxProcessCheckOutputBytes+1))
		time.Sleep(5 * time.Second)
		return
	case "pipe-parent":
		child := exec.Command(os.Args[0], "-test.run=^TestUncontainedRunnerBoundsTimeoutOverflowAndInheritedPipes$")
		child.Env = append(os.Environ(), "KBCHECK_UNCONTAINED_HELPER=pipe-child")
		child.Stdout = os.Stdout
		child.Stderr = os.Stderr
		if err := child.Start(); err != nil {
			os.Exit(2)
		}
		return
	case "pipe-child":
		for range 20 {
			fmt.Print("retained-pipe-output")
			time.Sleep(25 * time.Millisecond)
		}
		return
	}

	run := func(mode string, timeout time.Duration) (CheckResult, time.Duration) {
		t.Helper()
		t.Setenv("KBCHECK_UNCONTAINED_HELPER", mode)
		start := time.Now()
		result := runCommandWithoutOuterContainment(
			t.TempDir(), timeout, 100*time.Millisecond,
			os.Args[0], "-test.run=^TestUncontainedRunnerBoundsTimeoutOverflowAndInheritedPipes$",
		)
		return result, time.Since(start)
	}

	if result, elapsed := run("timeout", 50*time.Millisecond); result.ExitCode != 124 || elapsed > 2*time.Second {
		t.Fatalf("timeout result=%+v elapsed=%s", result, elapsed)
	}
	if result, elapsed := run("overflow", time.Second); result.ExitCode != 125 || elapsed > 2*time.Second {
		t.Fatalf("overflow result=%+v elapsed=%s", result, elapsed)
	}
	if result, elapsed := run("pipe-parent", time.Second); result.ExitCode == 0 || elapsed > 2*time.Second {
		t.Fatalf("inherited-pipe result=%+v elapsed=%s", result, elapsed)
	}
}

func TestDiscoverSkillRepoChecksIncludesNativeValidators(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-check", "SKILL.md"), "---\nname: kb-check\ndescription: test\n---\n")
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{
		"context-packet-selftest",
		"cross-model-benchmark-validate",
		"dishonest-completion-selftest",
		"execution-telemetry-selftest",
		"graph-routing-eval",
		"graph-routing-lifecycle-selftest",
		"kb-doctor-selftest",
		"kb-pipeline-selftest",
		"kb-release-gate-selftest",
		"kb-run-state-selftest",
		"kb-work-ready-set-selftest",
		"kb-work-scope-lease-selftest",
		"kb-work-slice-lease-selftest",
		"manifest-contract-selftest",
		"marketplace-promotion-selftest",
		"plan-worktree-lifecycle-selftest",
		"proof-governor-selftest",
		"provider-hygiene",
		"provider-hygiene-selftest",
		"review-reference-guard",
		"route-complexity-eval",
		"skill-eval",
		"skill-eval-baseline-selftest",
		"skill-eval-codex-dry-run",
		"skill-eval-ghcp-dry-run",
		"skill-eval-manifest-selftest",
		"skill-eval-observed-trace-dry-run",
		"skill-eval-quality",
		"skill-guidance",
		"skill-lint",
		"skill-marketplace-firebreak",
		"skill-marketplace-firebreak-selftest",
		"skill-surface-minimality",
		"skill-surface-minimality-selftest",
		"skill-surface-report",
		"workflow-governor-selftest",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
}

func TestDiscoverNestedDotnetProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "App", "App.csproj"), "<Project></Project>")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"dotnet-test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
}

func checkNames(checks []Check) []string {
	names := make([]string, 0, len(checks))
	for _, check := range checks {
		names = append(names, check.Name)
	}
	return names
}

// TestKBNativeRootsRecognized is the protected oracle for slice-015.
// It asserts the harness reads observations from .kb/observations.jsonl (kb-native
// ephemeral root) and does NOT require .atv/ to be present.
// RED: fails against pre-slice-015 code because minimality reads .atv/observations.jsonl.
// GREEN: passes after slice-015 changes minimality to read .kb/observations.jsonl.
func TestKBNativeRootsRecognized(t *testing.T) {
	root := t.TempDir()
	// Write a skill with evidence ONLY in .kb/observations.jsonl (kb-native root).
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-skill", "SKILL.md"),
		"---\nname: kb-skill\ndescription: test kb-native skill\n---\n# KB Skill\n")
	writeFile(t, filepath.Join(root, ".kb", "observations.jsonl"),
		`{"tool":"kb-skill","result":"used"}`+"\n")

	// .atv/ must NOT exist — confirm the harness does not require it.
	if _, err := os.Stat(filepath.Join(root, ".atv")); err == nil {
		t.Fatal("test setup error: .atv/ should not exist in the temp dir")
	}

	report, err := computeMinimality(root, ".github/skills", ".github/agents", 6)
	if err != nil {
		t.Fatalf("computeMinimality returned error: %v", err)
	}

	var found *minimalityRow
	for i := range report.SkillClassifications {
		if report.SkillClassifications[i].Name == "kb-skill" {
			found = &report.SkillClassifications[i]
			break
		}
	}
	if found == nil {
		t.Fatal("kb-skill not found in minimality report")
	}
	// The skill must have runtime evidence sourced from .kb/observations.jsonl.
	if found.EvidenceClass != "runtime" {
		t.Fatalf("expected kb-skill to have runtime evidence from .kb/observations.jsonl, got EvidenceClass=%q (classification=%q); .atv/observations.jsonl must NOT be required",
			found.EvidenceClass, found.Classification)
	}
}
