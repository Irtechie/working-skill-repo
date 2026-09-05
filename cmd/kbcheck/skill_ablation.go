package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Ablation records are imported evidence. This reducer never executes a
// referenced command and never turns unavailable evidence into a success.
func runSkillAblationCommand(root string, opts options, stdout, stderr io.Writer) int {
	if opts.resultRoot == "" || opts.output == "" {
		fmt.Fprintln(stderr, "skill-eval-ablation requires --result-root and --output")
		return 2
	}
	files, err := evalFiles(root, opts.resultRoot, "")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	out := map[string]any{"evidence_class": "imported", "eligible": 0, "excluded": 0, "conditions": map[string]int{}, "issues": []string{}}
	conditions := out["conditions"].(map[string]int)
	for _, file := range files {
		var row map[string]any
		if err := readJSONFile(file, &row); err != nil {
			out["excluded"] = out["excluded"].(int) + 1
			continue
		}
		condition := stringValue(row["condition"])
		proof, _ := row["independent_proof"].(map[string]any)
		if row["evidence_kind"] != "live" || stringValue(row["task_success"]) == "" || proof == nil || stringValue(proof["command"]) == "" || intValue(proof["exit_code"]) != 0 || condition == "" {
			out["excluded"] = out["excluded"].(int) + 1
			continue
		}
		out["eligible"] = out["eligible"].(int) + 1
		conditions[condition]++
	}
	if out["eligible"].(int) == 0 {
		out["issues"] = []string{"no independently evidenced live task outcomes admitted"}
	}
	path := resolveRepoPath(root, opts.output)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	writeJSONFile(path, out)
	if opts.json {
		writeJSON(stdout, out)
	} else {
		fmt.Fprintf(stdout, "Skill ablation: eligible=%d excluded=%d\n", out["eligible"], out["excluded"])
	}
	return 0
}
