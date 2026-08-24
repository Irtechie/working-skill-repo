package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type epistemicOracle struct {
	SchemaVersion     int                       `json:"schema_version"`
	FixtureID         string                    `json:"fixture_id"`
	Decision          string                    `json:"decision"`
	FinalState        string                    `json:"final_state"`
	ResolvingEvidence []string                  `json:"resolving_evidence"`
	Scored            bool                      `json:"scored"`
	ProcessOnly       bool                      `json:"process_only,omitempty"`
	Provenance        epistemicOracleProvenance `json:"provenance"`
}

type epistemicOracleProvenance struct {
	DecisionRule string `json:"decision_rule"`
	Rule         string `json:"rule"`
	Path         string `json:"path"`
	Pointer      string `json:"pointer"`
	Value        any    `json:"value,omitempty"`
}

type epistemicDetectionMetrics struct {
	CorrectProceed           int `json:"correct_proceed"`
	CorrectInvestigation     int `json:"correct_investigation"`
	MissedInvestigation      int `json:"missed_investigation"`
	UnnecessaryInvestigation int `json:"unnecessary_investigation"`
}

type epistemicOutcomeMetric struct {
	Eligible int `json:"eligible"`
	Correct  int `json:"correct"`
}

type epistemicCostMetrics struct {
	UserQuestions    int  `json:"user_questions"`
	ObservedCommands int  `json:"observed_commands"`
	ToolCalls        *int `json:"tool_calls,omitempty"`
	Turns            *int `json:"turns,omitempty"`
	ElapsedMS        *int `json:"elapsed_ms,omitempty"`
	InputTokens      *int `json:"input_tokens,omitempty"`
	OutputTokens     *int `json:"output_tokens,omitempty"`
	Tokens           *int `json:"tokens,omitempty"`
}

type epistemicMetrics struct {
	Detection       epistemicDetectionMetrics `json:"detection"`
	Resolution      epistemicOutcomeMetric    `json:"resolution"`
	Revision        epistemicOutcomeMetric    `json:"revision"`
	Cost            epistemicCostMetrics      `json:"cost"`
	VisibleCeremony bool                      `json:"visible_ceremony"`
	ProcessOnly     bool                      `json:"process_only,omitempty"`
}

type instructionSurface struct {
	Scope     string `json:"scope"`
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	LoadState string `json:"load_state"`
}

func loadEpistemicOracle(root, fixtureID string) (epistemicOracle, error) {
	var oracle epistemicOracle
	path := filepath.Join(root, "evals", "skill-eval", "epistemic", "oracles", fixtureID+".json")
	if err := readJSONFile(path, &oracle); err != nil {
		return oracle, err
	}
	if oracle.FixtureID != fixtureID {
		return oracle, fmt.Errorf("epistemic oracle fixture_id %q does not match %q", oracle.FixtureID, fixtureID)
	}
	return oracle, nil
}

