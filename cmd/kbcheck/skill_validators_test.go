package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillLintPassesValidSkillAndFailsBadFrontmatter(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

// TestSkillSyncReportSeparatesUncommittedSourceFromRealDrift pins the two cases
// a working-tree-only comparison cannot tell apart. Before HEAD awareness, both
// reported drift-required, so an uncommitted edit raised a sync failure that no
// sync could clear and real staleness was indistinguishable from it.
func TestSkillSyncReportSeparatesUncommittedSourceFromRealDrift(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	required := filepath.ToSlash(filepath.Join(root, "required"))
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), `{
	  "sync_targets": [
	    {"id":"source","path":".github/skills","classification":"source","required":true},
	    {"id":"required","path":"`+required+`","classification":"required","required":true}
	  ]
	}`)
	skillFile := filepath.Join(root, ".github", "skills", "demo", "SKILL.md")
	writeFile(t, skillFile, "committed content\n")

	gitOK(t, root, "init")
	gitOK(t, root, "config", "user.email", "test@example.com")
	gitOK(t, root, "config", "user.name", "Sync Report Test")
	gitOK(t, root, "add", ".")
	gitOK(t, root, "commit", "-m", "fixture")

	// The target holds exactly what was committed, while the source carries an
	// edit that has been released to nowhere.
	writeFile(t, filepath.Join(root, "required", "demo", "SKILL.md"), "committed content\n")
	writeFile(t, skillFile, "uncommitted local edit\n")

	result, err := computeSkillSyncReport(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeSkillSyncReport returned error: %v", err)
	}
	if !result.OK || result.RequiredIssues != 0 {
		t.Fatalf("target matching HEAD must not be drift, got %#v", result)
	}
	if status := syncStatusFor(t, result, "required"); status != "synced-at-head" {
		t.Fatalf("expected synced-at-head, got %q", status)
	}

	// A target matching neither HEAD nor the working tree is genuinely stale and
	// must still fail, or the fix would have silenced the check it repairs.
	writeFile(t, filepath.Join(root, "required", "demo", "SKILL.md"), "genuinely stale\n")
	result, err = computeSkillSyncReport(root, "config/skill-quality.json")
	if err != nil {
		t.Fatalf("computeSkillSyncReport returned error: %v", err)
	}
	if result.OK || result.RequiredIssues != 1 {
		t.Fatalf("real drift must still fail, got %#v", result)
	}
	if status := syncStatusFor(t, result, "required"); status != "drift-required" {
		t.Fatalf("expected drift-required, got %q", status)
	}
}

func syncStatusFor(t *testing.T, result skillSyncResult, target string) string {
	t.Helper()
	for _, row := range result.Rows {
		if row.Target == target {
			return row.Status
		}
	}
	t.Fatalf("no row for target %q in %#v", target, result.Rows)
	return ""
}

func TestSkillHashIgnoresRuntimeCachesButDetectsSourceChanges(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	target := filepath.Join(root, "target")
	writeFile(t, filepath.Join(source, "SKILL.md"), "same source\n")
	writeFile(t, filepath.Join(target, "SKILL.md"), "same source\n")
	writeFile(t, filepath.Join(target, "scripts", "__pycache__", "helper.cpython-311.pyc"), "generated bytecode")
	writeFile(t, filepath.Join(target, ".DS_Store"), "generated metadata")
	writeFile(t, filepath.Join(target, "Thumbs.db"), "generated metadata")

	sourceHash, err := skillHash(source)
	if err != nil {
		t.Fatalf("hash source: %v", err)
	}
	targetHash, err := skillHash(target)
	if err != nil {
		t.Fatalf("hash target: %v", err)
	}
	if sourceHash != targetHash {
		t.Fatalf("runtime cache changed skill hash: source=%s target=%s", sourceHash, targetHash)
	}

	writeFile(t, filepath.Join(target, "scripts", "helper.py"), "print('real source')\n")
	targetHash, err = skillHash(target)
	if err != nil {
		t.Fatalf("rehash target: %v", err)
	}
	if sourceHash == targetHash {
		t.Fatal("real source addition did not change skill hash")
	}
}

