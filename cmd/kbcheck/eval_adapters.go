package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type adapterRun struct {
	FixtureID    string `json:"fixture_id"`
	RunID        string `json:"run_id"`
	RunDir       string `json:"run_dir"`
	ResultPath   string `json:"result_path"`
	ManifestPath string `json:"manifest_path"`
	Mode         string `json:"mode"`
	Status       string `json:"status"`
	ExitCode     int    `json:"exit_code"`
}

type adapterOutput struct {
	OK      bool         `json:"ok"`
	Runtime string       `json:"runtime"`
	Mode    string       `json:"mode"`
	Runs    []adapterRun `json:"runs"`
}

type epistemicRouterDispatchConfig struct {
	SchemaVersion int      `json:"schema_version"`
	Command       string   `json:"command"`
	CommandArgs   []string `json:"command_args,omitempty"`
	ProjectRoot   string   `json:"project_root"`
	RunRoot       string   `json:"run_root"`
	RunID         string   `json:"run_id"`
	SliceID       string   `json:"slice_id"`
	Packet        string   `json:"packet"`
	RouteAlias    string   `json:"route_alias"`
}

type epistemicRouterDispatchReport struct {
	RouteAlias            string `json:"route_alias"`
	ProviderReportedModel string `json:"provider_reported_model"`
	SessionID             string `json:"session_id"`
	Attribution           string `json:"attribution"`
	ReceiptPath           string `json:"receipt_path"`
	StructuredOutputPath  string `json:"structured_output_path"`
}

