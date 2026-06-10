package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type skillQualityConfig struct {
	Lint                 lintConfig                 `json:"lint"`
	ReviewReferenceGuard reviewReferenceGuardConfig `json:"review_reference_guard"`
	SyncTargets          []syncTarget               `json:"sync_targets"`
}

type lintConfig struct {
	SkillRoot           string            `json:"skill_root"`
	AgentRoot           string            `json:"agent_root"`
	RequireArgumentHint string            `json:"require_argument_hint"`
	RequiredFrontmatter []string          `json:"required_frontmatter"`
	ScanExtensions      []string          `json:"scan_extensions"`
	HotPathWarningLines int               `json:"hot_path_warning_lines"`
	HotPathFailLines    int               `json:"hot_path_fail_lines"`
	AllowLongSkills     map[string]string `json:"allow_long_skills"`
}

type syncTarget struct {
	ID             string `json:"id"`
	Label          string `json:"label"`
	Path           string `json:"path"`
	Classification string `json:"classification"`
	Required       bool   `json:"required"`
}

type lintIssue struct {
	Severity string `json:"severity"`
	Path     string `json:"path"`
	Message  string `json:"message"`
}

type skillLintResult struct {
	OK       bool        `json:"ok"`
	Errors   []lintIssue `json:"errors"`
	Warnings []lintIssue `json:"warnings"`
	Summary  struct {
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	} `json:"summary"`
}

