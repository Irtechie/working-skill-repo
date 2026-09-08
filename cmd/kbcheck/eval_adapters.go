package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
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
		output.OK = output.OK && run.Status == "pass" && run.ExitCode == 0
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
	var result map[string]any
	process := CheckResult{}
	exitCode := 0
	status := "pass"
	if mode == "live" {
		workspace, prepareErr := prepareRoutingWorkspace(root, runDir)
		if prepareErr != nil {
			return adapterRun{}, prepareErr
		}
		live, captured, err := invokeLiveAgent(workspace, runtime, fixture, runID, opts)
		process = captured
		exitCode = process.ExitCode
		if err != nil {
			status = "fail"
			if exitCode == 0 {
				exitCode = 1
			}
			result = map[string]any{"id": runID, "fixture_id": fixtureID, "eval_run_id": runID, "runtime": runtime, "evidence_kind": "live", "adapter_error": err.Error()}
		} else {
			result = live
		}
	} else {
		result = dryRunResult(fixture, runtime, runID)
	}
	for path, content := range map[string]string{stdoutPath: process.Stdout, stderrPath: process.Stderr} {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return adapterRun{}, err
		}
	}
	if err := writeAdapterJSON(resultPath, result); err != nil {
		return adapterRun{}, err
	}
	if err := writeAdapterJSON(manifestPath, newRunManifest(root, runID, runtime, mode, fixture)); err != nil {
		return adapterRun{}, err
	}
	score, scoreErr := computeSkillEval(root, "", resultPath, "", false, runID, manifestPath)
	scoreBytes, _ := json.MarshalIndent(score, "", "  ")
	if err := os.WriteFile(filepath.Join(runDir, "score.json"), scoreBytes, 0o644); err != nil {
		return adapterRun{}, err
	}
	if scoreErr != nil || !score.OK {
		status = "fail"
		if exitCode == 0 {
			exitCode = 1
		}
	}
	return adapterRun{FixtureID: fixtureID, RunID: runID, RunDir: runDir, ResultPath: resultPath, ManifestPath: manifestPath, Mode: mode, Status: status, ExitCode: exitCode}, nil
}

// A prompt projection alone is insufficient: a live agent can read fixtures
// from its working directory. Expose skill payloads, never the source corpus.
func prepareRoutingWorkspace(root, runDir string) (string, error) {
	workspace := filepath.Join(runDir, "workspace")
	if err := os.MkdirAll(workspace, 0o755); err != nil {
		return "", err
	}
	source := filepath.Join(root, ".github", "skills")
	if info, err := os.Stat(source); err == nil && info.IsDir() {
		err = filepath.WalkDir(source, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("skill payload symlink refused: %s", path)
			}
			rel, err := filepath.Rel(source, path)
			if err != nil {
				return err
			}
			target := filepath.Join(workspace, ".github", "skills", rel)
			if entry.IsDir() {
				return os.MkdirAll(target, 0o755)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0o644)
		})
		if err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(filepath.Join(workspace, "AGENTS.md"), []byte("This is a read-only KB routing evaluation. Consult .github/skills or installed KB skills. Do not execute the requested work, bootstrap memory, or inspect parent directories, evaluator source, fixtures, expected answers, or previous eval results.\n"), 0o644); err != nil {
		return "", err
	}
	// A nested repository prevents parent-project instruction discovery.
	if out, err := exec.Command("git", "-C", workspace, "init", "--quiet").CombinedOutput(); err != nil {
		return "", fmt.Errorf("initialize routing workspace: %w: %s", err, out)
	}
	return workspace, nil
}

