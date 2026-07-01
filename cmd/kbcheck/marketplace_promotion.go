package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type marketplacePromoteResult struct {
	OK              bool             `json:"ok"`
	SkillID         string           `json:"skill_id"`
	Source          string           `json:"source"`
	MarketplacePath string           `json:"marketplace_path"`
	CatalogPath     string           `json:"catalog_path"`
	SHA256          string           `json:"sha256"`
	InstallTargets  []promoteSyncRow `json:"install_targets"`
	Firebreak       string           `json:"firebreak"`
}

type promoteSyncRow struct {
	Target string `json:"target"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type approvedCatalog struct {
	SchemaVersion int                    `json:"schemaVersion"`
	Skills        []approvedCatalogEntry `json:"skills"`
}

type approvedCatalogEntry struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"`
	Source          map[string]string      `json:"source"`
	MarketplacePath string                 `json:"marketplacePath"`
	SHA256          string                 `json:"sha256"`
	ApprovedBy      string                 `json:"approvedBy"`
	ApprovedAt      string                 `json:"approvedAt"`
	ApprovalReason  string                 `json:"approvalReason"`
	Evidence        map[string][]string    `json:"evidence"`
	Extra           map[string]interface{} `json:"-"`
}

func runMarketplacePromoteCommand(root string, opts options, stdout, stderr io.Writer) int {
	configPath := opts.config
	if configPath == "" {
		configPath = "config/skill-marketplace.json"
	}
	result, err := promoteMarketplaceSkill(root, configPath, opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if opts.json {
		writeJSON(stdout, result)
	} else {
		fmt.Fprintf(stdout, "Promoted marketplace skill: %s\n", result.SkillID)
		fmt.Fprintf(stdout, "marketplace=%s\n", result.MarketplacePath)
		fmt.Fprintf(stdout, "sha256=%s\n", result.SHA256)
		if len(result.InstallTargets) == 0 {
			fmt.Fprintln(stdout, "synced=none")
		}
		for _, target := range result.InstallTargets {
			fmt.Fprintf(stdout, "synced=%s path=%s\n", target.Target, target.Path)
		}
		fmt.Fprintf(stdout, "firebreak=%s\n", result.Firebreak)
	}
	return 0
}

func promoteMarketplaceSkill(root, configPath string, opts options) (marketplacePromoteResult, error) {
	var result marketplacePromoteResult
	if !opts.approved {
		return result, fmt.Errorf("promotion requires explicit --approved after human review")
	}
	if opts.source == "" {
		return result, fmt.Errorf("marketplace-promote requires --source")
	}
	if strings.TrimSpace(opts.approvalReason) == "" {
		return result, fmt.Errorf("marketplace-promote requires --approval-reason")
	}
	var config marketplaceConfig
	configFullPath := resolveRepoPath(root, configPath)
	if err := readJSONFile(configFullPath, &config); err != nil {
		return result, fmt.Errorf("marketplace config not found: %s", configFullPath)
	}
	marketplaceRoot := resolveRepoPath(root, config.Marketplace.LocalRoot)
	approvedSkillsPath := resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.ApprovedSkills)
	approvedCatalogPath := resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.ApprovedCatalog)
	quarantinePath := resolveMarketplacePath(marketplaceRoot, config.Marketplace.Directories.Quarantine)
	sourcePath := resolveRepoPath(root, opts.source)
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return result, fmt.Errorf("source skill path not found: %s", sourcePath)
	}
	sourceDir := sourcePath
	if !sourceInfo.IsDir() {
		sourceDir = filepath.Dir(sourcePath)
	}
	sourceSkillFile := filepath.Join(sourceDir, "SKILL.md")
	content, err := os.ReadFile(sourceSkillFile)
	if err != nil {
		return result, fmt.Errorf("source skill is missing SKILL.md: %s", sourceSkillFile)
	}
	frontmatter := extractFrontmatter(string(content))
	declaredName := frontmatterValue(frontmatter, "name")
	declaredDescription := frontmatterValue(frontmatter, "description")
	if declaredName == "" {
		return result, fmt.Errorf("source SKILL.md is missing frontmatter field 'name'")
	}
	if declaredDescription == "" {
		return result, fmt.Errorf("source SKILL.md is missing frontmatter field 'description'")
	}
	skillID := opts.skillID
	if skillID == "" {
		skillID = filepath.Base(sourceDir)
	}
	if declaredName != skillID {
		return result, fmt.Errorf("source SKILL.md frontmatter name %q does not match skill id %q", declaredName, skillID)
	}
	destinationDir := resolveMarketplacePath(marketplaceRoot, filepath.Join(config.Marketplace.Directories.ApprovedSkills, skillID))
	if pathUnder(destinationDir, quarantinePath) {
		return result, fmt.Errorf("refusing to place approved skill under quarantine: %s", destinationDir)
	}
	if err := copySkillDirectory(sourceDir, destinationDir, approvedSkillsPath); err != nil {
		return result, err
	}
	destinationSkillFile := filepath.Join(destinationDir, "SKILL.md")
	sha, err := fileSHA256(destinationSkillFile)
	if err != nil {
		return result, err
	}
	if err := updateApprovedCatalog(approvedCatalogPath, approvedCatalogEntry{
		ID:              skillID,
		Status:          "approved",
		Source:          promotionSource(opts, sourceDir),
		MarketplacePath: filepath.ToSlash(filepath.Join("skills", skillID)),
		SHA256:          sha,
		ApprovedBy:      opts.approvedBy,
		ApprovedAt:      time.Now().Format("2006-01-02"),
		ApprovalReason:  opts.approvalReason,
		Evidence: map[string][]string{
			"proofCommands": {
				fmt.Sprintf("kbcheck marketplace-promote --source %s --skill-id %s --approved", filepath.ToSlash(sourceSkillFile), skillID),
				fmt.Sprintf("kbcheck marketplace-firebreak --config %s", filepath.ToSlash(configPath)),
			},
		},
	}); err != nil {
		return result, err
	}
	synced := []promoteSyncRow{}
	for _, target := range parseCSV(opts.installTargets) {
		if strings.EqualFold(target, "none") {
			continue
		}
		targetRoot, err := promoteTargetRoot(target, opts)
		if err != nil {
			return result, err
		}
		targetDir := filepath.Join(targetRoot, skillID)
		if pathUnder(targetDir, quarantinePath) {
			return result, fmt.Errorf("refusing to sync runtime skill target into quarantine: %s", targetDir)
		}
		if err := copySkillDirectory(destinationDir, targetDir, targetRoot); err != nil {
			return result, err
		}
		targetHash, err := fileSHA256(filepath.Join(targetDir, "SKILL.md"))
		if err != nil {
			return result, err
		}
		if targetHash != sha {
			return result, fmt.Errorf("hash mismatch after syncing %q: expected %s, got %s", target, sha, targetHash)
		}
		synced = append(synced, promoteSyncRow{Target: target, Path: targetDir, SHA256: targetHash})
	}
	firebreak, err := computeMarketplaceFirebreak(root, configPath)
	if err != nil {
		return result, err
	}
	if !firebreak.OK {
		return result, fmt.Errorf("marketplace firebreak failed after promotion: %d issue(s)", firebreak.IssueCount)
	}
	result = marketplacePromoteResult{
		OK: true, SkillID: skillID, Source: sourceDir, MarketplacePath: destinationDir,
		CatalogPath: approvedCatalogPath, SHA256: sha, InstallTargets: synced, Firebreak: "passed",
	}
	return result, nil
}

