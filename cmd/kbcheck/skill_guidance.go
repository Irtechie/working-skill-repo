package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type skillGuidanceConfig struct {
	MaxSkillLines             int      `json:"max_skill_lines"`
	MaxHotPathLines           int      `json:"max_hot_path_lines"`
	HotPath                   []string `json:"hot_path"`
	DeprecatedSkills          []string `json:"deprecated_skills"`
	AuditPath                 string   `json:"audit_path"`
	RemovedSkillInventoryPath string   `json:"removed_skill_inventory_path"`
}

type skillGuidanceAuditRow struct {
	Name               string   `json:"name"`
	Purpose            string   `json:"purpose"`
	Callers            []string `json:"callers"`
	RoutingEvidence    string   `json:"routing_evidence"`
	RetainedCapability string   `json:"retained_capability"`
	Disposition        string   `json:"disposition"`
	Proof              string   `json:"proof"`
}

var skillReferencePattern = regexp.MustCompile(`references/[A-Za-z0-9_.-]+(?:/[A-Za-z0-9_.-]+)*`)

func runSkillGuidanceCommand(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-guidance.json"
	}
	config, err := loadSkillGuidanceConfig(resolveRepoPath(root, configPath))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	failures, err := validateSkillGuidance(root, config)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if len(failures) > 0 {
		for _, failure := range failures {
			fmt.Fprintln(stderr, failure)
		}
		return 1
	}
	fmt.Fprintln(stdout, "skill guidance: ok")
	return 0
}

func loadSkillGuidanceConfig(path string) (skillGuidanceConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return skillGuidanceConfig{}, fmt.Errorf("read skill guidance config: %w", err)
	}
	var config skillGuidanceConfig
	if err := json.Unmarshal(content, &config); err != nil {
		return skillGuidanceConfig{}, fmt.Errorf("parse skill guidance config: %w", err)
	}
	if config.MaxSkillLines <= 0 || config.MaxHotPathLines <= 0 || config.AuditPath == "" || config.RemovedSkillInventoryPath == "" {
		return skillGuidanceConfig{}, fmt.Errorf("invalid skill guidance config")
	}
	return config, nil
}

func validateSkillGuidance(root string, config skillGuidanceConfig) ([]string, error) {
	skillRoot := filepath.Join(root, ".github", "skills")
	entries, err := os.ReadDir(skillRoot)
	if err != nil {
		return nil, fmt.Errorf("read skill root: %w", err)
	}

	auditRows, err := loadSkillGuidanceAudit(resolveRepoPath(root, config.AuditPath))
	if err != nil {
		return nil, err
	}
	removed, err := loadStringList(resolveRepoPath(root, config.RemovedSkillInventoryPath))
	if err != nil {
		return nil, err
	}
	auditByName := map[string][]skillGuidanceAuditRow{}
	for _, row := range auditRows {
		auditByName[row.Name] = append(auditByName[row.Name], row)
	}
	hotPath := setOf(config.HotPath...)
	deprecated := setOf(config.DeprecatedSkills...)
	removedSet := setOf(removed...)
	skillNames := map[string]bool{}
	for _, entry := range entries {
		if entry.IsDir() {
			skillNames[entry.Name()] = true
		}
	}
	seenSkills := map[string]bool{}
	failures := []string{}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		skillPath := filepath.Join(skillRoot, name, "SKILL.md")
		content, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}
		seenSkills[name] = true
		lines := lineCount(string(content))
		if lines > config.MaxSkillLines {
			failures = append(failures, fmt.Sprintf("skill-too-long: %s lines=%d max=%d", name, lines, config.MaxSkillLines))
		}
		if hotPath[name] && lines > config.MaxHotPathLines {
			failures = append(failures, fmt.Sprintf("hot-path-too-long: %s lines=%d max=%d", name, lines, config.MaxHotPathLines))
		}
		if deprecated[name] {
			failures = append(failures, "deprecated-skill-present: "+name)
			if !removedSet[name] {
				failures = append(failures, "removed-inventory-missing: "+name)
			}
		}
		rows := auditByName[name]
		if len(rows) == 0 {
			failures = append(failures, "audit-row-missing: "+name)
		} else if len(rows) > 1 {
			failures = append(failures, "audit-row-duplicate: "+name)
		} else if reason := invalidAuditRow(rows[0]); reason != "" {
			failures = append(failures, "audit-row-invalid: "+name+" "+reason)
		}
		if len(rows) == 1 {
			failures = append(failures, impossibleSkillCallers(name, string(content), rows[0], skillNames)...)
		}
		failures = append(failures, validateSkillReferences(name, filepath.Join(skillRoot, name), string(content))...)
	}
	for name := range auditByName {
		if !seenSkills[name] && !deprecated[name] {
			failures = append(failures, "audit-row-stale: "+name)
		}
	}
	for _, name := range config.DeprecatedSkills {
		if !removedSet[name] {
			failures = append(failures, "removed-inventory-missing: "+name)
		}
	}
	sort.Strings(failures)
	return uniqueStrings(failures), nil
}

