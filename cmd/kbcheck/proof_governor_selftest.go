package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type proofGovernorFixtureCorpus struct {
	SchemaVersion int                            `json:"schema_version"`
	Scenarios     []proofGovernorFixtureScenario `json:"scenarios"`
}

type proofGovernorFixtureScenario struct {
	ID               string `json:"id"`
	ExpectedDecision string `json:"expected_decision"`
	ExpectedLaunches int    `json:"expected_launches"`
}

type proofGovernorObservedScenario struct {
	Decision string
	Launches int
}

func runProofGovernorSelftest(root string, stdout, stderr io.Writer) int {
	corpusPath := filepath.Join(root, "evals", "proof-governor", "fixtures.json")
	content, err := os.ReadFile(corpusPath)
	if err != nil {
		fmt.Fprintf(stderr, "read proof-governor fixtures: %v\n", err)
		return 1
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	var corpus proofGovernorFixtureCorpus
	if err := decoder.Decode(&corpus); err != nil {
		fmt.Fprintf(stderr, "parse proof-governor fixtures: %v\n", err)
		return 1
	}
	if corpus.SchemaVersion != proofGovernorSchemaVersion {
		fmt.Fprintln(stderr, "proof-governor fixtures use unsupported schema")
		return 1
	}
	observed, err := observeProofGovernorScenarios(root)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(corpus.Scenarios) != len(observed) {
		fmt.Fprintf(stderr, "proof-governor fixture coverage mismatch: fixtures=%d observed=%d\n", len(corpus.Scenarios), len(observed))
		return 1
	}
	for _, scenario := range corpus.Scenarios {
		got, ok := observed[scenario.ID]
		if !ok {
			fmt.Fprintf(stderr, "proof-governor scenario is not implemented: %s\n", scenario.ID)
			return 1
		}
		if got.Decision != scenario.ExpectedDecision || got.Launches != scenario.ExpectedLaunches {
			fmt.Fprintf(stderr, "%s: decision=%s launches=%d want decision=%s launches=%d\n",
				scenario.ID, got.Decision, got.Launches, scenario.ExpectedDecision, scenario.ExpectedLaunches)
			return 1
		}
	}
	fmt.Fprintf(stdout, "proof-governor selftest: passed scenarios=%d gui_launches=0\n", len(observed))
	return 0
}

func observeProofGovernorScenarios(repoRoot string) (map[string]proofGovernorObservedScenario, error) {
	observed := map[string]proofGovernorObservedScenario{}
	now := time.Now().UTC()

	root, cleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return nil, err
	}
	defer cleanup()
	registry := proofGovernorSelftestRegistry("plan-run", []string{"input.txt"})
	if err := writeProofGovernorSelftestRegistry(root, registry); err != nil {
		return nil, err
	}
	receipts := filepath.Join(root, "receipts")
	launches := 0
	passRunner := func(root string, check Check) CheckResult {
		launches++
		return CheckResult{ExitCode: 0}
	}
	full := executeProofGovernorPlan(root, registry, []string{"full"}, receipts, passRunner, now)
	if !full.OK || launches != 1 {
		return nil, fmt.Errorf("proof-governor full setup failed: %#v launches=%d", full, launches)
	}
	before := launches
	reused := executeProofGovernorPlan(root, registry, []string{"focused"}, receipts, passRunner, now.Add(time.Second))
	observed["passing-superset-reuse"] = proofGovernorObservedScenario{Decision: reused.Decision, Launches: launches - before}
	if err := os.WriteFile(filepath.Join(root, "unrelated.txt"), []byte("changed\n"), 0o644); err != nil {
		return nil, err
	}
	before = launches
	unrelated := executeProofGovernorPlan(root, registry, []string{"focused"}, receipts, passRunner, now.Add(2*time.Second))
	observed["unrelated-input-change"] = proofGovernorObservedScenario{Decision: unrelated.Decision, Launches: launches - before}
	if err := os.WriteFile(filepath.Join(root, "input.txt"), []byte("changed\n"), 0o644); err != nil {
		return nil, err
	}
	before = launches
	relevant := executeProofGovernorPlan(root, registry, []string{"focused"}, receipts, passRunner, now.Add(3*time.Second))
	observed["relevant-input-change"] = proofGovernorObservedScenario{Decision: relevant.Decision, Launches: launches - before}
	if len(relevant.ReceiptPaths) != 1 {
		return nil, fmt.Errorf("relevant rerun did not emit a receipt")
	}
	latest, issues := loadProofGovernorReceipts(receipts)
	if len(issues) != 0 || len(latest) < 2 {
		return nil, fmt.Errorf("proof receipt corpus invalid: %v", issues)
	}
	last := latest[len(latest)-1]
	if last.CheckSemanticsSHA256 == "" || last.RelevantInputsSHA256 == "" || last.EnvironmentSHA256 == "" {
		return nil, fmt.Errorf("proof receipt omits semantic digest fields")
	}

	semanticRegistry := registry
	semanticRegistry.Checks[1].Command = []string{"fake", "focused", "--changed"}
	if err := writeProofGovernorSelftestRegistry(root, semanticRegistry); err != nil {
		return nil, err
	}
	before = launches
	semantic := executeProofGovernorPlan(root, semanticRegistry, []string{"focused"}, receipts, passRunner, now.Add(4*time.Second))
	observed["semantic-command-drift"] = proofGovernorObservedScenario{Decision: semantic.Decision, Launches: launches - before}

	sharedRoot, sharedCleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return nil, err
	}
	defer sharedCleanup()
	sharedRegistry := proofGovernorSelftestRegistry("plan-run", []string{"input.txt"})
	if err := writeProofGovernorSelftestRegistry(sharedRoot, sharedRegistry); err != nil {
		return nil, err
	}
	sharedReceipts := filepath.Join(sharedRoot, "receipts")
	sharedLaunches := 0
	sharedRunner := func(root string, check Check) CheckResult {
		sharedLaunches++
		return CheckResult{ExitCode: 0}
	}
	if result := executeProofGovernorPlan(sharedRoot, sharedRegistry, []string{"full", "focused"}, sharedReceipts, sharedRunner, now); !result.OK {
		return nil, fmt.Errorf("shared dependency setup failed: %#v", result)
	}
	if err := os.WriteFile(filepath.Join(sharedRoot, "input.txt"), []byte("shared changed\n"), 0o644); err != nil {
		return nil, err
	}
	before = sharedLaunches
	shared := executeProofGovernorPlan(sharedRoot, sharedRegistry, []string{"focused"}, sharedReceipts, sharedRunner, now.Add(time.Second))
	observed["shared-dependency-change"] = proofGovernorObservedScenario{Decision: shared.Decision, Launches: sharedLaunches - before}

	failureRoot, failureCleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return nil, err
	}
	defer failureCleanup()
	failureRegistry := proofGovernorSingleSelftestRegistry("failed", "cli", "plan-run")
	if err := writeProofGovernorSelftestRegistry(failureRoot, failureRegistry); err != nil {
		return nil, err
	}
	failureReceipts := filepath.Join(failureRoot, "receipts")
	failureLaunches := 0
	failRunner := func(root string, check Check) CheckResult {
		failureLaunches++
		return CheckResult{ExitCode: 1}
	}
	_ = executeProofGovernorPlan(failureRoot, failureRegistry, []string{"failed"}, failureReceipts, failRunner, now)
	before = failureLaunches
	failedReplay := executeProofGovernorPlan(failureRoot, failureRegistry, []string{"failed"}, failureReceipts, failRunner, now.Add(time.Second))
	observed["failed-replay-ceiling"] = proofGovernorObservedScenario{Decision: failedReplay.Decision, Launches: failureLaunches - before}

	timeoutRoot, timeoutCleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return nil, err
	}
	defer timeoutCleanup()
	timeoutRegistry := proofGovernorSingleSelftestRegistry("timeout", "cli", "plan-run")
	if err := writeProofGovernorSelftestRegistry(timeoutRoot, timeoutRegistry); err != nil {
		return nil, err
	}
	timeoutLaunches := 0
	timeoutRunner := func(root string, check Check) CheckResult {
		timeoutLaunches++
		return CheckResult{ExitCode: 124, Stderr: "process tree killed after timeout"}
	}
	timeoutResult := executeProofGovernorPlan(timeoutRoot, timeoutRegistry, []string{"timeout"}, filepath.Join(timeoutRoot, "receipts"), timeoutRunner, now)
	if timeoutResult.OK || timeoutResult.ExitCode != 124 {
		return nil, fmt.Errorf("timeout scenario reported global pass: %#v", timeoutResult)
	}
	observed["partial-timeout-receipt"] = proofGovernorObservedScenario{Decision: timeoutResult.Decision, Launches: timeoutLaunches}

	namespaceRoot, namespaceCleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return nil, err
	}
	defer namespaceCleanup()
	workerRegistry := proofGovernorSingleSelftestRegistry("namespaced", "cli", "worker-run")
	if err := writeProofGovernorSelftestRegistry(namespaceRoot, workerRegistry); err != nil {
		return nil, err
	}
	namespaceReceipts := filepath.Join(namespaceRoot, "receipts")
	namespaceLaunches := 0
	namespaceRunner := func(root string, check Check) CheckResult {
		namespaceLaunches++
		return CheckResult{ExitCode: 0}
	}
	if result := executeProofGovernorPlan(namespaceRoot, workerRegistry, []string{"namespaced"}, namespaceReceipts, namespaceRunner, now); !result.OK {
		return nil, fmt.Errorf("namespace setup failed: %#v", result)
	}
	integratedRegistry := proofGovernorSingleSelftestRegistry("namespaced", "cli", "plan-run")
	if err := writeProofGovernorSelftestRegistry(namespaceRoot, integratedRegistry); err != nil {
		return nil, err
	}
	before = namespaceLaunches
	namespace := executeProofGovernorPlan(namespaceRoot, integratedRegistry, []string{"namespaced"}, namespaceReceipts, namespaceRunner, now.Add(time.Second))
	observed["preintegration-run-namespace"] = proofGovernorObservedScenario{Decision: namespace.Decision, Launches: namespaceLaunches - before}

	registryRoot, registryCleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return nil, err
	}
	defer registryCleanup()
	registryDrift := proofGovernorSingleSelftestRegistry("registry", "cli", "plan-run")
	if err := writeProofGovernorSelftestRegistry(registryRoot, registryDrift); err != nil {
		return nil, err
	}
	registryReceipts := filepath.Join(registryRoot, "receipts")
	registryLaunches := 0
	registryRunner := func(root string, check Check) CheckResult {
		registryLaunches++
		return CheckResult{ExitCode: 0}
	}
	if result := executeProofGovernorPlan(registryRoot, registryDrift, []string{"registry"}, registryReceipts, registryRunner, now); !result.OK {
		return nil, fmt.Errorf("registry setup failed: %#v", result)
	}
	registryContent, err := os.ReadFile(filepath.Join(registryRoot, "registry.json"))
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(registryRoot, "registry.json"), append(registryContent, '\n'), 0o644); err != nil {
		return nil, err
	}
	before = registryLaunches
	registryResult := executeProofGovernorPlan(registryRoot, registryDrift, []string{"registry"}, registryReceipts, registryRunner, now.Add(time.Second))
	observed["registry-drift"] = proofGovernorObservedScenario{Decision: registryResult.Decision, Launches: registryLaunches - before}

	unknown := executeProofGovernorPlan(registryRoot, registryDrift, []string{"unknown"}, filepath.Join(registryRoot, "unknown-receipts"), registryRunner, now)
	observed["unknown-impact"] = proofGovernorObservedScenario{Decision: unknown.Decision, Launches: 0}

	guiRoot, guiCleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return nil, err
	}
	defer guiCleanup()
	guiRegistry := proofGovernorSingleSelftestRegistry("desktop", "native-gui", "plan-run")
	if err := writeProofGovernorSelftestRegistry(guiRoot, guiRegistry); err != nil {
		return nil, err
	}
	guiLaunches := 0
	guiRunner := func(root string, check Check) CheckResult {
		guiLaunches++
		return CheckResult{ExitCode: 0}
	}
	gui := executeProofGovernorPlan(guiRoot, guiRegistry, []string{"desktop"}, filepath.Join(guiRoot, "receipts"), guiRunner, now)
	observed["native-gui-unattended"] = proofGovernorObservedScenario{Decision: gui.Decision, Launches: guiLaunches}

	snapshotScript, err := os.ReadFile(filepath.Join(repoRoot, ".github", "skills", "kb-regression-snapshot", "scripts", "kb-regression-snapshot.ps1"))
	if err != nil {
		return nil, err
	}
	scriptText := string(snapshotScript)
	if strings.Count(scriptText, "chromium.launch({ headless: true })") != 1 || !strings.Contains(scriptText, "Assert-DomElements") {
		return nil, fmt.Errorf("headless browser assertions are not batched through one launch site")
	}
	observed["headless-browser-batch"] = proofGovernorObservedScenario{Decision: proofGovernorRun, Launches: 1}

	if err := exerciseProofGovernorPublicCLI(); err != nil {
		return nil, err
	}
	return observed, nil
}