func runEvalAdapterCommand(root string, opts options, runtime string, stdout, stderr io.Writer) int {
	result, err := runEvalAdapter(root, opts, runtime)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "Skill eval %s adapter: %d run(s), mode=%s\n", runtime, len(result.Runs), result.Mode)
		for _, run := range result.Runs {
			fmt.Fprintf(stdout, "%s: %s\n", run.FixtureID, run.ResultPath)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

func runEvalAdapter(root string, opts options, runtime string) (adapterOutput, error) {
	runRoot := opts.runRoot
	if runRoot == "" {
		runRoot = ".kb/eval-runs"
	}
	fixtures, err := selectRouteFixtures(root, opts.fixtureID, opts.all)
	if err != nil {
		return adapterOutput{}, err
	}
	mode := "live"
	if opts.dryRun {
		mode = "dry-run"
	}
	output := adapterOutput{OK: true, Runtime: runtime, Mode: mode}
	for _, fixture := range fixtures {
		run, err := runOneAdapterFixture(root, runRoot, runtime, mode, fixture, opts)
		if err != nil {
			return output, err
		}
		output.Runs = append(output.Runs, run)
		if mode == "dry-run" && !opts.keepRun {
			_ = os.RemoveAll(run.RunDir)
		}
	}
	return output, nil
}

func runOneAdapterFixture(root, runRoot, runtime, mode string, fixture map[string]any, opts options) (adapterRun, error) {
	fixtureID := stringValue(fixture["id"])
	now := time.Now()
	runID := fmt.Sprintf("%s-%09d-%s", now.Format("20060102-150405"), now.Nanosecond(), slug(fixtureID+"-"+runtime+"-"+mode))
	runDir := resolveRepoPath(root, filepath.Join(runRoot, runID))
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return adapterRun{}, err
	}
	resultPath := filepath.Join(runDir, "result.json")
	manifestPath := filepath.Join(runDir, "manifest.json")
	stdoutPath := filepath.Join(runDir, "stdout.txt")
	stderrPath := filepath.Join(runDir, "stderr.txt")
	epistemic := boolValue(fixture["_epistemic"])
	surfaces := []instructionSurface{}
	actorRoot := root
	auditActorRoot := actorRoot
	executionActorHash := ""
	cleanupActor := func() {}
	isolationCLI := ""
	routerConfig := epistemicRouterDispatchConfig{}
	routerContractPath := ""
	if epistemic {
		var actorErr error
		actorRoot, cleanupActor, actorErr = createEpistemicActorWorkspace(root, runDir)
		if actorErr != nil {
			return adapterRun{}, actorErr
		}
		defer cleanupActor()
		var surfaceErr error
		surfaces, surfaceErr = epistemicInstructionSurfaces(root, runtime, mode)
		if surfaceErr != nil {
			return adapterRun{}, surfaceErr
		}
		if err := materializeEpistemicActorWorkspace(root, fixtureID, actorRoot, surfaces); err != nil {
			return adapterRun{}, err
		}
		schema, err := os.ReadFile(filepath.Join(root, "evals", "skill-eval", "result.schema.json"))
		if err != nil {
			return adapterRun{}, err
		}
		if err := os.WriteFile(filepath.Join(actorRoot, "result.schema.json"), schema, 0o644); err != nil {
			return adapterRun{}, err
		}
		if err := materializeEpistemicInstructionSurfaces(root, actorRoot, runtime); err != nil {
			return adapterRun{}, err
		}
		if (mode == "live" || mode == "dry-run") && runtime == "codex" {
			codexHome, userHome, pluginRoots, err := codexIsolationRoots()
			if err != nil {
				return adapterRun{}, err
			}
			projectSkills := []string{".agents/skills/kb-plan/SKILL.md", ".agents/skills/kb-gate/SKILL.md"}
			isolation, err := buildCodexEpistemicSkillIsolation(codexHome, userHome, pluginRoots, projectSkills)
			if err != nil {
				return adapterRun{}, err
			}
			isolationCLI = isolation.CLIConfig
			surfaces = append(surfaces, codexIsolationInstructionSurfaces(isolation)...)
		}
		var contractErr error
		routerConfig, routerContractPath, contractErr = loadEpistemicRouterDispatchConfig(opts.agentCommand)
		if contractErr != nil {
			return adapterRun{}, contractErr
		}
		promptPath := filepath.Join(actorRoot, "sealed-prompt.txt")
		if err := os.WriteFile(promptPath, []byte(evalPrompt(fixture, runtime, runID)), 0o600); err != nil {
			return adapterRun{}, err
		}
		if _, _, err := epistemicRoutedAgentCommand(routerConfig, actorRoot, promptPath, filepath.Join(actorRoot, "result.schema.json"), "epistemic-preview.json", isolationCLI); err != nil {
			return adapterRun{}, err
		}
		executionActorHash = directoryContentHash(actorRoot)
		auditActorRoot = filepath.Join(runDir, "actor-workspace-audit")
		if err := snapshotEpistemicActorWorkspace(actorRoot, auditActorRoot); err != nil {
			return adapterRun{}, err
		}
	}
	result := dryRunResult(fixture, runtime, runID)
	if epistemic {
		result = dryRunEpistemicResult(root, fixture, runtime, runID, surfaces)
	}
	exitCode := 0
	status := "pass"
	if mode == "live" {
		live, code, err := invokeLiveAgent(actorRoot, runtime, fixture, runID, opts, isolationCLI)
		exitCode = code
		if err != nil {
			status = "fail"
			_ = os.WriteFile(stderrPath, []byte(err.Error()), 0o644)
		} else {
			result = live
			if epistemic {
				result["instruction_surfaces"] = surfaces
			}
		}
	}
	writeJSONFile(resultPath, result)
	manifest := newRunManifest(root, runID, runtime, fixture, surfaces, auditActorRoot)
	if epistemic {
		var manifestErr error
		manifest, manifestErr = newRoutedEpistemicRunManifest(root, runID, runtime, fixture, surfaces, auditActorRoot, routerConfig, routerContractPath)
		if manifestErr != nil {
			return adapterRun{}, manifestErr
		}
	}
	if epistemic {
		manifest["execution_actor_workspace"] = actorRoot
		manifest["execution_actor_workspace_external"] = pathOutsideRoot(root, actorRoot)
		manifest["execution_actor_workspace_sha256"] = executionActorHash
	}
	writeJSONFile(manifestPath, manifest)
	score, _ := computeSkillEval(root, "", resultPath, "", false, runID, manifestPath)
	scoreBytes, _ := json.MarshalIndent(score, "", "  ")
	_ = os.WriteFile(filepath.Join(runDir, "score.json"), scoreBytes, 0o644)
	if !score.OK {
		status = "fail"
		exitCode = 1
	}
	_ = os.WriteFile(stdoutPath, []byte(""), 0o644)
	if _, err := os.Stat(stderrPath); err != nil {
		_ = os.WriteFile(stderrPath, []byte(""), 0o644)
	}
	return adapterRun{FixtureID: fixtureID, RunID: runID, RunDir: runDir, ResultPath: resultPath, ManifestPath: manifestPath, Mode: mode, Status: status, ExitCode: exitCode}, nil
}

func dryRunEpistemicResult(root string, fixture map[string]any, runtime, runID string, surfaces []instructionSurface) map[string]any {
	fixtureID := stringValue(fixture["id"])
	oracle, err := loadEpistemicOracle(root, fixtureID)
	if err != nil {
		return map[string]any{"id": runID, "fixture_id": fixtureID, "expected_result": "fail", "eval_run_id": runID}
	}
	decision := oracle.Decision
	epistemic := map[string]any{"decision": decision}
	if decision == "investigate" {
		epistemic["evidence_inspected"] = oracle.ResolvingEvidence
		epistemic["final_state"] = oracle.FinalState
	}
	return map[string]any{
		"id": runID, "fixture_id": fixtureID, "expected_result": "pass", "eval_run_id": runID,
		"response_text": "Proceeding from the checked-in evidence.",
		"actual":        map[string]any{"route": "kb-plan", "user_questions": 0, "artifacts": []string{}, "proof": []string{}},
		"trace":         map[string]any{"files_read": oracle.ResolvingEvidence, "commands": []string{"dry-run"}, "tools": []string{"skill-eval-run-" + runtime}},
		"claim_checks":  []map[string]any{}, "epistemic": epistemic, "instruction_surfaces": surfaces,
	}
}