func writeAdapterJSON(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func dryRunResult(fixture map[string]any, runtime, runID string) map[string]any {
	expected, _ := fixture["expected"].(map[string]any)
	fixtureID := stringValue(fixture["id"])
	return map[string]any{
		"id":            runID,
		"fixture_id":    fixtureID,
		"eval_run_id":   runID,
		"evidence_kind": "synthetic",
		"runtime":       runtime,
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

func invokeLiveAgent(root, runtime string, fixture map[string]any, runID string, opts options) (map[string]any, CheckResult, error) {
	command := opts.agentCommand
	if command == "" {
		command = runtime
		if runtime == "ghcp" {
			command = "copilot"
		}
	}
	if _, err := exec.LookPath(command); err != nil {
		return nil, CheckResult{ExitCode: 127}, fmt.Errorf("%s command unavailable; use --dry-run or install/authenticate CLI", command)
	}
	prompt := evalPrompt(fixture, runtime, runID)
	if runtime == "opencode" {
		result, process, err := invokeOpenCode(root, command, prompt, 5*time.Minute)
		if err == nil {
			err = validateAdapterIdentity(result, fixture, runID)
		}
		return result, process, err
	}
	args, err := routingAgentArgs(command, runtime, prompt)
	if err != nil {
		return nil, CheckResult{ExitCode: 127}, err
	}
	process := runProcessCheck(root, Check{Args: args, Timeout: 5 * time.Minute})
	if process.ExitCode != 0 {
		return nil, process, fmt.Errorf("%s failed: exit=%d", runtime, process.ExitCode)
	}
	var result map[string]any
	if runtime == "ghcp" {
		result, err = parseCopilotEventStream(process.Stdout)
	} else {
		err = json.Unmarshal([]byte(extractLastJSONObject(process.Stdout)), &result)
	}
	if err != nil {
		return nil, process, err
	}
	return result, process, validateAdapterIdentity(result, fixture, runID)
}

// Copilot's text renderer wraps JSON strings at the terminal width. Consume
// structured events instead and require the CLI's successful terminal result.
func parseCopilotEventStream(stream string) (map[string]any, error) {
	var content string
	finished := false
	for _, line := range strings.Split(stream, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event map[string]any
		if json.Unmarshal([]byte(line), &event) != nil || event == nil || finished {
			return nil, fmt.Errorf("malformed or trailing Copilot event")
		}
		data, _ := event["data"].(map[string]any)
		switch stringValue(event["type"]) {
		case "session.error":
			return nil, fmt.Errorf("Copilot session error")
		case "assistant.message":
			content = stringValue(data["content"])
		case "result":
			exit, ok := event["exitCode"].(float64)
			if !ok || exit != 0 || stringValue(event["sessionId"]) == "" {
				return nil, fmt.Errorf("Copilot unsuccessful terminal result")
			}
			finished = true
		case "":
			return nil, fmt.Errorf("Copilot event missing type")
		}
	}
	var result map[string]any
	if !finished || json.Unmarshal([]byte(content), &result) != nil || result == nil {
		return nil, fmt.Errorf("Copilot stream missing successful JSON response")
	}
	return result, nil
}

// Never send model prompts through Windows command-shell shims. Resolve only
// known package entrypoints, then pass the prompt as one native argv element.
func routingAgentArgs(command, runtime, prompt string) ([]string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		return nil, err
	}
	args := []string{path}
	if goruntime.GOOS == "windows" && strings.ToLower(filepath.Ext(path)) != ".exe" {
		base := filepath.Join(filepath.Dir(path), "node_modules")
		switch runtime {
		case "codex":
			path = filepath.Join(base, "@openai", "codex", "vendor", "x86_64-pc-windows-msvc", "bin", "codex.exe")
			args = []string{path}
		case "ghcp":
			path = filepath.Join(base, "@github", "copilot", "npm-loader.js")
			node, nodeErr := exec.LookPath("node.exe")
			if nodeErr != nil {
				return nil, nodeErr
			}
			args = []string{node, path}
		default:
			return nil, fmt.Errorf("unsupported routing runtime: %s", runtime)
		}
		if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
			return nil, fmt.Errorf("native %s entrypoint unavailable: %s", runtime, path)
		}
	}
	switch runtime {
	case "codex":
		return append(args, "exec", "--sandbox", "read-only", "--color", "never", prompt), nil
	case "ghcp":
		return append(args, "--prompt", prompt, "--silent", "--output-format", "json", "--stream", "off", "--no-ask-user", "--available-tools=view,grep,glob", "--deny-tool=shell", "--deny-tool=write"), nil
	default:
		return nil, fmt.Errorf("unsupported routing runtime: %s", runtime)
	}
}

func validateAdapterIdentity(result, fixture map[string]any, runID string) error {
	if stringValue(result["fixture_id"]) != stringValue(fixture["id"]) || stringValue(result["eval_run_id"]) != runID {
		return fmt.Errorf("agent result fixture_id/eval_run_id does not match this invocation")
	}
	result["evidence_kind"] = "live"
	return nil
}

// Resolve only a native executable. Windows npm shims are never evaluated with
// the prompt as shell text; the observed opencode-ai package bundles this binary.
func resolveOpenCodeExecutable(command string) (string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		return "", err
	}
	if goruntime.GOOS == "windows" {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".cmd", ".ps1", ".bat":
			path = filepath.Join(filepath.Dir(path), "node_modules", "opencode-ai", "bin", "opencode.exe")
		case ".exe":
		default:
			return "", fmt.Errorf("unsupported OpenCode launcher: native executable required")
		}
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return "", fmt.Errorf("OpenCode native executable unavailable: %s", path)
	}
	return path, nil
}

