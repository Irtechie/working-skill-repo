package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillGuidanceAcceptsCompleteCompactSurface(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeGuidanceFixture(t, root, "alpha", 4, "Use references/process.md when executing.")
	writeGuidanceFile(t, filepath.Join(root, ".github", "skills", "alpha", "references", "process.md"), "# Process\n\nShort.")
	writeGuidanceFile(t, filepath.Join(root, "config", "skill-guidance.json"), `{
  "max_skill_lines": 20,
  "max_hot_path_lines": 10,
  "hot_path": ["alpha"],
  "deprecated_skills": ["old-alias"],
  "audit_path": "config/skill-guidance-audit.json",
  "removed_skill_inventory_path": "config/removed-skills.json"
}`)
	writeGuidanceFile(t, filepath.Join(root, "config", "skill-guidance-audit.json"), `[
  {"name":"alpha","purpose":"bounded lane","callers":["user"],"routing_evidence":"explicit trigger","retained_capability":"runs alpha","disposition":"keep","proof":"fixture"}
]`)
	writeGuidanceFile(t, filepath.Join(root, "config", "removed-skills.json"), `["old-alias"]`)

	var stdout, stderr bytes.Buffer
	if code := runSkillGuidanceCommand(root, options{}, &stdout, &stderr); code != 0 {
		t.Fatalf("expected pass, code=%d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "skill guidance: ok") {
		t.Fatalf("missing success output: %s", stdout.String())
	}
}

func TestSkillGuidanceRejectsOversizeMissingAuditAndDeprecatedSkill(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeGuidanceFixture(t, root, "alpha", 12, "")
	writeGuidanceFixture(t, root, "old-alias", 3, "")
	writeGuidanceFile(t, filepath.Join(root, "config", "skill-guidance.json"), `{
  "max_skill_lines": 10,
  "max_hot_path_lines": 8,
  "hot_path": ["alpha"],
  "deprecated_skills": ["old-alias"],
  "audit_path": "config/skill-guidance-audit.json",
  "removed_skill_inventory_path": "config/removed-skills.json"
}`)
	writeGuidanceFile(t, filepath.Join(root, "config", "skill-guidance-audit.json"), `[
  {"name":"old-alias","purpose":"alias","callers":[],"routing_evidence":"none","retained_capability":"none","disposition":"remove","proof":"replacement"}
]`)
	writeGuidanceFile(t, filepath.Join(root, "config", "removed-skills.json"), `[]`)

	var stdout, stderr bytes.Buffer
	if code := runSkillGuidanceCommand(root, options{}, &stdout, &stderr); code == 0 {
		t.Fatal("expected guidance failures")
	}
	output := stderr.String()
	for _, want := range []string{"deprecated-skill-present: old-alias", "hot-path-too-long: alpha", "audit-row-missing: alpha", "removed-inventory-missing: old-alias"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in %s", want, output)
		}
	}
}

func TestSkillGuidanceRejectsNestedAndUnnavigableReferences(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeGuidanceFixture(t, root, "alpha", 4, "Load references/deep/process.md.")
	longReference := "# Process\n\n" + strings.Repeat("detail\n", 105)
	writeGuidanceFile(t, filepath.Join(root, ".github", "skills", "alpha", "references", "deep", "process.md"), longReference)
	writeGuidanceFile(t, filepath.Join(root, "config", "skill-guidance.json"), `{
  "max_skill_lines": 20,
  "max_hot_path_lines": 10,
  "hot_path": [],
  "deprecated_skills": [],
  "audit_path": "config/skill-guidance-audit.json",
  "removed_skill_inventory_path": "config/removed-skills.json"
}`)
	writeGuidanceFile(t, filepath.Join(root, "config", "skill-guidance-audit.json"), `[
  {"name":"alpha","purpose":"bounded lane","callers":["user"],"routing_evidence":"explicit trigger","retained_capability":"runs alpha","disposition":"keep","proof":"fixture"}
]`)
	writeGuidanceFile(t, filepath.Join(root, "config", "removed-skills.json"), `[]`)

	var stdout, stderr bytes.Buffer
	if code := runSkillGuidanceCommand(root, options{}, &stdout, &stderr); code == 0 {
		t.Fatal("expected reference failures")
	}
	output := stderr.String()
	for _, want := range []string{"nested-reference: alpha", "reference-missing-contents: alpha"} {
		if !strings.Contains(output, want) {
			t.Fatalf("missing %q in %s", want, output)
		}
	}
}

func writeGuidanceFixture(t *testing.T, root, name string, lines int, body string) {
	t.Helper()
	content := "---\nname: " + name + "\ndescription: focused skill\n---\n" + body + "\n"
	for strings.Count(content, "\n")+1 < lines {
		content += "detail\n"
	}
	writeGuidanceFile(t, filepath.Join(root, ".github", "skills", name, "SKILL.md"), content)
}

func writeGuidanceFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