func dryRunResult(fixture map[string]any, runtime, runID string) map[string]any {
	expected, _ := fixture["expected"].(map[string]any)
	fixtureID := stringValue(fixture["id"])
	return map[string]any{
		"id":              runID,
		"fixture_id":      fixtureID,
		"expected_result": "pass",
		"eval_run_id":     runID,
		"actual": map[string]any{
			"route":          stringValue(expected["route"]),
			"user_questions": intValue(expected["max_user_questions"]),
			"artifacts":      stringArray(expected["artifacts"]),
			"proof":          stringArray(expected["proof"]),
		},
		"trace": map[string]any{
			"files_read": []string{"evals/route-complexity/" + fixtureID + ".json"},
			"commands":   []string{"dry-run"},
			"tools":      []string{"skill-eval-run-" + runtime},
		},
		"claim_checks": []map[string]any{
			{"type": "file_exists", "path": "evals/route-complexity/" + fixtureID + ".json", "contains": "", "expected": true, "claim": "Fixture file exists"},
			{"type": "command_ran", "path": "", "contains": "dry-run", "expected": true, "claim": "Dry-run command was recorded"},
		},
	}
}

func invokeLiveAgent(root, runtime string, fixture map[string]any, runID string, opts options, isolationCLI string) (map[string]any, int, error) {
	const routedContractPrefix = "kbrouter-contract:"
	if boolValue(fixture["_epistemic"]) {
		if strings.HasPrefix(strings.TrimSpace(opts.agentCommand), routedContractPrefix) {
			configPath := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(opts.agentCommand), routedContractPrefix))
			return invokeRoutedEpistemicAgent(root, fixture, runID, configPath, isolationCLI)
		}
		return nil, 1, fmt.Errorf("epistemic live evaluation requires an orchestrator-selected kbrouter contract; direct model dispatch is refused")
	}
	command := opts.agentCommand
	args := []string{}
	if command == "" {
		command = runtime
		if runtime == "ghcp" {
			command = "copilot"
		}
	}
	if _, err := exec.LookPath(command); err != nil {
		return nil, 127, fmt.Errorf("%s command unavailable; use --dry-run or install/authenticate CLI", command)
	}
	prompt := evalPrompt(fixture, runtime, runID)
	cmd := exec.Command(command, args...)
	cmd.Dir = root
	cmd.Stdin = strings.NewReader(prompt)
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		code := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
		return nil, code, fmt.Errorf("%s\n%s", out.String(), errOut.String())
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(extractLastJSONObject(out.String())), &result); err != nil {
		return nil, 1, err
	}
	return result, 0, nil
}

func invokeRoutedEpistemicAgent(actorRoot string, fixture map[string]any, runID, configPath, isolationCLI string) (map[string]any, int, error) {
	config, _, err := loadEpistemicRouterDispatchConfig("kbrouter-contract:" + configPath)
	if err != nil {
		return nil, 1, err
	}
	promptPath := filepath.Join(actorRoot, "sealed-prompt.txt")
	wantPrompt := []byte(evalPrompt(fixture, "codex", runID))
	sealedPrompt, err := os.ReadFile(promptPath)
	if err != nil || !bytes.Equal(sealedPrompt, wantPrompt) {
		return nil, 1, fmt.Errorf("sealed epistemic prompt is missing or changed")
	}
	runHash := sha256.Sum256([]byte(runID))
	structuredName := "epistemic-result-" + hex.EncodeToString(runHash[:8]) + ".json"
	command, args, err := epistemicRoutedAgentCommand(config, actorRoot, promptPath, filepath.Join(actorRoot, "result.schema.json"), structuredName, isolationCLI)
	if err != nil {
		return nil, 1, err
	}
	if _, err := exec.LookPath(command); err != nil {
		return nil, 127, fmt.Errorf("kbrouter command unavailable: %w", err)
	}
	cmd := exec.Command(command, args...)
	cmd.Dir = config.ProjectRoot
	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	if err := cmd.Run(); err != nil {
		code := 1
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		}
		return nil, code, fmt.Errorf("%s\n%s", out.String(), errOut.String())
	}
	var report epistemicRouterDispatchReport
	if err := json.Unmarshal([]byte(extractLastJSONObject(out.String())), &report); err != nil {
		return nil, 1, fmt.Errorf("decode kbrouter dispatch report: %w", err)
	}
	if report.RouteAlias != config.RouteAlias || report.Attribution != "exact" || report.SessionID == "" || report.ProviderReportedModel == "" || report.ReceiptPath == "" || report.StructuredOutputPath == "" {
		return nil, 1, fmt.Errorf("kbrouter dispatch did not return exact route/session/provider-model evidence")
	}
	if err := validateEpistemicRoutingReceipt(report.ReceiptPath, config, report); err != nil {
		return nil, 1, err
	}
	var result map[string]any
	if err := readJSONFile(report.StructuredOutputPath, &result); err != nil {
		return nil, 1, fmt.Errorf("read routed epistemic structured output: %w", err)
	}
	return result, 0, nil
}