func invokeOpenCode(root, command, prompt string, timeout time.Duration) (map[string]any, CheckResult, error) {
	path, err := resolveOpenCodeExecutable(command)
	if err != nil {
		return nil, CheckResult{ExitCode: 127}, err
	}
	process := runProcessCheck(root, Check{Args: []string{path, "run", "--format", "json", prompt}, Timeout: timeout})
	if process.ExitCode != 0 {
		return nil, process, fmt.Errorf("opencode process failed with exit code %d", process.ExitCode)
	}
	final, err := parseOpenCodeEventStream(process.Stdout)
	return final, process, err
}

func parseOpenCodeEventStream(stream string) (map[string]any, error) {
	var text strings.Builder
	session := ""
	finished := false
	for _, line := range strings.Split(stream, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var event map[string]any
		if json.Unmarshal([]byte(line), &event) != nil || event == nil {
			return nil, fmt.Errorf("malformed OpenCode JSON event")
		}
		id := stringValue(event["sessionID"])
		if id == "" || (session != "" && session != id) {
			return nil, fmt.Errorf("missing or inconsistent OpenCode sessionID")
		}
		session = id
		if finished {
			return nil, fmt.Errorf("OpenCode event after final step_finish")
		}
		part, _ := event["part"].(map[string]any)
		switch stringValue(event["type"]) {
		case "error":
			return nil, fmt.Errorf("OpenCode error event")
		case "step_start", "tool_use":
			if part == nil {
				return nil, fmt.Errorf("OpenCode event missing part")
			}
			if stringValue(event["type"]) == "step_start" {
				text.Reset()
			}
		case "text":
			value, ok := part["text"].(string)
			if !ok || stringValue(part["type"]) != "text" {
				return nil, fmt.Errorf("invalid OpenCode text part")
			}
			text.WriteString(value)
		case "step_finish":
			if stringValue(part["type"]) != "step-finish" {
				return nil, fmt.Errorf("invalid OpenCode step_finish part")
			}
			reason := stringValue(part["reason"])
			if reason == "stop" {
				finished = true
			} else if reason != "tool-calls" {
				return nil, fmt.Errorf("OpenCode did not finish normally: %s", reason)
			}
		default:
			return nil, fmt.Errorf("unsupported OpenCode event type")
		}
	}
	if !finished {
		return nil, fmt.Errorf("OpenCode stream missing final step_finish")
	}
	var final map[string]any
	if err := json.Unmarshal([]byte(text.String()), &final); err != nil || final == nil {
		return nil, fmt.Errorf("OpenCode assistant text is not one JSON result")
	}
	return final, nil
}

func evalPrompt(fixture map[string]any, runtime, runID string) string {
	content, _ := json.MarshalIndent(publicEvalPromptFixture(fixture), "", "  ")
	fixtureID := stringValue(fixture["id"])
	return fmt.Sprintf(`You are running a KB skill-routing evaluation for %s.

Rules:
- Do not edit files.
- Do not run destructive commands.
- Do not execute the requested work.
- Decide the smallest correct KB route for the request.
- Return exactly one JSON object and no markdown, prose, or code fences.
- Set eval_run_id exactly to "%s".
- Fill trace.files_read and trace.commands only with files/commands you actually used.
- actual.user_questions is an integer count, not an array.
- actual.artifacts and actual.proof are arrays of proposed deliverables and verification commands for the selected route; they are not claims that you executed work.
- claim_checks is an empty array unless you have a supported deterministic check (file_exists, command_ran, or file_read); do not invent claim types or status fields.
- Do not inspect evaluator source, route fixtures, expected answers, or previous eval results. Use only the supplied public scenario and skill instructions.
- Consult the selected skill and applicable kb-check verification policy before proposing proof.
- Use the public category labels below where applicable; other routes may require additional labels. These are vocabulary choices, not a checklist to include indiscriminately.
  Artifact categories: changed file; verification note; reproduction evidence; root cause note; requirements or assumptions; slice plans; manifest; review findings; PR URL; pushed branch; test output; lint output.
  Proof categories: git diff --check; targeted text/render check if UI-visible; relevant narrow check exits 0; repro fails before fix; same path passes after fix; failing test passes; source/docs read; required checks and reviews pass; remote default contains delivered commit.
- Return claim_checks as [] for this routing-only assessment. The trace records actual reads/commands; proposed artifacts and proof must not be presented as completed work.
- Your final response must be ONLY the JSON object. Do not add a lead-in sentence or code fence. Use this field shape (replace the values): {"id":"...","fixture_id":"...","eval_run_id":"...","actual":{"route":"...","user_questions":0,"artifacts":[],"proof":[]},"trace":{"files_read":[],"commands":[],"tools":[]},"claim_checks":[]}.

Route fixture:
%s

Return a result object with id "%s-live-%s", fixture_id "%s", eval_run_id "%s", actual.route, actual.user_questions, actual.artifacts, actual.proof, trace.files_read, trace.commands, trace.tools, and claim_checks. The scorer, not you, determines pass or fail.
`, runtime, runID, string(content), runtime, fixtureID, fixtureID, runID)
}

