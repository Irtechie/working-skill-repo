package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStartupIntentContract keeps a direct question from being turned into a
// repository mutation just because project memory is incomplete. It protects
// the route boundary, not incidental prose.
func TestStartupIntentContract(t *testing.T) {
	t.Parallel()
	root := communicationContractRepoRoot(t)
	start := readStartupContract(t, root, ".github/skills/kb-start/SKILL.md")
	mapSkill := readStartupContract(t, root, ".github/skills/kb-map/SKILL.md")
	readOnly := readStartupContract(t, root, ".github/skills/kb-start/references/read-only-startup.md")
	mutating := readStartupContract(t, root, ".github/skills/kb-start/references/mutating-startup.md")
	hygiene := readStartupContract(t, root, ".github/skills/kb-start/references/session-hygiene.md")

	for _, want := range []string{
		"classify the request as **read-only**, **mutating**",
		"references/read-only-startup.md",
		"references/mutating-startup.md",
		"references/session-hygiene.md",
	} {
		if !strings.Contains(start, want) {
			t.Errorf("kb-start is missing startup-intent boundary %q", want)
		}
	}
	for _, forbidden := range []string{"cleanup, queue claims, bootstrap, refresh, session preservation", "Do not create directories"} {
		if !strings.Contains(readOnly, forbidden) {
			t.Errorf("read-only startup no longer forbids %q", forbidden)
		}
	}
	for _, want := range []string{"queue claim before mutation", "Missing", "kb-map-bootstrap", "terminal-cleanup --action sweep", "fails closed"} {
		if !strings.Contains(mutating, want) {
			t.Errorf("mutating startup lost required guard %q", want)
		}
	}
	if !strings.Contains(hygiene, "only at an actual restart decision") || !strings.Contains(hygiene, "session-preserve --action apply") {
		t.Error("session hygiene is not conditional or does not retain preservation guidance")
	}
	for _, want := range []string{
		"During a read-only lookup, do not bootstrap or create layout paths",
		"explicit `setup`, `refresh`, or mutating execution",
		"requires `kb-map-bootstrap` before the operation continues",
	} {
		if !containsStartupContract(mapSkill, want) {
			t.Errorf("kb-map lost intent-sensitive missing-memory behavior %q", want)
		}
	}
}

func containsStartupContract(content, want string) bool {
	return strings.Contains(strings.Join(strings.Fields(content), " "), strings.Join(strings.Fields(want), " "))
}

func TestStartupIntentReferencesResolve(t *testing.T) {
	t.Parallel()
	root := communicationContractRepoRoot(t)
	for _, name := range []string{"read-only-startup.md", "mutating-startup.md", "session-hygiene.md"} {
		path := filepath.Join(root, ".github", "skills", "kb-start", "references", name)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			t.Fatalf("kb-start reference %s is not installable: %v", name, err)
		}
	}
}

func readStartupContract(t *testing.T, root, relative string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}