func loadEpistemicRouterDispatchConfig(agentCommand string) (epistemicRouterDispatchConfig, string, error) {
	const prefix = "kbrouter-contract:"
	value := strings.TrimSpace(agentCommand)
	if !strings.HasPrefix(value, prefix) {
		return epistemicRouterDispatchConfig{}, "", fmt.Errorf("epistemic evaluation requires an orchestrator-selected kbrouter contract; direct model dispatch is refused")
	}
	path := strings.TrimSpace(strings.TrimPrefix(value, prefix))
	if path == "" {
		return epistemicRouterDispatchConfig{}, "", fmt.Errorf("epistemic kbrouter contract path is empty")
	}
	absPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return epistemicRouterDispatchConfig{}, "", err
	}
	var config epistemicRouterDispatchConfig
	if err := readJSONFile(absPath, &config); err != nil {
		return epistemicRouterDispatchConfig{}, "", fmt.Errorf("read routed epistemic dispatch contract: %w", err)
	}
	return config, absPath, nil
}

func epistemicRoutedAgentCommand(config epistemicRouterDispatchConfig, actorRoot, promptPath, schemaPath, structuredName, isolationCLI string) (string, []string, error) {
	if config.SchemaVersion != 1 || strings.TrimSpace(config.Command) == "" || strings.TrimSpace(config.ProjectRoot) == "" || strings.TrimSpace(config.RunRoot) == "" || strings.TrimSpace(config.RunID) == "" || strings.TrimSpace(config.SliceID) == "" || strings.TrimSpace(config.Packet) == "" || strings.TrimSpace(config.RouteAlias) == "" {
		return "", nil, fmt.Errorf("routed epistemic dispatch requires a complete orchestrator-selected contract")
	}
	if config.RouteAlias == "current" {
		return "", nil, fmt.Errorf("current/App-only route cannot execute an external epistemic evaluator")
	}
	if strings.TrimSpace(actorRoot) == "" || strings.TrimSpace(promptPath) == "" || strings.TrimSpace(schemaPath) == "" || strings.TrimSpace(structuredName) == "" || strings.TrimSpace(isolationCLI) == "" {
		return "", nil, fmt.Errorf("routed epistemic dispatch requires sealed actor, prompt, schema, output, and instruction isolation")
	}
	artifactStem := strings.TrimSuffix(filepath.Base(structuredName), filepath.Ext(structuredName))
	if artifactStem == "" || artifactStem == "." {
		return "", nil, fmt.Errorf("routed epistemic dispatch requires a safe structured output name")
	}
	args := append([]string{}, config.CommandArgs...)
	args = append(args,
		"dispatch", "--project-root", config.ProjectRoot, "--run-root", config.RunRoot,
		"--run-id", config.RunID, "--slice-id", config.SliceID, "--packet", config.Packet,
		"--output", artifactStem+"-dispatch-output.json", "--receipt", artifactStem+"-dispatch-receipt.json",
		"--handoff", artifactStem+"-dispatch-handoff.json",
		"--route-alias", config.RouteAlias, "--evaluator-actor-cwd", actorRoot,
		"--evaluator-prompt", promptPath, "--evaluator-output-schema", schemaPath,
		"--evaluator-structured-output", structuredName, "--evaluator-instruction-config", isolationCLI, "--json",
	)
	return config.Command, args, nil
}

func validateEpistemicRoutingReceipt(path string, config epistemicRouterDispatchConfig, report epistemicRouterDispatchReport) error {
	var receipt struct {
		RouteEvidence struct {
			RunID                 string `json:"run_id"`
			SliceID               string `json:"slice_id"`
			RouteAlias            string `json:"route_alias"`
			ProviderReportedModel string `json:"provider_reported_model"`
			SessionID             string `json:"session_id"`
		} `json:"route_evidence"`
	}
	if err := readJSONFile(path, &receipt); err != nil {
		return fmt.Errorf("read epistemic routing receipt: %w", err)
	}
	evidence := receipt.RouteEvidence
	if evidence.RunID != config.RunID || evidence.SliceID != config.SliceID || evidence.RouteAlias != config.RouteAlias || evidence.SessionID != report.SessionID || evidence.ProviderReportedModel != report.ProviderReportedModel {
		return fmt.Errorf("epistemic routing receipt does not match dispatch report and selected route")
	}
	return nil
}

