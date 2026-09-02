package main

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestReadmeIsFocusedProductFrontDoor(t *testing.T) {
	t.Parallel()
	root := communicationContractRepoRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(content)
	normalized := strings.TrimSuffix(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lineCount := len(strings.Split(normalized, "\n"))
	if lineCount < 350 || lineCount > 500 {
		t.Fatalf("README.md has %d lines; expected 350-500", lineCount)
	}

	diagrams := []string{
		"docs/assets/kb-workflow-overview.png",
		"docs/assets/kb-model-selection.png",
		"docs/assets/kb-memory-loop.png",
	}
	for _, required := range []string{
		"Portable workflow skills",
		"**Status:** actively used, pre-1.0",
		"KB is for developers",
		"You do not need Go",
		"## Start Here",
		"npx github:Irtechie/working-skill-repo --target all --profile core",
		"## The Six-Skill Loop",
		"## How the Workflow Fits Together",
		"## One Concrete Example",
		`kb-complete "Add CSV export to the invoice list, preserving the current filters"`,
		"## Task Routing",
		"## Planning and Execution",
		"## Model Routing with DDR",
		"DDR is the normal production routing contract",
		"It never routes below the planned tier",
		"A fresh bounded local-route evaluation remains future evidence",
		"## Project Mapping, Memory, and Handoffs",
		"## Verification Without Theater",
		"## Review, Delivery, and Recovery",
		"## What Makes KB Different",
		"## What Is in This Repository",
		"## Maintainer Commands",
		"## Platform and Security Reality",
		"## Read More",
		"## Credits",
		"kb-start",
		"kb-map",
		"kb-fix",
		"kb-plan",
		"kb-work",
		"kb-complete",
		"docs/README.md",
		"docs/context/architecture/skills.md",
		"docs/context/architecture/kb-workflow.md",
		"docs/context/operations/testing.md",
		"docs/context/eval-map.md",
		"docs/context/operations/skill-bundle-maintenance.md",
		"docs/context/architecture/private-skill-marketplace.md",
		"docs/context/research/2026-07-29-graph-engineering-definition-and-provenance.md",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("README.md missing product-front-door content %q", required)
		}
	}
	targets := regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)\)`).FindAllStringSubmatch(text, -1)
	actualDiagrams := make([]string, 0, len(targets))
	for _, target := range targets {
		actualDiagrams = append(actualDiagrams, target[1])
	}
	sort.Strings(diagrams)
	sort.Strings(actualDiagrams)
	if strings.Join(actualDiagrams, "\n") != strings.Join(diagrams, "\n") {
		t.Fatalf("README.md image targets mismatch\nactual:\n%s\nexpected:\n%s", strings.Join(actualDiagrams, "\n"), strings.Join(diagrams, "\n"))
	}
	for _, diagram := range actualDiagrams {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(diagram)))
		if err != nil {
			t.Errorf("README.md diagram %q does not resolve: %v", diagram, err)
		} else if !info.Mode().IsRegular() {
			t.Errorf("README.md diagram %q is not a regular file", diagram)
		}
	}

	for _, forbidden := range []string{
		"Graph-Compatible Workflow Milestones",
		"DDR route: <current|subagent> | primary:",
		"`pr-review-artifacts` branch",
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("README.md contains maintainer-only detail %q", forbidden)
		}
	}

	for _, pattern := range []string{`(?i)[a-z]:[\\/](?:users|dev)[\\/]`, `(?i)/users/[a-z0-9._-]+/`} {
		if regexp.MustCompile(pattern).MatchString(text) {
			t.Errorf("README.md contains a machine-private path matching %q", pattern)
		}
	}
}

func TestTrackedTextContainsNoMachinePrivatePaths(t *testing.T) {
	t.Parallel()
	root := communicationContractRepoRoot(t)
	patterns := machinePrivatePathPatterns()
	for path, content := range trackedTextFiles(t, root) {
		for _, pattern := range patterns {
			if pattern.Match(content) {
				t.Errorf("%s contains a machine-private path matching %q", path, pattern)
			}
		}
	}
}

func TestMachinePrivatePathPatternsIncludeEscapedWindowsPaths(t *testing.T) {
	t.Parallel()
	separator := strings.Repeat("\\", 2)
	examples := []string{
		"C:" + separator + "Users" + separator + "private-user" + separator + "secret",
		"E:" + separator + "Dev" + separator + "Tools",
		"prefix" + separator + "session-state" + separator + "run",
	}
	for _, example := range examples {
		matched := false
		for _, pattern := range machinePrivatePathPatterns() {
			if pattern.MatchString(example) {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("escaped machine-private path was not detected: %q", example)
		}
	}
}

func machinePrivatePathPatterns() []*regexp.Regexp {
	return []*regexp.Regexp{
		regexp.MustCompile(`(?i)[a-z]:[\\/]{1,2}(?:users|dev)[\\/]{1,2}`),
		regexp.MustCompile(`(?i)/users/[a-z0-9._-]+/`),
		regexp.MustCompile(`(?i)[\\/]{1,2}session-state[\\/]{1,2}`),
	}
}

func TestResearchIndexMatchesRetainedNotes(t *testing.T) {
	t.Parallel()
	root := communicationContractRepoRoot(t)
	researchDir := filepath.Join(root, "docs", "context", "research")
	entries, err := os.ReadDir(researchDir)
	if err != nil {
		t.Fatal(err)
	}
	var actual []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "README.md" {
			actual = append(actual, entry.Name())
		}
	}
	sort.Strings(actual)

	index, err := os.ReadFile(filepath.Join(researchDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	matches := regexp.MustCompile("`([^`]+\\.md)`").FindAllSubmatch(index, -1)
	indexed := make([]string, 0, len(matches))
	seen := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		name := string(match[1])
		if _, duplicate := seen[name]; duplicate {
			t.Fatalf("research index contains duplicate %q", name)
		}
		seen[name] = struct{}{}
		indexed = append(indexed, name)
	}
	sort.Strings(indexed)
	if strings.Join(indexed, "\n") != strings.Join(actual, "\n") {
		t.Fatalf("research index mismatch\nindexed:\n%s\nactual:\n%s", strings.Join(indexed, "\n"), strings.Join(actual, "\n"))
	}
}

func trackedTextFiles(t *testing.T, root string) map[string][]byte {
	t.Helper()
	code, output := runGitCommand(root, "ls-files", "-z")
	if code != 0 {
		t.Fatalf("git ls-files failed: %s", output)
	}
	files := make(map[string][]byte)
	for _, path := range strings.Split(output, "\x00") {
		if path == "" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatalf("read tracked file %s: %v", path, err)
		}
		if bytes.IndexByte(content, 0) >= 0 {
			continue
		}
		files[path] = content
	}
	return files
}
