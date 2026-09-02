package main

import (
	"os"
	"path/filepath"
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
	lineCount := len(strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n"))
	if lineCount >= 200 {
		t.Fatalf("README.md has %d lines; expected fewer than 200", lineCount)
	}

	for _, required := range []string{
		"Portable workflow skills",
		"**Status:** actively used, pre-1.0",
		"KB is for developers",
		"You do not need Go",
		"## Install",
		"npx github:Irtechie/working-skill-repo --target all --profile core",
		"## The Six-Skill Loop",
		"## One Concrete Example",
		`kb-complete "Add CSV export to the invoice list, preserving the current filters"`,
		"For solo-owner plan-to-accepted-PR delivery",
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

	for _, forbidden := range []string{
		"Graph-Compatible Workflow Milestones",
		`C:\Users\`,
		`E:\Dev\Tools\`,
		"DDR route: <current|subagent> | primary:",
		"`pr-review-artifacts` branch",
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("README.md contains maintainer-only detail %q", forbidden)
		}
	}
}