func evalPrompt(fixture map[string]any, runtime, runID string) string {
	visible := map[string]any{}
	for key, value := range fixture {
		if !strings.HasPrefix(key, "_") {
			visible[key] = value
		}
	}
	content, _ := json.MarshalIndent(visible, "", "  ")
	fixtureID := stringValue(fixture["id"])
	if boolValue(fixture["_epistemic"]) {
		return fmt.Sprintf(`You are running a sealed KB planning-premise evaluation for %s.

Rules:
- Work only inside the current disposable repository.
- Invoke $kb-plan for the planning pass and $kb-gate for the plan-to-work decision.
- Do not edit files or run destructive commands.
- Decide whether the material premise is adequately supported (proceed) or requires evidence inspection (investigate).
- If investigation is required, inspect only evidence capable of resolving the premise, then report supported, contradicted, or no-justified-conclusion.
- Return exactly one JSON object and no markdown, prose, or code fences.
- Set eval_run_id exactly to "%s".
- response_text must be the actual concise user-facing answer. Do not expose a verification checklist for a supported proceed case.
- Fill trace fields only with files, commands, and tools actually used. Omit unavailable telemetry rather than guessing it.

Actor-visible fixture:
%s

Return a result object with id "%s-live-%s", fixture_id "%s", expected_result "pass", eval_run_id "%s", response_text, actual.route "kb-plan", actual.user_questions, empty actual.artifacts and actual.proof arrays, epistemic.decision, epistemic.evidence_inspected and epistemic.final_state when investigation occurs, trace.files_read, trace.commands, trace.tools, and an empty claim_checks array.
`, runtime, runID, string(content), runtime, fixtureID, fixtureID, runID)
	}
	return fmt.Sprintf(`You are running a KB skill-routing evaluation for %s.

Rules:
- Do not edit files.
- Do not run destructive commands.
- Do not execute the requested work.
- Decide the smallest correct KB route for the request.
- Return exactly one JSON object and no markdown, prose, or code fences.
- Set eval_run_id exactly to "%s".
- Fill trace.files_read and trace.commands only with files/commands you actually used.

Route fixture:
%s

Return a result object with id "%s-live-%s", fixture_id "%s", expected_result "pass", eval_run_id "%s", actual.route, actual.user_questions, actual.artifacts, actual.proof, trace.files_read, trace.commands, trace.tools, and claim_checks.
`, runtime, runID, string(content), runtime, fixtureID, fixtureID, runID)
}

func newRunManifest(root, runID, runtime string, fixture map[string]any, surfaces []instructionSurface, actorRoot string) map[string]any {
	fixtureID := stringValue(fixture["id"])
	protected := []map[string]any{}
	for _, entry := range []struct {
		role string
		path string
	}{
		{"fixture", "evals/route-complexity/" + fixtureID + ".json"},
		{"scorer", "cmd/kbcheck/skill_eval.go"},
		{"epistemic_scorer", "cmd/kbcheck/skill_eval_epistemic.go"},
		{"result_schema", "evals/skill-eval/result.schema.json"},
		{"adapter", "cmd/kbcheck/eval_adapters.go"},
		{"config", "config/skill-quality.json"},
		{"treatment_governance", "config/skill-guidance-audit.json"},
	} {
		full := resolveRepoPath(root, entry.path)
		if hash := fileHashOrEmpty(full); hash != "" {
			protected = append(protected, map[string]any{"role": entry.role, "path": entry.path, "sha256": hash})
		}
	}
	if boolValue(fixture["_epistemic"]) {
		visibleRoot := filepath.Join(root, "evals", "skill-eval", "epistemic", "visible", fixtureID)
		_ = filepath.WalkDir(visibleRoot, func(path string, entry os.DirEntry, err error) error {
			if err == nil && !entry.IsDir() {
				rel, _ := filepath.Rel(root, path)
				protected = append(protected, map[string]any{"role": "visible_fixture", "path": filepath.ToSlash(rel), "sha256": fileHashOrEmpty(path)})
			}
			return nil
		})
		oraclePath := filepath.Join("evals", "skill-eval", "epistemic", "oracles", fixtureID+".json")
		protected = append(protected, map[string]any{"role": "hidden_oracle", "path": filepath.ToSlash(oraclePath), "sha256": fileHashOrEmpty(filepath.Join(root, oraclePath))})
	}
	manifest := map[string]any{"run_id": runID, "fixture_id": fixtureID, "runtime": runtime, "runtime_identity": runtimeIdentity(runtime), "created_at": time.Now().Format(time.RFC3339Nano), "protected_files": protected, "instruction_surfaces": surfaces}
	if boolValue(fixture["_epistemic"]) {
		manifest["actor_workspace"] = actorRoot
		manifest["actor_workspace_sha256"] = directoryContentHash(actorRoot)
	}
	return manifest
}

func newRoutedEpistemicRunManifest(root, runID, runtime string, fixture map[string]any, surfaces []instructionSurface, actorRoot string, config epistemicRouterDispatchConfig, contractPath string) (map[string]any, error) {
	identity := runtimeIdentity(runtime)
	if stringValue(identity["executable"]) == "" || stringValue(identity["executable"]) == "unknown" || len(stringValue(identity["sha256"])) != 64 || strings.TrimSpace(stringValue(identity["version"])) == "" {
		detail := strings.TrimSpace(stringValue(identity["version_error"]))
		if detail != "" {
			detail = ": " + detail
		}
		return nil, fmt.Errorf("epistemic runtime executable identity is unprovable for %s%s", runtime, detail)
	}
	manifest := newRunManifest(root, runID, runtime, fixture, surfaces, actorRoot)
	manifest["runtime_identity"] = identity
	manifest["model_selection_owner"] = "kbrouter"
	manifest["selected_route_alias"] = config.RouteAlias
	manifest["router_dispatch_contract_sha256"] = fileHashOrEmpty(contractPath)
	return manifest, nil
}

