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
	"regexp"
	"sort"
	"strings"
	"time"
)

type simpleIssue struct {
	Path    string `json:"path,omitempty"`
	Fixture string `json:"fixture,omitempty"`
	Message string `json:"message"`
}

func runBenchmarkValidateCommand(root string, opts options, stdout, stderr io.Writer) int {
	fixtureRoot := opts.fixtureRoot
	if fixtureRoot == "" {
		fixtureRoot = "evals/cross-model-benchmarks"
	}
	result, err := computeBenchmarkValidate(root, fixtureRoot)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "Cross-model benchmark fixtures: files=%d cases=%d issues=%d\n", result.Files, result.Cases, len(result.Issues))
		for _, issue := range result.Issues {
			fmt.Fprintf(stdout, "ERROR [%s] %s\n", issue.Path, issue.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

type benchmarkResult struct {
	OK          bool          `json:"ok"`
	FixtureRoot string        `json:"fixture_root"`
	Files       int           `json:"files"`
	Cases       int           `json:"cases"`
	Issues      []simpleIssue `json:"issues"`
}

func computeBenchmarkValidate(root, fixtureRoot string) (benchmarkResult, error) {
	result := benchmarkResult{OK: true, FixtureRoot: fixtureRoot}
	fullRoot := resolveRepoPath(root, fixtureRoot)
	files, err := filepath.Glob(filepath.Join(fullRoot, "*.json"))
	if err != nil || len(files) == 0 {
		if err != nil {
			return result, err
		}
		if _, statErr := os.Stat(fullRoot); statErr != nil {
			return result, fmt.Errorf("fixture root not found: %s", fixtureRoot)
		}
		result.Issues = append(result.Issues, simpleIssue{Path: fixtureRoot, Message: "No benchmark fixture JSON files found."})
		result.OK = false
		return result, nil
	}
	sort.Strings(files)
	result.Files = len(files)
	ids := map[string]bool{}
	for _, file := range files {
		relative := relativePath(root, file)
		var fixture map[string]any
		if err := readJSONFile(file, &fixture); err != nil {
			result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: "Invalid JSON: " + err.Error()})
			continue
		}
		if intValue(fixture["schema_version"]) != 1 {
			result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: "schema_version must be 1."})
		}
		if stringValue(fixture["suite"]) == "" {
			result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: "Missing suite."})
		}
		cases := arrayValue(fixture["cases"])
		if len(cases) == 0 {
			result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: "Missing cases."})
			continue
		}
		for _, rawCase := range cases {
			result.Cases++
			c, _ := rawCase.(map[string]any)
			for _, field := range []string{"id", "category", "prompt", "expected", "forbidden_failures", "scoring"} {
				if missingJSONField(c, field) {
					result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: fmt.Sprintf("Case missing required field '%s'.", field)})
				}
			}
			id := stringValue(c["id"])
			if id != "" {
				if ids[id] {
					result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: fmt.Sprintf("Duplicate case id '%s'.", id)})
				}
				ids[id] = true
			}
			expected, _ := c["expected"].(map[string]any)
			if expected != nil && !hasJSONField(expected, "must_include") {
				result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: fmt.Sprintf("Case '%s' expected block must include must_include.", id)})
			}
			if len(arrayValue(c["forbidden_failures"])) == 0 {
				result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: fmt.Sprintf("Case '%s' must list forbidden_failures.", id)})
			}
			if scoring, ok := c["scoring"].(map[string]any); ok && len(scoring) == 0 {
				result.Issues = append(result.Issues, simpleIssue{Path: relative, Message: fmt.Sprintf("Case '%s' scoring must define at least one dimension.", id)})
			}
		}
	}
	result.OK = len(result.Issues) == 0
	return result, nil
}

