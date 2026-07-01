package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type evalIssue struct {
	Result  string `json:"result,omitempty"`
	Case    string `json:"case,omitempty"`
	Message string `json:"message"`
}

type evalRow struct {
	File            string `json:"file"`
	FixtureID       string `json:"fixture_id,omitempty"`
	CaseID          string `json:"case_id,omitempty"`
	ExpectedResult  string `json:"expected_result"`
	ActualResult    string `json:"actual_result"`
	IssueCount      int    `json:"issue_count"`
	WarningCount    int    `json:"warning_count,omitempty"`
	TraceConfidence string `json:"trace_confidence,omitempty"`
	Computed        bool   `json:"computed,omitempty"`
	AmbiguousCount  int    `json:"ambiguous_count,omitempty"`
}

func runSkillEvalCommand(root string, opts options, stdout, stderr io.Writer) int {
	resultRoot := opts.resultRoot
	if resultRoot == "" {
		resultRoot = "evals/skill-eval/selftest"
	}
	result, err := computeSkillEval(root, resultRoot, opts.resultPath, opts.baseline, opts.updateBaseline, opts.requiredRunID, opts.manifestPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		mode := "results"
		if result.Selftest {
			mode = "selftest"
		}
		fmt.Fprintf(stdout, "Skill eval: %d %s files, %d issues\n", result.ResultCount, mode, len(result.Issues))
		for _, row := range result.Results {
			fmt.Fprintf(stdout, "%s: fixture=%s expected=%s actual=%s issues=%d\n", row.File, row.FixtureID, row.ExpectedResult, row.ActualResult, row.IssueCount)
		}
		for _, issue := range result.Issues {
			fmt.Fprintf(stdout, "ERROR [%s] %s\n", issue.Result, issue.Message)
		}
		for _, warning := range result.Warnings {
			fmt.Fprintf(stdout, "WARN  [%s] %s\n", warning.Result, warning.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

type skillEvalResult struct {
	OK          bool        `json:"ok"`
	ResultCount int         `json:"result_count"`
	Selftest    bool        `json:"selftest"`
	Results     []evalRow   `json:"results"`
	Baseline    any         `json:"baseline"`
	Issues      []evalIssue `json:"issues"`
	Warnings    []evalIssue `json:"warnings"`
}

func computeSkillEval(root, resultRoot, resultPath, baselinePath string, updateBaseline bool, requiredRunID, manifestPath string) (skillEvalResult, error) {
	fixtures, err := loadFixtureMap(root, "evals/route-complexity")
	if err != nil {
		return skillEvalResult{}, err
	}
	files, err := evalFiles(root, resultRoot, resultPath)
	if err != nil {
		return skillEvalResult{}, err
	}
	out := skillEvalResult{ResultCount: len(files), Selftest: resultPath == ""}
	out.Issues = append(out.Issues, validateRunManifest(root, manifestPath, requiredRunID)...)
	for _, file := range files {
		var result map[string]any
		if err := readJSONFile(file, &result); err != nil {
			out.Issues = append(out.Issues, evalIssue{Result: filepath.Base(file), Message: err.Error()})
			continue
		}
		issues, warnings, confidence := scoreSkillEvalResult(root, result, fixtures)
		expectedOutcome := expectedResult(result)
		actualPass := len(issues) == 0
		expectedPass := expectedOutcome == "pass"
		if out.Selftest && actualPass != expectedPass {
			if expectedPass {
				out.Issues = append(out.Issues, evalIssue{Result: filepath.Base(file), Message: "Self-test expected pass but scorer found issues."})
			} else {
				out.Issues = append(out.Issues, evalIssue{Result: filepath.Base(file), Message: "Self-test expected failure but scorer passed it."})
			}
		}
		if !out.Selftest || expectedPass {
			out.Issues = append(out.Issues, issues...)
		}
		out.Warnings = append(out.Warnings, warnings...)
		out.Results = append(out.Results, evalRow{File: filepath.Base(file), FixtureID: stringValue(result["fixture_id"]), ExpectedResult: expectedOutcome, ActualResult: passFail(actualPass), IssueCount: len(issues), WarningCount: len(warnings), TraceConfidence: confidence})
	}
	if baselinePath != "" {
		baseFull := resolveRepoPath(root, baselinePath)
		if updateBaseline {
			baseline := map[string]any{"schema_version": 1, "generated_at": time.Now().Format(time.RFC3339Nano), "result_root": resultRoot, "result_path": resultPath, "result_count": len(out.Results), "rows": out.Results}
			writeJSONFile(baseFull, baseline)
			out.Baseline = map[string]any{"path": baseFull, "updated": true, "issue_count": 0}
		} else {
			var baseline struct {
				Rows []evalRow `json:"rows"`
			}
			if err := readJSONFile(baseFull, &baseline); err != nil {
				out.Issues = append(out.Issues, evalIssue{Result: "baseline", Message: "BaselinePath does not exist: " + baselinePath})
			} else {
				issues := compareSkillEvalBaseline(baseline.Rows, out.Results)
				out.Issues = append(out.Issues, issues...)
				out.Baseline = map[string]any{"path": baseFull, "updated": false, "issue_count": len(issues)}
			}
		}
	}
	out.OK = len(out.Issues) == 0
	return out, nil
}

func scoreSkillEvalResult(root string, result map[string]any, fixtures map[string]map[string]any) ([]evalIssue, []evalIssue, string) {
	issues := []evalIssue{}
	warnings := []evalIssue{}
	id := stringValue(result["id"])
	if id == "" {
		id = "<missing-id>"
	}
	for _, field := range []string{"id", "fixture_id", "actual", "trace"} {
		if missingJSONField(result, field) {
			issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Missing top-level field '%s'.", field)})
		}
	}
	fixtureID := stringValue(result["fixture_id"])
	fixture, ok := fixtures[fixtureID]
	if !ok {
		issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Unknown fixture_id '%s'.", fixtureID)})
		return issues, warnings, "self-reported"
	}
	expected, _ := fixture["expected"].(map[string]any)
	actual, _ := result["actual"].(map[string]any)
	if actual == nil {
		actual = map[string]any{}
	}
	if stringValue(actual["route"]) != stringValue(expected["route"]) {
		issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Expected route '%s' but got '%s'.", stringValue(expected["route"]), stringValue(actual["route"]))})
	}
	if intValue(actual["user_questions"]) > intValue(expected["max_user_questions"]) {
		issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Expected at most %d user questions but got %d.", intValue(expected["max_user_questions"]), intValue(actual["user_questions"]))})
	}
	for _, artifact := range stringArray(expected["artifacts"]) {
		if !containsAny(stringArray(actual["artifacts"]), artifact) {
			issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Missing expected artifact evidence containing '%s'.", artifact)})
		}
	}
	for _, proof := range stringArray(expected["proof"]) {
		if !containsAny(stringArray(actual["proof"]), proof) {
			issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Missing expected proof evidence containing '%s'.", proof)})
		}
	}
	trace, _ := result["trace"].(map[string]any)
	if trace == nil {
		issues = append(issues, evalIssue{Result: id, Message: "Missing trace."})
		trace = map[string]any{}
	} else {
		for _, field := range []string{"files_read", "commands"} {
			if missingJSONField(trace, field) {
				issues = append(issues, evalIssue{Result: id, Message: "Missing trace." + field + "."})
			}
		}
	}
	for _, raw := range arrayValue(result["claim_checks"]) {
		check, _ := raw.(map[string]any)
		if !testEvalClaim(root, trace, check) {
			claim := stringValue(check["claim"])
			if claim == "" {
				claim = stringValue(check["type"])
			}
			issues = append(issues, evalIssue{Result: id, Message: "Claim check failed: " + claim})
		}
	}
	ruleIssues, ruleWarnings := traceRuleIssues(id, result, trace)
	issues = append(issues, ruleIssues...)
	warnings = append(warnings, ruleWarnings...)
	for _, artifact := range stringArray(result["claim_artifacts"]) {
		claimResult, err := computeSkillEvalClaims(root, "evals/skill-eval/claims", artifact)
		if err != nil || !claimResult.OK {
			issues = append(issues, evalIssue{Result: id, Message: "Claim artifact failed deterministic verification: " + artifact})
		}
	}
	confidence := "self-reported"
	if observed, _ := result["observed_trace"].(map[string]any); boolValue(observed["captured"]) {
		confidence = "observed"
	}
	return issues, warnings, confidence
}

func runSkillEvalClaimsCommand(root string, opts options, stdout, stderr io.Writer) int {
	claimRoot := opts.claimRoot
	if claimRoot == "" {
		claimRoot = "evals/skill-eval/claims"
	}
	result, err := computeSkillEvalClaims(root, claimRoot, opts.claimPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		mode := "claims"
		if result.Selftest {
			mode = "selftest"
		}
		fmt.Fprintf(stdout, "Skill claim eval: %d %s files, %d issues\n", result.CaseCount, mode, len(result.Issues))
		for _, row := range result.Results {
			fmt.Fprintf(stdout, "%s: case=%s expected=%s actual=%s issues=%d ambiguous=%d\n", row.File, row.CaseID, row.ExpectedResult, row.ActualResult, row.IssueCount, row.AmbiguousCount)
		}
		for _, issue := range result.Issues {
			fmt.Fprintf(stdout, "ERROR [%s] %s\n", issue.Case, issue.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

type claimEvalResult struct {
	OK        bool        `json:"ok"`
	CaseCount int         `json:"case_count"`
	Selftest  bool        `json:"selftest"`
	Results   []evalRow   `json:"results"`
	Issues    []evalIssue `json:"issues"`
}

func computeSkillEvalClaims(root, claimRoot, claimPath string) (claimEvalResult, error) {
	files, err := evalFiles(root, claimRoot, claimPath)
	if err != nil {
		return claimEvalResult{}, err
	}
	out := claimEvalResult{CaseCount: len(files), Selftest: claimPath == ""}
	for _, file := range files {
		var c map[string]any
		if err := readJSONFile(file, &c); err != nil {
			out.Issues = append(out.Issues, evalIssue{Case: filepath.Base(file), Message: err.Error()})
			continue
		}
		issues, ambiguous := scoreClaimCase(root, c)
		expected := expectedResult(c)
		actualPass := len(issues) == 0
		expectedPass := expected == "pass"
		if out.Selftest && actualPass != expectedPass {
			if expectedPass {
				out.Issues = append(out.Issues, evalIssue{Case: filepath.Base(file), Message: "Self-test expected pass but claim verifier found issues."})
			} else {
				out.Issues = append(out.Issues, evalIssue{Case: filepath.Base(file), Message: "Self-test expected failure but claim verifier passed it."})
			}
		}
		if !out.Selftest || expectedPass {
			out.Issues = append(out.Issues, issues...)
		}
		out.Results = append(out.Results, evalRow{File: filepath.Base(file), CaseID: stringValue(c["id"]), ExpectedResult: expected, ActualResult: passFail(actualPass), IssueCount: len(issues), AmbiguousCount: ambiguous})
	}
	out.OK = len(out.Issues) == 0
	return out, nil
}

func runSkillEvalQualityCommand(root string, opts options, stdout, stderr io.Writer) int {
	qualityRoot := opts.qualityRoot
	if qualityRoot == "" {
		qualityRoot = "evals/skill-eval/quality"
	}
	result, err := computeSkillEvalQuality(root, qualityRoot, opts.qualityPath, opts.minScore)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		mode := "quality cases"
		if result.Selftest {
			mode = "selftest"
		}
		fmt.Fprintf(stdout, "Skill quality eval: %d %s, %d issues\n", result.CaseCount, mode, len(result.Issues))
		for _, row := range result.Results {
			fmt.Fprintf(stdout, "%s: case=%s expected=%s actual=%s issues=%d\n", row.File, row.CaseID, row.ExpectedResult, row.ActualResult, row.IssueCount)
		}
		for _, issue := range result.Issues {
			fmt.Fprintf(stdout, "ERROR [%s] %s\n", issue.Case, issue.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

type qualityEvalResult struct {
	OK        bool        `json:"ok"`
	CaseCount int         `json:"case_count"`
	Selftest  bool        `json:"selftest"`
	MinScore  int         `json:"min_score"`
	Results   []evalRow   `json:"results"`
	Issues    []evalIssue `json:"issues"`
}

func computeSkillEvalQuality(root, qualityRoot, qualityPath string, minScore int) (qualityEvalResult, error) {
	fixtures, err := loadFixtureMap(root, "evals/route-complexity")
	if err != nil {
		return qualityEvalResult{}, err
	}
	files, err := evalFiles(root, qualityRoot, qualityPath)
	if err != nil {
		return qualityEvalResult{}, err
	}
	out := qualityEvalResult{CaseCount: len(files), Selftest: qualityPath == "", MinScore: minScore}
	for _, file := range files {
		var c map[string]any
		if err := readJSONFile(file, &c); err != nil {
			out.Issues = append(out.Issues, evalIssue{Case: filepath.Base(file), Message: err.Error()})
			continue
		}
		issues, computed := scoreQualityCase(c, fixtures, minScore)
		expected := expectedResult(c)
		actualPass := len(issues) == 0
		expectedPass := expected == "pass"
		if out.Selftest && actualPass != expectedPass {
			if expectedPass {
				out.Issues = append(out.Issues, evalIssue{Case: filepath.Base(file), Message: "Self-test expected pass but quality scorer found issues."})
			} else {
				out.Issues = append(out.Issues, evalIssue{Case: filepath.Base(file), Message: "Self-test expected failure but quality scorer passed it."})
			}
		}
		if !out.Selftest || expectedPass {
			out.Issues = append(out.Issues, issues...)
		}
		out.Results = append(out.Results, evalRow{File: filepath.Base(file), CaseID: stringValue(c["id"]), ExpectedResult: expected, ActualResult: passFail(actualPass), IssueCount: len(issues), Computed: computed})
	}
	out.OK = len(out.Issues) == 0
	return out, nil
}

func runSkillEvalRegressionCommand(root string, opts options, stdout, stderr io.Writer) int {
	runRoot := opts.runRoot
	if runRoot == "" {
		runRoot = ".kb/eval-runs"
	}
	report, err := computeSkillEvalRegression(root, runRoot, opts.baseline, opts.output)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, report)
	} else {
		fmt.Fprintf(stdout, "Skill eval regression report: rows=%d pass=%d non_pass=%d\n", intValue(report["row_count"]), intValue(report["pass_count"]), intValue(report["non_pass_count"]))
		fmt.Fprintf(stdout, "Report: %s\n", stringValue(report["output_path"]))
	}
	return 0
}

func computeSkillEvalRegression(root, runRoot, baselinePath, outputPath string) (map[string]any, error) {
	fullRoot := resolveRepoPath(root, runRoot)
	if _, err := os.Stat(fullRoot); err != nil {
		return nil, fmt.Errorf("RunRoot does not exist: %s", runRoot)
	}
	rows := []map[string]any{}
	summaries, _ := recursiveGlob(fullRoot, "summary.json")
	for _, summaryPath := range summaries {
		var summary map[string]any
		if err := readJSONFile(summaryPath, &summary); err != nil {
			continue
		}
		for _, raw := range arrayValue(summary["results"]) {
			result, _ := raw.(map[string]any)
			rows = append(rows, map[string]any{"source": summaryPath, "corpus_id": stringValue(summary["corpus_id"]), "runtime": stringValue(result["runtime"]), "fixture_id": stringValue(result["fixture_id"]), "mode": stringValue(result["mode"]), "status": stringValue(result["status"]), "exit_code": result["exit_code"], "duration_ms": result["duration_ms"], "result_path": stringValue(result["result_path"]), "result_bytes": fileSizeOrZero(stringValue(result["result_path"])), "stdout_bytes": fileSizeOrZero(stringValue(result["stdout"])), "stderr_bytes": fileSizeOrZero(stringValue(result["stderr"]))})
		}
	}
	if len(rows) == 0 {
		results, _ := recursiveGlob(fullRoot, "result.json")
		for _, resultPath := range results {
			var result map[string]any
			_ = readJSONFile(resultPath, &result)
			runtime := "unknown"
			if strings.Contains(resultPath, "ghcp") {
				runtime = "ghcp"
			} else if strings.Contains(resultPath, "codex") {
				runtime = "codex"
			}
			mode := "live"
			if strings.Contains(resultPath, "dry-run") {
				mode = "dry-run"
			}
			rows = append(rows, map[string]any{"source": resultPath, "corpus_id": "", "runtime": runtime, "fixture_id": stringValue(result["fixture_id"]), "mode": mode, "status": "pass", "exit_code": 0, "result_path": resultPath, "result_bytes": fileSizeOrZero(resultPath)})
		}
	}
	statusCounts := map[string]int{}
	passCount := 0
	totalResultBytes := int64(0)
	for _, row := range rows {
		status := stringValue(row["status"])
		statusCounts[status]++
		if status == "pass" {
			passCount++
		}
		totalResultBytes += int64(intValue(row["result_bytes"]))
	}
	report := map[string]any{"generated_at": time.Now().Format(time.RFC3339Nano), "run_root": fullRoot, "row_count": len(rows), "pass_count": passCount, "non_pass_count": len(rows) - passCount, "status_counts": statusCounts, "total_result_bytes": totalResultBytes, "rows": rows, "comparison": nil}
	if baselinePath != "" {
		var baseline map[string]any
		if err := readJSONFile(resolveRepoPath(root, baselinePath), &baseline); err != nil {
			return nil, fmt.Errorf("BaselinePath does not exist: %s", baselinePath)
		}
		report["comparison"] = map[string]any{"baseline": resolveRepoPath(root, baselinePath), "row_count_delta": len(rows) - intValue(baseline["row_count"]), "pass_count_delta": passCount - intValue(baseline["pass_count"]), "non_pass_count_delta": (len(rows) - passCount) - intValue(baseline["non_pass_count"]), "total_result_bytes_delta": totalResultBytes - int64(intValue(baseline["total_result_bytes"]))}
	}
	if outputPath == "" {
		outputPath = filepath.Join(fullRoot, "reports", "skill-eval-regression-"+time.Now().Format("20060102-150405")+".json")
	} else {
		outputPath = resolveRepoPath(root, outputPath)
	}
	writeJSONFile(outputPath, report)
	writeRegressionMarkdown(outputPath, report, rows)
	report["output_path"] = outputPath
	return report, nil
}

func runSkillEvalManifestSelftest(root string, stdout, stderr io.Writer) int {
	opts := options{fixtureID: "tiny-typo-fix", dryRun: true, keepRun: true}
	output, err := runEvalAdapter(root, opts, "codex")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(output.Runs) == 0 {
		fmt.Fprintln(stderr, "adapter produced no runs")
		return 1
	}
	run := output.Runs[0]
	defer os.RemoveAll(run.RunDir)
	good, err := computeSkillEval(root, "", run.ResultPath, "", false, run.RunID, run.ManifestPath)
	if err != nil || !good.OK {
		fmt.Fprintf(stderr, "Skill eval rejected a valid manifest: %v %#v\n", err, good.Issues)
		return 1
	}
	var manifest map[string]any
	if err := readJSONFile(run.ManifestPath, &manifest); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	protected := arrayValue(manifest["protected_files"])
	if len(protected) == 0 {
		fmt.Fprintln(stderr, "manifest missing protected_files")
		return 1
	}
	first, _ := protected[0].(map[string]any)
	first["sha256"] = strings.Repeat("0", 64)
	badPath := filepath.Join(run.RunDir, "manifest-bad.json")
	writeJSONFile(badPath, manifest)
	bad, err := computeSkillEval(root, "", run.ResultPath, "", false, run.RunID, badPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if bad.OK {
		fmt.Fprintln(stderr, "Skill eval accepted a manifest with a tampered protected-file SHA.")
		return 1
	}
	fmt.Fprintln(stdout, "Skill eval manifest selftest: valid manifest passed; tampered fixture SHA correctly rejected.")
	return 0
}

func runSkillEvalBaselineSelftest(root string, stdout, stderr io.Writer) int {
	tempRoot := filepath.Join(root, ".kb", "eval-baseline-selftest-"+randomShortHash())
	resultRoot := filepath.Join(tempRoot, "results")
	if err := os.MkdirAll(resultRoot, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(tempRoot)
	selftestFiles, _ := filepath.Glob(resolveRepoPath(root, "evals/skill-eval/selftest/*.json"))
	for _, file := range selftestFiles {
		content, _ := os.ReadFile(file)
		_ = os.WriteFile(filepath.Join(resultRoot, filepath.Base(file)), content, 0o644)
	}
	baselinePath := filepath.Join(tempRoot, "baseline.json")
	updated, err := computeSkillEval(root, resultRoot, "", baselinePath, true, "", "")
	if err != nil || !updated.OK {
		fmt.Fprintf(stderr, "Failed to create valid baseline: %v %#v\n", err, updated.Issues)
		return 1
	}
	compare, err := computeSkillEval(root, resultRoot, "", baselinePath, false, "", "")
	if err != nil || !compare.OK {
		fmt.Fprintf(stderr, "Unchanged baseline comparison failed: %v %#v\n", err, compare.Issues)
		return 1
	}
	passPath := filepath.Join(resultRoot, "pass-tiny-typo-fix.json")
	var pass map[string]any
	_ = readJSONFile(passPath, &pass)
	actual, _ := pass["actual"].(map[string]any)
	actual["proof"] = []string{}
	writeJSONFile(passPath, pass)
	regression, err := computeSkillEval(root, resultRoot, "", baselinePath, false, "", "")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if regression.OK {
		fmt.Fprintln(stderr, "Baseline comparison accepted a proof regression.")
		return 1
	}
	content, _ := os.ReadFile(resolveRepoPath(root, "evals/skill-eval/selftest/fail-proof-missing.json"))
	_ = os.WriteFile(filepath.Join(resultRoot, "fail-proof-missing.json"), content, 0o644)
	var negative map[string]any
	_ = readJSONFile(filepath.Join(resultRoot, "fail-proof-missing.json"), &negative)
	negativeActual, _ := negative["actual"].(map[string]any)
	negativeActual["proof"] = []string{"git diff --check"}
	negativeTrace, _ := negative["trace"].(map[string]any)
	negativeTrace["commands"] = []string{"git diff --check"}
	writeJSONFile(filepath.Join(resultRoot, "fail-proof-missing.json"), negative)
	negativeRegression, err := computeSkillEval(root, resultRoot, "", baselinePath, false, "", "")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if negativeRegression.OK {
		fmt.Fprintln(stderr, "Baseline comparison accepted a negative-fixture regression.")
		return 1
	}
	fmt.Fprintln(stdout, "Skill eval baseline selftest: baseline update passed; unchanged compare passed; proof regression correctly rejected; negative-fixture regression correctly rejected.")
	return 0
}

func validateRunManifest(root, manifestPath, requiredRunID string) []evalIssue {
	if manifestPath == "" {
		return nil
	}
	issues := []evalIssue{}
	full := resolveRepoPath(root, manifestPath)
	var manifest map[string]any
	if err := readJSONFile(full, &manifest); err != nil {
		return []evalIssue{{Result: "manifest", Message: "ManifestPath does not exist: " + manifestPath}}
	}
	if requiredRunID != "" && stringValue(manifest["run_id"]) != requiredRunID {
		issues = append(issues, evalIssue{Result: "manifest", Message: fmt.Sprintf("Expected manifest run_id '%s' but got '%s'.", requiredRunID, stringValue(manifest["run_id"]))})
	}
	protected := arrayValue(manifest["protected_files"])
	if len(protected) == 0 {
		return append(issues, evalIssue{Result: "manifest", Message: "Manifest is missing protected_files."})
	}
	for _, raw := range protected {
		entry, _ := raw.(map[string]any)
		role := stringValue(entry["role"])
		if role == "" {
			role = "<missing-role>"
		}
		pathValue := stringValue(entry["path"])
		expected := strings.ToLower(stringValue(entry["sha256"]))
		if pathValue == "" {
			issues = append(issues, evalIssue{Result: "manifest", Message: fmt.Sprintf("Protected file entry '%s' is missing path.", role)})
			continue
		}
		if expected == "" {
			issues = append(issues, evalIssue{Result: "manifest", Message: fmt.Sprintf("Protected file '%s' is missing sha256.", pathValue)})
			continue
		}
		actual := fileHashOrEmpty(resolveRepoPath(root, pathValue))
		if actual == "" {
			issues = append(issues, evalIssue{Result: "manifest", Message: fmt.Sprintf("Protected file '%s' is missing.", pathValue)})
			continue
		}
		if actual != expected {
			issues = append(issues, evalIssue{Result: "manifest", Message: fmt.Sprintf("Protected file '%s' changed for role '%s': expected %s but got %s.", pathValue, role, expected, actual)})
		}
	}
	return issues
}

func traceRuleIssues(id string, result, trace map[string]any) ([]evalIssue, []evalIssue) {
	issues := []evalIssue{}
	warnings := []evalIssue{}
	observed, _ := result["observed_trace"].(map[string]any)
	observedCaptured := boolValue(observed["captured"])
	if observedCaptured {
		for _, write := range stringArray(observed["writes"]) {
			if write != "" {
				issues = append(issues, evalIssue{Result: id, Message: "Observed write during routing eval: " + write})
			}
		}
		for _, deletePath := range stringArray(observed["deletes"]) {
			if deletePath != "" {
				issues = append(issues, evalIssue{Result: id, Message: "Observed delete during routing eval: " + deletePath})
			}
		}
	} else {
		warnings = append(warnings, evalIssue{Result: id, Message: "Observed writes/deletes were not captured; no-write routing invariant is unverified."})
	}
	rules, _ := result["trace_rules"].(map[string]any)
	if rules == nil {
		return issues, warnings
	}
	for _, rule := range []struct {
		Field string
		Trace []string
		Label string
	}{
		{"required_files_read", stringArray(trace["files_read"]), "required file read"},
		{"required_commands", stringArray(trace["commands"]), "required command"},
		{"required_tools", stringArray(trace["tools"]), "required tool"},
	} {
		for _, expected := range stringArray(rules[rule.Field]) {
			if !containsAny(rule.Trace, expected) {
				issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Missing trace rule evidence for %s containing '%s'.", rule.Label, expected)})
			}
		}
	}
	forbiddenCommands := stringArray(trace["commands"])
	forbiddenTools := stringArray(trace["tools"])
	if observedCaptured {
		forbiddenCommands = stringArray(observed["commands"])
		forbiddenTools = []string{}
		if len(stringArray(rules["forbidden_tools"])) > 0 {
			warnings = append(warnings, evalIssue{Result: id, Message: "observed_trace has no tools field in v1; forbidden_tools cannot be externally enforced."})
		}
	} else if len(stringArray(rules["forbidden_commands"])) > 0 || len(stringArray(rules["forbidden_tools"])) > 0 {
		warnings = append(warnings, evalIssue{Result: id, Message: "Forbidden command/tool rules are using self-reported trace because observed_trace was not captured."})
	}
	for _, forbidden := range stringArray(rules["forbidden_commands"]) {
		if containsAny(forbiddenCommands, forbidden) {
			issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Forbidden trace rule matched forbidden command containing '%s'.", forbidden)})
		}
	}
	for _, forbidden := range stringArray(rules["forbidden_tools"]) {
		if containsAny(forbiddenTools, forbidden) {
			issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Forbidden trace rule matched forbidden tool containing '%s'.", forbidden)})
		}
	}
	return issues, warnings
}

func testEvalClaim(root string, trace map[string]any, check map[string]any) bool {
	expected := true
	if hasJSONField(check, "expected") {
		expected = boolValue(check["expected"])
	}
	actual := false
	switch stringValue(check["type"]) {
	case "file_exists":
		_, err := os.Stat(resolveRepoPath(root, stringValue(check["path"])))
		actual = err == nil
	case "command_ran":
		actual = containsAny(stringArray(trace["commands"]), stringValue(check["contains"]))
	case "file_read":
		actual = containsAny(stringArray(trace["files_read"]), stringValue(check["contains"]))
	default:
		return false
	}
	return actual == expected
}

func scoreClaimCase(root string, c map[string]any) ([]evalIssue, int) {
	issues := []evalIssue{}
	ambiguous := 0
	id := stringValue(c["id"])
	if id == "" {
		id = "<missing-id>"
	}
	trace, _ := c["trace"].(map[string]any)
	if trace == nil {
		trace = map[string]any{"files_read": []any{}, "commands": []any{}, "tools": []any{}}
	}
	for _, raw := range arrayValue(c["claims"]) {
		claim, _ := raw.(map[string]any)
		if stringValue(claim["type"]) == "ambiguous" {
			ambiguous++
			continue
		}
		if !testEvalClaim(root, trace, claim) {
			text := stringValue(claim["claim"])
			if text == "" {
				text = stringValue(claim["type"])
			}
			issues = append(issues, evalIssue{Case: id, Message: "Claim check failed: " + text})
		}
	}
	return issues, ambiguous
}

func scoreQualityCase(c map[string]any, fixtures map[string]map[string]any, minScore int) ([]evalIssue, bool) {
	id := stringValue(c["id"])
	if id == "" {
		id = "<missing-id>"
	}
	if hasJSONField(c, "quality") {
		return []evalIssue{{Case: id, Message: "Hand-authored quality objects are not accepted; provide input_result so scores are computed."}}, false
	}
	input, _ := c["input_result"].(map[string]any)
	if input == nil {
		return []evalIssue{{Case: id, Message: "Missing input_result; quality scores must be computed from captured output."}}, false
	}
	quality := measureQuality(input, fixtures)
	issues := []evalIssue{}
	expectedQuality, _ := c["expected_quality"].(map[string]any)
	for _, dimension := range []string{"completeness", "maintainability", "relevance", "proof_quality", "right_sized_ceremony"} {
		score := quality[dimension]
		if score < minScore {
			issues = append(issues, evalIssue{Case: id, Message: fmt.Sprintf("Quality dimension '%s' scored %d, below threshold %d.", dimension, score, minScore)})
		}
		if expectedQuality != nil {
			if raw, ok := expectedQuality[dimension].(map[string]any); ok && hasJSONField(raw, "score") && score != intValue(raw["score"]) {
				issues = append(issues, evalIssue{Case: id, Message: fmt.Sprintf("Computed '%s' score %d, expected %d.", dimension, score, intValue(raw["score"]))})
			}
		}
	}
	return issues, true
}

func measureQuality(result map[string]any, fixtures map[string]map[string]any) map[string]int {
	actual, _ := result["actual"].(map[string]any)
	if actual == nil {
		actual = map[string]any{}
	}
	route := stringValue(actual["route"])
	artifacts := stringArray(actual["artifacts"])
	proof := stringArray(actual["proof"])
	questions := intValue(actual["user_questions"])
	quality := map[string]int{"completeness": 1, "maintainability": 5, "relevance": 3, "proof_quality": 1, "right_sized_ceremony": 4}
	if route != "" && len(artifacts) > 0 && len(proof) > 0 {
		quality["completeness"] = 5
	} else if route != "" && (len(artifacts) > 0 || len(proof) > 0) {
		quality["completeness"] = 3
	}
	allText := strings.Join(append(artifacts, proof...), " ")
	if len(artifacts) > 8 || len(proof) > 8 {
		quality["maintainability"] = 2
	}
	if regexpVague(allText) {
		quality["maintainability"] = 1
	}
	fixture := fixtures[stringValue(result["fixture_id"])]
	expected, _ := fixture["expected"].(map[string]any)
	if expected != nil {
		if route == stringValue(expected["route"]) {
			quality["relevance"] = 5
		} else {
			quality["relevance"] = 1
		}
	}
	if len(proof) > 0 {
		quality["proof_quality"] = 3
		if expected != nil {
			missing := false
			for _, item := range stringArray(expected["proof"]) {
				if !containsAny(proof, item) {
					missing = true
				}
			}
			if !missing {
				quality["proof_quality"] = 5
			}
		} else {
			quality["proof_quality"] = 4
		}
	}
	if expected != nil {
		maxQuestions := intValue(expected["max_user_questions"])
		tier := stringValue(expected["complexity_tier"])
		if questions > maxQuestions {
			quality["right_sized_ceremony"] = 1
		} else if tier == "small" && contains([]string{"kb-brainstorm", "kb-plan", "kb-epic"}, route) {
			quality["right_sized_ceremony"] = 1
		} else if tier == "large" && contains([]string{"kb-fix", "kb-work"}, route) {
			quality["right_sized_ceremony"] = 1
		} else if route != "" && questions <= maxQuestions {
			quality["right_sized_ceremony"] = 5
		}
	}
	return quality
}

func evalFiles(root, fileRoot, singlePath string) ([]string, error) {
	if singlePath != "" {
		full := resolveRepoPath(root, singlePath)
		if _, err := os.Stat(full); err != nil {
			return nil, fmt.Errorf("path does not exist: %s", singlePath)
		}
		return []string{full}, nil
	}
	fullRoot := resolveRepoPath(root, fileRoot)
	files, err := filepath.Glob(filepath.Join(fullRoot, "*.json"))
	if err != nil || len(files) == 0 {
		return nil, fmt.Errorf("root does not exist or has no JSON files: %s", fileRoot)
	}
	sort.Strings(files)
	return files, nil
}

func loadFixtureMap(root, fixtureRoot string) (map[string]map[string]any, error) {
	files, err := evalFiles(root, fixtureRoot, "")
	if err != nil {
		return nil, err
	}
	out := map[string]map[string]any{}
	for _, file := range files {
		var fixture map[string]any
		if err := readJSONFile(file, &fixture); err != nil {
			continue
		}
		if id := stringValue(fixture["id"]); id != "" {
			out[id] = fixture
		}
	}
	return out, nil
}

func compareSkillEvalBaseline(baseline, current []evalRow) []evalIssue {
	issues := []evalIssue{}
	byFile := map[string]evalRow{}
	for _, row := range current {
		byFile[row.File] = row
	}
	for _, old := range baseline {
		row, ok := byFile[old.File]
		if !ok {
			issues = append(issues, evalIssue{Result: "baseline", Message: fmt.Sprintf("Baseline row '%s' is missing from current results.", old.File)})
			continue
		}
		if row.FixtureID != old.FixtureID {
			issues = append(issues, evalIssue{Result: old.File, Message: fmt.Sprintf("Fixture changed from '%s' to '%s'.", old.FixtureID, row.FixtureID)})
		}
		if row.ExpectedResult != old.ExpectedResult {
			issues = append(issues, evalIssue{Result: old.File, Message: fmt.Sprintf("Expected result changed from '%s' to '%s'.", old.ExpectedResult, row.ExpectedResult)})
		}
		if old.ExpectedResult == "fail" && row.ActualResult != "fail" {
			issues = append(issues, evalIssue{Result: old.File, Message: fmt.Sprintf("Negative fixture regressed from fail to '%s'.", row.ActualResult)})
		}
		if old.ActualResult == "pass" && row.ActualResult != "pass" {
			issues = append(issues, evalIssue{Result: old.File, Message: fmt.Sprintf("Result regressed from pass to '%s'.", row.ActualResult)})
		}
		if row.IssueCount > old.IssueCount {
			issues = append(issues, evalIssue{Result: old.File, Message: fmt.Sprintf("Issue count regressed from %d to %d.", old.IssueCount, row.IssueCount)})
		}
	}
	return issues
}

func recursiveGlob(root, name string) ([]string, error) {
	files := []string{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		if entry.Name() == name {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

func fileSizeOrZero(path string) int64 {
	if path == "" {
		return 0
	}
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func writeRegressionMarkdown(jsonPath string, report map[string]any, rows []map[string]any) {
	mdPath := strings.TrimSuffix(jsonPath, filepath.Ext(jsonPath)) + ".md"
	lines := []string{"# Skill Eval Regression Report", "", "- Generated: " + stringValue(report["generated_at"]), "- Run root: " + stringValue(report["run_root"]), fmt.Sprintf("- Rows: %d", intValue(report["row_count"])), fmt.Sprintf("- Pass: %d", intValue(report["pass_count"])), fmt.Sprintf("- Non-pass: %d", intValue(report["non_pass_count"])), "", "| Runtime | Fixture | Mode | Status | Duration ms | Result bytes |", "|---|---|---|---|---:|---:|"}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %v | %v |", stringValue(row["runtime"]), stringValue(row["fixture_id"]), stringValue(row["mode"]), stringValue(row["status"]), row["duration_ms"], row["result_bytes"]))
	}
	_ = os.WriteFile(mdPath, []byte(strings.Join(lines, "\n")), 0o644)
}

func expectedResult(obj map[string]any) string {
	if value := stringValue(obj["expected_result"]); value != "" {
		return value
	}
	return "pass"
}

func passFail(pass bool) string {
	if pass {
		return "pass"
	}
	return "fail"
}

func containsAny(items []string, expected string) bool {
	needle := normalizeEvalText(expected)
	for _, item := range items {
		if strings.Contains(normalizeEvalText(item), needle) {
			return true
		}
	}
	return false
}

func normalizeEvalText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func regexpVague(value string) bool {
	for _, token := range []string{"stuff", "things", "misc", "various", "whatever", "???", "todo later"} {
		if strings.Contains(strings.ToLower(value), token) {
			return true
		}
	}
	return false
}