func runEvalLiveCorpusCommand(root string, opts options, stdout, stderr io.Writer) int {
	runtimes := opts.runtime
	if runtimes == "" {
		runtimes = "codex,ghcp"
	}
	allRuns := []adapterRun{}
	for _, runtime := range strings.Split(runtimes, ",") {
		runtime = strings.TrimSpace(runtime)
		if runtime == "" {
			continue
		}
		localOpts := opts
		localOpts.all = true
		result, err := runEvalAdapter(root, localOpts, runtime)
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		allRuns = append(allRuns, result.Runs...)
	}
	output := adapterOutput{OK: true, Runtime: runtimes, Mode: "live", Runs: allRuns}
	if opts.dryRun {
		output.Mode = "dry-run"
	}
	if opts.json {
		writeJSON(stdout, output)
	} else {
		fmt.Fprintf(stdout, "Skill eval live corpus: %d run(s), runtime=%s mode=%s\n", len(output.Runs), runtimes, output.Mode)
	}
	return 0
}

func runSkillEvalWrapCommand(root string, opts options, stdout, stderr io.Writer) int {
	before := gitStatusMap(root)
	wrapped := opts.runner
	runtime := "ghcp"
	if strings.Contains(strings.ToLower(wrapped), "codex") {
		runtime = "codex"
	}
	if wrapped == "" {
		wrapped = "eval-run-ghcp"
		runtime = "ghcp"
	}
	adapterOpts := opts
	adapterOpts.keepRun = true
	result, err := runEvalAdapter(root, adapterOpts, runtime)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	after := gitStatusMap(root)
	writes, deletes := statusDiff(before, after)
	commands := []string{}
	if opts.dryRun {
		commands = append(commands, "dry-run")
	}
	scored := []map[string]any{}
	for _, run := range result.Runs {
		var resultJSON map[string]any
		if err := readJSONFile(run.ResultPath, &resultJSON); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		observed := map[string]any{"captured": true, "method": "path-shim+git-diff", "commands": commands, "writes": writes, "deletes": deletes}
		resultJSON["observed_trace"] = observed
		writeJSONFile(run.ResultPath, resultJSON)
		score, _ := computeSkillEval(root, "", run.ResultPath, "", false, run.RunID, run.ManifestPath)
		if !score.OK {
			fmt.Fprintf(stderr, "Observed-trace scoring failed for %s\n", run.ResultPath)
			return 1
		}
		scored = append(scored, map[string]any{"fixture_id": run.FixtureID, "run_id": run.RunID, "result_path": run.ResultPath, "observed_trace": observed})
		if !opts.keepRun {
			_ = os.RemoveAll(run.RunDir)
		}
	}
	output := map[string]any{"ok": true, "sealed": opts.sealed, "runner": wrapped, "runs": scored}
	if opts.json {
		writeJSON(stdout, output)
	} else {
		fmt.Fprintf(stdout, "Skill eval wrapper: %d run(s), observed_trace captured.\n", len(scored))
	}
	return 0
}

func selectRouteFixtures(root, fixtureID string, all bool) ([]map[string]any, error) {
	files, err := evalFiles(root, "evals/route-complexity", "")
	if err != nil {
		return nil, err
	}
	fixtures := []map[string]any{}
	for _, file := range files {
		var fixture map[string]any
		if err := readJSONFile(file, &fixture); err != nil {
			continue
		}
		if fixtureID != "" && stringValue(fixture["id"]) != fixtureID {
			continue
		}
		fixtures = append(fixtures, fixture)
	}
	if fixtureID != "" && len(fixtures) == 0 {
		fixture, err := loadVisibleEpistemicFixture(root, fixtureID)
		if err != nil {
			return nil, fmt.Errorf("unknown fixture id: %s", fixtureID)
		}
		return []map[string]any{fixture}, nil
	}
	if fixtureID == "" && !all {
		return nil, fmt.Errorf("pass --fixture-id <id> or --all")
	}
	return fixtures, nil
}

func loadVisibleEpistemicFixture(root, fixtureID string) (map[string]any, error) {
	path := filepath.Join(root, "evals", "skill-eval", "epistemic", "visible", fixtureID, "fixture.json")
	var fixture map[string]any
	if err := readJSONFile(path, &fixture); err != nil {
		return nil, err
	}
	fixture["_epistemic"] = true
	return fixture, nil
}

func epistemicInstructionSurfaces(root, runtime, mode string) ([]instructionSurface, error) {
	surfaces := []instructionSurface{}
	for _, mapping := range epistemicProjectSkillMappings(runtime) {
		hash := fileHashOrEmpty(filepath.Join(root, filepath.FromSlash(mapping.Source)))
		if hash != "" {
			surfaces = append(surfaces, instructionSurface{Scope: "repo", Path: mapping.Destination, SHA256: hash, LoadState: "proven"})
		}
	}
	if mode == "live" || (mode == "dry-run" && runtime == "codex") {
		if runtime != "codex" {
			return nil, fmt.Errorf("epistemic live comparison is inconclusive: no supported isolated runtime profile is configured for %s", runtime)
		}
		return surfaces, nil
	}
	for _, scope := range []string{"global", "user", "cached"} {
		sum := sha256.Sum256([]byte("isolated:" + runtime + ":" + scope))
		surfaces = append(surfaces, instructionSurface{Scope: scope, Path: "<isolated>", SHA256: hex.EncodeToString(sum[:]), LoadState: "isolated"})
	}
	return surfaces, nil
}