// publicEvalPromptFixture is deliberately allowlisted. Fixtures are scorer
// inputs, not prompt templates: expected answers, guards, rubrics, and any
// future fields remain private unless this projection is consciously extended.
func publicEvalPromptFixture(fixture map[string]any) map[string]any {
	prompt := stringValue(fixture["user_prompt"])
	if prompt == "" {
		prompt = stringValue(fixture["prompt"])
	}
	return map[string]any{
		"id":          stringValue(fixture["id"]),
		"user_prompt": prompt,
		"repo_state":  fixture["repo_state"],
	}
}

func newRunManifest(root, runID, runtime, mode string, fixture map[string]any) map[string]any {
	fixtureID := stringValue(fixture["id"])
	protected := []map[string]any{}
	for _, entry := range []struct {
		role string
		path string
	}{
		{"fixture", "evals/route-complexity/" + fixtureID + ".json"},
		{"scorer", "cmd/kbcheck/skill_eval.go"},
		{"result_schema", "evals/skill-eval/result.schema.json"},
		{"adapter", "cmd/kbcheck/eval_adapters.go"},
		{"config", "config/skill-quality.json"},
	} {
		full := resolveRepoPath(root, entry.path)
		protected = append(protected, map[string]any{"role": entry.role, "path": entry.path, "sha256": fileHashOrEmpty(full)})
	}
	publicFixture, _ := json.Marshal(publicEvalPromptFixture(fixture))
	evidenceKind := "live"
	if mode == "dry-run" {
		evidenceKind = "synthetic"
	}
	return map[string]any{"run_id": runID, "fixture_id": fixtureID, "runtime": runtime, "mode": mode, "evidence_kind": evidenceKind, "created_at": time.Now().Format(time.RFC3339Nano), "raw_prompt_sha256": hashBytes(publicFixture), "protected_files": protected}
}

func hashBytes(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func runEvalLiveCorpusCommand(root string, opts options, stdout, stderr io.Writer) int {
	runtimes := opts.runtime
	if runtimes == "" {
		runtimes = "codex,ghcp"
	}
	allRuns := []adapterRun{}
	allOK := true
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
		allOK = allOK && result.OK
	}
	output := adapterOutput{OK: allOK, Runtime: runtimes, Mode: "live", Runs: allRuns}
	if opts.dryRun {
		output.Mode = "dry-run"
	}
	if opts.json {
		writeJSON(stdout, output)
	} else {
		fmt.Fprintf(stdout, "Skill eval live corpus: %d run(s), runtime=%s mode=%s\n", len(output.Runs), runtimes, output.Mode)
	}
	if !output.OK {
		return 1
	}
	return 0
}

func runSkillEvalWrapCommand(root string, opts options, stdout, stderr io.Writer) int {
	before := gitStatusMap(root)
	wrapped := opts.runner
	runtime := "ghcp"
	if strings.Contains(strings.ToLower(wrapped), "codex") {
		runtime = "codex"
	} else if strings.Contains(strings.ToLower(wrapped), "opencode") {
		runtime = "opencode"
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
	if !result.OK {
		fmt.Fprintln(stderr, "Adapter failed; raw run artifacts retained")
		if opts.json {
			writeJSON(stdout, result)
		}
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
		if err := writeAdapterJSON(run.ResultPath, resultJSON); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
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
		return nil, fmt.Errorf("unknown fixture id: %s", fixtureID)
	}
	if fixtureID == "" && !all {
		return nil, fmt.Errorf("pass --fixture-id <id> or --all")
	}
	return fixtures, nil
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
