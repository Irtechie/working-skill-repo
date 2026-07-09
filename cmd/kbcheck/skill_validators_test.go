package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillLintPassesValidSkillAndFailsBadFrontmatter(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
	  "lint": {
	    "skill_root": ".github/skills",
	    "require_argument_hint": "warning",
	    "required_frontmatter": ["name", "description"],
	    "scan_extensions": [".md"],
	    "hot_path_warning_lines": 10,
	    "hot_path_fail_lines": 20,
	    "allow_long_skills": {}
	  }
	}`)
	writeFile(t, filepath.Join(root, ".github", "skills", "good", "SKILL.md"), "---\nname: good\ndescription: ok\nargument-hint: test\n---\n# Good\n")

	result, err := computeSkillLint(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeSkillLint returned error: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected valid skill to pass: %#v", result.Errors)
	}

	writeFile(t, filepath.Join(root, ".github", "skills", "bad", "SKILL.md"), "---\nname: mismatch\n---\n# Bad\n")
	result, err = computeSkillLint(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeSkillLint returned error: %v", err)
	}
	if result.OK {
		t.Fatalf("expected bad frontmatter to fail, got %#v", result)
	}
	assertLintIssue(t, result.Errors, ".github/skills/bad/SKILL.md", "Missing required frontmatter field 'description'.")
	assertLintIssue(t, result.Errors, ".github/skills/bad/SKILL.md", "Frontmatter name 'mismatch' does not match folder 'bad'.")
	assertLintIssue(t, result.Warnings, ".github/skills/bad/SKILL.md", "Missing argument-hint frontmatter.")
}

func assertLintIssue(t *testing.T, issues []lintIssue, path, message string) {
	t.Helper()
	for _, issue := range issues {
		if issue.Path == path && issue.Message == message {
			return
		}
	}
	t.Fatalf("missing lint issue path=%q message=%q in %#v", path, message, issues)
}

func TestSkillSyncReportFindsRequiredDrift(t *testing.T) {
	root := t.TempDir()
	source := filepath.ToSlash(filepath.Join(root, "source"))
	required := filepath.ToSlash(filepath.Join(root, "required"))
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
	  "sync_targets": [
	    {"id":"source","path":"`+source+`","classification":"source","required":true},
	    {"id":"required","path":"`+required+`","classification":"required","required":true},
	    {"id":"optional","path":"missing-optional","classification":"optional","required":false}
	  ]
	}`)
	writeFile(t, filepath.Join(root, "source", "demo", "SKILL.md"), "source\n")
	writeFile(t, filepath.Join(root, "required", "demo", "SKILL.md"), "drift\n")

	result, err := computeSkillSyncReport(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeSkillSyncReport returned error: %v", err)
	}
	if result.OK || result.RequiredIssues != 1 {
		t.Fatalf("expected required drift, got %#v", result)
	}
}

func TestDoctorRepairsMarkedStaleRequiredTarget(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source", "demo")
	requiredRoot := filepath.Join(root, "required")
	required := filepath.Join(requiredRoot, "demo")
	writeFile(t, filepath.Join(source, "SKILL.md"), "v1\n")
	writeFile(t, filepath.Join(required, "SKILL.md"), "v1\n")
	oldHash, err := skillHash(source)
	if err != nil {
		t.Fatalf("hash old source: %v", err)
	}
	if err := writeSyncMarker(requiredRoot, "demo", oldHash); err != nil {
		t.Fatalf("write marker: %v", err)
	}
	writeFile(t, filepath.Join(source, "SKILL.md"), "v2\n")
	config := doctorTestConfig(t, root, filepath.Join(root, "source"), requiredRoot)

	report, err := computeDoctor(root, config, false)
	if err != nil {
		t.Fatalf("computeDoctor returned error: %v", err)
	}
	if report.OK || report.RequiredIssues != 1 {
		t.Fatalf("expected stale required issue, got %#v", report)
	}
	fixed, err := computeDoctor(root, config, true)
	if err != nil {
		t.Fatalf("computeDoctor fix returned error: %v", err)
	}
	if !fixed.OK || fixed.Fixed != 1 {
		t.Fatalf("expected repair to pass, got %#v", fixed)
	}
	sourceHash, _ := skillHash(source)
	targetHash, _ := skillHash(required)
	if targetHash != sourceHash {
		t.Fatalf("target hash %s did not match source %s", targetHash, sourceHash)
	}
}

func TestDoctorRefusesUnknownRequiredDrift(t *testing.T) {
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	requiredRoot := filepath.Join(root, "required")
	writeFile(t, filepath.Join(sourceRoot, "demo", "SKILL.md"), "source\n")
	writeFile(t, filepath.Join(requiredRoot, "demo", "SKILL.md"), "global-only-change\n")
	config := doctorTestConfig(t, root, sourceRoot, requiredRoot)

	result, err := computeDoctor(root, config, true)
	if err != nil {
		t.Fatalf("computeDoctor returned error: %v", err)
	}
	if result.OK || result.Refused != 1 || result.RequiredIssues != 1 {
		t.Fatalf("expected unknown drift refusal, got %#v", result)
	}
}

func TestResolveRepoPathExpandsHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		t.Skip("home directory unavailable")
	}
	got := resolveRepoPath(t.TempDir(), "~/.codex/skills")
	want := filepath.Join(home, ".codex", "skills")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func doctorTestConfig(t *testing.T, root, source, required string) string {
	t.Helper()
	config := filepath.Join(root, "config", "skill-quality.json")
	writeFile(t, config, `{
	  "sync_targets": [
	    {"id":"source","path":"`+filepath.ToSlash(source)+`","classification":"source","required":true},
	    {"id":"required","path":"`+filepath.ToSlash(required)+`","classification":"required","required":true}
	  ]
	}`)
	return config
}

func TestMarketplaceFirebreakFailsQuarantineActiveRoot(t *testing.T) {
	root := t.TempDir()
	market := filepath.ToSlash(filepath.Join(root, "market"))
	writeFile(t, filepath.Join(root, "config", "skill-marketplace.json"), `{
	  "marketplace": {
	    "local_root": "`+market+`",
	    "directories": {
	      "approved_skills": "skills",
	      "approved_catalog": "catalog/approved-skills.json",
	      "quarantine_catalog": "catalog/quarantined-skills.json",
	      "quarantine": "quarantine"
	    }
	  },
	  "project_local_paths": {"skills": ".github/skills"},
	  "quarantine_firebreak": {
	    "never_load_from_quarantine": true,
	    "additional_active_skill_roots": ["`+market+`/quarantine"]
	  }
	}`)

	result, err := computeMarketplaceFirebreak(root, "config/skill-marketplace.json")
	if err != nil {
		t.Fatalf("computeMarketplaceFirebreak returned error: %v", err)
	}
	if result.OK || result.IssueCount == 0 {
		t.Fatalf("expected quarantine active root failure, got %#v", result)
	}
}