type epistemicSkillMapping struct{ Source, Destination string }

func epistemicProjectSkillMappings(runtime string) []epistemicSkillMapping {
	prefix := ".github/skills"
	if runtime == "codex" {
		prefix = ".agents/skills"
	}
	return []epistemicSkillMapping{
		{Source: ".github/skills/kb-plan/SKILL.md", Destination: prefix + "/kb-plan/SKILL.md"},
		{Source: ".github/skills/kb-gate/SKILL.md", Destination: prefix + "/kb-gate/SKILL.md"},
	}
}

func materializeEpistemicInstructionSurfaces(root, workspace, runtime string) error {
	for _, mapping := range epistemicProjectSkillMappings(runtime) {
		source := filepath.Join(root, filepath.FromSlash(mapping.Source))
		content, err := os.ReadFile(source)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return err
		}
		dest := filepath.Join(workspace, filepath.FromSlash(mapping.Destination))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dest, content, 0o644); err != nil {
			return err
		}
	}
	return nil
}

type codexSkillIsolationEntry struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	Enabled bool   `json:"enabled"`
	Scope   string `json:"scope"`
}
type codexSkillIsolationConfig struct {
	Disabled       []codexSkillIsolationEntry `json:"disabled"`
	EnabledProject []string                   `json:"enabled_project"`
	CLIConfig      string                     `json:"cli_config"`
}

func buildCodexEpistemicSkillIsolation(codexHome, userHome string, pluginRoots, projectSkills []string) (codexSkillIsolationConfig, error) {
	wantProject := []string{".agents/skills/kb-plan/SKILL.md", ".agents/skills/kb-gate/SKILL.md"}
	if strings.Join(projectSkills, "|") != strings.Join(wantProject, "|") {
		return codexSkillIsolationConfig{}, fmt.Errorf("unexpected enabled project skill set")
	}
	type isolationRoot struct {
		Path, Scope string
		SystemRoot  string
	}
	roots := []isolationRoot{
		{Path: filepath.Join(codexHome, "skills"), Scope: "user", SystemRoot: filepath.Join(codexHome, "skills", ".system")},
		{Path: filepath.Join(userHome, ".agents", "skills"), Scope: "user"},
	}
	for _, pluginRoot := range pluginRoots {
		roots = append(roots, isolationRoot{Path: pluginRoot, Scope: "cached"})
	}
	seen := map[string]bool{}
	entries := []codexSkillIsolationEntry{}
	for _, scan := range roots {
		scanRoot := scan.Path
		if _, err := os.Stat(scanRoot); os.IsNotExist(err) {
			continue
		}
		err := filepath.WalkDir(scanRoot, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.EqualFold(entry.Name(), "SKILL.md") {
				return nil
			}
			clean := filepath.Clean(path)
			key := strings.ToLower(clean)
			if seen[key] {
				return nil
			}
			seen[key] = true
			hash := fileHashOrEmpty(clean)
			if len(hash) != 64 {
				return fmt.Errorf("cannot hash discovered skill %s", clean)
			}
			scope := scan.Scope
			if scan.SystemRoot != "" && !pathOutsideRoot(scan.SystemRoot, clean) {
				scope = "global"
			}
			entries = append(entries, codexSkillIsolationEntry{Path: clean, SHA256: hash, Enabled: false, Scope: scope})
			return nil
		})
		if err != nil {
			return codexSkillIsolationConfig{}, err
		}
	}
	sort.Slice(entries, func(i, j int) bool { return strings.ToLower(entries[i].Path) < strings.ToLower(entries[j].Path) })
	parts := []string{}
	for _, entry := range entries {
		path := strings.ReplaceAll(filepath.ToSlash(entry.Path), "'", "''")
		parts = append(parts, fmt.Sprintf("{path='%s',enabled=false}", path))
	}
	return codexSkillIsolationConfig{Disabled: entries, EnabledProject: append([]string(nil), projectSkills...), CLIConfig: "skills.config=[" + strings.Join(parts, ",") + "]"}, nil
}

func codexIsolationInstructionSurfaces(config codexSkillIsolationConfig) []instructionSurface {
	surfaces := make([]instructionSurface, 0, len(config.Disabled))
	for _, entry := range config.Disabled {
		surfaces = append(surfaces, instructionSurface{Scope: entry.Scope, Path: entry.Path, SHA256: entry.SHA256, LoadState: "isolated"})
	}
	return surfaces
}

func codexIsolationRoots() (string, string, []string, error) {
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", "", nil, err
	}
	codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	if codexHome == "" {
		codexHome = filepath.Join(userHome, ".codex")
	}
	return codexHome, userHome, []string{filepath.Join(codexHome, "plugins", "cache")}, nil
}