func runSkillLintCommand(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-quality.json"
	}
	result, err := computeSkillLint(root, configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "Skill lint: %d errors, %d warnings\n", len(result.Errors), len(result.Warnings))
		for _, issue := range result.Errors {
			fmt.Fprintf(stdout, "ERROR [%s] %s\n", issue.Path, issue.Message)
		}
		for _, warning := range result.Warnings {
			fmt.Fprintf(stdout, "WARN  [%s] %s\n", warning.Path, warning.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

func computeSkillLint(root, configPath string) (skillLintResult, error) {
	var result skillLintResult
	config, err := loadSkillQualityConfig(root, configPath)
	if err != nil {
		return result, err
	}
	skillRoot := resolveRepoPath(root, config.Lint.SkillRoot)
	skillNames := []string{}
	if info, err := os.Stat(skillRoot); err != nil || !info.IsDir() {
		result.Errors = append(result.Errors, lintIssue{"error", config.Lint.SkillRoot, "Skill root is missing."})
	} else {
		entries, _ := os.ReadDir(skillRoot)
		sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			skillName := entry.Name()
			skillNames = append(skillNames, skillName)
			skillFile := filepath.Join(skillRoot, skillName, "SKILL.md")
			relative := relativePath(root, skillFile)
			content, err := os.ReadFile(skillFile)
			if err != nil {
				result.Errors = append(result.Errors, lintIssue{"error", relative, "Missing SKILL.md."})
				continue
			}
			text := string(content)
			lines := countLines(text)
			frontmatter := extractFrontmatter(text)
			if frontmatter == "" {
				result.Errors = append(result.Errors, lintIssue{"error", relative, "Missing YAML frontmatter."})
			}
			for _, field := range config.Lint.RequiredFrontmatter {
				if !frontmatterHasField(frontmatter, field) {
					result.Errors = append(result.Errors, lintIssue{"error", relative, fmt.Sprintf("Missing required frontmatter field '%s'.", field)})
				}
			}
			if declared := frontmatterValue(frontmatter, "name"); declared != "" && declared != skillName {
				result.Errors = append(result.Errors, lintIssue{"error", relative, fmt.Sprintf("Frontmatter name '%s' does not match folder '%s'.", declared, skillName)})
			}
			if !frontmatterHasField(frontmatter, "argument-hint") {
				switch config.Lint.RequireArgumentHint {
				case "error":
					result.Errors = append(result.Errors, lintIssue{"error", relative, "Missing argument-hint frontmatter."})
				case "warning":
					result.Warnings = append(result.Warnings, lintIssue{"warning", relative, "Missing argument-hint frontmatter."})
				}
			}
			for _, ref := range regexp.MustCompile(`@\./([^\s)]+)`).FindAllStringSubmatch(text, -1) {
				if _, err := os.Stat(filepath.Join(skillRoot, skillName, filepath.FromSlash(ref[1]))); err != nil {
					result.Errors = append(result.Errors, lintIssue{"error", relative, fmt.Sprintf("Broken lazy reference '@./%s'.", ref[1])})
				}
			}
			if config.Lint.HotPathFailLines > 0 && lines > config.Lint.HotPathFailLines {
				if reason, ok := config.Lint.AllowLongSkills[skillName]; ok {
					result.Warnings = append(result.Warnings, lintIssue{"warning", relative, fmt.Sprintf("Skill has %d lines, exceeding hard limit but allowlisted: %s", lines, reason)})
				} else {
					result.Errors = append(result.Errors, lintIssue{"error", relative, fmt.Sprintf("Skill has %d lines, exceeding hard limit %d.", lines, config.Lint.HotPathFailLines)})
				}
			} else if config.Lint.HotPathWarningLines > 0 && lines > config.Lint.HotPathWarningLines {
				result.Warnings = append(result.Warnings, lintIssue{"warning", relative, fmt.Sprintf("Skill has %d lines, exceeding warning limit %d.", lines, config.Lint.HotPathWarningLines)})
			}
		}
	}
	addUnknownSkillReferenceWarnings(root, config, skillNames, &result)
	addConflictMarkerErrors(root, config, &result)
	result.OK = len(result.Errors) == 0
	result.Summary.Errors = len(result.Errors)
	result.Summary.Warnings = len(result.Warnings)
	return result, nil
}

type skillSyncResult struct {
	OK             bool      `json:"ok"`
	RequiredIssues int       `json:"required_issues"`
	Rows           []syncRow `json:"rows"`
}

type syncRow struct {
	Skill          string `json:"skill"`
	Target         string `json:"target"`
	Classification string `json:"classification"`
	Required       bool   `json:"required"`
	Status         string `json:"status"`
	SourceHash     string `json:"source_hash"`
	TargetHash     string `json:"target_hash"`
	Suggestion     string `json:"suggestion"`
}

func runSkillSyncReportCommand(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-quality.json"
	}
	result, err := computeSkillSyncReport(root, configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		printSkillSyncReport(stdout, result, opts.verboseOptional)
	}
	if !result.OK {
		return 1
	}
	return 0
}

func computeSkillSyncReport(root, configPath string) (skillSyncResult, error) {
	var result skillSyncResult
	config, err := loadSkillQualityConfig(root, configPath)
	if err != nil {
		return result, err
	}
	var source *syncTarget
	for i := range config.SyncTargets {
		if config.SyncTargets[i].Classification == "source" {
			source = &config.SyncTargets[i]
			break
		}
	}
	if source == nil {
		return result, fmt.Errorf("no source sync target configured")
	}
	sourceRoot := resolveRepoPath(root, source.Path)
	entries, err := os.ReadDir(sourceRoot)
	if err != nil {
		return result, fmt.Errorf("source skill root not found: %s", source.Path)
	}
	skills := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			skills = append(skills, entry.Name())
		}
	}
	sort.Strings(skills)
	for _, skill := range skills {
		sourceHash, _ := skillHash(filepath.Join(sourceRoot, skill))
		for _, target := range config.SyncTargets {
			if target.Classification == "source" {
				continue
			}
			targetRoot := resolveRepoPath(root, target.Path)
			status := "missing-target"
			targetHash := ""
			suggestion := "target path unavailable"
			if info, err := os.Stat(targetRoot); err == nil && info.IsDir() {
				if hash, err := skillHash(filepath.Join(targetRoot, skill)); err == nil {
					targetHash = hash
					if hash == sourceHash {
						status = "match"
						suggestion = "none"
					} else if target.Required {
						status = "drift-required"
						suggestion = "review diff before copying source -> target or merging target -> source"
					} else {
						status = "drift-optional"
						suggestion = "review diff before copying source -> target or merging target -> source"
					}
				} else if target.Required {
					status = "missing-required"
					suggestion = "copy source -> target if this skill is meant to ship there"
				} else {
					status = "missing-optional"
					suggestion = "copy source -> target if this skill is meant to ship there"
				}
			}
			row := syncRow{
				Skill: skill, Target: target.ID, Classification: target.Classification,
				Required: target.Required, Status: status, SourceHash: shortHash(sourceHash),
				TargetHash: shortHash(targetHash), Suggestion: suggestion,
			}
			result.Rows = append(result.Rows, row)
			if target.Required && status != "match" {
				result.RequiredIssues++
			}
		}
	}
	result.OK = result.RequiredIssues == 0
	return result, nil
}