func scoreEpistemicResult(result map[string]any, oracle epistemicOracle) (epistemicMetrics, []evalIssue) {
	metrics := epistemicMetrics{ProcessOnly: oracle.ProcessOnly}
	id := stringValue(result["id"])
	if id == "" {
		id = oracle.FixtureID
	}
	e, _ := result["epistemic"].(map[string]any)
	decision := stringValue(e["decision"])
	switch {
	case oracle.Decision == "proceed" && decision == "proceed":
		metrics.Detection.CorrectProceed = 1
	case oracle.Decision == "proceed" && decision == "investigate":
		metrics.Detection.UnnecessaryInvestigation = 1
	case oracle.Decision == "investigate" && decision == "proceed":
		metrics.Detection.MissedInvestigation = 1
	case oracle.Decision == "investigate" && decision == "investigate":
		metrics.Detection.CorrectInvestigation = 1
	}
	issues := []evalIssue{}
	if strings.TrimSpace(stringValue(result["response_text"])) == "" {
		issues = append(issues, evalIssue{Result: id, Message: "Epistemic result is missing the actual user-facing response_text."})
	}
	if !contains([]string{"proceed", "investigate"}, decision) {
		issues = append(issues, evalIssue{Result: id, Message: "Epistemic decision must be 'proceed' or 'investigate'."})
	} else if metrics.Detection.MissedInvestigation+metrics.Detection.UnnecessaryInvestigation > 0 {
		issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Expected epistemic decision '%s' but got '%s'.", oracle.Decision, decision)})
	}

	if oracle.Decision == "investigate" && decision == "investigate" {
		metrics.Resolution.Eligible = 1
		inspected := stringArray(e["evidence_inspected"])
		if containsAllNormalized(inspected, oracle.ResolvingEvidence) {
			metrics.Resolution.Correct = 1
		} else {
			issues = append(issues, evalIssue{Result: id, Message: "Investigation did not inspect evidence capable of resolving the premise."})
		}
		if !oracle.ProcessOnly && oracle.Scored {
			metrics.Revision.Eligible = 1
			if metrics.Resolution.Correct == 1 && stringValue(e["final_state"]) == oracle.FinalState {
				metrics.Revision.Correct = 1
			} else {
				issues = append(issues, evalIssue{Result: id, Message: fmt.Sprintf("Expected final epistemic state '%s' after resolving evidence.", oracle.FinalState)})
			}
		}
	}

	response := strings.ToLower(stringValue(result["response_text"]))
	if oracle.Decision == "proceed" && decision == "proceed" && visibleInvestigationCeremony(response) {
		metrics.VisibleCeremony = true
		issues = append(issues, evalIssue{Result: id, Message: "Supported proceed response exposed unnecessary verification ceremony."})
	}
	actual, _ := result["actual"].(map[string]any)
	metrics.Cost.UserQuestions = intValue(actual["user_questions"])
	trace, _ := result["trace"].(map[string]any)
	metrics.Cost.ObservedCommands = len(stringArray(trace["commands"]))
	telemetry, _ := result["telemetry"].(map[string]any)
	metrics.Cost.ToolCalls = optionalInt(telemetry, "tool_calls")
	metrics.Cost.Turns = optionalInt(telemetry, "turns")
	metrics.Cost.ElapsedMS = optionalInt(telemetry, "elapsed_ms")
	metrics.Cost.InputTokens = optionalInt(telemetry, "input_tokens")
	metrics.Cost.OutputTokens = optionalInt(telemetry, "output_tokens")
	if metrics.Cost.InputTokens != nil && metrics.Cost.OutputTokens != nil {
		total := *metrics.Cost.InputTokens + *metrics.Cost.OutputTokens
		metrics.Cost.Tokens = &total
	}
	return metrics, issues
}

func optionalInt(obj map[string]any, key string) *int {
	if obj == nil || !hasJSONField(obj, key) {
		return nil
	}
	v := intValue(obj[key])
	return &v
}

