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

func TestDiscoverSkillRepoChecksIncludesBenchmarkValidator(t *testing.T) {
	t.Setenv("KBCHECK_POWERSHELL", "pwsh")
	root := t.TempDir()
	writeFile(t, filepath.Join(root, ".github", "skills", "kb-check", "SKILL.md"), "---\nname: kb-check\ndescription: test\n---\n")
	writeFile(t, filepath.Join(root, "config", "skill-quality.json"), "{}")
	writeFile(t, filepath.Join(root, "scripts", "cross-model-benchmark-validate.ps1"), "exit 0")

	checks, err := DiscoverChecks(root)
	if err != nil {
		t.Fatalf("DiscoverChecks returned error: %v", err)
	}
	got := checkNames(checks)
	want := []string{"cross-model-benchmark-validate"}
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