func printSkillSyncReport(w io.Writer, result skillSyncResult, verboseOptional bool) {
	fmt.Fprintf(w, "Skill sync report: %d comparisons, %d required issues\n", len(result.Rows), result.RequiredIssues)
	groupCounts := map[string]int{}
	for _, row := range result.Rows {
		groupCounts[row.Target+", "+row.Status]++
	}
	keys := sortedMapKeys(groupCounts)
	for _, key := range keys {
		fmt.Fprintf(w, "%s: %d\n", key, groupCounts[key])
	}
	required := []syncRow{}
	optional := []syncRow{}
	for _, row := range result.Rows {
		if row.Required && row.Status != "match" {
			required = append(required, row)
		}
		if !row.Required && row.Status != "match" {
			optional = append(optional, row)
		}
	}
	if len(required) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Required issues:")
		for _, issue := range required {
			fmt.Fprintf(w, "ERROR [%s] %s: %s source=%s target=%s :: %s\n", issue.Target, issue.Skill, issue.Status, issue.SourceHash, issue.TargetHash, issue.Suggestion)
		}
	}
	if len(optional) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "Optional target differences: %d warning-only differences. Use --verbose-optional for per-skill rows.\n", len(optional))
		groups := map[string]int{}
		for _, row := range optional {
			groups[row.Target+", "+row.Status]++
		}
		for _, key := range sortedMapKeys(groups) {
			fmt.Fprintf(w, "WARN  %s: %d\n", key, groups[key])
		}
		if verboseOptional {
			fmt.Fprintln(w)
			fmt.Fprintln(w, "Optional target details:")
			for _, issue := range optional {
				fmt.Fprintf(w, "WARN  [%s] %s: %s source=%s target=%s :: %s\n", issue.Target, issue.Skill, issue.Status, issue.SourceHash, issue.TargetHash, issue.Suggestion)
			}
		}
	}
}

type marketplaceConfig struct {
	Marketplace struct {
		LocalRoot   string `json:"local_root"`
		Directories struct {
			ApprovedSkills    string `json:"approved_skills"`
			ApprovedCatalog   string `json:"approved_catalog"`
			QuarantineCatalog string `json:"quarantine_catalog"`
			Quarantine        string `json:"quarantine"`
		} `json:"directories"`
	} `json:"marketplace"`
	ProjectLocalPaths struct {
		Skills string `json:"skills"`
	} `json:"project_local_paths"`
	QuarantineFirebreak struct {
		NeverLoadFromQuarantine    bool     `json:"never_load_from_quarantine"`
		AdditionalActiveSkillRoots []string `json:"additional_active_skill_roots"`
	} `json:"quarantine_firebreak"`
}