func runRouteEvalCommand(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-quality.json"
	}
	result, err := computeRouteEval(root, configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "Route complexity eval: %d fixtures, %d issues\n", result.FixtureCount, len(result.Issues))
		for _, row := range result.Results {
			fmt.Fprintf(stdout, "%s: route=%s tier=%s score=%d guards=%s\n", row.ID, row.Route, row.Tier, row.Score, row.Guards)
		}
		for _, issue := range result.Issues {
			fmt.Fprintf(stdout, "ERROR [%s] %s\n", issue.Fixture, issue.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

type routeEvalResult struct {
	OK           bool          `json:"ok"`
	FixtureCount int           `json:"fixture_count"`
	Issues       []simpleIssue `json:"issues"`
	Results      []routeRow    `json:"results"`
}

type routeRow struct {
	ID     string `json:"id"`
	Route  string `json:"route"`
	Tier   string `json:"tier"`
	Score  int    `json:"score"`
	Guards string `json:"guards"`
}

func computeRouteEval(root, configPath string) (routeEvalResult, error) {
	var result routeEvalResult
	var raw struct {
		RouteComplexity struct {
			FixtureRoot      string   `json:"fixture_root"`
			AllowedPlatforms []string `json:"allowed_platforms"`
			AllowedRoutes    []string `json:"allowed_routes"`
			Rubric           struct {
				SmallMax    int `json:"small_max"`
				StandardMax int `json:"standard_max"`
			} `json:"rubric"`
		} `json:"route_complexity"`
	}
	if err := readJSONFile(resolveRepoPath(root, configPath), &raw); err != nil {
		return result, err
	}
	fixtureRoot := resolveRepoPath(root, raw.RouteComplexity.FixtureRoot)
	files, err := filepath.Glob(filepath.Join(fixtureRoot, "*.json"))
	if err != nil || len(files) == 0 {
		return result, fmt.Errorf("no route sizing fixtures found in %s", raw.RouteComplexity.FixtureRoot)
	}
	sort.Strings(files)
	result.FixtureCount = len(files)
	allGuards := []string{}
	for _, file := range files {
		name := filepath.Base(file)
		var fixture map[string]any
		if err := readJSONFile(file, &fixture); err != nil {
			result.Issues = append(result.Issues, simpleIssue{Fixture: name, Message: "Invalid JSON: " + err.Error()})
			continue
		}
		for _, field := range []string{"id", "platforms", "prompt", "repo_state", "expected", "complexity_signals", "guards"} {
			if missingJSONField(fixture, field) {
				result.Issues = append(result.Issues, simpleIssue{Fixture: name, Message: fmt.Sprintf("Missing top-level field '%s'.", field)})
			}
		}
		expected, _ := fixture["expected"].(map[string]any)
		for _, field := range []string{"route", "complexity_tier", "max_user_questions", "artifacts", "proof"} {
			if expected != nil && missingJSONField(expected, field) {
				result.Issues = append(result.Issues, simpleIssue{Fixture: name, Message: fmt.Sprintf("Missing expected field '%s'.", field)})
			}
		}
		signals, _ := fixture["complexity_signals"].(map[string]any)
		requiredSignals := []string{"subsystem_count", "uncertainty", "user_visible", "data_auth_security_risk", "external_dependency", "verification_surface", "rollback_difficulty", "expected_duration_hours"}
		missingSignals := false
		for _, field := range requiredSignals {
			if signals != nil && missingJSONField(signals, field) {
				missingSignals = true
				result.Issues = append(result.Issues, simpleIssue{Fixture: name, Message: fmt.Sprintf("Missing complexity_signals field '%s'.", field)})
			}
		}
		for _, platform := range stringArray(fixture["platforms"]) {
			if !contains(raw.RouteComplexity.AllowedPlatforms, platform) {
				result.Issues = append(result.Issues, simpleIssue{Fixture: name, Message: fmt.Sprintf("Unknown platform '%s'.", platform)})
			}
		}
		if expected != nil {
			route := stringValue(expected["route"])
			if route != "" && !contains(raw.RouteComplexity.AllowedRoutes, route) {
				result.Issues = append(result.Issues, simpleIssue{Fixture: name, Message: fmt.Sprintf("Unknown expected route '%s'.", route)})
			}
		}
		guards := stringArray(fixture["guards"])
		allGuards = append(allGuards, guards...)
		if signals != nil && expected != nil && !missingSignals && !missingJSONField(expected, "complexity_tier") {
			score := routeScore(signals)
			tier := "large"
			if score <= raw.RouteComplexity.Rubric.SmallMax {
				tier = "small"
			} else if score <= raw.RouteComplexity.Rubric.StandardMax {
				tier = "standard"
			}
			if tier != stringValue(expected["complexity_tier"]) {
				result.Issues = append(result.Issues, simpleIssue{Fixture: name, Message: fmt.Sprintf("Expected tier '%s' but rubric computed '%s' (score %d).", stringValue(expected["complexity_tier"]), tier, score)})
			}
			result.Results = append(result.Results, routeRow{ID: stringValue(fixture["id"]), Route: stringValue(expected["route"]), Tier: tier, Score: score, Guards: strings.Join(guards, ",")})
		}
	}
	for _, guard := range []string{"over-planning", "under-planning"} {
		if !contains(allGuards, guard) {
			result.Issues = append(result.Issues, simpleIssue{Fixture: "suite", Message: fmt.Sprintf("Missing required guard '%s'.", guard)})
		}
	}
	result.OK = len(result.Issues) == 0
	return result, nil
}

func routeScore(signals map[string]any) int {
	verificationWeights := map[string]int{"none": 0, "unit": 1, "integration": 2, "functional": 3, "full": 4}
	duration := floatValue(signals["expected_duration_hours"])
	durationScore := 4
	if duration <= 0.5 {
		durationScore = 0
	} else if duration <= 2 {
		durationScore = 1
	} else if duration <= 8 {
		durationScore = 2
	}
	score := intValue(signals["subsystem_count"]) + intValue(signals["uncertainty"]) + intValue(signals["data_auth_security_risk"]) + intValue(signals["rollback_difficulty"]) + durationScore
	if boolValue(signals["external_dependency"]) {
		score++
	}
	if boolValue(signals["user_visible"]) {
		score++
	}
	score += verificationWeights[stringValue(signals["verification_surface"])]
	return score
}

func runReleaseSelftestCommand(stdout, stderr io.Writer) int {
	root, err := os.MkdirTemp("", "kb-release-gate-*")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(root)
	writeFixture := func(name, testBody string) (string, error) {
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+name+"\n"), 0o644); err != nil {
			return "", err
		}
		if err := os.WriteFile(filepath.Join(dir, "fixture_test.go"), []byte(testBody), 0o644); err != nil {
			return "", err
		}
		_ = exec.Command("git", "-C", dir, "init").Run()
		_ = exec.Command("git", "-C", dir, "config", "user.email", "test@example.com").Run()
		_ = exec.Command("git", "-C", dir, "config", "user.name", "Release Gate Test").Run()
		_ = exec.Command("git", "-C", dir, "add", ".").Run()
		_ = exec.Command("git", "-C", dir, "commit", "-m", "fixture").Run()
		return dir, nil
	}
	passRoot, _ := writeFixture("passfixture", `package passfixture
import "testing"
func TestFixture(t *testing.T) {}
`)
	failRoot, _ := writeFixture("failfixture", `package failfixture
import "testing"
func TestFixture(t *testing.T) { t.Fatal("expected failure") }
`)
	if code := run([]string{"local-release", "--root", passRoot, "--json"}, io.Discard, io.Discard); code != 0 {
		fmt.Fprintf(stderr, "local-release should pass with successful required checks; exit=%d\n", code)
		return 1
	}
	if code := run([]string{"live-release", "--root", passRoot, "--json"}, io.Discard, io.Discard); code != 0 {
		fmt.Fprintf(stderr, "live-release should pass when live corpus runner is explicitly unavailable; exit=%d\n", code)
		return 1
	}
	if code := run([]string{"local-release", "--root", failRoot, "--json"}, io.Discard, io.Discard); code == 0 {
		fmt.Fprintln(stderr, "required native core failure should make the gate fail")
		return 1
	}
	fmt.Fprintln(stdout, "kb-release-gate selftest passed")
	return 0
}