func visibleInvestigationCeremony(text string) bool {
	for _, marker := range []string{"let me investigate", "before i continue", "verification checklist", "checklist:", "i need to verify", "i'll research"} {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func containsAllNormalized(actual, expected []string) bool {
	set := map[string]bool{}
	for _, value := range actual {
		set[filepath.ToSlash(filepath.Clean(value))] = true
	}
	for _, value := range expected {
		if !set[filepath.ToSlash(filepath.Clean(value))] {
			return false
		}
	}
	return true
}

func validateEpistemicCorpus(root string) []evalIssue {
	oracleRoot := filepath.Join(root, "evals", "skill-eval", "epistemic", "oracles")
	files, _ := filepath.Glob(filepath.Join(oracleRoot, "*.json"))
	sort.Strings(files)
	issues := []evalIssue{}
	for _, path := range files {
		var oracle epistemicOracle
		if err := readJSONFile(path, &oracle); err != nil {
			issues = append(issues, evalIssue{Result: filepath.Base(path), Message: err.Error()})
			continue
		}
		if oracle.ProcessOnly {
			continue
		}
		if !oracle.Scored || !contains([]string{"json-pointer-equals", "json-pointer-absent"}, oracle.Provenance.Rule) {
			issues = append(issues, evalIssue{Result: oracle.FixtureID, Message: "Correctness oracle lacks deterministic machine-checkable provenance."})
			continue
		}
		if err := verifyEpistemicProvenance(root, oracle); err != nil {
			issues = append(issues, evalIssue{Result: oracle.FixtureID, Message: err.Error()})
		}
	}
	return issues
}

func verifyEpistemicProvenance(root string, oracle epistemicOracle) error {
	fixturePath := filepath.Join(root, "evals", "skill-eval", "epistemic", "visible", oracle.FixtureID, "fixture.json")
	var fixture map[string]any
	if err := readJSONFile(fixturePath, &fixture); err != nil {
		return fmt.Errorf("visible fixture unavailable for decision provenance")
	}
	task := strings.ToLower(stringValue(fixture["task"]))
	switch oracle.Provenance.DecisionRule {
	case "task-cites-source-value":
		if oracle.Decision != "proceed" || !strings.Contains(task, strings.ToLower(oracle.Provenance.Path)) || !strings.Contains(task, "has feature.enabled=") {
			return fmt.Errorf("proceed label disagrees with task source-citation construction rule")
		}
	case "task-unsupported-assumption":
		if oracle.Decision != "investigate" || !strings.HasPrefix(strings.TrimSpace(task), "assume ") {
			return fmt.Errorf("investigate label disagrees with unsupported-assumption construction rule")
		}
	default:
		return fmt.Errorf("unsupported oracle decision rule %q", oracle.Provenance.DecisionRule)
	}
	path := filepath.Join(root, "evals", "skill-eval", "epistemic", "visible", oracle.FixtureID, filepath.FromSlash(oracle.Provenance.Path))
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("oracle provenance source unavailable: %s", oracle.Provenance.Path)
	}
	var value any
	if err := json.Unmarshal(content, &value); err != nil {
		return fmt.Errorf("oracle provenance source is not JSON: %s", oracle.Provenance.Path)
	}
	got, found := jsonPointer(value, oracle.Provenance.Pointer)
	switch oracle.Provenance.Rule {
	case "json-pointer-absent":
		if found {
			return fmt.Errorf("oracle provenance expected absent pointer %s", oracle.Provenance.Pointer)
		}
		if oracle.FinalState != "no-justified-conclusion" {
			return fmt.Errorf("final-state label disagrees with absent source truth")
		}
	case "json-pointer-equals":
		if !found || fmt.Sprint(got) != fmt.Sprint(oracle.Provenance.Value) {
			return fmt.Errorf("oracle provenance does not match source truth at %s", oracle.Provenance.Pointer)
		}
		wantState := "contradicted"
		if boolValue(oracle.Provenance.Value) {
			wantState = "supported"
		}
		if oracle.FinalState != wantState {
			return fmt.Errorf("final-state label disagrees with machine-checkable source truth")
		}
	default:
		return fmt.Errorf("unsupported oracle provenance rule %q", oracle.Provenance.Rule)
	}
	return nil
}

func jsonPointer(value any, pointer string) (any, bool) {
	if pointer == "" {
		return value, true
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, false
	}
	current := value
	for _, token := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[token]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func materializeEpistemicActorWorkspace(root, fixtureID, workspace string, surfaces []instructionSurface) error {
	for _, surface := range surfaces {
		if !contains([]string{"proven", "isolated"}, surface.LoadState) || len(surface.SHA256) != 64 {
			return fmt.Errorf("instruction surface %s:%s has unprovable load state", surface.Scope, surface.Path)
		}
	}
	source := filepath.Join(root, "evals", "skill-eval", "epistemic", "visible", fixtureID)
	if _, err := os.Stat(source); err != nil {
		return fmt.Errorf("visible epistemic fixture unavailable: %s", fixtureID)
	}
	return filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		dest := filepath.Join(workspace, rel)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, content, 0o644)
	})
}