func promotionSource(opts options, sourceDir string) map[string]string {
	source := map[string]string{"type": opts.sourceType, "path": filepath.ToSlash(sourceDir)}
	if opts.upstreamRepo != "" {
		source["upstreamRepo"] = filepath.ToSlash(opts.upstreamRepo)
	}
	return source
}

func promoteTargetRoot(target string, opts options) (string, error) {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case "codex":
		return filepath.Abs(opts.codexSkillsRoot)
	case "copilot":
		return filepath.Abs(opts.copilotSkillsRoot)
	case "agents":
		return filepath.Abs(opts.agentsSkillsRoot)
	default:
		return "", fmt.Errorf("unknown install target %q; use codex,copilot,agents,none", target)
	}
}

func parseCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func copySkillDirectory(sourceDir, destinationDir, requiredParent string) error {
	if !pathUnder(destinationDir, requiredParent) {
		return fmt.Errorf("refusing to write outside approved skills path: %s", destinationDir)
	}
	if normalizePathText(sourceDir) == normalizePathText(destinationDir) {
		return nil
	}
	if err := os.RemoveAll(destinationDir); err != nil {
		return err
	}
	if err := os.MkdirAll(destinationDir, 0o755); err != nil {
		return err
	}
	return filepath.WalkDir(sourceDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(sourceDir, path)
		if err != nil || rel == "." {
			return err
		}
		dest := filepath.Join(destinationDir, rel)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, content, info.Mode())
	})
}

