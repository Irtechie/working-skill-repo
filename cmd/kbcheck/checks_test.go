package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverPackageChecks(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "package.json"), `{"scripts":{"lint":"eslint .","test":"vitest","unused":"noop"}}`)
	writeFile(t, filepath.Join(root, "pnpm-lock.yaml"), "")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"lint", "test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
	if checks[0].Args[0] != "pnpm" {
		t.Fatalf("expected pnpm runner, got %v", checks[0].Args)
	}
}

func TestDiscoverSkillRepoChecksIncludesNativeValidators(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-check", "SKILL.md"), "---\nname: kb-check\ndescription: test\n---\n")
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{
		"atv-upstream-delta",
		"atv-upstream-delta-selftest",
		"cross-model-benchmark-validate",
		"kb-pipeline-selftest",
		"kb-release-gate-selftest",
		"kb-work-ready-set-selftest",
		"kb-work-scope-lease-selftest",
		"marketplace-promotion-selftest",
		"route-complexity-eval",
		"skill-eval",
		"skill-eval-baseline-selftest",
		"skill-eval-codex-dry-run",
		"skill-eval-ghcp-dry-run",
		"skill-eval-manifest-selftest",
		"skill-eval-observed-trace-dry-run",
		"skill-eval-quality",
		"skill-lint",
		"skill-marketplace-firebreak",
		"skill-marketplace-firebreak-selftest",
		"skill-surface-minimality",
		"skill-surface-minimality-selftest",
		"skill-surface-report",
		"skill-sync-report",
		"workflow-governor-selftest",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
}

func TestDiscoverNestedDotnetProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "src", "App", "App.csproj"), "<Project></Project>")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"dotnet-test"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("checks=%v want=%v", got, want)
	}
}

func checkNames(checks []Check) []string {
	names := make([]string, 0, len(checks))
	for _, check := range checks {
		names = append(names, check.Name)
	}
	return names
}