// impossibleSkillCallers reports audit rows that claim a skill caller for a
// skill the model can never invoke. Such a claim cannot be stale-but-true; it
// is unsatisfiable by construction.
func impossibleSkillCallers(name, content string, row skillGuidanceAuditRow, skillNames map[string]bool) []string {
	frontmatter := extractFrontmatter(content)
	if strings.TrimSpace(frontmatterValue(frontmatter, "disable-model-invocation")) != "true" {
		return nil
	}
	failures := []string{}
	for _, caller := range row.Callers {
		caller = strings.TrimSpace(caller)
		if caller == name || !skillNames[caller] {
			continue
		}
		failures = append(failures, fmt.Sprintf("audit-caller-impossible: %s caller=%s disable-model-invocation=true", name, caller))
	}
	return failures
}

func loadSkillGuidanceAudit(path string) ([]skillGuidanceAuditRow, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read skill guidance audit: %w", err)
	}
	var rows []skillGuidanceAuditRow
	if err := json.Unmarshal(content, &rows); err != nil {
		return nil, fmt.Errorf("parse skill guidance audit: %w", err)
	}
	return rows, nil
}

func loadStringList(path string) ([]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read string list: %w", err)
	}
	var values []string
	if err := json.Unmarshal(content, &values); err != nil {
		return nil, fmt.Errorf("parse string list: %w", err)
	}
	return values, nil
}

func invalidAuditRow(row skillGuidanceAuditRow) string {
	if row.Name == "" || row.Purpose == "" || row.RoutingEvidence == "" || row.RetainedCapability == "" || row.Disposition == "" || row.Proof == "" {
		return "required-field-empty"
	}
	if row.Disposition != "keep" && row.Disposition != "remove" {
		return "invalid-disposition"
	}
	return ""
}

func validateSkillReferences(name, skillDir, content string) []string {
	failures := []string{}
	for _, reference := range uniqueStrings(skillReferencePattern.FindAllString(content, -1)) {
		reference = strings.TrimRight(reference, ".,;:")
		relative := strings.TrimPrefix(reference, "references/")
		if strings.Contains(relative, "/") {
			failures = append(failures, "nested-reference: "+name+" "+reference)
		}
		path := filepath.Join(skillDir, filepath.FromSlash(reference))
		referenceContent, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, "reference-missing: "+name+" "+reference)
			continue
		}
		if strings.EqualFold(filepath.Ext(path), ".md") && lineCount(string(referenceContent)) > 100 {
			text := strings.ToLower(string(referenceContent))
			if !strings.Contains(text, "## contents") && !strings.Contains(text, "## table of contents") {
				failures = append(failures, "reference-missing-contents: "+name+" "+reference)
			}
		}
	}
	return failures
}

func lineCount(content string) int {
	if content == "" {
		return 0
	}
	return strings.Count(content, "\n") + 1
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}