type marketplaceIssue struct {
	Kind    string `json:"kind"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

type marketplaceFirebreakResult struct {
	OK                bool               `json:"ok"`
	MarketplaceRoot   string             `json:"marketplace_root"`
	QuarantinePath    string             `json:"quarantine_path"`
	CheckedSkillRoots []string           `json:"checked_skill_roots"`
	IssueCount        int                `json:"issue_count"`
	Issues            []marketplaceIssue `json:"issues"`
}

func runMarketplaceFirebreakCommand(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-marketplace.json"
	}
	result, err := computeMarketplaceFirebreak(root, configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "Skill marketplace firebreak: issues=%d\n", result.IssueCount)
		fmt.Fprintf(stdout, "quarantine=%s\n", result.QuarantinePath)
		for _, issue := range result.Issues {
			fmt.Fprintf(stdout, "ERROR [%s] %s: %s\n", issue.Kind, issue.Path, issue.Message)
		}
	}
	if !result.OK {
		return 1
	}
	return 0
}

func computeMarketplaceFirebreak(root, configPath string) (marketplaceFirebreakResult, error) {
	var result marketplaceFirebreakResult
	configFullPath := resolveRepoPath(root, configPath)
	var config marketplaceConfig
	if err := readJSONFile(configFullPath, &config); err != nil {
		return result, fmt.Errorf("marketplace config not found: %s", configFullPath)
	}
	marketplaceRoot := resolveRepoPath(root, config.Marketplace.LocalRoot)
	quarantinePath := resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.Quarantine)
	approvedSkillsPath := resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.ApprovedSkills)
	approvedCatalogPath := resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.ApprovedCatalog)
	quarantineCatalogPath := resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.QuarantineCatalog)
	result.MarketplaceRoot = marketplaceRoot
	result.QuarantinePath = quarantinePath
	if !config.QuarantineFirebreak.NeverLoadFromQuarantine {
		result.Issues = append(result.Issues, marketplaceIssue{"missing-firebreak-policy", configFullPath, "Config must set quarantine_firebreak.never_load_from_quarantine=true."})
	}
	roots := knownSkillRoots(root, marketplaceRoot, config)
	result.CheckedSkillRoots = roots
	for _, rootPath := range roots {
		testSkillRoot(rootPath, quarantinePath, &result.Issues)
	}
	testApprovedCatalog(marketplaceRoot, approvedCatalogPath, approvedSkillsPath, quarantinePath, &result.Issues)
	testQuarantineCatalog(quarantineCatalogPath, &result.Issues)
	result.IssueCount = len(result.Issues)
	result.OK = result.IssueCount == 0
	return result, nil
}

func runMarketplaceFirebreakSelftest(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-marketplace.json"
	}
	valid, err := computeMarketplaceFirebreak(root, configPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !valid.OK {
		fmt.Fprintf(stderr, "expected valid marketplace firebreak config to pass, got %d issues\n", valid.IssueCount)
		return 1
	}
	tmpParent := filepath.Join(root, ".atv", "tmp")
	if err := os.MkdirAll(tmpParent, 0o755); err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	tempRoot, err := os.MkdirTemp(tmpParent, "skill-marketplace-firebreak-selftest-*")
	if err != nil {
		fmt.Fprintf(stderr, "create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tempRoot)
	var config marketplaceConfig
	if err := readJSONFile(resolveRepoPath(root, configPath), &config); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	config.QuarantineFirebreak.NeverLoadFromQuarantine = true
	config.QuarantineFirebreak.AdditionalActiveSkillRoots = []string{filepath.ToSlash(filepath.Join(config.Marketplace.LocalRoot, config.Marketplace.Directories.Quarantine))}
	badPath := filepath.Join(tempRoot, "bad-config.json")
	content, _ := json.MarshalIndent(config, "", "  ")
	if err := os.WriteFile(badPath, content, 0o644); err != nil {
		fmt.Fprintf(stderr, "write bad config: %v\n", err)
		return 1
	}
	bad, err := computeMarketplaceFirebreak(root, badPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if bad.OK {
		fmt.Fprintln(stderr, "expected marketplace firebreak to fail when an active skill root points at quarantine")
		return 1
	}
	fmt.Fprintln(stdout, "Skill marketplace firebreak selftest: valid config passed; quarantined active root correctly rejected.")
	return 0
}

func loadSkillQualityConfig(root, configPath string) (skillQualityConfig, error) {
	var config skillQualityConfig
	fullPath := resolveRepoPath(root, configPath)
	if err := readJSONFile(fullPath, &config); err != nil {
		return config, fmt.Errorf("missing skill quality config: %s", configPath)
	}
	return config, nil
}

func readJSONFile(path string, value any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(content, value)
}

func resolveRepoPath(root, path string) string {
	path = expandPathText(path)
	if filepath.IsAbs(path) {
		abs, _ := filepath.Abs(path)
		return abs
	}
	abs, _ := filepath.Abs(filepath.Join(root, filepath.FromSlash(path)))
	return abs
}

func resolveMarketplacePath(marketplaceRoot, path string) string {
	path = expandPathText(path)
	if filepath.IsAbs(path) {
		abs, _ := filepath.Abs(path)
		return abs
	}
	abs, _ := filepath.Abs(filepath.Join(marketplaceRoot, filepath.FromSlash(path)))
	return abs
}

func expandPathText(path string) string {
	expanded := os.ExpandEnv(path)
	if expanded == "~" || strings.HasPrefix(expanded, "~/") || strings.HasPrefix(expanded, `~\`) {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			if expanded == "~" {
				return home
			}
			return filepath.Join(home, filepath.FromSlash(strings.TrimLeft(expanded[1:], `/\`)))
		}
	}
	return expanded
}

func relativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func countLines(text string) int {
	if text == "" {
		return 0
	}
	return len(regexp.MustCompile(`\r?\n`).Split(text, -1))
}

func extractFrontmatter(text string) string {
	if !strings.HasPrefix(text, "---") {
		return ""
	}
	lines := regexp.MustCompile(`\r?\n`).Split(text, -1)
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return ""
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], "\n")
		}
	}
	return ""
}

func frontmatterHasField(frontmatter, field string) bool {
	if frontmatter == "" {
		return false
	}
	pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(field) + `\s*:`)
	return pattern.MatchString(frontmatter)
}

func frontmatterValue(frontmatter, field string) string {
	if frontmatter == "" {
		return ""
	}
	pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(field) + `\s*:\s*(.+?)\s*$`)
	match := pattern.FindStringSubmatch(frontmatter)
	if len(match) < 2 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(match[1]), `"'`)
}

