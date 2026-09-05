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
	writeJSONFile(filepath.Join(records, "live.json"), map[string]any{"evidence_kind": "live", "condition": "reduced", "task_success": "pass", "independent_proof": map[string]any{"command": "go test ./...", "exit_code": 0}})
	var stdout, stderr strings.Builder
	if code := runSkillAblationCommand(root, options{resultRoot: "records", output: "report.json", json: true}, &stdout, &stderr); code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr.String())
	}
	var report map[string]any
	if err := readJSONFile(filepath.Join(root, "report.json"), &report); err != nil {
		t.Fatal(err)
	}
	if intValue(report["eligible"]) != 1 || intValue(report["excluded"]) != 1 {
		t.Fatalf("report=%#v", report)
	}
}
