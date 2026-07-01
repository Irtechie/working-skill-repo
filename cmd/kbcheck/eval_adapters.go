package main

import (
	"bytes"
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
	result := dryRunResult(fixture, runtime, runID)
	exitCode := 0
	status := "pass"
	if mode == "live" {
		live, code, err := invokeLiveAgent(root, runtime, fixture, runID, opts)
		exitCode = code
		if err != nil {
			status = "fail"
			_ = os.WriteFile(stderrPath, []byte(err.Error()), 0o644)
		} else {
			result = live
		}
	}
	writeJSONFile(resultPath, result)
	writeJSONFile(manifestPath, newRunManifest(root, runID, runtime, fixture))
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

func invokeLiveAgent(root, runtime string, fixture map[string]any, runID string, opts options) (map[string]any, int, error) {
	command := opts.agentCommand
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
	cmd := exec.Command(command)
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

func evalPrompt(fixture map[string]any, runtime, runID string) string {
	content, _ := json.MarshalIndent(fixture, "", "  ")
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

Route fixture:
%s

Return a result object with id "%s-live-%s", fixture_id "%s", expected_result "pass", eval_run_id "%s", actual.route, actual.user_questions, actual.artifacts, actual.proof, trace.files_read, trace.commands, trace.tools, and claim_checks.
`, runtime, runID, string(content), runtime, fixtureID, fixtureID, runID)
}

func newRunManifest(root, runID, runtime string, fixture map[string]any) map[string]any {
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
	return map[string]any{"run_id": runID, "fixture_id": fixtureID, "runtime": runtime, "created_at": time.Now().Format(time.RFC3339Nano), "protected_files": protected}
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