func addUnknownSkillReferenceWarnings(root string, config skillQualityConfig, skillNames []string, result *skillLintResult) {
	if len(skillNames) == 0 {
		return
	}
	known := map[string]bool{"land": true, "todo-resolve": true}
	for _, skill := range skillNames {
		known[skill] = true
	}
	files := []string{}
	for _, rootPath := range []string{".github/skills", "evals"} {
		full := resolveRepoPath(root, rootPath)
		_ = filepath.WalkDir(full, func(path string, entry os.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".md" || ext == ".json" {
				files = append(files, path)
			}
			return nil
		})
	}
	for _, file := range []string{"AGENTS.md", "README.md", "config/skill-quality.json"} {
		full := resolveRepoPath(root, file)
		if _, err := os.Stat(full); err == nil {
			files = append(files, full)
		}
	}
	sort.Strings(files)
	token := `(?:kb|ce)-[a-z0-9-]+|todo-[a-z0-9-]+|document-review|learn|evolve|tdd|klfg`
	patterns := []*regexp.Regexp{
		regexp.MustCompile("`/?(" + token + ")`"),
		regexp.MustCompile(`(?:^|[^A-Za-z0-9_-])/(?:` + token + `)(?:[^A-Za-z0-9_-]|$)`),
	}
	seen := map[string]bool{}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		relative := relativePath(root, file)
		for i, line := range regexp.MustCompile(`\r?\n`).Split(string(content), -1) {
			for _, pattern := range patterns {
				for _, match := range pattern.FindAllStringSubmatch(line, -1) {
					ref := ""
					if len(match) > 1 && match[1] != "" {
						ref = match[1]
					} else {
						raw := match[0]
						idx := strings.Index(raw, "/")
						if idx >= 0 {
							raw = raw[idx+1:]
						}
						ref = regexp.MustCompile(token).FindString(raw)
					}
					if ref == "" || known[ref] {
						continue
					}
					key := fmt.Sprintf("%s:%d:%s", relative, i+1, ref)
					if !seen[key] {
						seen[key] = true
						result.Warnings = append(result.Warnings, lintIssue{"warning", relative, fmt.Sprintf("Unknown skill reference '%s' at line %d.", ref, i+1)})
					}
				}
			}
		}
	}
}