func runSurfaceReportCommand(root string, opts options, stdout, stderr io.Writer) int {
	skillRoot := opts.skillRoot
	if skillRoot == "" {
		skillRoot = ".github/skills"
	}
	report, err := computeSurfaceReport(root, skillRoot, opts.route, opts.baseline)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.output != "" {
		full := resolveRepoPath(root, opts.output)
		_ = os.MkdirAll(filepath.Dir(full), 0o755)
		content, _ := json.MarshalIndent(report, "", "  ")
		_ = os.WriteFile(full, content, 0o644)
	}
	if opts.json {
		writeJSON(stdout, report)
	} else {
		fmt.Fprintf(stdout, "Skill surface report: routes=%d\n", len(report.Routes))
		for _, row := range report.Routes {
			fmt.Fprintf(stdout, "%s: skills=%d lines=%d token_estimate=%d hash=%s\n", row.Route, row.SkillCount, row.TotalLines, row.TokenEstimate, shortHash(row.CombinedHash))
			if len(row.Missing) > 0 {
				fmt.Fprintf(stdout, "WARN [%s] missing skills: %s\n", row.Route, strings.Join(row.Missing, ", "))
			}
		}
	}
	return 0
}

type surfaceReport struct {
	GeneratedAt string         `json:"generated_at"`
	SkillRoot   string         `json:"skill_root"`
	Routes      []surfaceRoute `json:"routes"`
	Comparison  any            `json:"comparison"`
}

