package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	out := map[string]any{"evidence_class": "imported", "eligible": 0, "excluded": 0, "conditions": map[string]int{}, "matched_groups": []map[string]any{}, "incomplete_groups": []string{}, "issues": []string{}}
	conditions := out["conditions"].(map[string]int)
	groups := map[string]map[string]map[string]any{}
	for _, file := range files {
		var row map[string]any
		if err := readJSONFile(file, &row); err != nil {
			out["excluded"] = out["excluded"].(int) + 1
			continue
		}
		condition := stringValue(row["condition"])
		proof, _ := row["independent_proof"].(map[string]any)
		keyParts := []string{stringValue(row["case"]), stringValue(row["repetition"]), stringValue(row["host"]), stringValue(row["config_fingerprint"]), stringValue(row["project_hash"]), stringValue(row["task_prompt_hash"])}
		missingKey := false
		for _, part := range keyParts {
			if part == "" {
				missingKey = true
			}
		}
		if row["evidence_kind"] != "live" || stringValue(row["task_success"]) == "" || proof == nil || stringValue(proof["command"]) == "" || intValue(proof["exit_code"]) != 0 || condition == "" || missingKey {
			out["excluded"] = out["excluded"].(int) + 1
			continue
		}
		out["eligible"] = out["eligible"].(int) + 1
		conditions[condition]++
		key := strings.Join(keyParts, "|")
		if groups[key] == nil {
			groups[key] = map[string]map[string]any{}
		}
		if _, duplicate := groups[key][condition]; duplicate {
			out["excluded"] = out["excluded"].(int) + 1
			out["eligible"] = out["eligible"].(int) - 1
			continue
		}
		groups[key][condition] = row
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		arms := groups[key]
		if len(arms) != 3 || arms["full"] == nil || arms["reduced"] == nil || arms["none"] == nil {
			out["incomplete_groups"] = append(out["incomplete_groups"].([]string), key)
			continue
		}
		out["matched_groups"] = append(out["matched_groups"].([]map[string]any), map[string]any{"group": key, "full_success": stringValue(arms["full"]["task_success"]), "reduced_success": stringValue(arms["reduced"]["task_success"]), "none_success": stringValue(arms["none"]["task_success"])})
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