func addConflictMarkerErrors(root string, config skillQualityConfig, result *skillLintResult) {
	extensions := map[string]bool{}
	for _, ext := range config.Lint.ScanExtensions {
		extensions[strings.ToLower(ext)] = true
	}
	conflict := regexp.MustCompile(`^(<<<<<<<|>>>>>>>)(\s|$)|^=======$`)
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !extensions[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for i, line := range regexp.MustCompile(`\r?\n`).Split(string(content), -1) {
			if conflict.MatchString(line) {
				result.Errors = append(result.Errors, lintIssue{"error", relativePath(root, path), fmt.Sprintf("Unresolved conflict marker at line %d.", i+1)})
			}
		}
		return nil
	})
}

func skillHash(skillDir string) (string, error) {
	info, err := os.Stat(skillDir)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("missing skill dir")
	}
	lines := []string{}
	err = filepath.WalkDir(skillDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		hash := sha256.Sum256(content)
		rel, _ := filepath.Rel(skillDir, path)
		lines = append(lines, filepath.ToSlash(rel)+"\t"+hex.EncodeToString(hash[:]))
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(lines)
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:]), nil
}

func shortHash(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}

func knownSkillRoots(root, marketplaceRoot string, config marketplaceConfig) []string {
	roots := []string{}
	if config.ProjectLocalPaths.Skills != "" {
		roots = append(roots, resolveRepoPath(root, config.ProjectLocalPaths.Skills))
	}
	if config.Marketplace.Directories.ApprovedSkills != "" {
		roots = append(roots, resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.ApprovedSkills))
	}
	if userProfile, err := os.UserHomeDir(); err == nil && userProfile != "" {
		for _, rel := range []string{".codex/skills", ".copilot/skills", ".agents/skills"} {
			roots = append(roots, filepath.Join(userProfile, filepath.FromSlash(rel)))
		}
	}
	for _, additional := range config.QuarantineFirebreak.AdditionalActiveSkillRoots {
		if additional != "" {
			roots = append(roots, resolveRepoPath(root, additional))
		}
	}
	seen := map[string]bool{}
	unique := []string{}
	for _, path := range roots {
		key := normalizePathText(path)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, path)
		}
	}
	return unique
}