type surfaceRoute struct {
	Route         string      `json:"route"`
	Skills        []skillInfo `json:"skills"`
	SkillCount    int         `json:"skill_count"`
	Missing       []string    `json:"missing"`
	TotalLines    int         `json:"total_lines"`
	TokenEstimate int         `json:"token_estimate"`
	CombinedHash  string      `json:"combined_hash"`
}

type skillInfo struct {
	Name          string `json:"name"`
	Exists        bool   `json:"exists"`
	Lines         int    `json:"lines"`
	TokenEstimate int    `json:"token_estimate"`
	Hash          string `json:"hash"`
}

func computeSurfaceReport(root, skillRoot, routeFilter, baselinePath string) (surfaceReport, error) {
	fullRoot := resolveRepoPath(root, skillRoot)
	routes := map[string][]string{
		"base":        {"kb-start", "kb-map", "kb-first-principles", "kb-check"},
		"kb-plan":     {"kb-start", "kb-map", "kb-plan", "kb-check"},
		"kb-work":     {"kb-start", "kb-map", "kb-work", "kb-check"},
		"kb-goal":     {"kb-start", "kb-map", "kb-goal", "kb-check"},
		"kb-epic":     {"kb-start", "kb-map", "kb-brainstorm", "kb-plan", "kb-epic", "kb-check"},
		"kb-complete": {"kb-start", "kb-map", "kb-complete", "kb-review", "kb-check", "learn", "evolve"},
	}
	order := []string{"base", "kb-plan", "kb-work", "kb-goal", "kb-epic", "kb-complete"}
	if routeFilter != "" {
		if _, ok := routes[routeFilter]; !ok {
			return surfaceReport{}, fmt.Errorf("unknown route '%s'. Known routes: %s", routeFilter, strings.Join(order, ", "))
		}
		order = []string{routeFilter}
	}
	report := surfaceReport{GeneratedAt: time.Now().Format(time.RFC3339Nano), SkillRoot: fullRoot}
	for _, route := range order {
		row := surfaceRoute{Route: route}
		hashParts := []string{}
		for _, skill := range routes[route] {
			info := readSkillInfo(fullRoot, skill)
			row.Skills = append(row.Skills, info)
			row.SkillCount++
			if !info.Exists {
				row.Missing = append(row.Missing, skill)
			}
			row.TotalLines += info.Lines
			row.TokenEstimate += info.TokenEstimate
			hashParts = append(hashParts, info.Name+":"+info.Hash)
		}
		row.CombinedHash = hashString(strings.Join(hashParts, "\n"))
		report.Routes = append(report.Routes, row)
	}
	if baselinePath != "" {
		var baseline surfaceReport
		if err := readJSONFile(resolveRepoPath(root, baselinePath), &baseline); err != nil {
			return report, fmt.Errorf("BaselinePath does not exist: %s", baselinePath)
		}
		type comparisonRow struct {
			Route       string `json:"route"`
			LineDelta   int    `json:"line_delta"`
			TokenDelta  int    `json:"token_delta"`
			HashChanged bool   `json:"hash_changed"`
		}
		comparison := []comparisonRow{}
		byRoute := map[string]surfaceRoute{}
		for _, row := range baseline.Routes {
			byRoute[row.Route] = row
		}
		for _, row := range report.Routes {
			if old, ok := byRoute[row.Route]; ok {
				comparison = append(comparison, comparisonRow{Route: row.Route, LineDelta: row.TotalLines - old.TotalLines, TokenDelta: row.TokenEstimate - old.TokenEstimate, HashChanged: row.CombinedHash != old.CombinedHash})
			}
		}
		report.Comparison = comparison
	}
	return report, nil
}

func readSkillInfo(skillRoot, name string) skillInfo {
	path := filepath.Join(skillRoot, name, "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return skillInfo{Name: name}
	}
	text := string(content)
	return skillInfo{Name: name, Exists: true, Lines: countLines(text), TokenEstimate: tokenEstimate(text), Hash: shortHash(hashString(text))}
}

func runPipelineCommand(root string, opts options, stdout, stderr io.Writer) int {
	switch {
	case opts.start != "":
		return startPipeline(root, opts.start, stdout, stderr)
	case opts.status:
		return pipelineStatus(root, opts.runID, stdout, stderr)
	default:
		fmt.Fprintln(stdout, "Usage:")
		fmt.Fprintln(stdout, "  kbcheck pipeline --start skill-bundle-proof-spike")
		fmt.Fprintln(stdout, "  kbcheck pipeline --status [--run-id <id>]")
		return 1
	}
}