func exerciseProofGovernorPublicCLI() error {
	root, cleanup, err := newProofGovernorSelftestRoot()
	if err != nil {
		return err
	}
	defer cleanup()
	registry := proofGovernorRegistry{SchemaVersion: proofGovernorSchemaVersion, Checks: []proofGovernorCheckSpec{{
		ID: "public", Namespace: proofGovernorNamespace{Goal: "selftest", Slice: "integrated", Run: "plan-run"},
		Command: []string{"go", "version"}, Covers: []string{"public"}, Inputs: []string{"input.txt"},
		WorkingDir: ".", TimeoutMS: 10_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 300,
	}}}
	if err := writeProofGovernorSelftestRegistry(root, registry); err != nil {
		return err
	}
	args := []string{
		"proof-run", "--root", root, "--registry", "registry.json",
		"--receipt-dir", "receipts", "--request", "public",
	}
	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != 0 {
		return fmt.Errorf("public proof-run failed: code=%d stderr=%s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(args, &stdout, &stderr); code != 0 || !strings.Contains(stdout.String(), "proof-run: REUSE") {
		return fmt.Errorf("public proof-run did not reuse unchanged proof: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	return nil
}

func newProofGovernorSelftestRoot() (string, func(), error) {
	root, err := os.MkdirTemp("", "kb-proof-governor-selftest-*")
	if err != nil {
		return "", func() {}, err
	}
	for name, content := range map[string]string{"input.txt": "stable\n", "unrelated.txt": "stable\n"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			_ = os.RemoveAll(root)
			return "", func() {}, err
		}
	}
	return root, func() { _ = os.RemoveAll(root) }, nil
}

func proofGovernorSelftestRegistry(run string, inputs []string) proofGovernorRegistry {
	namespace := proofGovernorNamespace{Goal: "selftest", Slice: "integrated", Run: run}
	return proofGovernorRegistry{SchemaVersion: proofGovernorSchemaVersion, Checks: []proofGovernorCheckSpec{
		{
			ID: "full", Namespace: namespace, Command: []string{"fake", "full"},
			Covers: []string{"full", "focused"}, Inputs: inputs, WorkingDir: ".",
			TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 300,
		},
		{
			ID: "focused", Namespace: namespace, Command: []string{"fake", "focused"},
			Covers: []string{"focused"}, Inputs: inputs, WorkingDir: ".",
			TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: "cli", MaxAgeSeconds: 300,
		},
	}}
}

func proofGovernorSingleSelftestRegistry(id, class, run string) proofGovernorRegistry {
	return proofGovernorRegistry{SchemaVersion: proofGovernorSchemaVersion, Checks: []proofGovernorCheckSpec{{
		ID: id, Namespace: proofGovernorNamespace{Goal: "selftest", Slice: "integrated", Run: run},
		Command: []string{"fake", id}, Covers: []string{id}, Inputs: []string{"input.txt"},
		WorkingDir: ".", TimeoutMS: 1_000, ExpectedExit: 0, ExecutionClass: class, MaxAgeSeconds: 300,
	}}}
}

func writeProofGovernorSelftestRegistry(root string, registry proofGovernorRegistry) error {
	content, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "registry.json"), append(content, '\n'), 0o644)
}