func testSkillRoot(rootPath, quarantinePath string, issues *[]marketplaceIssue) {
	if pathUnder(rootPath, quarantinePath) {
		*issues = append(*issues, marketplaceIssue{"active-root-in-quarantine", rootPath, "Active or approved skill roots must never point into marketplace quarantine."})
	}
	info, err := os.Stat(rootPath)
	if err != nil || !info.IsDir() {
		return
	}
	items := []string{rootPath}
	entries, _ := os.ReadDir(rootPath)
	for _, entry := range entries {
		if entry.IsDir() {
			items = append(items, filepath.Join(rootPath, entry.Name()))
		}
	}
	for _, item := range items {
		if pathUnder(item, quarantinePath) {
			*issues = append(*issues, marketplaceIssue{"skill-path-in-quarantine", item, "A loadable skill path is inside marketplace quarantine."})
		}
		if target, err := os.Readlink(item); err == nil && target != "" {
			targetPath := target
			if !filepath.IsAbs(targetPath) {
				targetPath = filepath.Join(filepath.Dir(item), targetPath)
			}
			if pathUnder(targetPath, quarantinePath) {
				*issues = append(*issues, marketplaceIssue{"skill-link-target-in-quarantine", item, "A loadable skill directory links into marketplace quarantine."})
			}
		}
	}
}

func testApprovedCatalog(marketplaceRoot, catalogPath, approvedSkillsPath, quarantinePath string, issues *[]marketplaceIssue) {
	var catalog struct {
		Skills []map[string]any `json:"skills"`
	}
	if err := readJSONFile(catalogPath, &catalog); err != nil {
		return
	}
	for _, skill := range catalog.Skills {
		name := stringProperty(skill, "name")
		if status := stringProperty(skill, "status"); status != "" && status != "approved" {
			*issues = append(*issues, marketplaceIssue{"approved-catalog-status", catalogPath, fmt.Sprintf("Approved catalog entry '%s' has non-approved status '%s'.", name, status)})
		}
		for _, field := range []string{"marketplacePath", "localPath", "path"} {
			value := stringProperty(skill, field)
			if value == "" {
				continue
			}
			resolved := resolveMarketplacePath(marketplaceRoot, value)
			if pathUnder(resolved, quarantinePath) {
				*issues = append(*issues, marketplaceIssue{"approved-catalog-quarantine-path", resolved, fmt.Sprintf("Approved catalog entry '%s' resolves field '%s' into quarantine.", name, field)})
			}
			if field == "marketplacePath" && !pathUnder(resolved, approvedSkillsPath) {
				*issues = append(*issues, marketplaceIssue{"approved-catalog-outside-approved-skills", resolved, fmt.Sprintf("Approved catalog entry '%s' must resolve marketplacePath under approved skills.", name)})
			}
		}
		if source, ok := skill["source"].(map[string]any); ok {
			if sourcePath := stringProperty(source, "path"); sourcePath != "" {
				resolved := resolveMarketplacePath(marketplaceRoot, sourcePath)
				if pathUnder(resolved, quarantinePath) {
					*issues = append(*issues, marketplaceIssue{"approved-source-in-quarantine", resolved, fmt.Sprintf("Approved catalog entry '%s' cannot load from quarantine as its source path.", name)})
				}
			}
		}
	}
}

func testQuarantineCatalog(catalogPath string, issues *[]marketplaceIssue) {
	var catalog struct {
		Skills []map[string]any `json:"skills"`
	}
	if err := readJSONFile(catalogPath, &catalog); err != nil {
		return
	}
	for _, skill := range catalog.Skills {
		if strings.EqualFold(stringProperty(skill, "status"), "approved") {
			name := stringProperty(skill, "name")
			*issues = append(*issues, marketplaceIssue{"quarantine-entry-approved", catalogPath, fmt.Sprintf("Quarantine catalog entry '%s' is marked approved; move it to the approved catalog after review instead.", name)})
		}
	}
}

func stringProperty(object map[string]any, name string) string {
	if value, ok := object[name]; ok && value != nil {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

func normalizePathText(path string) string {
	abs, _ := filepath.Abs(path)
	return strings.ToLower(strings.TrimRight(filepath.ToSlash(abs), "/"))
}

func pathUnder(path, parent string) bool {
	pathText := normalizePathText(path)
	parentText := normalizePathText(parent)
	return pathText == parentText || strings.HasPrefix(pathText, parentText+"/")
}

func sortedMapKeys(values map[string]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func writeJSON(w io.Writer, value any) {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}
