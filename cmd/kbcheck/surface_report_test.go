package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestSurfaceReportComparisonShowsAddedAndRemovedRoutes(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-start", "SKILL.md"), "# start\n")
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-map", "SKILL.md"), "# map\n")
	baselinePath := filepath.Join(root, "baseline.json")
	writeFile(t, baselinePath, `{
  "routes": [
    {"route":"base","total_lines":1,"token_estimate":1,"combined_hash":"old"},
    {"route":"removed-route","total_lines":4,"token_estimate":10,"combined_hash":"gone"}
  ]
}`)
	report, err := computeSurfaceReport(root, ".github/skills", "", baselinePath)
	if err != nil {
		t.Fatal(err)
	}
	content, _ := json.Marshal(report.Comparison)
	text := string(content)
	if !strings.Contains(text, `"status":"added"`) || !strings.Contains(text, `"status":"removed"`) {
		t.Fatalf("comparison did not expose route changes: %s", text)
	}
}