func TestDoctorRepairsMarkedStaleRequiredTarget(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

func TestDoctorRefreshesStaleMarkerOnMatchingTarget(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	source := filepath.Join(sourceRoot, "demo")
	requiredRoot := filepath.Join(root, "required")
	required := filepath.Join(requiredRoot, "demo")
	config := doctorTestConfig(t, root, sourceRoot, requiredRoot)

	// A target synced by some route that does not maintain the marker: content
	// is in step with source, but the marker still records an older source.
	writeFile(t, filepath.Join(source, "SKILL.md"), "v0\n")
	staleHash, err := skillHash(source)
	if err != nil {
		t.Fatalf("hash v0 source: %v", err)
	}
	writeFile(t, filepath.Join(source, "SKILL.md"), "v1\n")
	writeFile(t, filepath.Join(required, "SKILL.md"), "v1\n")
	if err := writeSyncMarker(requiredRoot, "demo", staleHash); err != nil {
		t.Fatalf("write stale marker: %v", err)
	}

	matched, err := computeDoctor(root, config, true)
	if err != nil {
		t.Fatalf("computeDoctor match returned error: %v", err)
	}
	if !matched.OK {
		t.Fatalf("expected matching target to pass, got %#v", matched)
	}
	v1Hash, err := skillHash(source)
	if err != nil {
		t.Fatalf("hash v1 source: %v", err)
	}
	if got := readSyncMarker(requiredRoot, "demo"); got != v1Hash {
		t.Fatalf("marker was not refreshed on match: got %s, want %s", got, v1Hash)
	}

	// The refreshed marker is what lets the next source advance be recognized as
	// a safe forward sync rather than unknown downstream drift.
	writeFile(t, filepath.Join(source, "SKILL.md"), "v2\n")
	fixed, err := computeDoctor(root, config, true)
	if err != nil {
		t.Fatalf("computeDoctor fix returned error: %v", err)
	}
	if !fixed.OK || fixed.Fixed != 1 || fixed.Refused != 0 {
		t.Fatalf("expected safe repair after marker refresh, got %#v", fixed)
	}
	sourceHash, _ := skillHash(source)
	targetHash, _ := skillHash(required)
	if targetHash != sourceHash {
		t.Fatalf("target hash %s did not match source %s", targetHash, sourceHash)
	}
}

func TestDoctorStillRefusesRealDownstreamDriftDespiteMarkerRefresh(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	sourceRoot := filepath.Join(root, "source")
	source := filepath.Join(sourceRoot, "demo")
	requiredRoot := filepath.Join(root, "required")
	required := filepath.Join(requiredRoot, "demo")
	config := doctorTestConfig(t, root, sourceRoot, requiredRoot)

	writeFile(t, filepath.Join(source, "SKILL.md"), "v1\n")
	writeFile(t, filepath.Join(required, "SKILL.md"), "v1\n")
	if _, err := computeDoctor(root, config, true); err != nil {
		t.Fatalf("computeDoctor seed returned error: %v", err)
	}

	// Both sides move: the marker now matches neither, which is genuine
	// two-sided drift and must still be refused rather than overwritten.
	writeFile(t, filepath.Join(source, "SKILL.md"), "v2\n")
	writeFile(t, filepath.Join(required, "SKILL.md"), "downstream-only\n")
	result, err := computeDoctor(root, config, true)
	if err != nil {
		t.Fatalf("computeDoctor returned error: %v", err)
	}
	if result.OK || result.Refused != 1 || result.Fixed != 0 {
		t.Fatalf("expected refusal on two-sided drift, got %#v", result)
	}
	if got, _ := os.ReadFile(filepath.Join(required, "SKILL.md")); string(got) != "downstream-only\n" {
		t.Fatalf("refused repair still overwrote the target: %q", string(got))
	}
}

func TestResolveRepoPathExpandsHome(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
