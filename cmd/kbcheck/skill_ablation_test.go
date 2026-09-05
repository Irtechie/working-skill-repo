package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillAblationExcludesSyntheticAndUnprovenRows(t *testing.T) {
	root := t.TempDir()
	records := filepath.Join(root, "records")
	if err := os.MkdirAll(records, 0o755); err != nil {
		t.Fatal(err)
	}
	writeJSONFile(filepath.Join(records, "synthetic.json"), map[string]any{"evidence_kind": "synthetic", "condition": "full", "task_success": "pass"})
	for _, condition := range []string{"full", "reduced", "none"} {
		writeJSONFile(filepath.Join(records, condition+".json"), map[string]any{"evidence_kind": "live", "condition": condition, "task_success": "pass", "case": "case-1", "repetition": "1", "host": "host-a", "config_fingerprint": "config", "project_hash": "project", "task_prompt_hash": "prompt", "independent_proof": map[string]any{"command": "go test ./...", "exit_code": 0}})
	}
	var stdout, stderr strings.Builder
	if code := runSkillAblationCommand(root, options{resultRoot: "records", output: "report.json", json: true}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var report map[string]any
	if err := readJSONFile(filepath.Join(root, "report.json"), &report); err != nil {
		t.Fatal(err)
	}
	if intValue(report["eligible"]) != 3 || intValue(report["excluded"]) != 1 || len(report["matched_groups"].([]any)) != 1 {
		t.Fatalf("report=%#v", report)
	}
}