func startPipeline(root, pipelineID string, stdout, stderr io.Writer) int {
	path := resolveRepoPath(root, fmt.Sprintf("config/pipelines/%s.json", pipelineID))
	var pipeline map[string]any
	if err := readJSONFile(path, &pipeline); err != nil {
		fmt.Fprintf(stderr, "Pipeline '%s' not found at config/pipelines/%s.json\n", pipelineID, pipelineID)
		return 1
	}
	runRoot := resolveRepoPath(root, ".kb/pipeline-runs")
	_ = os.MkdirAll(runRoot, 0o755)
	runID := time.Now().Format("20060102-150405-000") + "-" + randomShortHash() + "-" + slug(stringValue(pipeline["id"]))
	runDir := filepath.Join(runRoot, runID)
	_ = os.MkdirAll(filepath.Join(runDir, "phase-prompts"), 0o755)
	protected := []map[string]any{}
	for _, raw := range arrayValue(pipeline["protected_files"]) {
		entry, _ := raw.(map[string]any)
		pathValue := stringValue(entry["path"])
		full := resolveRepoPath(root, pathValue)
		exists := false
		hash := ""
		if content, err := os.ReadFile(full); err == nil {
			exists = true
			sum := sha256.Sum256(content)
			hash = hex.EncodeToString(sum[:])
		}
		protected = append(protected, map[string]any{"role": stringValue(entry["role"]), "path": pathValue, "sha256": hash, "exists": exists})
	}
	phases := arrayValue(pipeline["phases"])
	current := ""
	if len(phases) > 0 {
		if first, ok := phases[0].(map[string]any); ok {
			current = stringValue(first["id"])
		}
	}
	run := map[string]any{"run_id": runID, "pipeline_id": stringValue(pipeline["id"]), "started_at": time.Now().Format(time.RFC3339Nano), "status": "started", "run_dir": runDir, "phase_count": len(phases), "current_phase": current, "proof_commands": arrayValue(pipeline["proof_commands"]), "protected_files": protected}
	writeJSONFile(filepath.Join(runDir, "pipeline.json"), pipeline)
	writeJSONFile(filepath.Join(runDir, "run.json"), run)
	writeJSONFile(filepath.Join(runDir, "protected-files.json"), map[string]any{"run_id": runID, "protected_files": protected})
	writeJSONFile(filepath.Join(runDir, "proof.json"), map[string]any{"run_id": runID, "proof_commands": arrayValue(pipeline["proof_commands"]), "results": []any{}})
	for _, raw := range phases {
		phase, _ := raw.(map[string]any)
		phaseID := stringValue(phase["id"])
		prompt := fmt.Sprintf("# Phase: %s\n\nPipeline: %s\nFresh context: %v\nSkills: %s\nRequired outputs: %s\n\n## Instructions\n\n- Read pipeline.json in this run directory.\n- Read only the context needed for this phase.\n- Produce the required outputs listed above.\n- Do not execute later phases from this prompt.\n", phaseID, stringValue(pipeline["id"]), boolValue(phase["fresh_context"]), strings.Join(stringArray(phase["skills"]), ", "), strings.Join(stringArray(phase["required_outputs"]), ", "))
		_ = os.WriteFile(filepath.Join(runDir, "phase-prompts", phaseID+".md"), []byte(prompt), 0o644)
	}
	selected := "# Selected Pipeline\n\n- Run ID: " + runID + "\n- Pipeline: " + stringValue(pipeline["id"]) + "\n- Status: started\n- Run directory: " + runDir + "\n\n| Phase | Fresh Context | Skills | Required Outputs |\n|---|---|---|---|\n"
	for _, raw := range phases {
		phase, _ := raw.(map[string]any)
		selected += fmt.Sprintf("| %s | %v | %s | %s |\n", stringValue(phase["id"]), boolValue(phase["fresh_context"]), strings.Join(stringArray(phase["skills"]), ", "), strings.Join(stringArray(phase["required_outputs"]), ", "))
	}
	_ = os.WriteFile(filepath.Join(runDir, "selected-pipeline.md"), []byte(selected), 0o644)
	fmt.Fprintf(stdout, "KB pipeline started: %s\n", runID)
	fmt.Fprintf(stdout, "Run directory: %s\n", runDir)
	return 0
}

