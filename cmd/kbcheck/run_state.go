package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const lowRouteConfidence = 0.55

type routeHistoryEntry struct {
	Route        string
	Confidence   float64
	StateChanged bool
	ProgressKey  string
	StateHash    string
	Line         int
	Progress     bool
}

type runStateIssue struct {
	Code    string `json:"code"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

type runStateResult struct {
	OK      bool            `json:"ok"`
	Entries int             `json:"entries"`
	Issues  []runStateIssue `json:"issues"`
}

func runRunStateCommand(root string, opts options, stdout, stderr io.Writer) int {
	path := resolveInputPath(root, opts.history)
	result, err := validateRouteHistory(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		_ = encoder.Encode(result)
	} else if result.OK {
		fmt.Fprintln(stdout, "run state: ok")
	} else {
		for _, issue := range result.Issues {
			if issue.Line > 0 {
				fmt.Fprintf(stderr, "%s line %d: %s\n", issue.Code, issue.Line, issue.Message)
			} else {
				fmt.Fprintf(stderr, "%s: %s\n", issue.Code, issue.Message)
			}
		}
	}
	if !result.OK {
		return 2
	}
	return 0
}

func runRunStateSelftest(stdout, stderr io.Writer) int {
	temp, err := os.MkdirTemp("", "kb-run-state-selftest-*")
	if err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(temp)

	write := func(name, body string) string {
		path := filepath.Join(temp, name)
		_ = os.WriteFile(path, []byte(strings.TrimLeft(body, "\n")), 0o644)
		return path
	}
	mustPass := func(name, body string) bool {
		result, err := validateRouteHistory(write(name+".jsonl", body))
		if err != nil || !result.OK {
			fmt.Fprintf(stderr, "%s failed: result=%#v err=%v\n", name, result, err)
			return false
		}
		return true
	}
	mustFail := func(name, body, code string) bool {
		result, err := validateRouteHistory(write(name+".jsonl", body))
		if err != nil || result.OK || !hasRunStateIssue(result.Issues, code) {
			fmt.Fprintf(stderr, "%s not rejected with %s: result=%#v err=%v\n", name, code, result, err)
			return false
		}
		return true
	}

	if !mustPass("valid", `
{"route":"kb-start","confidence":0.82,"state_changed":true}
{"route":"kb-plan","confidence":0.74,"progress_key":"manifest-created"}
`) {
		return 1
	}
	if !mustFail("oscillation", `
{"route":"kb-plan","confidence":0.80}
{"route":"kb-work","confidence":0.80}
{"route":"kb-plan","confidence":0.80}
{"route":"kb-work","confidence":0.80}
`, "route-oscillation") {
		return 1
	}
	if !mustFail("low-confidence", `
{"route":"kb-start","confidence":0.40}
{"route":"kb-brainstorm","confidence":0.42}
{"route":"kb-plan","confidence":0.44}
`, "low-confidence-no-progress") {
		return 1
	}
	if !mustFail("no-progress", `
{"route":"kb-start","confidence":0.80}
{"route":"kb-start","confidence":0.81}
{"route":"kb-start","confidence":0.82}
{"route":"kb-start","confidence":0.83}
`, "no-progress-loop") {
		return 1
	}
	if !mustFail("malformed", `
{"confidence":0.9}
`, "malformed-row") {
		return 1
	}

	fmt.Fprintln(stdout, "KB run-state selftest: oscillation, low-confidence, and no-progress guards passed.")
	return 0
}

func validateRouteHistory(path string) (runStateResult, error) {
	file, err := os.Open(path)
	if err != nil {
		return runStateResult{}, fmt.Errorf("read route history: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	entries := []routeHistoryEntry{}
	issues := []runStateIssue{}
	var previousHash string
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		entry, rowIssues := parseRouteHistoryLine(line, lineNo, previousHash)
		if len(rowIssues) > 0 {
			issues = append(issues, rowIssues...)
			continue
		}
		if entry.StateHash != "" {
			previousHash = entry.StateHash
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return runStateResult{}, fmt.Errorf("scan route history: %w", err)
	}
	issues = append(issues, analyzeRouteHistory(entries)...)
	return runStateResult{OK: len(issues) == 0, Entries: len(entries), Issues: issues}, nil
}

func parseRouteHistoryLine(line string, lineNo int, previousHash string) (routeHistoryEntry, []runStateIssue) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		return routeHistoryEntry{}, []runStateIssue{{Code: "malformed-json", Line: lineNo, Message: err.Error()}}
	}
	entry := routeHistoryEntry{Line: lineNo}
	if value, ok := raw["route"]; ok {
		_ = json.Unmarshal(value, &entry.Route)
	}
	var confidence *float64
	if value, ok := raw["confidence"]; ok {
		var parsed float64
		if err := json.Unmarshal(value, &parsed); err == nil {
			confidence = &parsed
			entry.Confidence = parsed
		}
	}
	if value, ok := raw["state_changed"]; ok {
		_ = json.Unmarshal(value, &entry.StateChanged)
	}
	if value, ok := raw["progress_key"]; ok {
		_ = json.Unmarshal(value, &entry.ProgressKey)
	}
	if value, ok := raw["state_hash"]; ok {
		_ = json.Unmarshal(value, &entry.StateHash)
	}

	issues := []runStateIssue{}
	if strings.TrimSpace(entry.Route) == "" {
		issues = append(issues, runStateIssue{Code: "malformed-row", Line: lineNo, Message: "route is required"})
	}
	if confidence == nil {
		issues = append(issues, runStateIssue{Code: "malformed-row", Line: lineNo, Message: "confidence is required"})
	}
	entry.Progress = entry.StateChanged || entry.ProgressKey != "" || (entry.StateHash != "" && previousHash != "" && entry.StateHash != previousHash)
	return entry, issues
}

func analyzeRouteHistory(entries []routeHistoryEntry) []runStateIssue {
	issues := []runStateIssue{}
	lowNoProgressRun := 0
	noProgressRun := 0
	seenLowConfidence := false
	seenNoProgress := false
	seenOscillation := false
	for i, entry := range entries {
		if entry.Progress {
			lowNoProgressRun = 0
			noProgressRun = 0
		} else {
			noProgressRun++
			if entry.Confidence < lowRouteConfidence {
				lowNoProgressRun++
			} else {
				lowNoProgressRun = 0
			}
		}
		if lowNoProgressRun >= 3 && !seenLowConfidence {
			issues = append(issues, runStateIssue{Code: "low-confidence-no-progress", Line: entry.Line, Message: "three consecutive low-confidence routes made no progress; re-plan or ask for clarification"})
			seenLowConfidence = true
		}
		if noProgressRun >= 4 && !seenNoProgress {
			issues = append(issues, runStateIssue{Code: "no-progress-loop", Line: entry.Line, Message: "four consecutive route decisions made no progress"})
			seenNoProgress = true
		}
		if i >= 3 && !seenOscillation {
			a := entries[i-3].Route
			b := entries[i-2].Route
			c := entries[i-1].Route
			d := entry.Route
			if a == c && b == d && a != b {
				issues = append(issues, runStateIssue{Code: "route-oscillation", Line: entry.Line, Message: fmt.Sprintf("route history oscillates %s/%s/%s/%s", a, b, c, d)})
				seenOscillation = true
			}
		}
	}
	return issues
}

func hasRunStateIssue(issues []runStateIssue, code string) bool {
	for _, issue := range issues {
		if issue.Code == code {
			return true
		}
	}
	return false
}