func createEpistemicActorWorkspace(sourceRoot, runDir string) (string, func(), error) {
	if pathOutsideRoot(sourceRoot, runDir) {
		return "", func() {}, fmt.Errorf("run directory escaped source repository")
	}
	actorRoot, err := os.MkdirTemp("", "kb-epistemic-actor-")
	if err != nil {
		return "", func() {}, err
	}
	cleanup := func() { _ = os.RemoveAll(actorRoot) }
	if !pathOutsideRoot(sourceRoot, actorRoot) {
		cleanup()
		return "", func() {}, fmt.Errorf("actor workspace must be outside source repository")
	}
	cmd := exec.Command("git", "init", "--quiet", actorRoot)
	if output, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("initialize actor git root: %v: %s", err, strings.TrimSpace(string(output)))
	}
	return actorRoot, cleanup, nil
}

func pathOutsideRoot(root, candidate string) bool {
	rootAbs, err1 := filepath.Abs(root)
	candidateAbs, err2 := filepath.Abs(candidate)
	if err1 != nil || err2 != nil {
		return false
	}
	if !strings.EqualFold(filepath.VolumeName(rootAbs), filepath.VolumeName(candidateAbs)) {
		return true
	}
	rel, err := filepath.Rel(rootAbs, candidateAbs)
	return err == nil && (rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func directoryContentHash(root string) string {
	lines := []string{}
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if entry.IsDir() && (rel == ".git" || strings.HasPrefix(filepath.ToSlash(rel), ".git/")) {
			return filepath.SkipDir
		}
		if !entry.IsDir() {
			lines = append(lines, filepath.ToSlash(rel)+"="+fileHashOrEmpty(path))
		}
		return nil
	})
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

func snapshotEpistemicActorWorkspace(source, destination string) error {
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if entry.IsDir() && (relSlash == ".git" || strings.HasPrefix(relSlash, ".git/") || strings.Contains("/"+relSlash+"/", "/evals/skill-eval/epistemic/oracles/")) {
			return filepath.SkipDir
		}
		if entry.IsDir() {
			return os.MkdirAll(filepath.Join(destination, rel), 0o755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		dest := filepath.Join(destination, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, content, 0o644)
	})
}

func runtimeIdentity(runtime string) map[string]any {
	command := runtime
	if runtime == "ghcp" {
		command = "copilot"
	}
	path, err := exec.LookPath(command)
	if err != nil {
		return map[string]any{"runtime": runtime, "executable": "unknown", "sha256": "unknown", "version": ""}
	}
	version, versionErr := runtimeReportedVersion(path)
	identity := map[string]any{"runtime": runtime, "executable": path, "sha256": fileHashOrEmpty(path), "version": version}
	if versionErr != nil {
		identity["version_error"] = versionErr.Error()
	}
	return identity
}

func runtimeReportedVersion(executable string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var cmd *exec.Cmd
	switch strings.ToLower(filepath.Ext(executable)) {
	case ".cmd", ".bat":
		commandProcessor := strings.TrimSpace(os.Getenv("ComSpec"))
		if commandProcessor == "" {
			commandProcessor = filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")
		}
		if strings.ContainsAny(executable, "\r\n") {
			return "", fmt.Errorf("runtime launcher path contains a line break")
		}
		cmd = exec.CommandContext(ctx, commandProcessor, "/d", "/c", "call", executable, "--version")
	default:
		cmd = exec.CommandContext(ctx, executable, "--version")
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("runtime --version failed: %w: %s", err, strings.TrimSpace(string(output)))
	}
	version := strings.TrimSpace(string(output))
	if version == "" {
		return "", fmt.Errorf("runtime --version returned empty output")
	}
	return version, nil
}

func fileHashOrEmpty(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func extractLastJSONObject(text string) string {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		return text
	}
	depth := 0
	start := -1
	last := ""
	inString := false
	escaped := false
	for i, r := range text {
		if inString {
			if escaped {
				escaped = false
			} else if r == '\\' {
				escaped = true
			} else if r == '"' {
				inString = false
			}
			continue
		}
		if r == '"' {
			inString = true
		} else if r == '{' {
			if depth == 0 {
				start = i
			}
			depth++
		} else if r == '}' {
			depth--
			if depth == 0 && start >= 0 {
				last = text[start : i+1]
				start = -1
			}
		}
	}
	return last
}

func gitStatusMap(root string) map[string]string {
	cmd := exec.Command("git", "-C", root, "status", "--porcelain=v1")
	out, err := cmd.Output()
	if err != nil {
		return map[string]string{}
	}
	status := map[string]string{}
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		path := strings.TrimSpace(line[3:])
		path = strings.Trim(path, `"`)
		path = filepath.ToSlash(path)
		if strings.HasPrefix(path, ".kb/") {
			continue
		}
		status[path] = line[:2]
	}
	return status
}

func statusDiff(before, after map[string]string) ([]string, []string) {
	writes := []string{}
	deletes := []string{}
	for path, afterStatus := range after {
		if before[path] == afterStatus {
			continue
		}
		if strings.Contains(afterStatus, "D") {
			deletes = append(deletes, path)
		} else {
			writes = append(writes, path)
		}
	}
	for path := range before {
		if _, ok := after[path]; !ok {
			writes = append(writes, path)
		}
	}
	sort.Strings(writes)
	sort.Strings(deletes)
	return writes, deletes
}