func pipelineStatus(root, runID string, stdout, stderr io.Writer) int {
	runRoot := resolveRepoPath(root, ".kb/pipeline-runs")
	runDir := ""
	if runID != "" {
		runDir = filepath.Join(runRoot, runID)
	} else {
		entries, _ := os.ReadDir(runRoot)
		names := []string{}
		for _, entry := range entries {
			if entry.IsDir() {
				names = append(names, entry.Name())
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(names)))
		if len(names) > 0 {
			runDir = filepath.Join(runRoot, names[0])
		}
	}
	if runDir == "" {
		fmt.Fprintln(stderr, "No pipeline run found.")
		return 1
	}
	var run map[string]any
	if err := readJSONFile(filepath.Join(runDir, "run.json"), &run); err != nil {
		fmt.Fprintf(stderr, "Pipeline run is missing run.json: %s\n", runDir)
		return 1
	}
	fmt.Fprintf(stdout, "KB pipeline status: %s\n", stringValue(run["run_id"]))
	fmt.Fprintf(stdout, "Pipeline: %s\n", stringValue(run["pipeline_id"]))
	fmt.Fprintf(stdout, "Status: %s\n", stringValue(run["status"]))
	fmt.Fprintf(stdout, "Current phase: %s\n", stringValue(run["current_phase"]))
	fmt.Fprintf(stdout, "Run directory: %s\n", stringValue(run["run_dir"]))
	return 0
}

func runPipelineSelftest(root string, stdout, stderr io.Writer) int {
	var out strings.Builder
	if code := startPipeline(root, "skill-bundle-proof-spike", &out, stderr); code != 0 {
		return code
	}
	match := regexp.MustCompile(`KB pipeline started: ([^\r\n]+)`).FindStringSubmatch(out.String())
	if len(match) < 2 {
		fmt.Fprintln(stderr, "Pipeline start output did not include a run id.")
		return 1
	}
	runID := strings.TrimSpace(match[1])
	runDir := resolveRepoPath(root, ".kb/pipeline-runs/"+runID)
	defer os.RemoveAll(runDir)
	for _, required := range []string{"run.json", "pipeline.json", "selected-pipeline.md", "protected-files.json", "proof.json", "phase-prompts/map.md"} {
		if _, err := os.Stat(filepath.Join(runDir, filepath.FromSlash(required))); err != nil {
			fmt.Fprintf(stderr, "Pipeline run missing required artifact: %s\n", required)
			return 1
		}
	}
	if code := pipelineStatus(root, runID, io.Discard, stderr); code != 0 {
		return code
	}
	if code := startPipeline(root, "does-not-exist", io.Discard, io.Discard); code == 0 {
		fmt.Fprintln(stderr, "Pipeline accepted an unknown id.")
		return 1
	}
	fmt.Fprintln(stdout, "KB pipeline selftest: start/status passed; unknown pipeline id correctly rejected.")
	return 0
}

func writeJSONFile(path string, value any) {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	content, _ := json.MarshalIndent(value, "", "  ")
	_ = os.WriteFile(path, content, 0o644)
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func randomShortHash() string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(sum[:])[:8]
}

func slug(value string) string {
	value = strings.ToLower(value)
	value = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "pipeline"
	}
	return value
}

func tokenEstimate(text string) int {
	return len(regexp.MustCompile(`\s+`).Split(strings.TrimSpace(text), -1))
}

func hasJSONField(obj map[string]any, field string) bool {
	if obj == nil {
		return false
	}
	value, ok := obj[field]
	return ok && value != nil && fmt.Sprintf("%v", value) != ""
}

func missingJSONField(obj map[string]any, field string) bool {
	return !hasJSONField(obj, field)
}

func arrayValue(value any) []any {
	switch v := value.(type) {
	case []any:
		return v
	default:
		return []any{}
	}
}

func stringArray(value any) []string {
	out := []string{}
	for _, item := range arrayValue(value) {
		out = append(out, stringValue(item))
	}
	return out
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func intValue(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	default:
		var out int
		_, _ = fmt.Sscanf(stringValue(value), "%d", &out)
		return out
	}
}

func floatValue(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		var out float64
		_, _ = fmt.Sscanf(stringValue(value), "%f", &out)
		return out
	}
}

func boolValue(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	default:
		return strings.EqualFold(stringValue(value), "true")
	}
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func runGit(root string, args ...string) bool {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	return cmd.Run() == nil
}