func aggregateEpistemicRegression(root string, rows []map[string]any) map[string]any {
	totals := map[string]int{
		"correct_proceed": 0, "correct_investigation": 0, "missed_investigation": 0, "unnecessary_investigation": 0,
		"resolution_eligible": 0, "resolution_correct": 0, "revision_eligible": 0, "revision_correct": 0,
		"user_questions": 0, "observed_commands": 0, "visible_ceremony": 0,
	}
	count := 0
	for _, row := range rows {
		path := stringValue(row["result_path"])
		if path == "" {
			continue
		}
		var result map[string]any
		if err := readJSONFile(path, &result); err != nil {
			continue
		}
		if _, ok := result["epistemic"]; !ok {
			continue
		}
		oracle, err := loadEpistemicOracle(root, stringValue(result["fixture_id"]))
		if err != nil {
			continue
		}
		metrics, _ := scoreEpistemicResult(result, oracle)
		count++
		totals["correct_proceed"] += metrics.Detection.CorrectProceed
		totals["correct_investigation"] += metrics.Detection.CorrectInvestigation
		totals["missed_investigation"] += metrics.Detection.MissedInvestigation
		totals["unnecessary_investigation"] += metrics.Detection.UnnecessaryInvestigation
		totals["resolution_eligible"] += metrics.Resolution.Eligible
		totals["resolution_correct"] += metrics.Resolution.Correct
		totals["revision_eligible"] += metrics.Revision.Eligible
		totals["revision_correct"] += metrics.Revision.Correct
		totals["user_questions"] += metrics.Cost.UserQuestions
		totals["observed_commands"] += metrics.Cost.ObservedCommands
		if metrics.VisibleCeremony {
			totals["visible_ceremony"]++
		}
	}
	if count == 0 {
		return nil
	}
	out := map[string]any{"result_count": count}
	for key, value := range totals {
		out[key] = value
	}
	return out
}

func compareEpistemicRegression(rawBaseline any, current map[string]any) string {
	baseline, ok := rawBaseline.(map[string]any)
	if !ok || !hasEpistemicComparisonEvidence(baseline) || !hasEpistemicComparisonEvidence(current) || intValue(baseline["result_count"]) == 0 || intValue(current["result_count"]) == 0 {
		return "inconclusive"
	}
	worse := intValue(current["missed_investigation"]) > intValue(baseline["missed_investigation"]) ||
		intValue(current["unnecessary_investigation"]) > intValue(baseline["unnecessary_investigation"]) ||
		intValue(current["resolution_correct"]) < intValue(baseline["resolution_correct"]) ||
		intValue(current["revision_correct"]) < intValue(baseline["revision_correct"]) ||
		intValue(current["user_questions"]) > intValue(baseline["user_questions"]) ||
		intValue(current["visible_ceremony"]) > intValue(baseline["visible_ceremony"])
	if worse {
		return "reject"
	}
	better := intValue(current["missed_investigation"]) < intValue(baseline["missed_investigation"]) ||
		intValue(current["resolution_correct"]) > intValue(baseline["resolution_correct"]) ||
		intValue(current["revision_correct"]) > intValue(baseline["revision_correct"])
	if better {
		return "promote"
	}
	if intValue(baseline["missed_investigation"]) == 0 && intValue(current["missed_investigation"]) == 0 {
		return "promote"
	}
	return "inconclusive"
}

func hasEpistemicComparisonEvidence(metrics map[string]any) bool {
	for _, key := range []string{"result_count", "missed_investigation", "unnecessary_investigation", "resolution_correct", "revision_correct", "user_questions", "visible_ceremony"} {
		if !hasJSONField(metrics, key) {
			return false
		}
	}
	return true
}