func updateApprovedCatalog(path string, entry approvedCatalogEntry) error {
	var catalog approvedCatalog
	if _, err := os.Stat(path); err == nil {
		if err := readJSONFile(path, &catalog); err != nil {
			return err
		}
	}
	if catalog.SchemaVersion == 0 {
		catalog.SchemaVersion = 1
	}
	next := []approvedCatalogEntry{}
	for _, existing := range catalog.Skills {
		if existing.ID != entry.ID {
			next = append(next, existing)
		}
	}
	next = append(next, entry)
	sort.Slice(next, func(i, j int) bool { return next[i].ID < next[j].ID })
	catalog.Skills = next
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}

func fileSHA256(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:]), nil
}

func runMarketplacePromoteSelftest(root string, stdout, stderr io.Writer) int {
	tmpParent := filepath.Join(root, ".kb", "tmp")
	if err := os.MkdirAll(tmpParent, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	tempRoot, err := os.MkdirTemp(tmpParent, "promote-marketplace-skill-selftest-*")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer os.RemoveAll(tempRoot)
	sourceRoot := filepath.Join(tempRoot, "source", "selftest-skill")
	marketplaceRoot := filepath.Join(tempRoot, "marketplace")
	globalRoot := filepath.Join(tempRoot, "globals", "codex")
	configPath := filepath.Join(tempRoot, "config.json")
	badConfigPath := filepath.Join(tempRoot, "bad-config.json")
	if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(sourceRoot, "SKILL.md"), []byte(`---
name: selftest-skill
description: Selftest fixture skill for marketplace promotion.
argument-hint: "[selftest]"
---

# Selftest Skill

Used only by marketplace promotion selftest.
`), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := writeMarketplaceSelftestConfig(configPath, marketplaceRoot, "skills"); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	opts := options{
		source: sourceRoot, skillID: "selftest-skill", approvalReason: "selftest approval",
		approvedBy: "selftest", sourceType: "selftest", installTargets: "codex",
		codexSkillsRoot: globalRoot, approved: true,
	}
	result, err := promoteMarketplaceSkill(root, configPath, opts)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	approvedHash, err := fileSHA256(filepath.Join(marketplaceRoot, "skills", "selftest-skill", "SKILL.md"))
	if err != nil || approvedHash != result.SHA256 {
		fmt.Fprintln(stderr, "approved skill hash mismatch")
		return 1
	}
	syncedHash, err := fileSHA256(filepath.Join(globalRoot, "selftest-skill", "SKILL.md"))
	if err != nil || syncedHash != result.SHA256 {
		fmt.Fprintln(stderr, "synced skill hash mismatch")
		return 1
	}
	if err := writeMarketplaceSelftestConfig(badConfigPath, marketplaceRoot, "quarantine/approved-skills"); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	_, err = promoteMarketplaceSkill(root, badConfigPath, opts)
	if err == nil {
		fmt.Fprintln(stderr, "expected promotion to fail when approved path is inside quarantine")
		return 1
	}
	fmt.Fprintln(stdout, "Marketplace promotion selftest: happy path promoted and synced; quarantine approved path correctly rejected.")
	return 0
}

func writeMarketplaceSelftestConfig(path, marketplaceRoot, approvedSkills string) error {
	config := map[string]any{
		"schema_version": 1,
		"marketplace": map[string]any{
			"id": "selftest-marketplace", "local_root": filepath.ToSlash(marketplaceRoot),
			"remote": "", "trust_model": "private-approved-catalog",
			"directories": map[string]string{
				"approved_skills": approvedSkills, "pipelines": "pipelines", "harnesses": "harnesses",
				"approved_catalog": "catalog/approved-skills.json", "quarantine_catalog": "catalog/quarantined-skills.json",
				"quarantine": "quarantine", "scripts": "scripts",
			},
		},
		"project_local_paths": map[string]string{"skills": ".github/skills"},
		"quarantine_firebreak": map[string]any{
			"never_load_from_quarantine": true, "additional_active_skill_roots": []string{},
		},
	}
	for _, dir := range []string{"catalog", "quarantine", "skills"} {
		_ = os.MkdirAll(filepath.Join(marketplaceRoot, filepath.FromSlash(dir)), 0o755)
	}
	_ = os.WriteFile(filepath.Join(marketplaceRoot, "catalog", "approved-skills.json"), []byte(`{"schemaVersion":1,"skills":[]}`), 0o644)
	_ = os.WriteFile(filepath.Join(marketplaceRoot, "catalog", "quarantined-skills.json"), []byte(`{"schemaVersion":1,"skills":[]}`), 0o644)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o644)
}
